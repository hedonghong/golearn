package array

import (
	"fmt"
	"testing"
)

//有一个整型数组，数组元素不重复，数组元素先升序后降序，找出最大值。
//例如：1,3,5,7,9,8,6,4,2，请写一个函数找出数组最大的元素

//如何用二分法缩小空间呢？只要比较中间元素与其下一个元素大小即可
//
//如果中间元素大于其下一个元素大小，证明最大值在左侧，因此右指针左移
//如果中间元素小于其下一个元素大小，证明最大值在左侧，因此右指针左移

func maxNum(nums []int) int {
	length := len(nums)
	if length <= 0 {
		return -1
	}
	// 下表
	left := 0
	right := length - 1
	var mid int = 0
	for left <= right {
		mid = left + (right-left)/2
		if nums[mid] > nums[mid+1] {
			right = mid - 1
		} else {
			left = mid + 1
		}
	}
	return nums[left]
}

func TestMaxNum(t *testing.T) {
	arr := []int{1, 3, 5, 7, 9, 8, 6, 4, 2}
	fmt.Println(maxNum(arr))
}
