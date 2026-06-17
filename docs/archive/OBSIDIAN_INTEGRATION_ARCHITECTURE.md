# Obsidian Integration Technical Architecture

**Version:** 2.0.0  
**Last Updated:** 2026-05-22  
**Status:** Technical Design  
**Audience:** Developers, Architects, Technical Leads

---

## Overview

Technical architecture for integrating Obsidian with Ti Brain RAG system using hybrid approach:
- **Layer 1:** Obsidian MCP Server (real-time access via MCP protocol)
- **Layer 2:** obsidian-headless (scheduled sync automation)
- **Layer 3:** Web Clipper (external content capture)

**For executive summary, see:** `OBSIDIAN_SYNC_ARCHITECTURE.md`  
**For implementation plan, see:** `OBSIDIAN_INTEGRATION_MASTER_PLAN.md`

---

## System Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                         User Layer                            │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐          │
│  │ Obsidian     │  │ Obsidian     │  │ Web Browser  │          │
│  │ Desktop      │  │ Mobile       │  │ (Web Clipper)│          │
│  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘          │
└─────────┼──────────────────┼──────────────────┼─────────────────┘
          │                  │                  │
          └──────────────────┼──────────────────┘
                             │
┌────────────────────────────┼────────────────────────────────────┐
│                    Integration Layer                            │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │              obsidian-mcp-server (MCP)                    │  │
│  │  - 14 Tools (read, write, search, edit)                  │  │
│  │  - 3 Resources (vault, tags, status)                      │  │
│  │  - Path Policy (read/write restrictions)                  │  │
│  │  - Authentication (API key)                              │  │
│  └───────────────────────────┬──────────────────────────────┘  │
│                              │                                  │
│  ┌───────────────────────────┴──────────────────────────────┐  │
│  │              obsidian-headless (Sync)                     │  │
│  │  - Obsidian Sync client                                 │  │
│  │  - Continuous sync with watcher                         │  │
│  │  - Conflict resolution                                   │  │
│  │  - Configuration management                              │  │
│  └───────────────────────────┬──────────────────────────────┘  │
│                              │                                  │
│  ┌───────────────────────────┴──────────────────────────────┐  │
│  │              sync_obsidian.py (Script)                   │  │
│  │  - Frontmatter mapping                                  │  │
│  │  - Tag/scope validation                                 │  │
│  │  - Directory watching                                   │  │
│  │  - Auto-ingest trigger                                   │  │
│  └───────────────────────────┬──────────────────────────────┘  │
└──────────────────────────────┼──────────────────────────────────┘
                               │
┌──────────────────────────────┼──────────────────────────────────┐
│                    Processing Layer                            │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │              Frontmatter Parser (Go)                      │  │
│  │  - YAML parsing                                          │  │
│  │  - Tag validation                                        │  │
│  │  - Scope validation                                      │  │
│  │  - Frontmatter mapping                                   │  │
│  └───────────────────────────┬──────────────────────────────┘  │
│                              │                                  │
│  ┌───────────────────────────┴──────────────────────────────┐  │
│  │              Ingestion Pipeline (Python)                  │  │
│  │  - File discovery                                       │  │
│  │  - Frontmatter extraction                               │  │
│  │  - Vector embedding                                      │  │
│  │  - Database storage                                      │  │
│  └───────────────────────────┬──────────────────────────────┘  │
│                              │                                  │
│  ┌───────────────────────────┴──────────────────────────────┐  │
│  │              Query Pipeline (Python)                      │  │
│  │  - Keyword search                                       │  │
│  │  - Tag/scope filtering                                  │  │
│  │  - Vector search                                        │  │
│  │  - Result ranking                                       │  │
│  └───────────────────────────┬──────────────────────────────┘  │
└──────────────────────────────┼──────────────────────────────────┘
                               │
