# Biểu Tượng Ti Learning Lab

> **Cập Nhật Cuối**: 2026-05-08
> **Phiên Bản**: 4.0

---

## 🎯 Mục Tiêu

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

### Success Metrics
- **Application Rate**: Số lượng docs được apply vào real projects / Tổng docs
- **Verification Rate**: Số lượng docs có verification status / Tổng docs
- **Connection Rate**: Số lượng docs có cross-references / Tổng docs
- **Deprecation Rate**: Số lượng docs obsolete được deprecate / Tổng docs

---

## Bảng Trạng Thái

|| Trạng Thái | Emoji | Ý Nghĩa |
||-----------|-------|---------|
|| **Đang Hoạt Động** | ✅ | Đang sử dụng, cập nhật |
|| **Ổn Định** | 🟢 | Không cần thay đổi, tài liệu tham khảo |
|| **Đang Thực Hiện** | 🔄 | Đang được làm việc |
|| **Chờ Xét Duyệt** | ⏳ | Đã lập kế hoạch, chưa bắt đầu |
|| **Bị Chặn** | 🚫 | Bị chặn bởi phụ thuộc hoặc vấn đề |
|| **Không Còn Dùng** | ⚠️ | Cũ, nên xóa |
|| **Đã Duyệt** | ✅ | Kế hoạch đã được duyệt, sẵn sàng thực hiện |
|| **Hoàn Thành** | ✅ | Nhiệm vụ/dự án hoàn thành |
|| **Đã Lưu Trữ** | 📦 | Đã chuyển vào kho lưu trữ |

---

## Luồng Dữ Liệu

```
01_Learning/<project>/  ──┐
                          ├──→ 03_Knowledge ──→ Brain/Memory ──→ 04_Planning ──→ 05_Taskboard ──→ Task thực tế
02_Research/<project>/  ───┘
                              ↑
                              │
06_HandsOn (POC) ─────────────┘
07_Repositories (tham khảo) ──┘
```

---

## 01 — Tài Liệu Học Tập (01_Learning/)

> **Mục đích**: Đọc, nắm kiến thức nền tảng → **bổ sung vào 03_Knowledge**
> **Khi nào dùng**: Cần học concept mới, framework, pattern, best practice cho dự án cụ thể
> **Output**: Tài liệu hiểu biết (không phải code chạy)

|| Dự Án | Thư Mục | Trạng Thái | Cập Nhật Cuối |
||-------|---------|-----------|---------------|
|| Ti CLI | `01_Learning/CLI/` | ⏳ Chờ | - |
|| TiBrain | `01_Learning/TiBrain/` | ⏳ Chờ | - |
|| Router | `01_Learning/Router/` | ⏳ Chờ | - |
|| MCPHub | `01_Learning/MCPHub/` | ⏳ Chờ | - |
|| Core | `01_Learning/Core/` | ⏳ Chờ | - |
|| UI | `01_Learning/UI/` | ⏳ Chờ | - |
|| SDK | `01_Learning/SDK/` | ⏳ Chờ | - |
|| Providers | `01_Learning/Providers/` | ⏳ Chờ | - |
|| Registry | `01_Learning/Registry/` | ⏳ Chờ | - |
|| Monitoring | `01_Learning/Monitoring/` | ⏳ Chờ | - |
|| Deploy | `01_Learning/Deploy/` | ⏳ Chờ | - |
|| Docs | `01_Learning/Docs/` | ⏳ Chờ | - |
|| Tests | `01_Learning/Tests/` | ⏳ Chờ | - |
|| BestSource | `01_Learning/BestSource/` | ⏳ Chờ | - |
|| Configs | `01_Learning/Configs/` | ⏳ Chờ | - |

---

## 02 — Nghiên Cứu (02_Research/)

> **Mục đích**: Phân tích, thử nghiệm cho **dự án cụ thể** → **tạo Plan (04)** → thực hiện Task
> **Khi nào dùng**: Cần đánh giá giải pháp, benchmark, proof-of-concept trước khi quyết định cho dự án
> **Output**: Báo cáo phân tích, recommendation, decision record

