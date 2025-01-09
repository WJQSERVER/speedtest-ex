package main

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
)

/*
./demo -cfg ./demo.toml -port 8989 -user admin -password password -secret rand -auth -initcfg
./demo -cfg ./demo.toml -port 8989 -user admin -password password -secret secret -auth -initcfg
./demo -cfg ./demo.toml -port 8989 -user admin -password password -secret secret -auth
./demo -cfg ./demo.toml -port 8989
./demo -cfg ./demo.toml -port 8989 -initcfg
*/

// flags
/*
	cfgfilePtr := flag.String("cfg", "./config/config.toml", "config file path") // 配置文件路径
	portPtr := flag.Int("port", 0, "port to listen on")                          // 监听端口
	initcfgPtr := flag.Bool("initcfg", false, "init config mode to run")         // 初始化配置模式
	authPtr := flag.Bool("auth", false, "Enbale auth mode")                      // 授权模式
	userPtr := flag.String("user", "", "User name for auth mode")                // 用户名
	passwordPtr := flag.String("password", "", "Password for auth mode")         // 密码
	secretPtr := flag.String("secret", "", "Secret key for auth mode")           // 密钥
*/

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
