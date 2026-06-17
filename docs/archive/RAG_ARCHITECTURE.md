---
title: "RAG Architecture Design - Option 3 Implementation"
category: "architecture"
tier: "hot"
tags: ["rag", "architecture", "frontmatter", "tags", "scopes", "tibrain"]
priority: "P0"
last_updated: "2026-05-22"
version: "1.0.0"
---

# RAG Architecture Design - Option 3 Implementation

> **Version**: 1.0.0
> **Date**: 2026-05-22
> **Purpose**: Design RAG system with Option 3 (current structure + frontmatter parsing)
> **Approach**: Build on existing TiBrain RAG system, add frontmatter support

---

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────┐
│                    Knowledge Files (Current Structure)     │
│  apps/tibrain/knowledge/ti-ecosystem-docs/**/*.md         │
│  - Files stay in current locations                         │
│  - Frontmatter contains tags, scopes, metadata             │
└─────────────────────────────────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────┐
│                Frontmatter Parser (NEW)                     │
│  - Parse YAML frontmatter from markdown files              │
│  - Extract tags, scopes, category, tier, last_updated      │
│  - Validate against TAGS.md taxonomy                       │
└─────────────────────────────────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────┐
│              Ingestion Pipeline (ENHANCED)                  │
│  - Parse frontmatter                                        │
│  - Generate embeddings                                      │
│  - Store with tags/scopes metadata                         │
│  - Update rag_documents table with new columns             │
└─────────────────────────────────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────┐
│                 Vector Store (EXISTING)                     │
│  - SQLite with vector embeddings                           │
│  - Cosine similarity search                                 │
│  - Hybrid search (vector + keyword)                        │
└─────────────────────────────────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────┐
│               Query Pipeline (ENHANCED)                     │
│  - Filter by tags/scopes                                    │
│  - Vector search with metadata filters                      │
│  - Re-ranking with tag/scopes relevance                    │
│  - Context assembly with filtered results                  │
└─────────────────────────────────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────┐
│                    Response Generation                       │
│  - LLM response with citations                             │
│  - Tag/scope metadata in response                          │
│  - Confidence scores                                        │
└─────────────────────────────────────────────────────────────┘
```

---

## Database Schema Updates

### New Columns for rag_documents

```sql
-- Add tags and scopes support
ALTER TABLE rag_documents ADD COLUMN tags TEXT; -- JSON array
ALTER TABLE rag_documents ADD COLUMN scopes TEXT; -- JSON array
ALTER TABLE rag_documents ADD COLUMN category TEXT; -- From frontmatter
ALTER TABLE rag_documents ADD COLUMN tier TEXT; -- From frontmatter
ALTER TABLE rag_documents ADD COLUMN priority TEXT; -- From frontmatter

-- Indexes for tag/scope filtering
CREATE INDEX idx_rag_documents_tags ON rag_documents(tags);
CREATE INDEX idx_rag_documents_scopes ON rag_documents(scopes);
CREATE INDEX idx_rag_documents_category ON rag_documents(category);
CREATE INDEX idx_rag_documents_tier ON rag_documents(tier);
```

---

## Component Design

### 1. Frontmatter Parser

**File**: `apps/tibrain/internal/frontmatter/parser.go`

**Responsibilities**:
- Parse YAML frontmatter from markdown files
- Extract metadata: tags, scopes, category, tier, priority, last_updated
- Validate tags against TAGS.md taxonomy
- Return structured metadata

**API**:
```go
type Frontmatter struct {
    Title       string   `yaml:"title"`
    Category    string   `yaml:"category"`
    Tier        string   `yaml:"tier"`
    Tags        []string `yaml:"tags"`
    Scopes      []string `yaml:"scopes"`
    Priority    string   `yaml:"priority"`
    LastUpdated string   `yaml:"last_updated"`
    Version     string   `yaml:"version"`
}

func ParseFrontmatter(content string) (*Frontmatter, error)
func ValidateTags(tags []string) error
func ValidateScopes(scopes []string) error
```

### 2. Enhanced Ingestion Pipeline

**File**: `apps/tibrain/internal/ingestion/enhanced.go`

**Responsibilities**:
- Walk knowledge directory recursively
- Parse frontmatter from each markdown file
- Generate embeddings for content
- Store with tags/scopes metadata
- Update rag_documents table

**API**:
```go
type IngestionConfig struct {
    KnowledgeDir string
    ForceReindex  bool
    ValidateTags  bool
}

func IngestWithFrontmatter(config IngestionConfig) error
func ProcessFile(filePath string) (*RAGDocument, error)
func UpdateDocumentWithFrontmatter(doc *RAGDocument, fm *Frontmatter)
```

### 3. Enhanced Query Pipeline

**File**: `apps/tibrain/internal/query/enhanced.go`

**Responsibilities**:
- Accept tag/scope filters in queries
- Filter documents by tags/scopes
- Re-rank results by tag/scope relevance
- Return filtered results with metadata

**API**:
```go
type QueryFilter struct {
    Tags   []string
    Scopes []string
    Category string
    Tier   string
}

type EnhancedQuery struct {
    Query  string
    Filter QueryFilter
    TopK   int
}

