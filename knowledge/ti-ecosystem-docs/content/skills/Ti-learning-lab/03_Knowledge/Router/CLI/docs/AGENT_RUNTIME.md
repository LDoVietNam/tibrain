# Ti Agent Runtime (Go 1.23)

This pack adds the first practical agent-runtime layer learned from Kilo Code style architecture while keeping Ti small and Go-native.

## New command groups

```bash
ti session list
ti session create "investigate bug" --provider openrouter --model gpt-4o-mini
ti session append <id> user "what changed?"
ti session show <id>
ti session fork <id> "try another plan"
ti session export <id> session.json

ti tool list
ti tool run read_file path=README.md
ti tool run grep pattern=TODO path=.

ti permission init
ti permission list
ti permission mode ask
ti permission check bash "rm -rf build"
ti permission allow read_file:*.go
ti permission deny '*:.env'

ti context init
ti context scan .
ti context show .

ti mcp tools
ti mcp serve --transport stdio
ti mcp serve --transport sse --port 8765
```

## Design

- `session`: persistent SQLite-backed conversation/task memory.
- `tool`: exposes the internal tool registry so tools are not hardcoded inside an agent loop.
- `permission`: local policy gate for shell/file/destructive actions.
- `context`: project-level AGENTS.md learning.
- `mcp`: exposes Ti capabilities to MCP clients over stdio or HTTP/SSE.

## Safety defaults

- Permission mode defaults to `ask`.
- Secrets are not printed by cookie/auth commands.
- `AGENTS.md` is local project guidance, not a global secret store.
- MCP uses local tools; expose HTTP/SSE only on trusted networks.

## Go 1.23 notes

This remains compatible with Go 1.23.x. The agent-runtime commands use existing dependencies already present in `go.mod` (`cobra`, `modernc.org/sqlite`, `uuid`) and mostly standard library code.
