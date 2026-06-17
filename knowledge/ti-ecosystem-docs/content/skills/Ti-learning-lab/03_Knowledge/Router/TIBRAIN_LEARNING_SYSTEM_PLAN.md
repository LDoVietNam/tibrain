# TiBrain Learning System — Comprehensive Plan

> **Version**: 1.0.0  
> **Date**: 2026-04-29  
> **Status**: Plan Phase (Research-backed)  
> **Author**: devin (BEADS Protocol)

---

## 0. Research Findings

### Repos/Projects đã khảo sát

| Repo | Stars | Pattern học được |
|------|-------|-----------------|
| **mem0ai/mem0** | 54.4k | Memory layer 4 loại: semantic + episodic + procedural + temporal. Extract facts từ conversations → store dưới dạng atomic facts, không phải raw text. Smart deduplication. |
| **IAAR-Shanghai/Awesome-AI-Memory** | 799 | AI memory taxonomy: working memory (in-context) → episodic (session logs) → semantic (knowledge base) → procedural (skill patterns). Retrieval phải là hybrid: keyword + semantic. |
| **recallium-ai/recallium** | 33 | **Memory types cho developer**: decisions, debug sessions, research, in-progress work, rules, code patterns. Projects là unit of context (scoped, isolated). Session continuity tự động. |
| **sanonone/kektordb** | 70 | Go-based AI memory: vector search + temporal knowledge graph. Memory decay + contradiction detection + MCP integration. |
| **anhmtk/StillMe-Learning-AI-System-RAG** | 6 | **Self-evolving RAG**: 19-validator chain, continuous learning từ multiple sources (6x/day), validation-first (biết khi nào KHÔNG biết). |
| **Fr-e-d/GAAI-framework** | 134 | **Governance pattern**: Discovery (plan) vs Delivery (execute) isolated contexts. Cross-session memory qua `.gaai/project/contexts/memory/`. Backlog là contract. |

### Key Patterns từ Industry

**1. mem0 Architecture (Production-proven)**
```
Input text → LLM extract facts → Vector search existing memories
→ LLM decide: ADD / UPDATE / DELETE / NOOP → Store atomic facts
→ On retrieval: semantic search → re-rank by relevance + recency
```

**2. 4-Layer Memory Stack (Recallium + IAAR)**
```
L0: Identity (~50-100 tokens) — luôn load
L1: Essential Story (~500-800 tokens) — top weighted memories  
L2: On-demand (~200-500 tokens) — context-filtered
L3: Deep Search (unlimited) — full semantic search
```
→ **TiBrain đã có** layer này trong `memory.Palace`!

**3. StillMe Validation Chain**
```
Input → Fetch context → Validate (19 layers) → Score confidence
→ KNOWN/UNCERTAIN/UNKNOWN state → Response + citations
```

**4. GAAI Cross-session Memory**
```
.gaai/project/contexts/memory/
├── decisions.md     # Architectural decisions + rationale
├── patterns.md      # Recurring code patterns observed
├── context.md       # Current sprint + pending work
└── failures.md      # What didn't work + why
```

---

## 1. Vấn Đề Hiện Tại của TiBrain

### 1.1 3 Vòng Lặp Học Đang Bị Đứt

```
[Vòng 1] beads.md → markdown_parser → Entry ✅ XONG
                                             ↓
                                    ProcessBEADSEntry ❌ CHƯA GỌI
                                             ↓
                                      memory.Palace ❌ KHÔNG KẾT NỐI

[Vòng 2] Router requests → BEADS JSONL ❌ ROUTER KHÔNG LOG
                                    ↓
                             brain engine ❌ CHƯA WIRED

[Vòng 3] tool_usage_log → analytics ❌ KHÔNG CÓ API
```

### 1.2 Nguyên Nhân Gốc Rễ

- **Không có Orchestrator**: Không ai gọi `ProcessBEADSEntry` sau khi parse
- **Memory không có index**: Palace có drawers nhưng search chỉ là keyword, không có semantic
- **Skill learning bị passive**: Tool usage log ghi nhưng không feedback lại tool scores
- **BEADS chỉ là human-written**: Thiếu machine-generated entries với model/cost/latency
- **Không có feedback loop**: Ti không biết skill nào hoạt động tốt trong context nào

