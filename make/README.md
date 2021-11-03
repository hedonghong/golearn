

## 文件命令组成

```go
<target> : <prerequisites> 
[tab]  <commands>
```

上面第一行冒号前面的部分，叫做"目标"（target），冒号后面的部分叫做"前置条件"（prerequisites）；第二行必须由一个tab键起首，后面跟着"命令"（commands）。

"目标"是必需的，不可省略；"前置条件"和"命令"都是可选的，但是两者之中必须至少存在一个。

每条规则就明确两件事：构建目标的前置条件是什么，以及如何构建。下面就详细讲解，每条规则的这三个组成部分。

target

一个目标（target）就构成一条规则。目标通常是文件名，指明Make命令所要构建的对象，比如上文的 a.txt 。目标可以是一个文件名，也可以是多个文件名，之间用空格分隔。

除了文件名，目标还可以是某个操作的名字，这称为"伪目标"（phony target）。

```go
clean:
    rm *.o
```

上面代码的目标是clean，它不是文件名，而是一个操作的名字，属于"伪目标 "，作用是删除对象文件。


$ make  clean

但是，如果当前目录中，正好有一个文件叫做clean，那么这个命令不会执行。因为Make发现clean文件已经存在，就认为没有必要重新构建了，就不会执行指定的rm命令。

为了避免这种情况，可以明确声明clean是"伪目标"，写法如下。

```go
.PHONY: clean
clean:
    rm *.o temp
```


声明clean是"伪目标"之后，make就不会去检查是否存在一个叫做clean的文件，而是每次运行都执行对应的命令。像.PHONY这样的内置目标名还有不少，可以查看手册。

如果Make命令运行时没有指定目标，默认会执行Makefile文件的第一个目标。


$ make

上面代码执行Makefile文件的第一个目标。


prerequisites

前置条件通常是一组文件名，之间用空格分隔。它指定了"目标"是否重新构建的判断标准：只要有一个前置文件不存在，或者有过更新（前置文件的last-modification时间戳比目标的时间戳新），"目标"就需要重新构建。

```go
result.txt: source.txt
    cp source.txt result.txt
```

上面代码中，构建 result.txt 的前置条件是 source.txt 。如果当前目录中，source.txt 已经存在，那么make result.txt可以正常运行，否则必须再写一条规则，来生成 source.txt 。

```go
source.txt:
    echo "this is the source" > source.txt
```

上面代码中，source.txt后面没有前置条件，就意味着它跟其他文件都无关，只要这个文件还不存在，每次调用make source.txt，它都会生成。


$ make result.txt
$ make result.txt
上面命令连续执行两次make result.txt。第一次执行会先新建 source.txt，然后再新建 result.txt。第二次执行，Make发现 source.txt 没有变动（时间戳晚于 result.txt），就不会执行任何操作，result.txt 也不会重新生成。

如果需要生成多个文件，往往采用下面的写法。


source: file1 file2 file3
上面代码中，source 是一个伪目标，只有三个前置文件，没有任何对应的命令。


$ make source
执行make source命令后，就会一次性生成 file1，file2，file3 三个文件。这比下面的写法要方便很多。


$ make file1
$ make file2
$ make file3


commands

命令（commands）表示如何更新目标文件，由一行或多行的Shell命令组成。它是构建"目标"的具体指令，它的运行结果通常就是生成目标文件。

每行命令之前必须有一个tab键。如果想用其他键，可以用内置变量.RECIPEPREFIX声明。

```go
.RECIPEPREFIX = >
all:
> echo Hello, world
```

上面代码用.RECIPEPREFIX指定，大于号（>）替代tab键。所以，每一行命令的起首变成了大于号，而不是tab键。

需要注意的是，每行命令在一个单独的shell中执行。这些Shell之间没有继承关系。

```go
var-lost:
    export foo=bar
    echo "foo=[$$foo]"
```

上面代码执行后（make var-lost），取不到foo的值。因为两行命令在两个不同的进程执行。一个解决办法是将两行命令写在一行，中间用分号分隔。

```go
var-kept:
    export foo=bar; echo "foo=[$$foo]"
```

另一个解决办法是在换行符前加反斜杠转义。

```go
var-kept:
    export foo=bar; \
    echo "foo=[$$foo]"
```

最后一个方法是加上.ONESHELL:命令。

```go
.ONESHELL:
var-kept:
    export foo=bar; 
    echo "foo=[$$foo]"
```

## 其他语法

