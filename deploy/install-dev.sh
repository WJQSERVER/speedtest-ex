#! /bin/bash
# https://github.com/WJQSERVER/speedtest-ex
# LGPL-3.0 License
# Copyright (c) 2025 WJQSERVER

bin_dir_default="/usr/local/speedtest-ex"
bin_dir=$bin_dir_default

# install packages
install() {
    if [ $# -eq 0 ]; then
        echo "ARGS NOT FOUND"
        return 1
    fi

    for package in "$@"; do
        if ! command -v "$package" &>/dev/null; then
            echo "Installing $package..."
            if command -v dnf &>/dev/null; then
                dnf -y update && dnf install -y "$package"
            elif command -v yum &>/dev/null; then
                yum -y update && yum -y install "$package"
            elif command -v apt &>/dev/null; then
                apt update -y && apt install -y "$package"
            elif command -v apk &>/dev/null; then
                apk update && apk add "$package"
            else
                echo "UNKNOWN PACKAGE MANAGER"
                return 1
            fi
        else
            echo "$package is already installed."
        fi
    done

    return 0
}

# 检查是否为root用户
if [ "$EUID" -ne 0 ]; then
    echo "请以root用户运行此脚本"
    exit 1
fi

# 安装依赖包
install curl wget sed

# 查看当前架构是否为linux/amd64或linux/arm64
ARCH=$(uname -m)
if [ "$ARCH" != "x86_64" ] && [ "$ARCH" != "aarch64" ]; then
    echo "$ARCH 架构不被支持"
    exit 1
fi

# 重写架构值,改为amd64或arm64
if [ "$ARCH" == "x86_64" ]; then
    ARCH="amd64"
elif [ "$ARCH" == "aarch64" ]; then
    ARCH="arm64"
fi

# 获取监听端口
read -p "请输入程序监听的端口(默认8989): " PORT
if [ -z "$PORT" ]; then
    PORT=8989
fi

# 泛监听(0.0.0.0)
read -p "请键入程序监听的IP(默认泛监听0.0.0.0): " IP
if [ -z "$IP" ]; then
    IP="0.0.0.0"
fi

# 安装目录
read -p "请输入安装目录(默认${bin_dir_default}): " bin_dir
if [ -z "$bin_dir" ]; then
   bin_dir=${bin_dir_default}
fi

make_systemd_service() {
    cat <<EOF > /etc/systemd/system/speedtest-ex.service
[Unit]
Description=SpeedTest-EX
After=network.target

[Service]
ExecStart=/bin/bash -c '${bin_dir}/speedtest-ex -cfg ${bin_dir}/config/config.toml > ${bin_dir}/log/run.log 2>&1'
WorkingDirectory=${bin_dir}
Restart=always
User=root
Group=root

[Install]
WantedBy=multi-user.target

EOF
}

make_openrc_service() {
    cat <<EOF > /etc/init.d/speedtest-ex
#!/sbin/openrc-run

command="${bin_dir}/speedtest-ex"
command_args="-cfg ${bin_dir}/config/config.toml"
pidfile="/run/speedtest-ex.pid"
name="speedtest-ex"

depend() {
    need net
}

start_pre() {
    checkpath --directory --mode 0755 /run
}

start() {
    ebegin "Starting ${name}"
    start-stop-daemon --start --make-pidfile --pidfile ${pidfile} --background --exec ${command} -- ${command_args}
    eend $?
}

stop() {
    ebegin "Stopping ${name}"
    start-stop-daemon --stop --pidfile ${pidfile}
    eend $?
}
EOF
    chmod +x /etc/init.d/speedtest-ex
}

make_procd_service() {
    config_path="${bin_dir}/config/config.toml"

    cat <<EOF > ${bin_dir}/boot.sh
#!/bin/sh
cd ${bin_dir}
${bin_dir}/speedtest-ex -cfg ${config_path}

EOF

    chmod +x ${bin_dir}/boot.sh

    cat <<EOF > /etc/init.d/speedtest-ex
#!/bin/sh /etc/rc.common

START=99
USE_PROCD=1

start_service() {
    procd_open_instance
    procd_set_param command ${bin_dir}/boot.sh
    procd_close_instance
}

stop_service() {
    pid=\$(pidof speedtest-ex)
    [ -n "\$pid" ] && kill \$pid
}



EOF
    chmod +x /etc/init.d/speedtest-ex
}


# 创建目录
mkdir -p ${bin_dir}
mkdir -p ${bin_dir}/config
mkdir -p ${bin_dir}/log
mkdir -p ${bin_dir}/db

# 获取最新版本号
VERSION=$(curl -s https://raw.githubusercontent.com/WJQSERVER/speedtest-ex/dev/DEV-VERSION)
if [ -z "$VERSION" ]; then
    echo "无法获取版本号"
    exit 1
fi
wget -q -O ${bin_dir}/VERSION https://raw.githubusercontent.com/WJQSERVER/speedtest-ex/dev/DEV-VERSION

# 下载speedtest-ex
wget -q -O ${bin_dir}/speedtest-ex.tar.gz https://github.com/WJQSERVER/speedtest-ex/releases/download/${VERSION}/speedtest-ex-linux-${ARCH}.tar.gz
if [ $? -ne 0 ]; then
    echo "下载失败，请检查网络连接"
    exit 1
fi

install tar
tar -zxvf ${bin_dir}/speedtest-ex.tar.gz -C ${bin_dir}
mv ${bin_dir}/speedtest-ex-linux-${ARCH} ${bin_dir}/speedtest-ex
rm ${bin_dir}/speedtest-ex.tar.gz
chmod +x ${bin_dir}/speedtest-ex

# 下载配置文件
if [ -f ${bin_dir}/config/config.toml ]; then
    echo "配置文件已存在, 跳过下载"
    echo "[WARNING] 请检查配置文件是否正确，DEV版本升级时请注意配置文件兼容性"
    sleep 2
else
    wget -q -O ${bin_dir}/config/config.toml https://raw.githubusercontent.com/WJQSERVER/speedtest-ex/dev/deploy/config.toml
fi



# 询问是否开启鉴权(Y/N)
read -p "是否开启鉴权(Y/N): " isAuth
if [[ "$isAuth" =~ ^[Yy]$ ]]; then
    read -p "请输入用户名: " username
    read -p "请输入密码: " password
    read -p "请输入密钥(默认rand): " secret
    # 若secret不为空，则写入配置文件
    cd ${bin_dir}
    ./speedtest-ex -cfg ./config/config.toml -port ${PORT} -user ${username} -password ${password} -secret ${secret:-rand} -auth -initcfg
else
    cd ${bin_dir}
    ./speedtest-ex -cfg ./config/config.toml -port ${PORT} -initcfg
fi

# 判断发行版并创建相应的服务
if command -v systemctl &>/dev/null; then
    make_systemd_service
    systemctl daemon-reload
    systemctl enable speedtest-ex
    systemctl start speedtest-ex
elif [ -x /sbin/openrc ] || [ -x /usr/bin/openrc ]; then
    make_openrc_service
    rc-update add speedtest-ex default
    service speedtest-ex start
elif [ -x /sbin/procd ]; then
    make_procd_service
    /etc/init.d/speedtest-ex enable
    /etc/init.d/speedtest-ex start
else
    echo "不支持的服务管理器"
    exit 1
fi

echo "speedtest-ex 安装成功"