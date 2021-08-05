package ch3

import (
	"fmt"
	"testing"
)

func TestTryLock(t *testing.T) {
	m :=MyLocker()
	fmt.Println(m.TryLock())
	fmt.Println(m.TryLock())
	m.UnLock()
	fmt.Println(m.TryLock())
	m.UnLock()
}
