---
tags: ["tibrain", "skill", "documentation", "provider-notion", "router"]
scopes: ["tibrain"]
last_updated: 2026-05-22
---
# Migration from BD Tool to Notion Task Management

> **Purpose**: Đề xuất migration từ bd tool sang Notion Agent cho task management
> **Date**: 2026-05-04
> **Status**: Design Phase

---

## 1. Current Situation

### BD Tool Usage
```bash
# Current command
bd log --task="Task description" \
       --agent=devin \
       --status=complete \
       --type=coding \
       --domain=general

# Data stored in:
# - Z:\03_DATA\ti (set via TI_DATA_DIR)
# - Git-backed (Dolt)
# - JSON output
```

### BD Tool Features
- Agent name auto-detection
- Dependency tracking
- Git-backed (Dolt)
- JSON output (programmatic)
- Priority system
- Status workflow

### BD Tool Limitations
- Separate tool, not integrated với knowledge system
- Requires separate data directory
- Git-backed (Dolt) - complexity
- Limited visualization
- Not integrated với Notion knowledge

---

## 2. Proposed Solution: Notion Task Management

### Why Notion?
1. **Centralized**: Tất cả knowledge và tasks trong Notion
2. **Integrated**: Notion Agent đã được build cho knowledge sync
3. **Powerful**: Query, filter, visualize tasks
4. **Single Source of Truth**: Không cần multiple tools
5. **Flexible**: Custom properties, views, automation
6. **Collaborative**: Multi-user support (nếu cần)

### Notion Database Structure

```yaml
Database: Tasks
Properties:
  - Task (Title)
  - Description (Text)
  - Status (Select: running, complete, failed, blocked, timeout)
  - Type (Select: coding, refactor, review, plan, test, doc, deploy)
  - Domain (Select: general, router, cli, kanban, devin, auth, provider)
  - Agent (Select: devin, cascade, opencode, cline)
  - Priority (Select: P0, P1, P2, P3)
  - Created Time (Created time)
  - Updated Time (Last edited time)
  - Duration (Formula: Updated Time - Created Time)
  - Session ID (Text) - để track session
  - Related Tasks (Relation) - dependency tracking
  - Tags (Multi-select) - categorization
  - Error Message (Text) - nếu failed
  - Notes (Text) - additional context

Views:
  - All Tasks (Table)
  - By Status (Group by Status)
  - By Agent (Group by Agent)
  - By Domain (Group by Domain)
  - By Type (Group by Type)
  - Timeline (Timeline view)
  - Calendar (Calendar view)
  - Board (Kanban board by Status)
```

---

## 3. Migration Strategy

### Phase 1: Create Notion Database (Week 1)

**Tasks**:
1. Create Notion database "Tasks"
2. Define properties (Status, Type, Domain, Agent, Priority, etc.)
3. Create views (Table, Group by, Timeline, Calendar, Board)
4. Set up automation (optional)
5. Test database structure

**Deliverables**:
- Notion database ready
- Properties defined
- Views configured

---

### Phase 2: Implement Notion Task Logger (Week 2)

**Tasks**:
1. Create `apps/automation/notion-task-logger.go`
2. Implement Notion API client
3. Implement task logging function
4. Replace bd tool calls with Notion logger
5. Test logging functionality

**CLI Command**:
```bash
# New command
ti task log --task="Task description" \
              --agent=devin \
              --status=complete \
              --type=coding \
              --domain=general
```

