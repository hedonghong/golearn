package main
import (
	"github.com/micro/go-micro/v2/registry"
	"github.com/micro/go-micro/v2/registry/etcd"
	hedonghong "golearn/micro/proto/cap"
	"context"
	"fmt"
	"github.com/micro/go-micro/v2"
)


type CapServer struct {}
// 需要实现的方法 参数可以从上面imooc.pb.micro.go中获取
func (c *CapServer) SayHello(ctx context.Context, req *hedonghong.SayRequest , res *hedonghong.SayResponse) error {
	// 业务逻辑代码
	res.Answer = "我们口号是: \"" + req.Message + "\""
	return nil
}


func main(){
	// 创建新的服务
	service := micro.NewService(
		micro.Name("cap.hedonghong.server"),
		micro.Registry(etcd.NewRegistry(
			// 地址是我本地etcd服务器地址，不要照抄
			registry.Addrs("127.0.0.1:2379"),
		)),
	)
	// 初始化方法
	service.Init()
	// 注册我们的服务 RegisterCapHandler 就是我们自己在imooc.proto中生成的服务cap
	// imooc为我们自动生成的别名 原来为cap_imooc_service_imooc 重新起个别名
	hedonghong.RegisterCapHandler(service.Server(), new(CapServer))
	// 运行服务
	if err := service.Run(); err != nil {
		fmt.Println(err)
	}
}
