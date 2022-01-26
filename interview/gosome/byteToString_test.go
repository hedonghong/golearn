package gosome

import (
	"testing"
	"unsafe"
)

func TestByteToString(t *testing.T) {
	s := []byte("sky")
	// 第一种
	s1 := string(s)
	s1=s1
	// 第二种
	s2 := *(*string)(unsafe.Pointer(&s))
	s2 = s2
}
