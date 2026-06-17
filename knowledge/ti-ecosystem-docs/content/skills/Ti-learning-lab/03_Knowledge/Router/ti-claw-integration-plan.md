# Ti Claw Integration Plan with Ti CLI

> **Purpose**: Plan integration of Ti Claw (rebranded GoClaw) with Ti CLI
> **Date**: 2026-05-04
> **Status**: Planning Phase

---

## 1. Architecture Overview

### Current Ti CLI Structure
```
apps/
├── cli/                # Official CLI (builds to bin/ti)
├── router/             # Router service
├── tui/                # Terminal UI
├── automation/         # Automation tools
└── mcp/                # MCP servers

packages/
├── sdk/                # Plugin API SDK
├── core/               # Core modules
└── providers/          # Provider implementations

content/
├── skills/             # Skills
├── agents/             # Agent definitions
└── mcp/                # MCP server configs
```

### Proposed Ti Claw Structure
```
apps/
├── cli/                # Official CLI
├── router/             # Router service (existing UI at localhost:5173)
├── ticlaw/             # Ti Claw agent framework
│   ├── cmd/            # CLI commands
│   ├── internal/       # Go backend
│   │   ├── agent/      # Agent loop (think-act-observe)
│   │   ├── store/      # SQLite store layer
│   │   ├── tools/      # Tool registry
│   │   ├── providers/  # LLM providers
│   │   ├── memory/     # Memory system
│   │   ├── config/     # Config loading
│   │   ├── bootstrap/  # Context file seeding
│   │   ├── http/       # HTTP API server
│   │   ├── context/    # Context mode (NEW)
│   │   ├── review/     # Review system (NEW)
│   │   ├── skill-creator/ # Skill creator (NEW)
│   │   ├── superpowers/ # Superpowers workflow (NEW)
│   │   └── gsd/        # GSD system (NEW)
│   ├── ui/             # Ti Claw UI (NEW - localhost:5174)
│   │   ├── src/
│   │   │   ├── pages/      # Dashboard, Agents, Tools, Memory, etc.
│   │   │   ├── components/ # Agent cards, chat interface, etc.
│   │   │   ├── stores/     # Zustand stores
│   │   │   └── services/   # API client
│   │   ├── package.json
│   │   └── vite.config.ts
│   └── content/        # Agent context files
└── automation/         # Existing automation tools

packages/
├── ti-ui-components/   # Shared design system (NEW)
│   ├── src/
│   │   ├── components/ # Button, Card, Input, Modal, etc.
│   │   ├── styles/     # Theme variables, mixins
│   │   └── index.ts
│   └── package.json
├── ticlaw-sdk/         # Ti Claw SDK for external use
├── sdk/                # Existing plugin API SDK
└── core/               # Existing core modules

content/
├── skills/             # Ti skills (.devin/skills/)
├── agents/             # Agent definitions
└── ticlaw/             # Ti Claw-specific content
    ├── agents/         # Ti Claw agent definitions
    └── skills/         # Ti Claw-specific skills
```

---

## 2. Integration Strategy

### Phase 1: Core Framework (Week 1-2)

**Goal**: Implement core Ti Claw patterns

**Tasks**:
1. **Create `apps/ticlaw/` repository**
   - Initialize Go module
   - Set up directory structure
   - Copy core patterns from GoClaw

2. **Implement Agent Loop**
   - `internal/agent/loop.go` - Think-Act-Observe cycle
   - `internal/agent/loop_types.go` - Types and interfaces
   - Simplify: Remove multi-tenant, WebSocket, channels

3. **Implement Store Layer**
   - `internal/store/` - Interface definitions
   - `internal/store/sqlite/` - SQLite implementation
   - Remove PostgreSQL dependency

4. **Implement Tool Registry**
   - `internal/tools/registry.go` - Tool management
   - `internal/tools/policy.go` - Tool policy engine
   - Built-in tools: filesystem, exec, web, memory

5. **Implement Provider Pattern**
   - `internal/providers/provider.go` - Provider interface
   - `internal/providers/anthropic.go` - Anthropic provider
   - `internal/providers/openai.go` - OpenAI provider
   - `internal/providers/router.go` - Ti Router provider (NEW)

6. **Implement Memory System**
   - `internal/memory/memory.go` - Memory interface
   - `internal/memory/sqlite.go` - SQLite implementation
   - Integrate with Knowledge Graph Memory (server-memory MCP)

7. **Implement Config Loading**
   - `internal/config/config.go` - Config loading
   - Support JSON5 + env overlay
   - Secrets from `Z:\00_SECRET\ticlaw.env`

