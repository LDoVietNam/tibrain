# 🏛️ Architect Agent: Ti CLI Architecture Analysis

> **Agent**: `architect` (model: opus)
> **Source**: `~/.claude/agents/best/architect.md`
> **Task**: Analyze Ti CLI Hybrid Design - is it architecturally sound?

---

# 1. CURRENT STATE ANALYSIS

## Codebase Stats

| Metric | Ti CLI | Ti Router |
|--------|--------|-----------|
| **Go files** | 264 | 173 |
| **Internal packages** | 60+ (`internal/*`) | 30+ (`layers/*`) |
| **Codebase size** | 1.8MB internal | ~1MB |
| **Architecture** | Microkernel + Plugins | Layered |
| **Module path** | `github.com/ti/cli` | independent |

## Existing Architecture Pattern

### Ti CLI: **Microkernel + Plugin Pattern** (như VS Code, Eclipse)

```
┌──────────────────────────────────────────────┐
│  Cobra CLI (cmd/root.go)                      │
│  - Global flags: --config, --log-level       │
│  - Cobra command tree                        │
└──────────────────┬───────────────────────────┘
                   ↓
┌──────────────────────────────────────────────┐
│  Core Microkernel (internal/core)             │
│  - Platform interface                         │
│  - PluginRegistry                             │
│  - LifecycleManager                           │
└──────────────────┬───────────────────────────┘
                   ↓
┌──────────────────────────────────────────────┐
│  Plugin Manager (internal/plugins)            │
│  - Discovery, Loading                         │
│  - Health checks                              │
│  - gRPC bus                                   │
└──────────────────┬───────────────────────────┘
                   ↓
┌──────────────────────────────────────────────┐
│  60+ Internal Packages                        │
│  Brain, Memory, ContextPack, Planner, ...    │
└──────────────────────────────────────────────┘
```

### Ti Router: **Layered Architecture**

```
cmd/routerd/main.go (port 1806)
   ↓
layers/
  ├── server/        - HTTP server
  ├── http/          - HTTP middleware
  ├── routing/       - Route logic
  ├── translator/    - Format translation
  ├── provider/      - Provider integration
  ├── authentication/- Auth
  ├── rate/          - Rate limiting
  ├── resilience/    - Circuit breaker
  ├── usage/         - Usage tracking
  ├── metrics/       - Metrics
  ├── monitoring/    - Logging
  ├── audit/         - Audit logs
  ├── health/        - Health checks
  ├── notion/        - Notion integration
  ├── db/            - Database
  └── config/        - Configuration
```

## Identified Patterns ✅

1. **Plugin Architecture**: gRPC + lifecycle management
2. **Repository Pattern**: `internal/storage` with SQLite (modernc.org)
3. **Service Layer**: `internal/{brain, memory, contextpack, planner}`
4. **Layered Backend** (Router): Clear separation of concerns
5. **Configuration Layered**: 7-source precedence
6. **Pure Go**: No CGO dependencies (modernc.org/sqlite)
7. **Microkernel Core**: Minimal core, plugins do the work

## Technical Debt ⚠️

1. **agentcore is placeholder**: `internal/agentcore/placeholder.go` only
2. **15 modules SKIPPED**: secrets, tui, browser, browserdrivers, gmail (external deps)
3. **Two routers**: `internal/router` (Ti CLI router) vs `Ti Router` (separate Go module) - confusion
4. **Two configs**: `internal/config` and `internal/config.backup`
5. **Plugin system unfinished**: Devin Plugin (Python+gRPC) referenced but not yet implemented
6. **Translation layer**: Old Node.js code still in `noderouter.disabled/` reference

## Scalability Limitations

1. **In-memory state**: Brain's `taskProviderScore` keeps maps in memory
2. **Single Ti Router**: No horizontal scaling design (port 1806 single)
3. **SQLite single file**: Per-project SQLite limits concurrent writes
4. **gRPC plugin overhead**: Each plugin = process = memory cost

