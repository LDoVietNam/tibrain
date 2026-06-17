# Ti Learning Lab - Learning & Development Hub

**Location**: `Z:\10_WORKPLACE\Ti\Ti-learning-lab\`

---

## 🎯 Mục Đích

**Xây dựng hệ thống learning pipeline có cấu trúc để chuyển đổi kiến thức thụ động thành năng lực thực tế.**

### Pipeline
```
Learning → Research → Planning → Taskboard → HandsOn → Production
```

### Nguyên Tắc
1. **Actionable Knowledge** - Mọi tài liệu phải có thể áp dụng vào thực tế
2. **Verified Knowledge** - Kiến thức phải được test trong real projects
3. **Connected Knowledge** - Các stages phải có liên kết rõ ràng
4. **Quality over Quantity** - Chỉ giữ lại tài liệu có giá trị thực sự

Đây là nơi **học → nghiên cứu → lập kế hoạch → thực hiện → archive** theo workflow có cấu trúc.

---

## 📂 Cấu Trúc

```
Ti-learning-lab/
├── 01_Learning/              # Learning materials (tutorials, guides, patterns để HỌC)
│   ├── BrowserAutomation/    # Browser automation learning
│   ├── CLI/                  # CLI learning
│   ├── Configs/              # Configuration learning
│   ├── Core/                 # Core learning
│   ├── MCPHub/               # MCP Hub learning
│   ├── TiBrain/              # TiBrain learning
│   └── [project-name]/       # Individual project learning
├── 03_Knowledge/             # Reference materials (API docs, architecture, technical specs - KHÔNG phải để học)
│   ├── 00_META/              # Metadata (AUDIT_REPORT, CROSS_REFERENCE_EXAMPLE)
│   ├── agents/               # Agent framework
│   ├── cli/                  # CLI patterns
│   ├── patterns/             # Software patterns
│   ├── lessons/              # Lessons learned
│   └── [category]/           # Other categories
├── 05_Taskboard/             # Task tracking, beads
│   ├── beads.md             # Beads log
│   ├── storage/             # Schema and metadata files
│   └── tasks/               # Task definitions
├── 06_HandsOn/              # Thực hành (POC, prototype, spike)
│   ├── codex-autoresearch/  # Codex auto-research project
│   └── [project-name]/      # Individual projects
├── 07_Repositories/          # Repository references
│   ├── storage/             # Zip archives
│   ├── undetectable-fingerprint-browser/ # Browser fingerprint research
│   └── [repo-name]/         # Individual repos
├── 08_Archives/              # Archived projects
│   └── [archived-project]/  # Archived projects
├── templates/               # Templates cho stages
│   ├── 01_Learning_Template.md
│   ├── 06_HandsOn_Template.md
│   └── README.md
├── WORKFLOW.md              # Workflow guide
├── SUBFOLDER_AUDIT_REPORT.md # Subfolder audit report
├── AGENTS.md                 # Agent guidelines (root)
├── README.md                 # Documentation (this file)
├── MANIFEST.md               # Project manifest (metadata, tracking)
└── .devin/                   # Devin config

**Notes:**
- 📋 05_Taskboard/beads.md is copied from Z:\10_WORKPLACE\taskboard\beads.md (manual sync required)
- 🗑️ 02_Research/, 04_Planning/, 09_Specialized/ đã xóa (trống)
```

---

## 📚 Tài Liệu Hỗ Trợ

| Tài Liệu | Mô Tả | Location |
|----------|-------|----------|
| **AGENTS.md** | Rules cho agents khi làm việc trong project | [AGENTS.md](./AGENTS.md) |
| **AGENT_AUTO_LOAD_GUIDE.md** | Hướng dẫn cách agent tự động load rules | [AGENT_AUTO_LOAD_GUIDE.md](./AGENT_AUTO_LOAD_GUIDE.md) |
| **WORKFLOW.md** | Hướng dẫn workflow - khi nào dùng stage nào | [WORKFLOW.md](./WORKFLOW.md) |
| **templates/** | Templates cho các stages (Learning, HandsOn) | [templates/](./templates/) |
| **AUDIT_REPORT.md** | Báo cáo audit chất lượng content 03_Knowledge/ | [03_Knowledge/00_META/AUDIT_REPORT.md](./03_Knowledge/00_META/AUDIT_REPORT.md) |
| **CROSS_REFERENCE_EXAMPLE.md** | Example cách tạo cross-references giữa stages | [03_Knowledge/00_META/CROSS_REFERENCE_EXAMPLE.md](./03_Knowledge/00_META/CROSS_REFERENCE_EXAMPLE.md) |
| **SUBFOLDER_AUDIT_REPORT.md** | Báo cáo audit subfolder cleanup | [SUBFOLDER_AUDIT_REPORT.md](./SUBFOLDER_AUDIT_REPORT.md) |

---

## 🔄 Workflow Học Tập & Phát Triển

```
┌─────────────────────────────────────────────────────────────┐
│                     Workflow Pipeline                        │
└─────────────────────────────────────────────────────────────┘

