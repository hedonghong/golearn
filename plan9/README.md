
## plan9基础

    栈的加减调整
    我们栈是从高地址向低地址增涨的，所以减法可以理解为增加栈空间，加法为收缩栈空间
    SUBQ $0x18, SP // 对 SP 做减法，为函数分配函数栈帧

    ADDQ $0x18, SP // 对 SP 做加法，清除函数栈帧

    在内存和CPU的交互中存在一些数据的搬运
    常数在 plan9 汇编用 $num 表示，可以为负数，默认情况下为十进制。可以用 $0x123 的形式来表示十六进制数。
    
    MOVB $1, DI      // 1 byte 把1放到DI寄存器
    MOVW $0x10, BX   // 2 bytes
    MOVD $1, DX      // 4 bytes
    MOVQ $-10, AX     // 8 bytes

    B、W、D、Q是指搬多长的空间的值

    运算指令

    ADDQ  AX, BX   // BX += AX
    SUBQ  AX, BX   // BX -= AX
    IMULQ AX, BX   // BX *= AX
    类似数据搬运指令，同样可以通过修改指令的后缀来对应不同长度的操作数。例如 ADDQ/ADDW/ADDL/ADDB。
    
    跳转

    // 无条件跳转
    JMP addr   // 跳转到地址，地址可为代码中的地址，不过实际上手写不会出现这种东西
    JMP label  // 跳转到标签，可以跳转到同一函数内的标签位置
    JMP 2(PC)  // 以当前指令为基础，向前跳2行，向前/后跳转 x 行
    JMP -2(PC) // 同上，向后跳2行
    
    // 有条件跳转
    JZ target // 如果 zero flag 被 set 过，则跳转
    
    一般组合可能还有通过CMP比较指令结合JMP进行跳转，golang中g的栈的伸缩就是通过这两个指令完成
    CMP指令也可以接长度，比如CMPQ

    伪寄存器

    1、FP: 使用形如 symbol+offset(FP) 的方式，引用函数的输入参数。例如 arg0+0(FP)，arg1+8(FP)，使用 FP 不加 symbol 时，无法通过编译，在汇编层面来讲，symbol 并没有什么用，加 symbol 主要是为了提升代码可读性。另外，官方文档虽然将伪寄存器 FP 称之为 frame pointer，实际上它根本不是 frame pointer，按照传统的 x86 的习惯来讲，frame pointer 是指向整个 stack frame 底部的 BP 寄存器。假如当前的 callee 函数是 add，在 add 的代码中引用 FP，该 FP 指向的位置不在 callee 的 stack frame 之内，而是在 caller 的 stack frame 上。具体可参见之后的 栈结构 一章。
    2、PC: 实际上就是在体系结构的知识中常见的 pc 寄存器，在 x86 平台下对应 ip 寄存器，amd64 上则是 rip。除了个别跳转之外，手写 plan9 代码与 PC 寄存器打交道的情况较少。
    3、SB: 全局静态基指针，一般用来声明函数或全局变量，在之后的函数知识和示例部分会看到具体用法。
    4、SP: plan9 的这个 SP 寄存器指向当前栈帧的局部变量的开始位置，使用形如 symbol+offset(SP) 的方式，引用函数的局部变量。offset 的合法取值是 [-framesize, 0)，注意是个左闭右开的区间。假如局部变量都是 8 字节，那么第一个局部变量就可以用 localvar0-8(SP) 来表示。这也是一个词不表意的寄存器。与硬件寄存器 SP 是两个不同的东西，在栈帧 size 为 0 的情况下，伪寄存器 SP 和硬件寄存器 SP 指向同一位置。手写汇编代码时，如果是 symbol+offset(SP) 形式，则表示伪寄存器 SP。如果是 offset(SP) 则表示硬件寄存器 SP。务必注意。对于编译输出(go tool compile -S / go tool objdump)的代码来讲，目前所有的 SP 都是硬件寄存器 SP，无论是否带 symbol。

    我们这里对容易混淆的几点简单进行说明：
    
    伪 SP 和硬件 SP 不是一回事，在手写代码时，伪 SP 和硬件 SP 的区分方法是看该 SP 前是否有 symbol。如果有 symbol，那么即为伪寄存器，如果没有，那么说明是硬件 SP 寄存器。
    SP 和 FP 的相对位置是会变的，所以不应该尝试用伪 SP 寄存器去找那些用 FP + offset 来引用的值，例如函数的入参和返回值。
    官方文档中说的伪 SP 指向 stack 的 top，是有问题的。其指向的局部变量位置实际上是整个栈的栈底(除 caller BP 之外)，所以说 bottom 更合适一些。
    在 go tool objdump/go tool compile -S 输出的代码中，是没有伪 SP 和 FP 寄存器的，我们上面说的区分伪 SP 和硬件 SP 寄存器的方法，对于上述两个命令的输出结果是没法使用的。在编译和反汇编的结果中，只有真实的 SP 寄存器。
    FP 和 Go 的官方源代码里的 framepointer 不是一回事，源代码里的 framepointer 指的是 caller BP 寄存器的值，在这里和 caller 的伪 SP 是值是相等的。

    变量声明

    使用 DATA 结合 GLOBL 来定义一个变量，GLOBL 必须跟在 DATA 指令之后，使用 GLOBL 指令将变量声明为 global，额外接收两个参数，一个是 flag（RODATA），另一个是变量的总大小
    DATA    symbol+offset(SB)/width, value
    注意offset 该值相对于符号 symbol 的偏移，而不是相对于全局某个地址的偏移
    例子：

    DATA ·NameData(SB)/8,$"gopher" 变量NameData
    
    DATA birthYear+0(SB)/4, $1988
    GLOBL birthYear(SB), RODATA, $4

    声明数组：具有偏移量
    DATA bio<>+0(SB)/8, $"oh yes i"
    DATA bio<>+8(SB)/8, $"am here "
    GLOBL bio<>(SB), RODATA, $16

    <>，这个跟在符号名之后，表示该全局变量只在当前文件中生效，另外文件中引用该变量的话，会报 relocation target not found 的错误

     flag，还可以有其它的取值，可以查看runtime/textflag.h
    
    NOPROF = 1
    (For TEXT items.) Don’t profile the marked function. This flag is deprecated.
    DUPOK = 2
    It is legal to have multiple instances of this symbol in a single binary. The linker will choose one of the duplicates to use.
    NOSPLIT = 4
    (For TEXT items.) Don’t insert the preamble to check if the stack must be split. The frame for the routine, plus anything it calls, must fit in the spare space at the top of the stack segment. Used to protect routines such as the stack splitting code itself.
    RODATA = 8
    (For DATA and GLOBL items.) Put this data in a read-only section.
    NOPTR = 16
    (For DATA and GLOBL items.) This data contains no pointers and therefore does not need to be scanned by the garbage collector.
    WRAPPER = 32
    (For TEXT items.) This is a wrapper function and should not count as disabling recover.
    NEEDCTXT = 64
    (For TEXT items.) This function is a closure so it uses its incoming context register.

