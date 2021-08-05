package ch3

type MyLock struct {
	lock chan struct{}
}

//如果多个协程往这里面TryLock，是不是一直挂在待发送队列链里面
func MyLocker() *MyLock {
	m := MyLock{
		lock: make(chan struct{}, 1),
	}
	return &m
}

func (m * MyLock) TryLock() bool {
	select {
	case m.lock <- struct{}{}:
		return true
	default:
		return false
	}
}

func (m * MyLock) UnLock()  {
	<- m.lock
}

//其实可用atomic.CompareAndSwapInt32()