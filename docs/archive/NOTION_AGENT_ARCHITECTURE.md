# Notion Agent Architecture

> **Date**: 2026-05-23  
> **Concept**: Unified Notion Agent as single interface for all Notion operations  
> **Simplification**: Hide Notion Manager, Notion AI, Notion API complexity

---

## 🎯 Problem

### Complexity of Previous Approach:
- Obsidian → Notion sync (separate)
- Notion AI queries (via Notion Manager)  
- Notion database management (direct Notion API)
- 3 different interfaces for Notion operations

### Solution: Notion Agent
- **Single unified interface** for all Notion operations
- **Hide complexity** of Notion Manager, Notion AI, Notion API
- **Simplify external systems** - only talk to Notion Agent

---

## 🏗️ Notion Agent Architecture

### System Overview:

```
┌─────────────────────────────────────────────────────┐
│                 Notion Agent                         │
│            (Unified Notion Interface)               │
│                                                      │
│  ┌──────────────────────────────────────────────┐  │
│  │              Agent Core                       │  │
│  │  - Operation routing                          │  │
│  │  - Request/response handling                 │  │
│  │  - Error handling                            │  │
│  │  - Caching layer                             │  │
│  └──────────────────────────────────────────────┘  │
│                                                      │
│  ┌──────────┐  ┌──────────┐  ┌──────────────────┐  │
│  │  Sync    │  │  Query   │  │  Database        │  │
│  │  Module  │  │  Module  │  │  Manager         │  │
│  │          │  │          │  │                  │  │
│  │  - Obsi  │  │  - Notion│  │  - Pages         │  │
│  │    dian  │  │    AI    │  │  - Databases     │  │
│  │    sync  │  │  - Hybrid│  │  - Metadata      │  │
│  │          │  │    RAG   │  │  - Relations     │  │
│  └──────────┘  └──────────┘  └──────────────────┘  │
│                                                      │
│  ┌──────────────────────────────────────────────┐  │
│  │         Notion Backend Integration          │  │
│  │  - Notion Manager (API gateway)            │  │
│  │  - Notion API (direct)                      │  │
│  │  - Authentication                          │  │
│  │  - Rate limiting                           │  │
│  └──────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────┘
                    │
          ┌─────────┴──────────┐
          │                    │
          ↓                    ↓
┌────────────────┐    ┌────────────────┐
│  Notion        │    │  Notion        │
│  Manager       │    │  Database      │
│  (API Gateway) │    │  (Notion.com)  │
└────────────────┘    └────────────────┘
```

### External Interfaces:

```
┌─────────────┐         ┌─────────────┐         ┌─────────────┐
│  Obsidian   │         │  Ti Brain   │         │  User CLI   │
│  Vault      │         │  RAG        │         │             │
│             │         │             │         │             │
└──────┬──────┘         └──────┬──────┘         └──────┬──────┘
       │                      │                      │
       │                      │                      │
       └──────────────────────┴──────────────────────┘
                              │
                              ↓
                     ┌─────────────────┐
                     │  Notion Agent   │
                     │  (Unified API)  │
                     └─────────────────┘
```

---

## 🔧 Notion Agent API

### Core Operations:

#### 1. Sync Operations

```go
// Sync Obsidian vault to Notion database
type SyncRequest struct {
    VaultPath      string `json:"vault_path"`
    DatabaseID     string `json:"database_id"`
    Incremental    bool   `json:"incremental"`
    Direction      string `json:"direction"` // "push", "pull", "bidirectional"
}

type SyncResponse struct {
    Success       bool     `json:"success"`
    FilesSynced   int      `json:"files_synced"`
    Errors        []string `json:"errors"`
    Timestamp     string   `json:"timestamp"`
}
```

#### 2. Query Operations

