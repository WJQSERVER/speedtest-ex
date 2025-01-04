# Configuration File

Below is an example of the configuration file within a Docker container.

The configuration file is located at `{installation_directory}/speedtest-ex/config/config.toml`

```toml
[server]
host = "0.0.0.0" # Listening address
port = 8989 # Listening port
basePath = "" # Retained for compatibility with LiberSpeed; do not modify if not needed

[log]
logFilePath = "/data/speedtest-go/log/speedtest-go.log" # Path to the log file
maxLogSize = 5 # MB Maximum size of the log file

[ipinfo]
model = "ipinfo" # ip (self-hosted) or ipinfo (ipinfo.io)
ipinfo_url = "" # Please fill in the API address of ipinfo.io when self-hosting
ipinfo_api_key = "" # ipinfo.io API key, if available

[database]
model = "bolt"  # Database type, currently only supports BoltDB
path = "/data/speedtest-go/db/speedtest.db" # Path to the database file

[frontend]
chartlist = 100 # Default to display the most recent 100 entries
``` 
