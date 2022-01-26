package sort

import (
	"fmt"
	"testing"
)

/**
归并排序
将数据切小 多份
合并多份数字
 */
func MergeSort(arr []int) []int  {
	length := len(arr)
	if length <= 1 {
		return arr
	} else {
		mid := length/2
		leftArr := MergeSort(arr[:mid])
		rightArr := MergeSort(arr[mid:])
		return mergeFunc(leftArr, rightArr)
	}
}

func mergeFunc(leftArr, rightArr []int) []int  {
	leftIndex := 0
	rightIndex := 0
	returnArr := make([]int, 0, 0)
	for leftIndex < len(leftArr) && rightIndex < len(rightArr) {
		if leftArr[leftIndex] < rightArr[rightIndex] {
			returnArr = append(returnArr, leftArr[leftIndex])
			leftIndex++
		} else if leftArr[leftIndex] > rightArr[rightIndex] {
			returnArr = append(returnArr, rightArr[rightIndex])
			rightIndex++
		} else {
			returnArr = append(returnArr, leftArr[leftIndex])
			returnArr = append(returnArr, rightArr[rightIndex])
			leftIndex++
			rightIndex++
		}
	}

	for leftIndex < len(leftArr) {
		returnArr = append(returnArr, leftArr[leftIndex])
		leftIndex++
	}

	for rightIndex < len(rightArr) {
		returnArr = append(returnArr, rightArr[rightIndex])
		rightIndex++
	}

	return returnArr
}

func TestMergeSort(t *testing.T) {
	slice := []int{7,8,1,3,4,2,5,6,9,0}
	fmt.Println(slice)

	slice = MergeSort(slice)

	fmt.Println(slice)
}
