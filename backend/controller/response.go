package controller

import "github.com/gin-gonic/gin"

// APIResponse 统一接口响应结构。
type APIResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

// Success 返回成功响应。
func Success(c *gin.Context, data interface{}) {
	c.JSON(200, APIResponse{
		Code:    0,
		Message: "success",
		Data:    data,
	})
}

// Error 返回错误响应。
func Error(c *gin.Context, code int, message string) {
	c.JSON(200, APIResponse{
		Code:    code,
		Message: message,
		Data:    nil,
	})
}
