1、下载镜像
```go
docker pull prom/node-exporter
docker pull prom/prometheus
docker pull grafana/grafana
docker pull prom/pushgateway
docker pull prom/alertmanager
https://www.cnblogs.com/xiao987334176/p/13203164.html
```

2、安装

```go

安装
docker run -d -p 9100:9100 \
-v "/proc:/host/proc:ro" \
-v "/sys:/host/sys:ro" \
-v "/:/rootfs:ro" \
--net="host" \
prom/node-exporter

检查lsof -i:9100

访问：
http://192.168.91.132:9100/metrics

安装
mkdir ～/code/docker/prometheus
cd prometheus
vim prometheus.yml

内容：
global:
scrape_interval:     60s
evaluation_interval: 60s

scrape_configs:
- job_name: prometheus
static_configs:
- targets: ['localhost:9090']
labels:
instance: prometheus

- job_name: linux
static_configs:
- targets: ['192.168.91.132:9100']
labels:
instance: localhost

docker run  -d \
-p 9090:9090 \
-v ～/code/docker/prometheus/prometheus.yml:/etc/prometheus/prometheus.yml  \
prom/prometheus

访问：
http://192.168.91.132:9090/graph
http://192.168.91.132:9090/targets

安装
mkdir ～/code/docker/grafana/storage
chmod 777 -R ～/code/docker/grafana/storage

docker run -d \
-p 3000:3000 \
--name=grafana \
-v ～/code/docker/grafana/storage:/var/lib/grafana \
grafana/grafana

访问：
http://192.168.91.132:3000/
默认的用户名和密码都是admin 第一次需要修改 hedonghong/hedonghong

安装
docker run -d \
--name=pg \
-p 9091:9091 \
prom/pushgateway

访问：
http://192.168.91.132:9091/

修改：prometheus.yml

global:
scrape_interval:     60s
evaluation_interval: 60s

scrape_configs:
- job_name: prometheus
static_configs:
- targets: ['localhost:9090']
labels:
instance: prometheus

- job_name: linux
static_configs:
- targets: ['192.168.91.132:9100']
labels:
instance: localhost

- job_name: pushgateway
static_configs:
- targets: ['192.168.91.132:9091']
labels:
instance: pushgateway

重启：prometheus
docker restart 59ae7d9c8c3a

访问：
http://192.168.91.132:9090/targets 是否已经启动
后面程序推送数据接入即可


注意：上面安装的普罗修斯有时区不对的问题
docker exec -it prometheus date 查看时区
它的时区为：UTC，我需要更改为CST，也就是中国上海时区
好像需要重新封装镜像
```

3、php_fpm接入普罗米修斯

```go

开启fpm的状态
pm.status_path = /fpm_status

nginx配置转发
location ~ ^/(fpm_status|health)$ {
    fastcgi_pass 192.168.31.34:9000;
    fastcgi_param SCRIPT_FILENAME $document_root$fastcgi_script_name;
    include fastcgi_params;
}

访问：
http://192.168.31.34/fpm_status
会有下面指标
pool-fpm 池子名称，大多数为www
process manager – 进程管理方式,值：static, dynamic or ondemand. dynamic
start time – 启动日期,如果reload了php-fpm，时间会更新
start since – 运行时长
accepted conn – 当前池子接受的请求数
listen queue –请求等待队列，如果这个值不为0，那么要增加FPM的进程数量
max listen queue – 请求等待队列最高的数量
listen queue len – socket等待队列长度
idle processes – 空闲进程数量
active processes –活跃进程数量
total processes – 总进程数量
max active processes –最大的活跃进程数量（FPM启动开始算）
max children reached -大道进程最大数量限制的次数，如果这个数量不为0，那说明你的最大进程数量太小了，请改大一点。
slow requests –启用了php-fpm slow-log，缓慢请求的数量

封装fpm-exporter

https://github.com/bakins/php-fpm-exporter/releases

注意：若下载zip文件，需要自己手动用go环境编译。对于go语言不熟悉的人，会编译失败。
所以，下载已经编译好的文件，是比较稳妥的办法

创建目录
./
├── dockerfile
├── php-fpm-exporter.linux.amd64（不同机器换）
└── run.sh


dockerfile

FROM alpine:3.10
ADD php-fpm-exporter.linux.amd64 /php-fpm-exporter
ADD run.sh /
RUN chmod 755 /php-fpm-exporter /run.sh
EXPOSE 9190
ENTRYPOINT [ "/run.sh" ]

run.sh

#!/bin/sh

/php-fpm-exporter --addr 0.0.0.0:9190 --endpoint $endpoint

打包成镜像

docker build -t php-fpm-exporter:v1 .
	
运行镜像
docker run -d -it --restart=always --name php-fpm-exporter -e endpoint=http://192.168.31.34/fpm_status -p 9190:9190 php-fpm-exporter:v1

访问：
http://192.168.31.34:9191/metrics

配置普罗米修斯

- job_name: PHP-FPM
    static_configs:
    - targets: ['192.168.31.34:9190']
    labels:
    instance: localhost

访问targets确保state为up
```
