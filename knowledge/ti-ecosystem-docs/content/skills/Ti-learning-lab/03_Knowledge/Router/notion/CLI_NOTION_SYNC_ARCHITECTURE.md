---
tags: ["tibrain", "skill", "documentation", "provider-notion", "router"]
scopes: ["tibrain"]
last_updated: 2026-05-22
---
# Architecture - CLI Documentation Sync với Notion

> **Mục đích**: Định nghĩa architecture cho sync CLI documentation với Notion knowledge base
> **Ngày tạo**: 2026-05-05
> **Trạng thái**: Architecture Design
> **Version**: 1.0

---

## 1. Tổng quan

### Vấn đề
- CLI documentation mới được tạo trong `apps/cli/docs/` (7 files)
- Chưa có workflow sync đến Notion knowledge base
- Không có category cho CLI Architecture trong Notion
- Không integration với existing Notion sync systems

### Mục tiêu
- Sync CLI documentation đến Notion knowledge base
- Tạo category "CLI Architecture" trong Notion
- Integration với existing Notion sync systems (knowledge-sync plugin/agent)
- Maintain consistency với existing knowledge sync workflow

---

## 2. Existing Architecture

### 2.1 Knowledge Sync Plugin (Basic)

**Location**: `apps/automation/knowledge-sync.go`

**Architecture**:
```
KnowledgeSyncPlugin
├── Config
│   ├── KnowledgePath (hardcoded: Router)
│   ├── SyncLogPath
│   ├── NotionAPIKey
│   ├── NotionDatabaseID
│   └── NotionVersion
├── Features
│   ├── Scan markdown files
│   ├── SHA256 hash tracking
│   ├── Basic sync to Notion API
│   ├── Sync log tracking
│   └── BD tool logging
└── Properties
    ├── Tên
    ├── Type (Changelog)
    ├── Status (Done)
    └── Priority (Medium/High)
```

**Limitations**:
- Không có AI analysis
- Không có smart categorization
- Không có relationship detection
- Hardcoded KnowledgePath = `03_Knowledge/Router`
- Basic summary extraction

### 2.2 Knowledge Sync AI Agent (Advanced)

**Location**: `apps/automation/knowledge-sync-ai.go`

**Architecture**:
```
KnowledgeSyncAI
├── Config
│   ├── KnowledgePath (hardcoded: Router)
│   ├── SyncLogPath (AI version)
│   ├── NotionAPIKey
│   ├── NotionDatabaseID
│   └── NotionVersion
├── LLM Client
│   ├── Provider (openai, anthropic, router)
│   ├── API Key
│   └── Base URL
├── KnowledgeMemory
│   ├── Syncs (map[string]SyncLogEntry)
│   ├── Patterns (map[string][]string)
│   ├── Relationships (map[string][]string)
│   └── Priorities (map[string]string)
├── Tool Registry
│   ├── read_file
│   ├── analyze_content
│   ├── categorize
│   ├── generate_summary
│   ├── detect_relationships
│   └── sync_to_notion
└── Features
    ├── AI-powered content analysis
    ├── Auto-categorization
    ├── Priority determination
    ├── Summary generation
    ├── Relationship detection
    └── Learning system
```

**Status**: Incomplete (stub functions)

**Limitations**:
- Hardcoded KnowledgePath = `03_Knowledge/Router`
- Many functions are stubs (NotionClient, scanKnowledgeFiles, isSynced)
- KnowledgeMemory Save/Load not implemented
- NotionClient CreatePage is stub

### 2.3 Manual PowerShell Script

**Location**: `Ti-learning-lab/scripts/sync-knowledge-to-notion.ps1`

**Architecture**:
```
PowerShell Script
├── Config
│   ├── NOTION_TOKEN
│   ├── DB_ID
│   ├── KNOWLEDGE_DIR (hardcoded: Router)
│   ├── NOTION_API_BASE
│   └── NOTION_VERSION
├── Features
│   ├── Manual Notion API calls
│   ├── Hardcoded files list
│   ├── Basic properties
│   └── Content truncation (2000 chars)
└── Properties
    ├── Tên
    ├── Type (Changelog)
    ├── Status (Done)
    ├── Priority (High)
    └── Version (1.0.0)
```

**Limitations**:
- Manual, hardcoded
- No AI analysis
- No automation
- Content truncation
- No learning system

---

## 3. CLI Documentation Structure

### 3.1 Current Location

