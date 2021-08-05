package ch3

import (
	"fmt"
	"math/rand"
	"strconv"
	"sync"
	"testing"
	"time"
)

func TestMyMap(t *testing.T) {
	mymap := MyMap1{
		data: make(map[interface{}]interface{}),
	}
	mymap.Store(1, 1)
	mymap.Store(2, 2)
	mymap.Store(3, 3)
	mymap.LoadAndDelete(1)
	mymap.LoadOrStore(4, 4)
	mymap.Delete(2)
	fmt.Println(mymap.Load(3))
	fmt.Println(mymap.Load(4))
}

// go test -bench=. 或者 go test -bench=. xxx.go
// go test -benchtime=5s xxx.go //自定义测试时间
//显示基准测试名称，2000000000 表示测试的次数，也就是 testing.B 结构中提供给程序使用的 N。“0.33 ns/op”表示每一个操作耗费多少时间（纳秒）
//通过-benchtime参数可以自定义测试时间
//在命令行中添加-benchmem参数以显示内存分配情况
//“16 B/op”表示每一次调用需要分配 16 个字节，“2 allocs/op”表示每一次调用有两次分配。
//命令-cpu=1,2,4 表示当启动1个，2个，4个cpu时的情况
//命令-count=10 表示进行几次测试

//go get golang.org/x/perf/cmd/benchstat 测试对比工具

/*
testing包内置了支持生成CPU，内存和块的profile文件。

-cpuprofile=$FILE 将 CPU 分析结果写入 $FILE.
-memprofile=$FILE 将内存分析结果写入 $FILE, -memprofilerate=N 调整记录速率为 1/N.
-blockprofile=$FILE, 将块分析结果写入 $FILE.
使用这些标识中的任何一个同时都会保留二进制文件。

% go test -run=XXX -bench=. -cpuprofile=c.p bytes
% go tool pprof c.p
 进入命令行下可以
 top10 前10个最耗时间的 top10 -cum 以cum排序
 list runtime.scanobject 查看runtime.scanobject函数运行情况
 pdf 产生pdf文件报告
 go tool pprof -http=:30001 cpu.out 通过web查看

*/
func BenchmarkMyMap(b *testing.B) {
	mymap := MyMap1{
		data: make(map[interface{}]interface{}),
	}
	rand.Seed(time.Now().UnixNano())
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			s := strconv.Itoa(rand.Int())
			p := strconv.Itoa(rand.Int())
			mymap.Store(s, s)
			mymap.Load(p)
		}
	})
}

func BenchmarkSyncMap(b *testing.B) {
	syncMap := sync.Map{}
	rand.Seed(time.Now().UnixNano())
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			s := strconv.Itoa(rand.Int())
			p := strconv.Itoa(rand.Int())
			syncMap.Store(s, s)
			syncMap.Load(p)
		}
	})
}

/*
goos: darwin
goarch: amd64
pkg: golearn/ch3
BenchmarkMyMap-4         1000000              1278 ns/op
BenchmarkSyncMap-4       1000000              1545 ns/op
PASS
ok      golearn/ch3     2.864s
 */
