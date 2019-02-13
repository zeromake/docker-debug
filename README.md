# Docker-debug

## Overview

`docker-debug` which allows you to run a new container in running docker for debugging purpose. The new container will join the pid, network, user and ipc namespaces of the target container, so you can use arbitrary trouble-shooting tools without pre-install them in your production container image.

## Demo

## Quick Start

Install the docker debug
``` shell
go get -u github.com/zeromake/docker-debug
```
