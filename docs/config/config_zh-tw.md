# 配置文件

以下以Docker容器内的配置文件为例。

配置文件位於 `{安裝目錄}/speedtest-ex/config/config.toml`

```toml
[server]
host = "0.0.0.0" # 監聽地址
port = 8989 # 監聽端口
basePath = "" # 兼容LiberSpeed而保留, 無需求請不要修改

[log]
logFilePath = "/data/speedtest-ex/log/speedtest-ex.log" # 日誌文件路徑
maxLogSize = 5 # MB 日誌文件最大容量

[ipinfo]
model = "ipinfo" # ip(自托管) 或 ipinfo(ipinfo.io)
ipinfo_url = "" #自托管時請填寫您的IP程式的API地址
ipinfo_api_key = "" # ipinfo.io API key 若有可以填寫

[database]
model = "bolt"  # 資料庫類型, 目前僅支持BoltDB
path = "/data/speedtest-ex/db/speedtest.db" # 資料庫文件路徑

[frontend]
chartlist = 100 # 預設顯示最近100條數據
```