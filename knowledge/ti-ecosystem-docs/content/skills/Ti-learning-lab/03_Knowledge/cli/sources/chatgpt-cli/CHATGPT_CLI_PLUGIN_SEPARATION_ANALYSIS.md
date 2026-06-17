# ChatGPT CLI Plugin Separation Analysis

> **Date**: 2026-05-05
> **Purpose**: Analyze ChatGPT CLI source code to propose plugin separation strategy for Ti CLI
> **Goal**: Balance granularity with dependency management to avoid dependency hell

---

## Source Code Structure Analysis

### Directory Overview

```
chatgpt-cli-main/
├── agent/              # Agent framework
│   ├── core/          # BaseAgent, Budget, Policy, Runner, Clock, TranscriptBuffer
│   ├── factory/       # Agent factory for creating different agent types
│   ├── planexec/      # Plan-Execute agent implementation
│   ├── react/         # ReAct agent implementation
│   ├── tools/         # Tool interfaces (LLM, Shell, Files)
│   ├── types/         # Core types (Step, Plan, Result, Effects)
│   └── utils/         # Utilities (unified diff parsing)
├── api/               # API integration
│   ├── client/        # LLM API client, MCP integration, media handling
│   └── http/          # HTTP utilities
├── cache/             # Caching layer
├── config/            # Configuration management
├── history/           # Conversation history
├── internal/          # Internal utilities
│   └── fsio/          # File I/O abstractions
├── cmd/chatgpt/       # CLI entry point
└── docs/              # Documentation
```

### Dependency Graph

Based on the source code analysis, the dependency flow is:

```
main.go (CLI entry point)
  ├─> config/ (configuration)
  ├─> history/ (conversation history)
  ├─> api/client/ (LLM API client)
  │   └─> api/http/ (HTTP utilities)
  ├─> agent/factory/ (agent factory)
  │   ├─> agent/core/ (BaseAgent, Budget, Policy, Runner, Clock)
  │   ├─> agent/react/ (ReAct agent)
  │   │   ├─> agent/core/
  │   │   ├─> agent/tools/ (LLM, Shell, Files)
  │   │   └─> agent/types/
  │   └─> agent/planexec/ (Plan-Execute agent)
  │       ├─> agent/core/
  │       ├─> agent/tools/
  │       └─> agent/types/
  ├─> cache/ (caching)
  └─> internal/fsio/ (file I/O)
```

### Coupling Analysis

