package interview

import (
	"fmt"
	"testing"
)

//请实现⼀个算法，在不使⽤【额外数据结构和储存空间】的情况下，翻转⼀个给定的字
//符串(可以使⽤单个过程变量)。
//给定⼀个string，请返回⼀个string，为翻转后的字符串。保证字符串的⻓度⼩于等于
//5000。

func ReverseString(s string) (string, bool) {
	strLen := len(s)
	if strLen < 0 || strLen > 5000 {
		return s, false
	}
	rs := []rune(s)
	for i := 0; i < strLen/2; i++ {
		rs[i], rs[strLen-1-i] = rs[strLen-1-i], rs[i]
	}
	return string(rs), true
}

func TestReverseString(t *testing.T) {
	s := "jsdlw2"
	fmt.Println(ReverseString(s))
}