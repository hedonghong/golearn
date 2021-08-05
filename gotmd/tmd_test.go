package gotmd

import (
	"testing"
)

func TestTmd(t *testing.T) {
	e := New()
	e.Use(Logger())
	e.GET("/", func(ctx *Context) {
		ctx.JSON(200, map[string]interface{}{
			"name": "user",
			"password": "1234",
		})
	})
	e.Run(":9091")
}
