---
tags: ["tibrain", "provider", "documentation", "router", "skill"]
scopes: ["tibrain"]
last_updated: 2026-05-22
---
# Cookie Providers Documentation

**Last Updated**: 2026-05-08
**Scope**: Ti Router Cookie Providers

## 🎯 Overview

Ti Router supports **14 cookie providers** across 2 management systems:
- **API-based (8)**: Managed via HTTP API endpoints (`/api/cookie-providers/*`)
- **Env-based (2)**: Managed via environment variables in `providers.yaml`
- **LLMCookieBridge (4)**: Web-based cookie providers from LLMCookieBridge library

## 📋 Provider List

### Multi-Tier Providers (Appear in multiple categories)

| Provider | API Key | OAuth | Cookie | Cookie Type | Docs |
|----------|---------|-------|--------|-------------|------|
| **Gemini** | ✅ GOOGLE_API_KEY | - | ✅ | API-based | [gemini.md](./gemini.md) |
| **Gemini Web** | - | - | ✅ | LLMCookieBridge | [gemini-web.md](./gemini-web.md) |
| **ChatGPT** | - | ✅ | ✅ | API-based | [chatgpt.md](./chatgpt.md) |
| **ChatGPT Web** | - | - | ✅ | LLMCookieBridge | [chatgpt-web.md](./chatgpt-web.md) |
| **Claude** | ✅ ANTHROPIC_API_KEY | ✅ | ✅ | API-based | [claude.md](./claude.md) |
| **Claude Web** | - | - | ✅ | LLMCookieBridge | [claude-web.md](./claude-web.md) |
| **Perplexity** | ✅ PERPLEXITY_API_KEY | - | ✅ | API-based | [perplexity.md](./perplexity.md) |
| **Perplexity Web** | - | - | ✅ | LLMCookieBridge | [perplexity-web.md](./perplexity-web.md) |
| **Copilot** | - | ✅ | ✅ | API-based | [copilot.md](./copilot.md) |
| **Windsurf** | ✅ WINDSURF_API_KEY | ✅ | ✅ | API-based | [windsurf.md](./windsurf.md) |
| **Notion** | ✅ NOTION_INTEGRATION_TOKEN | - | ✅ | Env-based | [notion.md](./notion.md) |

### Cookie-Only Providers

| Provider | Cookie Type | Management | Docs |
|----------|-------------|------------|------|
| **Meta AI** | API-based | HTTP API | [meta.md](./meta.md) |
| **HuggingChat** | API-based | HTTP API | [huggingchat.md](./huggingchat.md) |

## 🔧 Management Systems

### 1. API-Based Cookie Providers (8 providers)

Managed via HTTP API endpoints in `cookie_providers.go`:

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/cookie-providers/list` | GET | List all cookie providers |
| `/api/cookie-providers/set` | POST | Set cookie for a provider |
| `/api/cookie-providers/get` | GET | Get specific cookie provider |
| `/api/cookie-providers/delete` | DELETE | Delete cookie for a provider |
| `/api/cookie-providers/validate` | POST | Validate all cookie providers |
| `/api/cookie-providers/best-model` | GET | Suggest best model for task |

**Providers**: Gemini, Copilot, Perplexity, Meta, HuggingChat, ChatGPT, Claude, Windsurf

### 2. Env-Based Cookie Providers (2 providers)

Managed via environment variables in `providers.yaml`:

| Provider | Env Variable | Used By |
|----------|--------------|---------|
| **Notion Cookie** | `NOTION_COOKIE` | notion-cookie, notion providers |

### 3. LLMCookieBridge Providers (4 providers)

Web-based cookie providers from LLMCookieBridge library:

| Provider | Cookie Required | Source |
|----------|----------------|--------|
| **Gemini Web** | `__Secure-1PSID`, `__Secure-1PSIDTS` | LLMCookieBridge |
| **ChatGPT Web** | `__Secure-next-auth.session-token` | LLMCookieBridge |
| **Claude Web** | Cookie header with `sessionKey=...` | LLMCookieBridge |
| **Perplexity Web** | `__Secure-next-auth.session-token` | LLMCookieBridge |

**Reference**: https://github.com/tkgo11/LLMCookieBridge

## 🎯 Provider Tiers

A single provider can appear in **multiple tiers**:

### Example: Claude
- **API Key Provider**: `ANTHROPIC_API_KEY` in router.env
- **OAuth Provider**: Claude Code OAuth (GitHub, GitLab-style)
- **Cookie Provider**: Claude Web cookie header

### Example: ChatGPT
- **OAuth Provider**: OpenAI Codex OAuth
- **Cookie Provider**: ChatGPT Web session token

### Example: Perplexity
- **API Key Provider**: `PERPLEXITY_API_KEY` in router.env
- **Cookie Provider**: Perplexity Web session token

## 📝 Usage Guidelines

### When to Use Cookie Providers

Use cookie-based authentication when:
1. You have an existing browser session
2. API key is not available or rate-limited
3. You want to use web-based features (free tiers, etc.)
4. For experimentation and testing

### Security Considerations

⚠️ **WARNING**: Cookie-based authentication targets reverse-engineered web endpoints that may change without notice. Use for:
- Experimentation
- Internal tools
- Migration utilities
- Research workflows

**NOT** for:
- Production SLA
- Long-term guaranteed integration
- Bypassing provider terms

### Cookie Refresh

Most cookie providers support automatic refresh:
- **API-based**: Use `/api/cookie-providers/validate` endpoint
- **LLMCookieBridge**: Built-in `refresh()` method with custom callbacks

## 🔗 Related Documentation

- [Ti Router Providers Config](../../../../Ti/apps/core/router/configs/providers.yaml)
- [Cookie Provider Implementation](../../../../Ti/apps/core/router/cmd/routerd/cookie_providers.go)
- [CONFIG_GUIDE.md](../../../../00_SECRET/CONFIG_GUIDE.md)
- [LLMCookieBridge GitHub](https://github.com/tkgo11/LLMCookieBridge)

---

**Generated by**: Devin CLI
**Date**: 2026-05-08
