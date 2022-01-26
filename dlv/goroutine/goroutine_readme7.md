## 在前一节中，都是从队列里面找需要调度的g，这次我们了解下抢占调度

在1.14中有如下回触发调度：

```go
1、go关键字
2、gc
3、系统调用
4、内存同步操作：atomic，mutex，channel 操作等会使 goroutine 阻塞，因此会被调度走。等条件满足后（例如其他 goroutine 解锁了）还会被调度上来继续运行
5、挂起/唤醒g
```

    在之前的章节中我们了解了main g和普通g的产生，并放入队列中等待调度运行。但是我们在实际应用中会有停止g调度的情况比如：gc的stw、gc的栈扫描、sysmon的后台监控（公平调度）、channel的写入读取、系统调用、网络io、定时器等。
    针对这些终止g的运行我们有一个概念叫做抢占，其中抢占有如下实现：
    1、协作式抢占（讲道理版）
        函数头、函数尾插入的栈扩容检查，在函数调用中插入会变指令
        0x0000 00000 (add.go:8)	TEXT	"".xxx(SB), ABIInternal, $24-32
        0x0000 00000 (add.go:8)	MOVQ	(TLS), CX
        0x0009 00009 (add.go:8)	CMPQ	SP, 16(CX)//检查是否sp < g.stackguard0，若是条到下面去操作morestack_noctxt(),切换到g0扩展g栈
        ...
        0x0060 00096 (add.go:8)	CALL	runtime.morestack_noctxt(SB)
        //morestack_noctxt在1.14中已经加入信号抢占了
    2、非协作式抢占（不讲道理版）
        在mstart1()->mstart0()->initsig()中，最终实现32种系统信号的回调函数处理注册，只不过不同系统绑定信号和处理函数的函数有点差异，但都是调用系统底层能力。
        最终体现的效果就是在出现对应的信号后，系统会切换到要处理该信号的函数上面去，执行完之后再切换回来。相当于在用户代码中加入了一段逻辑一样。
        mac代码中发送信号：
        func signalM(mp *m, sig int) {
            pthread_kill(pthread(mp.procid), uint32(sig))
        }
        信号的处理：
        //runtime/signal_unix.go
        func sighandler(sig uint32, info *siginfo, ctxt unsafe.Pointer, gp *g) {
            ...
            //会判断是否进行抢占，切换到对应的信号执行函数栈中去执行，再切换回来
            doSigPreempt(gp, c)
            ...
        }

        // asyncPreempt is implemented in assembly.
        func asyncPreempt()//汇编实现，切换运行线程，执行信号处理函数，再恢复线程
        
        //go:nosplit
        func asyncPreempt2() {
            gp := getg()
            gp.asyncSafePoint = true
            if gp.preemptStop {
                mcall(preemptPark)//preemptStop 是在 GC 的栈扫描中才会设置为 true
            } else {
                //除了栈扫描，其它抢占全部走这条分支
                mcall(gopreempt_m)
            }
            gp.asyncSafePoint = false
        }

        栈扫描抢占流程
        
        suspendG -> preemptM -> signalM 发信号。
        
        sighandler -> asyncPreempt -> 保存执行现场 -> asyncPreempt2 -> preemptPark
        
        preemptPark 和 gopark 类似，挂起当前正在执行的 goroutine，该 goroutine 之前绑定的线程就可以继续执行调度循环了。
        
        scanstack 执行完之后：
        
        resumeG -> ready -> runqput 会让之前被停下来的 goroutine 进当前 P 的队列或全局队列。
        
        其它流程
        
        preemptone -> preemptM - signalM 发信号。
        
        sighandler -> asyncPreempt -> 保存执行现场 -> asyncPreempt2 -> gopreempt_m
        
        gopreempt_m 会直接将被抢占的 goroutine 放进全局队列。
        
        无论是栈扫描流程还是其它流程，当 goroutine 程序被调度到时，都是从汇编中的 CALL ·asyncPreempt2(SB) 的下一条指令开始执行的，即 asyncPreempt 汇编函数的下半部分。
        
        这部分会将之前 goroutine 的现场完全恢复，就和抢占从来没有发生过一样。

    在channel原理中涉及到g的gopark()挂起、goready()唤醒。