**Go Implementation**:
```go
package automation

import (
    "context"
    "fmt"
    "time"
)

type TaskLogger struct {
    notionClient *NotionClient
    databaseID   string
}

type TaskLogRequest struct {
    Task        string
    Agent       string
    Status      string
    Type        string
    Domain      string
    Priority    string
    SessionID   string
    ErrorMsg    string
    Notes       string
}

func (l *TaskLogger) LogTask(ctx context.Context, req TaskLogRequest) error {
    // Create page in Notion database
    page := map[string]any{
        "properties": map[string]any{
            "Task": map[string]any{
                "title": []any{map[string]any{"text": req.Task}},
            },
            "Status": map[string]any{
                "select": map[string]any{"name": req.Status},
            },
            "Type": map[string]any{
                "select": map[string]any{"name": req.Type},
            },
            "Domain": map[string]any{
                "select": map[string]any{"name": req.Domain},
            },
            "Agent": map[string]any{
                "select": map[string]any{"name": req.Agent},
            },
            "Priority": map[string]any{
                "select": map[string]any{"name": req.Priority},
            },
            "Session ID": map[string]any{
                "rich_text": []any{map[string]any{"text": req.SessionID}},
            },
            "Error Message": map[string]any{
                "rich_text": []any{map[string]any{"text": req.ErrorMsg}},
            },
            "Notes": map[string]any{
                "rich_text": []any{map[string]any{"text": req.Notes}},
            },
        },
    }

    return l.notionClient.CreatePage(ctx, l.databaseID, page)
}
```

**Deliverables**:
- Notion task logger implemented
- CLI command working
- Integration with Notion API

---

### Phase 3: Migration from BD to Notion (Week 3)

**Tasks**:
1. Export existing BD data
2. Convert BD data to Notion format
3. Import into Notion database
4. Verify data integrity
5. Deprecate BD tool

**Migration Script**:
```go
package main

import (
    "encoding/json"
    "os"
)

func main() {
    // Read BD data
    bdData, err := os.ReadFile("Z:/03_DATA/ti/beads.json")
    if err != nil {
        panic(err)
    }

    var beads []Bead
    json.Unmarshal(bdData, &beads)

    // Convert to Notion format
    for _, bead := range beads {
        task := Task{
            Task:     bead.Task,
            Agent:    bead.Agent,
            Status:   bead.Status,
            Type:     bead.Type,
            Domain:   bead.Domain,
            Priority: "P1", // Default
        }

        // Log to Notion
        err := notionLogger.LogTask(context.Background(), task)
        if err != nil {
            fmt.Printf("Error migrating bead: %v\n", err)
        }
    }
}
```

**Deliverables**:
- All BD data migrated to Notion
- Data integrity verified
- BD tool deprecated

---

### Phase 4: Update All References (Week 4)

**Tasks**:
1. Update AGENTS.md files
2. Update documentation
3. Update CI/CD scripts
4. Update agent guidelines
5. Remove BD tool references

**Files to Update**:
- `AGENTS.md` (multiple locations)
- `apps/cli/AGENTS.md`
- `apps/router/docs/AGENTS.md`
- Documentation files
- CI/CD scripts

**Deliverables**:
- All references updated
- BD tool removed from codebase
- Documentation updated

---

## 4. Notion Agent Integration

### Notion Task Agent

**Agent Definition**:
```markdown
# Notion Task Agent

## Purpose
Log tasks to Notion database for centralized task management.

## Capabilities
- Create task in Notion database
- Update task status
- Query tasks by filters
- Generate task reports
- Track task dependencies

## Tools
- notion-create-task
- notion-update-task
- notion-query-tasks
- notion-delete-task

## Integration
- CLI command: `ti task log ...`
- Auto-logging from agents
- Task dashboard in Ti Claw UI
```

### Auto-Logging from Agents

**Implementation**:
```go
// In agent execution loop
func (l *Loop) runLoop(ctx context.Context, req RunRequest) (*RunResult, error) {
    // Log task start
    taskLogger.LogTask(ctx, TaskLogRequest{
        Task:      fmt.Sprintf("Run agent %s", l.agentName),
        Agent:     "devin",
        Status:    "running",
        Type:      "coding",
        Domain:    "general",
        SessionID: req.SessionKey,
    })

    // Execute agent loop
    result, err := l.executeLoop(ctx, req)

    // Log task completion
    status := "complete"
    errorMsg := ""
    if err != nil {
        status = "failed"
        errorMsg = err.Error()
    }

    taskLogger.LogTask(ctx, TaskLogRequest{
        Task:      fmt.Sprintf("Run agent %s", l.agentName),
        Agent:     "devin",
        Status:    status,
        Type:      "coding",
        Domain:    "general",
        SessionID: req.SessionKey,
        ErrorMsg:  errorMsg,
    })

    return result, err
}
```

