package worker

import (
	"context"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golearn/crontab/common"
	"time"
)

//mongodb存储日志
type LogSink struct {
	client *mongo.Client
	logCollection *mongo.Collection
	logChan chan *common.JobLog
	autoCommitChan chan *common.LogBatch
}

var (
	GLogSink *LogSink
)

func InitLogSink() (err error)  {
	var (
		client *mongo.Client
	)

	//1、建立连接
	if client, err = mongo.Connect(context.TODO(),
		options.Client().ApplyURI(GConfig.MongodbUri),
		options.Client().SetConnectTimeout(time.Duration(GConfig.MongodbConnectTimeout) * time.Millisecond),
		); err != nil {
		return
	}

	//2、选择db和collection
	GLogSink = &LogSink{
		client: client,
		logCollection: client.Database("cron").Collection("log"),
		logChan: make(chan *common.JobLog, 1000),
		autoCommitChan: make(chan *common.LogBatch, 1000),
	}

	go GLogSink.writeLoop()
	return
}

// 批量写入日志
func (l *LogSink) saveLogs(batch *common.LogBatch) {
	l.logCollection.InsertMany(context.TODO(), batch.Logs)
}

func (l *LogSink) writeLoop()  {
	var (
		log *common.JobLog
		logBatch *common.LogBatch // 当前的批次
		commitTimer *time.Timer
		timeoutBatch *common.LogBatch // 超时批次
	)
	for {
		select {
		case log = <- l.logChan:
			if logBatch == nil {
				logBatch = &common.LogBatch{}
				//让批次超时自动提交（1秒）
				commitTimer = time.AfterFunc(
					time.Duration(GConfig.JobLogCommitTimeout) * time.Millisecond,
					func(batch *common.LogBatch) func() {
						return func() {
							l.autoCommitChan <- batch
						}
					}(logBatch),
				)
			}
			//把日志追加到批次中
			logBatch.Logs = append(logBatch.Logs, log)
			//如果批次满了，就立即发送
			if len(logBatch.Logs) >= GConfig.JobLogBatchSize {
				//发送日志
				l.saveLogs(logBatch)
				//清空日志
				logBatch = nil
				//取消定时器
				commitTimer.Stop()
			}
		case timeoutBatch = <- l.autoCommitChan:
			//判断过期批次是否仍旧是当前批次
			if timeoutBatch != logBatch {
				continue
			}
			//把批次写入mongodb
			l.saveLogs(timeoutBatch)
			//清空logBatch
			logBatch = nil
		}
	}
}

// 发送日志
func (l *LogSink) Append(jobLog *common.JobLog) {
	select {
	case l.logChan <- jobLog:
	default:
		// 队列满了就丢弃
	}
}