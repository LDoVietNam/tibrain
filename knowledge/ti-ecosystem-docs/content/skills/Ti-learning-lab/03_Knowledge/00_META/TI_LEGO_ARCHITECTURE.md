# Ti LEGO Architecture

> **Last Updated**: 2026-05-07
> **Purpose**: Document Ti's LEGO-based architecture pattern for agents and developers

---

## 🧱 LEGO Architecture Model

Ti được xây dựng theo hướng **LEGO block composition** với 4 tầng abstraction:

| Term | LEGO Equivalent | Description | Examples |
|------|----------------|-------------|----------|
| **Kernel** | Core brick | Cốt lõi, nền tảng cho toàn bộ hệ sinh thái AI | Ti Core, Router |
| **Module** | Parent block | Folder/package lớn hoặc nhóm chức năng | `apps/core/cli/`, `apps/daemon/` |
| **Block** | Child block | File, function, hoặc plugin con | `internal/ui/dashboard.go`, `internal/providers/` |
| **Plugin** | Attachable block | Unit độc lập, Go plugin hoặc workflow node | MCP servers, Ticrew members, Skills |

---

## 🎯 Design Principles

### 1. **Infrastructure-First Design**
- **Tất cả LEGO blocks đều là infrastructure concern**
- Quản lý: trạng thái, dependency, I/O, runtime, scalability
- Health, status, diagnostics, runtime metrics cho mọi level
- Không có ngoại lệ - kernel/module/block/plugin đều cùng nature

### 2. **Scope-Based Organization**
- **Kernel**: Core infrastructure, cốt lõi hệ thống
- **Module**: Tập hợp blocks, quản lý dependency chain
- **Block**: Unit nhỏ, có status/error state
- **Plugin**: Unit độc lập, attachable, runtime-aware

### 3. **Composability**
- Mỗi block có thể lắp ghép với nhau
- Input/output rõ ràng
- Interface well-defined

### 4. **Independence**
- Plugin có thể hoạt động độc lập
- Block có thể test riêng biệt
- Module có thể deploy riêng

### 5. **Interoperability**
- Standard interfaces giữa blocks
- Common protocols (MCP, sidecar)
- Shared data structures

### 6. **Extensibility**
- Dễ thêm mới plugins
- Dễ mở rộng modules
- Dây chuyền mở rộng kernel

---

## 📊 Hierarchy Examples

### Example 1: Ti CLI Structure

```
Kernel: Ti Core (router, providers, configs)
  ↓
Module: apps/core/cli/
  ↓
Blocks:
  - cmd/ (CLI commands)
  - internal/ui/ (TUI components)
  - internal/providers/ (LLM providers)
  - internal/convert/ (code conversion)
  ↓
Plugins:
  - MCP servers (memory-gateway, github-mcp)
  - Ticrew members (notion-agent, router-agent)
  - Skills (docx, pptx, xlsx)
```

### Example 2: Ticrew Framework

```
Kernel: Ti Core
  ↓
Module: apps/core/cli/internal/plugins/ticrew/
  ↓
Blocks:
  - multi/orchestrator.go (agent orchestration)
  - cmd/agent.go (agent commands)
  - bootstrap/identity.go (agent identity)
  - scheduler/lanes.go (lane-based scheduling)
  ↓
Plugins (Ticrew Members):
  - notion-agent (Notion integration)
  - router-agent (Router integration)
  - custom-agent (User-defined agents)
```

### Example 3: Provider System

```
Kernel: Ti Core
  ↓
Module: apps/core/cli/internal/providers/
  ↓
Blocks:
  - router.go (Router provider)
  - provider.go (Base provider interface)
  - openai.go (OpenAI provider)
  ↓
Plugins:
  - Cookie providers (custom auth)
  - BYOK providers (user's own keys)
  - MCP providers (external providers)
```

---

## 🔌 Plugin Interface Contract

Mỗi plugin phải implement:

### 1. **Metadata**
```go
type PluginMetadata struct {
    Name        string
    Version     string
    Type        string // "mcp", "ticrew-member", "skill"
    Description string
    Capabilities []string
}
```

### 2. **Lifecycle**
- `Init()` - Khởi tạo plugin
- `Start()` - Bắt đầu plugin
- `Stop()` - Dừng plugin
- `HealthCheck()` - Check health status

