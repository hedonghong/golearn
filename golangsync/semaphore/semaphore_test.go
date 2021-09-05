package semaphore

import (
	"context"
	"fmt"
	"golang.org/x/sync/semaphore"
	"sync"
	"testing"
	"time"
)

func TestName(t *testing.T) {
	//在golang中goroutine的开启是非常廉价的
	//使用者可能无意开启了很多的goroutine导致泄漏等
	//为了控制goroutine的数量，我们可以使用的信号量
	//这里的信号量与linux里面的信号有区别。
	//for i := 0; i < math.MaxInt64; i++ {
	//	go func(i int) {
	//		time.Sleep(5 * time.Second)
	//	}(i)
	//}
	//一般可以通过goroutine池控制，比如ants、go-playground/pool、jeffail/tunny

	//runNums := 9 // 要运行的goroutine数量
	//limitNums := 3 // 同时运行的goroutine为3个
	//ch := make(chan bool, limitNums)
	//wg := sync.WaitGroup{}
	//wg.Add(runNums)
	//for i:=0; i < runNums; i++{
	//	go func(num int) {
	//		defer wg.Done()
	//		ch <- true // 发送信号
	//		fmt.Printf("%d 我在干活 at time %d\n",num,time.Now().Unix())
	//		time.Sleep(2 * time.Second)
	//		<- ch // 接收数据代表退出了
	//	}(i)
	//}
	//wg.Wait()

	//https://github.com/eddycjy/gsema/blob/master/sema.go

	//+++++++++++++++++分界线+++++++++

	names := []string{
		"asong1",
		"asong2",
		"asong3",
		"asong4",
		"asong5",
		"asong6",
		"asong7",
	}

	var (
		limit int64 = 3// 同时运行的goroutine上限
		weight int64 = 1 // 信号量的权重
	)

	sem := semaphore.NewWeighted(limit)
	var w sync.WaitGroup
	for _, name := range names {
		w.Add(1)
		//每次Acquire都减1，直到释放，否则后面的goroutine只能等待
		if err := sem.Acquire(context.Background(), weight); err != nil {
			break
		}
		go func(name string) {
			defer sem.Release(weight)
			defer w.Done()
			fmt.Println(name)
			time.Sleep(10 * time.Second) // 延时能更好的体现出来控制
		}(name)
	}
	w.Wait()

	fmt.Println("over--------")

	//+++++++++++++++++分界线+++++++++

	//serviceName := []string{
	//	"cart",
	//	"order",
	//	"account",
	//	"item",
	//	"menu",
	//}
	//eg,ctx := errgroup.WithContext(context.Background())
	//s := semaphore.NewWeighted(limit)
	//for index := range serviceName{
	//	name := serviceName[index]
	//	if err := s.Acquire(ctx,1); err != nil{
	//		fmt.Printf("Acquire failed and err is %s\n", err.Error())
	//		break
	//	}
	//	eg.Go(func() error {
	//		defer s.Release(1)
	//		return callService(name)
	//	})
	//}
	//
	//if err := eg.Wait(); err != nil{
	//	fmt.Printf("err is %s\n", err.Error())
	//	return
	//}
	//fmt.Printf("run success\n")
}

func callService(name string) error {
	fmt.Println("call ",name)
	time.Sleep(1 * time.Second)
	return nil
}
