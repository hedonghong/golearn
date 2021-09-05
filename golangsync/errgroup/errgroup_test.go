package errgroup

import (
	"fmt"
	"golang.org/x/sync/errgroup"
	"net/http"
	"sync"
	"testing"
)

func TestWaitGroup(t *testing.T) {
	// 声明一个等待组
	var wg sync.WaitGroup
	// 准备一系列的网站地址
	var urls = []string{
		"http://www.baidu.com/",
		"http://www.baidu.com/",
		"http://www.1234567.com/",//假的
	}
	// 遍历这些地址
	for _, url := range urls {
		// 每一个任务开始时, 将等待组增加1
		wg.Add(1)
		// 开启一个并发
		go func(url string) {
			// 使用defer, 表示函数完成时将等待组值减1
			defer wg.Done()
			// 使用http访问提供的地址
			_, err := http.Get(url)
			// 访问完成后, 打印地址和可能发生的错误
			fmt.Println(url, err)
			// 通过参数传递url地址
		}(url)
	}
	// 等待所有的任务完成
	wg.Wait()
	fmt.Println("over")

}

func TestErrgroup(t *testing.T) {
	var g errgroup.Group
	var urls = []string{
		"http://www.baidu.com/",
		"http://www.baidu.com/",
		"http://www.1234567.com/",//假的
	}
	for _, url := range urls {
		// Launch a goroutine to fetch the URL.
		url := url
		g.Go(func() error {
			// Fetch the URL.
			resp, err := http.Get(url)
			if err == nil { // 这里记得关掉
				resp.Body.Close()
			}
			return err
		})
	}
	// Wait for all HTTP fetches to complete.
	// err是返回的错误
	err := g.Wait()
	if err != nil {
		fmt.Printf("err: %v", err)
		return
	}
	fmt.Println("Successfully fetched all URLs.")
}
