# Ti Hybrid Design: Brain Server + CLI Wrapper + Prompt Injection

> **Final Design**: Kết hợp 3 layers để đạt mục tiêu:
> - **Smooth như Ampcode** (Memory MemPalace + Handoff)
> - **Versatile như Codex** (Plugin + Skills system)
> - **Quality như Claude Code** (Brain RL + Cross-agent learning)

---

## 🏗️ HYBRID ARCHITECTURE (3 LAYERS)

```
┌─────────────────────────────────────────────────────────────────────┐
│ LAYER 1: Ti CLI Wrapper (Meta-Agent với Custom Prompt Injection)    │
│                                                                       │
│  Z:\Ti\CLI\bin\ti.exe                                                 │
│                                                                       │
│  Commands:                                                            │
│  • ti claude "do X"     → Wrap Claude Code với context injection     │
│  • ti devin "do Y"      → Wrap Devin với context injection           │
│  • ti codex "do Z"      → Wrap Codex với context injection           │
│  • ti amp "do W"        → Wrap Ampcode với context injection         │
│  • ti chat "do K"       → Direct route qua Ti Brain → best agent     │
│                                                                       │
│  Flow:                                                                │
│  1. User input prompt                                                 │
│  2. Query Ti Brain → relevant context, patterns, memory, skills      │
│  3. Inject context vào prompt:                                        │
│     [Ti Context]                                                      │
│     - Project: ti-learning-lab                                        │
│     - Recent patterns: ...                                            │
│     - Relevant memory: ...                                            │
│     - Applicable skills: ...                                          │
│     [/Ti Context]                                                     │
│     User: do X                                                        │
│  4. Spawn agent (claude/devin/codex/amp) với enhanced prompt         │
│  5. Stream output to user                                            │
│  6. Capture & feed back vào Ti Brain (learning)                      │
└─────────────────────────────────────────────────────────────────────┘
                                  │
                                  ▼ HTTP/gRPC/MCP
┌─────────────────────────────────────────────────────────────────────┐
│ LAYER 2: Ti Router as Brain Gateway (port 1806)                      │
│                                                                       │
│  Z:\Ti\router\ - existing Ti Router (Go)                              │
│                                                                       │
│  Existing capabilities (KEEP):                                        │
│  • Multi-provider routing (Groq, OpenRouter, Gemini, ...)            │
│  • Auth, rate limiting, fallback chains                              │
│  • Combo routing, circuit breaker                                    │
│  • OpenAI-compatible API (/v1/chat/completions, /v1/models)          │
│                                                                       │
│  NEW additions:                                                       │
│  • /v1/brain/context          - Get Ti Brain context                 │
│  • /v1/brain/memory/recall    - Recall MemPalace memory              │
│  • /v1/brain/pattern/observe  - Add learned pattern                  │
│  • /v1/brain/skill/get        - Get skill content                    │
│  • /v1/brain/handoff          - Create handoff bundle (persistent)    │
│  • /v1/brain/handoff/{id}     - Recall handoff bundle                │
│  • /mcp                       - MCP server endpoint                  │
│                                                                       │
│  Why Router is perfect Brain Gateway:                                │
│  • Already runs on port 1806 (always available)                      │
│  • Already has auth, routing, providers                              │
│  • Just add brain endpoints (small extension)                         │
│  • All agents already use Router for AI calls                        │
└─────────────────────────────────────────────────────────────────────┘
                                  │
                                  ▼
┌─────────────────────────────────────────────────────────────────────┐
│ LAYER 3: Ti Brain Core (existing internal packages)                  │
│                                                                       │
│  Z:\Ti\CLI\internal\                                                  │
│                                                                       │
│  ┌─────────────────────────────────────────────────────────────┐  │
│  │  Universal Context Adapter (NEW)                              │  │
│  │  internal/contextio/                                          │  │
│  │  Reads/writes:                                                │  │
│  │  • AGENTS.md (universal)                                      │  │
│  │  • CLAUDE.md (Claude Code)                                   │  │
│  │  • .devin/ (Devin)                                            │  │
│  │  • .codex/skills/ (Codex)                                    │  │
│  │  • .claude/rules/ (Claude path-scoped)                       │  │
│  └─────────────────────────────────────────────────────────────┘  │
│                                                                       │
│  ┌─────────────────────────────────────────────────────────────┐  │
│  │  Brain Engine (EXISTING) - internal/brain/                   │  │
│  │  • RL learning, pattern recognition                          │  │
│  │  • Cross-agent insights                                       │  │
│  └─────────────────────────────────────────────────────────────┘  │
│                                                                       │
│  ┌─────────────────────────────────────────────────────────────┐  │
│  │  Memory MemPalace (EXISTING) - internal/memory/              │  │
│  │  L0 Identity → L1 Essential → L2 On-Demand → L3 Deep Search │  │
│  └─────────────────────────────────────────────────────────────┘  │
│                                                                       │
│  ┌─────────────────────────────────────────────────────────────┐  │
│  │  ContextPack (EXISTING) - internal/contextpack/              │  │
│  │  Smart context bundling per agent                             │  │
│  └─────────────────────────────────────────────────────────────┘  │
│                                                                       │
│  ┌─────────────────────────────────────────────────────────────┐  │
│  │  Planner Council (EXISTING) - internal/planner/              │  │
│  │  Multi-agent planning, council pattern                        │  │
│  └─────────────────────────────────────────────────────────────┘  │
│                                                                       │
│  Storage (HYBRID):                                                   │
│  • Per-project: <project>/.ti-brain/ (gitignored)                    │
│  • Global: ~/.ti/brain/ (cross-project patterns)                     │
│  • SQLite-backed (existing modernc.org/sqlite)                       │
└─────────────────────────────────────────────────────────────────────┘
```

