package web

import (
	"context"
	"errors"
	"net/http"
	"speedtest/config"
	"time"

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
		// timeout 设置为 3 秒
		timeout := 3 * time.Second
		//timeout := 10 * time.Nanosecond // for debug

		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		pinger.Count = 1 // 只发送一次 ping
		err = pinger.RunWithContext(ctx)
		if err != nil {
			if errors.Is(err, context.DeadlineExceeded) {
				c.JSON(http.StatusRequestTimeout, PingResult{IP: ip, Success: false, Error: "timeout"})
				return
			}
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
			Error:   "revping-not-online",
		}
		c.JSON(http.StatusOK, results)
	}
}