##  在gc的stw中：

```go
//runtime/proc.go
func stopTheWorldWithSema() {
	.....
	preemptall()
	.....
	// 等待剩余的 P 主动停下
	if wait {
		for {
			// wait for 100us, then try to re-preempt in case of any races
			// 等待 100us，然后重新尝试抢占
			if notetsleep(&sched.stopnote, 100*1000) {
				noteclear(&sched.stopnote)
				break
			}
			//并不是抢占，而是设置抢占标识，在1.14前会有一个由于用户g无法停止，进而无法gc操作，一直卡在这个循环里面的问题，后面为了这个问题做了优化，就是后面的所谓"信号"抢占
			preemptall()
		}
	}

```

##  在g的gc栈扫描中：

```go
//runtime/mgcmark.go
func markroot(gcw *gcWork, i uint32) {
	// Note: if you add a case here, please also update heapdump.go:dumproots.
	switch {
	......
	default:
		// the rest is scanning goroutine stacks
		var gp *g
		......

		// scanstack must be done on the system stack in case
		// we're trying to scan our own stack.
		systemstack(func() {
			//suspendG中会调用 preemptM -> signalM 对正在执行的 goroutine 所在的线程发送抢占信号
			stopped := suspendG(gp)
			scanstack(gp, gcw)
			resumeG(stopped)
		})
	}
}

```

##   在sysmon()中:

