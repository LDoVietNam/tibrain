# Build Ti CLI with Go 1.23

## Requirements

- Go 1.23.x
- Internet access for the first `go mod download`
- Optional: `protoc`, `protoc-gen-go`, `protoc-gen-go-grpc`, `golangci-lint`, `zip`

## Linux/macOS

```bash
go version
make check-go
cp configs/ti.example.json ti.json
cp .env.example ~/.ti/auth/.env
make deps
make build-small
./bin/ti-cli version
./bin/ti-cli doctor
```

## Windows PowerShell

```powershell
go version
Copy-Item configs\ti.example.json ti.json
New-Item -ItemType Directory -Force ~/.ti/auth | Out-Null
Copy-Item .env.example ~/.ti/auth/.env
.\build.ps1 deps
.\build.ps1 build-small
.\bin\ti-cli.exe version
.\bin\ti-cli.exe doctor
```

## Useful targets

```bash
make build       # normal local build
make build-small # portable stripped binary
make build-pgo   # use default.pgo when present
make test-fast
make release
```

## Docker

```bash
docker build --build-arg GO_VERSION=1.23 -t ti-cli:go123 .
docker run --rm ti-cli:go123 version
```
