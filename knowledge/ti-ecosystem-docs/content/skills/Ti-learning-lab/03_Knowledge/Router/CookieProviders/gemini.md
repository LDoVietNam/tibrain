---
tags: ["tibrain", "security", "provider-gemini", "documentation", "router"]
scopes: ["providers", "tibrain"]
last_updated: 2026-05-22
---
# Gemini Cookie Provider

**Provider**: Gemini (Google)
**Type**: Multi-Tier (API Key + Cookie)
**Cookie Type**: API-based
**Last Updated**: 2026-05-08

## 🎯 Overview

Gemini appears in **2 provider tiers** in Ti Router:
1. **API Key Provider**: `GOOGLE_API_KEY` (official API)
2. **Cookie Provider**: Gemini Web cookie (web-based)

## 🔑 Authentication Methods

### 1. API Key Authentication (Recommended)

**Environment Variable**: `GOOGLE_API_KEY`
**Location**: `Z:\00_SECRET\router.env`
**Status**: ⚠️ Not Configured

```bash
GOOGLE_API_KEY=AIzaSy-xxxx
```

**Pros**:
- Official Google API
- Stable endpoints
- Production-ready
- Full feature support

**Cons**:
- Requires API key
- Rate limits apply
- Billing based on usage

### 2. Cookie Authentication (Web-Based)

**Cookie Type**: API-based
**Management**: HTTP API endpoints
**Status**: ⚠️ Not Set

**API Endpoints**:
```bash
# Set Gemini cookie
POST /api/cookie-providers/set
{
  "provider": "gemini",
  "cookie": {
    "__Secure-1PSID": "your-1psid",
    "__Secure-1PSIDTS": "your-1psidts"
  },
  "expires_in": 2592000  # 30 days
}

# Get Gemini cookie status
GET /api/cookie-providers/get?provider=gemini

# Validate Gemini cookie
POST /api/cookie-providers/validate
```

**Cookie Format**:
- `__Secure-1PSID` - Primary session ID
- `__Secure-1PSIDTS` - Session timestamp

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

**File**: `../../../../Ti/apps/core/router\cmd\routerd\cookie_providers.go`

```go
var cookieProviders = map[string]*providers.CookieProvider{
    "gemini": providers.GoogleGemini,
}
```

### Cookie Structure

**Required Cookies**:
- `__Secure-1PSID` - Primary session ID
- `__Secure-1PSIDTS` - Session timestamp

**Bootstrap Process**:
Under the hood, Gemini bootstrap extracts web app state:
- `SNlM0e` access token
- Build label (`bl`)
- Session id (`f.sid`)
- Language metadata

### Models Supported

- gemini-2.5-flash
- gemini-2.5-pro
- gemini-1.5-flash

## 🔧 Usage Examples

### Setting Cookie via API

```bash
curl -X POST http://localhost:1807/api/cookie-providers/set \
  -H "Content-Type: application/json" \
  -d '{
    "provider": "gemini",
    "cookie": {
      "__Secure-1PSID": "your-1psid-here",
      "__Secure-1PSIDTS": "your-1psidts-here"
    },
    "expires_in": 2592000
  }'
```

### Checking Cookie Status

```bash
curl http://localhost:1807/api/cookie-providers/get?provider=gemini
```

### Validating All Cookies

```bash
curl -X POST http://localhost:1807/api/cookie-providers/validate
```

## 🌐 Gemini Web (LLMCookieBridge)

**Source**: https://github.com/tkgo11/LLMCookieBridge
**Cookie Type**: LLMCookieBridge
**Status**: ⚠️ Not Implemented in Ti Router

**Cookie Required**:
```python
cookies={
    "__Secure-1PSID": os.environ["GEMINI_1PSID"],
    "__Secure-1PSIDTS": os.environ["GEMINI_1PSIDTS"],
}
```

**Python Example**:
```python
import os
from llm_cookie_bridge import LLMCookieBridge

bridge = LLMCookieBridge.create(
    "gemini",
    cookies={
        "__Secure-1PSID": os.environ["GEMINI_1PSID"],
        "__Secure-1PSIDTS": os.environ["GEMINI_1PSIDTS"],
    },
)

async with bridge:
    response = await bridge.chat("Say hello")
    print(response.text)
```

**Bootstrap Process**:
Gemini bootstrap extracts web app state:
- `SNlM0e` access token
- Build label (`bl`)
- Session id (`f.sid`)
- Language metadata

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
| **Cookie (API-based)** | ⚠️ Low | ❌ No | Free/Varies | Low |
| **Cookie (LLMCookieBridge)** | ⚠️ Low | ❌ No | Free/Varies | Medium |

## 🔗 Related Documentation

- [Google AI Studio](https://aistudio.google.com)
- [Gemini API Docs](https://ai.google.dev/docs)
- [LLMCookieBridge Gemini Docs](https://github.com/tkgo11/LLMCookieBridge#gemini)
- [Ti Router Providers Config](../../../../Ti/apps/core/router/configs/providers.yaml)

## 🚀 Integration into Ti Router

### Step 1: Add Provider to cookie_providers.go

**File**: `../../../../Ti/apps/core/router\cmd\routerd\cookie_providers.go`

```go
var cookieProviders = map[string]*providers.CookieProvider{
    "gemini": providers.GoogleGemini,
}
```

### Step 2: Implement Cookie Provider

**File**: `../../../../Ti/apps/core/router\layers\providers\cookie\`

```go
// GoogleGemini cookie provider
var GoogleGemini = &providers.CookieProvider{
    Name:          "gemini",
    Type:          "cookie",
    Domain:        "google.com",
    RequiredKeys:  []string{"__Secure-1PSID", "__Secure-1PSIDTS"},
    Models:        []string{"gemini-2.5-flash", "gemini-2.5-pro"},
    Status:        "pending",
    Priority:      10,
}
```

### Step 3: Implement Bootstrap Logic

Learn from LLMCookieBridge's Gemini implementation:
```go
func (gp *GeminiProvider) Bootstrap() error {
    // Extract SNlM0e access token
    // Extract build label (bl)
    // Extract session id (f.sid)
    // Extract language metadata
}
```

### Step 4: Test Integration

```bash
# Set Gemini cookie
curl -X POST http://localhost:1807/api/cookie-providers/set \
  -H "Content-Type: application/json" \
  -d '{
    "provider": "gemini",
    "cookie": {
      "__Secure-1PSID": "your-1psid",
      "__Secure-1PSIDTS": "your-1psidts"
    },
    "expires_in": 2592000
  }'
```

## 📚 Learn from Other Routers

### 9Router (Node.js)
**GitHub**: https://github.com/decolua/9router

**Gemini Integration**: 9Router supports Gemini via OAuth and custom endpoints.
- Reference: `src/providers/gemini.ts`
- Cookie handling: Session token extraction
- Auto-refresh: Built-in token refresh

### LLMCookieBridge (Python)
**GitHub**: https://github.com/tkgo11/LLMCookieBridge

**Gemini Integration**: Full cookie-based implementation with bootstrap logic.
- Reference: `src/llm_cookie_bridge/providers/gemini.py`
- Bootstrap: Extracts SNlM0e, build label, session id
- Refresh: Reloads app bootstrap state

## 🔄 Auto-Update Mechanism

See [AUTO_UPDATE_MECHANISM.md](./AUTO_UPDATE_MECHANISM.md) for details.

---

**Generated by**: Devin CLI
**Date**: 2026-05-08
