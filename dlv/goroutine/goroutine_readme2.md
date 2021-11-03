
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
        
        if atomic.Load(&sched.npidle) != 0 && atomic.Load(&sched.nmspinning) == 0 && mainStarted {
            wakep()
        }
        releasem(_g_.m)
    }
```