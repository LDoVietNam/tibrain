# AI Agent Orchestration Repos - Research

> **Version**: 1.0.0  
> **Last Updated**: 2026-04-28  
> **Category**: Research  
> **Language**: Tiếng Việt

---

## 📋 Tổng Quan

Research về các open source repositories liên quan đến AI agent orchestration, multi-agent systems, và agent coordination để học hỏi architecture patterns và best practices.

## 🔍 Top Repositories

### 1. RuFlo (ruvnet/ruflo)

**URL**: https://github.com/ruvnet/ruflo  
**Stars**: 33.9k  
**Language**: TypeScript  
**Updated**: 9 hours ago

**Mô tả**: Leading agent orchestration platform cho Claude. Deploy intelligent multi-agent swarms, coordinate autonomous workflows.

**Key Features**:
- Multi-agent swarm orchestration
- Autonomous workflow coordination
- Claude Code integration
- Agent coordination mechanisms

**Architecture Patterns**:
- Swarm-based architecture
- Event-driven coordination
- Agent hierarchy

**License**: Open source

**Lessons cho Ti**:
- Swarm pattern cho multi-agent coordination
- Event-driven architecture cho agent communication
- Claude Code integration patterns

---

### 2. Edict (cft0808/edict)

**URL**: https://github.com/cft0808/edict  
**Stars**: 15.5k  
**Language**: Python  
**Updated**: Yesterday

**Mô tả**: OpenClaw Multi-Agent Orchestration System với 9 specialized AI agents, real-time dashboard, model config, và full audit trails.

**Key Features**:
- 9 specialized AI agents
- Real-time dashboard
- Model configuration
- Full audit trails
- Kanban integration

**Architecture Patterns**:
- Specialized agent architecture
- Real-time monitoring dashboard
- Audit trail system
- Kanban-based task management

**License**: Open source

**Lessons cho Ti**:
- Specialized agents pattern
- Real-time dashboard design
- Audit trail implementation
- Kanban integration cho task tracking

---

### 3. Multi-Agent Shogun (yohey-w/multi-agent-shogun)

**URL**: https://github.com/yohey-w/multi-agent-shogun  
**Stars**: 1.2k  
**Language**: Shell  
**Updated**: 9 days ago

**Mô tả**: Samurai-inspired multi-agent system cho Claude Code. Orchestrate parallel AI tasks via tmux với shogun → karo → ashigaru hierarchy.

**Key Features**:
- Tmux-based orchestration
- Parallel AI task execution
- Hierarchical agent structure (shogun → karo → ashigaru)
- Claude Code integration

**Architecture Patterns**:
- Hierarchy-based agent organization
- Tmux cho parallel execution
- Shell-based orchestration

**License**: Open source

**Lessons cho Ti**:
- Hierarchy pattern cho agent organization
- Tmux cho parallel execution
- Shell-based orchestration

---

### 4. Golutra (golutra/golutra)

**URL**: https://github.com/golutra/golutra  
**Stars**: 3.3k  
**Language**: Rust  
**Updated**: 21 days ago

**Mô tả**: Multi-agent AI orchestration platform cho automation, workflows, và developer tools. Transforms Codex, Claude Code, và OpenClaw.

**Key Features**:
- Multi-agent AI orchestration
- Workflow automation
- Developer tools integration
- Cross-platform support (Codex, Claude Code, OpenClaw)

**Architecture Patterns**:
- Rust-based architecture
- Cross-platform agent integration
- Workflow automation engine

**License**: Open source

**Lessons cho Ti**:
- Rust cho high-performance agent orchestration
- Cross-platform integration patterns
- Workflow automation design

---

### 5. Solace Agent Mesh (SolaceLabs/solace-agent-mesh)

**URL**: https://github.com/SolaceLabs/solace-agent-mesh  
**Stars**: 3.3k  
**Language**: Python  
**Updated**: 1 minute ago

**Mô tả**: Event-driven framework designed to build và orchestrate multi-agent AI systems. Enables seamless integration của AI agents với real-time event streaming.

**Key Features**:
- Event-driven architecture
- Real-time event streaming
- Multi-agent system orchestration
- Enterprise framework

**Architecture Patterns**:
- Event-driven architecture
- Real-time streaming
- Enterprise-grade framework
- MCP integration

**License**: Open source

**Lessons cho Ti**:
- Event-driven architecture cho agent coordination
- Real-time streaming integration
- Enterprise framework design
- MCP integration patterns

---

### 6. Claude Multi-Agent Project Manager (bobmatnyc/claude-mpm)

**URL**: https://github.com/bobmatnyc/claude-mpm  
**Stars**: 122  
**Language**: Python  
**Updated**: 2 days ago

**Mô tả**: Claude Multi-Agent Project Manager với multi-channel orchestration, GitHub-first SDK mode, và plugin system cho Claude.

**Key Features**:
- Multi-channel orchestration
- GitHub-first SDK mode
- Plugin system
- Project management

**Architecture Patterns**:
- Multi-channel orchestration
- Plugin-based extensibility
- GitHub integration
- SDK mode

**License**: Open source

**Lessons cho Ti**:
- Multi-channel orchestration pattern
- Plugin system design
- GitHub-first integration
- SDK mode implementation

---

### 7. MassGen (massgen/MassGen)

