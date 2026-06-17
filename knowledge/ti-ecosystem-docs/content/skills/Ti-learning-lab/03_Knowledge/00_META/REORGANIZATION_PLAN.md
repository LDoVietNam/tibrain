# 03_Knowledge Reorganization Plan

> **Date**: 2026-05-05
> **Purpose**: Reorganize 03_Knowledge folder for better structure and navigation

---

## Current Issues

1. **Duplicate folders**: `Agent/` vs `agents/`, `lessons/` vs `lessons-learned/`, `Router/` vs `ti-router/`
2. **Uncategorized files at root**: 18 .md files scattered at root level
3. **Mixed naming conventions**: Some use kebab-case, some use underscores
4. **Inconsistent structure**: Some folders have subfolders, some don't
5. **Outdated INDEX.md**: Doesn't reflect current structure

---

## Analysis

### Duplicate Folders

| Folder 1 | Folder 2 | Content Difference | Action |
|----------|----------|-------------------|--------|
| `Agent/` | `agents/` | Agent/ has ti-claw docs, agents/ has framework docs | Keep both, rename Agent/ → ticlaw/ |
| `lessons/` | `lessons-learned/` | lessons/ has files, lessons-learned/ has subfolders | Merge into lessons/ |
| `Router/` | `ti-router/` | Router/ has English docs, ti-router/ has Vietnamese docs | Keep both, Router/ for English, ti-router/ for Vietnamese |

### Uncategorized Files at Root

| File | Category | Destination |
|------|----------|-------------|
| `01_AI_ROUTER_BEST_PRACTICES.md` | Router | Router/ |
| `07_DOCS-README.md` | Docs | docs/ |
| `cache-agent-patterns-go.md` | Patterns | patterns/ |
| `frontend-ai-components.md` | Frontend | patterns/ (or create frontend/) |
| `go-cli-ai-patterns.md` | Go patterns | patterns/ |
| `html-ui-alternative.md` | UI | patterns/ (or create ui/) |
| `INDEX.md` | Index | Keep at root |
| `MANIFEST.md` | Meta | Keep at root |
| `metrics-cli-agent-patterns-go.md` | Metrics | metrics/ |
| `oauth-authentication-patterns-go.md` | OAuth | agents/auth/ |
| `prompt-engineering-repos.md` | Prompt Engineering | patterns/ (or create prompt-engineering/) |
| `provider-management.md` | Provider | Router/ |
| `README.md` | Meta | Keep at root |
| `registry.json` | Meta | Keep at root |
| `RESEARCH_ANALYSIS.md` | Research | research/ |
| `RESEARCH_INDEX.md` | Research | research/ |
| `TI_CLI_BEST_SOURCE_INTEGRATION.md` | CLI | cli/ |
| `TI_CLI_CLAUDE_FORMAT.md` | CLI | cli/ |

### Existing Folders (Keep)

- `api/` - API integration
- `archive/` - Archived files
- `auto-reg-tools/` - Auto registration tools
- `browser/` - Browser automation
- `browser-automation/` - Browser automation (duplicate?)
- `cache/` - Cache
- `cheatsheets/` - Cheatsheets
- `cli/` - CLI (already cleaned)
- `computer-vision/` - Computer vision
- `decisions/` - Decisions
- `devin/` - Devin (newly created)
- `docs/` - Documentation
- `donut-browser/` - Donut browser
- `hot-reload/` - Hot reload
- `metrics/` - Metrics
- `notion/` - Notion (newly created)
- `patterns/` - Patterns
- `protocols/` - Protocols
- `research/` - Research (newly created)
- `Router/` - Router
- `runbooks/` - Runbooks
- `skills/` - Skills
- `spectre/` - Spectre
- `tibrain/` - Tibrain
- `tool/` - Tool
- `分析sharedchatfun的cook/` - Chinese folder (keep)

---

## Proposed Structure

