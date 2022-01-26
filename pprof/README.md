# go-pprof-practice

[《golang pprof 实战》](https://blog.wolfogre.com/posts/go-ppof-practice/)代码实验用例。

1、生成报告runtime/pprof

import "runtime/pprof"

var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")

func main() {
    flag.Parse()
    if *cpuprofile != "" {
        f, err := os.Create(*cpuprofile)
        if err != nil {
            log.Fatal(err)
        }
        pprof.StartCPUProfile(f)
        defer pprof.StopCPUProfile()
    }
    ...

}

var memprofile = flag.String("memprofile", "", "write memory profile to this file")

func main() {
    // …………

    FindHavlakLoops(cfgraph, lsgraph)
    if *memprofile != "" {
        f, err := os.Create(*memprofile)
        if err != nil {
            log.Fatal(err)
        }
        pprof.WriteHeapProfile(f)
        f.Close()
        return
    }
    
    // …………
}


2、web界面net/http/pprof

# 下载 cpu profile，默认从当前开始收集 30s 的 cpu 使用情况，需要等待 30s
go tool pprof http://47.93.238.9:8080/debug/pprof/profile
# wait 120s
go tool pprof http://47.93.238.9:8080/debug/pprof/profile?seconds=120

# 下载 heap profile
go tool pprof http://47.93.238.9:8080/debug/pprof/heap

# 下载 goroutine profile
go tool pprof http://47.93.238.9:8080/debug/pprof/goroutine

# 下载 block profile
go tool pprof http://47.93.238.9:8080/debug/pprof/block

# 下载 mutex profile
go tool pprof http://47.93.238.9:8080/debug/pprof/mutex

除了上面讲到的两种方式（报告生成、命令行交互），还可以在浏览器里进行交互。先生成 profile 文件，再执行命令：

go tool pprof --http=:8080 ~/Downloads/profile


https://www.qcrao.com/2019/11/10/dive-into-go-pprof/
