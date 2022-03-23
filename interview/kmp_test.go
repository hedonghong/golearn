package interview

import (
	"fmt"
	"testing"
)

/*
	m: 主串
	n: 子串
	kmp:
		1、BF：暴力算法，复杂度最好O(m+n)，最坏O(mn)，也是比较好理解
			m与n先对齐，然后一一字符对比，若出现不匹配，则m往后一位，再重复匹配
*/

func TestBF(t *testing.T) {
	m := "abcedeabcdesf"
	//m := "abcedeabcedesf"
	n := "abcd"
	lengthm := len(m)
	lengthn := len(n)
	//
	// abcedeabcdesf
	//          abcd
	// 13 - 4 = 9
	// 只要m剩下的字符串如果都不够n了，那就说明要退出了
	//
	end := lengthm - lengthn
	start := 0
	for end >= start {
		subM := m[start:(start + lengthn)]
		if subM == n {
			fmt.Println(start)
			break
		} else {
			start++
		}
	}
}

//BF算法：暴力匹配算法或朴素匹配算法，即从被查找的主串的起始位置开始，依次向前比对每个位置上的字符，
//以找出与要查找的目标模式串值和数量都匹配的字符串
//算法复杂度：O(n*m) 注：n为主串的长度，m为模式串长度
func BfStr(str, target string) int {
	n := len([]rune(str))
	m := len([]rune(target))
	j := 0
	i := 0
	start := -1
	for ; i < n; i++ {
		if j < m && str[i:i+1] != target[j:j+1] {
			j = 0
			continue
		} else {
			j++
			if j >= m {
				start = i - m + 1
				break
			}
		}
	}
	return start
}
