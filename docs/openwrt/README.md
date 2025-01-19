# 在 OpenWrt 上运行 Speedtest-EX

## 下载构建
- 目前仅提供 `arm64` 以及 `amd64` 架构的 `apk`(OpenWrt-SNAPSHOT) 以及 `ipk` 文件
https://github.com/JohnsonRan/packages_builder/releases

## 进行安装
- 前往 OpenWrt 管理界面，转到 `系统 -> 软件包` 页面上传并安装软件包

## 开始使用
- 配置文件位于 `/etc/speedtest-ex/config.toml`
- 修改配置后执行 `/etc/init.d/speedtest-ex restart` 重启服务
- 默认运行在 `8989` 端口
- 若不再需要直接前往软件包处删除 `speedtest-ex` 即可

### 自行编译
项目也提供了可自行编译的库  
https://github.com/JohnsonRan/packages_net_speedtest-ex  
你可以用此库编译出任何架构的软件包以供安装