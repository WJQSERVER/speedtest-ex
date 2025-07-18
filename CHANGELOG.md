# 更新日志

1.1.3 - 2025-07-18
---
- CHANGE: 更新依赖, 提升性能与安全性

1.1.2 - 2025-06-13
---
- CHANGE: 更新依赖

25w17a - 2025-06-13
---
- PRE-RELEASE: 此版本是v1.1.2的预发布版本; 
- CHANGE: 更新依赖

1.1.1 - 2025-06-08
---
- CHANGE: 为`revping`加入非特权回退功能

25w16a - 2025-06-08
---
- PRE-RELEASE: 此版本是v1.1.1的预发布版本; 
- CHANGE: 为`revping`加入非特权回退功能

1.1.0 - 2025-06-07
---
- CHANGE: 转向Touka框架
- CHANGE: 弃用-cfg, 换为-c (保留兼容性支持, -cfg仍然正常工作)
- CHANGE: 下载数据生成增加流模式
- CHANGE: 加入Compress支持
- CHANGE: 优化测速脚本性能
- CHANGE: 对实现进行优化, 避免ws的一些关闭问题, 提升下载数据生成效率
- CHANGE: 改进前端内容handle处置
- ADD: 加入CLI

25w15b - 2025-06-07
---
- PRE-RELEASE: 此版本是v1.1.0的预发布版本; 
- CHANGE: 下载数据生成增加流模式
- CHANGE: 加入Compress支持

25w15a - 2025-06-06
---
- PRE-RELEASE: 此版本是v1.1.0的预发布版本; 
- CHANGE: 转向Touka框架
- CHANGE: 弃用-cfg, 换为-c (保留兼容性支持, -cfg仍然正常工作)
- CHANGE: 优化测速脚本性能
- CHANGE: 对实现进行优化, 避免ws的一些关闭问题, 提升下载数据生成效率
- CHANGE: 改进前端内容handle处置
- ADD: 加入CLI

1.0.1 - 2025-06-04
---
- CHANGE: 更新依赖
- CHANGE: 弃用对req的依赖, 转向touka-httpc

25w14b - 2025-06-04
---
- PRE-RELEASE: 此版本是v1.0.1的预发布版本; 
- CHANGE: 弃用对req的依赖, 转向touka-httpc

25w14a - 2025-06-04
---
- PRE-RELEASE: 此版本是v1.0.1的预发布版本; 
- CHANGE: 更新依赖

1.0.0 - 2025-03-19
---
- RELEASE: 首个正式稳定版

25w13a - 2025-03-19
---
- PRE-RELEASE: 此版本是v1.0.0的预发布版本; 此项目已趋于稳定;

0.0.12
---
- CHANGE: 关闭Gin上传缓冲区
- CHANGE: 更新Go版本至1.24
- CHANGE: 更新相关依赖库

25w12a
---
- CHANGE: 关闭Gin上传缓冲区
- CHANGE: 更新Go版本至1.24

0.0.11
---
- CHANGE: 优化`empty`部分实现
- CHANGE: 将`logger`升级至`v1.2.0`版本
- CHANGE: 更新相关依赖库
- CHANGE: 更新Go版本至1.23.6

25w11b
---
- CHANGE: 优化`empty`部分实现

25w11a
---
- CHANGE: 改进`empty`部分实现
- CHANGE: 将`logger`升级至`v1.2.0`版本

0.0.10
---
- CHANGE: 更新Go版本至1.23.5
- ADD: 加入`-version`命令行参数, 用于显示当前版本信息
- CHANGE: 将`net`换为`net/netip`

25w10a
---
- PRE-RELEASE: 作为0.0.10的预发布版本, 请勿用于生产环境
- CHANGE: 更新Go版本至1.23.5
- ADD: 加入`-version`命令行参数, 用于显示当前版本信息
- CHANGE: 将`net`换为`net/netip`

0.0.9
---
- CHANGE: 将`Revping`转为通过`WebSocket`通道传递数据
- CHANGE: 完善bin安装脚本
- REMOVE: 移除部分无用保留页面
- CHANGE: 关闭`gin`日志输出, 避免影响性能(go终端日志输出性能较差, 易成为性能瓶颈)
- CHANGE: 新增`[Speedtest]`配置块,`downDataChunkSize = 4 #mb` `downDataChunkCount = 4` 分别用于设置下载数据块大小与块数量, 配置更加灵活

25w09b
---
- PRE-RELEASE: 作为0.0.9的预发布版本, 请勿用于生产环境
- CHANGE: 关闭`gin`日志输出, 避免影响性能(go终端日志输出性能较差, 易成为性能瓶颈)
- CHANGE: 新增`[Speedtest]`配置块,`downDataChunkSize = 4 #mb` `downDataChunkCount = 4` 分别用于设置下载数据块大小与块数量, 配置更加灵活

