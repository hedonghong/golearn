package interview

import (
	"fmt"
	"testing"
)

//递归练习

//第一要素：明确你这个函数想要干什么，定义一个函数不管内部实现先

//第二要素：寻找递归结束条件

//第三要素：找出函数的等价关系式子

//学习例子：斐波那契数列 1 1 2 3 5 8

/*

第一步
func f(n int) int {

}

第二步
明显1，1，2
f(1) = 1
f(2) = 1
f(3) = f(1) + f(2) = 1 + 1 = 2

if n <= 2 {
	return 1
}

第三步：
f(n) = f(n-1) + f(n-2)


 */

func f(n int) int {
	if n <= 2 {
		return 1
	}
	return f(n-1) + f(n-2)
}

//1 1 2 3 5 8
func TestF(t *testing.T) {
	fmt.Println(f(5))//5
	fmt.Println(f(6))//8
}

//学习例子 小青蛙跳台阶：一次跳台阶，一次跳二台阶的方式，n个台阶有多少种方式

/*
第一步：

func f1(n int) int {

}

第二步：
0  0
1  1 ：1；
2  2 ：1，1；2
3  3 ：1，1，1；1，2；2，1；
4  5 ：1，1，1，1；1，2，1；2，1，1；1，1，2；2，2

if n <= 3 {
	return n
}

第三步：

f(3) = f(2)+ f(1) // 3
f(4) = f(3) + f(2) // 5
f(5) = f(4) + f(3) // 8
f(n) = f(n-1) + f(n-2)

 */

func f1(n int) int {
	//n <= 2 或者 n <= 3
	if n <= 3 {
		return n
	}
	return f1(n-1) + f1(n-2)
}

func TestF1(t *testing.T) {
	fmt.Println(f1(3))
	fmt.Println(f1(4))
	fmt.Println(f1(5))
}

//学习例子：反转单链表 head->1->2->3->4 head->4->3->2->1
/**

type Node struct {
	N int
	Next Node
}

第一步：
func f2(head Node) Node {

}

第二步：
if head == nil || head.Next == nil {
	return head
}

第三步：

head->1->2->3->4
假如
newHead->4->3->2<-1<-head

newHead->4->3->2->1->nil

代码：
newHead := f3(head.Next)
node2 := head.Next
node2.Next = head
head.Next= null
return newHead;
 */

type ElemNode struct {
	N int
	Next *ElemNode
}

func f3(head *ElemNode) *ElemNode {
	if head == nil || head.Next == nil {
		return head
	}
	newHead := f3(head.Next)
	node2 := head.Next
	node2.Next = head
	head.Next= nil
	return newHead
}

func TestF3(t *testing.T) {
	node4 := &ElemNode{N: 4, Next: nil}
	node3 := &ElemNode{N: 3, Next: node4}
	node2 := &ElemNode{N: 2, Next: node3}
	node1 := &ElemNode{N: 1, Next: node2}

	runNodeFun(node1)

	newNode := f3(node1)
	runNodeFun(newNode)
}