```go
func sysmon() {
	idle := 0 // how many cycles in succession we had not wokeup somebody
	for {
		......
		// retake P's blocked in syscalls
		// and preempt long running G's
		//syscall 太久的，需要将 P 从 M 上剥离；运行用户代码太久的，需要抢占停止该 goroutine 执行。
		//这里
		if retake(now) != 0 {
			idle = 0
		} else {
			idle++
		}
	}
}

在runtime/proc.go中定义了如果一个g运行了多久会被终止，让别的g有运行的机会
// forcePreemptNS is the time slice given to a G before it is
// preempted.
const forcePreemptNS = 10 * 1000 * 1000 // 10ms
这里定义了一个g最长能运行10ms

func retake(now int64) uint32 {
    n := 0
    // Prevent allp slice changes. This lock will be completely
    // uncontended unless we're already stopping the world.
    lock(&allpLock)
    // We can't use a range loop over allp because we may
    // temporarily drop the allpLock. Hence, we need to re-fetch
    // allp each time around the loop.
    for i := 0; i < len(allp); i++ {
        _p_ := allp[i]
        if _p_ == nil {
            // This can happen if procresize has grown
            // allp but not yet created new Ps.
            continue
        }
        pd := &_p_.sysmontick
        s := _p_.status
        sysretake := false
        //p的状态标明这个p正在做啥，最后preemptone中调用preemptM发生抢占信息
        if s == _Prunning || s == _Psyscall {
            // Preempt G if it's running for too long.
            t := int64(_p_.schedtick)
            if int64(pd.schedtick) != t {
                pd.schedtick = uint32(t)
                pd.schedwhen = now
            } else if pd.schedwhen+forcePreemptNS <= now {
                preemptone(_p_)
                // In case of syscall, preemptone() doesn't
                // work, because there is no M wired to P.
                sysretake = true
            }
        }
        //如果正在发生系统调用
        if s == _Psyscall {
            // Retake P from syscall if it's there for more than 1 sysmon tick (at least 20us).
            t := int64(_p_.syscalltick)
            if !sysretake && int64(pd.syscalltick) != t {
                pd.syscalltick = uint32(t)
                pd.syscallwhen = now
                continue
            }
            // On the one hand we don't want to retake Ps if there is no other work to do,
            // but on the other hand we want to retake them eventually
            // because they can prevent the sysmon thread from deep sleep.
            if runqempty(_p_) && atomic.Load(&sched.nmspinning)+atomic.Load(&sched.npidle) > 0 && pd.syscallwhen+10*1000*1000 > now {
                continue
            }
            // Drop allpLock so we can take sched.lock.
            unlock(&allpLock)
            // Need to decrement number of idle locked M's
            // (pretending that one more is running) before the CAS.
            // Otherwise the M from which we retake can exit the syscall,
            // increment nmidle and report deadlock.
            incidlelocked(-1)
            if atomic.Cas(&_p_.status, s, _Pidle) {
                if trace.enabled {
                    traceGoSysBlock(_p_)
                    traceProcStop(_p_)
                }
                n++
                _p_.syscalltick++
                handoffp(_p_)
            }
            incidlelocked(1)
            lock(&allpLock)
        }
    }
    unlock(&allpLock)
    return uint32(n)
}
```
##   系统调用在部分调用系统底层提供的函数时，有阻塞的情况，那么部分阻塞被golang调度器接管，挂起等待数据准备后之后在唤醒。


    我本身系统时苹果系统

    //runtime/sys_darwin.go下找到下面函数

    1、syscall_syscall

        func syscall_syscall(fn, a1, a2, a3 uintptr) (r1, r2, err uintptr) {
            entersyscall()
            libcCall(unsafe.Pointer(funcPC(syscall)), unsafe.Pointer(&fn))
            exitsyscall()
            return
        }
        func syscall()

    2、syscall_rawSyscall 不需要entersyscall、exitsyscall

    func syscall_rawSyscall(fn, a1, a2, a3 uintptr) (r1, r2, err uintptr) {
        libcCall(unsafe.Pointer(funcPC(syscall)), unsafe.Pointer(&fn))
        return
    }

    部分系统调用定义：

    系统调用	类型
    SYS_TIME	RawSyscall
    SYS_GETTIMEOFDAY	RawSyscall
    SYS_SETRLIMIT	RawSyscall
    SYS_GETRLIMIT	RawSyscall
    SYS_EPOLL_WAIT	Syscall

    entersyscall：准备工作

        会在获取当前程序计数器和栈位置之后调用 runtime.reentersyscall，它会完成 Goroutine 进入系统调用前的准备工作
    禁止线程上发生的抢占，防止出现内存不一致的问题；
    保证当前函数不会触发栈分裂或者增长；
    保存当前的程序计数器 PC 和栈指针 SP 中的内容；
    将 Goroutine 的状态更新至 _Gsyscall；
    将 Goroutine 的处理器和线程暂时分离并更新处理器的状态到 _Psyscall；
    释放当前线程上的锁；

    exitsyscall：调用结束恢复调度

    当系统调用结束后，会调用退出系统调用的函数 runtime.exitsyscall 为当前 Goroutine 重新分配资源，该函数有两个不同的执行路径：
    调用 runtime.exitsyscallfast；
    切换至调度器的 Goroutine 并调用 runtime.exitsyscall0

    这两种不同的路径会分别通过不同的方法查找一个用于执行当前 Goroutine 处理器 P，快速路径 runtime.exitsyscallfast 中包含两个不同的分支：
    
    1、如果 Goroutine 的原处理器处于 _Psyscall 状态，会直接调用 wirep 将 Goroutine 与处理器进行关联；
    2、如果调度器中存在闲置的处理器，会调用 runtime.acquirep 使用闲置的处理器处理当前 Goroutine；
    另一个相对较慢的路径 runtime.exitsyscall0 会将当前 Goroutine 切换至 _Grunnable 状态，并移除线程 M 和当前 Goroutine 的关联：
    
    当我们通过 runtime.pidleget 获取到闲置的处理器时就会在该处理器上执行 Goroutine；
    在其它情况下，我们会将当前 Goroutine 放到全局的运行队列中，等待调度器的调度；

## 主动让出cpu

