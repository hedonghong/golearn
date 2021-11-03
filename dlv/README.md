# go语言是怎么运行起来的
## 1、准备

```go
Dockerfile

FROM centos
RUN yum install golang -y \
&& yum install dlv -y \
&& yum install binutils -y \
&& yum install vim -y \
&& yum install gdb -y

docker build -t test 建造容器

docker run -it --rm test bash 运行容器
```

## 2、先了解下go build 

```go
不同系统可执行文件编译命令
Mac下编译Linux, Windows平台的64位可执行程序

CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build main.go
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build main.go
Linux下编译Mac, Windows平台的64位可执行程序

CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build main.go
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build main.go
Windows下编译Mac, Linux平台的64位可执行程序

SET CGO_ENABLED=0SET GOOS=darwin3 SET GOARCH=amd64 go build main.go
SET CGO_ENABLED=0 SET GOOS=linux SET GOARCH=amd64 go build  main.go
build 详解
go help build
构建编译由导入路径命名的包，以及它们的依赖关系，但它不会安装结果

使用
go build [-o 输出名] [-i] [编译标记] [包名]
如果参数为***.go文件或文件列表，则编译为一个个单独的包。
当编译单个main包（文件），则生成可执行文件。
当编译单个或多个包非主包时，只构建编译包，但丢弃生成的对象（.a），仅用作检查包可以构建。
当编译包时，会自动忽略’_test.go’的测试文件。
特定参数
名称	描述
-o	output 指定编译输出的名称，代替默认的包名。
-i	install 安装作为目标的依赖关系的包(用于增量编译提速)。
以下 build 参数可用在 build, clean, get, install, list, run, test
名称	描述
-a	完全编译，不理会-i产生的.a文件(文件会比不带-a的编译出来要大)
-n	仅打印输出build需要的命令，不执行build动作（少用）。
-p n	开多少核cpu来并行编译，默认为本机CPU核数（少用）。
-race	同时检测数据竞争状态，只支持 linux/amd64, freebsd/amd64, darwin/amd64 和 windows/amd64.
-msan	启用与内存消毒器的互操作。仅支持linux / amd64，并且只用Clang / LLVM作为主机C编译器（少用）。
-v	打印出被编译的包名（少用）.
-work	打印临时工作目录的名称，并在退出时不删除它（少用）。
-x	同时打印输出执行的命令名（-n）（少用）.
-asmflags ‘flag list’	传递每个go工具asm调用的参数（少用）
-buildmode mode	编译模式（少用）‘go help buildmode’
-compiler name	使用的编译器 == runtime.Compiler (gccgo or gc)（少用）.
-gccgoflags ‘arg list’	gccgo 编译/链接器参数（少用）
-gcflags ‘arg list’	垃圾回收参数（少用）.
-ldflags ‘flag list’	‘-s -w’: 压缩编译后的体积; -s: 去掉符号表; -w:去掉调试信息，不能gdb调试了;
-linkshared	链接到以前使用创建的共享库 -buildmode=shared.
-pkgdir dir	从指定位置，而不是通常的位置安装和加载所有软件包。例如，当使用非标准配置构建时，使用-pkgdir将生成的包保留在单独的位置。
-tags ‘tag list’	构建出带tag的版本.
以上命令，单引号/双引号均可。

对包的操作’go help packages’
对路径的描述’go help gopath’

```

## 3、编译命令简单使用

```go
go build -x xxx.go 观察编译的整个过程

一般是在临时目录生成临时编译文件-过程包括：编译-链接
```

## 4、可执行文件

