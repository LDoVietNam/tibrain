# Ti Sandbox Runtime

Ti Sandbox Runtime adds a safety layer for shell/tool/agent execution.

## What it provides

- `portable` backend: cross-platform guardrails using Go only.
- `bubblewrap` backend: stronger Linux process isolation when `bwrap` is installed.
- `docker` backend: optional container execution using `golang:1.23-bookworm` by default.
- Policy JSON with timeout, output cap, env allowlist, path denylist, and command denylist.
- JSONL audit log for every sandbox run.
- Tool integration: `ti tool run bash ... --sandbox` or `TI_TOOL_SANDBOX=1`.

The portable backend is not a kernel-level sandbox. It is designed as a safe default guardrail. For stronger isolation on Linux, install Bubblewrap and keep `backend: "auto"`.

## Commands

```bash
ti-cli sandbox init
ti-cli sandbox status
ti-cli sandbox doctor
ti-cli sandbox policy
ti-cli sandbox run -- go test ./...
ti-cli sandbox run --backend portable -- echo hello
ti-cli sandbox run --backend bubblewrap -- go test ./...
ti-cli sandbox run --backend docker -- go test ./...
ti-cli sandbox logs
ti-cli sandbox logs --json --limit 50
```

## Tool integration

```bash
ti-cli tool run bash command="go test ./..." --sandbox
TI_TOOL_SANDBOX=1 ti-cli tool run bash command="go test ./..."
```

## Default policy location

```text
~/.ti/sandbox/policy.json
```

You can override it:

```bash
ti-cli sandbox --policy ./configs/sandbox.example.json status
```

## Backend selection

With `backend: "auto"`, Ti selects:

1. `bubblewrap` on Linux when `bwrap` is installed.
2. `portable` otherwise.

You can force a backend per run:

```bash
ti-cli sandbox run --backend portable -- go test ./...
ti-cli sandbox run --backend docker -- go test ./...
```

## Recommended defaults

- Network: deny.
- Timeout: 120 seconds.
- Max output: 200 KB.
- Read/write scope: current workspace.
- Deny sensitive paths such as `.env`, `~/.ssh`, `~/.aws`, `~/.gnupg`.
- Log all runs to `~/.ti/data/sandbox/audit.jsonl`.

## Notes for Linux hardening

Install Bubblewrap for stronger sandboxing:

```bash
sudo apt-get install bubblewrap
```

Then verify:

```bash
ti-cli sandbox doctor
```
