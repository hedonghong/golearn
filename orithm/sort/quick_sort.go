package sort

import (
	"fmt"
	"testing"
)

// 原理：
// 把数组分为3段，
// 7,8,1,3,4,2,5,6,9,0
// 7 为中间段
// 小于7的为一段
// 大于7的为一段
// 反复上面过程


func QuickSort(arr []int) []int {
	length := len(arr)
	if length <=1 {
		return arr
	} else {
		middle := arr[0]
		minArr := make([]int, 0, 0) //比我小的
		maxArr := make([]int, 0, 0) //比我大的
		midArr := make([]int, 0, 0) //与我一样的
		midArr = append(midArr, middle) //先把自己加入
		for i := 1; i < length; i++ {
			if middle < arr[i] {
				maxArr = append(maxArr, arr[i])
			} else if middle > arr[i] {
				minArr = append(minArr, arr[i])
			} else {
				midArr = append(midArr, arr[i])
			}
		}
		minArr, maxArr = QuickSort(minArr),QuickSort(maxArr)
		myArr := append(append(minArr,midArr...), maxArr...)
		return myArr
	}
}


func QuickSort1(arr []int) []int {
	length := len(arr)
	if length <=1 {
		return arr
	} else {
		n := 0 // n >= 0 && n < length - 1 可以随机一个 0 - length-1的范围数
		middle := arr[0]
		minArr := make([]int, 0, 0) //比我小的
		maxArr := make([]int, 0, 0) //比我大的
		midArr := make([]int, 0, 0) //与我一样的
		midArr = append(midArr, middle) //先把自己加入
		for i := 0; i < length; i++ {
			if i == n {
				continue
			}
			if middle < arr[i] {
				maxArr = append(maxArr, arr[i])
			} else if middle > arr[i] {
				minArr = append(minArr, arr[i])
			} else {
				midArr = append(midArr, arr[i])
			}
		}
		minArr, maxArr = QuickSort(minArr),QuickSort(maxArr)
		myArr := append(append(minArr,midArr...), maxArr...)
		return myArr
	}
}

func TestQuick(t *testing.T) {
	slice := []int{7,8,1,3,4,2,5,6,9,0}
	fmt.Println(slice)

	QuickSort(slice)

	fmt.Println(slice)
}
