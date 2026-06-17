#!/usr/bin/env sh
set -eu

export GOTOOLCHAIN="${GOTOOLCHAIN:-local}"
export GOWORK="${GOWORK:-off}"
export GO_VERSION="${GO_VERSION:-1.23}"
make check-go
make release
