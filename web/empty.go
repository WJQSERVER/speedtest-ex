package web

import (
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
)

// empty 处理对/empty的请求，丢弃请求体并返回成功的状态码
func empty(c *gin.Context) {
	_, err := io.Copy(io.Discard, c.Request.Body)
	if err != nil {
		return
	}
	c.Status(http.StatusOK)
}
