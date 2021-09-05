package main

import (
	"context"
	"fmt"
	"github.com/micro/go-micro/v2"
	"github.com/micro/go-micro/v2/registry"
	"github.com/micro/go-micro/v2/registry/etcd"
	hedonghong "golearn/micro/proto/cap"
)

func main()  {
	// 实例化
	service := micro.NewService(
		micro.Name("cap.hedonghong.client"),
		micro.Registry(etcd.NewRegistry(
			// 地址是我本地etcd服务器地址，不要照抄
			registry.Addrs("127.0.0.1:2379"),
		)),
	)
	// 初始化
	service.Init()

	// 写需要访问的service
	capImooc := hedonghong.NewCapService("cap.hedonghong.server", service.Client())
	res, err := capImooc.SayHello(context.TODO(), &hedonghong.SayRequest{Message: "Go语言 微服务学习 你学废了吗！"})
	if err != nil {
		fmt.Println(err)
	}else{
		fmt.Println(res.Answer)
	}
}

