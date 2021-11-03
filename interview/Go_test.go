package interview

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

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