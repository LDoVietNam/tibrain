# Ti Wrap Interface Analysis & Recommendation

> **Mục tiêu**: Ti CLI wrap interface .devin để hoạt động:
> - **Mượt mà như Ampcode**
> - **Đa biến như Codex**
> - **Chất lượng cao như Claude Code**

---

## 🔍 PHÂN TÍCH 4 TOOLS

### 1. Ampcode (Smoothness)

**Triết lý**:
- "Frontier coding agent" - tận dụng tối đa các model mạnh nhất
- 4 nguyên tắc: unconstrained tokens, best models, raw model power, evolves with new models

**Architecture**:
```
Amp = Model + System Prompt + Tools

Modes:
  - smart  (Opus 4.7, default)
  - rush   (faster, cheaper, small tasks)
  - deep   (GPT-5.4, extended thinking)
  - large  (hidden mode)

Specialized Agents:
  - Oracle    → Review, planning, complex problems
  - Librarian → Codebase exploration, search
  - Painter   → UI/visual tasks
```

**Context Management** (CỰC KỲ QUAN TRỌNG):
- **Threads** = context windows (save/share/reference)
- **Handoff**: extract data → new fresh thread (giảm noise)
- **Edit & Restore**: cleanup context window
- **Reference Threads**: pull info từ thread khác qua `read_thread` tool
- **AGENTS.md**: multi-level (cwd, parent dirs, $HOME/.config/amp/)
- **@file mention**: include file content
- **Shell mode**: `$cmd` execute và include output
- Subagents: parallel work

**Memory storage**:
- ✅ `AGENTS.md` (primary, shared format!)
- ✅ Threads stored on Amp servers
- ❌ KHÔNG có `.amp/` directory local

**Smoothness mechanism**:
1. **Short focused conversations** → less noise → better quality
2. **Handoff** giữa threads để giữ context window nhỏ
3. **Multi-mode** chọn model phù hợp với task complexity
4. **Subagents** parallel execution

---

### 2. Claude Code (Quality)

**Triết lý**:
- Agentic coding tool - reads codebase, edits, runs commands
- Deep context awareness, persistent memory across sessions

**Memory Architecture** (HAI HỆ THỐNG):

| | CLAUDE.md | Auto Memory |
|---|---|---|
| **Who writes** | User | Claude itself |
| **Contains** | Instructions/rules | Learnings/patterns |
| **Scope** | Project/user/org | Per working tree |
| **Loaded** | Every session | First 200 lines/25KB |
| **Use for** | Standards, workflows | Build commands, insights |

**File structure**:
```
project/
├── CLAUDE.md                    # Project instructions
├── .claude/
│   ├── CLAUDE.md                # Alt location
│   ├── rules/                   # Path-scoped rules
│   │   ├── frontend.md          # Glob: src/components/**
│   │   └── backend.md           # Glob: api/**
│   ├── commands/                # Custom commands
│   │   ├── review-pr.md
│   │   └── deploy-staging.md
│   ├── hooks/                   # Pre/post action shells
│   ├── skills/                  # Procedural knowledge
│   ├── agents/                  # Subagents
│   └── memory/                  # Auto memory storage
└── AGENTS.md                    # Shared format support
```

**Quality mechanism**:
1. **Multi-level CLAUDE.md** (managed/project/user) với precedence
2. **Path-scoped rules** với glob patterns
3. **Auto memory** học từ corrections
4. **Hooks** automate workflows
5. **Subagents** với own auto memory
6. **Imports**: `@./other-file.md` to compose
7. **MCP** integration

**Memory loading order**:
```
managed policy → org → user → project → AGENTS.md → auto memory
```

---

### 3. Codex CLI (Versatility)

**Triết lý**:
- Multi-language, multi-model, plugin-based
- TypeScript + Rust hybrid (codex-cli + codex-rs)

**File structure**:
```
codex/
├── codex-cli/                   # TypeScript CLI
├── codex-rs/                    # Rust core (high perf)
├── .codex/
│   └── skills/
│       ├── babysit-pr/
│       ├── code-review/
│       ├── code-review-context/
│       ├── code-review-testing/
│       ├── codex-bug/
│       ├── codex-issue-digest/
│       ├── codex-pr-body/
│       └── remote-tests/
├── sdk/                         # Build custom workflows
└── patches/                     # Patch management
```

**Skill structure** (each skill = directory):
```
skill-name/
├── README.md                    # What it does
├── SKILL.md                     # Procedure
└── helpers/                     # Optional scripts
```

