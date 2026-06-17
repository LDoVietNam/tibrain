# LobeHub Research & Learning Notes

> **Date**: 2026-05-10
> **Purpose**: Study LobeHub architecture and patterns for TiCrew development
> **Status**: Research completed

---

## 🎯 LobeHub Overview

### What is LobeHub?
- **Open-source AI Agent platform** built on Next.js
- **Mission**: "Find, build, and collaborate with agent teammates that grow with you"
- **Architecture**: Human-agent co-evolving network
- **License**: Open source with 76.8k GitHub stars

### Key Features
- **Agent System**: Persistent AI teammates (not one-off conversations)
- **Multi-model Support**: 30+ AI providers (OpenAI, Claude, Gemini, etc.)
- **Plugin System**: MCP (Multi-Channel Plugin) integration
- **Knowledge Base**: RAG with document storage
- **Memory System**: Long-term context across sessions
- **Multi-channel**: Discord, WeChat, Telegram, Slack integration

---

## 🏗️ Architecture Analysis

### High-Level Architecture
```
Frontend Layer: Next.js RSC + React Router DOM hybrid SPA
     ↓
API Layer: RESTful WebAPI + tRPC Routers
     ↓
Runtime Layer: Model Runtime + Agent Runtime
     ↓
Data Layer: PostgreSQL + Redis + S3 Storage
```

### Frontend Architecture
**Hybrid Routing Approach**:
- **Next.js App Router**: Auth pages, SSR, static routes (`src/app/(backend)/`)
- **React Router DOM**: Main chat SPA, agent interfaces (`src/spa/` + `src/routes/`)

**Tech Stack**:
- UI: `@lobehub/ui`, antd
- CSS: antd-style (CSS-in-JS)
- State: Zustand (slice pattern)
- Data Fetching: SWR + tRPC
- i18n: react-i18next

### Backend Architecture
**Dual API Style**:
- **tRPC Routers**: Type-safe main business routes
  - `lambda/`: Main business (agent, session, message, topic, file, knowledge)
  - `async/`: Long-running operations (file processing, AI tasks)
- **RESTful WebAPI**: Standard HTTP endpoints

**Key Services**:
- Authentication: Better Auth (email/password + SSO)
- Database: PostgreSQL with Drizzle ORM
- Storage: S3-compatible (AWS S3, Cloudflare R2, MinIO)
- Cache: Redis for sessions and performance

---

## 🤖 Agent System Deep Dive

### Agent Components (6 Core Elements)
| Component | Purpose | Implementation |
|-----------|---------|----------------|
| **System Role** | Personality, expertise, behavioral guidelines | System prompt engineering |
| **AI Model** | LLM powering the agent | 30+ provider integrations |
| **Skills** | Capabilities like web search, code execution | Plugin system |
| **Integrations** | MCP connections and external services | MCP server integration |
| **Knowledge Base** | Documents and data for reference | RAG with vector storage |
| **Memory** | Personal context across sessions | Long-term memory pipeline |

### Agent Types
1. **Built-in Agents**: Agent Builder, Pages Agent, Memory Agent
2. **Custom Agents**: User-created for specific workflows
3. **Community Agents**: Ready-to-use shared agents

### Agent Configuration Structure
```typescript
interface AgentConfig {
  // Identity
  id: string;
  name: string;
  avatar?: string;
  description?: string;
  
  // Behavior
  systemRole: string;           // System prompt
  chatConfig: {
    model: string;              // e.g., 'gpt-4'
    provider: string;           // e.g., 'openai'
    temperature: number;        // 0-2
    maxTokens?: number;
  };
  
  // Capabilities
  tools?: string[];             // Tool IDs
  knowledgeBases?: string[];    // KB references
  integrations?: string[];      // MCP integrations
}
```

---

## 🔧 Technical Patterns

### 1. State Management (Zustand)
```typescript
// Slice pattern for organized state
const createAgentSlice = (set: StateCreator) => ({
  agents: [],
  currentAgent: null,
  createAgent: (agent) => set((state) => ({
    agents: [...state.agents, agent]
  })),
  updateAgent: (id, updates) => set((state) => ({
    agents: state.agents.map(agent => 
      agent.id === id ? { ...agent, ...updates } : agent
    )
  }))
});
```

### 2. Agent Runtime Pipeline
```
Context Assembly → Model Reasoning → Tool Execution → Response Generation
```

### 3. Plugin System (MCP)
- **Multi-Channel Plugin**: Standard for tool integration
- **Built-in Tools**: Web search, code execution, image generation
- **Custom Tools**: Developer-defined capabilities

### 4. Memory System
- **Short-term**: Conversation context
- **Long-term**: User preferences, learned patterns
- **Async Processing**: Background memory extraction with Upstash workflows

---

## 📁 Directory Structure Insights

### Key Directories
```
src/
├── app/               # Next.js App Router (auth, API routes)
├── spa/               # React Router DOM SPA
├── components/        # Reusable UI components
├── features/          # Business feature modules
├── store/             # Zustand state management
├── server/            # Server-side modules
│   ├── routers/       # tRPC routers
│   └── services/      # Business logic with DB access
├── libs/              # Third-party integrations
└── types/             # TypeScript definitions
```