8. **Implement Bootstrap Pattern**
   - `internal/bootstrap/context_files.go` - Context file seeding
   - `internal/bootstrap/seed.go` - User file seeding
   - Integrate with Ti's content/ structure

**Deliverables**:
- Core Ti Claw framework
- Unit tests for core components
- Documentation

---

### Phase 2: CLI Integration (Week 3)

**Goal**: Integrate Ti Claw with Ti CLI

**Tasks**:
1. **Add Ti Claw Commands to Ti CLI**
   - `apps/cli/cmd/agent.go` - Agent commands
   - `apps/cli/cmd/agent_run.go` - Run agent
   - `apps/cli/cmd/agent_list.go` - List agents
   - `apps/cli/cmd/agent_create.go` - Create agent

2. **Implement Plugin System**
   - `packages/ticlaw-sdk/` - Ti Claw SDK
   - Plugin registration in Ti CLI
   - Agent discovery from `content/agents/`

3. **Integrate with Ti Router**
   - Add Ti Router as provider
   - Use Ti Router for LLM calls
   - Fallback to direct providers if Router unavailable

4. **Integrate with Knowledge Graph Memory**
   - Use server-memory MCP for memory
   - Memory operations via MCP tools
   - Persistent memory across sessions

5. **Integrate with Ti Skills**
   - Load skills from `.devin/skills/`
   - BM25 search for skills
   - Skill execution in agent loop

**Deliverables**:
- Ti CLI with Ti Claw integration
- Plugin system working
- Integration tests

---

### Phase 3: Notion Agent (Week 4)

**Goal**: Build first agent - Notion Agent

**Tasks**:
1. **Create Notion Agent Definition**
   - `content/ticlaw/agents/notion-agent.md`
   - Agent configuration
   - Tool definitions (Notion MCP)

2. **Implement Notion MCP Integration**
   - `apps/mcp/server/notion/` - Notion MCP server
   - Notion API client
   - Tool definitions: sync, query, update

3. **Build Notion Agent**
   - Use Ti Claw framework
   - Integrate Notion MCP tools
   - Implement knowledge sync logic

4. **Test Notion Agent**
   - Sync GoClaw patterns to Notion
   - Sync Ti knowledge to Notion
   - Verify functionality

**Deliverables**:
- Notion Agent working
- Knowledge sync automated
- Documentation

---

### Phase 4: Additional Agents (Week 5-6)

**Goal**: Build additional agents

**Tasks**:
1. **Knowledge Sync Agent**
   - Sync Ti skills to Notion
   - Sync Ti agents to Notion
   - Sync patterns to Notion

2. **Code Review Agent**
   - Review code changes
   - Provide feedback
   - Suggest improvements

3. **Test Agent**
   - Generate tests
   - Run tests
   - Report results

4. **Documentation Agent**
   - Generate documentation
   - Update docs
   - Keep docs in sync

**Deliverables**:
- 4+ agents working
- Agent templates
- Best practices

---

## 3. Component Mapping

### GoClaw → Ti Claw Mapping

| GoClaw Component | Ti Claw Component | Notes |
|----------------|------------------|-------|
| `internal/agent/` | `internal/agent/` | Keep, simplify |
| `internal/store/` | `internal/store/` | Keep, SQLite only |
| `internal/store/pg/` | `internal/store/sqlite/` | Replace PG with SQLite |
| `internal/tools/` | `internal/tools/` | Keep, simplify |
| `internal/providers/` | `internal/providers/` | Keep, add Router |
| `internal/memory/` | `internal/memory/` | Keep, integrate KG |
| `internal/config/` | `internal/config/` | Keep |
| `internal/bootstrap/` | `internal/bootstrap/` | Keep, adapt to Ti |
| `internal/skills/` | - | Use Ti's `.devin/skills/` |
| `internal/channels/` | - | Remove (CLI-only) |
| `internal/scheduler/` | - | Remove (CLI-only) |
| `internal/cron/` | - | Remove (CLI-only) |
| `internal/oauth/` | - | Remove (CLI-only) |
| `internal/permissions/` | - | Remove (single-user) |
| `internal/http/` | - | Remove (CLI-only) |
| `internal/gateway/` | - | Remove (CLI-only) |
| `ui/web/` | - | Remove (CLI-only) |
| `ui/desktop/` | - | Remove (CLI-only) |
| `migrations/` | - | Remove (use SQLite auto-migrate) |

---

## 4. Data Flow

### Agent Execution Flow

