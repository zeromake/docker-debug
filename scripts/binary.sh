#!/bin/env bash
#
# Build a static binary for the host OS/ARCH
#

OS=$1
source ./scripts/variables.env
if [[ -n $OS ]]; then
    source ./scripts/variables.${OS}.env
fi
echo "Building statically linked $VERSION $TARGET"
export CGO_ENABLED=0
go build -o "${TARGET}" --ldflags "${LDFLAGS}" "${SOURCE}"

# ln -sf "$(basename "${TARGET}")" dist/docker-debug
