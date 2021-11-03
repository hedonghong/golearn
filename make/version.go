package main

import (
	"fmt"
	"os"
)

var buildstamp = ""
//var githash = ""
var goversion = ""

func main() {
	args := os.Args
	if len(args) == 2 && (args[1] == "--version" || args[1] == "-v") {
		//fmt.Printf("Git Commit Hash: %s\n", githash)
		fmt.Printf("UTC Build Time : %s\n", buildstamp)
		fmt.Printf("Golang Version : %s\n", goversion)
		return
	}
	fmt.Println("111")
}