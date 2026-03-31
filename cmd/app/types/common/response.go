package common

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// BaseResponse is the base response
type BaseResponse struct {
	Code int         `json:"code"`
	Msg  string      `json:"message"`
	Data interface{} `json:"data"`
}

// Success 统一成功响应。
func Success(c *gin.Context, data interface{}) {
	Response(c, http.StatusOK, nil, data)
}

// Fail 统一失败响应。
// 对于 5xx 默认不透传内部错误详情，避免敏感信息泄漏。
func Fail(c *gin.Context, httpStatus int, err error) {
	if err == nil {
		err = errors.New(http.StatusText(httpStatus))
	}
	Response(c, httpStatus, err, nil)
}

// FailMessage 使用指定错误消息返回失败响应。
func FailMessage(c *gin.Context, httpStatus int, message string) {
	if strings.TrimSpace(message) == "" {
		message = http.StatusText(httpStatus)
	}
	Response(c, httpStatus, errors.New(message), nil)
}

// AbortFailMessage 中断请求并返回失败响应。
func AbortFailMessage(c *gin.Context, httpStatus int, message string) {
	c.Abort()
	FailMessage(c, httpStatus, message)
}

// Response generate response
func Response(c *gin.Context, httpStatus int, err error, data interface{}) {
	if err != nil {
		msg := strings.TrimSpace(err.Error())
		if msg == "" {
			msg = http.StatusText(httpStatus)
		}
		if httpStatus >= http.StatusInternalServerError {
			// 统一 5xx 响应，避免暴露内部实现细节。
			msg = "internal server error"
		}
		c.JSON(httpStatus, BaseResponse{
			Code: httpStatus,
			Msg:  msg,
			Data: nil,
		})
		return
	}
	c.JSON(httpStatus, BaseResponse{
		Code: httpStatus,
		Msg:  "success",
		Data: data,
	})
}