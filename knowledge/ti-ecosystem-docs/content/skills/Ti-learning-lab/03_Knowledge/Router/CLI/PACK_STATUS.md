# Pack Status

This pack is optimized for Go 1.23 and a translate-first Ti architecture.

## New in v0.8.0

- `ti-cli token estimate/compress/run/cache`
- `ti-cli proxy run` alias for token-saving command proxy usage
- `ti-cli auth windsurf`
- `ti-cli auth token import/list/show/remove/path`
- Docs for auth tokens and token optimization
- Config sample for token optimizer

## Build note

Run on a machine with Go 1.23 and network access for modules:

```bash
GOTOOLCHAIN=local GOWORK=off go mod tidy
GOTOOLCHAIN=local GOWORK=off go test ./internal/translate ./internal/tokenopt ./internal/authtoken ./internal/blocks ./internal/sandbox
VERSION=v0.8.0 OUT=bin/ti-cli ./scripts/max-release.sh
```
