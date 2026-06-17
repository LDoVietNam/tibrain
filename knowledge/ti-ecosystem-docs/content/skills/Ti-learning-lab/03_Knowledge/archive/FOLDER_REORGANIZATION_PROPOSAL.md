# CLI Knowledge Folder Reorganization Proposal

> **Date**: 2026-05-05
> **Purpose**: Reorganize `Z:\10_WORKPLACE\Ti\Ti-learning-lab\03_Knowledge\cli` for better navigation and maintainability

---

## Current Structure Analysis

### Current Issues

1. **Flat structure with mixed topics** - All files at root level, hard to find related documents
2. **Inconsistent naming** - Some use `CLI_` prefix, some don't
3. **Scattered subdirectories** - `CLI UPGRADE/` and `devin/` contain related content
4. **No clear categorization** - Difficult to distinguish between architecture, features, integration, etc.

### Current File Count

- Root level: 28 .md files
- `CLI UPGRADE/`: 11 files + `docs/` with 24 files
- `devin/`: 5 files
- **Total**: 68 files

---

## Proposed Structure

### New Folder Organization

```
cli/
├── 00_INDEX.md                    # Index/README for this folder
├── 01_CORE/                       # Core CLI architecture
│   ├── ARCHITECTURE.md
│   ├── PLUGIN_ARCHITECTURE.md
│   ├── QUICK_START.md
│   ├── CLI_GUIDE.md
│   └── CODEBASE_MIGRATION_ASSESSMENT.md
├── 02_AGENT_FRAMEWORK/            # Agent-related documentation
│   ├── TI_AGENT_FRAMEWORK/
│   │   ├── TI_AGENT_FRAMEWORK_ARCHITECTURE.md
│   │   ├── TI_AGENT_FRAMEWORK_FEASIBILITY.md
│   │   └── TI_AGENT_FRAMEWORK_IMPLEMENTATION_PLAN.md
│   ├── TI_CREW/
│   │   ├── TI_CREW_ARCHITECTURE.md
│   │   ├── TI_CREW_FEASIBILITY.md
│   │   ├── TI_CREW_GOCLAW_INTEGRATION.md
│   │   ├── TI_CREW_IMPLEMENTATION_PLAN.md
│   │   └── TI_CREW_PLUGIN_ANALYSIS.md
│   ├── CHATGPT_CLI_PLUGIN_SEPARATION_ANALYSIS.md
│   ├── SUB_AGENT_GUIDE.md
│   └── cli-agent-status.md
├── 03_CLI_FEATURES/               # CLI feature documentation
│   ├── CLI_TOOL_SYSTEM.md
│   ├── CLI_COMMAND_VALIDATION.md
│   ├── CLI_ERROR_HANDLING.md
│   ├── CLI_FLAG_MANAGEMENT.md
│   └── CLI_EXAMPLES.md
├── 04_INTEGRATION/                # Integration with other systems
│   ├── CLI_INTEGRATION_ROADMAP.md
│   ├── CLI_MCP_OAUTH.md
│   └── ROUTER_MIDDLEWARE_PLUGIN_ARCHITECTURE.md
├── 05_NOTION_SYNC/                # Notion integration
│   ├── CLI_NOTION_SYNC_STRATEGY.md
│   ├── CLI_NOTION_SYNC_ARCHITECTURE.md
│   ├── CLI_NOTION_SYNC_INTEGRATION_POINTS.md
│   ├── CLI_NOTION_SYNC_EVALUATION.md
│   ├── CLI_NOTION_SYNC_APPROVAL.md
│   └── NOTION_DATABASE_CONFIGURATION.md
├── 06_ROUTER/                     # Router-related
│   ├── ROUTER_MIDDLEWARE_PLUGIN_ARCHITECTURE.md
│   └── ROUTER_RULES_ENFORCEMENT_ARCHITECTURE.md
├── 07_DEPLOYMENT/                # Deployment and maintenance
│   ├── DEPLOYMENT_MAINTENANCE_STRATEGY.md
│   └── (future deployment docs)
├── 08_CLI_UPGRADE_RESEARCH/       # CLI upgrade research
│   ├── Implementation.md
│   ├── implementation-plan.md
│   ├── key-concepts.md
│   ├── video-notes.md
│   ├── GO_SDK_PATTERNS.md
│   ├── AGENTS.md.reference
│   ├── AGENTS_VIETNAMESE.md
│   ├── .antigravity_rules.md
│   ├── .multi_tool_rules.md
│   ├── .opencode_rules.md
│   └── docs/
│       ├── AUTO-LOOP-EXTRACTION.md
│       ├── AUTO-LOOP-FINAL-STATUS.md
│       ├── CLAUDE_CODE_COMPLETE_EXTRACTION.md
│       ├── CLAUDE_CODE_PATTERNS.md
│       ├── CLAUDE_CODE_RUST_ANALYSIS.md
│       ├── CLI-FEATURE-MATRIX.md
│       ├── CLIPROXY_STATIC_FILES_ANALYSIS.md
│       ├── CodexCapabilityImprovement.md
│       ├── CONTINUOUS-LEARNING-LOOP.md
│       ├── DASHBOARD_BACKEND_INTEGRATION_ASSESSMENT.md
│       ├── EXTRACTION-FINAL-SUMMARY.md
│       ├── MODERN_DASHBOARD_UI_RESEARCH_2026.md
│       ├── My-Router_BACKEND_UPGRADE_PLAN.md
│       ├── NODE_MODULES-AGENT-PLAN.md
│       ├── NODE_MODULES.md
│       ├── OmniRoute-OAuth-Analysis.md
│       ├── OPENCODE_AGENT_LOOP.md
│       ├── OPENCODE_ANALYSIS_README.md
│       ├── OPENCODE_ARCHITECTURE_OVERVIEW.md
│       ├── OPENCODE_CONFIG_SYSTEM.md
│       ├── OPENCODE_PROVIDER_SYSTEM.md
│       ├── OpenSource-CLI-Analysis.md
│       ├── PATTERNS.md
│       ├── TI-ARCHITECTURE.md
│       ├── TI-CLI-BLUEPRINT.md
│       └── TI-RESEARCH-SOURCES.md
├── 09_DEVIN_PORTING/              # Devin-specific porting docs
│   ├── 01_API_KIEN_TRUC.md
│   ├── 02_GO_PORT_PLAN.md
│   ├── 03_OFFICIAL_CLI.md
│   ├── 04_PHAN_TICH_UU_DIEM_PORT.md
│   └── DEVIN_API_KIEN_TRUC.md
├── 10_LESSONS_LEARNED/            # Lessons learned from other CLIs
│   ├── JUNIE_CLI_LESSONS.md
│   └── (future lessons)
└── 11_ARCHIVE/                   # Deprecated or old documentation
    └── (move deprecated files here)
```

