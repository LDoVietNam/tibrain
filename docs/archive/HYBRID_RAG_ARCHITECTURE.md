# Hybrid RAG Architecture: Notion AI + Ti Brain RAG

> **Date**: 2026-05-23  
> **Insight**: Notion AI + Ti Brain RAG = Enhanced Information Retrieval  
> **Shift**: From "sync" to "AI-powered knowledge retrieval"

---

## 🧠 Problem Re-Definition

### Original Understanding:
- **Goal**: Sync Obsidian notes to Notion database
- **Value**: Knowledge backup, collaboration
- **Approach**: File conversion and sync

### New Understanding:
- **Goal**: Leverage Notion AI + Ti Brain RAG for enhanced queries
- **Value**: **Better information retrieval through hybrid search**
- **Approach**: Hybrid RAG with dual knowledge sources

---

## 🎯 Use Case Examples

### Query: "How do I implement OAuth PKCE flow?"

#### Without Hybrid:
- **Notion AI**: Searches notes → finds "OAuth PKCE Flow" → limited to current notes
- **Result**: Incomplete answer, missing broader context

#### With Hybrid:
- **Notion AI**: Searches notes → "OAuth PKCE Flow" + content
- **Ti Brain RAG**: Searches knowledge base → "OAuth patterns", "PKCE implementations"
- **Hybrid Result**: Comprehensive answer with:
  - Current implementation details
  - Best practices from broader knowledge
  - Code examples from multiple sources
  - Security considerations

---

## 🏗️ Hybrid RAG Architecture

### System Components:

#### Layer 1: Storage & Sync
```
Obsidian Vault (Markdown) → Notion Database (Storage)
```
- Obsidian → Notion sync (for storage layer)
- Notion AI queries storage
- Ti Brain sync (optional, for knowledge base)

#### Layer 2: Dual Knowledge Sources

**Source A: Notion AI (Notion Manager)**
- Queries current Notion database
- Real-time access to notes
- Notion AI built-in search
- Accessed via Notion Manager API gateway

**Source B: Ti Brain RAG**
- Broader knowledge base
- Cross-project knowledge
- Historical patterns
- Vector search + semantic search

#### Layer 3: Hybrid Search Engine
```
Query → Split → Parallel Search → Merge & Rank → Result
         ↓         ↓              ↓         ↓
    Notion AI  Ti Brain RAG  Weight  Enhanced
```

---

## 🔧 Technical Implementation

### 1. Notion AI Integration (via Notion Manager)

```go
// Go service for Notion AI queries
package notionai

type QueryRequest struct {
    Query string `json:"query"`
    Model  string `json:"model"` // e.g., "sonnet-4.6", "claude-sonnet"
    Context int    `json:"context"` // context window size
}

type QueryResponse struct {
    Answer string `json:"answer"`
    Sources []string `json:"sources"`
    Model  string `json:"model"`
}
```

### 2. Ti Brain RAG Integration (Existing)

```go
// Use existing Ti Brain RAG system
package tibrain

func QueryRAG(query string, topK int) ([]Result, error) {
    // Use existing RAG implementation
    // Vector search + semantic search
    return tibrain.RAGQuery(query, topK)
}
```

### 3. Hybrid Search Engine

```python
class HybridSearch:
    def __init__(self, notion_manager, tibrain_rag):
        self.notion_manager = notion_manager
        self.tibrain_rag = tibrain_rag
        self.reranker = CrossEncoderReranker()
    
    def search(self, query: str):
        # Parallel search
        notion_results = self.notion_manager.search(query)
        tibrain_results = self.tibrain_rag.search(query)
        
        # Merge and rank
        merged = self.merge_results(notion_results, tibrain_results)
        ranked = self.reranker.rank(merged)
        
        return ranked
```

---

## 📊 Data Flow

### Query Flow:

