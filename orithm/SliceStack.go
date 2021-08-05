package orithm

import "sync"

type SliceStack struct {
	stack []interface{}
	mu sync.Mutex
}

func (s *SliceStack) Push(v interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.stack = append(s.stack, v)
}

func (s *SliceStack) Pop() interface{} {
	s.mu.Lock()
	defer s.mu.Unlock()
	index := len(s.stack)
	if index <= 0 {
		return nil

	}
	v := s.stack[index-1]

	////1、无法释放空间
	//s.stack = s.stack[0 : index-1]
	////2、新建切片
	//newSlice := make([]interface{}, index-1, index-1)
	//for item := range s.stack {
	//	newSlice = append(newSlice, item)
	//}
	//s.stack = newSlice
	//3、copy
	newSlice1 := make([]interface{}, index-1, index-1)
	copy(newSlice1, s.stack)
	s.stack = newSlice1
	return v
}

type LinkStack struct {
	root *LinkNode
	size int
	mu sync.Mutex
}

type LinkNode struct {
	Value interface{}
	NextNode *LinkNode
}

func (l *LinkStack) Push(v interface{})  {
	l.mu.Lock()
	defer l.mu.Unlock()
	newLinkNode := &LinkNode{Value: v}
	if l.size == 0 {
		l.root = newLinkNode
	} else {
		headNode := l.root
		newLinkNode.NextNode = headNode
		l.root = newLinkNode
	}
	l.size++
}

func (l *LinkStack) Pop() interface{} {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.size <= 0 {
		return nil
	}
	nextNode := l.root.NextNode
	popNode  := l.root
	l.root    = nextNode
	l.size--
	return popNode.Value
}
