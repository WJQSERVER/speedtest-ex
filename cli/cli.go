package main

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/tls"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"path"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/briandowns/spinner" // 用于动态加载提示
	"github.com/fatih/color"        // 用于彩色输出
	"github.com/schollz/progressbar/v3"
)

// --- 1. 国际化 (i18n) 文本管理 ---

// Messages 结构体包含所有需要本地化的UI文本
type Messages struct {
	LangCode                   string
	FlagLangUsage              string
	FlagServerUsage            string
	FlagBasePathUsage          string
	FlagTimeoutUsage           string
	FlagDurationUsage          string
	FlagDownloadUsage          string
	FlagUploadUsage            string
	FlagBothUsage              string
	FlagUploadSizeUsage        string
	FlagDownloadChunksUsage    string
	FlagShowProgressUsage      string
	FlagShowVersionUsage       string
	FlagSkipTLSVerifyUsage     string
	StageTitleIP               string
	StageTitleDownload         string
	StageTitleUpload           string
	StageTitleLatency          string
	StageTitleSummary          string
	Fetching                   string
	YourIP                     string
	ServerVersion              string
	ModeDurationBased          string
	ModeSizeBased              string
	ModeChunkBased             string
	ModeServerDefault          string
	ProgressBarDownloading     string
	ProgressBarUploading       string
	ResultDownloadSpeed        string
	ResultUploadSpeed          string
	ResultLatencyHint          string
	Error                      string
	ErrorFetchIP               string
	ErrorFetchVersion          string
	ErrorDownloadTest          string
	ErrorUploadTest            string
	WarningNoDataTransferred   string
	TestInterrupted            string
	OverallTimeout             string
	NoTestSelected             string
	SummaryJSON                string
	OperationCancelled         string
	OperationTimedOut          string
	FinalError                 string
	FinalErrorWithDetails      string
	UploadTemplateInitFailed   string
	UploadSizeMustBePositive   string
	UploadTemplateChunkNotInit string
	DownloadNoData             string
	DownloadDurationZero       string
	UploadNoData               string
	UploadDurationZero         string
	InvalidServerURL           string
	EmptyServerURL             string
}

// 全局变量, 保存当前语言的文本
var msg *Messages

// loadMessages 根据语言标志加载对应的文本
func loadMessages(lang string) {
	switch strings.ToLower(lang) {
	case "en":
		msg = &enMessages
	default: // 默认中文
		msg = &zhMessages
	}
}

