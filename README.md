# bastion-web-proxy

A simple web proxy that supports:
* Serving a proxy auto-configuration file.
* Proxying through an SSH tunnel:
    * SOCKS5 local client connections
    * TCP local connections

## Building

Requirements:

* The `make` command (e.g. [GNU make](https://www.gnu.org/software/make/manual/make.html)).
* The [Golang toolchain](https://golang.org/doc/install) (version 1.18 or later).

In a shell, execute: `make` (or `make build`).

The build artifacts can be cleaned by using: `make clean`.

## Configuration

The proxy is configured via a YAML configuration file (`config.yaml` by default).

See `config/config.yaml.example` for an example.

## Running

Using this Git repository, in a shell, execute: `make run`

or

Using the release tarball, installed in your PATH: `bastion-web-proxy`

```
Usage of bastion-web-proxy:
      --config-location string   The path to the directory where the configuration file is stored. (default ".")
      --log-level string         The logging level (trace, debug, info...) (default "info")
```

## TODOs

* Automatically reconnect the SSH connection in case of disconnect
* Automatically disconnect the SSH connection when idling
* Implement a basic status page
* Test client-side connection error handling
* Generate the proxy auto-configuration file from the configuration?
