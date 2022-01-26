package gosome

import (
	"fmt"
	"testing"
	"time"
)

func TestFor(t *testing.T) {
	for i := 0; i < 10; i++ {
		//v:=i
		go func() {
			//fmt.Println(v) 9,0,1,2,3,4,5,6,7,8
			fmt.Println(i) //10,10,10,10,10,10,10,10,10,10
		}()
	}
	time.Sleep(10 * time.Second)
}

type user struct {
	name string
	age uint64
}

func TestForRange(t *testing.T)  {
	u := []user{
		{"asong",23},
		{"song",19},
		{"asong2020",18},
	}
	n := make([]*user,0,len(u))
	for _,v := range u{
		n = append(n, &v)
	}
	fmt.Println(n)
	for _,v := range n{
		fmt.Println(v)
	}
	/*
	[0xc0000a6040 0xc0000a6040 0xc0000a6040]
	&{asong2020 18}
	&{asong2020 18}
	&{asong2020 18}
	 */

	// 优化：
	// 第一种
	// o := v

	// 第二种
	// for k,_ := range u{
	//		n = append(n, &u[k])
	//	}


	// 同理问题
	u = []user{
		{"asong",23},
		{"song",19},
		{"asong2020",18},
	}
	for _,v := range u{
		if v.age != 18{
			v.age = 20
		}
	}
	fmt.Println(u)
	//[{asong 23} {song 19} {asong2020 18}]
	// 调整
	/*
	for k,v := range u{
		if v.age != 18{
			u[k].age = 18
		}
	}
	 */
}

func TestMap(t *testing.T) {
	var addTomap = func() {
		var t = map[string]string{
			"asong": "太帅",
			"song": "好帅",
			"asong1": "非常帅",
		}
		for k := range t {
			t["song2020"] = "真帅"
			fmt.Printf("%s%s ", k, t[k])
		}
	}
	for i := 0; i < 10; i++ {
		addTomap()
		fmt.Println()
	}
	/*
	在循环中心新增不一定能遍历到
	asong太帅 song好帅 asong1非常帅 song2020真帅
	asong太帅 song好帅 asong1非常帅
	asong太帅 song好帅 asong1非常帅 song2020真帅
	asong太帅 song好帅 asong1非常帅 song2020真帅
	asong太帅 song好帅 asong1非常帅 song2020真帅
	asong太帅 song好帅 asong1非常帅 song2020真帅
	asong太帅 song好帅 asong1非常帅 song2020真帅
	asong太帅 song好帅 asong1非常帅 song2020真帅
	asong太帅 song好帅 asong1非常帅 song2020真帅
	asong太帅 song好帅 asong1非常帅 song2020真帅
	*/
}

// 1、5*5 3*3 5*3

// 20*20 = 400
// 400/25 = 16
// 400/9  = 44 ... 4
// 400/15 = 26 ... 10

// 25*x + 9*y + 15*z <= 400