// 中文文本
var zhMessages = Messages{
	LangCode:                   "zh",
	FlagLangUsage:              "设置显示语言 ('zh' 或 'en')",
	FlagServerUsage:            "测速服务器的URL",
	FlagBasePathUsage:          "服务器上的API基础路径",
	FlagTimeoutUsage:           "整体操作的超时时间 (秒)",
	FlagDurationUsage:          "每个测试的持续时间 (秒). 0 表示基于大小/块.",
	FlagDownloadUsage:          "执行下载测试",
	FlagUploadUsage:            "执行上传测试",
	FlagBothUsage:              "执行下载和上传测试",
	FlagUploadSizeUsage:        "目标上传大小 (MB) (当持续时间为0时)",
	FlagDownloadChunksUsage:    "下载的块数量 (当持续时间为0时)",
	FlagShowProgressUsage:      "显示进度条",
	FlagShowVersionUsage:       "显示服务器版本并退出",
	FlagSkipTLSVerifyUsage:     "跳过TLS证书验证",
	StageTitleIP:               " IP 信息 ",
	StageTitleDownload:         " 下载测试 ",
	StageTitleUpload:           " 上传测试 ",
	StageTitleLatency:          " 延迟测试 ",
	StageTitleSummary:          " 测试总结 ",
	Fetching:                   "获取中",
	YourIP:                     "您的 IP",
	ServerVersion:              "服务器版本",
	ModeDurationBased:          "模式: 基于时长 (%d 秒)",
	ModeSizeBased:              "模式: 基于大小 (%d MB)",
	ModeChunkBased:             "模式: 基于块 (%d 块)",
	ModeServerDefault:          "模式: 服务器默认",
	ProgressBarDownloading:     "下载中",
	ProgressBarUploading:       "上传中",
	ResultDownloadSpeed:        "下载速度",
	ResultUploadSpeed:          "上传速度",
	ResultLatencyHint:          "请使用 ping 命令测试延迟:",
	Error:                      "错误",
	ErrorFetchIP:               "获取客户端IP失败",
	ErrorFetchVersion:          "获取服务器版本失败",
	ErrorDownloadTest:          "下载测试出错",
	ErrorUploadTest:            "上传测试出错",
	WarningNoDataTransferred:   "警告: 已达到测试时长, 但未传输任何数据.",
	TestInterrupted:            "测试被用户中断. 正在清理...",
	OverallTimeout:             "整体操作在 %d 秒后超时.",
	NoTestSelected:             "未选择任何测试. 请使用 -dl, -ul, 或 -both. 或使用 -v 获取版本.",
	SummaryJSON:                "--- 总结 (JSON) ---",
	OperationCancelled:         "操作被用户取消",
	OperationTimedOut:          "操作超时",
	FinalError:                 "最终错误",
	FinalErrorWithDetails:      "操作也受到以下影响: %s",
	UploadTemplateInitFailed:   "致命错误: 初始化上传模板块失败: %v",
	UploadSizeMustBePositive:   "基于大小的测试模式, 上传大小必须为正数",
	UploadTemplateChunkNotInit: "上传模板块未初始化, 无法执行上传测试",
	DownloadNoData:             "没有下载任何数据",
	DownloadDurationZero:       "下载时长为零或负数但有数据",
	UploadNoData:               "没有上传任何数据",
	UploadDurationZero:         "上传时长为零或负数但有数据",
	InvalidServerURL:           "无效的服务器URL",
	EmptyServerURL:             "服务器URL不能为空",
}

// 英文文本
var enMessages = Messages{
	LangCode:                   "en",
	FlagLangUsage:              "Set display language ('zh' or 'en')",
	FlagServerUsage:            "Speedtest server URL",
	FlagBasePathUsage:          "API base path on the server",
	FlagTimeoutUsage:           "Overall operation timeout in seconds",
	FlagDurationUsage:          "Duration of each test in seconds. 0 for size/chunk based.",
	FlagDownloadUsage:          "Perform download test",
	FlagUploadUsage:            "Perform upload test",
	FlagBothUsage:              "Perform both download and upload tests",
	FlagUploadSizeUsage:        "Target upload size in MB (if duration is 0)",
	FlagDownloadChunksUsage:    "Number of chunks for download (if duration is 0)",
	FlagShowProgressUsage:      "Show progress bar",
	FlagShowVersionUsage:       "Show server version and exit",
	FlagSkipTLSVerifyUsage:     "Skip TLS certificate verification",
	StageTitleIP:               " IP Information ",
	StageTitleDownload:         " Download Test ",
	StageTitleUpload:           " Upload Test ",
	StageTitleLatency:          " Latency Test ",
	StageTitleSummary:          " Test Summary ",
	Fetching:                   "Fetching",
	YourIP:                     "Your IP",
	ServerVersion:              "Server Version",
	ModeDurationBased:          "Mode: Duration-based (%d seconds)",
	ModeSizeBased:              "Mode: Size-based (%d MB)",
	ModeChunkBased:             "Mode: Chunk-based (%d chunks)",
	ModeServerDefault:          "Mode: Server default",
	ProgressBarDownloading:     "Downloading",
	ProgressBarUploading:       "Uploading",
	ResultDownloadSpeed:        "Download Speed",
	ResultUploadSpeed:          "Upload Speed",
	ResultLatencyHint:          "For latency, please use the ping command:",
	Error:                      "Error",
	ErrorFetchIP:               "Failed to get client IP",
	ErrorFetchVersion:          "Failed to get server version",
	ErrorDownloadTest:          "Download test error",
	ErrorUploadTest:            "Upload test error",
	WarningNoDataTransferred:   "Warning: Test duration reached, but no data was transferred.",
	TestInterrupted:            "Test interrupted by user. Cleaning up...",
	OverallTimeout:             "Overall operation timed out after %d seconds.",
	NoTestSelected:             "No test selected. Use -dl, -ul, or -both. Or -v for version.",
	SummaryJSON:                "--- Summary (JSON) ---",
	OperationCancelled:         "Operation cancelled by user",
	OperationTimedOut:          "Operation timed out",
	FinalError:                 "Final error",
	FinalErrorWithDetails:      "Operation also affected by: %s",
	UploadTemplateInitFailed:   "FATAL: Failed to initialize upload template chunk: %v",
	UploadSizeMustBePositive:   "Upload size must be positive for size-based test",
	UploadTemplateChunkNotInit: "Upload template chunk not initialized, cannot perform upload test",
	DownloadNoData:             "No data downloaded",
	DownloadDurationZero:       "Download duration zero/negative with data",
	UploadNoData:               "No data uploaded",
	UploadDurationZero:         "Upload duration zero/negative with data",
	InvalidServerURL:           "Invalid server URL",
	EmptyServerURL:             "Server URL cannot be empty",
}

