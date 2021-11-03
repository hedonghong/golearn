
## runtime mutex

[runtime mutex](https://colobu.com/2020/12/06/mutex-in-go-runtime/)

运行时的mutex数据结构很简单，如下所示，定义在runtime2.go中

```go
type mutex struct {}
```

对于dragonfly、freebsd、linux架构，mutex会使用基于Futex的实现， key就是一个uint32的值。 Linux提供的Futex(Fast user-space mutexes)用来构建用户空间的锁和信号量。Go 运行时封装了两个方法，用来sleep和唤醒当前线程：

src/runtime/lock_futex.go

futexsleep(addr uint32, val uint32, ns int64)：原子操作`if addr == val { sleep }`。
futexwakeup(addr *uint32, cnt uint32)：唤醒地址addr上的线程最多cnt次。
对于其他的架构，比如aix、darwin、netbsd、openbsd、plan9、solaris、windows，mutex会使用基于sema的实现，key就是M* waitm。Go 运行时封装了三个方法，用来创建信号量和sleep/wakeup：

src/runtime/lock_sema.go

func semacreate(mp *m):创建信号量
func semasleep(ns int64) int32： 请求信号量，请求不到会休眠一段时间
func semawakeup(mp *m)：唤醒mp
基于这两种实现，分别有不同的lock和unlock方法的实现，主要逻辑都是类似的，所以接下来我们只看基于Futex的lock/unlock。


