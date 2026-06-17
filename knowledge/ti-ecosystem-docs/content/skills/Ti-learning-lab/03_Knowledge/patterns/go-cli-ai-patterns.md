---
tags: ["tibrain", "cli", "go", "documentation", "provider"]
scopes: ["cli", "tibrain"]
last_updated: 2026-05-22
---
# Kien Thuc Nghien Cuu: Mau Kien Truc CLI Go cho AI Agents

**Ngay**: 2026-04-29
**Nguon**: GitHub Repos (skport/golang-cli-architecture, awf-project/cli, opencode-ai/opencode, ENTERPILOT/GOModel)
**Muc dich**: Xay dung Ti CLI voi kien truc chuan, ho tro nhieu provider AI

---

## 1. Kien Truc Layer Chuan cho Go CLI

### 1.1 Clean Architecture / Hexagonal

```
Controller (cmd/cobra) → Application (internal/app) → Domain (internal/core) → Infrastructure (internal/providers)
```

**Nguon hoc tap**: `skport/golang-cli-architecture`
- `cmd/` - Cobra commands
- `internal/app/` - Business logic, services
- `internal/core/` - Domain entities, interfaces
- `internal/provider/` - Concrete implementations

**Loi ich**:
- De test (mock domain interfaces)
- De thay doi provider (chi can swap implementation)
- Khong phu thuoc framework o tang domain

### 1.2 Modular Architecture (OpenCode Pattern)

```
cmd/                    - CLI interface
internal/app/           - Core services
internal/config/        - Config management  
internal/llm/           - LLM providers integration
internal/db/            - Database & migrations
internal/session/       - Session management
internal/tui/           - Terminal UI
internal/logging/       - Logging infrastructure
```

**Nguon hoc tap**: `opencode-ai/opencode`

---

## 2. Quan Ly Provider AI (Multi-Backend)

### 2.1 Auto-Detection Pattern (GOModel)

```go
// Tu dong phat hien provider tu env vars
providers := []string{}
if os.Getenv("OPENAI_API_KEY") != "" {
    providers = append(providers, "openai")
}
if os.Getenv("ANTHROPIC_API_KEY") != "" {
    providers = append(providers, "anthropic")
}
```

**Nguon hoc tap**: `ENTERPILOT/GOModel`

### 2.2 Registry Pattern

```go
type ProviderRegistry struct {
    providers map[string]Provider
}

func (r *ProviderRegistry) Register(p Provider) error
func (r *ProviderRegistry) Get(name string) (Provider, bool)
func (r *ProviderRegistry) List() []ProviderSnapshot
```

**Nguon hoc tap**: `awf-project/cli` (workflow states + providers)

### 2.3 Unified Interface

```go
type Provider interface {
    ID() string
    Name() string
    Models() []Model
    Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error)
    Stream(ctx context.Context, req ChatRequest) (StreamResponse, error)
    IsHealthy() bool
}
```

---

## 3. He Thong Config (7-Level Precedence)

### 3.1 Thu tu uu tien (OpenCode Pattern)

1. Env vars (cao nhat)
2. `./.ti.json` (project local)
3. `$XDG_CONFIG_HOME/ti/config.json` (global)
4. `$HOME/.ti/config.json` (fallback)
5. Defaults (thap nhat)

**Nguon hoc tap**: `opencode-ai/opencode` config system

### 3.2 Portable Mode

```go
// Kiem tra .ti/ trong thu muc hien tai truoc
func GetConfigDir() string {
    if _, err := os.Stat(".ti/config.json"); err == nil {
        return ".ti"
    }
    return os.Getenv("XDG_CONFIG_HOME") + "/ti"
}
```

---

## 4. Workflow Orchestration (AWF Pattern)

### 4.1 YAML Workflow