**URL**: https://github.com/massgen/MassGen  
**Stars**: 967  
**Language**: Python  
**Updated**: Yesterday

**Mô tả**: Open-source multi-agent scaling system chạy trong terminal, autonomously orchestrating frontier models và agents.

**Key Features**:
- Terminal-based execution
- Multi-agent scaling
- Autonomous orchestration
- Frontier model integration

**Architecture Patterns**:
- Terminal-based CLI
- Scaling system
- Autonomous orchestration

**License**: Open source

**Lessons cho Ti**:
- Terminal-based agent orchestration
- Scaling system design
- Autonomous orchestration patterns

---

### 8. ccswarm (nwiizo/ccswarm)

**URL**: https://github.com/nwiizo/ccswarm  
**Stars**: 137  
**Language**: Rust  
**Updated**: Mar 5

**Mô tả**: Multi-agent orchestration system sử dụng Claude Code với Git worktree isolation và specialized AI agents cho collaborative development.

**Key Features**:
- Claude Code integration
- Git worktree isolation
- Specialized AI agents
- Collaborative development

**Architecture Patterns**:
- Git worktree isolation
- Claude Code integration
- Specialized agents
- Collaborative development

**License**: Open source

**Lessons cho Ti**:
- Git worktree isolation pattern
- Claude Code integration
- Specialized agents design

---

## 🎓 Common Patterns

### 1. Event-Driven Architecture

**Repos sử dụng**: Solace Agent Mesh, RuFlo

**Pattern**:
- Agents communicate qua events
- Event bus cho coordination
- Real-time event streaming

**Benefits cho Ti**:
- Scalable agent communication
- Real-time coordination
- Loose coupling giữa agents

### 2. Hierarchy-Based Organization

**Repos sử dụng**: Multi-Agent Shogun, Edict

**Pattern**:
- Agents organized trong hierarchy
- Clear chain of command
- Specialized roles per level

**Benefits cho Ti**:
- Clear agent responsibilities
- Scalable organization
- Easy to manage complex workflows

### 3. Plugin System

**Repos sử dụng**: Claude MPM, RuFlo

**Pattern**:
- Core system + plugins
- Dynamic plugin loading
- Plugin lifecycle management

**Benefits cho Ti**:
- Extensibility without modifying core
- Custom agent capabilities
- Community-driven extensions

### 4. Real-Time Dashboard

**Repos sử dụng**: Edict, RuFlo

**Pattern**:
- Real-time agent monitoring
- Task visualization
- Performance metrics

**Benefits cho Ti**:
- Visibility vào agent operations
- Real-time debugging
- Performance monitoring

### 5. Git Integration

**Repos sử dụng**: Claude MPM, ccswarm

**Pattern**:
- Git-based task tracking
- Worktree isolation
- GitHub-first workflow

**Benefits cho Ti**:
- Version control cho agent tasks
- Isolated development environments
- GitHub integration

## 🔧 Technical Stack Comparison

| Repo | Language | Architecture | Key Features |
|------|----------|--------------|--------------|
| RuFlo | TypeScript | Swarm-based | Multi-agent swarm, autonomous workflows |
| Edict | Python | Specialized agents | 9 agents, dashboard, audit trails |
| Shogun | Shell | Hierarchy-based | Tmux orchestration, parallel tasks |
| Golutra | Rust | Cross-platform | Multi-platform integration, automation |
| Solace Mesh | Python | Event-driven | Real-time streaming, enterprise |
| Claude MPM | Python | Plugin-based | Multi-channel, GitHub integration |
| MassGen | Python | Scaling system | Terminal-based, autonomous |
| ccswarm | Rust | Git-based | Worktree isolation, collaborative |

## 📊 Recommendations cho Ti

### 1. Adopt Event-Driven Architecture

**Why**: Scalable, real-time coordination, loose coupling

**Implementation**:
- Use event bus cho agent communication
- Implement pub/sub pattern
- Real-time streaming integration

### 2. Implement Hierarchy-Based Organization

**Why**: Clear responsibilities, scalable organization

**Implementation**:
- Define agent hierarchy (Architect → Executor → Worker)
- Specialized roles per level
- Clear chain of command

### 3. Add Plugin System

**Why**: Extensibility, community contributions

**Implementation**:
- Core system + plugin architecture
- Dynamic plugin loading
- Plugin lifecycle management

### 4. Build Real-Time Dashboard

**Why**: Visibility, debugging, monitoring

**Implementation**:
- Real-time agent monitoring
- Task visualization
- Performance metrics

### 5. Integrate Git Workflows

**Why**: Version control, isolation, collaboration

**Implementation**:
- Git-based task tracking
- Worktree isolation
- GitHub integration

## 📚 References

- RuFlo: https://github.com/ruvnet/ruflo
- Edict: https://github.com/cft0808/edict
- Multi-Agent Shogun: https://github.com/yohey-w/multi-agent-shogun
- Golutra: https://github.com/golutra/golutra
- Solace Agent Mesh: https://github.com/SolaceLabs/solace-agent-mesh
- Claude MPM: https://github.com/bobmatnyc/claude-mpm
- MassGen: https://github.com/massgen/MassGen
- ccswarm: https://github.com/nwiizo/ccswarm

---

*Last Updated: 2026-04-28*
