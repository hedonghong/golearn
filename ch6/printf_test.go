package ch6

import (
	"fmt"
	"strconv"
	"testing"
)

//go test -bench=BenchmarkItoa -benchmem
/*
sky.he@bawangbieji ch6 % go test -bench=BenchmarkName -benchmem
goos: darwin
goarch: amd64
pkg: golearn/ch6
BenchmarkName-4         12487339                93.0 ns/op            16 B/op          2 allocs/op(每次操作从堆分配内存次数，逃逸这个一定有)
PASS
ok      golearn/ch6     1.274s

*/
func BenchmarkName(b *testing.B) {
	num := 10
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		fmt.Sprintf("%d", num)
	}
}

func BenchmarkFormat(b *testing.B) {
	num := int64(10)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		strconv.FormatInt(num, 10)
	}
}

func BenchmarkItoa(b *testing.B) {
	num := 10
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		strconv.Itoa(num)
	}
}

func TestJiaoTi(t *testing.T) {
	s := []rune("abcdefghijklmnopqrstuvwxyz是")
	slen := len(s)
	last := fmt.Sprintf("%c", s[slen-1])
	g1 := make(chan bool)
	g2 := make(chan bool)
	out := make(chan bool)
	go func() {
		for i := 0; i < slen; i += 2 {
			if ok := <-g1; ok {
				c := fmt.Sprintf("%c", s[i])
				fmt.Println("协程1", c)
				if c == last {
					out <- true
					return
				}
				g2 <- true
			}
		}
	}()
	go func() {
		for i := 1; i < slen; i += 2 {
			if ok := <-g2; ok {
				c := fmt.Sprintf("%c", s[i])
				fmt.Println("协程2", c)
				if c == last {
					out <- true
					return
				}
				if i != 10 {
					g1 <- true
				}
			}
		}
	}()
	g1 <- true
	<-out
}
