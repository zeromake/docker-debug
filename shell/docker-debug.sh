#!/bin/bash

set -e

inspectInfo=$(docker inspect "$1")
containerId=$(echo "$inspectInfo" | grep "Id" | cut -d'"' -f 4)
mergedDir=$(echo "$inspectInfo" | grep "MergedDir" | cut -d'"' -f 4)
targetName="container:$containerId"
name=$(docker run --network $targetName --pid $targetName --stop-signal SIGKILL -v $mergedDir:/mnt/container --rm -it -d nicolaka/netshoot:latest /usr/bin/env sh)
args="${@:2}"
docker exec -it "$name" $args
docker stop "$name" > /dev/null &