```go
// Query Notion AI with hybrid search
type QueryRequest struct {
    Query         string `json:"query"`
    UseTiBrain    bool   `json:"use_tibrain"` // Enable hybrid search
    Model         string `json:"model"`      // "sonnet-4.6", "claude-3.5", etc.
    TopK          int    `json:"top_k"`
    QueryType     string `json:"query_type"` // "current_note", "general", "technical"
}

type QueryResponse struct {
    Answer        string        `json:"answer"`
    Sources       []Source      `json:"sources"`
    Model         string        `json:"model"`
    HybridMode    bool          `json:"hybrid_mode"`
    SourcesUsed   []string      `json:"sources_used"` // ["notion_ai", "tibrain_rag"]
}
```

#### 3. Database Operations

```go
// Create/update Notion pages
type CreatePageRequest struct {
    Title         string                 `json:"title"`
    Content       string                 `json:"content"`
    DatabaseID    string                 `json:"database_id"`
    Properties   map[string]interface{} `json:"properties"`
}

type CreatePageResponse struct {
    PageID        string `json:"page_id"`
    Success       bool   `json:"success"`
}
```

#### 4. Management Operations

```go
// Get Notion Agent status
type StatusResponse struct {
    Status        string                 `json:"status"` // "running", "idle", "error"
    NotionManager Connected              `json:"notion_manager_connected"`
    DatabaseCount int                    `json:"database_count"`
    LastSync      string                 `json:"last_sync"`
    Version       string                 `json:"version"`
}
```

---

## 🎯 Notion Agent Modules

### 1. Sync Module

**Responsibilities:**
- Obsidian vault → Notion database sync
- Frontmatter mapping (use existing Go mapper)
- Incremental sync detection
- Conflict resolution

**Implementation:**
```go
package notionagent

type SyncModule struct {
    frontmatterMapper *frontmatter.Mapper
    vaultScanner     *VaultScanner
    notionAPI        *notion.Client
}

func (s *SyncModule) SyncVault(req SyncRequest) (*SyncResponse, error) {
    // 1. Scan vault
    files := s.vaultScanner.Scan(req.VaultPath)
    
    // 2. Map frontmatter
    for _, file := range files {
        mapped := s.frontmatterMapper.MapObsidianToTiBrain(file)
    }
    
    // 3. Push to Notion
    // ... existing sync logic
    
    return &SyncResponse{Success: true}
}
```

### 2. Query Module

**Responsibilities:**
- Notion AI queries (via Notion Manager)
- Hybrid RAG queries (Notion AI + Ti Brain)
- Result merging and reranking
- Source weighting

**Implementation:**
```go
package notionagent

type QueryModule struct {
    notionManager     *notion_manager.Client
    tiBrainRAG        *tibrain.RAGClient
    reranker         *CrossEncoderReranker
    sourceWeights     map[string]float64
}

func (q *QueryModule) Query(req QueryRequest) (*QueryResponse, error) {
    var notionResults, tibrainResults []Result
    
    // Parallel queries
    if req.UseTiBrain {
        notionResults = q.notionManager.Query(req.Query)
        tibrainResults = q.tiBrainRAG.Query(req.Query)
    } else {
        notionResults = q.notionManager.Query(req.Query)
    }
    
    // Merge and rerank
    if req.UseTiBrain {
        merged := q.mergeResults(notionResults, tibrainResults)
        ranked := q.reranker.Rank(merged)
        return &QueryResponse{Answer: ranked[0].Content}
    }
    
    return &QueryResponse{Answer: notionResults[0].Content}
}
```

### 3. Database Manager

**Responsibilities:**
- Create Notion databases
- Manage page properties
- Handle relationships
- Metadata management

**Implementation:**
```go
package notionagent

type DatabaseManager struct {
    notionAPI      *notion.Client
    schemaCache    map[string]DatabaseSchema
}

func (d *DatabaseManager) CreateDatabase(schema DatabaseSchema) (string, error) {
    // Create Notion database with schema
    // Cache schema for future operations
}
```

### 4. Notion Backend Integration

**Responsibilities:**
- Notion Manager API gateway connection
- Notion API authentication
- Rate limiting and retry logic
- Account pooling (if needed)