## 汇编函数写法

```go

func PrintName()

TEXT ·PrintName(SB), $16-0
中点 · 比较特殊，是一个 unicode 的中点，该点在 mac 下的输入方法是 option+shift+9
                              参数及返回值大小
                                  | 
 TEXT pkgname·add(SB),NOSPLIT,$32-32
       |        |               |
      包名     函数名         栈帧大小(局部变量+可能需要的额外调用函数的参数空间的总大小，但不包括调用其它函数时的 ret address 的大小)


$16-32 表示 $framesize-argsize，
    argsize = 参数大小求和+返回值大小求和
    $framesize =
    1、局部变量，及其每个变量的 size。
    2、在函数中是否有对其它函数调用时，如果有的话，调用时需要将 callee 的参数、返回值考虑在内。虽然 return address(rip)的值也是存储在 caller 的 stack frame 上的，但是这个过程是由 CALL 指令和 RET 指令完成 PC 寄存器的保存和恢复的，在手写汇编时，同样也是不需要考虑这个 PC 寄存器在栈上所需占用的 8 个字节的。
    3、原则上来说，调用函数时只要不把局部变量覆盖掉就可以了。稍微多分配几个字节的 framesize 也不会死。
    4、在确保逻辑没有问题的前提下，你愿意覆盖局部变量也没有问题。只要保证进入和退出汇编函数时的 caller 和 callee 能正确拿到返回值就可以。

```

