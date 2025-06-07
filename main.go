/*
LGPL-3.0 License

Copyright (c) 2025 WJQSERVER
*/

package main

import (
	"crypto/rand"
	"encoding/base64"
	"flag"
	"fmt"
	"log"
	"os"
	_ "time/tzdata"

	"speedtest/config"
	"speedtest/database"
	"speedtest/web"

	_ "github.com/breml/rootcerts"
)

var (
	cfg *config.Config
)

var (
	version string
)

var (
	cfgfile     string
	configfile  string
	port        int
	initcfg     bool
	auth        bool
	user        string
	password    string
	secret      string
	dev         bool
	showVersion bool
)

func ReadFlag() {
	cfgfilePtr := flag.String("cfg", "", "config file path(Deprecated)")          // 配置文件路径(弃用)
	configfilePtr := flag.String("c", "./config/config.toml", "config file path") // 配置文件路径
	portPtr := flag.Int("port", 0, "port to listen on")                           // 监听端口
	initcfgPtr := flag.Bool("initcfg", false, "init config mode to run")          // 初始化配置模式
	authPtr := flag.Bool("auth", false, "Enbale auth mode")                       // 授权模式
	userPtr := flag.String("user", "", "User name for auth mode")                 // 用户名
	passwordPtr := flag.String("password", "", "Password for auth mode")          // 密码
	secretPtr := flag.String("secret", "", "Secret key for auth mode")            // 密钥
	devPtr := flag.Bool("dev", false, "Development mode")                         // 开发模式
	versionPtr := flag.Bool("version", false, "Show version")                     // 显示版本

	flag.Parse()
	configfile = *configfilePtr
	cfgfile = *cfgfilePtr
	port = *portPtr
	initcfg = *initcfgPtr
	auth = *authPtr
	user = *userPtr
	password = *passwordPtr
	secret = *secretPtr
	dev = *devPtr
	showVersion = *versionPtr

	if cfgfile != "" && configfile == "./config/config.toml" {
		configfile = cfgfile
		fmt.Printf("-cfg is Deprecated, using -c \n")
	}
}

func loadConfig() {
	var err error
	// 初始化配置
	//cfg, err = config.LoadConfig(configfile)
	cfg, err = config.LoadConfig(configfile)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	fmt.Printf("Loaded config: %v\n", cfg)
}

func saveNewConfig() {
	err := config.SaveConfig(configfile, cfg)
	if err != nil {
		log.Printf("Failed to save config: %v", err)
	}
}

func updateConfig() {
	// 写入新配置
	saveNewConfig()
	// 重新加载配置
	loadConfig()
}

func generateSecret(length int) (string, error) {
	// 生成随机字节
	bytes := make([]byte, length)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}

	// 将随机字节编码为 Base64
	secret := base64.RawStdEncoding.EncodeToString(bytes)
	return secret, nil
}

func initConfig() {
	// 初始化配置

	// 端口
	if port != 0 {
		cfg.Server.Port = port
	}

	// 开启鉴权模式
	if auth {
		if user != "" && password != "" {
			cfg.Auth.Enable = true
			cfg.Auth.Username = user
			cfg.Auth.Password = password
		} else {
			fmt.Println("User name and password must be set for auth mode")
			return
		}
		if secret != "" {
			if secret == "rand" {
				var err error
				secret, err = generateSecret(32)
				if err != nil {
					fmt.Println("Failed to generate secret key:", err)
					return
				}
				fmt.Println("Generated secret key")
				cfg.Auth.Secret = secret
			} else if len(secret) < 8 {
				fmt.Println("Secret key must be at least 8 characters long")
				return
			}
			fmt.Println("Secret key:", secret)
			cfg.Auth.Secret = secret
		} else {
			fmt.Println("Secret key must be set for auth mode")
			return
		}
	}

	// 保存并重载
	updateConfig()

}

func debugOutput() {
	// 输出调试
	fmt.Printf("ConfigFile: %s\n", configfile)
	fmt.Printf("Port: %d\n", port)
	fmt.Printf("InitCfg: %t\n", initcfg)
	fmt.Printf("Auth: %t\n", auth)
	fmt.Printf("User: %s\n", user)
	fmt.Printf("Password: %s\n", password)
	fmt.Printf("Secret: %s\n", secret)
}

func init() {
	ReadFlag()
	if showVersion {
		fmt.Printf("SpeedTest-EX Version: %s\n", version)
		os.Exit(0)
	}
	loadConfig()
	if initcfg {
		initConfig()
		fmt.Printf("Config file initialized, exit.\n")
		os.Exit(0)
	} else {
		initConfig()
	}
	//updateConfig()
	web.RandomDataInit(cfg)
	web.InitEmptyBuf()
}

func main() {
	flag.Parse()
	database.SetDBInfo(cfg)
	if dev {
		version = "dev"
		debugOutput()
	}
	web.ListenAndServe(cfg, version)
}
