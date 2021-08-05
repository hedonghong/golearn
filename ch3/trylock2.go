package ch3

import (
	"runtime"
	"sync"
	"sync/atomic"
	"unsafe"
)

const mutexLocked = 1 << iota

//使用 unsafe 操作指针

type Mutex struct{
	sync.Mutex
}

func (m *Mutex) TryLock() bool {
	return atomic.CompareAndSwapInt32((*int32)(unsafe.Pointer(&m.Mutex)), 0, mutexLocked)
}

//自旋锁
type SpinLock1 struct {
	f uint32
}
func (sl *SpinLock1) Lock() {
	for !sl.TryLock() {
		runtime.Gosched()
	}
}
func (sl *SpinLock1) Unlock() {
	atomic.StoreUint32(&sl.f, 0)
}
func (sl *SpinLock1) TryLock() bool {
	return atomic.CompareAndSwapUint32(&sl.f, 0, 1)
}

//自旋锁优化
type spinLock uint32
func (sl *spinLock) Lock() {
	for !atomic.CompareAndSwapUint32((*uint32)(sl), 0, 1) {
		runtime.Gosched() //without this it locks up on GOMAXPROCS > 1
	}
}
func (sl *spinLock) Unlock() {
	atomic.StoreUint32((*uint32)(sl), 0)
}
func (sl *spinLock) TryLock() bool {
	return atomic.CompareAndSwapUint32((*uint32)(sl), 0, 1)
}
func SpinLock() sync.Locker {
	var lock spinLock
	return &lock
}

//channel方式

type ChanMutex chan struct{}
func (m *ChanMutex) Lock() {
	ch := (chan struct{})(*m)
	ch <- struct{}{}
}
func (m *ChanMutex) Unlock() {
	ch := (chan struct{})(*m)
	select {
	case <-ch:
	default:
		panic("unlock of unlocked mutex")
	}
}
func (m *ChanMutex) TryLock() bool {
	ch := (chan struct{})(*m)
	select {
	case ch <- struct{}{}:
		return true
	default:
	}
	return false
}