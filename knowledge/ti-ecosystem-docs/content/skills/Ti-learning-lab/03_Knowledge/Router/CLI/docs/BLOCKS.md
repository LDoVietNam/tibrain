# Ti Blocks Runtime

Ti Blocks records command, sandbox, tool, translate, and future agent runs as structured blocks instead of one long terminal stream. The goal is a Warp-like workflow layer for Ti without requiring Ti to become a full terminal emulator.

## Commands

```bash
ti-cli block path
ti-cli block list
ti-cli block show last
ti-cli block output last
ti-cli block explain last
ti-cli block export last block.md
ti-cli block rerun last
```

`last` resolves to the most recent block. You can also pass a full block ID or any unique prefix.

## Storage

By default blocks are stored at:

```text
~/.ti/data/blocks/
├── blocks.jsonl
└── artifacts/
    ├── blk-sandbox-*.out
    └── blk-tool-*.out
```

Override the store root with:

```bash
ti-cli block --store ./tmp/blocks list
```

## Auto-recorded operations

These commands record blocks by default:

```bash
ti-cli sandbox run -- go test ./...
ti-cli tool run bash command="go test ./..." --sandbox
ti-cli translate convert . --dry-run
```

Disable recording for a single run:

```bash
ti-cli sandbox run --no-block -- echo hello
ti-cli tool run bash command="echo hello" --no-block
ti-cli translate convert . --dry-run --no-block
```

## Rerun

Rerun sends the saved command through the sandbox runtime:

```bash
ti-cli block rerun last
ti-cli block rerun blk-sandbox-123 --backend portable --timeout 60
```

This intentionally prefers sandbox reruns over raw shell reruns.

## Explain

`ti-cli block explain` currently uses deterministic heuristics. It does not call an AI provider yet. Future versions can route block explanation through the selected provider/model.

## Design

Ti Blocks is meant to support:

- searchable command history by block
- safe reruns through sandbox policy
- exportable command/output artifacts
- TUI/web UI later
- session replay and audit trails
- AI explanations/fix suggestions on a specific block

It does not implement GPU rendering, terminal emulation, CRDT collaboration, or Warp-style shell DCS integration yet. Those belong to later shell/TUI phases.
