##非main goroutine 如何调度运行的？
    main goroutine退出时会直接执行exit系统调用退出整个进程，而非main goroutine退出时则会进入goexit函数完成最后的清理工作

```go
//test.go
package main

import "fmt"

var ch = make(chan int)

func g1(x int)  {
    fmt.Println(x)
    ch <- x
}

func main()  {
    go g1(1)
    fmt.Println(<-ch)
}

先根据之前的流程梳理下：
1、main goroutine启动后main函数中创建一个goroutine g1()函数

怎么验证非main goroutine是走goexit退出呢，使用dlv或者gdb
下面使用gdb：

>b main.g1 //在main.g1处打断点
>r //可以看到调用者是goexit
>bt //查看函数调用链，看起来g2真的是被runtime.goexit调用的
>disass //反汇编找ret的地址，这是为了在ret处下断点
>b *0x0000000000486a0c //在retq指令位置下断点
>c //跟踪进去
>disass //程序停在了ret指令处
>si //单步执行一条指令
>disass  //可以看出来g2已经返回到了goexit函数中

//经过上面，我们看下普通goroutine怎么退出，并进行循环调度的
// The top-most function running on a goroutine
// returns to goexit+PCQuantum.
TEXT runtime·goexit(SB),NOSPLIT,$0-0
BYTE	$0x90	// NOP
CALL	runtime·goexit1(SB)	// does not return
// traceback from goexit1 must hit code range of goexit
BYTE	$0x90	// NOP

//根据代码查找到runtime·goexit1()
//runtime/proc.go
// Finishes execution of the current goroutine.
func goexit1() {
	if raceenabled {
		racegoend()//与竞态检查有关，不关注
	}
	if trace.enabled {
		traceGoEnd()//与backtrace有关，不关注
	}
	mcall(goexit0)
}

//runtime/asm_amd64.s
// func mcall(fn func(*g))
// Switch to m->g0's stack, call fn(g).
// Fn must never return. It should gogo(&g->sched)
// to keep running g.
//参数是一个指向funcval对象的指针
//根据注释大概知道逻辑是切换到g0并且调用fn函数
TEXT runtime·mcall(SB), NOSPLIT, $0-8
//取出参数的值放入DI寄存器，它是funcval对象的指针，此场景中fn.fn是goexit0的地址
MOVQ	fn+0(FP), DI

get_tls(CX)
//AX = g，本场景g 是 g1
MOVQ	g(CX), AX	// save state in g->sched
//mcall返回地址放入BX
MOVQ	0(SP), BX	// caller's PC
//保存g1的调度信息，因为我们要从当前正在运行的g1切换到g0
MOVQ	BX, (g_sched+gobuf_pc)(AX)
LEAQ	fn+0(FP), BX	// caller's SP
MOVQ	BX, (g_sched+gobuf_sp)(AX)
MOVQ	AX, (g_sched+gobuf_g)(AX)
MOVQ	BP, (g_sched+gobuf_bp)(AX)

// switch to m->g0 & its stack, call fn
//下面三条指令主要目的是找到g0的指针
MOVQ	g(CX), BX
MOVQ	g_m(BX), BX
MOVQ	m_g0(BX), SI
//此刻，SI = g0， AX = g，所以这里在判断g 是否是 g0，如果g == g0则一定是哪里代码写错了
CMPQ	SI, AX	// if g == m->g0 call badmcall
JNE	3(PC)//如果没错，跳过3行指令
MOVQ	$runtime·badmcall(SB), AX
JMP	AX
MOVQ	SI, g(CX)	// g = m->g0 把g0的地址设置到线程本地存储之中
//恢复g0的栈顶指针到CPU的rsp寄存器，这一条指令完成了栈的切换，从g的栈切换到了g0的栈
MOVQ	(g_sched+gobuf_sp)(SI), SP	// sp = m->g0->sched.sp
//AX = g
PUSHQ	AX//fn的参数g入栈 
MOVQ	DI, DX//DI是结构体funcval实例对象的指针，它的第一个成员才是goexit0的地址
MOVQ	0(DI), DI//读取第一个成员到DI寄存器
CALL	DI//调用goexit0(g)
POPQ	AX
MOVQ	$runtime·badmcall2(SB), AX
JMP	AX
RET

    mcall的参数是一个函数，在Go语言的实现中，函数变量并不是一个直接指向函数代码的指针，而是一个指向funcval结构体对象的指针，funcval结构体对象的第一个成员fn才是真正指向函数代码的指针。
    
    mcall函数主要有两个功能：
    
    首先从当前运行的g(我们这个场景是g1)切换到g0，这一步包括保存当前g的调度信息，把g0设置到tls中，修改CPU的rsp寄存器使其指向g0的栈；
    
    以当前运行的g(我们这个场景是g1)为参数调用fn函数(此处为goexit0)。

    mcall做的事情跟gogo函数完全相反，gogo函数实现了从g0切换到某个goroutine去运行，而mcall实现了从某个goroutine切换到g0来运行，因此，mcall和gogo的代码非常相似，然而mcall和gogo在做切换时有个重要的区别：gogo函数在从g0切换到其它goroutine时首先切换了栈，然后通过跳转指令从runtime代码切换到了用户goroutine的代码，而mcall函数在从其它goroutine切换回g0时只切换了栈，并未使用跳转指令跳转到runtime代码去执行。为什么会有这个差别呢？原因在于在从g0切换到其它goroutine之前执行的是runtime的代码而且使用的是g0栈，所以切换时需要首先切换栈然后再从runtime代码跳转某个goroutine的代码去执行（切换栈和跳转指令不能颠倒，因为跳转之后执行的就是用户的goroutine代码了，没有机会切换栈了），然而从某个goroutine切换回g0时，goroutine使用的是call指令来调用mcall函数，mcall函数本身就是runtime的代码，所以call指令其实已经完成了从goroutine代码到runtime代码的跳转，因此mcall函数自身的代码就不需要再跳转了，只需要把栈切换到g0栈即可。


// goexit continuation on g0.
// gp 是g1() goroutine
func goexit0(gp *g) {
	_g_ := getg() // _g_ = g0

	//g马上退出，所以设置其状态为_Gdead
	//状态切换重要细节点
	//目前 gp 是g1() goroutine
	casgstatus(gp, _Grunning, _Gdead)
	if isSystemGoroutine(gp, false) {
		atomic.Xadd(&sched.ngsys, -1)
	}
	//清空g1 goroutine保存的一些信息
	gp.m = nil
	locked := gp.lockedm != 0
	gp.lockedm = 0
	_g_.m.lockedg = 0
	gp.preemptStop = false
	gp.paniconfault = false
	gp._defer = nil // should be true already but just in case.
	gp._panic = nil // non-nil for Goexit during panic. points at stack-allocated data.
	gp.writebuf = nil
	gp.waitreason = 0
	gp.param = nil
	gp.labels = nil
	gp.timer = nil

	if gcBlackenEnabled != 0 && gp.gcAssistBytes > 0 {
		// Flush assist credit to the global pool. This gives
		// better information to pacing if the application is
		// rapidly creating an exiting goroutines.
		scanCredit := int64(gcController.assistWorkPerByte * float64(gp.gcAssistBytes))
		atomic.Xaddint64(&gcController.bgScanCredit, scanCredit)
		gp.gcAssistBytes = 0
	}

	//g->m = nil, m->currg = nil 解绑g和m之关系
	dropg()

	if GOARCH == "wasm" { // no threads yet on wasm
		gfput(_g_.m.p.ptr(), gp)//g放回p的gFree池子中，防止反复创建g重复利用
		schedule() // never returns 继续循环调度
	}

	if _g_.m.lockedInt != 0 {
		print("invalid m->lockedInt = ", _g_.m.lockedInt, "\n")
		throw("internal lockOSThread error")
	}
	//g放回p的gFree池子中，防止反复创建g重复利用
	gfput(_g_.m.p.ptr(), gp)
	if locked {
		// The goroutine may have locked this thread because
		// it put it in an unusual kernel state. Kill it
		// rather than returning it to the thread pool.

		// Return to mstart, which will release the P and exit
		// the thread.
		if GOOS != "plan9" { // See golang.org/issue/22227.
			gogo(&_g_.m.g0.sched)
		} else {
			// Clear lockedExt on plan9 since we may end up re-using
			// this thread.
			_g_.m.lockedExt = 0
		}
	}
	//继续循环调度
	schedule()
}

    从g1栈切换到g0栈之后，下面开始在g0栈执行goexit0函数，该函数完成最后的清理工作：

    把g的状态从_Grunning变更为_Gdead；
    
    然后把g的一些字段清空成0值；

    调用dropg函数解除g和m之间的关系，其实就是设置g->m = nil, m->currg = nil；
    
    把g放入p的freeg队列缓存起来供下次创建g时快速获取而不用从内存分配。freeg就是g的一个对象池；
    
    调用schedule函数再次进行调度；
    
    
    调度循环
    
    我们说过，任何goroutine被调度起来运行都是通过schedule()->execute()->gogo()这个函数调用链完成的，而且这个调用链中的函数一直没有返回。以我们刚刚讨论过的g2 goroutine为例，从g2开始被调度起来运行到退出是沿着下面这条路径进行的
    
    schedule()->execute()->gogo()->g2()->goexit()->goexit1()->mcall()->goexit0()->schedule()
    可以看出，一轮调度是从调用schedule函数开始的，然后经过一系列代码的执行到最后又再次通过调用schedule函数来进行新一轮的调度，从一轮调度到新一轮调度的这一过程我们称之为一个调度循环，这里说的调度循环是指某一个工作线程的调度循环，而同一个Go程序中可能存在多个工作线程，每个工作线程都有自己的调度循环，也就是说每个工作线程都在进行着自己的调度循环。
    
    从前面的代码分析可以得知，上面调度循环中的每一个函数调用都没有返回，虽然g2()->goexit()->goexit1()->mcall()这几个函数是在g2的栈空间执行的，但剩下的函数都是在g0的栈空间执行的，那么问题就来了，在一个复杂的程序中，调度可能会进行无数次循环，也就是说会进行无数次没有返回的函数调用，大家都知道，每调用一次函数都会消耗一定的栈空间，而如果一直这样无返回的调用下去无论g0有多少栈空间终究是会耗尽的，那么这里是不是有问题？其实没有问题，关键点就在于，每次执行mcall切换到g0栈时都是切换到g0.sched.sp所指的固定位置，这之所以行得通，正是因为从schedule函数开始之后的一系列函数永远都不会返回，所以重用这些函数上一轮调度时所使用过的栈内存是没有问题的。

    ppt7图

我们用上图来总结一下工作线程的执行流程：

初始化，调用mstart函数；

调用mstart1函数，在该函数中调用save函数设置g0.sched.sp和g0.sched.pc等调度信息，其中g0.sched.sp指向mstart函数栈帧的栈顶；

依次调用schedule->execute->gogo函数执行调度；

运行用户的goroutine代码；

用户goroutine代码执行过程中调用runtime中的某些函数，然后这些函数调用mcall切换到g0.sched.sp所指的栈并最终再次调用schedule函数进入新一轮调度，之后工作线程一直循环执行着3～5这一调度循环直到进程退出为止。
```