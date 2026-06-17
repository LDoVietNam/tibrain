# Ti Crew - GoClaw Integration Plan

> **Version**: 1.0.0
> **Last Updated**: 2026-05-05
> **Purpose**: Tích hợp GoClaw làm nền tảng cho Ti Crew để hoàn thiện nhanh hơn

---

## 📋 Executive Summary

**Decision**: ✅ **Sử dụng GoClaw làm nền tảng cho Ti Crew**

**Rationale**:
- GoClaw có **tất cả features** mà Ti Crew cần
- Production-tested, multi-tenant, secure
- Timeline: **2-3 weeks** (vs 8 weeks build from scratch)
- Effort: **4-6 person-weeks** (vs 16 person-weeks)
- Can fork/customize cho Ti-specific needs

---

## 🎯 Integration Strategy

### Option 1: Fork GoClaw (RECOMMENDED)

**Approach**: Fork goclaw repository, tùy chỉnh cho Ti ecosystem

**Pros**:
- Full control over codebase
- Can remove unused features (messaging channels, etc.)
- Can add Ti-specific integrations (Beads, Knowledge Graph Memory, etc.)
- Can rebrand as "Ti Crew"

**Cons**:
- Need to maintain fork
- Upstream updates cần merge manual

**Timeline**: 2-3 weeks

### Option 2: Use as Dependency

**Approach**: Import goclaw as Go module, wrap with Ti-specific layer

**Pros**:
- Automatic upstream updates
- Less maintenance burden

**Cons**:
- Limited customization
- Cannot remove unused features
- Complex dependency management

**Timeline**: 3-4 weeks

### ✅ Recommendation: **Option 1 (Fork GoClaw)**

---

## 📁 Ti Crew Structure (dựa trên GoClaw)

```
apps/
├── cli/                # ticlaw
├── router/             # Router service
├── tui/                # Terminal UI
├── automation/         # Automation tools
├── mcp/                # MCP servers
└── ticrew/             # ✅ Ti Crew (forked from GoClaw)
    ├── cmd/            # CLI commands (goclaw → ticrew)
    │   ├── root.go    # Main entry point
    │   ├── onboard.go # Onboarding wizard
    │   ├── migrate.go # Database migrations
    │   └── gateway.go # Gateway startup
    ├── internal/
    │   ├── agent/     # Agent loop (think→act→observe)
    │   ├── bootstrap/ # System prompts + seeding
    │   ├── bus/       # Event bus
    │   ├── cache/     # Caching layer
    │   ├── config/    # Config loading
    │   ├── crypto/    # AES-256-GCM encryption
    │   ├── cron/      # Cron scheduling
    │   ├── gateway/   # WS + HTTP server
    │   │   └── methods/ # RPC handlers
    │   ├── knowledgegraph/ # Knowledge graph
    │   ├── mcp/       # MCP bridge/server
    │   ├── memory/    # Memory system (pgvector)
    │   ├── permissions/ # RBAC
    │   ├── providers/ # LLM providers
    │   ├── scheduler/ # Lane-based concurrency
    │   ├── sessions/  # Session management
    │   ├── skills/    # SKILL.md loader + BM25
    │   ├── store/     # Store interfaces + pg/ implementations
    │   ├── tasks/     # Task management
    │   ├── tools/     # Tool registry
    │   └── tracing/   # LLM call tracing
    ├── pkg/
    │   ├── protocol/  # Wire types
    │   └── browser/   # Browser automation
    ├── migrations/    # PostgreSQL migrations
    ├── ui/
    │   └── web/       # React SPA (Vite, Tailwind, Radix UI)
    ├── go.mod
    ├── go.sum
    ├── Makefile
    └── README.md
```

---

## 🔧 Customizations cho Ti Ecosystem

### 1. Agent Discovery from Ti Agents Directory

**Current**: GoClaw agents được quản lý qua database
**Ti Crew**: Auto-discover từ `content/agents/agents/` (104 agents)

**Implementation**:
```go
// internal/bootstrap/ti_agents.go
func DiscoverTiAgents(ctx context.Context, agentDir string) ([]*Agent, error) {
    entries, err := os.ReadDir(agentDir)
    if err != nil {
        return nil, err
    }

    agents := []*Agent{}
    for _, entry := range entries {
        if !strings.HasSuffix(entry.Name(), ".md") {
            continue
        }

        // Parse YAML frontmatter
        metadata, err := parseAgentFile(filepath.Join(agentDir, entry.Name()))
        if err != nil {
            continue
        }

        // Convert to GoClaw Agent format
        agent := &Agent{
            ID:          metadata.ID,
            Name:        metadata.Name,
            Type:        "predefined", // Shared context
            Personality: metadata.Description,
            Provider:    metadata.Model,
            Tools:       metadata.Tools,
            MCPs:        metadata.MCPs,
            // ... other fields
        }

        agents = append(agents, agent)
    }

    return agents, nil
}
```

