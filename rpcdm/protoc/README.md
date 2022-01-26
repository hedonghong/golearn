命令：

protoc -I proto/ --go_out=plugins=grpc:proto proto/helloworld.proto

-I 指定代码输出目录，忽略服务定义的包名，否则会根据包名创建目录
--go_out 指定代码输出目录，格式：--go_out=plugins=grpc:目录名
命令最后面的参数是proto协议文件 编译成功后在proto目录生成了helloworld.pb.go文件，里面包含了，我们的服务和接口定义。
--proto_path=PATH 与-I同义  它表示的是我们要在哪个路径下搜索proto文件，这个参数既可以用-I指定，也可以使用--proto_path=指定

go rpc

1、net/rpc god编码 tcp/http只能用于golang之间

2、net/rpc/jsonrpc 只支持tcp

3、protorpc库

rpc框架

grpc gRPC使用protobuf进行序列化和反序列化

rpcx

调试工具

grpcui-就是gRPC中的postman