# Ti CLI

> Go 1.23-ready optimized CLI pack.

## Quick start

```bash
go version
make check-go
make deps
make build-small
./bin/ti-cli version
./bin/ti-cli doctor
```

See [`docs/BUILD.md`](docs/BUILD.md) and [`docs/GO123_OPTIMIZATIONS.md`](docs/GO123_OPTIMIZATIONS.md).


---

# Ti CLI

Ti CLI is a Go command-line application built around a microkernel + plugin architecture for provider-backed AI chat, agent execution, skills, and BEADS learning utilities.

## What is included

- Cobra-based CLI entrypoint: `ti-cli`
- Provider registry for Ti Router, OpenAI, Anthropic, OpenRouter, and SharedChat-style auth
- Plugin/agent management commands
- Markdown skill discovery/search commands
- BEADS learning command tree
- Interactive `ask` loop
- `doctor` command for local environment checks
- Cross-platform build and release scripts

## Requirements

Required to build from source:

- Go version matching `go.mod` (`go 1.23` in this pack)
- Internet access once for `go mod download`, unless a `vendor/` directory is supplied

Optional:

- `protoc`, `protoc-gen-go`, `protoc-gen-go-grpc` for protobuf regeneration
- `golangci-lint` for linting
- `zip` for Unix package generation

## Quick start

```bash
cp configs/ti.example.json ti.json
mkdir -p ~/.ti/auth ~/.ti/data
cp .env.example ~/.ti/auth/.env
make deps
make build
./bin/ti-cli doctor
./bin/ti-cli version
```

Windows PowerShell:

```powershell
Copy-Item configs\ti.example.json ti.json
New-Item -ItemType Directory -Force ~/.ti/auth, ~/.ti/data
Copy-Item .env.example ~/.ti/auth/.env
.\build.ps1 deps
.\build.ps1 build
.\bin\ti-cli.exe doctor
.\bin\ti-cli.exe version
```

Set at least one backend before chatting:

```bash
export TI_ROUTER_URL=http://localhost:1806
# or
export OPENAI_API_KEY=...
export ANTHROPIC_API_KEY=...
export OPENROUTER_API_KEY=...
```

Then run:

```bash
./bin/ti-cli provider list
./bin/ti-cli ask
```

## Core commands

```bash
ti-cli version
ti-cli doctor
ti-cli ask
ti-cli provider list
ti-cli provider status openrouter
ti-cli skill list --skills-dir ./best_source/skills
ti-cli skill search router
ti-cli chat "explain this repo"
ti-cli agent status
ti-cli plugin list
ti-cli learn stats
```

See `docs/COMMANDS.md` for the command map.

## Build targets

Unix/macOS:

```bash
make help
make deps
make build
make test
make smoke
make build-all
make package
make release VERSION=2.0.0
```

Windows:

```powershell
.\build.ps1 help
.\build.ps1 deps
.\build.ps1 build
.\build.ps1 test
.\build.ps1 smoke
.\build.ps1 build-all
.\build.ps1 package
.\build.ps1 release -Version 2.0.0
```

Release outputs are written to `dist/`.

## Configuration

Config discovery order:

1. Defaults
2. `~/.config/ti/config.json`
3. `./ti.json`
4. `./.ti/config.json`
5. `TI_CONFIG`
6. Environment overrides

Use `configs/ti.example.json` as a starter. Keep secrets in `~/.ti/auth/.env` or environment variables, not in committed config.

## Packaging notes

The optimized pack intentionally treats `bin/` and `dist/` as generated artifacts. Build them locally with your target Go toolchain so version metadata and platform targets are correct.

See `PACK_STATUS.md` for exactly what was changed in this revision and what still must be provided by the build environment.

## Go 1.23 provider/devkit additions

This build also includes the provider and workflow patterns from the uploaded `chatgpt-cli` and `ai-devkit` packs:

- `compatible_providers` for OpenAI-compatible gateways, local routers, and custom auth headers.
- `cookie_http` compatible provider for authorized internal/self-hosted services that use browser-exported cookies.
- `auth import-cookies` / `auth cookies` / `auth status` commands with redacted output.
- `devkit` commands for AI DevKit phase templates and readiness linting.
- `prompt` commands to list/show/run reusable prompt files through any configured provider.

Examples:

```bash
ti-cli auth import-cookies sharedchat --file ./cookies.json
ti-cli auth status sharedchat
ti-cli provider list
ti-cli devkit init --all
git diff | ti-cli prompt run code-review --provider openrouter
```

Read:

- `docs/COOKIE_PROVIDER.md`
- `docs/COMPATIBLE_PROVIDERS.md`
- `docs/AI_DEVKIT_INTEGRATION.md`
