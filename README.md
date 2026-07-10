# TiBrain

> Dịch vụ **control-plane / knowledge-hub** viết bằng Go, chạy trên một cổng HTTP duy nhất, gom REST API, MCP (SSE) và Browser UI vào cùng một tiến trình.

Repo: [`github.com/ti/router/tibrain`](https://github.com/ti/router/tibrain)

---

## Overview

TiBrain là "bộ não" trung tâm điều phối cho hệ sinh thái agent/CLI. Nó cung cấp
một điểm truy cập thống nhất cho:

- **Điều phối agent (agent orchestration)** và bàn giao công việc (handoff) giữa các CLI.
- **Bộ nhớ nhận thức (cognitive memory)**: episodic, semantic, procedural.
- **RAG retrieval**: vector store, reranker, embeddings, adaptive retrieval, cache.
- **Knowledge indexing**: nạp và đánh chỉ mục tri thức phục vụ truy hồi.
- **MCP hub**: đăng ký và phục vụ tool qua Model Context Protocol (SSE).
- **Learning & quality**: thống kê theo mô hình BEADS LEARN.

Toàn bộ chạy trên **một cổng duy nhất** (mặc định `1810`): REST tại `/api/*`,
MCP SSE tại `/mcp`, và Browser UI phục vụ trực tiếp từ cùng server.

---

## Features

- **CLI registry & handoff** — Đăng ký các CLI/agent và điều phối bàn giao tác vụ.
- **MCP tool/registry** — Máy chủ MCP với registry tool, phục vụ qua SSE tại `/mcp`.
- **Cognitive memory** — Ba tầng bộ nhớ: episodic (sự kiện), semantic (khái niệm), procedural (quy trình).
- **RAG retrieval** — Vector store, reranker, embeddings, adaptive retrieval; cache tùy chọn qua Redis.
- **Knowledge indexing** — Nạp và đánh chỉ mục tri thức (có flag `--index-knowledge`).
- **Agent orchestration** — Điều phối luồng làm việc giữa nhiều agent.
- **Learning / quality** — Theo dõi chất lượng và thống kê học tập theo mô hình BEADS LEARN.
- **Cloudflare integration** — Tích hợp dịch vụ Cloudflare.
- **RTK handler** — Xử lý các tác vụ RTK.

---

## Architecture

Kiến trúc single-port: một HTTP server duy nhất định tuyến tới nhiều subsystem.

```
                        ┌──────────────────────────────────────┐
         HTTP :1810 ───►│            TiBrain Server             │
                        │  (single-port HTTP, main.go)          │
                        ├──────────────────────────────────────┤
    /api/*  (REST) ────►│  REST API layer                       │
    /mcp    (SSE)  ────►│  MCP hub (tool registry)              │
    /  (Browser UI) ───►│  Web UI                               │
    /health /ready ────►│  Health & readiness                   │
                        ├──────────────────────────────────────┤
                        │  Core subsystems:                     │
                        │   • CLI registry & handoff            │
                        │   • Agent orchestration               │
                        │   • Cognitive memory                  │
                        │     (episodic/semantic/procedural)    │
                        │   • RAG (vector store, reranker,      │
                        │     embeddings, adaptive retrieval)   │
                        │   • Knowledge indexing                │
                        │   • Learning / quality (BEADS LEARN)  │
                        │   • Cloudflare integration, RTK       │
                        └──────────────────────────────────────┘
                                     │            │
                                     ▼            ▼
                              Redis (cache,   Vector store /
                               optional)      knowledge index
```

---

## Installation

**Yêu cầu:** Go `1.25.8`, `make`.

> **Status: đang hoàn thiện.** Ở thời điểm hiện tại repo **chưa build xanh** vì
> một số package trong `internal/` (bao gồm `db`, `memory`, `tools`) đang thiếu.
> Cần bổ sung đầy đủ các package `internal/` này trước khi `make build` có thể
> hoàn tất thành công. Các bước dưới đây mô tả quy trình chuẩn khi mã nguồn đầy đủ.

Build binary vào thư mục `build/`:

```bash
make build
```

Binary sinh ra sẽ nằm trong `build/`.

---

## Configuration

TiBrain đọc cấu hình từ `config.yaml` kết hợp với biến môi trường (env) và cờ dòng lệnh (flags).
Thứ tự ưu tiên thông thường: flags > env > config.yaml.

### config.yaml

File cấu hình chính của dịch vụ. Đặt cạnh binary hoặc chỉ tới bằng đường dẫn khi chạy.

### Environment variables

Các thiết lập trong `config.yaml` có thể được ghi đè qua biến môi trường (ví dụ
cấu hình Redis cho cache RAG, khóa tích hợp Cloudflare, v.v.).

### Flags

| Flag | Mô tả |
| --- | --- |
| `--port` | Cổng HTTP để lắng nghe (mặc định `1810`). |
| `--index-knowledge` | Chạy quá trình đánh chỉ mục tri thức. |

Ví dụ:

```bash
./build/tibrain --port 1810
./build/tibrain --index-knowledge
```

---

## Running

Chạy bằng Makefile:

```bash
make run
```

Hoặc chạy binary trực tiếp:

```bash
./build/tibrain --port 1810
```

Server khởi động trên một cổng duy nhất và phục vụ:

- **REST API** — `http://localhost:1810/api/*`
- **MCP SSE** — `http://localhost:1810/mcp`
- **Browser UI** — `http://localhost:1810/`
- **Health / status** — `/health`, `/ready`, `/status`

Kiểm tra nhanh:

```bash
curl http://localhost:1810/health
curl http://localhost:1810/ready
curl http://localhost:1810/status
```

---

## API Summary

| Nhóm | Endpoint (tiêu biểu) | Mô tả |
| --- | --- | --- |
| **Health / status** | `GET /health` | Liveness — dịch vụ còn sống hay không. |
| | `GET /ready` | Readiness — kiểm tra các subsystem: hub, memory, rag, mcp. |
| | `GET /status` | Trạng thái tổng hợp của dịch vụ. |
| **CLI registry** | `/api/*` | Đăng ký CLI/agent và điều phối handoff giữa các CLI. |
| **MCP hub** | `/mcp` (SSE) | Máy chủ MCP với tool registry, phục qua Server-Sent Events. |
| **RAG** | `/api/*` | Truy hồi RAG: vector store, reranker, embeddings, adaptive retrieval. |
| **Knowledge** | `/api/*` | Nạp và đánh chỉ mục tri thức (xem thêm flag `--index-knowledge`). |
| **Agents** | `/api/*` | Điều phối agent (agent orchestration). |
| **Cognitive memory** | `/api/*` | Bộ nhớ nhận thức: episodic, semantic, procedural. |
| **RTK** | `/api/*` | Xử lý tác vụ RTK. |

> Chi tiết đường dẫn cụ thể của từng nhóm `/api/*` bám theo router thực tế trong mã nguồn.

---

## Development

Xác minh chất lượng trước khi commit (vet + lint + test-short):

```bash
make verify
```

Các lệnh phổ biến:

```bash
make build    # build binary vào build/
make run      # chạy dịch vụ
make verify   # vet + lint + test-short
```

Điểm vào chương trình: `main.go` (hàm `main`).

---

## Project Status / Notes

- **Trạng thái: đang hoàn thiện.** Repo hiện **chưa build được** do thiếu các
  package trong `internal/` (`db`, `memory`, `tools`). Cần bổ sung đầy đủ trước
  khi build thành công.
- Kiến trúc single-port: REST (`/api/*`) + MCP SSE (`/mcp`) + Browser UI cùng một cổng (mặc định `1810`).
- Cache RAG qua Redis là **tùy chọn** (optional).
- Readiness (`/ready`) phản ánh trạng thái của các subsystem: hub, memory, rag, mcp.
- Yêu cầu toolchain: Go `1.25.8`.

## Development notes

- **Submodules**: The repositories `mcp/`, `obsidian-headless/`, and `qdrant/` are Git submodules. After cloning, run `git submodule update --init --recursive` to fetch them.
- **Runtime data**: All runtime files (SQLite DB, WAL/SHM, Bleve index, 1MCP configuration) reside in the `data/` directory, which is ignored by Git.
- **Handoff logging**: Agents must log actions to the central handoff file at `Z:\02_CORE\_cli\.config\handoff.json` (see AGENTS.md for format).

## Handoff Logging

All agents working in this repository must log handoff events to the centralized handoff file located at:

`Z:\02_CORE\_cli\.config\handoff.json`

Each log entry should be a single-line JSON object appended to the file (newline delimited). Example:

```json
{"agent":"tibrain","action":"updated AGENTS.md and README.md with handoff logging guidance","timestamp":"2026-07-10T05:30:00+07:00","details":{"files":["AGENTS.md","README.md"]}}
```

Fields:
- `agent`: the agent identifier (must be "tibrain").
- `action`: short description of the change or task performed.
- `timestamp`: ISO 8601 timestamp with timezone offset.
- `details`: optional object with additional context (e.g., list of files changed, issue numbers).

This practice mirrors the bead logging convention and enables cross‑session traceability.