### 1.3 So Sánh với Industry

| Feature | mem0 | recallium | TiBrain hiện tại |
|---------|------|-----------|-----------------|
| Fact extraction từ logs | ✅ LLM-based | ✅ Pattern-based | ❌ |
| Cross-session memory | ✅ | ✅ | ❌ |
| Memory deduplication | ✅ | Partial | ❌ |
| Semantic search | ✅ | ✅ | ❌ (chỉ keyword) |
| Learning từ failures | ✅ | ✅ | ❌ |
| Skill performance tracking | N/A | N/A | ✅ (ModelPerformance) |
| Skill recommendation | N/A | N/A | ❌ |
| Real-time ingestion | ✅ | ✅ | ❌ |

---

## 2. Mục Tiêu

**TiBrain Learning System** = hệ thống Ti tự học từ mọi hoạt động để:

1. **Ghi nhớ**: Mọi quyết định, pattern, lỗi qua sessions
2. **Cải thiện**: Tự nâng quality score của skills dựa trên usage outcomes
3. **Đề xuất**: Recommend đúng skill/model cho từng context
4. **Tự biết giới hạn**: Biết khi nào không chắc (StillMe philosophy)

---

## 3. Architecture: TiBrain Learning System v2

```
┌─────────────────────────────────────────────────────────────┐
│                    DATA SOURCES (Đầu vào)                   │
├──────────────────┬───────────────────┬──────────────────────┤
│  beads.md        │  Router JSONL     │  tool_usage_log      │
│  (human logs)    │  (machine logs)   │  (TiBrain DB)        │
│  Markdown parser │  beads.Logger     │  SQL queries         │
└────────┬─────────┴─────────┬─────────┴──────────┬───────────┘
         │                   │                    │
         ▼                   ▼                    ▼
┌─────────────────────────────────────────────────────────────┐
│               INGESTION LAYER (Chuẩn hóa)                   │
│                                                             │
│  beads.Entry (ML-grade schema v2):                          │
│  Features: task_type, phase, domain, complexity, budget     │
│  Decision: model, agent, reasoning, alternatives            │
│  Outcome: quality, cost, latency, success, tokens           │
└─────────────────────────┬───────────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────────┐
│               LEARNING ENGINE (Xử lý)                       │
│                                                             │
│  ┌──────────────────┐  ┌──────────────────────────────────┐ │
│  │ beadslearn.Learn │  │  Fact Extractor                  │ │
│  │                  │  │  (inspired by mem0)               │ │
│  │ - ModelPerf stats│  │  - Extract atomic facts          │ │
│  │ - Quality trends │  │  - Lessons learned               │ │
│  │ - Best model API │  │  - Failure patterns              │ │
│  └────────┬─────────┘  └───────────────┬──────────────────┘ │
│           │                            │                     │
│           ▼                            ▼                     │
│  ┌──────────────────────────────────────────────────────┐   │
│  │              Skill Intelligence                       │   │
│  │  - Skill usage counter per context                   │   │
│  │  - Skill co-occurrence matrix → composition hints    │   │
│  │  - Dynamic quality score (usage-based, not static)   │   │
│  └──────────────────────┬───────────────────────────────┘   │
└───────────────────────────────────────────────────────────── ┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────────┐
│                 MEMORY LAYER (Lưu trữ)                      │
│                                                             │
│  memory.Palace (4-Layer Stack):                             │
│  L0: Ti Identity (ai Ti là, mục tiêu)                       │
│  L1: Essential Story (top decisions, recurring patterns)    │
│  L2: On-demand (project/domain-filtered)                    │
│  L3: Deep Search (full semantic — future: embeddings)       │
│                                                             │
│  + beads.Store (SQLite) — raw ML training data             │
│  + model-stats.json — model performance history            │
│  + skill-stats.json — skill usage + quality history        │
└─────────────────────────┬───────────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────────┐
│                   API LAYER (Truy vấn)                      │
│                                                             │
│  TiBrain REST API (port 1810):                              │
│  GET /v1/brain/stats          → Tổng hợp learning stats    │
│  GET /v1/brain/best-model     → Best model cho task_type   │
│  GET /v1/brain/route          → Full routing suggestion     │
│  GET /v1/skills/recommend     → Skill recommendation       │
│  GET /v1/skills/compose       → Skill composition hints    │
│  POST /v1/brain/ingest        → Manual data ingestion      │
└─────────────────────────────────────────────────────────────┘
```

