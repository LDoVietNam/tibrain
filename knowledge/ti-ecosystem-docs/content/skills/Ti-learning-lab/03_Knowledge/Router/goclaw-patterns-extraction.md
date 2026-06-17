# GoClaw Patterns Extraction for Ti Claw

> **Purpose**: Extract key patterns from GoClaw to rebrand as Ti Claw
> **Date**: 2026-05-04
> **Source**: `Ti-learning-lab/01_Learning/lab/07_Repositories/goclaw-main/`

---

## 1. Core Architecture Patterns

### 1.1 Agent Loop Pattern (Think-Act-Observe)

**Location**: `internal/agent/loop.go`, `internal/agent/loop_types.go`

**Pattern**:
```go
type Loop struct {
    provider      providers.Provider
    model         string
    contextWindow int
    maxIterations int
    maxToolCalls  int
    sessions      store.SessionStore
    tools         *tools.Registry
    // ... other fields
}

// Think → Act → Observe cycle
func (l *Loop) runLoop(ctx context.Context, req RunRequest) (*RunResult, error) {
    // 1. Build messages from session history
    // 2. Process media
    // 3. Think: Call LLM
    // 4. Act: Execute tools
    // 5. Observe: Process tool results
    // 6. Repeat until completion
}
```

**Key Components**:
- **Context Injection**: `injectContext()` - agent/tenant/user/workspace scoping
- **Message Building**: `buildMessages()` - resolves context files, history, summary
- **Media Handling**: `persistMedia()`, `reloadMediaForMessages()`
- **Tool Execution**: Tool registry with policy engine
- **Event System**: `AgentEvent` with delegation + routing context

**Events**:
- `run.started`, `run.completed`
- `chunk` (streaming tokens)
- `tool.call`, `tool.result`
- Auto-summarization at >75% context

---

### 1.2 Store Layer Pattern (Repository Pattern)

**Location**: `internal/store/`, `internal/store/pg/`

**Pattern**:
```go
// Interface-based design
type SessionStore interface {
    GetHistory(ctx context.Context, sessionKey string) []Message
    GetSummary(ctx context.Context, sessionKey string) string
    SetContextWindow(ctx context.Context, sessionKey string, window int)
    // ... other methods
}

// PostgreSQL implementation
type pgSessionStore struct {
    db *sql.DB
}

func (s *pgSessionStore) GetHistory(...) []Message {
    // Raw SQL with $1, $2 parameters
    // Uses execMapUpdate() helper
}
```

**Key Features**:
- Interface-based abstraction
- Raw SQL with `database/sql` + `pgx/v5/stdlib`
- Parameterized queries (SQL injection prevention)
- Context propagation: `store.WithAgentType(ctx)`, `store.WithUserID(ctx)`, `store.WithAgentID(ctx)`, `store.WithLocale(ctx)`
- Multi-tenant isolation

---

### 1.3 Tool Registry Pattern

**Location**: `internal/tools/`

**Pattern**:
```go
type Registry struct {
    tools map[string]Tool
    policy *PolicyEngine
}

type Tool interface {
    Name() string
    Description() string
    Execute(ctx context.Context, args map[string]any) (map[string]any, error)
}

type PolicyEngine struct {
    // Filters tools sent to LLM based on policy
}
```

**Built-in Tools**:
- `filesystem` - file operations
- `exec` - shell execution
- `web` - HTTP requests
- `memory` - memory operations
- `subagent` - agent delegation
- `read_image` - vision provider integration

**Tool Gating**:
- `TeamActionPolicy` - blocks certain tools in lite edition
- Per-agent tool policy from DB
- Per-tenant disabled tools

---

### 1.4 Provider Pattern (Adapter Pattern)

**Location**: `internal/providers/`

**Pattern**:
```go
type Provider interface {
    Complete(ctx context.Context, req CompleteRequest) (*CompleteResponse, error)
    CompleteWithTools(ctx context.Context, req CompleteRequest) (*CompleteResponse, error)
    Stream(ctx context.Context, req CompleteRequest) (<-chan Chunk, error)
}

// Implementations:
// - Anthropic (native HTTP+SSE)
// - OpenAI-compat (HTTP+SSE)
// - DashScope (Alibaba Qwen)
// - Claude CLI (stdio+MCP bridge)
// - ACP (Anthropic Console Proxy)
// - Codex (OpenAI)
```

