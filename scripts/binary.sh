#!/usr/bin/env bash
#
# Build a static binary for the host OS/ARCH
#


source ./scripts/variables.env

echo "Building statically linked $TARGET"
export CGO_ENABLED=0
go build -o "${TARGET}" --ldflags "${LDFLAGS}" "${SOURCE}"

ln -sf "$(basename "${TARGET}")" dist/docker-debug