---

## File Mapping

### 01_CORE/
| From | To |
|------|-----|
| `ARCHITECTURE.md` | `01_CORE/ARCHITECTURE.md` |
| `PLUGIN_ARCHITECTURE.md` | `01_CORE/PLUGIN_ARCHITECTURE.md` |
| `QUICK_START.md` | `01_CORE/QUICK_START.md` |
| `CLI_GUIDE.md` | `01_CORE/CLI_GUIDE.md` |
| `CODEBASE_MIGRATION_ASSESSMENT.md` | `01_CORE/CODEBASE_MIGRATION_ASSESSMENT.md` |

### 02_AGENT_FRAMEWORK/
| From | To |
|------|-----|
| `TI_AGENT_FRAMEWORK_ARCHITECTURE.md` | `02_AGENT_FRAMEWORK/TI_AGENT_FRAMEWORK/TI_AGENT_FRAMEWORK_ARCHITECTURE.md` |
| `TI_AGENT_FRAMEWORK_FEASIBILITY.md` | `02_AGENT_FRAMEWORK/TI_AGENT_FRAMEWORK/TI_AGENT_FRAMEWORK_FEASIBILITY.md` |
| `TI_AGENT_FRAMEWORK_IMPLEMENTATION_PLAN.md` | `02_AGENT_FRAMEWORK/TI_AGENT_FRAMEWORK/TI_AGENT_FRAMEWORK_IMPLEMENTATION_PLAN.md` |
| `TI_CREW_ARCHITECTURE.md` | `02_AGENT_FRAMEWORK/TI_CREW/TI_CREW_ARCHITECTURE.md` |
| `TI_CREW_FEASIBILITY.md` | `02_AGENT_FRAMEWORK/TI_CREW/TI_CREW_FEASIBILITY.md` |
| `TI_CREW_GOCLAW_INTEGRATION.md` | `02_AGENT_FRAMEWORK/TI_CREW/TI_CREW_GOCLAW_INTEGRATION.md` |
| `TI_CREW_IMPLEMENTATION_PLAN.md` | `02_AGENT_FRAMEWORK/TI_CREW/TI_CREW_IMPLEMENTATION_PLAN.md` |
| `TI_CREW_PLUGIN_ANALYSIS.md` | `02_AGENT_FRAMEWORK/TI_CREW/TI_CREW_PLUGIN_ANALYSIS.md` |
| `CHATGPT_CLI_PLUGIN_SEPARATION_ANALYSIS.md` | `02_AGENT_FRAMEWORK/CHATGPT_CLI_PLUGIN_SEPARATION_ANALYSIS.md` |
| `SUB_AGENT_GUIDE.md` | `02_AGENT_FRAMEWORK/SUB_AGENT_GUIDE.md` |
| `cli-agent-status.md` | `02_AGENT_FRAMEWORK/cli-agent-status.md` |

