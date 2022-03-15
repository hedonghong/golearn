package array

import (
	"fmt"
	"testing"
)

func twoSum(nums []int, target int) []int {
	numMap := make(map[int]int)
	for k, _ := range nums {
		anther := target - nums[k]
		if _, ok := numMap[anther]; ok {
			return []int{k, numMap[anther]}
		}
		numMap[nums[k]] = k
	}
	return nil
}

func TestTwoSum(t *testing.T) {
	// 找出数组中相加等于目标数字的
	nums := []int{1, 3, 5, 7, 8}
	fmt.Println(twoSum(nums, 9))
}
