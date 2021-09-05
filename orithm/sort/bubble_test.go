package sort

import (
	"fmt"
	"testing"
)

func TestBubble(t *testing.T) {
	var arr []int64 = []int64{3,4,1,7,3,9}
	fmt.Println("before:", arr)
	fmt.Println("after:", Bubble(arr))
}


//第一轮： 3,4,1,7,3,9  3,4对比，不用换，4，1对比，置换成 1，4 ....
func Bubble(arr []int64) []int64 {
	leng := len(arr)
	if leng <= 1 {
		return arr
	}
	for i := leng - 1; i > 0; i--  {
		noChange := true
		for j := 0; j < i; j++   {
			if arr[j] > arr[j+1] {
				arr[j], arr[j+1] = arr[j+1], arr[j]
				noChange = false
			}
		}
		if noChange {
			break
		}
	}
	return arr
}

func TestSelect(t *testing.T) {
	var arr []int64 = []int64{3,4,1,7,3,9}
	fmt.Println("before:", arr)
	fmt.Println("after:", Select(arr))
}

func Select(arr []int64) []int64 {
	leng := len(arr)
	if leng <= 1 {
		return arr
	}
	var minIndex int
	for i := 0; i < leng - 1; i++ {
		//从i~n-1的位置上选择最小值，放到i位置上，这样一直循环整个数组
		minIndex = i
		for j := i + 1; j < leng; j++ {
			if arr[j] < arr[minIndex] {
				minIndex = j
			}
		}
		arr[i], arr[minIndex] = arr[minIndex], arr[i]
	}
	return arr
}

//异或
func TestYihuo(t *testing.T) {
	//相同为0 不同为1
	m := 1^0
	n := 1^1

	fmt.Printf("m: %d, n: %d \n", m, n)

	//a = a^b^b a= b^a^b
	a := 1
	b := 2

	a = a^b
	b = a^b
	a = a^b

	fmt.Printf("a:%d, b:%d \n", a, b)

	intarr := []int{1, 2, 3, 3, 2} //一种数字出现一次，其余为偶次，求这个数
	
	eor := 0

	for _, i := range intarr {
		eor ^= i
	}
	fmt.Println("intarr one type:", eor)

	intarr = []int{1, 2, 3, 3, 2, 4} //两种数字出现一次，其余为偶次，求这两个数

	eor = 0
	for _, i := range intarr {
		eor ^= i
	}
	//假设 a != b eor = a^b
	//0001
	//0100
	//0101
	//那他们的2进制数 数位有不同的情况
	eora := eor & (^eor + 1) //^a 二进制取反，提取最右边的出现1的位
	//0101      -> 1010 + 0001 = 1011
	//1011
	//0001
	p := 0
	for _, j := range intarr {
		if (j & eora) == 1 {
			p ^= j
		}
	}

	fmt.Printf("intarr two type: %d, %d \n", p, eor^p)

}

func TestInsert(t *testing.T) {
	var arr []int64 = []int64{3,4,1,7,3,9}
	fmt.Println("before:", arr)
	fmt.Println("after:", Insert(arr))
}

//手里那着牌，新牌网里面合适的位置插入
func Insert(arr []int64) []int64 {
	leng := len(arr)
	if leng <= 1 {
		return arr
	}

	for i := 1; i < leng; i++ {
		for j := i; j > 0; j--  {
			if arr[j] < arr[j-1] {
				arr[j], arr[j-1] = arr[j-1], arr[j]
			} else {
				goto THIS
			}
		}
		THIS:
	}

	return arr
}