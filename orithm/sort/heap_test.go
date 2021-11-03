package sort

import (
	"fmt"
	"testing"
)
//大顶堆 用于升序排列
//小顶堆 用于降序排序

//满二叉树 是所有子叶子节点是左右满的
//是一个完全二叉树，允许右节点没有叶子，相差层高为1

//核心思想是：维护三角关系的大小顶，比如大顶堆的父节点一定比子节点大。交换顶点和最后一个叶子节点，这个时候订点节点已经出堆。重新排序新订点的三角关系。
//重复如此，直到排序结束。

//可以使用数组来进行堆的存储
//关系有下标为i的父节点下标为(i-1)/2 整数除法
//从n/2向下取整为根节点开始建堆，从底部到顶部构建大顶堆，最后一个非叶子节点开始
//那么第一非叶子节点就是 (9-1)/2，数组下标是0开始的
//i的左节点下标：i*2+1
//i的右节点下标：i*2+2

//s 切片，n切片的长度，维持三角关系的i下标
//  6
// / \
// 7  5
// 这样一个堆，如果6的元素下标是0，7是1，5是2，按照下面函数
//元素6和7交换，largest最终=1那么想象如果1下标元素换了，如果后面还是有元素呢
//那么必须重新维护一下三角关系
func heapify(s []int, n, i int)  {
	largest := i
	lson := i*2+1
	rson := i*2+2
	if (lson < n) && (s[largest] < s[lson]) {
		largest = lson
	}
	if (rson < n) && (s[largest] < s[rson]) {
		largest = rson
	}
	if largest != i {
		//交换值
		s[largest], s[i] = s[i], s[largest]
		//那么必须重新维护一下三角关系
		heapify(s, n, largest)
	}
}

//s切片，n是s的长度
func heapSort(s []int, n int)  {
	//建堆
	for i:= (n-1)/2; i >= 0; i-- {
		heapify(s, n, i)
	}
	//排序
	for i:= n - 1; i > 0; i-- {
		s[i], s[0] = s[0], s[i]
		heapify(s, i, 0)
	}
}

func TestHeap(t *testing.T) {
	slice := []int{7,8,1,3,4,2,5,6,9,0}
	fmt.Println(slice)

	heapSort(slice, len(slice))

	fmt.Println(slice)
}