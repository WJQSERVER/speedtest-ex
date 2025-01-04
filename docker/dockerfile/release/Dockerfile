FROM alpine:latest AS builder

ARG USER=WJQSERVER
ARG REPO=speedtest-ex
ARG APPLICATION=speedtest-ex
ARG TARGETOS
ARG TARGETARCH
ARG TARGETPLATFORM

# 拉取依赖
RUN apk add --no-cache wget curl

# 创建目录
RUN mkdir -p /data/${APPLICATION}
RUN mkdir -p /data/${APPLICATION}/config
RUN mkdir -p /data/${APPLICATION}/logs
RUN mkdir -p /data/${APPLICATION}/db

# 后端
RUN VERSION=$(curl -s https://raw.githubusercontent.com/${USER}/${REPO}/main/VERSION) && \
    wget -O /data/${APPLICATION}/${APPLICATION} https://github.com/${USER}/${REPO}/releases/download/$VERSION/${APPLICATION}-${TARGETOS}-${TARGETARCH}
RUN wget -O /data/${APPLICATION}/config.toml https://raw.githubusercontent.com/${USER}/${REPO}/main/config/config.toml
RUN wget -O /usr/local/bin/init.sh https://raw.githubusercontent.com/${USER}/${REPO}/main/docker/dockerfile/dev/init.sh

# 权限
RUN chmod +x /data/${APPLICATION}/${APPLICATION}
RUN chmod +x /usr/local/bin/init.sh

FROM alpine:latest

COPY --from=builder /data/${APPLICATION} /data/${APPLICATION}
COPY --from=builder /usr/local/bin/init.sh /usr/local/bin/init.sh

# 权限
RUN chmod +x /data/${APPLICATION}/${APPLICATION}
RUN chmod +x /usr/local/bin/init.sh

CMD ["/usr/local/bin/init.sh"]


