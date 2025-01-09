# 使用 Docker Compose 部署 SpeedTest-EX

## 创建目录

此处以`/root/data/docker_data/speedtest-ex`为例，创建目录并进入：

```bash
mkdir -p /root/data/docker_data/speedtest-ex
cd /root/data/docker_data/speedtest-ex
```

## 创建compose文件

```bash
touch docker-compose.yml
```

## 编辑compose文件

```yaml
version: '3'
services:
  speedtest-ex:
    image: 'wjqserver/speedtest-ex:latest'
    restart: always
    volumes:
      - './speedtest-ex/config:/data/speedtest-ex/config' # 配置文件
      - './speedtest-ex/log:/data/speedtest-ex/log' # 日志文件
      - './speedtest-ex/db:/data/speedtest-ex/db' # 数据库文件
    ports:
      - '8989:8989' # 端口映射
```

## 启动服务

```bash
docker compose up -d
```