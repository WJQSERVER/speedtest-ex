# 在 OpenWrt 上運行 Speedtest-EX

> 非項目官方適配，若有問題請轉至 https://github.com/JohnsonRan/packages_builder

## 下載構建
- 目前僅提供 `arm64` 以及 `amd64` 架構的 `apk` (OpenWrt-SNAPSHOT) 以及 `ipk` 文件
[下載連結](https://github.com/JohnsonRan/packages_builder/releases)

## 進行安裝
- 前往 OpenWrt 管理介面，轉到 `系統 -> 軟體包` 頁面上傳並安裝軟體包

## 開始使用
- 配置文件位於 `/etc/speedtest-ex/config.toml`
- 修改配置後執行 `/etc/init.d/speedtest-ex restart` 重啟服務
- 默認運行在 `8989` 端口
- 若不再需要直接前往軟體包處刪除 `speedtest-ex` 即可

### 自行編譯
項目也提供了可自行編譯的庫  
[編譯庫連結](https://github.com/JohnsonRan/packages_net_speedtest-ex)  
你可以用此庫編譯出任何架構的軟體包以供安裝

