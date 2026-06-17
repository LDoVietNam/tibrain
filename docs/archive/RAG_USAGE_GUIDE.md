---
title: "RAG System Usage Guide"
category: "infrastructure"
tier: "hot"
tags: ["rag", "tibrain", "documentation", "query", "ingestion"]
priority: "P0"
last_updated: "2026-05-22"
version: "1.0.0"
---

# RAG System Usage Guide

> **Version**: 1.0.0
> **Date**: 2026-05-22
> **Purpose**: Guide for using TiBrain RAG system with tag/scope filtering
> **Architecture**: Option 3 (current structure + frontmatter parsing)

---

## Overview

TiBrain RAG system enables semantic search across knowledge base with tag/scope-based filtering. Files stay in current structure, frontmatter contains metadata, and queries can filter by tags, scopes, or both.

For shared `OmniRoute` routing knowledge across Codex, Claude Code, Cursor, Cline, Open WebUI, and internal agents, see `docs/OMNIROUTE_MULTI_AGENT_GUIDE.md`.

---

## Architecture

```
Knowledge Files (Current Structure)
    ↓
Frontmatter Parser (Extract tags, scopes, metadata)
    ↓
Ingestion Pipeline (Parse + Embed + Store)
    ↓
SQLite Database (rag_documents with tags/scopes columns)
    ↓
Query Pipeline (Filter by tags/scopes + Search)
    ↓
Results (Filtered by metadata)
```

---

## Quick Start

### 1. Ingest Knowledge Base

```bash
cd apps/tibrain

# Ingest all files with validation
py scripts/ingest_kb.py --dir knowledge/ti-ecosystem-docs --validate-tags

# Ingest specific directory
py scripts/ingest_kb.py --dir knowledge/ti-ecosystem-docs/content/skills

# Force re-index all files
py scripts/ingest_kb.py --dir knowledge/ti-ecosystem-docs --force
```

### 2. Query Knowledge Base

```bash
# Basic query
py scripts/query_kb.py "authentication patterns"

# Filter by tags
py scripts/query_kb.py "provider" --tags authentication

# Filter by scopes
py scripts/query_kb.py "provider" --scopes auth

# Filter by both tags and scopes
py scripts/query_kb.py "testing" --tags testing --scopes code

# Filter by category
py scripts/query_kb.py "documentation" --category infrastructure

# Filter by tier
py scripts/query_kb.py "core rules" --tier hot

# Adjust result count
py scripts/query_kb.py "provider" --top-k 10
```

---

## Frontmatter Format

All markdown files must include frontmatter with tags and scopes:

```yaml
---
tags: ["tibrain", "provider", "documentation", "skill"]
scopes: ["tibrain"]
last_updated: "2026-05-22"
---
```

### Required Fields
- `tags`: Array of tags from TAGS.md taxonomy
- `scopes`: Array of scopes from SCOPES.md
- `last_updated`: YYYY-MM-DD format

### Optional Fields
- `title`: Document title
- `category`: Document category (core, quality, qa, pattern, safety, communication, infrastructure, domain, meta)
- `tier`: Document tier (hot, warm, cold)
- `priority`: Document priority (P0, P1, P2)
- `version`: Document version

---

## Tag/Scope Validation

### Valid Tags
- Core tags: authentication, authorization, rate-limiting, caching, translation, retry, monitoring, deployment, troubleshooting, security, performance, testing, documentation
- Provider tags: provider-antigravity, provider-openai, provider-claude, provider-gemini, provider-deepseek, provider-groq, provider-openrouter, provider-windsurf, provider-notion
- Technology tags: go, typescript, javascript, python, rust, java, mcp, oauth, oidc, jwt, sse, grpc, http, websocket, graphql, rest, sql, nosql, docker, kubernetes, terraform
- Component tags: router, cli, tibrain, ticrew, mcp-server, provider, plugin, skill, workflow, agent, dashboard, api, database, cache
- Pattern tags: pattern-auth-pkce, pattern-auth-device-code, pattern-auth-api-key, pattern-auth-cookie, pattern-retry-exponential, pattern-retry-circuit-breaker, pattern-cache-content, pattern-cache-session, pattern-translation-hub-spoke, pattern-translation-native-passthrough, pattern-monitoring-metrics, pattern-monitoring-logging, pattern-monitoring-tracing, pattern-deployment-blue-green, pattern-deployment-canary, pattern-testing-tdd, pattern-testing-bdd
- Domain tags: cli-tools, web-automation, browser-automation, code-generation, code-analysis, integration, messaging, storage, networking