┌──────────────────────────────┼──────────────────────────────────┐
│                    Storage Layer                               │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │              SQLite Database (RAG)                        │  │
│  │  - rag_documents table                                   │  │
│  │  - Vector embeddings                                     │  │
│  │  - Metadata (tags, scopes, category, tier)               │  │
│  │  - Full-text search index                                │  │
│  └───────────────────────────┬──────────────────────────────┘  │
│                              │                                  │
│  ┌───────────────────────────┴──────────────────────────────┐  │
│  │              Obsidian Vault (File System)                 │  │
│  │  - Markdown files                                        │  │
│  │  - Frontmatter (YAML)                                    │  │
│  │  - Media files (images, audio, video)                     │  │
│  │  - Plugin data (.obsidian/)                              │  │
│  └──────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────┘
```

---

## Component Specifications

### 1. obsidian-mcp-server

**Purpose:** Real-time MCP access to Obsidian vault

**Technology Stack:**
- Language: TypeScript
- Framework: @cyanheads/mcp-ts-core v0.9.1
- Runtime: Bun v1.3.11+ or Node v24+
- Transport: STDIO, HTTP

**Key Features:**
- 14 MCP tools (read, write, search, edit)
- 3 MCP resources (vault, tags, status)
- Path policy enforcement
- Authentication via API key
- Frontmatter support
- Tag reconciliation

**API Configuration:**
```bash
OBSIDIAN_API_KEY=<api-key>
OBSIDIAN_BASE_URL=http://127.0.0.1:27123
OBSIDIAN_VERIFY_SSL=false
OBSIDIAN_REQUEST_TIMEOUT_MS=30000
OBSIDIAN_ENABLE_COMMANDS=false
OBSIDIAN_READ_PATHS=
OBSIDIAN_WRITE_PATHS=
OBSIDIAN_READ_ONLY=false
```

**MCP Tools:**
| Tool | Input | Output | Use Case |
|------|-------|--------|----------|
| `obsidian_get_note` | path, format | content, frontmatter | Read notes for RAG ingestion |
| `obsidian_search_notes` | query, mode, filters | results with pagination | Search vault content |
| `obsidian_write_note` | path, content, section | created, previousSize, currentSize | Write Ti Brain content to Obsidian |
| `obsidian_manage_frontmatter` | path, operation, key, value | updated frontmatter | Sync tags/scopes |
| `obsidian_manage_tags` | path, operation, tags, location | updated tags | Tag reconciliation |
| `obsidian_list_notes` | path, depth, filters | file list | Directory watching |
| `obsidian_list_tags` | nameRegex | tags with counts | Tag validation |

**SLA:**
- Uptime: 99.9%
- Response time: < 500ms (tools), < 200ms (resources)
- Throughput: 100 requests/second

### 2. obsidian-headless

**Purpose:** Scheduled sync automation

**Technology Stack:**
- Language: JavaScript (Node.js)
- Runtime: Node v22+
- Package: obsidian-headless v0.0.9

**Key Features:**
- Obsidian Sync authentication
- Remote vault listing
- Local-remote sync setup
- One-time sync
- Continuous sync with watcher
- Sync configuration management

**Command API:**
```bash
ob login [--email] [--password] [--mfa]
ob logout
ob sync-list-remote
ob sync-list-local
ob sync-create-remote --name <name> [--encryption] [--password]
ob sync-setup --vault <id> [--path] [--password] [--device-name]
ob sync [--path] [--continuous]
ob sync-config [--path] [options]
ob sync-status [--path]
ob sync-unlink [--path]
```

**Sync Modes:**
- `bidirectional`: Full two-way sync (default)
- `pull-only`: Only download, ignore local changes
- `mirror-remote`: Only download, revert local changes

**SLA:**
- Sync success rate: 99.9%
- Sync latency: < 10s (continuous)
- Configuration validation: 100%

### 3. sync_obsidian.py

**Purpose:** Sync orchestration and frontmatter mapping

**Technology Stack:**
- Language: Python 3.11+
- Dependencies: PyYAML, watchdog

**Key Features:**
- Frontmatter mapping (Obsidian ↔ Ti Brain)
- Tag/scope validation
- Directory watching
- Auto-ingest trigger
- Conflict resolution
- Batch operations

**API:**
```python
class ObsidianSync:
    def __init__(self, vault_path: str, tibrain_path: str)
    def sync_obsidian_to_tibrain(self, file_path: Path) -> bool
    def sync_tibrain_to_obsidian(self, content: str, frontmatter: Dict, target_path: Path) -> bool
    def sync_directory(self, direction: str, pattern: str) -> Dict
    def run_obsidian_headless_sync(self, continuous: bool = False) -> bool
```

**CLI:**
```bash
python sync_obsidian.py \
  --direction obsidian-to-tibrain \
  --vault /path/to/vault \
  --tibrain /path/to/tibrain \
  --pattern "*.md" \
  --headless-sync \
  --continuous
