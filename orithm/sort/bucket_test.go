package sort

import (
	"fmt"
	"testing"
)

func BucketSort(arr []int, bucketSize int) []int {
	if len(arr) <= 1 {
		return arr
	}
	// 获取最大值
	max := arr[0]
	min := arr[0]
	for i:=1; i < len(arr); i++ {
		if max < arr[i] {
			max = arr[i]
		}
		if min > arr[i] {
			min = arr[i]
		}
	}

	// 桶分
	bucketCount := make([][]int, (max - min)/ bucketSize + 1)
	// 数据入桶
	for i:= 0; i < len(arr); i++ {
		bucketCount[(arr[i] - min)/ bucketSize] = append(bucketCount[(arr[i] - min)/ bucketSize], arr[i])
	}
	key := 0
	for _, bucket := range bucketCount {
		if len(bucket) <= 0 {
			continue
		}
		Bubble(bucket)
		for _, value := range bucket {
			arr[key] = value
			key++
		}
	}
	return arr
}


func TestBucketSort(t *testing.T) {
	slice := []int{17,81,12,35,14,72,95,16,79,10}
	fmt.Println(slice)

	slice = BucketSort(slice, 5)

	fmt.Println(slice)
}