---

## 4. Implementation Plan

### Phase 1: Wiring (Ưu tiên CAO — 2 ngày)

**Mục tiêu**: Làm cho learning pipeline chạy được với 170 beads.md entries hiện có.

#### Task 1.1 — `LearnFromMarkdown()` function
**File**: `Z:\Ti\core\pkg\beadslearn\markdown_parser.go`

```go
// LearnFromMarkdownFile parses beads.md và feed vào learning engine.
// Đây là bridge giữa human-written logs và ML learning pipeline.
func LearnFromMarkdownFile(beadsPath string, learn *Learn) (int, int, error) {
    entries, err := ParseMarkdownFile(beadsPath)
    // batch process all entries
    // skip entries với status "running"
    // return (processed, skipped, error)
}
```

**Chuẩn Learn**: Chỉ process entries có status DONE/failed (complete/failed)

#### Task 1.2 — `bd learn` CLI Command
**File mới**: `Z:\Ti\cmd\beads\learn.go`

```
bd learn --from=Z:\Ti\taskboard\beads.md [--dry-run] [--verbose]
```

Output:
```
Parsing beads.md... 170 entries found
Processing: 165 completed, 5 running (skipped)
Model stats updated: devin (165 tasks, 100% success, avg quality 0.95)
Memory drawers added: 23 new patterns stored
Saved: Z:\Ti\data\model-stats.json
Done in 0.3s
```

#### Task 1.3 — Wire Palace vào TiBrain server
**File**: `Z:\Ti\router\tibrain-server\main.go`

```go
// Trong main():
palace, _ := memory.NewPalace(cfg.DataDir)
learn := beadslearn.NewLearn(palace, beadsPath)

// Background goroutine: watch beads.md cho new entries
go watchBeadsFile(beadsPath, learn)
```

#### Task 1.4 — Brain Stats API
**File**: `Z:\Ti\router\tibrain-server\main.go`

```
GET /v1/brain/stats
Response: {
  "total_entries": 170,
  "success_rate": 0.97,
  "avg_quality": 0.95,
  "models": { "devin": {...}, "claude": {...} },
  "top_task_types": [...],
  "quality_trend": +0.05
}

GET /v1/brain/best-model?task_type=coding
Response: {
  "model": "claude-sonnet-4-5",
  "avg_quality": 0.92,
  "sample_count": 45,
  "confidence": "high"
}
```

**Deliverable**: `bd learn --from=beads.md` chạy được, stats accessible qua API

---

### Phase 2: Real-time Learning (3 ngày)

**Mục tiêu**: Router tự log mọi request vào BEADS, TiBrain học continuous.

