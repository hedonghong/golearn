## slice

runtime/slice.go

```go
//看看底层切片结构体
type slice struct {
	array unsafe.Pointer//指向数组指针
	len   int //长度
	cap   int //容量
}
很明显，切片的底层就是指向数组的，值得注意的是，由于array并不一定指向数组首位，还记得我们切片是可以通过数组截取获得的吗?
var intArr = [3]int{1, 2, 3}
var []intSlice = intArr[1:2]
```

先简单总结下数组截取切片的规律吧：

### 1、数组获取切片

arr[low:high:max]
len = high - low
cap = max - low

arr := [10]int{1,2,3,4,5,6,7,8,9,10}
s1:= arr[2:5:9] len:3 cap:7的切片
2是指下标，从0开始，5同理是下标
s1 -> {3,4,5}

s2:= s1[1:7] 没有指定max，则max=7; low = 1; high=7; len = 6; cap = 6
s2 -> {4,5,6,7,8,9}
没指定max则为原数组容量，不能超过原数组容量


### 2、空切片和nil切片的区别
先说下结论吧
1、nil切片和空切片指向的地址不一样。nil空切片引用数组指针地址为0（无指向任何实际地址）
2、空切片的引用数组指针地址是有的，且固定为一个值

var s1 []int //nil切片
s2 := make([]int,0)//空切片
s4 := make([]int,0)//空切片

fmt.Printf("s1 pointer:%+v, s2 pointer:%+v, s4 pointer:%+v, \n", *(*reflect.SliceHeader)(unsafe.Pointer(&s1)),*(*reflect.SliceHeader)(unsafe.Pointer(&s2)),*(*reflect.SliceHeader)(unsafe.Pointer(&s4)))
fmt.Printf("%v\n", (*(*reflect.SliceHeader)(unsafe.Pointer(&s1))).Data==(*(*reflect.SliceHeader)(unsafe.Pointer(&s2))).Data)
fmt.Printf("%v\n", (*(*reflect.SliceHeader)(unsafe.Pointer(&s2))).Data==(*(*reflect.SliceHeader)(unsafe.Pointer(&s4))).Data)
//s1 pointer:{Data:0 Len:0 Cap:0}, s2 pointer:{Data:824634830224 Len:0 Cap:0}, s4 pointer:{Data:824634830224 Len:0 Cap:0},
//false
//true
### 3、切片不会缩容

当我们用一个被扩展到很大的切片后，由于切片的扩展机制问题，我们切片不会缩容，所以只能通过切片拷贝的方式进行。

```go
    countries := []string{"USA", "Singapore", "Germany", "India", "Australia"}
    neededCountries := countries[:len(countries)-2]
    countriesCpy := make([]string, len(neededCountries))
    copy(countriesCpy, neededCountries) //copies neededCountries to countriesCpy
    
    //原来的countries和neededCountries都可以通过垃圾回收处理掉
```

## slice的底层实现

```go
//编译器会判断是逃逸，逃逸的话，会通过下面方法创建切片，看到最后会返回一个地址，主要是因为由调用方再合成一个切片结构体，下面方法只提供具有一定长度和容量的内存地址。
该函数主要工作是计算切片占用的内存空间并在堆上申请一片连续的内存，它使用如下的方式计算占用的内存：
内存空间=切片中元素大小×切片容量
func makeslice(et *_type, len, cap int) unsafe.Pointer {
	mem, overflow := math.MulUintptr(et.size, uintptr(cap))
	if overflow || mem > maxAlloc || len < 0 || len > cap {
		// NOTE: Produce a 'len out of range' error instead of a
		// 'cap out of range' error when someone does make([]T, bignumber).
		// 'cap out of range' is true too, but since the cap is only being
		// supplied implicitly, saying len is clearer.
		// See golang.org/issue/4085.
		mem, overflow := math.MulUintptr(et.size, uintptr(len))
		if overflow || mem > maxAlloc || len < 0 {
			panicmakeslicelen()
		}
		panicmakeslicecap()
	}

	return mallocgc(mem, et, true)
}
```

## slice 追加

