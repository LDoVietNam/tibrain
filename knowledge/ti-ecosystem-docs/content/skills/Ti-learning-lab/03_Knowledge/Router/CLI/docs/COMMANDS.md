# Command Map

| Command | Purpose |
| --- | --- |
| `ti-cli version` | Print build version, commit, and build date. |
| `ti-cli doctor` | Check Go/protoc/lint availability, config loading, and provider bootstrap. |
| `ti-cli ask` | Start an interactive provider-backed chat loop. |
| `ti-cli provider list` | Show registered providers and health. |
| `ti-cli provider status <name>` | Inspect one provider. |
| `ti-cli skill list` | List markdown skills from `skills_dir`. |
| `ti-cli skill show <name>` | Print one skill. |
| `ti-cli skill search <query>` | Search skill content. |
| `ti-cli chat <prompt>` | Route a prompt through the plugin system. |
| `ti-cli agent list/status/execute` | Manage agent plugins. |
| `ti-cli plugin list/load/unload/status/execute` | Manage low-level plugins. |
| `ti-cli learn ...` | Inspect BEADS learning stats, patterns, budget, and mode. |

Run `ti-cli <command> --help` for command-specific flags.

## Added in Go 1.23 provider/devkit pack

```bash
# Cookie auth state
ti-cli auth import-cookies sharedchat --file ./cookies.json
ti-cli auth cookies
ti-cli auth status sharedchat
ti-cli auth remove-cookies sharedchat
ti-cli auth path

# OpenAI-compatible / cookie HTTP providers
ti-cli provider list
ti-cli provider status local-openai
ti-cli provider status custom-cookie

# AI DevKit workflow templates
ti-cli devkit init --all
ti-cli devkit phase
ti-cli devkit phase show requirements
ti-cli devkit command
ti-cli devkit command show code-review
ti-cli devkit lint --json

# Prompt file runner
git diff | ti-cli prompt run code-review --provider openrouter
ti-cli prompt list
ti-cli prompt show execute-plan
ti-cli prompt run ./my-prompt.md "extra context" --provider local-openai
```

See also:

- `docs/COOKIE_PROVIDER.md`
- `docs/COMPATIBLE_PROVIDERS.md`
- `docs/AI_DEVKIT_INTEGRATION.md`

## Sandbox Runtime

```bash
ti-cli sandbox init
ti-cli sandbox status
ti-cli sandbox doctor
ti-cli sandbox policy
ti-cli sandbox run -- echo hello
ti-cli sandbox run --backend portable -- go test ./...
ti-cli sandbox run --backend bubblewrap -- go test ./...
ti-cli sandbox run --backend docker -- go test ./...
ti-cli sandbox logs
ti-cli sandbox logs --json --limit 50
```

Tool integration:

```bash
ti-cli tool run bash command="go test ./..." --sandbox
TI_TOOL_SANDBOX=1 ti-cli tool run bash command="go test ./..."
```

## Translate Engine

```bash
ti-cli translate scan [path]
ti-cli translate explain [path]
ti-cli translate convert [path] --dry-run
ti-cli translate convert [path] --out .ti/packs/imported
ti-cli translate validate .ti/packs/imported
```

Supported source formats in this build: Codex `AGENTS.md`, Claude Code `CLAUDE.md` and `.claude/commands|agents`, OpenCode commands/agents/skills, Cursor/Windsurf rules, and MCP server config. Executable hooks/plugins are converted with sandbox/risk metadata instead of being enabled directly.

## Blocks Runtime

```bash
ti-cli block path
ti-cli block list
ti-cli block show <id|last>
ti-cli block output <id|last>
ti-cli block rerun <id|last>
ti-cli block export <id|last> [file]
ti-cli block explain <id|last>
```

Auto-recorded commands:

```bash
ti-cli sandbox run -- echo hello
ti-cli tool run bash command="go test ./..." --sandbox
ti-cli translate convert . --dry-run
```

Disable recording with `--no-block` on supported commands.

## v0.6.0 Max Optimization Commands

### Provider Translate

```bash
ti-cli provider scan .
ti-cli provider import . --dry-run
ti-cli provider import ./opencode.json --out .ti/providers.generated.json
```

### Pack Manager

```bash
ti-cli pack list
ti-cli pack show .ti/packs/imported
ti-cli pack disable .ti/packs/imported
ti-cli pack enable .ti/packs/imported
```

### Blocks Search/Stats

```bash
ti-cli block search "permission denied"
ti-cli block stats
```

### Optimization Diagnostics

```bash
ti-cli optimize doctor
ti-cli optimize build-plan
```

### Shell Completion

```bash
ti-cli completion bash
ti-cli completion zsh
ti-cli completion fish
ti-cli completion powershell
```

## Token optimizer / proxy

```bash
ti-cli token estimate [file|-]
ti-cli token compress [file|-] --command "go test ./..."
ti-cli token run -- go test ./...
ti-cli token cache path
ti-cli token cache stats
ti-cli token cache clear
ti-cli proxy run -- git diff
```

## Auth token manager

```bash
ti-cli auth windsurf
ti-cli auth token import windsurf --provider windsurf --file ~/.ti/auth/windsurf.token
ti-cli auth token list
ti-cli auth token show windsurf
ti-cli auth token remove windsurf
ti-cli auth token path
```
