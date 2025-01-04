# Deploying SpeedTest-EX Using Docker Compose

## Create Directory

Here, we will use `/root/data/docker_data/speedtest-ex` as an example. Create the directory and navigate into it:

```bash
mkdir -p /root/data/docker_data/speedtest-ex
cd /root/data/docker_data/speedtest-ex
```

## Create Compose File

```bash
touch docker-compose.yml
```

## Edit Compose File

```yaml
version: '3'
services:
  speedtest-ex:
    image: 'wjqserver/speedtest-ex:latest'
    restart: always
    volumes:
      - './speedtest-ex/config:/data/speedtest-ex/config' # Configuration file
      - './speedtest-ex/log:/data/speedtest-ex/log' # Log file
      - './speedtest-ex/db:/data/speedtest-ex/db' # Database file
    ports:
      - '8989:8989' # Port mapping
```

## Start the Service

```bash
docker-compose up -d
```