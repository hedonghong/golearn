package response

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

const (
	SUCCESS = 0
	FAILURE = 20001
)

type Response struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

func Result(code int, msg string, data interface{}, c *gin.Context) {
	// 开始时间
	c.JSON(http.StatusOK, Response {
		code,
		msg,
		data,
	})
}

func Ok(c *gin.Context) {
	Result(SUCCESS, "success", map[string]interface{}{}, c)
}

func OkWithMessage(message string, c *gin.Context) {
	Result(SUCCESS, message, map[string]interface{}{}, c)
}

func OkWithData(data interface{}, c *gin.Context) {
	Result(SUCCESS, "success", data, c)
}

func OkWithDetailed(data interface{}, message string, c *gin.Context) {
	Result(SUCCESS, message, data, c)
}

func Fail(c *gin.Context) {
	Result(FAILURE, "failure", map[string]interface{}{}, c)
}

func FailWithMessage(message string, c *gin.Context) {
	Result(FAILURE, message, map[string]interface{}{}, c)
}

func FailWithDetailed(data interface{}, message string, c *gin.Context) {
	Result(FAILURE, message, data, c)
}
