package resp

import "github.com/gin-gonic/gin"

type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func JSON(c *gin.Context, httpCode int, code int, message string, data interface{}) {
	c.JSON(httpCode, Response{Code: code, Message: message, Data: data})
}

func OK(c *gin.Context, data interface{}) {
	JSON(c, 200, 0, "ok", data)
}

func Fail(c *gin.Context, httpCode int, message string) {
	JSON(c, httpCode, httpCode, message, nil)
}
