# Running Speedtest-EX on OpenWrt

## Download Build
- Currently only provides `apk`(OpenWrt-SNAPSHOT) and `ipk` files for `arm64` and `amd64` architectures
https://github.com/JohnsonRan/packages_builder/releases

## Installation
- Go to OpenWrt management interface, navigate to `System -> Software` page to upload and install the package

## Getting Started
- Configuration file is located at `/etc/speedtest-ex/config.toml`
- After modifying the configuration, execute `/etc/init.d/speedtest-ex restart` to restart the service
- Default running on port `8989`
- If no longer needed, simply remove `speedtest-ex` from the software packages

### Compile by Yourself
- You can use this repo to compile packages for any architecture for installation  
https://github.com/JohnsonRan/packages_net_speedtest-ex