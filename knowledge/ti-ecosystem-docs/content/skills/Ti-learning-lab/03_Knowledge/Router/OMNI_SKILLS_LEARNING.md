# Awesome Omni Skills - Tóm Tắt Học Tập

**Nguồn**: `Z:\Ti\Ti-learning-lab\03_Knowledge\skills\awesome-omni-skills-main`
**Ngày**: 2026-04-29
**Mục đích**: Trích xuất các pattern kiến trúc cho hệ thống tool/skill của Ti Router

---

## Tổng Quan

Awesome Omni Skills là một hệ thống catalog skill và runtime toàn diện cho các trợ lý lập trình AI. Nó cung cấp:
- **3480 skill gốc** chia thành 17 danh mục
- **7 bundle được biên tập** cho các use case khác nhau
- **4 runtime surface**: CLI, API, MCP, A2A
- **9 client có thể cài đặt**: Claude Code, Cursor, Gemini CLI, Codex CLI, Kiro, Antigravity, Goose, Qwen Code, OpenCode
- **16 client có thể cấu hình MCP**

---

## Các Pattern Kiến Trúc Chính

### 1. Định Dạng Skill (SKILL.md)

**Cấu trúc**:
```yaml
---
name: skill-name
description: "Mô tả skill"
version: "0.0.1"
category: development
tags: ["tag1", "tag2"]
complexity: intermediate
risk: safe
tools: ["claude-code", "cursor"]
source: community
author: "author-name"
date_added: "2026-04-15"
date_updated: "2026-04-25"
---

# Tiêu Đề Skill

## Tổng Quan
## Khi Nào Sử Dụng Skill Này
## Bảng Vận Hành
## Quy Trình
## Ví Dụ
## Thực Hành Tốt Nhất
## Khắc Phục Sự Cố
## Các Skill Liên Quan
## Tài Nguyên Bổ Sung
```

**Thông Tin Chủ Chốt**:
- YAML frontmatter cho metadata có thể đọc được bởi máy
- Ranh giới kích hoạt rõ ràng trong "Khi Nào Sử Dụng"
- Bảng vận hành để hướng dẫn quyết định
- Các skill liên quan cho routing handoff
- Các file hỗ trợ (references, examples, scripts, agents, assets)

### 2. Kiến Trúc Catalog

**Các thành phần**:
- `catalog.json` - Manifest được tạo từ tất cả skills
- `catalog.db` - Database SQLite để tìm kiếm và lọc
- `bundles.json` - Các bundle skill được định nghĩa trước
- Archives - `.tar.gz` và `.zip` với checksums

**Pipeline tạo**:
```
skills/ → validate → metadata → catalog.json → catalog.db → archives
```

**Tính năng chính**:
- Chuẩn hóa và phân loại lại taxonomy
- Điểm chất lượng (quality score, security score)
- Tìm kiếm và lọc theo category, tags, quality
- Lập kế hoạch và so sánh bundle

### 3. Runtime Surfaces

#### Surface CLI
- `npx awesome-omni-skills` - Entry point chính
- Visual UI với Ink (React cho CLI)
- Các luồng cài đặt được hướng dẫn
- Khám phá trước khi cài đặt (lệnh `find`)
- Cài đặt theo mục tiêu cụ thể (`--claude`, `--cursor`, v.v.)

#### Surface API
- HTTP API chỉ đọc
- OpenAPI 3.1 với Swagger UI
- Endpoints: search, bundles, compare, install plans, downloads
- Port mặc định 3333

#### Surface MCP
- Nhiều transport: stdio, stream, SSE
- Các công cụ khám phá và đề xuất
- Xem trước cài đặt và local sidecar
- Tạo cấu hình cho 16 client MCP
- Local sidecar với các công cụ filesystem

#### Surface A2A
- Điều phối agent-to-agent
- Quản lý lifecycle task
- Handoff và polling
- Streaming và hủy bỏ
- Persistence
- Port mặc định 3335

### 4. Mục Tiêu Cài Đặt

**Mục tiêu tích hợp sẵn**:
```javascript
{
  id: "claude",
  name: "Claude Code",
  path: "~/.claude/skills",
  invocation: "Use skill-name to ..."
}
```

**Mục tiêu tùy chỉnh**:
- Các cấu hình tùy chỉnh đã lưu
- Hỗ trợ đường dẫn tùy chỉnh
- Pattern invocation theo mục tiêu cụ thể

**Logic cài đặt**:
- Cài đặt dựa trên symlink
- Kiểm tra an toàn (file tồn tại, permissions)
- Backup trước khi ghi đè
- Rollback khi thất bại

### 5. Validation và Bảo Mật

**Pipeline Validation**:
- Validation skill (YAML frontmatter, các section bắt buộc)
- Tạo metadata
- Cổng bảo mật quan trọng (chặn các pattern không an toàn)
- Chuẩn hóa taxonomy
- Verification archive

**Các Cổng Bảo Mật**:
- Chặn nội dung remote được pipe vào shell
- Chặn các hướng dẫn exposing prompts/secrets
- Quét ClamAV và VirusTotal
- Checksums SHA-256
- Artifacts được ký
- Verification release được thực thi bởi CI

### 6. Cấu Trúc Monorepo

```
awesome-omni-skills/
├── packages/
│   ├── cli/              # Công cụ CLI
│   ├── catalog-core/     # Logic catalog
│   ├── install-targets/  # Mục tiêu cài đặt
│   ├── server-api/       # Server API
│   ├── server-mcp/       # Server MCP
│   ├── server-a2a/       # Server A2A
│   └── i18n-runtime/     # Quốc tế hóa
├── skills/               # Input gốc
├── skills_omni/          # Output được biên tập (Tiếng Anh)
├── tools/scripts/        # Scripts build/validation
├── dist/                 # Artifacts được tạo
└── data/                 # Bundles, metadata
```

