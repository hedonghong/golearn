## select作用

1、channel非阻塞收发，随机执行

```go
ch1 := make(chan int, 2)
ch2 := make(chan int, 2)
for {
	select {
	    ...
	    case <- ch1:
		case ch1 <- xx:
        default:
    }
}
```

2、阻塞等待

```go
方式1：select {}

方式2：for {}
```

3、超时控制
若有需要请求第三方，但我方有返回时间要去，可以舍得超时channel使用select监听到底是超时还是请求返回。

## select结构

在 runtime/select.go

```go
// Select case descriptor.
// Known to compiler.
// Changes here must also be made in src/cmd/internal/gc/select.go's scasetype.
type scase struct {
	c           *hchan         // chan case中的channel指针
	elem        unsafe.Pointer // data element 读或者写的缓冲区地址
	kind        uint16 //case的类型
	pc          uintptr // race pc (for race detector / msan)
	releasetime int64
}

//kind可以下下面几种类型
kind = const (
    caseNil = iota
    caseRecv
    caseSend
    caseDefault
)
caseRecv：case语句中尝试读取scase.c中的数据，case <-Chan；
caseSend：case语句中尝试向scase.c中写入数据，case Chan <- Send；
caseDefault：default
```
reflect_rselect组装selectgo需要调用的参数
func reflect_rselect(cases []runtimeSelect) (int, bool) {

在一个select中，所有的case语句会构成一个scase结构体的数组，default是特殊的case。
接下来，执行select时，实际调用selectgo()函数
func selectgo(cas0 *scase, order0 *uint16, ncases int) (int, bool) {

cas0 为上文提到的case语句抽象出的结构体scase数组的第一个元素地址
order0 为一个两倍cas0数组长度的buffer，保存scase随机序列pollorder和scase中channel地址序列lockorder。
nncases 表示scase数组的长度

大概理下调用:
reflect/value.go(Select()) -> reflect/value.go(rselect())->runtime/select.go(selectgo())

核心是selectgo，先说下selectgo的核心思路：
1、打乱所有case
2、锁住所有case中的channel
3、遍历所有case中的channel是否有读写，若有解锁所有channel，并返回对应channel的数据；
若没有则分有没有default，有default则返回对应的scase，解锁所有channel；
若没有default则阻塞当前goroutine，把当前g加入到channel对应的等待队列中去，解锁所有channel，当前g等待被唤醒；
回忆下：hchan中的等待队列
recvq    waitq  // list of recv waiters //等待接收的 goroutine 队列
sendq    waitq  // list of send waiters //等待发送的 goroutine 队列
4、有channel可以读写，g被唤醒，对所有channel加锁。遍历找到可读写的channel和g，移除前一步加入到channel收发队列等待的g，返回scase值，解锁

##selectgo源码实现

```go

// selectgo implements the select statement.
//
// cas0 points to an array of type [ncases]scase, and order0 points to
// an array of type [2*ncases]uint16. Both reside on the goroutine's
// stack (regardless of any escaping in selectgo).
//
// selectgo returns the index of the chosen scase, which matches the
// ordinal position of its respective select{recv,send,default} call.
// Also, if the chosen scase was a receive operation, it reports whether
// a value was received.
func selectgo(cas0 *scase, order0 *uint16, ncases int) (int, bool) {
	if debugSelect {
		print("select: cas0=", cas0, "\n")
	}

	//使劲造，搞一个[1<<16]cas1的指针数组，一个两倍（[1 << 17])的指针数组order1
	cas1 := (*[1 << 16]scase)(unsafe.Pointer(cas0))
	order1 := (*[1 << 17]uint16)(unsafe.Pointer(order0))

	//截取获取0-ncases，容量为ncases的切片scases
	scases := cas1[:ncases:ncases]
	//同理
	pollorder := order1[:ncases:ncases]
	//截取order1后半段切片
	lockorder := order1[ncases:][:ncases:ncases]

	// Replace send/receive cases involving nil channels with
	// caseNil so logic below can assume non-nil channel.
	//scases.c为nil的设置一个空scases
	for i := range scases {
		cas := &scases[i]
		if cas.c == nil && cas.kind != caseDefault {
			*cas = scase{}
		}
	}

	var t0 int64
	if blockprofilerate > 0 {
		t0 = cputicks()
		for i := 0; i < ncases; i++ {
			scases[i].releasetime = -1
		}
	}

	// The compiler rewrites selects that statically have
	// only 0 or 1 cases plus default into simpler constructs.
	// The only way we can end up with such small sel.ncase
	// values here is for a larger select in which most channels
	// have been nilled out. The general code handles those
	// cases correctly, and they are rare enough not to bother
	// optimizing (and needing to test).

	// generate permuted order
	// 打乱，重新排序一下
	for i := 1; i < ncases; i++ {
		j := fastrandn(uint32(i + 1))
		pollorder[i] = pollorder[j]
		pollorder[j] = uint16(i)
	}

	// sort the cases by Hchan address to get the locking order.
	// simple heap sort, to guarantee n log n time and constant stack footprint.
	//根据c的地址进行堆排序，可以参考下orithm/sort/heap_test.go
	for i := 0; i < ncases; i++ {
		j := i
		// Start with the pollorder to permute cases on the same channel.
		c := scases[pollorder[i]].c
		for j > 0 && scases[lockorder[(j-1)/2]].c.sortkey() < c.sortkey() {
			k := (j - 1) / 2
			lockorder[j] = lockorder[k]
			j = k
		}
		lockorder[j] = pollorder[i]
	}
	for i := ncases - 1; i >= 0; i-- {
		o := lockorder[i]
		c := scases[o].c
		lockorder[i] = lockorder[0]
		j := 0
		for {
			k := j*2 + 1
			if k >= i {
				break
			}
			if k+1 < i && scases[lockorder[k]].c.sortkey() < scases[lockorder[k+1]].c.sortkey() {
				k++
			}
			if c.sortkey() < scases[lockorder[k]].c.sortkey() {
				lockorder[j] = lockorder[k]
				j = k
				continue
			}
			break
		}
		lockorder[j] = o
	}

	if debugSelect {
		for i := 0; i+1 < ncases; i++ {
			if scases[lockorder[i]].c.sortkey() > scases[lockorder[i+1]].c.sortkey() {
				print("i=", i, " x=", lockorder[i], " y=", lockorder[i+1], "\n")
				throw("select: broken sort")
			}
		}
	}

	// lock all the channels involved in the select
	//锁住所有的channel
	sellock(scases, lockorder)

	var (
		gp     *g
		sg     *sudog
		c      *hchan
		k      *scase
		sglist *sudog
		sgnext *sudog
		qp     unsafe.Pointer
		nextp  **sudog
	)

	//开始循环判断，查找可读写的channel
loop:
	// pass 1 - look for something already waiting
	var dfli int
	var dfl *scase
	var casi int
	var cas *scase
	var recvOK bool
	for i := 0; i < ncases; i++ {
		casi = int(pollorder[i])
		cas = &scases[casi]
		c = cas.c

		switch cas.kind {
		case caseNil:
			continue

		case caseRecv:
			//如果是接收方，检查发送队列是否为nil，所有有G在等待发送
			sg = c.sendq.dequeue()
			if sg != nil {
				//不为nil走recv
				goto recv
			}
			//缓冲区有元素，走bufrecv
			if c.qcount > 0 {
				goto bufrecv
			}
			//channel已经被关闭了，走rclose
			if c.closed != 0 {
				goto rclose
			}

		case caseSend:
			//同理
			if raceenabled {
				racereadpc(c.raceaddr(), cas.pc, chansendpc)
			}
			if c.closed != 0 {
				goto sclose
			}
			sg = c.recvq.dequeue()
			if sg != nil {
				goto send
			}
			//有空位，直接发
			if c.qcount < c.dataqsiz {
				goto bufsend
			}

		case caseDefault:
			dfli = casi
			dfl = cas
		}
	}

	//走default 解锁所有channel
	if dfl != nil {
		selunlock(scases, lockorder)
		casi = dfli
		cas = dfl
		goto retc
	}

	// pass 2 - enqueue on all chans
	//行吧没有default，又没有Channel可读写的
	//准备将当前g加入channel的等待队列了
	gp = getg()
	if gp.waiting != nil {
		throw("gp.waiting != nil")
	}
	nextp = &gp.waiting
	for _, casei := range lockorder {
		casi = int(casei)
		cas = &scases[casi]
		if cas.kind == caseNil {
			continue
		}
		c = cas.c
		sg := acquireSudog()
		sg.g = gp
		sg.isSelect = true
		// No stack splits between assigning elem and enqueuing
		// sg on gp.waiting where copystack can find it.
		sg.elem = cas.elem
		sg.releasetime = 0
		if t0 != 0 {
			sg.releasetime = -1
		}
		sg.c = c
		// Construct waiting list in lock order.
		*nextp = sg
		nextp = &sg.waitlink

		switch cas.kind {
		case caseRecv:
			c.recvq.enqueue(sg)

		case caseSend:
			c.sendq.enqueue(sg)
		}
	}

	// wait for someone to wake us up
	gp.param = nil
	// Signal to anyone trying to shrink our stack that we're about
	// to park on a channel. The window between when this G's status
	// changes and when we set gp.activeStackChans is not safe for
	// stack shrinking.
	atomic.Store8(&gp.parkingOnChan, 1)
	//阻塞当前G，等待被唤醒
	gopark(selparkcommit, nil, waitReasonSelect, traceEvGoBlockSelect, 1)
	//唤醒后的代码，
	gp.activeStackChans = false

	sellock(scases, lockorder)

	gp.selectDone = 0
	sg = (*sudog)(gp.param)
	gp.param = nil

	// pass 3 - dequeue from unsuccessful chans
	// otherwise they stack up on quiet channels
	// record the successful case, if any.
	// We singly-linked up the SudoGs in lock order.
	casi = -1
	cas = nil
	sglist = gp.waiting
	// Clear all elem before unlinking from gp.waiting.
	for sg1 := gp.waiting; sg1 != nil; sg1 = sg1.waitlink {
		sg1.isSelect = false
		sg1.elem = nil
		sg1.c = nil
	}
	gp.waiting = nil

	for _, casei := range lockorder {
		k = &scases[casei]
		if k.kind == caseNil {
			continue
		}
		if sglist.releasetime > 0 {
			k.releasetime = sglist.releasetime
		}
		if sg == sglist {
			// sg has already been dequeued by the G that woke us up.
			casi = int(casei)
			cas = k
		} else {
			c = k.c
			if k.kind == caseSend {
				c.sendq.dequeueSudoG(sglist)
			} else {
				c.recvq.dequeueSudoG(sglist)
			}
		}
		sgnext = sglist.waitlink
		sglist.waitlink = nil
		//解除所有的sudog
		releaseSudog(sglist)
		sglist = sgnext
	}

	//没有case，继续循环吧
	if cas == nil {
		// We can wake up with gp.param == nil (so cas == nil)
		// when a channel involved in the select has been closed.
		// It is easiest to loop and re-run the operation;
		// we'll see that it's now closed.
		// Maybe some day we can signal the close explicitly,
		// but we'd have to distinguish close-on-reader from close-on-writer.
		// It's easiest not to duplicate the code and just recheck above.
		// We know that something closed, and things never un-close,
		// so we won't block again.
		goto loop
	}

	c = cas.c

	if debugSelect {
		print("wait-return: cas0=", cas0, " c=", c, " cas=", cas, " kind=", cas.kind, "\n")
	}

	if cas.kind == caseRecv {
		recvOK = true
	}

	if raceenabled {
		if cas.kind == caseRecv && cas.elem != nil {
			raceWriteObjectPC(c.elemtype, cas.elem, cas.pc, chanrecvpc)
		} else if cas.kind == caseSend {
			raceReadObjectPC(c.elemtype, cas.elem, cas.pc, chansendpc)
		}
	}
	if msanenabled {
		if cas.kind == caseRecv && cas.elem != nil {
			msanwrite(cas.elem, c.elemtype.size)
		} else if cas.kind == caseSend {
			msanread(cas.elem, c.elemtype.size)
		}
	}

	selunlock(scases, lockorder)
	goto retc

bufrecv:
	// can receive from buffer
	if raceenabled {
		if cas.elem != nil {
			raceWriteObjectPC(c.elemtype, cas.elem, cas.pc, chanrecvpc)
		}
		raceacquire(chanbuf(c, c.recvx))
		racerelease(chanbuf(c, c.recvx))
	}
	if msanenabled && cas.elem != nil {
		msanwrite(cas.elem, c.elemtype.size)
	}
	recvOK = true
	qp = chanbuf(c, c.recvx)
	if cas.elem != nil {
		typedmemmove(c.elemtype, cas.elem, qp)
	}
	typedmemclr(c.elemtype, qp)
	c.recvx++
	if c.recvx == c.dataqsiz {
		c.recvx = 0
	}
	c.qcount--
	selunlock(scases, lockorder)
	goto retc

bufsend:
	// can send to buffer
	if raceenabled {
		raceacquire(chanbuf(c, c.sendx))
		racerelease(chanbuf(c, c.sendx))
		raceReadObjectPC(c.elemtype, cas.elem, cas.pc, chansendpc)
	}
	if msanenabled {
		msanread(cas.elem, c.elemtype.size)
	}
	typedmemmove(c.elemtype, chanbuf(c, c.sendx), cas.elem)
	c.sendx++
	if c.sendx == c.dataqsiz {
		c.sendx = 0
	}
	c.qcount++
	selunlock(scases, lockorder)
	goto retc

recv:
	// can receive from sleeping sender (sg)
	recv(c, sg, cas.elem, func() { selunlock(scases, lockorder) }, 2)
	if debugSelect {
		print("syncrecv: cas0=", cas0, " c=", c, "\n")
	}
	recvOK = true
	goto retc

rclose:
	// read at end of closed channel
	selunlock(scases, lockorder)
	recvOK = false
	if cas.elem != nil {
		typedmemclr(c.elemtype, cas.elem)
	}
	if raceenabled {
		raceacquire(c.raceaddr())
	}
	goto retc

send:
	// can send to a sleeping receiver (sg)
	if raceenabled {
		raceReadObjectPC(c.elemtype, cas.elem, cas.pc, chansendpc)
	}
	if msanenabled {
		msanread(cas.elem, c.elemtype.size)
	}
	send(c, sg, cas.elem, func() { selunlock(scases, lockorder) }, 2)
	if debugSelect {
		print("syncsend: cas0=", cas0, " c=", c, "\n")
	}
	goto retc

retc:
	if cas.releasetime > 0 {
		blockevent(cas.releasetime-t0, 1)
	}
	return casi, recvOK

sclose:
	// send on closed channel
	selunlock(scases, lockorder)
	panic(plainError("send on closed channel"))
}
```
