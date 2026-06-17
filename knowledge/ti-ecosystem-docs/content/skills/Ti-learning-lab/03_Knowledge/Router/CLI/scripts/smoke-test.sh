#!/usr/bin/env sh
set -eu

bin="${1:-./bin/ti-cli}"
"$bin" version
"$bin" --help >/dev/null
"$bin" doctor || true
