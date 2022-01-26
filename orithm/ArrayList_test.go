package orithm

import "testing"

func TestArrayList(t *testing.T) {
	arrList := NewArrayList()
	arrList.Append(1)//0
	arrList.Append(2)//1
	arrList.Append(3)//2
	arrList.Append(4)//3
	arrList.Append(5)//4

	arrList.Insert(3, 31)
}
