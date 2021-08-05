package ch3

import "sync"

type task struct{}

type MyMap struct {
	m   map[int]task
	mux sync.RWMutex
}

func (m *MyMap) finishJob(t task, id int) {
	m.mux.Lock()
	defer m.mux.Unlock()

	// finish task
	delete(m.m, id)
}

func (m *MyMap) DoMyJob(taskID int) {
	//或者去掉defer m.mux.RUnlock() 直接用 m.mux.RUnlock()
	setjob := func () task {
		m.mux.RLock()
		defer m.mux.RUnlock()
		t := m.m[taskID]
		return t
	}

	t := setjob()

	m.finishJob(t, taskID)
}