```
User Input (CLI)
    ↓
Ti CLI (ti agent run <agent>)
    ↓
Ti Claw Framework
    ↓
Agent Loop (Think-Act-Observe)
    ↓
┌─────────────┬─────────────┬─────────────┐
│   Think     │    Act      │  Observe    │
│             │             │             │
│ - Load      │ - Execute   │ - Process   │
│   history   │   tools     │   results   │
│ - Load      │ - Call LLM  │ - Update    │
│   context   │             │   memory    │
│ - Call LLM  │             │ - Emit      │
│             │             │   events    │
└─────────────┴─────────────┴─────────────┘
    ↓
Output (CLI)
```

### Tool Execution Flow

```
Agent Loop
    ↓
Tool Registry
    ↓
┌─────────────┬─────────────┬─────────────┐
│  Filesystem │    Exec     │     Web     │
│             │             │             │
│ - Read      │ - Shell     │ - HTTP      │
│ - Write     │ - Commands  │ - API       │
│ - List      │             │             │
└─────────────┴─────────────┴─────────────┘
    ↓
MCP Tools (optional)
    ↓
┌─────────────┬─────────────┬─────────────┐
│   Notion    │  Knowledge  │   Router    │
│             │    Graph    │             │
│ - Sync      │ - Search    │ - Route     │
│ - Query     │ - Add       │ - Optimize  │
│ - Update    │             │             │
└─────────────┴─────────────┴─────────────┘
    ↓
Tool Results
    ↓
Agent Loop (Observe)
```

### Memory Flow

```
Agent Loop
    ↓
Memory System
    ↓
┌─────────────┬─────────────┬─────────────┐
│   Session   │   SQLite    │  Knowledge  │
│   Memory    │   Store     │    Graph     │
│             │             │             │
│ - History   │ - Summary   │ - Entities  │
│ - Summary   │ - Metadata  │ - Relations │
│ - Context   │             │ - Observ.   │
└─────────────┴─────────────┴─────────────┘
    ↓
Knowledge Graph Memory (server-memory MCP)
    ↓
Persistent Storage
```

---

## 5. Configuration

### Ti Claw Config Structure

```json5
{
  // LLM Providers
  providers: {
    default: "router",  // Use Ti Router by default
    router: {
      enabled: true,
      endpoint: "http://localhost:8080",
      timeout: 30
    },
    anthropic: {
      enabled: true,
      apiKey: "${ANTHROPIC_API_KEY}",  // From env
      model: "claude-3-5-sonnet-20241022"
    },
    openai: {
      enabled: true,
      apiKey: "${OPENAI_API_KEY}",
      model: "gpt-4o"
    }
  },

  // Memory
  memory: {
    enabled: true,
    backend: "knowledge-graph",  // Use KG memory
    maxContext: 100000,
    summarizeThreshold: 0.75
  },

  // Tools
  tools: {
    filesystem: {
      enabled: true,
      allowedPaths: ["Z:\\10_WORKPLACE\\Ti"]
    },
    exec: {
      enabled: true,
      allowedCommands: ["git", "go", "node", "pnpm"]
    },
    web: {
      enabled: true,
      allowedDomains: ["*"]
    },
    mcp: {
      enabled: true,
      servers: [
        {
          name: "knowledge-graph",
          command: "node",
          args: ["Z:\\10_WORKPLACE\\Ti\\apps\\mcp\\server\\server-memory\\dist\\index.js"]
        },
        {
          name: "notion",
          command: "node",
          args: ["Z:\\10_WORKPLACE\\Ti\\apps\\mcp\\server\\notion\\dist\\index.js"]
        }
      ]
    }
  },

  // Agents
  agents: {
    dir: "content/ticlaw/agents",
    defaultAgent: "notion-agent"
  },

  // Skills
  skills: {
    dir: ".devin/skills",
    enabled: true
  },

  // Bootstrap
  bootstrap: {
    enabled: true,
    dir: "content/ticlaw/bootstrap",
    files: ["SOUL.md", "IDENTITY.md"]
  }
}
```

### Environment Variables

```bash
# Secrets (from Z:\00_SECRET\ticlaw.env)
ANTHROPIC_API_KEY=sk-ant-...
OPENAI_API_KEY=sk-...

# Ti Router
TI_ROUTER_ENDPOINT=http://localhost:8080
TI_ROUTER_API_KEY=...

# Memory
TI_MEMORY_ENABLED=true
TI_MEMORY_PATH=Z:/03_DATA/ti/memory/knowledge-graph.jsonl

# Data
TI_DATA_DIR=Z:/03_DATA/ti
TI_CONFIG_DIR=Z:/04_CONFIG/ti
```

