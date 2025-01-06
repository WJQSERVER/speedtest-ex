package web

import (
	"net/http"
	"speedtest/config"

	"github.com/gin-gonic/gin"
	ping "github.com/prometheus-community/pro-bing"
)

type PingResult struct {
	IP      string  `json:"ip"`
	Success bool    `json:"success"`
	RTT     float64 `json:"rtt"`
	Error   string  `json:"error,omitempty"`
}

func pingIP(c *gin.Context, cfg *config.Config) {
	ip := c.ClientIP() // 获取客户端 IP 地址
	if ip == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "IP address is required"})
		return
	}

	if cfg.RevPing.Enable {

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
	} else {
		results := PingResult{
			IP:      ip,
			Success: false,
			RTT:     0,
		}
		c.JSON(http.StatusOK, results)
	}
}