1、注释
井号（#）在Makefile中表示注释。

```go
#这个是注释
```

2、回声
正常情况下，make会打印每条命令，然后再执行，这就叫做回声（echoing）。

```go
test:
    # 这是测试
```
在命令的前面加上@，就可以关闭回声。
```go
test:
    @# 这是测试
```

3、通配符
通配符（wildcard）用来指定一组符合条件的文件名。Makefile 的通配符与 Bash 一致，主要有星号（*）、问号（？）和 [...] 。比如， *.o 表示所有后缀名为o的文件。

4、模式匹配

Make命令允许对文件名，进行类似正则运算的匹配，主要用到的匹配符是%。比如，假定当前目录下有 f1.c 和 f2.c 两个源码文件，需要将它们编译为对应的对象文件。

%.o: %.c

等同于下面的写法。

f1.o: f1.c
f2.o: f2.c

5、变量和赋值符

Makefile 允许使用等号自定义变量。

```go
txt = Hello World
test:
    @echo $(txt)
```

有时，变量的值可能指向另一个变量。

v1 = $(v2)

上面代码中，变量 v1 的值是另一个变量 v2。这时会产生一个问题，v1 的值到底在定义时扩展（静态扩展），还是在运行时扩展（动态扩展）？如果 v2 的值是动态的，这两种扩展方式的结果可能会差异很大。

为了解决类似问题，Makefile一共提供了四个赋值运算符 （=、:=、？=、+=）

VARIABLE = value
# 在执行时扩展，允许递归扩展。

VARIABLE := value
# 在定义时扩展。

VARIABLE ?= value
# 只有在该变量为空时才设置值。

VARIABLE += value
# 将值追加到变量的尾端。

6、内置变量

