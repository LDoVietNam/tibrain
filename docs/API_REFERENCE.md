# Ti Brain API Reference

## Overview

Ti Brain provides multiple API interfaces for different use cases:

1. **Ti Brain REST API** - Main HTTP API for knowledge management
2. **Obsidian MCP Server API** - Model Context Protocol for AI agents
3. **Obsidian Local REST API** - Direct Obsidian vault access

---

## 1. Ti Brain REST API

**Base URL**: `http://localhost:1810`

### Health & Status

#### GET `/api/health`
Health check endpoint.

**Response**:
```json
{
  "status": "healthy",
  "timestamp": "2026-05-23T01:00:00Z",
  "service": "TiBrain",
  "version": "2.0.0",
  "port": 1810,
  "uptime": "active"
}
```

#### GET `/api/status`
Detailed system status.

**Response**:
```json
{
  "service": "TiBrain",
  "status": "active",
  "timestamp": "2026-07-01T03:11:57.5051414+07:00",
  "port": 1810,
  "knowledge_base": {
    "indexed": true,
    "knowledge_bases": 0,
    "total_documents": 0
  },
  "agents": {
    "active_agents": 1,
    "total_agents": 1,
    "agents": ["ti-brain"]
  },
  "tools_registered": 221,
  "orchestration": "ready"
}
```

### Knowledge Base

#### GET `/api/knowledge`
Get knowledge base statistics.

#### GET `/api/knowledge/index`
Get knowledge base index.

#### GET `/api/knowledge/status`
Get knowledge base status.

### Agent Management

#### GET `/api/agents`
List registered agents.

#### POST `/api/agents/register`
Register a new agent.

### Orchestration

#### POST `/api/orchestrate`
Orchestrate multi-agent workflows.

#### POST `/api/agent/request`
Send request to Ti Agent orchestrator.

#### GET `/api/agent/memory`
Get agent memory state.

### RAG (Retrieval-Augmented Generation)

#### POST `/api/rag/query`
Query the RAG system.

**Request**:
```json
{
  "query": "What is OAuth PKCE flow?",
  "top_k": 5,
  "filters": {
    "tags": ["authentication", "oauth"],
    "scopes": ["auth"]
  }
}
```

**Response**:
```json
{
  "results": [
    {
      "content": "PKCE is an extension to Authorization Code flow...",
      "metadata": {
        "file": "OAuth-PKCE-Flow.md",
        "tags": ["authentication", "oauth"],
        "score": 0.95
      }
    }
  ]
}
```

#### GET `/api/rag/status`
Get RAG system status.

#### POST `/api/rag/feedback`
Provide feedback on RAG results.

#### POST `/api/rag/ingest`
Ingest new documents into RAG system.

### Adaptive Retrieval And Promotion Boundary

The v2 retrieval flow implements the runtime pipeline:

`observe -> verify -> score -> distill -> promote -> execute`

Important behavior:

- The learning plane can inspect broad corpus evidence.
- The execution plane only consumes promoted patterns.
- Raw worker output is never written directly into runtime logic.

#### POST `/api/v2/retrieve`
Run adaptive retrieval with fast-path, optional slow-path enrichment, runtime scoring, and candidate distillation.

**Request**:
```json
{
  "query": "How should routing knowledge be verified before promotion?",
  "scopes": ["scope:local_cli"],
  "context_source": {
    "agent_id": "omniroute_cli",
    "port": 1806,
    "session_type": "chat"
  }
}
```

**Response**:
```json
{
  "id": "query-id",
  "query": "How should routing knowledge be verified before promotion?",
  "documents": ["doc-1"],
  "response": "Based on the available documentation...",
  "confidence": 0.61,
  "runtime": {
    "mode": "adaptive_slow_path",
    "cache_hit": false,
    "verified": true,
    "used_slow_path": true,
    "fast_candidates": 2,
    "slow_candidates": 4,
    "selected_documents": 1,
    "confidence_score": 0.52,
    "groundedness_score": 0.63,
    "coverage_score": 0.5,
    "quality_score": 0.57,
    "distilled_candidate_id": "candidate-id",
    "stage_durations_ms": {
      "fast_path": 8,
      "slow_path": 2,
      "assemble": 1,
      "answer": 3,
      "persist": 1
    }
  }
}
```

#### GET `/api/v2/retrieve/traces`
List recent runtime traces recorded by adaptive retrieval.

**Query params**:

- `limit` - optional, default `20`

#### GET `/api/v2/brain/patterns`
List pattern candidates produced by verified retrieval.

**Query params**:

- `status` - optional filter such as `candidate` or `promoted`
- `limit` - optional, default `20`

#### POST `/api/v2/brain/patterns/promote`
Promote a distilled pattern candidate into the execution-plane registry.

**Request**:
```json
{
  "candidate_id": "candidate-id"
}
```

**Response**:
```json
{
  "candidate_id": "candidate-id",
  "status": "promoted"
}
```

#### GET `/api/v2/runtime/registry`
Return the promoted runtime registry view. This endpoint intentionally excludes raw or unverified candidates.