### Monorepo Structure
```
packages/
├── @lobechat/model-runtime    # AI provider integrations
├── @lobechat/context-engine    # Tool assembly engine
├── @lobechat/database         # Database schemas
├── @lobechat/builtin-tools    # Built-in tool implementations
└── @lobehub/ui                 # UI component library
```

---

## 🎨 UI/UX Patterns

### Design Principles
- **Glassmorphism**: Modern transparent design
- **Progressive Enhancement**: Features layer on top of core functionality
- **Agent-Centric**: Everything revolves around agent interactions

### Key UI Components
- **Agent Builder**: Smart creation with AI assistance
- **Chat Interface**: Multi-modal conversations
- **Knowledge Base Integration**: In-chat document reference
- **Tool Execution**: Real-time tool usage visualization

---

## 🚀 Deployment Architecture

### Self-Hosting Options
1. **Docker Compose** (Recommended)
2. **Vercel** (Serverless)
3. **Cloud Platforms** (Zeabur, Sealos, Dokploy)

### Required Services
- **PostgreSQL**: Primary database
- **Redis**: Session storage and caching
- **S3 Storage**: File uploads and knowledge bases
- **AI Provider APIs**: At least one (OpenAI, Anthropic, etc.)

### Optional Services
- **Langfuse**: LLM observability
- **OpenTelemetry**: Distributed tracing
- **Searxng**: Privacy-focused web search
- **RustFS/MinIO**: Self-hosted S3 alternative

---

## 💡 Key Learnings for TiCrew

### 1. Architecture Patterns to Adopt
- **Hybrid Routing**: SSR for auth, SPA for main app
- **Dual API Style**: tRPC for type safety, REST for compatibility
- **Agent-Centric Design**: Everything revolves around agents

### 2. Technical Decisions
- **Zustand over Redux**: Simpler state management
- **tRPC for APIs**: Type-safe client-server communication
- **MCP for Plugins**: Standard tool integration
- **PostgreSQL + Redis**: Proven data layer

### 3. Agent System Design
- **6-Component Model**: System Role, Model, Skills, Integrations, Knowledge, Memory
- **Persistent Agents**: Not one-off conversations
- **Knowledge Integration**: RAG with document storage
- **Memory Pipeline**: Long-term context retention

### 4. UI/UX Principles
- **Agent Builder**: Smart creation with AI assistance
- **Glassmorphism Design**: Modern transparent aesthetics
- **Progressive Enhancement**: Core features first

---

## 🔄 Implementation Roadmap for TiCrew

### Phase 1: Core Architecture
- [ ] Set up Next.js + React Router DOM hybrid
- [ ] Implement Zustand state management
- [ ] Create tRPC routers for API
- [ ] Set up PostgreSQL + Redis

### Phase 2: Agent System
- [ ] Implement 6-component agent model
- [ ] Create Agent Builder interface
- [ ] Add basic Skills system
- [ ] Implement Memory pipeline

### Phase 3: Advanced Features
- [ ] MCP plugin integration
- [ ] Knowledge base with RAG
- [ ] Multi-channel support
- [ ] Advanced tool execution

### Phase 4: Production Ready
- [ ] Docker deployment setup
- [ ] Monitoring and observability
- [ ] Performance optimization
- [ ] Security hardening

---

## 📊 Comparison: TiCrew vs LobeHub

| Aspect | TiCrew (Current) | LobeHub (Reference) |
|--------|------------------|---------------------|
| **Architecture** | HTML/CSS/JS (simple) | Next.js + React (complex) |
| **Agent System** | Basic router integration | 6-component model |
| **State Management** | None (vanilla JS) | Zustand slices |
| **API Layer** | Direct router calls | tRPC + REST |
| **Database** | None | PostgreSQL + Redis |
| **Plugin System** | None | MCP integration |
| **Knowledge Base** | None | RAG with storage |
| **Memory** | None | Long-term pipeline |

---

## 🎯 Next Steps for TiCrew

### Immediate Actions
1. **Study LobeHub Source Code**: Deep dive into implementation details
2. **Adopt Agent Model**: Implement 6-component agent system
3. **Add State Management**: Introduce Zustand for complex state
4. **Create API Layer**: Build tRPC-style type-safe APIs

### Medium-term Goals
1. **Plugin System**: Implement MCP for tool integration
2. **Knowledge Base**: Add RAG capabilities
3. **Memory Pipeline**: Implement long-term context
4. **Modern UI**: Upgrade from vanilla JS to React

### Long-term Vision
1. **Multi-channel Support**: Discord, Slack, etc.
2. **Advanced Agent Builder**: AI-assisted agent creation
3. **Enterprise Features**: Teams, collaboration, security
4. **Ecosystem**: Community agents, marketplace

---

*Research completed by Devin CLI*  
*Date: 2026-05-10*  
*Status: Ready for implementation planning*