# 使用 Docker Compose 部署 SpeedTest-EX

## 創建目錄

此處以 `/root/data/docker_data/speedtest-ex` 為例，創建目錄並進入：

```bash
mkdir -p /root/data/docker_data/speedtest-ex
cd /root/data/docker_data/speedtest-ex
```

## 創建 Compose 文件

```bash
touch docker-compose.yml
```

## 編輯 Compose 文件

```yaml
version: '3'
services:
  speedtest-ex:
    image: 'wjqserver/speedtest-ex:latest'
    restart: always
    volumes:
      - './speedtest-ex/config:/data/speedtest-ex/config' # 配置文件
      - './speedtest-ex/log:/data/speedtest-ex/log' # 日誌文件
      - './speedtest-ex/db:/data/speedtest-ex/db' # 資料庫文件
    ports:
      - '8989:8989' # 端口映射
```

## 啟動服務

```bash
docker compose up -d
```