package ch3

import (
	"context"
	"sync"
)

type MyContext struct {
	child map[interface{}] context.Context
	m sync.Mutex
}

func (m * MyContext) WithValue(parent context.Context, key, val interface{}) context.Context {
	c := context.WithValue(parent, key, val)
	m.m.Lock()
	m.child[key] = c
	m.m.Unlock()
	return c
}

func (m *MyContext) Value(key interface{}) interface{} {
	if c, ok := m.child[key]; ok {
		return c.Value(key)
	}
	return nil
}
