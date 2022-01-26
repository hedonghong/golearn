package orithm

import "fmt"

type TreeNode struct {
	Data interface{}
	Left *TreeNode
	Right *TreeNode
}

/*
构建一棵树后，我们希望遍历它，有四种遍历方法：

先序遍历：先访问根节点，再访问左子树，最后访问右子树。
后序遍历：先访问左子树，再访问右子树，最后访问根节点。
中序遍历：先访问左子树，再访问根节点，最后访问右子树。
层次遍历：每一层从左到右访问每一个节点。
*/
//递归
//先序遍历
func PreOrder(tree *TreeNode)  {
 	if tree == nil {
 		return
	}
	//先访问根节点
	fmt.Println(tree.Data)
	//再访问左子树
	PreOrder(tree.Left)
	//最后访问右子树
	PreOrder(tree.Right)
}

//中序遍历
func MidOrder(tree *TreeNode)  {
	if tree == nil {
		return
	}
	//先访问左子树
	MidOrder(tree.Left)
	//再访问根节点
	fmt.Println(tree.Data)
	//最后访问右子树
	MidOrder(tree.Right)
}

//后序遍历
func PostOrder(tree *TreeNode)  {
	if tree == nil {
		return
	}
	//先访问左子树
	PostOrder(tree.Left)
	//再访问右子树
	PostOrder(tree.Right)
	//最后访问根节点
	fmt.Println(tree.Data)
}

//层次遍历，广度遍历
func LayerOrder(tree *TreeNode)  {
	if tree == nil {
		return
	}
	queue := new(LinkStack)

	queue.Push(tree)

	for queue.size > 0 {
		node := queue.Pop().(*TreeNode)
		fmt.Println(node.Data)
		if node.Left != nil {
			queue.Push(node.Left)
		}
		if node.Right != nil {
			queue.Push(node.Right)
		}
	}
}