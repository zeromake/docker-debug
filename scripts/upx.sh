#!/bin/env sh
source ./scripts/variables.env

upx -q ${TARGET} -o ${TARGET}-upx

ln -sf "$(basename "${TARGET}-upx")" dist/docker-debug