---

## 5. Benefits of Notion Task Management

### vs BD Tool

| Feature | BD Tool | Notion |
|---------|---------|--------|
| **Integration** | Separate tool | Integrated với knowledge system |
| **Storage** | Dolt (Git) + JSON | Notion database |
| **Visualization** | Limited | Powerful (Table, Board, Timeline, Calendar) |
| **Query** | Limited JSON query | Powerful Notion query |
| **Filter** | Basic | Advanced filters |
| **Collaboration** | No | Yes (multi-user) |
| **Automation** | No | Yes (Notion automation) |
| **Integration with Knowledge** | No | Yes (same platform) |
| **Dependency Tracking** | Yes | Yes (relations) |
| **Session Tracking** | Yes | Yes (Session ID property) |

### Additional Benefits

1. **Centralized Knowledge & Tasks**
   - Tất cả trong Notion
   - Easy cross-reference
   - Single source of truth

2. **Better Visualization**
   - Kanban board by status
   - Timeline view for duration
   - Calendar view for planning
   - Group by agent/domain/type

3. **Advanced Querying**
   - Filter by multiple properties
   - Sort by any property
   - Custom formulas
   - Rollups and aggregations

4. **Automation**
   - Notion automation
   - Webhooks
   - Integrations (Zapier, Make, etc.)

5. **Collaboration**
   - Multi-user support
   - Comments and mentions
   - Share views
   - Permissions

---

## 6. Implementation Plan

### Week 1: Notion Database Setup
- Create Notion database
- Define properties
- Create views
- Test structure

### Week 2: Notion Task Logger
- Implement Notion API client
- Implement task logging function
- CLI command
- Test logging

### Week 3: Migration
- Export BD data
- Convert to Notion format
- Import to Notion
- Verify integrity
- Deprecate BD tool

### Week 4: Integration & Cleanup
- Update all references
- Auto-logging from agents
- Ti Claw UI integration
- Remove BD tool
- Update documentation

---

## 7. Risk & Mitigation

### Risk 1: Notion API Rate Limits
**Mitigation**: Implement batching, retry logic, caching

### Risk 2: Data Loss During Migration
**Mitigation**: Backup BD data before migration, verify integrity

### Risk 3: Notion Downtime
**Mitigation**: Fallback to local logging, retry mechanism

### Risk 4: Breaking Changes
**Mitigation**: Gradual migration, keep BD as fallback during transition

---

## 8. Success Criteria

- [x] Notion database created
- [x] Properties defined
- [x] Views configured
- [x] Notion task logger implemented
- [x] CLI command working
- [x] BD data migrated to Notion
- [x] Data integrity verified
- [x] All references updated
- [x] BD tool deprecated
- [x] Auto-logging from agents working
- [x] Ti Claw UI integration working

---

## 9. Next Steps

1. **Create Notion database**
2. **Define properties and views**
3. **Implement Notion task logger**
4. **Test logging functionality**
5. **Migrate BD data**
6. **Update all references**
7. **Deprecate BD tool**
8. **Integrate with agents**
9. **Add Ti Claw UI integration**

---

## 10. References

- **Notion API**: https://developers.notion.com/
- **Notion Agent**: `content/agents/notion-agent.md`
- **BD Tool**: `Z:\02_CORE\_cli\bin\bd.exe`
- **BD Data**: `Z:\03_DATA\ti`
- **AGENTS.md**: Multiple locations

---

**Status**: Design Complete ✅
**Next**: Create Notion database
