# Ti CLI Architecture Overview

> **Last Updated**: 2026-05-05
> **Version**: 3.0.0-ti-ecosystem-go1.23

## 📐 High-Level Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                         Ti CLI                             │
│                    (Go-based Runtime)                       │
└─────────────────────────────────────────────────────────────┘
                              │
              ┌───────────────┼───────────────┐
              │               │               │
        ┌─────▼─────┐  ┌────▼────┐  ┌─────▼──────┐
        │   Router  │  │  Agents │  │  Plugins   │
        │  Service  │  │ System  │  │  System    │
        └─────┬─────┘  └────┬────┘  └─────┬──────┘
              │            │              │
        ┌─────▼─────┐  ┌────▼────┐  ┌─────▼──────┐
        │  Layers   │  │ Skills  │  │  Context   │
        │  (84+)    │  │  &     │  │  Packing   │
        │           │  │ Learning│  │           │
        └───────────┘  └─────────┘  └────────────┘
```

## 🏗️ Component Architecture

### 1. Core Components

#### **CLI Entry Point** (`apps/cli/cmd/`)
- **root.go**: Main CLI entry point with Cobra framework
- **Commands**: 60+ commands organized by domain
- **Initialization**: Config loading, Router service startup, plugin host

#### **Router Integration** (`apps/router/`)
- **Layers Architecture**: 84+ modular layers
  - `provider/`: 22+ AI provider implementations
  - `routing/`: Model routing, load balancing, adaptive routing
  - `resilience/`: Circuit breaker, rate limiting, caching
  - `memory/`: Knowledge graph memory
  - `rtk/`: Token compression, predictive analytics
  - `engine/`: Decision engine for adaptive routing
- **HTTP Server**: Port 1807 with health checks
- **Configuration**: YAML-based provider config (22+ providers)

#### **Plugin System** (`apps/cli/internal/plugins/`)
- **Old System**: Built-in plugins (devin, logger, permission, etc.)
- **New System (PACK 1)**: Plugin API with kernel-based architecture
- **Compatibility Layer**: Backward compatibility during migration
- **Discovery**: Auto-discovery from `.ti/plugins/` directory

#### **Agent System** (`apps/cli/internal/agents/`)
- **Devin Adapter**: Claude Code integration
- **Native Agents**: Logger, Permission, Status, Tmux
- **Agent Registry**: Dynamic agent management
- **Agent Loop**: Autonomous task execution

#### **Sub-Agent System** (`apps/cli/internal/subagent/`)
- **External CLI Orchestration**: Spawn Claude Code, Codex, Gemini, etc. as sub-agents
- **Parallel Execution**: Run multiple agents concurrently with aggregation modes (merge, vote, race)
- **Profile System**: Agent profiles with YAML front-matter (role, permissions, tools)
- **Runner**: Execute external CLIs via stdio with timeout and context management
- **CLI Registry Integration**: Dynamic CLI discovery and configuration

#### **Multi-Agent Orchestrator** (`apps/cli/internal/plugins/ticlaw/multi/`)
- **Agent Registry**: Agent type management with capabilities
- **Message Bus**: Pub/sub inter-agent communication
- **Task Queue**: Priority-based task distribution
- **Agent Instances**: Running agent lifecycle management
- **Workflow Engine**: Multi-agent coordination and collaboration

#### **Context Management** (`apps/cli/internal/context/`)
- **Scanner**: File system scanning with ignore patterns
- **Packer**: Context compression and serialization
- **Budget**: Token budget management
- **Manager**: Context lifecycle management

### 2. Data Flow

```
User Command
    │
    ▼
┌─────────────┐
│ CLI Parser  │ (Cobra)
└──────┬──────┘
       │
       ▼
┌─────────────┐
│ Config Load │ (kernel.Config)
└──────┬──────┘
       │
       ▼
┌─────────────┐
│ Router Init │ (router.Service)
└──────┬──────┘
       │
   ┌───┴───┐
   │       │
   ▼       ▼
Plugins  Agents
   │       │
   └───┬───┘
       │
       ▼
┌─────────────┐
│ Execution  │
└──────┬──────┘
       │
       ▼
