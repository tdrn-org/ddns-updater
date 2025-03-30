## About DDNS Updater
ddns-updater is a command line tool to perform dynamic DNS (DDNS) updates. It provides various finder mechanisms to gather the
running host's ip addresses and update mechanisms for several DNS backends.

The following address detection mechanisms are supported:
* Interface based (examining the addresses of the running host's interfaces)
* UPnP based (querying the local network's router for the external IPv4 address)
* Web based (querying one or more web based services to determine the running host's ip addresses)

The following DNS backends are supported:
* AWS/Route53 (updating AWS/Route53 zone information)
* Web (invoking a web based service to update DNS)

### Installation
To install ddns-updater you have to download a suitable [release archive](https://github.com/tdrn-org/ddns-updater/releases) and extract it or build it from source by cloning the repository and issueing a simple
```
make build
```
To build ddns-updater, Go version 1.20 or higher is required. The resulting binary will be written to **./build/bin**.
Copy the either extracted or built tool binary to a location of your choice (e.g. /usr/local/bin/).

### Configuration
The ddns-updater tool retrieves most of its configuration from a configuration file. See [ddns-updater.toml](https://github.com/tdrn-org/ddns-updater/blob/master/ddns-updater.toml) as a reference and adapt it to your need. The default location of the
configuration file is /etc/ddns-updater/ddns-updater.toml. An explicit location
can be given via the --config command line option (see next section).

### Usage
The ddns-updater tool supports the following command line: 
```
ddns-updater [--verbose|--debug] [--config=<config file>] [--pretend] [--force] [--reset-cache]

--verbose
	Enable verbose output.
--debug
	Enable debug output.
--config=<config file>
	Read configuration from file <config file>.
	If this option is not set, configuration is read from /etc/ddns-updater/ddns-updater.toml.
--pretend
	Gather the host's ip addresses and prepare a DDNS update, but do not apply it.
--force
	Forces an update (even if the DNS information is considered up-to-date).
--reset-cache
	Clears the tool's cache to restart from scratch.
```

### Changelog:
See [CHANGELOG.md](https://github.com/tdrn-org/ddns-updater/blob/master/CHANGELOG.md).

### License
This project is subject to the the MIT License.
See [LICENSE](https://github.com/tdrn-org/ddns-updater/blob/master/LICENSE) information for details.