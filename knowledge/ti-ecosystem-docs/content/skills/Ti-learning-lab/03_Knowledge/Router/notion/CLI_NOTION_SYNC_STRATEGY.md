---
tags: ["tibrain", "documentation", "cli", "skill", "provider-notion"]
scopes: ["cli", "tibrain"]
last_updated: 2026-05-22
---
# Chiến Lược Sync CLI Documentation với Notion

> **Mục đích**: Định nghĩa chiến lược sync CLI documentation đến Notion knowledge base
> **Ngày tạo**: 2026-05-05
> **Trạng thái**: Design Phase
> **Ưu tiên**: High

---

## 1. Tổng quan

### Vấn đề hiện tại
- CLI documentation mới được tạo trong `apps/cli/docs/` (6 files)
- Chưa có workflow sync đến Notion knowledge base
- Không có category cho CLI Architecture trong Notion
- Không có integration với AI-powered analysis system

### Mục tiêu
- Sync toàn bộ CLI documentation đến Notion
- Tạo category "CLI Architecture" trong Notion knowledge base
- Integration với AI-powered analysis system (knowledge-sync-ai agent)
- Tự động sync với changelog tracking
- Learning system để improve categorization over time

---

## 2. Architecture hiện tại

### Knowledge Sync Workflow (Hiện tại)

```
Research & Document (Markdown)
  ↓
AI-Powered Sync (knowledge-sync-ai agent)
  ↓
Verification & Learning
  ↓
Task Logging (BD tool)
```

**AI Agent Capabilities:**
- Content Analysis (LLM reads và comprehends)
- Categorization (auto-categorize vào UI/Backend patterns, architecture, etc.)
- Prioritization (assign priority levels)
- Summary Generation (intelligent summaries)
- Relationship Detection (find connections với existing knowledge)
- Learning (remember patterns across sync operations)
- Memory (build knowledge graph)

### Location hiện tại
- **Knowledge Path**: `Z:\10_WORKPLACE\Ti\Ti-learning-lab\03_Knowledge\Router\`
- **Sync Log**: `Z:\10_WORKPLACE\Ti\Ti-learning-lab\scripts\sync-log.json`
- **Notion Database**: `34fb60e8155a80c787a6ce3ea9f631af`

---

## 3. CLI Documentation Structure

### Files hiện tại trong `apps/cli/docs/`

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

### Category đề xuất trong Notion

**Main Category**: CLI Architecture

**Subcategories:**
1. **Flag Management** - Flag management system
2. **Error Handling** - Error handling patterns
3. **Command Validation** - Command validation system
4. **Tool System** - Tool architecture
5. **MCP Integration** - MCP OAuth integration
6. **Examples** - Code examples và patterns

---

## 4. Chiến lược Sync

### Option 1: AI-Powered Sync (Khuyên dùng) ✅

**Workflow:**
```
1. Move CLI docs từ apps/cli/docs/ đến Ti-learning-lab/03_Knowledge/CLI/
2. Update learning lab index
3. Run AI sync agent: go run . knowledge-sync-ai --path="CLI/"
4. AI agent analyzes, categorizes, detects relationships
5. Sync to Notion với AI-generated metadata
6. Verify sync trong Notion
```

**Ưu điểm:**
- Consistent với workflow hiện tại
- AI analysis và categorization
- Learning system improves over time
- Knowledge graph integration
- Automatic relationship detection với existing knowledge

**Nhược điểm:**
- Cần move files (thay đổi location)
- Phụ thuộc vào AI agent availability

### Option 2: Manual Sync (Fallback)

**Workflow:**
```
1. Tạo script PowerShell mới cho CLI docs sync
2. Sync trực tiếp đến Notion API
3. Manual categorization
4. Manual relationship mapping
```

**Ưu điểm:**
- Không cần move files
- Control hoàn toàn
- Không phụ thuộc AI

**Nhược điểm:**
- Không có AI analysis
- Manual categorization
- Không có learning system
- Không có relationship detection

### Option 3: Hybrid (Flexible)

**Workflow:**
```
1. Move CLI docs đến Ti-learning-lab/03_Knowledge/CLI/
2. Use AI sync cho regular sync
3. Manual sync cho urgent updates
```

**Ưu điểm:**
- Best của cả 2 options
- Flexibility
- AI analysis + manual control

---

## 5. Implementation Plan (Option 1: AI-Powered Sync)

### Phase 1: Preparation (Day 1)

**Tasks:**
1. Create directory: `Ti-learning-lab/03_Knowledge/CLI/`
2. Move CLI docs từ `apps/cli/docs/` đến directory mới
3. Update learning lab index với CLI knowledge entries
4. Update references trong codebase (nếu cần)
5. Test file structure

**Deliverables:**
- CLI docs moved to Ti-learning-lab
- Learning lab index updated
- File structure verified

### Phase 2: AI Sync Integration (Day 2)

**Tasks:**
1. Test AI sync agent với CLI path: `go run . knowledge-sync-ai --path="CLI/" --dry-run`
2. Review AI-generated metadata
3. Adjust AI prompts nếu cần (cho CLI-specific categorization)
4. Run full sync: `go run . knowledge-sync-ai --path="CLI/"`
5. Verify sync trong Notion

**Deliverables:**
- AI sync working với CLI docs
- Notion pages created với AI metadata
- Sync log updated

### Phase 3: Notion Database Setup (Day 2-3)

**Tasks:**
1. Create "CLI Architecture" category trong Notion (nếu chưa có)
2. Define properties cho CLI docs:
   - Category (Select: Flag Management, Error Handling, etc.)
   - Status (Select: Draft, Review, Published)
   - Priority (Select: P0, P1, P2, P3)
   - Component (Select: internal/cli, internal/tool, internal/mcp, etc.)
   - Related Knowledge (Relation)
   - Tags (Multi-select)
3. Create views:
   - All CLI Docs (Table)
   - By Category (Group by Category)
   - By Component (Group by Component)
   - By Status (Group by Status)
   - Timeline (Timeline view)
4. Test database structure

**Deliverables:**
- Notion database configured
- Properties defined
- Views configured

### Phase 4: Verification & Learning (Day 3)

**Tasks:**
1. Verify tất cả CLI docs synced
2. Check AI categorization accuracy
3. Check relationship detection
4. Review AI-generated summaries
5. Test learning system (run sync again để xem improvement)
6. Update documentation nếu cần

**Deliverables:**
- All CLI docs verified
- AI categorization reviewed
- Learning system tested
- Documentation updated

### Phase 5: Automation Setup (Day 4)

**Tasks:**
1. Setup CI/CD hook để auto-sync khi docs thay đổi
2. Create pre-commit hook để validate docs trước sync
3. Setup scheduled sync (daily/weekly)
4. Monitor sync logs
5. Setup alerts cho sync failures

**Deliverables:**
- Automation setup
- Monitoring configured
- Alerts configured

---

## 6. Notion Database Schema

### Properties cho CLI Docs

```yaml
Database: CLI Architecture Knowledge

