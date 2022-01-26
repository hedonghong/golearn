package gosome

import (
	"reflect"
	"testing"
	"unicode/utf8"
	"unsafe"
)

func TestStringToRune(t *testing.T) {
	s := "sky"
	s1 := []rune(s)
	s1=s1
}


func TestRuneToString(t *testing.T) {
	s := "sky"
	s1 := []rune(s)

	// 计算字节长度
	var l int
	for _, r := range s1 {
		l += utf8.RuneLen(r)
	}
	s2 := *((*string)(unsafe.Pointer(&reflect.StringHeader{
		Data: (*((*reflect.SliceHeader)(unsafe.Pointer(&s1)))).Data,
		Len: l,
	})))
	s2=s2
}
