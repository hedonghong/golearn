package main

import (
	"fmt"
	"github.com/julienschmidt/httprouter"
	"golearn/web/demo"
	_ "golearn/web/demo"
	"golearn/web/framework"
	"net/http"
)

func hello(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	fmt.Println("hello")
	w.Write([]byte("hello world"))
}

func hi(w http.ResponseWriter, r *http.Request)  {
	fmt.Println("hi run")
	w.Write([]byte("hi run"))
}

func main()  {
	//router := httprouter.New()
	//router.GET("/", hello)
	//http.ListenAndServe(":30001", router)
	defer framework.Db.Close()

	router := framework.New()
	router.Use(framework.Logger())
	router.GET("/", hello)
	router.GET("/demo", demo.Demo)
	v1GroupRouter := router.Group("/v1")
	v1GroupRouter.GET("/hi", framework.TimeMiddleware(hi))

	router.Run(":40001")
}
