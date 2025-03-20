# Running Speedtest-EX on OpenWrt

> This is not an official adaptation of the project. If you have any issues, please refer to https://github.com/JohnsonRan/InfinitySubstance

## Add feed
- Currently, only `arm64` and `amd64` architecture `apk` (OpenWrt-SNAPSHOT) and `ipk` packages are provided.
```shell
# only needs to be run once
curl -s -L https://github.com/JohnsonRan/InfinitySubstance/raw/main/feed.sh | ash
```

## Installation
- Navigate to the OpenWrt management interface and go to the `System -> Packages` page, search and install `speedtest-ex`

## Getting Started
- The configuration file can be found at `/etc/speedtest-ex/config.toml`
- After modifying the configuration, execute `/etc/init.d/speedtest-ex restart` to restart the service.
- It runs by default on port `8989`.
- If no longer needed, simply go to the packages section to remove `speedtest-ex`.

### Self-Compilation
The project also provides libraries for self-compilation.  
https://github.com/JohnsonRan/packages_net_speedtest-ex  
You can use this library to compile packages for any architecture for installation.
