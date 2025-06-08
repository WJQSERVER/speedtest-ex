package web

import (
	"context"
	"errors"
	"net/http"
	"speedtest/config"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/infinite-iroha/touka"
	ping "github.com/prometheus-community/pro-bing"
)

// 基于WS实现的RevPing
type PingResult struct {
	IP      string  `json:"ip"`
	Success bool    `json:"success"`
	RTT     float64 `json:"rtt"`
	Error   string  `json:"error,omitempty"`
}

var (
	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true // 允许所有来源
		},
	}
	useUnprivilegedPing      = false    // 全局标志, 记录是否应强制使用非特权ping
	checkUnprivilegedPingMux sync.Mutex // 互斥锁, 确保首次检测和标志设置是线程安全的
)

func pingIP(ip string, cfg *config.Config, c *touka.Context) (PingResult, error) {
	if ip == "" {
		return PingResult{}, errors.New("IP address is required")
	}

	if !cfg.RevPing.Enable {
		return PingResult{IP: ip, Success: false, Error: "revping-not-online"}, nil
	}

	// 内部函数, 封装单次ping的逻辑, 便于重试
	runPing := func(privileged bool) (*ping.Statistics, error) {
		pinger, err := ping.NewPinger(ip)
		if err != nil {
			return nil, err
		}
		pinger.SetPrivileged(privileged)
		pinger.Count = 1

		timeout := 3 * time.Second
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		err = pinger.RunWithContext(ctx)
		if err != nil {
			return nil, err
		}
		stats := pinger.Statistics()
		if stats.PacketsRecv == 0 {
			return nil, context.DeadlineExceeded // 如果没有收到包, 视作超时
		}
		return stats, nil
	}

	// --- 自动检测与切换的核心逻辑 ---

	// 首先, 根据全局标志决定是否一开始就使用非特权模式
	privilegedAttempt := !useUnprivilegedPing

	stats, err := runPing(privilegedAttempt)

	// 如果第一次尝试(特权模式)失败了
	if err != nil {
		// 检查是否是权限错误, 并且我们确实是在特权模式下尝试的
		isPermissionError := strings.Contains(err.Error(), "permission denied") || strings.Contains(err.Error(), "operation not permitted")

		checkUnprivilegedPingMux.Lock()
		// 再次检查, 防止并发写入
		if !useUnprivilegedPing && privilegedAttempt && isPermissionError {
			c.Warnf("Permission denied for ICMP ping, switching to unprivileged (UDP) mode for all subsequent pings.")
			useUnprivilegedPing = true // 设置全局标志
			checkUnprivilegedPingMux.Unlock()

			// 立即用非特权模式重试一次
			stats, err = runPing(false)
		} else {
			checkUnprivilegedPingMux.Unlock()
		}
	}

	// --- 处理最终结果 ---

	if err != nil {
		// 如果(重试后)仍然有错误
		if errors.Is(err, context.DeadlineExceeded) {
			return PingResult{IP: ip, Success: false, Error: "timeout"}, nil
		}
		// 对于其他错误(包括权限错误, 如果重试也失败了), 返回具体错误信息
		return PingResult{IP: ip, Success: false, Error: err.Error()}, err
	}

	// 如果成功
	return PingResult{
		IP:      ip,
		Success: true,
		RTT:     stats.AvgRtt.Seconds() * 1000,
	}, nil
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
		result, err := pingIP(ip, cfg, c)
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
				result, err := pingIP(ip, cfg, c)
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
