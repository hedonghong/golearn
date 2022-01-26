package gosome

import (
	"reflect"
	"testing"
	"unsafe"
)

func TestStringToByte(t *testing.T) {
	s := "sky"
	// 第一种
	s1 := []byte(s)
	s1=s1
	// 第二种
	s2 := *(*[]byte)(unsafe.Pointer(&reflect.SliceHeader{
		Data: (*(*reflect.StringHeader)(unsafe.Pointer(&s))).Data,
		Len: len(s),
		Cap: len(s),
	}))
	s2 = s2
	//第一种使用[]byte这种直接转化，也是我们常用的方式，第二种是使用unsafe的方式。这两种区别就在于一个是重新分配了内存，
	//另一个是复用了原来的内存。
}
