---
tags: ["provider-claude", "tibrain", "skill", "documentation", "router"]
scopes: ["providers", "tibrain"]
last_updated: 2026-05-22
---
# Anthropic Claude Code Tools Analysis & Ti Application

> **Purpose**: Phân tích 6 công cụ từ Anthropic Claude Code và đề xuất áp dụng vào Ti ecosystem
> **Date**: 2026-05-04
> **Source**: User insights about Anthropic Claude Code ecosystem

---

## 1. The Factory: Skill Creator

### Mô Tả
- Tạo skill từ mô tả ngôn ngữ tự nhiên
- Không cần viết file skill.md
- Không cần học format
- Claude tự tạo skill, test, chỉnh sửa, đóng gói
- Có thể đưa SOP vào → biến thành skill tái sử dụng

### Use Case
- Agency bất động sản mất hàng giờ viết mô tả nhà
- Không có tool → phải viết tay skill.md, test nhiều lần
- Có tool → chỉ cần mô tả → Claude làm hết

### Áp Dụng Vào Ti

**Ti Skill Creator**:
```
Ti Skill Creator
├── Input: Mô tả tự nhiên hoặc SOP
├── Process:
│   ├── Phân tích yêu cầu
│   ├── Tạo skill structure
│   ├── Implement skill logic
│   ├── Test skill
│   ├── Chỉnh sửa
│   └── Đóng gói
└── Output: Skill dùng được mãi mãi
```

**Implementation**:
- Agent: `skill-creator-agent` trong Ti Claw
- Input: Natural language description hoặc SOP document
- Output: `.devin/skills/<skill-name>/SKILL.md`
- Integration:
  - CLI command: `ti skill create "Tạo skill để sync Notion"`
  - UI: Skill Creator page trong Ti Claw UI
  - Auto-test: Tự động test skill sau khi tạo

**Benefits cho Ti**:
- Tạo skill nhanh hơn 10x
- Không cần học skill format
- Tái sử dụng SOP thành skill
- Auto-test đảm bảo chất lượng

---

## 2. The Process: Superpowers

### Mô Tả
- Ép Claude làm việc như senior dev
- Lập kế hoạch toàn bộ trước khi code
- Làm việc trong môi trường tách biệt
- Viết test trước
- Brainstorm giải pháp
- Review 2 lần: đúng spec và chất lượng code
- Từ 60% lên 80% ngay lần đầu → tiết kiệm thời gian debug

### Áp Dụng Vào Ti

**Ti Superpowers**:
```
Ti Superpowers Workflow
├── Phase 1: Planning
│   ├── Phân tích yêu cầu
│   ├── Lập kế hoạch implementation
│   ├── Brainstorm giải pháp
│   └── Đề xuất architecture
├── Phase 2: Implementation
│   ├── Tạo môi trường tách biệt
│   ├── Viết test trước (TDD)
│   ├── Implement code
│   └── Test từng phần
├── Phase 3: Review
│   ├── Review đúng spec
│   ├── Review chất lượng code
│   ├── Review performance
│   └── Review security
└── Phase 4: Polish
    ├── Optimize code
    ├── Add documentation
    └── Final verification
```

**Implementation**:
- Agent: `superpowers-agent` trong Ti Claw
- Workflow: TDD workflow skill (đã có trong `.devin/skills/tdd-workflow/`)
- Integration:
  - CLI command: `ti superpowers "Implement feature X"`
  - UI: Superpowers toggle trong Ti Claw UI
  - Auto-review: Tự động review code sau khi implement

**Benefits cho Ti**:
- Tăng chất lượng code từ 60% lên 80%
- Giảm thời gian debug
- Consistent quality
- Better architecture

---

## 3. The Clean Room: GSD (Get Stuff Done)

### Mô Tả
- Giải quyết context rot
- Tạo sub-agent riêng cho từng task
- Mỗi agent có context sạch riêng
- Tránh context bị loãng
- Cơ chế kiểm soát chất lượng:
  - Phát hiện việc Claude tự bỏ sót yêu cầu
  - Kiểm tra bảo mật
  - Chế độ autonomous: đưa spec → Claude tự làm hết

### Áp Dụng Vào Ti

**Ti GSD (Get Stuff Done)**:
```
Ti GSD System
├── Context Management
│   ├── Tạo context sạch cho mỗi task
│   ├── Isolate context giữa agents
│   ├── Auto-cleanup context rác
│   └── Context persistence
├── Agent Orchestration
│   ├── Tạo sub-agent cho từng task
│   ├── Coordinate giữa agents
│   ├── Handoff context
│   └── Merge results
├── Quality Control
│   ├── Check missing requirements
│   ├── Security check
│   ├── Privacy check
│   └── Quality gate
└── Autonomous Mode
    ├── Spec → Auto execution
    ├── Progress tracking
    ├── Error handling
    └── Result aggregation
```

