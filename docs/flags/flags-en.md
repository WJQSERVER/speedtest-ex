# Input Parameters

``` bash
speedtest-ex -cfg /path/to/config/config.toml # Path to the configuration file (required)

# The following parameters are optional
-port 8080 # Set the server port, default is 8989

-auth # Enable authentication, default is off

-username admin # Set the username (authentication must be enabled)

-password admin # Set the password (authentication must be enabled)

-secret rand # Set the secret key (authentication must be enabled) (rand is randomly generated)

-initcfg # Initialise configuration mode, input and save configuration for quick setup (will exit after saving configuration)

-dev # Enable development mode, default is off (do not enable for non-development users)

```