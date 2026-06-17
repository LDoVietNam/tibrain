#!/usr/bin/env sh
set -eu

VERSION="${VERSION:-dev}"
COMMIT="${COMMIT:-$(git rev-parse --short HEAD 2>/dev/null || echo none)}"
DATE="${DATE:-$(date -u +%Y-%m-%dT%H:%M:%SZ)}"
OUT="${OUT:-bin/ti-cli}"

mkdir -p "$(dirname "$OUT")"
GOTOOLCHAIN=local GOWORK=off CGO_ENABLED=0 go build \
  -trimpath \
  -buildvcs=false \
  -tags "netgo,osusergo" \
  -ldflags "-s -w -buildid= -X github.com/ti/cli/cmd.version=$VERSION -X github.com/ti/cli/cmd.commit=$COMMIT -X github.com/ti/cli/cmd.date=$DATE" \
  -o "$OUT" ./cmd/ti

echo "built $OUT"
