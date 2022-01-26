package slice

import (
	"fmt"
	"os"
	"reflect"
	"testing"
	"unsafe"
)

func TestSlice(t *testing.T) {
	var arr = [10]int {1,2,3,4,5,6,7,8,9,10}
	s1 := arr[2:5:9]
	s2 := s1[1:7]
	fmt.Println(arr)//[1 2 3 4 5 6 7 8 9 10]
	fmt.Println(s1)//[3 4 5]
	fmt.Fprintf(os.Stdout,"s1 len:%d, cap:%d \n", len(s1), cap(s1))//s1 len:3, cap:7
	fmt.Println(s2)//[4 5 6 7 8 9]
	fmt.Fprintf(os.Stdout,"s2 len:%d, cap:%d \n", len(s2), cap(s2))//s2 len:6, cap:6
}

func TestNilEmptySlice(t *testing.T) {
	var s1 []int
	s2 := make([]int,0)
	s3 := make([]int,0)

	fmt.Printf("s1 pointer:%+v, s2 pointer:%+v, s3 pointer:%+v, \n", *(*reflect.SliceHeader)(unsafe.Pointer(&s1)),*(*reflect.SliceHeader)(unsafe.Pointer(&s2)),*(*reflect.SliceHeader)(unsafe.Pointer(&s3)))
	fmt.Printf("%v\n", (*(*reflect.SliceHeader)(unsafe.Pointer(&s1))).Data==(*(*reflect.SliceHeader)(unsafe.Pointer(&s2))).Data)
	fmt.Printf("%v\n", (*(*reflect.SliceHeader)(unsafe.Pointer(&s2))).Data==(*(*reflect.SliceHeader)(unsafe.Pointer(&s3))).Data)
}
