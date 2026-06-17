---
tags: ["tibrain", "provider-claude", "authentication", "security", "documentation"]
scopes: ["providers", "auth", "tibrain"]
last_updated: 2026-05-22
---
# Claude Cookie Provider

**Provider**: Claude
**Type**: Multi-Tier (API Key + OAuth + Cookie)
**Cookie Type**: API-based
**Last Updated**: 2026-05-08

## 🎯 Overview

Claude appears in **3 provider tiers** in Ti Router:
1. **API Key Provider**: `ANTHROPIC_API_KEY` (official API)
2. **OAuth Provider**: Claude Code OAuth (browser-based)
3. **Cookie Provider**: Claude Web cookie header (web-based)

## 🔑 Authentication Methods

### 1. API Key Authentication (Recommended)

**Environment Variable**: `ANTHROPIC_API_KEY`
**Location**: `Z:\00_SECRET\router.env`
**Status**: ⚠️ Not Configured

```bash
ANTHROPIC_API_KEY=sk-ant-xxxx
```

**Pros**:
- Official API
- Stable endpoints
- Production-ready
- Full feature support

**Cons**:
- Requires API key
- Rate limits apply
- Billing based on usage

### 2. OAuth Authentication

**OAuth Provider**: Claude Code
**OAuth Endpoint**: https://github.com/anthropics/claude-code (via OAuth flow)
**Status**: ⚠️ Not Configured

**OAuth Config**:
```bash
CLAUDE_OAUTH_CLIENT_ID=
CLAUDE_OAUTH_CLIENT_SECRET=
CLAUDE_OAUTH_TOKEN_FILE=Z:\06_AUTH\claude.oauth
```

**Pros**:
- Browser-based authentication
- No API key management
- Auto token refresh

**Cons**:
- Requires OAuth setup
- More complex configuration
- Depends on OAuth flow

### 3. Cookie Authentication (Web-Based)

**Cookie Type**: API-based
**Management**: HTTP API endpoints
**Status**: ⚠️ Not Set

**API Endpoints**:
```bash
# Set Claude cookie
POST /api/cookie-providers/set
{
  "provider": "claude",
  "cookie": {
    "sessionKey": "your-session-key"
  },
  "expires_in": 2592000  # 30 days
}

# Get Claude cookie status
GET /api/cookie-providers/get?provider=claude

# Validate Claude cookie
POST /api/cookie-providers/validate
```

**Cookie Format**:
- Full cookie header string containing `sessionKey=...`
- Example: `sessionKey=xxxx; other-cookies=yyyy`

**Pros**:
- Uses existing browser session
- No API key needed
- Access to web-based features
- Free tier access possible

**Cons**:
- Targets reverse-engineered endpoints
- May break without notice
- Not production-ready
- Security risks

## 📋 Implementation Details

### Cookie Provider Registration

**File**: `../../../../Ti/apps/core/router/cmd/routerd/cookie_providers.go`

```go
var cookieProviders = map[string]*providers.CookieProvider{
    "claude": providers.ClaudePro,
}
```

### Cookie Structure

**Required Cookies**:
- `sessionKey` - Primary authentication cookie
- Other Claude-specific cookies may be required

### Models Supported

- claude-sonnet-4-5
- claude-opus-4-5
- claude-haiku-4-5

## 🔧 Usage Examples

### Setting Cookie via API

```bash
curl -X POST http://localhost:1807/api/cookie-providers/set \
  -H "Content-Type: application/json" \
  -d '{
    "provider": "claude",
    "cookie": {
      "sessionKey": "your-session-key-here"
    },
    "expires_in": 2592000
  }'
```

### Checking Cookie Status

```bash
curl http://localhost:1807/api/cookie-providers/get?provider=claude
```

### Validating All Cookies

```bash
curl -X POST http://localhost:1807/api/cookie-providers/validate
```

## 🌐 Claude Web (LLMCookieBridge Alternative)

**Source**: https://github.com/tkgo11/LLMCookieBridge
**Cookie Type**: LLMCookieBridge
**Status**: ⚠️ Not Implemented in Ti Router

**Cookie Required**:
```python
cookie_header=os.environ["CLAUDE_COOKIE_HEADER"]
```

**Cookie Format**: Full cookie header string with `sessionKey=...`

**Python Example**:
```python
from llm_cookie_bridge import LLMCookieBridge

bridge = LLMCookieBridge.create(
    "claude",
    cookie_header=os.environ["CLAUDE_COOKIE_HEADER"],
)

async with bridge:
    response = await bridge.chat("Say hello")
    print(response.text)
```

## ⚠️ Security Considerations

**Cookie Authentication Risks**:
- Targets reverse-engineered web endpoints
- May break without notice
- Not suitable for production
- Session expiration issues
- Potential account flagging

**Best Practices**:
- Use API key authentication for production
- Use cookie authentication only for testing/experimentation
- Implement proper cookie refresh logic
- Monitor for endpoint changes
- Have fallback to API key

## 📊 Comparison

| Method | Stability | Production Ready | Cost | Setup Complexity |
|--------|-----------|------------------|------|------------------|
| **API Key** | ✅ High | ✅ Yes | Paid | Low |
| **OAuth** | ✅ High | ✅ Yes | Paid | Medium |
| **Cookie (API-based)** | ⚠️ Low | ❌ No | Free/Varies | Low |
| **Cookie (LLMCookieBridge)** | ⚠️ Low | ❌ No | Free/Varies | Medium |