### Valid Scopes
- auth: Authentication and authorization
- resilience: Resilience patterns (rate-limiting, retry, caching)
- integration: Integration patterns (translation, mcp, grpc, sse, websocket, graphql, rest)
- observability: Observability patterns (monitoring, logging, tracing)
- infrastructure: Infrastructure patterns (deployment, docker, kubernetes, terraform)
- providers: Provider-specific knowledge
- cli: CLI-related knowledge
- tibrain: TiBrain-related knowledge
- ticrew: Ticrew-related knowledge
- web: Web-related knowledge
- code: Code-related knowledge

---

## Query Examples

### Example 1: Find Authentication Patterns
```bash
py scripts/query_kb.py "authentication" --tags authentication
```

### Example 2: Find Provider Documentation
```bash
py scripts/query_kb.py "provider" --tags provider --scopes providers
```

### Example 3: Find Testing Documentation
```bash
py scripts/query_kb.py "testing" --tags testing --scopes code
```

### Example 4: Find Hot Tier Core Rules
```bash
py scripts/query_kb.py "core rules" --tier hot --category core
```

### Example 5: Find CLI Patterns
```bash
py scripts/query_kb.py "cli patterns" --tags cli --scopes cli
```

---

## Advanced Usage

### Multiple Tag Filters
```bash
# Files must have BOTH tags
py scripts/query_kb.py "documentation" --tags tibrain documentation
```

### Multiple Scope Filters
```bash
# Files must have BOTH scopes
py scripts/query_kb.py "resilience" --scopes resilience tibrain
```

### Combined Filters
```bash
# Files must have tags AND scopes
py scripts/query_kb.py "auth" --tags authentication --scopes auth
```

### Category + Tier
```bash
# Filter by category and tier
py scripts/query_kb.py "documentation" --category infrastructure --tier warm
```

---

## Database Schema

### rag_documents Table
```sql
CREATE TABLE rag_documents (
    id TEXT PRIMARY KEY,
    title TEXT NOT NULL,
    content TEXT NOT NULL,
    path TEXT NOT NULL,
    category TEXT NOT NULL DEFAULT 'general',
    tags TEXT,              -- JSON array of tags
    scopes TEXT,            -- JSON array of scopes
    tier TEXT DEFAULT 'warm',
    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL,
    status TEXT NOT NULL DEFAULT 'active',
    vector_id TEXT,
    metadata TEXT,
    content_hash TEXT,
    file_size INTEGER DEFAULT 0,
    last_indexed INTEGER,
    indexing_status TEXT DEFAULT 'pending'
);
```

### Indexes
- `idx_rag_documents_tags` - Fast tag filtering
- `idx_rag_documents_scopes` - Fast scope filtering
- `idx_rag_documents_category` - Fast category filtering
- `idx_rag_documents_tier` - Fast tier filtering
- `idx_rag_documents_updated` - Fast sorting by update time

---

## Performance

### Ingestion Performance
- Target: 1000 files/minute
- Current: ~6 files/second (360 files/minute)
- Bottleneck: Frontmatter parsing and embedding generation

### Query Performance
- Keyword search: < 100ms
- Tag/scope filtering: < 50ms
- Combined query: < 150ms
- Vector search: < 500ms (when embedding service available)

### Optimization Tips
1. Use specific tags/scopes to reduce result set
2. Use category/tier filters for broad categories
3. Limit results with `--top-k` parameter
4. Cache frequently used queries

---

## Troubleshooting

### Issue: No results found
**Cause**: Tags/scopes not matching database format
**Solution**: Check database format: `py -c "import sqlite3; conn = sqlite3.connect('tibrain.db'); cursor = conn.cursor(); cursor.execute('SELECT tags FROM rag_documents LIMIT 1'); print(cursor.fetchone())"`

### Issue: Invalid tag warnings
**Cause**: Tag not in TAGS.md taxonomy
**Solution**: Check TAGS.md or use `--validate-tags` flag