---

# 2. REQUIREMENTS GATHERING

## Functional Requirements (User Goal)

| ID | Requirement | Priority |
|----|-------------|----------|
| FR1 | Ti wraps interface .devin (and .claude, .codex, .amp) | P0 |
| FR2 | Brain Server (HTTP/gRPC + MCP) | P0 |
| FR3 | CLI Wrapper với custom prompt injection | P0 |
| FR4 | Cross-agent learning (Devin ↔ Claude ↔ Codex) | P1 |
| FR5 | Universal Context Adapter (read multi-format) | P0 |
| FR6 | Per-project + global brain storage (hybrid) | P1 |
| FR7 | Native UX preservation (agents work normally) | P0 |

## Non-Functional Requirements

### Performance
- **Smooth like Ampcode**: Latency <200ms cho brain queries
- **Context window**: Stay under 10k tokens (MemPalace)
- **Streaming**: Real-time output relay

### Versatility (Codex-like)
- **Multi-agent**: Support 30+ agents (best_source) + future
- **Multi-model**: Already have via Ti Router
- **Pluggable**: Add new agent types easily

### Quality (Claude Code-like)
- **Cross-session learning**: Brain RL persists
- **Pattern recognition**: From all agent runs
- **Context-aware**: AGENTS.md + auto memory equivalent

