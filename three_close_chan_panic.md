[TOC]

---

### 1、代码准备

```go
//close nil chan
package main

func main()  {
	var ch chan int
	close(ch)
	close(ch)
	//panic: close of nil channel
}

```

```go
//close close chan
package main

func main()  {
	ch := make(chan int)
	close(ch)
	close(ch)
	//panic: close of closed channel
}


```

```go
//close chan bu write
package main

func main()  {
	ch := make(chan int)
	close(ch)
	ch <- 1
	//panic: send on closed channel
}

```

### 2、找出panic处

#### 2.1 dlv debug close_nil_chan.go 输出运行栈

```go
Warning: debugging optimized function
	runtime.curg._panic.arg: interface {}(string) "close of nil channel"
  1184:	// fatalpanic implements an unrecoverable panic. It is like fatalthrow, except
  1185:	// that if msgs != nil, fatalpanic also prints panic messages and decrements
  1186:	// runningPanicDefers once main is blocked from exiting.
  1187:	//
  1188:	//go:nosplit
=>1189:	func fatalpanic(msgs *_panic) {
  1190:		pc := getcallerpc()
  1191:		sp := getcallersp()
  1192:		gp := getg()
  1193:		var docrash bool
  1194:		// Switch to the system stack to avoid any stack growth, which
(dlv) bt
0  0x000000000042e6e0 in runtime.fatalpanic
   at /usr/lib/golang/src/runtime/panic.go:1189
1  0x000000000042e13e in runtime.gopanic
   at /usr/lib/golang/src/runtime/panic.go:1064
2  0x0000000000404d66 in runtime.closechan
   at /usr/lib/golang/src/runtime/chan.go:342
3  0x00000000004601c3 in main.main

  上面在 runtime.closechan 处打断点，当经过两次的运行时出现painc，同样经过2.2的汇编代码查询

```

#### 2.2 go tool compile -N -l -S close_nil_chan.go

```go

0x001d 00029 (close_nil_chan.go:4)      MOVQ    $0, "".ch+8(SP)
0x0026 00038 (close_nil_chan.go:5)      MOVQ    $0, (SP)
0x002e 00046 (close_nil_chan.go:5)      CALL    runtime.closechan(SB)
0x0033 00051 (close_nil_chan.go:6)      PCDATA  $0, $1
0x0033 00051 (close_nil_chan.go:6)      PCDATA  $1, $0
0x0033 00051 (close_nil_chan.go:6)      MOVQ    "".ch+8(SP), AX
0x0038 00056 (close_nil_chan.go:6)      PCDATA  $0, $0
0x0038 00056 (close_nil_chan.go:6)      MOVQ    AX, (SP)
0x003c 00060 (close_nil_chan.go:6)      CALL    runtime.closechan(SB)

//针对关闭已经关闭的chan或者关闭nil chan 都可以看runtime.closechan(SB) 方法
//具体在runtime/chan.go:340的

func closechan(c *hchan) {
    if c == nil {
    panic(plainError("close of nil channel"))
    }
    
    lock(&c.lock)
    if c.closed != 0 {
    unlock(&c.lock)
    panic(plainError("close of closed channel"))
    }
    .....
}


//向关闭的chan进行写操作，我这次先从获取汇编操作开始，并且直接给runtime.chansend1()进行dlv断点调试
//汇编后
0x004e 00078 (close_chan_but_write.go:6)        MOVQ    AX, (SP)
0x0052 00082 (close_chan_but_write.go:6)        PCDATA  $0, $1
0x0052 00082 (close_chan_but_write.go:6)        LEAQ    ""..stmp_0(SB), AX
0x0059 00089 (close_chan_but_write.go:6)        PCDATA  $0, $0
0x0059 00089 (close_chan_but_write.go:6)        MOVQ    AX, 8(SP)
0x005e 00094 (close_chan_but_write.go:6)        CALL    runtime.chansend1(SB)

//dlv部分情况
> runtime.chansend1() /usr/lib/golang/src/runtime/chan.go:127 (PC: 0x40431c)
Warning: debugging optimized function
122:	}
123:
124:	// entry point for c <- x from compiled code
125:	//go:nosplit
126:	func chansend1(c *hchan, elem unsafe.Pointer) {
=> 127:		chansend(c, elem, true, getcallerpc())
128:	}
129:
130:	/*
   131:	 * generic single channel send/recv
   132:	 * If block is not nil,
(dlv)

可以知道chansend1底层调用了chansend方法

//继续

(dlv) n
> runtime.chansend() /usr/lib/golang/src/runtime/chan.go:143 (PC: 0x404361)
Warning: debugging optimized function
   138:	 * when a channel involved in the sleep has
   139:	 * been closed.  it is easiest to loop and re-run
   140:	 * the operation; we'll see that it's now closed.
   141:	 */
142:	func chansend(c *hchan, ep unsafe.Pointer, block bool, callerpc uintptr) bool {
=> 143:		if c == nil {
144:			if !block {
145:				return false
146:			}
147:			gopark(nil, nil, waitReasonChanSendNilChan, traceEvGoStop, 2)
148:			throw("unreachable")
(dlv) p c
*runtime.hchan {
qcount: 0,
dataqsiz: 0,
buf: unsafe.Pointer(0xc000056070),
elemsize: 8,
closed: 1,
elemtype: *runtime._type {size: 8, ptrdata: 0, hash: 4149441018, tflag: tflagUncommon|tflagExtraStar|tflagNamed|tflagRegularMemory (15), align: 8, fieldAlign: 8, kind: 2, equal: runtime.memequal64, gcdata: *1, str: 681, ptrToThis: 23296},
sendx: 0,
recvx: 0,
recvq: runtime.waitq {
first: *runtime.sudog nil,
last: *runtime.sudog nil,},
sendq: runtime.waitq {
first: *runtime.sudog nil,
last: *runtime.sudog nil,},
lock: runtime.mutex {key: 0},}
(dlv)

> runtime.chansend() /usr/lib/golang/src/runtime/chan.go:187 (PC: 0x404849)
Warning: debugging optimized function
182:
183:		lock(&c.lock)
184:
185:		if c.closed != 0 {
186:			unlock(&c.lock)
=> 187:			panic(plainError("send on closed channel"))
188:		}
189:
190:		if sg := c.recvq.dequeue(); sg != nil {
191:			// Found a waiting receiver. We pass the value we want to send
192:			// directly to the receiver, bypassing the channel buffer (if any).
(dlv)

//上面dlv 通过dlv的p c可以知道chan c.closed = 1 发送数据时会发生panic

//针对对已经关闭的chan进行send操作可以看runtime.chansend1(SB)
//具体在runtime/chan.go:126
//go:nosplit
func chansend1(c *hchan, elem unsafe.Pointer) {
chansend(c, elem, true, getcallerpc())
}

func chansend(c *hchan, ep unsafe.Pointer, block bool, callerpc uintptr) bool {
	....
		
    lock(&c.lock)
    if c.closed != 0 {
        unlock(&c.lock)
        panic(plainError("send on closed channel"))
    }
        ....
}

```