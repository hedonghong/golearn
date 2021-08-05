package orithm

import "sync"

type node struct {
	Value interface{}
	Pre *node
	Next *node
}

type DoubleLink struct {
	mx *sync.RWMutex
	Size uint
	Head *node
	Tail *node
}

func (n *node) GetValue() interface{} {
	return n.Value
}

func (n *node) GetPreNode() interface{} {
	return n.Pre
}

func (n *node) GetNextNode() interface{} {
	return n.Next
}

func (n *node) HashNext() bool {
	return n.Next != nil
}

func (n *node) HashPre() bool {
	return n.Pre != nil
}

func (n *node) IsNil() bool {
	return n == nil
}

func NewDoubleLink() *DoubleLink  {
	return &DoubleLink{
		mx:&sync.RWMutex{},
		Size: 0,
		Head: &node{},
		Tail: &node{},
	}
}

func (d *DoubleLink) GetSize() uint {
	return d.Size
}

func (d *DoubleLink) GetHead() *node {
	return d.Head
}

func (d *DoubleLink) GetTail() *node {
	return d.Tail
}

func (d *DoubleLink) Append(value interface{})  {
	d.mx.Lock()
	defer d.mx.Unlock()
	newNode := &node{
		Value: value,
	}
	if d.GetSize() == 0 {
		d.Head.Next = newNode
		newNode.Pre = d.Head
		d.Tail.Pre = newNode
		newNode.Next = d.Tail
	} else {
		pre := d.Tail.Pre
		pre.Next = newNode
		newNode.Pre = pre
		newNode.Next = d.Tail
		d.Tail.Pre = newNode
	}
	d.Size++
}

func (d *DoubleLink) InsertFromHead(n uint, v interface{}) bool {
	d.mx.Lock()
	defer d.mx.Unlock()
	if n > d.GetSize() {
		panic("index out of link")
	}
	var i uint = 1
	thisNode := d.Head
	for ;i <= n; i++  {
		thisNode = thisNode.Next
	}
	newNode := &node{
		Value: v,
	}
	newNode=newNode
	return true
}
