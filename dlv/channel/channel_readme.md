
## 基于 go version go1.14.12 darwin/amd64 源码分析
### 1、channel底层数据结构

```go

type hchan struct {
	qcount   uint // total data in the queue 队列中元素总数
	dataqsiz uint // size of the circular queue 缓冲队列元素的允许放置的个数
	buf      unsafe.Pointer // points to an array of dataqsiz elements 缓冲队列，循环数组
	elemsize uint16 //队列元素的大小
	closed   uint32 //是否已经关闭 0否 1是
	elemtype *_type // element type 元素的类型
	sendx    uint   // send index 已发送元素在循环数组中的索引
	recvx    uint   // receive index 已接收元素在循环数组中的索引
	recvq    waitq  // list of recv waiters //等待接收的 goroutine 队列
	sendq    waitq  // list of send waiters //等待发送的 goroutine 队列
	lock mutex //并发读写锁
}

//waitq是 sudog 的一个双向链表，而 sudog 实际上是对 goroutine 的一个封装（晚点再介绍）


type waitq struct {
    first *sudog
    last  *sudog
}


//创建chan，返回hchan的指针
func makechan(t *chantype, size int) *hchan {
    elem := t.elem
    
    // compiler checks this but be safe.
    if elem.size >= 1<<16 {
        throw("makechan: invalid channel element type")
    }
    if hchanSize%maxAlign != 0 || elem.align > maxAlign {
        throw("makechan: bad alignment")
    }
    
    //mem 元素大小*个数
    mem, overflow := math.MulUintptr(elem.size, uintptr(size))
    if overflow || mem > maxAlloc-hchanSize || size < 0 {
        panic(plainError("makechan: size out of range"))
    }
    
    // Hchan does not contain pointers interesting for GC when elements stored in buf do not contain pointers.
    // buf points into the same allocation, elemtype is persistent.
    // SudoG's are referenced from their owning thread so they can't be collected.
    // TODO(dvyukov,rlh): Rethink when collector can move allocated objects.
    var c *hchan
    switch {
    //大小等于 0的元素类型：struct{}
    case mem == 0:
    // Queue or element size is zero.
    //不存在缓冲区，那么就只会为 runtime.hchan 分配一段内存空间
    c = (*hchan)(mallocgc(hchanSize, nil, true))
    // Race detector uses this location for synchronization.
    c.buf = c.raceaddr()
    case elem.ptrdata == 0:
    // Elements do not contain pointers.
    // Allocate hchan and buf in one call.
    //如果元素类型不含指针 或者 size 大小为 0（无缓冲类型）
    //分配 "hchan 结构体大小 + 元素大小*个数" 的内存
    //会为当前的 Channel 和底层的数组分配一块连续的内存空间
    c = (*hchan)(mallocgc(hchanSize+mem, nil, true))
    c.buf = add(unsafe.Pointer(c), hchanSize)
    default:
    // Elements contain pointers.
    //为 runtime.hchan 和缓冲区分配内存 这个不一定连续了哈
    c = new(hchan)
    c.buf = mallocgc(mem, elem, true)
    }
    
    c.elemsize = uint16(elem.size)//元素大小
    c.elemtype = elem//原因类型
    c.dataqsiz = uint(size)//缓冲大小
    
    if debugChan {
        print("makechan: chan=", c, "; elemsize=", elem.size, "; dataqsiz=", size, "\n")
    }
    return c
}


func chansend(c *hchan, ep unsafe.Pointer, block bool, callerpc uintptr) bool {
    if c == nil {
        if !block {
        return false
        }
        //永远阻塞住
        gopark(nil, nil, waitReasonChanSendNilChan, traceEvGoStop, 2)
        throw("unreachable")
    }
    
    if debugChan {
    print("chansend: chan=", c, "\n")
    }

    if raceenabled {
        racereadpc(c.raceaddr(), callerpc, funcPC(chansend))
    }

// Fast path: check for failed non-blocking operation without acquiring the lock.
//
// After observing that the channel is not closed, we observe that the channel is
// not ready for sending. Each of these observations is a single word-sized read
// (first c.closed and second c.recvq.first or c.qcount depending on kind of channel).
// Because a closed channel cannot transition from 'ready for sending' to
// 'not ready for sending', even if the channel is closed between the two observations,
// they imply a moment between the two when the channel was both not yet closed
// and not ready for sending. We behave as if we observed the channel at that moment,
// and report that the send cannot proceed.
//
// It is okay if the reads are reordered here: if we observe that the channel is not
// ready for sending and then observe that it is not closed, that implies that the
// channel wasn't closed during the first observation.
    //如果是非阻塞的，并且没有关闭channel，但是缓冲为0又没有接受着，又或者缓冲大于0，但是满了，直接返回吧
    if !block && c.closed == 0 && ((c.dataqsiz == 0 && c.recvq.first == nil) ||
        (c.dataqsiz > 0 && c.qcount == c.dataqsiz)) {
        return false
    }

    var t0 int64
    if blockprofilerate > 0 {
        t0 = cputicks()
    }

    //读写都是原子性的
    lock(&c.lock)

    if c.closed != 0 {
        unlock(&c.lock)
        panic(plainError("send on closed channel"))
    }

    //接收队列不为空说明存储读等待了，发调用send拷贝数据
    if sg := c.recvq.dequeue(); sg != nil {
        // Found a waiting receiver. We pass the value we want to send
        // directly to the receiver, bypassing the channel buffer (if any).
        send(c, sg, ep, func() { unlock(&c.lock) }, 3)
        return true
    }

    //如果队列中数据小于允许缓冲大小，没满，那就塞进去
    if c.qcount < c.dataqsiz {
        // Space is available in the channel buffer. Enqueue the element to send.
    	//获取下一个buf存储的位置
        qp := chanbuf(c, c.sendx)
        if raceenabled {
            raceacquire(qp)
            racerelease(qp)
        }
        //将发送的数据拷贝到缓冲区中并增加 sendx 索引和 qcount 计数器
        typedmemmove(c.elemtype, qp, ep)
        c.sendx++
        if c.sendx == c.dataqsiz {
            c.sendx = 0
        }
        c.qcount++
        unlock(&c.lock)
        return true
    }

    if !block {
        unlock(&c.lock)
        return false
    }

    // Block on the channel. Some receiver will complete our operation for us.
    //没有接收，也满了，行吧
    //Channel 阻塞地发送数据会执行下面的代码
    gp := getg()
    mysg := acquireSudog()
    mysg.releasetime = 0
    if t0 != 0 {
        mysg.releasetime = -1
    }
    // No stack splits between assigning elem and enqueuing mysg
    // on gp.waiting where copystack can find it.
    mysg.elem = ep
    mysg.waitlink = nil
    mysg.g = gp
    mysg.isSelect = false
    mysg.c = c
    gp.waiting = mysg
    gp.param = nil
    //发送g封装成sudg结构入到等待发送的g队列中
    c.sendq.enqueue(mysg)
    // Signal to anyone trying to shrink our stack that we're about
    // to park on a channel. The window between when this G's status
    // changes and when we set gp.activeStackChans is not safe for
    // stack shrinking.
    atomic.Store8(&gp.parkingOnChan, 1)
    //挂起g，等待唤醒
    gopark(chanparkcommit, unsafe.Pointer(&c.lock), waitReasonChanSend, traceEvGoBlockSend, 2)
    // Ensure the value being sent is kept alive until the
    // receiver copies it out. The sudog has a pointer to the
    // stack object, but sudogs aren't considered as roots of the
    // stack tracer.
    //这后面是被唤醒之后需要做一些释放工作
    KeepAlive(ep)
    
    // someone woke us up.
    if mysg != gp.waiting {
        throw("G waiting list is corrupted")
    }
    gp.waiting = nil
    gp.activeStackChans = false
    if gp.param == nil {
        if c.closed == 0 {
            throw("chansend: spurious wakeup")
        }
        panic(plainError("send on closed channel"))
    }
    gp.param = nil
    if mysg.releasetime > 0 {
        blockevent(mysg.releasetime-t0, 2)
    }
    mysg.c = nil
    releaseSudog(mysg)
    return true
}

func send(c *hchan, sg *sudog, ep unsafe.Pointer, unlockf func(), skip int) {
    if raceenabled {
        if c.dataqsiz == 0 {
            racesync(c, sg)
        } else {
            // Pretend we go through the buffer, even though
            // we copy directly. Note that we need to increment
            // the head/tail locations only when raceenabled.
            qp := chanbuf(c, c.recvx)
            raceacquire(qp)
            racerelease(qp)
            raceacquireg(sg.g, qp)
            racereleaseg(sg.g, qp)
            c.recvx++
            if c.recvx == c.dataqsiz {
                c.recvx = 0
            }
            c.sendx = c.recvx // c.sendx = (c.sendx+1) % c.dataqsiz
        }
    }
    if sg.elem != nil {
    	//拷贝元素到接受者
        sendDirect(c.elemtype, sg, ep)
        sg.elem = nil
    }
    gp := sg.g
    unlockf()
    gp.param = unsafe.Pointer(sg)
    if sg.releasetime != 0 {
        sg.releasetime = cputicks()
    }
    //唤醒接受者
    goready(gp, skip+1)
}

//从一个有缓冲的 channel 里读数据，当 channel 被关闭，依然能读出有效值。只有当返回的 ok 为 false 时，读出的数据才是无效的
func chanrecv(c *hchan, ep unsafe.Pointer, block bool) (selected, received bool) {
    // raceenabled: don't need to check ep, as it is always on the stack
    // or is new memory allocated by reflect.
    
    if debugChan {
        print("chanrecv: chan=", c, "\n")
    }
    
    if c == nil {
        if !block {
            return
        }
        //chan为nil 永远阻塞
        gopark(nil, nil, waitReasonChanReceiveNilChan, traceEvGoStop, 2)
        throw("unreachable")
    }
    
    // Fast path: check for failed non-blocking operation without acquiring the lock.
    //
    // After observing that the channel is not ready for receiving, we observe that the
    // channel is not closed. Each of these observations is a single word-sized read
    // (first c.sendq.first or c.qcount, and second c.closed).
    // Because a channel cannot be reopened, the later observation of the channel
    // being not closed implies that it was also not closed at the moment of the
    // first observation. We behave as if we observed the channel at that moment
    // and report that the receive cannot proceed.
    //
    // The order of operations is important here: reversing the operations can lead to
    // incorrect behavior when racing with a close.
    //非阻塞，但不能读取，直接返回
    if !block && (c.dataqsiz == 0 && c.sendq.first == nil ||
        c.dataqsiz > 0 && atomic.Loaduint(&c.qcount) == 0) &&
        atomic.Load(&c.closed) == 0 {
        return
    }
    
    var t0 int64
    if blockprofilerate > 0 {
        t0 = cputicks()
    }
    
    lock(&c.lock)
    
    //chan 已经关闭，并且没有元素可再读，清楚ep指针对应数据，返回
    if c.closed != 0 && c.qcount == 0 {
        if raceenabled {
            raceacquire(c.raceaddr())
        }
        unlock(&c.lock)
        if ep != nil {
            typedmemclr(c.elemtype, ep)
        }
        return true, false
    }
    
    //发送队列不为空，调用recv读数据
    if sg := c.sendq.dequeue(); sg != nil {
        // Found a waiting sender. If buffer is size 0, receive value
        // directly from sender. Otherwise, receive from head of queue
        // and add sender's value to the tail of the queue (both map to
        // the same buffer slot because the queue is full).
        recv(c, sg, ep, func() { unlock(&c.lock) }, 3)
        return true, true
    }
    
    //队列里面元素不为空，直接读buff
    if c.qcount > 0 {
        // Receive directly from queue
        qp := chanbuf(c, c.recvx)
        if raceenabled {
        raceacquire(qp)
        racerelease(qp)
        }
        if ep != nil {
            typedmemmove(c.elemtype, ep, qp)
        }
        typedmemclr(c.elemtype, qp)
        c.recvx++
        if c.recvx == c.dataqsiz {
            c.recvx = 0
        }
        c.qcount--
        unlock(&c.lock)
        return true, true
    }
    
    if !block {
        unlock(&c.lock)
        return false, false
    }
    
    // no sender available: block on this channel.
    //没有数据可读，放入sudog，挂起，等待唤醒
    gp := getg()
    mysg := acquireSudog()
    mysg.releasetime = 0
    if t0 != 0 {
        mysg.releasetime = -1
    }
    // No stack splits between assigning elem and enqueuing mysg
    // on gp.waiting where copystack can find it.
    mysg.elem = ep
    mysg.waitlink = nil
    gp.waiting = mysg
    mysg.g = gp
    mysg.isSelect = false
    mysg.c = c
    gp.param = nil
    c.recvq.enqueue(mysg)
    // Signal to anyone trying to shrink our stack that we're about
    // to park on a channel. The window between when this G's status
    // changes and when we set gp.activeStackChans is not safe for
    // stack shrinking.
    atomic.Store8(&gp.parkingOnChan, 1)
    gopark(chanparkcommit, unsafe.Pointer(&c.lock), waitReasonChanReceive, traceEvGoBlockRecv, 2)
    
    // someone woke us up
    if mysg != gp.waiting {
        throw("G waiting list is corrupted")
    }
    gp.waiting = nil
    gp.activeStackChans = false
    if mysg.releasetime > 0 {
        blockevent(mysg.releasetime-t0, 2)
    }
    closed := gp.param == nil
    gp.param = nil
    mysg.c = nil
    releaseSudog(mysg)
    return true, !closed
}


func closechan(c *hchan) {
    if c == nil {
        panic(plainError("close of nil channel"))
    }
    
    lock(&c.lock)
    if c.closed != 0 {
        unlock(&c.lock)
        panic(plainError("close of closed channel"))
    }
    
    if raceenabled {
        callerpc := getcallerpc()
        racewritepc(c.raceaddr(), callerpc, funcPC(closechan))
        racerelease(c.raceaddr())
    }
    
    c.closed = 1
    
    var glist gList
    
    // release all readers
    //找出所有接受者，赋予一个该类型的默认值，并把对应的G放入glist中后面唤醒
    //sudog释放
    for {
        sg := c.recvq.dequeue()
        if sg == nil {
            break
        }
        if sg.elem != nil {
            typedmemclr(c.elemtype, sg.elem)
            sg.elem = nil
        }
        if sg.releasetime != 0 {
            sg.releasetime = cputicks()
        }
        gp := sg.g
        gp.param = nil
        if raceenabled {
            raceacquireg(gp, c.raceaddr())
        }
            glist.push(gp)
        }
    
    // release all writers (they will panic)
	//找出所有发送者，并把对应的G放入glist中后面唤醒，唤醒后会发现channel已经关闭，报painc
	//sudog释放
    for {
        sg := c.sendq.dequeue()
        if sg == nil {
            break
        }
        sg.elem = nil
        if sg.releasetime != 0 {
            sg.releasetime = cputicks()
        }
        gp := sg.g
        gp.param = nil
        if raceenabled {
            raceacquireg(gp, c.raceaddr())
        }
        glist.push(gp)
    }
    unlock(&c.lock)
    
    // Ready all Gs now that we've dropped the channel lock.
    //遍历唤醒所有接受者发送者，该干嘛干嘛，该报错报错
    for !glist.empty() {
        gp := glist.pop()
        gp.schedlink = 0
        goready(gp, 3)
    }
}
```