```go
// Gosched yields the processor, allowing other goroutines to run. It does not
// suspend the current goroutine, so execution resumes automatically.
func Gosched() {
	checkTimeouts()
	//最终是调用了goschedImpl
	mcall(gosched_m)
}
// Gosched continuation on g0.
func gosched_m(gp *g) {
    if trace.enabled {
    traceGoSched()
    }
    goschedImpl(gp)
}

func goschedImpl(gp *g) {
	//获取当前g的状态
    status := readgstatus(gp)
    if status&^_Gscan != _Grunning {
        dumpgstatus(gp)
        throw("bad g status")
    }
    //把g改为可_Grunnable状态
    casgstatus(gp, _Grunning, _Grunnable)
    //切断g与m的关系
	// dropg 移除 m 与当前 Goroutine m->curg（简称 gp ）之间的关联。
	// 通常，调用方将 gp 的状态设置为非 _Grunning 后立即调用 dropg 完成工作。
	// 调用方也有责任在 gp 将使用 ready 时重新启动时进行相关安排。
	// 在调用 dropg 并安排 gp ready 好后，调用者可以做其他工作，但最终应该
	// 调用 schedule 来重新启动此 m 上的 Goroutine 的调度。
    dropg()
    lock(&sched.lock)
    //把g放入全局队列等待调度
    globrunqput(gp)
    unlock(&sched.lock)
    //重新调度
    schedule()
}
func dropg() {
    _g_ := getg()
    
    setMNoWB(&_g_.m.curg.m, nil)
    setGNoWB(&_g_.m.curg, nil)
}
// setMNoWB 当使用 muintptr 不可行时，在没有 write barrier 下执行 *mp = new
//go:nosplit
//go:nowritebarrier
func setMNoWB(mp **m, new *m) {
    (*muintptr)(unsafe.Pointer(mp)).set(new)
}
// setGNoWB 当使用 guintptr 不可行时，在没有 write barrier 下执行 *gp = new
//go:nosplit
//go:nowritebarrier
func setGNoWB(gp **g, new *g) {
    (*guintptr)(unsafe.Pointer(gp)).set(new)
}
```

## 挂起/唤醒g

系统调用；
channel读写条件不满足；
抢占式调度时间片结束；

gopark函数做的主要事情分为两点：

解除当前goroutine的m的绑定关系，将当前goroutine状态机切换为等待状态；
调用一次schedule()函数，在局部调度器P发起一轮新的调度。

```go
func gopark(unlockf func(*g, unsafe.Pointer) bool, lock unsafe.Pointer, reason waitReason, traceEv byte, traceskip int) {
	//挂起的原因：runtime/runtime2.go-948行，有各种挂起原因
	if reason != waitReasonSleep {
		checkTimeouts() // timeouts may expire while two goroutines keep the scheduler busy
	}
	mp := acquirem()
	gp := mp.curg
	status := readgstatus(gp)
	if status != _Grunning && status != _Gscanrunning {
		throw("gopark: bad g status")
	}
	mp.waitlock = lock
	mp.waitunlockf = unlockf
	gp.waitreason = reason
	mp.waittraceev = traceEv
	mp.waittraceskip = traceskip
	releasem(mp)
	// can't do anything that might move the G between Ms here.
	mcall(park_m)
}
// park continuation on g0.
func park_m(gp *g) {
    _g_ := getg()
    
    if trace.enabled {
    traceGoPark(_g_.m.waittraceev, _g_.m.waittraceskip)
    }
    
    //g改为_Gwaiting状态
    casgstatus(gp, _Grunning, _Gwaiting)
    //解绑g与m关系
    dropg()
    
    //不是很明白，好像调试相关
    if fn := _g_.m.waitunlockf; fn != nil {
        ok := fn(gp, _g_.m.waitlock)
        _g_.m.waitunlockf = nil
        _g_.m.waitlock = nil
        if !ok {
            if trace.enabled {
                traceGoUnpark(gp, 2)
            }
            casgstatus(gp, _Gwaiting, _Grunnable)
            execute(gp, true) // Schedule it back, never returns.
        }
    }
    //重新调度
    schedule()
}
```

