package main

import (
	"flag"
	"fmt"
	"golearn/crontab/worker"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	_ "net/http/pprof"
)

var (
	configFile string
)

func initEnv()  {
	//设置程序所能使用的cpu与内核相同
	runtime.GOMAXPROCS(runtime.NumCPU())
}

func initArgs()  {
	//worker -config ./worker.json
	flag.StringVar(&configFile, "config", "./worker.json", "设置配置文件路径")
	flag.Parse()
}

func main()  {
	var (
		err error
		sig chan os.Signal//定义退出信号
	)

	//初始化命令行参数
	initArgs()

	//初始化线程
	initEnv()

	//加载配置
	if err = worker.InitConfig(configFile); err != nil {
		goto ERR
	}

	// 服务注册
	if err = worker.InitRegister(); err != nil {
		goto ERR
	}

	//启动日志协程
	if err = worker.InitLogSink(); err != nil {
		goto ERR
	}

	//启动执行器
	if err = worker.InitExecutor(); err != nil {
		goto ERR
	}

	//启动调度器
	if err = worker.InitScheduler(); err != nil {
		goto ERR
	}

	//启动任务管理器
	if err = worker.InitJobMgr(); err != nil {
		goto ERR
	}

	//用于监控pprof
	go func() {
		fmt.Println(http.ListenAndServe(":50002", nil))
	}()

	//正常退出
	sig = make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGTERM, syscall.SIGINT)
	<-sig
	//或者select {}
	return

	//错误退出
ERR:
	fmt.Println("错误", err)
}