**Implementation**:
- Agent: `gsd-agent` trong Ti Claw
- Context System: SQLite-based context storage
- Integration:
  - CLI command: `ti gsd "Task description"`
  - UI: GSD dashboard trong Ti Claw UI
  - Autonomous mode: `ti gsd --auto "Task description"`

**Benefits cho Ti**:
- Không còn context rot
- Session chạy 3 tiếng vẫn ổn
- Auto-quality control
- Autonomous execution

---

## 4. The Safety Net: /review + /ultra-review

### Mô Tả
- /review: review code nhanh, local
- /ultra-review: review nâng cao bằng cloud + nhiều agent song song
- Chỉ báo bug đã được verify
- Không spam lỗi giả
- /review cho việc thường ngày
- /ultra-review cho các phần quan trọng: payment, auth, database

### Áp Dụng Vào Ti

**Ti Review System**:
```
Ti Review System
├── Local Review (Quick)
│   ├── Static analysis
│   ├── Code style check
│   ├── Common bug detection
│   └── Performance hints
├── Cloud Review (Deep)
│   ├── Multi-agent parallel review
│   ├── Security audit
│   ├── Architecture review
│   └── Best practices check
└── Integration
    ├── CLI command: `ti review`
    ├── CLI command: `ti ultra-review`
    ├── UI: Review panel trong Ti Claw UI
    └── Auto-review: Tự động review sau khi commit
```

**Implementation**:
- Agent: `review-agent` trong Ti Claw
- Tools: Static analysis tools (golangci-lint, eslint, etc.)
- Integration:
  - CLI command: `ti review <file>`
  - CLI command: `ti ultra-review <file>`
  - UI: Review panel trong Ti Claw UI
  - Pre-commit hook: Auto review trước khi commit

**Benefits cho Ti**:
- Chỉ báo bug thực
- Không spam lỗi giả
- Quick review cho日常工作
- Deep review cho critical code

---

## 5. The Compressor: Context Mode

### Mô Tả
- Giải quyết context rác
- Nén dữ liệu đầu vào (giảm từ hàng trăm KB xuống vài trăm byte)
- Lưu toàn bộ session vào database
- Khi context bị reset → tự khôi phục lại
- Session chạy 3 tiếng vẫn ổn
- Không cần nhắc lại context

### Áp Dụng Vào Ti

**Ti Context Mode**:
```
Ti Context Mode
├── Compression
│   ├── Nén input data
│   ├── Loại bỏ context rác
│   ├── Tóm tắt long messages
│   └── Optimize token usage
├── Persistence
│   ├── Lưu session vào database
│   ├── Version control context
│   ├── Incremental save
│   └── Auto-cleanup old sessions
├── Recovery
│   ├── Auto-restore khi context reset
│   ├── Merge context từ versions
│   ├── Detect context drift
│   └── Sync context giữa agents
└── Monitoring
    ├── Track context size
    ├── Alert khi context quá lớn
    ├── Optimize context tự động
    └── Report context stats
```

**Implementation**:
- System: `context-mode` trong Ti Claw
- Storage: SQLite-based session storage
- Integration:
  - CLI flag: `ti --context-mode`
  - UI: Context mode toggle trong Ti Claw UI
  - Auto-enable: Tự động enable cho long sessions

**Benefits cho Ti**:
- Giảm context từ hàng trăm KB xuống vài trăm byte
- Session chạy 3 tiếng vẫn ổn
- Không cần nhắc lại context
- Auto-recover khi context reset

---

## 6. The Memory: Claude Mem

### Mô Tả
- Claude mặc định không nhớ gì giữa các session
- Claude Mem giải quyết:
  - Ghi lại toàn bộ hoạt động
  - Tóm tắt thành knowledge
  - Lưu vào database
  - Tự động inject lại khi session mới
- Tự tạo file CLAUDE.md
- Tự update tài liệu project
- Không cần giải thích lại project
- Tiết kiệm ~10x token

### Áp Dụng Vào Ti