### Issue: Database locked
**Cause**: Multiple concurrent writes
**Solution**: Wait for ingestion to complete before querying

### Issue: Missing frontmatter
**Cause**: File doesn't have frontmatter
**Solution**: Add frontmatter with tags/scopes to file

---

## Maintenance

### Regular Ingestion
```bash
# Weekly re-index of updated files
py scripts/ingest_kb.py --dir knowledge/ti-ecosystem-docs --validate-tags
```

### Tag Taxonomy Updates
1. Update TAGS.md with new tags
2. Update frontmatter parser in `internal/frontmatter/parser.go`
3. Re-ingest knowledge base
4. Validate with `--validate-tags` flag

### Database Cleanup
```bash
# Remove inactive documents
py -c "import sqlite3; conn = sqlite3.connect('tibrain.db'); conn.execute('DELETE FROM rag_documents WHERE status != \"active\"'); conn.commit()"
```

---

## Integration with docs-lifecycle.md

### CREATE Phase
1. Add frontmatter to new document
2. Include tags and scopes
3. Run ingestion script
4. Validate tags against taxonomy

### READ Phase
1. Use query script with filters
2. Filter by tags/scopes for specific domains
3. Use category/tier for broad searches

### UPDATE Phase
1. Update frontmatter tags/scopes
2. Re-run ingestion script
3. Database will update automatically

### MAINTAIN Phase
1. Regular re-indexing
2. Tag taxonomy updates
3. Database cleanup

---

## Best Practices

### Tag Selection
- Use specific tags instead of generic ones
- Use 3-5 tags per file
- Include component tags (router, cli, tibrain)
- Include domain tags (auth, resilience, integration)

### Scope Selection
- Use 1-3 scopes per file
- Match scope to primary domain
- Use scopes for broad categorization

### Frontmatter Quality
- Always include tags and scopes
- Keep last_updated current
- Use valid categories and tiers
- Validate before ingestion

### Query Optimization
- Use specific filters for faster queries
- Combine tags and scopes for precision
- Use category/tier for broad searches
- Limit results with `--top-k`

---

## Obsidian Integration

### Overview
Ti Brain RAG system integrates with Obsidian via hybrid architecture:
- **MCP Server**: Real-time read/write access via `obsidian-mcp-server`
- **obsidian-headless**: Scheduled sync automation
- **Web Clipper**: External content capture

### Quick Start

#### 1. Setup Obsidian MCP Server
```bash
cd apps/tibrain/obsidian-mcp-server
bun install
cp .env.example .env
# Edit .env and set OBSIDIAN_API_KEY
bun run build
```

#### 2. Configure MCP Client
```json
{
  "mcpServers": {
    "obsidian": {
      "type": "stdio",
      "command": "bun",
      "args": ["apps/tibrain/obsidian-mcp-server/dist/index.js"],
      "env": {
        "MCP_TRANSPORT_TYPE": "stdio",
        "MCP_LOG_LEVEL": "info",
        "OBSIDIAN_API_KEY": "${OBSIDIAN_API_KEY}"
      }
    }
  }
}
```

#### 3. Sync Obsidian to Ti Brain
```bash
cd apps/tibrain

# Sync from Obsidian vault to Ti Brain
python scripts/sync_obsidian.py \
  --direction obsidian-to-tibrain \
  --vault /path/to/obsidian/vault \
  --headless-sync

# Continuous sync with obsidian-headless
python scripts/sync_obsidian.py \
  --direction obsidian-to-tibrain \
  --vault /path/to/obsidian/vault \
  --continuous
```

### Frontmatter Mapping

#### Obsidian → Ti Brain
| Obsidian Field | Ti Brain Field | Validation |
|----------------|----------------|------------|
| `tags` | `tags` | Validated against TAGS.md |
| `scopes` | `scopes` | Validated against SCOPES.md |
| `category` | `category` | Mapped to Ti Brain categories |
| `tier` | `tier` | Validated T1/T2/T3 |
| `priority` | `priority` | Validated P0/P1/P2/P3 |

#### Ti Brain → Obsidian
| Ti Brain Field | Obsidian Field | Notes |
|----------------|----------------|-------|
| `tags` | `tags` | Direct copy |
| `scopes` | `scopes` | Direct copy |
| `category` | `category` | Direct copy |
| `tier` | `tier` | Direct copy |
| `priority` | `priority` | Direct copy |

