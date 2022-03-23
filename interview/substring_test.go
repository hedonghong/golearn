package interview

import (
	"fmt"
	"testing"
)

// 截取字符串

func TestSubString1(t *testing.T) {
	s := "abcdef"
	fmt.Println(s[1:4])
	s1 := "Go 语言"
	fmt.Println(s1[1:4])
	// 上面很明显对中文不友好，应该是说对多字节文字不友好
	s2 := "Go 语言"
	rs := []rune(s2)
	fmt.Println(string(rs[1:4]))
	// 使用rune 处理中文问题
}
