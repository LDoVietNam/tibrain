# Ti CLI Go 1.23 Optimizations

This pack is prepared for Go 1.23.x instead of Go 1.26.

## What changed

- `go.mod` is pinned to `go 1.23`.
- `.go-version` is set to `1.23.2`.
- Build scripts use `GOTOOLCHAIN=local` by default to avoid surprise downloads of a newer toolchain.
- Release builds use `-trimpath`, `-buildvcs=false`, stripped symbols, and an empty build ID for more reproducible artifacts.
- `build-small` and release cross-builds use `CGO_ENABLED=0` plus `netgo,osusergo` tags for more portable binaries.
- `build-pgo` supports `default.pgo` when you have a CPU profile and falls back to `-pgo=off`.
- `version` now prints the Go runtime and target platform.
- `doctor` now prints the Go runtime and `GOTOOLCHAIN` so build mismatches are visible quickly.
- CI is renamed to `ci-go123` and tests/builds only against Go 1.23.x.

## First build

```bash
go version
make check-go
make deps
make build-small
./bin/ti-cli version
./bin/ti-cli doctor
```

Windows:

```powershell
go version
.\build.ps1 deps
.\build.ps1 build-small
.\bin\ti-cli.exe version
.\bin\ti-cli.exe doctor
```

## PGO flow

1. Generate a useful profile from realistic CLI usage/tests.
2. Save it as `default.pgo` at repo root.
3. Build:

```bash
make build-pgo
```

Without `default.pgo`, the target builds with `-pgo=off`.

## Dependency note

This pack intentionally avoids Go 1.26-only module metadata. If you previously built with Go 1.26, run:

```bash
go clean -modcache
GOTOOLCHAIN=local go mod tidy
```

Then commit the refreshed `go.sum`.
