package web

import (
	"fmt"
	"net/http"
	"regexp"
	"speedtest/config"
	"speedtest/database"
	"speedtest/results"
	"time"

	"github.com/fenthope/compress"
	"github.com/fenthope/cors"
	"github.com/fenthope/reco"
	"github.com/fenthope/record"
	"github.com/fenthope/sessions"
	"github.com/fenthope/sessions/cookie"

	"github.com/infinite-iroha/touka"
)

var pagesPathRegex = regexp.MustCompile(`^[\w/]+$`)

// ListenAndServe 启动HTTP服务器并设置路由处理程序
func ListenAndServe(cfg *config.Config, version string) error {
	router := touka.New()
	router.Use(touka.Recovery())
	router.Use(record.Middleware())
	router.Use(compress.Compression(compress.DefaultCompressionConfig()))
	var (
		logPath string
		logSize int
	)
	if cfg.Log.LogFilePath != "" {
		logPath = cfg.Log.LogFilePath
	} else {
		logPath = "speedtest-ex.log"
	}
	if cfg.Log.MaxLogSize != 0 {
		logSize = cfg.Log.MaxLogSize
	} else {
		logSize = 5
	}

	router.SetLogger(reco.Config{
		Level:           reco.LevelInfo,
		Mode:            reco.ModeText,
		TimeFormat:      time.RFC3339,
		FilePath:        logPath,
		EnableRotation:  true,
		MaxFileSizeMB:   int64(logSize),
		MaxBackups:      5,
		CompressBackups: true,
		Async:           true,
		DefaultFields:   nil,
	})

	if cfg.Auth.Enable {
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
		router.POST("/api/login", func(c *touka.Context) {
			AuthLogin(c, cfg)
		})

		// 添加登出路由
		router.GET("/api/logout", func(c *touka.Context) {
			AuthLogout(c)
		})
	}

	// 版本信息接口
	router.GET("/api/version", func(c *touka.Context) {
		c.JSON(200, touka.H{
			"Version": version,
		})
	})

	backendUrl := "/backend"
	// 记录遥测数据
	router.POST(backendUrl+"/results/telemetry", func(c *touka.Context) {
		results.Record(c, cfg)
	})
	// 获取客户端 IP 地址
	router.GET(backendUrl+"/getIP", func(c *touka.Context) {
		getIP(c, cfg)
	})
	// 垃圾数据接口
	if cfg.Speedtest.DownloadGenStream {
		router.GET(backendUrl+"/garbage", garbageStream)
	} else {
		router.GET(backendUrl+"/garbage", garbage)
	}
	// 空接口
	router.ANY(backendUrl+"/empty", empty)
	// 获取图表数据
	router.GET(backendUrl+"/api/chart-data", func(c *touka.Context) {
		GetChartData(database.DB, cfg, c)
	})

	basePath := cfg.Server.BasePath
	// 记录遥测数据
	router.POST(basePath+"/results/telemetry", func(c *touka.Context) {
		results.Record(c, cfg)
	})
	// 获取客户端 IP 地址
	router.GET(basePath+"/getIP", func(c *touka.Context) {
		getIP(c, cfg)
	})
	// 垃圾数据接口
	router.GET(basePath+"/garbage", garbage)
	// 空接口
	router.ANY(basePath+"/empty", empty)
	// 获取图表数据
	router.GET(basePath+"/api/chart-data", func(c *touka.Context) {
		GetChartData(database.DB, cfg, c)
	})
	// 反向ping ws
	router.GET(basePath+"/ws", func(c *touka.Context) {
		handleWebSocket(c, cfg)
	})

	// PHP 前端默认值兼容性
	router.ANY(basePath+"/empty.php", empty)
	router.GET(basePath+"/garbage.php", garbageStream)
	router.GET(basePath+"/getIP.php", func(c *touka.Context) {
		getIP(c, cfg)
	})
	router.POST(basePath+"/results/telemetry.php", func(c *touka.Context) {
		results.Record(c, cfg)
	})
	router.SetUnMatchFS(http.FS(pages))
	return StartServer(cfg, router)
}

func StartServer(cfg *config.Config, r *touka.Engine) error {
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
