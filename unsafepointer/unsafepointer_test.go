package unsafepointer

import (
	"fmt"
	"testing"
	"unsafe"
)

type User struct {
	name string
	age int
}

func TestUnsafepointer1(t *testing.T) {
	//a := 10
	//b := &a
	//var c *float64 = (*float64)(b)
	////Cannot convert expression of type '*int' to type '*float64'

	a := 10
	b := &a
	var c *float64 = (*float64)(unsafe.Pointer(b))
	*c = *c * 3
	fmt.Println(a)


	u := new(User)
	fmt.Println(u)

	//第一个修改 user 的 name 值的时候，因为 name 是第一个字段，所以不用偏移，我们获取 user 的指针，然后通过 unsafe.Pointer 转为 *string 进行赋值操作即可。
	//第二个修改 user 的 age 值的时候，因为 age 不是第一个字段，所以我们需要内存偏移，内存偏移牵涉到的计算只能通过 uintptr，所我们要先把 user 的指针地址转为 uintptr，然后我们再通过 unsafe.Offsetof(u.age) 获取需要偏移的值，进行地址运算(+)偏移即可。
	pName := (*string)(unsafe.Pointer(u))
	*pName = "sky"

	//这里我们可以看到，我们第二个偏移的表达式非常长，但是也千万不要把他们分段，不能像下面这样。
	//temp := uintptr(unsafe.Pointer(u)) + unsafe.Offsetof(u.age)
	//pAge := (*int)(unsafe.Pointer(temp))
	//这里会牵涉到 GC，如果我们的这些临时变量被 GC，那么导致的内存操作就错了，我们最终操作的，就不知道是哪块内存了，会引起莫名其妙的问题
	pAge := (*int)(unsafe.Pointer(
		uintptr(unsafe.Pointer(u)) + unsafe.Offsetof(u.age)))
	*pAge = 20
	fmt.Println(u)


	//[]byte和string其实内部的存储结构都是一样的，但 Go 语言的类型系统禁止他俩互换。
	//如果借助unsafe.Pointer，我们就可以实现在零拷贝的情况下，将[]byte数组直接转换成string类型
	bytes := []byte{104, 101, 108, 108, 111}

	p := unsafe.Pointer(&bytes) //强制转换成unsafe.Pointer，编译器不会报错
	str := *(*string)(p) //然后强制转换成string类型的指针，再将这个指针的值当做string类型取出来
	fmt.Println(str) //输出 "hello"
}

func TestUnsafepointer2(t *testing.T) {
	//ArbitraryType仅用于文档目的，实际上并不是unsafe包的一部分,它表示任意Go表达式的类型。
	//type ArbitraryType int
	//任意类型的指针，类似于C的*void
	//type Pointer *ArbitraryType
	//确定结构在内存中占用的确切大小
	//func Sizeof(x ArbitraryType) uintptr
	//返回结构体中某个field的偏移量
	//func Offsetof(x ArbitraryType) uintptr
	//返回结构体中某个field的对其值（字节对齐的原因）
	//func Alignof(x ArbitraryType) uintptr
	u1 := User{
		name: "xxx",//8+8
		age: 1,//8
	}
	fmt.Println(unsafe.Sizeof(u1))//24
	fmt.Println(unsafe.Alignof(u1))//对齐是8
	u := new(User)
	fmt.Println(unsafe.Offsetof(u.name))//string 8+8 2个字节
	fmt.Println(unsafe.Offsetof(u.age))
}

func TestSss(t *testing.T) {
	type Part2 struct {
		e byte//1
		c int8//1
		a bool//1
		b int32//4
		d int64//8
	}
	part2 := Part2{}
	fmt.Printf("part2 size: %d, align: %d\n", unsafe.Sizeof(part2), unsafe.Alignof(part2))
}

//每种类型的所占字节数不一样
type S struct {
	A uint32//4字节+4padding
	B uint64//8字节
	C uint64//8字节
	D uint64//8字节
	E struct{}//0字节 + 8padding，struct一般不放在末尾，防止对齐占用字节，先说S.E后面隐藏着一个8字节的padding，因为内存对齐的关系
}

type S1 struct {
	E struct{}//0字节，struct一般不放在末尾，防止对齐占用字节，先说S.E后面隐藏着一个8字节的padding，因为内存对齐的关系
	A uint32//4字节
	B uint64//8字节
	C uint64//8字节
	D uint64//8字节
}

type S2 struct {
	A chan int//8
	B chan string//8
	C []int//24
	D []string//24
	E map[string]string//8
}

type S3 struct {
	A uint64//8
	B uint32
	C int32//B+C = 4 + 4 = 8
	D int8 //1+7padding = 8
	E struct{}// 由于上一个字段没有满，直接不用填充
}

type S4 struct {
	A uint64//8
	B uint32
	C int32//B+C = 4 + 4 = 8
	D int8 //1+7padding = 8
	E [0]int32// 由于上一个字段没有满 ，直接不用填充 0 * 4 = 0
}