### Web Clipper Integration

#### Frontmatter Template
```yaml
---
title: "{{title}}"
url: "{{url}}"
tags: ["web-clip", "external"]
scopes: ["external"]
category: "external"
tier: "T3"
priority: "P2"
last_updated: "{{date}}"
source: "obsidian-clipper"
---
```

#### Auto-Ingest Clipped Content
```bash
# Watch clipped content folder
python scripts/sync_obsidian.py \
  --direction obsidian-to-tibrain \
  --vault /path/to/obsidian/vault \
  --pattern "Clipped/*.md"
```

### MCP Tools Available

| Tool | Purpose | Usage |
|------|---------|-------|
| `obsidian_get_note` | Read notes with frontmatter | Parse → RAG ingestion |
| `obsidian_search_notes` | Search vault | Query RAG system |
| `obsidian_write_note` | Create/update notes | Ti Brain → Obsidian |
| `obsidian_manage_frontmatter` | CRUD frontmatter | Sync tags/scopes |
| `obsidian_manage_tags` | Add/remove/list tags | Tag reconciliation |
| `obsidian_list_notes` | List vault files | Directory watching |

### obsidian-headless Commands

| Command | Purpose |
|---------|---------|
| `ob login` | Authenticate with Obsidian Sync |
| `ob sync-list-remote` | List remote vaults |
| `ob sync-setup` | Configure local-remote sync |
| `ob sync` | One-time sync |
| `ob sync --continuous` | Continuous sync with watcher |
| `ob sync-config` | Configure sync settings |
| `ob sync-status` | Show sync status |

### Configuration

#### MCP Server Environment Variables
```bash
OBSIDIAN_API_KEY=<local-rest-api-key>
OBSIDIAN_BASE_URL=http://127.0.0.1:27123
OBSIDIAN_VERIFY_SSL=false
OBSIDIAN_REQUEST_TIMEOUT_MS=30000
OBSIDIAN_ENABLE_COMMANDS=false
OBSIDIAN_READ_PATHS=          # Unset = full vault
OBSIDIAN_WRITE_PATHS=         # Unset = full vault
OBSIDIAN_READ_ONLY=false
```

#### Path Policy Examples
```bash
# Read everywhere, write only in specific folders
OBSIDIAN_WRITE_PATHS=projects/,scratch/

# Read only specific folder, write only in subfolder
OBSIDIAN_READ_PATHS=public/
OBSIDIAN_WRITE_PATHS=public/inbox/

# Read-only deployment
OBSIDIAN_READ_ONLY=true
```

### Troubleshooting

#### MCP Server Issues
- **Connection refused**: Check Obsidian is running with Local REST API enabled
- **Authentication failed**: Verify `OBSIDIAN_API_KEY` is correct
- **Path forbidden**: Check `OBSIDIAN_READ_PATHS` / `OBSIDIAN_WRITE_PATHS` configuration

#### obsidian-headless Issues
- **Login failed**: Verify Obsidian Sync credentials
- **Sync stalled**: Check network connection and vault size
- **Conflict errors**: Review conflict strategy configuration

#### Sync Script Issues
- **Frontmatter parsing failed**: Validate YAML syntax
- **Tag validation failed**: Check TAGS.md taxonomy
- **Vector embedding failed**: Verify embedding model availability

### References

- **Obsidian Integration Architecture**: `docs/OBSIDIAN_INTEGRATION_ARCHITECTURE.md`
- **obsidian-mcp-server**: `obsidian-mcp-server/README.md`
- **obsidian-headless**: `obsidian-headless/README.md`
- **RAG Architecture**: `docs/RAG_ARCHITECTURE.md`
- **Tag Taxonomy**: `Ti-learning-lab/04_Planning/TAGS.md`
- **Scope Definitions**: `Ti-learning-lab/04_Planning/SCOPES.md`
- **Tag Mapping**: `Ti-learning-lab/04_Planning/TAG_MAPPING.md`
- **Docs Lifecycle**: `content/rules/docs-lifecycle.md`

---

*Last Updated: 2026-05-22*
*Status: Production Ready*
