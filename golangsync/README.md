# 官方扩展包学习

[golang/sync](https://github.com/golang/sync)

[官方标准库](https://studygolang.com/pkgdoc)

[热心市民标准库库学习笔记](https://books.studygolang.com/The-Golang-Standard-Library-by-Example/)

## 1、errgroup

### 1.1 解决问题场景：多个goroutine并发操作，当一个遇到问题的时候，退出
    通过errgroup中TestWaitGroup与TestErrgroup运行对比，明显前者有个等待的时间，后者是非常飞快地输出错误了。
    goroutine中有有可能访问出错

### 1.2 定义：errgroup是为了处理一组任务的自任务的goroutine组，提供同步，错误传递和上下文取消。

### 1.3 源码

```go
//一个总任务的子任务的goroutine的集合
type Group struct {
	cancel func()//context中的cancel
	wg sync.WaitGroup//多个任务调度
	errOnce sync.Once//每个只执行一次
	err     error
}

//返回一个从ctx派生的context的新的group
//当函数第一次传递给go时，派生的context被取消，返回的非零错误或第一次返回
//wait返回
func WithContext(ctx context.Context) (*Group, context.Context) {
    ctx, cancel := context.WithCancel(ctx)
    return &Group{cancel: cancel}, ctx
}

//利用了sync.WaitGroup的wait
func (g *Group) Wait() error {
    g.wg.Wait()
    //不为nil就取消goroutine返回错误
    if g.cancel != nil {
        g.cancel()
    }
    return g.err
}

//Go在一个新的goroutine里面调用给的function
//第一次调用返回的一个非nil的error，它的错误就将会被返回在wait
func (g *Group) Go(f func() error) {
    g.wg.Add(1)
    go func() {
    defer g.wg.Done()
        //执行用户给的function，有错误，则执行一次错误赋值
        //并且运行上下文的取消函数，取消所有goroutine的运行
        if err := f(); err != nil {
            g.errOnce.Do(func() {
                g.err = err
                if g.cancel != nil {
                    g.cancel()
                }
            })
        }
    }()
}

```

### 1.4 更多例子，可以阅读源码中官方测试例子

### 1.5 kratos之sync/errgroup 提供带recover和并行数的errgroup， err包含详细的错误堆栈信息

    https://pkg.go.dev/github.com/bilibili/kratos/pkg/sync/errgroup
    https://pkg.go.dev/github.com/go-kratos/kratos
    https://github.com/go-kratos/kratos/tree/v1.0.x/pkg/sync/errgroup

## 2、semaphore
https://mp.weixin.qq.com/s/JC14dWffHub0nfPlPipsHQ
### 2.1 解决问题场景：防止goroutine启动过多，造成泄漏
    可以看下semaphore文件夹中的例子，其中例子抄录自互联网

### 2.2 定义：信号标，一个同步对象，用于保持计算小于最大计算值。即内部保持一个变量，释放为+，为0的需要等待，以计数的方式实现并发量的控制，比如goroutine的启动数

### 2.3 源码

```go
type Weighted struct {
     size    int64 // 设置一个最大权值
     cur     int64 // 标识当前已被使用的资源数
     mu      sync.Mutex // 互斥锁，提供临界区保护
     waiters list.List // 阻塞等待的调用者列表，使用链表数据结构保证先进先出的顺序，存储的数据是waiter对象
}

type waiter struct {
    n     int64 // 等待调用者权重值
    ready chan<- struct{} //这就是一个channel，利用channel的close机制实现唤醒，close channel就是唤醒
}

// NewWeighted为并发访问创建一个新的加权信号量，该信号量具有给定的最大权值。
func NewWeighted(n int64) *Weighted {
    w := &Weighted{size: n}
    return w
}

//阻塞获取权值的方法 - Acquire
func (s *Weighted) Acquire(ctx context.Context, n int64) error {
    s.mu.Lock() // 加锁保护临界区
    // 有资源可用并且没有等待获取权值的goroutine
    if s.size-s.cur >= n && s.waiters.Len() == 0 {
        s.cur += n // 加权
        s.mu.Unlock() // 释放锁
        return nil
    }
    // 要获取的权值n大于最大的权值了
    if n > s.size {
        // 先释放锁，确保其他goroutine调用Acquire的地方不被阻塞
        s.mu.Unlock()
        // 阻塞等待context的返回
        <-ctx.Done()
        return ctx.Err()
    }
    // 走到这里就说明现在没有资源可用了
    // 创建一个channel用来做通知唤醒
    ready := make(chan struct{})
    // 创建waiter对象
    w := waiter{n: n, ready: ready}
    // waiter按顺序入队
    elem := s.waiters.PushBack(w)
    // 释放锁，等待唤醒，别阻塞其他goroutine
    s.mu.Unlock()
    
    //下面是属于再次尝试的代码
    // 阻塞等待唤醒
    select {
    // context关闭
    case <-ctx.Done():
        err := ctx.Err() // 先获取context的错误信息
        s.mu.Lock()
        select {
            case <-ready:
                // 在context被关闭后被唤醒了，那么试图修复队列，假装我们没有取消
                err = nil
            default:
            // 判断是否是第一个元素
            isFront := s.waiters.Front() == elem
            // 移除第一个元素
            s.waiters.Remove(elem)
            // 如果是第一个元素且有资源可用通知其他waiter
            if isFront && s.size > s.cur {
                s.notifyWaiters()
            }
        }
        s.mu.Unlock()
        return err
    // 被唤醒了
    case <-ready:
        return nil
    }
}
流程一：有资源可用时并且没有等待权值的goroutine，走正常加权流程；

流程二：想要获取的权值n大于初始化时设置最大的权值了，这个goroutine永远不会获取到信号量，所以阻塞等待context的关闭；

流程三：前两步都没问题的话，就说明现在系统没有资源可用了，这时就需要阻塞等待唤醒，在阻塞等待唤醒这里有特殊逻辑；

//不阻塞获取权值的方法 - TryAcquire
func (s *Weighted) TryAcquire(n int64) bool {
    s.mu.Lock() // 加锁
    // 有资源可用并且没有等待获取资源的goroutine
    success := s.size-s.cur >= n && s.waiters.Len() == 0
    if success {
        s.cur += n
    }
    s.mu.Unlock()
    return success
}

//释放权重
func (s *Weighted) Release(n int64) {
    s.mu.Lock()
    // 释放资源
    s.cur -= n
    // 释放资源大于持有的资源，则会发生panic
    if s.cur < 0 {
        s.mu.Unlock()
        panic("semaphore: released more than held")
    }
    // 通知其他等待的调用者
    s.notifyWaiters()
    s.mu.Unlock()
}

//唤醒waiter，在Acquire和Release方法中都调用了notifyWaiters
func (s *Weighted) notifyWaiters() {
    for {
        // 获取等待调用者队列中的队员
        next := s.waiters.Front()
        // 没有要通知的调用者了
        if next == nil {
            break // No more waiters blocked.
        }
    
        // 断言出waiter信息
        w := next.Value.(waiter)
        if s.size-s.cur < w.n {
            // 没有足够资源为下一个调用者使用时，继续阻塞该调用者，遵循先进先出的原则，
            // 避免需要资源数比较大的waiter被饿死
            //
            // 考虑一个场景，使用信号量作为读写锁，现有N个令牌，N个reader和一个writer
            // 每个reader都可以通过Acquire（1）获取读锁，writer写入可以通过Acquire（N）获得写锁定
            // 但不包括所有的reader，如果我们允许reader在队列中前进，writer将会饿死-总是有一个令牌可供每个reader
            break
        }
    
        // 获取资源
        s.cur += w.n
        // 从waiter列表中移除
        s.waiters.Remove(next)
        // 使用channel的close机制唤醒waiter
        close(w.ready)
    }
}

```

## 3、singleflight

### 3.1 解决问题场景：防止缓存失效，瞬间并发请求数据库等导致数据库压力多大。只有允许一个有效请求。防止缓存击穿。

### 3.2 具体请看golearn/golangsync/singleflight中测试例子

### 3.3 源码

```go

//一类需要防止并发的任务分组
type Group struct {
	mu sync.Mutex       // 互斥锁，并发控制
	m  map[string]*call // 所要调用的事件相关信息，比如多次调用，使用传入key作为map唯一健，不用初始化，后面DO()方法会初始化，使用懒加载
}

type call struct {
    wg sync.WaitGroup
    // 函数的返回值，在 wg 返回前只会写入一次
    val interface{}
    err error
    // 使用调用了 Forgot 方法
    forgotten bool
    // 统计调用次数以及返回的 channel
    dups  int
    chans []chan<- Result
}

func (g *Group) Do(key string, fn func() (interface{}, error)) (v interface{}, err error, shared bool) {
    g.mu.Lock()
    
    // 前面提到的懒加载
    if g.m == nil {
        g.m = make(map[string]*call)
    }
    
    // 会先去看 key 是否已经存在
    if c, ok := g.m[key]; ok {
        // 如果存在就会解锁
        c.dups++
        g.mu.Unlock()
        
        // 然后等待 WaitGroup 执行完毕，只要一执行完，所有的 wait 都会被唤醒
        c.wg.Wait()
        
        // 这里区分 panic 错误和 runtime 的错误，避免出现死锁，后面可以看到为什么这么做
        if e, ok := c.err.(*panicError); ok {
            panic(e)
        } else if c.err == errGoexit {
            runtime.Goexit()
        }
        return c.val, c.err, true
    }
    
    // 如果我们没有找到这个 key 就 new call
    c := new(call)
    
    // 然后调用 waitgroup 这里只有第一次调用会 add 1，其他的都会调用 wait 阻塞掉
    // 所以这要这次调用返回，所有阻塞的调用都会被唤醒
    c.wg.Add(1)
    g.m[key] = c
    g.mu.Unlock()
    
    // 然后我们调用 doCall 去执行
    g.doCall(c, key, fn)
    return c.val, c.err, c.dups > 0
}

//使用了两个 defer 巧妙的将 runtime 的错误和我们传入 function 的 panic 区别开来避免了由于传入的 function panic 导致的死锁
func (g *Group) doCall(c *call, key string, fn func() (interface{}, error)) {
    normalReturn := false
    recovered := false
    // 第一个 defer 检查 runtime 错误
    defer func() {
        // c.err错误搭上errGoexit标识
        if !normalReturn && !recovered {
            c.err = errGoexit
        }
        
        c.wg.Done()
        g.mu.Lock()
        defer g.mu.Unlock()
        if !c.forgotten {
            delete(g.m, key)
        }
        
        if e, ok := c.err.(*panicError); ok {
            //为了防止等待channel永远被阻塞，需要保证这个panic无法恢复
            if len(c.chans) > 0 {
                go panic(e)
                select {}//让goroutine阻塞住
            } else {
                panic(e)
            }
        } else if c.err == errGoexit {
            // Already in the process of goexit, no need to call again
        } else {
            // Normal return
            for _, ch := range c.chans {
                ch <- Result{c.val, c.err, c.dups > 0}
            }
        }
    }()

    // 使用一个匿名函数来执行
    func() {
        defer func() {
            if !normalReturn {
                // 如果 panic 了我们就 recover 掉，然后 new 一个 panic 的错误
                // 后面在上层重新 panic
                if r := recover(); r != nil {
                    c.err = newPanicError(r)
                }
            }
        }()
        
        c.val, c.err = fn()
        
        // 如果 fn 没有 panic 就会执行到这一步，如果 panic 了就不会执行到这一步
        // 所以可以通过这个变量来判断是否 panic 了
        normalReturn = true
    }()

    // 如果 normalReturn 为 false 就表示，我们的 fn panic 了
    // 如果执行到了这一步，也说明我们的 fn  recover 住了，不是直接 runtime exit
    if !normalReturn {
        recovered = true
    }
}

//异步Do 主要实现上就是，如果调用 DoChan 会给 call.chans 添加一个 channel 这样等第一次调用执行完毕之后就会循环向这些 channel 写入数据
func (g *Group) DoChan(key string, fn func() (interface{}, error)) <-chan Result {
    ch := make(chan Result, 1)
    g.mu.Lock()
    if g.m == nil {
        g.m = make(map[string]*call)
    }
    if c, ok := g.m[key]; ok {
        c.dups++
        c.chans = append(c.chans, ch)
        g.mu.Unlock()
        return ch
    }
    c := &call{chans: []chan<- Result{ch}}
    c.wg.Add(1)
    g.m[key] = c
    g.mu.Unlock()
    
    go g.doCall(c, key, fn)
    
    return ch
}

//forget 用于手动释放某个 key 下次调用就不会阻塞等待了
func (g *Group) Forget(key string) {
    g.mu.Lock()
    if c, ok := g.m[key]; ok {
        c.forgotten = true
    }
    delete(g.m, key)
    g.mu.Unlock()
}

```

### 3.4 

```go
1. 一个阻塞，全员等待
使用 singleflight 我们比较常见的是直接使用 Do 方法，但是这个极端情况下会导致整个程序 hang 住，如果我们的代码出点问题，有一个调用 hang 住了，那么会导致所有的请求都 hang 住

还是之前的例子，我们加一个 select 模拟阻塞

func singleflightGetArticle(sg *singleflight.Group, id int) (string, error) {
	v, err, _ := sg.Do(fmt.Sprintf("%d", id), func() (interface{}, error) {
		// 模拟出现问题，hang 住
		select {}
		return getArticle(id)
	})

	return v.(string), err
}
执行就会发现死锁了

fatal error: all goroutines are asleep - deadlock!

goroutine 1 [select (no cases)]:
这时候我们可以使用 DoChan 结合 select 做超时控制

func singleflightGetArticle(ctx context.Context, sg *singleflight.Group, id int) (string, error) {
	result := sg.DoChan(fmt.Sprintf("%d", id), func() (interface{}, error) {
		// 模拟出现问题，hang 住
		select {}
		return getArticle(id)
	})

	select {
	case r := <-result:
		return r.Val.(string), r.Err
	case <-ctx.Done():
		return "", ctx.Err()
	}
}
调用的时候传入一个含 超时的 context 即可，执行时就会返回超时错误

❯ go run ./1.go
panic: context deadline exceeded
2. 一个出错，全部出错
这个本身不是什么问题，因为 singleflight 就是这么设计的，但是实际使用的时候 如果我们一次调用要 1s，我们的数据库请求或者是 下游服务可以支撑 10rps 的请求的时候这会导致我们的错误阈提高，因为实际上我们可以一秒内尝试 10 次，但是用了 singleflight 之后只能尝试一次，只要出错这段时间内的所有请求都会受影响

这种情况我们可以启动一个 Goroutine 定时 forget 一下，相当于将 rps 从 1rps 提高到了 10rps

go func() {
       time.Sleep(100 * time.Millisecond)
       // logging
       g.Forget(key)
   }()
```

## 4、syncmap

### 4.1 解决问题：golang 在多goroutine 环境下并发写map会产生panic，可以说是"线程不安全"
    官方针对map上述问题推出syncmap，当然也可以自己封装结构体加互斥锁的，不过就是效率有点问题，那么syncmap怎么解决效率问题？可以在后面源码解读部分看到答案【读写分离】

### 4.2 源码，最好有对atomic.Value 、unsafe.Pointer 知识点了解，源码中用到了，否则看到一脸懵

```go

//实现了读写分离的map read map 和 dirty map
type Map struct {
   mu Mutex //就像你看到的，一把锁
   read atomic.Value // 是一个atomic.Value 存储的是 readOnly 结构体，利用 atomic.Value提供原子操作读写 map 【一个读map】
   dirty map[interface{}]*entry // 【一个写map】
   misses int
}

type readOnly struct {
    m       map[interface{}]*entry//普通的一个map
    amended bool // 是否存在dirty map中有，m中没有的？true 是
}

type entry struct {
    p unsafe.Pointer // *interface{} 一个指针，存的是readOnly的指针
}

//根据key加载一个值，源码英语备注我不去掉，可以结合中文理解
func (m *Map) Load(key interface{}) (value interface{}, ok bool) {
	//从atomic.Value加载中加载读map
    read, _ := m.read.Load().(readOnly)
    e, ok := read.m[key]
    //不存在，并且发现存在dirty map中有，m中没有的
    //amended 表明在dirty map中被删除
    if !ok && read.amended {
        m.mu.Lock()
        // Avoid reporting a spurious miss if m.dirty got promoted while we were
        // blocked on m.mu. (If further loads of the same key will not miss, it's
        // not worth copying the dirty map for this key.)
        //加锁成功后，再次进行检查，防止并发情况
        read, _ = m.read.Load().(readOnly)
        e, ok = read.m[key]
        if !ok && read.amended {
        	//从dirty map读下出来
            e, ok = m.dirty[key]
            // Regardless of whether the entry was present, record a miss: this key
            // will take the slow path until the dirty map is promoted to the read
            // map.
            //后面介绍，主要是把dirty map 拷贝到 read map中
            m.missLocked()
        }
        m.mu.Unlock()
    }
    if !ok {
        return nil, false
    }
    //从atomic.Value读取存储的e.p的值
    return e.load()
}

//若从read map查找失败的次数大于 dirty map的个数时，从dirty map 拷贝到 read map中
func (m *Map) missLocked() {
    m.misses++
    if m.misses < len(m.dirty) {
        return
    }
    m.read.Store(readOnly{m: m.dirty})
    m.dirty = nil
    m.misses = 0
}
//atomic.Value读操作，读取之前存入的有效值
func (e *entry) load() (value interface{}, ok bool) {
    p := atomic.LoadPointer(&e.p)
    //是为nil或者是被打上删除标识的了，那就返回nil吧，expunged是删除的标识
    if p == nil || p == expunged {
        return nil, false
    }
    //先转为*interface{}指针，再*从指针里面读取值，原理可以看下atomic.Value篇
    return *(*interface{})(p), true
}

//看到这里，如果Load方法都看懂了，那存入的方法也很简单了
func (m *Map) Store(key, value interface{}) {
	//一样，从read map中尝试读取下，有些情况是软删除的
	//tryStore操作是如果存在，不是被软删除就进行更新了
	//提个问题，在哪里打上删除标识的？怎么样才会打上？看下去吧
    read, _ := m.read.Load().(readOnly)
    if e, ok := read.m[key]; ok && e.tryStore(&value) {
        return
    }
    //运行到这里说明不存在于read map或者key被标识为删除
    m.mu.Lock()
    //加锁再检查，很正常的双重检验
    read, _ = m.read.Load().(readOnly)
    if e, ok := read.m[key]; ok {
    	//确保read map中的key不被标识为已删除
        if e.unexpungeLocked() {
            // The entry was previously expunged, which implies that there is a
            // non-nil dirty map and this entry is not in it.
        	//行吧，存入一下到dirty map
            m.dirty[key] = e
        }
        //atomic.Value写操作
        e.storeLocked(&value)
    } else if e, ok := m.dirty[key]; ok {
        //atomic.Value写操作，可以说是更新一下
        e.storeLocked(&value)
    } else {
    	//都不存在与read map 和 dirty map，并且不存在dirty中有,read没有的
        if !read.amended {
            // We're adding the first new key to the dirty map.
            // Make sure it is allocated and mark the read-only map as incomplete.
        	//存入dirty map之前确保read中的amended 标识为true，表示有存在dirty，read map不存在，用于读
            m.dirtyLocked()//从read map 拷贝到 dirty map，如果发现
            m.read.Store(readOnly{m: read.m, amended: true})
        }
        //存入dirty map
        m.dirty[key] = newEntry(value)
    }
    m.mu.Unlock()
}

//从read map 拷贝到 dirty map
func (m *Map) dirtyLocked() {
    if m.dirty != nil {
        return
    }
    //这部需要结合后面的删除函数联想下
    read, _ := m.read.Load().(readOnly)
    m.dirty = make(map[interface{}]*entry, len(read.m))
    for k, e := range read.m {
        if !e.tryExpungeLocked() {
            m.dirty[k] = e
        }
    }
}

//检查是否已经删除，是nil给e.p打上删除标识
func (e *entry) tryExpungeLocked() (isExpunged bool) {
    p := atomic.LoadPointer(&e.p)
    for p == nil {
        if atomic.CompareAndSwapPointer(&e.p, nil, expunged) {
            return true
        }
        p = atomic.LoadPointer(&e.p)
    }
    return p == expunged
}

// tryStore stores a value if the entry has not been expunged.
// 尝试更新
// If the entry is expunged, tryStore returns false and leaves the entry
// unchanged.
func (e *entry) tryStore(i *interface{}) bool {
	//循环？因为是用了CompareAndSwapPointer【了解下自旋锁】，所以循环
    for {
        p := atomic.LoadPointer(&e.p)
        if p == expunged {
            return false
        }
        if atomic.CompareAndSwapPointer(&e.p, p, unsafe.Pointer(i)) {
            return true
        }
    }
}

// unexpungeLocked ensures that the entry is not marked as expunged.
// 确保read map中key不被标识为已删除
// If the entry was previously expunged, it must be added to the dirty map
// before m.mu is unlocked.
func (e *entry) unexpungeLocked() (wasExpunged bool) {
    return atomic.CompareAndSwapPointer(&e.p, expunged, nil)
}

//存在返回，不存在保存方法操作大同小异，读懂上面都差不多
func LoadOrStore
//存在并且没有被软删除则更新并且返回方法操作大同小异，读懂上面都差不多
func tryLoadOrStore

//删除操作，先删除dirty 再设置e.p为nil，最终在store方法中存在给e.p搭上expunged删除标识的可能
// Delete deletes the value for a key.
func (m *Map) Delete(key interface{}) {
    read, _ := m.read.Load().(readOnly)
    e, ok := read.m[key]
    if !ok && read.amended {
        m.mu.Lock()
        read, _ = m.read.Load().(readOnly)
        e, ok = read.m[key]
        if !ok && read.amended {
        	//read map中没有，并且发现dity map存在read map中没有的
        	//删除一下dity map
            delete(m.dirty, key)
        }
        m.mu.Unlock()
    }
    if ok {
    	//真正删除e.p中的值，设置为nil
        e.delete()
    }
}

//如果read map中存储并且没有标识为已经删除，那设置下为nil
func (e *entry) delete() (hadValue bool) {
    for {
    	//如果
        p := atomic.LoadPointer(&e.p)
        if p == nil || p == expunged {
            return false
        }
        //真正删除e.p中的值，设置为nil
        if atomic.CompareAndSwapPointer(&e.p, p, nil) {
            return true
        }
    }
}

//迭代方法，从read map中循环，但存在amended为true的情况，
//这个时候需要上互斥锁，把dirty map 覆盖给 read map 清空dirty map，设置查询miss次数为misses=0
func Range

```