1. Learning (01_Learning/)
   ↓ Học patterns, best practices, guides
   → Đọc tài liệu, hiểu patterns
   → Lưu notes vào appropriate subfolders

2. Research (02_Research/)
   ↓ Nghiên cứu sâu, analysis, investigation
   → Tạo analysis reports
   → Document findings
   → Validate assumptions

3. Planning (04_Planning/)
   ↓ Lập kế hoạch từ research findings
   → Tạo enhancement plans
   → Viết proposals
   → Định nghĩa roadmaps

4. Taskboard (05_Taskboard/)
   ↓ Tạo tasks từ plans
   → Log vào beads.md
   → Track progress
   → Assign to agents

5. Projects (06_Projects/)
   ↓ Thực hiện development
   → Learning experiments
   → Active development
   → Production ready

6. Archives (08_Archives/)
   ↓ Archive khi hoàn thành
   → Backup completed projects
   → Keep for reference
```

---

## 🔄 Non-Linear Workflow Scenarios

Linear workflow không phù hợp với mọi scenarios. Dưới đây là các non-linear workflows được support:

### Scenario 1: Iterative Learning & Research

```
Learning (01_Learning/)
   ↓ Học pattern mới
Research (02_Research/)
   ↓ Nghiên cứu sâu
Learning (01_Learning/) ← BACK TO LEARNING
   ↓ Học thêm từ research findings
Research (02_Research/)
   ↓ Validate với real data
Planning (04_Planning/)
```

**Use case:** Khi research findings reveal gaps trong knowledge cần học thêm.

### Scenario 2: Parallel Activities

```
Learning (01_Learning/) ──────┐
   ↓ Học MCP integration        │ PARALLEL
Research (02_Research/) ──────┤
   ↓ Analyze current state      │
Planning (04_Planning/) ←──────┘
   ↓ Combine findings
```

**Use case:** Khi learning và research có thể làm song song để save time.

### Scenario 3: Multiple Projects from Single Plan

```
Planning (04_Planning/)
   ↓ ROUTER_ENHANCEMENT_PLAN.md
Taskboard (05_Taskboard/)
   ↓ Create multiple task groups
Projects (06_Projects/) ──────┬─→ learning/router-phase1/
                              ├─→ active/router-phase2/
                              └─→ learning/router-phase3/
```

**Use case:** Khi một plan lớn chia thành multiple projects có thể chạy parallel hoặc sequential.

### Scenario 4: Research-First Approach

```
Research (02_Research/)
   ↓ Analyze problem space
Planning (04_Planning/)
   ↓ Create research plan
Research (02_Research/)
   ↓ Execute research
Learning (01_Learning/) ← LEARN FROM RESEARCH
   ↓ Document patterns
Planning (04_Planning/)
   ↓ Create implementation plan
```

**Use case:** Khi cần deep research trước khi học patterns (complex domains).

### Scenario 5: Quick Learning → Direct Implementation

```
Learning (01_Learning/)
   ↓ Learn simple pattern
Projects (06_Projects/)
   ↓ Direct implementation (skip research/planning)
Taskboard (05_Taskboard/)
   ↓ Log tasks
```

**Use case:** Khi pattern đơn giản, không cần deep research hoặc formal planning.

### Scenario 6: Planning → Learning Loop

```
Planning (04_Planning/)
   ↓ Identify knowledge gaps
Learning (01_Learning/)
   ↓ Learn missing skills
Planning (04_Planning/) ← UPDATE PLAN
   ↓ Refine plan with new knowledge
