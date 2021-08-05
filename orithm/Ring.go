package orithm

import (
	"errors"
	"fmt"
	"sync"
)

//CircleQueue 环型队列
type CircleQueue struct {
	MaxSize int
	Array   [5]int
	Front   int
	Rear    int
}

//Push 向队列中添加一个值
func (q *CircleQueue) Push(val int) (err error) {
	//先判断队列是否已满
	if q.IsFull() {
		return errors.New("队列已满")
	}
	q.Array[q.Rear] = val
	//队尾不包含元素
	//q.Rear++
	q.Rear = (q.Rear + 1) % q.MaxSize
	return
}

//Pop 得到一个值
func (q *CircleQueue) Pop() (val int, err error) {
	if q.IsEmpty() {
		return -1, errors.New("队列已空")
	}
	//队首包含元素
	val = q.Array[q.Front]
	//q.Front++
	q.Front = (q.Front + 1) % q.MaxSize
	return val, err
}

//IsFull 队列是否满了
func (q *CircleQueue) IsFull() bool {
	return (q.Rear+1)%q.MaxSize == q.Front
}

//IsEmpty 队列是否为空
func (q *CircleQueue) IsEmpty() bool {
	return q.Front == q.Rear
}

//Size 队列的大小
func (q *CircleQueue) Size() int {
	return (q.Rear + q.MaxSize - q.Front) % q.MaxSize
}

//Show 显示队列
func (q *CircleQueue) Show() {
	//取出当前队列有多少元素
	size := q.Size()
	if size == 0 {
		fmt.Println("队列为空")
	}
	//辅助变量，指向Front
	tmpFront := q.Front
	for i := 0; i < size; i++ {
		fmt.Printf("queue[%d]=%v\t", tmpFront, q.Array[tmpFront])
		tmpFront = (tmpFront + 1) % q.MaxSize
	}

}


// 数组栈，后进先出
type ArrayStack struct {
	array []string   // 底层切片
	size  int        // 栈的元素数量
	lock  sync.Mutex // 为了并发安全使用的锁
}

// 入栈
func (stack *ArrayStack) Push(v string) {
	stack.lock.Lock()
	defer stack.lock.Unlock()

	// 放入切片中，后进的元素放在数组最后面
	stack.array = append(stack.array, v)

	// 栈中元素数量+1
	stack.size = stack.size + 1
}

func (stack *ArrayStack) Pop() string {
	stack.lock.Lock()
	defer stack.lock.Unlock()

	// 栈中元素已空
	if stack.size == 0 {
		panic("empty")
	}

	// 栈顶元素
	v := stack.array[stack.size-1]

	// 切片收缩，但可能占用空间越来越大
	//stack.array = stack.array[0 : stack.size-1]

	// 创建新的数组，空间占用不会越来越大，但可能移动元素次数过多
	newArray := make([]string, stack.size-1, stack.size-1)
	for i := 0; i < stack.size-1; i++ {
		newArray[i] = stack.array[i]
	}
	stack.array = newArray

	// 栈中元素数量-1
	stack.size = stack.size - 1
	return v
}
