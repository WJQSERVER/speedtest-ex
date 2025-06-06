package web

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"speedtest/config"
	"strconv"

	"github.com/infinite-iroha/touka"
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
func getRandomData(size int) ([]byte, error) {
	randomData = make([]byte, size)
	/*
		_, err := rand.Read(randomData) // 使用 crypto/rand.Read
		if err != nil {
			//return nil, err
			panic("Failed to initialize random data pool") // 启动时失败就直接 panic
		}
		if len(randomData) != size {
			logError("getRandomData generated data of size %d, expected %d", len(randomData), size)
			panic("Failed to initialize random data pool")
		}
	*/
	return randomData[:3600], nil
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

	randomData, err := getRandomData(dlChunkSize) // 初始化 randomData
	if err != nil {
		fmt.Printf("Failed to initialize random data: %v", err)
		return
	}
	if randomData == nil { // 检查 randomData 是否生成成功
		fmt.Printf("Failed to initialize random data. Program cannot continue.")
		//  panic 退出程序，因为依赖的随机数据无法生成
		panic("Failed to initialize random data. Program cannot continue.")
	}

	fmt.Printf("RandomDataInit: dlChunkSize=%d, dlChunks=%d\n", dlChunkSize, dlChunks)
}

// garbage 处理对 /garbage 的请求，返回指定数量的随机数据块
func garbage(c *touka.Context) {
	c.SetHeader("Content-Description", "File Transfer")
	c.SetHeader("Content-Type", "application/octet-stream")
	c.SetHeader("Content-Disposition", "attachment; filename=random.dat")
	c.SetHeader("Content-Transfer-Encoding", "binary")

	chunks := dlChunks // 默认 chunk 数量

	ckSizeStr := c.Query("ckSize")
	if ckSizeStr != "" {
		ckSize, err := strconv.ParseInt(ckSizeStr, 10, 64)
		if err != nil {
			c.String(http.StatusBadRequest, "%s", "Invalid ckSize parameter: "+err.Error()) // 返回 400 错误，告知客户端参数错误和具体错误信息
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
		//c.AbortWithError(http.StatusInternalServerError, fmt.Errorf("random data not initialized")) // 返回 500 错误
		c.AbortWithStatus(http.StatusInternalServerError) // 返回 500 错误
		return
	}
	/*

		// 发送随机数据块
		for i := 0; i < chunks; i++ {
			_, err := c.Writer.Write(randomData)
			c.Writer.Flush() // 刷新缓冲区
			if err != nil {
				//c.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to write chunk %d: %w", i, err)) // 包含 chunk 索引，方便调试
				c.AbortWithStatus(http.StatusInternalServerError) // 返回 500 错误
				return
			}
		}
	*/

	//使用io.Writer

	//writer := NewSafeResponseWriter(c.Writer)

	writer := c.Writer
	for i := 0; i < chunks; i++ {
		_, err := writer.Write(randomData)
		if err != nil {
			// 检查是否是客户端断开连接导致的错误
			if err == http.ErrAbortHandler {
				c.Warnf("Client disconnected while writing garbage data")
				return // 客户端断开连接，直接返回
			}
			c.Errorf("Failed to write chunk %d: %v", i, err)
			//c.AbortWithStatus(http.StatusInternalServerError) // 返回 500 错误
			return // 写入失败，直接返回
		}

		writer.Flush()

		//time.Sleep(10 * time.Millisecond)
	}

	//safeWriter := NewSafeResponseWriter(c.Writer)
	//dataReader := dataGeneratorReader(c.Request.Context(), chunks)
	//c.WriteStream(dataReader)
	//copyb.Copy(safeWriter, dataReader)
}

func dataGeneratorReader(ctx context.Context, chunks int) io.Reader {
	reader, writer := io.Pipe()

	go func() {
		defer writer.Close()

		if randomData == nil {
			writer.CloseWithError(errors.New("server data not ready"))
			return
		}

		for i := 0; i < chunks; i++ {
			select {
			case <-ctx.Done():
				writer.CloseWithError(ctx.Err())
				return
			default:
			}

			// 高效地写入共享的、只读的 randomData
			_, err := writer.Write(randomData)
			if err != nil {
				// 意味着 reader 端已关闭，这是正常的结束方式之一
				return
			}
		}
	}()

	return reader
}