```go
在cmd/compile/internal/gc/ssa.go
有一段解析，这里分两种情况一直是append(s, e1, e2, e3)，一种是s = append(s, e1, e2, e3)
// If inplace is false, process as expression "append(s, e1, e2, e3)":
//
// ptr, len, cap := s
// newlen := len + 3
// if newlen > cap {
//     ptr, len, cap = growslice(s, newlen)
//     newlen = len + 3 // recalculate to avoid a spill
// }
// // with write barriers, if needed:
// *(ptr+len) = e1
// *(ptr+len+1) = e2
// *(ptr+len+2) = e3
// return makeslice(ptr, newlen, cap)
//
//
// If inplace is true, process as statement "s = append(s, e1, e2, e3)":
//
// a := &s
// ptr, len, cap := s
// newlen := len + 3
// if uint(newlen) > uint(cap) {
//    newptr, len, newcap = growslice(ptr, len, cap, newlen)
//    vardef(a)       // if necessary, advise liveness we are writing a new a
//    *a.cap = newcap // write before ptr to avoid a spill
//    *a.ptr = newptr // with write barrier
// }
// newlen = len + 3 // recalculate to avoid a spill
// *a.len = newlen
// // with write barriers, if needed:
// *(ptr+len) = e1
// *(ptr+len+1) = e2
// *(ptr+len+2) = e3
```

## slice扩容

简单说下结论：

1、如果期望容量大于当前容量的两倍就会使用期望容量；

2、如果当前切片的长度小于 1024 就会将容量翻倍；

3、如果当前切片的长度大于 1024 就会每次增加 25% 的容量，直到新容量大于期望容量；

