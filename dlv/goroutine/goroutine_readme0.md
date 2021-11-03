
## 进程、线程、goroutine

    1、进度是系统资源分配的基本单位
    2、线程是CPU调度的基本单位，共享进程的资源
    3、协程一种用户态的轻量级线程，协程的调度完全由用户控制。协程拥有自己的寄存器上下文和栈。协程调度切换时，将寄存器上下文和栈保存到其他地方，在切回来的时候，恢复先前保存的寄存器上下文和栈，直接操作栈则基本没有内核切换的开销，可以不加锁的访问全局变量，所以上下文的切换非常快。 协程在子程序内部可中断的，然后转而执行别的子程序，在适当的时候再返回来接着执行。

    基于系统线程进程切换太重，内存使用太重，创造出goroutine。默认大小栈为2k。
    
    内核对系统线程的调度简单的归纳为：在执行操作系统代码时，内核调度器按照一定的算法挑选出一个线程并把该线程保存在内存之中的寄存器的值放入CPU对应的寄存器从而恢复该线程的运行。
    
    万变不离其宗，系统线程对goroutine的调度与内核对系统线程的调度原理是一样的，实质都是通过保存和修改CPU寄存器的值来达到切换线程/goroutine的目的。

    因此，为了实现对goroutine的调度，需要引入一个数据结构来保存CPU寄存器的值以及goroutine的其它一些状态信息，在Go语言调度器源代码中，这个数据结构是一个名叫g的结构体，它保存了goroutine的所有信息，该结构体的每一个实例对象都代表了一个goroutine，调度器代码可以通过g对象来对goroutine进行调度，当goroutine被调离CPU时，调度器代码负责把CPU寄存器的值保存在g对象的成员变量之中，当goroutine被调度起来运行时，调度器代码又负责把g对象的成员变量所保存的寄存器的值恢复到CPU的寄存器。

    要实现对goroutine的调度，仅仅有g结构体对象是不够的，至少还需要一个存放所有（可运行）goroutine的容器，便于工作线程寻找需要被调度起来运行的goroutine，于是Go调度器又引入了schedt结构体，一方面用来保存调度器自身的状态信息，另一方面它还拥有一个用来保存goroutine的运行队列。因为每个Go程序只有一个调度器，所以在每个Go程序中schedt结构体只有一个实例对象，该实例对象在源代码中被定义成了一个共享的全局变量，这样每个工作线程都可以访问它以及它所拥有的goroutine运行队列，我们称这个运行队列为全局运行队列。

    既然说到全局运行队列，读者可能猜想到应该还有一个局部运行队列。确实如此，因为全局运行队列是每个工作线程都可以读写的，因此访问它需要加锁，然而在一个繁忙的系统中，加锁会导致严重的性能问题。于是，调度器又为每个工作线程引入了一个私有的局部goroutine运行队列，工作线程优先使用自己的局部运行队列，只有必要时才会去访问全局运行队列，这大大减少了锁冲突，提高了工作线程的并发性。在Go调度器源代码中，局部运行队列被包含在p结构体的实例对象之中，每一个运行着go代码的工作线程都会与一个p结构体的实例对象关联在一起。

    除了上面介绍的g、schedt和p结构体，Go调度器源代码中还有一个用来代表工作线程的m结构体，每个工作线程都有唯一的一个m结构体的实例对象与之对应，m结构体对象除了记录着工作线程的诸如栈的起止位置、当前正在执行的goroutine以及是否空闲等等状态信息之外，还通过指针维持着与p结构体的实例对象之间的绑定关系。于是，通过m既可以找到与之对应的工作线程正在运行的goroutine，又可以找到工作线程的局部运行队列等资源。

G() ->  M -> P->(G,G,G,G,G，有本地256个队列和reqnext下一个g)  schedt(G,G,G,G)全局


## 具体数据结构体

所在位置：runtime/runtime2.go

