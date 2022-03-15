介绍
gops是一个命令行工具，配合agent，可以用来很方便的诊断运行中的go程序，使用简单，官方维护

安装
go get github.com/google/gops
// 或者
go install github.com/google/gops@latest

go get github.com/google/gops
// 或者
go install github.com/google/gops@latest

使用
gops本身可以查看所有go程序的进程，如果一个程序使用了agent，gops可以报告更多的信息，比如stack，内存统计，trace等，使用了agent的程序会用*号标记

使用agent
go func() {
cfg := agent.Options{
Addr:            ":2022", //远程调试使用，绑定对应的进程pid
ShutdownCleanup: true,
}
if err := agent.Listen(cfg); err != nil {
panic(err)
}
}()
gops支持远程诊断，只需要将换成Host:Port即可

列出所有go进程
root@2kpBjdU4m:~# gops
PID    PPID   Name                     Version  Location

52222  850    docker-proxy             go1.13.8 /usr/bin/docker-proxy
162404 162302 gopls                    go1.17.5 /root/go/bin/gopls
169840 169738 gopls                    go1.17.5 /root/go/bin/gopls
171257 162544 web                    * go1.17.5 /home/2kpBjdU4m/hub/gen-server/web
171289 170774 gops                     go1.17.5 /root/go/bin/gops
850    1      dockerd                  go1.13.8 /usr/bin/dockerd
查看进程详情
root@2kpBjdU4m:~# gops 171257
parent PID:	162544
threads:	9
memory usage:	0.266%
cpu usage:	99.498%
username:	root
cmd+args:	./web
elapsed time:	02:18
local/remote:	172.19.0.1:44384 <-> 172.19.0.12:3306 (ESTABLISHED)
local/remote:	172.19.0.1:48492 <-> 172.19.0.7:27017 (ESTABLISHED)
local/remote:	172.19.0.1:45504 <-> 172.19.0.5:9200 (ESTABLISHED)
local/remote:	172.19.0.1:48494 <-> 172.19.0.7:27017 (ESTABLISHED)
local/remote:	172.19.0.1:48490 <-> 172.19.0.7:27017 (ESTABLISHED)
local/remote:	:::2022 <-> :::0 (LISTEN)
可以指定收集时间
root@2kpBjdU4m:~# gops 172381 10s
parent PID:	162544
threads:	9
memory usage:	0.277%
cpu usage:	99.678%
cpu usage (10s):	100.400%
username:	root
cmd+args:	./web
elapsed time:	03:06
local/remote:	172.19.0.1:48600 <-> 172.19.0.7:27017 (ESTABLISHED)
local/remote:	172.19.0.1:48602 <-> 172.19.0.7:27017 (ESTABLISHED)
local/remote:	172.19.0.1:45614 <-> 172.19.0.5:9200 (ESTABLISHED)
local/remote:	172.19.0.1:44496 <-> 172.19.0.12:3306 (ESTABLISHED)
local/remote:	172.19.0.1:48604 <-> 172.19.0.7:27017 (ESTABLISHED)
local/remote:	:::2022 <-> :::0 (LISTEN)
local/remote:	:::2023 <-> :::0 (LISTEN)
查看stack信息
root@2kpBjdU4m:~# gops stack 172381
goroutine 85 [running]:
runtime/pprof.writeGoroutineStacks({0x1033f60, 0xc00050e030})
/usr/local/go/src/runtime/pprof/pprof.go:693 +0x70
runtime/pprof.writeGoroutine({0x1033f60, 0xc00050e030}, 0x0)
/usr/local/go/src/runtime/pprof/pprof.go:682 +0x2b
runtime/pprof.(*Profile).WriteTo(0xe88671, {0x1033f60, 0xc00050e030}, 0x0)
/usr/local/go/src/runtime/pprof/pprof.go:331 +0x14b
github.com/google/gops/agent.handle({0x7fb5e4678718, 0xc00050e030}, {0xc0000382a0, 0x9, 0xc0000a3fc0})
/root/go/pkg/mod/github.com/google/gops@v0.3.22/agent/agent.go:201 +0x15d
github.com/google/gops/agent.listen()
/root/go/pkg/mod/github.com/google/gops@v0.3.22/agent/agent.go:145 +0x19a
created by github.com/google/gops/agent.Listen
/root/go/pkg/mod/github.com/google/gops@v0.3.22/agent/agent.go:123 +0x365

