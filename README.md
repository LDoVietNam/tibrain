# TiBrain

Minimal startup guide for the standalone TiBrain service.

## Start

```powershell
cd Z:\01_PROJECTS\apps\tibrain
.\tibrain.exe --port 1810
```

If you want to run from source instead of the bundled binary:

```powershell
go run . --port 1810
```

> **Note**: Knowledge base currently empty. Run `python rag_api_server.py` hoặc sync Obsidian vault để khởi động RAG indexing.

## Verify

- Browser overview UI: `http://localhost:1810/overview`
- Health: `http://localhost:1810/api/health`
- Status: `http://localhost:1810/api/status`
- JSON overview: `http://localhost:1810/api/overview`
- Base URL for API clients: `http://localhost:1810`
- Adaptive retrieval: `POST /api/v2/retrieve`
- Runtime registry: `GET /api/v2/runtime/registry`
- Knowledge base: `GET /api/knowledge` (currently empty)

## Runtime Model

TiBrain now separates the brain into two planes:

- Learning plane: `observe -> verify -> score -> distill -> promote`
- Execution plane: `execute`

Key rules:

- Raw worker or substrate output may be ingested and analyzed, but it must not write directly into runtime logic.
- Only verified, scored, and distilled knowledge can be promoted into the runtime registry.
- Retrieval traces and promotion candidates are persisted so the promotion boundary is inspectable.

## Local Load-Shedding Defaults

For single-node local runtime, TiBrain now defaults to the lighter path:

- vector search is disabled unless `TIBRAIN_ENABLE_VECTOR_SEARCH=true`
- generative answer refinement is disabled unless `TIBRAIN_ENABLE_LLM_ANSWERS=true`
- embedding calls are skipped entirely when `rag_vector_index` is empty
- API handlers reuse one live `RAGSystemManager` instead of rebuilding it per request
- knowledge-base `doc_count` updates are batched during indexing instead of recalculated per file
- SQLite runs in a local-lean profile tuned for single-node service startup and low write overhead
- fresh databases bootstrap `tool_registry` at the latest schema and skip per-column startup migration scans

Optional tuning:

- `TIBRAIN_EMBEDDING_TIMEOUT_MS`
- `TIBRAIN_LLM_TIMEOUT_MS`

## Data
- Config: `Z:\01_PROJECTS\apps\tibrain\.env`
- DB: `Z:\01_PROJECTS\apps\tibrain\data\tibrain_data\tibrain.db`
- RAG Index: `Z:\01_PROJECTS\apps\tibrain\data\rag_index\`
- Reports: `Z:\01_PROJECTS\apps\tibrain\data\reports\`

## Reference

- Full API reference: `docs/API_REFERENCE.md`
- Data table catalog: `docs/DATA_TABLE_CATALOG.md`
- Brain/runtime architecture: `docs/BRAIN_RUNTIME_ARCHITECTURE.md`
- Archived design and implementation docs: `docs/archive/`
