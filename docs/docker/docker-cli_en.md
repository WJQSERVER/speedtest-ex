# Deploying SpeedTest-EX Using Docker-Cli

```bash
# Run the container
docker run -d \
  --name speedtest-ex \
  --restart always \
  -v ./speedtest-ex/config:/data/speedtest-ex/config \
  -v ./speedtest-ex/log:/data/speedtest-ex/log \
  -v ./speedtest-ex/db:/data/speedtest-ex/db \
  -p 8989:8989 \
  wjqserver/speedtest-ex:latest
```