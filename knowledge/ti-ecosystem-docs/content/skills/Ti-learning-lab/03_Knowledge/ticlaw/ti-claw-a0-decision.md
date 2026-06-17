# Ti Claw Phase A0 Decision — Architecture & Module Mapping

> **Date**: 2026-05-05  
> **Status**: Completed  
> **Decision**: **Option B — Plugin trong Ti CLI**  

---

## Executive Summary

After auditing GoClaw vs Ti CLI modules, the optimal architecture is **Ti Claw as a plugin within Ti CLI**. This maximizes code reuse, minimizes duplication, and leverages existing infrastructure.

---

## Architecture Decision

### Chosen: Option B — Plugin in Ti CLI

**Rationale**:
- Ti CLI already has **74% of agent infrastructure** (agentcore, brain, memory, MCP, tools)
- GoClaw's agent loop (38KB) is more formal than CLI's (6KB) → **port GoClaw patterns only where CLI lacks**
- Single codebase → easier maintenance
- Plugin system exists (gRPC microkernel)

**Structure**:
```
apps/cli/internal/plugins/ticlaw/
├── agent/          # Port GoClaw agent loop patterns
├── scheduler/      # Port lane-based scheduler
├── skills/         # Port SKILL.md + BM25
├── bootstrap/      # Port SOUL.md, IDENTITY.md
└── cmd/            # CLI commands: ti agent run/list/skill
```

---

## Module Mapping Decision

### Reuse from Ti CLI (✅ Already Available)

| Module | Ti CLI Path | Purpose | Verdict |
|--------|-------------|---------|---------|
| Agent Execution | `internal/agent/` | Basic think→act→observe | **Reuse + extend** |
| Brain/Learning | `internal/brain/` | RL, learning engine | **Reuse as-is** |
| Memory | `internal/memory/` | Palace memory | **Reuse as-is** |
| MCP Bridge | `internal/mcp/` | MCP server/client | **Reuse as-is** |
| Tools | `internal/tools/` | File tools | **Reuse as-is** |
| Providers | `internal/providers/router.go` | LLM gateway | **Reuse Router** |
| Sessions | `internal/session/` | Session management | **Reuse as-is** |
| Permissions | `internal/permission/` | Safety checks | **Reuse as-is** |

### Port from GoClaw (🔄 Need to Add)

| Module | GoClaw Path | Purpose | Implementation |
|--------|-------------|---------|----------------|
| Agent Loop (formal) | `internal/agent/loop.go` | Think→Act→Observe with delegation | **Port patterns** into CLI's agent |
| Scheduler | `internal/scheduler/` | Lane-based concurrency | **Port as new module** |
| Skills | `internal/skills/` | SKILL.md + BM25 search | **Port as new module** |
| Bootstrap | `internal/bootstrap/` | SOUL.md, IDENTITY.md seeding | **Port as new module** |
| Context Files | Various | Agent-level config | **Port pattern** |

### Remove from GoClaw (❌ Not Needed)

| Module | Reason |
|--------|--------|
| `internal/channels/` | External messaging (Telegram, Discord) |
| `internal/tts/` | Text-to-speech |
| `internal/oauth/` | OAuth (Router already has 12 OAuth) |
| `internal/permissions/` | RBAC (Lite edition doesn't need) |
| `internal/knowledgegraph/` | Use Ti's existing KG |
| `internal/sandbox/` | Code execution (optional later) |

---

## POC Results

### Router Compatibility ✅

**Test**: `TestTiRouterProvider` confirms Ti Router is OpenAI-compatible
- Basic completion works
- Provider interface matches
- Tool calling ready (needs MCP tools configured)

### MCP Bridge ✅

**Status**: Ti CLI already has MCP server bridge
- 24+ tools in `apps/mcp/`
- Ready for agent integration

---

## Implementation Plan

### Phase A2: Fork/Plugin Setup

1. **Create plugin structure**:
   ```bash
   mkdir -p apps/cli/internal/plugins/ticlaw/{agent,scheduler,skills,bootstrap,cmd}
   ```

2. **Port GoClaw agent loop patterns**:
   - Extract delegation logic from `goclaw/internal/agent/loop.go`
   - Integrate with CLI's `internal/agent/loop.go`

3. **Port scheduler**:
   - Lane-based concurrency for multi-agent scenarios

4. **Port skills system**:
   - SKILL.md format + BM25 search
   - Compatible with existing `content/skills/`

### Phase A3: Ti Integration

1. **Provider**: Point to Router :1807 (already working)
2. **MCP**: Register existing Ti MCP tools
3. **Memory**: Bridge to knowledge-graph.jsonl
4. **CLI Commands**: `ti agent run/list/skill`

---

## Success Criteria

- [x] Architecture decision documented
- [x] Module mapping completed
- [x] POC: Router provider works
- [ ] Plugin structure created
- [ ] Agent loop patterns ported
- [ ] Basic agent runs end-to-end

---

## Next Steps

1. **Start Phase A1**: Quick Wins (Notion Task Management) — can proceed in parallel
2. **Begin Phase A2**: Create plugin structure
3. **Port GoClaw scheduler** (highest priority missing piece)

---

**Conclusion**: Plugin approach maximizes reuse while adding only essential GoClaw patterns. Ti Claw becomes a "supercharged" agent plugin within Ti CLI ecosystem.
