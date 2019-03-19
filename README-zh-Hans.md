# Docker-debug

[English](README.md) ∙ [简体中文](README-zh-Hans.md)

## Overview

`docker-debug` 是一个运行中的 `docker` 容器故障排查方案,
在运行中的 `docker` 上额外启动一个容器，并将目标容器的 `pid`, `network`, `uses`, `filesystem` 和 `ipc` 命名空间注入到新的容器里，
因此，您可以使用任意故障排除工具，而无需在生产容器镜像中预先安装额外的工具环境。

## Demo
[![asciicast](https://asciinema.org/a/234638.svg)](https://asciinema.org/a/234638)
## Quick Start

安装 `docker-debug` 命令行工具
``` shell
# MacOS
curl -Lo docker-debug https://github.com/zeromake/docker-debug/releases/download/v0.1.0/docker-debug-darwin-amd64

# Linux
curl -Lo docker-debug https://github.com/zeromake/docker-debug/releases/download/v0.1.0/docker-debug-linux-amd64

chmod +x ./docker-debug
sudo mv docker-debug /usr/local/bin/

# Windows
curl -Lo docker-debug.exe https://github.com/zeromake/docker-debug/releases/download/v0.1.0/docker-debug-windows-amd64.exe
```

或者到 [release page](https://github.com/zeromake/docker-debug/releases/lastest) 下载最新可执行文件并添加到 PATH。

我们来试试！
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

## 默认镜像
docker-debug 使用 `nicolaka/netshoot` 作为默认镜像来运行额外容器。
你可以通过命令行 `flag(--image)` 覆盖默认镜像，或者直接修改配置文件 `~/.docker-debug/config.toml` 中的 `image`。
``` toml
# 默认镜像
image = "nicolaka/netshoot:latest"
# 大多数docker操作的超时，默认为 10s。
timeout = 10000000000
# 默认使用哪个配置来连接docker
config_default = "default"

# docker 连接配置
[config]
  # docker 默认连接配置
  [config.default]
    host = "unix:///var/run/docker.sock"
    # 是否为 tls
    tls = false
    # 证书目录
    cert_dir = ""
    # 证书密码
    cert_password = ""
```

## 详细
1. 在 `docker` 中查找镜像，没有调用 `docker` 拉取镜像。
2. 查找目标容器, 没找到返回报错。
3. 通过自定义镜像创建一个容器并挂载 `ipc`, `pid`, `network`, `etc`, `filesystem`。
4. 在新容器中创建并运行 `docker exec`。
5. 在新容器中进行调试。
6. 等待调试容器退出运行，把调试用的额外容器清理掉。

## Reference & Thank
1. [kubectl-debug](https://github.com/aylei/kubectl-debug): `docker-debug` 想法来自这个 kubectl 调试工具。
2. [Docker核心技术与实现原理](https://draveness.me/docker): `docker-debug` 的文件系统挂载原理来自这个博文。
3. [docker-engine-api-doc](https://docs.docker.com/engine/api/latest): docker engine api 文档。