### 03_CLI_FEATURES/
| From | To |
|------|-----|
| `CLI_TOOL_SYSTEM.md` | `03_CLI_FEATURES/CLI_TOOL_SYSTEM.md` |
| `CLI_COMMAND_VALIDATION.md` | `03_CLI_FEATURES/CLI_COMMAND_VALIDATION.md` |
| `CLI_ERROR_HANDLING.md` | `03_CLI_FEATURES/CLI_ERROR_HANDLING.md` |
| `CLI_FLAG_MANAGEMENT.md` | `03_CLI_FEATURES/CLI_FLAG_MANAGEMENT.md` |
| `CLI_EXAMPLES.md` | `03_CLI_FEATURES/CLI_EXAMPLES.md` |

### 04_INTEGRATION/
| From | To |
|------|-----|
| `CLI_INTEGRATION_ROADMAP.md` | `04_INTEGRATION/CLI_INTEGRATION_ROADMAP.md` |
| `CLI_MCP_OAUTH.md` | `04_INTEGRATION/CLI_MCP_OAUTH.md` |
| `ROUTER_MIDDLEWARE_PLUGIN_ARCHITECTURE.md` | `04_INTEGRATION/ROUTER_MIDDLEWARE_PLUGIN_ARCHITECTURE.md` |

### 05_NOTION_SYNC/
| From | To |
|------|-----|
| `CLI_NOTION_SYNC_STRATEGY.md` | `05_NOTION_SYNC/CLI_NOTION_SYNC_STRATEGY.md` |
| `CLI_NOTION_SYNC_ARCHITECTURE.md` | `05_NOTION_SYNC/CLI_NOTION_SYNC_ARCHITECTURE.md` |
| `CLI_NOTION_SYNC_INTEGRATION_POINTS.md` | `05_NOTION_SYNC/CLI_NOTION_SYNC_INTEGRATION_POINTS.md` |
| `CLI_NOTION_SYNC_EVALUATION.md` | `05_NOTION_SYNC/CLI_NOTION_SYNC_EVALUATION.md` |
| `CLI_NOTION_SYNC_APPROVAL.md` | `05_NOTION_SYNC/CLI_NOTION_SYNC_APPROVAL.md` |
| `NOTION_DATABASE_CONFIGURATION.md` | `05_NOTION_SYNC/NOTION_DATABASE_CONFIGURATION.md` |

