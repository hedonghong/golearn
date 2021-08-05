package ch6

import (
	"fmt"
	"io/ioutil"
	"net/http"
)

var bdiduChan chan []byte
var bingChan chan []byte

func ch6()  {
	bdiduChan = make(chan []byte, 1)
	bingChan  = make(chan []byte, 1)

	go func() {
		baiduResp,_ := http.Get("https://www.baidu.com")
		defer baiduResp.Body.Close()
		bodyc, _ := ioutil.ReadAll(baiduResp.Body)
		//time.Sleep(10 * time.Second)
		bdiduChan <- bodyc
	}()

	go func() {
		bingResp,_ := http.Get("https://bing.com")
		defer bingResp.Body.Close()
		bodyc, _ := ioutil.ReadAll(bingResp.Body)
		bingChan <- bodyc
	}()

	select {
	case html := <- bdiduChan:
		fmt.Println("baidu.com")
		fmt.Println(string(html))
	case html := <- bingChan:
		fmt.Println("bing.com")
		fmt.Println(string(html))
	}

}
