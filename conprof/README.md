# 线上性能采集

相同类型的项目
conprof，profefe，Pyroscope

## quick start docker-compose安装docker-compose.yml

```go
version: '3'
services:
  pyroscope:
    image: "pyroscope/pyroscope:latest"
    ports:
      - "4040:4040"
    command:
      - "server"
```

## [localPyroscope](http://localhost:4040/ )
    https://pyroscope.io/docs/server-configuration/ 具体配置

## 应用如何接入?

```go
package main

import (
	"github.com/pyroscope-io/pyroscope/pkg/agent/profiler"
)


func main() {
	profiler.Start(profiler.Config{
		ApplicationName: "simple.golang.app",
		ServerAddress:   "http://pyroscope:4040", // this will run inside docker-compose, hence `pyroscope` for hostname
		// by default all profilers are enabled,
		// but you can select the ones you want to use:
		ProfileTypes: []profiler.ProfileType{
			profiler.ProfileCPU,
			profiler.ProfileAllocObjects,
			profiler.ProfileAllocSpace,
			profiler.ProfileInuseObjects,
			profiler.ProfileInuseSpace,
		},
	})
	//application code
}

```

## 另外如何做到按需收集
    现在的应用可能跑在物理机上，也可能跑在 docker 中，因此在获取用量时，需要我们自己去做一些定制。物理机中的数据采集，可以使用 gopsutil，docker 中的数据采集，可以参考少量 cgroups 中的实现
https://github.com/shirou/gopsutil  https://zhuanlan.zhihu.com/p/126362239
https://github.com/containerd/cgroups

https://xargin.com/continuous-profiling/