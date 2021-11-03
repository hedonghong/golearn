
# 1、defer的生效顺序

defer的执行顺序是倒序执行（同入栈先进后出）

```go
type _defer struct {
	siz       int32 //是延迟函数参数和结果的内存大小
	started   bool //是否已经开始执行
	openDefer bool //是否开发编码的方式实现的defer，如果是后面painc涉及到栈扫描defer，把defer函数解析到对应所在函数的栈上面，不分配在堆上，但出现painc会导致需要扫描栈查看需要执行的defer
	sp        uintptr //sp、pc分别代表栈指针和调用方的程序计数器
	pc        uintptr
	fn        *funcval //defer 关键字中传入的函数
	_panic    *_panic //触发延迟调用的结构体，可能为空
	link      *_defer //下一个defer结构体
    heap      bool //是否_defer在堆上分配，一般涉及循环defer、
    ...
}
```

# 2、defer与return,函数返回值之间的顺序

return最先执行->return负责将结果写入返回值中->接着defer开始执行一些收尾工作->最后函数携带当前返回值退出

返回值的表达方式，我们知道根据是否提前声明有两种方式：一种是func test() int 另一种是 func test() (i int)，所以两种情况都来说说

```go
// 匿名参数函数
func Anonymous() int {
     var i int
     defer func() {
      i++
      fmt.Println("defer2 value is ", i)
     }()
    
     defer func() {
      i++
      fmt.Println("defer1 in value is ", i)
     }()
    
     return i
}

//命名参数函数
func HasName() (j int) {
    defer func() {
    j++
    fmt.Println("defer2 in value", j)
    }()
    
    defer func() {
    j++
    fmt.Println("defer1 in value", j)
    }()
    
    return j
}

1. Anonymous()的返回值为0
2. HasName()的返回值为2
```

# 3、defer定义和执行两个步骤，做的事情

会先将defer后函数的参数部分的值(或者地址)给先下来【你可以理解为()里头的会先确定】，后面函数执行完，才会执行defer后函数的{}中的逻辑