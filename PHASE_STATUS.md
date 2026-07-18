# PHASE_STATUS.md — TiBrain Agent OS

> Audit thực tế ngày 2026-07-18. Mọi kết luận dựa trên file thật trong repo `Z:\01_PROJECTS\apps\tibrain`
> (module `github.com/ti/router/tibrain`, go 1.26.1). Không suy diễn.

## 1. Audit summary (verified)

| Hạng mục | Thực trạng (file thật) |
|---|---|
| Entry point | `main.go` duy nhất ở root, **2627 LOC**, định nghĩa `type Hub`, `func NewHub`, `type Server`, 41 `mux.HandleFunc`. |
| Build | **FAIL**. `main.go:2508` gọi `NewMCPServerManager(...)` nhưng hàm này **không định nghĩa ở đâu** (grep toàn repo = 0). Thêm: `internal/mcp/server.go` sai API mcp-go v0.56.0 (11 `AddTool` 4-arg); `internal/rag` thiếu `go-redis`; `agent/` thiếu package `providers`. |
| AGENTS.md vs thực tế | AGENTS.md §6.4 liệt kê `api_server.go`, `mcp_server.go`, `rag_system.go`, `mcp_registry_tools.go`… nhưng **root chỉ có `main.go`**. AGENTS.md là docs lỗi thời; không dùng làm căn cứ tồn tại file. |
| `cmd/` | Tồn tại nhưng **0 file `.go`** → chưa có CLI. |
| `bin/` | Chỉ chứa `bash.exe`, `git.exe`, `sh.exe` (MSYS2, không phải CLI TiBrain). |
| `.go` files | 80 file (loại trừ `pkg/mod`). 16 file `*_test.go` (internal/db, frontmatter, knowledge, logger, memory, notionprovider, rag×7, ticonsole×2). |
| `go.mod` | Toàn bộ deps là `// indirect` (chưa `go mod tidy`). mcp-go v0.56.0, modernc.org/sqlite v1.54.0. |
| Phase-required dirs | `internal/{execution,pipeline,verification,trace,prompts,skills,capabilities,policy,prediction,intent,routing,runtime,planner,workflow,learning,storage,metrics}` → **đều CHƯA tồn tại**. |
| Existing internal/ | analytics(7), api(1), async(1), config(1), core(1), crypto(1), db(9+2test), frontmatter(3+1test), knowledge(10+2test), logger(2+1test), mcp(1,broken), mcpclient(4), memory(2+1test), notionprovider(3+1test), orchestration(5), prompt(1=F2), rag(13+7test), ticonsole(5+2test), tools(7). |
| `agent/` (root pkg) | `swarm.go`, `failover_pool.go` — BROKEN (import `providers` không tồn tại). |

## 2. Cross-cutting blocker (phải giải quyết trước Phase 1)

**Repo không build được.** Root cause: `NewMCPServerManager` undefined trong `main.go`.
→ Phase 1 không thể chạy `go build ./...` nếu chưa định nghĩa hàm này (hoặc loại bỏ reference).
→ Quyết định: định nghĩa `MCPServerManager` (internal/mcphub hoặc internal/mcp) như một phần của Phase 1/3 groundwork.

## 3. Phase status table

Decision ∈ {KEEP, REFACTOR, REWRITE, CREATE}