**Key Features**:
- All use `RetryDo()` for retries
- Loads from `llm_providers` table with encrypted API keys
- AES-256-GCM encryption for API keys
- SSE streaming support

---

### 1.5 Event Bus Pattern

**Location**: `internal/bus/`

**Pattern**:
```go
type EventPublisher interface {
    Publish(event Event) error
}

type Event struct {
    Type string
    Data any
    // ... metadata
}
```

**Usage**:
- Agent lifecycle events
- Media file events
- Tool execution events
- WebSocket protocol (v3): Frame types `req`/`res`/`event`

---

## 2. Configuration Patterns

### 2.1 Config Loading (JSON5 + Env Overlay)

**Location**: `internal/config/`

**Pattern**:
```go
// Load from JSON5 file
config := LoadConfig("config.json5")

// Overlay with environment variables
config = config.WithEnvVars()
```

**Secrets Management**:
- Secrets in `.env.local` or env vars
- NEVER in config.json
- AES-256-GCM encryption for API keys in DB

---

### 2.2 Agent Types

**Location**: `internal/agent/`

**Pattern**:
```go
// "open" - per-user context (7 files)
// "predefined" - shared context + USER.md per-user
type AgentType string
```

**Context Files**:
- `agent_context_files` (agent-level)
- `user_context_files` (per-user)
- Routed via `ContextFileInterceptor`

---

## 3. Scheduling Patterns

### 3.1 Lane-based Concurrency

**Location**: `internal/scheduler/`

**Pattern**:
```go
// Lanes: main, subagent, cron
// Each lane has its own concurrency limit
type Scheduler struct {
    lanes map[string]*Lane
}

type Lane struct {
    maxConcurrency int
    activeRuns     atomic.Int32
}
```

**Adaptive Throttle**:
- Uses real context window from session
- Not hardcoded 200K
- Lane-based isolation

---

### 3.2 Cron Scheduling

**Location**: `internal/cron/`

**Pattern**:
```go
// Support: at, every, cron expressions
type CronJob struct {
    Schedule string // "at 9am", "every 1h", "0 * * * *"
    Handler  func()
}
```

---

## 4. Memory Patterns

### 4.1 Memory System (pgvector)

**Location**: `internal/memory/`

**Pattern**:
```go
type MemoryStore interface {
    Add(ctx context.Context, sessionKey string, content string) error
    Search(ctx context.Context, sessionKey string, query string) ([]Memory, error)
}
```

**Features**:
- pgvector for similarity search
- Per-session memory
- Configurable memory size limits

---

### 4.2 Session Management

**Location**: `internal/sessions/`

**Pattern**:
```go
type SessionStore interface {
    GetHistory(ctx context.Context, sessionKey string) []Message
    GetSummary(ctx context.Context, sessionKey string) string
    SetSummary(ctx context.Context, sessionKey string, summary string) error
}
```

**Auto-summarization**:
- At >75% context usage
- Per-session summarization lock (sync.Map)
- Summary stored in session

---

## 5. Security Patterns

### 5.1 Input Guard

**Location**: `internal agent/input_guard.go`

**Pattern**:
```go
type InputGuard struct {
    injectionAction string // "log", "warn", "block", "off"
}

func (g *InputGuard) Scan(input string) GuardResult {
    // Detect injection attempts
    // Return action based on config
}
```

**Security Features**:
- Rate limiting
- Input guard (detection-only)
- CORS
- Shell deny patterns
- SSRF protection
- Path traversal prevention
- All security logs: `slog.Warn("security.*")`

---

### 5.2 RBAC (Role-Based Access Control)

**Location**: `internal/permissions/`

**Pattern**:
```go
type Role string
const (
    RoleAdmin    Role = "admin"
    RoleOperator Role = "operator"
    RoleViewer   Role = "viewer"
)

func HasPermission(ctx context.Context, role Role, permission string) bool {
    // Check permissions
}
```

---

## 6. Integration Patterns

### 6.1 MCP (Model Context Protocol)

**Location**: `internal/mcp/`

**Pattern**:
```go
// MCP bridge/server
type MCPServer struct {
    tools []MCPTool
}

type MCPTool struct {
    Name        string
    Description string
    Handler     func(ctx context.Context, args map[string]any) (map[string]any, error)
}
```