### 3. **Interface**
```go
type Plugin interface {
    Metadata() PluginMetadata
    Init(config Config) error
    Start(ctx context.Context) error
    Stop(ctx context.Context) error
    HealthCheck() HealthStatus
}
```

---

## 🎨 Block Design Guidelines

### 1. **Single Responsibility**
- Một block = một responsibility
- Không mix concerns
- Dễ test, dễ maintain

### 2. **Clear Boundaries**
- Input/output rõ ràng
- Interface well-defined
- Không leak implementation details

### 3. **Dependency Direction**
- High-level modules không phụ thuộc low-level
- Abstractions không phụ thuộc details
- Dependency inversion principle

### 4. **Error Handling**
- Plugin có status, error state
- Block handle errors gracefully
- Fallback mechanisms

---

## 🚀 Usage Patterns

### Pattern 1: Plugin Registration

```go
// Register MCP plugin
registry.Register(mcp.NewMemoryGateway())

// Register Ticrew member
ticrew.RegisterMember(notionAgent)

// Register skill
skills.Register(docx.NewSkill())
```

### Pattern 2: Block Composition

```go
// Compose blocks into module
cliModule := &CLIModule{
    Commands: []Command{
        dashboard.NewCommand(),
        chat.NewCommand(),
        provider_wizard.NewCommand(),
    },
    UI: &UI{
        Dashboard: dashboard.NewTUI(),
        Chat: chat.NewTUI(),
    },
}
```

### Pattern 3: Plugin Discovery

```go
// Auto-discover plugins
plugins := discoverer.Discover("plugins/")
for _, plugin := range plugins {
    plugin.Init(config)
    plugin.Start(ctx)
}
```

---

## 📋 Decision Framework

**Important**: Tất cả LEGO blocks đều là infrastructure concern. Câu hỏi không phải "infrastructure vs feature", mà là **scope và reusability**.

Khi quyết định implement feature mới:

| Question | Answer → Action |
|----------|----------------|
| Là core infrastructure, được dùng bởi toàn hệ thống? | Yes → Kernel |
| Là large functional group, tập hợp nhiều blocks? | Yes → Module |
| Là unit nhỏ, file/function, có scope cụ thể? | Yes → Block |
| Là independent unit, attachable, runtime-aware? | Yes → Plugin |
| **Cần được reuse bởi nhiều modules?** | **Yes → Higher level (Kernel/Module)** |
| **Chỉ cần trong scope của một module cụ thể?** | **Yes → Lower level (Block/Plugin trong module)** |

### Examples

- "Router provider system" → Module (providers/) - large functional group
- "Dashboard TUI" → Block (internal/ui/dashboard.go) - unit nhỏ trong UI module
- "GitHub MCP server" → Plugin - independent unit, attachable
- "Agent detection logic" → **Scope-based decision**:
  - Nếu chỉ Ticrew cần → Block trong Ticrew module
  - Nếu UI, Config, Convert đều cần → Block trong core infrastructure
- "Ticrew framework" → Module (plugins/ticrew/) - large functional group

### Reusability Heuristics

| Scenario | Recommended Location | Rationale |
|----------|---------------------|-----------|
| **Single module use** | Block/Plugin trong module đó | Minimize coupling, simple dependency |
| **Multiple modules use** | Core infrastructure (Kernel/Module) | Avoid duplication, single source of truth |
| **External third-party** | Plugin (attachable) | Independent lifecycle, optional dependency |
| **Core platform feature** | Kernel | Always available, no optional dependencies |

---

## 🔗 Related Docs

- <ref_file file="Z:\10_WORKPLACE\Ti\Ti-learning-lab\AGENTS.md" /> - Agent guidelines
- <ref_file file="Z:\10_WORKPLACE\Ti\Ti-learning-lab\03_Knowledge\cli\CONVERT_TRANS_ARCHITECTURE.md" /> - Convert architecture
- <ref_file file="Z:\10_WORKPLACE\Ti\Ti-learning-lab\03_Knowledge\cli\UI_ARCHITECTURE.md" /> - UI architecture

---

## 📝 Notes

- LEGO architecture cho phép rapid composition
- Plugin system cho phép third-party extensions
- Module boundaries cho phép independent development
- Block granularity cho phép reusability

---

*Last Updated: 2026-05-07*
*Version: 1.0*
