# Rules Index - Danh Mục Rules

> **Version**: 3.0.0
> **Last Updated**: 2026-05-09
> **Purpose**: Danh mục tất cả rules cho AI agents
> **Location**: `Z:\10_WORKPLACE\Ti\content\rules\`

---

## 📋 Tổng Quan

**Total Rules**: 21 rules
- 5 Core Workflow Rules
- 4 Quality Rules
- 5 QA System Rules
- 1 Classification Rule
- 6 NEW Pattern Rules (from Cline)

**Ngôn ngữ**: Tiếng Việt (trừ thuật ngữ kỹ thuật)

---

## 🎯 Các Rules Theo Category

### Core Workflow Rules (5 rules)

|| # | Rule | File | Mô tả | Priority |
||---|------|------|-------|----------|
|| 1 | **Exploration** | `core/exploration.md` | Khám phá codebase, trace execution paths, thu thập context | P0 |
|| 2 | **Planning** | `core/planning.md` | Breakdown task thành beads, chọn execution path, select patterns | P0 |
|| 3 | **Implementation** | `core/implementation.md` | Incremental slices, simplicity first, scope discipline, test per slice | P0 |
|| 4 | **Review** | `core/review.md` | Five-axis review (Correctness, Readability, Architecture, Security, Performance) | P1 |
|| 5 | **Beads** | `core/beads.md` | Log checkpoints, track decisions, handoff giữa sessions/agents | P0 |

---

### Quality Rules (4 rules)

|| # | Rule | File | Mô tả | Priority |
||---|------|------|-------|----------|
|| 6 | **Mandatory Workflow Usage** | `quality/mandatory-workflow-usage.md` | BẮT BUỘC sử dụng workflow-orchestrator skill | P0 |
|| 7 | **Mandatory Quality Principles** | `quality/mandatory-quality-principles.md` | Context Collection, Decision Heuristics, Assumption Documentation | P0 |
|| 8 | **Mandatory Workflow Selection** | `quality/mandatory-workflow-selection.md` | Workflow selection criteria, classification factors | P0 |
|| 9 | **Mandatory Quality Gates** | `quality/mandatory-quality-gates.md` | Assumption Verification, Rollback, Lessons Learned Logging | P0 |

---

### QA System Rules (5 rules)

||| # | Rule | File | Mô tả | Priority |
|||---|------|------|-------|----------|
||| 10 | **Task Classification** | `task-classification.md` | Classify task type (bug-fix, feature, refactor) | P0 |
||| 11 | **Pre-Flight Checklist** | `pre-flight-checklist.md` | Task understanding, context collection, risk assessment (23 items) | P0 |
||| 12 | **Post-Flight Checklist** | `post-flight-checklist.md` | Deliverables verification, quality verification (27 items) | P0 |
||| 13 | **Quality Metrics Dashboard** | `quality-metrics-dashboard.md` | Track quality metrics over time | P1 |
||| 14 | **Auto-Skill Invocation** | `auto-skill-invocation.md` | Auto-load skills based on task type and language | P1 |

---

### Classification Rules (1 rule)

||| # | Rule | File | Mô tả | Priority |
|||---|------|------|-------|----------|
||| 15 | **Classification Rules** | `classification-rules.md` | 10 rules để tránh lỗi phân loại sai | P0 |

---

### 🆕 Pattern Rules from Cline (6 rules - NEW in v3.0.0)

||| # | Rule | File | Mô tả | Priority |
|||---|------|------|-------|----------|
||| 16 | **Tribal Knowledge** | `tribal-knowledge.md` | Khi nào nên add rules, patterns non-obvious | P0 |
||| 17 | **Writing Principles** | `writing-principles.md` | ⭐⭐⭐ Nguyên tắc viết docs tốt (evidence thay vì adjectives) | P0 |
||| 18 | **Network Patterns** | `network.md` | Proxy-aware network patterns, HTTP_PROXY support | P1 |
||| 19 | **Storage Architecture** | `storage-architecture.md` | Atomic writes, in-memory cache patterns | P1 |
||| 20 | **Hook System** | `hooks/README.md` | Hook system patterns (PreToolUse, PostToolUse, etc.) | P1 |
||| 21 | **Other Rules** | `*.md` | Existing rules | Various |

---

## 🚀 Workflow Chuẩn Trước Khi Bắt Đầu Task

### 1. P0 Core Rules (Bắt buộc)
- Đọc `core/exploration.md` để biết cách explore codebase
- Đọc `core/planning.md` để biết cách breakdown task
- Đọc `core/implementation.md` để biết cách implement theo slices
- Đọc `core/beads.md` để biết cách log progress và decisions

### 2. P0 Quality Rules (Bắt buộc)
- Đọc `quality/mandatory-workflow-usage.md`
- Đọc `quality/mandatory-quality-principles.md`
- Đọc `quality/mandatory-workflow-selection.md`
- Đọc `quality/mandatory-quality-gates.md`

### 3. P0 QA System Rules (Bắt buộc)
- Đọc `task-classification.md`
- Đọc `pre-flight-checklist.md`
- Đọc `post-flight-checklist.md`

### 4. 🆕 P0 Pattern Rules (Bắt buộc - NEW)
- Đọc `tribal-knowledge.md` - Biết khi nào add rules
- Đọc `writing-principles.md` - ⭐⭐⭐ Nguyên tắc viết docs

---

## 📂 Rules Theo Use Case

### Khi Bắt Đầu Task Mới
- `task-classification.md`
- `pre-flight-checklist.md`
- `core/exploration.md`
- `core/planning.md`
- `tribal-knowledge.md`
- `writing-principles.md`

### Khi Viết Documentation
- `writing-principles.md` ⭐⭐⭐
  - Use evidence, not adjectives
  - Claim must pass "Three Tests"
  - Keep it brief - devs don't waste time

### Khi Gặp Blocker / Cần Handoff
- `core/beads.md`
- `tribal-knowledge.md` (add to rules if pattern discovered)

### Khi Review / Trước Khi Hoàn Thành
- `post-flight-checklist.md`
- `quality-metrics-dashboard.md`
- `core/review.md`

### Khi Select Workflow
- `quality/mandatory-workflow-selection.md`

---

## 🔍 Tìm Kiếm Rules

### Theo Priority
```bash
# P0 Core Rules
cat Z:\10_WORKPLACE\Ti\content\rules\core\exploration.md
cat Z:\10_WORKPLACE\Ti\content\rules\core\planning.md
cat Z:\10_WORKPLACE\Ti\content\rules\core\implementation.md
cat Z:\10_WORKPLACE\Ti\content\rules\core\beads.md

