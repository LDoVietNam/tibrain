# Added Agent Runtime Commands

## Session

```bash
ti session path
ti session list [--limit 20] [--json]
ti session stats [--json]
ti session create [title] [--provider NAME] [--model MODEL] [--phase PHASE] [--project NAME]
ti session show <id>
ti session append <id> <system|user|assistant|tool> <content>
ti session fork <id> [title]
ti session archive <id>
ti session delete <id>
ti session export <id> [file]
ti session import <file>
```

## Tool

```bash
ti tool list [--json]
ti tool run <name> key=value...
ti tool run <name> '{"path":"README.md"}'
echo '{"path":"README.md"}' | ti tool run read_file --stdin
```

## Permission

```bash
ti permission path
ti permission init
ti permission list [--json]
ti permission mode ask|auto|yes|deny|plan|smart
ti permission allow <tool:pattern>
ti permission deny <tool:pattern>
ti permission ask <tool:pattern>
ti permission check <tool> [path-or-command]
```

## Project Context

```bash
ti context init
ti context scan [path]
ti context show [path]
```

## MCP

```bash
ti mcp tools
ti mcp serve --transport stdio
ti mcp serve --transport sse --port 8765
ti mcp serve --transport http --port 8765
```
