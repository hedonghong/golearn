# unsafe.Pointer

## 1、Golang指针 区别

    1、 *类型:普通指针类型，用于传递对象地址，不能进行指针运算。
    2、 unsafe.Pointer:通用指针类型，用于转换不同类型的指针，不能进行指针运算，不能读取内存存储的值（必须转换到某一类型的普通指针）。unsafe.Pointer 是桥梁，可以让任意类型的指针实现相互转换，也可以将任意类型的指针转换为 uintptr 进行指针运算。
    3、 uintptr:用于指针运算，GC 不把 uintptr 当指针，uintptr 无法持有对象。uintptr 类型的目标会被回收。比如要在某个指针地址上加上一个偏移量，做完加减法后，转换成Pointer，通过*操作，取值，修改值，随意。


## 2、官方文档对unsafe.Pointer的描述的四个规则：
    （1）任何类型的指针都可以被转化为Pointer
    （2）Pointer可以被转化为任何类型的指针
    （3）uintptr可以被转化为Pointer
    （4）Pointer可以被转化为uintptr

## 3、涉及到内存对齐，为了不用再去翻阅查询内存对接，下面介绍下

### 3.1 查看 unsafepointer_test.go - TestUnsafepointer2、3

### 3.2 为什么要对齐
    操作系统并非一个字节一个字节访问内存，而是按2, 4, 8这样的字长来访问。因此，当CPU从存储器读数据到寄存器，或者从寄存器写数据到存储器，IO的数据长度通常是字长。如 32 位系统访问粒度是 4 字节（bytes），64 位系统的是 8 字节。
    当被访问的数据长度为 n 字节且该数据地址为n字节对齐，那么操作系统就可以高效地一次定位到数据，无需多次读取、处理对齐运算等额外操作。
    数据结构应该尽可能地在自然边界上对齐。如果访问未对齐的内存，CPU需要做两次内存访问。
    word 通常在32位架构上为4，在64位架构上为8。
    对齐倍数必须是2，4，8 即unsafe.Alignof()结果必然是2，4，8
看下go官方文档 [Size and alignment guarantees](https://golang.org/ref/spec#Size_and_alignment_guarantees) 对于go数据类型的大小保证和对齐保证

```go
    type                                 size in bytes
    bool                                  1
    byte, uint8, int8                     1
    uint16, int16                         2
    uint32, int32, float32                4
    uint64, int64, float64, complex64     8
    complex128                           16
    uint, int                            1 word 
    uintptr                              1 word
    string                               2 word
    指针                                  1 word
    slice                                3 word
    map                                  1 word
    channel                              1 word
	func                                 1 word
    interface                            2 word
    非空struct                            字段尺寸+填充尺寸
    数组                                  数组类型尺寸+长度  [4]bool = 4 * 1
    struct{} 和 [0]T{}                   0
```
如果结构或数组类型包含大小为零的字段（或元素），则其大小为零。 两个不同的零大小变量在内存中可能具有相同的地址。
也就是说Struct{}和[0]T{}的大小为0；不同类型的大小为0的变量可能指向同一快地址。

### 3.3 总结下
    内存对齐是为了让cpu更高效的访问内存中的数据
    struct的对齐是：如果类型t的对齐保证是n，那么类型t的每个值的地址在运行时必须是n的倍数
    struct内字段如果填充过多，可以尝试重排，使字段排列更紧密，减少内存浪费
    零大小字段要避免作为struct最后一个字段，会有内存浪费
    32位系统上对64位字的原子访问要保证其是8bytes对齐的；当让如果不必要的话，还是用加锁(mutex)的方式更清晰简单

    零大小字段（zero sized field）是指struct{}，大小为 0，按理作为字段时不需要对齐，但当在作为结构体最后一个字段（final field）时需要对齐的。即开篇我们讲到的面试题的情况，假设有指针指向这个final zero field, 返回的地址将在结构体之外（即指向了别的内存），如果此指针一直存活不释放对应的内存，就会有内存泄露的问题（该内存不因结构体释放而释放），go会对这种final zero field也做填充，使对齐。当然，有一种情况不需要对这个final zero field做额外填充，也就是这个末尾的上一个字段未对齐，需要对这个字段进行填充时，final zero field就不需要再次填充，而是直接利用了上一个字段的填充。



## 4、指针类型转换 - 查看 unsafepointer_test.go - TestUnsafepointer1 例子
Go 语言在设计的时候，为了编写方便、效率高以及降低复杂度，被设计成为一门强类型的静态语言。强类型意味着一旦定义了，它的类型就不能改变了；静态意味着类型检查在运行前就做了。

同时为了安全的考虑，Go 语言是不允许两个指针类型进行转换的。

一般使用 *T 作为一个指针类型，表示一个指向类型T变量的指针。为了安全的考虑，两个不同的指针类型不能相互转换，比如 *int 不能转为 *float64

