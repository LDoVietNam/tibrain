# Ti Learning Lab - Workflow Guide

> **Last Updated**: 2026-05-06
> **Purpose**: Hướng dẫn khi nào dùng stage nào trong pipeline

---

## 🎯 Pipeline Overview

```
Learning → Research → Planning → Taskboard → HandsOn → Production
```

---

## 📋 Decision Tree: Khi Nào Dùng Stage Nào?

### Scenario 1: Học Pattern Mới (Simple)

```
Có cần deep research không?
├─ Không → Learning → HandsOn → Taskboard (log)
└─ Có → Learning → Research → Planning → Taskboard → HandsOn
```

**Example:**
- Pattern đơn giản: "Cách dùng React hooks"
- **Workflow**: Learning → HandsOn → Taskboard

---

### Scenario 2: Implement Feature Mới (Complex)

```
Có cần research không?
├─ Không → Planning → Taskboard → HandsOn
└─ Có → Learning → Research → Planning → Taskboard → HandsOn
```

**Example:**
- Feature mới: "Thêm MCP integration vào CLI"
- **Workflow**: Learning → Research → Planning → Taskboard → HandsOn

---

### Scenario 3: Nghiên Cứu Công Nghệ Mới

```
Cần hiểu sâu về công nghệ?
├─ Không → Research → Planning → Taskboard
└─ Có → Learning → Research → Planning → Taskboard
```

**Example:**
- Nghiên cứu: "So sánh Go vs Rust cho CLI"
- **Workflow**: Learning → Research → Planning → Taskboard

---

### Scenario 4: Fix Bug (Quick)

```
Bug có cần research không?
├─ Không → HandsOn → Taskboard (log)
└─ Có → Research → HandsOn → Taskboard
```

**Example:**
- Bug đơn giản: "Fix typo in README"
- **Workflow**: HandsOn → Taskboard

---

### Scenario 5: Refactor Code

```
Refactor có cần plan không?
├─ Không → HandsOn → Taskboard
└─ Có → Planning → Taskboard → HandsOn
```

**Example:**
- Refactor lớn: "Refactor CLI plugin architecture"
- **Workflow**: Planning → Taskboard → HandsOn

---

## 🔄 Stage Definitions

### 01_Learning - Học Tài Liệu

**Khi nào dùng:**
- Cần học concept mới, framework, pattern, best practice
- Đọc tutorials, guides, documentation
- Nắm kiến thức nền tảng

**Output:**
- Notes, summaries, key takeaways
- Link đến resources đã đọc
- Không phải code chạy

**Duration:**
- Short: 1-2 hours (simple pattern)
- Long: 1-2 days (complex framework)

**Kết thúc khi:**
- Đã hiểu concept
- Đã viết notes
- Có thể giải thích cho người khác

---

### 02_Research - Nghiên Cứu Sâu

**Khi nào dùng:**
- Cần đánh giá giải pháp (A vs B)
- Cần benchmark performance
- Cần proof-of-concept trước khi quyết định
- Cần investigation cho problem phức tạp

**Output:**
- Analysis report
- Recommendation
- Decision record
- Test results/data

**Duration:**
- Short: 2-4 hours (simple comparison)
- Long: 1-3 days (deep investigation)

**Kết thúc khi:**
- Có recommendation rõ ràng
- Có data/evidence support
- Có thể đưa ra decision

---

### 03_Knowledge - Tổng Hợp Kiến Thức

**Khi nào dùng:**
- Tổng hợp từ Learning + Research thành kiến thức có cấu trúc
- Tạo patterns, runbooks, decision logs
- Lưu reference materials (API docs, architecture specs)

**Output:**
- Pattern docs
- Runbooks
- Decision logs
- Cheatsheets

**Duration:**
- Ongoing - cập nhật khi có kiến thức mới

**Kết thúc khi:**
- Kiến thức có cấu trúc
- Dễ tra cứu
- Dễ onboard người mới

---

### 04_Planning - Lập Kế Hoạch

**Khi nào dùng:**
- Sau khi research xong, cần breakdown task
- Cần định nghĩa acceptance criteria
- Cần estimate timeline/resources
- Cần roadmap cho large project

**Output:**
- Plan với acceptance criteria
- Timeline estimate
- Resource requirements
- Task breakdown

**Duration:**
- Short: 1-2 hours (small feature)
- Long: 1-2 days (large enhancement)

**Kết thúc khi:**
- Tasks được breakdown rõ ràng
- Acceptance criteria được định nghĩa
- Timeline được estimate