```
User Query: "OAuth PKCE implementation"
         ↓
┌─────────────────────────────────┐
│  Query Router                   │
│  (Analyze query type)          │
└─────────────┬───────────────────┘
              │
       ┌───────┴──────────┐
       │                  │
       ↓                  ↓
┌─────────────┐    ┌─────────────┐
│  Notion AI   │    │  Ti Brain   │
│  (search     │    │  RAG        │
│  local       │    │  (knowledge │
│  notes)     │    │  base)      │
└─────┬───────┘    └─────┬──────┘
      │                  │
      └──────────────────┘
                ↓
         ┌────────────────┐
         │   Result Merger   │
         └────────┬─────────┘
                  │
         ┌────────┴────────┐
         │   Cross-Encoder  │
         │   Reranker      │
         └────────┬─────────┘
                  ↓
         Enhanced Result
```

---

## 🎯 Key Features

### 1. Parallel Search
- Notion AI searches current database
- Ti Brain RAG searches knowledge base
- Execute in parallel for low latency

### 2. Result Merging
- Combine results from both sources
- Remove duplicates
- Apply cross-encoder reranking
- Weight sources based on query type

### 3. Source Weighting Strategy

**Query Type:**
- **Current Note Queries**: Notion AI weight 0.7, Ti Brain 0.3
- **General Knowledge**: Ti Brain weight 0.7, Notion AI 0.3
- **Code/Technical**: Ti Brain weight 0.8, Notion AI 0.2

### 4. Context Augmentation
- Notion AI results: Add current note context
- Ti Brain RAG results: Add broader knowledge base context
- Combined: Provide both current and broader context

### 5. Answer Synthesis
- Use Notion AI to synthesize final answer
- Augment with Ti Brain insights
- Provide source citations from both

---

## 🔧 Implementation Details

### Phase 1: Notion AI Integration

#### API Gateway (Notion Manager):
```bash
# Start Notion Manager
cd /z/10_WORKPLACE/Ti/apps/integrations/notion_manager
go run ./cmd/notion_manager

# Query via Notion AI
export ANTHROPIC_BASE_URL=http://localhost:8081
curl http://localhost:8081/v1/messages \
  -d '{"model": "sonnet-4.6", "messages": [{"role": "user", "content": "How do I implement OAuth?"}]}'
```

#### Python Notion AI Client:
```python
class NotionAIClient:
    def __init__(self, notion_manager_url: str, api_key: str):
        self.url = notion_manager_url
        self.api_key = api_key
    
    def query(self, query: str, model: str = "sonnet-4.6"):
        # Query through Notion Manager
        response = requests.post(
            f"{self.url}/v1/messages",
            headers={"Authorization": f"Bearer {self.api_key}"},
            json={
                "model": model,
                "messages": [{"role": "user", "content": query}]
            }
        )
        return response.json()
```

### Phase 2: Ti Brain RAG Integration

#### Use Existing RAG System:
```go
// Ti Brain already has RAG system
// Use existing: internal/rag_system.go
// Query via: POST /api/rag/query
```

#### Python Ti Brain Client:
```python
class TiBrainRAGClient:
    def __init__(self, tibrain_url: str):
        self.url = tibrain_url
    
    def query(self, query: str, top_k: int = 5):
        response = requests.post(
            f"{self.url}/api/rag/query",
            json={
                "query": query,
                "top_k": top_k
            }
        )
        return response.json()
```

### Phase 3: Hybrid Search Engine

#### Cross-Encoder Reranking:
```python
from sentence_transformers import CrossEncoder
import numpy as np

class CrossEncoderReranker:
    def __init__(self):
        self.model = CrossEncoder('ms-marco-electra-base')
    
    def rank(self, results):
        # Re-rank combined results
        # Use cross-encoder for semantic relevance
        return sorted(results, key=lambda x: x['score'], reverse=True)
```

#### Result Merger:
```python
def merge_results(notion_results, tibrain_results):
    merged = []
    
    # Add source tags
    for r in notion_results:
        r['source'] = 'notion_ai'
        merged.append(r)
    
    for r in tibrain_results:
        r['source'] = 'tibrain_rag'
        merged.append(r)
    
    return merged
```

