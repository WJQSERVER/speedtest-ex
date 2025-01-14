package web

import (
	"fmt"
	"net/http"
	"speedtest/config"
	"strconv"

	"github.com/gin-gonic/gin"
)

const (
	// 块尺寸为 1 MiB
	chunkSize = 1 * 1024 * 1024
	dfchunks  = 4
	//chunkSize = 16 * 1024 * 1024
)

var (
	// 随机数据块
	dlChunkSize int
	dlChunks    int
	randomData  = getRandomData(dlChunkSize)
)

// RandomDataInit initializes the random data blocks for speed testing based on configuration settings.
// It sets the chunk size and number of chunks for download data, either using provided configuration
// values or falling back to default settings. The function generates random data blocks with the
// specified chunk size and logs the initialization parameters.
//
// Parameters:
//   - cfg: A configuration object containing speedtest-related settings for chunk size and count.
//
// The function modifies global variables:
//   - dlChunkSize: Set to the configured chunk size in bytes, or default to chunkSize
//   - dlChunks: Set to the configured number of chunks, or default to dfchunks
//   - randomData: Generated random data blocks using the determined chunk size
//
// Example:
//   config := &config.Config{
//     Speedtest: config.SpeedtestConfig{
//       DownDataChunkSize: 2,   // 2 MiB chunks
//       DownDataChunkCount: 8,  // 8 chunks
//     },
//   }
//   RandomDataInit(config)
func RandomDataInit(cfg *config.Config) {
	if cfg.Speedtest.DownDataChunkSize != 0 {
		dlChunkSize = cfg.Speedtest.DownDataChunkSize * 1024 * 1024
	} else {
		dlChunkSize = chunkSize
	}
	if cfg.Speedtest.DownDataChunkCount != 0 {
		dlChunks = cfg.Speedtest.DownDataChunkCount
	} else {
		dlChunks = dfchunks
	}
	randomData = getRandomData(dlChunkSize)
	fmt.Printf("RandomDataInit: dlChunkSize=%d, dlChunks=%d\n", dlChunkSize, dlChunks)
}

// If an error occurs during data writing, the request is aborted with an internal server error.
func garbage(c *gin.Context) {
	// 设置响应头信息

	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Type", "application/octet-stream")
	c.Header("Content-Disposition", "attachment; filename=random.dat")
	c.Header("Content-Transfer-Encoding", "binary")

	// 默认chunk数量为4
	chunks := dlChunks

	// 从查询参数中获取ckSize
	ckSize := c.Query("ckSize")
	if ckSize != "" {
		// 尝试将ckSize转换为整数
		i, err := strconv.ParseInt(ckSize, 10, 64)
		if err != nil {
			return
		} else {
			// 如果转换成功，限制最大chunk数量为1024，并确保chunk数量大于0
			if i > 1024 {
				chunks = 1024
			} else if i > 0 {
				chunks = int(i)
			}
		}
	}

	// 发送随机数据块
	for i := 0; i < chunks; i++ {
		// 尝试写入随机数据到客户端
		if _, err := c.Writer.Write(randomData); err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
	}
}
