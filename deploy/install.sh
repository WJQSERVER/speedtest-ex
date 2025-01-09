# /bin/bash
# https://github.com/WJQSERVER/speedtest-ex

bin_dir_default="/usr/local/speedtest-ex"
bin_dir=bin_dir_default

# install packages
install() {
    if [ $# -eq 0 ]; then
        echo "ARGS NOT FOUND"
        return 1
    fi

    for package in "$@"; do
        if ! command -v "$package" &>/dev/null; then
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
    echo " $ARCH 架构不被支持"
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

# 创建目录
mkdir -p ${bin_dir}
mkdir -p ${bin_dir}/config
mkdir -p ${bin_dir}/log
mkdir -p ${bin_dir}/db

# 获取最新版本号
VERSION=$(curl -s https://raw.githubusercontent.com/WJQSERVER/speedtest-ex/main/VERSION)
wget -q -O ${bin_dir}/VERSION https://raw.githubusercontent.com/WJQSERVER/speedtest-ex/main/VERSION

# 下载speedtest-ex
wget -q -O ${bin_dir}/speedtest-ex https://github.com/WJQSERVER/speedtest-ex/releases/download/${VERSION}/speedtest-ex-linux-${ARCH}.tar.gz
install tar
tar -zxvf ${bin_dir}/speedtest-ex-linux-${ARCH}.tar.gz -C ${bin_dir}
chmod +x ${bin_dir}/speedtest-ex


# 下载配置文件
if [ -f ${bin_dir}/config/config.toml ]; then
    echo "配置文件已存在, 跳过下载"
    echo "[WARNING] 请检查配置文件是否正确，DEV版本升级时请注意配置文件兼容性"
    sleep 2
else
    wget -q -O ${bin_dir}/config/config.toml https://raw.githubusercontent.com/WJQSERVER/speedtest-ex/main/config/config.toml
fi

# 下载systemd服务文件
if [ "$bin_dir" = "${bin_dir_default}" ]; then
    make_systemd_service()
    # todo

else
    make_systemd_service()
fi

# todo配置修改

# 启动speedtest-ex
systemctl daemon-reload
systemctl enable speedtest-ex
systemctl start speedtest-ex

echo "speedtest-ex 安装成功"
