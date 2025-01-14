# Configuration File

Below is an example of the configuration file within a Docker container.

The configuration file is located at `{installation_directory}/speedtest-ex/config/config.toml`

```toml
[server]
host = "0.0.0.0" # Listening address
port = 8989 # Listening port
basePath = "" # Retained for compatibility with LiberSpeed; do not modify if not needed

[Speedtest]
downDataChunkSize = 4 #mb Download data chunk size
downDataChunkCount = 4 # Download data chunk count

[log]
logFilePath = "/data/speedtest-ex/log/speedtest-ex.log" # Path to the log file
maxLogSize = 5 # MB Maximum size of the log file

[ipinfo]
model = "ipinfo" # ip (self-hosted) or ipinfo (ipinfo.io)
ipinfo_url = "" # Please fill in the API address of your self-hosted ipinfo service when self-hosting
ipinfo_api_key = "" # ipinfo.io API key, if available

[database]
model = "bolt"  # Database type, currently only supports BoltDB
path = "/data/speedtest-ex/db/speedtest.db" # Path to the database file

[frontend]
chartlist = 100 # Default to display the most recent 100 entries

[revping]
enable = true # Enable reverse ping test

[auth]
enable = false # Enable authentication
username = "admin" # Username for authentication
password = "password" # Password for authentication
secret = "secret" # Secret key for Generating Session Cookies. You should change this to a secure value.
``` 
