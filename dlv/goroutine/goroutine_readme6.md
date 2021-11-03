## 调度策略
    了解了main goroutine和普通goroutine的生死后，我们继续了解下具体的调度策略吧，如果我们有一堆的g，我该用什么的策略调度运行起来呢？

    下面围绕：调度策略、       调度时机、切换机制

```go
//在goroutine_readme4中大概我们有这几个地方涉及到查找可运行的g进行调度：

if gp == nil {
// Check the global runnable queue once in a while to ensure fairness.
// Otherwise two goroutines can completely occupy the local runqueue
// by constantly respawning each other.
//为了保证调度的公平性，每进行61次调度就需要优先从全局运行队列中获取goroutine，
//因为如果只调度本地队列中的g，那么全局运行队列中的goroutine将得不到运行
    if _g_.m.p.ptr().schedtick%61 == 0 && sched.runqsize > 0 {
        lock(&sched.lock)//所有工作线程都能访问全局运行队列，所以需要加锁
        gp = globrunqget(_g_.m.p.ptr(), 1)//从全局运行队列中获取1个goroutine
        unlock(&sched.lock)
    }
}
if gp == nil {
    //从与m关联的p的runnext下一个g和runq本地运行队列中获取goroutine
    gp, inheritTime = runqget(_g_.m.p.ptr())
    // We can see gp != nil here even if the M is spinning,
    // if checkTimers added a local goroutine via goready.
}
if gp == nil {
    //重点来了，就是findrunnable函数
    //如果从本地运行队列和全局运行队列都没有找到需要运行的goroutine，
    //则调用findrunnable函数从其它工作线程的运行队列中偷取，如果偷取不到，则当前工作线程进入睡眠，
    //直到获取到需要运行的goroutine之后findrunnable函数才会返回。
    gp, inheritTime = findrunnable() // blocks until work is available
}
```

## globrunqget()

```go

// Try get a batch of G's from the global runnable queue.
// Sched must be locked.
//参数是当前p和需要获取的g个数
func globrunqget(_p_ *p, max int32) *g {
	if sched.runqsize == 0 {
		return nil
	}

	//根据p的数量平分全局运行队列中的goroutines
	//每个p大概可以有多少个g
	n := sched.runqsize/gomaxprocs + 1
	//上面计算n的方法可能导致n大于全局运行队列中的goroutine数量
	if n > sched.runqsize {
		n = sched.runqsize
	}
	//最多取max个goroutine
	if max > 0 && n > max {
		n = max
	}
	//如果要取的n个g都比本地队列的个数的一半都大（256/2 = 128），那就取一半吧
	if n > int32(len(_p_.runq))/2 {
		n = int32(len(_p_.runq)) / 2
	}

	//重新设置全局队列中的g数量
	sched.runqsize -= n

	//从队列中拿出g
	//直接通过函数返回gp，其它的goroutines通过runqput放入本地运行队列
	gp := sched.runq.pop()
	n--
	for ; n > 0; n-- {
        //全局运行队列中pop出一个goroutine
		gp1 := sched.runq.pop()
		//放入到本地队列
		runqput(_p_, gp1, false)
	}
	return gp
}
```

## runqget()

