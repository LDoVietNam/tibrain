# Dự án Ti

> **Mục đích**: Hệ sinh thái Ti - Nền tảng phát triển powered by AI
> **Vị trí**: `Z:\10_WORKPLACE\Ti`
> **Cập nhật lần cuối**: 2026-05-15
> **Phiên bản**: 2.0.0

---

## 🎯 Tổng quan

Ti là một hệ sinh thái phát triển powered by AI với các thành phần:

- **Ti CLI**: Công cụ dòng lệnh chính
- **TiCrew**: Hệ thống AI agents và RAG
- **Learning Lab**: Nền tảng học tập và kiến thức
- **Content Hub**: Kho lưu trữ skills, workflows, agents

## 📁 Cấu trúc dự án

```
Z:\10_WORKPLACE\Ti\
├── docs/                     # Trung tâm tài liệu
│   ├── rag/                 # Tài liệu hệ thống RAG
│   ├── agents/              # Tài liệu agents
│   ├── cli/                 # Tài liệu CLI
│   ├── development/         # Hướng dẫn phát triển
│   └── deployment/          # Hướng dẫn triển khai
├── apps/                     # Ứng dụng
│   ├── cli/                 # Ứng dụng CLI chính thức
│   ├── auto_reg/            # Công cụ tự động đăng ký
│   ├── ticrew/              # Ứng dụng TiCrew
│   │   └── rag-backend/     # Hệ thống backend RAG
│   └── mcp/                 # Máy chủ MCP
├── packages/                 # Thư viện chia sẻ
│   ├── sdk/                 # SDK Plugin API
│   ├── core/                # Modules chính
│   └── providers/           # Triển khai providers
├── content/                  # Skills, Workflows, Agents
│   ├── skills/              # AI Skills
│   ├── agents/              # Cấu hình Agents
│   ├── workflows/           # Workflows tự động hóa
│   └── rules/               # Quy tắc hệ thống
├── configs/                  # Files cấu hình
├── scripts/                  # Scripts xây dựng & tiện ích
├── tests/                    Bộ kiểm thử
├── examples/                 # Ví dụ mã nguồn
├── tools/                    # Công cụ phát triển
└── Ti-learning-lab/          # Tài liệu học tập
```

## 🚀 Bắt đầu nhanh

### Xây dựng CLI
```bash
cd apps/cli
go build -o ../../bin/ti .
```

### Xây dựng hệ thống RAG
```bash
cd apps/ticrew/rag-backend
go build -o rag-backend ./cmd
./rag-backend
```

### Truy cập tài liệu
```bash
# Tài liệu chính
open docs/README.md

# Tài liệu hệ thống RAG
open docs/rag/README.md
```

## 📚 Tài liệu

### 📋 Trung tâm tài liệu
- **[docs/README.md](docs/README.md)** - Tổng quan tài liệu
- **[docs/rag/README.md](docs/rag/README.md)** - Tài liệu hệ thống RAG
- **[docs/agents/README.md](docs/agents/README.md)** - Tài liệu agents
- **[docs/cli/README.md](docs/cli/README.md)** - Tài liệu CLI

### 📋 Tài liệu thành phần
- **[apps/README.md](apps/README.md)** - Tổng quan ứng dụng
- **[packages/README.md](packages/README.md)** - Tổng quan packages
- **[content/README.md](content/README.md)** - Tổng quan nội dung

### 📋 Kiến trúc
- **[docs/rag/architecture.md](docs/rag/architecture.md)** - Kiến trúc hệ thống RAG

## 🛠️ Phát triển

### Điều kiện tiên quyết
- Go 1.26+
- Qdrant (cho hệ thống RAG)
- Neo4j (cho đồ thị kiến thức)

### Lệnh xây dựng
```bash
# Xây dựng tất cả ứng dụng
make build

# Xây dựng ứng dụng cụ thể
make build-cli
make build-ticrew

# Chạy kiểm thử
make test

# Dọn dẹp build artifacts
make clean
```

### Quy trình phát triển
1. Fork repository
2. Tạo nhánh tính năng
3. Thực hiện thay đổi
4. Chạy kiểm thử
5. Gửi pull request

## 🤖 AI Agents

### Hướng dẫn cho AI Agents
- **[AGENTS.md](AGENTS.md)** - Hướng dẫn chi tiết cho AI agents (Claude, Devin, Gemini, v.v.)

### Tài nguyên cho Agents
- **[content/](content/)** - Skills, workflows, agents, và rules
- **[content/rules/](content/rules/)** - Quy tắc và guidelines bắt buộc
- **[content/workflows/](content/workflows/)** - Quy trình tự động hóa
- **[content/skills/](content/skills/)** - Kỹ năng AI có thể tái sử dụng

## 🤖 Cộng đồng

### Đóng góp
- [GitHub Repository](https://github.com/ti/ti)
- [Máy chủ Discord](https://discord.gg/ti)
- [Tài liệu](https://ti.github.io/docs)

### Hỗ trợ
- [Vấn đề](https://github.com/ti/ti/issues)
- [Thảo luận](https://github.com/ti/ti/discussions)
- [Wiki](https://github.com/ti/ti/wiki)

---

*Cập nhật lần cuối: 2026-05-15*