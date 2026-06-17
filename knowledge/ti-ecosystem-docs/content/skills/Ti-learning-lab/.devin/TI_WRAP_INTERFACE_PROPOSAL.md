# Ti Wrap Interface .devin Proposal

> **Ti wrap interface .devin - Agent chỉ gọi API, Ti quản lý toàn bộ memory**

---

## 🎯 Mục Tiêu

Thay vì:
- Agent đọc local .devin files
- Agent chạy sync script thủ công
- Agent quản lý context/knowledge thủ công

Thành:
- Agent chỉ gọi API của Ti
- Ti wrap interface .devin
- Ti quản lý local .devin + centralized brain
- Ti tự động sync, track, manage

---

## 📊 Architecture So Sánh

### Current (Manual .devin):
```
Agent → Đọc .devin/context/project-context.md
Agent → Đọc .devin/knowledge/repo-index.md
Agent → Update .devin/memory/session-memory.md
Agent → Chạy sync-brain.sh thủ công
```

**Nhược điểm:**
- ❌ Agent cần biết .devin structure
- ❌ Agent cần quản lý sync thủ công
- ❌ Không consistent interface
- ❌ Khó maintain

### Proposed (Ti Wrap):
```
Agent → Ti API (.devin context, knowledge, memory)
Ti → Quản lý .devin local
Ti → Auto sync với Ti Brain
Ti → Track progress
```

**Ưu điểm:**
- ✅ Agent đơn giản (chỉ gọi API)
- ✅ Ti quản lý toàn bộ
- ✅ Consistent interface
- ✅ Auto sync, auto track
- ✅ Easier maintain

---

## 🏗️ Proposed Architecture

### Option 1: Ti API Wrapper
```
┌─────────────────────────────────────────────────────────────┐
│                     Ti API Layer                             │
│  Z:\Ti\pkg\brain\ (hoặc Z:\Ti\CLI\internal\brain)            │
└──────────────────────┬──────────────────────────────────────┘
                       │
        ┌──────────────┴──────────────┐
        │                             │
┌───────▼────────┐           ┌───────▼────────┐
│  Local .devin    │           │  Central Brain   │
│  (Ti quản lý)     │           │  (Ti Brain)     │
└──────────────────┘           └──────────────────┘
```

### Option 2: Ti Brain Only (Simplified)
```
┌─────────────────────────────────────────────────────────────┐
│                     Ti Brain                                │
│  Z:\Ti\brain\ (centralized cho tất cả)                     │
│                                                              │
│  ┌──────────────────────────────────────────────────────┐  │
│  │  Project: ti-learning-lab/                          │  │
│  │    ├── context/                                     │  │
│  │    ├── knowledge/                                   │  │  │
│  │    └── memory/                                      │  │
│  └──────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────┘
```

---

## 🔧 Implementation Options

### Option 1: Extend Ti Brain API
Thêm API endpoints vào Ti Brain để quản lý .devin:

```go
// Z:\Ti\pkg\brain\api.go
package brain

type DevinAPI struct {
    project string
    devinPath string
}

func NewDevinAPI(project string) *DevinAPI {
    return &DevinAPI{
        project: project,
        devinPath: fmt.Sprintf("/z/Ti/%s/.devin", project),
    }
}

// Context API
func (d *DevinAPI) GetProjectContext() (string, error) {
    return readFile(d.devinPath + "/context/project-context.md")
}

func (d *DevinAPI) UpdateProjectContext(content string) error {
    return writeFile(d.devinPath + "/context/project-context.md", content)
}

// Knowledge API
func (d *DevinAPI) GetPatterns() (string, error) {
    return readFile(d.devinPath + "/knowledge/patterns.md")
}

func (d *DevinAPI) AddPattern(pattern string) error {
    // Append to patterns.md
    return appendFile(d.devinPath + "/knowledge/patterns.md", pattern)
}

// Memory API
func (d *DevinAPI) GetSessionMemory() (string, error) {
    return readFile(d.devinPath + "/memory/session-memory.md")
}

func (d *DevinAPI) UpdateSessionMemory(content string) error {
    return writeFile(d.devinPath + "/memory/session-memory.md", content)
}

// Sync API
func (d *DevinAPI) SyncWithBrain() error {
    // Auto sync với Ti Brain
    return syncBrain(d.devinPath)
}
```