┌─────────────┐
│  Response  │
└─────────────┘
```

### 3. Key Patterns

#### **Dependency Injection**
- Setter-based DI for Router components
- Constructor injection for plugins
- Service locator pattern for global services

#### **Error Handling**
- Structured errors with codes and severity
- Context information and recovery suggestions
- Stack trace capture for debugging
- Fluent builder pattern for error construction

#### **Configuration**
- Layered configuration (env → file → defaults)
- Hot-reload for provider config
- Validation with detailed error messages

#### **Plugin Architecture**
- Plugin interface with lifecycle methods
- Capability-based command attachment
- Health monitoring with circuit breakers
- Graceful degradation

## 🔌 Integration Points

### **External Services**
- **Router Service**: HTTP API on port 1807
- **MCP Servers**: stdio, http, websocket transports
- **AI Providers**: 22+ providers via Router
- **Notion**: Knowledge sync integration
- **GitHub**: PR review, repo management
- **External CLIs**: Claude Code, Codex, Gemini, Qwen, etc. (via sub-agent system)

### **Internal Services**
- **Knowledge Graph Memory**: Persistent memory via server-memory
- **BD Tool**: Task tracking and logging
- **Skills System**: Dynamic skill discovery and execution
- **Beads**: Continuous learning system
- **Sub-Agent System**: External CLI orchestration (see SUB_AGENT_GUIDE.md)
- **Multi-Agent Orchestrator**: Advanced multi-agent workflows (Ti Claw plugin)

## 📊 Performance Characteristics

### **Startup Time**
- CLI startup: ~0.02s (basic commands)
- Router registration: ~0.5s (if Router available)
- Plugin loading: ~0.1s (with auto-load disabled)

### **Context Operations**
- Small projects (<100 files): <1s
- Medium projects (100-1000 files): 2-5s
- Large projects (>1000 files): 5-15s

### **Memory Usage**
- Base CLI: ~50MB
- With Router: ~100MB
- With context: ~150-200MB

## 🔒 Security Architecture

### **Permission System**
- Permission sandbox for file operations
- Policy-based access control
- User-level and system-level permissions

### **Secrets Management**
- Centralized secrets in `Z:\00_SECRET\`
- Environment variable loading
- No secrets in code or config files

### **Authentication**
- Router token validation
- Provider API key management
- Cookie-based authentication for some providers

## 🧪 Testing Architecture

### **Unit Tests**
- Package-level tests (100+ test files)
- Mock-based testing for external dependencies
- Coverage focus on core logic

### **Integration Tests**
- E2E test suite (`apps/cli/e2e/`)
- Real CLI execution testing
- Multi-command workflow testing

### **Performance Tests**
- Benchmark tests for critical paths
- Load testing for Router service
- Memory profiling for context operations

## 📦 Deployment Architecture

### **Build Process**
```bash
cd apps/cli
go build -o ../../bin/ti .
```

### **Distribution**
- Single binary distribution
- No external dependencies for basic operations
- Optional Router service for advanced features

### **Configuration**
- Config file: `~/.ti-cli/config.yaml`
- Environment variables for secrets
- Provider config: `configs/providers.yaml`

## 🔄 Update Strategy

### **Compatibility**
- Backward compatible CLI interface
- Plugin compatibility layer for migration
- Config migration paths

### **Rollback**
- Binary rollback capability
- Config versioning
- Database schema versioning (for memory)

## 📈 Monitoring & Observability

### **Metrics**
- Router metrics: requests, errors, latency
- Agent metrics: execution time, success rate
- Context metrics: pack time, cache hit rate

### **Logging**
- Structured logging with levels
- Request/response logging for Router
- Error logging with stack traces

### **Health Checks**
- CLI health: `ti doctor`
- Router health: `ti auto router status`
- Plugin health: `ti plugin-compat status`

## 🎯 Future Architecture Plans

### **Short Term**
- Complete plugin migration to PACK 1
- Enhanced incremental context packing
- Improved error recovery mechanisms

### **Medium Term**
- Web dashboard for monitoring
- Cloud integration (AWS, GCP, Azure)
- Mobile companion app

### **Long Term**
- Distributed agent orchestration
- Multi-region Router deployment
- Advanced AI capabilities (reasoning, planning)

## 📚 Additional Documentation

- **Quick Start**: [QUICK_START.md](QUICK_START.md)
- **Sub-Agent System**: [SUB_AGENT_GUIDE.md](SUB_AGENT_GUIDE.md)
- **CLI Knowledge Sync**: [CLI_NOTION_SYNC_STRATEGY.md](CLI_NOTION_SYNC_STRATEGY.md)
- **Architecture Details**: [TI-HYBRID-DESIGN.md](../../docs/01-architecture/TI-HYBRID-DESIGN.md)
