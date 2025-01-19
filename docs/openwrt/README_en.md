# Running Speedtest-EX on OpenWrt

> This is not an official adaptation of the project. If you have any issues, please refer to https://github.com/JohnsonRan/packages_net_speedtest-ex 

## Download and Build
- Currently, only `arm64` and `amd64` architecture `apk` (OpenWrt-SNAPSHOT) and `ipk` files are provided.
https://github.com/JohnsonRan/packages_builder/releases

## Installation
- Navigate to the OpenWrt management interface and go to the `System -> Packages` page to upload and install the package.

## Getting Started
- The configuration file can be found at `/etc/speedtest-ex/config.toml`
- After modifying the configuration, execute `/etc/init.d/speedtest-ex restart` to restart the service.
- It runs by default on port `8989`.
- If no longer needed, simply go to the packages section to remove `speedtest-ex`.

### Self-Compilation
The project also provides libraries for self-compilation.  
https://github.com/JohnsonRan/packages_net_speedtest-ex  
You can use this library to compile packages for any architecture for installation.