```

**SLA:**
- Sync latency: < 5s (single file)
- Validation accuracy: 95%
- Auto-ingest success rate: 99%

### 4. Frontmatter Parser (Go)

**Purpose:** Parse and validate frontmatter

**Technology Stack:**
- Language: Go 1.20+
- Dependencies: gopkg.in/yaml.v3

**Key Features:**
- YAML parsing
- Tag validation against TAGS.md
- Scope validation against SCOPES.md
- Frontmatter mapping
- Error reporting

**API:**
```go
type Parser struct {
    validTags   map[string]bool
    validScopes map[string]bool
}

func (p *Parser) ParseFrontmatter(content string) (*Frontmatter, error)
func (p *Parser) ValidateTags(tags []string) error
func (p *Parser) ValidateScopes(scopes []string) error
func (p *Parser) ExtractContent(content string) string

type Mapper struct {
    validTags   map[string]bool
    validScopes map[string]bool
}

func (m *Mapper) MapObsidianToTiBrain(obsidian ObsidianFrontmatter) (Frontmatter, error)
func (m *Mapper) MapTiBrainToObsidian(tibrain Frontmatter) (ObsidianFrontmatter, error)
func ValidateTiBrainFrontmatter(frontmatter Frontmatter) error
func MergeFrontmatter(base, override Frontmatter) Frontmatter
```

**SLA:**
- Parse time: < 100ms per file
- Validation accuracy: 95%
- Mapping accuracy: 100%

### 5. Ingestion Pipeline (Python)

**Purpose:** Ingest files into RAG database

**Technology Stack:**
- Language: Python 3.11+
- Dependencies: SQLite, sentence-transformers

**Key Features:**
- File discovery
- Frontmatter extraction
- Vector embedding
- Database storage
- Incremental updates

**API:**
```python
def ingest_directory(directory: str, validate_tags: bool = True, force: bool = False)
def ingest_file(file_path: str, validate_tags: bool = True)
def extract_frontmatter(content: str) -> Dict
def validate_tags(tags: List[str]) -> List[str]
def generate_embedding(content: str) -> np.ndarray
def store_document(document: Dict) -> bool
```

**SLA:**
- Ingestion rate: 1000 files/minute
- Embedding time: < 500ms per file
- Storage success rate: 99.9%

### 6. Query Pipeline (Python)

**Purpose:** Query RAG database with filters

**Technology Stack:**
- Language: Python 3.11+
- Dependencies: SQLite, sentence-transformers

**Key Features:**
- Keyword search
- Tag/scope filtering
- Vector search
- Result ranking
- Category/tier filtering

**API:**
```python
def query_database(query: str, tags: List[str] = None, scopes: List[str] = None, 
                  category: str = None, tier: str = None, top_k: int = 5) -> List[Dict]
