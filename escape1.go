package main

import (
	"fmt"
	"unsafe"
)

func i_escapes(x int) *int {
	var i int
	i = x
	return &i
}

func j_escapes(x int) *int {
	var j int = x
	j = x
	return &j
}

func k_escapes(x int) *int {
	k := x
	return &k
}

func in_escapes(k int) *int {
	return &k
}

func fxfx() {
	defer func() {
		if err := recover(); err != nil {
			println("recover here")
		}
	}()

	defer func() {
		panic(1)
	}()

	defer func() {
		panic(2)
	}()
	fmt.Println("sss")
}

type m struct {//16
	a int
	b int8
	c int8
	d int16
	e int32
}
type item struct {//32+16 = 48
	a uint64
	b float64
	c int
	d int
	e m
}

func main()  {
	//m := m{}
	//m := make(map[uint64]map[uint64]m)
	m := item{}
	fmt.Println(unsafe.Sizeof(m))
	//fxfx()
}