func runNodeFun(head *ElemNode) {
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

//优化递归思路
/*

1、考虑重复计算是否可以缓存起来，比如上面的f(n) = f(n-1) + f(n-2)
	比如map存储起来，因为f(n-1)后面肯定有和f(n-2) 重复计算的部分
	通过map缓存起来，直接取用
2、考虑是否可以自底向上
	对于递归的问题，我们一般都是从上往下递归的，直到递归到最底，再一层一层着把值返回。
	但是递归存在占用大量栈空间的问题，大n递归的栈空间不够用会有问题。那么考虑自底向上计算
	f(1) = 1; f(2) = 2; f(3) = f(2) + f(1) = 3

	下面f4函数就是这样子通过循环取消递归，这种方法，其实也被称之为递推

 */

func f4(n int) int {
	if n <= 2 {
		return n
	}
	n1 := 1
	n2 := 2

	sum := 0

	for i := 3; i <= n; i++ {
		sum = n1 + n2
		n1 = n2
		n2 = sum
	}
	return sum
}


func TestF4(t *testing.T) {
	node4 := &ElemNode{N: 4, Next: nil}
	node3 := &ElemNode{N: 3, Next: node4}
	node2 := &ElemNode{N: 2, Next: node3}
	node1 := &ElemNode{N: 1, Next: node2}

	runNodeFun(node1)

	newNode := reverseLink(node1)
	runNodeFun(newNode)
}

//head->1->2->3->4
//一个节点一个节点地搞，pre保存上一个节点
func reverseLink(node *ElemNode) *ElemNode {
	var next *ElemNode = nil//当前节点的后驱
	var pre *ElemNode = nil//当前节点的前驱
	for node != nil {
		//node 1
		next = node.Next//2
		//当前节点的后驱指向前驱
		node.Next = pre//nil
		pre = node //1
		//处理下一个节点
		node = next //2
	}
	return pre
}

/*
题目：给定一个单向链表的头结点head,以及两个整数from和to ,在单项链表上把第from个节点和第to个节点这一部分进行反转

列如：
 1->2->3->4->5->null,from=2,to=4

结果：1->4->3->2->5->null

列如：

1->2->3->null from=1,to=3

结果为3->2->1->null

要求】
1、如果链表长度为N，时间复杂度要求为O（N),额外空间复杂度要求为O（1）

2、如果不满足1<=from<=to<=N,则不调整

*/


// 1->  2->3->4->  5->null,from=2,to=4
func reverseParkLink(node *ElemNode, from, to int) *ElemNode {
	if node.N > from {
		return node
	}
	var node1 *ElemNode = node
	var fpre *ElemNode//from 上一个
	var tpos *ElemNode//to 后一个
	var len int = 1
	for  {
		if node1 == nil {
			break
		}
		if len == from-1 {
			fpre = node1//1
		}
		if len == to + 1 {
			tpos = node1//5
		}
		len++
		node1 = node1.Next
	}
	if from > to || from < 1 || to > len {
		return node
	}
	if fpre == nil {
		//包含第一个元素 from = 1
		node1 = node //1 要连接tpos元素的 node1.next = tpos
	} else {
		//不包含第一个元素 from > 1
		node1 = fpre.Next//2 //要连接tpos元素的 node1.next = tpos
	}
	//cur = 3->4->5->nil cur是下一个要操作的元素，会一直变化
	cur := node1.Next//3 设置下第一个要操作的元素
	//2->5->nil 连接tpos
	node1.Next = tpos
	var next *ElemNode
	for cur != tpos {
		next = cur.Next//4 获取下一个要操作的元素
		cur.Next = node1//3->2 把当前操作的元素设置为下一个元素的next
		node1 = cur//<-3 记录倒叙后的头部元素->3->2 后面->4->3->2
		cur = next//<-4 设置下一个原色到cur中循环
	}
	if fpre != nil {
		////1->(插入)->2->5->nil 更改为 1->4->3->2->5->nil
		fpre.Next = node1
		//返回原来
		return node
	}
	return node1
}

func TestF5(t *testing.T) {
	node5 := &ElemNode{N: 5, Next: nil}
	node4 := &ElemNode{N: 4, Next: node5}
	node3 := &ElemNode{N: 3, Next: node4}
	node2 := &ElemNode{N: 2, Next: node3}
	node1 := &ElemNode{N: 1, Next: node2}

	runNodeFun(node1)

	newNode := reverseParkLink(node1, 2,4)
	runNodeFun(newNode)
}



/*
题目：

链表:1->2->3->4->5->6->7->8->null, K = 3。那么 6->7->8，3->4->5，1->2各位一组。调整后：1->2->5->4->3->8->7->6->null。其中 1，2不调整，因为不够一组

 */

func reverseGroup(head *ElemNode, k int) *ElemNode {

	//第一段
	var tempHead *ElemNode = head
	for i := 1; i < k && tempHead != nil; i++  {
		tempHead = tempHead.Next
	}
	if tempHead == nil {
		return head
	}

	//下一段的head元素保存
	var nextHead *ElemNode = tempHead.Next
	//把第一段的最后一个元素的next改为nil
	tempHead.Next = nil
	//把第一段的元素就行逆序
	newHead := reverseLink(head)
	//把第二段进行分组逆序
	newTemp := reverseGroup(nextHead, k)
	//把第一组和第二组连接起来
	head.Next = newTemp
	//返回新分组逆序链表
	return newHead
}

func reverseList(head *ElemNode) *ElemNode {
	var (
		next *ElemNode
		pre *ElemNode
	)
	for head != nil {
		next = head.Next
		head.Next = pre
		pre = head
		head = next
	}
	return pre
}

func TestF6(t *testing.T) {
	node8 := &ElemNode{N: 8, Next: nil}
	node7 := &ElemNode{N: 7, Next: node8}
	node6 := &ElemNode{N: 6, Next: node7}
	node5 := &ElemNode{N: 5, Next: node6}
	node4 := &ElemNode{N: 4, Next: node5}
	node3 := &ElemNode{N: 3, Next: node4}
	node2 := &ElemNode{N: 2, Next: node3}
	node1 := &ElemNode{N: 1, Next: node2}

	runNodeFun(node1)

	newNode := reverseGroup(node1, 3)
	runNodeFun(newNode)
}


/*
题目：
例如： 链表:1->2->3->4->5->6->7->8->null, K = 3。那么 6->7->8，3->4->5，1->2各位一组。调整后：1->2->5->4->3->8->7->6->null。其中 1，2不调整，因为不够一组。

先逆序一次，分组逆序，在逆序一次
 */

func TestF7(t *testing.T) {
	node8 := &ElemNode{N: 8, Next: nil}
	node7 := &ElemNode{N: 7, Next: node8}
	node6 := &ElemNode{N: 6, Next: node7}
	node5 := &ElemNode{N: 5, Next: node6}
	node4 := &ElemNode{N: 4, Next: node5}
	node3 := &ElemNode{N: 3, Next: node4}
	node2 := &ElemNode{N: 2, Next: node3}
	node1 := &ElemNode{N: 1, Next: node2}
	runNodeFun(node1)

	newNode := reverseLink(node1)
	runNodeFun(newNode)

	newNode1 := reverseGroup(newNode, 3)

	newNode2 := reverseLink(newNode1)
	runNodeFun(newNode2)
}