**Thông Tin Chủ Chốt**:
- Monorepo dựa trên workspace
- Các package được chia sẻ cho runtime surfaces
- Scripts Python cho validation
- Node.js cho runtime servers
- Sự phân tách rõ ràng giữa input và output được biên tập

---

## Các Pattern Áp Dụng Cho Ti Router

### 1. Registry Tool/Skill

**TiBrain hiện tại**: Registry tool với tích hợp MCP
**Pattern Omni**: Catalog với điểm chất lượng, bundles, search

**Đề xuất**:
- Thêm điểm chất lượng cho tools (0-100)
- Thêm điểm bảo mật (0-100)
- Tạo tool bundles (ví dụ: "file-ops", "network-ops")
- Triển khai search và lọc
- Tạo tool manifests (JSON + SQLite)

### 2. Định Dạng Tool

**TiBrain hiện tại**: Định nghĩa tool trong code Go
**Pattern Omni**: SKILL.md với YAML frontmatter

**Đề xuất**:
- Cân nhắc YAML frontmatter cho định nghĩa tool
- Thêm section "Khi Nào Sử Dụng" cho ranh giới kích hoạt
- Thêm "Các Tool Liên Quan" cho routing handoff
- Thêm file hỗ trợ (examples, scripts)

### 3. Runtime Surfaces

**TiBrain hiện tại**: Chỉ HTTP API
**Pattern Omni**: CLI, API, MCP, A2A

**Đề xuất**:
- Thêm công cụ CLI để quản lý TiBrain
- Thêm server MCP để khám phá và thực thi tool
- Thêm runtime A2A để điều phối agent
- Giữ HTTP API cho các tích hợp bên ngoài

### 4. Validation và Bảo Mật

**TiBrain hiện tại**: Validation cơ bản
**Pattern Omni**: Các cổng bảo mật toàn diện

**Đề xuất**:
- Thêm cổng bảo mật quan trọng cho đăng ký tool
- Chặn các pattern không an toàn (remote shell, secret exposure)
- Thêm checksums cho tool archives
- Triển khai verification CI cho các thay đổi tool

### 5. Mục Tiêu Cài Đặt

**TiBrain hiện tại**: Không có khái niệm cài đặt
**Pattern Omni**: Cài đặt đa client

**Đề xuất**:
- Định nghĩa mục tiêu cài đặt (Router, Devin, Claude)
- Thêm logic cài đặt tool
- Hỗ trợ đường dẫn tùy chỉnh
- Thêm rollback khi thất bại

### 6. Bundles

**TiBrain hiện tại**: Không có bundles
**Pattern Omni**: Các bundle skill được định nghĩa trước

**Đề xuất**:
- Tạo tool bundles cho các workflow phổ biến
- Ví dụ: "file-ops", "network-ops", "build-tools"
- Cho phép cài đặt bundle
- Lập kế hoạch và so sánh bundle

---

## Ưu Tiên Triển Khai

### Ưu Tiên Cao
1. **Điểm Chất Lượng Tool** - Thêm điểm chất lượng/bảo mật vào registry tool
2. **Tool Bundles** - Tạo tool bundles được định nghĩa trước
3. **Công Cụ CLI** - Thêm công cụ CLI để quản lý TiBrain
4. **Search và Lọc** - Triển khai search tool theo category, chất lượng

### Ưu Tiên Trung Bình
5. **Server MCP** - Thêm server MCP để khám phá tool
6. **Cài Đặt Tool** - Thêm logic cài đặt tool
7. **Cổng Validation** - Thêm validation bảo mật cho đăng ký tool
8. **Runtime A2A** - Thêm điều phối agent-to-agent

### Ưu Tiên Thấp
9. **Output Được Biên Tập** - Tách biệt input gốc và output được biên tập
10. **Quốc Tế Hóa** - Thêm hỗ trợ i18n
11. **Visual UI** - Thêm visual UI dựa trên Ink
12. **Tạo Archive** - Thêm tool archives với checksums

---

## Các Bài Học Chủ Chốt

1. **Phân Tách Mối Quan Tâm**: Sự phân tách rõ ràng giữa input gốc và output được biên tập
2. **Nhiều Runtime Surfaces**: Cùng một catalog được tiêu thụ bởi CLI, API, MCP, A2A
3. **Bảo Mật Đầu Tiên**: Các cổng bảo mật toàn diện trước khi xuất bản
4. **Chỉ Số Chất Lượng**: Điểm chất lượng và bảo mật cho tools
5. **Bundles**: Các bộ sưu tập được định nghĩa trước cho các use case phổ biến
6. **Cài Đặt**: Cài đặt đa client với các kiểm tra an toàn
7. **Validation**: Pipeline validation toàn diện
8. **Quốc Tế Hóa**: Hỗ trợ i18n tích hợp sẵn
9. **Monorepo**: Monorepo dựa trên workspace cho các package được chia sẻ
10. **Khám Phá**: Search, so sánh và lập kế hoạch trước khi cài đặt

---

## Tham Chiếu

- Repository: `https://github.com/diegosouzapw/awesome-omni-skills`
- Hướng Dẫn CLI: `docs/users/CLI-USER-GUIDE.md`
- Bundles: `docs/users/BUNDLES.md`
- Bắt Đầu: `docs/users/GETTING-STARTED.md`
