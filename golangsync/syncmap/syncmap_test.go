package syncmap

import (
	"sync"
	"testing"
	"time"
)

func TestMap(t *testing.T) {

	m := make(map[int]int)

	i := 0
	for i < 10 {
		go func() {
			for y:=0; y < 1000; y++ {
				m[i]=i
			}
		}()
		i++
	}
	time.Sleep(10 * time.Second)
}

func TestSyncMap(t *testing.T) {
	var m sync.Map
	i := 0
	for i < 10 {
		go func() {
			for y:=0; y < 1000; y++ {
				m.Store(i, i)
			}
		}()
		i++
	}
	time.Sleep(10 * time.Second)
}