**Features**:
- Tool bridging to MCP servers
- Resource access
- Prompt templates

---

### 6.2 Channel Integration

**Location**: `internal/channels/`

**Pattern**:
```go
// Supported channels:
// - Telegram (telego)
// - Feishu/Lark
// - Zalo
// - Discord
// - WhatsApp

type Channel interface {
    SendMessage(ctx context.Context, chatID string, message string) error
}
```

**Telegram Formatting**:
- LLM output → `SanitizeAssistantContent()`
- → `markdownToTelegramHTML()`
- → `chunkHTML()`
- → `sendHTML()`
- Tables rendered as ASCII in `<pre>` tags

---

## 7. Bootstrap Patterns

### 7.1 Context File Seeding

**Location**: `internal/bootstrap/`

**Pattern**:
```go
type ContextFile struct {
    Name    string
    Content string
}

func SeedUserFiles(ctx context.Context, agentID uuid.UUID, userID, agentType string, isNew bool) error {
    // Seed BOOTSTRAP.md, USER.md, etc.
    // Only if isNew or user has zero files
}
```

**Auto-cleanup**:
- BOOTSTRAP.md removed after 3 user messages
- `bootstrapAutoCleanupTurns = 3`

---

### 7.2 Skill Loading

**Location**: `internal/skills/`

**Pattern**:
```go
type Loader struct {
    skills map[string]Skill
}

type Skill struct {
    Name        string
    Description string
    Content     string
}

// BM25 search for skills
func (l *Loader) Search(query string) []Skill {
    // BM25 algorithm
}
```

---

## 8. Internationalization (i18n)

**Location**: `internal/i18n/`

**Pattern**:
```go
// Backend: message catalog
func T(locale, key string, args ...any) string {
    // Return translated string
}

// Supported: en (default), vi, zh
// Locale propagated via store.WithLocale(ctx)
```

**Web UI**:
- `i18next` with namespace-split locale files
- Location: `ui/web/src/i18n/locales/{lang}/`

**Rules**:
- New user-facing strings: add key to `internal/i18n/keys.go`, add translations to all 3 catalog files
- New UI strings: add key to all locale JSON files
- Bootstrap templates (SOUL.md, etc.) stay English-only (LLM consumption)

---

## 9. Desktop Edition (Lite) Patterns

**Location**: `ui/desktop/`

**Pattern**:
```go
// Build tag: //go:build sqliteonly
// Desktop binary includes only SQLite, no PostgreSQL

type Edition string
const (
    EditionStandard Edition = "standard"
    EditionLite      Edition = "lite"
)

// Lite limits:
// - 5 agents, 1 team, 5 members, 50 sessions
// - No channels, heartbeat, file storage UI, skill self-manage, KG, RBAC, multi-tenant
```

**Secrets**:
- OS keyring (`go-keyring`) with file fallback
- Data dir: `~/.goclaw/data/`
- Workspace: `~/.goclaw/workspace/`

---

## 10. Mobile UI/UX Patterns

**Location**: `ui/web/`

**Pattern**:
```css
/* Viewport height: use h-dvh, never h-screen */
.container {
    height: 100dvh;
}

/* Input font-size: 16px on mobile to prevent iOS auto-zoom */
input {
    font-size: 16px; /* text-base md:text-sm */
}

/* Safe areas for notched devices */
.safe-top {
    padding-top: env(safe-area-inset-top);
}
```

**Rules**:
- Touch targets: ≥44px on touch devices
- Tables: wrap in overflow-x-auto, min-w-[600px]
- Grid layouts: mobile-first responsive
- Dialogs: full-screen on mobile, centered on desktop
- Virtual keyboard: use `useVirtualKeyboard()` hook
- ErrorBoundary key: strip dynamic segments for stable key

---

## 11. Migration Patterns

**Location**: `migrations/`

**Pattern**:
```sql
-- Migration files in migrations/
-- Bump RequiredSchemaVersion in internal/upgrade/version.go
```

**Rules**:
- When adding new migration, bump `RequiredSchemaVersion`
- All user inputs use parameterized queries (`$1, $2`)
- Queries optimized (no N+1, no unnecessary full table scans)
- WHERE clauses, JOINs, ORDER BY use existing indices