2、channel应用场景

```go
停止信号
任务定时
解耦生产方和消费方
控制并发数
```

3、channel如何优雅关闭

```go
有两个不那么优雅地关闭 channel 的方法：

1、defer-recover 机制，defer-recover 在兜底。

2、sync.Once 来保证只关闭一次。

那到底应该如何优雅地关闭 channel？

根据 sender 和 receiver 的个数，分下面几种情况：

一个 sender，一个 receiver
一个 sender， M 个 receiver
N 个 sender，一个 reciver
N 个 sender， M 个 receiver
1、对于 1，2，只有一个 sender 的情况就不用说了，直接从 sender 端关闭就好了，没有问题。重点关注第 3，4 种情况。

2、3情况增加一个传递关闭信号的 channel，receiver 通过信号 channel 下达关闭数据 channel 指令。senders 监听到关闭信号后，停止接收数据。比如直接关闭信号队列，发送端用select 接收关闭信号即可

3、4情况需要增加一个中间人，M 个 receiver 都向它发送关闭 dataCh 的“请求”，中间人收到第一个请求后，就会直接下达关闭 dataCh 的指令（通过关闭 stopCh，这时就不会发生重复关闭的情况，因为 stopCh 的发送方只有中间人一个）。另外，这里的 N 个 sender 也可以向中间人发送关闭 dataCh 的请求。

发生 panic 的情况有三种：向一个关闭的 channel 进行写操作；关闭一个 nil 的 channel；重复关闭一个 channel

func main() {
    rand.Seed(time.Now().UnixNano())
    const Max = 100000
    const NumReceivers = 10
    const NumSenders = 1000
    dataCh := make(chan int, 100)
    stopCh := make(chan struct{})
    // It must be a buffered channel.
    toStop := make(chan string, 1)
    var stoppedBy string
    // moderator
    go func() {
        stoppedBy = <-toStop
        close(stopCh)
    }()
    // senders
    for i := 0; i < NumSenders; i++ {
        go func(id string) {
        for {
        value := rand.Intn(Max)
        if value == 0 {
        select {
        case toStop <- "sender#" + id:
        default:
        }
        return
        }
        select {
        case <- stopCh:
        return
        case dataCh <- value:
        }
        }
        }(strconv.Itoa(i))
    }
    // receivers
    for i := 0; i < NumReceivers; i++ {
        go func(id string) {
        for {
            select {
                case <- stopCh:
                    return
                case value := <-dataCh:
                if value == Max-1 {
                    select {
                    case toStop <- "receiver#" + id:
                    default:
                    }
                    return
                }
                fmt.Println(value)
            }
        }
        }(strconv.Itoa(i))
    }
    select {
        case <- time.After(time.Hour):
    }
}
```