#### Task 2.1 — Router BEADS Logging
**File**: `Z:\Ti\router\` (xác định middleware layer)

Mỗi request qua router tạo 1 BEADS entry:
```go
entry := beads.Entry{
    Task:        req.Messages[last].Content[:100],
    Model:       selectedModel,
    Agent:       req.AgentID,
    TaskType:    inferTaskType(req),
    Domain:      inferDomain(req),
    Complexity:  estimateComplexity(req),
    // Filled after response:
    Quality:     0, // Updated after response
    LatencyMs:   elapsed.Milliseconds(),
    Cost:        calculateCost(usage),
    PromptTokens: usage.PromptTokens,
    CompletionTokens: usage.CompletionTokens,
    Success:     err == nil,
}
```

#### Task 2.2 — BEADS JSONL → SQLite Sync
**File**: `Z:\Ti\router\tibrain-server\main.go`

```go
// Background goroutine: watch beads.jsonl, batch insert vào beads.db
func syncBEADSStore(jsonlPath string, store *beads.Store) {
    // tail file, parse new lines, insert to SQLite
    // dedup by timestamp+session_id
}
```

#### Task 2.3 — `/v1/brain/route` Smart Routing Endpoint
```
POST /v1/brain/route
Body: {
  "task_type": "coding",
  "domain": "backend",
  "complexity": 0.7,
  "budget": 0.05,
  "phase": "implement"
}
Response: {
  "recommended_model": "claude-sonnet-4-5",
  "reasoning": "Best quality (0.92) for coding/backend with 45 data points",
  "alternatives": [
    {"model": "gemini-2-flash", "avg_quality": 0.85, "cost_savings": "60%"}
  ],
  "confidence": 0.87
}
```

**Logic**: `GetBestModel()` + cost-aware fallback nếu không đủ data

**Deliverable**: Router tự log + TiBrain có routing recommendations dựa trên history

---

### Phase 3: Skill Intelligence (3-4 ngày)

**Mục tiêu**: Từ usage patterns học skill nào tốt trong context nào, recommend tự động.

#### Task 3.1 — Skill Usage Tracking Schema
**File**: `Z:\Ti\router\tibrain-server\main.go` (migrateDatabase)

```sql
CREATE TABLE IF NOT EXISTS skill_usage (
    id           INTEGER PRIMARY KEY AUTOINCREMENT,
    skill_id     TEXT NOT NULL,
    skill_name   TEXT NOT NULL,
    agent_id     TEXT NOT NULL,
    task_type    TEXT NOT NULL,
    domain       TEXT NOT NULL,
    project      TEXT NOT NULL,
    quality      REAL NOT NULL DEFAULT 0,
    success      INTEGER NOT NULL DEFAULT 1,
    timestamp    INTEGER NOT NULL,
    session_id   TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS skill_cooccurrence (
    skill_a      TEXT NOT NULL,
    skill_b      TEXT NOT NULL,
    count        INTEGER NOT NULL DEFAULT 1,
    avg_quality  REAL NOT NULL DEFAULT 0,
    PRIMARY KEY (skill_a, skill_b)
);
```

#### Task 3.2 — Dynamic Skill Scoring
Thay vì static `quality_score` trong `tool_registry`, tính dynamic:

```go
// DynamicSkillScore = weighted average:
// - 70% usage-based quality (từ skill_usage table)
// - 30% curated quality (từ tool_registry)
// Decay factor: recent usage có weight cao hơn (half-life 30 ngày)
```

#### Task 3.3 — Skill Recommendation API
```
GET /v1/skills/recommend?task_type=django&domain=backend&project=my-api
Response: {
  "skills": [
    {
      "name": "django-tdd",
      "score": 0.95,
      "reason": "Used 12x for django/backend, 100% success rate",
      "compose_with": ["python-testing", "django-security"]
    },
    {
      "name": "django-security",
      "score": 0.88,
      "reason": "Co-occurs with django-tdd in 8/12 sessions"
    }
  ]
}
```

#### Task 3.4 — Co-occurrence Matrix Builder
Mỗi session kết thúc với quality ≥ 0.7:
- Record tất cả skills được dùng trong session
- Update `skill_cooccurrence` table
- Tính composition score

#### Task 3.5 — Fact Extractor (inspired by mem0)
Từ mỗi BEADS entry có `Lessons Learned` section:
```go
func extractFacts(entry beads.Entry, lessonsText string) []Fact {
    // Parse bullet points từ Lessons Learned
    // Convert thành atomic facts
    // Deduplicate với existing Palace drawers
    // Store với source = entry.SessionID
}
```

**Deliverable**: `/v1/skills/recommend` trả về contextual skill list với evidence

---

### Phase 4: Self-awareness & Validation (tương lai)

Inspired by **StillMe validation chain**:

- **Confidence scoring**: Biết khi nào recommendation có ít data → trả `confidence: low`
- **Contradiction detection**: Khi skill A được ghi nhận cả tốt lẫn xấu → flag conflict
- **Epistemic states**: `KNOWN` (≥10 data points, consistent) / `UNCERTAIN` (3-9 points) / `UNKNOWN` (<3 points)
- **Hallucination prevention**: Không recommend skill nếu chưa có evidence

---

## 5. Data Flow Chi Tiết

```
beads.md entry (human-written)
    │
    ▼ markdown_parser.go
beads.Entry struct
    │
    ├──► beadslearn.ProcessBEADSEntry()
    │        ├──► updateModelStats() → model-stats.json
    │        ├──► scoreDrawers() → memory.Palace
    │        └──► storeQADrawer() → memory.Palace (quality ≥ 0.7)
    │
    └──► beads.Store.Insert() → beads.db (SQLite)
                                    │
                                    ▼
                              Analytics queries
                              (stats, trends, best-model)

Router request
    │
    ▼ (Phase 2) beads.Logger
beads.Entry (auto-generated)
    │
    ▼ background sync
beads.db → learn engine → same flow as above
```

---

## 6. File Structure

```
Z:\Ti\
├── core\
│   └── pkg\
│       └── beadslearn\
│           ├── learn.go               ✅ Có - ModelPerformance tracking
│           ├── markdown_parser.go     ✅ Có - Parse beads.md
│           ├── orchestrator.go        ❌ CẦN TẠO - Wires everything
│           └── fact_extractor.go      ❌ CẦN TẠO - Extract atomic facts
├── cmd\
│   └── beads\
│       └── learn.go                   ❌ CẦN TẠO - CLI command
└── router\
    └── tibrain-server\
        ├── main.go                    ✅ Có - TiBrain server (cần thêm endpoints)
        ├── learning.go                ❌ CẦN TẠO - Learning integration
        └── skill_intelligence.go      ❌ CẦN TẠO - Skill recommendation
```

---

## 7. Success Metrics

| Metric | Phase 1 | Phase 2 | Phase 3 |
|--------|---------|---------|---------|
| Entries processed | 170 (beads.md) | +realtime | continuous |
| `bd learn` command | ✅ | - | - |
| `/v1/brain/stats` | ✅ | ✅ enriched | ✅ |
| `/v1/brain/best-model` | ✅ | ✅ more data | ✅ |
| `/v1/brain/route` | ❌ | ✅ | ✅ |
| `/v1/skills/recommend` | ❌ | ❌ | ✅ |
| Skill co-occurrence | ❌ | ❌ | ✅ |
| Atomic fact extraction | ❌ | ❌ | ✅ |

---

## 8. Lessons Learned từ Research

1. **mem0 insight**: Đừng store raw text — extract atomic facts. "Task X succeeded with quality 0.9 using model Y" → cô đọng hơn nhiều.
2. **recallium insight**: Projects là unit of context. Memory nên scoped theo project, không global.
3. **StillMe insight**: Biết khi nào KHÔNG biết quan trọng hơn luôn có answer. Trả `confidence: unknown` tốt hơn hallucinate.
4. **GAAI insight**: Cross-session memory phải explicit, không implicit. Write decisions.md, patterns.md sau mỗi session.
5. **kektordb insight**: Temporal decay quan trọng. Lessons từ 6 tháng trước ít relevant hơn lessons từ tuần trước.

---

## 9. Risks & Mitigations

| Risk | Probability | Impact | Mitigation |
|------|-------------|--------|------------|
| beads.md parsing errors | Low | Medium | Parser đã tested với 170 entries |
| Memory Palace không scale | Medium | Medium | SQLite cho <100k entries là ổn, sau đó migrate |
| Skill scoring không accurate | Medium | High | Start với 70/30 weighted blend, tune theo feedback |
| Router logging overhead | Low | Low | Async goroutine, không block request path |
| Circular learning (bad data → bad model) | Low | High | Min 3 data points, validation chain trước khi update |

---

## 10. Next Immediate Actions

Theo BEADS protocol, implement theo thứ tự:

1. **[NOW]** Task 1.1: `LearnFromMarkdownFile()` + Task 1.2: `bd learn` CLI
2. **[Day 2]** Task 1.3: Wire Palace vào TiBrain + Task 1.4: Brain Stats API
3. **[Day 3-5]** Phase 2: Router logging + Smart routing
4. **[Day 6-9]** Phase 3: Skill intelligence

---

*Research sources: mem0ai/mem0 (54.4k⭐), IAAR-Shanghai/Awesome-AI-Memory (799⭐), recallium-ai/recallium (33⭐), sanonone/kektordb (70⭐), anhmtk/StillMe-RAG (6⭐), Fr-e-d/GAAI-framework (134⭐)*