# P0 Quality Rules
cat Z:\10_WORKPLACE\Ti\content\rules\quality\mandatory-workflow-usage.md
cat Z:\10_WORKPLACE\Ti\content\rules\quality\mandatory-quality-principles.md
cat Z:\10_WORKPLACE\Ti\content\rules\quality\mandatory-workflow-selection.md
cat Z:\10_WORKPLACE\Ti\content\rules\quality\mandatory-quality-gates.md

# P0 QA System Rules
cat Z:\10_WORKPLACE\Ti\content\rules\task-classification.md
cat Z:\10_WORKPLACE\Ti\content\rules\pre-flight-checklist.md
cat Z:\10_WORKPLACE\Ti\content\rules\post-flight-checklist.md

# 🆕 P0 Pattern Rules (NEW)
cat Z:\10_WORKPLACE\Ti\content\rules\tribal-knowledge.md
cat Z:\10_WORKPLACE\Ti\content\rules\writing-principles.md

# P1 Rules
cat Z:\10_WORKPLACE\Ti\content\rules\core\review.md
cat Z:\10_WORKPLACE\Ti\content\rules\quality-metrics-dashboard.md
cat Z:\10_WORKPLACE\Ti\content\rules\auto-skill-invocation.md
```

### Theo Category
```bash
# Core Workflow Rules
ls Z:\10_WORKPLACE\Ti\content\rules\core\

# Quality Rules
ls Z:\10_WORKPLACE\Ti\content\rules\quality\

# Pattern Rules (NEW)
ls Z:\10_WORKPLACE\Ti\content\rules\hooks\
```

---

## 📊 Thống Kê

### v3.0.0 (2026-05-09) - NEW

|| Category | Số Lượng | Files |
||----------|----------|-------|
|| **Core Workflow Rules** | 5 | exploration.md, planning.md, implementation.md, review.md, beads.md |
|| **Quality Rules** | 4 | mandatory-workflow-usage.md, mandatory-quality-principles.md, mandatory-workflow-selection.md, mandatory-quality-gates.md |
|| **QA System Rules** | 5 | task-classification.md, pre-flight-checklist.md, post-flight-checklist.md, quality-metrics-dashboard.md, auto-skill-invocation.md |
|| **Classification Rules** | 1 | classification-rules.md |
|| **🆕 Pattern Rules (from Cline)** | 6 | tribal-knowledge.md, writing-principles.md, network.md, storage-architecture.md, hooks/README.md, existing rules |
|| **Tổng cộng** | **21+** | - |

---

## 📝 Writing Principles Quick Reference (⭐⭐⭐)

Extract from `writing-principles.md`:

### ❌ WRONG - Adjectives
- "blazingly fast"
- "production ready"
- "enterprise-grade"

### ✅ RIGHT - Evidence
- "200 times faster"
- "scaling to 100 servers or 1 million documents per second"
- "30% faster recommendation system"

### Three Tests for Great Claims
1. **Visualized** - Can you picture it?
2. **Proven False** - Is it falsifiable?
3. **Only You** - Can only you say it?

---

## 🔗 Related Files

- **Agent Guidelines**: `Z:\10_WORKPLACE\Ti\content\agent-guidelines.md`
- **Workflows**: `Z:\10_WORKPLACE\Ti\content\workflows\INDEX.md`
- **Skills**: `Z:\10_WORKPLACE\Ti\content\skills\`
- **Beads Project**: https://github.com/gastownhall/beads

---

## 🔄 Recent Updates

### v3.0.0 (2026-05-09) - ⭐ BIG UPDATE
- Added **6 NEW Pattern Rules** from Cline:
  - `tribal-knowledge.md` - When/how to add tribal knowledge
  - `writing-principles.md` - ⭐⭐⭐ Evidence-based writing
  - `network.md` - Proxy-aware network patterns
  - `storage-architecture.md` - Atomic writes, cache patterns
  - `hooks/README.md` - Hook system patterns
- Updated total rules: 15 → **21+**
- Added "Writing Principles Quick Reference" section
- Updated Workflow Chuẩn to include new P0 rules
- Added category "Pattern Rules from Cline"

### v2.1.0 (2026-06-05)
- Added QA System Rules category (5 rules)
- Updated total rules: 9 → 15

### v2.0.0 (2026-05-05)
- Added Quality Rules category (4 rules)
- Updated total rules: 5 → 9

---

*Last Updated: 2026-05-09*
