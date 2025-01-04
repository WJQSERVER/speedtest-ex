package web

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"regexp"
	"speedtest/config"
	"speedtest/database"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// RawIspInfo 结构体表示原始 ISP 信息
type RawIspInfo struct {
	IP       string `json:"ip"`
	Hostname string `json:"hostname"`
	City     string `json:"city"`
	Region   string `json:"region"`
	Country  string `json:"country"`
	Loc      string `json:"loc"`
	Org      string `json:"org"`
	Postal   string `json:"postal"`
	Timezone string `json:"timezone"`
	Readme   string `json:"readme"`
}

// 通用响应格式
type CommonIPInfoResponse struct {
	IP        string `json:"ip"`
	Org       string `json:"org"`       // ipinfo = org, self-host = ASN + ISP
	Region    string `json:"region"`    // ipinfo = region, self-host = nil
	City      string `json:"city"`      // ipinfo = city, self-host = nil
	Country   string `json:"country"`   // ipinfo = Country, self-host = CountryCode
	Continent string `json:"continent"` // ipinfo = nil, self-host = continent_name
}

// IspInfo 结构体表示处理后的 ISP 信息
type IspInfo struct {
	ProcessedString string               `json:"processedString"`
	RawIspInfo      CommonIPInfoResponse `json:"rawIspInfo"`
}

// SimpleRateLimiter 定义一个简单的计数器限流器
type SimpleRateLimiter struct {
	mu        sync.Mutex
	maxReqs   int           // 最大请求数
	window    time.Duration // 时间窗口
	count     int           // 当前请求计数
	resetTime time.Time     // 窗口重置时间
}

var rateLimiter = NewSimpleRateLimiter(5, 10*time.Second)

// NewSimpleRateLimiter 创建一个新的限流器
func NewSimpleRateLimiter(maxReqs int, window time.Duration) *SimpleRateLimiter {
	return &SimpleRateLimiter{
		maxReqs:   maxReqs,
		window:    window,
		count:     0,
		resetTime: time.Now().Add(window),
	}
}

// Allow 检查是否允许请求
func (rl *SimpleRateLimiter) Allow() bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()

	// 如果当前时间超过了窗口重置时间，重置计数器
	if now.After(rl.resetTime) {
		rl.count = 0
		rl.resetTime = now.Add(rl.window)
	}

	// 检查是否超过最大请求数
	if rl.count < rl.maxReqs {
		rl.count++
		return true
	}

	return false
}

// GetChartData 处理获取图表数据的请求
func GetChartData(db database.DataAccess, cfg *config.Config, c *gin.Context) {

	if !rateLimiter.Allow() {
		// 如果限流，返回 429 Too Many Requests
		c.JSON(http.StatusTooManyRequests, gin.H{
			"error": "Too Many Requests",
		})
		return
	}

	// 获取最近N条记录
	records, err := db.GetLastNRecords(cfg.Frontend.Chartlist)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 转换数据格式用于图表显示
	var chartData []map[string]interface{}
	for _, record := range records {
		// 转换字符串为浮点数
		download, _ := strconv.ParseFloat(record.Download, 64)
		upload, _ := strconv.ParseFloat(record.Upload, 64)
		ping, _ := strconv.ParseFloat(record.Ping, 64)
		jitter, _ := strconv.ParseFloat(record.Jitter, 64)

		// 解码 ISP 信息
		var ispInfo IspInfo
		err := json.Unmarshal([]byte(record.ISPInfo), &ispInfo)
		if err != nil {
			logError("解码 ISP 信息失败: %s", err)
			ispInfo = IspInfo{ProcessedString: "未知", RawIspInfo: CommonIPInfoResponse{}}
		}
		// 对IP信息进行预处理
		psIP, psRemaining := GetIPFromProcessedString(ispInfo.ProcessedString)
		ispIP := PreprocessIPInfo(ispInfo.RawIspInfo.IP)
		psIP = PreprocessIPInfo(psIP)
		newProcessedString := fmt.Sprintf("%s%s", psIP, psRemaining)
		// 更新ProcessedString字段
		ispInfo.ProcessedString = newProcessedString
		// 更新ispInfo.RawIspInfo.IP字段
		ispInfo.RawIspInfo.IP = ispIP
		// 重新编码ISP信息
		ispInfoJSON, err := json.Marshal(ispInfo)
		if err != nil {
			logError("编码 ISP 信息失败: %s", err)
			ispInfoJSON = []byte("{}")
		}
		record.ISPInfo = string(ispInfoJSON)

		logInfo("%s", record.ISPInfo)
		//logInfo("%s", ispInfo.ProcessedString)
		chartData = append(chartData, map[string]interface{}{
			"timestamp": record.Timestamp,
			"download":  download,
			"upload":    upload,
			"ping":      ping,
			"jitter":    jitter,
			"isp":       record.ISPInfo,
		})
	}

	c.JSON(http.StatusOK, chartData)
}

// GetIPFromProcessedString 分割ProcessedString字段, 取出IP
func GetIPFromProcessedString(processedString string) (string, string) {
	// 查找 ' - ' 的位置
	index := strings.Index(processedString, " - ")
	if index == -1 {
		return "", ""
	}

	ip := processedString[:index]
	if !isIP(ip) {
		logWarning("IP信息不符合规范: %s", ip)
		return "", ""
	}

	// 取出IP和剩余部分
	remaining := processedString[index:] // 包含 ' - ' 和后面的内容
	return ip, remaining
}

// 检测IP是否符合规范
func isIP(ip string) bool {
	// 使用net包进行检测
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return false
	}
	return true
}

// 对IP信息进行预处理, 一定程度上减少隐私问题
func PreprocessIPInfo(ip string) string {
	// 判断是否为特殊IP
	if isSpecialIP(ip) {
		return ip // 直接返回原始 IP
	}
	// 处理 IPv4 地址
	if net.ParseIP(ip).To4() != nil {
		parts := strings.Split(ip, ".")
		if len(parts) == 4 {
			return fmt.Sprintf("%s.%s.%s.x", parts[0], parts[1], parts[2]) // 保留前 24 位，最后一部分用 x 代替
		}
	}

	// 处理 IPv6 地址
	if net.ParseIP(ip).To16() != nil {
		parts := strings.Split(ip, ":")
		if len(parts) >= 3 {
			return fmt.Sprintf("%s:%s:%s::", parts[0], parts[1], parts[2]) // 保留前 48 位
		}
	}

	return ip // 如果不匹配，返回原始 IP
}

var specialIPPatterns = []*regexp.Regexp{
	localIPv6Regex,
	linkLocalIPv6Regex,
	localIPv4Regex,
	privateIPv4Regex10,
	privateIPv4Regex172,
	privateIPv4Regex192,
	linkLocalIPv4Regex,
	cgnatIPv4Regex,
	unspecifiedAddressRegex,
	broadcastAddressRegex,
}

// 特殊IP模式
func isSpecialIP(ip string) bool {
	for _, pattern := range specialIPPatterns {
		if pattern.MatchString(ip) {
			return true
		}
	}
	return false
}
