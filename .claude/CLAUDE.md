# TiBrain — Working Agreement for Claude

## Project at a glance
- **Path**: `Z:\01_PROJECTS\tibrain` (NOT `Z:\10_WORKPLACE\Ti\apps\tibrain` — old location)
- **Port**: `1810` (not 1807 — that's Router)
- **Module**: `github.com/ti/router/tibrain`
- **Runtime model**: Learning plane (`observe -> verify -> score -> distill -> promote`) + Execution plane (`execute`)
- **Storage**: SQLite (primary) + Neo4j (graph) + Redis (cache) + Qdrant (vector, optional)

## Build & test
- **ALWAYS use `make`, never `go build` / `go test` directly**
- `make build` — optimized binary in `build/tibrain(.exe)`
- `make test` — full tests with race detector + coverage
- `make test-coverage-html` — open `coverage.html` in browser
- `make verify` — pre-commit gate (vet + lint + test-short)
- `make lint` — golangci-lint (REQUIRED before commit)
- `make fmt` — gofmt

## API surface (high-traffic endpoints)
- `GET  /api/health` — health check
- `GET  /api/status` — full status with KB/agent/RAG counts
- `POST /api/rag/query` — RAG query with intelligent routing
- `POST /api/v2/retrieve` — adaptive retrieval (v2)
- `GET  /api/v2/retrieve/traces` — runtime traces
- `GET  /api/v2/runtime/registry` — promoted patterns
- `GET  /api/agents` — agent registry
- `POST /api/agents/register` — register new agent
- `POST /api/knowledge/index` — trigger knowledge indexing

## Architecture conventions
- Top-level `.go` files still in `package main` (refactor in progress, see Tier 2)
- New code goes in `internal/<concern>/` subpackages
- API handlers stay in `api_server.go` until extracted to `internal/api/`
- Storage backends: `internal/storage/{sqlite,neo4j,redis,qdrant}/` when refactored

## Gotchas — read before touching
- **Vector search OFF by default** — set `TIBRAIN_ENABLE_VECTOR_SEARCH=true` to enable
- **LLM answers OFF by default** — set `TIBRAIN_ENABLE_LLM_ANSWERS=true` to enable
- **.env contains secrets** — never commit (already in .gitignore)
- **Reports** are written to `Z:\01_PROJECTS\tibrain\reports\` (NOT old `Z:\10_WORKPLACE\Ti\apps\tibrain\reports\`)
- **RTK subprocess** runs with `cmd.Dir = Z:\01_PROJECTS` (workspace root, not tibrain)
- **`AutoIngest` without a directory arg** scans `Z:\01_PROJECTS` (workspace)
- **Test coverage is ~5%** (only 2 test files for 42 source files) — add tests for any new logic
- **Both SQLite drivers required** (mattn + modernc) — pick one when refactoring

## Known incomplete features
- `KnowledgeBaseProcessor` — was placeholder, removed (signature is now `NewAPIServer(hub, im)`)
- Docs-mcp-server integration — `/api/docs/search` returns placeholder response