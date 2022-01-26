package sort

import (
	"fmt"
	"testing"
)

//循环从左到右查找最小的元素与数组第一个未交互的元素交互
// 4 2 9 1
//1、1 2 9 4
//2、1 2 9 4
//3、1 2 4 9
func SelectMinSort(arr []int64) []int64 {
	len := len(arr)
	for i := 0; i < len; i++ {
		min:= arr[i]
		minIndex := i
		for j := i+1; j < len;j++ {
			if arr[j] < min {
				min = arr[j]
				minIndex = j
			}
		}
		if i != minIndex {
			arr[i], arr[minIndex] = arr[minIndex], arr[i]
		}
	}
	return arr
}

//优化：在找最新的同时，找下最大的，那就可以节省一半循环
func SelectMinMaxSort(arr []int64) []int64 {
	len := len(arr)
	for i := 0; i < len/2; i++ {
		minIndex := i
		maxIndex := i
		for j := i+1; j < len - i;j++ {
			//最大
			if arr[j] > arr[maxIndex] {
				maxIndex = j
				continue
			}
			//最小
			if arr[j] < arr[minIndex] {
				minIndex = j
			}
		}

		if maxIndex == i && minIndex == len-i-1 {
			//最大值在开头，最小值在最后，直接交互
			arr[maxIndex], arr[minIndex] = arr[minIndex], arr[maxIndex]
		} else {
			//否则最小值放在开头，最大值放在结尾
			arr[minIndex], arr[i] = arr[i],arr[minIndex]
			arr[maxIndex], arr[len-i-1] = arr[len-i-1],arr[maxIndex]
		}
	}
	return arr
}

func TestSelectSort(t *testing.T) {
	arr := []int64{4, 2, 9, 1}
	fmt.Println(SelectMinMaxSort(arr))
	arr = []int64{5, 9, 1, 6, 8, 14, 6, 49, 25, 4, 6, 3}
	fmt.Println(SelectMinMaxSort(arr))
}
