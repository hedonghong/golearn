package master

import (
	"context"
	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"golearn/crontab/common"
	"time"
)

//cron/workers/
type WorkerMgr struct {
	client *clientv3.Client
	kv clientv3.KV
	lease clientv3.Lease
}

var (
	GWorkerMgr *WorkerMgr
)

func (w *WorkerMgr) ListWorkers() (workers []string, err error)  {
	var (
		getResp *clientv3.GetResponse
		kv *mvccpb.KeyValue
		workerIP string
	)

	//初始化数组
	workers = make([]string, 0)

	//获取目录前缀下的所有kv
	if getResp, err = w.kv.Get(context.TODO(), common.JOB_WORKER_DIR, clientv3.WithPrefix()); err != nil {
		return
	}
	//解析每个节点ip
	for _, kv = range getResp.Kvs {
		workerIP = common.ExtractWorkerIP(string(kv.Value))
		workers = append(workers, workerIP)
	}
	return
}

func InitWorkerMgr() (err error)  {
	var (
		config clientv3.Config
		client *clientv3.Client
		kv clientv3.KV
		lease clientv3.Lease
	)

	// 初始化配置
	config = clientv3.Config{
		Endpoints: GConfig.Endpoints, // 集群地址
		DialTimeout: time.Duration(GConfig.DialTimeout) * time.Millisecond, // 连接超时
	}

	// 建立连接
	if client, err = clientv3.New(config); err != nil {
		return
	}

	// 得到KV和Lease的API子集
	kv = clientv3.NewKV(client)
	lease = clientv3.NewLease(client)

	GWorkerMgr = &WorkerMgr{
		client :client,
		kv: kv,
		lease: lease,
	}
	return
}
