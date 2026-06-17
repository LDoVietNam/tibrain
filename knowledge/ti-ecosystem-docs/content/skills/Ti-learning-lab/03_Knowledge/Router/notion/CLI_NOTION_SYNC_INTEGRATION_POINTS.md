---
tags: ["tibrain", "documentation", "cli", "skill", "testing"]
scopes: ["code", "cli", "tibrain"]
last_updated: 2026-05-22
---
# Integration Points - CLI Documentation Sync với Notion

> **Mục đích**: Định nghĩa chi tiết các integration points cho CLI documentation sync
> **Ngày tạo**: 2026-05-05
> **Trạng thái**: Integration Design
> **Version**: 1.0

---

## 1. File System Integration

### Point 1.1: Move CLI Documentation

**Source**: `apps/cli/docs/`
**Destination**: `Ti-learning-lab/03_Knowledge/CLI/`

**Files to move**:
```
apps/cli/docs/
├── CLI_GUIDE.md                          → Ti-learning-lab/03_Knowledge/CLI/
├── CLI_FLAG_MANAGEMENT.md                → Ti-learning-lab/03_Knowledge/CLI/
├── CLI_ERROR_HANDLING.md                 → Ti-learning-lab/03_Knowledge/CLI/
├── CLI_COMMAND_VALIDATION.md             → Ti-learning-lab/03_Knowledge/CLI/
├── CLI_TOOL_SYSTEM.md                    → Ti-learning-lab/03_Knowledge/CLI/
├── CLI_MCP_OAUTH.md                      → Ti-learning-lab/03_Knowledge/CLI/
└── CLI_EXAMPLES.md                       → Ti-learning-lab/03_Knowledge/CLI/
```

**Command**:
```bash
# Move files
mv apps/cli/docs/*.md Ti-learning-lab/03_Knowledge/CLI/

# Verify
ls -la Ti-learning-lab/03_Knowledge/CLI/
```

**Validation**:
- [ ] All 7 files moved successfully
- [ ] File permissions preserved
- [ ] No files left in apps/cli/docs/
- [ ] Files readable

---

### Point 1.2: Update References

**Files to search for references**:
```bash
# Search for "apps/cli/docs" references
grep -r "apps/cli/docs" --include="*.md" --include="*.go" .

# Search for CLI docs references
grep -r "CLI_GUIDE.md" --include="*.md" --include="*.go" .
grep -r "CLI_FLAG_MANAGEMENT.md" --include="*.md" --include="*.go" .
grep -r "CLI_ERROR_HANDLING.md" --include="*.md" --include="*.go" .
grep -r "CLI_COMMAND_VALIDATION.md" --include="*.md" --include="*.go" .
grep -r "CLI_TOOL_SYSTEM.md" --include="*.md" --include="*.go" .
grep -r "CLI_MCP_OAUTH.md" --include="*.md" --include="*.go" .
grep -r "CLI_EXAMPLES.md" --include="*.md" --include="*.go" .
```

**Expected references**:
- `AGENTS.md` (nếu có CLI docs links)
- `docs/CLI_GUIDE.md` (internal links to other CLI docs)
- Any README files
- Any Go files with doc comments

**Update pattern**:
```markdown
# Before
[CLI Guide](apps/cli/docs/CLI_GUIDE.md)

# After
[CLI Guide](Ti-learning-lab/03_Knowledge/CLI/CLI_GUIDE.md)
```

**Validation**:
- [ ] All references found
- [ ] All references updated
- [ ] No broken links
- [ ] Build successful

---

## 2. Notion Database Integration

### Point 2.1: Create CLI Architecture Category

**Database**: `34fb60e8155a80c787a6ce3ea9f631af`
**Category Name**: CLI Architecture
**Subcategories**:
- Flag Management
- Error Handling
- Command Validation
- Tool System
- MCP Integration
- Examples
- Overview

**Notion API Call**:
```bash
# Create category via Notion API
curl -X POST 'https://api.notion.com/v1/databases/34fb60e8155a80c787a6ce3ea9f631af' \
  -H 'Authorization: Bearer $NOTION_API_KEY' \
  -H 'Notion-Version: 2022-06-28' \
  -H 'Content-Type: application/json' \
  --data '{
    "properties": {
      "Category": {
        "select": {
          "options": [
            {"name": "CLI Architecture", "color": "blue"}
          ]
        }
      }
    }
  }'
```