|| Dự Án | Thư Mục | Trạng Thái | Cập Nhật Cuối |
||-------|---------|-----------|---------------|
|| Ti CLI | `02_Research/CLI/` | ⏳ Chờ | - |
|| TiBrain | `02_Research/TiBrain/` | ⏳ Chờ | - |
|| Router | `02_Research/Router/` | ⏳ Chờ | - |
|| MCPHub | `02_Research/MCPHub/` | ⏳ Chờ | - |
|| Core | `02_Research/Core/` | ⏳ Chờ | - |
|| UI | `02_Research/UI/` | ⏳ Chờ | - |
|| SDK | `02_Research/SDK/` | ⏳ Chờ | - |
|| Providers | `02_Research/Providers/` | ⏳ Chờ | - |
|| Registry | `02_Research/Registry/` | ⏳ Chờ | - |
|| Monitoring | `02_Research/Monitoring/` | ⏳ Chờ | - |
|| Deploy | `02_Research/Deploy/` | ⏳ Chờ | - |
|| Docs | `02_Research/Docs/` | ⏳ Chờ | - |
|| Tests | `02_Research/Tests/` | ⏳ Chờ | - |
|| BestSource | `02_Research/BestSource/` | ⏳ Chờ | - |
|| Configs | `02_Research/Configs/` | ⏳ Chờ | - |

---

## 03 — Kiến Thức Tổng Hợp (03_Knowledge/)

> **Mục đích**: Tổng hợp từ **01_Learning + 02_Research** → thành kiến thức có cấu trúc cho **Brain/Memory**
> **Khi nào dùng**: Tra cứu nhanh, onboard agent mới, verify quyết định
> **Output**: Pattern, runbook, decision log, cheatsheet

### Cấu Trúc Hiện Tại (Updated 2026-05-08)

```
03_Knowledge/
├── 00_META/                   # Metadata about reorganization ✅
│   ├── AUDIT_REPORT.md        # Content audit report
│   ├── CROSS_REFERENCE_EXAMPLE.md # Cross-reference examples
│   └── TI_LEGO_ARCHITECTURE.md # Ti LEGO architecture model (Kernel, Module, Block, Plugin) 🆕
├── agents/                    # Agent framework and orchestration ✅
├── api/                       # API integration patterns
├── archive/                   # Archived files (includes old Chinese folder)
├── auto-reg-tools/            # Auto registration tools
├── cache/                     # Cache
├── cli/                       # CLI documentation ✅
│   ├── core/                  # Core CLI architecture
│   ├── features/              # CLI features
│   ├── integration/           # CLI integration
│   ├── deployment/            # CLI deployment
│   └── sources/               # CLI source analysis
│       └── chatgpt-cli/        # ChatGPT CLI analysis
├── computer-vision/           # Computer vision
├── devin/                     # Devin porting ✅
├── docs/                      # Documentation references
├── frontend/                  # Frontend patterns ✅
├── hot-reload/                 # Hot reload
├── lessons/                   # Lessons learned ✅
├── metrics/                   # Metrics ✅
├── notion/                    # Notion integration ✅
├── patterns/                  # Patterns ✅
├── prompt-engineering/        # Prompt engineering
├── protocols/                 # Protocols
├── provider/                  # Provider management
├── research/                  # Research findings ✅
├── Router/                    # Router (English) ✅
│   └── CookieProviders/       # Cookie providers documentation 🆕
├── sources/                   # Sources (placeholder)
├── storage/                   # Large files & binaries (zip archives) 🆕
├── tibrain/                   # Tibrain
├── ticlaw/                    # Ti-claw
├── ti-router/                 # Router (Vietnamese) ✅
└── tool/                      # Tool
```

**Notes:**
- ✅ = Has INDEX.md for navigation
- 🆕 = Added in this cleanup (2026-05-08: CookieProviders/)
- Removed: `browser-automation/`, `cheatsheets/`, `decisions/`, `runbooks/`, `spectre/`, `donut-browser/` (empty folders)
- Moved to archive: `分析sharedchatfun的cook/` (Chinese folder, unrelated)
- Moved to storage: Large zip files from Router/ and browser/
- Moved to 07_Repositories/: `browser/` → `07_Repositories/browser-docs/`, `skills/` → `07_Repositories/skills-docs/` (source code repos)
- Renamed: `cli/00_INDEX.md` → `cli/INDEX.md` (consistent naming)