// --- 2. 核心逻辑与常量 ---

const (
	defaultUserAgent        = "SimpleSpeedtestCLI/1.3 (Go)"
	apiPathBackend          = "/backend"
	apiPathPublic           = "/api"
	uploadTemplateChunkSize = 64 * 1024
)

var (
	// Flags
	langFlag          = flag.String("lang", "zh", "Set display language ('zh' or 'en')")
	serverURLFlag     = flag.String("s", "http://localhost:8989", "Speedtest server URL")
	basePathFlag      = flag.String("base-path", "", "API base path on the server")
	timeoutFlag       = flag.Int("t", 70, "Overall operation timeout in seconds")
	durationFlag      = flag.Int("d", 10, "Duration of each test in seconds. 0 for size/chunk based.")
	downloadFlag      = flag.Bool("dl", false, "Perform download test")
	uploadFlag        = flag.Bool("ul", false, "Perform upload test")
	bothTestsFlag     = flag.Bool("both", false, "Perform both download and upload tests")
	uploadSizeMBFlag  = flag.Int("ul-size", 10, "Target upload size in MB (if duration is 0)")
	dlChunksFlag      = flag.Int("dl-chunks", 0, "Number of chunks for download (if duration is 0)")
	showProgressFlag  = flag.Bool("p", true, "Show progress bar")
	showVersionFlag   = flag.Bool("v", false, "Show server version and exit")
	skipTLSVerifyFlag = flag.Bool("skip-tls-verify", false, "Skip TLS certificate verification")

	// Global
	httpClient          *http.Client
	apiPrefix           string
	serverHost          string
	uploadTemplateChunk []byte
	currentSpeed        atomic.Value // 用于在进度条旁显示实时速度
)

type SpeedTestData struct {
	ClientIP          string  `json:"client_ip,omitempty"`
	DownloadSpeedMBps float64 `json:"download_speed_mbps,omitempty"`
	UploadSpeedMBps   float64 `json:"upload_speed_mbps,omitempty"`
	ServerVersion     string  `json:"server_version,omitempty"`
	Error             string  `json:"error,omitempty"`
}

// ClientIPInfo 用于解析 getIP 接口返回的JSON
type ClientIPInfo struct {
	ProcessedString string `json:"processedString"`
	RawIspInfo      struct {
		IP string `json:"ip"`
	} `json:"rawIspInfo"`
}

// --- 初始化与帮助函数 ---

func init() {
	// 在解析flag之前加载语言, 以便flag的帮助信息是正确的语言
	// 这是一个小技巧: 提前扫描 os.Args 找 -lang
	for i, arg := range os.Args {
		if (arg == "-lang" || arg == "--lang") && i+1 < len(os.Args) {
			loadMessages(os.Args[i+1])
			return
		}
	}
	loadMessages("zh") // 默认
}

func initUploadTemplateChunk() {
	uploadTemplateChunk = make([]byte, uploadTemplateChunkSize)
	_, err := rand.Read(uploadTemplateChunk)
	if err != nil {
		color.Red(msg.UploadTemplateInitFailed, err)
		uploadTemplateChunk = nil
	}
}

