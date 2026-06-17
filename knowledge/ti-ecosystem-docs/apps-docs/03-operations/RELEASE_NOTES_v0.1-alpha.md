# Ti Complete Ecosystem Pack v0.1-alpha

Includes:

- `CLI/` with command-copilot aliases, `ti tui`, and `ti mcp web` wrapper.
- `router/` structure with provider routing examples.
- `mcp/web-tunnel` standalone MCP file server with UI and Quick Tunnel support.
- `core/` shared schema/contracts skeleton.
- `best_source/` curated default skills, agents, and workflows.
- `taskboard/`, `tibrain/`, `monitoring/`, `Ti-learning-lab/`, `tests/` and supporting folders.

Build locally:

```bash
make build-cli
make build-mcp
```

Note: build/test were not fully verified in the packaging environment because Go dependency downloads timed out.