```yaml
name: code-review
version: "1.0.0"
inputs:
  - name: file
    type: string
    required: true
states:
  initial: read
  read:
    type: step
    command: cat "{{.inputs.file}}"
    on_success: analyze
  analyze:
    type: agent
    provider: claude
    prompt: "Review this code..."
    on_success: report
```

**Nguon hoc tap**: `awf-project/cli`

### 4.2 State Machine

- `initial` → `step/agent` → `on_success/on_failure` → `terminal`
- Ho tro retry, timeout, parallel execution

---

## 5. Plugin System

### 5.1 Plugin Registry

```bash
# Cac lenh plugin chuan
awf plugin list
awf plugin install owner/repo
awf plugin update [name]
awf plugin remove <name>
awf plugin search [query]
awf plugin enable/disable <name>
```

**Nguon hoc tap**: `awf-project/cli`

---

## 6. Session Management

### 6.1 Session Store

- SQLite cho persistence
- Resume conversation
- Export/import session
- Token tracking

**Nguon hoc tap**: `opencode-ai/opencode` (internal/session)

---

## 7. Security & Best Practices

### 7.1 Secret Management
- Khong hardcode API keys
- Dung env vars hoac keyring
- Khong log secrets

### 7.2 Sandboxing
- Dry-run mode: `ti run --dry-run`
- Interactive mode: `ti run --interactive`  
- Isolated execution cho workflow khong tin cay

**Nguon hoc tap**: `awf-project/cli` (security disclaimer)

---

## 8. Tich Hop cho Ti CLI

### 8.1 Nhung gi da co san trong Ti CLI
- Config 7-level precedence ✅
- Provider registry interface ✅
- SharedChatProvider (cookie-based) ✅
- Cobra commands (root, chat, plugin) ✅

### 8.2 Nhung gi can bo sung
- **Provider auto-detection** tu env vars
- **OpenRouter/Anthropic/OpenAI providers** (API key based)
- **Session store** (SQLite)
- **ti doctor** command (health check)
- **ti auth** command (credential management)
- **ti config** command (config CRUD)
- **ti run** command (workflow execution)
- **ti ask** command (quick chat with auto-route)
- **Streaming support** cho chat responses
- **Plugin system** (gRPC / native Go)

### 8.3 Kien truc de xuat cho Ti CLI v1

```
cmd/
  root.go       - Cobra root command
  ask.go        - Quick ask command
  run.go        - Workflow execution
  config.go     - Config management
  auth.go       - Credential management
  provider.go   - Provider commands (da co)
  skill.go      - Skill commands (da co)
  doctor.go     - Health check
  session.go    - Session management

internal/
  config/       - Config loading (da co)
  providers/    - Provider registry + implementations (da co phan registry)
    bootstrap.go    - Auto-detect + register providers
    sharedchat.go   - Cookie-based provider (da co)
    openrouter.go   - OpenRouter provider
    anthropic.go    - Anthropic provider
    openai.go       - OpenAI provider
  app/          - Core application services
    chat.go         - Chat orchestration
    workflow.go     - Workflow runner
  session/      - Session store (SQLite)
  skills/       - Skill reader (da co trong cmd/)
```

---

## 9. Cac Buoc Trien Khai (Tu nghien cuu)

1. **Phase 1**: Auto-detect providers tu env vars + register vao bootstrap
2. **Phase 2**: Them OpenRouter/Anthropic/OpenAI providers
3. **Phase 3**: Session store (SQLite) + `ti session` commands
4. **Phase 4**: `ti doctor` health check
5. **Phase 5**: `ti auth` credential management
6. **Phase 6**: `ti config` config CRUD
7. **Phase 7**: `ti ask` quick chat voi auto-route
8. **Phase 8**: `ti run` workflow execution (YAML states)
9. **Phase 9**: Streaming support
10. **Phase 10**: Plugin system

---

**Ket luan**: Can tich hop Clean Architecture + Auto-Detection + Session Management de bien Ti CLI thanh production-ready AI orchestration tool.
