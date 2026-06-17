# 🔍 Devin Handoff Mechanism Analysis

> **Date**: 2026-04-27
> **Scope**: Compare Devin's `cloud_handoff` vs Ampcode's handoff

---

## 📊 Devin's Handoff Types

### 1. **cloud_handoff** (Native Tool)

**Tool Definition** (from Devin's system prompt):

```
Hand off a task to a remote cloud Devin session.
Use when user explicitly asks to hand off to a cloud agent/remote agent.
ONLY call when user mentions "handoff", "cloud agent", "remote devin", etc.
NEVER call without explicit user request.
```

**Characteristics**:
- **Purpose**: Transfer task from local Devin → cloud Devin session
- **Trigger**: EXPLICIT user request only (not automatic)
- **Context**: Git repo name + branch auto-included
- **Filesystem**: Separate (cloud agent has different structure)
- **Paths**: Use relative paths from repo root (NO absolute paths)

**When to use**:
- User explicitly says "handoff to cloud devin"
- User explicitly says "remote agent"
- User explicitly says "handing off to devin"

**When NOT to use**:
- Automatic agent-to-agent chaining
- Implicit workflow orchestration
- Without user request

---

### 2. **run_subagent** (Background/Foreground Subagents)

**Tool Definition** (from Devin's system prompt):

```
Launch an independent subagent to handle a task autonomously.
Profiles: builder, code-reviewer, subagent_explore, subagent_general
```

**Characteristics**:
- **Purpose**: Parallel or sequential subagent execution
- **Trigger**: Agent decision (can be automatic)
- **Profiles**:
  - `subagent_explore` (read-only, codebase exploration)
  - `subagent_general` (full access, code changes)
  - `builder` (build-focused)
  - `code-reviewer` (review-focused)
- **Modes**:
  - `is_background=true` → parallel, non-blocking
  - `is_background=false` → foreground, blocking

**When to use**:
- Self-contained multi-step tasks
- Parallel execution of independent subtasks
- Codebase exploration without main agent context
- Tasks requiring different profiles

**When NOT to use**:
- Single straightforward task
- Tasks that can be done in <3 trivial steps
- Purely informational requests

---

### 3. **Orchestrate Pattern** (from best_source)

**File**: `Z:\Ti\best_source\commands\orchestrate.md`

**Characteristics**:
- **Purpose**: Sequential agent workflow orchestration
- **Pattern**: agent1 → handoff → agent2 → handoff → agent3
- **Workflow Types**: feature, bugfix, refactor, security, custom
- **Handoff Document**: Structured markdown between agents

**Example**:
```
planner → tdd-guide → code-reviewer → security-reviewer
```

**When to use**:
- Complex multi-agent workflows
- Sequential specialized agents
- Feature implementation pipeline

---

## 🔍 Ampcode Handoff (Reference)

**Mechanism** (from research):

1. **Context Management with Handoff**
   - Share context between agents
   - Thread sharing
   - Shell mode persistence

2. **Handoff Document Format**
   - Structured context transfer
   - Previous agent output
   - Next agent requirements

---

## 📊 Comparison: Devin vs Ampcode Handoff

| Aspect | Devin cloud_handoff | Devin run_subagent | Ampcode Handoff |
|--------|-------------------|-------------------|-----------------|
| **Purpose** | Local → Cloud transfer | Parallel/Sequential subtasks | Agent-to-agent chaining |
| **Trigger** | EXPLICIT user only | Agent decision | Agent/Automatic |
| **Context Transfer** | Git repo + branch | Full context or read-only | Thread + context sharing |
| **Filesystem** | Separate (cloud) | Same (local) | Same (local) |
| **Profiles** | N/A (same Devin) | Explore/General/Builder/Reviewer | N/A (same Ampcode) |
| **Auto** | ❌ NEVER auto | ✅ Can be auto | ✅ Can be auto |
| **Use Case** | Cloud offload | Parallel work, exploration | Workflow orchestration |

---

## 🎯 Implications for Ti Hybrid Design

### Current Design Assumptions:

From `TI_HYBRID_DESIGN.md`:

```
Phase 3: Ti CLI Wrapper (Meta-Agent)
- Meta-Agent orchestration via handoff mechanism
- Agent-to-agent context sharing
```

**Problem**: Design assumed "handoff" = automatic agent chaining (Ampcode style).

**Reality**: Devin's `cloud_handoff` is EXPLICIT user-triggered only.

---

### Updated Design Recommendations:

#### Option A: Use `run_subagent` for Agent Orchestration ✅ RECOMMENDED

**Implementation**:
```go
// In Ti CLI Wrapper
func (w *Wrapper) Orchestrate(task string, agents []string) {
    // Use run_subagent for sequential agents
    subagent1 := w.launchSubagent("subagent_general", task_part1)
    output1 := subagent1.wait()

    // Handoff via structured document
    handoff := w.createHandoffDoc(output1, "agent2")
    subagent2 := w.launchSubagent("subagent_general", task_part2 + handoff)
    output2 := subagent2.wait()
}
```

**Pros**:
- ✅ Works with Devin's native capabilities
- ✅ Can be automatic (no user trigger needed)
- ✅ Supports parallel execution
- ✅ Same filesystem (no path issues)

**Cons**:
- ⚠️ Subagents run in same session (not isolation like cloud)
- ⚠️ Limited to Devin's profiles

---

#### Option B: Use `cloud_handoff` for Cloud Offload ⚠️ LIMITED USE

**Implementation**:
```go
// In Ti CLI Wrapper
func (w *Wrapper) HandoffToCloud(task string) {
    // Only when user EXPLICITLY requests
    if w.userRequestedCloudHandoff() {
        w.callCloudHandoff(task, context)
    }
}
```

**Pros**:
- ✅ True cloud isolation
- ✅ Separate filesystem (clean slate)
- ✅ More compute resources

**Cons**:
- ❌ EXPLICIT user request only (not automatic)
- ❌ Path differences (need relative paths)
- ❌ Not suitable for automatic orchestration

---

#### Option C: Implement Custom Handoff Layer 🔧 HYBRID

**Implementation**:
```go
// In Ti Brain Server
type HandoffManager struct {
    memory  *MemPalace
    context *ContextPack
}

func (hm *HandoffManager) CreateHandoff(fromAgent, toAgent string, output string) HandoffDoc {
    // Store in Ti's MemPalace (L2/L3)
    hm.memory.Store("handoff:"+fromAgent+":"+toAgent, output)

    // Create structured document
    return HandoffDoc{
        From:    fromAgent,
        To:      toAgent,
        Context: hm.context.PackForAgent(toAgent),
        Output:  output,
    }
}

func (hm *HandoffManager) RecallHandoff(fromAgent, toAgent string) string {
    return hm.memory.Recall("handoff:"+fromAgent+":"+toAgent)
}
```

**Pros**:
- ✅ Persistent handoff storage (Ti's MemPalace)
- ✅ Cross-session handoff (survives restarts)
- ✅ Works with ALL agents (Claude, Devin, Codex, Ampcode)
- ✅ Leverages Ti's learning engine

**Cons**:
- ⚠️ Requires Ti Brain Server implementation
- ⚠️ More complexity (but justified)

---

## 🎯 RECOMMENDED APPROACH

### Phase 2 (Brain Endpoints) - Add Handoff API

**New Endpoints**:

```yaml
/v1/brain/handoff:
  POST:
    description: Store handoff document in MemPalace
    body:
      from_agent: string
      to_agent: string
      context: object
      output: string
    response:
      handoff_id: string

/v1/brain/handoff/{handoff_id}:
  GET:
    description: Recall handoff document
    response:
      from_agent: string
      to_agent: string
      context: object
      output: string
```

### Phase 3 (Ti CLI Wrapper) - Use `run_subagent` + Ti Handoff API

**Pattern**:
```go
func (w *Wrapper) Orchestrate(task string, workflow []string) {
    var lastOutput string

    for i, agentName := range workflow {
        // Pack context from Ti Brain
        context := w.brainClient.GetContext(agentName)

        // If not first agent, load previous handoff
        if i > 0 {
            handoff := w.brainClient.GetHandoff(workflow[i-1], agentName)
            context = mergeContext(context, handoff)
        }

        // Launch subagent (Devin's run_subagent)
        subagent := w.runSubagent("subagent_general", task, context)
        output := subagent.wait()

        // Store handoff in Ti Brain
        w.brainClient.StoreHandoff(agentName, workflow[i+1], output)
        lastOutput = output
    }

    return lastOutput
}
```

### Phase 4 (Cloud Handoff Integration) - Optional

**When user explicitly requests cloud**:
```go
func (w *Wrapper) HandleCloudHandoffRequest(task string) {
    // User explicitly asked for cloud
    if w.userRequestedCloud {
        // Use Devin's cloud_handoff tool
        // But pack context from Ti Brain first
        context := w.brainClient.GetContext("cloud")
        w.callCloudHandoff(task, context)
    }
}
```

---

## 📋 Updated Implementation Plan

### Phase 2a: Brain Handoff Endpoints (Day 1-2)

1. **Define OpenAPI spec** for `/v1/brain/handoff` endpoints
2. **Implement handoff storage** in MemPalace (L2/L3)
3. **Add handoff retrieval** API
4. **Test** with curl/Postman

### Phase 2b: Integration with Devin (Day 3-4)

1. **Create Ti CLI Wrapper prototype**
2. **Use `run_subagent`** for agent orchestration
3. **Integrate Ti Brain handoff API**
4. **Test** with simple workflow: `planner → code-reviewer`

### Phase 2c: Cloud Handoff Support (Day 5-7) - OPTIONAL

1. **Add cloud handoff trigger** (explicit user request only)
2. **Pack context** from Ti Brain before handoff
3. **Handle relative paths** for cloud filesystem
4. **Test** with actual cloud session

---

## 🔧 MCP Tool Updates

**New MCP Tools**:

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

{
  "name": "ti_list_handoffs",
  "description": "List all handoffs for current session",
  "parameters": {}
}
```

---

## ✅ Benefits of Updated Design

1. **Leverages Devin's native capabilities** (`run_subagent` for orchestration)
2. **Adds persistent handoff storage** via Ti's MemPalace (cross-session)
3. **Supports cloud handoff** when explicitly requested
4. **Works with ALL agents** (Claude, Devin, Codex, Ampcode)
5. **Maintains Ti's learning value** (handoffs contribute to brain)

---

## 🚨 Risks & Mitigations

### Risk 1: `run_subagent` Isolation
- **Issue**: Subagents share session context (not true isolation)
- **Mitigation**: Use Ti Brain for persistent storage + explicit context packing

### Risk 2: Cloud Handoff Path Issues
- **Issue**: Cloud filesystem different from local
- **Mitigation**: Use relative paths from repo root only

### Risk 3: Handoff Duplication
- **Issue**: Both Ti Brain and agent's native handoff
- **Mitigation**: Ti Brain as source of truth, agent handoff as cache

---

## 📊 Summary

| Handoff Type | Use in Ti Hybrid | Implementation |
|-------------|------------------|----------------|
| **Devin cloud_handoff** | Cloud offload (explicit) | Optional Phase 4 |
| **Devin run_subagent** | Agent orchestration | Phase 3 (primary) |
| **Ti Brain Handoff API** | Persistent storage | Phase 2 (core) |
| **Ampcode-style handoff** | Workflow pattern | Via Ti Brain API |

---

**Conclusion**: Devin HAS handoff capability but DIFFERENT from Ampcode. Use `run_subagent` for orchestration + Ti Brain for persistent handoff storage. `cloud_handoff` only for explicit cloud offload.

---

**Last Updated**: 2026-04-27
