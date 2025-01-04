![SpeedTest-EX Logo](https://raw.githubusercontent.com/WJQSERVER/speedtest-ex/main/web/pages/favicon_inverted.png)

# SpeedTest-EX

本專案是基於[speedtest-go](https://github.com/librespeed/speedtest-go)專案的大幅度重構。[speedtest-go](https://github.com/librespeed/speedtest-go)是使用Go語言重新實現的[librespeed](https://github.com/librespeed/speedtest)後端，而本專案是基於[speedtest-go](https://github.com/librespeed/speedtest-go)專案再次重構。LiberSpeed是一個開源的網路測速專案，其使用PHP實現後端，而本專案則使用Go語言和Gin框架實現後端，使程式更加輕量化且易於部署。

**❗ 注意**：基於網頁測速的原理，程式會生成無用塊供測速者下載來計算真實下行帶寬，一定程度上存在被惡意刷流量的風險，在對外分享你的測速頁面後，請注意觀察伺服器流量使用情況，避免流量使用異常。

## 特性
- 輕量化：無需額外環境，僅需下載二進制檔案即可運行。（Docker映像也更加輕量化）
- 易於部署：內嵌前端頁面，無需額外配置即可部署。
- 高效：基於Gin框架，並發處理能力強，響應速度快。

## 與speedtest-go的區別
- **Web框架**：speedtest-go使用Chi框架，本專案使用Gin框架。
- **IPinfo**：speedtest-go使用ipinfo.io API獲取IP資訊，本專案兼容ipinfo.io API，也可使用[WJQSERVER-STUDIO/ip](https://github.com/WJQSERVER-STUDIO/ip)為自托管服務提供IP資訊。
- **結果圖表**：本專案加入了結果圖表，方便用戶查看測速結果。
- **更加清晰的配置檔案**：改進配置檔案結構。
- **前端頁面**：內嵌前端頁面，無需額外配置即可部署。（仍與liberspeed前端保持兼容性）
- **重寫**：對大部分元件進行重寫與優化，使程式更加易於維護的同時提升部分性能。

## 部署與使用
### Docker部署
參看[docker-cli部署SpeedTest-EX](https://github.com/WJQSERVER/speedtest-ex/blob/main/docs/docker/docker-cli_zh-tw.md)  
參看[docker-compose部署SpeedTest-EX](https://github.com/WJQSERVER/speedtest-ex/blob/main/docs/docker/docker-compose_zh-tw.md)

### 配置檔案
參看[配置檔案說明](https://github.com/WJQSERVER/speedtest-ex/blob/main/docs/config/config_zh-tw.md)

## 截圖

![SpeedTest-EX Index Page](https://webp.wjqserver.com/speedtest-ex/index.png)

![SpeedTest-EX Chart Page](https://webp.wjqserver.com/speedtest-ex/chart.png)

## 许可证
版權 (C) 2016-2020 Federico Dossena

版權 (C) 2020 Maddie Zhan

版權 (C) 2025 WJQSERVER  

本程式是自由軟體：您可以根據自由軟體基金會發布的GNU較寬鬆公共授權條款重新分發和/或修改它，授權版本為3，或（根據您的選擇）任何更高版本。本程式的分發是希望它能有用，但不提供任何保證；甚至不提供適銷性或特定用途適用性的隱含保證。關於更多詳細資訊，請參見GNU通用公共授權。您應該已經收到與本程式一起提供的GNU較寬鬆公共授權的副本。如果沒有，請訪問<https://www.gnu.org/licenses/lgpl>。