```go
不同的系统下规范不一，在mac叫mach-O windows叫PE linux叫ELF
ELF由几部分构成
    ELF header 一般程序入口在这里
    Section header
    Sections

操作系统执行可执行文件
解析ELF header -> 加载文件内容至内存 -> 从enttry point开始执行代码

如何通过readelf命令找到执行文件执行入口
readelf -h ./xx       ./xx是编译好的可执行文件

一般会有输出
Entry point address: 0x455780

debug 模式：调试源代码
exec 模式：调试二进制执行文件
attach 模式：调试远程进程 for example：dlv attach 进行id

一般在使用dlv exec之前会使用go build -gcflags=all="-N -l"进行关闭内联优化编译
一般生成环境中不希望二进制文件被调试 go build -ldflags "-s -w"


(dlv) help
The following commands are available:
args -------------------------------- 打印函数参数。
break (alias: b) -------------------- 设置断点，例如：b 包名.函数名，或 b 文件名:行数。
breakpoints (alias: bp) ------------- 显式断点清单。
call -------------------------------- Resumes process, injecting a function call (EXPERIMENTAL!!!)
check (alias: checkpoint) ----------- 在当前位置创建检查点。
checkpoints ------------------------- 打印现有检查点的信息。
clear ------------------------------- 删除断点。
clear-checkpoint (alias: clearcheck)  删除检查点。
clearall ---------------------------- 删除多个断点。
condition (alias: cond) ------------- 设置断点条件。
config ------------------------------ 更改配置参数。
continue (alias: c) ----------------- 运行到断点或程序终止。
disassemble (alias: disass) --------- 反汇编程序。
down -------------------------------- 下移当前帧。
edit (alias: ed) -------------------- Open where you are in $DELVE_EDITOR or $EDITOR
exit (alias: quit | q) -------------- Exit the debugger.
frame ------------------------------- 设置当前帧，或在其他帧上执行命令。
funcs ------------------------------- 打印函数列表。
goroutine --------------------------- 显示或更改当前 goroutine
goroutines -------------------------- List program goroutines.
help (alias: h) --------------------- Prints the help message.
list (alias: ls | l) ---------------- Show source code.
locals ------------------------------ 打印局部变量。
next (alias: n) --------------------- 转到下一行。
on ---------------------------------- 在命中断点时执行命令。
print (alias: p) -------------------- Evaluate an expression.
regs -------------------------------- Print contents of CPU registers.打印寄存器
restart (alias: r) ------------------ Restart process from a checkpoint or event.
rewind (alias: rw) ------------------ 向后运行，直到断点或程序终止。
set --------------------------------- 更改变量的值。
source ------------------------------ Executes a file containing a list of delve commands
sources ----------------------------- 打印源文件列表。
stack (alias: bt) ------------------- 打印堆栈跟踪。
step (alias: s) --------------------- 单步执行程序。
step-instruction (alias: si) -------- 单步单 CPU 指令。
stepout ----------------------------- 跳出当前函数。
thread (alias: tr) ------------------ 切换到指定的线程。
threads ----------------------------- 打印每个跟踪线程的信息。
trace (alias: t) -------------------- 设置跟踪点。
types ------------------------------- 打印类型列表。
up ---------------------------------- 将当前帧上移。
vars -------------------------------- 打印包变量。3
whatis ------------------------------ 打印表达式的类型。

readelf -h ./hello

dlv exec ./hello
b *0x45d900
c 跳到下一个断点
si 进入函数
s 下一步

https://zhuanlan.zhihu.com/p/256970674 gdb
files 
b *0x45d900
run
s


runtime/rt0_linux_amd64.s
runtime/asm_amd64.s-runtime·rt0_go(SB)

runtime/proc.go
runtime/runtime1.go
runtime/runtime2.go

//栈数据的范围 [lo li]
type stack struct {
	//栈顶，低地址
    lo uintptr
    //栈底，高地址
    hi uintptr
}

<-- _StackPreempt


https://segmentfault.com/a/1190000019753885
https://www.cnblogs.com/luozhiyun/p/14844710.html
https://www.cnblogs.com/abozhang/tag/goroutine%E8%B0%83%E5%BA%A6%E5%99%A8/
https://xargin.com/go-and-plan9-asm/

事先知：

寄存器
有4个核心的伪寄存器，这4个寄存器是编译器用来维护上下文、特殊标识等作用的：
FP(Frame pointer): arguments and locals
PC(Program counter): jumps and branches
SB(Static base pointer): global symbols
SP(Stack pointer): top of stack

所有用户空间的数据都可以通过FP/SP(局部数据、输入参数、返回值)和SB(全局数据)访问。 通常情况下，不会对SB/FP寄存器进行运算操作，通常情况以会以SB/FP/SP作为基准地址，进行偏移解引用等操作。

在AMD64环境，伪PC寄存器其实是IP指令计数器寄存器的别名。伪FP寄存器对应的是函数的帧指针，一般用来访问函数的参数和返回值。伪SP栈指针对应的是当前函数栈帧的底部（不包括参数和返回值部分），一般用于定位局部变量。伪SP是一个比较特殊的寄存器，因为还存在一个同名的SP真寄存器。真SP寄存器对应的是栈的顶部，一般用于定位调用其它函数的参数和返回值。

伪寄存器一般需要一个标识符和偏移量为前缀，如果没有标识符前缀则是真寄存器。比如(SP)、+8(SP)没有标识符前缀为真SP寄存器，而a(SP)、b+8(SP)有标识符为前缀表示伪寄存器

MOV 指令有有好几种后缀 MOVB MOVW MOVL MOVQ 分别对应的是 1 字节 、2 字节 、4 字节、8 字节



高地址
Goroutine stack
+-------------------+  <-- _g_.stack.hi
|                   |
+-------------------+
|                   |
+-------------------+
|                   |
+-------------------+  <-- _g_.sched.sp
|                   |
+-------------------+
|                   |
+-------------------+
|                   |
+-------------------+
|                   |
+-------------------+
....
|                   |
+-------------------+  <-- _g_.stackguard0
|                   |   |   |
+-------------------+   |   | _StackSmall
|                   |   |   |
+-------------------+   |  ---
|                   |   |
+-------------------+   |  _StackGuard
|                   |   |
+-------------------+  <-- _g_.stack.lo
低地址

func xxx(a, b, c int) (e, f, g int) {
e, f, g = a, b, c
return
}

高地址
|   返回值g          |
+-------------------+
|   返回值f          |
+-------------------+
|   返回值e          |
+-------------------+
|   参数c            |
+-------------------+
|   参数b            |
+-------------------+
|   参数a            |
+-------------------+ <- 伪FP
|    函数返回地址     |
+-------------------+ <- 伪SP 和 硬件SP
|                   |
+-------------------+
低地址

1 package main
2
3 import "fmt"
4
5 func hello(msg string) {
6         fmt.Println(msg)
7 }
8
9 func main() {
10         go hello("sky")
11 }

-l: 禁止内联

-N: 禁止优化

-S: 输出到标准输出

方法一: go tool compile
使用go tool compile -N -l -S hello.go生成汇编代码//// 禁止优化

方法二: go tool objdump
首先先编译程序: go tool compile -N -l hello.go,
使用go tool objdump once.o反汇编出代码 (或者使用go tool objdump -s Do once.o反汇编特定的函数：)：

方法三: go build -gcflags -S
使用go build -gcflags -S once.go也可以得到汇编代码：

go tool compile 和 go build -gcflags -S 生成的是过程中的汇编，和最终的机器码的汇编可以通过go tool objdump生成。

//runtime/string.go//golang字符串底层实现
type stringStruct struct {
    str unsafe.Pointer
    len int
}
具体的传参过程：

"". 代表的是这个函数的命名空间，SB是个伪寄存器，全名为Static Base，代表对应函数的地址

0x001d 00029 (hello.go:10)	PCDATA	$0, $0
0x001d 00029 (hello.go:10)	PCDATA	$1, $0
0x001d 00029 (hello.go:10)	MOVL	$16, (SP) //newproc 中需要函数参数的大小siz 字符串是16
0x0024 00036 (hello.go:10)	PCDATA	$0, $1
0x0024 00036 (hello.go:10)	LEAQ	"".hello·f(SB), AX
0x002b 00043 (hello.go:10)	PCDATA	$0, $0
0x002b 00043 (hello.go:10)	MOVQ	AX, 8(SP) //把函数存入栈中8(SP)
0x0030 00048 (hello.go:10)	PCDATA	$0, $1
0x0030 00048 (hello.go:10)	LEAQ	go.string."sky"(SB), AX //将字符串"sky"放入寄存器AX
0x0037 00055 (hello.go:10)	PCDATA	$0, $0
0x0037 00055 (hello.go:10)	MOVQ	AX, 16(SP) //将AX中的内容放入栈中16(SP)
0x003c 00060 (hello.go:10)	MOVQ	$3, 24(SP) //将字符串长度3存入栈24(SP)
0x0045 00069 (hello.go:10)	CALL	runtime.newproc(SB)
0x004a 00074 (hello.go:11)	MOVQ	32(SP), BP
0x004f 00079 (hello.go:11)	ADDQ	$40, SP
0x0053 00083 (hello.go:11)	RET

//func newproc(siz int32, fn *funcval) {
func newproc(siz int32, fn *funcval) {
    //获取第一个参数地址
    argp := add(unsafe.Pointer(&fn), sys.PtrSize)
    gp := getg()
    //获取调用者的指令地址，也就是调用newproc时由call指令压栈的函数返回地址
    pc := getcallerpc()
    //systemstack的作用是切换到g0栈执行作为参数的函数
    //用g0系统栈创建goroutine对象
    //传递的参数包括fn函数入口地址，agrp参数起始地址，siz参数长度
    //调用放PC(gorutine)
    systemstack(func() {
    newproc1(fn, argp, siz, gp, pc)
    })
}

type funcval struct {
    fn uintptr
    // variable-size, fn-specific data here
}
它是一个变长结构，第一个字段是一个指针 fn，内存中，紧挨着 fn 的是函数的参数

这个函数需要两个参数，一个是参数大小，一个是方法地址，在汇编代码中分别通过MOVL $16, (SP)和MOVQ AX, 8(SP)实现的，0个参数，AX地址所指向的函数。

我们知道，goroutine 和线程一样，都有自己的栈，不同的是 goroutine 的初始栈比较小，只有 2K，而且是可伸缩的，这也是创建 goroutine 的代价比创建线程代价小的原因。

换句话说，每个 goroutine 都有自己的栈空间，newproc 函数会新创建一个新的 goroutine 来执行 fn 函数，在新 goroutine 上执行指令，就要用新 goroutine 的栈。而执行函数需要参数，这个参数又是在老的 goroutine 上，所以需要将其拷贝到新 goroutine 的栈上。拷贝的起始位置就是栈顶，这好办，那拷贝多少数据呢？由 siz 来确定。

fn 与函数参数
栈顶是 siz，再往上是函数的地址，再往上就是传给 hello 函数的参数，string 在这里是一个地址。因此前面代码里先 push 参数的地址，再 push 参数大小。因此，argp 跳过 fn，向上跳一个指针的长度，拿到 fn 参数的地址。

接着通过 getcallerpc 获取调用者的指令地址，也就是调用 newproc 时由 call 指令压栈的函数返回地址，也就是 runtime·rt0_go 函数里 CALL runtime·newproc(SB) 指令后面的 POPQ AX 这条指令的地址。

最后，调用 systemstack 函数在 g0 栈执行 fn 函数。由于本文讲述的是初始化过程中，由 runtime·rt0_go 函数调用，本身是在 g0 栈执行，因此会直接执行 fn 函数。而如果是我们在程序中写的 go xxx 代码，在执行时，就会先切换到 g0 栈执行，然后再切回来。

栈布局
|                 |       高地址
0x20  +-----------------+
|        BP        |
0x18  +-----------------+
|        3        |
+-----------------+
| &"sky"  |
0x10  +-----------------+ <-- fn + sys.PtrSize
|      hello      |
0x08  +-----------------+ <-- fn
|       siz       |
0x00  +-----------------+ <-- SP
|    newproc PC(调用者返回地址 return address)   |
+-----------------+ callerpc: 要运行的 Goroutine 的 PC
|                 |
|                 |       低地址

若是在asm_amd64.s中TEXT runtime·rt0_go(SB),NOSPLIT,$0
MOVQ	$runtime·mainPC(SB), AX		// entry
PUSHQ	AX
PUSHQ	$0			// arg size
CALL	runtime·newproc(SB)
POPQ	AX
POPQ	AX
一般CALL runtime·newproc(SB) 指令后面的 POPQ AX 这条指令的地址作为return address

//声明一个变量runtime·mainPC+ = runtime·main()
DATA	runtime·mainPC+0(SB)/8,$runtime·main(SB)
GLOBL	runtime·mainPC(SB),RODATA,$8

CALL	runtime·newproc(SB) 后等于在g0中新建了一个main goroutine，并且优先放在runext中等待执行


```

