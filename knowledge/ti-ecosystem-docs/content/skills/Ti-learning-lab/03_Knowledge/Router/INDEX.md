---
tags: ["tibrain", "documentation", "router", "skill", "javascript"]
scopes: ["cli", "tibrain"]
last_updated: 2026-05-22
---
# Router Knowledge Base

> **Folder**: Router Knowledge  
> **Purpose**: Tổng hợp kiến thức về router architecture, patterns, và implementations
> **Language**: English (see [`../ti-router/`](../ti-router/) for Vietnamese version)

---

## �️ Active Plans (2026-05-05)

3 kế hoạch chính đang thực thi, tách biệt theo scope:

| Plan | Scope | Status | Start Phase |
|------|-------|--------|-------------|
| [`ROUTER_PLAN.md`](./ROUTER_PLAN.md) | Router completion — LLM gateway | Planning | R-0 (lint fix, URGENT) |
| [`TI_CLAW_PLAN.md`](./TI_CLAW_PLAN.md) | Ti Claw agent framework | Planning | A0 (study & decide) |
| [`TI_CLI_PLAN.md`](./TI_CLI_PLAN.md) | Ti CLI + multi-CLI orchestration | Planning | CLI-0 (audit & design) |

### Quick Navigation

- **Router issues (lint, UI, perf)** → `ROUTER_PLAN.md`
- **Build agent framework** → `TI_CLAW_PLAN.md`
- **Multi-CLI orchestration, sub-agents, registry** → `TI_CLI_PLAN.md`

### Context Sources (Primary)
- `router-context.json` — Router architecture (28 providers, 12 OAuth)
- `cli-context.json` — Ti CLI architecture (43/58 modules)
- `automation-context.json` — Automation framework
- `dashboard-context.json` — Dashboard

---

## �� Structure

```
Router/
├── 9router/                      # 9Router (JavaScript) research notes
│   ├── 00-9router-complete.md
│   ├── 01-account-selection-fallback.md
│   ├── 02-combo-system.md
│   ├── 03-oauth-token-refresh.md
│   ├── 04-format-translation-sse.md
│   ├── 05-proxy-usage.md
│   └── README.md
├── freellmapi/                   # FreeLLMAPI (TypeScript) research notes
│   ├── encrypted-key-storage.md
│   ├── enhanced-analytics.md
│   ├── health-probes.md
│   ├── integration-guide.md
│   ├── per-key-rate-tracking.md
│   └── sticky-sessions.md
├── oauth-session-pool/           # OAuth Session Pool & MCP Tool Routing
│   ├── analysis.md               # Comprehensive analysis
│   └── README.md                 # Folder overview
├── ClawRouter_ANALYSIS.md        # ClawRouter analysis
├── adaptive-routing-guide.md     # Adaptive routing guide
├── GO_HANDLER_PATTERNS.md        # Go HTTP handler patterns (Vietnamese)
├── GO_SECURITY_BEST_PRACTICES.md # Go security best practices (error handling, input validation, secrets management)
└── INDEX.md                      # This file
```

---

## 📚 Knowledge Categories

### 1. 9Router (JavaScript Reference)
**Purpose**: Account selection, fallback, OAuth flows

|| File | Description | Key Learnings |
||------|-------------|---------------|
|| `01-account-selection-fallback.md` | Multi-account selection & fallback | Mutex, model-level locking, exponential backoff |
|| `03-oauth-token-refresh.md` | OAuth token refresh | PKCE, device code flow, per-provider refresh |

**Status**: ✅ Research complete

---

### 2. FreeLLMAPI (TypeScript Reference)
**Purpose**: Rate limiting, sticky sessions, analytics

|| File | Description | Implementation Status |
||------|-------------|----------------------|
|| `per-key-rate-tracking.md` | Per-key RPM/RPD/TPM/TPD tracking | ✅ Implemented in Ti Router |
|| `sticky-sessions.md` | Multi-turn session consistency | ✅ Implemented in Ti Router |
|| `enhanced-analytics.md` | Latency, success rate, error distribution | ⚠️ Core implemented, needs DB migration |

**Status**: ✅ Research complete, partial implementation

---

### 3. OAuth Session Pool & MCP Tool Routing
**Purpose**: OAuth session pool management, MCP tool routing

|| File | Description | Status |
||------|-------------|--------|
|| `oauth-session-pool/analysis.md` | Comprehensive analysis, implementation plan | ✅ Research complete |
|| `oauth-session-pool/README.md` | Folder overview | ✅ Complete |

**Gap**: ❌ Not implemented yet

**Implementation Plan**:
- Phase 1: OAuth Session Pool (5 steps)
- Phase 2: MCP Tool Routing (3 steps)

**Cost**: $584-880/year (Option 1: Add to Ti Router)

---

### 4. Other Analysis
|| File | Description |
||------|-------------|
|| `ClawRouter_ANALYSIS.md` | ClawRouter analysis |
|| `adaptive-routing-guide.md` | Adaptive routing guide |

### 5. Go Patterns (Vietnamese)
|| File | Description | Status |
||------|-------------|--------|
|| `GO_HANDLER_PATTERNS.md` | Go HTTP handler patterns, best practices, lessons learned | ✅ Complete |
|| `GO_SECURITY_BEST_PRACTICES.md` | Go security best practices, error handling, input validation, secrets management | ✅ Complete |

**Key Learnings**:
- Dependency injection patterns
- Interface design (accept interfaces, return structs)
- Thread safety with mutex
- Error handling with context
- Handler organization patterns
- Secure error handling (Split Brain pattern)
- Input validation and sanitization
- Secrets management best practices
- Resource limits and DoS prevention
- Security audit checklist

---

## 🎯 Implementation Status

|| Feature | 9Router | FreeLLMAPI | Ti Router |
||---------|---------|-----------|-----------|
|| OAuth Flow | ✅ | N/A | ✅ (partial) |
|| Multi-account per provider | ✅ | N/A | ❌ |
|| Account selection strategy | ✅ | N/A | ❌ |
|| Model-level locking | ✅ | N/A | ❌ |
|| Exponential backoff | ✅ | N/A | ❌ |
|| Per-key rate tracking | N/A | ✅ | ✅ |
|| Sticky sessions | N/A | ✅ | ✅ |
|| Enhanced analytics | N/A | ✅ | ⚠️ (core) |
|| OAuth session pool | ✅ | N/A | ❌ |
|| MCP tool routing | N/A | N/A | ❌ |

---

## 📝 References

### Ti Router Implementation
- `Z:\Ti\router\` - Main router codebase
- `layers/authentication\` - OAuth infrastructure
- `layers/authentication/rate_tracker.go` - Per-key rate tracking
- `layers/authentication/sticky_session.go` - Sticky sessions
- `layers/authentication/analytics.go` - Enhanced analytics
- `layers/http/mcp/mcp.go` - MCP handler (placeholder)

### External Repositories
- `Z:\Ti\Ti-learning-lab\05_Repositories\router\9router\` - 9Router source
- `Z:\Ti\Ti-learning-lab\05_Repositories\router\freellmapi-main\` - FreeLLMAPI source

---

## ✅ Next Steps

1. **OAuth Session Pool**: Implement Phase 1 (5 steps)
2. **MCP Tool Routing**: Implement Phase 2 (3 steps)
3. **DB Migration**: Complete enhanced analytics migration

---

**Last Updated**: 2026-04-29  
**Agent**: Claude Code