### Security
- **Auth**: Bearer tokens (existing in Ti Router)
- **Secrets**: `Z:\00_SECRET\` integration (existing)
- **Permissions**: Per-agent tool permissions

---

# 3. DESIGN PROPOSAL ASSESSMENT

## Hybrid Design (3 Layers) - **REVIEW**

### ✅ Strengths

**Layer 1: Ti CLI Wrapper**
- ✅ Native UX preserved (user runs `ti claude`, agents still work)
- ✅ Custom prompt injection = HIGH VALUE differentiator
- ✅ Output capture for learning = closes the feedback loop

**Layer 2: Ti Router as Brain Gateway**
- ✅ Reuse existing port 1806 (no new infra)
- ✅ Already has auth, routing, providers
- ✅ Adding `/v1/brain/*` is extensions, not rewrites
- ✅ MCP-native = future-proof

**Layer 3: Ti Brain Core**
- ✅ Reuse `internal/brain`, `internal/memory`, `internal/contextpack`
- ✅ MemPalace 4-layer = better than Claude's CLAUDE.md
- ✅ Storage hybrid (per-project + global) = best of both

### ⚠️ Concerns / Risks

**Concern 1: Wrapper Complexity**
- Ti CLI Wrapper needs to:
  - Spawn agent process (Claude/Devin/Codex/Amp)
  - Inject prompts via stdin/PTY
  - Capture output via PTY
  - Handle streaming back to user
- **Risk**: PTY handling is OS-specific, complex
- **Risk**: Some agents (Claude Code) may not support stdin injection
- **Mitigation**: Start with file-based prompt injection (write to `.claude/auto-inject.md` reading from there)

**Concern 2: Ti Router Bloat**
- Currently Ti Router does: routing, auth, rate, resilience, providers, metrics
- Adding: brain, memory, contextpack, planner = significant scope
- **Risk**: Single point of failure
- **Risk**: Performance degradation
- **Mitigation**: 
  - Brain endpoints in separate goroutines
  - Cache brain responses (already exists in Ti Router)
  - Optional: separate `Ti Brain Server` process if Router becomes too heavy

**Concern 3: Two Brains?**
- `Z:\Ti\CLI\internal\brain` (RL learning, classifier)
- `Z:\Ti\router\layers\brain` (would be NEW, query interface)
- **Risk**: Code duplication
- **Mitigation**: 
  - Ti Router brain layer just **wraps** Ti CLI's brain package
  - Keep single source of truth in `internal/brain`
  - Router exposes via API

**Concern 4: Cross-Agent Privacy**
- If Ti learns from Devin session about User A's secrets, doesn't share with User B's Claude session
- **Risk**: Data leak across users
- **Mitigation**:
  - Per-user brain storage
  - Sanitize patterns before promoting to global
  - Skill: `safety-guard` from best_source

**Concern 5: Performance Overhead**
- Each `ti claude "task"` = 2 extra calls (brain context + brain learn)
- **Risk**: Latency tăng
- **Mitigation**:
  - Brain context fetch: async + cached
  - Brain learn: fire-and-forget (background goroutine)
  - User toggleable: `ti claude --no-brain "task"` for raw

---

# 4. TRADE-OFF ANALYSIS

## Decision 1: Brain Server Location

**Options**:

### Option A: Embed in Ti Router (port 1806)
- **Pros**: Reuse existing infra, single process, share auth/cache
- **Cons**: Bloat, single point of failure, mixing concerns
- **Decision**: ✅ **CHOSEN** for Phase 2 prototype

### Option B: Separate Ti Brain Server (port 18891)
- **Pros**: Separation of concerns, can scale independently
- **Cons**: New infra, deployment complexity
- **Decision**: ⚠️ **DEFER** to Phase 5+ if Router too heavy

### Option C: Embedded library (no server)
- **Pros**: Fastest, no IPC
- **Cons**: Each agent embeds = bloat, no cross-agent sharing
- **Decision**: ❌ **REJECT** - defeats purpose

## Decision 2: Prompt Injection Method

**Options**:

### Option A: File-based injection (write to `.claude/auto-inject.md`)
- **Pros**: Simple, no PTY needed, agent reads on next session
- **Cons**: Stale data between sessions
- **Decision**: ✅ **CHOSEN** for Phase 3 v1

### Option B: PTY-based injection (intercept stdin/stdout)
- **Pros**: Real-time, dynamic context per turn
- **Cons**: Complex, OS-specific, brittle
- **Decision**: ⚠️ **PHASE 3 v2** if file-based not enough

### Option C: MCP injection (agent calls Ti Brain MCP tools)
- **Pros**: Native to Claude/Codex MCP support
- **Cons**: Requires MCP-aware agent
- **Decision**: ✅ **PARALLEL** with Option A (best for Claude/Codex)

## Decision 3: Storage Architecture

**Options**:

### Option A: Per-project SQLite + Global SQLite (hybrid)
- **Pros**: Local fast, global shared, privacy-aware
- **Cons**: Sync logic needed
- **Decision**: ✅ **CHOSEN**

### Option B: Single global SQLite
- **Pros**: Simpler, no sync
- **Cons**: Privacy issues, scaling
- **Decision**: ❌ **REJECT**

### Option C: Cloud-backed (Notion/Supabase)
- **Pros**: Multi-device sync
- **Cons**: External dep, latency, privacy
- **Decision**: 🔮 **FUTURE** Phase 6+

## Decision 4: Agent Loading Strategy

**Options**:

### Option A: Auto-load best_source on startup
- **Pros**: Always available, simple
- **Cons**: Slower startup, all agents in memory
- **Decision**: ✅ **CHOSEN** with lazy loading

### Option B: On-demand load
- **Pros**: Fast startup
- **Cons**: First-call latency
- **Decision**: ⚠️ Combine with A (eager metadata, lazy body)

---

# 5. ARCHITECTURE DECISION RECORDS (ADRs)

## ADR-001: Hybrid Architecture (Brain Server + CLI Wrapper)

### Context
Ti needs to work as brain layer for multiple agents (Devin, Claude Code, Codex, Ampcode) while preserving native UX of each.

### Decision
Implement **3-layer hybrid**:
- Layer 1: Ti CLI Wrapper (meta-agent với prompt injection)
- Layer 2: Ti Router Brain Gateway (port 1806 + brain endpoints + MCP)
- Layer 3: Ti Brain Core (existing internal packages)

### Consequences

**Positive**:
- Native UX preserved (no migration burden)
- Reuse existing Ti CLI brain/memory/contextpack
- MCP-native for Claude/Codex
- Cross-agent learning capability

**Negative**:
- Complexity: 3 layers to maintain
- Potential performance overhead (2 extra API calls per task)
- Wrapper is OS-specific (PTY handling on Windows different from Unix)

### Alternatives Considered
- **Option A (Meta-Agent only)**: Wrap all agents - rejected (loses native UX)
- **Option C (Library only)**: Can't with closed-source agents - rejected
- **Manual .devin/**: User maintains - rejected (no learning)

### Status
**Accepted**

### Date
2026-04-28

---

## ADR-002: Use Ti Router as Brain Gateway

### Context
Ti needs HTTP/gRPC/MCP interface for agents to query brain.

### Decision
**Embed brain endpoints in Ti Router** (port 1806) rather than separate process.

### Consequences

**Positive**:
- Reuse auth, rate limiting, caching
- Single port to manage (1806)
- All agents already use Router for AI calls

**Negative**:
- Bloat in Ti Router
- Single point of failure
- Mixed concerns (routing + brain)

### Alternatives Considered
- **Separate Ti Brain Server (port 18891)**: Cleaner but more deployment overhead
- **Library only**: Can't expose to agents

### Status
**Accepted with monitoring** - revisit if Router latency degrades

### Date
2026-04-28

---

## ADR-003: AGENTS.md as Universal Format

### Context
Multiple agents (Claude/Codex/Amp) support different memory formats. Need unified standard.

### Decision
Adopt **AGENTS.md** (supported by Amp + Codex + Claude Code) as primary format. Other formats (`.devin/`, `.claude/CLAUDE.md`) become bridges.

### Consequences

**Positive**:
- Universal compatibility
- Single source of truth per project
- No duplication

**Negative**:
- Migration effort from `.devin/`
- Some agents may not respect AGENTS.md fully

### Alternatives Considered
- **Custom format**: Lock-in, no compatibility
- **Multiple files synced**: Sync hell

### Status
**Accepted**

### Date
2026-04-28

---

# 6. SCALABILITY PLAN

## Current Scale (1 user, 1 project)
- **Brain**: In-memory, ~10MB
- **MemPalace**: SQLite ~50MB
- **Ti Router**: Single instance, port 1806
- ✅ **Sufficient**

## Near-term (1 user, 10 projects)
- **Brain**: Per-project SQLite + global SQLite
- **MemPalace**: ~500MB total
- **Ti Router**: Single instance OK
- ✅ **Sufficient with hybrid storage**

## Mid-term (1 user, 100 projects, 1000 sessions/day)
- **Brain**: Index optimization, vacuum scheduling
- **MemPalace**: Compression, archival
- **Ti Router**: Cache aggressive, ratelimit user
- ⚠️ **Need monitoring**

## Long-term (Multi-user, cloud)
- **Brain**: Per-user partitioning
- **MemPalace**: PostgreSQL backend, vector DB for semantic search
- **Ti Router**: Horizontal scaling, load balancer
- 🔮 **Architecture pivot needed**

---

# 7. RED FLAGS CHECKLIST

Audit Ti's Hybrid Design against architect's red flags:

- [ ] **Big Ball of Mud**: ❌ NO - Clear 3-layer separation
- [ ] **Golden Hammer**: ⚠️ Risk - Reusing Ti Router for too many things
- [ ] **Premature Optimization**: ❌ NO - MemPalace 4-layer is justified
- [ ] **Not Invented Here**: ✅ Using best_source, MCP standard, AGENTS.md
- [ ] **Analysis Paralysis**: ⚠️ Risk - Spent time on docs, need to ship
- [ ] **Magic**: ⚠️ Risk - Prompt injection feels magical, needs clear docs
- [ ] **Tight Coupling**: ✅ Each layer can work independently
- [ ] **God Object**: ⚠️ Ti Router doing too much

---

# 8. RECOMMENDATIONS

## Architectural Soundness: **GOOD with Caveats** (8/10)

### ✅ What's Right

1. **3-layer hybrid is sound** - clear separation, native UX preserved
2. **Reuse Ti's existing intelligence** (brain, memory, contextpack) - no waste
3. **MCP-native** for future-proofing
4. **AGENTS.md universal** - smart convergence
5. **Hybrid storage** - good privacy/sharing balance

### ⚠️ What Needs Attention

1. **Define Brain API contract FIRST** before implementing
   - OpenAPI spec for `/v1/brain/*`
   - MCP tool definitions
2. **Performance budget**: Set max latency for brain calls (e.g., 100ms p95)
3. **Wrapper fallback**: If brain server down, agents still work
4. **Privacy boundary**: Clear rules for what crosses to global brain
5. **Telemetry**: Instrument all 3 layers from day 1

### 🔧 Recommended Adjustments

1. **Split Brain Gateway from Ti Router** when Router > 50% CPU on brain calls
   - Threshold: 1000 brain QPS
   - Plan: Separate `Ti Brain` binary
2. **Start with file-based prompt injection** (Option A in Decision 2)
   - Phase 3a: Write to `.claude/auto-inject.md`
   - Phase 3b: Add PTY injection (advanced)
3. **MCP server FIRST** for Claude/Codex compatibility
   - Phase 2 priority: MCP endpoint
   - HTTP endpoints secondary
4. **Test with 1 agent end-to-end** before adding others
   - Recommendation: Start with **Claude Code** (best MCP support)

---

# 9. FINAL VERDICT

## Is Ti CLI Hybrid Design Architecturally Sound?

**YES, with these conditions**:

1. ✅ **Define API contracts** before implementation (OpenAPI + MCP)
2. ✅ **Start with file-based injection** before PTY
3. ✅ **Test with 1 agent** (Claude Code) end-to-end
4. ✅ **Set performance budgets** (100ms p95 brain calls)
5. ✅ **Plan to split Brain Server** if Router becomes overloaded
6. ✅ **Implement telemetry** from day 1

## Recommended First Sprint (1 week)

```
Day 1: API contract design
  - OpenAPI spec for /v1/brain/*
  - MCP tool definitions
  
Day 2-3: Brain endpoints in Ti Router
  - /v1/brain/context
  - /v1/brain/memory/recall
  - /mcp endpoint (basic)
  
Day 4-5: Test with Claude Code via MCP
  - Configure Claude Code to use Ti as MCP server
  - Verify ti_get_context, ti_recall_memory work
  
Day 6: Telemetry + monitoring
  - Add metrics: brain QPS, latency, errors
  - Add health check
  
Day 7: Documentation + ADRs
  - Update AGENTS.md
  - Finalize ADRs
```

## What NOT to Do (Phase 1)

- ❌ Don't implement CLI Wrapper yet (Phase 3)
- ❌ Don't implement PTY injection (Phase 3b)
- ❌ Don't try all 4 agents at once (focus Claude Code)
- ❌ Don't migrate `.devin/` until AGENTS.md tested
- ❌ Don't optimize prematurely (MVP first)

---

# 10. NEXT STEPS

1. **Review this analysis** - confirm architectural decisions
2. **Choose first sprint focus** (Brain endpoints recommended)
3. **Write OpenAPI spec** for brain endpoints
4. **Use `@planner` agent** to break down sprint into tasks
5. **Use `/orchestrate feature` command** for multi-agent workflow

---

**Architecture Status**: 🟢 **APPROVED with adjustments**

**Confidence**: HIGH (8/10)

**Risk Level**: MEDIUM (3 layers = surface area, but mitigations clear)

**Recommended Action**: **Proceed with Phase 2 (Brain Endpoints)** with adjustments above

---

*Generated by simulating `architect` agent (model: opus) from `~/.claude/agents/best/architect.md`*
*Date: 2026-04-28*