#### High Coupling (Tightly Coupled)
- **agent/react/** depends on: agent/core/, agent/tools/, agent/types/
- **agent/planexec/** depends on: agent/core/, agent/tools/, agent/types/
- **agent/core/** depends on: agent/types/
- **agent/tools/** depends on: agent/types/

#### Medium Coupling
- **api/client/** depends on: api/http/, history/, internal/fsio/
- **agent/factory/** depends on: agent/core/, agent/react/, agent/planexec/

#### Low Coupling (Loosely Coupled)
- **config/** - standalone configuration management
- **history/** - standalone conversation history
- **cache/** - standalone caching layer
- **internal/fsio/** - standalone file I/O abstractions

---

## Plugin Separation Strategy

### Strategy Overview

Given the tight coupling in the agent framework, I propose a **hybrid approach**:

1. **Granular Plugins** for loosely coupled modules (config, history, cache, fsio)
2. **Coarse-Grained Plugins** for tightly coupled modules (agent framework, API client)
3. **Interface-Based Design** to allow future splitting if needed

### Proposed Plugin Structure

```
plugins/
├── chatgpt-config/           # Configuration management
│   ├── config.go
│   ├── manager.go
│   └── store.go
├── chatgpt-history/          # Conversation history
│   ├── history.go
│   ├── manager.go
│   └── store.go
├── chatgpt-cache/            # Caching layer
│   ├── cache.go
│   ├── entry.go
│   └── store.go
├── chatgpt-fsio/             # File I/O abstractions
│   ├── reader.go
│   └── writer.go
├── chatgpt-api/              # API client (coarse-grained)
│   ├── client/
│   │   ├── llm.go
│   │   ├── mcp.go
│   │   ├── media.go
│   │   └── client.go
│   └── http/
│       └── http.go
├── chatgpt-agent-core/       # Agent framework core (coarse-grained)
│   ├── core/
│   │   ├── base_agent.go
│   │   ├── budget.go
│   │   ├── policy.go
│   │   ├── runner.go
│   │   ├── clock.go
│   │   └── transcript_buffer.go
│   ├── types/
│   │   └── types.go
│   └── tools/
│       ├── llm.go
│       ├── shell.go
│       └── files.go
├── chatgpt-agent-react/      # ReAct agent (coarse-grained)
│   ├── react_agent.go
│   └── factory.go
└── chatgpt-agent-planexec/   # Plan-Execute agent (coarse-grained)
    ├── planexec_agent.go
    └── factory.go
```

### Rationale

#### 1. Granular Plugins for Loosely Coupled Modules

**chatgpt-config**, **chatgpt-history**, **chatgpt-cache**, **chatgpt-fsio**

- These modules have minimal dependencies
- Can be used independently by other plugins
- Easy to test and maintain
- Clear boundaries

#### 2. Coarse-Grained Plugins for Tightly Coupled Modules

**chatgpt-api**, **chatgpt-agent-core**, **chatgpt-agent-react**, **chatgpt-agent-planexec**

- These modules have tight coupling (agent/react → agent/core → agent/tools → agent/types)
- Splitting them would create dependency hell
- Keeping them together maintains internal cohesion
- External interfaces remain clean

#### 3. Interface-Based Design

Even with coarse-grained plugins, we can use interfaces to allow future splitting:

```go
// chatgpt-agent-core provides interfaces
type Budget interface {
    AllowIteration(now time.Time) error
    Snapshot(now time.Time) BudgetSnapshot
}

type Policy interface {
    AllowTool(tool string) error
    AllowShellCommand(cmd string) error
    AllowFileOp(op string, path string) error
}

// chatgpt-agent-react implements these interfaces
type ReActAgent struct {
    Budget  core.Budget
    Policy  core.Policy
    // ...
}
```

### Dependency Management

#### Plugin Dependencies

```
chatgpt-config (no dependencies)
chatgpt-history (no dependencies)
chatgpt-cache (no dependencies)
chatgpt-fsio (no dependencies)

chatgpt-api
  ├─> chatgpt-fsio (optional, for file operations)
  └─> chatgpt-history (optional, for conversation context)

chatgpt-agent-core
  └─> chatgpt-fsio (for file operations)

chatgpt-agent-react
  ├─> chatgpt-agent-core
  └─> chatgpt-api (for LLM tool)

chatgpt-agent-planexec
  ├─> chatgpt-agent-core
  └─> chatgpt-api (for LLM tool)
```

#### Plugin Loading Order

1. Load base plugins (config, history, cache, fsio)
2. Load API plugin
3. Load agent-core plugin
4. Load agent implementations (react, planexec)

---

## Alternative: Maximum Granularity (Not Recommended)

If we were to split into the smallest possible plugins, we would have:

```
plugins/
├── chatgpt-config/
├── chatgpt-history/
├── chatgpt-cache/
├── chatgpt-fsio/
├── chatgpt-http/
├── chatgpt-api-client/
├── chatgpt-api-llm/
├── chatgpt-api-mcp/
├── chatgpt-api-media/
├── chatgpt-agent-types/
├── chatgpt-agent-clock/
├── chatgpt-agent-base/
├── chatgpt-agent-budget/
├── chatgpt-agent-policy/
├── chatgpt-agent-runner/
├── chatgpt-agent-transcript/
├── chatgpt-agent-tools-llm/
├── chatgpt-agent-tools-shell/
├── chatgpt-agent-tools-files/
├── chatgpt-agent-utils/
├── chatgpt-agent-react/
└── chatgpt-agent-planexec/
```

### Problems with Maximum Granularity

1. **Dependency Hell**: 20+ plugins with complex dependency graph
2. **Version Management**: Breaking changes in one plugin cascade to many dependents
3. **Testing Difficulty**: Integration tests become complex
4. **Plugin Loading**: Complex loading order and dependency resolution
5. **Maintenance Overhead**: Small changes require updating multiple plugins

### When Maximum Granularity Might Be Useful

- If different agent implementations need different tool sets
- If API components need to be used independently
- If budget/policy enforcement needs to be customized per agent type

---

## Recommended Approach

### Phase 1: Start with Coarse-Grained Plugins

Implement the recommended 8-plugin structure first:

1. **chatgpt-config** - Configuration management
2. **chatgpt-history** - Conversation history
3. **chatgpt-cache** - Caching layer
4. **chatgpt-fsio** - File I/O abstractions
5. **chatgpt-api** - API client (LLM, MCP, media)
6. **chatgpt-agent-core** - Agent framework core
7. **chatgpt-agent-react** - ReAct agent
8. **chatgpt-agent-planexec** - Plan-Execute agent

### Phase 2: Refine Based on Usage

After implementation, monitor usage patterns:

- If API components are used independently, consider splitting chatgpt-api
- If agent tools need customization, consider splitting chatgpt-agent-core
- If budget/policy enforcement varies, extract into separate plugins

### Phase 3: Future-Proof with Interfaces

Even with coarse-grained plugins, use interfaces to allow future splitting:

```go
// Define clear interfaces in chatgpt-agent-core
type Budget interface { ... }
type Policy interface { ... }
type Runner interface { ... }

// Implementations can be extracted later if needed
```

---

## Implementation Considerations

### Plugin Registration

```go
// plugins/chatgpt-agent-react/plugin.go
package chatgpt_agent_react

import (
    "github.com/kardolus/ti-cli/plugin"
)

func init() {
    plugin.Register(&Plugin{
        Name: "chatgpt-agent-react",
        Dependencies: []string{
            "chatgpt-agent-core",
            "chatgpt-api",
        },
        Init: func(deps plugin.Dependencies) error {
            // Initialize plugin with dependencies
            return nil
        },
    })
}
```

### Dependency Injection

```go
// Use dependency injection to decouple plugins
type ReActAgentDeps struct {
    Core   *chatgpt_agent_core.Plugin
    API    *chatgpt_api.Plugin
    Config *chatgpt_config.Plugin
}

func NewReActAgent(deps ReActAgentDeps) (*ReActAgent, error) {
    // Create agent using dependencies
}
```

### Plugin Configuration

```yaml
# plugins.yaml
plugins:
  - name: chatgpt-config
    enabled: true
  - name: chatgpt-history
    enabled: true
  - name: chatgpt-cache
    enabled: true
  - name: chatgpt-fsio
    enabled: true
  - name: chatgpt-api
    enabled: true
    dependencies:
      - chatgpt-fsio
  - name: chatgpt-agent-core
    enabled: true
    dependencies:
      - chatgpt-fsio
  - name: chatgpt-agent-react
    enabled: true
    dependencies:
      - chatgpt-agent-core
      - chatgpt-api
  - name: chatgpt-agent-planexec
    enabled: true
    dependencies:
      - chatgpt-agent-core
      - chatgpt-api
```

---

## Conclusion

### Summary

The ChatGPT CLI source code has a clear modular structure with:

- **Loosely coupled modules**: config, history, cache, fsio → suitable for granular plugins
- **Tightly coupled modules**: agent framework, API client → suitable for coarse-grained plugins

### Recommendation

Implement the **8-plugin hybrid approach**:

1. 4 granular plugins for loosely coupled modules
2. 4 coarse-grained plugins for tightly coupled modules
3. Interface-based design to allow future splitting

This approach balances:
- **Granularity**: Enough separation for flexibility
- **Dependency Management**: Avoids dependency hell
- **Maintainability**: Clear boundaries and cohesive modules
- **Extensibility**: Interfaces allow future refinement

### Next Steps

1. Implement the 8-plugin structure
2. Test plugin loading and dependency resolution
3. Validate that all functionality works
4. Monitor usage patterns for future refinement