---

## 📊 Search Scenarios

### Scenario 1: Current Note Query
**Query**: "What does this note say about OAuth?"

**Weights**: Notion AI 0.7, Ti Brain 0.3
**Flow**:
- Notion AI: Search current note
- Ti Brain RAG: Search similar patterns (lighter search)
- Merge with Notion AI priority

### Scenario 2: General Knowledge
**Query**: "What are OAuth best practices?"

**Weights**: Notion AI 0.3, Ti Brain 0.7
**Flow**:
- Notion AI: Search for OAuth mentions in notes
- Ti Brain RAG: Search OAuth best practices knowledge base
- Merge with Ti Brain priority

### Scenario 3: Code/Technical
**Query**: "How to implement PKCE in Go?"

**Weights**: Notion AI 0.2, Ti Brain RAG 0.8
**Flow**:
- Notion AI: Search for Go PKCE implementations
- Ti Brain RAG: Search Go OAuth patterns, best practices
- Merge with Ti Brain priority

---

## 🎯 Benefits vs Simple Sync

### Before (Simple Sync):
- Obsidian → Notion storage only
- No AI capabilities
- No enhanced retrieval
- Basic backup/collaboration

### After (Hybrid RAG):
- Obsidian + Notion AI + Ti Brain RAG
- **Enhanced query capabilities**
- **Better information retrieval**
- **AI-powered insights**
- **Dual knowledge sources**
- **Intelligent result ranking**

---

## 🚀 Implementation Plan

### Sprint 1: Notion AI Integration (Week 1)

#### Deliverables:
1. Notion AI client (Python)
2. Integration with Notion Manager
3. Basic query interface
4. Test with sample queries

### Sprint 2: Hybrid Search Engine (Week 2)

#### Deliverables:
1. Result merger
2. Cross-encoder reranking
3. Source weighting logic
4. Answer synthesis

### Sprint 3: Ti Brain RAG Integration (Week 3)

#### Deliverables:
1. Connect to existing Ti Brain RAG
2. Implement dual query logic
3. Context augmentation
4. Enhanced answer synthesis

### Sprint 4: Obsidian Storage Layer (Week 4)

#### Deliverables:
1. Obsidian → Notion sync (for storage)
2. Maintain storage as knowledge base
3. Incremental updates
4. Ensure query sources are fresh

---

## 🎯 Success Metrics

### Quality Metrics:
- Query relevance score: target >0.8
- Source diversity: results from both sources
- Latency: <3 seconds per query
- Answer completeness: >90%

### User Experience Metrics:
- Query satisfaction: target >0.85
- Result diversity: target >70%
- Source attribution: clear
- Speed: <5 seconds per query

---

## 🔧 Configuration

### Environment Variables:
```bash
# Notion Manager
export ANTHROPIC_BASE_URL=http://localhost:8081
export ANTHROPIC_API_KEY=notion-manager-api-key

# Ti Brain
export TIBRAIN_URL=http://localhost:1810

# Obsidian Vault
export OBSIDIAN_VAULT_PATH=/path/to/vault

# Notion Database
export NOTION_DATABASE_ID=database-id
export NOTION_TOKEN=notion-api-token
```

### Config File:
```yaml
notion_manager:
  url: http://localhost:8081
  api_key: your-api-key

tibrain_rag:
  url:  http://localhost:1810
  top_k: 5

obsidian:
  vault_path: /path/to/vault
  
weights:
  current_note:
    notion_ai: 0.7
    tibrain_rag: 0.3
  general_knowledge:
    notion_ai:  your-0.3
    tibrain_rag: 0.7
  technical:
    notion_ai: 0.2
    tibrain_rag: 0.8
```

---

## 🎯 Final Architecture Recommendation