Projects (06_Projects/)
```

**Use case:** Khi planning reveals cần học thêm skills trước khi implement.

---

## 🎯 Workflow Selection Guide

| Scenario | Recommended Workflow | When to Use |
|----------|---------------------|-------------|
| **Standard feature** | Linear (Learning → Research → Planning → Task → Project) | New features requiring research |
| **Simple fix** | Quick Learning → Direct Implementation | Bug fixes, small changes |
| **Complex domain** | Research-First Approach | New domains, unfamiliar tech |
| **Knowledge gaps** | Planning → Learning Loop | Plan reveals missing skills |
| **Large project** | Multiple Projects from Single Plan | Large enhancement with phases |
| **Exploratory** | Iterative Learning & Research | R&D, experimental features |
| **Time-critical** | Parallel Activities | Need to save time, independent tasks |

---

---

## 📋 MANIFEST.md - Project Manifest

**Purpose**: Track metadata, progress, and status của tất cả activities trong Ti-learning-lab.

**Location**: `Z:\10_WORKPLACE\Ti\Ti-learning-lab\MANIFEST.md`

**Structure**:
```markdown
# Ti Learning Lab Manifest

> **Last Updated**: YYYY-MM-DD
> **Version**: 1.0

---

## Learning Materials (01_Learning/)

| Category | File | Status | Last Updated |
|----------|------|--------|--------------|
| agents | AGENTS.md | ✅ Active | 2026-04-30 |
| workflows | AI_WORKFLOW_BEST_PRACTICES.md | ✅ Active | 2026-04-30 |
| integration | CLAUDE_WINDSURF_INTEGRATION.md | ✅ Active | 2026-04-30 |
| performance | PERFORMANCE_TUNING.md | ✅ Active | 2026-04-30 |

---

## Research (02_Research/)

| Category | Topic | Status | Last Updated |
|----------|-------|--------|--------------|
| analysis | [topic] | 🔄 In Progress | YYYY-MM-DD |
| investigation | [topic] | ⏳ Pending | YYYY-MM-DD |

---

## Planning (04_Planning/)

| Type | Plan | Status | Priority | Last Updated |
|------|------|--------|----------|--------------|
| enhancement | ROUTER_ENHANCEMENT_PLAN.md | 🔄 In Progress | High | 2026-04-30 |
| proposal | WORKFLOW_PROPOSAL.md | ✅ Approved | Medium | 2026-04-30 |

---

## Projects (06_Projects/)

| Project | Type | Status | Last Updated |
|---------|------|--------|--------------|
| [project-name] | learning/active | 🔄 In Progress | YYYY-MM-DD |

---

## Taskboard (05_Taskboard/)

| Last Beads Entry | Status | Date |
|-------------------|--------|------|
| [entry title] | DONE | YYYY-MM-DD |

---

## Statistics

- **Total Learning Materials**: X
- **Active Research Topics**: X
- **Pending Plans**: X
- **Active Projects**: X
- **Completed Tasks**: X
```

---

## 🎯 Quy Tắc Sử Dụng

### Learning (01_Learning/)
- ✅ NÊN: Lưu patterns, best practices, guides học được
- ✅ NÊN: Organize theo category (agents, workflows, integration, performance)
- ❌ KHÔNG: Lưu temporary notes (dùng 02_Research/ hoặc 04_Planning/)

### Research (02_Research/)
- ✅ NÊN: Document analysis, investigation findings
- ✅ NÊN: Link đến learning materials (01_Learning/)
- ❌ KHÔNG: Lưu plans (dùng 04_Planning/)

### Knowledge Base (03_Knowledge/)
- ✅ NÊN: Lưu reference materials (API docs, architecture specs, technical documentation)
- ✅ NÊN: Static documentation không thay đổi thường xuyên
- ❌ KHÔNG: Lưu learning materials (dùng 01_Learning/)
- ❌ KHÔNG: Lưu research findings (dùng 02_Research/)

**Boundary Clarification:**
- **01_Learning/** = Materials để HỌC (tutorials, guides, patterns you're learning)
- **03_Knowledge/** = Reference materials (API docs, architecture, specs you reference)

Example:
- "How to integrate MCP" → 01_Learning/integration/
- "MCP API specification" → 03_Knowledge/api/

### Planning (04_Planning/)
- ✅ NÊN: Tạo plans từ research findings
- ✅ NÊN: Link đến research (02_Research/)
- ✅ NÊN: Tạo tasks cho taskboard (05_Taskboard/)
- ❌ KHÔNG: Lưu learning materials (dùng 01_Learning/)

### Taskboard (05_Taskboard/)
- ✅ NÊN: Log tất cả activities theo BEADS protocol
- ✅ NÊN: Link đến plans (04_Planning/)
- ✅ NÊN: Update MANIFEST.md sau mỗi task
- ❌ KHÔNG: Lưu research hoặc plans

### Projects (06_Projects/)
- ✅ NÊN: Tạo projects từ tasks
- ✅ NÊN: Link đến taskboard entries
- ✅ NÊN: Move to production khi ready
- ❌ KHÔNG: Lưu learning materials (dùng 01_Learning/)

---

## 🔍 Structure Validation

Run validation script to check structure integrity:

```bash
cd Z:\10_WORKPLACE\Ti\Ti-learning-lab
bash validate-structure.sh
```

The script checks:
- Expected folders exist
- No unexpected folders
- MANIFEST.md integrity (State Legend present)
- README.md completeness (Non-linear workflows documented)
- Deprecated folders (06_Knowledge)

**Pre-commit Hook:**
- Automatic validation runs when committing Ti-learning-lab files
- Commit blocked if validation fails
- Ensures structure integrity before commits

---

## 🚀 Quick Start

### Học pattern mới
```bash
# 1. Đọc learning materials
cd Z:\10_WORKPLACE\Ti\Ti-learning-lab\01_Learning\workflows