---

### 05_Taskboard - Theo Dõi Nhiệm Vụ

**Khi nào dùng:**
- Trong quá trình làm task thực tế
- Log progress theo beads protocol
- Track status của tasks

**Output:**
- beads.md (log entries)
- Task progress tracking

**Duration:**
- Ongoing - cập nhật khi làm task

**Kết thúc khi:**
- Task được log
- Status được cập nhật

---

### 06_HandsOn - Thực Hành

**Khi nào dùng:**
- Áp dụng Learning + Research vào code thực tế
- POC, prototype, spike
- Test solution trước khi merge vào production

**Output:**
- Code chạy được
- Proof-of-concept
- Demo

**Duration:**
- Short: 2-4 hours (simple POC)
- Long: 1-3 days (complex prototype)

**Kết thúc khi:**
- Code chạy được
- Có demo/POC
- Có thể đưa ra decision

---

## 🎯 Workflow Examples

### Example 1: Học MCP Integration (Simple)

```
1. Learning: Đọc docs về MCP
   - 01_Learning/MCPHub/integration.md
   - Notes: key concepts, API endpoints

2. HandsOn: Tạo POC
   - 06_HandsOn/mcp-poc/
   - Implement basic MCP server

3. Taskboard: Log progress
   - 05_Taskboard/beads.md
   - Log: "MCP POC completed"
```

**Duration:** 1 day

---

### Example 2: Router Enhancement (Complex)

```
1. Learning: Học routing patterns
   - 01_Learning/Router/patterns.md
   - Notes: load balancing, circuit breaker

2. Research: Analyze current router
   - 02_Research/Router/analysis.md
   - Benchmark current performance
   - Recommendation: Use RTK compression

3. Planning: Tạo enhancement plan
   - 04_Planning/Router/enhancement-plan.md
   - Breakdown tasks
   - Estimate timeline: 2 weeks

4. Taskboard: Create tasks
   - 05_Taskboard/beads.md
   - Log: "Router enhancement started"

5. HandsOn: Implement
   - 06_HandsOn/router-enhancement/
   - Implement RTK compression
   - Test performance

6. Taskboard: Update progress
   - Log: "Router enhancement completed"
```

**Duration:** 2-3 weeks

---

### Example 3: Research Tech Stack (Research-First)

```
1. Research: Compare Go vs Rust
   - 02_Research/TechStack/go-vs-rust.md
   - Benchmark: performance, memory, compile time
   - Recommendation: Go for CLI

2. Learning: Học Go patterns
   - 01_Learning/Go/patterns.md
   - Notes: idiomatic Go, best practices

3. Planning: Tạo migration plan
   - 04_Planning/Migration/go-migration.md
   - Breakdown tasks
   - Estimate timeline: 1 month

4. Taskboard: Create tasks
   - 05_Taskboard/beads.md
   - Log: "Go migration started"
```

**Duration:** 1-2 months

---

## 📊 Quick Reference

| Scenario | Workflow | Duration |
|----------|----------|----------|
| Learn simple pattern | Learning → HandsOn → Taskboard | 1 day |
| Implement new feature | Learning → Research → Planning → Taskboard → HandsOn | 1-2 weeks |
| Research tech stack | Learning → Research → Planning → Taskboard | 1-2 days |
| Fix simple bug | HandsOn → Taskboard | 1 hour |
| Refactor large code | Planning → Taskboard → HandsOn | 1-2 weeks |
| POC new tech | Learning → Research → HandsOn | 2-3 days |

---

## 🚀 Best Practices

1. **Linear vs Non-Linear**
   - Linear: Learning → Research → Planning → Taskboard → HandsOn (standard)
   - Non-Linear: Có thể skip stages nếu không cần

2. **Link Stages**
   - Luôn link giữa stages (Learning → Research → Planning → Taskboard → HandsOn)
   - Dùng relative paths để dễ navigate

3. **Update MANIFEST.md**
   - Sau khi hoàn thành task, update MANIFEST.md
   - Track status, last updated

4. **Log với bd tool**
   - Dùng `bd log` command (không edit manual beads.md)
   - Agent name: `devin`
   - Status: `running`, `complete`, `failed`, `blocked`

5. **Review Regularly**
   - Review docs trong 03_Knowledge/ hàng tháng
   - Deprecate docs không còn dùng
   - Update docs khi có kiến thức mới

---

*Last Updated: 2026-05-06*
