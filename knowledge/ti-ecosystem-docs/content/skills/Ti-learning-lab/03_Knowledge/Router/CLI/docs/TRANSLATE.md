# Ti Translate Engine v0.7.0

Ti Translate converts AI coding tool formats into a Ti Pack (`ti.pack.yaml`) and can export a Ti-normalized project back to selected external formats.

## Commands

```bash
ti-cli translate formats

ti-cli translate scan .
ti-cli translate explain .
ti-cli translate conflicts .
ti-cli translate convert . --dry-run
ti-cli translate convert . --out .ti/packs/imported
ti-cli translate validate .ti/packs/imported

ti-cli translate convert . --to claude-code --out ./exported-claude
ti-cli translate export . --to opencode --out ./exported-opencode

ti-cli translate optimize .
```

## Supported source formats

| Source | Detected files | Converted objects |
|---|---|---|
| Codex/OpenAI | `AGENTS.md`, `AGENTS.override.md` | instructions |
| Claude Code | `CLAUDE.md`, `.claude/commands/*.md`, `.claude/agents/*.md`, `.claude/settings*.json`, `.claude/hooks/*.json` | instructions, commands, agents, hooks, permissions, risks |
| OpenCode | `.opencode/commands/*.md`, `.opencode/agents/*.md`, `.opencode/skills/*.md`, `.opencode/plugins/*.{ts,js,mjs}`, `opencode.json[c]` | commands, agents, skills, plugins, providers, hooks, permissions |
| Cursor | `.cursor/rules/*.{md,mdc}`, `.cursorrules` | scoped instructions |
| Windsurf | `.windsurf/rules/*.{md,mdc}`, `.windsurfrules` | scoped instructions |
| MCP | `mcp.json`, `.mcp.json` | MCP servers |
| Aider | `.aider.conf.yml`, `aider.conf.yaml`, `CONVENTIONS.md` | conservative workflows/instructions |
| Roo | `.roo/**/*.{md,json,yaml,yml}` | instructions, commands, agents |
| Devin/Kilo-like | files containing `devin`/`kilo` with JSON/YAML extensions | conservative workflows/providers |

## Target formats

```text
ti-pack      full Ti IR output
claude-code  CLAUDE.md + .claude/commands + .claude/agents
codex        AGENTS.md
opencode     .opencode/commands + .opencode/agents + .opencode/skills
```

## Safety model

The translator does not directly execute imported plugins or hooks.

Executable imports are marked with:

```yaml
sandbox: true
risks:
  - level: high
    action: review before enabling; run through Ti sandbox
```

Before enabling imported workflows/plugins, run:

```bash
ti-cli sandbox init
ti-cli translate explain .
ti-cli translate convert . --dry-run
```

## Ti IR objects

The generated `ti.pack.yaml` may contain:

```text
instructions
commands
agents
skills
workflows
hooks
plugins
mcpServers
providers
permissions
conflicts
risks
```

## Conflict detection

`ti translate conflicts` currently detects practical conflicts such as package manager disagreements and Go version disagreements across imported instruction files.

```bash
ti-cli translate conflicts . --strict
```

## Lockfile

Every Ti Pack also gets:

```text
compat.lock.json
```

It stores source path, format, kind, SHA-256, and import time so future imports can detect changed source files.
