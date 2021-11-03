package interview

import (
	"fmt"
	"sync"
	"testing"
)

//每一个 thread 都会被传入一个 printNumber() 以输出一个整数。修改已给的代码，使其输出序列为 010203040506…，该序列长度必须为 2n。
//
//完整中文说明如下：
//
//假设有这么一个类：
//
//class ZeroEvenOdd {
//public ZeroEvenOdd(int n) { ... }      // 构造函数
//public void zero(printNumber) { ... }  // 仅打印出 0
//public void even(printNumber) { ... }  // 仅打印出 偶数
//public void odd(printNumber) { ... }   // 仅打印出 奇数
//}
//相同的一个 ZeroEvenOdd 类实例将会传递给三个不同的线程：
//
//线程 A 将调用 zero()，它只输出 0 。
//线程 B 将调用 even()，它只输出偶数。
//线程 C 将调用 odd()，它只输出奇数。
//每个线程都有一个 printNumber 方法来输出一个整数。请修改给出的代码以输出整数序列 010203040506... ，其中序列的长度必须为 2n。
//
//示例 1：
//
//输入：n = 2
//输出："0102"
//说明：三条线程异步执行，其中一个调用 zero()，另一个线程调用 even()，最后一个线程调用odd()。正确的输出为 "0102"。
//示例 2：
//
//输入：n = 5
//输出："0102030405"

type ZeroEventOdd struct {

	Wg sync.WaitGroup

	Num int
	Start int

	ZeroCh chan int
	EvenCh chan int
	OddCh chan int
}

func (z *ZeroEventOdd) zero(printNumber func(int))  {
	defer z.Wg.Done()
	z.Start++
	for _ = range z.ZeroCh {
		printNumber(0)
		if (z.Start%2) == 0 {
			z.EvenCh <- z.Start
		} else {
			z.OddCh <- z.Start
		}
	}
}

func (z *ZeroEventOdd) even(printNumber func(int))  {
	defer z.Wg.Done()
	for v := range z.EvenCh {
		printNumber(v)
		if v == z.Num {
			z.PrintEnd()
			return
		}
		z.Start++
		z.ZeroCh <- 0
	}
}

func (z *ZeroEventOdd) odd(printNumber func(int))  {
	defer z.Wg.Done()
	for v := range z.OddCh {
		printNumber(v)
		if v == z.Num {
			z.PrintEnd()
			return
		}
		z.Start++
		z.ZeroCh <- 0
	}
}

func PrintNumber(x int) {
	fmt.Print(x)
}

func (z *ZeroEventOdd) PrintEnd() {
	z.Start = 0
	z.Num = 0
	close(z.ZeroCh)
	close(z.EvenCh)
	close(z.OddCh)
}


func TestZeroEventOdd(t *testing.T) {
	n := 5
	zeo := &ZeroEventOdd{
		Num: n,
		Start: 0,
		ZeroCh: make(chan int),
		EvenCh: make(chan int),
		OddCh: make(chan int),
	}

	zeo.Wg.Add(3)
	go zeo.zero(PrintNumber)
	go zeo.even(PrintNumber)
	go zeo.odd(PrintNumber)
	zeo.ZeroCh <- 0
	zeo.Wg.Wait()
}

func TestPrintNumberLetter(t *testing.T) {
	var printNumCh, printLetCh chan struct{}

	wg := sync.WaitGroup{}

	wg.Add(1)
	go func() {
		var i int = 0
		for _ = range printNumCh {
			fmt.Print(i)
			i++
			fmt.Print(i)
			printLetCh <- struct{}{}
		}
	}()

	go func() {
		defer wg.Done()
		s := 'A'
		for s <= 'Z' {
			<- printLetCh
			fmt.Print(s)
			s++
			fmt.Print(s)
			s++
			printNumCh <- struct{}{}
		}
	}()
	printNumCh <- struct{}{}
	wg.Wait()
}