```go
// Get g from local runnable queue.
// If inheritTime is true, gp should inherit the remaining time in the
// current time slice. Otherwise, it should start a new time slice.
// Executed only by the owner P.
func runqget(_p_ *p) (gp *g, inheritTime bool) {
	// If there's a runnext, it's the next G to run.
	for {
		//从runnext成员中获取goroutine  runnext 只保存最新创建的一个g
		//并且运行也会从这里开始
		next := _p_.runnext
		if next == 0 {
			break
		}
		//不为空则返回该goroutine，把runnext成员清零
		if _p_.runnext.cas(next, 0) {
			return next.ptr(), true
		}
	}

	//从循环队列中获取goroutine，该队列最多可包含256个goroutine
	for {
		//队列中拿取goroutine都使用了cas操作
		//因为可能有其他工作线程此时此刻也正在访问这两个成员，从这里偷取可运行的goroutine
		//代码中对runqhead的操作使用了atomic.LoadAcq和atomic.CasRel，它们分别提供了load-acquire和cas-release语义
		h := atomic.LoadAcq(&_p_.runqhead) // load-acquire, synchronize with other consumers
		t := _p_.runqtail
		if t == h {
			return nil, false
		}
		gp := _p_.runq[h%uint32(len(_p_.runq))].ptr()
		if atomic.CasRel(&_p_.runqhead, h, h+1) { // cas-release, commits consume
			return gp, false
		}
	}
}

对于atomic.LoadAcq来说，其语义主要包含如下几条：

原子读取，也就是说不管代码运行在哪种平台，保证在读取过程中不会有其它线程对该变量进行写入；

位于atomic.LoadAcq之后的代码，对内存的读取和写入必须在atomic.LoadAcq读取完成后才能执行，编译器和CPU都不能打乱这个顺序；

当前线程执行atomic.LoadAcq时可以读取到其它线程最近一次通过atomic.CasRel对同一个变量写入的值，与此同时，位于atomic.LoadAcq之后的代码，不管读取哪个内存地址中的值，都可以读取到其它线程中位于atomic.CasRel（对同一个变量操作）之前的代码最近一次对内存的写入。

对于atomic.CasRel来说，其语义主要包含如下几条：

原子的执行比较并交换的操作；

位于atomic.CasRel之前的代码，对内存的读取和写入必须在atomic.CasRel对内存的写入之前完成，编译器和CPU都不能打乱这个顺序；

线程执行atomic.CasRel完成后其它线程通过atomic.LoadAcq读取同一个变量可以读到最新的值，与此同时，位于atomic.CasRel之前的代码对内存写入的值，可以被其它线程中位于atomic.LoadAcq（对同一个变量操作）之后的代码读取到。

因为可能有多个线程会并发的修改和读取runqhead，以及需要依靠runqhead的值来读取runq数组的元素，所以需要使用atomic.LoadAcq和atomic.CasRel来保证上述语义。

我们可能会问，为什么读取p的runqtail成员不需要使用atomic.LoadAcq或atomic.load？因为runqtail不会被其它线程修改，只会被当前工作线程修改，此时没有人修改它，所以也就不需要使用原子相关的操作。

CAS操作与ABA问题

我们知道使用cas操作需要特别注意ABA的问题，那么runqget函数这两个使用cas的地方会不会有问题呢？答案是这两个地方都不会有ABA的问题。原因分析如下：

首先来看对runnext的cas操作。只有跟_p_绑定的当前工作线程才会去修改runnext为一个非0值，其它线程只会把runnext的值从一个非0值修改为0值，然而跟_p_绑定的当前工作线程正在此处执行代码，所以在当前工作线程读取到值A之后，不可能有线程修改其值为B(0)之后再修改回A。

再来看对runq的cas操作。当前工作线程操作的是_p_的本地队列，只有跟_p_绑定在一起的当前工作线程才会因为往该队列里面添加goroutine而去修改runqtail，而其它工作线程不会往该队列里面添加goroutine，也就不会去修改runqtail，它们只会修改runqhead，所以，当我们这个工作线程从runqhead读取到值A之后，其它工作线程也就不可能修改runqhead的值为B之后再第二次把它修改为值A（因为runqtail在这段时间之内不可能被修改，runqhead的值也就无法越过runqtail再回绕到A值），也就是说，代码从逻辑上已经杜绝了引发ABA的条件。
```

## findrunnable()
    该函数相当复杂，它还做了与gc和netpoll等工作，而且该函数是在找不到需要运行g之前是不会返回的，没有g最后会休眠等待被唤醒。

//该函数分两部分
func findrunnable() {
top:
stop:
}

