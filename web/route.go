package web

import (
	"fmt"
	"io"
	"regexp"
	"speedtest/config"
	"speedtest/database"
	"speedtest/results"

	"github.com/WJQSERVER-STUDIO/go-utils/logger"
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
)

var (
	logw       = logger.Logw
	logInfo    = logger.LogInfo
	logWarning = logger.LogWarning
	logError   = logger.LogError
)

var pagesPathRegex = regexp.MustCompile(`^[\w/]+$`)

// ListenAndServe 启动HTTP服务器并配置路由处理程序。它设置Gin路由器、会话管理、CORS、身份验证和各种API端点。
// 
// 参数:
//   - cfg: 服务器配置，包含认证、服务器和其他设置
//   - version: 当前服务器版本字符串
//
// 返回值:
//   - error: 如果服务器启动失败，返回错误信息
//
// 主要功能:
//   - 配置Gin服务器为发布模式
//   - 设置会话管理（如果启用认证）
//   - 配置跨域资源共享（CORS）
//   - 添加登录和登出路由（如果启用认证）
//   - 提供版本信息接口
//   - 配置后端和基础路径路由
//   - 添加遥测、IP获取、垃圾数据和图表数据等接口
//   - 设置WebSocket端点
//   - 为PHP前端提供兼容性路由
//   - 处理未匹配的请求
func ListenAndServe(cfg *config.Config, version string) error {
	// gin.SetMode(gin.DebugMode)
	gin.SetMode(gin.ReleaseMode)
	gin.LoggerWithWriter(io.Discard)
	router := gin.New()
	router.UseH2C = true
	router.Use(gin.Recovery())

	if cfg.Auth.Enable == true {
		// 设置 session 中间件
		store := cookie.NewStore([]byte(cfg.Auth.Secret))
		store.Options(sessions.Options{
			Path:     "/",
			MaxAge:   86400 * 7, // 7 days
			HttpOnly: true,
		})
		router.Use(sessions.Sessions("mysession", store))
		// 应用 session 中间件
		router.Use(SessionMiddleware())

	}

	// CORS
	router.Use(cors.New(cors.Config{
		AllowAllOrigins: true,
		AllowMethods:    []string{"GET", "POST", "OPTIONS", "HEAD"},
		AllowHeaders:    []string{"*"},
	}))

	if cfg.Auth.Enable {
		// 添加登录路由
		router.POST("/api/login", func(c *gin.Context) {
			AuthLogin(c, cfg)
		})

		// 添加登出路由
		router.GET("/api/logout", func(c *gin.Context) {
			AuthLogout(c)
		})
	}

	// 版本信息接口
	router.GET("/api/version", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"Version": version,
		})
	})

	backendUrl := "/backend"
	// 记录遥测数据
	router.POST(backendUrl+"/results/telemetry", func(c *gin.Context) {
		results.Record(c, cfg)
	})
	// 获取客户端 IP 地址
	router.GET(backendUrl+"/getIP", func(c *gin.Context) {
		getIP(c, cfg)
	})
	// 垃圾数据接口
	router.GET(backendUrl+"/garbage", garbage)
	// 空接口
	router.Any(backendUrl+"/empty", empty)
	// 获取图表数据
	router.GET(backendUrl+"/api/chart-data", func(c *gin.Context) {
		GetChartData(database.DB, cfg, c)
	})
	// 反向ping
	/*
		router.GET(backendUrl+"/revping", func(c *gin.Context) {
			pingIP(c, cfg)
		})
	*/

	basePath := cfg.Server.BasePath
	// 记录遥测数据
	router.POST(basePath+"/results/telemetry", func(c *gin.Context) {
		results.Record(c, cfg)
	})
	// 获取客户端 IP 地址
	router.GET(basePath+"/getIP", func(c *gin.Context) {
		getIP(c, cfg)
	})
	// 垃圾数据接口
	router.GET(basePath+"/garbage", garbage)
	// 空接口
	router.Any(basePath+"/empty", empty)
	// 获取图表数据
	router.GET(basePath+"/api/chart-data", func(c *gin.Context) {
		GetChartData(database.DB, cfg, c)
	})
	// 反向ping
	/*
		router.GET(basePath+"/revping", func(c *gin.Context) {
			pingIP(c, cfg)
		})
	*/
	// 反向ping ws
	router.GET(basePath+"/ws", func(c *gin.Context) {
		handleWebSocket(c, cfg)
	})

	// PHP 前端默认值兼容性
	router.Any(basePath+"/empty.php", empty)
	router.GET(basePath+"/garbage.php", garbage)
	router.GET(basePath+"/getIP.php", func(c *gin.Context) {
		getIP(c, cfg)
	})
	router.POST(basePath+"/results/telemetry.php", func(c *gin.Context) {
		results.Record(c, cfg)
	})

	//router.NoRoute(gin.WrapH(http.FileServer(http.FS(pages))))
	// 处理所有请求
	router.NoRoute(func(c *gin.Context) {
		PagesEmbedFS(c)
	})

	return StartServer(cfg, router)
}

func StartServer(cfg *config.Config, r *gin.Engine) error {
	addr := cfg.Server.Host

	if addr == "" {
		addr = "0.0.0.0"
	}

	port := cfg.Server.Port
	if port == 0 {
		port = 8989
	}

	if err := r.Run(fmt.Sprintf("%s:%d", addr, port)); err != nil {
		return fmt.Errorf("failed to run server: %w", err)
	}

	return nil
}