```
┌─────────────────────────────────────────────┐
│                 Obsidian Vault               │
│           (User writes/organizes)               │
└───────────────┬─────────────────────────────────┘
                │
                │ (sync for storage)
                ↓
        ┌───────────┴────────────────┐
        │      Notion Database    │
        │    (storage + Notion AI)   │
        └───────┬──────────────────────┘
                │
                │
        ┌───────┴────────────────┬──────────┐
        │                      │          │
        ↓                      ↓          ↓
┌──────────────┐    ┌─────────────┐  ┌────────────┐
│   Notion AI   │    │  Ti Brain   │  │  Obsidian  │
│   Query       │    │  RAG Query  │  │  Storage   │
│   (via         │    │  (existing)  │
│  Manager)    │    │             │
└──────┬──────┘    └─────┬──────┘  └────────────┘
       │                │
       └────────────────┘
                ↓
        ┌────────────────────────┐
        │  Hybrid Search Engine  │
        │  (Merge + Rerank)     │
        └──────┬─────────────────┘
               ↓
        ┌────────────────────────┐
        │  Enhanced Query Result   │
        │  (answer + sources)    │
        └────────────────────────┘
```

---

## 💡 Key Insights from Critical Thinking

### Insight 1: Notion AI Changes Everything
- Notion AI is a powerful query engine built-in
- Using Notion AI + Ti Brain RAG creates hybrid search
- Value proposition shifts from "sync" to "AI-powered retrieval"

### Insight 2: Source Weighting is Critical
- Different query types need different source priorities
- Technical queries need Ti Brain more (broader knowledge)
- Current note queries need Notion AI more (local context)
- Dynamic weighting improves relevance

### Insight 3: Storage Still Needed
- Notion AI needs data to query
- Obsidian → Notion sync still required
- But sync becomes maintenance task, not core value
- Focus shifts from sync to query enhancement

### Insight 4: Overcoming Notion AI Limitations
- Notion AI limited to database content
- Ti Brain RAG provides external knowledge
- Hybrid approach overcomes both limitations
- Result: better than either alone

---

## 🎯 New Value Proposition

### Before (Simple Sync):
- **Value**: Backup, collaboration
- **AI**: None
- **Query**: Manual search only

### After (Hybrid RAG):
- **Value**: **Enhanced information retrieval**
- **AI**: Notion AI + Ti Brain RAG + hybrid reranking
- **Query**: **AI-powered with dual knowledge sources**

---

## 🚨 Risks & Mitigation

### Risk 1: Notion AI Context Window
- **Risk**: Notion AI has limited context
- **Mitigation**: Ti Brain RAG provides additional context

### Risk 2: Result Quality
- **Risk**: Poor merging/reranking
- **Mitigation**: Test and tune cross-encoder model, manual review

### Risk: Notion AI Rate Limits
- **Risk**: Query throttling
- Mitigation: Notion Manager account pooling, caching

### Risk: Integration Complexity
- **Risk**: Three systems = three failure points
- **Mitigation**: Graceful degradation (fallback to single source)

---

## ✅ Recommendation

### **Primary Approach: Hybrid RAG System**

**Architecture**:
1. Obsidian → Notion sync (storage layer)
2. Notion AI queries (via Notion Manager)
3. Ti Brain RAG queries (existing system)
4. Hybrid search engine (merge + rerank)
5. Enhanced answer synthesis

**Timeline**: 4 weeks (Sprint 1-4)
**Effort**: 60-80 hours
**Value**: High (AI-powered knowledge retrieval)

---

## 🎯 Next Steps

**Option A**: Implement Hybrid RAG System (Recommended)
- Sprint 1-4 plan as outlined
- Enhanced query capabilities
- AI-powered knowledge retrieval

**Option B**: Start with Notion AI + simple Ti Brain
- Just add Notion AI first
- Basic result merging
- Iterate from there

**Option C**: Prototype concept first
- Quick POC of hybrid search
- Test with sample data
- Validate value proposition

**Option D**: Your specific idea/concern

Which approach would you like to proceed with? This fundamentally changes the scope from "sync" to "AI-powered knowledge retrieval system".