```go
// Finds a runnable goroutine to execute.
// Tries to steal from other P's, get g from local or global queue, poll network.
func findrunnable() (gp *g, inheritTime bool) {
	_g_ := getg()//_g_ = g0

	// The conditions here and in handoffp must agree: if
	// findrunnable would return a G to run, handoffp must start
	// an M.

top:
	//获取当前的p
	_p_ := _g_.m.p.ptr()
	//gc相关，暂时跳过
	if sched.gcwaiting != 0 {
		gcstopm()
		goto top
	}
	if _p_.runSafePointFn != 0 {
		runSafePointFn()
	}

	//检查时间堆，是否有需要执行的g
	now, pollUntil, _ := checkTimers(_p_, 0)

	//跳过
	if fingwait && fingwake {
		if gp := wakefing(); gp != nil {
			ready(gp, 0, true)
		}
	}
	//跳过
	if *cgo_yield != nil {
		asmcgocall(*cgo_yield, nil)
	}

	//double check
	// local runq
	//再次看一下本地运行队列是否有需要运行的goroutine
	if gp, inheritTime := runqget(_p_); gp != nil {
		return gp, inheritTime
	}

    //double check
	// global runq
	//再看看全局运行队列是否有需要运行的goroutine
	if sched.runqsize != 0 {
		lock(&sched.lock)
		gp := globrunqget(_p_, 0)
		unlock(&sched.lock)
		if gp != nil {
			return gp, false
		}
	}

	//跳过，网络IO相关
	// Poll network.
	// This netpoll is only an optimization before we resort to stealing.
	// We can safely skip it if there are no waiters or a thread is blocked
	// in netpoll already. If there is any kind of logical race with that
	// blocked thread (e.g. it has already returned from netpoll, but does
	// not set lastpoll yet), this thread will do blocking netpoll below
	// anyway.
	if netpollinited() && atomic.Load(&netpollWaiters) > 0 && atomic.Load64(&sched.lastpoll) != 0 {
		if list := netpoll(0); !list.empty() { // non-blocking
			gp := list.pop()
			injectglist(&list)
			casgstatus(gp, _Gwaiting, _Grunnable)
			if trace.enabled {
				traceGoUnpark(gp, 0)
			}
			return gp, false
		}
	}

	// Steal work from other P's.
	procs := uint32(gomaxprocs)
	ranTimer := false
	// If number of spinning M's >= number of busy P's, block.
	// This is necessary to prevent excessive CPU consumption
	// when GOMAXPROCS>>1 but the program parallelism is low.
	//如果工作的m比p还多，那就是满负荷情况下，不要去找g了，浪费cpu了
	//已经有足够多的工作线程正在寻找可运行的goroutine，让他们去找就好了，自己偷个懒去睡觉
	if !_g_.m.spinning && 2*atomic.Load(&sched.nmspinning) >= procs-atomic.Load(&sched.npidle) {
		goto stop
	}
	if !_g_.m.spinning {
		//设置m的状态为spinning，表示正在找g
		_g_.m.spinning = true
		//处于spinning状态的m数量加一
		atomic.Xadd(&sched.nmspinning, 1)
	}
	//从其它p的本地运行队列盗取goroutine
	//这里充满了设计
	for i := 0; i < 4; i++ {
		//随机，但不重复地从其他p获取g（偷）
		for enum := stealOrder.start(fastrand()); !enum.done(); enum.next() {
			if sched.gcwaiting != 0 {
				goto top
			}
			stealRunNextG := i > 2 // first look for ready queues with more than 1 g
			p2 := allp[enum.position()]
			if _p_ == p2 {
				continue
			}
			//偷g，如果成功就返回
			if gp := runqsteal(_p_, p2, stealRunNextG); gp != nil {
				return gp, false
			}

			// Consider stealing timers from p2.
			// This call to checkTimers is the only place where
			// we hold a lock on a different P's timers.
			// Lock contention can be a problem here, so avoid
			// grabbing the lock if p2 is running and not marked
			// for preemption. If p2 is running and not being
			// preempted we assume it will handle its own timers.
			//还偷时间堆上要执行的g
			if i > 2 && shouldStealTimers(p2) {
				tnow, w, ran := checkTimers(p2, now)
				now = tnow
				if w != 0 && (pollUntil == 0 || w < pollUntil) {
					pollUntil = w
				}
				if ran {
					// Running the timers may have
					// made an arbitrary number of G's
					// ready and added them to this P's
					// local run queue. That invalidates
					// the assumption of runqsteal
					// that is always has room to add
					// stolen G's. So check now if there
					// is a local G to run.
					if gp, inheritTime := runqget(_p_); gp != nil {
						return gp, inheritTime
					}
					ranTimer = true
				}
			}
		}
	}
	if ranTimer {
		// Running a timer may have made some goroutine ready.
		goto top
	}

stop:

	//gc相关，查找符合gc标识的g并进行调度运行
	// We have nothing to do. If we're in the GC mark phase, can
	// safely scan and blacken objects, and have work to do, run
	// idle-time marking rather than give up the P.
	if gcBlackenEnabled != 0 && _p_.gcBgMarkWorker != 0 && gcMarkWorkAvailable(_p_) {
		_p_.gcMarkWorkerMode = gcMarkWorkerIdleMode
		gp := _p_.gcBgMarkWorker.ptr()
		casgstatus(gp, _Gwaiting, _Grunnable)
		if trace.enabled {
			traceGoUnpark(gp, 0)
		}
		return gp, false
	}

	delta := int64(-1)
	if pollUntil != 0 {
		// checkTimers ensures that polluntil > now.
		delta = pollUntil - now
	}

	// wasm only:
	// If a callback returned and no other goroutine is awake,
	// then pause execution until a callback was triggered.
	if beforeIdle(delta) {
		// At least one goroutine got woken.
		goto top
	}

	// Before we drop our P, make a snapshot of the allp slice,
	// which can change underfoot once we no longer block
	// safe-points. We don't need to snapshot the contents because
	// everything up to cap(allp) is immutable.
	allpSnapshot := allp

	// return P and block
	lock(&sched.lock)
	if sched.gcwaiting != 0 || _p_.runSafePointFn != 0 {
		unlock(&sched.lock)
		goto top
	}
	//再次检查下全局队列是否有g
	if sched.runqsize != 0 {
		gp := globrunqget(_p_, 0)
		unlock(&sched.lock)
		return gp, false
	}
	//最后还是没有好吧，m解除p的绑定，把p放入闲置池中，然后m洗洗睡了
	if releasep() != _p_ {
		throw("findrunnable: wrong p")
	}
	pidleput(_p_)
	unlock(&sched.lock)

	// Delicate dance: thread transitions from spinning to non-spinning state,
	// potentially concurrently with submission of new goroutines. We must
	// drop nmspinning first and then check all per-P queues again (with
	// #StoreLoad memory barrier in between). If we do it the other way around,
	// another thread can submit a goroutine after we've checked all run queues
	// but before we drop nmspinning; as the result nobody will unpark a thread
	// to run the goroutine.
	// If we discover new work below, we need to restore m.spinning as a signal
	// for resetspinning to unpark a new worker thread (because there can be more
	// than one starving goroutine). However, if after discovering new work
	// we also observe no idle Ps, it is OK to just park the current thread:
	// the system is fully loaded so no spinning threads are required.
	// Also see "Worker thread parking/unparking" comment at the top of the file.
	wasSpinning := _g_.m.spinning
	if _g_.m.spinning {
		//m即将睡眠，状态不再是spinning
		_g_.m.spinning = false
		if int32(atomic.Xadd(&sched.nmspinning, -1)) < 0 {
			throw("findrunnable: negative nmspinning")
		}
	}

	// check all runqueues once again
	//特么还不死心，休眠之前再看一下是否有工作要做
	for _, _p_ := range allpSnapshot {
		if !runqempty(_p_) {
			lock(&sched.lock)
			_p_ = pidleget()
			unlock(&sched.lock)
			if _p_ != nil {
				acquirep(_p_)
				if wasSpinning {
					_g_.m.spinning = true
					atomic.Xadd(&sched.nmspinning, 1)
				}
				goto top
			}
			break
		}
	}

	// Check for idle-priority GC work again.
	if gcBlackenEnabled != 0 && gcMarkWorkAvailable(nil) {
		lock(&sched.lock)
		_p_ = pidleget()
		if _p_ != nil && _p_.gcBgMarkWorker == 0 {
			pidleput(_p_)
			_p_ = nil
		}
		unlock(&sched.lock)
		if _p_ != nil {
			acquirep(_p_)
			if wasSpinning {
				_g_.m.spinning = true
				atomic.Xadd(&sched.nmspinning, 1)
			}
			// Go back to idle GC check.
			goto stop
		}
	}

	// poll network
	if netpollinited() && (atomic.Load(&netpollWaiters) > 0 || pollUntil != 0) && atomic.Xchg64(&sched.lastpoll, 0) != 0 {
		atomic.Store64(&sched.pollUntil, uint64(pollUntil))
		if _g_.m.p != 0 {
			throw("findrunnable: netpoll with p")
		}
		if _g_.m.spinning {
			throw("findrunnable: netpoll with spinning")
		}
		if faketime != 0 {
			// When using fake time, just poll.
			delta = 0
		}
		list := netpoll(delta) // block until new work is available
		atomic.Store64(&sched.pollUntil, 0)
		atomic.Store64(&sched.lastpoll, uint64(nanotime()))
		if faketime != 0 && list.empty() {
			// Using fake time and nothing is ready; stop M.
			// When all M's stop, checkdead will call timejump.
			stopm()
			goto top
		}
		lock(&sched.lock)
		_p_ = pidleget()
		unlock(&sched.lock)
		if _p_ == nil {
			injectglist(&list)
		} else {
			acquirep(_p_)
			if !list.empty() {
				gp := list.pop()
				injectglist(&list)
				casgstatus(gp, _Gwaiting, _Grunnable)
				if trace.enabled {
					traceGoUnpark(gp, 0)
				}
				return gp, false
			}
			if wasSpinning {
				_g_.m.spinning = true
				atomic.Xadd(&sched.nmspinning, 1)
			}
			goto top
		}
	} else if pollUntil != 0 && netpollinited() {
		pollerPollUntil := int64(atomic.Load64(&sched.pollUntil))
		if pollerPollUntil == 0 || pollerPollUntil > pollUntil {
			netpollBreak()
		}
	}
	//算了受不鸟了，没事做，睡觉了，等待唤醒再跳转到top看看有没有事做
	stopm()
	goto top
}

第一点，工作线程M的自旋状态(spinning)。工作线程在从其它工作线程的本地运行队列中盗取goroutine时的状态称为自旋状态。从上面代码可以看到，当前M在去其它p的运行队列盗取goroutine之前把spinning标志设置成了true，同时增加处于自旋状态的M的数量，而盗取结束之后则把spinning标志还原为false，同时减少处于自旋状态的M的数量，从后面的分析我们可以看到，当有空闲P又有goroutine需要运行的时候，这个处于自旋状态的M的数量决定了是否需要唤醒或者创建新的工作线程。

第二点，盗取算法。盗取过程用了两个嵌套for循环。内层循环实现了盗取逻辑，从代码可以看出盗取的实质就是遍历allp中的所有p，查看其运行队列是否有goroutine，如果有，则取其一半到当前工作线程的运行队列，然后从findrunnable返回，如果没有则继续遍历下一个p。但这里为了保证公平性，遍历allp时并不是固定的从allp[0]即第一个p开始，而是从随机位置上的p开始，而且遍历的顺序也随机化了，并不是现在访问了第i个p下一次就访问第i+1个p，而是使用了一种伪随机的方式遍历allp中的每个p，防止每次遍历时使用同样的顺序访问allp中的元素。

```

