# 传入参数

``` bash
speedtest-ex -cfg /path/to/config/config.toml # 传入配置文件路径(必选)

# 以下参数可选
-port 8080 # 设置服务端口，默认8989

-auth #开启鉴权，默认关闭

-username admin # 设置用户名(需要开启鉴权)

-password admin # 设置密码(需要开启鉴权)

-secret rand # 设置密钥(需要开启鉴权) (rand为随机生成)

-initcfg # 初始化配置模式, 传入并保存配置, 用于快速初始化配置(保存配置后将会退出)

-dev # 开启开发模式，默认关闭(非开发用户请不要开启)

-version # 显示版本信息

```