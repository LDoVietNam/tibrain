#!/usr/bin/env sh
set -eu

want="${GO_VERSION:-1.23}"
gov="$(go env GOVERSION 2>/dev/null || go version | awk '{print $3}')"
case "$gov" in
  go"$want"*) echo "Go OK: $gov" ;;
  *)
    echo "Expected Go ${want}.x, got ${gov}" >&2
    exit 1
    ;;
esac
