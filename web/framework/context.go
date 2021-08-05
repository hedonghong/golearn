package framework

import (
	"fmt"
	"github.com/julienschmidt/httprouter"
	"net/http"
)

//middleware
type HandlerFunc func(ctx *Context)

type Context struct {
	Writer http.ResponseWriter
	Request *http.Request
	// middleware
	handlers []HandlerFunc
	index    int
	router *httprouter.Router
	length int
}

func (c *Context) Next() {
	c.index++
	if c.length == 0 {
		c.router.ServeHTTP(c.Writer, c.Request)
	}
	for ; c.index < c.length+1; c.index++ {
		if c.index == c.length {
			c.router.ServeHTTP(c.Writer, c.Request)
		} else {
			c.handlers[c.index](c)
		}
	}
}

//兼容http.HandleFunc
func TimeMiddleware(next http.HandlerFunc) httprouter.Handle {
	return func(wr http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		fmt.Println("timeMiddleware start")
		// next handler
		next.ServeHTTP(wr, r)
		fmt.Println("timeMiddleware end")
	}
}