goroutine 1 [chan receive, 4 minutes]:
main.main()
/home/2kpBjdU4m/hub/gen-server/cmd/web/main.go:108 +0x82b
查看内存统计
root@2kpBjdU4m:~# gops memstats 172381
alloc: 2.61MB (2740864 bytes)
total-alloc: 5.63MB (5904904 bytes)
sys: 15.08MB (15811592 bytes)
lookups: 0
mallocs: 38142
frees: 21689
heap-alloc: 2.61MB (2740864 bytes)
heap-sys: 7.38MB (7733248 bytes)
heap-idle: 2.99MB (3137536 bytes)
heap-in-use: 4.38MB (4595712 bytes)
heap-released: 2.47MB (2588672 bytes)
heap-objects: 16453
stack-in-use: 640.00KB (655360 bytes)
stack-sys: 640.00KB (655360 bytes)
stack-mspan-inuse: 70.92KB (72624 bytes)
stack-mspan-sys: 80.00KB (81920 bytes)
stack-mcache-inuse: 4.69KB (4800 bytes)
stack-mcache-sys: 16.00KB (16384 bytes)
other-sys: 935.69KB (958148 bytes)
gc-sys: 4.69MB (4917416 bytes)
next-gc: when heap-alloc >= 4.00MB (4194304 bytes)
last-gc: 2022-02-28 17:54:14.439966835 +0800 CST
gc-pause-total: 476.578µs
gc-pause: 111034
gc-pause-end: 1646042054439966835
num-gc: 3
num-forced-gc: 0
gc-cpu-fraction: 7.774838758506689e-06
enable-gc: true
debug-gc: false
查看runtime stats
root@2kpBjdU4m:~# gops stats 172381
goroutines: 27
OS threads: 11
GOMAXPROCS: 4
num CPU: 4
查看trace
gops允许你收集5s runtime tracer，然后提供浏览器访问

root@2kpBjdU4m:~# gops trace 172381
Tracing now, will take 5 secs...
Trace dump saved to: /tmp/trace1196940716
2022/02/28 17:57:30 Parsing trace...
2022/02/28 17:57:30 Splitting trace...
2022/02/28 17:57:30 Opening browser. Trace viewer is listening on http://127.0.0.1:35507
强制gc垃圾收集
运行gops gc立即执行垃圾回收，程序会被阻塞，直到gc完成

pprof实时交互
gops支持内存和cpu pprof分析，在收集数据之后，调用go tool pprof，进入交互样本分析界面

root@2kpBjdU4m:~# gops pprof-heap 172381
Profile dump saved to: /tmp/heap_profile2317160970
Binary file saved to: /tmp/binary3943624338
File: binary3943624338
Type: inuse_space
Time: Feb 28, 2022 at 6:08pm (CST)
Entering interactive mode (type "help" for commands, "o" for options)
(pprof) top
Showing nodes accounting for 3620.55kB, 100% of 3620.55kB total
Showing top 10 nodes out of 40
flat  flat%   sum%        cum   cum%
1025kB 28.31% 28.31%     1025kB 28.31%  runtime.allocm
544.67kB 15.04% 43.35%   544.67kB 15.04%  github.com/xdg-go/stringprep.init
514.63kB 14.21% 57.57%   514.63kB 14.21%  regexp.makeOnePass.func1
512.20kB 14.15% 71.72%   512.20kB 14.15%  runtime.malg
512.04kB 14.14% 85.86%   512.04kB 14.14%  github.com/segmentio/kafka-go/protocol.structDecodeFuncOf.func1.1
512.02kB 14.14%   100%   512.02kB 14.14%  github.com/segmentio/kafka-go/protocol.structEncodeFuncOf
0     0%   100%   514.63kB 14.21%  github.com/go-playground/validator/v10.init
0     0%   100%  1024.05kB 28.28%  github.com/segmentio/kafka-go/protocol.Register
0     0%   100%   512.04kB 14.14%  github.com/segmentio/kafka-go/protocol.decodeFuncOf
0     0%   100%   512.02kB 14.14%  github.com/segmentio/kafka-go/protocol.encodeFuncOf
(pprof) 

更多
[官方文档：](https://github.com/google/gops)