def keyword_search(query: str, limit: int = 10) -> List[Dict]
def filter_by_tags(results: List[Dict], tags: List[str]) -> List[Dict]
def filter_by_scopes(results: List[Dict], scopes: List[str]) -> List[Dict]
def vector_search(query: str, top_k: int = 5) -> List[Dict]
def rank_results(results: List[Dict]) -> List[Dict]
```

**SLA:**
- Query latency: < 500ms
- Result relevance: > 80%
- Filter accuracy: 100%

---

## Data Model

### Frontmatter Schema

#### Ti Brain Frontmatter
```yaml
---
title: string                    # Document title
tags: string[]                   # Tags from TAGS.md
scopes: string[]                 # Scopes from SCOPES.md
category: string                 # Category (core, quality, qa, pattern, safety, communication, infrastructure, domain, meta)
tier: string                     # Tier (T1, T2, T3)
priority: string                 # Priority (P0, P1, P2, P3)
last_updated: string             # ISO 8601 timestamp
version: string?                 # Optional version
---
```

#### Obsidian Frontmatter (Extended)
```yaml
---
title: string
tags: string[]
scopes: string[]?
category: string?
tier: string?
priority: string?
last_updated: string?
version: string?
url: string?                     # For web clips
source: string?                  # Source tracking
---
```

### Database Schema

#### rag_documents Table
```sql
CREATE TABLE rag_documents (
    id TEXT PRIMARY KEY,
    title TEXT NOT NULL,
    content TEXT NOT NULL,
    path TEXT NOT NULL,
    category TEXT NOT NULL DEFAULT 'general',
    tags TEXT,                    -- JSON array
    scopes TEXT,                  -- JSON array
    tier TEXT DEFAULT 'warm',
    priority TEXT DEFAULT 'P2',
    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL,
    status TEXT NOT NULL DEFAULT 'active',
    vector_id TEXT,
    metadata TEXT,
    content_hash TEXT,
    file_size INTEGER DEFAULT 0,
    last_indexed INTEGER,
    indexing_status TEXT DEFAULT 'pending',
    source TEXT DEFAULT 'tibrain',
    url TEXT,
    conflict_id TEXT,
    version INTEGER DEFAULT 1
);
```

#### sync_conflicts Table
```sql
CREATE TABLE sync_conflicts (
    id TEXT PRIMARY KEY,
    document_id TEXT NOT NULL,
    conflict_type TEXT NOT NULL,
    local_version TEXT NOT NULL,
    remote_version TEXT NOT NULL,
    detected_at INTEGER NOT NULL,
    resolved_at INTEGER,
    resolution TEXT,
    resolved_by TEXT,
    metadata TEXT
);
```

#### sync_audit Table
```sql
CREATE TABLE sync_audit (
    id TEXT PRIMARY KEY,
    operation TEXT NOT NULL,
    source TEXT NOT NULL,
    document_id TEXT,
    file_path TEXT,
    operation_time INTEGER NOT NULL,
    status TEXT NOT NULL,
    error_message TEXT,
    duration_ms INTEGER,
    metadata TEXT
);
```

---

## Frontmatter Mapping Logic

### Obsidian → Ti Brain
```python
def map_obsidian_to_tibrain(obsidian_fm: Dict) -> Dict:
    tibrain_fm = {
        'title': obsidian_fm.get('title', ''),
        'tags': validate_tags(obsidian_fm.get('tags', [])),
        'scopes': validate_scopes(obsidian_fm.get('scopes', [])),
        'category': obsidian_fm.get('category', 'general'),
        'tier': obsidian_fm.get('tier', 'T3'),
        'priority': obsidian_fm.get('priority', 'P2'),
        'last_updated': obsidian_fm.get('last_updated', datetime.now().isoformat()),
    }
    if 'version' in obsidian_fm:
        tibrain_fm['version'] = obsidian_fm['version']
    return tibrain_fm
```

### Ti Brain → Obsidian
```python
def map_tibrain_to_obsidian(tibrain_fm: Dict) -> Dict:
    obsidian_fm = {
        'title': tibrain_fm.get('title', ''),
        'tags': tibrain_fm.get('tags', []),
        'scopes': tibrain_fm.get('scopes', []),
        'category': tibrain_fm.get('category', 'general'),
        'tier': tibrain_fm.get('tier', 'T3'),
        'priority': tibrain_fm.get('priority', 'P2'),
        'last_updated': tibrain_fm.get('last_updated', ''),
    }
    if 'version' in tibrain_fm:
        obsidian_fm['version'] = tibrain_fm['version']
    return obsidian_fm
```

---

## Security Implementation

### Authentication
```go
// MCP Server authentication
type AuthConfig struct {
    APIKey     string
    EnableAuth  bool
    TokenExpiry time.Duration
}

func (c *AuthConfig) ValidateToken(token string) bool {
    return token == c.APIKey
}
```

### Path Policy
```go
// Path policy enforcement
type PathPolicy struct {
    readPaths  []string
    writePaths []string
    readOnly   bool
}

func (p *PathPolicy) AssertReadable(path string) error {
    if p.readOnly && !p.isWritePath(path) {
        return errors.New("read-only mode")
    }
    if !p.matchesAny(path, p.readPaths) && !p.matchesAny(path, p.writePaths) {
        return errors.New("path forbidden")
    }
    return nil
}
```

### Input Validation
```python
# Input sanitization
def sanitize_path(path: str) -> str:
    # Prevent directory traversal
    path = os.path.normpath(path)
    if path.startswith(".."):
        raise ValueError("Invalid path")
    return path

def validate_frontmatter(frontmatter: Dict) -> bool:
    # Validate required fields
    required = ['title', 'tags', 'scopes']
    for field in required:
        if field not in frontmatter:
            return False
    return True
```

---

## Performance Optimization

### Database Optimization
```sql
-- Indexes for fast filtering
CREATE INDEX idx_rag_documents_tags ON rag_documents(tags);
CREATE INDEX idx_rag_documents_scopes ON rag_documents(scopes);
CREATE INDEX idx_rag_documents_category ON rag_documents(category);
CREATE INDEX idx_rag_documents_tier ON rag_documents(tier);
CREATE INDEX idx_rag_documents_updated ON rag_documents(updated_at);
CREATE INDEX idx_rag_documents_source ON rag_documents(source);
```

### Caching Strategy
```python
# Frontmatter cache
from functools import lru_cache