---

## 12. Testing Patterns

**Location**: `tests/integration/`

**Pattern**:
```bash
go test -v ./tests/integration/
go test -race ./tests/integration/  # With race detector
```

**Post-Implementation Checklist**:
```bash
go fix ./...                        # Apply Go version upgrades
go build ./...                      # Compile check (PG build)
go build -tags sqliteonly ./...     # Compile check (Desktop/SQLite build)
go vet ./...                        # Static analysis
go test -race ./tests/integration/  # Integration tests with race detector
```

---

## 13. Go Conventions

**From CLAUDE.md**:
- Use `errors.Is(err, sentinel)` instead of `err == sentinel`
- Use `switch/case` instead of `if/else if` chains on the same variable
- Use `append(dst, src...)` instead of loop-based append
- Always handle errors; don't ignore return values
- Prefer explicit configuration over runtime heuristics
- Prefer the simplest solution that addresses the root cause directly

---

## 14. Ti Claw Adaptation Plan

### Components to Keep

1. **Agent Loop** - Think-Act-Observe pattern
2. **Store Layer** - Interface-based with SQLite (for Ti CLI)
3. **Tool Registry** - Tool management
4. **Provider Pattern** - LLM provider abstraction
5. **Event Bus** - Event system
6. **Session Management** - History + summary
7. **Memory System** - SQLite-based (simplified)
8. **Config Loading** - JSON5 + env overlay
9. **Bootstrap Pattern** - Context file seeding
10. **Skill Loading** - BM25 search

### Components to Remove

1. **Multi-tenant** - Ti is single-user CLI
2. **PostgreSQL** - Use SQLite only
3. **Web UI** - Ti is CLI-first
4. **Desktop UI** - Not needed for CLI
5. **Channels** - Not needed for CLI
6. **RBAC** - Single-user, no permissions
7. **WebSocket** - CLI uses stdin/stdout
8. **Cron Scheduling** - Not needed for CLI
9. **Lane-based Scheduler** - Simplified to single-threaded
10. **OAuth** - Not needed for CLI

### Components to Tweak

1. **Agent Types** - Keep only "predefined" (Ti has fixed agent definitions)
2. **Context Files** - Adapt to Ti's content/ structure
3. **Skills** - Integrate with Ti's .devin/skills/
4. **Memory** - Use Ti's Knowledge Graph Memory (server-memory)
5. **Providers** - Keep Anthropic, OpenAI, add Ti Router as provider
6. **Tools** - Adapt to Ti's MCP servers
7. **i18n** - Keep Vietnamese (Ti is VN-focused)
8. **Security** - Simplify (no rate limiting, no CORS, no SSRF)

### New Components for Ti

1. **Ti CLI Integration** - Plugin system for apps/cli/
2. **Ti Router Provider** - Use Ti Router as LLM provider
3. **Knowledge Graph Memory** - Integrate with server-memory MCP
4. **Notion Agent** - First agent for knowledge sync
5. **Ti Skill System** - Integrate with .devin/skills/
6. **OmniRoute Patterns** - Integrate routing patterns
7. **Ti Config** - Use Ti's config system

---

## 15. Next Steps

1. **Create Ti Claw repository** - Fork GoClaw or start from scratch
2. **Implement core patterns** - Agent loop, store, tools, providers
3. **Integrate with Ti CLI** - Plugin system
4. **Build Notion Agent** - First use case
5. **Extract patterns to Notion** - Document all patterns
6. **Test and iterate** - Verify functionality

---

## 16. References

- **GoClaw Source**: `Ti-learning-lab/01_Learning/lab/07_Repositories/goclaw-main/`
- **GoClaw CLAUDE.md**: `Ti-learning-lab/01_Learning/lab/07_Repositories/goclaw-main/CLAUDE.md`
- **Ti CLI**: `apps/cli/`
- **Ti Router**: `apps/router/`
- **Ti Skills**: `.devin/skills/`
- **Ti Agents**: `content/agents/`
- **Knowledge Graph Memory**: `content/mcp/KNOWLEDGE_GRAPH_USAGE_GUIDE.md`

---

**Status**: Patterns extracted ✅
**Next**: Sync to Notion via Notion Agent
