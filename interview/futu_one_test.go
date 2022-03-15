package interview

import (
	"errors"
	"fmt"
	"testing"
	"time"
)

func TestSlice(t *testing.T) {
	months := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12}
	q2 := months[3:6]
	q2e := q2[:4]
	fmt.Println(q2, len(q2), cap(q2))
	// 4,5,6    3  12-3=9
	fmt.Println(q2e, len(q2e), cap(q2e))
	// 4,5,6,7 4   9
}

type TestStruct struct{}

func NilOrNot(v interface{}) bool {
	return v == nil
}

func Print(v interface{}) {
	println(v)
}

func TestInterFace(t *testing.T) {
	//nil := "ssss"
	//fmt.Println(nil)
	// 上面可以，但是nil不再是nil

	//fmt.Println(nil == nil) //报错
	type Test struct{}
	v := Test{}
	Print(v)
	var s *TestStruct
	fmt.Println(s == nil)
	fmt.Println(NilOrNot(s))
}

func TestNil(t *testing.T) {
	// 指针类型的nil比较
	fmt.Println((*int64)(nil) == (*int64)(nil))
	// channel 类型的nil比较
	fmt.Println((chan int)(nil) == (chan int)(nil))
	// func类型的nil比较
	// fmt.Println((func())(nil) == (func())(nil)) // func() 只能与nil进行比较
	// interface类型的nil比较
	fmt.Println((interface{})(nil) == (interface{})(nil))
	// map类型的nil比较
	// fmt.Println((map[string]int)(nil) == (map[string]int)(nil)) // map 只能与nil进行比较
	// slice类型的nil比较
	// fmt.Println(([]int)(nil) == ([]int)(nil)) // slice 只能与nil进行比较
}

func myAppend(s []int) {
	s = append(s, 5)
}

func myAdd(s []int) {
	for i := range s {
		s[i] = s[i] + 5
	}
}

func TestSlice1(t *testing.T) {
	s := []int{1, 2, 3, 4}
	myAppend(s)
	fmt.Println(s)
	myAdd(s)
	fmt.Println(s)
}

func foo() (err error) {
	defer func() {
		fmt.Println(err)
		err = errors.New("a")
	}()

	defer func(e error) {
		fmt.Println(e)
		e = errors.New("b")
	}(err)

	err = errors.New("c")
	return
}

func TestError(t *testing.T) {
	fmt.Println(foo())
	// nil c a
}

func TestArr(t *testing.T) {
	//对于大数组，如果使用for range遍历，遍历前的转换过程会很浪费内存，可以优化：
	//(1)对数组取地址遍历for i, n := range &arr；(2)对数组做切片引用for i, n := range arr[:]；
	arr := [1000]int{1, 2, 3, 4}
	for i, v := range &arr {
		fmt.Println(i, v)
	}
	for x, y := range arr[:] {
		fmt.Println(x, y)
	}
}

// @TODO 重点复习
func TestFor(t *testing.T) {
	total, sum := 0, 0
	// 有个陷阱就是 i会加到11 并且到了go里面都是一样的，因为i地址是一样，并且协程运行不确定
	for i := 1; i <= 10; i++ {
		sum += i
		go func() {
			fmt.Println(i)
			//total += i
		}()
	}
	// 0 55
	// 50 55
	// 不确定，55
	fmt.Printf("total %d sum %d", total, sum)
	time.Sleep(time.Second * 5)
}
