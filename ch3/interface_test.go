package ch3

import (
	"fmt"
	"reflect"
	"testing"
)

type TestStruct struct{}

func NilOrNot(v interface{}) bool {
	t2 := reflect.TypeOf(v)
	fmt.Println("v type", t2.Kind())
	fmt.Println("v value", reflect.ValueOf(v))
	return v == nil
}

func TestInterface(t *testing.T) {
	var s *TestStruct
	t1 := reflect.TypeOf(s)
	fmt.Println("s type", t1.Kind())
	fmt.Println("s value", reflect.ValueOf(s))
	fmt.Println(s == nil)      // #=> true
	fmt.Println(NilOrNot(s))   // #=> false
}

func TestPanic(t *testing.T) {
	defer fmt.Println("in main1")
	defer fmt.Println("in main2")
	defer func() {
		defer func() {
			panic("panic again and again")
		}()
		panic("panic again")
	}()

	panic("panic once")
}
