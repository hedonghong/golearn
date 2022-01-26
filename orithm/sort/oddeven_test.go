package sort

import (
	"fmt"
	"testing"
)
/**
奇偶排序
先进行奇数位的交换
再进行偶数位的交换
最终不用交换就是有序
 */
func OddEven(arr []int) []int {
	length := len(arr)
	if length <= 1 {
		return arr
	}
	isSorted := false


	for ;isSorted == false;{
		isSorted = true
		// length - 1 是因为下面arr[i+1]
		for i := 1; i < length - 1;i += 2 {
			if arr[i] > arr[i+1] {
				arr[i], arr[i+1] = arr[i+1], arr[i]
				isSorted = false
			}
		}
		for i := 0; i < length - 1;i += 2 {
			if arr[i] > arr[i+1] {
				arr[i], arr[i+1] = arr[i+1], arr[i]
				isSorted = false
			}
		}
	}
	return arr
}

func TestOddEven(t *testing.T) {
	slice := []int{7,8,1,3,4,2,5,6,9,0}
	fmt.Println(slice)

	slice = OddEven(slice)

	fmt.Println(slice)
}