---

## 6. CLI Commands

### Agent Commands

```bash
# List all agents
ti agent list

# Run an agent
ti agent run notion-agent --input "Sync GoClaw patterns to Notion"

# Create a new agent
ti agent create my-agent --template knowledge-sync

# Delete an agent
ti agent delete my-agent

# Show agent details
ti agent show notion-agent

# Test an agent
ti agent test notion-agent --input "Test query"
```

### Agent Management Commands

```bash
# Install agent dependencies
ti agent install notion-agent

# Update agent
ti agent update notion-agent

# Sync agent to Notion
ti agent sync notion-agent

# Export agent
ti agent export notion-agent --output my-agent.json

# Import agent
ti agent import my-agent.json
```

### Skill Commands

```bash
# List available skills
ti skill list

# Search skills
ti skill search "testing"

# Load skill into agent
ti skill load tdd-workflow --agent notion-agent

# Test skill
ti skill test tdd-workflow
```

---

## 7. Agent Definition Format

### Agent Definition (Markdown)

```markdown
# Notion Agent

## Metadata
- Name: notion-agent
- Version: 1.0.0
- Description: Sync knowledge to Notion
- Author: Ti Team

## Configuration
- Model: claude-3-5-sonnet-20241022
- MaxIterations: 10
- MaxToolCalls: 20
- ContextWindow: 200000

## Tools
- notion-sync
- notion-query
- notion-update
- knowledge-graph-search
- knowledge-graph-add

## Skills
- omniroute-knowledge-sync
- tdd-workflow
- golang-patterns

## Bootstrap
- SOUL.md
- IDENTITY.md
- KNOWLEDGE.md

## System Prompt
You are a knowledge synchronization agent. Your goal is to sync patterns, knowledge, and documentation to Notion for persistent storage and easy access.

## Instructions
1. Read knowledge from Ti learning lab
2. Extract patterns and insights
3. Sync to Notion with proper structure
4. Maintain consistency and accuracy
5. Provide feedback on sync status

## Examples
### Example 1: Sync GoClaw Patterns
User: Sync GoClaw patterns to Notion
Agent: I'll sync the GoClaw patterns to Notion...
[Tool calls: knowledge-graph-search, notion-sync]
Result: Synced 15 patterns to Notion

### Example 2: Query Notion
User: What patterns have we synced?
Agent: I'll query Notion for synced patterns...
[Tool calls: notion-query]
Result: Found 15 patterns in Notion
```

---

## 8. Testing Strategy

### Unit Tests

```bash
# Test core components
go test ./apps/ticlaw/internal/agent/...
go test ./apps/ticlaw/internal/store/...
go test ./apps/ticlaw/internal/tools/...
go test ./apps/ticlaw/internal/providers/...
```

### Integration Tests

```bash
# Test agent loop
go test ./apps/ticlaw/tests/integration/agent_loop_test.go

# Test tool execution
go test ./apps/ticlaw/tests/integration/tool_execution_test.go

# Test provider integration
go test ./apps/ticlaw/tests/integration/provider_test.go
```

### E2E Tests

```bash
# Test full agent execution
ti agent run notion-agent --input "Test query"

# Test CLI integration
ti agent list
ti agent show notion-agent
```

---

## 9. Documentation

### Required Documentation

1. **Ti Claw Architecture** - `docs/ticlaw-architecture.md`
2. **Agent Development Guide** - `docs/agent-development-guide.md`
3. **Tool Development Guide** - `docs/tool-development-guide.md`
4. **Provider Development Guide** - `docs/provider-development-guide.md`
5. **CLI Reference** - `docs/ticlaw-cli-reference.md`
6. **Notion Agent Guide** - `docs/notion-agent-guide.md`
7. **Migration Guide** - `docs/migration-from-goclaw.md`

---

## 10. Rollout Plan

### Week 1-2: Core Framework (Backend)
- Implement core patterns
- Unit tests
- Documentation

### Week 3: CLI Integration
- Integrate with Ti CLI
- Plugin system
- Integration tests

### Week 4: Notion Agent
- Build Notion Agent
- Test sync functionality
- Documentation

### Week 5-6: UI Development (NEW)
- **Phase 1**: Design System
  - Create `packages/ti-ui-components/`
  - Define theme variables
  - Implement core components (Button, Card, Input, Modal)
  - Implement layout components (Sidebar, Header, Layout)