## 地址运算

```go
LEA 指令，在amd64 平台地址都是 8 个字节，所以直接就用 LEAQ 就好

LEAQ (BX)(AX*8), CX
代码中的 8 代表 scale，scale 只能是 0、2、4、8
如果写成其它值:
LEAQ (BX)(AX*3), CX
会报错./xx.s:6: bad scale: 3

// 用 LEAQ 的话，即使是两个寄存器值直接相加，也必须提供 scale
// 下面这样是不行的
// LEAQ (BX)(AX), CX
// asm: asmidx: bad address 0/2064/2067
// 正确的写法是
LEAQ (BX)(AX*1), CX


// 在寄存器运算的基础上，可以加上额外的 offset
LEAQ 16(BX)(AX*1), CX
```

##伪寄存器 SP 、伪寄存器 FP 和硬件寄存器 SP #
来写一段简单的代码证明伪 SP、伪 FP 和硬件 SP 的位置关系。 spspfp.s:

#include "textflag.h"

// func output(int) (int, int, int)
TEXT ·output(SB), $8-48
    MOVQ 24(SP), DX // 不带 symbol，这里的 SP 是硬件寄存器 SP
    MOVQ DX, ret3+24(FP) // 第三个返回值
    MOVQ perhapsArg1+16(SP), BX // 当前函数栈大小 > 0，所以 FP 在 SP 的上方 16 字节处
    MOVQ BX, ret2+16(FP) // 第二个返回值
    MOVQ arg1+0(FP), AX
    MOVQ AX, ret1+8(FP)  // 第一个返回值
    RET

    //上面汇编都是操作一个参数赋予给三个返回值，24(SP)，perhapsArg1+16(SP)，arg1+0(FP)都是指向参数的地址，并操作值
    package main
    
    import (
        "fmt"
    )
    
    func output(int) (int, int, int) // 汇编函数声明
    
    func main() {
        a, b, c := output(987654321)
        fmt.Println(a, b, c)
    }
    //987654321 987654321 987654321
    //栈结构是这样的:
    ------
    ret2 (8 bytes)
    ------
    ret1 (8 bytes)
    ------
    ret0 (8 bytes)
    ------
    arg0 (8 bytes)
    ------ FP
    ret addr (8 bytes)
    ------
    caller BP (8 bytes)
    ------ pseudo SP
    frame content (8 bytes)
    ------ hardware SP

## 汇编调用非汇编函数

#include "textflag.h"

// func output(a,b int) int
TEXT ·output(SB), NOSPLIT, $24-24
    MOVQ a+0(FP), DX // arg a
    MOVQ DX, 0(SP) // arg x
    MOVQ b+8(FP), CX // arg b
    MOVQ CX, 8(SP) // arg y
    CALL ·add(SB) // 在调用 add 之前，已经把参数都通过物理寄存器 SP 搬到了函数的栈顶
    MOVQ 16(SP), AX // add 函数会把返回值放在这个位置
    MOVQ AX, ret+16(FP) // return result
    RET

    package main
    
    import "fmt"
    
    func add(x, y int) int {
        return x + y
    }
    
    func output(a, b int) int
    
    func main() {
        s := output(10, 13)
        fmt.Println(s)
    }

## 汇编中的循环

#include "textflag.h"

// func sum(sl []int64) int64
TEXT ·sum(SB), NOSPLIT, $0-32
    MOVQ $0, SI
    MOVQ sl+0(FP), BX // &sl[0], addr of the first elem
    MOVQ sl+8(FP), CX // len(sl)
    INCQ CX           // CX++, 因为要循环 len 次

