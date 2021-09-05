package main

import "fmt"

type Person1 struct {
	age int
}

type T interface {
	foo()
}

type t1 int64

func (t t1) foo()  {
	fmt.Println("t1t1")
}

type t2 string

func (t t2) foo()  {
	fmt.Println("t2t2")
}

func chane(x interface{})  {
	if xx,ok := x.(T); ok {
		xx.foo()
	} else {
		fmt.Println("fail")
	}
}

func main() {
	var a = &Person1{111}

	println(a)

	var (
		b t1
		c t2
	)

	b=33
	c="fff"
	chane(b)
	chane(c)
}
