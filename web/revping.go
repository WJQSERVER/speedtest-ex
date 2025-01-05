package web

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-ping/ping"
)

type PingResult struct {
	IP      string  `json:"ip"`
	Success bool    `json:"success"`
	RTT     float64 `json:"rtt"`
	Error   string  `json:"error,omitempty"`
}

func pingIP(c *gin.Context) {
	ip := c.ClientIP() // 获取客户端 IP 地址
	if ip == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "IP address is required"})
		return
	}

	pinger, err := ping.NewPinger(ip)
	if err != nil {
		c.JSON(http.StatusInternalServerError, PingResult{IP: ip, Success: false, Error: err.Error()})
		return
	}

	pinger.Count = 1 // 只发送一次 ping
	err = pinger.Run()
	if err != nil {
		c.JSON(http.StatusInternalServerError, PingResult{IP: ip, Success: false, Error: err.Error()})
		return
	}

	stats := pinger.Statistics()
	result := PingResult{
		IP:      ip,
		Success: true,
		RTT:     stats.AvgRtt.Seconds() * 1000, // 转换为毫秒
	}

	// 日志输出
	// logInfo("icmp ping result: %v", result) // for debug

	c.JSON(http.StatusOK, result)
}