start:
    DECQ CX       // CX--
    JZ   done
    ADDQ (BX), SI // SI += *BX
    ADDQ $8, BX   // 指针移动
    JMP  start

done:
    // 返回地址是 24 是怎么得来的呢？
    // 可以通过 go tool compile -S math.go 得知
    // 在调用 sum 函数时，会传入三个值，分别为:
    // slice 的首地址、slice 的 len， slice 的 cap
    // 不过我们这里的求和只需要 len，但 cap 依然会占用参数的空间
    // 就是 16(FP)
    MOVQ SI, ret+24(FP)
    RET

package main

func sum([]int64) int64

func main() {
    println(sum([]int64{1, 2, 3, 4, 5}))
}

## slice汇编操作

slice 在传递给函数的时候，实际上会展开成三个参数:

1、首元素地址
2、slice 的 len
3、slice 的 cap

## string汇编操作

```go
package main

//go:noinline
func stringParam(s string) {}

func main() {
    var x = "abcc"
    stringParam(x)
}

```

用 go tool compile -S 输出其汇编:

0x001d 00029 (stringParam.go:11)    LEAQ    go.string."abcc"(SB), AX  // 获取 RODATA 段中的字符串地址
0x0024 00036 (stringParam.go:11)    MOVQ    AX, (SP) // 将获取到的地址放在栈顶，作为第一个参数
0x0028 00040 (stringParam.go:11)    MOVQ    $4, 8(SP) // 字符串长度作为第二个参数
0x0031 00049 (stringParam.go:11)    PCDATA  $0, $0 // gc 相关
0x0031 00049 (stringParam.go:11)    CALL    "".stringParam(SB) // 调用 stringParam 函数

## struct汇编操作

```go

package main

type address struct {
    lng int
    lat int
}

type person struct {
    age    int
    height int
    addr   address
}

func readStruct(p person) (int, int, int, int)

func main() {
    var p = person{
        age:    99,
        height: 88,
        addr: address{
            lng: 77,
            lat: 66,
        },
    }
    a, b, c, d := readStruct(p)
    println(a, b, c, d)
}

#include "textflag.h"

TEXT ·readStruct(SB), NOSPLIT, $0-64
    MOVQ arg0+0(FP), AX
    MOVQ AX, ret0+32(FP)
    MOVQ arg1+8(FP), AX
    MOVQ AX, ret1+40(FP)
    MOVQ arg2+16(FP), AX
    MOVQ AX, ret2+48(FP)
    MOVQ arg3+24(FP), AX
    MOVQ AX, ret3+56(FP)
    RET

会输出 99, 88, 77, 66，这表明即使是内嵌结构体，在内存分布上依然是连续的
```

## map的汇编操作
map的赋值，在读过源码的同学都知道是对期地址进行赋值

```go
package main

func main() {
    var m = map[int]int{}
    m[43] = 1
    var n = map[string]int{}
    n["abc"] = 1
    println(m, n)
}

0x0085 00133 (m.go:7)   LEAQ    type.map[int]int(SB), AX //maptype
0x008c 00140 (m.go:7)   MOVQ    AX, (SP)
0x0090 00144 (m.go:7)   LEAQ    ""..autotmp_2+232(SP), AX//hmap
0x0098 00152 (m.go:7)   MOVQ    AX, 8(SP)
0x009d 00157 (m.go:7)   MOVQ    $43, 16(SP)//uint64
0x00a6 00166 (m.go:7)   PCDATA  $0, $1
0x00a6 00166 (m.go:7)   CALL    runtime.mapassign_fast64(SB)
0x00ab 00171 (m.go:7)   MOVQ    24(SP), AX //最后把mapassign_fast64返回的地址给AX
0x00b0 00176 (m.go:7)   MOVQ    $1, (AX) //往AX上面赋值1

上面汇编都在为mapassign_fast64函数准备参数
func mapassign_fast64(t *maptype, h *hmap, key uint64) unsafe.Pointer {
	
}
```