### Categories và Status

|| Category | Thư Mục | Trạng Thái | Cập Nhật Cuối |
||----------|---------|-----------|---------------|
|| **CLI** | `cli/core/` | ✅ Đang Hoạt Động | 2026-05-05 |
|| | `cli/features/` | ✅ Đang Hoạt Động | 2026-05-05 |
|| | `cli/integration/` | ✅ Đang Hoạt Động | 2026-05-05 |
|| | `cli/deployment/` | ✅ Đang Hoạt Động | 2026-05-05 |
|| | `cli/sources/` | ✅ Đang Hoạt Động | 2026-05-05 |
|| **Agent Framework** | `agents/` | ✅ Đang Hoạt Động | 2026-05-05 |
|| **Router** | `Router/` | ✅ Đang Hoạt Động | 2026-05-05 |
|| | `ti-router/` | ✅ Đang Hoạt Động | 2026-05-05 |
|| **Research** | `research/` | ✅ Đang Hoạt Động | 2026-05-05 |
|| **Patterns** | `patterns/` | ✅ Đang Hoạt Động | 2026-05-05 |
|| **Lessons** | `lessons/` | ✅ Đang Hoạt Động | 2026-05-05 |
|| **Ti-claw** | `ticlaw/` | ✅ Đang Hoạt Động | 2026-05-05 |
|| **Devin** | `devin/` | ✅ Đang Hoạt Động | 2026-05-05 |
|| **Notion** | `notion/` | ✅ Đang Hoạt Động | 2026-05-05 |
|| **Frontend** | `frontend/` | ✅ Đang Hoạt Động | 2026-05-05 |
|| **Metrics** | `metrics/` | ✅ Đang Hoạt Động | 2026-05-05 |
|| **Decisions** | `decisions/` | ⏳ Chờ | - |
|| **Protocols** | `protocols/` | ⏳ Chờ | - |
|| **Runbooks** | `runbooks/` | ⏳ Chờ | - |
|| **Skills** | `skills/` | ⏳ Chờ | - |
|| **Tibrain** | `tibrain/` | ⏳ Chờ | - |
|| **API** | `api/` | ⏳ Chờ | - |
|| **Browser** | `browser/` | ⏳ Chờ | - |
|| **Computer Vision** | `computer-vision/` | ⏳ Chờ | - |
|| **Cache** | `cache/` | ⏳ Chờ | - |
|| **Docs** | `docs/` | ⏳ Chờ | - |
|| **Hot Reload** | `hot-reload/` | ⏳ Chờ | - |
|| **Spectre** | `spectre/` | ⏳ Chờ | - |
|| **Tool** | `tool/ | ⏳ Chờ | - |
|| **Provider** | `provider/` | ⏳ Chờ | - |
|| **Prompt Engineering** | `prompt-engineering/` | ⏳ Chờ | - |
|| **Cheatsheets** | `cheatsheets/` | ⏳ Chờ | - |
|| **Auto-reg-tools** | `auto-reg-tools/ | ⏳ Chờ | - |
|| **Archive** | `archive/` | ⏳ Chờ | - |

### Brain/Memory Integration

```powershell
# Import CLI pattern vao TiBrain LTM
curl -X POST http://localhost:1809/v1/tibrain/memory/append `
  -H "Content-Type: application/json" `
  -d '{
    "cli": "ti",
    "category": "cli-core",
    "content": "[03_Knowledge/cli/core/ARCHITECTURE.md]: Ti CLI plugin architecture with dynamic loading",
    "tags": ["cli", "plugin", "architecture"]
  }'
```

---

## 04 — Lập Kế Hoạch (04_Planning/)

> **Mục đích**: Tạo kế hoạch chi tiết dựa trên **02_Research** → đưa vào 05_Taskboard
> **Khi nào dùng**: Sau khi research xong, cần breakdown task
> **Output**: Plan với acceptance criteria, timeline, resources

