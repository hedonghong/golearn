package master

import (
	"context"
	"encoding/json"
	"github.com/coreos/etcd/clientv3"
	"golearn/crontab/common"
	"time"
)

//任务管理器
type JobMgr struct {
	client *clientv3.Client
	kv clientv3.KV
	lease clientv3.Lease
}

var (
	GJobMgr *JobMgr
)

func InitJobMgr() error {
	//链接配置
	config := clientv3.Config{
		Endpoints: GConfig.Endpoints,//集群地址
		DialTimeout: time.Duration(GConfig.DialTimeout) * time.Millisecond,//链接超时时间
	}

	//建立链接
	client, err := clientv3.New(config)
	if err != nil {
		return err
	}

	//kv客户端
	kv := clientv3.NewKV(client)

	//租约客户端
	lease := clientv3.NewLease(client)

	GJobMgr = &JobMgr{
		client: client,
		kv: kv,
		lease: lease,
	}
	return nil
}

func (j *JobMgr) SaveJob(newJob *common.Job) (oldJob *common.Job, err error) {
	//任务保存到/cron/jobs/任务名 -> json
	var (
		jobJson []byte
		putResp *clientv3.PutResponse
		oldJobTemp common.Job
	)
	jobKey := common.JOB_SAVE_DIR+newJob.Name
	if jobJson, err = json.Marshal(newJob); err != nil {
		return
	}
	if putResp, err = j.kv.Put(context.TODO(), jobKey, string(jobJson), clientv3.WithPrevKV()); err != nil {
		return
	}
	//如果是更新，返回旧值
	if putResp.PrevKv != nil {
		//更新
		if err = json.Unmarshal(putResp.PrevKv.Value, &oldJobTemp); err != nil {
			//因为etcd本来就有，其实这里的报错无所谓
			return
		}
		oldJob = &oldJobTemp
		return
	} else {
		//新建，不用做什么
	}
	return
}

func (j *JobMgr) DeleteJob(name string) (oldJob *common.Job, err error) {
	var (
		deleteResp *clientv3.DeleteResponse
		oldJobTemp common.Job
	)
	jobKey := common.JOB_SAVE_DIR+name
	if deleteResp, err = j.kv.Delete(context.TODO(), jobKey, clientv3.WithPrevKV()); err != nil {
		return
	}
	//如果是更新，返回旧值
	if len(deleteResp.PrevKvs) != 0 {
		//更新
		if err = json.Unmarshal(deleteResp.PrevKvs[0].Value, &oldJobTemp); err != nil {
			//如果报错，其实也无所谓，反正报错了，这里先返回
			return
		}
		oldJob = &oldJobTemp
		return
	}
	return
}

func (j *JobMgr) listJob() (jobList []*common.Job, err error) {
	var (
		getResp *clientv3.GetResponse
	)
	jobList = make([]*common.Job, 0)
	dirKey := common.JOB_SAVE_DIR
	if getResp, err = j.kv.Get(context.TODO(), dirKey, clientv3.WithPrefix()); err != nil {
		return
	}
	if len(getResp.Kvs) != 0 {
		for _,jobPair := range getResp.Kvs {
			job := &common.Job{}
			if err = json.Unmarshal(jobPair.Value, job); err != nil {
				err = nil
				continue
			}
			jobList = append(jobList, job)
		}
	}
	return
}

//杀死一个正在运行的任务
func (j *JobMgr) KillJob(name string) (err error)  {
	var (
		leaseResp *clientv3.LeaseGrantResponse
	)
	killerKey := common.JOB_KILLER_DIR+name
	//让woker监听到一次put操作，创建一个租约自动过期，不用再多余清除删除的key
	if leaseResp, err = j.lease.Grant(context.TODO(), 1); err != nil {
		return
	}
	leaseId := leaseResp.ID

	if _, err = j.kv.Put(context.TODO(), killerKey, "", clientv3.WithLease(leaseId)); err != nil {
		return
	}
	return
}