### Option 2: Ti Brain Centralized (Simplified)
Đưa .devin content vào Ti Brain luôn, không có local copy:

```
Z:\Ti\brain\
└── projects\
    └── ti-learning-lab\
        ├── context/
        ├── knowledge/
        └── memory/
```

Agent chỉ cần gọi:
```go
brain.GetProjectContext("ti-learning-lab")
brain.AddPattern("ti-learning-lab", pattern)
brain.UpdateSessionMemory("ti-learning-lab", memory)
```

---

## 🎯 Recommendation

**Option 2: Ti Brain Centralized (Simplified)**

**Lý do:**
1. Đơn giản hơn - không cần sync
2. Consistent với Ti architecture hiện tại
3. Ti Brain đã có infrastructure
4. Agent chỉ cần gọi API, không cần biết về .devin
5. Auto-update đã có trong Ti Brain

**Implementation:**
1. Move `.devin/` content vào `Z:\Ti\brain\projects\ti-learning-lab/`
2. Add API endpoints to Ti Brain để quản lý project-specific knowledge
3. Agent chỉ cần gọi `brain.GetProjectContext("ti-learning-lab")`
4. Ti Brain Smart Auto-Update sẽ tự động detect changes

---

## 📋 Implementation Plan

### Phase 1: Move .devin to Ti Brain
```bash
# Move .devin to Ti Brain
mv /z/Ti/Ti-learning-lab/.devin /z/Ti/brain/projects/ti-learning-lab

# Update Ti Brain README
# Add ti-learning-lab to projects/ section
```

### Phase 2: Add Project API to Ti Brain
```go
// Z:\Ti\pkg\brain\project.go
func GetProjectContext(project string) (string, error)
func AddPattern(project, pattern string) error
func AddLesson(project, lesson string) error
func UpdateSessionMemory(project, memory string) error
```

### Phase 3: Update Agent Workflow
```go
// Agent chỉ cần gọi:
context := brain.GetProjectContext("ti-learning-lab")
patterns := brain.GetPatterns("ti-learning-lab")
memory := brain.GetSessionMemory("ti-learning-lab")

// Update:
brain.AddPattern("ti-learning-lab", pattern)
brain.AddLesson("ti-learning-lab", lesson)
brain.UpdateSessionMemory("ti-learning-lab", memory)
```

---

## 🔄 Workflow

### Agent Workflow (Simplified):
```
1. Agent bắt đầu task
2. Agent gọi: brain.GetProjectContext("ti-learning-lab")
3. Agent research, học patterns
4. Agent gọi: brain.AddPattern("ti-learning-lab", pattern)
5. Agent hoàn thành task
6. Agent gọi: brain.UpdateSessionMemory("ti-learning-lab", memory)
```

### Ti Workflow (Automated):
```
1. Agent gọi API
2. Ti lưu vào Z:\Ti\brain\projects\ti-learning-lab\
3. Ti Smart Auto-Update detect changes
4. Ti auto-update indexes
5. Ti log beads
```

---

## ✅ Benefits

### 1. **Simplicity**
- Agent không cần biết về .devin structure
- Chỉ cần gọi brain API
- Consistent interface

### 2. **Centralized Management**
- Ti quản lý toàn bộ knowledge
- Không cần sync thủ công
- Auto-update tự động

### 3. **Consistency**
- Same API cho tất cả projects
- Same structure
- Same workflow

### 4. **Scalability**
- Dễ thêm project mới
- Dễ scale lên nhiều projects
- Dễ maintain

---

## 🎯 Recommendation

**Option 2: Ti Brain Centralized (Simplified)**

**Steps:**
1. Move `.devin/` → `Z:\Ti\brain\projects\ti-learning-lab/`
2. Add project API to Ti Brain
3. Update agent workflow
4. Delete sync script (không cần)
5. Update documentation

**Next**: Bạn muốn tôi implement Option 2 không?

---

**Last Updated**: 2026-04-28
