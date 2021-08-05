package gotmd

import (
	"fmt"
	"time"
)

func Logger() HandlerFunc  {
	return func(ctx *Context) {
		//
		startTime := time.Now()
		ctx.Next()
		endTime := time.Since(startTime)
		fmt.Println(endTime.Seconds())
	}
}
