
##创建main goroutine

```go
    // runtime/asm_amd64.s
    // TEXT runtime·rt0_go(SB),NOSPLIT,$0
	// create a new goroutine to start program
	//mainPC是runtime.main
	MOVQ	$runtime·mainPC(SB), AX		// entry
	PUSHQ	AX  // AX = &funcval{runtime·main}
	//newproc的第一个参数入栈，该参数表示runtime.main函数需要的参数大小，因为runtime.main没有参数，所以这里是0
	PUSHQ	$0			// arg size
	//创建main goroutine
	//runtime.main最终会调用我们写的main.main函数，重点放在newproc这个函数
	//newproc函数用于创建新的goroutine，它有两个参数，先说第二个参数fn，新创建出来的goroutine将从fn这个函数开始执行，而这个fn函数可能也会有参数，newproc的第一个参数正是fn函数的参数以字节为单位的大小。
	//newproc函数将创建一个新的goroutine来执行fn函数，而这个新创建的goroutine与当前这个goroutine会使用不同的栈，因此就需要在创建goroutine的时候把fn需要用到的参数先从当前goroutine的栈上拷贝到新的goroutine的栈上之后才能让其开始执行，而newproc函数本身并不知道需要拷贝多少数据到新创建的goroutine的栈上去，所以需要用参数的方式指定拷贝多少数据。
	CALL	runtime·newproc(SB)
	POPQ	AX
	POPQ	AX

	// start this M
	//主线程进入调度循环，运行刚刚创建的goroutine
	CALL	runtime·mstart(SB)

    //上面的mstart永远不应该返回的，如果返回了，一定是代码逻辑有问题，直接abort
	CALL	runtime·abort(SB)	// mstart should never return
	RET

	// Prevent dead-code elimination of debugCallV1, which is
	// intended to be called by debuggers.
	MOVQ	$runtime·debugCallV1(SB), AX
	RET

    
```

```go

type funcval struct {
    fn uintptr
    // variable-size, fn-specific data here
}
//runtime/proc.go
//go:nosplit 无需检查栈伸缩
func newproc(siz int32, fn *funcval) {
    //函数调用参数入栈顺序是从右向左，而且栈是从高地址向低地址增长的
    //注意：argp指向fn函数的第一个参数，而不是newproc函数的参数
    //参数fn在栈上的地址+8的位置存放的是fn函数的第一个参数
	//获取参数大小，后面用于拷贝参数到新goroutine栈
	argp := add(unsafe.Pointer(&fn), sys.PtrSize)
	////获取正在运行的g，初始化时是m0.g0
	gp := getg()
    //getcallerpc()返回一个地址，也就是调用newproc时由call指令压栈的函数返回地址，
    //对于我们现在这个场景来说，pc就是CALLruntime·newproc(SB)指令后面的POPQ AX这条指令的地址
	pc := getcallerpc()//POPQ	AX 指令地址
    //systemstack的作用是切换到g0栈执行作为参数的函数
    //我们这个场景现在本身就在g0栈，因此什么也不做，直接调用作为参数的函数
	systemstack(func() {
		//第一个参数fn是新创建的goroutine需要执行的函数
		//第二个参数argp是fn函数的第一个参数的地址
		//第三个参数是fn函数的参数以字节为单位的大小
		//第四个和第五个当前g和下一条执行指令地址
		newproc1(fn, argp, siz, gp, pc)
	})
}

举例：
func xxx(x,y,z int64) {
	
}

func main(){
	go xxx(1,2,3)
}
```

|--|栈底（高地址）|
|-----------| ----------- |
|-|fn:参数3|
|-|fn:参数2|
|-|fn:参数1|
|-|fn:xxx函数地址|