# 2. Nghiên cứu sâu nếu cần
cd Z:\10_WORKPLACE\Ti\Ti-learning-lab\02_Research\analysis

# 3. Tạo plan nếu cần implement
cd Z:\10_WORKPLACE\Ti\Ti-learning-lab\04_Planning\enhancement-plans

# 4. Tạo tasks
cd Z:\10_WORKPLACE\Ti\Ti-learning-lab\05_Taskboard
# Edit beads.md

# 5. Implement trong project
cd Z:\10_WORKPLACE\Ti\Ti-learning-lab\06_Projects\learning
```

### Track progress
```bash
# Update manifest
cd Z:\10_WORKPLACE\Ti\Ti-learning-lab
# Edit MANIFEST.md

# Check taskboard
cat Z:\10_WORKPLACE\Ti\Ti-learning-lab\05_Taskboard\beads.md
```

---

## 📊 Workflow Examples

### Example 1: Học MCP Integration
```
1. Learning: Đọc 01_Learning/integration/SKILLS_INTEGRATION.md
2. Research: Tạo analysis report trong 02_Research/analysis/
3. Planning: Tạo plan trong 04_Planning/proposals/
4. Taskboard: Log tasks vào 05_Taskboard/beads.md
5. Projects: Implement trong 06_Projects/learning/mcp-integration-poc/
6. Archives: Archive khi hoàn thành vào 08_Archives/
```

### Example 2: Router Enhancement
```
1. Learning: Đọc 01_Learning/performance/PERFORMANCE_TUNING.md
2. Research: Analyze current router trong 02_Research/analysis/
3. Planning: Update ROUTER_ENHANCEMENT_PLAN.md trong 04_Planning/enhancement-plans/
4. Taskboard: Create tasks trong 05_Taskboard/beads.md
5. Projects: Implement trong 06_Projects/active/router-enhancement/
6. Archives: Archive khi hoàn thành
```

---

## 🔗 Liên Kết Với Hệ Thống Ti

|| Khi nào | Làm gì | Where |
||---------|--------|-------|
|| **Đọc agent guidelines** | Reference AGENTS.md | `Z:\10_WORKPLACE\Ti\Ti-learning-lab\AGENTS.md` hoặc `Z:\Ti\AGENTS.md` |
|| **Check taskboard** | Đọc beads log | `Z:\10_WORKPLACE\Ti\Ti-learning-lab\05_Taskboard\beads.md` hoặc `Z:\10_WORKPLACE\Ti\taskboard\beads.md` |
|| **Production ready** | Move project | `Z:\Router\` hoặc `Z:\Agents-hub\` |
|| **Research findings** | Reference | `Z:\Ti\knowledge_base\04_Research\` |

---

## 📞 Liên Hệ & Support

|| Vai trò | Location | Contact |
||---------|----------|--------|
|| Learning lead | `Z:\10_WORKPLACE\Ti\Ti-learning-lab\` | User |
|| Research team | `Z:\Ti\knowledge_base\04_Research\` | Jarvis |
|| Production team | `Z:\Router\`, `Z:\Agents-hub\` | Spectre |

---

## 📖 Related Docs

|| Doc | Location | Purpose |
||-----|----------|--------|
|| Ti AGENTS.md | `Z:\Ti\AGENTS.md` | Main agent guidelines |
|| Taskboard | `Z:\10_WORKPLACE\Ti\taskboard\beads.md` | Main task tracking |
|| Knowledge Base | `Z:\Ti\knowledge_base\README.md` | Main knowledge structure |

---

*Last Updated: 2026-04-30*
*Version: 3.0 - Workflow-based structure*