**Implementation:**
```go
package notionagent

type NotionBackend struct {
    notionManager *notion_manager.Client
    notionAPI     *notion.Client
    accountPool   *AccountPool
    rateLimiter   *RateLimiter
}

func (n *NotionBackend) QueryNotionAI(query string) (string, error) {
    // Try Notion Manager first (better rate limits)
    if n.notionManager.IsAvailable() {
        return n.notionManager.Query(query)
    }
    
    // Fallback to direct Notion AI
    return n.notionAPI.QueryAI(query)
}
```

---

## 🚀 Implementation Plan

### Phase 1: Notion Agent Core (Week 1)

#### Deliverables:
1. Notion Agent Go service
2. Core operation routing
3. Basic API endpoints
4. Error handling

#### Tasks:
- [ ] Create notionagent Go package
- [ ] Define API structures
- [ ] Implement HTTP server
- [ ] Add logging and monitoring

### Phase 2: Sync Module (Week 2)

#### Deliverables:
1. Obsidian vault sync integration
2. Frontmatter mapper integration
3. Incremental sync
4. Conflict resolution

#### Tasks:
- [ ] Integrate existing Go frontmatter mapper
- [ ] Implement vault scanner
- [ ] Add sync detection logic
- [ ] Test with demo vault

### Phase 3: Query Module (Week 3)

#### Deliverables:
1. Notion Manager integration
2. Ti Brain RAG integration
3. Hybrid search engine
4. Result reranking

#### Tasks:
- [ ] Connect to Notion Manager
- [ ] Connect to Ti Brain RAG
- [ ] Implement parallel queries
- [ ] Add cross-encoder reranker

### Phase 4: Database Manager (Week 4)

#### Deliverables:
1. Notion database operations
2. Schema management
3. Property handling
4. Relationship management

#### Tasks:
- [ ] Implement database creation
- [ ] Add schema validation
- [ ] Handle page properties
- [ ] Manage relationships

---

## 🎯 External Integrations

### Obsidian Integration:

```python
# Obsidian plugin for Notion Agent
class NotionAgentPlugin:
    def __init__(self, agent_url: str):
        self.agent_url = agent_url
    
    def sync_vault(self, vault_path: str, database_id: str):
        response = requests.post(
            f"{self.agent_url}/api/sync",
            json={
                "vault_path": vault_path,
                "database_id": database_id,
                "direction": "push"
            }
        )
        return response.json()
    
    def query(self, query: str, use_tibrain: bool = True):
        response = requests.post(
            f"{self.agent_url}/api/query",
            json={
                "query": query,
                "use_tibrain": use_tibrain
            }
        )
        return response.json()
```

### Ti Brain Integration:

```go
// Ti Brain RAG client for Notion Agent
package tibrain

type RAGClient struct {
    baseURL string
}

func (r *RAGClient) Query(query string, topK int) ([]Result, error) {
    resp, err := http.Post(
        fmt.Sprintf("%s/api/rag/query", r.baseURL),
        "application/json",
        bytes.NewBuffer([]byte(fmt.Sprintf(`{"query": "%s", "top_k": %d}`, query, topK))),
    )
    // ... parse response
}
```

---

## 📊 System Configuration

### Environment Variables:
```bash
# Notion Agent
export NOTION_AGENT_PORT=8082
export NOTION_AGENT_LOG_LEVEL=info

# Notion Manager
export ANTHROPIC_BASE_URL=http://localhost:8081
export ANTHROPIC_API_KEY=notion-manager-api-key

# Notion API
export NOTION_TOKEN=notion-api-token
export NOTION_DATABASE_ID=database-id

# Ti Brain
export TIBRAIN_RAG_URL=http://localhost:1810

# Obsidian
export OBSIDIAN_VAULT_PATH=/path/to/vault
```

### Config File:
```yaml
notion_agent:
  port: 8082
  log_level: info
  
notion_manager:
  url: http://localhost:8081
  api_key: your-api-key
  
notion_api:
  token: your-notion-token
  database_id: your-database-id

tibrain_rag:
  url: http://localhost:1810
  enabled: true
  
obsidian:
  vault_path: /path/to/vault
  
sync:
  incremental: true
  conflict_resolution: "notion_wins"
  
query:
  default_model: sonnet-4.6
  default_top_k: 5
  use_hybrid: true
  
weights:
  current_note:
    notion_ai: 0.7
    tibrain_rag: 0.3
  general_knowledge:
    notion_ai: 0.3
    tibrain_rag: 0.7
  technical:
    notion_ai: 0.2
    tibrain_rag: 0.8
```