Properties:
  - Title (Title) - Tên documentation
  - Category (Select):
    - Flag Management
    - Error Handling
    - Command Validation
    - Tool System
    - MCP Integration
    - Examples
  - Status (Select):
    - Draft
    - Review
    - Published
    - Deprecated
  - Priority (Select):
    - P0 (Critical)
    - P1 (High)
    - P2 (Medium)
    - P3 (Low)
  - Component (Select):
    - internal/cli
    - internal/tool
    - internal/mcp
    - internal/copilot
    - cmd
  - Type (Select):
    - Documentation
    - Guide
    - Reference
    - Examples
  - Related Knowledge (Relation) - Links đến related knowledge
  - Tags (Multi-select) - Additional categorization
  - AI Summary (Text) - AI-generated summary
  - AI Category (Select) - AI-detected category
  - AI Priority (Select) - AI-assigned priority
  - AI Relationships (Multi-select) - AI-detected relationships
  - File Path (Text) - Original file path
  - Created Time (Created time)
  - Updated Time (Last edited time)
  - Sync Status (Select):
    - Pending
    - Synced
    - Failed
  - Last Sync (Date) - Last sync timestamp
  - Notes (Text) - Additional notes
```

### Views

1. **All CLI Docs** (Table view)
   - Columns: Title, Category, Status, Priority, Component, Updated Time

2. **By Category** (Group by Category)
   - Groups: Flag Management, Error Handling, Command Validation, Tool System, MCP Integration, Examples

3. **By Component** (Group by Component)
   - Groups: internal/cli, internal/tool, internal/mcp, internal/copilot, cmd

4. **By Status** (Group by Status)
   - Groups: Draft, Review, Published, Deprecated

5. **Timeline** (Timeline view)
   - Track documentation updates over time

6. **Kanban** (Board view by Status)
   - Columns: Draft → Review → Published → Deprecated

---

## 7. AI Sync Configuration

### CLI-Specific AI Prompts

```go
// CLI-specific categorization prompts
const CLICategorizationPrompt = `
Categorize this CLI documentation into one of these categories:
- Flag Management: Flag management system, registry, environment variables
- Error Handling: Error patterns, error codes, error recovery
- Command Validation: Validators, validation rules, input validation
- Tool System: Tool interface, tool registry, built-in tools
- MCP Integration: MCP OAuth, MCP servers, MCP protocol
- Examples: Code examples, usage patterns, integration examples

Assign priority based on:
- P0: Core architecture components (Flag Management, Error Handling)
- P1: Important features (Command Validation, Tool System)
- P2: Integration features (MCP Integration)
- P3: Examples and references

Detect relationships with existing knowledge:
- OmniRoute UI Patterns (if relevant to CLI UI)
- OmniRoute Backend Patterns (if relevant to CLI backend)
- Other CLI knowledge (if exists)
`
```

### AI Sync Commands

```bash
# Dry run để preview
cd Z:\10_WORKPLACE\Ti\apps\automation
go run . knowledge-sync-ai --path="Z:\10_WORKPLACE\Ti\Ti-learning-lab\03_Knowledge\CLI" --dry-run