## runqsteal()

```go
// Steal half of elements from local runnable queue of p2
// and put onto local runnable queue of p.
// Returns one of the stolen elements (or nil if failed).
func runqsteal(_p_, p2 *p, stealRunNextG bool) *g {
	t := _p_.runqtail
	n := runqgrab(p2, &_p_.runq, t, stealRunNextG)
	if n == 0 {
		return nil
	}
	n--
	gp := _p_.runq[(t+n)%uint32(len(_p_.runq))].ptr()
	if n == 0 {
		return gp
	}
	h := atomic.LoadAcq(&_p_.runqhead) // load-acquire, synchronize with consumers
	if t-h+n >= uint32(len(_p_.runq)) {
		throw("runqsteal: runq overflow")
	}
	atomic.StoreRel(&_p_.runqtail, t+n) // store-release, makes the item available for consumption
	return gp
}

// Grabs a batch of goroutines from _p_'s runnable queue into batch.
// Batch is a ring buffer starting at batchHead.
// Returns number of grabbed goroutines.
// Can be executed by any P.
func runqgrab(_p_ *p, batch *[256]guintptr, batchHead uint32, stealRunNextG bool) uint32 {
	for {
		h := atomic.LoadAcq(&_p_.runqhead) // load-acquire, synchronize with other consumers
		t := atomic.LoadAcq(&_p_.runqtail) // load-acquire, synchronize with the producer
		n := t - h//计算队列中有多少个goroutine
		n = n - n/2//取队列中goroutine个数的一半
		if n == 0 {
			if stealRunNextG {
				// Try to steal from _p_.runnext.
				if next := _p_.runnext; next != 0 {
					if _p_.status == _Prunning {
						// Sleep to ensure that _p_ isn't about to run the g
						// we are about to steal.
						// The important use case here is when the g running
						// on _p_ ready()s another g and then almost
						// immediately blocks. Instead of stealing runnext
						// in this window, back off to give _p_ a chance to
						// schedule runnext. This will avoid thrashing gs
						// between different Ps.
						// A sync chan send/recv takes ~50ns as of time of
						// writing, so 3us gives ~50x overshoot.
						if GOOS != "windows" {
							usleep(3)
						} else {
							// On windows system timer granularity is
							// 1-15ms, which is way too much for this
							// optimization. So just yield.
							osyield()
						}
					}
					if !_p_.runnext.cas(next, 0) {
						continue
					}
					batch[batchHead%uint32(len(batch))] = next
					return 1
				}
			}
			return 0
		}
		//小细节：按理说队列中的goroutine个数最多就是len(_p_.runq)，
		//所以n的最大值也就是len(_p_.runq)/2，那为什么需要这个判断呢？
		//这里关键点在于读取runqhead和runqtail是两个操作而非一个原子操作，当我们读取runqhead之后但还未读取runqtail之前，如果有其它线程快速的在增加（这是完全有可能的，其它偷取者从队列中偷取goroutine会增加runqhead，而队列的所有者往队列中添加goroutine会增加runqtail）这两个值，则会导致我们读取出来的runqtail已经远远大于我们之前读取出来放在局部变量h里面的runqhead了，也就是代码注释中所说的h和t已经不一致了，所以这里需要这个if判断来检测异常情况。
		
		if n > uint32(len(_p_.runq)/2) { // read inconsistent h and t
			continue
		}
		for i := uint32(0); i < n; i++ {
			g := _p_.runq[(h+i)%uint32(len(_p_.runq))]
			batch[(batchHead+i)%uint32(len(batch))] = g
		}
		if atomic.CasRel(&_p_.runqhead, h, h+n) { // cas-release, commits consume
			return n
		}
	}
}
```

