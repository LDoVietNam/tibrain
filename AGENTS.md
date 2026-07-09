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