**Versatility mechanism**:
1. **Skills as directories** (modular, composable)
2. **TypeScript + Rust** (best of both worlds)
3. **SDK** for custom workflows
4. **MCP** for external integration
5. **Multi-model** support
6. **Patches** system

---

### 4. Devin (.devin/)

**File structure**:
```
.devin/
├── README.md
├── context/
│   ├── project-context.md
│   ├── repo-context.md
│   └── task-context.md
├── knowledge/
│   ├── repo-index.md
│   ├── patterns.md
│   └── lessons.md
├── memory/
│   ├── session-memory.md
│   └── long-term-memory.md
├── skills/
│   └── [skill-name]/
└── config/
    └── settings.json
```

**Đặc điểm**:
- File-based (markdown)
- Cấu trúc hơi gần với Claude Code
- Skills directory style giống Codex
- Knowledge/memory rõ ràng

---

## 🎯 KEY INSIGHTS

### Insight 1: **`AGENTS.md` đang trở thành STANDARD** 🌟

| Tool | Format |
|------|--------|
| Ampcode | ✅ `AGENTS.md` (primary) |
| Codex | ✅ `AGENTS.md` (shared) |
| Claude Code | ✅ `AGENTS.md` (supported) + CLAUDE.md |
| Devin | ⚠️ `.devin/` (custom) |

**→ Convergence**: AGENTS.md là format chung, Claude/Codex/Amp đều support.

### Insight 2: **Context Windows là vấn đề lớn**

Tất cả 3 tool đều cần manage context window:
- **Ampcode**: Threads, Handoff, Edit & Restore
- **Claude Code**: Auto memory + CLAUDE.md (200 lines/25KB limit)
- **Codex**: Skills (load on demand)

**→ Ti Memory MemPalace** đã giải quyết vấn đề này tốt hơn (4-layer, on-demand).

### Insight 3: **Ti CLI có những thứ chúng KHÔNG có**

| Feature | Ampcode | Claude Code | Codex | Ti CLI |
|---------|---------|-------------|-------|---------|
| RL Learning | ❌ | ⚠️ Auto memory | ❌ | ✅ Brain Engine |
| 4-Layer Memory | ❌ | ⚠️ CLAUDE.md | ❌ | ✅ MemPalace |
| Smart Context Bundling | ⚠️ Manual | ⚠️ Imports | ❌ | ✅ ContextPack |
| Multi-Agent Council | ⚠️ Subagents | ⚠️ Subagents | ❌ | ✅ Planner Council |
| Cross-Agent Sharing | ❌ | ❌ | ❌ | ✅ (potential) |
| Plugin Architecture | ⚠️ MCP only | ⚠️ MCP only | ⚠️ Skills | ✅ Microkernel |

**→ Ti CLI có lợi thế kiến trúc lớn**

### Insight 4: **Cross-Agent Brain Sharing là KILLER FEATURE**

Không tool nào hiện tại làm được:
- Claude Code học từ Devin sessions
- Devin học từ Codex skills  
- Ampcode patterns share với Claude Code

**→ Ti CLI có thể làm điều này**

---

## 💡 3 OPTIONS PHÂN TÍCH

### Option A: Ti as Meta-Agent (Wrap All)

**Concept**: Ti CLI sở hữu `.devin/.claude/.codex/.amp/`, agents trở thành plugins.

```
User → Ti CLI → translates context → invokes Devin/Claude/Codex/Amp
                                  ← combines outputs
```

**Pros**:
- ✅ Single source of truth
- ✅ Auto translation between formats
- ✅ Cross-agent learning

**Cons**:
- ❌ User phải dùng Ti CLI thay vì agents trực tiếp
- ❌ Phức tạp - phải implement plugin cho mỗi agent
- ❌ Performance overhead (Ti → agent → Ti)
- ❌ Khó debug khi có issue
- ❌ User mất native experience của agent

**Verdict**: ❌ Không recommend

---

### Option B: Ti as Brain Server (Recommended)

**Concept**: Ti CLI chạy như HTTP/gRPC server, agents gọi API khi cần brain features.

```
User → Devin/Claude/Codex/Amp (native UX)
              ↓ optional API call
          Ti Brain Server
              ↓ HTTP/gRPC
       Brain + Memory + Patterns
```