---

## 🎯 HYBRID FLOW (Best of All Worlds)

### Use Case 1: User runs `ti claude "fix the auth bug"`

```
1. Ti CLI Wrapper receives prompt
   ↓
2. Wrapper calls Ti Router: GET /v1/brain/context?project=ti-router&task=fix-bug
   ↓
3. Ti Router → Brain Core
   - Memory recalls: "previous auth bugs in this project"
   - Patterns finds: "JWT validation issues common in Go"
   - Skills suggests: "use security-guard agent"
   - Handoff data: "prev session debugged similar"
   ↓
4. Ti Router returns enhanced context bundle
   ↓
5. Wrapper injects context into prompt:
   ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
   [Ti Context]
   Project: ti-router (Go)
   Recent patterns:
   - JWT issues common, check token expiry first
   - Auth middleware in layers/auth/
   Memory:
   - 2 weeks ago: similar bug in OAuth refresh
   - Solution: token refresh window 5min
   Applicable skills:
   - security-guard: vulnerability check
   [/Ti Context]
   
   User: fix the auth bug
   ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
   ↓
6. Wrapper spawns: claude --enhanced-prompt
   ↓
7. Claude works (native UX), output streamed to user
   ↓
8. After Claude done: Wrapper captures output
   ↓
9. Wrapper calls: POST /v1/brain/pattern/observe
   - Records: "auth bug fix in ti-router using JWT"
   - Updates: pattern weights
   - Saves: to MemPalace L1/L2
   ↓
10. Next time user runs auth-related task → smarter response
```

### Use Case 2: Direct API call (Devin/Claude calls Ti Router as MCP)

```
1. Devin starts task: "implement combo routing"
   ↓
2. Devin's MCP integration auto-discovers Ti Router MCP server
   ↓
3. Devin calls MCP tool: ti_recall_memory(query="combo routing")
   ↓
4. Ti Router → Brain → returns relevant memories
   ↓
5. Devin uses memories as context (native MCP flow)
   ↓
6. Devin completes task, calls: ti_observe_pattern(event="implemented combo")
   ↓
7. Ti Brain updates patterns
```

### Use Case 3: Cross-Agent Learning

```
1. Claude Code completes auth bug fix
   → Pattern: "JWT validation, token expiry"
   → Saved to Ti Brain
   ↓
2. Next day, Devin starts NEW project
   → Calls: ti_get_context(project="new-project")
   → Brain returns: "Pattern from another project: JWT expiry first check"
   → Devin applies the learned pattern
   ↓
3. Brain learns: "JWT pattern is universal across projects"
   → Promotes to global brain
```

