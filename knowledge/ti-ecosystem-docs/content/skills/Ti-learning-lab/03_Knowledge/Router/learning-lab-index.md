# Ti Learning Lab Index

> **Location**: `Z:\10_WORKPLACE\Ti\Ti-learning-lab\`
> **Purpose**: Index of learning resources - UI, Backend, Patterns, Skills
> **Last Updated**: 2026-05-04

---

## Overview

Ti Learning Lab là kho lưu trữ kiến thức học được từ các open-source projects, patterns, và best practices. Bài viết này index toàn bộ resources có sẵn.

---

## Structure

```
Ti-learning-lab/
├── 01_Learning/
│   └── lab/
│       ├── 03_Knowledge/           # Knowledge base (MD files + Code)
│       │   ├── browser/            # Browser automation projects
│       │   ├── agents/             # Agent patterns (MD only)
│       │   ├── api/                # API patterns (MD only)
│       │   ├── skills/             # Skills definitions
│       │   ├── auto-reg-tools/     # Auto registration tools
│       │   └── *.md                # Various knowledge docs
│       └── 06_HandsOn/             # Hands-on projects
└── 03_Knowledge/                   # Router-specific knowledge
```

---

## 1. Browser Automation (UI + Backend)

### Location: `01_Learning/lab/03_Knowledge/browser/`

#### 1.1 Skyvern (Full Stack)
**Path**: `browser/skyvern/`

**Frontend** (`skyvern-frontend/`):
- **Tech Stack**: React + TypeScript + Vite + Radix UI + Tailwind
- **Key Features**:
  - Task/Workflow management
  - Credentials management
  - Browser session persistence
  - Real-time WebSocket updates
  - Visual workflow builder
- **Patterns Learned**:
  - Multi-view navigation (SwitchBarNavigation)
  - Table với expandable rows
  - Dialog/Modal patterns
  - Custom hooks (useCredentialModalState, useBackgroundCredentialTest)
  - Zustand state management
  - TanStack Query for data fetching
- **Key Files**:
  - `src/routes/credentials/` - Credentials management
  - `src/routes/tasks/detail/` - Task details with multi-view
  - `src/routes/workflows/` - Workflow page with table
  - `src/components/SwitchBarNavigation.tsx` - Navigation pattern

**Backend** (`skyvern/`):
- **Tech Stack**: Python + FastAPI + SQLAlchemy + PostgreSQL
- **Key Features**:
  - Browser automation với Playwright
  - LLM integration (OpenAI, Claude, etc.)
  - Workflow execution engine
  - Credential management (Bitwarden integration)
  - Real-time streaming (WebSocket)
- **Patterns Learned**:
  - Service layer architecture
  - Database models với SQLAlchemy
  - API handlers với FastAPI
  - Background task processing
  - Error handling và logging
- **Key Files**:
  - `skyvern/core/` - Core business logic
  - `skyvern/services/` - Service layer
  - `skyvern/webeye/` - Browser automation
  - `skyvern/schemas/` - Pydantic models

#### 1.2 Browser-Use (Python Backend)
**Path**: `browser/browser-use/`

**Tech Stack**: Python + Playwright + LangChain

**Key Features**:
- Browser automation agent
- LLM-powered actions
- DOM interaction
- Screenshot analysis
- File operations

**Patterns Learned**:
- Agent architecture
- Tool registration
- State management
- Error recovery
- Observability

**Key Files**:
- `browser_use/agent/` - Agent implementation
- `browser_use/controller/` - Browser controller
- `browser_use/dom/` - DOM interaction
- `browser_use/tools/` - Tool definitions

#### 1.3 Donut Browser (Go Backend)
**Path**: `browser/donutbrowser/`, `browser/donutbrowser-main/`, `browser/donutbrowser-custom/`

**Tech Stack**: Go + Chromium

**Key Features**:
- Custom browser implementation
- Profile management
- Extension support
- Sync functionality

**Patterns Learned**:
- Go patterns cho browser
- Profile isolation
- Extension API
- Sync protocols

---

## 2. Agent Patterns (Knowledge Only)

### Location: `01_Learning/lab/03_Knowledge/agents/`

#### 2.1 Auth Agents
**Path**: `agents/auth/`

**Files**:
- `auth-agent-status.md` - Auth agent status patterns

#### 2.2 Orchestration
**Path**: `agents/orchestration/`

**Files**:
- `ai-agent-orchestration-repos.md` - AI agent orchestration repositories

---

## 3. API Patterns (Knowledge Only)

### Location: `01_Learning/lab/03_Knowledge/api/`

#### 3.1 API Integration
**Path**: `api/`

**Files**:
- `api-integration-state-management.md` - API integration state management

---

## 4. Skills Definitions

### Location: `01_Learning/lab/03_Knowledge/skills/`

#### 4.1 Omni Skills
**Path**: `skills/awesome-omni-skills-main/`

**Content**: Skill definitions cho Claude Code agents

---

## 5. Auto Registration Tools

### Location: `01_Learning/lab/03_Knowledge/auto-reg-tools/`

#### 5.1 Router Providers
**Path**: `auto-reg-tools/router-providers/`

**Providers**: claude, codex, cursor, gemini, gmail, kiro, modal, openai, other, qwen, windsurf

**Content**:
- Provider-specific documentation
- Auth flow patterns
- Integration guides

---

## 6. Router-Specific Knowledge

### Location: `03_Knowledge/Router/`

#### 6.0 Skills & Automation
**Skills Location**: `Z:\10_WORKPLACE\Ti\.devin\skills\`

**Key Skills**:
- `omniroute-knowledge-sync` - Document OmniRoute patterns and automatically sync to Notion with changelog entries
- `tdd-workflow` - Test-driven development workflow
- `frontend-patterns` - Frontend development patterns
- `backend-patterns` - Backend architecture patterns
- `golang-testing` - Go testing patterns
- `kotlin-testing` - Kotlin testing patterns

**Workflow Integration**:
- Use `omniroute-knowledge-sync` when researching OmniRoute codebase
- Automatically syncs findings to Notion knowledge base
- Logs tasks via BD tool for tracking
- Creates changelog entries in Notion database

#### 6.1 Backend Patterns
**Files**:
- `cache-agent-patterns-go.md` - Cache agent patterns in Go
- `go-cli-ai-patterns.md` - Go CLI AI patterns
- `metrics-cli-agent-patterns-go.md` - Metrics CLI agent patterns in Go
- `oauth-authentication-patterns-go.md` - OAuth authentication patterns in Go
- `provider-management.md` - Provider management patterns
- `omniroute-patterns.md` - OmniRoute backend architecture patterns (newly created)
- `omniroute-backend-analysis-ti-router.md` - OmniRoute backend analysis for Ti Router (newly created)

#### 6.2 Frontend Patterns
**Files**:
- `frontend-ai-components.md` - Frontend AI components
- `skyvern-ui-patterns.md` - Skyvern UI patterns (newly created)
- `omniroute-ui-patterns.md` - OmniRoute UI patterns (newly created)
- `omniroute-ui-application-plan.md` - OmniRoute UI application plan for Ti Router (newly created)

#### 6.3 General Knowledge
**Files**:
- `01_AI_ROUTER_BEST_PRACTICES.md` - AI router best practices
- `07_DOCS-README.md` - Documentation README
- `INDEX.md` - Knowledge index
- `MANIFEST.md` - Knowledge manifest
- `prompt-engineering-repos.md` - Prompt engineering repositories
- `RESEARCH_ANALYSIS.md` - Research analysis
- `RESEARCH_INDEX.md` - Research index

---

## 7. What Can Be Learned

### 7.1 From Skyvern (Full Stack)

**Frontend**:
- ✅ Multi-view navigation patterns
- ✅ Table với expandable rows
- ✅ Dialog/Modal patterns
- ✅ Custom hooks patterns
- ✅ State management với Zustand
- ✅ Data fetching với TanStack Query
- ✅ Real-time updates với WebSocket
- ✅ Form validation với React Hook Form
- ✅ Code editor integration (CodeMirror)
- ✅ Workflow builder UI

**Backend**:
- ✅ Service layer architecture
- ✅ Database models với SQLAlchemy
- ✅ API handlers với FastAPI
- ✅ Background task processing
- ✅ Browser automation với Playwright
- ✅ Credential management (Bitwarden)
- ✅ Real-time streaming (WebSocket)
- ✅ Error handling và logging
- ✅ Testing patterns (pytest)
- ✅ Docker containerization

### 7.2 From Browser-Use (Python Backend)

**Agent Architecture**:
- ✅ Agent implementation patterns
- ✅ Tool registration system
- ✅ State management
- ✅ Error recovery
- ✅ Observability
- ✅ DOM interaction patterns
- ✅ Screenshot analysis
- ✅ File operations

### 7.3 From Donut Browser (Go Backend)

**Browser Implementation**:
- ✅ Go patterns cho browser
- ✅ Profile isolation
- ✅ Extension API
- ✅ Sync protocols
- ✅ Chromium integration

### 7.4 From Router Knowledge

**Backend (Go)**:
- ✅ Cache agent patterns
- ✅ CLI AI patterns
- ✅ Metrics patterns
- ✅ OAuth authentication
- ✅ Provider management
- ✅ OmniRoute backend patterns
- ✅ OmniRoute backend analysis (so sánh với Ti Router)

**Frontend**:
- ✅ AI components
- ✅ UI patterns (from Skyvern)
- ✅ UI patterns (from OmniRoute)

---

## 8. Learning Path Recommendations

### For Frontend Developers
1. Start with **Skyvern Frontend** (`skyvern-frontend/`)
2. Study **Credentials Management** pattern
3. Learn **Task Details** multi-view pattern
4. Explore **Workflow Page** table pattern
5. Apply patterns to Ti Router UI

### For Backend Developers
1. Start with **Skyvern Backend** (`skyvern/`)
2. Study **Service Layer** architecture
3. Learn **Database Models** với SQLAlchemy
4. Explore **Browser Automation** với Playwright
5. Study **API Handlers** với FastAPI
6. Apply patterns to Ti Router (Go equivalent)

### For Full Stack Developers
1. Study **Skyvern** full stack (frontend + backend)
2. Explore **Browser-Use** agent patterns
3. Learn **Donut Browser** Go patterns
4. Apply to Ti Router full stack

---

## 9. How to Use This Index

### To Find UI Patterns
1. Go to `browser/skyvern/skyvern-frontend/`
2. Explore `src/routes/` for page patterns
3. Explore `src/components/` for component patterns
4. Check `skyvern-ui-patterns.md` for documented patterns

### To Find Backend Patterns
1. Go to `browser/skyvern/skyvern/`
2. Explore `skyvern/core/` for business logic
3. Explore `skyvern/services/` for service layer
4. Explore `browser/browser-use/` for agent patterns
5. Check `03_Knowledge/Router/` for Go-specific patterns

### To Find Agent Patterns
1. Go to `browser/browser-use/`
2. Explore `browser_use/agent/`
3. Check `agents/` folder for documented patterns

---

## 10. Future Learning Opportunities

### Not Yet Documented
- **Skyvern Backend**: Service layer, database models, API handlers
- **Browser-Use**: Agent architecture, tool registration, state management
- **Donut Browser**: Go browser patterns, profile isolation
- **Skyvern Frontend**: Workflow builder, code editor integration, WebSocket

### Recommended Next Steps
1. ✅ Document Skyvern UI patterns (COMPLETED)
2. ✅ Document OmniRoute backend patterns (COMPLETED)
3. ✅ Document OmniRoute UI patterns (COMPLETED)
4. ✅ Analyze OmniRoute backend patterns vs Ti Router (COMPLETED)
5. Document Skyvern backend patterns
6. Document Browser-Use agent patterns
7. Document Donut Browser Go patterns
8. Create comparison docs between different implementations
9. Create best practices guide combining all patterns

---

## 7. Router Samples (Full Stack - TypeScript/Node.js)

### Location: `07_Repositories/router/OmniRoute/`

### 7.1 OmniRoute (Full Stack)
**Tech Stack**: Next.js 16, TypeScript 5.9, SQLite (better-sqlite3), Tailwind CSS v4

**Features**:
- 160+ AI providers
- Multi-tier fallback (Subscription → API Key → Cheap → Free)
- 13 routing strategies (priority, weighted, round-robin, P2C, etc.)
- MCP Server (29 tools, 3 transports, 10 scopes)
- A2A Protocol (agent-to-agent communication)
- Electron desktop app
- SSE streaming
- Format translation (OpenAI ↔ Claude ↔ Gemini)
- OAuth flows (9 providers)
- Cookie-based authentication
- Circuit breaker
- Rate limiting (token bucket)
- Token refresh

**Architecture Patterns Learned**:
- **Database Layer**: Domain modules (22 files), versioned migrations (21 files), WAL journaling, encryption at rest
- **Executor Pattern**: Strategy pattern với BaseExecutor, provider-specific executors, fallback mechanism, retry logic
- **Combo Routing Engine**: 13 routing strategies, combo-first design, circuit breaker integration, fallback chains
- **Translator Pattern**: OpenAI as canonical format, normalize → translate → denormalize, tool call normalization
- **MCP Server**: Zod validation, scope enforcement, audit logging (SHA-256), 29 tools across 2 phases
- **API Routes**: CORS preflight, Zod validation, optional auth, policy enforcement, prompt injection guard
- **Services (36+)**: Rate limit manager, token refresh, account fallback, quota monitoring, session management

**UI Patterns Learned**:
- **DataTable**: Configurable data table with sticky header, row click, loading/empty states
- **Modal**: Modal with overlay, escape key, focus trap, body scroll lock, multiple sizes
- **Card**: Card with title/subtitle/icon/action, sub-components (Section, Row, ListItem)
- **Button**: Variants (primary, secondary, outline, ghost, danger), sizes, icons, loading state
- **Input/Select**: Label, error, hint, icon support, accessibility
- **Toggle**: Switch component with sizes, label, description
- **Badge**: Variants, sizes, dot, icon
- **FilterBar**: Search input + filter chips with dropdown
- **OAuthModal**: Complex multi-step OAuth flow (waiting → input → success → error)
- **EmptyState**: Empty state with icon, title, description, action button
- **Loading**: Spinner, page loading, skeleton, card skeleton
- **SegmentedControl**: Tab-like control with icons

**Key Files**:
- `src/lib/db/core.ts` - DB singleton, schema, helpers
- `src/lib/db/migrationRunner.ts` - Versioned migrations
- `open-sse/executors/base.ts` - BaseExecutor pattern
- `open-sse/services/combo.ts` - Combo routing engine
- `open-sse/translator/index.ts` - Format translation
- `open-sse/mcp-server/server.ts` - MCP server implementation
- `src/app/api/v1/chat/completions/route.ts` - API route pattern
- `src/shared/components/` - Shared UI components (DataTable, Modal, Card, Button, etc.)

**Documentation**:
- `omniroute-patterns.md` - Backend architecture patterns
- `omniroute-ui-patterns.md` - UI component patterns (newly created)

**Applied to Ti Router**:
- ✅ Database layer: Domain modules, migrations, encryption (can apply)
- ✅ Executor pattern: Strategy pattern cho providers (can apply)
- ✅ Combo routing: Multi-strategy routing engine (can apply)
- ✅ Translator: Format conversion between providers (can apply)
- ✅ MCP Server: Enhanced tools, scope enforcement, audit logging (can apply)
- ✅ UI Components: DataTable, Modal, Card, Button, Input, Select, Badge, FilterBar, EmptyState, Loading, Toggle, SegmentedControl (can apply)

---

## 11. References

- **Skyvern**: https://github.com/skyvern-ai/skyvern
- **Browser-Use**: https://github.com/browser-use/browser-use
- **Donut Browser**: Custom browser implementation
- **Ti Router**: `Z:\10_WORKPLACE\Ti\apps\router\`
- **Ti Learning Lab**: `Z:\10_WORKPLACE\Ti\Ti-learning-lab\`
