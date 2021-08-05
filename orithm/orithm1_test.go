package orithm

import (
	"container/ring"
	"fmt"
	"testing"
)
// go test orithm1_test.go
// go test -v orithm1_test.go
// go test -v -run TestRescuvie orithm1_test.go
func TestRescuvie(t *testing.T) {
	fmt.Println(Rescuvie(5))
}

func Rescuvie(n int) int {
	if n == 0 {
		return 1
	}
	return n*Rescuvie(n-1)
}

func TestRescuvieTail(t *testing.T) {
	fmt.Println(RescuvieTail(5, 1))
}

func RescuvieTail(n, a int ) int {
	if  n == 1 {
		return a
	}
	return RescuvieTail(n-1, a*n)
}

func TestBinarySearch(t *testing.T) {
	array := []int{1, 5, 9, 15, 81, 89, 123, 189, 333}
	//fmt.Println(BinarySearch(array, 400, 0, len(array)-1))
	//fmt.Println(BinarySearch(array, 123, 0, len(array)-1))
	fmt.Println(BinarySearch2(array, 400, 0, len(array)-1))
	fmt.Println(BinarySearch2(array, 123, 0, len(array)-1))
}

func BinarySearch(array []int, target, l, r int) int  {
	if l > r {
		return -1
	}
	temp := (l+r)/2
	mid := array[temp]
	if mid == target {
		return mid
	} else if mid > target {
		return BinarySearch(array, target, l, temp-1)
	} else {
		return BinarySearch(array, target, temp+1, r)
	}
}

func BinarySearch2(array []int, target, l, r int) int {
	ltemp := l
	rtemp := r
	for {
		if ltemp > rtemp {
			return -1
		}
		temp := (ltemp+rtemp)/2
		mid := array[temp]
		if mid == target {
			return mid
		} else if mid > target {
			rtemp = temp-1
		} else {
			ltemp = temp+1
		}
	}
}


func TestRing(t *testing.T) {
	r := ring.New(6)
	n := r.Len()
	for i := 0; i < n; i++ {
		r.Value = i
		r = r.Next()
	}
	r.Link(&ring.Ring{Value: 6})
	r.Do(func(p interface{}) {
		fmt.Println(p.(int))
	})
	//temp := r.Move(3+1)
	//temp.Do(func(p interface{}) {
	//	fmt.Println(p.(int))
	//})
	//0, 1, 2, 3, 4, 5
	r.Unlink(3)
	//r.Do(func(p interface{}) {
	//	fmt.Println(p.(int))
	//})
	//0, 4, 5
}
