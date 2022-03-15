package main

import (
	"fmt"
	"testing"
	"time"
)

func TestChan(t *testing.T) {
	ch := make(chan int, 10)

	go func() {
		for i := 0; i < 1000; i++ {
			fmt.Println("has:", i)
			select {
			case ch <- i:
				fmt.Println("in:", i)
			default:
				time.Sleep(1 * time.Second)
				fmt.Println("wait in:")
			}
		}
	}()

	for  {
		select {
		case i := <- ch:
			time.Sleep(10 * time.Second)
			fmt.Println("out:", i)
		default:
			time.Sleep(1 * time.Second)
			fmt.Println("wait out:")
		}
	}
}
func TestSSS(t *testing.T) {
	arr := []int{3, 4, 5, 1, 2}
	k := 0
	l := len(arr)
	for x1,x := range arr {
		l1 := l-(x1+1)
		for y1,y := range arr {
			if x1 == y1 {
				continue
			}
			if x > y {
				k+=1
				if l1 == k {
					goto SS
				}
			} else {
				break
			}
		}
	}
	SS:
	fmt.Println(l-k)
}
