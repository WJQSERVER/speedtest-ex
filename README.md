![SpeedTest-EX Logo](https://raw.githubusercontent.com/WJQSERVER/speedtest-ex/main/web/pages/favicon_inverted.png)

# SpeedTest-EX

本项目是基于[speedtest-go](https://github.com/librespeed/speedtest-go)项目的大幅度重构.
[speedtest-go](https://github.com/librespeed/speedtest-go)是使用Go语言重新实现的[librespeed](https://github.com/librespeed/speedtest)后端, 而本项目是基于[speedtest-go](https://github.com/librespeed/speedtest-go)项目再次重构.
LiberSpeed是一个开源的网络测速项目, 其使用php实现后端, 而本项目则使用Go语言和Gin框架实现后端, 使程序更加轻量化且易于部署.

zh-cn | [en](https://github.com/WJQSERVER/speedtest-ex/blob/main/docs/README_en.md) | [zh-tw](https://github.com/WJQSERVER/speedtest-ex/blob/main/docs/README_zh-tw.md) | [cantonese|hk](https://github.com/WJQSERVER/speedtest-ex/blob/main/docs/README_hk.md)

[SpeedTest-EX 讨论群组](https://t.me/speedtestex)

**❗ 注意**：基于网页测速的原理，程序会生成无用块供测速者下载来计算真实下行带宽，一定程度上存在被恶意刷流量的风险，在对外分享你的测速页面后，请注意观察服务器流量使用情况，避免流量使用异常。

## 特性

- 轻量化: 无需额外环境, 仅需下载二进制文件即可运行.(Docker镜像也更加轻量化)
- 易于部署: 内嵌前端页面, 无需额外配置即可部署.
- 高效: 基于Gin框架, 并发处理能力强, 响应速度快.

## 与speedtest-go的区别

- Web框架: speedtest-go使用Chi框架, 本项目使用Gin框架.
- IPinfo: speedtest-go使用ipinfo.io API获取IP信息, 本项目兼容ipinfo.io API, 也可使用[WJQSERVER-STUDIO/ip](https://github.com/WJQSERVER-STUDIO/ip)为自托管服务提供IP信息.
- 结果图表: 本项目加入了结果图表, 方便用户查看测速结果.
- 更加清晰的配置文件: 改进配置文件结构
- 前端页面: 内嵌前端页面, 无需额外配置即可部署.(仍与liberspeed前端保持兼容性)
- 重写: 对大部分组件进行重写与优化, 使程序更加易于维护的同时提升部分性能.

## 部署与使用

### Docker部署

参看[docker-cli部署SpeedTest-EX](https://github.com/WJQSERVER/speedtest-ex/blob/main/docs/docker/docker-cli_zh-cn.md)

参看[docker-compose部署SpeedTest-EX](https://github.com/WJQSERVER/speedtest-ex/blob/main/docs/docker/docker-compose_zh-cn.md)

### 配置文件

参看[配置文件说明](https://github.com/WJQSERVER/speedtest-ex/blob/main/docs/config/config_zh-cn.md)

## 前端页面

![SpeedTest-EX Index Page](https://webp.wjqserver.com/speedtest-ex/index.png)

![SpeedTest-EX Chart Page](https://webp.wjqserver.com/speedtest-ex/chart.png)


## License
Copyright (C) 2016-2020 Federico Dossena

Copyright (C) 2020 Maddie Zhan

Copyright (C) 2025 WJQSERVER

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Lesser General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU Lesser General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/lgpl>.