**Validation**:
- [ ] Category created in Notion
- [ ] Category visible in database
- [ ] Category color set correctly

---

### Point 2.2: Configure Properties

**Required Properties**:

| Property Name | Type | Description | CLI-Specific |
|--------------|------|-------------|--------------|
| Tên | Title | Document name | ✅ |
| Type | Select | Document type (Changelog, Guide, etc.) | ✅ (Guide) |
| Status | Select | Document status (Done, In Progress, etc.) | ✅ (Done) |
| Priority | Select | Priority (P0, P1, P2, P3) | ✅ (AI-determined) |
| Category | Select | Knowledge category | ✅ (CLI Architecture) |
| Component | Select | Component name | ✅ (Flag Management, etc.) |
| Related Knowledge | Relation | Links to other knowledge | ✅ |
| Tags | Multi-select | Tags | ✅ |
| AI Summary | Text | AI-generated summary | ✅ |
| AI Category | Select | AI-determined category | ✅ |
| AI Priority | Select | AI-determined priority | ✅ |
| AI Relationships | Multi-select | AI-detected relationships | ✅ |
| File Path | Text | Source file path | ✅ |
| Sync Status | Select | Sync status (Synced, Pending, Error) | ✅ |
| Last Sync | Date | Last sync timestamp | ✅ |

**Validation**:
- [ ] All properties configured
- [ ] Property types correct
- [ ] CLI-specific properties added

---

## 3. AI Sync Agent Integration

### Point 3.1: Complete Stub Functions

**File**: `apps/automation/knowledge-sync-ai.go`

**Functions to complete**:

#### 3.1.1: NotionClient CreatePage

**Current**: Stub
**Required**: Full implementation

```go
func (nc *NotionClient) CreatePage(ctx context.Context, title, content, category string, properties map[string]interface{}) (string, error) {
    // Implementation needed
    // 1. Create page in Notion
    // 2. Set properties
    // 3. Add content
    // 4. Return page ID/URL
}
```

#### 3.1.2: scanKnowledgeFiles

**Current**: Stub
**Required**: Scan directory for markdown files

```go
func (ks *KnowledgeSyncAI) scanKnowledgeFiles() ([]string, error) {
    // Implementation needed
    // 1. Read KnowledgePath
    // 2. Scan for .md files
    // 3. Return file paths
}
```

#### 3.1.3: isSynced

**Current**: Stub
**Required**: Check if file already synced

```go
func (ks *KnowledgeSyncAI) isSynced(filePath string) bool {
    // Implementation needed
    // 1. Check sync log
    // 2. Compare file hash
    // 3. Return true if synced and unchanged
}
```

#### 3.1.4: KnowledgeMemory Save/Load

**Current**: Stub
**Required**: Save and load knowledge memory

```go
func (km *KnowledgeMemory) Save(path string) error {
    // Implementation needed
    // 1. Marshal to JSON
    // 2. Write to file
}

func (km *KnowledgeMemory) Load(path string) error {
    // Implementation needed
    // 1. Read from file
    // 2. Unmarshal JSON
    // 3. Populate memory
}
```

**Validation**:
- [ ] All stub functions completed
- [ ] Functions tested
- [ ] Error handling added
- [ ] Logging added

---

### Point 3.2: Add CLI-Specific AI Prompts

**File**: `apps/automation/knowledge-sync-ai.go`

**Prompts to add**:

#### 3.2.1: CLI Categorization Prompt

```go
const CLICategorizationPrompt = `Bạn là một AI chuyên gia phân tích kiến thức về CLI Architecture.

Hãy phân tích nội dung CLI documentation sau và categorize nó:

Content:
{{content}}

Output format (JSON):
{
    "category": "CLI Architecture",
    "subcategory": "Flag Management | Error Handling | Command Validation | Tool System | MCP Integration | Examples | Overview",
    "component": "Tên component cụ thể",
    "priority": "P0 | P1 | P2 | P3",
    "tags": ["tag1", "tag2", "tag3"],
    "summary": "Tóm tắt ngắn gọn (2-3 câu)"
}`
```