|| Loại | Kế Hoạch | Trạng Thái | Ưu Tiên | Cập Nhật Cuối | Ghi Chú |
||------|----------|-----------|---------|---------------|---------|

---

## 05 — Bảng Nhiệm Vụ (05_Taskboard/)

> **Mục đích**: Theo dõi nhiệm vụ đang thực hiện (beads log, task list)
> **Khi nào dùng**: Trong quá trình làm task thực tế
> **Output**: beads.md, task progress

---

## 06 — Thực Hành (06_HandsOn/)

> **Mục đích**: Áp dụng **01_Learning + 02_Research** vào code thực tế (POC, prototype)
> **Khi nào dùng**: Cần thử code trước khi merge vào dự án chính
> **Output**: Code chạy được, proof-of-concept, demo

|| Loại | Nội Dung | Trạng Thái | Ghi Chú |
||------|----------|-----------|---------|
|| POC | Thử nghiệm giải pháp mới | ⏳ Trống | Không merge, chỉ học |
|| Prototype | Mẫu sớm cho feature | ⏳ Trống | Có thể refactor sau |
|| Spike | Tìm hiểu công nghệ lạ | ⏳ Trống | Giới hạn thời gian |

---

## 07 — Kho Mã Nguồn (07_Repositories/)

> **Mục đích**: Clone repo bên ngoài về để **nghiên cứu, tham khảo, extract patterns**
> **Khi nào dùng**: Cần học từ open source, tham khảo implementation
> **Quy tắc**: Chỉ đọc/reference — không sửa code gốc, copy pattern có credit

|| Repository | Ngôn Ngữ | Trạng Thái | Mục Đích Clone |
||-----------|---------|-----------|---------------|

---

## 08 — Lưu Trữ (08_Archives/)

> **Mục đích**: Lưu trữ tài liệu cũ không còn active nhưng cần giữ lịch sử

---

## 09 — Chuyên Môn (09_Specialized/)

> **Mục đích**: Tài liệu chuyên sâu theo domain (security, performance, DevOps...)

---

## Lịch Sử Thay Đổi

