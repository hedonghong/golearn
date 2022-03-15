package fasthttp

import (
	"fmt"
	"github.com/valyala/fasthttp"
	"testing"
)

var portServer string = ":20014"
var porthost string = "http://127.0.0.1:20014"

func TestFastHttpServer(t *testing.T) {
	fasthttp.ListenAndServe(portServer, func(ctx *fasthttp.RequestCtx) {
		httpHandle(ctx)
	})
}

func httpHandle(ctx *fasthttp.RequestCtx)  {
	// 避免高并发下使用，被复用的request互相影响
	newRequest := &fasthttp.Request{}
	ctx.Request.CopyTo(newRequest)

	//fmt.Println(newRequest.PostArgs())
	fmt.Println(string(ctx.QueryArgs().Peek("test")))

	body := newRequest.Body()

	fmt.Println(string(body))

	ctx.Response.AppendBodyString("ok")
	ctx.Response.SetStatusCode(200)
}

func TestFasthttpGet(t *testing.T) {
	status, resp, err := fasthttp.Get(nil, porthost)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Println(status)
	fmt.Println(string(resp))
}

func TestFasthttpPost(t *testing.T) {
	args := &fasthttp.Args{}
	args.Add("test", "TestFasthttpPost")
	status, resp, err := fasthttp.Post(nil, porthost+"?"+"test=ss", args)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Println(status)
	fmt.Println(string(resp))
}