func initHttpClient() {
	tr := &http.Transport{
		TLSClientConfig:   &tls.Config{InsecureSkipVerify: *skipTLSVerifyFlag},
		DialContext:       (&net.Dialer{Timeout: 15 * time.Second, KeepAlive: 30 * time.Second}).DialContext,
		ForceAttemptHTTP2: true, MaxIdleConns: 10, IdleConnTimeout: 60 * time.Second,
		TLSHandshakeTimeout: 15 * time.Second, ResponseHeaderTimeout: 20 * time.Second,
	}
	httpClient = &http.Client{Transport: tr}
}

func determineApiPrefixAndHost() error {
	if *serverURLFlag == "" {
		return fmt.Errorf(msg.EmptyServerURL)
	}
	parsedURL, err := url.Parse(*serverURLFlag)
	if err != nil {
		return fmt.Errorf("%s: %w", msg.InvalidServerURL, err)
	}
	serverHost = parsedURL.Hostname()
	if *basePathFlag != "" && *basePathFlag != "/" {
		apiPrefix = strings.TrimSuffix(*basePathFlag, "/")
	} else {
		apiPrefix = ""
	}
	return nil
}

func buildURL(isPublicAPI bool, endpointParts ...string) string {
	var currentApiPrefix string
	if apiPrefix != "" {
		currentApiPrefix = apiPrefix
	} else if isPublicAPI {
		currentApiPrefix = apiPathPublic
	} else {
		currentApiPrefix = apiPathBackend
	}
	fullPath := path.Join(append([]string{currentApiPrefix}, endpointParts...)...)
	if !strings.HasPrefix(fullPath, "/") {
		fullPath = "/" + fullPath
	}
	return strings.TrimSuffix(*serverURLFlag, "/") + fullPath
}

// 美化输出的辅助函数
func printStageTitle(title string) {
	color.New(color.BgCyan, color.FgBlack).Printf(" %s ", title)
	fmt.Println()
}

func printResult(label, value string) {
	color.New(color.FgHiWhite).Printf("  %-20s ", label+":")
	color.New(color.FgHiGreen, color.Bold).Println(value)
}

/*
func formatSpeed(mbps float64) string {
	bps := mbps * 1000 * 1000 * 8
	if bps < 1000 {
		return fmt.Sprintf("%.2f bps", bps)
	}
	if bps < 1000*1000 {
		return fmt.Sprintf("%.2f Kbps", bps/1000)
	}
	if bps < 1000*1000*1000 {
		return fmt.Sprintf("%.2f Mbps", bps/(1000*1000))
	}
	return fmt.Sprintf("%.2f Gbps", bps/(1000*1000*1000))
}*/
// formatSpeed 接收一个以 Mbps (兆比特/秒) 为单位的值, 并将其格式化为可读的字符串
func formatSpeed(mbps float64) string {
	if mbps < 1 {
		return fmt.Sprintf("%.2f Kbps", mbps*1000)
	}
	if mbps < 1000 {
		return fmt.Sprintf("%.2f Mbps", mbps)
	}
	return fmt.Sprintf("%.2f Gbps", mbps/1000)
}

// 动态 spinner
func newSpinner(message string) *spinner.Spinner {
	s := spinner.New(spinner.CharSets[14], 100*time.Millisecond, spinner.WithWriter(os.Stderr))
	s.Suffix = " " + message + "..."
	s.Color("cyan")
	return s
}

// --- 核心测速逻辑 ---

