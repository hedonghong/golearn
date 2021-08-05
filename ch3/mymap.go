package ch3

import "sync"

// 只支持 int 即可。
type MyMap1 struct {
	mu sync.RWMutex
	data map[interface{}]interface{}
}

func (m *MyMap1) Load(key interface{}) (value interface{}, ok bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if d, ok := m.data[key]; ok {
		return d, true
	}
	return nil, false
}

func (m *MyMap1) Store(key, value interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data[key] = value
}

func (m *MyMap1) Delete(key interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.data[key]; ok {
		delete(m.data, key)
	}
}

func (m *MyMap1) LoadOrStore(key, value interface{}) (actual interface{}, loaded bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if d, ok := m.data[key]; ok {
		return d, true
	} else {
		m.data[key] = value
		return value, false
	}
}

func (m *MyMap1) LoadAndDelete(key interface{}) (value interface{}, loaded bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if d, ok := m.data[key]; ok {
		delete(m.data, key)
		return d, true
	} else {
		return nil, false
	}
}