```go
//唤醒g
func goready(gp *g, traceskip int) {
	systemstack(func() {
		ready(gp, traceskip, true)
	})
}

// Mark gp ready to run.
func ready(gp *g, traceskip int, next bool) {
    if trace.enabled {
        traceGoUnpark(gp, traceskip)
    }
    
    status := readgstatus(gp)
    
    // Mark runnable.
    _g_ := getg()
    //不准抢占
    mp := acquirem() // disable preemption because it can be holding p in a local var
    if status&^_Gscan != _Gwaiting {
        dumpgstatus(gp)
        throw("bad g->status in ready")
    }
    //改为_Grunnable状态
    // status is Gwaiting or Gscanwaiting, make Grunnable and put on runq
    casgstatus(gp, _Gwaiting, _Grunnable)
    //放入队列等待调度
    runqput(_g_.m.p.ptr(), gp, next)
    //有空闲的p而且没有正在偷取goroutine的工作线程，则需要唤醒p出来工作
    if atomic.Load(&sched.npidle) != 0 && atomic.Load(&sched.nmspinning) == 0 {
        wakep()
    }
    releasem(mp)
}

func wakep() {
    // be conservative about spinning threads
	//是否存在正在到处找g的sched，如果有就是说本身没什么可运行的g，那我就不去再唤醒其他线程了，
	//因为大家都很闲，并且有线程到处找g了。没必要再添乱了
    if !atomic.Cas(&sched.nmspinning, 0, 1) {
        return
    }
    startm(nil, true)
}

// Schedules some M to run the p (creates an M if necessary).
// If p==nil, tries to get an idle P, if no idle P's does nothing.
// May run with m.p==nil, so write barriers are not allowed.
// If spinning is set, the caller has incremented nmspinning and startm will
// either decrement nmspinning or set m.spinning in the newly started M.
//go:nowritebarrierrec
func startm(_p_ *p, spinning bool) {
    lock(&sched.lock)
    if _p_ == nil {//没有指定p的话需要从p的空闲队列中获取一个p
        _p_ = pidleget()//从p的空闲队列中获取空闲p
        if _p_ == nil {//还是空
            unlock(&sched.lock)
            if spinning {
                // The caller incremented nmspinning, but there are no idle Ps,
                // so it's okay to just undo the increment and give up.
            	//spinning为true表示进入这个函数之前已经对sched.nmspinning加了1，需要还原
                if int32(atomic.Xadd(&sched.nmspinning, -1)) < 0 {
                    throw("startm: negative nmspinning")
                }
            }
            return //没有空闲的p，直接返回
        }
    }
    //从m空闲队列中获取正处于睡眠之中的工作线程，所有处于睡眠状态的m都在此队列中
    mp := mget()
    //没有处于睡眠状态的工作线程
    if mp == nil {
        // No M is available, we must drop sched.lock and call newm.
        // However, we already own a P to assign to the M.
        //
        // Once sched.lock is released, another G (e.g., in a syscall),
        // could find no idle P while checkdead finds a runnable G but
        // no running M's because this new M hasn't started yet, thus
        // throwing in an apparent deadlock.
        //
        // Avoid this situation by pre-allocating the ID for the new M,
        // thus marking it as 'running' before we drop sched.lock. This
        // new M will eventually run the scheduler to execute any
        // queued G's.
        id := mReserveID()
        unlock(&sched.lock)
        
        var fn func()
        if spinning {
            // The caller incremented nmspinning, so set m.spinning in the new M.
            fn = mspinning
        }
        newm(fn, _p_, id)//创建新的工作线程
        return
    }
    unlock(&sched.lock)
    if mp.spinning {
        throw("startm: m is spinning")
    }
    if mp.nextp != 0 {
        throw("startm: m has p")
    }
    if spinning && !runqempty(_p_) {
        throw("startm: p has runnable gs")
    }
    // The caller incremented nmspinning, so set m.spinning in the new M.
    mp.spinning = spinning
    mp.nextp.set(_p_)
    //唤醒处于休眠状态的工作线程
    notewakeup(&mp.park)
}

func notewakeup(n *note) {
    var v uintptr
    for {
        v = atomic.Loaduintptr(&n.key)
        //设置n.key = locked = 1, 被唤醒的线程通过查看该值是否等于1来确定是被其它线程唤醒还是意外从睡眠中苏醒
        if atomic.Casuintptr(&n.key, v, locked) {
            break
        }
    }
    
    // Successfully set waitm to locked.
    // What was it before?
    switch {
        case v == 0:
        // Nothing was waiting. Done.
        case v == locked:
        // Two notewakeups! Not allowed.
        throw("notewakeup - double wakeup")
        default:
        // Must be the waiting m. Wake it up.
		//调用semawakeup（linux是futexwakeup）唤醒
        semawakeup((*m)(unsafe.Pointer(v)))
    }
}

func semawakeup(mp *m) {
    pthread_mutex_lock(&mp.mutex)
    mp.count++
    if mp.count > 0 {
        pthread_cond_signal(&mp.cond)
    }
    pthread_mutex_unlock(&mp.mutex)
}
```
##  newm(fn, _p_, id)创建新的工作线程

