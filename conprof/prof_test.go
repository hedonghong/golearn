package conprof

import (
	"fmt"
	"github.com/mosn/holmes"
	"github.com/pyroscope-io/pyroscope/pkg/agent/profiler"
	"github.com/shirou/gopsutil/mem"
	"net/http"
	"testing"
	"time"
)

func TestGogsutil(t *testing.T) {
	profiler.Start(profiler.Config{
		ApplicationName: "simple.golang.app",
		ServerAddress:   "http://pyroscope:4040", // this will run inside docker-compose, hence `pyroscope` for hostname
		// by default all profilers are enabled,
		// but you can select the ones you want to use:
		ProfileTypes: []profiler.ProfileType{
			profiler.ProfileCPU,
			profiler.ProfileAllocObjects,
			profiler.ProfileAllocSpace,
			profiler.ProfileInuseObjects,
			profiler.ProfileInuseSpace,
		},
	})
	for  {
		v, _ := mem.VirtualMemory()
		fmt.Printf("Total: %v, Available: %v, UsedPercent:%f%%\n", v.Total, v.Available, v.UsedPercent)
		fmt.Println(v)
		time.Sleep(time.Second * 3)
	}
}

func Hello(w http.ResponseWriter, r *http.Request)  {
	var a = make([]byte, 1073741824)
	_ = a
	fmt.Fprintln(w, "hello")
}

func TestMosnHolmes(t *testing.T) {
	h, _ := holmes.New(
		holmes.WithCollectInterval("2s"),
		holmes.WithCoolDown("1m"),
		holmes.WithDumpPath("./"),
		holmes.WithTextDump(),
		//WithMemDump(30, 25, 80) means dump will happen when memory usage > 10% && memory usage > 125% * previous memory usage or memory usage > 80%
		holmes.WithMemDump(1, 1, 1),
	)
	h.EnableMemDump()
	h.Start()

	http.HandleFunc("/", Hello)
	http.ListenAndServe(":12345", nil)
	//h.Stop()
}