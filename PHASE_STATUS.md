# TiBrain — Phase Status Report

**Audit Date:** 2026-07-20
**Baseline Commit:** `0e8a143` (origin/main) → local main `0a842f6`
**Build Status:** ✅ SUCCESS (clean clone: `go build .` + `go vet ./...` pass)
**Module Path:** `github.com/ti/router/tibrain` · **Go:** 1.25.5 · **Port:** 1810

## Phase 0 — Reproducible and trustworthy baseline

| Task | Status | Notes |
|------|--------|-------|
| 1. Pin canonical repo state | ✅ DONE | `docs/architecture/repository-baseline.md` exists; owner `LDoVietNam/tibrain`, branch `main`, port `1810`. |
| 2. Clean clone build | ✅ DONE | `.gitignore` unignores `internal/`, `go.mod`, `go.sum`. Fresh clone builds + vets clean. |
| 3. Remove placeholder paths | ⚠️ PARTIAL | `executeLocalTool()` still returns placeholder; most endpoints call real impls. |
| 4. Normalize config | ✅ DONE | `config.yaml` port 1811→1810, no sample API keys; `config.example.yaml` added. |
| 5. Protocol conformance fixtures | ❌ TODO | Only static `protocol_conformance.go`; no `testdata/mcp/2025-11-25/*` or `conformance_test.go`. |

## Implemented (current tree)

| Area | Files |
|------|-------|
| MCP gateway | `internal/mcp/{server,mcp,hub_client,proxy_client,tool_registry,tools_filesystem}.go` — Streamable HTTP `/mcp` + legacy SSE `/mcp/sse` |
| Core / DB | `internal/core`, `internal/db` (db.go replaces old hub.go), `internal/crypto` |
| Tools / Policy | `internal/tools`, `internal/policy` (default-deny), `internal/memory` |
| Support pkgs | `internal/{api,async,cli,execution,knowledge,lazyrouter,notionprovider,ports,prediction,skills,storage,trace,verification}` |
| Top-level | `main.go`, `api_server.go`, `code_graph.go`, `integration.go`, `knowledge_indexer.go`, `mcp_hub_client.go`, `protocol_conformance.go`, `retrieval_router.go`, `ti_agent.go` |

## MCP endpoints

- `POST/GET/DELETE /mcp` — Streamable HTTP (mark3labs/mcp-go)
- `GET /mcp/sse` — legacy SSE bridge
- `GET /health`, `GET /ready`, `GET /status`

## Registered tools (real impl)

- `fs.read_file` (path-bounded to allowed roots)
- `brain.status`

> ⚠️ Other tools listed in `tool_registry.go` (`write_file`, `delete_file`, …) are **registered but have no handler** — implement before advertising.

## Known issues / next steps

1. **Task 3:** replace `executeLocalTool()` placeholder (main.go) with real local tool execution.
2. **Task 5:** add MCP 2025-11-25 conformance fixtures + tests.
3. **Phase 1+:** catalog pipeline, capability graph, full gateway, security/approval, domain packs — not started.
4. `go test ./...` fails on Windows due to toolchain `fork/exec` issue (env-level, not code); build + vet are the reliable gates.