# Full sync với AI analysis
go run . knowledge-sync-ai --path="Z:\10_WORKPLACE\Ti\Ti-learning-lab\03_Knowledge\CLI"

# Sync với learning enabled
go run . knowledge-sync-ai --path="Z:\10_WORKPLACE\Ti\Ti-learning-lab\03_Knowledge\CLI" --learn

# Force re-sync (ignore hash)
go run . knowledge-sync-ai --path="Z:\10_WORKPLACE\Ti\Ti-learning-lab\03_Knowledge\CLI" --force
```

---

## 8. Learning Lab Index Update

### Add CLI Knowledge Section

```markdown
## 6. CLI Architecture (Ti CLI)

### Location: `03_Knowledge/CLI/`

#### 6.1 Flag Management System
**File**: `CLI_FLAG_MANAGEMENT.md`
- Centralized flag management
- Type-safe flag definitions
- Environment variable support
- Thread-safe registry

#### 6.2 Error Handling Patterns
**File**: `CLI_ERROR_HANDLING.md`
- Structured error handling
- Error codes and messages
- Exit code support
- Error chaining

#### 6.3 Command Validation
**File**: `CLI_COMMAND_VALIDATION.md`
- Validator components
- Built-in validators
- Rule-based validation
- Batch validation

#### 6.4 Tool System
**File**: `CLI_TOOL_SYSTEM.md`
- Tool interface
- Tool registry
- Built-in tools
- JSON Schema validation

#### 6.5 MCP OAuth Integration
**File**: `CLI_MCP_OAUTH.md`
- OAuth 2.0 support
- OAuth callback handler
- Token storage
- Security features

#### 6.6 CLI Examples
**File**: `CLI_EXAMPLES.md`
- Complete command examples
- Integration patterns
- Testing examples
- Best practices

**Status**: ✅ Documented (Phase 4 Complete)
**Next**: Sync to Notion with AI-powered analysis
```

---

## 9. Best Practices

### Documentation Writing
- Use Vietnamese cho content (technical terms trong English)
- Include code examples với file references
- Add "When to Use" sections cho patterns
- Provide application guidance
- Keep content clear và actionable

### AI Sync Usage
- Use AI sync agent thay vì manual Notion API calls
- Run dry run trước để preview changes
- Use force sync chỉ khi cần thiết
- Check sync status trước full sync
- Monitor AI agent output cho errors

### Task Logging
- AI agent automatically logs với BD tool
- Agent name: `devin`
- Type: `doc`
- Domain: `cli`
- Status: `complete`

### Error Handling
- AI agent handles retries automatically
- Check sync log cho failed syncs
- Verify NOTION_API_KEY trong `Z:\00_SECRET\notion.env`
- Check Notion database ID là đúng
- Review AI agent output cho errors

---

## 10. Success Criteria

- [ ] CLI docs moved đến `Ti-learning-lab/03_Knowledge/CLI/`
- [ ] Learning lab index updated với CLI knowledge
- [ ] AI sync agent working với CLI path
- [ ] Notion database configured với CLI Architecture category
- [ ] All CLI docs synced với AI metadata
- [ ] AI categorization accurate
- [ ] Relationship detection working
- [ ] Learning system tested
- [ ] Automation setup (CI/CD hooks)
- [ ] Monitoring configured

---

## 11. Risk & Mitigation

### Risk 1: AI Categorization Inaccurate
**Mitigation**: 
- Review AI-generated metadata
- Adjust AI prompts nếu cần
- Manual override capability

### Risk 2: File Move Issues
**Mitigation**: 
- Update all references trong codebase
- Test build sau move
- Keep backup trước move

### Risk 3: Notion API Rate Limits
**Mitigation**: 
- Implement batching
- Retry logic
- Caching

### Risk 4: Sync Failures
**Mitigation**: 
- Monitor sync logs
- Fallback to manual sync
- Alerting setup

---

## 12. Next Steps

1. **Create CLI directory** trong Ti-learning-lab
2. **Move CLI docs** đến directory mới
3. **Update learning lab index**
4. **Test AI sync** với CLI path
5. **Configure Notion database**
6. **Run full sync**
7. **Verify sync**
8. **Setup automation**

---

## 13. References

- **Notion Sync Skill**: `.windsurf/skills/omniroute-knowledge-sync/`
- **AI Sync Agent**: `apps/automation/knowledge-sync-ai.go`
- **Knowledge Sync Plugin**: `apps/automation/knowledge-sync.go`
- **CLI Documentation**: `apps/cli/docs/`
- **Learning Lab Index**: `Ti-learning-lab/03_Knowledge/Router/learning-lab-index.md`
- **Notion Secrets**: `Z:\00_SECRET\notion.env`
- **Sync Log**: `Ti-learning-lab/scripts/sync-log.json`

---

**Trạng thái**: Design Complete ✅
**Next**: Phase 1 - Preparation (Move CLI docs to Ti-learning-lab)
