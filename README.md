# Docker-debug

[English](README.md) ∙ [简体中文](README-zh-Hans.md)

## Overview

`docker-debug` is an troubleshooting running docker container,
which allows you to run a new container in running docker for debugging purpose.
The new container will join the `pid`, `network`, `user`, `filesystem` and `ipc` namespaces of the target container, 
so you can use arbitrary trouble-shooting tools without pre-installing them in your production container image.

## Demo
[![asciicast](https://asciinema.org/a/235025.svg)](https://asciinema.org/a/235025)
## Quick Start

Install the `docker-debug` cli

**mac brew**
```shell
brew tap zeromake/docker-debug
brew install docker-debug
```

**download binary file**
``` shell
# MacOS
curl -Lo docker-debug https://github.com/zeromake/docker-debug/releases/download/v0.2.1/docker-debug-darwin-amd64

# Linux
curl -Lo docker-debug https://github.com/zeromake/docker-debug/releases/download/v0.2.1/docker-debug-linux-amd64

chmod +x ./docker-debug
sudo mv docker-debug /usr/local/bin/

# Windows
curl -Lo docker-debug.exe https://github.com/zeromake/docker-debug/releases/download/v0.2.1/docker-debug-windows-amd64.exe
```

download the latest binary from the [release page](https://github.com/zeromake/docker-debug/releases/lastest) and add it to your PATH.

Try it out!
``` shell
# docker-debug [OPTIONS] CONTAINER COMMAND [ARG...] [flags]
docker-debug CONTAINER COMMAND

# More flags
docker-debug --help

# info
docker-debug info
```

## Build from source
Clone this repo and:
``` shell
go build -o docker-debug ./cmd/docker-debug
mv docker-debug /usr/local/bin
```

## Default image
docker-debug uses nicolaka/netshoot as the default image to run debug container.
You can override the default image with cli flag, or even better, with config file ~/.docker-debug/config.toml
``` toml
version = "0.2.1"
image = "nicolaka/netshoot:latest"
mount_dir = "/mnt/container"
timeout = 10000000000
config_default = "default"

[config]
  [config.default]
    host = "unix:///var/run/docker.sock"
    tls = false
    cert_dir = ""
    cert_password = ""
```

## Todo
- [x] support windows7(Docker Toolbox)
- [ ] support windows10
- [ ] refactoring code
- [ ] add testing
- [x] add changelog
- [x] add README_CN.md
- [x] add brew package
- [x] docker-debug version manage config file
- [x] cli command set mount target container filesystem
- [ ] mount volume filesystem
- [ ] cli command document on readme
- [ ] config file document on readme
- [ ] add http api and web shell

## Details
1. find image docker is has, not has pull the image.
2. find container name is has, not has return error.
3. from customize image runs a new container in the container's namespaces (ipc, pid, network, etc, filesystem) with the STDIN stay open.
4. create and run a exec on new container.
5. Debug in the debug container.
6. then waits for the debug container to exit and do the cleanup.

## Reference & Thank
1. [kubectl-debug](https://github.com/aylei/kubectl-debug): `docker-debug` inspiration is from to this a kubectl debug tool.
2. [Docker核心技术与实现原理](https://draveness.me/docker): `docker-debug` filesystem is from the blog.
3. [docker-engine-api-doc](https://docs.docker.com/engine/api/latest): docker engine api document.
