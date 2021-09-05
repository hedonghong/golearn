
1、step 1

go get github.com/micro/go-micro/v2
go get github.com/micro/protoc-gen-micro/v2
<!-- go get -u github.com/golang/protobuf/protoc-gen-go -->
go get github.com/gogo/protobuf/protoc-gen-gofast

2、切换注册中心

```go
1、
	service := micro.NewService(
		micro.Name("cap.hedonghong.server"),
		micro.Registry(etcd.NewRegistry(
			// 地址是我本地etcd服务器地址，不要照抄
			registry.Addrs("127.0.0.1:2379"),
		)),
	)
2、
--registry=etcd
--registry_address=127.0.0.1:2379
```

3、protoc转go代码

```go
protoc *.proto --gofast_out=. --micro_out=.
```

4、micro/micro 代码生成工具安装

```go
go get github.com/micro/micro/v2

安装之后可以命令
micro help

cd go-workspace
# 调用micro生成代码
# 默认情况下Micro生成的代码会放到GOPATH/src中，通过配置--gopath=false可以选择在当前目录下
micro new --gopath=false user (user模块)
```