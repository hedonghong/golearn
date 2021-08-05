package main

import (
	"flag"
	"fmt"
	"golearn/crontab/master"
	"os"
	"os/signal"
	"runtime"
	"syscall"
)

var (
	configFile string
)

func initEnv()  {
	//设置程序所能使用的cpu与内核相同
	runtime.GOMAXPROCS(runtime.NumCPU())
}

func initArgs()  {
	//master -config ./master.json
	flag.StringVar(&configFile, "config", "./master.json", "设置配置文件路径")
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
	if err = master.InitConfig(configFile); err != nil {
		goto ERR
	}

	//启动任务管理器
	if err = master.InitJobMgr(); err != nil {
		goto ERR
	}

	//启动api http服务
	if err = master.InitApiServer(); err != nil {
		goto ERR
	}

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