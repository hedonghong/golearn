package framework

import (
	"fmt"
	"time"
)

func Logger() HandlerFunc {
	return func(c *Context) {
		startTime := time.Now()
		fmt.Println("start"+c.Request.URL.Path)
		c.Next()
		fmt.Println("end"+c.Request.URL.Path)
		endTime := time.Since(startTime)
		fmt.Println(endTime.Seconds())
	}
}
