package main

import (
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"runtime"
	"time"

	"golearn/pprof/animal"
)

func main1() {
	//https://www.jianshu.com/p/a054fda87918
	//https://github.com/wolfogre/go-pprof-practice
	//1、cpu占用 prof
	//go tool pprof http://localhost:6060/debug/pprof/profile
	//top 只接浏览器上面看
	//top
	//list 关键字  查下代码
	//2、内存占用
	//go tool pprof http://localhost:6060/debug/pprof/heap
	//go tool pprof flag http://127.0.0.1:4500/debug/pprof/heap
	//flag : --inuse_space 分析常驻内存  --alloc_objects分析临时内存
	//top
	//list 关键字  查下代码
	//3、gc
	//GODEBUG=gctrace=1 ./main | grep gc
	//使用gcvis图形节目看gc
	//go get github.com/davecheney/gcvis
	//a、使用相當簡單 gcvis YOUR_APP -index -http=:6060
	//gcvis main（编译后代码入口）
	//b、GODEBUG=gctrace=1  （编译后代码入口）  |& gcvis
	//c、频繁gc优化
	//go tool pprof http://localhost:6060/debug/pprof/allocs
	//top
	//list 关键字  查下代码
	//4、排查协程泄露 协程过多注意
	//go tool pprof http://localhost:6060/debug/pprof/goroutine
	//top
	//list 关键字  查下代码
	//5、排查锁的争用
	//go tool pprof http://localhost:6060/debug/pprof/mutex
	//top
	//list 关键字  查下代码
	//6、排查阻塞操作
	//go tool pprof http://localhost:6060/debug/pprof/block
	//top
	//list 关键字  查下代码

	//curl/wget 下载pprof点
	//curl localhost:8000/debug/pprof/heap > heap.base
	//curl localhost:8000/debug/pprof/heap > heap.current
	//若要时间可以seconds=5 下载5秒左右的情况
	//对比命令
	//go tool pprof -http=:8080 -base heap.base heap.current
	log.SetFlags(log.Lshortfile | log.LstdFlags)
	log.SetOutput(os.Stdout)

	runtime.GOMAXPROCS(1)// 限制 CPU 使用数，避免过载
	runtime.SetMutexProfileFraction(1)// 开启对锁调用的跟踪
	runtime.SetBlockProfileRate(1)// 开启对阻塞操作的跟踪

	go func() {
		if err := http.ListenAndServe(":6060", nil); err != nil {
			log.Fatal(err)
		}
		os.Exit(0)
	}()

	for {
		for _, v := range animal.AllAnimals {
			v.Live()
		}
		time.Sleep(time.Second)
	}
}
/*
上面这种是不修改一行代码的情况下，完全使用外部工具/参数，无侵入式的 GC 监控。

另一种办法是直接读取 runtime.MemStats (runtime/mstats.go) 的内容。其实上面这种办法也是读取了 runtime.memstats (跟 runtime.MemStats 是同一个东西，一个对内，一个对外)。这也意味着要修改我们的程序代码。

我逛了一圈，发现不少人也是这么干的。

NSQ 对 GC 监控 https://github.com/nsqio/nsq/blob/master/nsqd/statsd.go#L117
beego 对 GC 的监控： https://github.com/astaxie/beego/blob/master/toolbox/profile.go#L96
Go port of Coda Hale’s Metrics library https://github.com/rcrowley/go-metrics
A Golang library for exporting performance and runtime metrics to external metrics systems (i.e. statsite, statsd)
https://github.com/armon/go-metrics/
总之，都是读取了 runtime.MemStats ，然后定时发往一些时序数据库之类的，然后展示。

代码基本都是这样：

    memStats := &runtime.MemStats{}
    runtime.ReadMemStats(memStats)
如果希望获取 gcstats:

  gcstats := &debug.GCStats{PauseQuantiles: make([]time.Duration, 100)}
  debug.ReadGCStats(gcstats)
如果你用了 open-falcon 作为监控工具的话，还可以用 github.com/niean/goperfcounter , 配置一下即可使用。

{      "bases": [“runtime”, “debug”], // 分别对应 runtime.MemStats, debug.GCStats  }
如果读者看过 ReadMemStats 的实现的话，应该知道里面调用了 stopTheWorld 。 卧槽，这会不会出事啊！

Russ Cox 说，

We use ReadMemStats internally at Google. I am not sure of the period but it’s something like what you’re talking about (maybe up to once a minute, I forget).

Stopping the world is not a huge problem; stopping the world for a long time is. ReadMemStats stops the world for only a fixed, very short amount of time. So calling it every 10-20 seconds should be fine.

Don’t take my word for it: measure how long it takes and decide whether you’re willing to give up that much of every 10-20 seconds. I expect it would be under 1/1000th of that time (10 ms) .

refer: https://groups.google.com/forum/#!searchin/golang-nuts/ReadMemStats/golang-nuts/mTnw5k4pZdo/rpK69Fns2MsJ

另外， https://github.com/rcrowley/go-metrics 也提到了(go-metrics/runtime.go L:68)

runtime.ReadMemStats(&memStats) // This takes 50-200us .

我觉得一般业务，只要对性能没有很变态的要求，1毫秒内都还能接受吧，也看你读取的频率有多高。

由于每家公司使用的监控系统大相径庭，很难有大一统的解决办法，所以本文只是提供思路以及不严谨的考证。祝大家玩的开心！
 */
