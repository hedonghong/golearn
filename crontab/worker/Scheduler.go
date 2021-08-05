package worker

import (
	"fmt"
	"golearn/crontab/common"
	"time"
)

//任务调度
type Scheduler struct {
	//TODO 下面map可以用sync.map
	jobEventChan chan *common.JobEvent //etcd任务事件队列
	jobPlanTable map[string]*common.JobSchedulePlan //任务调度计划表
	jobExecutingTable map[string] *common.JobExecuteInfo //任务执行表
	jobResultChan chan *common.JobExecuteResult //任务结果队列
}

var (
	GScheduler * Scheduler
)

// 初始化调度器
func InitScheduler() (err error) {
	GScheduler = &Scheduler{
		jobEventChan: make(chan *common.JobEvent, 1000),
		jobPlanTable: make(map[string]*common.JobSchedulePlan),
		jobExecutingTable: make(map[string]*common.JobExecuteInfo),
		jobResultChan: make(chan *common.JobExecuteResult, 1000),
	}
	// 启动调度协程
	go GScheduler.scheduleLoop()
	return
}

//处理任务事件
func (s *Scheduler) handleJobEvent(jobEvent *common.JobEvent) {
	switch jobEvent.EventType {
	case common.JOB_EVENT_SAVE:
		jobSchedulePlan, err := common.BuildJobSchedulePlan(jobEvent.Job)
		if err != nil {
			return
		}
		s.jobPlanTable[jobEvent.Job.Name] = jobSchedulePlan
	case common.JOB_EVENT_DELETE:
		if _, jobExisted := s.jobPlanTable[jobEvent.Job.Name]; jobExisted {
			delete(s.jobPlanTable, jobEvent.Job.Name)
		}
	case common.JOB_EVENT_KILL:
		//取消掉command执行，判断任务是否在执行中
		if jobExecuteInfo, jobExecuting := s.jobExecutingTable[jobEvent.Job.Name]; jobExecuting {
			jobExecuteInfo.CancelFunc()//触发取消上下文参数
		}
	}
}

//执行任务
func (s *Scheduler) TryStartJob(jobPlan *common.JobSchedulePlan)  {
	//任务正在执行，跳过本次执行
	if _, jobExecuting := s.jobExecutingTable[jobPlan.Job.Name]; jobExecuting {
		return
	}

	//构建执行状态信息
	jobExecuteInfo := common.BuildJobExecuteInfo(jobPlan)

	//保存执行的状态
	s.jobExecutingTable[jobPlan.Job.Name] = jobExecuteInfo

	//执行任务
	fmt.Println("执行任务:", jobExecuteInfo.Job.Name, jobExecuteInfo.PlanTime, jobExecuteInfo.RealTime)

	//调用执行任务器
	GExecultor.ExecuteJob(jobExecuteInfo)
}

//计算任务调度时间
func (s *Scheduler) TrySchedule() (scheduleAfter time.Duration) {

	var (
		nearTime *time.Time
	)

	//如果任务列表为空，睡眠1秒
	if len(s.jobPlanTable) == 0 {
		scheduleAfter = 1 * time.Second
		return
	}

	//当前时间
	now := time.Now()

	//遍历所有任务
	for _, jobPlan := range s.jobPlanTable {
		//如果任务应执行时间过了，或者下次执行时间等于当前时间，那就执行任务吧
		if jobPlan.NextTime.Before(now) || jobPlan.NextTime.Equal(now) {
			s.TryStartJob(jobPlan)
			//更新下次执行时间
			jobPlan.NextTime = jobPlan.Expr.Next(now)
		}

		//统计最近一个要过期的任务时间
		if nearTime == nil || jobPlan.NextTime.Before(*nearTime) {
			nearTime = &jobPlan.NextTime
		}
	}
	//下次调度间隔时间（最近要执行的任务调度时间 - 当前时间）
	scheduleAfter = (*nearTime).Sub(now)
	return
}

func (s *Scheduler) handleJobResult(result *common.JobExecuteResult)  {
	//删除正在执行状态
	delete(s.jobExecutingTable, result.ExecuteInfo.Job.Name)

	//生成执行日志
	if result.Err != common.ERR_LOCK_ALREADY_REQUIRED {
		jobLog := &common.JobLog{
			JobName: result.ExecuteInfo.Job.Name,
			Command: result.ExecuteInfo.Job.Command,
			Output: string(result.OutPut),
			PlanTime: result.ExecuteInfo.PlanTime.UnixNano()/1000/1000,
			ScheduleTime: result.ExecuteInfo.RealTime.UnixNano()/1000/1000,
			StartTime: result.StartTime.UnixNano()/1000/1000,
			EndTime: result.EndTime.UnixNano()/1000/1000,
		}
		if result.Err != nil {
			jobLog.Err = result.Err.Error()
		} else {
			jobLog.Err = ""
		}
		GLogSink.Append(jobLog)
	}
}

// 调度协程
func (s *Scheduler) scheduleLoop() {

	var (
		scheduleAfter time.Duration
	)
	//初始化一次(1秒)
	scheduleAfter = s.TrySchedule()

	//调度的延迟定时器
	scheduleTimer := time.NewTimer(scheduleAfter)

	//定时任务common.Job
	for  {
		select {
		//监听任务变化事件
		case jobEvent := <- s.jobEventChan:
			//对内存中维护的任务表做增删改查
			s.handleJobEvent(jobEvent)
		//最近的任务到期了，运行一下
		case <- scheduleTimer.C:
		//监听任务执行结果
		case jobResultChan := <- s.jobResultChan:
			s.handleJobResult(jobResultChan)
		}
		//调度一次任务
		scheduleAfter = s.TrySchedule()
		//重置调度间隔
		scheduleTimer.Reset(scheduleAfter)
	}
}

// 推送任务变化事件
func (s *Scheduler) PushJobEvent(jobEvent *common.JobEvent) {
	s.jobEventChan <- jobEvent
}

// 回传任务执行结果
func (s *Scheduler) PushJobResult(jobResult *common.JobExecuteResult) {
	s.jobResultChan <- jobResult
}