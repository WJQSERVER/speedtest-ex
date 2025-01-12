package config

import (
	"os"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Server    ServerConfig
	Speedtest SpeedtestConfig
	Log       LogConfig
	IPinfo    IPinfoConfig
	Database  DatabaseConfig
	Frontend  FrontendConfig
	RevPing   RevPingConfig
	Auth      AuthConfig
}

/*
[server]
host = ""
port = 8989
basePath = ""
*/
type ServerConfig struct {
	Host     string `toml:"host"`
	Port     int    `toml:"port"`
	BasePath string `toml:"basePath"`
}

/*
[Speedtest]
downDataChunkSize = 4 #mb
downDataChunkCount = 4
*/

type SpeedtestConfig struct {
	DownDataChunkSize  int `toml:"downDataChunkSize"`  // mb
	DownDataChunkCount int `toml:"downDataChunkCount"` // 下载数据块数量
}

/*
[log]
logFilePath = "/data/speedtest-go/log/speedtest-go.log"
maxLogSize = 5 # MB
*/
type LogConfig struct {
	LogFilePath string `toml:"logFilePath"`
	MaxLogSize  int    `toml:"maxLogSize"` // MB
}

/*
[ipinfo]
model = "ip" # ip(self-hosted) or ipinfo(ipinfo.io) todo
ipinfo_url = "https://ip.1888866.xyz" #self-hosted only
ipinfo_api_key = "" # ipinfo.io API key
*/
type IPinfoConfig struct {
	Model     string `toml:"model"`
	IPinfoURL string `toml:"ipinfo_url"`
	IPinfoKey string `toml:"ipinfo_api_key"`
}

/*
[database]
model = "bolt"
path = "/data/speedtest-go/db/speedtest.db"
*/
type DatabaseConfig struct {
	Model string `toml:"model"`
	Path  string `toml:"path"` // bolt file path
}

/*
[frontend]
chartlist = 100 # 默认显示最近100条数据
*/
type FrontendConfig struct {
	Chartlist int `toml:"chartlist"`
}

/*
[revping]
enable = true # 是否开启反向延迟测试
*/
type RevPingConfig struct {
	Enable bool `toml:"enable"`
}

/*
[auth]
enable = false # 是否开启鉴权
username = "admin" # 鉴权用户名
password = "password" # 鉴权密码
secret = "secret" # 加密密钥, 用于生产session cookie, 请务必修改
*/
type AuthConfig struct {
	Enable   bool   `toml:"enable"`
	Username string `toml:"username"`
	Password string `toml:"password"`
	Secret   string `toml:"secret"`
}

// LoadConfig 从 TOML 配置文件加载配置
func LoadConfig(filePath string) (*Config, error) {
	var config Config
	if _, err := toml.DecodeFile(filePath, &config); err != nil {
		return nil, err
	}
	return &config, nil
}

// SaveConfig 保存配置到 TOML 配置文件
func SaveConfig(filePath string, config *Config) error {
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := toml.NewEncoder(file)
	if err := encoder.Encode(config); err != nil {
		return err
	}
	return nil
}