#### 3.2.2: CLI Relationship Detection Prompt

```go
const CLIRelationshipPrompt = `Bạn là một AI chuyên gia phát hiện relationships trong kiến thức CLI Architecture.

Hãy phân tích nội dung sau và phát hiện relationships với các kiến thức khác:

Content:
{{content}}
Existing Knowledge:
{{existing_knowledge}}

Output format (JSON):
{
    "relationships": [
        {
            "type": "related_to | depends_on | extends | implements",
            "target": "Tên kiến thức liên quan",
            "reason": "Lý do relationship"
        }
    ]
}`
```

**Validation**:
- [ ] CLI prompts added
- [ ] Prompts tested
- [ ] Output format validated

---

### Point 3.3: Add CLI Category Mapping

**File**: `apps/automation/knowledge-sync-ai.go`

**Mapping**:
```go
var CLICategoryMapping = map[string]string{
    "CLI_FLAG_MANAGEMENT.md":     "Flag Management",
    "CLI_ERROR_HANDLING.md":      "Error Handling",
    "CLI_COMMAND_VALIDATION.md":  "Command Validation",
    "CLI_TOOL_SYSTEM.md":         "Tool System",
    "CLI_MCP_OAUTH.md":           "MCP Integration",
    "CLI_EXAMPLES.md":            "Examples",
    "CLI_GUIDE.md":               "Overview",
}
```

**Validation**:
- [ ] Category mapping added
- [ ] All files mapped
- [ ] Mapping tested

---

## 4. CLI Command Integration

### Point 4.1: Add CLI Sync Command

**File**: `apps/cli/cmd/knowledge-sync.go`

**Current**: Basic sync command
**Required**: Add CLI-specific sync

```go
// Add flag for CLI sync
var cliPathFlag = flag.String("cli", "", "Sync CLI documentation (path to CLI knowledge dir)")

// In main
if *cliPathFlag != "" {
    // Use CLI-specific prompts
    // Use CLI category mapping
    // Use CLI sync log
}
```

**Command**:
```bash
# Sync CLI docs with AI
go run apps/automation/*.go knowledge-sync-ai --cli="Ti-learning-lab/03_Knowledge/CLI/"

# Sync CLI docs with basic plugin
go run apps/automation/*.go knowledge-sync --path="Ti-learning-lab/03_Knowledge/CLI/"
```

**Validation**:
- [ ] CLI flag added
- [ ] Command tested
- [ ] Help text updated

---

## 5. Environment Integration

### Point 5.1: Configure Environment Variables

**Required Variables**:
```bash
# Notion API
export NOTION_API_KEY=$(cat Z:\00_SECRET\notion.env)
export NOTION_DATABASE_ID=34fb60e8155a80c787a6ce3ea9f631af

# LLM (for AI sync)
export LLM_PROVIDER=router
export LLM_API_KEY=$(cat Z:\00_SECRET\router.env)
export LLM_BASE_URL=https://api.router.dev

# BD Tool
export TI_DATA_DIR=Z:\03_DATA\ti
```

**Validation**:
- [ ] All variables set
- [ ] Variables accessible
- [ ] API keys valid

---

## 6. BD Tool Integration

### Point 6.1: Add CLI Sync Task

**File**: `apps/automation/knowledge-sync-ai.go`

**Current**: Basic BD tool logging
**Required**: Add CLI-specific task logging

```go
// In sync function
bdTool := bd.NewBDTool("Z:\\02_CORE\\_cli\\bin\\bd.exe")
task := &bd.Task{
    Name:        "CLI Knowledge Sync",
    Description: "Sync CLI documentation to Notion with AI analysis",
    Status:      "In Progress",
    StartTime:   time.Now(),
}

// After sync
task.Status = "Done"
task.EndTime = time.Now()
task.Result = fmt.Sprintf("Synced %d CLI docs to Notion", len(syncedFiles))
bdTool.LogTask(task)
```

**Validation**:
- [ ] BD tool logging added
- [ ] Task name CLI-specific
- [ ] Task description clear

---