```go
    // newproc1()
    // Create a new g running fn with narg bytes of arguments starting
    // at argp. callerpc is the address of the go statement that created
    // this. The new g is put on the queue of g's waiting to run.
    func newproc1(fn *funcval, argp unsafe.Pointer, narg int32, callergp *g, callerpc uintptr) {
        //因为已经切换到g0栈，所以无论什么场景都有 _g_ = g0，当然这个g0是指当前工作线程的g0
        //对于我们这个场景来说，当前工作线程是主线程，所以这里的g0 = m0.g0，其他场景对应的goroutine
    	_g_ := getg()
        
        if fn == nil {
            _g_.m.throwing = -1 // do not dump full stacks
            throw("go of nil func value")
        }
        acquirem() // disable preemption because it can be holding p in a local var
        siz := narg
        siz = (siz + 7) &^ 7
        
        // We could allocate a larger initial stack if necessary.
        // Not worth it: this is almost always an error.
        // 4*sizeof(uintreg): extra space added below
        // sizeof(uintreg): caller's LR (arm) or return address (x86, in gostartcall).
        if siz >= _StackMin-4*sys.RegSize-sys.RegSize {
            throw("newproc: function arguments too large for new goroutine")
        }
        //初始化时_p_ = g0.m.p，从前面的分析可以知道其实就是allp[0]
        //这里说的初始化时，其他情况获取对应的p
        _p_ := _g_.m.p.ptr()
        //从p的本地缓冲里获取一个没有使用的g，初始化时没有，返回nil
        newg := gfget(_p_)
        //没有g，创建一个
        if newg == nil {
            //new一个g结构体对象，然后从堆上为其分配栈，并设置g的stack成员和两个stackgard成员
            newg = malg(_StackMin)
            casgstatus(newg, _Gidle, _Gdead)//初始化g的状态为_Gdead
            //放入全局变量allgs切片中
            allgadd(newg) // publishes with a g->status of Gdead so GC scanner doesn't look at uninitialized stack.
        }
        if newg.stack.hi == 0 {
            throw("newproc1: newg missing stack")
        }
        
        if readgstatus(newg) != _Gdead {
            throw("newproc1: new g is not Gdead")
        }
        //查资料说是调整g的栈顶置针，不是很理解
        totalSize := 4*sys.RegSize + uintptr(siz) + sys.MinFrameSize // extra space in case of reads slightly beyond frame
        totalSize += -totalSize & (sys.SpAlign - 1)                  // align to spAlign
        sp := newg.stack.hi - totalSize
        spArg := sp
        if usesLR {
            // caller's LR
            *(*uintptr)(unsafe.Pointer(sp)) = 0
            prepGoExitFrame(sp)
            spArg += sys.MinFrameSize
        }
        if narg > 0 {
        	//把参数从执行newproc函数的栈（初始化时是g0栈）拷贝到新g的栈
            memmove(unsafe.Pointer(spArg), argp, uintptr(narg))
            // This is a stack-to-stack copy. If write barriers
            // are enabled and the source stack is grey (the
            // destination is always black), then perform a
            // barrier copy. We do this *after* the memmove
            // because the destination stack may have garbage on
            // it.
            if writeBarrier.needed && !_g_.m.curg.gcscandone {
                f := findfunc(fn.fn)
                stkmap := (*stackmap)(funcdata(f, _FUNCDATA_ArgsPointerMaps))
                if stkmap.nbit > 0 {
                    // We're in the prologue, so it's always stack map index 0.
                    bv := stackmapdata(stkmap, 0)
                    bulkBarrierBitmap(spArg, spArg, uintptr(bv.n)*sys.PtrSize, 0, bv.bytedata)
                }
            }
        }
        //上面代码从堆上分配一个g结构体对象并为这个newg分配一个大小为2048字节（2k大小的栈空间）的栈，并设置好newg的stack成员，然后把newg需要执行的函数的参数从执行newproc函数的栈（初始化时是g0栈）拷贝到newg的栈，malg()->newg的stack.hi和stack.lo分别指向了其栈空间的起止位置。
        
        //把newg.sched结构体成员的所有成员设置为0
        memclrNoHeapPointers(unsafe.Pointer(&newg.sched), unsafe.Sizeof(newg.sched))
        //设置newg的sched成员，调度器需要依靠这些字段才能把goroutine调度到CPU上运行。
        newg.sched.sp = sp//newg的栈顶
        newg.stktopsp = sp
        //骚操作，非用户main goroutine 的下一条命令设置为goexit的命令的下一行指令
        //这样做原因是让goroutine退出后可以继续循环调度，并且打扫现场
        //newg.sched.pc表示当newg被调度起来运行时从这个地址开始执行指令
        //把pc设置成了goexit这个函数偏移1（sys.PCQuantum等于1）的位置，
        newg.sched.pc = funcPC(goexit) + sys.PCQuantum // +PCQuantum so that previous instruction is in same function
        newg.sched.g = guintptr(unsafe.Pointer(newg))
        //调整sched成员和newg的栈
        gostartcallfn(&newg.sched, fn)
(
    // adjust Gobuf as if it executed a call to fn
    // and then did an immediate gosave.
    func gostartcallfn(gobuf *gobuf, fv *funcval) {
        var fn unsafe.Pointer
        if fv != nil {
            //fn: goroutine的入口地址，不同goroutine不同，这里是初始化时对应的是runtime.main
            fn = unsafe.Pointer(fv.fn)
        } else {
            fn = unsafe.Pointer(funcPC(nilfunc))
        }
        //gostartcallfn首先从参数fv中提取出函数地址（初始化时是runtime.main），然后继续调用gostartcall函数
        gostartcall(gobuf, fn, unsafe.Pointer(fv))
    }

    //继续跟踪下去
    // adjust Gobuf as if it executed a call to fn with context ctxt
    // and then did an immediate gosave.
    func gostartcall(buf *gobuf, fn, ctxt unsafe.Pointer) {
    	//newg的栈顶，目前newg栈上只有fn函数的参数，sp指向的是fn的第一参数
        sp := buf.sp
        if sys.RegSize > sys.PtrSize {
            sp -= sys.PtrSize
            *(*uintptr)(unsafe.Pointer(sp)) = 0
        }
        //为返回地址预留空间
        sp -= sys.PtrSize
        //这里在伪装fn是被goexit函数调用的，使得fn执行完后返回到goexit继续执行，从而完成清理工作
        *(*uintptr)(unsafe.Pointer(sp)) = buf.pc//返回地址：在栈上放入goexit+1的地址
        buf.sp = sp//重新设置newg的栈顶寄存器
        //这里才真正让newg的ip寄存器指向fn函数，注意，这里只是在设置newg的一些信息，newg还未执行
        //等到newg被调度起来运行时，调度器会把buf.pc放入cpu的IP寄存器
        //从而使newg得以在cpu上真正的运行起来
        //这个场景为runtime.main函数的地址，不同goroutine函数不同
        buf.pc = uintptr(fn)
        buf.ctxt = ctxt
    }
)
        //主要用于traceback
        newg.gopc = callerpc
        newg.ancestors = saveAncestors(callergp)
        //设置newg的startpc为fn.fn，该成员主要用于函数调用栈的traceback和栈收缩
        //newg真正从哪里开始执行并不依赖于这个成员，而是sched.pc
        newg.startpc = fn.fn
        if _g_.m.curg != nil {
            newg.labels = _g_.m.curg.labels
        }
        if isSystemGoroutine(newg, false) {
            atomic.Xadd(&sched.ngsys, +1)
        }
        //设置g的状态为_Grunnable，表示这个g代表的goroutine可以运行了
        casgstatus(newg, _Gdead, _Grunnable)
        
        if _p_.goidcache == _p_.goidcacheend {
            // Sched.goidgen is the last allocated id,
            // this batch must be [sched.goidgen+1, sched.goidgen+GoidCacheBatch].
            // At startup sched.goidgen=0, so main goroutine receives goid=1.
            _p_.goidcache = atomic.Xadd64(&sched.goidgen, _GoidCacheBatch)
            _p_.goidcache -= _GoidCacheBatch - 1
            _p_.goidcacheend = _p_.goidcache + _GoidCacheBatch
        }
        newg.goid = int64(_p_.goidcache)
        _p_.goidcache++
        if raceenabled {
            newg.racectx = racegostart(callerpc)
        }
        if trace.enabled {
            traceGoCreate(newg, newg.startpc)
        }
        //把newg放入_p_的运行队列，初始化的时候一定是p的本地运行队列，其它时候可能因为本地队列满了而放入全局队列
        runqput(_p_, newg, true)
        
        // 如果当前有空闲的P，但是没有自旋的M(nmspinning等于0)，并且主函数已执行，则唤醒或新建一个M来调度一个P执行，意思是没有正在找g运行的p和m时，那就启动把
		//唤醒或新建一个M会通过调用wakep函数
            //首先交换nmspinning到1, 成功再继续, 多个线程同时执行wakep函数只有一个会继续
            //调用startm函数
            //    调用pidleget从"空闲P链表"获取一个空闲的P
            //    调用mget从"空闲M链表"获取一个空闲的M
            //    如果没有空闲的M, 则调用newm新建一个M
            //        newm会新建一个m的实例, m的实例包含一个g0, 然后调用newosprocclone一个系统线程
            //        newosproc会调用syscall clone创建一个新的线程
            //        线程创建后会设置TLS, 设置TLS中当前的g为g0, 然后执行mstart
            //    调用notewakeup(&mp.park)唤醒线程
        if atomic.Load(&sched.npidle) != 0 && atomic.Load(&sched.nmspinning) == 0 && mainStarted {
            wakep()
        }
        releasem(_g_.m)
    }
    // 可以参考：https://blog.csdn.net/u010853261/article/details/84790392
```
## runqput()