**Architecture**:
```
┌─────────────────────────────────────────────────────────────────┐
│                  Ti CLI Brain Server                              │
│  (port: configurable, e.g., 18891)                                │
│                                                                   │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │  Universal Context Adapter (NEW)                          │  │
│  │  Reads multiple formats:                                  │  │
│  │  • AGENTS.md (primary - shared standard)                  │  │
│  │  • CLAUDE.md (Claude Code)                                │  │
│  │  • .devin/ (Devin)                                        │  │
│  │  • .codex/skills/ (Codex)                                 │  │
│  │  • .claude/rules/, .claude/commands/                      │  │
│  └──────────────────────────────────────────────────────────┘  │
│                                                                   │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │  Ti Brain (EXISTING - internal/brain)                     │  │
│  │  • RL learning                                            │  │
│  │  • Cross-agent pattern sharing                            │  │
│  │  • Provider scoring                                       │  │
│  └──────────────────────────────────────────────────────────┘  │
│                                                                   │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │  Ti Memory MemPalace (EXISTING - internal/memory)         │  │
│  │  • L0: Identity (always loaded, ~50-100 tokens)           │  │
│  │  • L1: Essential Story (top weighted, ~500-800 tokens)    │  │
│  │  • L2: On-Demand (filtered, ~200-500 tokens)              │  │
│  │  • L3: Deep Search (semantic, unlimited)                  │  │
│  └──────────────────────────────────────────────────────────┘  │
│                                                                   │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │  Ti ContextPack (EXISTING - internal/contextpack)          │  │
│  │  + Planner Council (EXISTING - internal/planner)          │  │
│  │  • Smart context bundling                                  │  │
│  │  • Multi-agent planning                                   │  │
│  └──────────────────────────────────────────────────────────┘  │
│                                                                   │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │  HTTP/gRPC API                                             │  │
│  │  GET  /context/{agent}/{project}                          │  │
│  │  POST /memory/learn                                        │  │
│  │  GET  /skill/{name}                                       │  │
│  │  POST /pattern/observe                                     │  │
│  │  GET  /handoff/{thread_id}                                │  │
│  └──────────────────────────────────────────────────────────┘  │
└──────────────────────────────────────────────────────────────────┘
                          │
        ┌────────────────┼────────────────┐
        │                │                │
        ▼                ▼                ▼
┌──────────────┐ ┌──────────────┐ ┌──────────────┐
│   Devin       │ │ Claude Code  │ │  Codex CLI   │
│   .devin/     │ │  .claude/    │ │  .codex/     │
│  AGENTS.md    │ │ AGENTS.md    │ │ AGENTS.md    │
└──────────────┘ └──────────────┘ └──────────────┘
        │                │                │
        ▼                ▼                ▼
   Optional Ti    Optional Ti    Optional Ti
   Brain shim     Brain shim     Brain shim
   (MCP server)  (MCP server)   (MCP server)
```

**Implementation**:
1. Ti CLI có sẵn `internal/brain/`, `internal/memory/`, `internal/contextpack/`
2. Add `internal/contextio/` - Universal Context Adapter (đọc tất cả format)
3. Add `internal/api/server/` - HTTP/gRPC server
4. Expose qua **MCP server** - agents (Claude/Devin) auto-detect MCP và dùng

**Pros**:
- ✅ Native UX cho mỗi agent
- ✅ Optional - agents vẫn hoạt động không cần Ti
- ✅ Cross-agent brain sharing
- ✅ Reuse existing Ti infrastructure
- ✅ MCP-native (Claude Code, Codex đều support)
- ✅ Easy to add new agents (chỉ cần MCP support)

**Cons**:
- ⚠️ Phải maintain MCP server
- ⚠️ Performance overhead (small với gRPC)

**Verdict**: ✅✅✅ **RECOMMEND**

---

### Option C: Ti as Library (Embedded)

**Concept**: Ti CLI là Go library, embed vào mỗi agent.

**Pros**:
- ✅ Fastest performance
- ✅ Direct memory access

**Cons**:
- ❌ Không thể với Claude Code (closed source)
- ❌ Không thể với Codex (Rust)
- ❌ Chỉ có thể với Devin (open ecosystem)

**Verdict**: ❌ Không khả thi với closed-source agents

---

## 🎯 RECOMMENDATION: **Option B + AGENTS.md Convergence**

### Phase 1: AGENTS.md Convergence (1 day)

**Mục tiêu**: Đưa tất cả agents về AGENTS.md format chung.

```bash
# Mỗi project có AGENTS.md (Ampcode, Codex, Claude Code đều support)
project/
├── AGENTS.md                    # 🌟 PRIMARY (universal)
├── .devin/                      # Devin-specific bridge
│   └── README.md → links to AGENTS.md
├── .claude/                     # Claude Code-specific bridge
│   ├── rules/                   # Path-scoped (Claude only)
│   └── commands/                # Slash commands (Claude only)
└── .codex/                      # Codex-specific bridge
    └── skills/                  # Codex skills
```