func getServerVersion(ctx context.Context) (string, error) {
	s := newSpinner(msg.Fetching + " " + msg.ServerVersion)
	s.Start()
	defer s.Stop()
	req, err := http.NewRequestWithContext(ctx, "GET", buildURL(true, "version"), nil)
	if err != nil {
		return "", err
	}
	// ... (rest of the logic is same)
	req.Header.Set("User-Agent", defaultUserAgent)
	resp, err := httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("server status %s", resp.Status)
	}
	var vi struct {
		Version string `json:"Version"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&vi); err != nil {
		return "", err
	}
	return vi.Version, nil
}

func getClientIP(ctx context.Context) (*ClientIPInfo, error) {
	s := newSpinner(msg.Fetching + " " + msg.YourIP)
	s.Start()
	defer s.Stop()

	req, err := http.NewRequestWithContext(ctx, "GET", buildURL(false, "getIP"), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", defaultUserAgent)

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server status %s", resp.Status)
	}

	var ipInfo ClientIPInfo
	// 尝试解析JSON
	if err := json.NewDecoder(resp.Body).Decode(&ipInfo); err != nil {
		// 如果解析失败, 可能是旧版后端, 尝试作为纯文本读取
		// (这是一个增强兼容性的后备方案)
		bodyBytes, readErr := io.ReadAll(io.MultiReader(json.NewDecoder(resp.Body).Buffered(), resp.Body))
		if readErr == nil {
			ipStr := strings.TrimSpace(string(bodyBytes))
			if ipStr != "" {
				// 手动构建结构体
				return &ClientIPInfo{
					ProcessedString: ipStr,
					RawIspInfo: struct {
						IP string `json:"ip"`
					}{IP: ipStr},
				}, nil
			}
		}
		return nil, err // 如果后备方案也失败, 则返回原始错误
	}

	// 如果 IP 字段为空, 但 ProcessedString 存在, 尝试从中提取
	if ipInfo.RawIspInfo.IP == "" && ipInfo.ProcessedString != "" {
		parts := strings.Split(ipInfo.ProcessedString, " ")
		if len(parts) > 0 {
			ipInfo.RawIspInfo.IP = parts[0]
		}
	}

	return &ipInfo, nil
}

func performDownloadTest(mainCtx context.Context) (float64, error) {
	printStageTitle(msg.StageTitleDownload)
	// ... (The provided download logic is already quite good, we'll just beautify the output and progress bar)
	targetURL := buildURL(false, "garbage")
	useDurationMode := *durationFlag > 0

	if !useDurationMode && *dlChunksFlag > 0 {
		targetURL = fmt.Sprintf("%s?ckSize=%d", targetURL, *dlChunksFlag)
		color.Cyan("  " + fmt.Sprintf(msg.ModeChunkBased, *dlChunksFlag))
	} else if useDurationMode {
		color.Cyan("  " + fmt.Sprintf(msg.ModeDurationBased, *durationFlag))
	} else {
		color.Cyan("  " + msg.ModeServerDefault)
	}

	var testCtx context.Context
	var testCancel context.CancelFunc
	if useDurationMode {
		testCtx, testCancel = context.WithTimeout(mainCtx, time.Duration(*durationFlag)*time.Second)
	} else {
		testCtx, testCancel = context.WithCancel(mainCtx)
	}
	defer testCancel()

	req, err := http.NewRequestWithContext(testCtx, "GET", targetURL, nil)
	if err != nil {
		return 0, err
	}
	req.Header.Set("User-Agent", defaultUserAgent)

	start := time.Now()
	resp, err := httpClient.Do(req)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) && useDurationMode { /* expected */
		} else if errors.Is(err, context.Canceled) {
			return 0, err
		} else {
			return 0, fmt.Errorf("http client Do error: %w", err)
		}
	}
	if resp != nil {
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return 0, fmt.Errorf("server status %s for download", resp.Status)
		}
	}

	var bar *progressbar.ProgressBar
	if *showProgressFlag {
		bar = progressbar.NewOptions64(
			-1, // Always indeterminate for better live speed display
			progressbar.OptionSetWriter(os.Stderr),
			progressbar.OptionShowBytes(true),
			progressbar.OptionSetWidth(25),
			progressbar.OptionThrottle(100*time.Millisecond),
			progressbar.OptionOnCompletion(func() { fmt.Fprint(os.Stderr, "\n") }),
			progressbar.OptionSpinnerType(14),
			progressbar.OptionClearOnFinish(),
			progressbar.OptionSetDescription(fmt.Sprintf("[cyan]%s[reset]", msg.ProgressBarDownloading)),
			progressbar.OptionSetRenderBlankState(true),
		)
	}

	var written int64
	var readerForCopy io.Reader
	if resp != nil && resp.Body != nil {
		readerForCopy = resp.Body
	} else {
		readerForCopy = bytes.NewReader(nil)
	}

	// Ticker for updating speed on the progress bar
	ticker := time.NewTicker(250 * time.Millisecond)
	defer ticker.Stop()
	go func() {
		for range ticker.C {
			if bar != nil && !bar.IsFinished() {
				currentWritten := atomic.LoadInt64(&written)
				elapsed := time.Since(start).Seconds()
				if elapsed > 0 {
					speedMBps := (float64(currentWritten) / elapsed) / (1000 * 1000)
					bar.Describe(fmt.Sprintf("[cyan]%s[reset] %s", msg.ProgressBarDownloading, formatSpeed(speedMBps)))
				}
			}
		}
	}()

	if bar != nil {
		pr := progressbar.NewReader(readerForCopy, bar)
		readerForCopy = &pr
	}

	// The rest of the download loop is very similar, but we use atomic for 'written'
	buf := make([]byte, 64*1024)
	readLoopErr := error(nil)
	for {
		select {
		case <-testCtx.Done():
			readLoopErr = testCtx.Err()
			goto endReadLoop
		default:
		}
		n, rErr := readerForCopy.Read(buf)
		if n > 0 {
			atomic.AddInt64(&written, int64(n))
		}
		if rErr != nil {
			if rErr != io.EOF {
				readLoopErr = rErr
			}
			goto endReadLoop
		}
	}
endReadLoop:
	if bar != nil && !bar.IsFinished() {
		_ = bar.Finish()
	}

	actualDuration := time.Since(start)
	finalWritten := atomic.LoadInt64(&written)

	if finalWritten == 0 {
		if useDurationMode && (errors.Is(readLoopErr, context.DeadlineExceeded) || errors.Is(err, context.DeadlineExceeded)) {
			color.Yellow(msg.WarningNoDataTransferred)
			return 0, nil
		}
		if errors.Is(readLoopErr, context.Canceled) || errors.Is(err, context.Canceled) {
			return 0, readLoopErr
		}
		return 0, fmt.Errorf(msg.DownloadNoData)
	}
	if actualDuration <= 0 {
		return 0, fmt.Errorf(msg.DownloadDurationZero)
	}

	//speedMBps := (float64(finalWritten) / actualDuration.Seconds()) / (1000 * 1000)
	//return speedMBps, nil
	speedMbps := (float64(finalWritten) * 8 / actualDuration.Seconds()) / (1000 * 1000)
	return speedMbps, nil
}

func performUploadTest(mainCtx context.Context) (float64, error) {
	if uploadTemplateChunk == nil {
		return 0, fmt.Errorf(msg.UploadTemplateChunkNotInit)
	}
	printStageTitle(msg.StageTitleUpload)
	// ... (The provided upload logic is also great, we'll beautify its output)
	targetURL := buildURL(false, "empty")
	useDurationMode := *durationFlag > 0
	uploadTargetSizeBytes := int64(*uploadSizeMBFlag) * 1000 * 1000

	if useDurationMode {
		color.Cyan("  " + fmt.Sprintf(msg.ModeDurationBased, *durationFlag))
	} else {
		if *uploadSizeMBFlag <= 0 {
			return 0, fmt.Errorf(msg.UploadSizeMustBePositive)
		}
		color.Cyan("  " + fmt.Sprintf(msg.ModeSizeBased, *uploadSizeMBFlag))
	}

	var testCtx context.Context
	var testCancel context.CancelFunc
	if useDurationMode {
		testCtx, testCancel = context.WithTimeout(mainCtx, time.Duration(*durationFlag)*time.Second)
	} else {
		testCtx, testCancel = context.WithCancel(mainCtx)
	}
	defer testCancel()

	pipeReader, pipeWriter := io.Pipe()
	var bytesWrittenByGoroutine int64
	errChan := make(chan error, 1)

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		defer pipeWriter.Close()

		var bar *progressbar.ProgressBar
		if *showProgressFlag {
			totalSizeForBar := uploadTargetSizeBytes
			if useDurationMode {
				totalSizeForBar = -1
			}
			bar = progressbar.NewOptions64(
				totalSizeForBar,
				progressbar.OptionSetWriter(os.Stderr),
				progressbar.OptionShowBytes(true),
				progressbar.OptionSetWidth(25),
				progressbar.OptionThrottle(100*time.Millisecond),
				progressbar.OptionOnCompletion(func() { fmt.Fprint(os.Stderr, "\n") }),
				progressbar.OptionSpinnerType(14),
				progressbar.OptionClearOnFinish(),
				progressbar.OptionSetDescription(fmt.Sprintf("[cyan]%s[reset]", msg.ProgressBarUploading)),
				progressbar.OptionSetRenderBlankState(true),
			)
		}

		countingPipeWriter := &countingWriter{writer: pipeWriter, counter: &bytesWrittenByGoroutine}
		var writerForProgress io.Writer
		if bar != nil {
			writerForProgress = io.MultiWriter(countingPipeWriter, bar)
		} else {
			writerForProgress = countingPipeWriter
		}

		currentChunk := make([]byte, uploadTemplateChunkSize)
		for {
			select {
			case <-testCtx.Done():
				errChan <- testCtx.Err()
				return
			default:
			}
			copy(currentChunk, uploadTemplateChunk)
			_, wErr := writerForProgress.Write(currentChunk)
			if wErr != nil {
				if errors.Is(wErr, io.ErrClosedPipe) && testCtx.Err() != nil {
					errChan <- testCtx.Err()
				} else {
					errChan <- fmt.Errorf("error writing to upload pipe: %w", wErr)
				}
				return
			}
			if !useDurationMode && atomic.LoadInt64(&bytesWrittenByGoroutine) >= uploadTargetSizeBytes {
				break
			}
		}
		errChan <- nil
	}()

	req, err := http.NewRequestWithContext(testCtx, "POST", targetURL, pipeReader)
	if err != nil {
		_ = pipeReader.CloseWithError(err)
		wg.Wait()
		return 0, err
	}
	req.Header.Set("User-Agent", defaultUserAgent)
	req.Header.Set("Content-Type", "application/octet-stream")
	if !useDurationMode {
		req.ContentLength = uploadTargetSizeBytes
	} else {
		req.Header.Set("Transfer-Encoding", "chunked")
	}

	start := time.Now()
	resp, err := httpClient.Do(req)
	_ = pipeReader.Close()
	wg.Wait()

	var goroutineErr error
	select {
	case goroutineErr = <-errChan:
	default:
	}

	if err != nil { // HTTP client error
		if errors.Is(err, context.Canceled) {
			return 0, err
		}
		if !(errors.Is(err, context.DeadlineExceeded) && useDurationMode) {
			return 0, err
		}
	}
	if resp != nil {
		defer resp.Body.Close()
		_, _ = io.Copy(io.Discard, resp.Body)
	}

	actualDuration := time.Since(start)
	actualUploadedBytes := atomic.LoadInt64(&bytesWrittenByGoroutine)

	if actualUploadedBytes == 0 {
		if useDurationMode && (testCtx.Err() == context.DeadlineExceeded) {
			color.Yellow(msg.WarningNoDataTransferred)
			return 0, nil
		}
		if errors.Is(err, context.Canceled) || errors.Is(goroutineErr, context.Canceled) {
			return 0, context.Canceled
		}
		return 0, fmt.Errorf(msg.UploadNoData)
	}
	if actualDuration <= 0 {
		return 0, fmt.Errorf(msg.UploadDurationZero)
	}

	//speedMBps := (float64(actualUploadedBytes) / actualDuration.Seconds()) / (1000 * 1000)
	//return speedMBps, nil
	speedMbps := (float64(actualUploadedBytes) * 8 / actualDuration.Seconds()) / (1000 * 1000)
	return speedMbps, nil
}

type countingWriter struct {
	writer  io.Writer
	counter *int64
}

func (cw *countingWriter) Write(p []byte) (n int, err error) {
	n, err = cw.writer.Write(p)
	atomic.AddInt64(cw.counter, int64(n))
	return
}

// --- main 函数 ---

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Speedtest-EX CLI\n\n")
		flag.PrintDefaults()
	}
	flag.Parse()

	loadMessages(*langFlag)
	initUploadTemplateChunk()
	initHttpClient()
	if err := determineApiPrefixAndHost(); err != nil {
		color.Red("%s: %v", msg.Error, err)
		os.Exit(1)
	}

	mainCtx, mainCancel := context.WithTimeout(context.Background(), time.Duration(*timeoutFlag)*time.Second)
	defer mainCancel()

	sigCtx, stopSignalHandler := signal.NotifyContext(mainCtx, syscall.SIGINT, syscall.SIGTERM)
	defer stopSignalHandler()

	results := SpeedTestData{}
	var performedAction bool

	if *showVersionFlag {
		performedAction = true
		version, err := getServerVersion(sigCtx)
		if err != nil {
			results.Error = fmt.Sprintf("%s: %v", msg.ErrorFetchVersion, err)
			color.Red(results.Error)
		} else {
			results.ServerVersion = version
			printResult(msg.ServerVersion, version)
		}
		// ... exit logic remains similar
		return
	}

	// IP Info Stage
	printStageTitle(msg.StageTitleIP)
	ipInfo, err := getClientIP(sigCtx)
	if err != nil {
		color.Red("  %s: %v", msg.ErrorFetchIP, err)
		if results.Error != "" {
			results.Error += "; "
		}
		results.Error += fmt.Sprintf("%s: %v", msg.ErrorFetchIP, err)
	} else {
		// 在最终的JSON总结中, 只保存纯IP地址
		results.ClientIP = ipInfo.RawIspInfo.IP
		// 在屏幕上, 打印更友好的 processedString
		printResult(msg.YourIP, ipInfo.ProcessedString)
	}
	fmt.Println()

	// Latency Stage
	printStageTitle(msg.StageTitleLatency)
	color.HiWhite("  " + msg.ResultLatencyHint)
	color.HiYellow("  ping %s\n\n", serverHost)

	// Tests
	if *downloadFlag || *bothTestsFlag {
		performedAction = true
		if sigCtx.Err() == nil {
			dlSpeed, err := performDownloadTest(sigCtx)
			if err != nil && !errors.Is(err, context.Canceled) {
				color.Red("  %s: %v", msg.ErrorDownloadTest, err)
				results.Error += fmt.Sprintf("%s: %v; ", msg.ErrorDownloadTest, err)
			} else if err == nil {
				results.DownloadSpeedMBps = dlSpeed
				printResult(msg.ResultDownloadSpeed, formatSpeed(dlSpeed))
			}
		}
	}
	if (*downloadFlag || *bothTestsFlag) && (*uploadFlag || *bothTestsFlag) {
		fmt.Println() // Add a space between tests
	}
	if *uploadFlag || *bothTestsFlag {
		performedAction = true
		if sigCtx.Err() == nil {
			ulSpeed, err := performUploadTest(sigCtx)
			if err != nil && !errors.Is(err, context.Canceled) {
				color.Red("  %s: %v", msg.ErrorUploadTest, err)
				results.Error += fmt.Sprintf("%s: %v; ", msg.ErrorUploadTest, err)
			} else if err == nil {
				results.UploadSpeedMBps = ulSpeed
				printResult(msg.ResultUploadSpeed, formatSpeed(ulSpeed))
			}
		}
	}
	fmt.Println()

	if !performedAction && sigCtx.Err() == nil {
		color.Yellow(msg.NoTestSelected)
		flag.Usage()
	}

	if sigCtx.Err() != nil {
		if errors.Is(sigCtx.Err(), context.DeadlineExceeded) {
			color.Red("\n" + fmt.Sprintf(msg.OverallTimeout, *timeoutFlag))
			results.Error += msg.OperationTimedOut
		} else if errors.Is(sigCtx.Err(), context.Canceled) {
			color.Yellow("\n" + msg.TestInterrupted)
			results.Error += msg.OperationCancelled
		}
	}

	// Summary Stage
	printStageTitle(msg.StageTitleSummary)
	jsonData, _ := json.MarshalIndent(results, "  ", "  ")
	fmt.Println("  " + string(jsonData))

	if results.Error != "" {
		os.Exit(1)
	}
}
