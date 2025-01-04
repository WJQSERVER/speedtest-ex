# 配置文件

以下以Docker容器内的配置文件为例。

配置文件位于 `{安装目录}/speedtest-ex/config/config.toml`

```toml
[server]
host = "0.0.0.0" # 监听地址
port = 8989 # 监听端口
basePath = "" # 兼容LiberSpeed而保留, 无需求请不要修改

[log]
logFilePath = "/data/speedtest-go/log/speedtest-go.log" # 日志文件路径
maxLogSize = 5 # MB 日志文件最大容量

[ipinfo]
model = "ipinfo" # ip(自托管) 或 ipinfo(ipinfo.io)
ipinfo_url = "" #自托管时请填写ipinfo.io的API地址
ipinfo_api_key = "" # ipinfo.io API key 若有可以填写

[database]
model = "bolt"  # 数据库类型, 目前仅支持BoltDB
path = "/data/speedtest-go/db/speedtest.db" # 数据库文件路径

[frontend]
chartlist = 100 # 默认显示最近100条数据


```