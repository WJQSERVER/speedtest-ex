package web

import (
	"io"
	"net/http"
	"sync"

	"github.com/WJQSERVER-STUDIO/go-utils/copyb"
	"github.com/infinite-iroha/touka"
)

var (
	BufferPool *sync.Pool
	BufferSize int = 32 * 1024 // 32KB
)

func InitEmptyBuf() {
	// 初始化固定大小的缓存池
	BufferPool = &sync.Pool{
		New: func() interface{} {
			return make([]byte, BufferSize)
		},
	}
}

// empty 处理对/empty的请求，丢弃请求体并返回成功的状态码
func empty(c *touka.Context) {

	var err error
	/*
		// 使用固定32KB缓冲池
		buffer := BufferPool.Get().([]byte)
		defer BufferPool.Put(buffer)

		_, err = io.CopyBuffer(io.Discard, c.Request.Body, buffer)
		if err != nil {
			logWarning("empty > io.CopyBuffer error: %v", err)
			return
		}
		c.Status(http.StatusOK)
	*/

	//_, err = io.Copy(io.Discard, c.Request.Body)
	//if err != nil {
	//	return
	//}
	c.Status(http.StatusOK)

	_, err = copyb.Copy(io.Discard, c.Request.Body)
	if err != nil {
		return
	}
	c.Status(http.StatusOK)

	// for debug
	/*
		bodySize, err := io.Copy(io.Discard, c.Request.Body)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			return
		}
		c.Status(http.StatusOK)
		logInfo("empty > body size: %d", bodySize)
	*/
}
