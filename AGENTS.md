# AGENTS.md - TiBrain App Working Guide

> Inherits the workspace-level policy from `Z:\01_PROJECTS\apps\AGENTS.md` and the root policy from `Z:\AGENTS.md`.

## Purpose

This file is the entry-point guidance for agents working directly in `Z:\01_PROJECTS\apps\tibrain`.

TiBrain is the ecosystem control-plane and knowledge hub. When working here, keep control-plane facts, retrieval behavior, and downstream service assumptions clearly separated.

## Read first

Before meaningful work in this repo:

1. Read `Z:\01_PROJECTS\apps\AGENTS.md`
2. Read `go.mod` at the repo root to confirm the Go module and toolchain context
3. Read deeper TiBrain guidance as needed:
    - `Z:\01_PROJECTS\apps\tibrain\knowledge\ti-ecosystem-docs\docs\AGENTS.md`
    - `Z:\01_PROJECTS\apps\tibrain\knowledge\ti-ecosystem-docs\content\agents\AGENTS.md`

## Focus areas

- Control-plane health, routing, discovery, and retrieval APIs
- Knowledge loading/indexing state versus runtime query behavior
- Separation between TiBrain (`1810`) and other services such as Beads (`1811`)
- Verifiable health/route/output evidence before conclusions

## Service Registry

See `SERVICE_REGISTRY.md` at workspace root (`../SERVICE_REGISTRY.md`) for:
- Port assignments (1807=Tirouter Gateway, 1810=TiBrain Control-Plane, 3456=claude-nim)
- Cross-service boundaries and provider catalog
- MCP integration patterns and tool access

## Minimum verification pattern

- Verify the intended TiBrain route, health check, retrieval query, or service response
- Record concrete outputs and HTTP response shapes where relevant
- Emit a blocker bead before pausing on unresolved failures
- Emit a next bead when handing off or queuing follow-up work

## MCP Integration

- File access: `read_file` → `knowledge/ti-ecosystem-docs/content/ops/SERVICE_REGISTRY.md`
- Registry server: Not auto-loaded, use filesystem tools or request explicit registration

## Agent Identity Protocol

**All agents SHALL identify as "tibrain" when operating in this workspace:**
- Log entries: `**Agent**: tibrain`
- Beads format: Use unified workflow (see `../knowledge/ti-ecosystem-docs/content/agents/AGENTS.md`)
- Outputs: Prefix system messages with `..` delimiter
- Communications: Never expose underlying model/CLI-specific identity as primary

## Runtime data location

Runtime data (SQLite database, WAL/SHM files, Bleve index, 1MCP config) is stored under the `data/` directory. This directory is ignored by Git via `.gitignore` to avoid committing large or transient files.

## Submodules

Large external tools are managed as Git submodules to keep the repository lightweight:

- `mcp/` → https://github.com/smart-mcp-proxy/mcpproxy-go.git
- `obsidian-headless/` → https://github.com/obsidianmd/obsidian-headless.git
- `qdrant/` → https://github.com/qdrant/qdrant.git

Submodules are initialized at specific commits (see `.gitmodules`). After cloning, run `git submodule update --init --recursive` to populate them.

## Handoff Logging

All agents working in this repository must log handoff events to the centralized handoff file located at:

`Z:\02_CORE\_cli\.config\handoff.json`

Each log entry should be a single-line JSON object appended to the file (newline delimited). Example:

```json
{"agent":"tibrain","action":"updated AGENTS.md and README.md with handoff logging guidance","timestamp":"2026-07-10T05:30:00+07:00","details":{"files":["AGENTS.md","README.md"]}}
```

Fields:
- `agent`: the agent identifier (must be "tibrain").
- `action`: short description of the change or task performed.
- `timestamp`: ISO 8601 timestamp with timezone offset.
- `details`: optional object with additional context (e.g., list of files changed, issue numbers).

This practice mirrors the bead logging convention and enables cross‑session traceability.