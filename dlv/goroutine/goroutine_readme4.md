## 真的调度起步
    schedule -> execute -> gogo -> goexit -> schedule
```go
// One round of scheduler: find a runnable goroutine and execute it.
// Never returns.
func schedule() {
	//_g_ = 每个工作线程m对应的g0，初始化时是m0的g0
	_g_ := getg()

	if _g_.m.locks != 0 {
		throw("schedule: holding locks")
	}

	if _g_.m.lockedg != 0 {
		stoplockedm()
		execute(_g_.m.lockedg.ptr(), false) // Never returns.
	}

	// We should not schedule away from a g that is executing a cgo call,
	// since the cgo call is using the m's g0 stack.
	if _g_.m.incgo {
		throw("schedule: in cgo")
	}
	//总的思路从各处获取需要运行的g，获取不到再休眠，否则一直在运行g或获取g

top:
	pp := _g_.m.p.ptr()
	pp.preempt = false

	if sched.gcwaiting != 0 {
		gcstopm()
		goto top
	}
	if pp.runSafePointFn != 0 {
		runSafePointFn()
	}

	// Sanity check: if we are spinning, the run queue should be empty.
	// Check this before calling checkTimers, as that might call
	// goready to put a ready goroutine on the local run queue.
	if _g_.m.spinning && (pp.runnext != 0 || pp.runqhead != pp.runqtail) {
		throw("schedule: spinning with local work")
	}

	//timer堆相关，后面了解golang如何做定时器的看这里
	checkTimers(pp, 0)

	var gp *g
	var inheritTime bool

	// Normal goroutines will check for need to wakeP in ready,
	// but GCworkers and tracereaders will not, so the check must
	// be done here instead.
	tryWakeP := false
	if trace.enabled || trace.shutdown {
		gp = traceReader()
		if gp != nil {
			casgstatus(gp, _Gwaiting, _Grunnable)
			traceGoUnpark(gp, 0)
			tryWakeP = true
		}
	}
	if gp == nil && gcBlackenEnabled != 0 {
		gp = gcController.findRunnableGCWorker(_g_.m.p.ptr())
		tryWakeP = tryWakeP || gp != nil
	}
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

	// This thread is going to run a goroutine and is not spinning anymore,
	// so if it was marked as spinning we need to reset it now and potentially
	// start a new spinning M.
	// m 是 spinning 自旋中
	if _g_.m.spinning {
		resetspinning()
	}

	if sched.disable.user && !schedEnabled(gp) {
		// Scheduling of this goroutine is disabled. Put it on
		// the list of pending runnable goroutines for when we
		// re-enable user scheduling and look again.
		lock(&sched.lock)
		if schedEnabled(gp) {
			// Something re-enabled scheduling while we
			// were acquiring the lock.
			unlock(&sched.lock)
		} else {
			sched.disable.runnable.pushBack(gp)
			sched.disable.n++
			unlock(&sched.lock)
			goto top
		}
	}

	// If about to schedule a not-normal goroutine (a GCworker or tracereader),
	// wake a P if there is one.
	if tryWakeP {
		if atomic.Load(&sched.npidle) != 0 && atomic.Load(&sched.nmspinning) == 0 {
			wakep()
		}
	}
	if gp.lockedm != 0 {
		// Hands off own p to the locked m,
		// then blocks waiting for a new p.
		startlockedm(gp)
		goto top
	}
	//当前运行的是runtime的代码，函数调用栈使用的是g0的栈空间
    //调用execte切换到gp的代码和栈空间去运行
    //不过execte 只是做切换准备
	execute(gp, inheritTime)
}

// 暂时不管schedule细节，先去execute看看，了解一个大概
// 上面schedule获取的是main goroutine取出来运行

// Schedules gp to run on the current M.
// If inheritTime is true, gp inherits the remaining time in the
// current time slice. Otherwise, it starts a new time slice.
// Never returns.
//
// Write barriers are allowed because this is called immediately after
// acquiring a P in several places.
//
//go:yeswritebarrierrec
func execute(gp *g, inheritTime bool) {
	//获取g0
    _g_ := getg()
    
    // Assign gp.m before entering _Grunning so running Gs have an
    // M.
    //记录当前运行的g
    _g_.m.curg = gp
    //当前的g关联m
    gp.m = _g_.m
    //设置当前g的状态为_Grunning
    casgstatus(gp, _Grunnable, _Grunning)
    gp.waitsince = 0
    gp.preempt = false
    //重新设置stackguard0
    gp.stackguard0 = gp.stack.lo + _StackGuard
    if !inheritTime {
        _g_.m.p.ptr().schedtick++
    }
    
    // Check whether the profiler needs to be turned on or off.
    hz := sched.profilehz
    if _g_.m.profilehz != hz {
        setThreadCPUProfiler(hz)
    }
    
    if trace.enabled {
        // GoSysExit has to happen when we have a P, but before GoStart.
        // So we emit it here.
        if gp.syscallsp != 0 && gp.sysblocktraced {
            traceGoSysExit(gp.sysexitticks)
        }
        traceGoStart()
    }
    //gogo函数完成从g0到gp的的切换：CPU执行权的转让以及栈的切换
    //goroutine的切换从本质上来说就是CPU寄存器以及函数调用栈的切换，然而不管是go还是c这种高级语言都无法精确控制CPU寄存器的修改，因而高级语言在这里也就无能为力了，只能依靠汇编指令来达成目的
    //需要调用切换寄存器和栈内存的保存的信息，如sp pc等
    //回忆下gp.sched保存的是什么，忘记的可以翻下goroutine_readme0.md
    gogo(&gp.sched)
}

// runtime/asm_amd64.s
// func gogo(buf *gobuf)
// restore state from Gobuf; longjmp
TEXT runtime·gogo(SB), NOSPLIT, $16-8
MOVQ	buf+0(FP), BX		// gobuf gp.sched --> BX BX = buf
MOVQ	gobuf_g(BX), DX // DX = gp.sched.g
//下面这行代码没有实质作用，检查gp.sched.g是否是nil，如果是nil进程会crash死掉
MOVQ	0(DX), CX		// make sure g != nil
get_tls(CX)
//把要运行的g的指针放入线程本地存储，这样后面的代码就可以通过线程本地存储
//获取到当前正在执行的goroutine的g结构体对象，从而找到与之关联的m和p
MOVQ	DX, g(CX)
//把CPU的SP寄存器设置为sched.sp，完成了栈的切换
//接后面的命令也是一样设置调度上下文到CPU相关寄存器
MOVQ	gobuf_sp(BX), SP	// restore SP
MOVQ	gobuf_ret(BX), AX
MOVQ	gobuf_ctxt(BX), DX
MOVQ	gobuf_bp(BX), BP
//清空sched的值，因为我们已把相关值放入CPU对应的寄存器了，不再需要，这样做可以少gc的工作量
MOVQ	$0, gobuf_sp(BX)	// clear to help garbage collector
MOVQ	$0, gobuf_ret(BX)
MOVQ	$0, gobuf_ctxt(BX)
MOVQ	$0, gobuf_bp(BX)
//把sched.pc值放入BX寄存器
MOVQ	gobuf_pc(BX), BX
//JMP把BX寄存器的包含的地址值放入CPU的IP寄存器，于是，CPU跳转到该地址继续执行指令
JMP	BX

//1、把gp.sched的成员恢复到CPU的寄存器完成状态以及栈的切换；
//2、跳转到gp.sched.pc所指的指令地址（runtime.main）处执行。


// runtime/proc.go
// The main goroutine.
// 初始化后运行的第一个内容就是runtime.main()
func main() {
	//g = main goroutine，不再是g0了
	g := getg()
    
    // Racectx of m0->g0 is used only as the parent of the main goroutine.
    // It must not be used for anything else.
    g.m.g0.racectx = 0
    
    // Max stack size is 1 GB on 64-bit, 250 MB on 32-bit.
    // Using decimal instead of binary GB and MB because
    // they look nicer in the stack overflow failure message.
    // 64位系统上每个goroutine的栈最大可达1G，32位是 250 MB
    if sys.PtrSize == 8 {
        maxstacksize = 1000000000
    } else {
        maxstacksize = 250000000
    }
    
    // Allow newproc to start new Ms.
    mainStarted = true
    
    if GOARCH != "wasm" { // no threads on wasm yet, so no sysmon
    	//现在执行的是main goroutine，所以使用的是main goroutine的栈，需要切换到g0栈去执行newm()
    	//重点细节兜底监控线程sysmon，做一些公平调度，抢占等工作
        systemstack(func() {
        	//创建监控线程，该线程独立于调度器，不需要跟p关联即可运行
            newm(sysmon, nil, -1)
        })
    }
    
    // Lock the main goroutine onto this, the main OS thread,
    // during initialization. Most programs won't care, but a few
    // do require certain calls to be made by the main thread.
    // Those can arrange for main.main to run in the main thread
    // by calling runtime.LockOSThread during initialization
    // to preserve the lock.
    lockOSThread()
    
    if g.m != &m0 {
        throw("runtime.main not on m0")
    }
    
    ///调用runtime包的初始化函数
    doInit(&runtime_inittask) // must be before defer
    if nanotime() == 0 {
        throw("nanotime returning zero")
    }
    
    // Defer unlock so that runtime.Goexit during init does the unlock too.
    needUnlock := true
    defer func() {
        if needUnlock {
            unlockOSThread()
        }
    }()
    
    // Record when the world started.
    runtimeInitTime = nanotime()
    
    //开启垃圾回收器
    gcenable()
    
    main_init_done = make(chan bool)
    if iscgo {
        if _cgo_thread_start == nil {
            throw("_cgo_thread_start missing")
        }
        if GOOS != "windows" {
            if _cgo_setenv == nil {
                throw("_cgo_setenv missing")
            }
            if _cgo_unsetenv == nil {
                throw("_cgo_unsetenv missing")
            }
        }
        if _cgo_notify_runtime_init_done == nil {
            throw("_cgo_notify_runtime_init_done missing")
        }
        // Start the template thread in case we enter Go from
        // a C-created thread and need to create a new thread.
        startTemplateThread()
        cgocall(_cgo_notify_runtime_init_done, nil)
    }
    
    //用户编写的main包的初始化函数，会递归的调用我们import进来的包的初始化函数
    doInit(&main_inittask)
    
    close(main_init_done)
    
    needUnlock = false
    unlockOSThread()
    
    if isarchive || islibrary {
        // A program compiled with -buildmode=c-archive or c-shared
        // has a main, but it is not executed.
        return
    }
    fn := main_main // make an indirect call, as the linker doesn't know the address of the main package when laying down the runtime
	//调用用户的main.main函数
    fn()
    if raceenabled {
        racefini()
    }
    
    // Make racy client program work: if panicking on
    // another goroutine at the same time as main returns,
    // let the other goroutine finish printing the panic trace.
    // Once it does, it will exit. See issues 3934 and 20018.
    if atomic.Load(&runningPanicDefers) != 0 {
        // Running deferred functions should not take long.
        for c := 0; c < 1000; c++ {
            if atomic.Load(&runningPanicDefers) == 0 {
                break
            }
            Gosched()
        }
    }
    if atomic.Load(&panicking) != 0 {
        gopark(nil, nil, waitReasonPanicWait, traceEvGoStop, 1)
    }
    
    //进入系统调用，退出进程，可以看出main goroutine并未返回，而是直接进入系统调用退出进程了
    //所以在main函数运行完之后退出了，不等其他goroutine，注意这个是这个runtime.main
    //与其他goroutine的区别
    exit(0)
    //保护性代码，如果exit意外返回，下面的代码也会让该进程crash死掉
    for {
        var x *int32
        *x = 0
    }
}

1、启动一个sysmon系统监控线程，该线程负责整个程序的gc、抢占调度以及netpoll等功能的监控，在抢占调度一章我们再继续分析sysmon是如何协助完成goroutine的抢占调度的；

2、执行runtime包的初始化；

3、执行main包以及main包import的所有包的初始化；

4、执行main.main函数；

5、从main.main函数返回后调用exit系统调用退出进程；


runtime.main执行完main包的main函数之后就直接调用exit系统调用结束进程了，它并没有返回到调用它的函数（还记得是从哪里开始执行的runtime.main吗？），其实runtime.main是main goroutine的入口函数，并不是直接被调用的，而是在schedule()->execute()->gogo()这个调用链的gogo函数中用汇编代码直接跳转过来的，所以从这个角度来说，goroutine确实不应该返回，没有地方可返回啊！可是从前面的分析中我们得知，在创建goroutine的时候已经在其栈上放好了一个返回地址，伪造成goexit函数调用了goroutine的入口函数，这里怎么没有用到这个返回地址啊？其实那是为非main goroutine准备的，非main goroutine执行完成后就会返回到goexit继续执行，而main goroutine执行完成后整个进程就结束了，这是main goroutine与其它goroutine的一个区别


```