```
apps/cli/docs/
├── CLI_GUIDE.md                          # Updated với architecture overview
├── CLI_FLAG_MANAGEMENT.md                # Flag management documentation
├── CLI_ERROR_HANDLING.md                 # Error handling documentation
├── CLI_COMMAND_VALIDATION.md             # Command validation documentation
├── CLI_TOOL_SYSTEM.md                    # Tool system documentation
├── CLI_MCP_OAUTH.md                      # MCP OAuth documentation
└── CLI_EXAMPLES.md                       # Comprehensive examples
```

### 3.2 Proposed Location (Ti-learning-lab)

```
Ti-learning-lab/
├── 03_Knowledge/
│   ├── Router/                            # Existing Router knowledge
│   └── CLI/                               # NEW: CLI knowledge
│       ├── CLI_NOTION_SYNC_STRATEGY.md    # Sync strategy (đã tạo)
│       ├── CLI_GUIDE.md
│       ├── CLI_FLAG_MANAGEMENT.md
│       ├── CLI_ERROR_HANDLING.md
│       ├── CLI_COMMAND_VALIDATION.md
│       ├── CLI_TOOL_SYSTEM.md
│       ├── CLI_MCP_OAUTH.md
│       └── CLI_EXAMPLES.md
```

---

## 4. Integration Options

### Option A: Extend Existing Plugin (RECOMMENDED) ✅

**Approach**: Modify existing knowledge-sync plugin để support CLI path

**Changes Required**:
1. Modify `knowledge-sync.go`:
   - Remove hardcoded KnowledgePath
   - Add `--path` flag support (đã có, nhưng cần test)
   - Add CLI-specific properties
   - Add CLI-specific categorization logic

2. Modify `knowledge-sync-ai.go`:
   - Complete stub functions
   - Add CLI-specific AI prompts
   - Add CLI category mapping
   - Implement KnowledgeMemory Save/Load
   - Implement NotionClient CreatePage

**Pros**:
- Reuse existing infrastructure
- Consistent with Router knowledge sync
- Single source of truth
- Less code duplication

**Cons**:
- Need to complete AI agent stubs
- Risk of breaking existing Router sync

### Option B: Separate CLI Sync Plugin

**Approach**: Create dedicated CLI sync plugin

**Changes Required**:
1. Create `cli-sync.go` plugin
2. Copy/adapt logic from knowledge-sync.go
3. Add CLI-specific features
4. Separate sync log: `sync-log-cli.json`

**Pros**:
- Isolated from Router sync
- Can evolve independently
- No risk to existing sync

**Cons**:
- Code duplication
- Multiple sync logs to maintain
- Inconsistent with existing workflow

### Option C: Manual Script (Fallback)

**Approach**: Create PowerShell script for CLI docs sync

**Changes Required**:
1. Create `sync-cli-docs-to-notion.ps1`
2. Hardcode CLI docs list
3. Manual Notion API calls
4. Manual categorization

**Pros**:
- Simple, immediate
- No dependencies

**Cons**:
- No AI analysis
- Manual maintenance
- Not consistent with workflow

---

## 5. Integration Points

### 5.1 CLI → Ti-learning-lab

**Point 1**: Move CLI docs
- **From**: `apps/cli/docs/`
- **To**: `Ti-learning-lab/03_Knowledge/CLI/`
- **Impact**: Update all references trong codebase

**Point 2**: Update references
- `AGENTS.md` (nếu có CLI docs references)
- `docs/CLI_GUIDE.md` (update path references)
- Any internal links to CLI docs

### 5.2 Ti-learning-lab → Notion

**Point 1**: AI Sync Agent
- **Tool**: `knowledge-sync-ai.go`
- **Command**: `go run . knowledge-sync-ai --path="CLI/"`
- **Config**: NOTION_API_KEY, NOTION_DATABASE_ID

**Point 2**: Basic Sync Plugin (Fallback)
- **Tool**: `knowledge-sync.go`
- **Command**: `go run . knowledge-sync --path="CLI/"`
- **Config**: NOTION_API_KEY, NOTION_DATABASE_ID

### 5.3 Notion Database Structure

**Point 1**: Create CLI Architecture Category
- **Location**: Notion database `34fb60e8155a80c787a6ce3ea9f631af`
- **Category**: CLI Architecture
- **Subcategories**: Flag Management, Error Handling, Command Validation, Tool System, MCP Integration, Examples