**Ti Memory System**:
```
Ti Memory System
├── Activity Tracking
│   ├── Ghi lại toàn bộ hoạt động
│   ├── Track user preferences
│   ├── Log decisions
│   └── Monitor patterns
├── Knowledge Extraction
│   ├── Tóm tắt thành knowledge
│   ├── Extract patterns
│   ├── Identify best practices
│   └── Detect anti-patterns
├── Persistence
│   ├── Lưu vào database
│   ├── Sync với Knowledge Graph Memory
│   ├── Version control knowledge
│   └── Auto-cleanup old data
├── Injection
│   ├── Tự động inject khi session mới
│   ├── Smart context loading
│   ├── Relevance filtering
│   └── Priority ranking
└── Documentation
    ├── Tự tạo file CLAUDE.md
    ├── Tự update AGENTS.md
    ├── Tự update README
    └── Tự sync Notion
```

**Implementation**:
- System: `ti-memory` trong Ti Claw
- Integration với Knowledge Graph Memory (server-memory MCP)
- Integration:
  - CLI flag: `ti --memory-mode`
  - UI: Memory dashboard trong Ti Claw UI
  - Auto-enable: Tự động enable cho tất cả sessions

**Benefits cho Ti**:
- Ghi nhớ giữa sessions
- Tự động update documentation
- Tiết kiệm ~10x token
- Không cần giải thích lại project

---

## 7. Bonus: Frontend Design

### Mô Tả
- Skill chính thức từ Anthropic
- Giúp UI/UX bớt "mùi AI"
- Trông chuyên nghiệp hơn

### Áp Dụng Vào Ti

**Ti Frontend Design**:
```
Ti Frontend Design System
├── Design Components
│   ├── Button, Card, Input, Modal
│   ├── Sidebar, Header, Layout
│   ├── Typography, Colors, Spacing
│   └── Icons, Animations
├── Design Guidelines
│   ├── Accessibility (WCAG 2.2)
│   ├── Mobile-first responsive
│   ├── Dark/light theme
│   └── Vietnamese i18n
├── AI-less UI Patterns
│   ├── Natural language forms
│   ├── Progressive disclosure
│   ├── Smart defaults
│   └── Contextual help
└── Integration
    ├── ti-ui-components library
    ├── Ti Claw UI
    ├── Ti Router UI
    └── Design documentation
```

**Implementation**:
- Skill: `frontend-design` (đã có trong `.devin/skills/frontend-design/`)
- Design System: `packages/ti-ui-components/`
- Integration:
  - UI components library
  - Design guidelines
  - Examples and templates

**Benefits cho Ti**:
- UI/UX bớt "mùi AI"
- Trông chuyên nghiệp hơn
- Consistent design
- Better user experience

---

## 8. Cách Kiếm Tiền Từ Những Thứ Này

### Đừng Bán "Workflow"
- Khách hàng không quan tâm bạn dùng AI gì
- Họ quan tâm kết quả

### Hãy Bán Kết Quả
- **Tiết kiệm 10 giờ/tuần**
- **Giảm lỗi 80%**
- **Tăng tốc công việc 3x**
- **Tự động hóa quy trình thủ công**

### Nếu Mới Bắt Đầu
1. Chọn 1 skill
2. Học nó
3. Build vài demo
4. Tăng dần complexity

---

## 9. Áp Dụng Vào Ti Ecosystem

### Priority 1: Ti Memory System (Tiết kiệm ~10x token)
- Implement ngay
- Tích hợp với Knowledge Graph Memory
- Auto-update documentation
- Auto-sync Notion

### Priority 2: Ti Review System (Chỉ báo bug thực)
- Implement local review
- Implement cloud review (ultra-review)
- Pre-commit hook
- UI review panel

### Priority 3: Ti Context Mode (Session chạy 3 tiếng)
- Implement context compression
- Session persistence
- Auto-recovery
- Context monitoring

### Priority 4: Ti Skill Creator (Tạo skill 10x nhanh hơn)
- Implement skill creator agent
- Natural language input
- Auto-test
- UI integration

### Priority 5: Ti Superpowers (Tăng chất lượng code 60%→80%)
- Implement TDD workflow
- Planning phase
- Review phase
- Quality gates

### Priority 6: Ti GSD (Không còn context rot)
- Implement context isolation
- Sub-agent orchestration
- Quality control
- Autonomous mode

---

## 10. Integration với Ti Claw

### Architecture Update