### 2. Beads Integration

**Current**: GoClaw không có Beads
**Ti Crew**: Integrate Beads system cho task tracking

**Implementation**:
```go
// internal/beads/client.go
type BeadsClient struct {
    dataDir string
}

func (bc *BeadsClient) Log(task string, agent string, status string) error {
    // Call bd.exe or direct JSON write
    cmd := exec.Command("Z:\\02_CORE\\_cli\\bin\\bd.exe",
        "log",
        "--task", task,
        "--agent", agent,
        "--status", status,
        "--type", "coding",
        "--domain", "general")
    return cmd.Run()
}
```

### 3. Knowledge Graph Memory Integration

**Current**: GoClaw có knowledge graph riêng
**Ti Crew**: Integrate với server-memory (Ti's Knowledge Graph Memory)

**Implementation**:
```go
// internal/memory/server_memory.go
type ServerMemoryClient struct {
    mcpClient *mcp.Client
}

func (smc *ServerMemoryClient) Search(query string) ([]*Entity, error) {
    // Call MCP server-memory
    return smc.mcpClient.Call("search_nodes", map[string]any{
        "query": query,
    })
}
```

### 4. Ti-Specific Tools

**Current**: GoClaw có built-in tools
**Ti Crew**: Add Ti-specific tools

**Tools to add**:
- `ti_convert`: Convert configs to Go code
- `ti_sync`: Sync skills to hub.db
- `ti_router`: Router management
- `ti_mcp`: MCP server management

### 5. Remove Unused Features

**Features to remove** (to reduce complexity):
- Messaging channels (Telegram, Discord, Slack, Zalo, WhatsApp)
- Desktop edition (Wails)
- Browser automation (Rod)
- TTS/STT
- Media generation (image/audio/video)

**Rationale**: Ti Crew là internal tool, không cần public-facing features

---

## 🗓️ Implementation Plan (2-3 weeks)

### Week 1: Fork + Setup

| Task | Duration | Owner |
|------|----------|-------|
| Fork goclaw to Ti repository | 0.5 day | Senior Dev |
| Rename branding (goclaw → ticrew) | 0.5 day | Senior Dev |
| Setup PostgreSQL + pgvector | 0.5 day | DevOps |
| Run onboard wizard + seed data | 0.5 day | Senior Dev |
| Test basic gateway functionality | 0.5 day | Senior Dev |
| **Total** | **2.5 days** | - |

### Week 2: Ti-Specific Integrations

| Task | Duration | Owner |
|------|----------|-------|
| Implement Ti agent discovery | 1 day | Senior Dev |
| Integrate Beads system | 0.5 day | Senior Dev |
| Integrate Knowledge Graph Memory | 1 day | Senior Dev |
| Add Ti-specific tools | 1 day | Senior Dev |
| Test end-to-end (discover 104 agents) | 0.5 day | Senior Dev |
| **Total** | **4 days** | - |

### Week 3: Cleanup + Documentation

| Task | Duration | Owner |
|------|----------|-------|
| Remove unused features (channels, desktop) | 1 day | Senior Dev |
| Update documentation | 0.5 day | Senior Dev |
| Performance testing | 0.5 day | Senior Dev |
| Security review | 0.5 day | Senior Dev |
| Final testing + deployment | 0.5 day | Senior Dev |
| **Total** | **3 days** | - |

**Total Timeline**: **9.5 days (~2 weeks)**

---

## 📊 Updated Effort Comparison

| Phase | Original Plan | GoClaw Integration | Savings |
|-------|---------------|---------------------|---------|
| Foundation | 2 weeks | 0.5 week | **1.5 weeks** |
| Core Features | 2 weeks | 1 week | **1 week** |
| Advanced Features | 2 weeks | 0.5 week | **1.5 weeks** |
| UI & Tools | 2 weeks | 0.5 week | **1.5 weeks** |
| **Total** | **8 weeks** | **2 weeks** | **6 weeks** |

**Effort Savings**: **75%** (16 person-weeks → 4 person-weeks)

---

## 🎯 Success Criteria

- ✅ Load 104 agents từ `content/agents/agents/`
- ✅ Route tasks đến đúng agent với 80%+ accuracy
- ✅ Execute multi-agent workflows với agent teams
- ✅ Track metrics với OpenTelemetry
- ✅ Integrate Beads system
- ✅ Integrate Knowledge Graph Memory
- ✅ Deploy trong 2-3 weeks

---

## 📝 Next Steps

1. **Fork GoClaw**: Clone to `apps/ticrew/`
2. **Rename branding**: goclaw → ticrew
3. **Setup infrastructure**: PostgreSQL + pgvector
4. **Implement Ti integrations**: Agent discovery, Beads, Memory
5. **Test & deploy**: End-to-end testing

Bạn muốn tôi bắt đầu fork GoClaw và setup Ti Crew không?