@lru_cache(maxsize=1000)
def parse_frontmatter_cached(content: str) -> Dict:
    return parse_frontmatter(content)

# Query cache
import redis
redis_client = redis.Redis(host='localhost', port=6379, db=0)

def cache_query_results(query: str, results: List[Dict], ttl: int = 300):
    redis_client.setex(f"query:{query}", ttl, json.dumps(results))
```

### Batch Processing
```python
# Batch ingestion
def batch_ingest(files: List[str], batch_size: int = 100):
    for i in range(0, len(files), batch_size):
        batch = files[i:i + batch_size]
        for file in batch:
            ingest_file(file)
        # Commit batch
        commit_batch()
```

---

## Error Handling

### Error Categories
```go
type ErrorCategory int

const (
    ErrorCategoryValidation ErrorCategory = iota
    ErrorCategoryNetwork
    ErrorCategoryFileSystem
    ErrorCategoryDatabase
    ErrorCategoryConflict
)

type SyncError struct {
    Category    ErrorCategory
    Message     string
    Path        string
    Timestamp   time.Time
    Retryable   bool
}

func (e *SyncError) Error() string {
    return fmt.Sprintf("[%s] %s: %s", e.Category, e.Path, e.Message)
}
```

### Retry Logic
```python
# Exponential backoff
import time
from functools import wraps

def retry(max_attempts: int = 3, backoff_factor: float = 2.0):
    def decorator(func):
        @wraps(func)
        def wrapper(*args, **kwargs):
            attempt = 0
            while attempt < max_attempts:
                try:
                    return func(*args, **kwargs)
                except Exception as e:
                    attempt += 1
                    if attempt == max_attempts:
                        raise
                    wait_time = backoff_factor ** attempt
                    time.sleep(wait_time)
        return wrapper
    return decorator
```

---

## Testing Strategy

### Unit Tests
```go
// frontmatter/parser_test.go
func TestParseFrontmatter(t *testing.T) {
    content := `---
title: Test
tags: [test]
---
`
    parser := NewParser()
    fm, err := parser.ParseFrontmatter(content)
    assert.NoError(t, err)
    assert.Equal(t, "Test", fm.Title)
    assert.Equal(t, []string{"test"}, fm.Tags)
}
```

### Integration Tests
```python
# tests/test_sync_integration.py
def test_obsidian_to_tibrain_sync():
    sync = ObsidianSync(vault_path, tibrain_path)
    result = sync.sync_obsidian_to_tibrain(test_file)
    assert result is True
    assert os.path.exists(target_file)
```

### E2E Tests
```python
# tests/test_e2e_sync.py
def test_full_sync_cycle():
    # Create file in Obsidian
    create_obsidian_note(test_note)
    
    # Sync to Ti Brain
    sync.sync_obsidian_to_tibrain(test_file)
    
    # Verify in database
    docs = query_database("test query")
    assert len(docs) > 0
    
    # Modify in Ti Brain
    modify_tibrain_document(doc_id)
    
    # Sync back to Obsidian
    sync.sync_tibrain_to_obsidian(content, frontmatter, target_path)
    
    # Verify in Obsidian
    obsidian_note = read_obsidian_note(target_path)
    assert "modified" in obsidian_note
```

---

## References

- **Executive Summary:** `OBSIDIAN_SYNC_ARCHITECTURE.md`
- **Implementation Plan:** `OBSIDIAN_INTEGRATION_MASTER_PLAN.md`
- **obsidian-mcp-server:** `../obsidian-mcp-server/README.md`
- **obsidian-mcp-server CLAUDE:** `../obsidian-mcp-server/CLAUDE.md`
- **obsidian-headless:** `../obsidian-headless/README.md`
- **RAG Architecture:** `RAG_ARCHITECTURE.md`
- **RAG Usage Guide:** `RAG_USAGE_GUIDE.md`
- **Frontmatter Parser:** `../internal/frontmatter/parser.go`
- **Frontmatter Mapping:** `../internal/frontmatter/mapping.go`

---

*Document Version: 2.0.0*  
*Last Updated: 2026-05-22*  
*Next Review: 2026-08-22*  
*Owner: Ti Brain Development Team*
