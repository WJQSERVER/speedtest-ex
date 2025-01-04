package config

import (
	"github.com/BurntSushi/toml"
)

type Config struct {
	Server   ServerConfig
	Log      LogConfig
	IPinfo   IPinfoConfig
	Database DatabaseConfig
	Frontend FrontendConfig
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

// LoadConfig 从 TOML 配置文件加载配置
func LoadConfig(filePath string) (*Config, error) {
	var config Config
	if _, err := toml.DecodeFile(filePath, &config); err != nil {
		return nil, err
	}
	return &config, nil
}
