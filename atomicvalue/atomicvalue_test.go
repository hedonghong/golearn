package atomicvalue

import (
	"fmt"
	"sync/atomic"
	"testing"
	"time"
)

func LoadConfig(s string) map[string]string {
	fmt.Println("开始获取配置处理", s)
	m := make(map[string]string)
	time.Sleep(1 * time.Second)
	/** 读取配置文件等处理*/
	m["file"] = ""
	fmt.Println("结束配置读取", s)
	return m
}

func TestAtomicValue(t *testing.T) {

	//简单的数据类型来些
	var b int64
	//并发下计数
	atomic.AddInt64(&b, 1)
	//并发下赋值
	atomic.StoreInt64(&b, 33)
	//并发下读数
	atomic.LoadInt64(&b)
	//并发下将新值保存，并返回旧值
	atomic.SwapInt64(&b, 44)
	//并发下比较修改 乐观锁 b若=0，则改为1 返回bool
	atomic.CompareAndSwapInt64(&b, 0, 1)

	//存入任何类型
	var a atomic.Value
	a.Store("xx")
	fmt.Println(a.Load())
	//并发场景下可用于对相同变量的进行原子操作

	//有时产生一个值，需要一系列指令处理，想要对这整个指令集同步，比如加载配置，这时就需要atomic.Value上场
	var con atomic.Value
	con.Store(LoadConfig(""))
	c := con.Load()
	fmt.Println(c)
}

func TestAtomicValue1(t *testing.T) {
	var con atomic.Value
	go con.Store(LoadConfig("协程A"))
	go con.Store(LoadConfig("协程B"))

	select {}
}
