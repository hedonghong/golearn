package recover

import (
	"fmt"
	"log"
	"runtime"
	"sync"
	"testing"
	"time"
)

func RecoverFromPanic(funcName string)  {
	if e := recover(); e != nil {
		buf := make([]byte, 64 << 10)
		buf = buf[:runtime.Stack(buf, false)]
		log.Printf("[%s] func_name: %v, stack:%s", funcName, e, string(buf))
		paincError := fmt.Errorf("%v", e)
		fmt.Println(paincError)
		//do report
		ReportPanic(paincError.Error(), funcName, string(buf))
	}
}

type PanicReq struct {
	Service string `json:"service"`
	ErrorInfo string `json:"error_info"`
	Stack string `json:"stack"`
	LogId string `json:"log_id"`
	FuncName string `json:"func_name"`
	Host string `json:"host"`
	PodName string `json:"pod_name"`
}

var panicReportOnce sync.Once

func ReportPanic(errorInfo, funcName, stack string)  {
	panicReportOnce.Do(func() {
		defer func() {recover()}()
		go func() {
			paniceReq := &PanicReq{
				Service:   "xxx", //读取当前环境
				ErrorInfo: errorInfo,
				Stack:     stack,
				FuncName: funcName,
				Host: "127.0.0.1",//读取当前服务器Ip
				PodName: "xxx",//读取当前PodName
			}
			fmt.Println(paniceReq)
			//jsonByte, err := json.Marshal(paniceReq)
			//if err != nil {
			//	return
			//}
			//httpReq, err := http.NewRequest("POST", "xxxx.com", bytes.NewBuffer(jsonByte))
			//if err != nil {
			//	return
			//}
			//httpReq.Header.Set("Content-Type", "application/json")
			//client := &http.Client{
			//	Timeout: 5 * time.Second,
			//}
			//resp, err := client.Do(httpReq)
			//if err != nil {
			//	return
			//}
			//defer resp.Body.Close()
			//body, err := ioutil.ReadAll(resp.Body)
			////查看回复
			//bodyString := string(body)
			//fmt.Println(bodyString)
		}()
	})
}

func TestRecover(t *testing.T) {
	fmt.Println("test start")
	defer RecoverFromPanic("TestRecover")
	panic("test 一个错误")
	fmt.Println("test end")
	time.Sleep(10 * time.Second)
}
