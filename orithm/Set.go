package orithm

import "sync"

type Set struct {
	m map[int]struct{}
	cap int
	len int
	sync.RWMutex
}

func NewSet(cap int) *Set  {
	return &Set{
		m: make(map[int]struct{}, cap),
		cap: cap,
	}
}

func (s *Set) Add(item int)  {
	s.Lock()
	defer s.Unlock()
	s.m[item] = struct{}{}
	s.len = len(s.m)
}

func (s *Set) Remove(item int) {
	s.Lock()
	defer s.Unlock()
	delete(s.m, item)
	s.len = len(s.m)
}

func (s *Set) Has(item int) bool {
	s.RLock()
	defer s.RUnlock()
	_, ok := s.m[item]
	return ok
}

func (s *Set) Len() int {
	return s.len
}

func (s *Set) Clear(){
	s.Lock()
	defer s.Unlock()
	s.m = make(map[int]struct{}, s.cap)
	s.len = 0
}

func (s *Set) isEmpty() bool {
	if s.len == 0 {
		return true
	}
	return false
}

