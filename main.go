/*
LGPL-3.0 License

Copyright (c) 2025 WJQSERVER
*/

package main

import (
	"flag"
	"fmt"
	"log"
	_ "time/tzdata"

	"speedtest/config"
	"speedtest/database"
	"speedtest/web"

	"github.com/WJQSERVER-STUDIO/go-utils/logger"
	_ "github.com/breml/rootcerts"
	"github.com/gin-gonic/gin"
)

var (
	cfg        *config.Config
	configfile = "./config.toml"
	router     *gin.Engine
)

// 日志模块
var (
	logw       = logger.Logw
	logInfo    = logger.LogInfo
	logWarning = logger.LogWarning
	logError   = logger.LogError
)

func ReadFlag() {
	cfgfile := flag.String("cfg", configfile, "config file path")
	flag.Parse()
	configfile = *cfgfile
}

func loadConfig() {
	var err error
	// 初始化配置
	cfg, err = config.LoadConfig(configfile)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	fmt.Printf("Loaded config: %v\n", cfg)
}

func setupLogger() {
	// 初始化日志模块
	var err error
	err = logger.Init(cfg.Log.LogFilePath, cfg.Log.MaxLogSize) // 传递日志文件路径
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	logw("Logger initialized")
	logw("Init Completed")
}

var (
	optConfig = flag.String("c", "", "config file to be used, defaults to settings.toml in the same directory")
)

func init() {
	ReadFlag()
	loadConfig()
	setupLogger()
}

func main() {
	flag.Parse()
	database.SetDBInfo(cfg)
	web.ListenAndServe(cfg)
	defer logger.Close()
}
