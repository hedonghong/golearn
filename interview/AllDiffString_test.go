package interview

import (
	"fmt"
	"strings"
	"testing"
	"unicode"
)

//请实现⼀个算法，确定⼀个字符串的所有字符【是否全都不同】。这⾥我们要求【不允
//许使⽤额外的存储结构】。 给定⼀个string，请返回⼀个bool值,true代表所有字符全都
//不同，false代表存在相同的字符。 保证字符串中的字符为【ASCII字符】。字符串的⻓
//度⼩于等于【3000】。
func isUniqueString(s string) bool {
	if strings.Count(s, "") > 3000 {
		return false
	}

	for _, v := range s {
		if v > 127 {
			return false
		}
		if strings.Count(s, string(v)) > 1 {
			return false
		}
	}
	return true
}

func isUniqueString1(s string) bool {
	if strings.Count(s, "") > 3000 {
		return false
	}

	for k, v := range s {
		if v > 127 {
			return false
		}
		if strings.Index(s, string(v)) != k {
			return false
		}
	}
	return true
}

func TestAllDiffString(t *testing.T) {
	s := "AKSSJ"
	s1 := "AKOSDL"
	fmt.Println(isUniqueString(s))
	fmt.Println(isUniqueString1(s1))
}

//给定两个字符串，请编写程序，确定其中⼀个字符串的字符重新排列后，能否变成另⼀
//个字符串。 这⾥规定【⼤⼩写为不同字符】，且考虑字符串重点空格。给定⼀个string s1和⼀个string s2，请返回⼀个bool，代表两串是否重新排列后可相同。 保证两串的
//⻓度都⼩于等于5000。

func isRegroup(s1, s2 string) bool {
	sl1 := len([]rune(s1))
	sl2 := len([]rune(s2))
	if sl1 > 5000 || sl2 > 5000 || sl1 != sl2 {
		return false
	}
	for _,v := range s1 {
		if strings.Count(s1,string(v)) != strings.Count(s2,string(v)) {
			return false
		}
	}
	return true
}

//请编写⼀个⽅法，将字符串中的空格全部替换为“%20”。 假定该字符串有⾜够的空间存
//放新增的字符，并且知道字符串的真实⻓度(⼩于等于1000)，同时保证字符串由【⼤⼩
//写的英⽂字⺟组成】。 给定⼀个string为原始的串，返回替换后的string。

func replaceBlank(s string) (string, bool) {
	if len(s) > 1000 {
		return s, false
	}
	for _, v := range s {
		if string(v) != " " && unicode.IsLetter(v) == false {
			return s, false
		}
	}
	return strings.Replace(s, " ", "%20", -1), true
}

func TestReplaceBlank(t *testing.T) {
	s := " jgh kk"
	fmt.Println(replaceBlank(s))
}