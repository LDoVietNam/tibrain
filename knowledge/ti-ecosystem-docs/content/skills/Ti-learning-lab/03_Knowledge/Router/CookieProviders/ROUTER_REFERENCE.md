---
tags: ["documentation", "tibrain", "provider", "pattern-auth-cookie", "authentication"]
scopes: ["auth", "tibrain"]
last_updated: 2026-05-22
---
# Router Reference - Learn from Other Routers

**Last Updated**: 2026-05-08
**Purpose**: Document integration patterns from other AI routers

## 🎯 Overview

This document catalogs AI router implementations to learn integration patterns for cookie providers, OAuth, and multi-tier authentication.

## 📚 Router Catalog

### 1. 9Router (Node.js)

**GitHub**: https://github.com/decolua/9router
**Language**: Node.js/TypeScript
**Port**: 20128
**Type**: AI Router + Token Saver

#### Supported Providers
- OAuth: Claude Code, Antigravity, Codex, GitHub, Cursor
- Free: Kiro AI, OpenCode Free, Vertex AI
- API Key (40+): OpenRouter, GLM, Kimi, MiniMax, OpenAI, Anthropic, Gemini, DeepSeek, Groq, xAI, Mistral, Perplexity, Together, Fireworks, Cerebras, Cohere, NVIDIA, SiliconFlow

#### Key Features
- RTK Token Saver (20-40% token savings)
- 3-Tier Fallback (Subscription → Cheap → Free)
- Multi-account support
- Auto token refresh
- Format translation (OpenAI ↔ Claude ↔ Gemini)

#### Integration Patterns to Learn

**Cookie Handling** (`src/providers/claude.ts`):
```typescript
// Claude cookie-based authentication
class ClaudeProvider {
  async authenticate(cookie: string) {
    // Parse cookie header
    // Extract session token
    // Handle refresh
  }
}
```

**OAuth Flow** (`src/oauth/`):
- Standard OAuth 2.0 implementation
- Token refresh mechanism
- Error handling for expired tokens

**Provider Abstraction** (`src/providers/base.ts`):
- Unified provider interface
- Format translation layer
- Error handling patterns

#### Reference Files
- `src/providers/` - Provider implementations
- `src/oauth/` - OAuth implementations
- `src/formats/` - Format translators
- `.env.example` - Environment configuration

### 2. LLMCookieBridge (Python)

**GitHub**: https://github.com/tkgo11/LLMCookieBridge
**Language**: Python
**Type**: Cookie-based AI client library

#### Supported Providers
- Google Gemini web
- ChatGPT / OpenAI web
- Claude web
- Perplexity web

#### Key Features
- Unified async interface
- Cookie/session bridge
- Streaming support
- Best-effort session refresh
- Custom refresh callbacks
- Minimal dependency footprint

#### Integration Patterns to Learn

**Cookie Authentication** (`src/llm_cookie_bridge/providers/claude.py`):
```python
class ClaudeProvider:
    def __init__(self, cookie_header: str):
        # Parse cookie header string
        # Extract sessionKey
        # Discover organization UUID

    async def refresh(self):
        # Discover active organization UUID
        # Update session state
```

**Session Management** (`src/llm_cookie_bridge/providers/gemini.py`):
```python
class GeminiProvider:
    async def _bootstrap(self):
        # Extract SNlM0e access token
        # Extract build label (bl)
        # Extract session id (f.sid)
        # Extract language metadata
```

**Refresh Callbacks**:
```python
async def refresh_cookies(provider_name: str):
    # Custom cookie renewal logic
    return CookieRefreshResult(
        cookies={"sessionKey": "new-value"},
        metadata={"source": "external-secret-store"}
    )
```

#### Reference Files
- `src/llm_cookie_bridge/providers/` - Provider implementations
- `README.md` - Usage documentation
- `tests/` - Test patterns

### 3. OpenRouter (Go)

**GitHub**: https://github.com/openrouter/openrouter (Not open source)
**Language**: Go
**Type**: AI Model Gateway

#### Supported Providers
- 40+ AI providers via unified API

#### Key Features
- Unified API interface
- Model catalog
- Cost tracking
- Rate limiting

#### Integration Patterns to Learn

**Provider Abstraction**:
- Single API endpoint for all providers
- Model catalog with pricing
- Unified response format

**Authentication**:
- API key based
- OAuth support for some providers

#### Reference
- API Documentation: https://openrouter.ai/docs
- Not open source, but patterns documented in API docs

### 4. Kilo CLI (TypeScript)

**GitHub**: https://github.com/Kilo-Org/kilo
**Language**: TypeScript
**Type**: AI Coding Agent (fork of OpenCode)

#### Supported Providers
- OpenAI-compatible endpoints
- Custom provider configuration
- MCP-based providers

#### Key Features
- Plugin system
- MCP integration
- Multi-session support
- Agent Manager

#### Integration Patterns to Learn

**Provider Configuration** (`packages/opencode/src/config/`):
- JSON/JSONC config format
- Multiple config locations with precedence
- Config merging strategy

**Plugin System** (`plugins/`):
- Dynamic plugin loading
- Plugin configuration
- Plugin lifecycle management

#### Reference Files
- `packages/opencode/src/config/config.ts` - Config system
- `packages/kilo-gateway/` - Provider gateway
- `.opencode/` - Plugin directories

### 5. Cursor (Proprietary)

**Website**: https://cursor.sh
**Type**: AI Code Editor
**Language**: TypeScript/React

#### Supported Providers
- Claude Code (OAuth)
- OpenAI (API key)
- Custom endpoints

#### Integration Patterns to Learn