25w09a
---
- PRE-RELEASE: 作为0.0.9的预发布版本, 请勿用于生产环境
- CHANGE: 将`Revping`转为通过`WebSocket`通道传递数据
- CHANGE: 完善bin安装脚本
- REMOVE: 移除部分保留页面

0.0.8
---
- CHANGE: 大量扩充可传入的flag
- CHANGE: 修改`config`模块, 加入保存配置与重载配置
- CHANGE: 加入通过`crypto/rand`生成secret key的功能
- CHANGE: 增加前端版本号显示

25w08b
---
- PRE-RELEASE: 作为0.0.8的预发布版本, 请勿用于生产环境
- CHANGE: 大量扩充可传入的flag
- CHANGE: 修改`config`模块, 加入保存配置与重载配置
- CHANGE: 加入通过`crypto/rand`生成secret key的功能
- CHANGE: 增加前端版本号显示

25w08a
---
- PRE-RELEASE: 由于当日Github故障造成的CI/CD流程错误, 此版本弃用

0.0.7
---
- ADD: 加入鉴权功能与对应的前端页面, 同时对核心`speedtest.js`与`speedtest_worker.js`内请求后端的部分进行了适应性修改 (实验性, 需要更多测试)
- CHANGE: 改进前端静态文件处理, 进行一定改进
- CHANGE: 优化前端显示
- CHANGE: 改进`revping`功能, 加入熔断, 若由于`timeout`或`revping-not-online`导致的无法返回结果, 在连续失败后会停止发起请求, 避免阻塞
- CHANGE: 对`route.go`进行了优化, 独立部分处理逻辑

25w07b
---
- PRE-RELEASE: 作为0.0.7的预发布版本, 请勿用于生产环境
- CHANGE: 改进`revping`功能, 加入熔断, 若由于`timeout`或`revping-not-online`导致的无法返回结果, 在连续失败后会停止发起请求, 避免阻塞

25w07a
---
- PRE-RELEASE: 作为0.0.7的预发布版本, 请勿用于生产环境
- CHANGE: 优化前端显示
- ADD: 加入鉴权功能与对于的前端页面, 同时对核心`speedtest.js`与`speedtest_worker.js`内请求后端的部分进行了适应性修改 (实验性, 需要更多测试)
- CHANGE: 改进前端静态文件处理, 进行一定改进
- CHANGE: 对`route.go`进行了优化, 独立部分处理逻辑

0.0.6
---
- FIX: 修复工作流配置导致的编译问题
- CHANGE: 改进revping前端显示

25w06a
---
- PRE-RELEASE: 作为0.0.6的预发布版本, 请勿用于生产环境
- FIX: 修复工作流配置导致的编译问题
- CHANGE: 改进revping前端显示

0.0.5
---
- FIX: 修复遥测数据上传问题
- PR: [PR #8](https://github.com/WJQSERVER/speedtest-ex/pull/8)

25w05a
---
- PRE-RELEASE: 作为的修复的预发布版本, 请勿用于生产环境
- FIX: 修复遥测数据上传问题
- PR: [PR #7](https://github.com/WJQSERVER/speedtest-ex/pull/7)

0.0.4
---
- ADD: 加入反向ping显示
- ADD: 加入单下载流测速功能

25w04a
---
- PRE-RELEASE: 作为0.0.4的预发布版本, 请勿用于生产环境
- ADD: 加入反向ping显示
- ADD: 加入单下载流测速功能

0.0.3
---
- RELEASE: 0.0.3正式版本发布
- CHANGE: 优化前端页面显示
- CHANGE: 更新文档

25w03a
---
- PRE-RELEASE: 作为0.0.3的预发布版本, 请勿用于生产环境
- FIX: 修复前端页面显示问题

0.0.2
---
- RELEASE: 0.0.2正式版本发布
- CHANGE: 优化部分错误处理
- CHANGE&FIX: 修正文档错误

25w02a
---
- PRE-RELEASE: 作为0.0.2的预发布版本, 请勿用于生产环境
- CHANGE&FIX: 修正文档错误
- CHANGE: 优化部分错误处理

0.0.1
---
- RELEASE: 0.0.1正式版本发布

25w01d
---
- PRE-RELEASE: 作为0.0.1的预发布版本, 请勿用于生产环境
- FIX: 修复配置问题

25w01c
---
- PRE-RELEASE: 作为0.0.1的预发布版本, 请勿用于生产环境
- FIX: 修复了一些已知问题
- CHANGE: 更新依赖库

25w01b
---
- PRE-RELEASE: 首个对外预发布版本

25w01a
---
- PRE-RELEASE: 由于.gitignore造成的CI/CD流程错误，此版本弃用
