package web

import (
	"bytes"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"speedtest/config"
	"strconv"

	"github.com/WJQSERVER-STUDIO/go-utils/copyb"
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
	data := make([]byte, size)
	_, err := rand.Read(data) // 使用 crypto/rand.Read
	if err != nil {
		//return nil, err
		panic("Failed to initialize random data pool") // 启动时失败就直接 panic
	}
	if len(data) != size {
		panic(fmt.Sprintf("getRandomData generated data of size %d, expected %d", len(randomData), size))
	}

	return data, nil
}

// RandomDataInit 初始化随机数据块，在程序启动时调用
func RandomDataInit(cfg *config.Config) {
	var err error
	dlChunkSize = defaultChunkSize
	if cfg.Speedtest.DownDataChunkSize > 0 {
		dlChunkSize = cfg.Speedtest.DownDataChunkSize * 1024 * 1024
	}

	dlChunks = defaultChunks
	if cfg.Speedtest.DownDataChunkCount > 0 {
		dlChunks = cfg.Speedtest.DownDataChunkCount
	}

	randomData, err = getRandomData(dlChunkSize) // 初始化 randomData
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

	// 发送随机数据块
	//for i := 0; i < chunks; i++ {
	//	_, err := c.Writer.Write(randomData)
	//	c.Writer.Flush() // 刷新缓冲区
	//	if err != nil {
	//			//c.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to write chunk %d: %w", i, err)) // 包含 chunk 索引，方便调试
	//			c.AbortWithStatus(http.StatusInternalServerError) // 返回 500 错误
	//			return
	//	}
	//	}

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

		//time.Sleep(5 * time.Millisecond)
	}
}

var (
	// randomSeedData 是在程序启动时生成并常驻内存的随机数据"种子".
	// 所有下载测试都将重复使用这个数据块来生成流.
	randomSeedData []byte
)

// RandomDataInit 初始化随机数据种子, 必须在程序启动时调用一次.
func RandomDataInitStream(cfg *config.Config) {
	var err error
	dlChunkSize = defaultChunkSize
	if cfg.Speedtest.DownDataChunkSize > 0 {
		dlChunkSize = cfg.Speedtest.DownDataChunkSize * 1024 * 1024
	}

	dlChunks = defaultChunks
	if cfg.Speedtest.DownDataChunkCount > 0 {
		dlChunks = cfg.Speedtest.DownDataChunkCount
	}

	fmt.Printf("Initializing random data seed (%d bytes)...\n", dlChunkSize)
	// 使用 crypto/rand 生成高质量的随机种子
	seed := make([]byte, dlChunkSize)
	_, err = rand.Read(seed)
	if err != nil {
		// 如果在启动时无法生成种子数据, 这是一个致命错误, 程序无法继续.
		panic(fmt.Sprintf("FATAL: Failed to initialize random data seed: %v", err))
	}
	randomSeedData = seed
	fmt.Println("Random data seed initialized successfully.")
}

// garbage 处理下载请求, 高效地流式传输重复的随机数据.
func garbageStream(c *touka.Context) {
	// 设置标准的下载文件响应头
	c.SetHeader("Content-Description", "File Transfer")
	c.SetHeader("Content-Type", "application/octet-stream")
	c.SetHeader("Content-Disposition", "attachment; filename=random.dat")
	c.SetHeader("Content-Transfer-Encoding", "binary")

	// 定义默认值和限制 (与之前相同)
	const defaultChunks = 4         // 默认下载的块数量
	const maxChunks = 4096          // 可以适当调大最大值
	chunkSize := int64(dlChunkSize) // 每个块的大小就是我们种子的大小

	chunks := int64(defaultChunks)
	ckSizeStr := c.Query("ckSize")
	if ckSizeStr != "" {
		ckSize, err := strconv.ParseInt(ckSizeStr, 10, 64)
		if err != nil || ckSize <= 0 {
			c.String(http.StatusBadRequest, "Invalid ckSize parameter: must be a positive integer")
			return
		}
		chunks = ckSize
		if chunks > maxChunks {
			chunks = maxChunks
		}
	}

	// 计算总共需要发送的数据量 (字节)
	totalBytesToSend := chunkSize * chunks

	// --- 高性能流式处理的核心 ---

	// 1. 创建一个 readers 切片.
	// 我们需要重复种子数据 `chunks` 次, 所以创建 `chunks` 个指向种子的 Reader.
	readers := make([]io.Reader, chunks)
	for i := range readers {
		// bytes.NewReader 从一个字节切片创建一个 io.Reader.
		// 这个操作非常轻量, 它只是创建了一个指向内存中 `randomSeedData` 的视图.
		readers[i] = bytes.NewReader(randomSeedData)
	}

	// 2. 使用 io.MultiReader 将多个 Reader 连接成一个单一的逻辑流.
	// 当一个 Reader 被读完(返回EOF)时, MultiReader 会自动开始读取下一个.
	// 这就实现了一个重复 `randomSeedData` `chunks` 次的虚拟大文件流.
	repeatingStream := io.MultiReader(readers...)

	// 3. 使用 io.Copy 将这个逻辑流高效地写入HTTP响应.
	// 注意: 这里不再需要 io.LimitedReader, 因为 MultiReader 本身就是有限的.

	// 测试用代码, 限制速度, 避免回环CPU瓶颈造成的速度下降
	//limitSpeed, _ := toukautil.ParseRate("9gbps")
	//limitwriter := toukautil.NewRateLimitedWriter(c.Writer, limitSpeed, int(limitSpeed), c.Request.Context())
	//written, err := copyb.Copy(limitwriter, repeatingStream)

	written, err := copyb.Copy(c.Writer, repeatingStream)
	if err != nil {
		c.Warnf("Error during streaming garbage data: %v. Bytes written: %d", err, written)
		return
	}

	// (可选) 调试日志
	if written != totalBytesToSend {
		c.Warnf("Data stream truncated: expected to write %d bytes, but only wrote %d", totalBytesToSend, written)
	}
}