### Cross-Brain Communication

#### POST `/api/brain/router`
Communicate with Router Brain.

#### POST `/api/brain/ti`
Communicate with Ti Brain.

### Tools

#### GET `/api/tools`
List available tools.

### Obsidian Integration

#### GET `/api/obsidian`
Get Obsidian vault integration status.

**Response**:
```json
{
  "status": "configured",
  "vault_path": null,
  "documents": 0,
  "note": "Set OBSIDIAN_VAULT_PATH env var to enable auto-ingestion",
  "timestamp": "2026-07-01T03:11:57.5051414+07:00"
}
```

### v1 API (Open-WebUI Integration)

#### POST `/api/v1/rag/query`
RAG query endpoint (v1).

#### GET `/api/v1/knowledge/documents`
List knowledge documents (v1).

#### POST `/api/v1/knowledge/ingest`
Ingest documents (v1).

#### GET `/api/v1/graph/nodes`
Get graph nodes (v1).

#### GET `/api/v1/router/routes`
Get router routes (v1).

#### POST `/api/v1/memory/store`
Store memory (v1).

#### GET `/api/v1/analytics`
Get analytics (v1).

---

## 2. Obsidian MCP Server API

**Protocol**: Model Context Protocol (MCP)
**Transport**: STDIO or HTTP
**Base URL**: `http://localhost:3000` (HTTP mode)

### MCP Tools (14 total)

#### `obsidian_get_note`
Get note content by path.

**Parameters**:
```json
{
  "path": "Folder/Note.md"
}
```

#### `obsidian_search_notes`
Search notes by query (supports Omnisearch if available).

**Parameters**:
```json
{
  "query": "OAuth PKCE",
  "context_length": 200
}
```

#### `obsidian_list_notes`
List notes in vault with filtering.

**Parameters**:
```json
{
  "folder": "Authentication",
  "recursive": true,
  "limit": 50
}
```

#### `obsidian_write_note`
Create or update a note.

**Parameters**:
```json
{
  "path": "NewNote.md",
  "content": "# New Note\n\nContent here..."
}
```

#### `obsidian_append_to_note`
Append content to a note.

**Parameters**:
```json
{
  "path": "ExistingNote.md",
  "content": "\n## New Section\n\nContent"
}
```

#### `obsidian_delete_note`
Delete a note.

**Parameters**:
```json
{
  "path": "OldNote.md"
}
```

#### `obsidian_manage_frontmatter`
Manage YAML frontmatter.

**Parameters**:
```json
{
  "path": "Note.md",
  "operation": "set",
  "key": "tags",
  "value": ["authentication", "oauth"]
}
```

#### `obsidian_list_tags`
List all tags in vault.

#### `obsidian_list_commands` (opt-in)
List Obsidian commands (requires OBSIDIAN_ENABLE_COMMANDS=true).

#### `obsidian_execute_command` (opt-in)
Execute Obsidian command (requires OBSIDIAN_ENABLE_COMMANDS=true).

### MCP Resources (3 total)

#### `obsidian://vault/{+path}`
Get note by vault path.

#### `obsidian://tags`
Get all tags in vault.

#### `obsidian://status`
Get MCP server status.

### Environment Variables

```bash
OBSIDIAN_API_KEY=your-api-key
OBSIDIAN_BASE_URL=http://127.0.0.1:27123
OBSIDIAN_VERIFY_SSL=false
OBSIDIAN_REQUEST_TIMEOUT_MS=30000
OBSIDIAN_ENABLE_COMMANDS=false
OBSIDIAN_READ_PATHS=
OBSIDIAN_WRITE_PATHS=
OBSIDIAN_READ_ONLY=false
```

---

## 3. Obsidian Local REST API

**Base URL**: `http://127.0.0.1:27123`
**Authentication**: Bearer token (API key)

### Endpoints

#### GET `/`
Root endpoint - plugin info.

#### GET `/vault/{path}`
Get note by vault path.

#### POST `/vault/{path}`
Create/update note.

#### DELETE `/vault/{path}`
Delete note.

#### GET `/search/`
Search notes (basic).

#### GET `/search/omnisearch`
Search notes (Omnisearch plugin required).

#### GET `/tags/`
List all tags.

#### GET `/commands/`
List available commands.

#### POST `/commands/{commandId}`
Execute command.

---

## 4. Python Sync Script API

**Script**: `sync_obsidian.py`
**Usage**: Command-line interface

### Commands

#### Obsidian to Ti Brain Sync
```bash
python sync_obsidian.py \
  --direction obsidian-to-tibrain \
  --vault /path/to/vault \
  --tibrain /path/to/tibrain \
  --pattern "*.md"
```

#### Ti Brain to Obsidian Sync
```bash
python sync_obsidian.py \
  --direction tibrain-to-obsidian \
  --vault /path/to/vault \
  --tibrain /path/to/tibrain \
  --pattern "*.md"
```

#### Continuous Sync with File Watching
```bash
python sync_obsidian.py \
  --direction obsidian-to-tibrain \
  --vault /path/to/vault \
  --tibrain /path/to/tibrain \
  --pattern "*.md" \
  --watch
```