```go
// 把g放在本地队列，其实先放runnext
// runqput tries to put g on the local runnable queue.
// 如果next = false 就直接放在本地队列
// If next is false, runqput adds g to the tail of the runnable queue.
// 如果next = true 就直接runnext放
// If next is true, runqput puts g in the _p_.runnext slot.
// 如果runq满了，放在全局队列
// If the run queue is full, runnext puts g on the global queue.
// Executed only by the owner P.
//runqput(_p_, newg, true)
func runqput(_p_ *p, gp *g, next bool) {
	if randomizeScheduler && next && fastrand()%2 == 0 {
		next = false
	}

	if next {
	retryNext:
		//runnext看下是否有g，并且有没有其他线程在并发操作
		oldnext := _p_.runnext
		//有其他线程在并发操作那再跳到retryNext
		if !_p_.runnext.cas(oldnext, guintptr(unsafe.Pointer(gp))) {
			goto retryNext
		}
		//如果原来是空，那就直接返回了
		if oldnext == 0 {
			return
		}
		// Kick the old runnext out to the regular run queue.
		// 原本存放在runnext的gp需要放入runq的尾部，逻辑在后面
		gp = oldnext.ptr()
	}

retry:
	//可能有其它线程正在并发修改runqhead成员，所以需要跟其它线程同步，偷g也有可能的
	h := atomic.LoadAcq(&_p_.runqhead) // load-acquire, synchronize with consumers
	t := _p_.runqtail
	//判断队列是否满了（256），没有满入到if里面
	if t-h < uint32(len(_p_.runq)) {
		//队列还没有满，可以放入
		_p_.runq[t%uint32(len(_p_.runq))].set(gp)
        //虽然没有其它线程并发修改这个runqtail，但其它线程会并发读取该值以及p的runq成员
        //这里使用StoreRel是为了：
        //1，原子写入runqtail
        //2，防止编译器和CPU乱序，保证上一行代码对runq的修改发生在修改runqtail之前
        //3，可见行屏障，保证当前线程对运行队列的修改对其它线程立马可见
		atomic.StoreRel(&_p_.runqtail, t+1) // store-release, makes the item available for consumption
		return
	}
	//p的本地运行队列已满，需要放入全局运行队列
	if runqputslow(_p_, gp, h, t) {
		return
	}
	// the queue is not full, now the put above must succeed
	goto retry//兜底
}
```

