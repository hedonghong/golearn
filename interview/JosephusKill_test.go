package interview

import (
	"fmt"
	"testing"
)

/*
环形单链表约瑟夫问题
m个人围成环形列表，每k个人出列
 */


type People struct {
	N int
	Next *People
}
//1->2->3->4->5->6->->7->8->nil
//1->2->3->4->5->6->->7->8->1


func josephusKill(head *People, k int) *People {
	if head == nil || head.Next == nil {
		return head
	}
	//变成环形链表
	lastPeople := head
	for  {
		if lastPeople.Next == nil {
			break
		} else {
			lastPeople = lastPeople.Next
		}
	}

	lastPeople.Next = head

	var count = 1
	var pre *People
	/**
	//1->2->4->5->6->7->8->1
	//1->2->4->6->7->8->1
	//1->2->4->6->8->1
	//2->4->6->8->2
	//2->6->8->2
	//2->6->2
	//2->2
	 */
	for  {
		if head.Next != head {
			if count == k {//1->2->3->4
				pre.Next = head.Next
				count = 1
			} else {
				pre = head
				head = head.Next
				count++
			}
		} else {
			return head
		}
	}
}

func runNodeFun1(head *People) {
	for  {
		if head != nil {
			fmt.Println(head)
			if head.Next != nil {
				head = head.Next
			} else {
				break
			}
		}
	}
}

func TestJosephusKill(t *testing.T)  {
	node8 := &People{N: 8, Next: nil}//7
	node7 := &People{N: 7, Next: node8}
	node6 := &People{N: 6, Next: node7}
	node5 := &People{N: 5, Next: node6}
	node4 := &People{N: 4, Next: node5}
	node3 := &People{N: 3, Next: node4}
	node2 := &People{N: 2, Next: node3}
	node1 := &People{N: 1, Next: node2}//0

	runNodeFun1(node1)

	fmt.Println(josephusKill(node1, 2))
	fmt.Println(josehusMath(8, 2))
}

func josehusMath(n, m int) int {
	if n == 1 {
		return 0
	}
	return (josehusMath(n-1, m)+ m) % n
}
