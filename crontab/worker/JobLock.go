package worker

import (
	"context"
	"github.com/coreos/etcd/clientv3"
	"golearn/crontab/common"
)

//分布式锁 TXN事务
type JobLock struct {
	kv clientv3.KV
	lease clientv3.Lease
	leaseId clientv3.LeaseID //租约ID

	jobName string //任务名
	cancelFunc context.CancelFunc //用于终止自动续租
	isLocked bool //是否已经上锁
}

func InitJobLock(jobName string, kv clientv3.KV, lease clientv3.Lease) (jobLock *JobLock) {
	jobLock = &JobLock{
		kv: kv,
		lease: lease,
		jobName: jobName,
	}
	return
}

//尝试上锁
func (j *JobLock) TryLock() (err error) {
	var (
		leaseId clientv3.LeaseID
		leaseResp *clientv3.LeaseGrantResponse
		cancelCtx context.Context
		cancelFunc context.CancelFunc
		keepRespChan <- chan *clientv3.LeaseKeepAliveResponse
		txn clientv3.Txn
		lockKey string
		txnResp *clientv3.TxnResponse
	)
	//1、创建租约5秒
	if leaseResp, err = j.lease.Grant(context.TODO(), 5); err != nil {
		return
	}
	//context取消自动续租
	cancelCtx, cancelFunc = context.WithCancel(context.TODO())

	//获取租约id
	leaseId = leaseResp.ID

	//2、自动续租
	if keepRespChan, err = j.lease.KeepAlive(cancelCtx, leaseId); err != nil {
		goto FAIL
	}

	//3、处理续租答应的协议
	go func() {
		var keepResp *clientv3.LeaseKeepAliveResponse
		for  {
			select {
			//自动续租应答
			case keepResp = <- keepRespChan:
				if keepResp == nil {
					goto END
				}
			}
		}
		END:
	}()

	//4、创建事务
	txn = j.kv.Txn(context.TODO())

	//锁路径
	lockKey = common.JOB_LOCK_DIR + j.jobName

	//5、事务抢锁
	txn.If(clientv3.Compare(clientv3.CreateRevision(lockKey), "=", 0)).
		Then(clientv3.OpPut(lockKey, "", clientv3.WithLease(leaseId))).
		Else(clientv3.OpGet(lockKey))

	//提交事务
	if txnResp, err = txn.Commit(); err != nil {
		goto FAIL
	}

	//6、成功返回，失败释放租约
	if !txnResp.Succeeded {//锁被占用
		err = common.ERR_LOCK_ALREADY_REQUIRED
		goto FAIL
	}

	//抢锁成功
	j.leaseId = leaseId
	j.cancelFunc = cancelFunc
	j.isLocked = true
	return

	FAIL:
		cancelFunc() //取消自动续租
		j.lease.Revoke(context.TODO(), leaseId)
	return
}

//释放锁
func (j *JobLock) Unlock()  {
	if j.isLocked {
		j.cancelFunc()//取消续租
		j.lease.Revoke(context.TODO(), j.leaseId)//取消租约
	}
}
