package ch3

import "testing"

func TestDealLock(t *testing.T) {
	var taskMap = &MyMap{
		m: map[int]task{},
	}
	taskMap.DoMyJob(1)
}
