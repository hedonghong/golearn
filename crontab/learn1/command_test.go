package learn1

import (
	"context"
	"fmt"
	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"github.com/gorhill/cronexpr"
	"os/exec"
	"runtime"
	"testing"
	"time"
)

//没有捕获输出的
func TestCommand(t *testing.T) {
	cmd := exec.Command("/bin/ls", "./")
	err := cmd.Run()
	fmt.Println(err)
}

func TestCommand1(t *testing.T) {
	var outPut[]byte
	var err error
	cmd := exec.Command("/bin/ls", "/Users/sky.he/go/src/golearn")
	if outPut,err = cmd.CombinedOutput(); err != nil {
		fmt.Println(err)
	}
	fmt.Println(string(outPut))
	err = cmd.Run()
	fmt.Println(err)
}

func TestCommand2(t *testing.T) {
	var result chan string
	result = make(chan string, 1)
	ctx, cancelfun := context.WithCancel(context.Background())
	go func() {
		var (
			err error
			outPut []byte
		)
		cmd := exec.CommandContext(ctx, "/bin/sleep", "2")
		if outPut,err = cmd.CombinedOutput(); err != nil {
			fmt.Println(err)
		}
		cmd.Run()
		result <- string(outPut)
	}()

	time.Sleep(1 * time.Second)
	select {
	case re := <- result:
		fmt.Println("result")
		fmt.Println(re)
	case <- time.NewTicker(1 * time.Second).C:
		fmt.Println("time out")
		cancelfun()
	}
	time.Sleep(10 * time.Second)
}

func TestCommand3(t *testing.T) {
	var (
		expre *cronexpr.Expression
		err error
	)
	cronStr := "*/1 * * * *"
	if expre, err = cronexpr.Parse(cronStr); err != nil {
		fmt.Println(err)
		return
	}
	now := time.Now()
	//fmt.Println(expre)
	nt := expre.Next(now)
	//fmt.Println(nt.Format("2006-01-02 15:04:05"))
	time.AfterFunc(nt.Sub(now), func() {
		fmt.Println("time up")
	})
	time.Sleep(1800 * time.Second)
}

func TestCommand4(t *testing.T) {
	now := time.Now()
	expr := cronexpr.MustParse("*/5 * * * * * *")//5秒一次
	cronJob := &CronJob{
		expr: expr,
		nextTime: expr.Next(now),
	}

	var scheduleTable map[string]*CronJob
	scheduleTable = make(map[string]*CronJob)
	scheduleTable["job1"] = cronJob
	
	go func() {
		for {
			now1 := time.Now()
			for name, job := range scheduleTable {
				if job.nextTime.Before(now1) || job.nextTime.Equal(now1) {
					go func(jobName string) {
						fmt.Println(jobName)
					}(name)
					job.nextTime = job.expr.Next(now1)
				}
			}
			runtime.Gosched()
		}
	}()

	time.Sleep(15 * time.Second)
}

func TestCommand5(t *testing.T) {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{"localhost:2379"},
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		fmt.Println(err)
		return
	}
	defer cli.Close()

	//put
	kvCli := clientv3.NewKV(cli)
	putResp, err1 := kvCli.Put(context.TODO(), "/hello/dfd", "world")
	fmt.Println(putResp.Header.Revision)
	fmt.Println(err1)

	//put 并且返回前一个值
	putResp2, err2 := kvCli.Put(context.TODO(), "/hello/dfd", "world1", clientv3.WithPrevKV())
	fmt.Println(putResp2.PrevKv)
	if putResp2.PrevKv != nil {
		fmt.Println(string(putResp2.PrevKv.Value))
	}
	fmt.Println(err2)

	getResp, err3 := kvCli.Get(context.TODO(), "/hello/dfd")
	fmt.Println(getResp.Kvs)
	fmt.Println(err3)
}

func TestCommand6(t *testing.T) {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{"localhost:2379"},
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		fmt.Println(err)
		return
	}
	defer cli.Close()

	//put
	kvCli := clientv3.NewKV(cli)
	kvCli.Put(context.TODO(), "/etcd/test/1", "1")
	kvCli.Put(context.TODO(), "/etcd/test/2", "2")
	kvCli.Put(context.TODO(), "/etcd/test/3", "3")

	//根据目前前缀批量读取
	getResp, _ := kvCli.Get(context.TODO(), "/etcd/test/", clientv3.WithPrefix())
	fmt.Println(getResp.Kvs)
}

func TestCommand7(t *testing.T) {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{"localhost:2379"},
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		fmt.Println(err)
		return
	}
	defer cli.Close()

	//put
	kvCli := clientv3.NewKV(cli)

	//根据目前删除，并且设置WithPrevKV把历史值放在PrevKvs中
	//设置clientv3.WithPrefix() 则是前缀批量删除
	delResp, _ := kvCli.Delete(context.TODO(), "/etcd/test/1", clientv3.WithPrevKV())
	fmt.Println(delResp.Deleted)
	fmt.Println(delResp.PrevKvs)
	for k, ev := range delResp.PrevKvs {
		fmt.Println(k)
		fmt.Println(string(ev.Value))
	}
}

