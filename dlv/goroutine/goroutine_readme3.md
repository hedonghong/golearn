## 经过前两个md，创建main goroutine后，如何把main goroutine进行调度运行
    g0做一些调度的工作，其余需要切换到对应的g中去进行

```go
    //回忆下我们创建了main goroutine后后面的汇编指令
    // create a new goroutine to start program
    MOVQ	$runtime·mainPC(SB), AX		// entry
    PUSHQ	AX
    PUSHQ	$0			// arg size
    CALL	runtime·newproc(SB)
    POPQ	AX
    POPQ	AX
    
    // start this M
    CALL	runtime·mstart(SB)

    //接下来看看mstart()如何运行调度main goroutine


	// mstart is the entry-point for new Ms.
	//
	// This must not split the stack because we may not even have stack
	// bounds set up yet.
	//
	// May run during STW (because it doesn't have a P yet), so write
	// barriers are not allowed.
	// 不用伸缩栈检测
	//go:nosplit
	// 不用写屏障
	//go:nowritebarrierrec
	func mstart() {
		//_g_ = g0
        _g_ := getg()
        
        //对于启动过程来说，g0的stack.lo早已完成初始化，所以onStack = false
        //下面会根据判断设置goroutine的栈范围
        osStack := _g_.stack.lo == 0
        if osStack {
            // Initialize stack bounds from system stack.
            // Cgo may have left stack size in stack.hi.
            // minit may update the stack bounds.
            size := _g_.stack.hi
            if size == 0 {
                size = 8192 * sys.StackGuardMultiplier
            }
            _g_.stack.hi = uintptr(noescape(unsafe.Pointer(&size)))
            _g_.stack.lo = _g_.stack.hi - size + 1024
        }
        // Initialize stack guard so that we can start calling regular
        // Go code.
        _g_.stackguard0 = _g_.stack.lo + _StackGuard
        // This is the g0, so we can also call go:systemstack
        // functions, which check stackguard1.
        _g_.stackguard1 = _g_.stackguard0
        mstart1()
        
        // Exit this thread.
        switch GOOS {
            case "windows", "solaris", "illumos", "plan9", "darwin", "aix":
            // Windows, Solaris, illumos, Darwin, AIX and Plan 9 always system-allocate
            // the stack, but put it in _g_.stack before mstart,
            // so the logic above hasn't set osStack yet.
            osStack = true
        }
        mexit(osStack)
        //为什么g0已经执行到mstart1这个函数了而且还会继续调用其它函数，但g0的调度信息中的pc和sp却要设置在mstart函数中？难道下次切换到g0时要从mstart函数中的 if 语句继续执行？可是从mstart函数可以看到，if语句之后就要退出线程了！这看起来很奇怪，不过随着分析的进行，我们会看到这里为什么要这么做。
    }

    func mstart1() {
    	//启动过程时 _g_ = m0的g0
        _g_ := getg()
        
        if _g_ != _g_.m.g0 {
            throw("bad runtime·mstart")
        }
        
        // Record the caller for use as the top of stack in mcall and
        // for terminating the thread.
        // We're never coming back to mstart1 after we call schedule,
        // so other calls can reuse the current frame.
		//getcallerpc()获取mstart1执行完的返回地址，由于不返回，实际用不到
		//就是这句上面的 Exit this thread. switch GOOS { 这个指令
		//getcallersp()获取调用mstart1时的栈顶地址
		//代码中的getcallerpc()返回的是mstart调用mstart1时被call指令压栈的返回地址，getcallersp()函数返回的是调用mstart1函数之前mstart函数的栈顶地址
		//save函数来保存g0的调度信息，save这一行代码非常重要，是我们理解调度循环的关键点之一
        save(getcallerpc(), getcallersp())
        asminit()//在AMD64 Linux平台中，这个函数什么也没做，是个空函数
        minit()//与信号相关的初始化
        
        // Install signal handlers; after minit so that minit can
        // prepare the thread to be able to handle the signals.
        //启动时_g_.m是m0，所以会执行下面的mstartm0函数
        if _g_.m == &m0 {
        	//信号相关的初始化：注册信息回调处理函数等
            mstartm0()
        }
        
        //初始化过程中fn == nil
        if fn := _g_.m.mstartfn; fn != nil {
            fn()
        }
        //m0已经绑定了allp[0]，不是m0的话还没有p，所以需要获取一个p
        if _g_.m != &m0 {
            acquirep(_g_.m.nextp.ptr())
            _g_.m.nextp = 0
        }
        //schedule函数永远不会返回，真的循环调度从这里开始
        schedule()
        //schedule -> execute -> gogo -> goexit -> schedule
    }

    // save updates getg().sched to refer to pc and sp so that a following
    // gogo will restore pc and sp.
    //
    // save must not have write barriers because invoking a write barrier
    // can clobber getg().sched.
    //
    //go:nosplit
    //go:nowritebarrierrec
    //初始化更新保存_g_的sched调度信息
    func save(pc, sp uintptr) {
        _g_ := getg()
        
        _g_.sched.pc = pc
        _g_.sched.sp = sp
        _g_.sched.lr = 0
        _g_.sched.ret = 0
        _g_.sched.g = guintptr(unsafe.Pointer(_g_))
        // We need to ensure ctxt is zero, but can't have a write
        // barrier here. However, it should always already be zero.
        // Assert that.
        if _g_.sched.ctxt != nil {
            badctxt()
        }
    }
```