Make命令提供一系列内置变量，比如，$(CC) 指向当前使用的编译器，$(MAKE) 指向当前使用的Make工具。这主要是为了跨平台的兼容性，详细的内置变量清单[手册](https://www.gnu.org/software/make/manual/html_node/Implicit-Variables.html)

```go
output:
    $(CC) -o output input.c
```

7、自动变量

Make命令还提供一些自动变量，它们的值与当前规则有关。主要有以下几个。
（1）$@

$@指代当前目标，就是Make命令当前构建的那个目标。比如，make foo的 $@ 就指代foo。

```go
a.txt b.txt: 
    touch $@
```

等同于下面的写法

```go
a.txt:
    touch a.txt
b.txt:
    touch b.txt
```

2）$<

$< 指代第一个前置条件。比如，规则为 t: p1 p2，那么$< 就指代p1。

```go
a.txt: b.txt c.txt
    cp $< $@ 
```

等同于下面的写法

```go
a.txt: b.txt c.txt
    cp b.txt a.txt 
```

（3）$?

$? 指代比目标更新的所有前置条件，之间以空格分隔。比如，规则为 t: p1 p2，其中 p2 的时间戳比 t 新，$?就指代p2。

（4）$^

$^ 指代所有前置条件，之间以空格分隔。比如，规则为 t: p1 p2，那么 $^ 就指代 p1 p2 。

（5）$*

$* 指代匹配符 % 匹配的部分， 比如% 匹配 f1.txt 中的f1 ，$* 就表示 f1。

（6）$(@D) 和 $(@F)

$(@D) 和 $(@F) 分别指向 $@ 的目录名和文件名。比如，$@是 src/input.c，那么$(@D) 的值为 src ，$(@F) 的值为 input.c。

（7）$(<D) 和 $(<F)

$(<D) 和 $(<F) 分别指向 $< 的目录名和文件名。

所有的自动变量清单，请看[手册](https://www.gnu.org/software/make/manual/html_node/Automatic-Variables.html)。下面是自动变量的一个例子。

```go
dest/%.txt: src/%.txt
    @[ -d dest ] || mkdir dest
    cp $< $@
```

上面代码将 src 目录下的 txt 文件，拷贝到 dest 目录下。首先判断 dest 目录是否存在，如果不存在就新建，然后，$< 指代前置文件（src/%.txt）， $@ 指代目标文件（dest/%.txt）。

8、 判断和循环

Makefile使用 Bash 语法，完成判断和循环。

```go
ifeq ($(CC),gcc)
  libs=$(libs_for_gcc)
else
  libs=$(normal_libs)
endif
```

上面代码判断当前编译器是否 gcc ，然后指定不同的库文件。

```go
LIST = one two three
all:
    for i in $(LIST); do \
        echo $$i; \
    done

# 等同于

all:
    for i in one two three; do \
        echo $i; \
    done

```
上面代码的运行结果。
```go
one
two
three
```

9、函数
Makefile 还可以使用函数，格式如下。

```go
$(function arguments)
# 或者
${function arguments}
```

Makefile提供了许多[内置函数](https://www.gnu.org/software/make/manual/html_node/Functions.html)，可供调用。下面是几个常用的内置函数。

（1）shell 函数

shell 函数用来执行 shell 命令

```go
srcfiles := $(shell echo src/{00..99}.txt)
```

（2）wildcard 函数

wildcard 函数用来在 Makefile 中，替换 Bash 的通配符。

```go
srcfiles := $(wildcard src/*.txt)
```

（3）subst 函数

subst 函数用来文本替换，格式如下。

```go
$(subst from,to,text)
```

下面的例子将字符串"feet on the street"替换成"fEEt on the strEEt"。


$(subst ee,EE,feet on the street)
下面是一个稍微复杂的例子。

```go
comma:= ,
empty:=
# space变量用两个空变量作为标识符，当中是一个空格
space:= $(empty) $(empty)
foo:= a b c
bar:= $(subst $(space),$(comma),$(foo))
# bar is now `a,b,c'.
```
（4）patsubst函数

patsubst 函数用于模式匹配的替换，格式如下。

```go
$(patsubst pattern,replacement,text)
````
下面的例子将文件名"x.c.c bar.c"，替换成"x.c.o bar.o"。

```go
$(patsubst %.c,%.o,x.c.c bar.c)
```

（5）替换后缀名

替换后缀名函数的写法是：变量名 + 冒号 + 后缀名替换规则。它实际上patsubst函数的一种简写形式。

```go
min: $(OUTPUT:.js=.min.js)
```
上面代码的意思是，将变量OUTPUT中的后缀名 .js 全部替换成 .min.js 。

## 例子

```go
.PHONY: cleanall cleanobj cleandiff

cleanall : cleanobj cleandiff
        rm program

cleanobj :
        rm *.o

cleandiff :
        rm *.diff
```

make build: 编译
make vendor: 下载依赖
make api: 生成协议代码
make json: easyjson 代码生成
make test: 运行单元测试
make benchmark: 运行性能测试
make stat: 代码复杂度统计，代码行数统计
make clean: 清理 build 目录
make deep_clean: 清理所有代码以外的其他文件
make third: 下载所有依赖的第三方工具
make protoc: 下载 protobuf 工具
make glide: 下载 glide 依赖管理工具
make golang: 下载 golang 环境
make cloc: 下载 cloc 统计工具
make gocyclo: 下载 gocyclo 圈复杂度计算工具
make easyjson: 下载 easyjson 工具

```go
export PATH:=${PATH}:${GOPATH}/bin:$(shell pwd)/third/go/bin:$(shell pwd)/third/protobuf/bin:$(shell pwd)/third/cloc-1.76

.PHONY: all
all: third vendor api json build test stat

build: cmd/rta_server/*.go internal/*/*.go scripts/version.sh Makefile vendor api json
    @echo "编译"
    @rm -rf build/ && mkdir -p build/bin/ && \
    go build -ldflags "-X 'main.AppVersion=`sh scripts/version.sh`'" cmd/rta_server/main.go && \
    mv main build/bin/rta_server && \
    cp -r configs build/configs/

vendor: glide.lock glide.yaml
    @echo "下载 golang 依赖"
    glide install

api: vendor
    @echo "生成协议文件"
    @rm -rf api && mkdir api && \
    cd vendor/gitlab.mobvista.com/vta/rta_proto.git/ && \
    protoc --go_out=plugins=grpc:. *.proto && \
    cd - && \
    cp vendor/gitlab.mobvista.com/vta/rta_proto.git/* api/

json: internal/rcommon/rta_common_easyjson.go

internal/rcommon/rta_common_easyjson.go: internal/rcommon/rta_common.go Makefile
    easyjson internal/rcommon/rta_common.go

.PHONY: test
test: vendor api json
    @echo "运行单元测试"
    go test -cover internal/rranker/*.go
    go test -cover internal/rserver/*.go
    go test -cover internal/rworker/*.go
    go test -cover internal/rloader/*.go
    go test -cover internal/rrecall/*.go
    go test -cover internal/rmaster/*.go
    go test -cover internal/rsender/*.go

benchmark: benchmarkloader benchmarkall

.PHONY: benchmarkloader
benchmarkloader: vendor api json
    @echo "运行 loader 性能测试"
    go test -timeout 2h -bench BenchmarkS3Loader_Load -benchmem -cpuprofile cpu.out -memprofile mem.out -run=^$$ internal/rloader/*
    go tool pprof -svg ./rloader.test cpu.out > cpu.benchmarkloader.svg
    go tool pprof -svg ./rloader.test mem.out > mem.benchmarkloader.svg

.PHONY: benchmarkserver
benchmarkserver: vendor api json
    @echo "运行 server 性能测试"
    go test -timeout 2h -bench BenchmarkServer -benchmem -cpuprofile cpu.out -memprofile mem.out -run=^$$ internal/rserver/*
    go tool pprof -svg ./rserver.test cpu.out > cpu.benchmarkserver.svg
    go tool pprof -svg ./rserver.test mem.out > mem.benchmarkserver.svg

.PHONY: benchmarkall
benchmarkall: vendor api json
    @echo "运行 server 性能测试"
    go test -timeout 2h -bench BenchmarkAll -benchmem -cpuprofile cpu.out -memprofile mem.out -run=^$$ internal/rserver/*
    go tool pprof -svg ./rserver.test cpu.out > cpu.benchmarkall.svg    
    go tool pprof -svg ./rserver.test mem.out > mem.benchmarkall.svg

.PHONY: benchmarkcache
benchmarkcache: vendor api json
    @echo "测试 redis 集群性能"
    go test -timeout 5m -bench BenchmarkRtaCacheBatch -benchmem -cpuprofile cpu.out -memprofile mem.out -run=^$$ internal/rserver/*

.PHONY: stat
stat: cloc gocyclo
    @echo "代码行数统计"
    @ls internal/*/* scripts/* configs/* Makefile | xargs cloc --by-file
    @echo "圈复杂度统计"
    @ls internal/*/* | grep -v _test | xargs gocyclo
    @ls internal/*/* | grep -v _test | xargs gocyclo | awk '{sum+=$$1}END{printf("总圈复杂度: %s", sum)}'

.PHONY: clean
clean:
    rm -rf build

.PHONY: deep_clean
deep_clean:
    rm -rf vendor api build third

third: protoc glide golang cloc gocyclo easyjson

.PHONY: protoc
protoc: golang
    @hash protoc 2>/dev/null || { \
        echo "安装 protobuf 代码生成工具 protoc" && \
        mkdir -p third && cd third && \
        wget https://github.com/google/protobuf/releases/download/v3.2.0/protobuf-cpp-3.2.0.tar.gz && \
        tar -xzvf protobuf-cpp-3.2.0.tar.gz && \
        cd protobuf-3.2.0 && \
        ./configure --prefix=`pwd`/../protobuf && \
        make -j8 && \
        make install && \
        cd ../.. && \
        protoc --version; \
    }
    @hash protoc-gen-go 2>/dev/null || { \
        echo "安装 protobuf golang 插件 protoc-gen-go" && \
        go get -u github.com/golang/protobuf/{proto,protoc-gen-go}; \
    }

.PHONY: glide
glide: golang
    @mkdir -p $$GOPATH/bin
    @hash glide 2>/dev/null || { \
        echo "安装依赖管理工具 glide" && \
        curl https://glide.sh/get | sh; \
    }

.PHONY: golang
golang:
    @hash go 2>/dev/null || { \
        echo "安装 golang 环境 go1.10" && \
        mkdir -p third && cd third && \
        wget https://dl.google.com/go/go1.10.linux-amd64.tar.gz && \
        tar -xzvf go1.10.linux-amd64.tar.gz && \
        cd .. && \
        go version; \
    }

.PHONY: cloc
cloc:
    @hash cloc 2>/dev/null || { \
        echo "安装代码统计工具 cloc" && \
        mkdir -p third && cd third && \
        wget https://github.com/AlDanial/cloc/archive/v1.76.zip && \
        unzip v1.76.zip; \
    }

.PHONY: gocyclo
gocyclo: golang
    @hash gocyclo 2>/dev/null || { \
        echo "安装代码圈复杂度统计工具 gocyclo" && \
        go get -u github.com/fzipp/gocyclo; \
    }

.PHONY: easyjson
easyjson: golang
    @hash easyjson 2>/dev/null || { \
        echo "安装 json 编译工具 easyjson" && \
        go get -u github.com/mailru/easyjson/...; \
    }
```