## 5、go 执行
```go
go 四座大山
scheduler 、netpoll 、memory 、Garbage Collector(垃圾回收)

runtime._rt0_amd64_linux -> runtine._rt0_amd64 -> runtime.rt0_go
m0 是go程序启动后建立的第一个线程
rt0_go会做：
1、argc argv 处理
2、全局m0 g0初始化
3、获取cpu核心数
4、初始化内置数据结构
5、开始执行用户的main函数 （scheduler在这里循环调度 g协程）

go协程的生成

go fun(){} 实际上在go里面就是一个协程，被scheduler调度执行
后面会遇到GMP 模型，M真正协程可以执行的地方，P管理G，一个P对应多个G，其中G需要P才可以执行，不过也有例外
P scheduler -> schedInit:procresize  普通的叫schedule loop 不用P就可以运作的叫sysmon loop会做为兜底，一些监控，中断，垃圾回收等工作
G go func() -> newproc
M 按需创建 -> newm -> clone

P中有存放G的地方   优先级
自身指针-runnext   高  无锁 只能存放一个G
本地队列local run queue 中 是一个256数量的数组 无锁
全局队列global run queue 低 有锁 是一个链表

本地创建的新G会较高的优先级，放在runnext，目前golang无法安排G的优先级，只能按照scheduler进行调度

go func() {} 过程是怎么样的？
runtime.newproc
runtime.newproc1
malg&&allgadd 分配内存
runqput  g入队 -> 存入runnext，若满了则把老G踢往local run queue,通过满也会把一办的G踢往global run queue
wakep

go协程的消费
是一个循环代码：
schedule -> runtime.execute -> runtime.gogo -> runtime.goexit -> schedule
每执行61次就会检查一次全局队列，循环的时候p.schedtick会累加

runableGCWorker
globalrunqget(上锁)每执行61次就会检查一次全局队列
runqget
findrunnable(没有g的时候会寻找)
    top部分-有协程
        wakefind
        runqget
        globalrunqget(上锁)如果本地local run queue没有也会从全局队列中获取大约128个
        netpoll->injectglist->globalrunqput上锁
        runqsteal->runqgrab 从其他的p偷一半g回来（一半只是说法，准确看代码）
    stop部分-可能没有协程-需要暂停
        gcMarkWork
        check again all runq
        gcMarkWork
        netpoll
        stopm

重点函数比如gopark  goready  retake
```


## 6、编译过程

词法分析  语法分析  语义分析  中间代码生成  中间代码优化  机器代码生成

## 7、四大数据结构，数组，切片，map，字符串