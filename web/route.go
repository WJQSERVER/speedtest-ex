package web

import (
	"embed"
	"fmt"
	"io/fs"
	"net/http"
	"speedtest/config"
	"speedtest/database"
	"speedtest/results"

	"github.com/WJQSERVER-STUDIO/go-utils/logger"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

var (
	logw       = logger.Logw
	logInfo    = logger.LogInfo
	logWarning = logger.LogWarning
	logError   = logger.LogError
)

var (
	//go:embed pages/*
	assetsFS embed.FS
)

// ListenAndServe 启动HTTP服务器并设置路由处理程序
func ListenAndServe(cfg *config.Config) error {
	// gin.SetMode(gin.DebugMode)
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	router.UseH2C = true

	// CORS
	router.Use(cors.New(cors.Config{
		AllowAllOrigins: true,
		AllowMethods:    []string{"GET", "POST", "OPTIONS", "HEAD"},
		AllowHeaders:    []string{"*"},
	}))

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
	router.GET(backendUrl+"/revping", func(c *gin.Context) {
		pingIP(c, cfg)
	})

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
	router.GET(basePath+"/revping", func(c *gin.Context) {
		pingIP(c, cfg)
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

	// assets 嵌入文件系统
	pages, err := fs.Sub(assetsFS, "pages")
	if err != nil {
		logError("Failed when processing pages: %s", err)
	}
	router.NoRoute(gin.WrapH(http.FileServer(http.FS(pages))))

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