**Lợi ích**:
- ✅ Single source of truth (AGENTS.md)
- ✅ Mỗi agent vẫn có specific features
- ✅ Không phải duplicate content

### Phase 2: Ti Brain Server (1 week)

**Mục tiêu**: Ti CLI chạy như brain server, expose qua MCP.

```bash
# Ti CLI commands
ti brain serve --port 18891          # Start brain server
ti brain mcp                          # Start MCP server
ti memory recall "<query>"            # Recall from MemPalace
ti pattern observe <event>            # Add pattern
```

**MCP Tools**:
```
ti_get_context(agent, project)
ti_recall_memory(layer, query)
ti_observe_pattern(event)
ti_get_skill(name)
ti_handoff(from_thread, to_goal)
```

### Phase 3: Cross-Agent Brain (2 weeks)

**Mục tiêu**: Ti Brain học từ tất cả agents.

```
Devin completes task → Ti Brain observes → patterns learned
Claude Code starts task → Ti Brain shares patterns → better quality
Codex runs skill → Ti Brain logs → next time faster
```

### Phase 4: Three Goals Met

| Goal | How Ti Achieves It |
|------|---------------------|
| **Smooth like Ampcode** | MemPalace 4-layer (small context window) + Handoff API + ContextPack smart bundling |
| **Versatile like Codex** | Skills system + Plugin architecture + Multi-agent support + AGENTS.md universal |
| **Quality like Claude Code** | Brain RL learning + Cross-agent insights + Multi-level rules + Hooks |

---

## 📋 IMMEDIATE NEXT STEPS

### Step 1: Decision
Confirm Option B (Brain Server) là approach đúng.

### Step 2: AGENTS.md First
Trước khi build Ti Brain Server, **convert .devin/ → AGENTS.md style**:
- Tạo `AGENTS.md` ở root projects
- `.devin/` chỉ chứa Devin-specific (bridge)
- Tương thích với Claude Code và Codex

### Step 3: Universal Context Adapter
Build `internal/contextio/` trong Ti CLI:
- Đọc AGENTS.md, CLAUDE.md, .devin/, .codex/skills/
- Convert thành unified `ContextBundle`
- Reuse `internal/contextpack/`

### Step 4: MCP Server
Expose Ti Brain qua MCP:
- Claude Code: tự động detect MCP
- Devin: support MCP tools
- Codex: MCP integration

### Step 5: Test với 1 Project
Apply on Ti-Learning-Lab:
- AGENTS.md ở root
- Ti Brain MCP server running
- Test với Claude Code + Devin

---

## 🤔 OPEN QUESTIONS

1. **Storage**: Ti Brain memory lưu ở đâu?
   - Option 1: Per-project (`.ti-brain/` trong mỗi project)
   - Option 2: Global (`~/.ti/brain/` user-level)
   - Option 3: Hybrid (local + global) ← recommended

2. **Sync**: Cross-project pattern sharing?
   - Auto-share patterns across all projects của user
   - Privacy-aware (loại bỏ project-specific data)

3. **MCP vs HTTP/gRPC**: Cách nào primary?
   - MCP: native cho Claude Code, đang trở thành standard
   - HTTP/gRPC: more flexible, but agents cần custom client
   - **Recommend**: MCP primary, HTTP/gRPC fallback

4. **Agent Plugins**: Ti CLI's plugin system cho Devin/Claude/Codex?
   - Ti CLI đã có plugin architecture
   - Có thể add Devin plugin (Python bridge), Claude plugin, etc.
   - Nhưng chỉ cần khi muốn invoke agent từ Ti CLI

---

## 📊 COMPARISON SUMMARY

| Aspect | Manual .devin | Option A (Wrap) | **Option B (Brain Server)** | Option C (Library) |
|--------|---------------|------------------|----------------------------|---------------------|
| Native UX | ✅ | ❌ | ✅ | ❌ |
| Cross-agent learning | ❌ | ✅ | ✅ | ❌ |
| Easy to maintain | ✅ | ❌ | ✅ | ⚠️ |
| Performance | ✅ | ❌ | ✅ | ✅ |
| Future-proof | ⚠️ | ❌ | ✅ | ❌ |
| Reuse Ti infra | ❌ | ✅ | ✅ | ✅ |
| Smooth | ⚠️ | ❌ | ✅ | ✅ |
| Versatile | ⚠️ | ✅ | ✅ | ❌ |
| Quality | ⚠️ | ✅ | ✅ | ✅ |

**Winner**: **Option B - Ti as Brain Server**

---

**Last Updated**: 2026-04-28