### 06_ROUTER/
| From | To |
|------|-----|
| `ROUTER_MIDDLEWARE_PLUGIN_ARCHITECTURE.md` | `06_ROUTER/ROUTER_MIDDLEWARE_PLUGIN_ARCHITECTURE.md` |
| `ROUTER_RULES_ENFORCEMENT_ARCHITECTURE.md` | `06_ROUTER/ROUTER_RULES_ENFORCEMENT_ARCHITECTURE.md` |

### 07_DEPLOYMENT/
| From | To |
|------|-----|
| `DEPLOYMENT_MAINTENANCE_STRATEGY.md` | `07_DEPLOYMENT/DEPLOYMENT_MAINTENANCE_STRATEGY.md` |

### 08_CLI_UPGRADE_RESEARCH/
| From | To |
|------|-----|
| `CLI UPGRADE/Implementation.md` | `08_CLI_UPGRADE_RESEARCH/Implementation.md` |
| `CLI UPGRADE/implementation-plan.md` | `08_CLI_UPGRADE_RESEARCH/implementation-plan.md` |
| `CLI UPGRADE/key-concepts.md` | `08_CLI_UPGRADE_RESEARCH/key-concepts.md` |
| `CLI UPGRADE/video-notes.md` | `08_CLI_UPGRADE_RESEARCH/video-notes.md` |
| `CLI UPGRADE/GO_SDK_PATTERNS.md` | `08_CLI_UPGRADE_RESEARCH/GO_SDK_PATTERNS.md` |
| `CLI UPGRADE/AGENTS.md.reference` | `08_CLI_UPGRADE_RESEARCH/AGENTS.md.reference` |
| `CLI UPGRADE/AGENTS_VIETNAMESE.md` | `08_CLI_UPGRADE_RESEARCH/AGENTS_VIETNAMESE.md` |
| `CLI UPGRADE/.antigravity_rules.md` | `08_CLI_UPGRADE_RESEARCH/.antigravity_rules.md` |
| `CLI UPGRADE/.multi_tool_rules.md` | `08_CLI_UPGRADE_RESEARCH/.multi_tool_rules.md` |
| `CLI UPGRADE/.opencode_rules.md` | `08_CLI_UPGRADE_RESEARCH/.opencode_rules.md` |
| `CLI UPGRADE/docs/*` | `08_CLI_UPGRADE_RESEARCH/docs/*` |

### 09_DEVIN_PORTING/
| From | To |
|------|-----|
| `devin/01_API_KIEN_TRUC.md` | `09_DEVIN_PORTING/01_API_KIEN_TRUC.md` |
| `devin/02_GO_PORT_PLAN.md` | `09_DEVIN_PORTING/02_GO_PORT_PLAN.md` |
| `devin/03_OFFICIAL_CLI.md` | `09_DEVIN_PORTING/03_OFFICIAL_CLI.md` |
| `devin/04_PHAN_TICH_UU_DIEM_PORT.md` | `09_DEVIN_PORTING/04_PHAN_TICH_UU_DIEM_PORT.md` |
| `devin/DEVIN_API_KIEN_TRUC.md` | `09_DEVIN_PORTING/DEVIN_API_KIEN_TRUC.md` |

### 10_LESSONS_LEARNED/
| From | To |
|------|-----|
| `JUNIE_CLI_LESSONS.md` | `10_LESSONS_LEARNED/JUNIE_CLI_LESSONS.md` |

---

## Benefits

1. **Better Navigation** - Files grouped by topic/category
2. **Clearer Hierarchy** - Numbered folders indicate logical order
3. **Easier Maintenance** - Related files are together
4. **Scalable Structure** - Easy to add new categories
5. **Consistent Naming** - Standardized folder and file names

---

## Implementation Steps

1. Create new folder structure
2. Move files according to mapping
3. Create 00_INDEX.md as navigation guide
4. Update any cross-references in files
5. Delete old empty folders

---

## Notes

- `ROUTER_MIDDLEWARE_PLUGIN_ARCHITECTURE.md` appears in both 04_INTEGRATION and 06_ROUTER - needs decision on primary location
- Consider creating symlinks if a file needs to be in multiple locations
- Update AGENTS.md or other files that reference these paths