```go
// Create a new m. It will start off with a call to fn, or else the scheduler.
// fn needs to be static and not a heap allocated closure.
// May run with m.p==nil, so write barriers are not allowed.
//
// id is optional pre-allocated m ID. Omit by passing -1.
//go:nowritebarrierrec
func newm(fn func(), _p_ *p, id int64) {
	//allocm函数从堆上分配一个m结构体对象，然后调用newm1函数
    mp := allocm(_p_, fn, id)
    mp.nextp.set(_p_)
    mp.sigmask = initSigmask
    if gp := getg(); gp != nil && gp.m != nil && (gp.m.lockedExt != 0 || gp.m.incgo) && GOOS != "plan9" {
        // We're on a locked M or a thread that may have been
        // started by C. The kernel state of this thread may
        // be strange (the user may have locked it for that
        // purpose). We don't want to clone that into another
        // thread. Instead, ask a known-good thread to create
        // the thread for us.
        //
        // This is disabled on Plan 9. See golang.org/issue/22227.
        //
        // TODO: This may be unnecessary on Windows, which
        // doesn't model thread creation off fork.
        lock(&newmHandoff.lock)
        if newmHandoff.haveTemplateThread == 0 {
            throw("on a locked thread with no template thread")
        }
        mp.schedlink = newmHandoff.newm
        newmHandoff.newm.set(mp)
        if newmHandoff.waiting {
            newmHandoff.waiting = false
            notewakeup(&newmHandoff.wake)
        }
        unlock(&newmHandoff.lock)
        return
    }
    newm1(mp)
}

func newm1(mp *m) {
    if iscgo {
        var ts cgothreadstart
        if _cgo_thread_start == nil {
            throw("_cgo_thread_start missing")
        }
        ts.g.set(mp.g0)
        ts.tls = (*uint64)(unsafe.Pointer(&mp.tls[0]))
        ts.fn = unsafe.Pointer(funcPC(mstart))
        if msanenabled {
            msanwrite(unsafe.Pointer(&ts), unsafe.Sizeof(ts))
        }
        execLock.rlock() // Prevent process clone.
        asmcgocall(_cgo_thread_start, unsafe.Pointer(&ts))
        execLock.runlock()
        return
    }
    execLock.rlock() // Prevent process clone.
    //newosproc的主要任务是调用clone函数创建一个系统线程，而新建的这个系统线程将从mstart函数开始运行
    newosproc(mp)
    execLock.runlock()
}

//newosproc方法在不同系统有不同的实现如runtime/os_darwin.go，runtime/os_linux.go

// May run with m.p==nil, so write barriers are not allowed.
//go:nowritebarrierrec
func newosproc(mp *m) {
    stk := unsafe.Pointer(mp.g0.stack.hi)
    if false {
        print("newosproc stk=", stk, " m=", mp, " g=", mp.g0, " id=", mp.id, " ostk=", &mp, "\n")
    }
    
    // Initialize an attribute object.
    var attr pthreadattr
    var err int32
    //初始化线程
    err = pthread_attr_init(&attr)
    if err != 0 {
        write(2, unsafe.Pointer(&failthreadcreate[0]), int32(len(failthreadcreate)))
        exit(1)
    }
    
    // Find out OS stack size for our own stack guard.
    //初始化线程栈信息
    var stacksize uintptr
    if pthread_attr_getstacksize(&attr, &stacksize) != 0 {
        write(2, unsafe.Pointer(&failthreadcreate[0]), int32(len(failthreadcreate)))
        exit(1)
    }
    mp.g0.stack.hi = stacksize // for mstart
    //mSysStatInc(&memstats.stacks_sys, stacksize) //TODO: do this?
    
    // Tell the pthread library we won't join with this thread.
    if pthread_attr_setdetachstate(&attr, _PTHREAD_CREATE_DETACHED) != 0 {
        write(2, unsafe.Pointer(&failthreadcreate[0]), int32(len(failthreadcreate)))
        exit(1)
    }
    
    // Finally, create the thread. It starts at mstart_stub, which does some low-level
    // setup and then calls mstart.
    //创建线程，并且新线程从mstart()函数开始运行
    var oset sigset
    sigprocmask(_SIG_SETMASK, &sigset_all, &oset)
    //涉及到底层调用了，想了解的可以去查下pthread_create的创建，下面只是简单提下
    err = pthread_create(&attr, funcPC(mstart_stub), unsafe.Pointer(mp))
    sigprocmask(_SIG_SETMASK, &oset, nil)
    if err != 0 {
        write(2, unsafe.Pointer(&failthreadcreate[0]), int32(len(failthreadcreate)))
        exit(1)
    }
}

对于mac系统，如果成功创建线程，pthread_create() 函数返回数字 0，那么新线程进行新的调度循环，本身线程得到返回值0
反之返回非零值。各个非零值都对应着不同的宏，指明创建失败的原因，常见的宏有以下几种：
EAGAIN：系统资源不足，无法提供创建线程所需的资源。
EINVAL：传递给 pthread_create() 函数的 attr 参数无效。
EPERM：传递给 pthread_create() 函数的 attr 参数中，某些属性的设置为非法操作，程序没有相关的设置权限。

对于Linux用clone系统调用完成后实际上就多了一个操作系统线程，新创建的子线程和当前线程都得从系统调用返回然后继续执行后面的代码，那么从系统调用返回之后我们怎么知道哪个是父线程哪个是子线程，从而来决定它们的执行流程？使用过fork系统调用的读者应该知道，我们需要通过返回值来判断父子线程，系统调用的返回值如果是0则表示这是子线程，不为0则表示这个是父线程。用c代码来描述大概就是这个样子：
    
if (pthread_create(...) == 0) { //子线程
    子线程代码
} else {//父线程
    父线程代码
}

然这里只有一次clone调用，但它却返回了2次，一次返回到父线程，一次返回到子线程，然后2个线程各自执行自己的代码流程。

回到clone函数，下面代码的第一条指令就在判断系统调用的返回值，如果是子线程则跳转到后面的代码继续执行，如果是父线程，它创建子线程的任务已经完成，所以这里把返回值保存在栈上之后就直接执行ret指令返回到newosproc函数了。

//mstart_stub
//runtime/sys_darwin_amd64.s
TEXT runtime·mstart_stub(SB),NOSPLIT,$0
    // DI points to the m.
    // We are already on m's g0 stack.
    
    // Save callee-save registers.
    //保存当前调用这的寄存器信息
    SUBQ	$40, SP
    MOVQ	BX, 0(SP)
    MOVQ	R12, 8(SP)
    MOVQ	R13, 16(SP)
    MOVQ	R14, 24(SP)
    MOVQ	R15, 32(SP)
    
    MOVQ	m_g0(DI), DX // g
    
    // Initialize TLS entry.
    // See cmd/link/internal/ld/sym.go:computeTLSOffset.
    MOVQ	DX, 0x30(GS)
    
    // Someday the convention will be D is always cleared.
    CLD
    
    //调用mstart
    CALL	runtime·mstart(SB)
    //回忆一下，mstart函数首先会去设置m.g0的stackguard成员，然后调用mstart1()函数把当前工作线程的g0的调度信息保存在m.g0.sched成员之中，最后通过调用schedule函数进入调度循环。
    
    // Restore callee-save registers.
    //返回后再从栈中恢复寄存器信息，达到切换现场
    MOVQ	0(SP), BX
    MOVQ	8(SP), R12
    MOVQ	16(SP), R13
    MOVQ	24(SP), R14
    MOVQ	32(SP), R15
    
    // Go is all done with this OS thread.
    // Tell pthread everything is ok (we never join with this thread, so
    // the value here doesn't really matter).
    XORL	AX, AX
    
    ADDQ	$40, SP
    RET

```