| Phase | Module | Current Status | Gap | Decision | Files Affected |
|---|---|---|---|---|---|
| 1 | internal/execution | ABSENT | state machine (received→…→completed/failed/cancelled), ExecutionContext, retry classification | CREATE | internal/execution/ |
| 1 | internal/pipeline | ABSENT | pipeline orchestration qua states | CREATE | internal/pipeline/ |
| 1 | internal/verification | ABSENT | `Verifier` interface, verify không chỉ check non-empty | CREATE | internal/verification/ |
| 1 | internal/trace | ABSENT | trace mỗi bước (intent_detected, tool_called, …) | CREATE | internal/trace/ |
| 1 | main.go (Hub/Server) | EXISTS 2627 LOC, references undefined `NewMCPServerManager` → broken | build break; monolith | REFACTOR | main.go |
| 2 | internal/prompts | ABSENT (chỉ có `internal/prompt` số ít) | PromptSource, Normalizer, capability extraction | CREATE | internal/prompts/ |
| 2 | internal/skills | ABSENT | Skill registry, ranking, risk_level | CREATE | internal/skills/ |
| 2 | internal/capabilities | ABSENT | capability model | CREATE | internal/capabilities/ |
| 2 | internal/prompt (F2) | EXISTS 1 file (Preflight/Feedback/Catalog) | chưa normalize thành skill model; spec yêu cầu `internal/prompts` | REFACTOR | internal/prompt/prompt.go → internal/prompts/ |
| 3 | internal/mcp | EXISTS 1 file (server.go), BROKEN (mcp-go v0.56.0 API) | sửa 11 `AddTool`, align registry+policy | REFACTOR | internal/mcp/server.go |
| 3 | internal/tools | EXISTS 7 files (tool_definitions, local_tool_executor, quality_gate, rtk_handler, register_chrome_mcp, cloudflare_*) | thiếu Tool Registry schema, capability mapping, health, quarantine | REFACTOR | internal/tools/* |
| 3 | internal/policy | ABSENT | policy engine (read/write/execute/network/filesystem/secret/external/dangerous), quarantine | CREATE | internal/policy/ |
| 4 | internal/prediction | ABSENT | PredictionEngine, explainable ranking | CREATE | internal/prediction/ |
| 4 | internal/intent | ABSENT | IntentClassifier interface + rule-based impl | CREATE | internal/intent/ |
| 4 | internal/routing | ABSENT | agent/model/budget selector | CREATE | internal/routing/ |
| 5 | internal/runtime | ABSENT | ReAct loop, checkpoint, cancel/resume | CREATE | internal/runtime/ |
| 5 | internal/planner | ABSENT | plan generation (steps, capabilities, verification_rules, rollback) | CREATE | internal/planner/ |
| 5 | internal/workflow | ABSENT | workflow execution | CREATE | internal/workflow/ |
| 5 | agent/ (root pkg) | EXISTS 2 files, BROKEN (missing `providers`) | tích hợp vào internal/runtime hoặc sửa providers | REFACTOR | agent/swarm.go, agent/failover_pool.go |
| 6 | cmd/ti | ABSENT (`cmd/` trống) | unified CLI entrypoint, mọi command, `--json` | CREATE | cmd/ti/main.go |
| 6 | tibrain/bin/cli | bin/ chỉ có MSYS exe, KHÔNG có CLI source | tạo CLI binary duy nhất | CREATE | cmd/ti/ (hoặc bin/cli) |
| 7 | internal/learning | ABSENT | feedback, score update, promotion/degradation, replay | CREATE | internal/learning/ |
| 7 | internal/storage | ABSENT (có `internal/db` 9 files + 2 test) | dùng `internal/db` (SQLite) làm storage; thêm schema feedback/metrics | REFACTOR | internal/db (+ internal/storage/) |
| 7 | internal/metrics | ABSENT (có `internal/analytics` 7 files) | metrics collection; tái sử dụng analytics nếu phù hợp | REFACTOR/CREATE | internal/metrics/ (hoặc internal/analytics) |

## 4. Ghi chú quy trình

- Mọi module CREATE/REFACTOR phải có interface rõ ràng (contract-first) và test (`*_test.go`).
- `go build ./...` + `go test ./...` là gate bắt buộc mỗi phase.
- Khi tạo file dưới `internal/`, do `.gitignore` có rule `internal/` (tạo sau commit đầu), phải `git add -f` cho file mới.
- Remote: local repo = `thien123331/tibrain`; `LDoVietNam/tibrain` đã archive (stale). Cần làm rõ đích push trước khi commit phase. Tạm ghi nhận, không block local implement.