```go
// Put g and a batch of work from local runnable queue on global queue.
// Executed only by the owner P.
func runqputslow(_p_ *p, gp *g, h, t uint32) bool {
	//获取本地队列的一半+1自身（gp）
	var batch [len(_p_.runq)/2 + 1]*g

	// First, grab a batch from local queue.
	n := t - h
	n = n / 2
	//到了这里，应该是本地队列是满了的，否则就是有问题了
	if n != uint32(len(_p_.runq)/2) {
		throw("runqputslow: queue is not full")
	}
	//取本地队列的一半到batch
	for i := uint32(0); i < n; i++ {
		batch[i] = _p_.runq[(h+i)%uint32(len(_p_.runq))].ptr()
	}
	if !atomic.CasRel(&_p_.runqhead, h, h+n) { // cas-release, commits consume
		//如果cas操作失败，说明已经有其它工作线程从_p_的本地运行队列偷走了一些goroutine，所以直接返回
		return false
	}
	//把gp加上
	batch[n] = gp

	//打乱顺序，忽略，这个应该是调试的时候用到
	if randomizeScheduler {
		for i := uint32(1); i <= n; i++ {
			j := fastrandn(i + 1)
			batch[i], batch[j] = batch[j], batch[i]
		}
	}

	// Link the goroutines.
	//全局运行队列是一个链表，这里首先把所有需要放入全局运行队列的g链接起来，
	//减少后面对全局链表的锁住时间，从而降低锁冲突
	for i := uint32(0); i < n; i++ {
		batch[i].schedlink.set(batch[i+1])
	}
	//形成链表
	var q gQueue
	q.head.set(batch[0])
	q.tail.set(batch[n])

	// Now put the batch on global queue.
	//操作全局队列，需要加锁，尽量小的锁
	lock(&sched.lock)
	//行吧，批量直接放入全局队列了
	globrunqputbatch(&q, int32(n+1))
	unlock(&sched.lock)
	return true
}

// Put a batch of runnable goroutines on the global runnable queue.
// This clears *batch.
// Sched must be locked.
func globrunqputbatch(batch *gQueue, n int32) {
    sched.runq.pushBackAll(*batch)
    sched.runqsize += n
    *batch = gQueue{}
}
```
