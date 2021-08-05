package master

import (
	"encoding/json"
	"golearn/crontab/common"
	"net"
	"net/http"
	"net/http/pprof"
	"strconv"
	"time"
)

var (
	GApiServer *ApiServer
)

type ApiServer struct {
	httpServer *http.Server
}

func InitApiServer() error {
	var mux *http.ServeMux
	mux = http.NewServeMux()

	mux.HandleFunc("/job/save", handleJobSave)
	mux.HandleFunc("/job/delete", handleJobDelete)
	mux.HandleFunc("/job/list", handleJobList)
	mux.HandleFunc("/job/kill", handleJobKill)
	//mux.HandleFunc("/job/log", handleJobLog)
	mux.HandleFunc("/worker/list", handleWorkerList)
	//pprof
	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)

	listener, err := net.Listen("tcp", ":"+strconv.Itoa(GConfig.ApiPort))
	if err != nil {
		return err
	}
	httpServer := &http.Server{
		ReadHeaderTimeout: time.Duration(GConfig.ApiReadTimeOut) * time.Millisecond,
		WriteTimeout: time.Duration(GConfig.ApiWriteTimeOut) * time.Millisecond,
		Handler: mux,
	}
	GApiServer = &ApiServer{
		httpServer: httpServer,
	}

	go httpServer.Serve(listener)
	return nil
}

//保存任务
//POST job={"name":"jobName", "command":"echo hello", "cronExpr":"* * * * * * *"}
func handleJobSave(w http.ResponseWriter, r *http.Request)  {
	var (
		err error
		job common.Job
		postJob string
		oldJob *common.Job
		respByte []byte
	)
	//任务保存到etcd中
	if err = r.ParseForm(); err != nil {
		goto ERR//后面不再允许使用 := 方式命名变量
	}

	postJob = r.PostForm.Get("job")
	if err = json.Unmarshal([]byte(postJob), &job); err != nil {
		goto ERR
	}

	if oldJob, err = GJobMgr.SaveJob(&job); err != nil {
		goto ERR
	}

	//正常应答
	if respByte, err = common.BuildResponse(0, "success", oldJob); err != nil {
		goto ERR
	}
	w.Write(respByte)
	return
	ERR:
	//错误应答
		if respByte, err = common.BuildResponse(1, err.Error(), nil); err == nil {
			w.Write(respByte)
		}
}

//删除任务
//POST name=jobname
func handleJobDelete(w http.ResponseWriter, r *http.Request)  {
	var (
		err error
		name string
		oldJob *common.Job
		respByte []byte
	)
	if err = r.ParseForm(); err != nil {
		goto ERR
	}

	name = r.PostForm.Get("name")

	//删除任务
	if oldJob, err = GJobMgr.DeleteJob(name); err != nil {
		goto ERR
	}
	//正常应答
	if respByte, err = common.BuildResponse(0, "success", oldJob); err != nil {
		goto ERR
	}
	w.Write(respByte)
	return
	ERR:
		//错误应答
		if respByte, err = common.BuildResponse(1, err.Error(), nil); err == nil {
			w.Write(respByte)
		}
}

//任务列表
func handleJobList(w http.ResponseWriter, r *http.Request)  {
	var (
		err error
		jobList []*common.Job
		respByte []byte
	)
	if jobList, err = GJobMgr.listJob(); err != nil {
		goto ERR
	}

	//正常应答
	if respByte, err = common.BuildResponse(0, "success", jobList); err != nil {
		goto ERR
	}
	w.Write(respByte)
	return
	ERR:
		//错误应答
		if respByte, err = common.BuildResponse(1, err.Error(), nil); err == nil {
			w.Write(respByte)
		}
}

func handleJobKill(w http.ResponseWriter, r *http.Request)  {
	var (
		err error
		name string
		respByte []byte
	)

	if err = r.ParseForm(); err != nil {
		goto ERR
	}
	name = r.PostForm.Get("name")
	if err = GJobMgr.KillJob(name); err != nil {
		goto ERR
	}
	//正常应答
	if respByte, err = common.BuildResponse(0, "success", nil); err != nil {
		goto ERR
	}
	return
	ERR:
	//错误应答
	if respByte, err = common.BuildResponse(1, err.Error(), nil); err == nil {
		w.Write(respByte)
	}
}

func handleWorkerList(w http.ResponseWriter, r *http.Request) {
	var (
		workerArr []string
		err error
		bytes []byte
	)

	if workerArr, err = GWorkerMgr.ListWorkers(); err != nil {
		goto ERR
	}

	// 正常应答
	if bytes, err = common.BuildResponse(0, "success", workerArr); err == nil {
		w.Write(bytes)
	}
	return

ERR:
	if bytes, err = common.BuildResponse(-1, err.Error(), nil); err == nil {
		w.Write(bytes)
	}
}