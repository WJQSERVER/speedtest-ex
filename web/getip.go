package web

import (
	"regexp"
	"speedtest/config"
	"speedtest/ipinfo"
	"speedtest/results"

	"github.com/gin-gonic/gin"
)

// 预编译的正则表达式变量
var (
	localIPv6Regex          = regexp.MustCompile(`^::1$`)                            // 匹配本地 IPv6 地址
	linkLocalIPv6Regex      = regexp.MustCompile(`^fe80:`)                           // 匹配链路本地 IPv6 地址
	localIPv4Regex          = regexp.MustCompile(`^127\.`)                           // 匹配本地 IPv4 地址
	privateIPv4Regex10      = regexp.MustCompile(`^10\.`)                            // 匹配私有 IPv4 地址（10.0.0.0/8）
	privateIPv4Regex172     = regexp.MustCompile(`^172\.(1[6-9]|2\d|3[01])\.`)       // 匹配私有 IPv4 地址（172.16.0.0/12）
	privateIPv4Regex192     = regexp.MustCompile(`^192\.168\.`)                      // 匹配私有 IPv4 地址（192.168.0.0/16）
	linkLocalIPv4Regex      = regexp.MustCompile(`^169\.254\.`)                      // 匹配链路本地 IPv4 地址（169.254.0.0/16）
	cgnatIPv4Regex          = regexp.MustCompile(`^100\.([6-9][0-9]|1[0-2][0-7])\.`) // 匹配 CGNAT IPv4 地址（100.64.0.0/10）
	unspecifiedAddressRegex = regexp.MustCompile(`^0\.0\.0\.0$`)                     // 匹配未指定地址（0.0.0.0）
	broadcastAddressRegex   = regexp.MustCompile(`^255\.255\.255\.255$`)             // 匹配广播地址（255.255.255.255）
	removeASRegexp          = regexp.MustCompile(`AS\d+\s`)                          // 用于去除 ISP 信息中的自治系统编号
)

func getIP(c *gin.Context, cfg *config.Config) {
	clientIP := c.ClientIP() // 获取客户端 IP 地址
	// clientIP := "1.1.1.1"  // for debug
	var ret results.Result // 创建结果结构体实例
	ret = ipinfo.GetIP(clientIP, cfg)
	c.JSON(200, ret) // 返回 JSON 响应
}
