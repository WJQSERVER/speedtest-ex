#!/bin/sh

APPLICATION=speedtest-go

# 检查并复制 config.toml
if [ ! -f /data/${APPLICATION}/config/config.toml ]; then
    cp /data/${APPLICATION}/config.toml /data/${APPLICATION}/config/config.toml
fi

# 启动 Go 应用
/data/${APPLICATION}/${APPLICATION} -cfg /data/${APPLICATION}/config/config.toml > /data/${APPLICATION}/log/run.log 2>&1 &

# 保持脚本运行
while true; do
    sleep 1
done