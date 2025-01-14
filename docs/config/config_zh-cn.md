# 配置文件

以下以Docker容器内的配置文件为例。

配置文件位于 `{安装目录}/speedtest-ex/config/config.toml`

```toml
[server]
host = "0.0.0.0" # 监听地址
port = 8989 # 监听端口
basePath = "" # 兼容LiberSpeed而保留, 无需求请不要修改

[Speedtest]
downDataChunkSize = 4 #mb 下载数据分块大小
downDataChunkCount = 4 # 下载数据分块数量

[log]
logFilePath = "/data/speedtest-ex/log/speedtest-ex.log" # 日志文件路径
maxLogSize = 5 # MB 日志文件最大容量

[ipinfo]
model = "ipinfo" # ip(自托管) 或 ipinfo(ipinfo.io)
ipinfo_url = "" #使用自托管时请填写您的ip服务地址
ipinfo_api_key = "" # ipinfo.io API key 若有可以填写

[database]
model = "bolt"  # 数据库类型, 目前仅支持BoltDB
path = "/data/speedtest-ex/db/speedtest.db" # 数据库文件路径

[frontend]
chartlist = 100 # 默认显示最近100条数据

[revping]
enable = true # 是否开启反向延迟测试

[auth]
enable = false # 是否开启鉴权
username = "admin" # 鉴权用户名
password = "password" # 鉴权密码
secret = "secret" # 加密密钥, 用于生产session cookie, 请务必修改

```