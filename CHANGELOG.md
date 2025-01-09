# 更新日志

25w08a
---
- PRE-RELEASE: 作为0.0.8的预发布版本, 请勿用于生产环境
- CHANGE: 大量扩充可传入的flag
- CHANGE: 修改`config`模块, 加入保存配置与重载配置
- CHANGE: 加入通过`crypto/rand`生成secret key的功能
- CHANGE: 增加前端版本号显示

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