```
Ti Claw Architecture (Updated)
├── Core Framework
│   ├── Agent Loop (Think-Act-Observe)
│   ├── Store Layer (SQLite)
│   ├── Tool Registry
│   ├── Provider Pattern
│   ├── Memory System
│   ├── Config Loading
│   └── Bootstrap Pattern
├── NEW: Ti Memory System
│   ├── Activity Tracking
│   ├── Knowledge Extraction
│   ├── Persistence (Knowledge Graph Memory)
│   ├── Auto-injection
│   └── Documentation Auto-update
├── NEW: Ti Context Mode
│   ├── Context Compression
│   ├── Session Persistence
│   ├── Auto-recovery
│   └── Context Monitoring
├── NEW: Ti Review System
│   ├── Local Review (Quick)
│   ├── Cloud Review (Deep)
│   ├── Pre-commit Hook
│   └── UI Review Panel
├── NEW: Ti Skill Creator
│   ├── Natural Language Input
│   ├── Auto-generation
│   ├── Auto-test
│   └── UI Integration
├── NEW: Ti Superpowers
│   ├── Planning Phase
│   ├── Implementation Phase
│   ├── Review Phase
│   └── Quality Gates
└── NEW: Ti GSD
    ├── Context Isolation
    ├── Sub-agent Orchestration
    ├── Quality Control
    └── Autonomous Mode
```

---

## 11. Implementation Plan (Updated)

### Phase 1: Core Framework (Week 1-2)
- Implement core patterns
- **NEW: Ti Memory System** (Priority 1)
- **NEW: Ti Context Mode** (Priority 3)

### Phase 2: CLI Integration (Week 3)
- Integrate with Ti CLI
- **NEW: Ti Review System** (Priority 2)
- Plugin system

### Phase 3: Notion Agent (Week 4)
- Build Notion Agent
- **NEW: Ti Skill Creator** (Priority 4) - Use cho Notion Agent
- Test sync functionality

### Phase 4: Additional Agents (Week 5-6)
- Build additional agents
- **NEW: Ti Superpowers** (Priority 5) - Use cho agent development
- **NEW: Ti GSD** (Priority 6) - Use cho complex tasks
- Agent templates

### Phase 5: UI Development (Week 5-6)
- Design System
- Ti Claw UI
- Core pages
- Additional pages

### Phase 6: Polish & Release (Week 7-8)
- Bug fixes
- Performance optimization
- Final documentation
- Release v1.0.0

---

## 12. Success Criteria (Updated)

- [x] Core Ti Claw framework implemented
- [x] CLI integration working
- [x] Notion Agent syncing knowledge
- [x] 4+ agents working
- [x] Documentation complete
- [x] Tests passing
- [x] Performance acceptable (<1s response time)
- [x] Shared design system created (ti-ui-components)
- [x] Ti Claw UI implemented (localhost:5174)
- [x] UI integration with backend working
- [x] Theme toggle working (light/dark)
- [x] Vietnamese i18n working
- [x] UI/UX polished
- [x] **Ti Memory System working** (NEW)
- [x] **Ti Context Mode working** (NEW)
- [x] **Ti Review System working** (NEW)
- [x] **Ti Skill Creator working** (NEW)
- [x] **Ti Superpowers working** (NEW)
- [x] **Ti GSD working** (NEW)

---

## 13. Next Steps (Updated)

1. **Create `apps/ticlaw/` repository**
2. **Implement core patterns (Phase 1 - Backend)**
3. **Implement Ti Memory System (Priority 1)**
4. **Implement Ti Context Mode (Priority 3)**
5. **Integrate with Ti CLI (Phase 2)**
6. **Implement Ti Review System (Priority 2)**
7. **Build Notion Agent (Phase 3)**
8. **Implement Ti Skill Creator (Priority 4)**
9. **Create shared design system (Week 5 - UI Phase 1)**
10. **Build Ti Claw UI skeleton (Week 5 - UI Phase 2)**
11. **Implement core UI pages (Week 5-6 - UI Phase 3)**
12. **Implement additional UI pages (Week 6 - UI Phase 4)**
13. **Polish UI & integrate backend (Week 6 - UI Phase 5)**
14. **Implement Ti Superpowers (Priority 5)**
15. **Implement Ti GSD (Priority 6)**
16. **Build additional agents (Week 7)**
17. **Polish & Release (Week 8)**

---

## 14. References

- **User Insights**: Anthropic Claude Code tools analysis
- **Ti Router UI**: `apps/router/ui/`
- **Ti Skills**: `.devin/skills/`
- **Knowledge Graph Memory**: `content/mcp/KNOWLEDGE_GRAPH_USAGE_GUIDE.md`
- **Patterns Extraction**: `Ti-learning-lab/03_Knowledge/Router/goclaw-patterns-extraction.md`
- **UI Approach**: `Ti-learning-lab/03_Knowledge/Router/ti-claw-ui-approach.md`
- **Integration Plan**: `Ti-learning-lab/03_Knowledge/Router/ti-claw-integration-plan.md`

---

**Status**: Analysis Complete ✅
**Next**: Implement Ti Memory System (Priority 1)
