# Plan: Port Devin CLI Python → Go

> Ngày: 2026-04-29 | Agent: claude | Dự án: github.com/ti/cli/pkg/devin

## Cấu trúc Package

```
pkg/devin/
├── client.go      # HTTP core, V1/V3 URL routing
├── config.go      # Multi-profile config, env var precedence
├── types.go       # Session, Message, APIError
├── options.go     # NewClient Option pattern
├── sessions.go    # Session API
├── knowledge.go   # Knowledge API
├── playbooks.go   # Playbook API
├── secrets.go     # Secret API
├── schedules.go   # Schedule API
├── repos.go       # Repository API
└── *_test.go      # Unit tests
```

## Design Decisions

| Decision | Lý do |
|----------|-------|
| Option pattern (`WithXxx`) | Dễ mở rộng, backward compatible |
| Service structs (`Sessions`, `Knowledge`) | Tách domain, dễ test |
| Config riêng package-level | Dùng standalone hoặc inject |
| V3 default, V1 legacy | Match Python implementation |
| Context-aware polling | Caller control timeout |

## Implementation Order

1. `types.go` + `options.go` (30 min)
2. `config.go` (45 min)
3. `client.go` (60 min)
4. `sessions.go` (90 min)
5. Unit tests (60 min)
6. Các service khác (từng cái 20-30 min)
7. CLI commands (90 min)

## Key Patterns từ Research

- `go-anthropic`: `NewClient(apiKey)`, typed errors (`errors.As`)
- `agent-sdk-go`: Option pattern, modular architecture
- Python devin-cli: ClientProxy V1/V3, 403/404 fallback