#### With obsidian-headless Sync
```bash
python sync_obsidian.py \
  --direction obsidian-to-tibrain \
  --vault /path/to/vault \
  --tibrain /path/to/tibrain \
  --pattern "*.md" \
  --headless-sync
```

---

## 5. Go Frontmatter Mapper API

**Package**: `internal/frontmatter`
**Language**: Go

### Functions

#### `NewMapper(validTags, validScopes)`
Create a new frontmatter mapper.

#### `MapObsidianToTiBrain(obsidian Frontmatter)`
Map Obsidian frontmatter to Ti Brain format.

#### `MapTiBrainToObsidian(tibrain Frontmatter)`
Map Ti Brain frontmatter to Obsidian format.

#### `ValidateTiBrainFrontmatter(frontmatter Frontmatter)`
Validate Ti Brain frontmatter structure.

#### `MergeFrontmatter(base, override Frontmatter)`
Merge two frontmatter structures.

### Usage Example

```go
mapper := NewMapper(validTags, validScopes)
tibrainFm := mapper.MapObsidianToTiBrain(obsidianFm)
err := ValidateTiBrainFrontmatter(tibrainFm)
```

---

## API Usage Examples

### Example 1: Query Ti Brain Knowledge Base

```bash
curl -X POST http://localhost:1810/api/rag/query \
  -H "Content-Type: application/json" \
  -d '{
    "query": "OAuth PKCE implementation",
    "top_k": 5,
    "filters": {
      "tags": ["authentication", "oauth"],
      "scopes": ["auth"]
    }
  }'
```

### Example 2: Search Obsidian via MCP

```bash
# Start MCP server
cd obsidian-mcp-server
npm run start:stdio

# MCP client will call obsidian_search_notes tool
```

### Example 3: Sync Obsidian to Ti Brain

```bash
cd /path/to/tibrain
uv run --with pyyaml --with watchdog python sync_obsidian.py \
  --direction obsidian-to-tibrain \
  --vault ~/vaults/my-vault \
  --tibrain /path/to/tibrain \
  --pattern "*.md"
```

### Example 4: Get Obsidian Integration Status

```bash
curl http://localhost:1810/api/obsidian
```

### Example 5: Direct Obsidian API Access

```bash
curl -H "Authorization: Bearer YOUR_API_KEY" \
  http://127.0.0.1:27123/vault/Authentication/OAuth-PKCE-Flow.md
```

---

## Authentication

### Ti Brain REST API
Currently no authentication (development mode).

### Obsidian MCP Server
API key via `OBSIDIAN_API_KEY` environment variable.

### Obsidian Local REST API
Bearer token authentication:
```
Authorization: Bearer YOUR_API_KEY
```

---

## Rate Limiting

No rate limiting currently implemented (development mode).

---

## Error Responses

### Standard Error Format
```json
{
  "error": "Error message",
  "code": "ERROR_CODE",
  "timestamp": "2026-05-23T01:00:00Z"
}
```

### Common Error Codes
- `NOT_FOUND`: Resource not found
- `VALIDATION_ERROR`: Invalid input
- `SERVICE_UNAVAILABLE`: Service down
- `TIMEOUT`: Request timeout
- `UNAUTHORIZED`: Authentication failed

---

## SDK Examples

### Python SDK (Future)
```python
import tibrain

# Initialize client
client = tibrain.Client(base_url="http://localhost:1810")

# Query knowledge base
results = client.rag_query(
    query="OAuth PKCE",
    filters={"tags": ["authentication"]}
)

# Sync Obsidian
sync = tibrain.ObsidianSync(
    vault_path="~/vault",
    tibrain_path="~/tibrain"
)
sync.sync_directory("obsidian-to-tibrain", "*.md")
```

### Go SDK (Future)
```go
package main

import (
    "github.com/ti/tibrain/sdk"
)

func main() {
    client := sdk.NewClient("http://localhost:1810")
    
    results, err := client.RAGQuery(&sdk.QueryRequest{
        Query: "OAuth PKCE",
        Filters: map[string][]string{
            "tags": {"authentication", "oauth"},
        },
    })
}
```

---

## Monitoring & Debugging

### Health Checks
```bash
# Ti Brain API
curl http://localhost:1810/api/health

# Obsidian MCP Server
curl http://localhost:3000/api/status  # if HTTP mode
```

### Logging
- Ti Brain API: Console output
- MCP Server: Context-aware logging
- Sync Script: Python logging module

---

## API Versioning

- **Current Version**: v1.0
- **Compatibility**: Breaking changes will increment major version
- **Deprecation**: Deprecated endpoints will be announced 30 days before removal

---

## Support

For API issues:
- Check `docs/OBSIDIAN_DEPLOYMENT_GUIDE.md`
- Check `docs/OBSIDIAN_INTEGRATION_ARCHITECTURE.md`
- Review error logs and troubleshooting guides

---

**Last Updated**: 2026-07-01
**API Version**: 1.0.0
