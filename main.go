/*
LGPL-3.0 License

Copyright (c) 2025 WJQSERVER
*/

package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	_ "time/tzdata"

	"speedtest/config"
	"speedtest/database"
	"speedtest/web"

	"github.com/WJQSERVER-STUDIO/go-utils/logger"
	_ "github.com/breml/rootcerts"
	"github.com/gin-gonic/gin"
)

var (
	cfg    *config.Config
	router *gin.Engine
)

var (
	version string
)

// 日志模块
var (
	logw       = logger.Logw
	logInfo    = logger.LogInfo
	logWarning = logger.LogWarning
	logError   = logger.LogError
)

var (
	cfgfile  string
	port     int
	initcfg  bool
	auth     bool
	user     string
	password string
	secret   string
	dev      bool
)

func ReadFlag() {
	cfgfilePtr := flag.String("cfg", "./config/config.toml", "config file path") // 配置文件路径
	portPtr := flag.Int("port", 0, "port to listen on")                          // 监听端口
	initcfgPtr := flag.Bool("initcfg", false, "init config mode to run")         // 初始化配置模式
	authPtr := flag.Bool("auth", false, "Enbale auth mode")                      // 授权模式
	userPtr := flag.String("user", "", "User name for auth mode")                // 用户名
	passwordPtr := flag.String("password", "", "Password for auth mode")         // 密码
	secretPtr := flag.String("secret", "", "Secret key for auth mode")           // 密钥
	devPtr := flag.Bool("dev", false, "Development mode")                        // 开发模式

	flag.Parse()
	//configfile = *cfgfile
	cfgfile = *cfgfilePtr
	port = *portPtr
	initcfg = *initcfgPtr
	auth = *authPtr
	user = *userPtr
	password = *passwordPtr
	secret = *secretPtr
	dev = *devPtr
}

func loadConfig() {
	var err error
	// 初始化配置
	//cfg, err = config.LoadConfig(configfile)
	cfg, err = config.LoadConfig(cfgfile)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	fmt.Printf("Loaded config: %v\n", cfg)
}

func saveNewConfig() {
	err := config.SaveConfig(cfgfile, cfg)
	if err != nil {
		log.Printf("Failed to save config: %v", err)
	}
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

func debugOutput() {
	// 输出调试
	fmt.Printf("ConfigFile: %s\n", cfgfile)
	fmt.Printf("Port: %d\n", port)
	fmt.Printf("InitCfg: %t\n", initcfg)
	fmt.Printf("Auth: %t\n", auth)
	fmt.Printf("User: %s\n", user)
	fmt.Printf("Password: %s\n", password)
	fmt.Printf("Secret: %s\n", secret)
}

func init() {
	ReadFlag()
	loadConfig()
	if initcfg {
		initConfig()
		fmt.Printf("Config file initialized, exit.\n")
		os.Exit(0)
	} else {
		initConfig()
	}
	//updateConfig()
	setupLogger()
}

func main() {
	flag.Parse()
	database.SetDBInfo(cfg)
	if dev {
		version = "dev"
		debugOutput()
	}
	web.ListenAndServe(cfg, version)
	defer logger.Close()
}