## stopm()

```go
// Stops execution of the current m until new work is available.
// Returns with acquired P.
func stopm() {
	_g_ := getg()

	if _g_.m.locks != 0 {
		throw("stopm holding locks")
	}
	if _g_.m.p != 0 {
		throw("stopm holding p")
	}
	if _g_.m.spinning {
		throw("stopm spinning")
	}

	lock(&sched.lock)
	//把m结构体对象放入sched.midle空闲队列
	mput(_g_.m)
	unlock(&sched.lock)
	//进入睡眠状态
	notesleep(&_g_.m.park)
	
	
	//被其它工作线程唤醒
	noteclear(&_g_.m.park)
	acquirep(_g_.m.nextp.ptr())
	_g_.m.nextp = 0
}

stopm的核心是调用mput把m结构体对象放入sched的midle空闲队列，然后通过notesleep(&m.park)函数让自己进入睡眠状态。

note是go runtime实现的一次性睡眠和唤醒机制，一个线程可以通过调用notesleep(*note)进入睡眠状态，而另外一个线程则可以通过notewakeup(*note)把其唤醒。note的底层实现机制跟操作系统相关，不同系统使用不同的机制，比如linux下使用的futex系统调用，而mac下则是使用的pthread_cond_t条件变量，note对这些底层机制做了一个抽象和封装，这种封装给扩展性带来了很大的好处，比如当睡眠和唤醒功能需要支持新平台时，只需要在note层增加对特定平台的支持即可，不需要修改上层的任何代码。

回到stopm，当从notesleep函数返回后，需要再次绑定一个p，然后返回到findrunnable函数继续重新寻找可运行的goroutine，一旦找到可运行的goroutine就会返回到schedule函数，并把找到的goroutine调度起来运行，如何把goroutine调度起来运行的代码我们已经分析过了。现在继续看notesleep函数。

```

## notesleep()

    不同的系统平台，实现不一样分布有：
    1、runtime/lock_sema.go - notesleep() - aix、darwin、netbsd、openbsd、plan9、solaris、windows
    2、runtime/lock_futex.go - notesleep() - dragonfly、freebsd、linux架构

    涉及到知识点偏系统底层，可以选择去了解：
    lock_sema.go：

    1、futexsleep(addr uint32, val uint32, ns int64)：原子操作`if addr == val { sleep  }`。
    2、futexwakeup(addr *uint32, cnt uint32)：唤醒地址addr上的线程最多cnt次。

    lock_futex.go->runtime/os_linux.go->futexsleep()->runtime/sys_linux_amd64.s->runtime·futex(SB)

    1、func semacreate(mp *m):创建信号量
    2、func semasleep(ns int64) int32： 请求信号量，请求不到会休眠一段时间
    3、func semawakeup(mp *m)：唤醒mp

    这里多说一句runtime.Mutex也是用了上面这些知识实现
