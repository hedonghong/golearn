package inited

import (
	"github.com/gin-gonic/gin"
	v1 "golearn/goextra/gin/api/v1"
	"golearn/goextra/gin/middleware"
	"net/http"
)

func Routers() *gin.Engine {
	var router = gin.Default()
	router.Use(middleware.CORSMiddleware(), middleware.ZapMiddleware())
	router.POST("/userLogin", v1.UserLogin)
	router.GET("/demo", func(c *gin.Context) {

		okMap := make(map[string] interface{})
		okMap["code"] = 200
		okMap["msg"] = "ok"

		c.JSON(http.StatusOK, okMap)
	})
	group := router.Group("/api")
	authGroup := group.Use(middleware.JwtMiddleware())
	{
		authGroup.GET("/getUser", v1.GetSysUser)
		authGroup.POST("/postUser", v1.PostSysUser)
	}
	return router
}