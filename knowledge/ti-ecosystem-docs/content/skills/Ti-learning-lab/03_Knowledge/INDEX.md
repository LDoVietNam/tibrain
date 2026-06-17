# Ti Knowledge Index

> **Single source of truth** for all Ti documentation, protocols, research, and agent guides.
> **Location**: `Z:\10_WORKPLACE\Ti\Ti-learning-lab\03_Knowledge\`
> **Last Updated**: 2026-05-05

---

## Structure Overview

```
03_Knowledge/
├── INDEX.md                    ← You are here
├── README.md                   ← Overview
├── MANIFEST.md                 ← Manifest
├── registry.json               ← Registry
│
├── 00_META/                    ← Meta files and plans ✅
│
├── agents/                     ← Agent framework and orchestration ✅
├── ticlaw/                     ← Ti-claw specific documentation ✅
├── api/                        ← API integration ✅
├── archive/                    ← Archived files ✅
├── auto-reg-tools/             ← Auto registration tools ✅
├── cache/                      ← Cache ✅
├── cli/                        ← CLI documentation ✅
├── computer-vision/            ← Computer vision ✅
├── devin/                      ← Devin porting ✅
├── docs/                       ← Documentation references ✅
├── frontend/                   ← Frontend patterns ✅
├── hot-reload/                 ← Hot reload ✅
├── lessons/                    ← Lessons learned ✅
├── metrics/                    ← Metrics ✅
├── notion/                     ← Notion integration ✅
├── patterns/                   ← Patterns ✅
├── prompt-engineering/         ← Prompt engineering ✅
├── protocols/                  ← Protocols ✅
├── provider/                   ← Provider management ✅
├── research/                   ← Research findings ✅
├── Router/                     ← Router (English) ✅
├── sources/                    ← Sources ✅
├── storage/                    ← Large files & binaries ✅
├── tibrain/                    ← Tibrain ✅
├── ti-router/                  ← Router (Vietnamese) ✅
└── tool/                       ← Tool ✅
```

**Notes:**
- ✅ = Has INDEX.md for navigation
- `browser/` and `skills/` moved to `07_Repositories/` (source code repos)
- Removed empty folders: `browser-automation/`, `cheatsheets/`, `decisions/`, `runbooks/`, `spectre/`, `donut-browser/`

---

## Quick Navigation

|| Need | Go to |
||------|-------|
|| Agent automation rules | `agents/` |
|| Agent orchestration research | `agents/` |
|| Ti-claw specific docs | `ticlaw/` |
|| API integration patterns | `api/` |
|| Authentication & OAuth | `agents/` |
|| CLI documentation | `cli/` |
|| Computer vision | `computer-vision/` |
|| Devin porting | `devin/` |
|| Frontend patterns | `frontend/` |
|| Lessons learned | `lessons/` |
|| Metrics | `metrics/` |
|| Notion integration | `notion/` |
|| Patterns | `patterns/` |
|| Prompt engineering | `prompt-engineering/` |
|| Provider management | `provider/` |
|| Research findings | `research/` |
|| Router architecture (English) | `Router/` |
|| Router architecture (Vietnamese) | `ti-router/` |
|| Tibrain | `tibrain/` |
|| Browser automation docs | `07_Repositories/browser-docs/` |
|| Skills docs | `07_Repositories/skills-docs/` |

---

## Detailed Structure

### 📁 00_META - Meta Files

Meta documentation and planning files.

- **[INDEX.md](00_META/INDEX.md)** - Meta folder index
- **[REORGANIZATION_PLAN.md](00_META/REORGANIZATION_PLAN.md)** - Folder reorganization plan

### 🤖 agents - Agent Framework

Agent framework, orchestration, and automation guides.

- **[INDEX.md](agents/INDEX.md)** - Agent framework knowledge base
- `TI_AGENT_FRAMEWORK_ARCHITECTURE.md` - Agent framework architecture
- `TI_AGENT_FRAMEWORK_FEASIBILITY.md` - Feasibility analysis
- `SUB_AGENT_GUIDE.md` - Sub-agent guide
- `AGENT_QUALITY_WORKFLOW.md` - Quality workflow

### 🦞 ticlaw - Ti-Claw

Ti-claw specific documentation and guides.

- **[INDEX.md](ticlaw/INDEX.md)** - Ti-Claw knowledge base

### 🔌 api - API Integration

API integration patterns and state management.

- **[INDEX.md](api/INDEX.md)** - API integration knowledge base
- `api-integration-state-management.md` - State management patterns

### 📦 archive - Archive

Archived logs and old files.

- **[INDEX.md](archive/INDEX.md)** - Archive index
- `分析sharedchatfun的cook/` - Chinese folder (moved 2026-05-05)

### 🛠️ auto-reg-tools - Auto Registration Tools

Auto registration tools and provider configurations.

- **[INDEX.md](auto-reg-tools/INDEX.md)** - Auto registration tools knowledge base

### 💾 cache - Cache

Cache-related documentation.

- **[INDEX.md](cache/INDEX.md)** - Cache knowledge base
- `cache-agent-status.md` - Cache agent status

### 💻 cli - CLI Documentation

Ti CLI documentation (cleaned and organized).

- **[INDEX.md](cli/INDEX.md)** - CLI knowledge base
- `core/` - Core CLI architecture
- `features/` - CLI features
- `integration/` - CLI integration
- `deployment/` - CLI deployment
- `sources/` - CLI source analysis

### 👁️ computer-vision - Computer Vision

Computer vision documentation.

- **[INDEX.md](computer-vision/INDEX.md)** - Computer vision knowledge base

### 🔄 devin - Devin Porting

Devin API architecture and Go port plans.

- **[INDEX.md](devin/INDEX.md)** - Devin knowledge base

### 📚 docs - Documentation

Documentation references.

- **[INDEX.md](docs/INDEX.md)** - Documentation knowledge base
- `07_DOCS-README.md` - Documentation README

### 🎨 frontend - Frontend Patterns

Frontend patterns and components.

- **[INDEX.md](frontend/INDEX.md)** - Frontend knowledge base
- `frontend-ai-components.md` - Frontend AI components

### ⚡ hot-reload - Hot Reload

Hot reload documentation.

- **[INDEX.md](hot-reload/INDEX.md)** - Hot reload knowledge base
- `hot-reload-architecture-design.md` - Architecture design
- `hot-reload-implementation.md` - Implementation

### 📖 lessons - Lessons Learned

Lessons learned from various projects and implementations.

- **[INDEX.md](lessons/INDEX.md)** - Lessons learned knowledge base
- `CLI_PROXY_API_LESSONS.md` - CLI proxy API lessons
- `JUNIE_CLI_LESSONS.md` - Junie CLI lessons
- `TI_INTERNAL_LESSONS.md` - Ti internal lessons
- `UI_LESSONS.md` - UI lessons
- `cli/` - CLI lessons
- `mcphub/` - MCP hub lessons
- `router/` - Router lessons
- `tibrain/` - Tibrain lessons

### 📊 metrics - Metrics

Metrics documentation.

- **[INDEX.md](metrics/INDEX.md)** - Metrics knowledge base
- `metrics-cli-agent-patterns-go.md` - CLI agent metrics patterns

### 📝 notion - Notion Integration

Notion synchronization and integration.

- **[INDEX.md](notion/INDEX.md)** - Notion integration knowledge base
- `CLI_NOTION_SYNC_APPROVAL.md` - Notion sync approval
- `CLI_NOTION_SYNC_ARCHITECTURE.md` - Notion sync architecture
- `CLI_NOTION_SYNC_EVALUATION.md` - Notion sync evaluation
- `CLI_NOTION_SYNC_INTEGRATION_POINTS.md` - Notion sync integration points
- `CLI_NOTION_SYNC_STRATEGY.md` - Notion sync strategy
- `NOTION_DATABASE_CONFIGURATION.md` - Notion database configuration

### 🔧 patterns - Patterns

Various patterns and best practices.

- **[INDEX.md](patterns/INDEX.md)** - Patterns knowledge base
- `cache-agent-patterns-go.md` - Cache agent patterns (Go)
- `go-cli-ai-patterns.md` - Go CLI AI patterns
- `html-ui-alternative.md` - HTML UI alternative
- `oauth-authentication-patterns-go.md` - OAuth authentication patterns (Go)

### 💬 prompt-engineering - Prompt Engineering

Prompt engineering repositories and techniques.

- **[INDEX.md](prompt-engineering/INDEX.md)** - Prompt engineering knowledge base
- `prompt-engineering-repos.md` - Prompt engineering repos

### 📡 protocols - Protocols

Protocol documentation.

- **[INDEX.md](protocols/INDEX.md)** - Protocols knowledge base
- `beads-protocol.md` - Beads protocol

### 🏢 provider - Provider Management

Provider management documentation.

- **[INDEX.md](provider/INDEX.md)** - Provider management knowledge base
- `provider-management.md` - Provider management

### 🔬 research - Research

Research findings and analysis.

- **[INDEX.md](research/INDEX.md)** - Research knowledge base
- `RESEARCH_ANALYSIS.md` - Research analysis
- `RESEARCH_INDEX.md` - Research index
- `docs/` - Detailed research documentation
- `AGENTS_VIETNAMESE.md` - Vietnamese agents documentation
- `GO_SDK_PATTERNS.md` - Go SDK patterns
- `Implementation.md` - Implementation notes
- `implementation-plan.md` - Implementation plan
- `key-concepts.md` - Key concepts
- `video-notes.md` - Video notes

### 🌐 Router - Router (English)

Router architecture, patterns (English documentation).

- **[INDEX.md](Router/INDEX.md)** - Router knowledge base (English)
- `01_AI_ROUTER_BEST_PRACTICES.md` - AI router best practices
- `9router/` - 9router documentation
- `adaptive-routing-guide.md` - Adaptive routing guide
- `anthropic-claude-code-tools-analysis.md` - Anthropic Claude Code tools analysis
- `anthropic-messages-api.md` - Anthropic Messages API
- `automation-context.json` - Automation context
- `oauth-session-pool/` - OAuth session pool
- `github-research/` - GitHub research

### 🌐 ti-router - Router (Vietnamese)

Router architecture, patterns (Vietnamese documentation).

- **[INDEX.md](ti-router/INDEX.md)** - Router knowledge base (Vietnamese)
- `kiến-trúc-hệ-thống.md` - System architecture
- `providers/` - Providers

### 📦 sources - Sources

Documentation sources and references.

- **[INDEX.md](sources/INDEX.md)** - Sources knowledge base
- `README.md` - Sources structure

### 💾 storage - Storage

Large files and binaries.

- **[INDEX.md](storage/INDEX.md)** - Storage knowledge base
- Large zip files moved from Router/ and browser/

### 🧠 tibrain - Tibrain

Tibrain documentation.

- **[INDEX.md](tibrain/INDEX.md)** - Tibrain knowledge base

### 🛠️ tool - Tool

Tool documentation.

- **[INDEX.md](tool/INDEX.md)** - Tool knowledge base
- `tool-search-implementation-plan.md` - Tool search implementation
- `tool-search-learnings.md` - Tool search learnings

---

## External Repositories (07_Repositories/)

Source code repositories moved from 03_Knowledge/:

- **browser-docs/** - Browser automation source code and documentation (moved from `browser/`)
- **skills-docs/** - Skills repositories (moved from `skills/`)

---

## Rule: 1 Nơi Duy Nhất

- **Không** tạo doc mới ngoài folder này
- Draft → `06_Learning/`. Chuẩn hóa xong → appropriate category folder
- Cũ ở `Z:\07_DOCS\` đã redirect về đây

---

## Recent Changes (2026-05-05)

### Deep Cleanup
- Removed empty folders: `donut-browser/`, `cheatsheets/`, `decisions/`, `runbooks/`, `spectre/`
- Moved `browser/` to `07_Repositories/browser-docs/` (source code repos)
- Moved `skills/` to `07_Repositories/skills-docs/` (source code repos)
- Renamed `cli/00_INDEX.md` → `cli/INDEX.md` (consistent naming)
- Created INDEX.md for all folders (26 folders now have INDEX.md)
- Updated MANIFEST.md with latest structure

### Previous Reorganization
- Created `00_META/` for meta files
- Created `frontend/` for frontend patterns
- Created `prompt-engineering/` for prompt engineering
- Created `provider/` for provider management
- Renamed `Agent/` → `ticlaw/` for clarity
- Merged `lessons-learned/` into `lessons/`
- Moved uncategorized files from root to appropriate folders
- Created `devin/`, `notion/`, `research/` for better organization
- Cleaned `cli/` folder to contain only CLI-specific documentation
- Added `cli/sources/` for CLI source analysis

---

## Maintenance

When adding new documentation:
1. Choose the appropriate folder based on category
2. Add an entry to this INDEX.md
3. Use clear, descriptive filenames
4. Update this INDEX.md's "Last Updated" date
5. Keep only relevant documentation in this folder
6. Create INDEX.md for new folders