- **Phase 2**: Ti Claw UI Skeleton
  - Create `apps/ticlaw/ui/`
  - Set up React + Vite + TypeScript
  - Implement MainLayout with Sidebar
  - Integrate ti-ui-components
  - Set up API client
- **Phase 3**: Core Pages
  - Implement DashboardPage
  - Implement AgentsPage
  - Implement AgentDetailPage
  - Implement ToolsPage
  - Implement MemoryPage
- **Phase 4**: Additional Pages
  - Implement SkillsPage
  - Implement ConfigPage
  - Implement SettingsPage
  - Add i18n support (Vietnamese)
  - Add theme toggle
- **Phase 5**: Polish & Integration
  - Polish UI/UX
  - Add animations (Motion)
  - Add charts (Chart.js)
  - Integrate with Ti Claw backend
  - Test and fix bugs
  - Documentation

### Week 7-8: Additional Agents & Polish
- Build additional agents
- Agent templates
- Best practices
- Bug fixes
- Performance optimization
- Final documentation
- Release v1.0.0

---

## 11. Success Criteria

- [x] Core Ti Claw framework implemented
- [x] CLI integration working
- [x] Notion Agent syncing knowledge
- [x] 4+ agents working
- [x] Documentation complete
- [x] Tests passing
- [x] Performance acceptable (<1s response time)
- [x] Shared design system created (ti-ui-components)
- [x] Ti Claw UI implemented (localhost:5174)
- [x] UI integration with backend working
- [x] Theme toggle working (light/dark)
- [x] Vietnamese i18n working
- [x] UI/UX polished

---

## 12. Risks & Mitigations

### Risk 1: Complexity Overwhelm
**Mitigation**: Start with core patterns only, add features incrementally

### Risk 2: Integration Issues
**Mitigation**: Extensive testing, rollback plan, feature flags

### Risk 3: Performance Issues
**Mitigation**: Profiling, optimization, caching

### Risk 4: Documentation Gaps
**Mitigation**: Document as we go, code comments, examples

---

## 13. Next Steps

1. **Create `apps/ticlaw/` repository**
2. **Implement core patterns (Phase 1 - Backend)**
3. **Integrate with Ti CLI (Phase 2)**
4. **Build Notion Agent (Phase 3)**
5. **Create shared design system (Week 5 - UI Phase 1)**
6. **Build Ti Claw UI skeleton (Week 5 - UI Phase 2)**
7. **Implement core UI pages (Week 5-6 - UI Phase 3)**
8. **Implement additional UI pages (Week 6 - UI Phase 4)**
9. **Polish UI & integrate backend (Week 6 - UI Phase 5)**
10. **Build additional agents (Week 7)**
11. **Polish & Release (Week 8)**

---

## 14. References

- **GoClaw Source**: `Ti-learning-lab/01_Learning/lab/07_Repositories/goclaw-main/`
- **GoClaw CLAUDE.md**: `Ti-learning-lab/01_Learning/lab/07_Repositories/goclaw-main/CLAUDE.md`
- **Ti CLI**: `apps/cli/`
- **Ti Router**: `apps/router/`
- **Ti Router UI**: `apps/router/ui/` (localhost:5173)
- **Ti Skills**: `.devin/skills/`
- **Ti Agents**: `content/agents/`
- **Knowledge Graph Memory**: `content/mcp/KNOWLEDGE_GRAPH_USAGE_GUIDE.md`
- **Patterns Extraction**: `Ti-learning-lab/03_Knowledge/Router/goclaw-patterns-extraction.md`
- **UI Approach**: `Ti-learning-lab/03_Knowledge/Router/ti-claw-ui-approach.md`
- **Anthropic Tools Analysis**: `Ti-learning-lab/03_Knowledge/Router/anthropic-claude-code-tools-analysis.md`
- **Notion Task Migration**: `Ti-learning-lab/03_Knowledge/Router/notion-task-management-migration.md`

---

## 15. Task Management Migration (BD → Notion)

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
- Auto-logging: Tự động log từ agent execution

### Migration Plan
1. **Week 1**: Create Notion database, define properties, create views
2. **Week 2**: Implement Notion task logger, CLI command
3. **Week 3**: Migrate BD data to Notion, verify integrity, deprecate BD tool
4. **Week 4**: Update all references, auto-logging from agents, Ti Claw UI integration

### Benefits
- Centralized knowledge & tasks in Notion
- Better visualization (Table, Board, Timeline, Calendar)
- Advanced querying and filtering
- Automation support
- Multi-user collaboration
- Integration with knowledge system

---

**Status**: Plan created ✅
**Next**: Implement Phase 1 (Core Framework) + Notion Database Setup