---

## 🎯 Benefits vs Previous Approach

### Before (Multiple Interfaces):
- Obsidian → Notion sync (separate)
- Notion AI → Notion Manager (separate)
- Notion API → Direct (separate)
- 3 different interfaces, 3 points of integration

### After (Notion Agent):
- **Single interface** for all Notion operations
- **Hidden complexity** of Notion Manager, Notion AI, Notion API
- **Simplified external systems** - only talk to Notion Agent
- **Centralized control** - one place to manage all Notion operations

---

## 🔧 Notion Agent API Examples

### Example 1: Sync Obsidian Vault

```bash
# Sync vault to Notion database
curl -X POST http://localhost:8082/api/sync \
  -H "Content-Type: application/json" \
  -d '{
    "vault_path": "/path/to/vault",
    "database_id": "database-id",
    "direction": "push",
    "incremental": true
  }'
```

Response:
```json
{
  "success": true,
  "files_synced": 42,
  "errors": [],
  "timestamp": "2026-05-23T12:00:00Z"
}
```

### Example 2: Query with Hybrid Search

```bash
# Query with Ti Brain RAG
curl -X POST http://localhost:8082/api/query \
  -H "Content-Type: application/json" \
  -d '{
    "query": "How do I implement OAuth PKCE?",
    "use_tibrain": true,
    "model": "sonnet-4.6",
    "query_type": "technical",
    "top_k": 5
  }'
```

Response:
```json
{
  "answer": "OAuth PKCE is implemented by...",
  "sources": [
    {
      "source": "notion_ai",
      "content": "...",
      "confidence": 0.8
    },
    {
      "source": "tibrain_rag", 
      "content": "...",
      "confidence": 0.9
    }
  ],
  "model": "sonnet-4.6",
  "hybrid_mode": true,
  "sources_used": ["notion_ai", "tibrain_rag"]
}
```

### Example 3: Create Notion Page

```bash
# Create new page
curl -X POST http://localhost:8082/api/pages \
  -H "Content-Type: application/json" \
  -d '{
    "title": "New Note",
    "content": "# Note content...",
    "database_id": "database-id",
    "properties": {
      "tags": ["oauth", "security"],
      "scope": "implementation"
    }
  }'
```

---

## 🚀 Deployment

### Development:
```bash
# Start Notion Agent
go run ./cmd/notion-agent

# Start Notion Manager
cd /z/10_WORKPLACE/Ti/apps/integrations/notion_manager
go run ./cmd/notion_manager

# Start Ti Brain RAG (if running as service)
# ... existing Ti Brain service
```

### Production:
```bash
# Build Notion Agent
go build -o notion-agent ./cmd/notion-agent

# Run as service
./notion-agent --config /etc/notion-agent/config.yaml
```

---

## ✅ Recommendation

### **Primary Approach: Notion Agent**

**Architecture**:
- Single Notion Agent service
- Unified API for all Notion operations
- Hide complexity of Notion Manager, Notion AI, Notion API
- Simplify external system integration

**Timeline**: 4 weeks (Phase 1-4)
**Effort**: 60-80 hours
**Value**: High (simplified integration, centralized control)

---

## 🎯 Next Steps

**Option A**: Implement Notion Agent (Recommended)
- Phase 1-4 as outlined
- Unified interface for all Notion operations
- Simplified external integration

**Option B**: Start with core only (Phase 1-2)
- Sync + basic query first
- Expand later

**Option C**: Prototype concept first
- Quick POC of Notion Agent
- Validate unified interface concept

**Option D**: Your specific idea/concern

Notion Agent simplifies everything by abstracting all Notion-related operations into a single, unified service. This is a much cleaner approach than multiple interfaces.