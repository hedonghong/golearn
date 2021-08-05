package orithm

import (
	"fmt"
	"testing"
)

func TestSliceStack(t *testing.T) {
	sliceStack := new(SliceStack)
	fmt.Println(sliceStack.Pop())
	sliceStack.Push(1)
	sliceStack.Push(2)
	sliceStack.Push(3)
	sliceStack.Push(4)
	fmt.Println(sliceStack.Pop())
}

func TestLinkStack(t *testing.T) {
	linkStack := new(LinkStack)
	fmt.Println(linkStack.Pop())
	linkStack.Push(1)
	linkStack.Push(2)
	linkStack.Push(3)
	linkStack.Push(4)
	fmt.Println(linkStack.Pop())
	fmt.Println(linkStack.Pop())
}
