# 傳入參數

``` bash
speedtest-ex -cfg /path/to/config/config.toml # 傳入配置文件路徑(必選)

# 以下參數可選
-port 8080 # 設置服務端口，默認8989

-auth #開啟鑒權，默認關閉

-username admin # 設置用戶名(需要開啟鑒權)

-password admin # 設置密碼(需要開啟鑒權)

-secret rand # 設置密鑰(需要開啟鑒權) (rand為隨機生成)

-initcfg # 初始化配置模式, 傳入並保存配置, 用於快速初始化配置(保存配置後將會退出)

-dev # 開啟開發模式，默認關閉(非開發用戶請不要開啟)

-version # 顯示版本資訊

```