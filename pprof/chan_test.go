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
