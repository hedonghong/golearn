
# 获取go汇编代码

1、方法一: go tool compile
使用go tool compile -N -l -S once.go生成汇编代码：

2、方法二: go tool objdump
首先先编译程序: go tool compile -N -l once.go,

使用go tool objdump once.o反汇编出代码 (或者使用go tool objdump -s Do once.o反汇编特定的函数：)：

3、方法三: go build -gcflags -S
使用go build -gcflags -S once.go也可以得到汇编代码：


# dlv调试

1、dlv exec ./main  或  dlv debug ./main.go

p 打印 变量

bt 打印当前栈

n 下一步

si 进入函数

b 打断点

disass 反编译代码

# golint 

go get -u golang.org/x/lint/golint 安装

ls $GOPATH/bin 查看

golint file/dir 使用

golint校验常见的问题如下所示

don't use ALL_CAPS in Go names; use CamelCase
不能使用下划线命名法，使用驼峰命名法
exported function Xxx should have comment or be unexported
外部可见程序结构体、变量、函数都需要注释
var statJsonByte should be statJSONByte
var taskId should be taskID
通用名词要求大写
iD/Id -> ID
Http -> HTTP
Json -> JSON
Url -> URL
Ip -> IP
Sql -> SQL
don't use an underscore in package name
don't use MixedCaps in package name; xxXxx should be xxxxx
包命名统一小写不使用驼峰和下划线
comment on exported type Repo should be of the form "Repo ..." (with optional leading article)
注释第一个单词要求是注释程序主体的名称，注释可选不是必须的
type name will be used as user.UserModel by other packages, and that stutters; consider calling this Model
外部可见程序实体不建议再加包名前缀
if block ends with a return statement, so drop this else and outdent its block
if语句包含return时，后续代码不能包含在else里面
should replace errors.New(fmt.Sprintf(...)) with fmt.Errorf(...)
errors.New(fmt.Sprintf(…)) 建议写成 fmt.Errorf(…)
receiver name should be a reflection of its identity; don't use generic names such as "this" or "self"
receiver名称不能为this或self
error var SampleError should have name of the form ErrSample
错误变量命名需以 Err/err 开头
should replace num += 1 with num++
should replace num -= 1 with num--
a+=1应该改成a++，a-=1应该改成a–

GOLAND配合使用

新增tool:golint配合：Goland -> Tools -> External Tools 新建一个tool 配置如下

点击+新增

name:golint group: external tools

tool setting
    program: $GOPATH/bin/golint
    agrs : -set_exit_status $FilePath$
    workdir: $GOPATH/bin/golint

新增快捷键: Goland -> preference -> Keymap -> External Tools -> External Tools -> golint 右键新增快捷键
或者 右键  External Tools -》 golint 对当前文件检测

. gitlab提交限制
为了保证项目代码规范，我们可以在gitlab上做一层约束限制，当代码提交到gitlab的时候先做golint校验，校验不通过则不让提交代码。

我们可以为Go项目创建gitlab CI流程，通过.gitlab-ci.yml配置CI流程会自动使用govet进行代码静态检查、gofmt进行代码格式化检查、golint进行代码规范检查、gotest进行单元测试

例如go-common项目 .gitlab.yml文件如下，相关的脚本可以查看scripts【https://github.com/chenguolin/golang/tree/master/go-common/scripts】

# gitlab CI/CD pipeline配置文件
# 默认使用本地定制过的带有golint的golang镜像
image: golang:custom

stages:
- test

before_script:
- mkdir -p /go/src/gitlab.local.com/golang
- ln -s `pwd` /go/src/gitlab.local.com/golang/go-common && cd /go/src/gitlab.local.com/golang/go-common

# test stage
# job 1 test go vet
job_govet:
stage: test
script:
- bash ./scripts/ci-govet-check.sh
tags:
- dev
# job 2 test go fmt
job_gofmt:
stage: test
script:
- bash ./scripts/ci-gofmt-check.sh
tags:
- dev
# job 3 test go lint
job_golint:
stage: test
script:
- bash ./scripts/ci-golint-check.sh
tags:
- dev
# job 4 test go unit test
job_unit:
stage: test
script:
- bash ./scripts/ci-gotest-check.sh
tags:
- dev

另一个比较流行的代码规范检测

https://golangci-lint.run/usage/configuration/

# go大拿博客地址

[鸟窝](https://colobu.com/archives/)