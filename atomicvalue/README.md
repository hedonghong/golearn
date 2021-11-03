## atomic.Value 原子操作
    https://blog.betacat.io/post/golang-atomic-value-exploration/
## 1、原子操作
    Mutex由操作系统实现，而atomic包中的原子操作则由底层硬件直接提供支持。在 CPU 实现的指令集里，有一些指令被封装进了atomic包，这些指令在执行的过程中是不允许中断（interrupt）的，因此原子操作可以在lock-free的情况下保证并发安全，并且它的性能也能做到随 CPU 个数的增多而线性扩展。
    就是说有可能一次赋值操作分了两次写入，写操作被优先级高的操作拆成两次

## 2、场景可用于并发环境下某变量等的原子操作

    atomic包中支持六种类型
    int32
    uint32
    int64
    uint64
    uintptr
    unsafe.Pointer

## 3、unsafe.Pointer 与 atomic.Value

    Go语言并不支持直接操作内存，但是它的标准库提供一种不保证向后兼容的指针类型unsafe.Pointer，
    让程序可以灵活的操作内存，它的特别之处在于：可以绕过Go语言类型系统的检查
    也就是说：如果两种类型具有相同的内存结构，我们可以将unsafe.Pointer当作桥梁，让这两种类型的指针相互转换，从而实现同一份内存拥有两种解读方式
    我们可以自定义的结构体进行更新，具体可以看sync.Map，里面的read map就是用了这部分的知识。

## 4、atomic.Value的Store() 写 与 Load () 读
    可以看下 atomicvalue_test.go

## 5、源码

```go
//一个普通的结构体，v可以保存任何类型的值
type Value struct {
	v interface{}
}

//一个空interface  type表示类型，data表示数据 unsafe.Pointer指针
type ifaceWords struct {
    typ  unsafe.Pointer
    data unsafe.Pointer
}

//写函数
// Store sets the value of the Value to x.
// All calls to Store for a given Value must use values of the same concrete type.
// Store of an inconsistent type panics, as does Store(nil).
func (v *Value) Store(x interface{}) {
    if x == nil {
        panic("sync/atomic: store of nil value into Value")
    }
    //将旧和新的都转为ifaceWords类型
    vp := (*ifaceWords)(unsafe.Pointer(v))//旧值
    xp := (*ifaceWords)(unsafe.Pointer(&x))//新值
    //循环乐观锁CompareAndSwapPointer
    for {
    	//获取当前值的类型
        typ := LoadPointer(&vp.typ)
        //第一次写入
        if typ == nil {
            // Attempt to start first store.
            // Disable preemption so that other goroutines can use
            // active spin wait to wait for completion; and so that
            // GC does not see the fake type accidentally.
        	//听说是让g死死占用着这个p
            runtime_procPin()
            //第一次写入判断是否类型为nil并且把nil值设置为^uintptr(0)的unsafe.Pointer
            //如果失败，则证明已经有别的线程抢先完成了赋值操作，那它就解除抢占锁，然后重新回到 for 循环第一步     //如果设置成功，那证明当前线程抢到了这个"乐观锁”，它可以安全的把v设为传入的新值了（// Complete first store.到return这段）。注意，这里是先写data字段，然后再写typ字段。因为我们是以typ字段的值作为写入完成与否的判断依据的
            if !CompareAndSwapPointer(&vp.typ, nil, unsafe.Pointer(^uintptr(0))) {
            	//g释放p
                runtime_procUnpin()
                continue
            }
            // Complete first store.
            //写入新值
            StorePointer(&vp.data, xp.data)
            StorePointer(&vp.typ, xp.typ)
            runtime_procUnpin()
            return
        }
        //第一次写入还未完成，如果看到 typ字段还是^uintptr(0)这个中间类型，证明刚刚的第一次写入还没有完成，所以它会继续循环，“忙等"到第一次写入完成。
        if uintptr(typ) == ^uintptr(0) {
            // First store in progress. Wait.
            // Since we disable preemption around the first store,
            // we can wait with active spinning.
            continue
        }
        //第一次写入已完成 - 首先检查上一次写入的类型与这一次要写入的类型是否一致，如果不一致则抛出异常。反之，则直接把这一次要写入的值写入到data字段。
        // First store completed. Check type and overwrite data.
        if typ != xp.typ {
            panic("sync/atomic: store of inconsistently typed value into Value")
        }
        StorePointer(&vp.data, xp.data)
        return
    }
}

这个逻辑的主要思想就是，为了完成多个字段的原子性写入，我们可以抓住其中的一个字段，以它的状态来标志整个原子写入的状态。这个想法我在 TiDB 的事务实现中看到过类似的，他们那边叫Percolator模型，主要思想也是先选出一个primaryRow，然后所有的操作也是以primaryRow的成功与否作为标志。

//读函数
// It returns nil if there has been no call to Store for this Value.
func (v *Value) Load() (x interface{}) {
    vp := (*ifaceWords)(unsafe.Pointer(v))
    typ := LoadPointer(&vp.typ)
    if typ == nil || uintptr(typ) == ^uintptr(0) {
        // First store not yet completed.
        return nil
    }
    data := LoadPointer(&vp.data)
    //往x变量的指针变量写入内容
    //当前看到的typ和data构造出一个新的interface{}返回出去
    xp := (*ifaceWords)(unsafe.Pointer(&x))
    xp.typ = typ
    xp.data = data
    return
}

//TODO 下面两个函数查资料好像是说 需要了解下才行
//其中runtime_procPin方法可以将一个goroutine死死占用当前使用的P
//不允许其他的goroutine抢占，而runtime_procUnpin则是释放方法
func runtime_procPin()
func runtime_procUnpin()
```