## 7. CI/CD Integration (Future)

### Point 7.1: Add Pre-commit Hook

**File**: `.git/hooks/pre-commit` (or use pre-commit framework)

**Purpose**: Sync CLI docs on commit

```bash
#!/bin/bash
# Check if CLI docs changed
if git diff --cached --name-only | grep -q "Ti-learning-lab/03_Knowledge/CLI/"; then
    echo "CLI docs changed, syncing to Notion..."
    go run apps/automation/*.go knowledge-sync-ai --cli="Ti-learning-lab/03_Knowledge/CLI/"
fi
```

**Validation**:
- [ ] Hook added
- [ ] Hook executable
- [ ] Hook tested

---

### Point 7.2: Add Scheduled Sync

**File**: `.github/workflows/sync-cli-docs.yml` (or use cron)

**Purpose**: Scheduled sync to Notion

```yaml
name: Sync CLI Docs to Notion
on:
  schedule:
    - cron: '0 0 * * *'  # Daily at midnight
  workflow_dispatch:

jobs:
  sync:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - name: Setup Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.21'
      - name: Sync CLI docs
        env:
          NOTION_API_KEY: ${{ secrets.NOTION_API_KEY }}
          LLM_API_KEY: ${{ secrets.LLM_API_KEY }}
        run: go run apps/automation/*.go knowledge-sync-ai --cli="Ti-learning-lab/03_Knowledge/CLI/"
```

**Validation**:
- [ ] Workflow added
- [ ] Secrets configured
- [ ] Workflow tested

---

## 8. Testing Integration

### Point 8.1: Add Integration Tests

**File**: `apps/automation/knowledge-sync-ai_test.go`

**Tests**:
```go
func TestCLISync(t *testing.T) {
    // Test CLI sync with dry-run
    ks := &KnowledgeSyncAI{
        Config: Config{
            KnowledgePath: "Ti-learning-lab/03_Knowledge/CLI/",
            DryRun:        true,
        },
    }
    err := ks.Sync()
    assert.NoError(t, err)
}

func TestCLICategorization(t *testing.T) {
    // Test CLI categorization prompt
    content := readTestFile("CLI_FLAG_MANAGEMENT.md")
    result, err := ks.analyzeContent(content)
    assert.NoError(t, err)
    assert.Equal(t, "CLI Architecture", result.Category)
    assert.Equal(t, "Flag Management", result.Subcategory)
}
```

**Validation**:
- [ ] Tests added
- [ ] Tests passing
- [ ] Coverage > 80%

---

## 9. Monitoring Integration

### Point 9.1: Add Sync Metrics

**Metrics to track**:
- Number of files synced
- Sync duration
- Sync errors
- AI analysis duration
- Notion API rate limits
- Learning system accuracy

**Implementation**:
```go
type SyncMetrics struct {
    FilesSynced      int
    SyncDuration     time.Duration
    AnalysisDuration time.Duration
    Errors           int
    RateLimitHits    int
}
```

**Validation**:
- [ ] Metrics collected
- [ ] Metrics logged
- [ ] Metrics dashboard (optional)

---

## Summary Checklist

### File System
- [ ] Move CLI docs to Ti-learning-lab
- [ ] Update all references
- [ ] Test build

### Notion
- [ ] Create CLI Architecture category
- [ ] Configure properties
- [ ] Test API access

### AI Agent
- [ ] Complete stub functions
- [ ] Add CLI prompts
- [ ] Add category mapping
- [ ] Test AI analysis

### CLI Command
- [ ] Add CLI sync flag
- [ ] Test command
- [ ] Update help text

### Environment
- [ ] Configure variables
- [ ] Validate API keys
- [ ] Test access

### BD Tool
- [ ] Add CLI task logging
- [ ] Test logging

### CI/CD (Future)
- [ ] Add pre-commit hook
- [ ] Add scheduled sync
- [ ] Configure secrets

### Testing
- [ ] Add integration tests
- [ ] Test all scenarios
- [ ] Verify coverage

### Monitoring
- [ ] Add metrics collection
- [ ] Log metrics
- [ ] Create dashboard

---

**Trạng thái**: Integration Points Defined ✅
**Next**: Evaluate sync options
