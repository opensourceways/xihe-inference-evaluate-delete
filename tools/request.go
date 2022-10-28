package tools

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func Success(c *gin.Context, data interface{}, options ...RespOption) {
	c.JSON(http.StatusOK, SuccessReturn(data, c, options...))
}

func Failure(c *gin.Context, err error) {
	c.JSON(http.StatusOK, HandleBadReturn(err, nil))
}

func Response(c *gin.Context, data interface{}, err error) {
	if err != nil {
		Failure(c, err)
	} else {
		Success(c, data)
	}
}

func QueryFailure(c *gin.Context, err error) {
	c.JSON(http.StatusOK, QueryHandleBadReturn(err, nil))
}

type RespOption func(m map[string]interface{})

func SuccessReturn(data interface{}, c *gin.Context, options ...RespOption) map[string]interface{} {
	var info = make(map[string]interface{})
	info["code"] = 0
	info["msg"] = "success"
	info["nowTime"] = time.Now().Unix()
	info["data"] = data
	for _, option := range options {
		option(info)
	}
	return info
}

func HandleBadReturn(err error, data interface{}) map[string]interface{} {
	var info = make(map[string]interface{})
	info["code"] = -1
	info["msg"] = err.Error()
	info["nowTime"] = time.Now().Unix()
	info["data"] = data
	return info
}

func QueryHandleBadReturn(err error, data interface{}) map[string]interface{} {
	var info = make(map[string]interface{})
	info["code"] = -1
	info["msg"] = "failed"
	info["nowTime"] = time.Now().Unix()
	info["err"] = err.Error()
	info["data"] = data
	return info
}