- **2026-04-30**: Cấu trúc 01/02 theo dự án thực tế trong `Z:\\10_WORKPLACE\\Ti\\`
- **2026-04-30**: Tách 03_Knowledge thành patterns/runbooks/decisions/lessons/cheatsheets cho Brain/Memory
- **2026-04-30**: Đổi `06_Projects/` → `06_HandsOn/`
- **2026-05-05**: Sắp xếp lại 03_Knowledge/ theo cấu trúc thực tế (30+ folders)
- **2026-05-05**: Cập nhật MANIFEST.md với cấu trúc thực tế của 03_Knowledge/
- **2026-05-05**: Đổi naming convention từ số (01_, 02_, v.v.) sang tên rõ nghĩa (core, features, v.v.) cho cli/
- **2026-05-05**: Cleanup 03_Knowledge/ - xóa browser-automation/, move folder tiếng Trung đến archive/, move zip files đến storage/
- **2026-05-05**: Tạo INDEX.md cho 10 folders chính (agents, patterns, lessons, devin, notion, frontend, metrics, research, browser, storage, ti-router)
- **2026-05-05**: Cross-reference Router/ (English) và ti-router/ (Vietnamese)
- **2026-05-05**: Deep cleanup - xóa 5 folders trống (donut-browser, cheatsheets, decisions, runbooks, spectre)
- **2026-05-05**: Move browser/ và skills/ đến 07_Repositories/ (source code repos)
- **2026-05-05**: Rename cli/00_INDEX.md → cli/INDEX.md (consistent naming)
- **2026-05-06**: Định nghĩa goal mới: "Chuyển đổi kiến thức thụ động thành năng lực thực tế"
- **2026-05-06**: Priority 1 cleanup - xóa 05_Tracking/, move codex-autoresearch/ → 06_HandsOn/, move undetectable-fingerprint-browser/ → 07_Repositories/, xóa scripts/ và backups/
- **2026-05-06**: Priority 2 - tạo WORKFLOW.md (hướng dẫn workflow) và templates/ cho các stages
- **2026-05-06**: Priority 3 - tạo AUDIT_REPORT.md (báo cáo audit chất lượng content 03_Knowledge/)
- **2026-05-06**: Priority 4 - tạo CROSS_REFERENCE_EXAMPLE.md (example cách tạo cross-references)
- **2026-05-06**: Update README.md với links mới và cấu trúc mới
- **2026-05-06**: Version 3.1
- **2026-05-06**: Subfolder audit - xóa 02_Research/, 04_Planning/, 09_Specialized/ (trống)
- **2026-05-06**: Move zip files từ 06_HandsOn/ và 07_Repositories/ vào storage/
- **2026-05-06**: Move git repos từ 06_HandsOn/ vào 07_Repositories/
- **2026-05-06**: Cleanup 05_Taskboard/ - xóa failed files, move schema files vào storage/
- **2026-05-06**: Move files ở root của 01_Learning/ vào appropriate folders
- **2026-05-06**: Xóa duplicate folders (taskboard/, archive/)
- **2026-05-06**: Version 3.2
- **2026-05-07**: Tạo CONVERT_TRANS_ARCHITECTURE.md - tài liệu chi tiết về kiến trúc convert/trans của Ti CLI
- **2026-05-07**: Bổ sung OpenAPI Generator vào tài liệu convert/trans - hỗ trợ OpenAPI → Go SDK conversion
- **2026-05-07**: Triển khai thực sự OpenAPI Generator - tạo file openapi_generator.go, update registry, detectSourceFormat, và CLI commands
- **2026-05-07**: Tạo UI_ARCHITECTURE.md - tài liệu chi tiết về kiến trúc UI của Ti (Ti TUI + UIGenerator)
- **2026-05-07**: Cập nhật UI_ARCHITECTURE.md - bổ sung section về Ti CLI Internal UI Capabilities (Dashboard TUI, Chat TUI, Provider Wizard, Config TUI, Memory TUI, Plugin TUI)
- **2026-05-07**: Cập nhật UI_ARCHITECTURE.md - bổ sung section về Ti UI Generation Capabilities (UIGenerator vs Frontend-Design Skill)
- **2026-05-07**: Tạo TI_LEGO_ARCHITECTURE.md - tài liệu về kiến trúc LEGO của Ti (Kernel, Module, Block, Plugin)
- **2026-05-07**: Cập nhật AGENTS.md để reference TI_LEGO_ARCHITECTURE.md
- **2026-05-07**: Cập nhật MANIFEST.md để track TI_LEGO_ARCHITECTURE.md
- **2026-05-07**: Ti CLI UI Improvements - Fix Memory TUI command registration, add sample config data for Config TUI, refactor tui.go (1041 lines) into chat.go, remove legacy code, fix magic numbers with constants.go, remove duplicate code patterns (minIntWizard), create UI_COMPONENTS.md documentation
- **2026-05-07**: Version 3.9
- **2026-05-08**: Tạo Cookie Providers documentation trong 03_Knowledge/Router/CookieProviders/ - docs cho 10 cookie providers (Claude, Gemini, ChatGPT, Perplexity, Copilot, Windsurf, Notion, Meta, HuggingChat) với integration steps
- **2026-05-08**: Tạo ROUTER_REFERENCE.md - tham khảo từ 9Router, LLMCookieBridge, Kilo CLI, OpenRouter
- **2026-05-08**: Tạo AUTO_UPDATE_MECHANISM.md - GitHub action để monitor router updates
- **2026-05-08**: Move cookie provider docs từ Z:\07_DOCS\cookie-providers\ đến Z:\10_WORKPLACE\Ti\Ti-learning-lab\03_Knowledge\Router\CookieProviders\
- **2026-05-08**: Update CONFIG_GUIDE.md với đường dẫn mới đến cookie provider docs
- **2026-05-08**: Update cross-references trong tất cả cookie provider docs với đường dẫn mới
- **2026-05-08**: Version 4.0
