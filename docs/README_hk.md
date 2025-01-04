![SpeedTest-EX Logo](https://raw.githubusercontent.com/WJQSERVER/speedtest-ex/main/web/pages/favicon_inverted.png)

# SpeedTest-EX

本項目係基於[speedtest-go](https://github.com/librespeed/speedtest-go)項目嘅大幅度重構。[speedtest-go](https://github.com/librespeed/speedtest-go)係使用Go語言重新實現嘅[librespeed](https://github.com/librespeed/speedtest)後端，而本項目係基於[speedtest-go](https://github.com/librespeed/speedtest-go)項目再次重構。LiberSpeed係一個開源嘅網絡測速項目，其使用PHP實現後端，而本項目則使用Go語言同Gin框架實現後端，使程序更加輕量化且易於部署。

**❗ 注意**：基於網頁測速嘅原理，程序會生成無用塊供測速者下載來計算真實下行帶寬，一定程度上存在被惡意刷流量嘅風險，喺對外分享你嘅測速頁面後，請注意觀察伺服器流量使用情況，避免流量使用異常。

## 特性
- 輕量化：無需額外環境，僅需下載二進制文件即可運行。（Docker鏡像亦更加輕量化）
- 易於部署：內嵌前端頁面，無需額外配置即可部署。
- 高效：基於Gin框架，並發處理能力強，響應速度快。

## 與speedtest-go嘅區別
- **Web框架**：speedtest-go使用Chi框架，本項目使用Gin框架。
- **IPinfo**：speedtest-go使用ipinfo.io API獲取IP信息，本項目兼容ipinfo.io API，亦可使用[WJQSERVER-STUDIO/ip](https://github.com/WJQSERVER-STUDIO/ip)為自托管服務提供IP信息。
- **結果圖表**：本項目加入咗結果圖表，方便用戶查看測速結果。
- **更加清晰嘅配置文件**：改進配置文件結構。
- **前端頁面**：內嵌前端頁面，無需額外配置即可部署。（仍與liberspeed前端保持兼容性）
- **重寫**：對大部分組件進行重寫與優化，使程序更加易於維護嘅同時提升部分性能。

## 部署與使用
### Docker部署
參看[docker-cli部署SpeedTest-EX](https://github.com/WJQSERVER/speedtest-ex/blob/main/docs/docker/docker-cli_zh-tw.md)  
參看[docker-compose部署SpeedTest-EX](https://github.com/WJQSERVER/speedtest-ex/blob/main/docs/docker/docker-compose_zh-tw.md)

### 配置文件
參看[配置文件說明](https://github.com/WJQSERVER/speedtest-ex/blob/main/docs/config/config_zh-tw.md)

## 许可证
版權 (C) 2016-2020 Federico Dossena  
版權 (C) 2020 Maddie Zhan  
版權 (C) 2025 WJQSERVER  

本程序係自由軟件：你可以根據自由軟件基金會發布嘅GNU較寬鬆公共許可證條款重新分發同/或修改佢，許可證版本為3，或（根據你嘅選擇）任何更高版本。本程序嘅分發係希望佢能有用，但唔提供任何保證；甚至唔提供適銷性或特定用途適用性嘅隱含保證。關於更多詳細信息，請參見GNU通用公共許可證。你應該已經收到與本程序一齊提供嘅GNU較寬鬆公共許可證嘅副本。如果冇，請訪問<https://www.gnu.org/licenses/lgpl>。
