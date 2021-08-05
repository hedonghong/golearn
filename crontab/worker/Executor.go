package worker

import (
	"golearn/crontab/common"
	"math/rand"
	"os/exec"
	"time"
)

//任务执行器
type Executor struct {

}

var (
	GExecultor *Executor
)

//初始化执行器
func InitExecutor() error {
	GExecultor = &Executor{}
	return nil
}

//启动一个协程执行一个任务
func (e *Executor) ExecuteJob(jobExecInfo *common.JobExecuteInfo)  {
	go func() {
		var (
			err error
		)
		//初始化执行结果
		result := &common.JobExecuteResult{
			ExecuteInfo: jobExecInfo,
			OutPut: make([]byte, 0),
		}

		//初始化分布式锁
		jobLock := GJobMgr.CreateJobLock(jobExecInfo.Job.Name)


		//记录任务开始时间
		result.StartTime = time.Now()

		//上锁
		// 随机睡眠(0~1s)
		time.Sleep(time.Duration(rand.Intn(1000)) * time.Millisecond)

		err = jobLock.TryLock()
		defer jobLock.Unlock()
		//上锁并解锁
		if err != nil {
			result.Err = err
			result.EndTime = time.Now()
		} else {
			//上锁成功，重置任务启动时间
			result.StartTime = time.Now()

			//执行shell命令
			cmd := exec.CommandContext(jobExecInfo.CancelCtx, "/bin/bash", "-C",jobExecInfo.Job.Command)

			//执行并捕获输出
			outPut, err := cmd.CombinedOutput()

			//记录任务结束时间
			result.EndTime = time.Now()
			result.OutPut = outPut
			result.Err = err
		}
		//任务执行完成后，把执行的结果返回给Scheduler，Scheduler会从executingTable中删除掉执行记录
		GScheduler.PushJobResult(result)
	}()
}
