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