```
03_Knowledge/
├── INDEX.md                    # Main index
├── README.md                   # Overview
├── MANIFEST.md                 # Manifest
├── registry.json               # Registry
│
├── 00_META/                    # Meta files
│   ├── REORGANIZATION_PLAN.md
│   └── (future meta files)
│
├── agents/                     # Agent framework and orchestration
│   ├── 02_AGENT_FRAMEWORK/
│   ├── AGENTS.md.reference
│   ├── auth/
│   └── orchestration/
│
├── ticlaw/                     # Ti-claw specific (renamed from Agent/)
│   ├── ti-claw-a0-decision.md
│   ├── ti-claw-a1-quick-wins.md
│   └── ti-claw-a2-plugin-setup.md
│
├── api/                        # API integration
│
├── archive/                    # Archived files
│
├── auto-reg-tools/             # Auto registration tools
│
├── browser/                    # Browser automation
│
├── browser-automation/          # (merge into browser/ or keep separate)
│
├── cache/                      # Cache
│
├── cheatsheets/                # Cheatsheets
│
├── cli/                        # CLI (already cleaned)
│
├── computer-vision/            # Computer vision
│
├── decisions/                  # Decisions
│
├── devin/                      # Devin porting
│
├── docs/                       # Documentation
│   └── 07_DOCS-README.md
│
├── donut-browser/              # Donut browser
│
├── frontend/                   # Frontend patterns (new)
│   └── frontend-ai-components.md
│
├── hot-reload/                 # Hot reload
│
├── lessons/                    # Lessons learned (merged)
│   ├── CLI_PROXY_API_LESSONS.md
│   ├── JUNIE_CLI_LESSONS.md
│   ├── TI_INTERNAL_LESSONS.md
│   ├── UI_LESSONS.md
│   ├── cli/ (from lessons-learned/cli/)
│   ├── mcphub/ (from lessons-learned/mcphub/)
│   ├── router/ (from lessons-learned/router/)
│   └── tibrain/ (from lessons-learned/tibrain/)
│
├── metrics/                    # Metrics
│   └── metrics-cli-agent-patterns-go.md
│
├── notion/                     # Notion integration
│
├── patterns/                   # Patterns
│   ├── cache-agent-patterns-go.md
│   ├── go-cli-ai-patterns.md
│   ├── html-ui-alternative.md
│   ├── oauth-authentication-patterns-go.md
│   └── prompt-engineering-repos.md
│
├── protocols/                  # Protocols
│
├── prompt-engineering/         # Prompt engineering (new)
│   └── prompt-engineering-repos.md
│
├── provider/                   # Provider management (new)
│   └── provider-management.md
│
├── research/                   # Research
│   ├── RESEARCH_ANALYSIS.md
│   └── RESEARCH_INDEX.md
│
├── Router/                     # Router (English)
│   ├── 01_AI_ROUTER_BEST_PRACTICES.md
│   ├── provider-management.md
│   └── (existing Router/ content)
│
├── ti-router/                  # Router (Vietnamese)
│   └── (existing ti-router/ content)
│
├── runbooks/                   # Runbooks
│
├── skills/                     # Skills
│
├── spectre/                    # Spectre
│
├── tibrain/                    # Tibrain
│
├── tool/                       # Tool
│
└── 分析sharedchatfun的cook/     # Chinese folder
```

---

## Action Items

### Phase 1: Create New Folders
- [ ] Create `00_META/`
- [ ] Create `frontend/`
- [ ] Create `prompt-engineering/`
- [ ] Create `provider/`

### Phase 2: Rename Folders
- [ ] Rename `Agent/` → `ticlaw/`
- [ ] Delete `lessons-learned/` after merging

### Phase 3: Move Files from Root
- [ ] Move `01_AI_ROUTER_BEST_PRACTICES.md` → `Router/`
- [ ] Move `07_DOCS-README.md` → `docs/`
- [ ] Move `cache-agent-patterns-go.md` → `patterns/`
- [ ] Move `frontend-ai-components.md` → `frontend/`
- [ ] Move `go-cli-ai-patterns.md` → `patterns/`
- [ ] Move `html-ui-alternative.md` → `patterns/`
- [ ] Move `metrics-cli-agent-patterns-go.md` → `metrics/`
- [ ] Move `oauth-authentication-patterns-go.md` → `patterns/` (or agents/auth/)
- [ ] Move `prompt-engineering-repos.md` → `prompt-engineering/`
- [ ] Move `provider-management.md` → `provider/` (or Router/)
- [ ] Move `RESEARCH_ANALYSIS.md` → `research/`
- [ ] Move `RESEARCH_INDEX.md` → `research/`
- [ ] Move `TI_CLI_BEST_SOURCE_INTEGRATION.md` → `cli/`
- [ ] Move `TI_CLI_CLAUDE_FORMAT.md` → `cli/`
- [ ] Move `REORGANIZATION_PLAN.md` → `00_META/`

### Phase 4: Merge Duplicate Folders
- [ ] Merge `lessons-learned/` subfolders into `lessons/`
- [ ] Delete `lessons-learned/`

### Phase 5: Update INDEX.md
- [ ] Update INDEX.md with new structure
- [ ] Add navigation links
- [ ] Update "Last Updated" date

---

## Decisions Needed

1. **browser vs browser-automation**: Should these be merged or kept separate?
2. **oauth-authentication-patterns-go.md**: Should go to `patterns/` or `agents/auth/`?
3. **provider-management.md**: Should go to `provider/` or `Router/`?
