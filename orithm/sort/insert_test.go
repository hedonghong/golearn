package sort

import (
	"fmt"
	"testing"
)

//4, 2, 9, 1
//1、[4] 2, 9, 1
//2、[2,4] 9,1
//3、[2,4,9] 1
//4、[1,2,4,9]
//小规模用插入排序
func InsertSort(arr []int64) []int64  {
	len := len(arr)
	for i := 1; i < len - 1; i++ {
		check := arr[i]
		j := i-1
		//如果比左边排序好的数都小，让位处理
		//如果比左边排好的都大，那就不管了
		if check < arr[j] {
			for ; j >= 0 && check < arr[j]; j-- {
				arr[j+1] = arr[j]
			}
			arr[j+1] = check
		}
	}

	return arr
}

func TestInsertSort(t *testing.T) {
	arr := []int64{4, 2, 9, 1}
	fmt.Println(InsertSort(arr))
	arr = []int64{5, 9, 1, 6, 8, 14, 6, 49, 25, 4, 6, 3}
	fmt.Println(InsertSort(arr))
}
