
## 主程启动

```go
TEXT _rt0_amd64(SB),NOSPLIT,$-8
	MOVQ	0(SP), DI	// argc
	LEAQ	8(SP), SI	// argv
	JMP	runtime·rt0_go(SB)

TEXT runtime·rt0_go(SB),NOSPLIT,$0
    // copy arguments forward on an even stack
    MOVQ	DI, AX		// argc
    MOVQ	SI, BX		// argv
    SUBQ	$(4*8+7), SP		// 2args 2auto
    //调整栈顶寄存器使其按16字节对齐
	//让栈顶寄存器SP指向的内存的地址为16的倍数，之所以要按16字节对齐，是因为CPU有一组SSE指令，这些指令中出现的内存地址必须是16的倍数，最后两条指令把argc和argv搬到新的位置
    ANDQ	$~15, SP
    //argc放在SP+ 16字节处
    MOVQ	AX, 16(SP)
    //argv放在SP+ 24字节处
    MOVQ	BX, 24(SP)

    // create istack out of the given (operating system) stack.
    // _cgo_init may update stackguard.
    //g0的主要作用是提供一个栈供runtime代码执行，因此这里主要对g0的几个与栈有关的成员进行了初始化，从这里可以看出g0的栈大约有64K，地址范围为 SP - 64*1024 + 104 ～ SP
    MOVQ	$runtime·g0(SB), DI//g0的地址放入DI寄存器
    //从系统线程的栈空分出一部分当作g0的栈，然后初始化g0的栈信息和stackgard
    LEAQ	(-64*1024+104)(SP), BX//BX=SP- 64*1024 + 104
    MOVQ	BX, g_stackguard0(DI)//g0.stackguard0 =SP- 64*1024 + 104
    MOVQ	BX, g_stackguard1(DI)//g0.stackguard1 =SP- 64*1024 + 104
    MOVQ	BX, (g_stack+stack_lo)(DI)//g0.stack.lo =SP- 64*1024 + 104
    MOVQ	SP, (g_stack+stack_hi)(DI)//g0.stack.hi =SP

    ....//CPU型号检查以及cgo初始化相关的代码

    LEAQ	runtime·m0+m_tls(SB), DI//取m0的tls成员的地址到DI寄存器
    CALL	runtime·settls(SB)//调用settls设置线程本地存储，settls函数的参数在DI寄存器中
    
(//跳到settls 看看
    TEXT runtime·settls(SB),NOSPLIT,$32
    #ifdef GOOS_android
    // Android stores the TLS offset in runtime·tls_g.
    SUBQ	runtime·tls_g(SB), DI
    #else
    //DI寄存器中存放的是m.tls[0]的地址，m的tls成员是一个数组，读者如果忘记了可以回头看一下m结构体的定义
    //下面这一句代码把DI寄存器中的地址加8，为什么要+8呢，主要跟ELF可执行文件格式中的TLS实现的机制有关
    //执行下面这句指令之后DI寄存器中的存放的就是m.tls[1]的地址了
    ADDQ	$8, DI	// ELF wants to use -8(FS)
    #endif
    //下面通过arch_prctl系统调用设置FS段基址
    MOVQ	DI, SI//SI存放arch_prctl系统调用的第二个参数
    MOVQ	$0x1002, DI	// ARCH_SET_FS//arch_prctl的第一个参数
    MOVQ	$SYS_arch_prctl, AX//系统调用编号
    SYSCALL//调用
    CMPQ	AX, $0xfffffffffffff001//比较AX 小于等于$0xfffffffffffff001 正常RET返回，否则crash掉
    JLS	2(PC)
    MOVL	$0xf1, 0xf1  // crash
    RET
	//这里通过arch_prctl系统调用把m0.tls[1]的地址设置成了fs段的段基址。CPU中有个叫fs的段寄存器与之对应，而每个线程都有自己的一组CPU寄存器值，操作系统在把线程调离CPU运行时会帮我们把所有寄存器中的值保存在内存中，调度线程起来运行时又会从内存中把这些寄存器的值恢复到CPU，这样，在此之后，工作线程代码就可以通过fs寄存器来找到m.tls
)
    
    // store through it, to make sure it works
    //runtime/go_tls.h 定义了get_tls
    //#define	get_tls(r)	MOVQ TLS, r
    //把当前g赋予给r
    //TLS 是一个由 runtime 维护的虚拟寄存器，保存了指向当前 g 的指针
    //验证settls是否可以正常工作，如果有问题则abort退出程序
    get_tls(BX)//获取fs段基地址并放入BX寄存器，其实就是m0.tls[1]的地址，get_tls的代码由编译器生成
    MOVQ	$0x123, g(BX)//把整型常量0x123拷贝到fs段基地址偏移-8的内存位置，也就是m0.tls[0] =0x123
    MOVQ	runtime·m0+m_tls(SB), AX//AX=m0.tls[0]
    CMPQ	AX, $0x123//检查m0.tls[0]的值是否是通过线程本地存储存入的0x123来验证tls功能是否正常
    JEQ 2(PC)//跳过两行指令，就是ok:后
    CALL	runtime·abort(SB)//如果线程本地存储不能正常工作，退出程序
    ok:
    // set the per-goroutine and per-mach "registers"
    get_tls(BX)//获取fs段基址到BX寄存器
    LEAQ	runtime·g0(SB), CX//CX=g0的地址
    MOVQ	CX, g(BX)//把g0的地址保存在线程本地存储里面，也就是m0.tls[0]=&g0
    LEAQ	runtime·m0(SB), AX//AX=m0的地址

    // save m->g0 = g0
    //把m0和g0关联起来m0->g0 =g0，g0->m =m0
    MOVQ	CX, m_g0(AX)//m0.g0 = g0
    // save m0 to g0->m
    MOVQ	AX, g_m(CX) //g0.m = m0 
(
    上面的代码首先把g0的地址放入主线程的线程本地存储中，然后通过
    
    1
    2
    m0.g0 = &g0
    g0.m = &m0

    MOVQ	CX, g(BX)//把g0的地址保存在线程本地存储里面，也就是m0.tls[0]=&g0
    把m0和g0绑定在一起，这样，之后在主线程中通过get_tls可以获取到g0，通过g0的m成员又可以找到m0，于是这里就实现了m0和g0与主线程之间的关联。
）
    //准备调用args函数，这四条指令把参数放在栈上
    MOVL	16(SP), AX		// copy argc
    MOVL	AX, 0(SP)  // argc放在栈顶
MOVQ	24(SP), AX		// copy argv
    MOVQ	AX, 8(SP) //argv放在SP + 8的位置
    //处理操作系统传递过来的参数和env
    CALL	runtime·args(SB)
    //对于linx来说，osinit唯一功能就是获取CPU的核数并放在global变量ncpu中
    //执行的结果是全局变量 ncpu = CPU核数
    CALL	runtime·osinit(SB)
(runtime/os_linux.go
func osinit() {
    ncpu = getproccount()
    physHugePageSize = getHugePageSize()
    osArchInit()
}
)
    
	//调度系统初始化
    CALL	runtime·schedinit(SB)
(//runtime/proc.go
    func schedinit() {
        // raceinit must be the first call to race detector.
        // In particular, it must be done before mallocinit below calls racemapshadow.
		//getg函数在源代码中没有对应的定义，由编译器插入类似下面两行代码
		//get_tls(CX) 
		//MOVQ g(CX), BX; BX存器里面现在放的是当前g结构体对象的地址
        _g_ := getg()// _g_ = &g0
        if raceenabled {
        _g_.racectx, raceprocctx0 = raceinit()
        }
        //设置最多启动10000个操作系统线程，也是最多10000个M
        sched.maxmcount = 10000
        
        tracebackinit()
        moduledataverify()
        stackinit()
        mallocinit()
        fastrandinit() // must run before mcommoninit
        //初始化m0，因为从前面的代码我们知道g0->m = &m0
        mcommoninit(_g_.m, -1)
        cpuinit()       // must run before alginit
        alginit()       // maps must not be used before this call
        modulesinit()   // provides activeModules
        typelinksinit() // uses maps, activeModules
        itabsinit()     // uses activeModules
        
        msigsave(_g_.m)
        initSigmask = _g_.m.sigmask
        
        goargs()
        goenvs()
        parsedebugvars()
        gcinit()
        
        sched.lastpoll = uint64(nanotime())
        procs := ncpu//系统中有多少核，就创建和初始化多少个p结构体对象
        if n, ok := atoi32(gogetenv("GOMAXPROCS")); ok && n > 0 {
            procs = n//如果环境变量指定了GOMAXPROCS，则创建指定数量的p
        }
        ////创建和初始化全局变量allp
        if procresize(procs) != nil {
            throw("unknown runnable goroutine during bootstrap")
        }
        
        // For cgocheck > 1, we turn on the write barrier at all times
        // and check all pointer writes. We can't do this until after
        // procresize because the write barrier needs a P.
        if debug.cgocheck > 1 {
        writeBarrier.cgo = true
        writeBarrier.enabled = true
        for _, p := range allp {
        p.wbBuf.reset()
        }
        }
        
        if buildVersion == "" {
        // Condition should never trigger. This code just serves
        // to ensure runtime·buildVersion is kept in the resulting binary.
        buildVersion = "unknown"
        }
        if len(modinfo) == 1 {
        // Condition should never trigger. This code just serves
        // to ensure runtime·modinfo is kept in the resulting binary.
        modinfo = ""
        }
    }
    
    
    //mcommoninit()
	func mcommoninit(mp *m, id int64) {
        _g_ := getg()//_g = g0
        
        // g0 stack won't make sense for user (and is not necessary unwindable).
        if _g_ != _g_.m.g0 {
            callers(1, mp.createstack[:])
        }
        
        lock(&sched.lock)
        
        if id >= 0 {
            mp.id = id
        } else {
            mp.id = mReserveID()//获取一下m.id
            //其中mReserveID() -> checkmcount() //检查已创建系统线程是否超过了数量限制（10000）
        }
        
        mp.fastrand[0] = uint32(int64Hash(uint64(mp.id), fastrandseed))
        mp.fastrand[1] = uint32(int64Hash(uint64(cputicks()), ^fastrandseed))
        if mp.fastrand[0]|mp.fastrand[1] == 0 {
            mp.fastrand[1] = 1
        }

        //创建用于信号处理的gsignal，只是简单的从堆上分配一个g结构体对象,然后把栈设置好就返回了
        mpreinit(mp)
        if mp.gsignal != nil {
            mp.gsignal.stackguard1 = mp.gsignal.stack.lo + _StackGuard
        }
        
        // Add to allm so garbage collector doesn't free g->m
        // when it is just in a register or thread-local storage.
		//把m挂入全局链表allm之中
        mp.alllink = allm
        
        // NumCgoCall() iterates over allm w/o schedlock,
        // so we need to publish it safely.
        atomicstorep(unsafe.Pointer(&allm), unsafe.Pointer(mp))
        unlock(&sched.lock)
        
        // Allocate memory to hold a cgo traceback if the cgo call crashes.
        if iscgo || GOOS == "solaris" || GOOS == "illumos" || GOOS == "windows" {
            mp.cgoCallers = new(cgoCallers)
        }
    }

    //procresize()
    //procresize创建和初始化p结构体对象，在这个函数里面会创建指定个数（根据cpu核数或环境变量确定）的p结构体对象放在全变量allp里, 并把m0和allp[0]绑定在一起
	func procresize(nprocs int32) *p {
        old := gomaxprocs//gomaxprocs初始化时为0
        if old < 0 || nprocs <= 0 {
            throw("procresize: invalid arg")
        }
        if trace.enabled {
            traceGomaxprocs(nprocs)
        }
        
        // update statistics
        now := nanotime()
        if sched.procresizetime != 0 {
            sched.totaltime += int64(old) * (now - sched.procresizetime)
        }
        sched.procresizetime = now
        
        // Grow allp if necessary.
		//初始化时 len(allp) == 0
        if nprocs > int32(len(allp)) {
            // Synchronize with retake, which could be running
            // concurrently since it doesn't run on a P.
            lock(&allpLock)
            if nprocs <= int32(cap(allp)) {
                allp = allp[:nprocs]
            } else {
                //初始化时进入此分支，创建allp 切片
                nallp := make([]*p, nprocs)
                // Copy everything up to allp's cap so we
                // never lose old allocated Ps.
                copy(nallp, allp[:cap(allp)])
                allp = nallp
            }
            unlock(&allpLock)
        }
        
        // initialize new P's
		//循环创建nprocs个p并完成基本初始化
        for i := old; i < nprocs; i++ {
            pp := allp[i]
            if pp == nil {
                pp = new(p)//调用内存分配器从堆上分配一个struct p
            }
            pp.init(i)
            atomicstorep(unsafe.Pointer(&allp[i]), unsafe.Pointer(pp))
        }
        
        _g_ := getg()// _g_ = g0
        if _g_.m.p != 0 && _g_.m.p.ptr().id < nprocs {
            //初始化时m0->p还未初始化，所以不会执行这个分支
            // continue to use the current P
            _g_.m.p.ptr().status = _Prunning
            _g_.m.p.ptr().mcache.prepareForSweep()
        } else {
            // release the current P and acquire allp[0].
            //
            // We must do this before destroying our current P
            // because p.destroy itself has write barriers, so we
            // need to do that from a valid P.
            if _g_.m.p != 0 {
//初始化时这里不执行
                if trace.enabled {
                    // Pretend that we were descheduled
                    // and then scheduled again to keep
                    // the trace sane.
                    traceGoSched()
                    traceProcStop(_g_.m.p.ptr())
                }
                _g_.m.p.ptr().m = 0
            }
            _g_.m.p = 0
            _g_.m.mcache = nil
            p := allp[0]
            p.m = 0
            p.status = _Pidle
            acquirep(p)//把p和m0关联起来，其实是这两个strct的成员相互赋值
            if trace.enabled {
                traceGoStart()
            }
        }
        
        // release resources from unused P's
        //释放过多的p
        for i := nprocs; i < old; i++ {
            p := allp[i]
            p.destroy()//大有文章，不深入
            // can't free P itself because it can be referenced by an M in syscall
        }
        
        // Trim allp.
        if int32(len(allp)) != nprocs {
            lock(&allpLock)
            allp = allp[:nprocs]
            unlock(&allpLock)
        }
        
        //下面这个for 循环把所有空闲的p放入空闲链表
        var runnablePs *p
        for i := nprocs - 1; i >= 0; i-- {
            p := allp[i]
            if _g_.m.p.ptr() == p {//allp[0]跟m0关联了，所以是不能放
                continue
            }
            p.status = _Pidle
            if runqempty(p) {//初始化时除了allp[0]其它p全部执行这个分支，放入空闲链表
                pidleput(p)
            } else {
                p.m.set(mget())
                p.link.set(runnablePs)
                runnablePs = p
            }
        }
        stealOrder.reset(uint32(nprocs))
        var int32p *int32 = &gomaxprocs // make compiler check that gomaxprocs is an int32
        atomic.Store((*uint32)(unsafe.Pointer(int32p)), uint32(nprocs))
        return runnablePs
    }
)

    // create a new goroutine to start program
    MOVQ	$runtime·mainPC(SB), AX		// entry
    PUSHQ	AX
    PUSHQ	$0			// arg size
    CALL	runtime·newproc(SB)
    POPQ	AX
    POPQ	AX
    
    // start this M
    CALL	runtime·mstart(SB)
    
    CALL	runtime·abort(SB)	// mstart should never return
    RET
    
    // Prevent dead-code elimination of debugCallV1, which is
    // intended to be called by debuggers.
    MOVQ	$runtime·debugCallV1(SB), AX
    RET

```