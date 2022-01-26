package orithm

import (
	"errors"
	"fmt"
)

type List interface {
	Size() int
	Get(index int) (interface{}, error)
	Set(index int, newval interface{}) error
	Insert(index int, newval interface{}) error
	Append(newval interface{})
	Clear()
	Delete(index int) error
	String() string
}

//1、用切片唯一问题就是不会缩容，可以通过算法优化
//2、没并发锁
type ArrayList struct {
 	dataStore []interface{}
 	theSize int
}
//可以增加容量和实际数据保存情况，进行对比调整，设定临界值
//若容量大于多少多少，但实际数据保存多少则进行缩容，即重新循环拷贝切片

func NewArrayList() *ArrayList  {
	return &ArrayList{
		dataStore: make([]interface{}, 0),
		theSize: 0,
	}
}

func (a *ArrayList) Clear() {
	a.dataStore = make([]interface{}, 0)
	a.theSize = 0
}

func (a *ArrayList) Size() int {
	return a.theSize
}

func (a *ArrayList) Get(index int) (interface{}, error) {
	if index < 0 || index > a.theSize {
		return nil,errors.New("越界")
	}
	return a.dataStore[index], nil
}

func (a *ArrayList) Append(newval interface{}) {
	a.dataStore = append(a.dataStore, newval)
	a.theSize++
}

func (a *ArrayList) String() string {
	return fmt.Sprint(a.dataStore)
}

func (a *ArrayList) Delete(index int) error {
	if index < 0 || index > a.theSize {
		return errors.New("越界")
	}
	a.dataStore = append(a.dataStore[:index], a.dataStore[index+1:]...)
	a.theSize--
	return nil
}
func (a *ArrayList) Set(index int, newval interface{}) error {
	if index < 0 || index > a.theSize {
		return errors.New("越界")
	}
	a.dataStore[index] = newval
	return nil
}
func (a *ArrayList) Insert(index int, newval interface{}) error {
	if index < 0 || index > a.theSize {
		return errors.New("越界")
	}
	//0,1,2,3,4,5 index 若index = 3
	//1,2,3,4,5,6 value

	//0,1,2,3,4,5,6 index
	//1,2,3,3,4,5,6 value
	for size := a.theSize; index < size; size-- {
		if size == a.theSize {
			a.Append(0)
		}
		a.dataStore[size] = a.dataStore[size-1]
	}
	a.dataStore[index] = newval
	a.theSize++
	return nil
}