---

## 📋 IMPLEMENTATION PLAN (Parallel Phases)

### Phase 1: AGENTS.md Convergence (1 day) - LOW RISK

**Goal**: Standardize on AGENTS.md as universal format.

```bash
# Trong mỗi project:
project/
├── AGENTS.md                # 🌟 PRIMARY (universal)
├── .devin/
│   ├── README.md            # Bridge to AGENTS.md
│   └── skills/              # Devin-specific skills
├── .claude/
│   ├── rules/               # Claude path-scoped rules
│   ├── commands/            # Claude slash commands
│   └── hooks/               # Claude hooks
└── .codex/
    └── skills/              # Codex skills
```

**Tasks**:
1. Generate `AGENTS.md` from existing `.devin/context/project-context.md`
2. Move `.devin/knowledge/patterns.md` → `.devin/skills/patterns/`
3. Add to `.gitignore`: `.ti-brain/`, session memory
4. Document in README

**Verification**: Claude Code, Codex, Ampcode đều đọc được AGENTS.md.

---

### Phase 2: Ti Router Brain Endpoints (3-5 days) - MEDIUM RISK

**Goal**: Add brain endpoints to existing Ti Router (port 1806).

**File**: `Z:\Ti\router\layers\brain\` (NEW package)

```go
// brain.go
package brain

import (
    "github.com/ti/cli/internal/brain"
    "github.com/ti/cli/internal/memory"
    "github.com/ti/cli/internal/contextpack"
)

type BrainGateway struct {
    engine     *brain.Engine
    palace     *memory.Palace
    contextSvc *contextpack.Service
}

func (b *BrainGateway) GetContext(project, task string) (*ContextBundle, error)
func (b *BrainGateway) RecallMemory(query string, layer int) ([]Drawer, error)
func (b *BrainGateway) ObservePattern(event string) error
func (b *BrainGateway) GetSkill(name string) (*Skill, error)
func (b *BrainGateway) CreateHandoff(fromAgent, toAgent string, context, output string) (string, error)
func (b *BrainGateway) RecallHandoff(handoffID string) (*HandoffBundle, error)
```

**HTTP routes** (in `cmd/routerd/main.go`):
```go
mux.HandleFunc("/v1/brain/context", brainGateway.handleGetContext)
mux.HandleFunc("/v1/brain/memory/recall", brainGateway.handleRecallMemory)
mux.HandleFunc("/v1/brain/pattern/observe", brainGateway.handleObservePattern)
mux.HandleFunc("/v1/brain/skill", brainGateway.handleGetSkill)
mux.HandleFunc("POST /v1/brain/handoff", brainGateway.handleCreateHandoff)
mux.HandleFunc("GET /v1/brain/handoff/{id}", brainGateway.handleRecallHandoff)
```

**MCP server endpoint**:
```go
mux.HandleFunc("/mcp", brainGateway.handleMCP) // JSON-RPC for MCP
```

**Tasks**:
1. Create `Z:\Ti\router\layers\brain\` package
2. Wire to existing `internal/brain`, `internal/memory`, `internal/contextpack`
3. Add HTTP handlers in `cmd/routerd/main.go`
4. Add MCP server (JSON-RPC over HTTP)
5. Test: `curl http://localhost:1806/v1/brain/context?project=test`

---

### Phase 3: Ti CLI Wrapper (5-7 days) - MEDIUM-HIGH RISK

**Goal**: Build meta-agent CLI wrapper with agent orchestration.

**File**: `Z:\Ti\CLI\cmd\agents\` (NEW)

**Key Insight**: Devin has 3 handoff types:
- `cloud_handoff`: Local → Cloud (EXPLICIT user only)
- `run_subagent`: Parallel/Sequential subtasks (automatic, recommended)
- `Orchestrate`: Sequential workflow (via best_source pattern)

**Design Choice**: Use `run_subagent` for orchestration + Ti Brain for persistent handoff storage.

```go
// agents/claude.go
package agents

