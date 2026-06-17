# CLI Folder Cleanup Plan

> **Date**: 2026-05-05
> **Purpose**: Keep only CLI-specific files in cli/ folder, move rest to appropriate folders

---

## Classification

### ✅ CLI-SPECIFIC (KEEP in cli/)

**01_CORE/ - CLI Core Architecture**
- ARCHITECTURE.md
- CLI_GUIDE.md
- CODEBASE_MIGRATION_ASSESSMENT.md
- PLUGIN_ARCHITECTURE.md
- QUICK_START.md

**03_CLI_FEATURES/ - CLI Features**
- CLI_COMMAND_VALIDATION.md
- CLI_ERROR_HANDLING.md
- CLI_EXAMPLES.md
- CLI_FLAG_MANAGEMENT.md
- CLI_TOOL_SYSTEM.md

**04_INTEGRATION/ - CLI Integration (partial)**
- CLI_INTEGRATION_ROADMAP.md
- CLI_MCP_OAUTH.md

**07_DEPLOYMENT/ - CLI Deployment**
- DEPLOYMENT_MAINTENANCE_STRATEGY.md

### ❌ NOT CLI-SPECIFIC (MOVE)

**02_AGENT_FRAMEWORK/ → agents/**
- CHATGPT_CLI_PLUGIN_SEPARATION_ANALYSIS.md
- cli-agent-status.md
- SUB_AGENT_GUIDE.md
- TI_CREW/ (5 files)

**04_INTEGRATION/ROUTER_MIDDLEWARE_PLUGIN_ARCHITECTURE.md → Router/**

**05_NOTION_SYNC/ → notion/** (create folder)
- CLI_NOTION_SYNC_APPROVAL.md
- CLI_NOTION_SYNC_ARCHITECTURE.md
- CLI_NOTION_SYNC_EVALUATION.md
- CLI_NOTION_SYNC_INTEGRATION_POINTS.md
- CLI_NOTION_SYNC_STRATEGY.md
- NOTION_DATABASE_CONFIGURATION.md

**06_ROUTER/ → Router/**
- ROUTER_RULES_ENFORCEMENT_ARCHITECTURE.md

**08_CLI_UPGRADE_RESEARCH/ → research/**
- AGENTS_VIETNAMESE.md
- GO_SDK_PATTERNS.md
- Implementation.md
- implementation-plan.md
- key-concepts.md
- video-notes.md
- docs/ (26 files)

**09_DEVIN_PORTING/ → devin/** (create folder)
- 01_API_KIEN_TRUC.md
- 02_GO_PORT_PLAN.md
- 03_OFFICIAL_CLI.md
- 04_PHAN_TICH_UU_DIEM_PORT.md
- DEVIN_API_KIEN_TRUC.md

**10_LESSONS_LEARNED/ → lessons/**
- JUNIE_CLI_LESSONS.md

**11_ARCHIVE/ → archive/**
- FOLDER_REORGANIZATION_PROPOSAL.md

---

## Final CLI Structure (After Cleanup)

```
cli/
├── 00_INDEX.md
├── 01_CORE/
│   ├── ARCHITECTURE.md
│   ├── CLI_GUIDE.md
│   ├── CODEBASE_MIGRATION_ASSESSMENT.md
│   ├── PLUGIN_ARCHITECTURE.md
│   └── QUICK_START.md
├── 02_FEATURES/              (renamed from 03_CLI_FEATURES)
│   ├── CLI_COMMAND_VALIDATION.md
│   ├── CLI_ERROR_HANDLING.md
│   ├── CLI_EXAMPLES.md
│   ├── CLI_FLAG_MANAGEMENT.md
│   └── CLI_TOOL_SYSTEM.md
├── 03_INTEGRATION/           (renamed from 04_INTEGRATION, cleaned)
│   ├── CLI_INTEGRATION_ROADMAP.md
│   └── CLI_MCP_OAUTH.md
└── 04_DEPLOYMENT/            (renamed from 07_DEPLOYMENT)
    └── DEPLOYMENT_MAINTENANCE_STRATEGY.md
```

---

## Destination Folders

- **agents/** - Already exists
- **Router/** - Already exists
- **research/** - Already exists
- **lessons/** - Already exists
- **archive/** - Already exists
- **notion/** - Need to create
- **devin/** - Need to create
