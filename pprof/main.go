package main

/**
go trace 实验
 */

import (
"context"
"fmt"
"os"
"runtime"
"runtime/trace"
"sync"
)

/**
 go tool trace trace.out 浏览器会打开，1.14.4左右的版本会有问题，打开一片空白
 */
func main(){
	// 为了看协程抢占，这里设置了一个cpu 跑
	runtime.GOMAXPROCS(1)

	f, _ := os.Create("trace.out")
	defer f.Close()

	_ = trace.Start(f)
	defer trace.Stop()

	ctx,  task := trace.NewTask(context.Background(), "sumTask")
	defer task.End()

	var wg sync.WaitGroup
	wg.Add(10)
	for i := 0; i < 10; i ++ {
		// 启动10个协程，只是做一个累加运算
		go func(region string) {
			defer wg.Done()

			// 标记region
			trace.WithRegion(ctx, region, func() {
				var sum, k int64
				for ; k < 1000000000; k ++ {
					sum += k
				}
				fmt.Println(region, sum)
			})
		}(fmt.Sprintf("region_%02d", i))
	}
	wg.Wait()
}
