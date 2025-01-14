package web

import (
	"context"
	"errors"
	"net/http"
	"speedtest/config"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	ping "github.com/prometheus-community/pro-bing"
)

// nonWS (预留)
/*
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
}*/

// new (基于WS实现的RevPing)

type PingResult struct {
	IP      string  `json:"ip"`
	Success bool    `json:"success"`
	RTT     float64 `json:"rtt"`
	Error   string  `json:"error,omitempty"`
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // 允许所有来源
	},
}

// pingIP performs a reverse ping operation on the specified IP address.
// It checks if reverse pinging is enabled in the configuration and attempts to ping the IP.
// If the IP is empty or pinging is disabled, it returns an appropriate error result.
// 
// The function supports a single ping attempt with a 3-second timeout. If the ping is successful,
// it returns the round-trip time in milliseconds. If a timeout occurs or an error is encountered,
// it returns a PingResult with failure details.
// 
// Parameters:
//   - ip: The target IP address to ping
//   - cfg: Configuration object containing reverse ping settings
//
// Returns:
//   - PingResult: Contains ping operation details including IP, success status, RTT, and any error
//   - error: Any critical error encountered during the ping process
//
// Possible error conditions:
//   - Empty IP address
//   - Pinger creation failure
//   - Ping timeout
//   - Network-related errors
//
// Example:
//   result, err := pingIP("8.8.8.8", config)
//   if err != nil {
//     // Handle error
//   }
//   if result.Success {
//     fmt.Printf("Ping successful, RTT: %.2f ms", result.RTT)
//   }
func pingIP(ip string, cfg *config.Config) (PingResult, error) {
	if ip == "" {
		return PingResult{}, errors.New("IP address is required")
	}

	if cfg.RevPing.Enable {
		pinger, err := ping.NewPinger(ip)
		if err != nil {
			return PingResult{IP: ip, Success: false, Error: err.Error()}, err
		}

		timeout := 3 * time.Second
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		pinger.Count = 1
		err = pinger.RunWithContext(ctx)
		if err != nil {
			if errors.Is(err, context.DeadlineExceeded) {
				return PingResult{IP: ip, Success: false, Error: "timeout"}, nil
			}
			return PingResult{IP: ip, Success: false, Error: err.Error()}, err
		}

		stats := pinger.Statistics()
		return PingResult{
			IP:      ip,
			Success: true,
			RTT:     stats.AvgRtt.Seconds() * 1000,
		}, nil
	} else {
		return PingResult{
			IP:      ip,
			Success: false,
			RTT:     0,
			Error:   "revping-not-online",
		}, nil
	}
}

// handleWebSocket establishes a WebSocket connection for continuous IP ping monitoring. It upgrades an HTTP connection to WebSocket, retrieves the client's IP address, and starts a goroutine that periodically sends ping results to the client every 2 seconds. The function handles WebSocket connection lifecycle, including initial ping, periodic updates, and graceful connection closure.
func handleWebSocket(c *gin.Context, cfg *config.Config) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		logError("WebSocket upgrade error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "无法升级到 WebSocket"})
		return
	}
	defer conn.Close()

	ip := c.ClientIP() // 获取客户端 IP 地址

	// 启动一个 goroutine 来定期推送数据
	go func() {
		// 首次推送
		result, err := pingIP(ip, cfg)
		if err != nil {
			logWarning("Ping error: %v", err)
		} else {
			err = conn.WriteJSON(result)
			if err != nil {
				logWarning("WebSocket write error: %v", err)
				return // 如果写入失败，退出 goroutine
			}
		}

		ticker := time.NewTicker(2 * time.Second) // 每 2 秒推送一次数据
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				// 调用 pingIP 函数获取 Ping 结果
				result, err := pingIP(ip, cfg)
				if err != nil {
					logWarning("Ping error: %v", err)
					continue // 继续下一个周期
				}

				err = conn.WriteJSON(result)
				if err != nil {
					logWarning("WebSocket write error: %v", err)
					return // 如果写入失败，退出 goroutine
				}
			}
		}
	}()

	// 处理客户端的关闭连接
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			logWarning("WebSocket read error: %v", err)
			break // 读取消息出错，退出循环
		}
	}
}