## 🔗 Related Documentation

- [Claude Official API](https://docs.anthropic.com)
- [Claude Code](https://claude.ai/code)
- [LLMCookieBridge Claude Docs](https://github.com/tkgo11/LLMCookieBridge#claude)
- [Ti Router Providers Config](../../../../Ti/apps/core/router/configs/providers.yaml)

## 🚀 Integration into Ti Router

### Step 1: Add Provider to cookie_providers.go

**File**: `../../../../Ti/apps/core/router/cmd/routerd/cookie_providers.go`

```go
var cookieProviders = map[string]*providers.CookieProvider{
    "claude": providers.ClaudePro,
}
```

### Step 2: Implement Cookie Provider

**File**: `../../../../Ti/apps/core/router/layers/providers/cookie/`

Create or update cookie provider implementation:
```go
// ClaudePro cookie provider
var ClaudePro = &providers.CookieProvider{
    Name:          "claude",
    Type:          "cookie",
    Domain:        "claude.ai",
    RequiredKeys:  []string{"sessionKey"},
    Models:        []string{"claude-sonnet-4-5", "claude-opus-4-5"},
    Status:        "pending",
    Priority:      15,
}
```

### Step 3: Add HTTP Endpoints

**File**: `../../../../Ti/apps/core/router/cmd/routerd/cookie_providers.go`

Endpoints are already implemented:
- `POST /api/cookie-providers/set` - Set cookie
- `GET /api/cookie-providers/get` - Get cookie status
- `POST /api/cookie-providers/validate` - Validate cookie

### Step 4: Test Integration

```bash
# Set Claude cookie
curl -X POST http://localhost:1807/api/cookie-providers/set \
  -H "Content-Type: application/json" \
  -d '{
    "provider": "claude",
    "cookie": {"sessionKey": "your-session-key"},
    "expires_in": 2592000
  }'

# Validate cookie
curl -X POST http://localhost:1807/api/cookie-providers/validate

# Test model access
curl -H "Authorization: Bearer sk-jarvis-dev" \
  http://localhost:1807/v1/chat/completions \
  -d '{
    "model": "claude/claude-sonnet-4-5",
    "messages": [{"role": "user", "content": "Hello"}]
  }'
```

## 📚 Learn from Other Routers

### 9Router (Node.js)
**GitHub**: https://github.com/decolua/9router

**Claude Integration**: 9Router supports Claude via OAuth and custom endpoints.
- Reference: `src/providers/claude.ts`
- Cookie handling: Uses session tokens
- Auto-refresh: Built-in token refresh

**Key Learnings**:
- Cookie refresh mechanism
- Session management
- Error handling for expired sessions

### LLMCookieBridge (Python)
**GitHub**: https://github.com/tkgo11/LLMCookieBridge

**Claude Integration**: Full cookie-based implementation.
- Reference: `src/llm_cookie_bridge/providers/claude.py`
- Cookie format: Full cookie header string
- Refresh: Custom refresh callback support

**Key Learnings**:
- Cookie header parsing
- Organization UUID discovery
- Session recovery logic

### OpenRouter (Go)
**GitHub**: https://github.com/openrouter/openrouter

**Claude Integration**: API-based with multiple auth methods.
- Reference: Not open source, but API docs available
- Auth: API key + OAuth support

**Key Learnings**:
- Multi-auth support architecture
- Provider abstraction layer

## 🔄 Auto-Update Mechanism

### GitHub Action for Monitoring

**File**: `.github/workflows/monitor-claude-updates.yml`

```yaml
name: Monitor Claude Provider Updates

on:
  schedule:
    - cron: '0 0 * * *'  # Daily check
  workflow_dispatch:

jobs:
  monitor-updates:
    runs-on: ubuntu-latest
    steps:
      - name: Check Anthropic API updates
        run: |
          # Check for API changes
          curl -s https://docs.anthropic.com/api-releases | jq .

      - name: Check 9Router updates
        run: |
          # Check for 9Router Claude implementation updates
          curl -s https://api.github.com/repos/decolua/9router/commits?path=src/providers/claude.ts | jq .

      - name: Check LLMCookieBridge updates
        run: |
          # Check for LLMCookieBridge Claude implementation updates
          curl -s https://api.github.com/repos/tkgo11/LLMCookieBridge/commits?path=src/llm_cookie_bridge/providers/claude.py | jq .

      - name: Notify Router Agent
        run: |
          # Send webhook to router agent to trigger update check
          curl -X POST http://localhost:1807/api/router-agent/check-updates \
            -H "Content-Type: application/json" \
            -d '{"provider": "claude"}'
```

### Router Agent Auto-Update

**Implementation**: Router agent monitors provider updates and suggests code changes.

**Update Flow**:
1. GitHub Action detects provider update
2. Sends webhook to Router Agent
3. Router Agent analyzes changes
4. Generates diff/patch suggestions
5. Applies updates with approval

**File**: `../../../../Ti/apps/core/router/cmd/routerd/router_agent.go`

```go
// Check for provider updates
func (ra *RouterAgent) CheckProviderUpdates(provider string) error {
    // Fetch latest implementation from reference routers
    // Compare with current implementation
    // Generate update suggestions
    // Store in memory for review
}
```

---

**Generated by**: Devin CLI
**Date**: 2026-05-08
