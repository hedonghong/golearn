package interview

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

/*
1、iota只能在常量的表达式中使用。

fmt.Println(iota)
//编译错误： undefined: iota

2、每次 const 出现时，都会让 iota 初始化为0.

const a = iota // a=0
const (
  b = iota     //b=0
  c            //c=1   相当于c=iota
)
3、自定义类型
自增长常量经常包含一个自定义枚举类型，允许你依靠编译器完成自增设置。

type Stereotype int

const (
    TypicalNoob Stereotype = iota // 0
    TypicalHipster                // 1   TypicalHipster = iota
    TypicalUnixWizard             // 2  TypicalUnixWizard = iota
    TypicalStartupFounder         // 3  TypicalStartupFounder = iota
)
4、可跳过的值

//如果两个const的赋值语句的表达式是一样的，那么可以省略后一个赋值表达式。
type AudioOutput int

const (
    OutMute AudioOutput = iota // 0
    OutMono                    // 1
    OutStereo                  // 2
    _
    _
    OutSurround                // 5
)

5、位掩码表达式
type Allergen int

const (
    IgEggs Allergen = 1 << iota         // 1 << 0 which is 00000001
    IgChocolate                         // 1 << 1 which is 00000010
    IgNuts                              // 1 << 2 which is 00000100
    IgStrawberries                      // 1 << 3 which is 00001000
    IgShellfish                         // 1 << 4 which is 00010000
)

6、定义数量级
type ByteSize float64

const (
    _           = iota                   // ignore first value by assigning to blank identifier
    KB ByteSize = 1 << (10 * iota) // 1 << (10*1)
    MB                                   // 1 << (10*2)
    GB                                   // 1 << (10*3)
    TB                                   // 1 << (10*4)
    PB                                   // 1 << (10*5)
    EB                                   // 1 << (10*6)
    ZB                                   // 1 << (10*7)
    YB                                   // 1 << (10*8)
)
7、定义在一行的情况

const (
    Apple, Banana = iota + 1, iota + 2
    Cherimoya, Durian   // = iota + 1, iota + 2
    Elderberry, Fig     //= iota + 1, iota + 2
)
// Apple: 1
// Banana: 2
// Cherimoya: 2
// Durian: 3
// Elderberry: 3
// Fig: 4

8、中间插队
const (
    i = iota
    j = 3.14
    k = iota
    l
)
//那么打印出来的结果是 i=0,j=3.14,k=2,l=3

 */
const (
	a = iota
	b = iota
)
const (
	name = "menglu"
	c = iota
	d = iota
)

func TestXX(t *testing.T) {
	fmt.Println(a)
	fmt.Println(b)
	fmt.Println(c)
	fmt.Println(d)
}

func TestXX2(t *testing.T) {
	str1 := []string{"a", "b", "c"}
	str2 := str1[1:]
	str2[1] = "new"
	fmt.Println(str1)
	str2 = append(str2, "z", "x", "y")
	fmt.Println(str1)
}

type Student struct {
	Name string
}

func TestXX3(t *testing.T) {
	fmt.Println(&Student{Name: "menglu"} == &Student{Name: "menglu"})
	fmt.Println(Student{Name: "menglu"} == Student{Name: "menglu"})

	fmt.Println([...]string{"1"} == [...]string{"1"})
	//s1 := []string{"1"}
	//s2 := []string{"1"}
	//fmt.Println(s1 == s2)
	//数组长度和类型相同比较，切片不能直接比较
}

func TestXX4(t *testing.T) {
	type student struct {
		Age int
	}

	kv1 :=map[string]student{"xx":{Age: 1}}
	kv2 :=map[string]*student{"xx":{Age: 1}}
	kv3 := kv1["xx"]//map不允许直接通过健找到对应的对象进行赋值
	//map中的元素不是变量，因此不能寻址。
	//1）map作为一个封装好的数据结构，由于它底层可能会由于数据扩张而进行迁移，所以拒绝直接寻址，避免产生野指针；
	//2）map中的key在不存在的时候，赋值语句其实会进行新的k-v值的插入，所以拒绝直接寻址结构体内的字段，以防结构体不存在的时候可能造成的错误；
	//3）这可能和map的并发不安全性相关

	kv3.Age = 2
	kv2["xx"].Age = 3
	fmt.Println(kv1)
	fmt.Println(kv2)
	fmt.Println(kv3)

	s := []student{{Age: 21}}
	s[0].Age = 22
	fmt.Println(s)
}

var (
	mu sync.Mutex
	chain string
)

func TestXX5(t *testing.T) {
	chain = "main"
	A()
	fmt.Println(chain)
}

func A()  {
	mu.Lock()
	defer mu.Unlock()
	chain = chain+"--A"
	B()
}

func B()  {
	chain = chain+"--B"
	fmt.Println(chain)
	C()
}

func C()  {
	mu.Lock()
	defer mu.Unlock()
	chain = chain+"--C"
}



var mu1 sync.RWMutex
var count int

func TestXX6(t *testing.T) {
	go A1()
	time.Sleep(2*time.Second)
	mu1.Lock()
	defer mu1.Unlock()
	count++
	fmt.Println(count)
}

func A1()  {
	mu1.RLock()
	defer mu1.RUnlock()
	B1()
}

func B1()  {
	time.Sleep(5*time.Second)
	C1()
}

func C1()  {
	mu1.RLock()
	defer mu1.RUnlock()
}


type user struct {
	name string
	age int
}

var u = user{name: "Ankur", age: 25}
var g = &u
func modifyUser(pu *user) {
	fmt.Println("modifyUser Received Vaule", pu)
	pu.name = "Anand"
}
func printUser(u <-chan *user) {
	time.Sleep(2 * time.Second)
	fmt.Println("printUser goRoutine called", <-u)
}

func TestXX7(t *testing.T) {
	c := make(chan *user, 5)
	c <- g
	fmt.Println(g)
	// modify g  直接改变了g的地址，而不是修改
	g = &user{name: "Ankur Anand", age: 100}
	go printUser(c)
	go modifyUser(g)
	time.Sleep(5 * time.Second)
	fmt.Println(g)


	// 拷贝进去的是地址，如果g发生改变，接收方也会相应改变
	d := make(chan *user, 5)
	d <- g
	fmt.Println(g)
	g.name = "heihie"
	go printUser(d)
	time.Sleep(5 * time.Second)
}


var wg1 sync.WaitGroup

func TestXX8(t *testing.T) {

	wg1.Add(1)
	go func() {
		time.Sleep(time.Millisecond)
		wg1.Done()
		wg1.Add(1)
	}()
	wg1.Wait()
}