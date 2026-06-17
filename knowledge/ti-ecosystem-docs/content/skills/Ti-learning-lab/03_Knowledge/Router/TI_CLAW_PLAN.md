# Ti Claw Plan - Agent Framework

> **Purpose**: Kế hoạch cho Ti Claw (rebrand GoClaw cho cá nhân) — agent framework
> **Date**: 2026-05-05 (v4 — tách khỏi Router/CLI plans)
> **Status**: Planning
> **Usage**: Personal use only — không lo license, không deadline cứng
> **Strategy**: **Rebrand GoClaw → Ti Claw** + reuse triệt để Router/CLI/MCP đã có
> **Scope**: Chỉ focus Ti Claw. Router và CLI tách riêng.
> **Sister Plans**:
> - `ROUTER_PLAN.md` — Router completion (LLM gateway)
> - `TI_CLI_PLAN.md` — Ti CLI + multi-CLI orchestration

---

## Table of Contents

1. [Executive Summary](#1-executive-summary)
2. [Framework Selection](#2-framework-selection)
3. [Fork vs Rewrite Decision](#3-fork-vs-rewrite-decision)
4. [Ti Ecosystem Assets Reuse](#4-ti-ecosystem-assets-reuse)
5. [Patterns Extraction](#5-patterns-extraction)
6. [Multi-Framework Learning](#6-multi-framework-learning)
7. [Architecture Design](#7-architecture-design)
8. [Integration Plan](#8-integration-plan)
9. [UI Approach](#9-ui-approach)
10. [Anthropic Tools Integration](#10-anthropic-tools-integration)
11. [Task Management Migration](#11-task-management-migration)
12. [Roadmap Reanalysis](#12-roadmap-reanalysis)
13. [Implementation Timeline](#13-implementation-timeline)
14. [Success Criteria](#14-success-criteria)
15. [Key Risks & Mitigations](#15-key-risks--mitigations)
16. [References](#16-references)

---

## 1. Executive Summary

### Overview
Ti Claw là framework agent **cá nhân** cho Ti ecosystem, rebrand từ GoClaw. Vì là personal use, chiến lược tập trung vào: **rebrand thẳng tay, customize thoải mái, tái sử dụng tối đa Router/CLI/MCP đã có**.

### Quan trọng: Ti CLI ĐÃ CÓ NHIỀU MODULE AGENT
Sau khi đọc `cli-context.json`, phát hiện Ti CLI đã có:
- ✅ `agentcore`, `agent`, `routeagent` — Agent execution core
- ✅ `brain` — Learning engine với RL
- ✅ `ticore` — Plan/Verify
- ✅ `planner`, `taskinput` — Planning + task input
- ✅ `beads`, `beadsgraph`, `beadslearn` — BEADS learning
- ✅ `verify`, `repair` — Verification + auto-repair
- ✅ `mcp` — MCP server bridge
- ✅ Microkernel + Plugin architecture (gRPC)

→ **Ti Claw có thể là plugin trong Ti CLI**, không nhất thiết phải là app riêng!

### Key Decisions (v3 Refined)
1. **Framework**: Rebrand GoClaw → Ti Claw (personal, không lo license)
2. **Edition**: SQLite-only (như GoClaw Lite)
3. **Architecture choice** (cần quyết định):
   - **Option A**: `apps/ticlaw/` — app riêng, port 18790
   - **Option B**: `apps/cli/internal/plugins/ticlaw/` — plugin trong CLI (recommended)
4. **LLM Provider**: Ti Router (:1807) — đã có 28 providers + 12 OAuth, không cần redo
5. **MCP Tools**: Reuse 24+ tools trong `apps/mcp/`
6. **Memory**: Reuse CLI's `memory` module (Palace memory) + knowledge-graph.jsonl
7. **Brain/Learning**: Reuse CLI's `brain` + `beads*` modules thay vì build mới
8. **Plan/Verify**: Reuse CLI's `ticore`, `planner`, `verify`, `repair`
9. **UI**: Extend `apps/router/ui/` (đã có React, vừa build)
10. **Task Management**: Migrate BD → Notion qua `notion_manager`

### Related Plans

Plan này chỉ focus **Ti Claw agent framework**. Router và CLI completion có plans riêng:
- **`ROUTER_PLAN.md`** — Router LLM gateway completion (UI polish, backend fixes, performance)
- **`TI_CLI_PLAN.md`** — Ti CLI + multi-CLI orchestration (registry, sub-agents, parallel execution)

### No Hard Deadlines
- Personal project → ưu tiên chất lượng > tốc độ
- Mỗi phase chỉ proceed khi phase trước stable
- Parallel work across plans OK, xem "Cross-Plan Coordination" trong Section 13

---

## 2. Framework Selection

### Research Results

| Framework | Stars | Pros | Cons | Verdict |
|-----------|-------|------|------|---------|
| **LangChain Go** | 9.2k | LLM library phổ biến | Chỉ là LLM library, không có agent loop | ❌ Không phù hợp |
| **Blades** | 763 | Framework đầy đủ, middleware ecosystem | Quá complex, không phù hợp CLI | ❌ Không phù hợp |
| **Orloj** | 84 | Full-stack orchestration, governance | Quá nặng cho Ti use case | ❌ Không phù hợp |
| **Cordum** | 470 | Governance-focused, safety kernel | Không phải agent framework | ❌ Không phù hợp |
| **Autopus-ADK** | 97 | Multi-agent orchestration | Chưa mature | ❌ Không phù hợp |
| **GoClaw** | - | Đầy đủ patterns, đã có trong learning lab | Enterprise-grade, cần simplify | ✅ **CHỌN** |

### Why GoClaw?
- Đầy đủ patterns (Agent Loop, Store, Tools, Providers, Memory, Config)
- Đã có trong learning lab → Dễ research và extract patterns
- Mature, production-use
- Dễ customize cho Ti use case
- Documentation chi tiết (CLAUDE.md)
- **Có sẵn Lite edition (SQLite + Wails)** → pattern hay cho Ti desktop future
- **Multi-provider ready** (Anthropic, OpenAI, Claude CLI, DashScope, Codex)

---

## 3. Fork vs Rewrite Decision

### Decision: **FORK GoClaw**

### Comparison

| Criterion | Fork GoClaw | Rewrite from Scratch |
|-----------|-------------|----------------------|
| **Time to first agent** | 2-3 tuần | 6-8 tuần |
| **Production-ready code** | ✅ Ngay từ đầu | ❌ Phải hardening lại |
| **Upstream updates** | ✅ Có thể cherry-pick | ❌ Không |
| **Edge cases** | ✅ GoClaw đã xử lý | ❌ Phải tự phát hiện |
| **Learning opportunity** | ⚠️ Trung bình | ✅ Cao |
| **Customization** | ⚠️ Cần strip down | ✅ Hoàn toàn tự do |
| **Risk** | Low | High |

### Fork Strategy

```
1. Clone GoClaw → apps/ticlaw/
2. Rename package: github.com/GoClaw/... → ti/apps/ticlaw/...
3. Strip modules không dùng phase đầu:
   - internal/channels/ (Telegram, Feishu, Zalo, Discord, WhatsApp)
   - internal/tts/ (TTS providers)
   - internal/oauth/ (nếu không cần)
   - internal/knowledgegraph/ (dùng Ti's existing)
   - internal/permissions/ (Lite edition không cần RBAC)
4. Config build: //go:build sqliteonly
5. Keep core: agent/, scheduler/, tools/, providers/, memory/, mcp/, skills/
6. Add Ti-specific:
   - internal/providers/tirouter.go — Ti Router provider
   - internal/mcp/tibridge.go — Bridge tới Ti MCP servers
7. Track upstream via git remote cho cherry-pick updates
```

### Upgrade Path
- Giữ `.upstream-goclaw` reference commit
- Khi GoClaw release patterns mới (vd: safety kernel) → cherry-pick
- Document các file đã fork khỏi upstream

---

## 4. Ti Ecosystem Assets Reuse

### Ti Router — Đã có 28 providers + 12 OAuth

**Path**: `apps/router/` (port 1807)
**Module**: `github.com/ti/router`, Go 1.25

| Capability | Status | GoClaw equivalent |
|-----------|--------|-------------------|
| OpenAI-compat API | ✅ `/v1/chat/completions`, `/v1/messages`, `/v1/embeddings` | `internal/providers/openai.go` |
| 28 Providers | ✅ gemini, openai, claude, perplexity, groq, windsurf... | `internal/providers/` |
| 12 OAuth | ✅ anthropic, codex, gemini, qwen, github, cursor, kiro... | `internal/oauth/` |
| Cache | ✅ `resilience.Cache` + prompt cache | `internal/cache/` |
| Rate Limit | ✅ Per-account, per-provider | — |
| Circuit Breaker | ✅ `routing.HealthMonitor` | — |
| Routing | ✅ Semantic + adaptive routing, model selection | — |
| RTK Compression | ✅ Token compression | — |
| Memory Store | ✅ Routing decision storage | `internal/memory/` |
| Decision Engine | ✅ Adaptive routing | — |
| Audit Logging | ✅ Compliance audit | `internal/tracing/` |
| Session Tracking | ✅ Fingerprinting | `internal/sessions/` |

→ **Ti Claw KHÔNG CẦN providers/oauth riêng** — point thẳng vào Router :1807

### Ti CLI — Đã có agent execution core

**Path**: `apps/cli/` (74% migrated, 43/58 modules)
**Module**: `github.com/ti/cli`
**Architecture**: Microkernel + Plugin (gRPC)

| Module | Status | Mục đích | Reuse strategy |
|--------|--------|----------|----------------|
| `agentcore` | ✅ | Agent execution core | **REUSE** thay GoClaw `internal/agent/` |
| `agent` | ✅ | Agent execution loop | **REUSE** |
| `routeagent` | ✅ | Routing engine + AgentMux | **REUSE** |
| `brain` | ✅ | Learning engine với RL | **REUSE** thay vì rewrite |
| `ticore` | ✅ | Plan/Verify | **REUSE** thay GoClaw planner |
| `planner` | ✅ | Planning system | **REUSE** |
| `taskinput` | ✅ | Task input processing | **REUSE** |
| `verify` | ✅ | Verification system | **REUSE** |
| `repair` | ✅ | Auto-repair | **REUSE** |
| `beads` | ✅ | BEADS logging | **REUSE** |
| `beadsgraph` | ✅ | Task graph viz | **REUSE** |
| `beadslearn` | ✅ | BEADS learning | **REUSE** |
| `qa`, `qalog` | ✅ | QA system | **REUSE** |
| `mcp` | ✅ | MCP server bridge | **REUSE** |
| `memory` | ✅ | Palace memory | **REUSE** thay GoClaw memory |
| `session` | ✅ | Session management | **REUSE** |
| `permission` | ✅ | Permission system | **REUSE** |
| `safety` | ✅ | Safety checks | **REUSE** |
| `tools` | ✅ | File tools | **REUSE** |
| `commitflow` | ✅ | Git commit flow | **REUSE** |
| `optimize` | ✅ | Code optimization | **REUSE** |
| `patch` | ✅ | Patch management | **REUSE** |
| `automation` | ✅ | Automation framework | **REUSE** |
| `cookie` | ✅ | Browser cookies | **REUSE** |
| `learn` | ✅ | Learning engine (16 files) | **REUSE** |

→ **Ti CLI đã có gần như đầy đủ infrastructure cho agent**. Ti Claw chỉ cần thêm:
- Agent loop pattern từ GoClaw (think→act→observe formalization)
- Skill system (SKILL.md + BM25)
- Lane-based scheduler
- Bootstrap (SOUL.md, IDENTITY.md)

### Other Assets

| Asset | Path | Reuse |
|-------|------|-------|
| **Ti Router UI** | `apps/router/ui/` (React 19, Vite 6) | UI base — extend thêm /agents, /tools |
| **Ti MCP Hub** | `apps/mcp/` (24+ tools) | MCP bridge target |
| **notion_manager** | `notion_manager/` (port 8081) | Notion API gateway |
| **Knowledge Graph** | `Z:/03_DATA/ti/memory/knowledge-graph.jsonl` | Memory backend |
| **TiBrain** | `apps/mcp/server/tibrain-mcp/` (Python) | Handoff + skills tools |
| **Ti Skills** | `content/skills/` (100+) | Skill library |
| **Ti Agents** | `content/agents/` (100+) | Agent library |
| **Ti Rules** | `content/rules/` (13 P0/P1) | Agent guidelines |

### Critical Insight

**Ti CLI đã 74% có một agent framework rồi**. Câu hỏi quan trọng:

1. **Có nên build Ti Claw separate không**, khi CLI đã có hầu hết?
2. **Hay merge GoClaw patterns vào CLI** — chỉ port những phần CLI thiếu?

**Recommendation**: Option B (Plugin-based)
- Ti Claw = plugin trong CLI's plugin system
- Reuse CLI's brain/agent/memory/beads
- Thêm GoClaw-specific patterns: agent loop formalization, skill system, scheduler
- CLI hiện tại đã có Plugin Loader, gRPC bus → easy integrate

---

## 5. Patterns Extraction

### 16 Pattern Categories

#### 1. Core Architecture Patterns
- **Agent Loop Pattern** (Think-Act-Observe)
- **Store Layer Pattern** (Repository Pattern)
- **Tool Registry Pattern**
- **Provider Pattern** (Adapter Pattern)
- **Event Bus Pattern**

#### 2. Configuration Patterns
- **Config Loading** (JSON5 + Env Overlay)
- **Agent Types** (open/predefined)
- **Context Files** (agent-level + per-user)

#### 3. Scheduling Patterns
- **Lane-based Concurrency**
- **Cron Scheduling**

#### 4. Memory Patterns
- **Memory System** (pgvector)
- **Session Management**
- **Auto-summarization**

#### 5. Security Patterns
- **Input Guard**
- **RBAC** (Role-Based Access Control)

#### 6. Integration Patterns
- **MCP** (Model Context Protocol)
- **Channel Integration**

#### 7. Bootstrap Patterns
- **Context File Seeding**
- **Skill Loading** (BM25 search)

#### 8. Internationalization (i18n)
- **Backend**: Message catalog
- **Web UI**: i18next
- **Supported**: en, vi, zh

#### 9. Desktop Edition (Lite) Patterns
- **Build tag**: `//go:build sqliteonly`
- **Edition system**: Standard vs Lite
- **Secrets**: OS keyring

#### 10. Mobile UI/UX Patterns
- **Viewport height**: `h-dvh`
- **Input font-size**: 16px on mobile
- **Safe areas**: Notched devices
- **Touch targets**: ≥44px

#### 11. Migration Patterns
- **Migration files**: `migrations/`
- **Version tracking**: `internal/upgrade/version.go`

#### 12. Testing Patterns
- **Integration tests**: `tests/integration/`
- **Race detector**: `go test -race`
- **Post-implementation checklist**

#### 13. Go Conventions
- `errors.Is(err, sentinel)` instead of `err == sentinel`
- `switch/case` instead of `if/else if`
- `append(dst, src...)` instead of loop-based append
- Always handle errors

### Component Mapping

#### Keep (10 components)
- Agent Loop, Store Layer, Tool Registry, Provider Pattern
- Event Bus, Session Management, Memory System
- Config Loading, Bootstrap Pattern, Skill Loading

#### Remove (10 components)
- Multi-tenant, PostgreSQL, Web UI, Desktop UI
- Channels, RBAC, WebSocket, Cron Scheduling
- Lane-based Scheduler, OAuth

#### Tweak (8 components)
- Agent Types, Context Files, Skills, Memory
- Providers, Tools, i18n, Security

#### New (7 components)
- Ti CLI Integration, Ti Router Provider, Knowledge Graph Memory
- Notion Agent, Ti Skill System, OmniRoute Patterns, Ti Config

---

## 6. Multi-Framework Learning

### Strategy: GoClaw Core + Patterns từ các frameworks khác

GoClaw là nền tảng, nhưng có vài patterns **quan trọng** nên học từ frameworks khác:

### Frameworks to Study

| Framework | Pattern đáng học | Apply vào Ti Claw |
|-----------|------------------|-------------------|
| **LangGraph** | State machines, graph-based agent orchestration | Multi-step workflows phức tạp |
| **CrewAI** | Role specialization (Researcher, Writer, Reviewer) | Predefined agent templates |
| **AutoGen** | Conversational multi-agent | Agent-to-agent handoff |
| **Letta (MemGPT)** | Long-context memory management (tier 1/2/3) | Memory optimization khi context >75% |
| **OpenAI Swarm** | Lightweight handoff pattern | Sub-agent dispatch |
| **Claude Skills** | Skill-as-code với frontmatter | GoClaw có sẵn — dùng pattern này |
| **Anthropic Claude Code** | Superpowers, GSD, context mode | Xem Section 10 |
| **Devin** | Long-horizon planning, self-correction | Tích hợp như một tool |
| **Cursor** | File-level context selection | Inspire cho Ti CLI |

### Research Priority

1. **Letta (MemGPT)** — Priority 1: Memory tiering critical cho long session
2. **LangGraph** — Priority 2: Graph orchestration cho complex workflows
3. **CrewAI** — Priority 3: Agent role templates
4. **Claude Code** — Priority 4: Đã có analysis, xem Section 10
5. **AutoGen / Swarm** — Priority 5: Multi-agent patterns khi cần

### Integration Approach

- **Không fork frameworks khác** — chỉ extract patterns
- Document pattern vào `Ti-learning-lab/03_Knowledge/Agent/patterns/`
- Implement như **optional modules** trong Ti Claw
- Ví dụ:
  - `apps/ticlaw/internal/memory/tiered.go` — Letta-inspired tiered memory
  - `apps/ticlaw/internal/workflow/graph.go` — LangGraph-inspired state machine
  - `apps/ticlaw/internal/agents/crew.go` — CrewAI-inspired role templates

### Deliverable for Week 0

Document: `Ti-learning-lab/03_Knowledge/Agent/framework-patterns-comparison.md`
- So sánh chi tiết các patterns
- Chọn top 3-5 patterns đáng adopt
- Map vào Ti Claw modules

---

## 7. Architecture Design

### Integration Topology

```
┌─────────────────────────────────────────────────────┐
│  Ti CLI (apps/cli/)                                 │
│  - ti agent run, ti skill list, ti task log         │
└────────────────────────┬────────────────────────────┘
                         │ CLI commands
                         ▼
┌─────────────────────────────────────────────────────┐
│  Ti Claw (apps/ticlaw/) — FORK of GoClaw            │
│  Port: 18790                                        │
│  ┌──────────────────────────────────────────────┐   │
│  │ internal/agent/   — Think→Act→Observe loop   │   │
│  │ internal/scheduler/ — Lane-based (main/sub)  │   │
│  │ internal/tools/   — Tool registry            │   │
│  │ internal/skills/  — SKILL.md + BM25          │   │
│  │ internal/memory/  — SQLite + KG bridge       │   │
│  │ internal/mcp/     — MCP bridge               │   │
│  │ internal/providers/tirouter.go (NEW)         │   │
│  │ internal/bootstrap/ — Agent context          │   │
│  └──────────────────────────────────────────────┘   │
└─────┬───────────┬──────────────┬────────────────────┘
      │ Provider  │ MCP bridge   │ Skills
      ▼           ▼              ▼
┌──────────┐ ┌────────────┐ ┌────────────────────┐
│ Router   │ │ MCP Hub    │ │ content/skills/    │
│ :1807    │ │ apps/mcp/  │ │ .devin/skills/     │
│ (reused) │ │ 24+ tools  │ │ (reused)           │
└─────┬────┘ └─────┬──────┘ └────────────────────┘
      │            │
      ▼            ▼
 ┌─────────┐  ┌──────────────┐
 │ OpenAI  │  │ notion_mgr   │
 │ Anthro. │  │ :8081        │
 │ Claude  │  │ (reused)     │
 └─────────┘  └──────────────┘

Memory backend:
 - Z:/03_DATA/ti/memory/knowledge-graph.jsonl (reused)
 - apps/ticlaw/data/sqlite.db (new, agent sessions)

UI:
 - apps/router/ui/ (extended with /agents, /tools routes)
 - Served from Router :1807 (embedded)
```

### Proposed Structure

```
apps/
├── cli/                # Official CLI (existing) — add: ti agent, ti task
├── router/             # Router service (existing, :1807)
│   ├── layers/         # Router backend
│   └── ui/             # React UI (EXTENDED — add agent/task pages)
├── mcp/                # MCP Hub (existing, 24+ tools) — reuse as-is
├── ticlaw/             # Ti Claw — FORK from GoClaw
│   ├── cmd/            # CLI entry (goclaw onboard, goclaw migrate, etc.)
│   ├── main.go         # Main binary (sqliteonly build tag)
│   ├── internal/       # FORKED from GoClaw, stripped
│   │   ├── agent/      # KEEP — Agent loop
│   │   ├── scheduler/  # KEEP — Lane-based concurrency
│   │   ├── bootstrap/  # KEEP — SOUL.md, IDENTITY.md seeding
│   │   ├── config/     # KEEP — JSON5 + env overlay
│   │   ├── crypto/     # KEEP — AES-256-GCM for API keys
│   │   ├── gateway/    # KEEP — WS + HTTP
│   │   ├── http/       # KEEP — HTTP API
│   │   ├── mcp/        # KEEP + EXTEND — Bridge tới Ti MCP
│   │   ├── memory/     # KEEP + EXTEND — KG bridge
│   │   ├── providers/  # KEEP + ADD tirouter.go
│   │   ├── sessions/   # KEEP — Session management
│   │   ├── skills/     # KEEP — SKILL.md + BM25
│   │   ├── store/      # KEEP — SQLite + interface
│   │   ├── tasks/      # KEEP — Task management
│   │   ├── tools/      # KEEP — Tool registry
│   │   ├── tracing/    # KEEP (optional) — OTel tracing
│   │   ├── upgrade/    # KEEP — Schema version
│   │   ├── hooks/      # KEEP — Extensibility
│   │   └── edition/    # KEEP — Lite/Standard
│   │   ── STRIPPED (phase 1):
│   │   ── channels/ (Telegram, Feishu, Zalo, Discord, WhatsApp)
│   │   ── tts/
│   │   ── oauth/ (tùy chọn)
│   │   ── permissions/ (lite không cần)
│   │   ── knowledgegraph/ (dùng Ti's existing)
│   │   ── sandbox/ (optional later)
│   │   ── channels/
│   ├── migrations/     # KEEP — SQLite migration files
│   ├── skills/         # KEEP — GoClaw skill examples
│   └── content/        # NEW — Ti-specific agent context files
│       ├── SOUL.md
│       ├── IDENTITY.md
│       └── agents/
│           └── notion-agent.md
└── automation/         # Existing automation tools

content/                # Shared Ti content (existing)
├── skills/             # Ti skills — compatible với Ti Claw skills
├── agents/             # Agent definitions
└── mcp/                # MCP configs

packages/               # Shared libs (existing)
├── sdk/                # Plugin API SDK
└── core/               # Core modules

Z:/03_DATA/ti/memory/knowledge-graph.jsonl  # Memory (reused)
```

### Key Architecture Principles

1. **One source of truth**: Router (1807) là gateway duy nhất cho LLM
2. **MCP as contract**: Tools connect qua MCP, không qua Go interface trực tiếp
3. **Skill compatibility**: Ti Claw skills = GoClaw SKILL.md format = compatible với Anthropic Skills
4. **Lite first**: SQLite-only build, desktop-ready (Wails support future)
5. **UI consolidation**: Một UI (router/ui) cho toàn bộ Ti ecosystem, không split

---

## 8. Integration Plan

### Phase 0: Study & Decide (Week 0, 3-5 ngày)
**Goal**: Xác định chi tiết trước khi code

**Tasks:**
1. Audit GoClaw internal/ modules — xác định chính xác cái gì keep/strip
2. Research Letta, LangGraph, CrewAI patterns (xem Section 6)
3. Test clone GoClaw + build `sqliteonly` cục bộ
4. Verify Ti Router OpenAI-compat endpoint hoạt động với GoClaw provider
5. Confirm MCP bridge GoClaw có thể gọi Ti MCP servers

**Deliverables**:
- `framework-patterns-comparison.md`
- Confirmed fork strategy
- Proof-of-concept: GoClaw → Router provider

---

### Phase 1: Fork & Strip (Week 2)
**Goal**: Fork GoClaw thành `apps/ticlaw/` và strip modules không cần

**Tasks:**
1. Clone GoClaw → `apps/ticlaw/`
2. Rename Go module path
3. Strip: channels/, tts/, permissions/, knowledgegraph/, sandbox/
4. Set up `sqliteonly` build tag
5. Verify `./ticlaw onboard` chạy được
6. Setup upstream tracking: `.upstream-goclaw` ref

**Deliverables**:
- `apps/ticlaw/` running với SQLite
- Stripped modules documented
- Build chỉ cần: `go build -tags sqliteonly ./...`

---

### Phase 2: Ti Integration (Week 3)
**Goal**: Connect Ti Claw với Ti ecosystem

**Tasks:**
1. Add `internal/providers/tirouter.go` — Router as provider
2. Register Ti MCP servers vào `internal/mcp/` bridge
3. Add `internal/memory/kgbridge.go` — Knowledge graph bridge
4. Add CLI commands: `ti agent run`, `ti agent list`, `ti task log`
5. Config `.env.local` — API keys via Ti Router

**Deliverables**:
- Ti Claw agent chạy được với Router provider
- 24+ Ti MCP tools callable từ agent
- `ti agent run --name hello` end-to-end test pass

---

### Phase 3: Notion Agent (Week 4)
**Goal**: Build first agent - Notion Agent

**Tasks:**
1. Create `content/ticlaw/agents/notion-agent.md` (predefined agent)
2. Write SOUL.md, IDENTITY.md cho Notion agent
3. Implement Notion skill (wrap `notion_manager` :8081 API)
4. Test: "Tạo task trong Notion DB" end-to-end
5. Test: "List tasks with status=pending"

**Deliverables**:
- Notion Agent working
- Notion skill registered
- Documentation cho pattern agent definition

---

### Phase 4: Additional Agents (Week 7-8)
**Goal**: Build additional agents

**Tasks:**
1. Knowledge Sync Agent
2. Code Review Agent
3. Test Agent
4. Documentation Agent

**Deliverables**:
- 4+ agents working
- Agent templates
- Best practices

---

## 9. UI Approach

### UI Options Comparison (Revised)

| Option | Pros | Cons | Verdict |
|--------|------|------|---------|
| **Extend Router UI** | ✅ Không duplicate, consistent, 1 entry point | Cần coordinate khi update | ✅ **CHỌN** |
| **Separate UI + Shared Design System** | Cá nhân hóa độc lập | Duplicate code, 2 places to maintain | ❌ KHÔNG CHỌN |
| **Hybrid (Shared Components)** | Shared components | Complex setup cho phase đầu | ⏸️ DEFER |
| **GoClaw UI (ui/web)** | Đã có sẵn React | Không match Ti design, duplicate với router/ui | ❌ KHÔNG CHỌN |

### Selected Approach: Extend Router UI

**Rationale:**
- Ti đã có `apps/router/ui/` (React 19, Vite 6) — tương đồng với GoClaw stack
- Vừa build xong, serve từ Router :1807 (embedded)
- Thêm routes cho Ti Claw không cần tạo UI mới
- Single UI = single URL = single session auth

### Ti Claw UI Pages (Extended trong router/ui)

**Existing Routes (Router UI):**
```
/login, /dashboard, /config, /providers, /auth-files, /system
```

**New Routes (added by Ti Claw):**
```
/agents            → Agents list + templates
/agents/:id        → Agent detail + conversation UI
/agents/:id/chat   → Chat interface
/tools             → Tool registry (MCP tools)
/skills            → Skill library
/tasks             → Task board (từ Notion)
/sessions          → Session history
/memory            → Knowledge graph viewer
```

### UI Implementation Plan (Week 7)

**Phase 1**: API integration
- Ti Claw expose HTTP API tại :18790
- Router proxy tới Ti Claw: `/api/ticlaw/*` → `localhost:18790/*`
- UI call `/api/ticlaw/agents`, `/api/ticlaw/tools`, etc.

**Phase 2**: New pages in `router/ui/src/pages/`
- `AgentsPage.tsx`, `AgentDetailPage.tsx`
- `ToolsPage.tsx`, `SkillsPage.tsx`
- `TasksPage.tsx` (kết nối Notion)

**Phase 3**: Chat interface
- WebSocket connection tới Ti Claw gateway
- Streaming chat, tool call visualization
- Use pattern từ GoClaw ui/web/

**Phase 4**: Polish
- i18n (vi, en)
- Theme toggle
- Mobile responsive

**Optional (future)**: Shared Design System `packages/ti-ui-components/` chỉ khi cần support thêm 1 UI khác (vd: mobile app). **Không build trước**.

---

## 10. Anthropic Tools Integration

### 6 Tools from Anthropic Claude Code

#### 1. Skill Creator
**Purpose**: Tạo skill từ mô tả ngôn ngữ tự nhiên
**Ti Application**: Ti Skill Creator Agent
**Priority**: High (tạo skill 10x nhanh hơn)

#### 2. Superpowers
**Purpose**: Ép Claude làm việc như senior dev
**Ti Application**: Ti Superpowers Workflow
**Priority**: Medium (tăng chất lượng code 60%→80%)

#### 3. GSD (Get Stuff Done)
**Purpose**: Giải quyết context rot, sub-agent orchestration
**Ti Application**: Ti GSD System
**Priority**: High (không còn context rot)

#### 4. /review + /ultra-review
**Purpose**: Review code, chỉ báo bug thực
**Ti Application**: Ti Review System
**Priority**: Medium (chỉ báo bug thực)

#### 5. Context Mode
**Purpose**: Nén context, tránh rác
**Ti Application**: Ti Context Mode
**Priority**: High (session chạy 3 tiếng)

#### 6. Claude Mem
**Purpose**: Ghi nhớ giữa sessions
**Ti Application**: Ti Memory System
**Priority**: High (tiết kiệm ~10x token)

### Integration Priority

1. **Ti Memory System** (Priority 1) - Tiết kiệm ~10x token
2. **Ti Context Mode** (Priority 2) - Session chạy 3 tiếng
3. **Ti Review System** (Priority 3) - Chỉ báo bug thực
4. **Ti Skill Creator** (Priority 4) - Tạo skill 10x nhanh hơn
5. **Ti Superpowers** (Priority 5) - Tăng chất lượng code
6. **Ti GSD** (Priority 6) - Không còn context rot

---

## 11. Task Management Migration

### Current: BD Tool
- Location: `Z:\02_CORE\_cli\bin\bd.exe`
- Data: `Z:\03_DATA\ti`
- Git-backed: Dolt
- Command: `bd log --task="..." --agent=devin --status=complete --type=coding --domain=general`

### New: Notion Task Management
- Database: Notion "Tasks" database
- Properties: Task, Status, Type, Domain, Agent, Priority, Session ID, Error Message, Notes
- Views: Table, Group by, Timeline, Calendar, Board
- Command: `ti task log --task="..." --agent=devin --status=complete --type=coding --domain=general`

### Migration Plan (Week 1)

#### Day 1-2: Create Notion Database
- Create database "Tasks"
- Define properties
- Create views

#### Day 3-5: Implement Notion Task Logger
- Implement Notion API client
- CLI command: `ti task log`
- Test logging

#### Day 6-7: Migrate BD Data
- Export BD data
- Convert to Notion format
- Import to Notion
- Deprecate BD tool

#### Day 8-10: Optimize Task Board
- Optimize views
- Add formulas, filters
- Add automation

### Benefits vs BD Tool
- Centralized knowledge & tasks in Notion
- Better visualization (Table, Board, Timeline, Calendar)
- Advanced querying and filtering
- Automation support
- Multi-user collaboration
- Integration with knowledge system

---

## 12. Roadmap Reanalysis

### Task Dependencies

| Task | Dependencies | Can Start Now? | Quick Win? |
|------|--------------|----------------|------------|
| **Chọn nền tảng agent** | Không cần | ✅ Done | ✅ Yes |
| **Build Notion Agent** | Không cần Ti Claw | ✅ Yes | ✅ Yes |
| **Migrate BD → Notion** | Cần Notion database | ✅ Yes | ✅ Yes |
| **Notion Task Board** | Cần Notion database | ✅ Yes | ✅ Yes |
| **Auto-task Logic** | Cần Notion Agent + Agent Framework | ❌ No | ❌ No |
| **Ti Claw Framework** | Không cần | ✅ Yes | ❌ No (foundation) |
| **Devin + GitHub** | Cần GitHub integration | ✅ Yes | ✅ Yes |

### Strategy: Quick Wins First

**Old Plan (Sequential)**
```
Week 1-2: Ti Claw Core Foundation
Week 3: CLI Integration
Week 4: Notion Agent
Week 5-6: UI Development
Week 7-8: Additional Agents + Polish
```

**New Plan (Quick Wins First)**
```
Week 1: Quick Wins (Task Management in Notion) ✅
Week 2: Notion Agent ✅
Week 3-4: Ti Claw Foundation
Week 5-6: Advanced Features
Week 7: Auto-Task Logic (GSD)
Week 8: Devin + GitHub Integration
```

### Why Quick Wins First?
- ✅ Value sớm (Week 1-2)
- ✅ Validate assumptions với real usage
- ✅ Motivation cao với quick wins
- ✅ Foundation được xây trên thực tế
- ✅ Không chờ đợi foundation quá lâu

---

## 13. Implementation Roadmap (Phase-based, no deadlines)

### Scope: Ti Claw Only

File này chỉ tập trung vào **Ti Claw agent framework**. Các plans liên quan đã tách ra:
- **Router completion**: xem `ROUTER_PLAN.md`
- **Ti CLI + multi-CLI orchestration**: xem `TI_CLI_PLAN.md`

Không có deadline cứng. Mỗi phase có **gate criteria** — chỉ proceed khi đạt.

### Ti Claw Roadmap

```
A0 Study & Decide ⭐ START HERE
 ↓
A1 Quick Wins (Notion Tasks)
 ↓
A2 Fork/Plugin Setup
 ↓
A3 Ti Integration
 ↓
A4 Notion Agent (first agent)
 ↓
A5 Advanced Features
 ↓
A6 Multi-Agent + Devin/GitHub
```

**Critical path**: A0 → A2 → A3 → A4 (validate framework)

**Dependencies**:
- A3 cần Router stable → check `ROUTER_PLAN.md` Phase R-0/R-1
- A3 cần CLI foundation → check `TI_CLI_PLAN.md` Phase CLI-1/CLI-2

---

### Phase A0: Study & Decide ⭐ START HERE
**Goal**: Confirm architecture choice (App vs Plugin)

**Tasks:**
- Audit GoClaw `internal/` — đối chiếu với CLI's modules đã có
- Decide: `apps/ticlaw/` (Option A) hay `apps/cli/internal/plugins/ticlaw/` (Option B)
- Research patterns: Letta tiered memory, LangGraph state machines, CrewAI roles
- POC: GoClaw → Router :1807 (test 1 LLM call)
- POC: GoClaw MCP client → Ti MCP server (test 1 tool call)

**Gate criteria**:
- ✅ Architecture decision documented
- ✅ POC scripts pass
- ✅ List of GoClaw modules to keep vs reuse-from-CLI

---

### Phase A1: Quick Wins (Notion Task Management)
**Goal**: Deliver value ngay, không cần framework

**Tasks:**
- Create Notion Database (Tasks)
- `ti task log` command via Ti CLI
- Migrate BD data to Notion
- Optimize task board views

**Gate criteria**:
- ✅ `ti task log` working
- ✅ All BD data in Notion
- ✅ Daily workflow uses Notion

---

### Phase A2: Fork/Plugin Setup
**Goal**: Setup Ti Claw codebase

**If Option A (separate app):**
- Clone GoClaw → `apps/ticlaw/`, rename module
- Strip: channels, tts, oauth, permissions (lite không cần)
- Build `sqliteonly`

**If Option B (CLI plugin) — RECOMMENDED:**
- Create `apps/cli/internal/plugins/ticlaw/`
- Port GoClaw `internal/agent/` patterns
- Port `internal/skills/` (SKILL.md + BM25)
- Port `internal/scheduler/` (lane-based)
- Reuse CLI's `agentcore`, `brain`, `memory`, `beads`

**Gate criteria**:
- ✅ Build successful
- ✅ Basic agent loop runs end-to-end (echo agent)

---

### Phase A3: Ti Integration
**Goal**: Wire Ti Claw vào ecosystem

**Tasks:**
- Provider: point Ti Claw provider tới Router :1807
- MCP bridge: register 24+ Ti MCP tools
- Memory: connect knowledge-graph.jsonl
- CLI commands: `ti agent run`, `ti agent list`, `ti skill list`

**Gate criteria**:
- ✅ `ti agent run --name hello` end-to-end
- ✅ Agent gọi được MCP tool
- ✅ Memory persist giữa sessions

---

### Phase A4: Notion Agent (First Agent)
**Goal**: Validate framework với real use case

**Tasks:**
- Agent definition + SOUL.md, IDENTITY.md
- Notion skill (wrap notion_manager :8081)
- Tests: create/list/update tasks

**Gate criteria**:
- ✅ Notion agent autonomous flow works
- ✅ Auto-create task from `ti agent run "tạo task X"`

---

### Phase A5: Advanced Features
**Goal**: Memory, Context, Review

**Tasks:**
- Tiered memory (Letta pattern)
- Context auto-summarization >75%
- Review system (bug detection)
- Skill creator
- Superpowers workflow
- GSD context isolation

**Gate criteria**:
- ✅ Long session (3h+) không context overflow
- ✅ Review skill catch real bugs
- ✅ Skill creator generate valid SKILL.md

---

### Phase A6: Multi-Agent + Devin/GitHub
**Goal**: Multi-agent orchestration + external integration

**Tasks:**
- Sub-agent dispatch (Swarm pattern)
- Agent-to-agent handoff (AutoGen pattern)
- GitHub skill (PR, issue, commits)
- Devin as tool/skill

**Gate criteria**:
- ✅ Agent A delegate → Agent B → result back
- ✅ Agent → GitHub PR end-to-end

---

## Cross-Plan Coordination

### Dependencies trên các plans khác

**Router (`ROUTER_PLAN.md`)**:
- Ti Claw Phase A3 cần Router stable (Phase R-0 + R-1 done)
- Ti Claw UI cần Router UI working (Phase R-2)

**Ti CLI (`TI_CLI_PLAN.md`)**:
- Ti Claw Phase A2 cần CLI foundation có `agent` commands (Phase CLI-2/CLI-4)
- Ti Claw có thể là plugin trong CLI (Option B)
- CLI's agent orchestrator có thể spawn Ti Claw như sub-agent

### Recommended Priority Across Plans

```
Week 1:
- Router Phase R-0 (fix lint)      [ROUTER_PLAN.md]
- Ti Claw Phase A0 (study)         [this file]
- CLI Phase CLI-0 (audit)          [TI_CLI_PLAN.md]

Week 2:
- Router Phase R-1 (UI smoke test)
- Ti Claw Phase A1 (Notion quick wins)
- CLI Phase CLI-1 (registry foundation)

Week 3+:
- Router Phase R-2/R-3 (polish)
- Ti Claw Phase A2 (fork/plugin)
- CLI Phase CLI-2/CLI-3 (config + router integration)

Week 4+:
- Ti Claw Phase A3 (integration)
- CLI Phase CLI-4 (sub-agent runner)

...continue iteratively
```

---

## 14. Success Criteria

### Overall Success Criteria
- [x] Framework selected (GoClaw → Ti Claw)
- [x] Patterns extracted
- [x] Integration plan created
- [x] UI approach selected
- [x] Task management migration planned
- [x] Roadmap reanalyzed
- [ ] Notion database created
- [ ] Notion task logger implemented
- [ ] BD data migrated to Notion
- [ ] Notion Agent working
- [ ] Ti Claw core framework implemented
- [ ] CLI integration working
- [ ] Advanced features implemented
- [ ] GSD system implemented
- [ ] GitHub integration working
- [ ] Devin optimization working

### Per Week Success Criteria

#### Week 1: Quick Wins
- [ ] Notion database created
- [ ] Properties defined
- [ ] Views configured
- [ ] Notion task logger implemented
- [ ] CLI command working
- [ ] BD data migrated
- [ ] Data integrity verified
- [ ] Task board optimized

#### Week 2: Notion Agent
- [ ] Notion Agent simple version working
- [ ] Notion MCP integration working
- [ ] Sync logic working
- [ ] Enhanced version working

#### Week 3-4: Ti Claw Foundation
- [ ] Core framework implemented
- [ ] CLI integration working
- [ ] Tests passing

#### Week 5-6: Advanced Features
- [ ] Ti Memory System working
- [ ] Ti Context Mode working
- [ ] Ti Review System working

#### Week 7: Auto-Task Logic
- [ ] GSD system implemented
- [ ] Context isolation working
- [ ] Sub-agent orchestration working
- [ ] Quality control working

#### Week 8: Devin + GitHub
- [ ] GitHub integration working
- [ ] Auto-commit working
- [ ] Auto-PR working
- [ ] Devin optimization working

---

## 15. Key Risks & Mitigations

### Risk 1: Architecture choice paralysis (App vs Plugin)
**Severity**: High
**Risk**: Không quyết định được Option A vs B → stuck
**Mitigation**:
- Phase A0 timebox: Max 1-2 days
- Default → **Option B (Plugin)** vì CLI đã có 74% infrastructure
- Có thể switch sau nếu cần

### Risk 2: Duplicate code giữa Ti Claw và CLI
**Severity**: Medium
**Risk**: Port GoClaw's agent loop trong khi CLI đã có `agentcore`
**Mitigation**:
- Phase A0 task: đối chiếu module-by-module
- Reuse-first principle: chỉ port nếu CLI thiếu
- Document mapping table

### Risk 3: Router không phải OpenAI-compat 100%
**Severity**: Medium
**Risk**: GoClaw's OpenAI provider gọi endpoints Router chưa support
**Mitigation**:
- Phase A0 POC test cụ thể: streaming, tool calls, function calling
- Nếu missing features: Add vào Router (Phase B-something)

### Risk 4: MCP bridge incompatibility
**Severity**: Low (CLI đã có MCP module)
**Risk**: GoClaw MCP client format khác Ti MCP
**Mitigation**:
- Phase A0 POC MCP call end-to-end
- Reuse CLI's `mcp` module nếu cần adapter

### Risk 5: Scope creep với Multi-framework
**Severity**: Medium
**Risk**: Research quá nhiều frameworks → delay
**Mitigation**:
- Timebox research max 2 ngày
- Top 3 patterns only (Letta tiered memory ưu tiên)
- Implement sau Phase A4

### Risk 6: SQLite performance với nhiều agents
**Severity**: Low
**Risk**: SQLite-only không đủ
**Mitigation**:
- Personal use: 1-3 agents concurrent đủ
- GoClaw Desktop edition test 5 agents OK
- Migration path tới Postgres có sẵn

### Risk 7: Notion API rate limits
**Severity**: Low
**Risk**: Notion Agent throttled
**Mitigation**:
- Cache via Router
- Batch operations
- Backoff trong notion_manager

### Risk 8: Router/CLI parallel work conflict
**Severity**: Medium
**Risk**: 3 tracks song song có thể conflict
**Mitigation**:
- Coordination strategy in Section 13
- Daily pick 1 task per track max
- Stable interfaces giữa tracks

---

## 16. References

### Context Files (Primary Sources) ⭐
- **Router Context**: `Ti-learning-lab/03_Knowledge/Router/router-context.json` (28 providers, 12 OAuth)
- **CLI Context**: `Ti-learning-lab/03_Knowledge/Router/cli-context.json` (43/58 modules, microkernel)
- **Automation Context**: `Ti-learning-lab/03_Knowledge/Router/automation-context.json`
- **Dashboard Context**: `Ti-learning-lab/03_Knowledge/Router/dashboard-context.json`

### Source Documents
- **GoClaw Source**: `Ti-learning-lab/07_Repositories/goclaw-main/`
- **GoClaw CLAUDE.md**: `Ti-learning-lab/07_Repositories/goclaw-main/CLAUDE.md`
- **Ti CLI**: `apps/cli/`
- **Ti Router**: `apps/router/`
- **Ti Router UI**: `apps/router/ui/`
- **Ti MCP Hub**: `apps/mcp/`
- **notion_manager**: `notion_manager/`
- **Ti Skills**: `content/skills/` (100+)
- **Ti Agents**: `content/agents/` (100+)
- **Ti Rules**: `content/rules/` (13 P0/P1)
- **Knowledge Graph**: `Z:/03_DATA/ti/memory/knowledge-graph.jsonl`

### Detailed Documents
- **Patterns Extraction**: `Ti-learning-lab/03_Knowledge/Router/goclaw-patterns-extraction.md`
- **Integration Plan**: `Ti-learning-lab/03_Knowledge/Router/ti-claw-integration-plan.md`
- **Research Summary**: `Ti-learning-lab/03_Knowledge/Router/ti-claw-research-summary.md`
- **UI Approach**: `Ti-learning-lab/03_Knowledge/Router/ti-claw-ui-approach.md`
- **Anthropic Tools Analysis**: `Ti-learning-lab/03_Knowledge/Router/anthropic-claude-code-tools-analysis.md`
- **Notion Migration**: `Ti-learning-lab/03_Knowledge/Router/notion-task-management-migration.md`
- **Roadmap Reanalysis**: `Ti-learning-lab/03_Knowledge/Router/ti-roadmap-reanalysis.md`

---

## Summary

### Key Decisions (v3 — Personal Use)
1. **Framework**: Rebrand GoClaw → Ti Claw (personal, không lo license)
2. **Edition**: SQLite-only
3. **Architecture**: Default **Plugin trong CLI** (Option B), reuse CLI's agentcore/brain
4. **LLM Provider**: Ti Router (:1807) — đã có 28 providers + 12 OAuth
5. **MCP Tools**: Reuse 24+ Ti MCP tools
6. **Memory**: Reuse CLI's `memory` module + knowledge-graph.jsonl
7. **Brain/Learning**: Reuse CLI's `brain` + `beads*`
8. **UI**: Extend `apps/router/ui/`
9. **Task Mgmt**: Migrate BD → Notion qua notion_manager
10. **No deadlines**: Phase-based với gate criteria

### Asset Reuse Summary

**Don't rebuild — already exist trong Ti:**
- ✅ Agent execution (`agentcore`, `agent`, `routeagent`)
- ✅ Brain/Learning (`brain`, RL, beads)
- ✅ Plan/Verify (`ticore`, `planner`, `verify`, `repair`)
- ✅ Memory (`memory` Palace)
- ✅ Sessions (`session`)
- ✅ MCP bridge (`mcp`)
- ✅ Permission, Safety, Tools
- ✅ Provider routing (Router 28 providers)
- ✅ OAuth (Router 12 providers)
- ✅ Cache, Rate Limit, Circuit Breaker (Router)
- ✅ UI base (router/ui React 19)

**Need to add (from GoClaw):**
- Agent loop formalization (think→act→observe pattern)
- Skill system (SKILL.md + BM25)
- Lane-based scheduler
- Bootstrap (SOUL.md, IDENTITY.md)
- Agent context files structure

### Ti Claw Roadmap

```
A0 Study & Decide ⭐ START HERE
 ↓
A1 Quick Wins (Notion Task Mgmt)
 ↓
A2 Fork/Plugin Setup
 ↓
A3 Ti Integration (Router + MCP)
 ↓
A4 Notion Agent (first agent)
 ↓
A5 Advanced Features (memory, context, review)
 ↓
A6 Multi-Agent + Devin/GitHub
```

### Next Action

**Phase A0** — Architecture Decision
- Default: Plugin in CLI (Option B)
- Confirm by checking GoClaw modules vs CLI modules
- Timebox: 1-2 ngày

### Related Plans (work in parallel)

- **`ROUTER_PLAN.md`** → Start Phase R-0 (fix lint) first, blocks nothing in Ti Claw
- **`TI_CLI_PLAN.md`** → Start Phase CLI-0 (audit) parallel với A0
- See "Cross-Plan Coordination" Section 13 for priority order

---

**Status**: Ti Claw Plan v4 (split from master) ✅
**Next**: Phase A0 — Study & Decide
**Last Update**: 2026-05-05 (tách Router và CLI ra files riêng)