```go

//用来记录goroutine使用栈的起始和结束位置
type stack struct {
    lo uintptr 栈顶，指向内存低地址，结束
    hi uintptr 栈底，指向内存高地址 开始
}

type gobuf struct {
    // The offsets of sp, pc, and g are known to (hard-coded in) libmach.
    //
    // ctxt is unusual with respect to GC: it may be a
    // heap-allocated funcval, so GC needs to track it, but it
    // needs to be set and cleared from assembly, where it's
    // difficult to have write barriers. However, ctxt is really a
    // saved, live register, and we only ever exchange it between
    // the real register and the gobuf. Hence, we treat it as a
    // root during stack scanning, which means assembly that saves
    // and restores it doesn't need write barriers. It's still
    // typed as a pointer so that any other writes from Go get
    // write barriers.
    sp   uintptr 保存cpu的rsp的寄存器的值 栈顶位置
    pc   uintptr 保存cpu的rip的寄存器的值 下一条指令地址
    g    guintptr 记录当前gobuf属于哪个goroutine
    ctxt unsafe.Pointer
    ret  sys.Uintreg 保存系统调用的返回值，因为从系统调用返回之后如果p被其他goroutine占用，那就保存在这里，这个goroutine放入到全局队列等待被执行，进而取得其返回值
    lr   uintptr
    bp   uintptr // 保存cpu的rip寄存器的值 for GOEXPERIMENT=framepointer
}

//g结构体，代表一个goroutine
type g struct {
    // Stack parameters.
    // stack describes the actual stack memory: [stack.lo, stack.hi).
    // stackguard0 is the stack pointer compared in the Go stack growth prologue.
    // It is stack.lo+StackGuard normally, but can be StackPreempt to trigger a preemption.
    // stackguard1 is the stack pointer compared in the C stack growth prologue.
    // It is stack.lo+StackGuard on g0 and gsignal stacks.
    // It is ~0 on other goroutine stacks, to trigger a call to morestackc (and crash).
	//记录该goroutine的使用的栈
    stack       stack   // offset known to runtime/cgo
    //用于栈溢出检查，实现栈的自由扩展和收缩，抢占调度用到stackguard0
    stackguard0 uintptr // offset known to liblink
    stackguard1 uintptr // offset known to liblink
    
    _panic       *_panic // innermost panic - offset known to liblink
    _defer       *_defer // innermost defer
    //此goroutine正在哪个m中执行
    m            *m      // current m; offset known to arm liblink
    //保存调度信息，主要是几个寄存器的值
    sched        gobuf
    syscallsp    uintptr        // if status==Gsyscall, syscallsp = sched.sp to use during gc
    syscallpc    uintptr        // if status==Gsyscall, syscallpc = sched.pc to use during gc
    stktopsp     uintptr        // expected sp at top of stack, to check in traceback
    param        unsafe.Pointer // passed parameter on wakeup
    atomicstatus uint32
    stackLock    uint32 // sigprof/scang lock; TODO: fold in to atomicstatus
    goid         int64
    //指向全局g队列中的下一个g
    //所有全局g队列的g形成一个链表
    schedlink    guintptr
    waitsince    int64      // approx time when the g become blocked
    waitreason   waitReason // if status==Gwaiting
    //抢占调度的标志，如果需要抢占调度，设置preempt为true
    preempt       bool // preemption signal, duplicates stackguard0 = stackpreempt
    preemptStop   bool // transition to _Gpreempted on preemption; otherwise, just deschedule
    preemptShrink bool // shrink stack at synchronous safe point
    
    // asyncSafePoint is set if g is stopped at an asynchronous
    // safe point. This means there are frames on the stack
    // without precise pointer information.
    asyncSafePoint bool
    
    paniconfault bool // panic (instead of crash) on unexpected fault address
    gcscandone   bool // g has scanned stack; protected by _Gscan bit in status
    throwsplit   bool // must not split stack
    // activeStackChans indicates that there are unlocked channels
    // pointing into this goroutine's stack. If true, stack
    // copying needs to acquire channel locks to protect these
    // areas of the stack.
    activeStackChans bool
    // parkingOnChan indicates that the goroutine is about to
    // park on a chansend or chanrecv. Used to signal an unsafe point
    // for stack shrinking. It's a boolean value, but is updated atomically.
    parkingOnChan uint8
    
    raceignore     int8     // ignore race detection events
    sysblocktraced bool     // StartTrace has emitted EvGoInSyscall about this goroutine
    sysexitticks   int64    // cputicks when syscall has returned (for tracing)
    traceseq       uint64   // trace event sequencer
    tracelastp     puintptr // last P emitted an event for this goroutine
    lockedm        muintptr
    sig            uint32
    writebuf       []byte
    sigcode0       uintptr
    sigcode1       uintptr
    sigpc          uintptr
    gopc           uintptr         // pc of go statement that created this goroutine
    ancestors      *[]ancestorInfo // ancestor information goroutine(s) that created this goroutine (only used if debug.tracebackancestors)
    startpc        uintptr         // pc of goroutine function
    racectx        uintptr
    waiting        *sudog         // sudog structures this g is waiting on (that have a valid elem ptr); in lock order
    cgoCtxt        []uintptr      // cgo traceback context
    labels         unsafe.Pointer // profiler labels
    timer          *timer         // cached timer for time.Sleep
    selectDone     uint32         // are we participating in a select and did someone win the race?
    
    // Per-G GC state
    
    // gcAssistBytes is this G's GC assist credit in terms of
    // bytes allocated. If this is positive, then the G has credit
    // to allocate gcAssistBytes bytes without assisting. If this
    // is negative, then the G must correct this by performing
    // scan work. We track this in bytes to make it fast to update
    // and check for debt in the malloc hot path. The assist ratio
    // determines how this corresponds to scan work debt.
    gcAssistBytes int64
}


0x0000 MOVQ	(TLS), CX   ;; store current *g in CX
0x0009 CMPQ	SP, 16(CX)  ;; compare SP and g.stackguard0
0x000d JLS	58	    ;; jumps to 0x3a if SP <= g.stackguard0
TLS 是一个由 runtime 维护的虚拟寄存器，保存了指向当前 g 的指针，这个 g 的数据结构会跟踪 goroutine 运行时的所有状态值。
我们可以看到 16(CX) 对应的是 g.stackguard0，是 runtime 维护的一个阈值，该值会被拿来与栈指针(stack-pointer)进行比较以判断一个 goroutine 是否马上要用完当前的栈空间。

因此 prologue 只要检查当前的 SP 的值是否小于或等于 stackguard0 的阈值就行了，如果是的话，就跳到 epilogue 部分去。

//m结构体代表工作线程，保存了m自身使用的栈信息，当前正在运行的goroutine，与m绑定的p等信息
type m struct {
	//记录调度代码时需要使用这个栈，执行用户goroutine时进行栈切换
    g0      *g     // goroutine with scheduling stack
    morebuf gobuf  // gobuf arg to morestack
    divmod  uint32 // div/mod denominator for arm - known to liblink
    
    // Fields not known to debuggers.
    procid        uint64       // for debuggers, but offset not hard-coded
    gsignal       *g           // signal-handling g
    goSigStack    gsignalStack // Go-allocated signal handling stack
    sigmask       sigset       // storage for saved signal mask
    //通过TLS（线程本地存储）实现m结构对象与工作线程之间的绑定
    tls           [6]uintptr   // thread-local storage (for x86 extern register)
    mstartfn      func()
    //指向正在运行的goroutine的g结构体对象
    curg          *g       // current running goroutine
    caughtsig     guintptr // goroutine running during fatal signal
    //记录当前与m绑定的p结构体对象
    p             puintptr // attached p for executing go code (nil if not executing go code)
    nextp         puintptr
    oldp          puintptr // the p that was attached before executing a syscall
    id            int64
    mallocing     int32
    throwing      int32
    preemptoff    string // if != "", keep curg running on this m
    locks         int32
    dying         int32
    profilehz     int32
    //spinning状态，表示当前工作线程正在试图从其他工作线程的本地队列中偷取g
    spinning      bool // m is out of work and is actively looking for work
    blocked       bool // m is blocked on a note
    newSigstack   bool // minit on C thread called sigaltstack
    printlock     int8
    incgo         bool   // m is executing a cgo call
    freeWait      uint32 // if == 0, safe to free g0 and delete m (atomic)
    fastrand      [2]uint32
    needextram    bool
    traceback     uint8
    ncgocall      uint64      // number of cgo calls in total
    ncgo          int32       // number of cgo calls currently in progress
    cgoCallersUse uint32      // if non-zero, cgoCallers in use temporarily
    cgoCallers    *cgoCallers // cgo traceback if crashing in cgo call
    //没有g需要运行时，工作线程睡眠在这个park成员上
    //其他线程通过这个park唤醒该工作线程
    park          note
    //记录所有工作线程的一个链表
    alllink       *m // on allm
    schedlink     muintptr
    //线程内存缓存，分配小内存从这个分配与内存管理有关
    mcache        *mcache
    lockedg       guintptr
    createstack   [32]uintptr // stack that created this thread.
    lockedExt     uint32      // tracking for external LockOSThread
    lockedInt     uint32      // tracking for internal lockOSThread
    nextwaitm     muintptr    // next m waiting for lock
    waitunlockf   func(*g, unsafe.Pointer) bool
    waitlock      unsafe.Pointer
    waittraceev   byte
    waittraceskip int
    startingtrace bool
    syscalltick   uint32
    freelink      *m // on sched.freem
    
    // these are here because they are too large to be on the stack
    // of low-level NOSPLIT functions.
    libcall   libcall
    libcallpc uintptr // for cpu profiler
    libcallsp uintptr
    libcallg  guintptr
    syscall   libcall // stores syscall parameters on windows
    
    vdsoSP uintptr // SP for traceback while in VDSO call (0 if not in call)
    vdsoPC uintptr // PC for traceback while in VDSO call
    
    // preemptGen counts the number of completed preemption
    // signals. This is used to detect when a preemption is
    // requested, but fails. Accessed atomically.
    preemptGen uint32
    
    // Whether this is a pending preemption signal on this M.
    // Accessed atomically.
    signalPending uint32
    
    dlogPerM
    
    mOS
}

//p结构体，用于保存工作线程执行go代码需要的必要资源，比如goroutine的运行队列，内存分配用到的缓存等等
type p struct {
    id          int32
    //p状态
    status      uint32 // one of pidle/prunning/...
    link        puintptr
    schedtick   uint32     // incremented on every scheduler call
    syscalltick uint32     // incremented on every system call
    sysmontick  sysmontick // last tick observed by sysmon
    m           muintptr   // back-link to associated m (nil if idle)
    mcache      *mcache
    pcache      pageCache
    raceprocctx uintptr
    
    deferpool    [5][]*_defer // pool of available defer structs of different sizes (see panic.go)
    deferpoolbuf [5][32]*_defer
    
    // Cache of goroutine ids, amortizes accesses to runtime·sched.goidgen.
    goidcache    uint64
    goidcacheend uint64
    
    // Queue of runnable goroutines. Accessed without lock.
    //本地g队列 队列头
    runqhead uint32
	//本地g队列 队列尾
    runqtail uint32
    //本地g队列
    runq     [256]guintptr
    // runnext, if non-nil, is a runnable G that was ready'd by
    // the current G and should be run next instead of what's in
    // runq if there's time remaining in the running G's time
    // slice. It will inherit the time left in the current time
    // slice. If a set of goroutines is locked in a
    // communicate-and-wait pattern, this schedules that set as a
    // unit and eliminates the (potentially large) scheduling
    // latency that otherwise arises from adding the ready'd
    // goroutines to the end of the run queue.
    //p中最新创建的g，执行也是首先从这个拿
    runnext guintptr
    
    // Available G's (status == Gdead)
    gFree struct {
    gList
    n int32
    }
    
    sudogcache []*sudog
    sudogbuf   [128]*sudog
    
    // Cache of mspan objects from the heap.
    mspancache struct {
    // We need an explicit length here because this field is used
    // in allocation codepaths where write barriers are not allowed,
    // and eliminating the write barrier/keeping it eliminated from
    // slice updates is tricky, moreso than just managing the length
    // ourselves.
    len int
    buf [128]*mspan
    }
    
    tracebuf traceBufPtr
    
    // traceSweep indicates the sweep events should be traced.
    // This is used to defer the sweep start event until a span
    // has actually been swept.
    traceSweep bool
    // traceSwept and traceReclaimed track the number of bytes
    // swept and reclaimed by sweeping in the current sweep loop.
    traceSwept, traceReclaimed uintptr
    
    palloc persistentAlloc // per-P to avoid mutex
    
    _ uint32 // Alignment for atomic fields below
    
    // The when field of the first entry on the timer heap.
    // This is updated using atomic functions.
    // This is 0 if the timer heap is empty.
    timer0When uint64
    
    // Per-P GC state
    gcAssistTime         int64    // Nanoseconds in assistAlloc
    gcFractionalMarkTime int64    // Nanoseconds in fractional mark worker (atomic)
    gcBgMarkWorker       guintptr // (atomic)
    gcMarkWorkerMode     gcMarkWorkerMode
    
    // gcMarkWorkerStartTime is the nanotime() at which this mark
    // worker started.
    gcMarkWorkerStartTime int64
    
    // gcw is this P's GC work buffer cache. The work buffer is
    // filled by write barriers, drained by mutator assists, and
    // disposed on certain GC state transitions.
    gcw gcWork
    
    // wbBuf is this P's GC write barrier buffer.
    //
    // TODO: Consider caching this in the running G.
    wbBuf wbBuf
    
    runSafePointFn uint32 // if 1, run sched.safePointFn at next safe point
    
    // Lock for timers. We normally access the timers while running
    // on this P, but the scheduler can also do it from a different P.
    timersLock mutex
    
    // Actions to take at some time. This is used to implement the
    // standard library's time package.
    // Must hold timersLock to access.
    timers []*timer
    
    // Number of timers in P's heap.
    // Modified using atomic instructions.
    numTimers uint32
    
    // Number of timerModifiedEarlier timers on P's heap.
    // This should only be modified while holding timersLock,
    // or while the timer status is in a transient state
    // such as timerModifying.
    adjustTimers uint32
    
    // Number of timerDeleted timers in P's heap.
    // Modified using atomic instructions.
    deletedTimers uint32
    
    // Race context used while executing timer functions.
    timerRaceCtx uintptr
    
    // preempt is set to indicate that this P should be enter the
    // scheduler ASAP (regardless of what G is running on it).
    preempt bool
    
    pad cpu.CacheLinePad
}

//调度器的状态信息和全局g队列等
type schedt struct {
    // accessed atomically. keep at top to ensure alignment on 32-bit systems.
    goidgen   uint64
    lastpoll  uint64 // time of last network poll, 0 if currently polling
    pollUntil uint64 // time to which current poll is sleeping
    
    lock mutex
    
    // When increasing nmidle, nmidlelocked, nmsys, or nmfreed, be
    // sure to call checkdead().
    //空闲m链表
    midle        muintptr // idle m's waiting for work
    //空闲m的数量
    nmidle       int32    // number of idle m's waiting for work
    nmidlelocked int32    // number of locked m's waiting for work
    mnext        int64    // number of m's that have been created and next M ID
    maxmcount    int32    // maximum number of m's allowed (or die)
    nmsys        int32    // number of system m's not counted for deadlock
    nmfreed      int64    // cumulative number of freed m's
    
    ngsys uint32 // number of system goroutines; updated atomically
    
    //空闲的p链表
    pidle      puintptr // idle p's
	//空闲的p链表数量
    npidle     uint32
    nmspinning uint32 // See "Worker thread parking/unparking" comment in proc.go.
    
    // Global runnable queue.
    //全局g链表
    runq     gQueue
    //全局g链表的g数量
    runqsize int32
    
    // disable controls selective disabling of the scheduler.
    //
    // Use schedEnableUser to control this.
    //
    // disable is protected by sched.lock.
    disable struct {
    // user disables scheduling of user goroutines.
    user     bool
    runnable gQueue // pending runnable Gs
    n        int32  // length of runnable
    }
    
    // Global cache of dead G's.
    //是所有已经退出的g对应g结构体对象组成的链表
    //用于缓存g结构体对象，避免每次创建goroutine时都从新分配内存
    gFree struct {
    lock    mutex
    stack   gList // Gs with stacks
    noStack gList // Gs without stacks
    n       int32
    }
    
    // Central cache of sudog structs.
    sudoglock  mutex
    sudogcache *sudog
    
    // Central pool of available defer structs of different sizes.
    deferlock mutex
    deferpool [5]*_defer
    
    // freem is the list of m's waiting to be freed when their
    // m.exited is set. Linked through m.freelink.
    freem *m
    
    gcwaiting  uint32 // gc is waiting to run
    stopwait   int32
    stopnote   note
    sysmonwait uint32
    sysmonnote note
    
    // safepointFn should be called on each P at the next GC
    // safepoint if p.runSafePointFn is set.
    safePointFn   func(*p)
    safePointWait int32
    safePointNote note
    
    profilehz int32 // cpu profiling rate
    
    procresizetime int64 // nanotime() of last change to gomaxprocs
    totaltime      int64 // ∫gomaxprocs dt up to procresizetime
}

//一些重要的全局变量
var (
	//保存所有的g数量
    allglen    uintptr
    //所有的m构成的一个链表，包括下面的m0
    allm       *m
    //保存所有的p，len(allp) == gomaxprocs
    allp       []*p  // len(allp) == gomaxprocs; may change at safe points, otherwise immutable
    allpLock   mutex // Protects P-less reads of allp and all writes
    //p的最大值，默认等于ncpu，但可以通过GOMAXPROCS修改
    gomaxprocs int32
    //系统中cpu核的数量，程序启动时由runtime代码初始化
    ncpu       int32
    forcegc    forcegcstate
    sched      schedt
    newprocs   int32
    
    // Information about what cpu features are available.
    // Packages outside the runtime should not use these
    // as they are not an external api.
    // Set on startup in asm_{386,amd64}.s
    processorVersionInfo uint32
    isIntel              bool
    lfenceBeforeRdtsc    bool
    
    goarm                uint8 // set by cmd/link on arm systems
    framepointer_enabled bool  // set by cmd/link
)
//runtime/proc.go
var (
    // 代表进程的主线程
    m0           m
    // m0的g0，也就是m0.g0 = &g0
    g0           g
    raceprocctx0 uintptr
)

```



