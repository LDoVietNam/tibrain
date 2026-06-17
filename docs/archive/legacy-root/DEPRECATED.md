# TiBrain App (Port 1810) - DEPRECATED

> **Status**: DEPRECATED / DISABLED
> **Date**: 2026-05-17
> **Replacement**: Unified Brain Server (Port 1808) in `apps/router/brain-server/`

---

## Why Deprecated?

The Ti ecosystem now uses a **Unified Brain Server** on port 1808 (part of Ti Router) instead of the standalone TiBrain app on port 1810.

This provides:
- **Single source of truth**: All intelligence endpoints in one server
- **Unified storage**: Memory Palace database (SQLite drawers)
- **Simplified operations**: One server instead of two
- **Consistent API**: All endpoints under one port

---

## Migration Status

### ✅ Migrated Components (34 endpoints)

| Component | Endpoints | Status |
|-----------|-----------|--------|
| RAG | /rag/query, /rag/search, /rag/analytics | ✅ |
| CLI Context | /cli/context, /cli/cache, /cli/help | ✅ |
| CLI Registry | /v1/tibrain/cli/* (4 endpoints) | ✅ |
| Handoff Tracking | /v1/tibrain/handoff/* (3 endpoints) | ✅ |
| Cognitive Memory | /v1/tibrain/memory/* (4 endpoints) | ✅ |
| MCP Registry | /v1/tibrain/mcp/* (4 endpoints) | ✅ |
| MCP Hub | /v1/tibrain/mcp-hub/* (4 endpoints) | ✅ |
| Tool Registry | /v1/tibrain/tool/* (5 endpoints) | ✅ |
| Agent Processing | /v1/tibrain/agent/process | ✅ |
| BEADS Stats | /v1/tibrain/beads/stats/* (5 endpoints) | ✅ |

### ⏭️ Not Migrated (Complex/External Dependencies)

| Component | Reason |
|-----------|--------|
| Neo4j Graph Store | Requires external Neo4j instance |
| Knowledge Indexing | Complex initialization, can be run manually |
| Ti Agent Orchestrator | Complex orchestration logic |

---

## How to Use the New Brain Server

### Start the Brain Server
```bash
cd apps/router
go run brain-server/main.go
```

### Access Endpoints
- **Health**: http://localhost:1808/health
- **RAG**: http://localhost:1808/rag/query
- **CLI Registry**: http://localhost:1808/v1/tibrain/cli/list
- **Memory**: http://localhost:1808/v1/tibrain/memory/stats

---

## Porting Code

If you have code referencing port 1810, update it to port 1808:

```go
// Old
client := tibrain.NewEnhancedTiBrainClient("http://localhost:1810")

// New
client := tibrain.NewEnhancedTiBrainClient("http://localhost:1808")
```

---

## Keeping this Directory

This directory is kept for reference and historical purposes. The code here is no longer actively maintained.

For new development, use `apps/router/brain-server/`.