func WrapClaude(prompt string, opts WrapperOptions) error {
    // 1. Call Ti Router brain
    ctx := callBrain("/v1/brain/context", project, task)

    // 2. Build enhanced prompt
    enhanced := injectContext(prompt, ctx)

    // 3. Spawn claude
    cmd := exec.Command("claude", "--prompt", enhanced)
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr

    // 4. Capture output via PTY (for stream learning)
    output, err := runWithPTY(cmd)

    // 5. Feed back to brain
    callBrain("/v1/brain/pattern/observe", output)

    return err
}

// agents/orchestrator.go - NEW
package agents

func Orchestrate(task string, workflow []string) error {
    var lastOutput string

    for i, agentName := range workflow {
        // 1. Pack context from Ti Brain
        context := callBrain("/v1/brain/context", project, task)

        // 2. Load previous handoff (if not first agent)
        if i > 0 {
            handoff := callBrain("/v1/brain/handoff/"+workflow[i-1]+"-"+agentName)
            context = mergeContext(context, handoff)
        }

        // 3. Launch subagent (Devin's run_subagent)
        // For Claude/Codex/Amp: spawn process
        // For Devin: use run_subagent tool
        subagent := launchAgent(agentName, task, context)
        output := subagent.wait()

        // 4. Store handoff in Ti Brain (persistent, cross-session)
        callBrain("POST /v1/brain/handoff", fromAgent, toAgent, context, output)
        lastOutput = output
    }

    return nil
}
```

**Commands**:
```bash
ti claude "do X"           # Wraps Claude Code
ti devin "do Y"            # Wraps Devin
ti codex "do Z"            # Wraps Codex
ti amp "do W"              # Wraps Ampcode
ti chat "do K"             # Direct route via Ti Router → best agent
ti orchestrate "task" planner,tdd-guide,code-reviewer  # Multi-agent workflow
ti brain serve             # Standalone brain server (if Ti Router not running)
ti brain mcp               # MCP server endpoint
```

**Tasks**:
1. Create `cmd/agents/` package
2. Implement wrappers for Claude, Devin, Codex, Ampcode
3. Implement orchestrator with run_subagent pattern
4. Integrate Ti Brain handoff API
5. PTY capture for stream learning
6. Custom prompt injection logic
7. Add to `cmd/root.go`

---

### Phase 4: Cross-Agent Brain Learning (Ongoing)

**Goal**: Brain learns from all agents.

---

## 🔍 Devin Handoff Integration (Critical Design Decision)

### Devin's Handoff Types (from DEVIN_HANDOFF_ANALYSIS.md):

| Type | Purpose | Trigger | Use in Ti Hybrid |
|------|---------|---------|------------------|
| **cloud_handoff** | Local → Cloud transfer | EXPLICIT user only | Optional cloud offload |
| **run_subagent** | Parallel/Sequential subtasks | Agent decision | **Primary orchestration** |
| **Orchestrate** | Sequential workflow | Agent/Automatic | Via best_source pattern |

### Design Decision:

**Use `run_subagent` for agent orchestration + Ti Brain for persistent handoff storage.**

**Why not `cloud_handoff`?**
- ❌ EXPLICIT user request only (not automatic)
- ❌ Separate filesystem (path issues)
- ❌ Not suitable for automatic orchestration

**Why `run_subagent` + Ti Brain?**
- ✅ Automatic orchestration (no user trigger needed)
- ✅ Same filesystem (no path issues)
- ✅ Supports parallel execution
- ✅ Ti Brain provides persistent storage (cross-session)
- ✅ Works with ALL agents (Claude, Devin, Codex, Ampcode)

### Implementation Pattern:

```go
// Ti CLI Wrapper orchestration
func (w *Wrapper) Orchestrate(task string, workflow []string) {
    for i, agentName := range workflow {
        // 1. Pack context from Ti Brain
        context := w.brainClient.GetContext(agentName)

        // 2. Load previous handoff (if not first agent)
        if i > 0 {
            handoff := w.brainClient.GetHandoff(workflow[i-1], agentName)
            context = mergeContext(context, handoff)
        }

        // 3. Launch subagent (Devin's run_subagent)
        subagent := w.runSubagent("subagent_general", task, context)
        output := subagent.wait()

        // 4. Store handoff in Ti Brain (persistent)
        w.brainClient.StoreHandoff(agentName, workflow[i+1], output)
    }
}
```

### MCP Tool Updates:

```json
{
  "name": "ti_create_handoff",
  "description": "Store handoff document in Ti's MemPalace",
  "parameters": {
    "from_agent": "string",
    "to_agent": "string",
    "context": "object",
    "output": "string"
  }
}

