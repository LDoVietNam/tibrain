# OmniRoute Knowledge for Multi-Agent Workflows

## Mục tiêu

Tài liệu này mô tả cách đồng bộ tri thức từ `OmniRoute` vào `TiBrain` để nhiều agent/CLI khác nhau cùng dùng chung một lớp RAG về routing provider, model capability, RTK và tài liệu vận hành.

## Nguồn tri thức được nạp

- Provider knowledge pack từ `OmniRoute`
- Repo docs như `README.md`, `CLAUDE.md`, `RTK_INTEGRATION.md`, `CHANGELOG.md`
- Một phần `docs/*` của `OmniRoute`

## Cách sync nhanh trên Windows

Chạy từ thư mục `Z:\10_WORKPLACE\Ti\apps\tibrain`:

```powershell
py scripts/sync_omniroute_knowledge.py
```

Hoặc dùng `Makefile`:

```powershell
make sync-omniroute
```

## API ingest formats

### 1. Provider knowledge pack

```json
{
  "format": "omniroute_knowledge_pack",
  "file_path": "Z:\\10_WORKPLACE\\Ti\\tools\\router\\OmniRoute\\docs\\omniroute_provider_knowledge.json",
  "source": "OmniRoute provider knowledge sync"
}
```

### 2. Repo docs

```json
{
  "format": "omniroute_repo_docs",
  "repo_path": "Z:\\10_WORKPLACE\\Ti\\tools\\router\\OmniRoute",
  "source": "OmniRoute repository docs sync"
}
```

## Cách dùng cho nhiều agent

### Codex CLI

- Dùng khi cần trả lời về provider routing, fallback, RTK, capability, alias model
- Query nên nhấn mạnh `OmniRoute`, `router_knowledge`, tên provider, hoặc feature cụ thể
- Hữu ích cho tác vụ codegen cần hiểu policy routing trước khi sửa code

### Claude Code

- Dùng khi cần nghiên cứu repo, đối chiếu model routing, hoặc sinh patch dựa trên tri thức `OmniRoute`
- Có thể hỏi theo kiểu: “dựa trên knowledge OmniRoute trong TiBrain, provider nào hỗ trợ reasoning + tools?”

### Cline / Roo / OpenCode-style agents

- Dùng cho IDE assistant cần context kỹ thuật gọn thay vì quét cả repo lớn mỗi lần
- Phù hợp khi agent cần chọn model/provider tương thích task code, embeddings, hoặc RTK flow

### Cursor

- Dùng khi muốn editor agent tham chiếu tri thức routing tập trung thay vì nhúng prompt thủ công dài
- Nên kết hợp với query có category `router_knowledge`

### Open WebUI / custom chat frontends

- Có thể gọi `TiBrain` qua `/api/v1/rag/query`
- Dùng để xây “Router Brain” chat chuyên trả lời về `OmniRoute`, provider, model, limits, RTK

### Ti nội bộ agents

- `TiBrain` có thể đóng vai trò knowledge hub chung cho `router`, `cli`, `mcp`, `ops` agents
- Khuyến nghị mọi agent truy cập cùng một nguồn RAG để tránh lệch tri thức giữa các tool

## Query patterns gợi ý

- “OmniRoute hỗ trợ provider nào cho embeddings?”
- “RTK trong OmniRoute dùng để làm gì?”
- “Provider nào có model hỗ trợ vision + tools?”
- “Tài liệu OmniRoute nói gì về fallback routing?”
- “Khác nhau giữa knowledge pack và repo docs trong TiBrain là gì?”

## Gợi ý vận hành

- Sync lại sau khi `OmniRoute` thay đổi provider registry hoặc tài liệu lớn
- Dùng knowledge pack cho câu hỏi cấu trúc/provider nhanh
- Dùng repo docs cho câu hỏi giải thích feature, workflow, lịch sử thay đổi
- Nếu `CHANGELOG.md` quá lớn, cân nhắc tách chunk/summarize ở bước tiếp theo

## Hạn chế hiện tại

- Import repo docs hiện mới lấy tập tài liệu chọn lọc, không ingest toàn bộ repo code
- Chưa có filter query riêng cho `source_system=OmniRoute`
- Chưa có chunk/summarization chuyên biệt cho tài liệu rất dài như changelog

## Bước nên làm tiếp

- Thêm query filter theo `source_system` và `source_type`
- Thêm chunking/summarization cho `CHANGELOG.md`
- Thêm job sync định kỳ cho `OmniRoute`