func QueryWithFilters(query EnhancedQuery) ([]RAGDocument, error)
func FilterByTags(docs []RAGDocument, tags []string) []RAGDocument
func FilterByScopes(docs []RAGDocument, scopes []string) []RAGDocument
```

### 4. CLI Scripts

**File**: `apps/tibrain/scripts/ingest_kb.py`

**Usage**:
```bash
# Ingest all files with frontmatter parsing
py scripts/ingest_kb.py --dir knowledge/ti-ecosystem-docs --validate-tags

# Force re-index
py scripts/ingest_kb.py --dir knowledge/ti-ecosystem-docs --force

# Ingest specific directory
py scripts/ingest_kb.py --dir knowledge/ti-ecosystem-docs/content/skills
```

**File**: `apps/tibrain/scripts/query_kb.py`

**Usage**:
```bash
# Query with tag filter
py scripts/query_kb.py --query "authentication patterns" --tags authentication,security

# Query with scope filter
py scripts/query_kb.py --query "testing requirements" --scopes auth,code

# Query with both tags and scopes
py scripts/query_kb.py --query "provider implementation" --tags provider --scopes providers

# Query without filters
py scripts/query_kb.py --query "how to implement new feature"
```

---

## Implementation Plan

### Phase 1: Frontmatter Parser
- [ ] Create `internal/frontmatter/parser.go`
- [ ] Implement YAML parsing
- [ ] Implement tag validation against TAGS.md
- [ ] Implement scope validation against SCOPES.md
- [ ] Add unit tests

### Phase 2: Database Schema Updates
- [ ] Add new columns to rag_documents table
- [ ] Create indexes for tag/scope filtering
- [ ] Migration script for existing data
- [ ] Test schema changes

### Phase 3: Enhanced Ingestion
- [ ] Update ingestion pipeline to use frontmatter parser
- [ ] Implement recursive directory walking
- [ ] Store tags/scopes in database
- [ ] Add validation flags
- [ ] Test ingestion with sample files

### Phase 4: Enhanced Query
- [ ] Add tag/scope filter support to query pipeline
- [ ] Implement filtering logic
- [ ] Add re-ranking by tag/scope relevance
- [ ] Test queries with filters

### Phase 5: CLI Scripts
- [ ] Create ingest_kb.py script
- [ ] Create query_kb.py script
- [ ] Add command-line argument parsing
- [ ] Test scripts end-to-end

### Phase 6: Documentation
- [ ] Update docs-lifecycle.md with RAG integration
- [ ] Create RAG usage guide
- [ ] Add examples and best practices
- [ ] Update TAGS.md with RAG metadata requirements

---

## Benefits of Option 3

### ✅ Advantages
1. **No file movement** — Files stay in current structure
2. **Multi-tag support** — Files can have multiple tags/scopes
3. **Flexible filtering** — Filter by tags, scopes, or both
4. **Easy maintenance** — No symlinks or duplicates
5. **Leverages existing RAG** — Builds on top of current system
6. **Frontmatter-driven** — Metadata in files, not external database

### 📊 Comparison with Other Options

| Aspect | Option 1 (Tag Folders) | Option 2 (Scope Folders) | Option 3 (Current) |
|--------|----------------------|-------------------------|-------------------|
| File movement | Required | Required | None |
| Multi-tag support | Via symlinks | No | ✅ Yes |
| Maintenance | High (symlinks) | Medium | Low |
| RAG complexity | High | Medium | Low |
| Flexibility | Medium | Low | High |

---

## Performance Considerations

### Indexing Strategy
- Tags stored as JSON array in single column
- Scopes stored as JSON array in single column
- SQLite JSON1 extension for querying
- Indexes on tags/scopes columns for fast filtering

### Query Performance
- Tag filtering: `WHERE tags LIKE '%"tag_name"%'`
- Scope filtering: `WHERE scopes LIKE '%"scope_name"%'`
- Combined filters: AND/OR logic
- Vector search first, then filter by metadata

### Caching Strategy
- Cache query results by (query, tags, scopes) tuple
- Cache frontmatter parsing results
- Cache tag validation results

---

## Quality Assurance

### Testing Strategy
1. **Unit tests** for frontmatter parser
2. **Integration tests** for ingestion pipeline
3. **End-to-end tests** for query pipeline
4. **Performance tests** for large-scale ingestion

### Validation Checks
- Frontmatter must be valid YAML
- Tags must exist in TAGS.md
- Scopes must exist in SCOPES.md
- Category must be valid
- Tier must be valid (hot/warm/cold)

### Error Handling
- Graceful fallback for missing frontmatter
- Warning for invalid tags/scopes
- Retry logic for failed ingestion
- Logging for all operations

---

## Migration Path

### For Existing Documents
1. Run ingestion script with `--force` flag
2. Parse frontmatter from all files
3. Update database with tags/scopes
4. Validate all tags against taxonomy
5. Generate report of validation issues

### For New Documents
1. Follow frontmatter template from docs-lifecycle.md
2. Include tags and scopes in frontmatter
3. Run ingestion script on new files
4. Validate tags before ingestion

---

## Success Metrics

### Coverage
- 100% of tagged files ingested with frontmatter
- 100% of tags validated against taxonomy
- 100% of scopes validated against SCOPES.md

### Performance
- Ingestion time: < 1000 files/minute
- Query time: < 500ms with filters
- Cache hit rate: > 80%

### Quality
- Zero invalid tags in database
- Zero invalid scopes in database
- All queries return relevant results

---

*This architecture leverages existing TiBrain RAG infrastructure while adding frontmatter support for tag/scope-based filtering. No file movement required, multi-tag support built-in, and easy maintenance.*