**OAuth Integration**:
- Browser-based OAuth flow
- Token storage in local config
- Auto token refresh

**Provider Selection**:
- Model selection UI
- Provider configuration
- Fallback mechanisms

## 🔍 Integration Patterns Summary

### Cookie Authentication Patterns

| Pattern | Source | Description |
|---------|--------|-------------|
| **Cookie Header Parsing** | LLMCookieBridge | Parse full cookie header string |
| **Session Bootstrap** | LLMCookieBridge | Extract tokens from web app state |
| **Cookie Refresh** | 9Router | Auto-refresh with custom callbacks |
| **Session Recovery** | LLMCookieBridge | Recover from expired sessions |
| **Cookie Validation** | 9Router | Validate cookie health and expiry |

### OAuth Authentication Patterns

| Pattern | Source | Description |
|---------|--------|-------------|
| **OAuth 2.0 Flow** | 9Router | Standard OAuth 2.0 implementation |
| **Token Storage** | Kilo CLI | Store tokens in local config |
| **Auto Refresh** | 9Router | Automatic token refresh before expiry |
| **Multi-Provider OAuth** | 9Router | Support multiple OAuth providers |

### Provider Abstraction Patterns

| Pattern | Source | Description |
|---------|--------|-------------|
| **Unified Interface** | LLMCookieBridge | Single interface for all providers |
| **Format Translation** | 9Router | Translate between OpenAI/Claude/Gemini formats |
| **Provider Registry** | 9Router | Dynamic provider registration |
| **Config Precedence** | Kilo CLI | Multiple config locations with precedence |

## 🚀 Applying Patterns to Ti Router

### Cookie Provider Integration

**Learn from**: LLMCookieBridge

**Implementation Steps**:
1. Parse cookie header string (not individual cookies)
2. Bootstrap session to extract internal tokens
3. Implement custom refresh callbacks
4. Handle session recovery on expiry

**Reference**: `../../../../Ti/apps/core/router/layers/providers/cookie/`

### OAuth Provider Integration

**Learn from**: 9Router

**Implementation Steps**:
1. Implement OAuth 2.0 flow
2. Store tokens in secure location
3. Implement auto-refresh mechanism
4. Handle token expiry gracefully

**Reference**: `../../../../Ti/apps/core/router/layers/authentication/`

### Multi-Tier Provider Support

**Learn from**: 9Router + LLMCookieBridge

**Implementation Steps**:
1. Create provider registry with multiple auth methods
2. Implement auth method selection logic
3. Support fallback between auth methods
4. Unified interface for all auth types

**Reference**: `../../../../Ti/apps/core/router/layers/provider/`

## 🔄 Auto-Update Mechanism

### GitHub Actions Setup

Create workflow to monitor reference router updates:

**File**: `.github/workflows/monitor-router-updates.yml`

```yaml
name: Monitor Router Updates

on:
  schedule:
    - cron: '0 0 * * *'  # Daily check
  workflow_dispatch:

jobs:
  monitor-9router:
    runs-on: ubuntu-latest
    steps:
      - name: Check 9Router updates
        run: |
          curl -s https://api.github.com/repos/decolua/9router/commits | jq .

  monitor-llmcookiebridge:
    runs-on: ubuntu-latest
    steps:
      - name: Check LLMCookieBridge updates
        run: |
          curl -s https://api.github.com/repos/tkgo11/LLMCookieBridge/commits | jq .

  notify-router-agent:
    needs: [monitor-9router, monitor-llmcookiebridge]
    runs-on: ubuntu-latest
    steps:
      - name: Notify Router Agent
        run: |
          curl -X POST http://localhost:1807/api/router-agent/check-updates
```

### Router Agent Update Logic

**Implementation**: Router agent analyzes changes and suggests updates.

**Update Flow**:
1. Detect reference router update
2. Fetch diff/changes
3. Analyze impact on Ti Router
4. Generate patch suggestions
5. Request approval for application

**File**: `../../../../Ti/apps/core/router/cmd/routerd/router_agent.go`

```go
func (ra *RouterAgent) CheckRouterUpdates() error {
    // Fetch latest commits from reference routers
    // Compare with current implementation
    // Generate update suggestions
    // Store in memory for review
}
```

## 📝 Provider-Specific References

### Claude
- **9Router**: `src/providers/claude.ts` (OAuth + Cookie)
- **LLMCookieBridge**: `src/llm_cookie_bridge/providers/claude.py` (Cookie)

### Gemini
- **9Router**: `src/providers/gemini.ts` (OAuth + Cookie)
- **LLMCookieBridge**: `src/llm_cookie_bridge/providers/gemini.py` (Cookie)

### ChatGPT
- **9Router**: `src/providers/openai.ts` (OAuth + Cookie)
- **LLMCookieBridge**: `src/llm_cookie_bridge/providers/chatgpt.py` (Cookie)

### Perplexity
- **9Router**: `src/providers/perplexity.ts` (API Key + Cookie)
- **LLMCookieBridge**: `src/llm_cookie_bridge/providers/perplexity.py` (Cookie)

## 🔗 Quick Reference Links

| Router | GitHub | Docs | Key Files |
|--------|--------|------|-----------|
| **9Router** | https://github.com/decolua/9router | https://9router.com | `src/providers/` |
| **LLMCookieBridge** | https://github.com/tkgo11/LLMCookieBridge | README.md | `src/llm_cookie_bridge/providers/` |
| **Kilo CLI** | https://github.com/Kilo-Org/kilo | Docs in repo | `packages/kilo-gateway/` |
| **OpenRouter** | - | https://openrouter.ai/docs | API docs |

---

**Generated by**: Devin CLI
**Date**: 2026-05-08