**Point 2**: Add Properties
- Category (Select)
- Status (Select)
- Priority (Select)
- Component (Select)
- Type (Select)
- Related Knowledge (Relation)
- Tags (Multi-select)
- AI Summary (Text)
- AI Category (Select)
- AI Priority (Select)
- AI Relationships (Multi-select)
- File Path (Text)
- Sync Status (Select)
- Last Sync (Date)

---

## 6. Data Flow Diagram

```
CLI Documentation (apps/cli/docs/)
  ↓
[Phase 1] Move to Ti-learning-lab/03_Knowledge/CLI/
  ↓
[Phase 2] AI Sync Agent Analysis
  ├─→ LLM Analysis (content, concepts, patterns)
  ├─→ Categorization (CLI Architecture subcategories)
  ├─→ Priority Determination (P0-P3)
  ├─→ Summary Generation (intelligent summary)
  └─→ Relationship Detection (with existing knowledge)
  ↓
[Phase 3] Sync to Notion
  ├─→ Create page with AI metadata
  ├─→ Set properties (Category, Status, Priority, etc.)
  ├─→ Add content (truncated if needed)
  └─→ Return Notion URL
  ↓
[Phase 4] Update Sync Log
  ├─→ Record file hash
  ├─→ Record Notion URL
  └─→ Record sync timestamp
  ↓
[Phase 5] Learning System
  ├─→ Learn patterns
  ├─→ Update KnowledgeMemory
  └─→ Improve future categorization
  ↓
[Phase 6] BD Tool Logging
  └─→ Log task completion
```

---

## 7. Dependencies

### 7.1 External Dependencies

1. **Notion API**
   - API Key: `NOTION_API_KEY` (from `Z:\00_SECRET\notion.env`)
   - Database ID: `34fb60e8155a80c787a6ce3ea9f631af`
   - Rate Limits: Need to handle

2. **LLM Provider** (cho AI analysis)
   - Provider: `LLM_PROVIDER` (openai, anthropic, router)
   - API Key: `LLM_API_KEY`
   - Base URL: `LLM_BASE_URL`

### 7.2 Internal Dependencies

1. **BD Tool**
   - Path: `Z:\02_CORE\_cli\bin\bd.exe`
   - Environment: `TI_DATA_DIR=Z:\03_DATA\ti`

2. **Ti LLM Client**
   - Package: `github.com/ti/llm`
   - Configuration: llm.Config

---

## 8. Risks

### Risk 1: Breaking Existing Router Sync

**Probability**: Medium
**Impact**: High

**Mitigation**:
- Test Router sync before making changes
- Keep Router sync as default path
- Use `--path` flag for CLI sync
- Separate sync logs

### Risk 2: AI Agent Incomplete

**Probability**: High
**Impact**: High

**Mitigation**:
- Complete stub functions first
- Test with basic sync before AI sync
- Fallback to basic sync if AI fails
- Manual categorization as backup

### Risk 3: Notion API Rate Limits

**Probability**: Medium
**Impact**: Medium

**Mitigation**:
- Add delays between syncs (100ms)
- Batch operations
- Retry logic
- Monitor rate limits

### Risk 4: File Move Breaks References

**Probability**: Low
**Impact**: Medium

**Mitigation**:
- Search for all references before move
- Update all references
- Test build after move
- Keep backup

---

## 9. Success Criteria

- [ ] Architecture document approved
- [ ] Integration points defined
- [ ] Sync option selected
- [ ] AI agent stub functions completed (if Option A/B)
- [ ] CLI docs moved to Ti-learning-lab
- [ ] References updated
- - [ ] Build successful after move
- [ ] Notion database configured with CLI Architecture category
- [ ] Test sync with dry-run
- [ ] Full sync successful
- [ ] AI analysis working (if using AI sync)
- [ ] Learning system working
- [ ] BD tool logging successful

---

## 10. Next Steps

1. **Get approval** for architecture and integration approach
2. **Complete AI agent stubs** (if Option A/B selected)
3. **Move CLI docs** to Ti-learning-lab
4. **Update references** in codebase
5. **Test build** after move
6. **Configure Notion database** with CLI Architecture category
7. **Test sync** with dry-run
8. **Run full sync**
9. **Verify sync** in Notion
10. **Setup automation** (CI/CD hooks, scheduled sync)

---

**Trạng thái**: Architecture Design Complete ✅
**Next**: Get user approval for integration approach
