package interview

import (
	"fmt"
	"testing"
)

//数组、链表去重

//O(1)

//数组去重

//链表去重

type
Node struct {
	Value int64
	Next *Node
}

//数组版
func TestArrDistinct(t *testing.T) {
	intArr := []int{1, 2, 2, 3, 3, 4, 4, 5, 5, 5, 6}
	var (
		slow int
		fast int
	)
	slow , fast = 0, 1
	for fast < len(intArr) {
		if intArr[slow] != intArr[fast] {
			slow++
			intArr[slow] = intArr[fast]
		}
		fast++
	}
	//因为切片[)左闭又开，所以再加1
	slow++
	fmt.Println(intArr[:slow])
}

func TestNode(t *testing.T) {
	node1 := Node{Value: 1}
	node2 := Node{Value: 2}
	node2_1 := Node{Value: 2}
	node3 := Node{Value: 3}
	node3_1 := Node{Value: 3}
	node4 := Node{Value: 4}
	node4_1 := Node{Value: 4}
	node5 := Node{Value: 5}
	node5_1 := Node{Value: 5}
	node6 := Node{Value: 6}

	node1.Next = &node2
	node2.Next = &node2_1
	node2_1.Next = &node3
	node3.Next = &node3_1
	node3_1.Next = &node4
	node4.Next = &node4_1
	node4_1.Next = &node5
	node5.Next = &node5_1
	node5_1.Next = &node6
	node6.Next = nil

	head := &node1
	next := &node1
	for  {
		fmt.Println(*next)
		next = next.Next
		if next == nil {
			break
		}
	}

	slow , fast := &node1, node1.Next
	for fast != nil {
		if slow.Value != fast.Value {
			slow.Next = fast
			slow = slow.Next
		}
		fast = fast.Next
	}
	slow.Next = nil
	for  {
		fmt.Println(*head)
		head = head.Next
		if head == nil {
			break
		}
	}
}