但这样的说法也不是很正确，咋们看下源码
```go
// runtime/slice.go
growslice 在追加期间处理切片增长。 它传递切片元素类型、旧切片和所需的新最小容量，并返回一个至少具有该容量的新切片，并将旧数据复制到其中。 新切片的长度设置为旧切片的长度，而不是新请求的容量。 这是为了方便代码生成。 旧切片的长度立即用于计算追加期间写入新值的位置。 TODO：当旧的后端消失时，重新考虑这个决定。 SSA 后端可能更喜欢新的长度或仅返回 ptr/cap 并节省堆栈空间
// growslice handles slice growth during append.
// It is passed the slice element type, the old slice, and the desired new minimum capacity,
// and it returns a new slice with at least that capacity, with the old data
// copied into it.
// The new slice's length is set to the old slice's length,
// NOT to the new requested capacity.
// This is for codegen convenience. The old slice's length is used immediately
// to calculate where to write new values during an append.
// TODO: When the old backend is gone, reconsider this decision.
// The SSA backend might prefer the new length or to return only ptr/cap and save stack space.

//old 旧slice   cap新cap
func growslice(et *_type, old slice, cap int) slice {
	if raceenabled {
		callerpc := getcallerpc()
		racereadrangepc(old.array, uintptr(old.len*int(et.size)), callerpc, funcPC(growslice))
	}
	if msanenabled {
		msanread(old.array, uintptr(old.len*int(et.size)))
	}

	if cap < old.cap {
		panic(errorString("growslice: cap out of range"))
	}

	//空切片
	if et.size == 0 {
		// append should not create a slice with nil pointer but non-zero len.
		// We assume that append doesn't need to preserve old.array in this case.
		return slice{unsafe.Pointer(&zerobase), old.len, cap}
	}

	newcap := old.cap
	//先初始化两倍原来的容量
	doublecap := newcap + newcap
	//如果要求的容量还要大于两倍
	if cap > doublecap {
		//1、如果期望容量大于当前容量的两倍就会使用期望容量；
		newcap = cap
	} else {
		//2、如果当前切片的长度小于 1024 就会将容量翻倍；
		if old.len < 1024 {
			newcap = doublecap
		} else {
			//3、如果当前切片的长度大于 1024 就会每次增加 25% 的容量，直到新容量大于期望容量；
			// Check 0 < newcap to detect overflow
			// and prevent an infinite loop.
			for 0 < newcap && newcap < cap {
				newcap += newcap / 4
			}
			// Set newcap to the requested cap when
			// the newcap calculation overflowed.
			if newcap <= 0 {
				newcap = cap
			}
		}
	}

	//举例s切片原来长度和cap都是2，若新增三个元素，那么根据上面newcap=5，那个走下面逻辑
	//传入roundupsize的参数就是5*8 = 40
	//上面规律不是很正确的地方来了
	//由于有内存对齐
	//上面逻辑确定切片的大致容量，下面还需要根据切片中的元素大小对齐内存，当数组中元素所占的字节大小为 1、8 或者 2 的倍数时，运行时会使用如下代码对齐内存，最终才会得出容量，其实一般都会比上面等于或者大点点
	var overflow bool
	var lenmem, newlenmem, capmem uintptr
	// Specialize for common values of et.size.
	// For 1 we don't need any division/multiplication.
	// For sys.PtrSize, compiler will optimize division/multiplication into a shift by a constant.
	// For powers of 2, use a variable shift.
	switch {
	case et.size == 1:
		lenmem = uintptr(old.len)
		newlenmem = uintptr(cap)
		capmem = roundupsize(uintptr(newcap))
		overflow = uintptr(newcap) > maxAlloc
		newcap = int(capmem)
	case et.size == sys.PtrSize//代码中ptrSize是指一个指针的大小，在64位机上是8
		lenmem = uintptr(old.len) * sys.PtrSize
		newlenmem = uintptr(cap) * sys.PtrSize
		capmem = roundupsize(uintptr(newcap) * sys.PtrSize)
		overflow = uintptr(newcap) > maxAlloc/sys.PtrSize
		newcap = int(capmem / sys.PtrSize)
	case isPowerOfTwo(et.size):
		var shift uintptr
		if sys.PtrSize == 8 {
			// Mask shift for better code generation.
			shift = uintptr(sys.Ctz64(uint64(et.size))) & 63
		} else {
			shift = uintptr(sys.Ctz32(uint32(et.size))) & 31
		}
		lenmem = uintptr(old.len) << shift
		newlenmem = uintptr(cap) << shift
		capmem = roundupsize(uintptr(newcap) << shift)
		overflow = uintptr(newcap) > (maxAlloc >> shift)
		newcap = int(capmem >> shift)
	default:
		lenmem = uintptr(old.len) * et.size
		newlenmem = uintptr(cap) * et.size
		capmem, overflow = math.MulUintptr(et.size, uintptr(newcap))
		capmem = roundupsize(capmem)
		newcap = int(capmem / et.size)
	}

	// The check of overflow in addition to capmem > maxAlloc is needed
	// to prevent an overflow which can be used to trigger a segfault
	// on 32bit architectures with this example program:
	//
	// type T [1<<27 + 1]int64
	//
	// var d T
	// var s []T
	//
	// func main() {
	//   s = append(s, d, d, d, d)
	//   print(len(s), "\n")
	// }
	if overflow || capmem > maxAlloc {
		panic(errorString("growslice: cap out of range"))
	}

	//创建新空间，拷贝数据过去
	var p unsafe.Pointer
	if et.ptrdata == 0 {
		p = mallocgc(capmem, nil, false)
		// The append() that calls growslice is going to overwrite from old.len to cap (which will be the new length).
		// Only clear the part that will not be overwritten.
		memclrNoHeapPointers(add(p, newlenmem), capmem-newlenmem)
	} else {
		// Note: can't use rawmem (which avoids zeroing of memory), because then GC can scan uninitialized memory.
		p = mallocgc(capmem, et, true)
		if lenmem > 0 && writeBarrier.enabled {
			// Only shade the pointers in old.array since we know the destination slice p
			// only contains nil pointers because it has been cleared during alloc.
			bulkBarrierPreWriteSrcOnly(uintptr(p), uintptr(old.array), lenmem)
		}
	}
	memmove(p, old.array, lenmem)

	return slice{p, old.len, newcap}
}

// Returns size of the memory block that mallocgc will allocate if you ask for the size.
// 总需要内存容量size
//举例s切片原来长度和cap都是2，若新增三个元素，那么根据上面newcap=5，那个走下面逻辑
//传入roundupsize的参数就是5*8 = 40
const _MaxSmallSize = 32768
const smallSizeMax = 1024
const smallSizeDiv = 8
func roundupsize(size uintptr) uintptr {
    if size < _MaxSmallSize {
        if size <= smallSizeMax-8 {
        	//按照上面例子，，我们走这里
        	// size = 40， (size+smallSizeDiv-1)/smallSizeDiv = (40 + 8 - 1)/ 8 = 47/8 = 5
        	//class_to_size[size_to_class8[5]]
        	//class_to_size[4]
        	//最终是48
        	//回到上一个函数就是48/8 = 6，那么容量就是6
            return uintptr(class_to_size[size_to_class8[(size+smallSizeDiv-1)/smallSizeDiv]])
        } else {
            return uintptr(class_to_size[size_to_class128[(size-smallSizeMax+largeSizeDiv-1)/largeSizeDiv]])
        }
    }
    if size+_PageSize < size {
        return size
    }
    return alignUp(size, _PageSize)
}

// alignUp rounds n up to a multiple of a. a must be a power of 2.
func alignUp(n, a uintptr) uintptr {
    return (n + a - 1) &^ (a - 1)
}
```

