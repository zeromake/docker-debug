# Docker-debug

[![Build Status](https://github.com/zeromake/docker-debug/actions/workflows/release.yml/badge.svg)](https://github.com/zeromake/docker-debug/actions/workflows/release.yml)
[![Go Report Card](https://goreportcard.com/badge/zeromake/docker-debug)](https://goreportcard.com/report/zeromake/docker-debug)

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
brew install zeromake/docker-debug/docker-debug
```

**download binary file**

<details>
<summary>
<kbd>use bash or zsh</kbd>
</summary>

``` bash
# get latest tag
VERSION=`curl -w '%{url_effective}' -I -L -s -S https://github.com/zeromake/docker-debug/releases/latest -o /dev/null | awk -F/ '{print $NF}'`

# MacOS Intel
curl -Lo docker-debug https://github.com/zeromake/docker-debug/releases/download/${VERSION}/docker-debug-darwin-amd64

# MacOS M1
curl -Lo docker-debug https://github.com/zeromake/docker-debug/releases/download/${VERSION}/docker-debug-darwin-arm64

# Linux
curl -Lo docker-debug https://github.com/zeromake/docker-debug/releases/download/${VERSION}/docker-debug-linux-amd64

chmod +x ./docker-debug
sudo mv docker-debug /usr/local/bin/

# Windows
curl -Lo docker-debug.exe https://github.com/zeromake/docker-debug/releases/download/${VERSION}/docker-debug-windows-amd64.exe
```

</details>

<details>
<summary>
<kbd>use fish</kbd>
</summary>

``` fish
# get latest tag
set VERSION (curl -w '%{url_effective}' -I -L -s -S https://github.com/zeromake/docker-debug/releases/latest -o /dev/null | awk -F/ '{print $NF}')

# MacOS Intel
curl -Lo docker-debug https://github.com/zeromake/docker-debug/releases/download/$VERSION/docker-debug-darwin-amd64

# MacOS M1
curl -Lo docker-debug https://github.com/zeromake/docker-debug/releases/download/$VERSION/docker-debug-darwin-arm64

# Linux
curl -Lo docker-debug https://github.com/zeromake/docker-debug/releases/download/$VERSION/docker-debug-linux-amd64

chmod +x ./docker-debug
sudo mv docker-debug /usr/local/bin/

# Windows
curl -Lo docker-debug.exe https://github.com/zeromake/docker-debug/releases/download/$VERSION/docker-debug-windows-amd64.exe
```
</details>


download the latest binary from the [release page](https://github.com/zeromake/docker-debug/releases/lastest) and add it to your PATH.

**Try it out!**
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
version = "0.7.5"
image = "nicolaka/netshoot:latest"
mount_dir = "/mnt/container"
timeout = 10000000000
config_default = "default"

[config]
  [config.default]
    version = "1.40"
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
- [x] mount volume filesystem
- [x] docker connection config on cli command
- [x] `-v` cli args support
- [ ] docker-debug signal handle smooth exit
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

## Contributors

### Code Contributors

This project exists thanks to all the people who contribute. [[Contribute](CONTRIBUTING.md)].
<a href="https://github.com/zeromake/docker-debug/graphs/contributors"><img src="https://opencollective.com/docker-debug/contributors.svg?width=890&button=false" /></a>

### Financial Contributors

Become a financial contributor and help us sustain our community. [[Contribute](https://opencollective.com/docker-debug/contribute)]

#### Individuals

<a href="https://opencollective.com/docker-debug"><img src="https://opencollective.com/docker-debug/individuals.svg?width=890"></a>

#### Organizations

Support this project with your organization. Your logo will show up here with a link to your website. [[Contribute](https://opencollective.com/docker-debug/contribute)]

<a href="https://opencollective.com/docker-debug/organization/0/website"><img src="https://opencollective.com/docker-debug/organization/0/avatar.svg"></a>
<a href="https://opencollective.com/docker-debug/organization/1/website"><img src="https://opencollective.com/docker-debug/organization/1/avatar.svg"></a>
<a href="https://opencollective.com/docker-debug/organization/2/website"><img src="https://opencollective.com/docker-debug/organization/2/avatar.svg"></a>
<a href="https://opencollective.com/docker-debug/organization/3/website"><img src="https://opencollective.com/docker-debug/organization/3/avatar.svg"></a>
<a href="https://opencollective.com/docker-debug/organization/4/website"><img src="https://opencollective.com/docker-debug/organization/4/avatar.svg"></a>
<a href="https://opencollective.com/docker-debug/organization/5/website"><img src="https://opencollective.com/docker-debug/organization/5/avatar.svg"></a>
<a href="https://opencollective.com/docker-debug/organization/6/website"><img src="https://opencollective.com/docker-debug/organization/6/avatar.svg"></a>
<a href="https://opencollective.com/docker-debug/organization/7/website"><img src="https://opencollective.com/docker-debug/organization/7/avatar.svg"></a>
<a href="https://opencollective.com/docker-debug/organization/8/website"><img src="https://opencollective.com/docker-debug/organization/8/avatar.svg"></a>
<a href="https://opencollective.com/docker-debug/organization/9/website"><img src="https://opencollective.com/docker-debug/organization/9/avatar.svg"></a>
