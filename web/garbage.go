package web

import (
	"crypto/rand"
	"fmt"
	"net/http"
	"speedtest/config"
	"strconv"

	"github.com/gin-gonic/gin"
)

const (
	// defaultChunkSize 块尺寸为 1 MiB
	defaultChunkSize = 1 * 1024 * 1024
	defaultChunks    = 4    // 默认 chunk 数量
	maxChunks        = 1024 // 最大允许的 chunk 数量，防止滥用
)

var (
	// 随机数据块
	dlChunkSize int
	dlChunks    int
	randomData  []byte // 声明为 slice，在 init 中初始化
)

// getRandomData 生成指定大小的随机数据块
func getRandomData(length int) []byte {
	data := make([]byte, length)
	_, err := rand.Read(data) // 使用 crypto/rand.Read 获取随机数据，返回读取的字节数和 error
	if err != nil {
		logError("Failed to generate random data: %v", err)
		//  如果随机数据生成失败，返回 nil 或 空 slice，并由调用方处理错误
		return nil //  返回 nil，表示生成失败
	}
	return data
}

// RandomDataInit 初始化随机数据块，在程序启动时调用
func RandomDataInit(cfg *config.Config) {
	dlChunkSize = defaultChunkSize
	if cfg.Speedtest.DownDataChunkSize > 0 {
		dlChunkSize = cfg.Speedtest.DownDataChunkSize * 1024 * 1024
	}

	dlChunks = defaultChunks
	if cfg.Speedtest.DownDataChunkCount > 0 {
		dlChunks = cfg.Speedtest.DownDataChunkCount
	}

	randomData = getRandomData(dlChunkSize) // 初始化 randomData
	if randomData == nil {                  // 检查 randomData 是否生成成功
		logError("Failed to initialize random data. Program cannot continue.")
		//  panic 退出程序，因为依赖的随机数据无法生成
		panic("Failed to initialize random data. Program cannot continue.")
	}

	fmt.Printf("RandomDataInit: dlChunkSize=%d, dlChunks=%d\n", dlChunkSize, dlChunks)
}

// garbage 处理对 /garbage 的请求，返回指定数量的随机数据块
func garbage(c *gin.Context) {
	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Type", "application/octet-stream")
	c.Header("Content-Disposition", "attachment; filename=random.dat")
	c.Header("Content-Transfer-Encoding", "binary")

	chunks := dlChunks // 默认 chunk 数量

	ckSizeStr := c.Query("ckSize")
	if ckSizeStr != "" {
		ckSize, err := strconv.ParseInt(ckSizeStr, 10, 64)
		if err != nil {
			c.String(http.StatusBadRequest, "Invalid ckSize parameter: "+err.Error()) // 返回 400 错误，告知客户端参数错误和具体错误信息
			return
		}

		if ckSize > maxChunks {
			chunks = maxChunks
		} else if ckSize > 0 {
			chunks = int(ckSize)
		} else {
			c.String(http.StatusBadRequest, "ckSize must be greater than 0") // 返回 400 错误，告知客户端参数错误
			return
		}
	}

	//  检查 randomData 是否为空，防止在初始化失败的情况下继续运行
	if randomData == nil {
		c.AbortWithError(http.StatusInternalServerError, fmt.Errorf("random data not initialized")) // 返回 500 错误
		return
	}

	// 发送随机数据块
	for i := 0; i < chunks; i++ {
		_, err := c.Writer.Write(randomData)
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to write chunk %d: %w", i, err)) // 包含 chunk 索引，方便调试
			return
		}
	}
}
