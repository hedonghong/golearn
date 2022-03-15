package middleware

import (
	"github.com/gin-gonic/gin"
	"golearn/goextra/gin/model/response"
	"golearn/goextra/gin/util"
)

func JwtMiddleware() gin.HandlerFunc  {
	return func(c *gin.Context) {
		token := c.Request.Header.Get("token")
		if token == "" {
			response.FailWithMessage("请求未携带token，无权限访问", c)
			c.Abort()
			return
		}
		j := util.NewJWT()
		claims, err := j.ParserToken(token)
		if err != nil {
			response.FailWithMessage(err.Error(), c)
			c.Abort()
			return
		}
		c.Set("claims", claims)
		c.Next()
	}
}
