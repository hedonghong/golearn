package main

import (
	"log"
	_ "net/http/pprof"
	"os"
	"runtime/pprof"
	"time"
)

// this is the main function
func main() {
	// 主函数中添加
	//1、http://localhost:9999/debug/pprof/
	//2、go tool pprof -http=:1234http://localhost:40002/debug/pprof/profile
	//http.ListenAndServe("0.0.0.0:40002", nil)


	f, err := os.OpenFile("./cpu.prof", os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		log.Fatal("could not create CPU profile: ", err)
	}
	defer f.Close()

	if err := pprof.StartCPUProfile(f); err != nil {
		log.Fatal("could not start CPU profile: ", err)
	}
	defer pprof.StopCPUProfile()


	go workForever()

	//if *memprofile != "" {
	//	f, err := os.Create(*memprofile)
	//	defer f.Close()
	//	if err != nil {
	//		log.Fatal("could not create memory profile: ", err)
	//	}
	//	runtime.GC() // get up-to-date statistics
	//	if err := pprof.WriteHeapProfile(f); err != nil {
	//		log.Fatal("could not write memory profile: ", err)
	//	}
	//
	//}
	pprof.StopCPUProfile()
	select {
	}
}

func counter() {
	slice := make([]int, 0)
	c := 1
	for i := 0; i < 100000; i++ {
		c = i + 1 + 2 + 3 + 4 + 5
		slice = append(slice, c)
	}
}

func workForever() {
	for {
		go counter()
		time.Sleep(1 * time.Second)
	}
}
