package worker

import (
	"context"
	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"golearn/crontab/common"
	"time"
)

//任务管理器
type JobMgr struct {
	client *clientv3.Client
	kv clientv3.KV
	lease clientv3.Lease
	watcher clientv3.Watcher
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

	watcher := clientv3.NewWatcher(client)

	GJobMgr = &JobMgr{
		client: client,
		kv: kv,
		lease: lease,
		watcher: watcher,
	}

	//启动监听任务
	GJobMgr.WatchJobs()
	//启动监听killer
	GJobMgr.watchKiller()
	return nil
}

//监听任务变化
func (j *JobMgr) WatchJobs() (err error)  {

	var (
		getResp *clientv3.GetResponse
		job *common.Job
	)
	//1、get从cron/jobs/目录下所有任务，并且获知当前集群revision
	if getResp, err = j.kv.Get(context.TODO(), common.JOB_SAVE_DIR, clientv3.WithPrefix()); err != nil {
		return
	}
	for _, kvPaire := range getResp.Kvs {
		if job, err = common.UnpackJob(kvPaire.Value); err == nil {
			jobEvent := common.BuildJobEvent(common.JOB_EVENT_SAVE, job)
			// 同步给scheduler(调度协程)
			GScheduler.PushJobEvent(jobEvent)
		}
	}
	//2、从该revision向后监听变化事件
	go func() {
		var (
			watchResp clientv3.WatchResponse
			watchEvent *clientv3.Event
			job *common.Job
			jobEvent *common.JobEvent
		)
		watchStartRevision := getResp.Header.Revision + 1
		watchChan := j.watcher.Watch(context.TODO(), common.JOB_SAVE_DIR,
			clientv3.WithRev(watchStartRevision), clientv3.WithPrefix())
		for watchResp = range watchChan {
			for _, watchEvent = range watchResp.Events {
				switch watchEvent.Type {
				//任务新建或者更新
				case mvccpb.PUT:
					//TODO 反序列化 推送给调度协程
					if job, err = common.UnpackJob(watchEvent.Kv.Value); err != nil {
						continue
					}
					//构造一个event新增或更新事件
					jobEvent = common.BuildJobEvent(common.JOB_EVENT_SAVE, job)
				//任务删除
				case mvccpb.DELETE:
					//TODO 推删除事件给调度协程
					jobName := common.ExtractJobName(string(watchEvent.Kv.Key))
					//构造一个event删除事件
					jobEvent = common.BuildJobEvent(common.JOB_EVENT_DELETE, &common.Job {
						Name: jobName,
					})
				}
				//推删除事件给调度协程
				GScheduler.PushJobEvent(jobEvent)
			}
		}

	}()

	return
}

// 监听强杀任务通知
func (j *JobMgr) watchKiller() {

	//监听/cron/killer/目录的变化
	go func() {
		var (
			watchResp clientv3.WatchResponse
			watchEvent *clientv3.Event
		)
		watchChan := j.watcher.Watch(context.TODO(), common.JOB_KILLER_DIR, clientv3.WithPrefix())
		for watchResp = range watchChan {
			for _, watchEvent = range watchResp.Events {
				switch watchEvent.Type {
				case mvccpb.PUT://有杀死任务事件
					jobName := common.ExtractKillerName(string(watchEvent.Kv.Key))
					job := &common.Job{
						Name: jobName,
					}
					jobEvent := common.BuildJobEvent(common.JOB_EVENT_KILL, job)
					//事件通知调度协程
					GScheduler.PushJobEvent(jobEvent)
				case mvccpb.DELETE://killer标记过期，被自动删除

				}
			}
		}
	}()
}

// 创建任务执行锁
func (j *JobMgr) CreateJobLock(jobName string) (jobLock *JobLock){
	jobLock = InitJobLock(jobName, j.kv, j.lease)
	return
}
