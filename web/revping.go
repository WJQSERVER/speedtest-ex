package web

import (
	"context"
	"errors"
	"net/http"
	"speedtest/config"
	"time"

	"github.com/gorilla/websocket"
	"github.com/infinite-iroha/touka"
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

func handleWebSocket(c *touka.Context, cfg *config.Config) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.Errorf("WebSocket upgrade error: %v", err)
		c.JSON(http.StatusInternalServerError, touka.H{"error": "无法升级到 WebSocket"})
		return
	}
	defer conn.Close()

	// 在此处设置一个上下文，用于控制后续所有goroutine的生命周期
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel() // 确保在 handleWebSocket 函数退出时取消上下文

	ip := c.ClientIP() // 获取客户端 IP 地址

	// 启动一个 goroutine 来定期推送数据
	go func() {
		// 首次推送
		result, err := pingIP(ip, cfg)
		if err != nil {
			c.Warnf("Ping error: %v", err)
		} else {
			err = conn.WriteJSON(result)
			if err != nil {
				c.Warnf("WebSocket write error: %v", err)
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
					c.Warnf("Ping error: %v", err)
					continue // 继续下一个周期
				}

				err = conn.WriteJSON(result)
				if err != nil {
					c.Warnf("WebSocket write error: %v", err)
					return // 如果写入失败，退出 goroutine
				}
			case <-ctx.Done():
				// 如果上下文被取消，退出 goroutine
				c.Infof("WebSocket context cancelled, closing connection.")
				return
			}
		}
	}()

	go func() {
		// 处理客户端的关闭连接
		for {
			select {
			default:
				_, _, err := conn.ReadMessage()
				if err != nil {
					c.Warnf("WebSocket read error: %v", err)
					cancel()
					return
				}
			case <-ctx.Done():
				// 如果上下文被取消，退出 goroutine
				return
			}

		}
	}()

	<-ctx.Done()
}