{
  "name": "ti_recall_handoff",
  "description": "Recall handoff document from Ti's MemPalace",
  "parameters": {
    "from_agent": "string",
    "to_agent": "string"
  }
}
```

---

## 🎯 Cross-Agent Brain Learning (Ongoing)

**Goal**: Brain learns from all agents.

**Flow**:
```
Each agent run → captured → analyzed → patterns extracted → MemPalace

Pattern types:
- Code patterns (refactoring, fixes)
- Tool usage patterns
- Error patterns + solutions
- Project-specific conventions
- Cross-project universal patterns
```

**Promotion logic** (existing `internal/memory/promotion.go`):
```
L3 (Deep Search) → L2 (On-Demand) when frequently accessed
L2 → L1 (Essential) when applicable to many tasks
L1 → L0 (Identity) when fundamental
```

---

## 🎯 SUCCESS METRICS

### Smoothness (Ampcode-like)
- ✅ Context window stays small (MemPalace L0+L1 ~600-900 tokens)
- ✅ Handoff API for thread switching (persistent storage in Ti Brain)
- ✅ Agent orchestration via run_subagent (automatic)
- ✅ Edit & Restore via wrapper

### Versatility (Codex-like)
- ✅ Plugin architecture (existing)
- ✅ Skills system (Codex-style directories)
- ✅ Multi-agent support
- ✅ AGENTS.md universal

### Quality (Claude Code-like)
- ✅ Brain RL learning (cross-session)
- ✅ Pattern recognition
- ✅ Multi-level rules (managed/project/user)
- ✅ Cross-agent insights

### NEW: Cross-Agent Intelligence
- ✅ Devin learns from Claude sessions
- ✅ Codex skills shared with Ampcode
- ✅ Universal patterns promoted to global brain

---

## 🚀 START PARALLEL

### Track A: AGENTS.md Migration (1 day)
- [ ] Create `AGENTS.md` từ `.devin/context/project-context.md`
- [ ] Update `.gitignore`
- [ ] Document bridge structure

### Track B: Ti Router Brain Endpoints Prototype (3 days)
- [ ] Create `Z:\Ti\router\layers\brain\`
- [ ] Wire to existing Ti CLI brain/memory/contextpack
- [ ] Add HTTP handlers (context, memory, pattern, skill, handoff)
- [ ] Implement handoff storage in MemPalace (L2/L3)
- [ ] Add MCP server (JSON-RPC over HTTP)
- [ ] Test endpoints

### Track C: Ti CLI Wrapper Prototype (3 days)
- [ ] Create `cmd/agents/claude.go`
- [ ] Implement basic prompt injection
- [ ] Implement orchestrator with run_subagent pattern
- [ ] Integrate Ti Brain handoff API
- [ ] Test với Claude Code
- [ ] Test workflow: planner → code-reviewer
- [ ] Iterate

---

## 💡 KEY DESIGN PRINCIPLES

1. **Native UX First** - Agents work normally, Ti enhances
2. **Optional Integration** - Agents work without Ti Brain
3. **Reuse Existing** - Ti CLI, Ti Router, Ti Brain đã có sẵn
4. **Universal Format** - AGENTS.md là chuẩn chung
5. **Cross-Agent Learning** - Killer feature, no one else has it
6. **Hybrid Storage** - Local (per-project) + Global (cross-project)
7. **MCP-Native** - Future-proof với MCP standard

---

**Last Updated**: 2026-04-28
**Status**: Design approved, ready to implement
**References**:
- `DEVIN_HANDOFF_ANALYSIS.md` - Devin's handoff mechanism analysis
- `ARCHITECT_ANALYSIS.md` - Architect agent review
- `AGENT_QUALITY_REVIEW.md` - Self-evaluation of architect analysis