func TestCommand8(t *testing.T) {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{"localhost:2379"},
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		fmt.Println(err)
		return
	}
	defer cli.Close()

	//put
	kvCli := clientv3.NewKV(cli)

	//开一个租约
	leaseCli := clientv3.NewLease(cli)
	leaseResp , _ := leaseCli.Grant(context.TODO(), 10)//秒
	fmt.Println(leaseResp)
	leaseId := leaseResp.ID

	//超时
	canelCtx, _ := context.WithTimeout(context.TODO(), 10 * time.Second)
	//若要续租，5秒内都会需求，每秒续租，5秒后不再续租
	keepAliveChan, _ := leaseCli.KeepAlive(canelCtx, leaseId)

	go func() {
		for  {
			select {
				case kaResp := <- keepAliveChan:
					if kaResp == nil {
						fmt.Println("租约失败")
						return
					}
					fmt.Println(kaResp.ID)
			}
		}
	}()

	//设置自动10秒过期 , WithLease()设置关联租约
	kvCli.Put(context.TODO(), "/etcd/timeout", "10", clientv3.WithLease(leaseId))

	var i int
	for {
		getResp, _ := kvCli.Get(context.TODO(), "/etcd/timeout")
		fmt.Println(getResp.Kvs)
		time.Sleep(2 * time.Second)
		i++
		if i > 12 {
			break
		}
	}
}

func TestCommand9(t *testing.T) {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{"localhost:2379"},
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		fmt.Println(err)
		return
	}
	defer cli.Close()

	//put
	kvCli := clientv3.NewKV(cli)

	//模拟健值不断变化
	go func() {
		for {
			kvCli.Put(context.TODO(), "/etcd/command9", "com9")

			kvCli.Delete(context.TODO(), "/etcd/command9")

			time.Sleep(1 * time.Second)
		}
	}()

	getResp, _ := kvCli.Get(context.TODO(), "/etcd/command9")
	if len(getResp.Kvs) != 0 {
		fmt.Println(getResp.Kvs[0].Value)
	}

	//当前etcd集群事物ID，自增 从+1后的版本开始监听
	watchId := getResp.Header.Revision + 1

	watcher := clientv3.NewWatcher(cli)

	/*
	//如果想监听一段事件
	cancelCtx, cancelFunc := context.WithCancel(context.TODO())
	time.AfterFunc(5 * time.Second, func() {
		cancelFunc()
	})
	//cancelCtx 把这个放在下面的监听上下文中
	 */

	watchChan := watcher.Watch(context.TODO(), "/etcd/command9", clientv3.WithRev(watchId))

	for watchResp := range watchChan {
		//Events 事件
		for _, ev := range watchResp.Events {
			switch ev.Type {
			case mvccpb.PUT:
				fmt.Println("修改", string(ev.Kv.Value), "revision",
					ev.Kv.CreateRevision, ev.Kv.ModRevision)
			case mvccpb.DELETE:
				fmt.Println("删除", string(ev.Kv.Value), ev.Kv.ModRevision)
			}
		}
	}
}

func TestCommand10(t *testing.T) {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{"localhost:2379"},
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		fmt.Println(err)
		return
	}
	defer cli.Close()

	kvCli := clientv3.NewKV(cli)

	//创建op
	putOp := clientv3.OpPut("/etcd/command10", "com10")
	//执行op
	opResp, _ := kvCli.Do(context.TODO(), putOp)
	fmt.Println(opResp.Put().Header.Revision)

	getOp := clientv3.OpGet("/etcd/command10")
	getResp, _ := kvCli.Do(context.TODO(), getOp)
	fmt.Println(string(getResp.Get().Kvs[0].Value))
}

func TestCommand11(t *testing.T) {
	//lease锁自动过期 续租
	//op操作
	//tx事物

	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{"localhost:2379"},
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		fmt.Println(err)
		return
	}
	defer cli.Close()

	kvCli := clientv3.NewKV(cli)

	//1、上锁 创建租约，自动续租，拿着租约去抢占key
	leaser := clientv3.NewLease(cli)
	leaseResp, _ := leaser.Grant(context.TODO(), 5)//5秒锁
	leaseId := leaseResp.ID

	cancelCtx, cancelFun := context.WithCancel(context.TODO())
	defer cancelFun()//取消自动续租
	defer leaser.Revoke(context.TODO(), leaseId)//删除租约
	leaseAliveChan, _ := leaser.KeepAlive(cancelCtx, leaseId)


	go func() {
		for  {
			select {
			case aliveResp := <- leaseAliveChan:
				if aliveResp == nil {
					goto END
				} else {
					fmt.Println("续租成功", aliveResp.ID)
				}
			}
		}
		END:
	}()

	//抢占key
	txn := kvCli.Txn(context.TODO())
	//当这个key不存在
	txn.If(clientv3.Compare(clientv3.CreateRevision("/etcd/command11"), "=", 0)).
	//新建这个key
		Then(clientv3.OpPut("/etcd/command11", "xxx", clientv3.WithLease(leaseId))).
	//否则读取下这个key
		Else(clientv3.OpGet("/etcd/command11"))
	//提交事物
	txnResp, _ := txn.Commit()

	//判断是否抢占锁 - 下面是没抢到锁
	if !txnResp.Succeeded {
		fmt.Println("锁被占用", string(txnResp.Responses[0].GetResponseRange().Kvs[0].Value))
		return
	}

	//2、处理业务

	fmt.Println("处理业务")
	time.Sleep(5 * time.Second)

	//3、释放锁 取消自动续租 释放租约（会立即删除对应的key）
	/*
		defer cancelFun()//取消自动续租
		defer leaser.Revoke(context.TODO(), leaseId)//删除租约
	 */
}