type S5 struct {
	A uint64//8
	B uint32
	C int32//B+C = 4 + 4 = 8
	D uint64 //8
	E struct{}//上一个字段对齐，需要填充 8
}

func TestUnsafepointer3(t *testing.T) {
	//上面的struct S，占用多大的内存
	fmt.Println(unsafe.Offsetof(S{}.E))//32
	fmt.Println(unsafe.Sizeof(S{}.E))//0
	fmt.Println(unsafe.Sizeof(S{}))//40
	fmt.Println(unsafe.Sizeof(S1{}))//32
	fmt.Println(unsafe.Sizeof(S2{}))//72
	fmt.Println(unsafe.Sizeof(S3{}))//24
	fmt.Println(unsafe.Sizeof(S4{}))//24
	fmt.Println(unsafe.Sizeof(S5{}))//32

	// 注意下面例子，连续声明的表量对应的内存地址会怎么样？相差24吗？猜猜
	var S6,S7 S3
	fmt.Println(unsafe.Sizeof(S6))//24
	fmt.Println(unsafe.Sizeof(S7))//24

	fmt.Println("++++++++++")
	fmt.Printf("S6.E offset:%d , S6.E sizeof:%d, S6 sizeof: %d, S6.address:%v \n", unsafe.Offsetof(S6.E), unsafe.Sizeof(S6.E), unsafe.Sizeof(S6), &S6.A)

	fmt.Printf("S7.E offset:%d , S7.E sizeof:%d, S7 sizeof: %d, S7.address:%v \n", unsafe.Offsetof(S7.E), unsafe.Sizeof(S7.E), unsafe.Sizeof(S7), &S7.A)

	ptr6 := uintptr(unsafe.Pointer(&S6))
	//S6.E offset:17 , S6.E sizeof:0, S6 sizeof: 24, S6.address:0xc00001a220
	ptr7 := uintptr(unsafe.Pointer(&S7))
	//S7.E offset:17 , S7.E sizeof:0, S7 sizeof: 24, S7.address:0xc00001a240
	fmt.Println(ptr7 - ptr6)//32 内存是连续的，为什么偏移了32
	//因为对于s6这个对象，它的大小是24bytes，而go在内存分配时，会从span中拿大于或等于40的最小的span中的一个块给这个对象，而sizeclass中这个块的大小值为32，所以虽然s6的大小是24bytes，但是实际分配给这个对象的内存大小是32，这里面涉及到golang的内存分配和管理，golang在runtime中枚举67种内存span
	//在Go语言中,对象被划分为三种大小
	//小于16字节 – 微对象
	//大于16字节,小于32KB – 小对象
	//大于32KB – 大对象

	fmt.Println("++++++++++")

	var S8,S9 S5
	fmt.Println(unsafe.Sizeof(S8))//32
	fmt.Println(unsafe.Sizeof(S9))//32

	fmt.Println("++++++++++")
	fmt.Printf("S8.E offset:%d , S8.E sizeof:%d, S8 sizeof: %d, S8.address:%v \n", unsafe.Offsetof(S8.E), unsafe.Sizeof(S8.E), unsafe.Sizeof(S8), &S8.A)

	fmt.Printf("S9.E offset:%d , S9.E sizeof:%d, S9 sizeof: %d, S9.address:%v \n", unsafe.Offsetof(S9.E), unsafe.Sizeof(S9.E), unsafe.Sizeof(S9), &S9.A)

	ptr8 := uintptr(unsafe.Pointer(&S8))
	ptr9 := uintptr(unsafe.Pointer(&S9))
	fmt.Println(ptr9 - ptr8)
	//S8.E offset:24 , S8.E sizeof:0, S8 sizeof: 32, S8.address:0xc00013c120
	//S9.E offset:24 , S9.E sizeof:0, S9 sizeof: 32, S9.address:0xc00013c140
	//32

	var S10,S11 S
	fmt.Println(unsafe.Sizeof(S10))//32
	fmt.Println(unsafe.Sizeof(S11))//32

	fmt.Println("++++++++++")
	fmt.Printf("S10.E offset:%d , S10.E sizeof:%d, S10 sizeof: %d, S10.address:%v \n", unsafe.Offsetof(S10.E), unsafe.Sizeof(S10.E), unsafe.Sizeof(S10), &S10.A)

	fmt.Printf("S11.E offset:%d , S11.E sizeof:%d, S11 sizeof: %d, S11.address:%v \n", unsafe.Offsetof(S11.E), unsafe.Sizeof(S11.E), unsafe.Sizeof(S11), &S11.A)

	ptr10 := uintptr(unsafe.Pointer(&S10))
	ptr11 := uintptr(unsafe.Pointer(&S11))
	fmt.Println(ptr11 - ptr10)
	//S10.E offset:32 , S10.E sizeof:0, S10 sizeof: 40, S10.address:0xc00009e060
	//S11.E offset:32 , S11.E sizeof:0, S11 sizeof: 40, S11.address:0xc00009e090
	//48
}
