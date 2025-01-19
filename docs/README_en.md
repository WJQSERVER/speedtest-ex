![SpeedTest-EX Logo](https://raw.githubusercontent.com/WJQSERVER/speedtest-ex/main/web/pages/favicon_inverted.png)

# SpeedTest-EX

This project is a significant refactor of the [speedtest-go](https://github.com/librespeed/speedtest-go) project. [speedtest-go](https://github.com/librespeed/speedtest-go) is a Go language reimplementation of the backend for [librespeed](https://github.com/librespeed/speedtest), while this project is a further refactor based on [speedtest-go](https://github.com/librespeed/speedtest-go). LiberSpeed is an open-source network speed testing project that uses PHP for its backend, whereas this project uses Go and the Gin framework for its backend, making the program more lightweight and easier to deploy.

**❗ Note**: Based on the principle of web speed testing, the program generates useless blocks for testers to download to calculate the actual downstream bandwidth, which carries a risk of being maliciously flooded with traffic. After sharing your speed test page externally, please monitor the server's traffic usage to avoid abnormal traffic usage.

## Features
- **Lightweight**: No additional environment required; just download the binary file to run. (Docker images are also more lightweight)
- **Easy to deploy**: Embedded frontend page, no additional configuration needed for deployment.
- **Efficient**: Based on the Gin framework, with strong concurrency handling capabilities and fast response times.

## Differences from speedtest-go
- **Web Framework**: speedtest-go uses the Chi framework, while this project uses the Gin framework.
- **IPinfo**: speedtest-go uses the ipinfo.io API to obtain IP information; this project is compatible with the ipinfo.io API and can also use [WJQSERVER-STUDIO/ip](https://github.com/WJQSERVER-STUDIO/ip) to provide IP information for self-hosted services.
- **Result Charts**: This project includes result charts for easier viewing of speed test results.
- **Clearer Configuration Files**: Improved configuration file structure.
- **Frontend Page**: Embedded frontend page, no additional configuration needed for deployment. (Still maintains compatibility with the liberspeed frontend)
- **Rewritten**: Most components have been rewritten and optimized, making the program easier to maintain while improving some performance.

## Deployment and Usage
### Deploy with Docker
Refer to [docker-cli deployment of SpeedTest-EX](https://github.com/WJQSERVER/speedtest-ex/blob/main/docs/docker/docker-cli_en.md)  
Refer to [docker-compose deployment of SpeedTest-EX](https://github.com/WJQSERVER/speedtest-ex/blob/main/docs/docker/docker-compose_en.md)

### Deploy Speedtest-EX on OpenWrt
Refer to [Running Speedtest-EX on OpenWrt](docs/openwrt/README_en.md)

### Configuration File
Refer to [configuration file description](https://github.com/WJQSERVER/speedtest-ex/blob/main/docs/config/config_en.md)

## Frontend Page

![SpeedTest-EX Index Page](https://webp.wjqserver.com/speedtest-ex/index.png)

![SpeedTest-EX Chart Page](https://webp.wjqserver.com/speedtest-ex/chart.png)

## License
Copyright (C) 2016-2020 Federico Dossena

Copyright (C) 2020 Maddie Zhan

Copyright (C) 2025 WJQSERVER

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Lesser General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU Lesser General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/lgpl>.
