package sort

import (
	"fmt"
	"testing"
)
/*
这是一种分组插入方法，最后一次迭代就相当于是直接插入排序，其他迭代相当于每次移动 n 个距离的直接插入排序，这些整数是两个数之间的距离，我们称它们为增量。

我们取数列长度的一半为增量，以后每次减半，直到增量为1。

举个简单例子，希尔排序一个 12 个元素的数列：[5 9 1 6 8 14 6 49 25 4 6 3]，增量 d 的取值依次为：6，3，1：

x 表示不需要排序的数
取 d = 6 对 [5 x x x x x 6 x x x x x] 进行直接插入排序，没有变化。
取 d = 3 对 [5 x x 6 x x 6 x x 4 x x] 进行直接插入排序，排完序后：[4 x x 5 x x 6 x x 6 x x]。
取 d = 1 对 [4 9 1 5 8 14 6 49 25 6 6 3] 进行直接插入排序，因为 d=1 完全就是直接插入排序了。
 */

func ShellSort(arr []int64) []int64 {
	len := len(arr)
	//step每次减半，直到步长为1
	for step := len/2; step >= 1; step /= 2 {
		for i := step; i < len; i += step {
			for j := i-step; j >= 0; j -= step {
				if arr[j+step] < arr[j] {
					arr[j], arr[j+step] = arr[j+step], arr[j]
					continue
				}
				break
			}
		}
	}
	return arr
}

func TestShellSort(t *testing.T) {
	arr := []int64{4, 2, 9, 1}
	fmt.Println(ShellSort(arr))
	arr = []int64{5, 9, 1, 6, 8, 14, 6, 49, 25, 4, 6, 3}
	fmt.Println(ShellSort(arr))
}
