---
tags: ["tibrain", "pattern-auth-cookie", "documentation", "http", "provider"]
scopes: ["integration", "tibrain"]
last_updated: 2026-05-22
---
# Cookie-Based AI Provider Implementations from GitHub

## Overview
Research on GitHub repositories that use cookie-based authentication for AI providers to improve Ti Router's cookie provider implementation.

## Key Repositories

### 1. WebAI-to-API (Amm1rr/WebAI-to-API)
**URL:** https://github.com/Amm1rr/WebAI-to-API

**Features:**
- Connects to Gemini web interface using browser cookies
- Exposes as API endpoint without API KEY
- Supports ChatGPT, Claude, DeepSeek, Grok
- Lightweight, fast, efficient for personal use
- Secondary server powered by gpt4free library

**Cookie Implementation:**
- Uses `app/utils/browser.py` to retrieve cookies from browser
- Configuration in `config.conf` for cookie storage
- Auto-retrieval from default browser if cookies are empty
- Supports multiple browsers: firefox, brave, chrome, edge, safari
- Key cookies: `gemini_cookie_1psid`, `gemini_cookie_1psidts`

**Architecture:**
- FastAPI application
- Modular structure: endpoints, services, configs, utilities
- Session management for cookie-based auth
- Gemini client wrapper for API interaction

### 2. CLIProxyAPI (router-for-me/CLIProxyAPI)
**URL:** https://github.com/router-for-me/CLIProxyAPI

**Features:**
- Wraps Gemini CLI, Antigravity, ChatGPT Codex, Claude Code
- Provides OpenAI/Gemini/Claude/Codex compatible API
- Multi-account CLI access
- Supports OAuth for Claude Code
- OpenAI Responses compatible

**Cookie Implementation:**
- Uses CLI-based authentication
- OAuth support for Claude Code
- Multi-account management
- Session-based authentication

### 3. Gemini-API (dsdanielpark/Gemini-API)
**URL:** https://github.com/dsdanielpark/Gemini-API

**Features:**
- Unofficial Python wrapper for Google Gemini
- Uses cookie values for authentication
- REST format responses
- Synchronous clients preferred (rate limiting)
- Derived from Bard API project

**Cookie Implementation:**
```python
from gemini import Gemini

# Auto-collect cookies from browser
client = Gemini(auto_cookies=True)

# Pass specific cookies
client = Gemini(auto_cookies=True, target_cookies="all")
```

**Key Patterns:**
- Auto-cookie collection from browser
- Target cookie selection
- Session management
- Rate limiting awareness

### 4. 9router (decolua/9router)
**URL:** https://github.com/decolua/9router

**Features:**
- Unlimited FREE AI coding
- Connect Claude Code, Codex, Cursor, Cline, Copilot, Antigravity
- 40+ providers
- Auto-fallback
- RTK -40% tokens
- Never hit limits

**Cookie Implementation:**
- Multi-provider cookie support
- GitHub OAuth integration
- Monthly reset system
- Provider switching

## Additional Repositories (Round 2)

### 5. auth2api (AmazingAng/auth2api)
**URL:** https://github.com/AmazingAng/auth2api

**Features:**
- Lightweight Claude OAuth to OpenAI-compatible API proxy
- Multiple providers: Claude OAuth, OpenAI Codex OAuth, Cursor local-login
- Multi-account support with sticky routing
- Per-account cooldown, refresh, and stats
- OpenAI-compatible API endpoints
- Claude native passthrough
- Streaming, tools, images, reasoning support
- Per-account health handling

**Cookie Implementation:**
- OAuth-based authentication (not direct cookies)
- Multiple provider support (anthropic, codex, cursor)
- Token refresh with concurrency lock
- Per-account usage tracking
- Health monitoring and cooldown
- Admin status endpoint for account management

**Key Patterns:**
- OAuth token management
- Multi-account pooling
- Automatic failover
- Per-account health handling
- Token refresh mechanisms

### 6. Claude-API (KoushikNavuluri/Claude-API)
**URL:** https://github.com/KoushikNavuluri/Claude-API

**Features:**
- Unofficial API for Claude AI
- Python-based wrapper
- Cookie-based authentication
- Conversation management
- Message sending with attachments
- Chat history support

**Cookie Implementation:**
```python
# Get cookie from browser developer tools
cookie = os.environ.get('cookie')
claude_api = Client(cookie)
```

**Key Patterns:**
- Browser cookie extraction (network tab or storage tab)
- Environment variable for cookie storage
- Session management
- Conversation ID tracking

### 7. perplexity-ai (helallao/perplexity-ai)
**URL:** https://github.com/helallao/perplexity-ai

**Features:**
- Unofficial API Wrapper for Perplexity.ai
- Account generator with web interface
- MCP Server support
- Web interface for account management
- Streaming responses
- Async usage

**Cookie Implementation:**
- PERPLEXITY_COOKIES environment variable
- Cookie extraction via browser inspector
- cURL to Python conversion for cookies
- Support for pro, reasoning, and deep research modes
- Anonymous mode without cookies (free queries only)

**Key Patterns:**
- Environment variable for cookies
- Browser inspector cookie extraction
- cURL conversion for cookie parsing
- Account generation with cookies
- Mode-based cookie requirements

### 8. sydney.py (vsakkas/sydney.py)
**URL:** https://github.com/vsakkas/sydney.py

**Features:**
- Python Client for Copilot (formerly Bing Chat)
- Cookie-based authentication for regions
- Captcha challenge handling
- Message verification

**Cookie Implementation:**
- Manual message writing for cookie verification
- CaptchaChallenge error handling
- Region-specific cookie requirements
- Success message verification

**Key Patterns:**
- Captcha challenge handling
- Region-specific cookie logic
- Manual verification process
- Error handling for cookie failures

## Updated Common Patterns Across Repositories

### 1. Browser Cookie Extraction Methods
- **Network Tab**: Copy as cURL (bash) → convert to Python
- **Storage Tab**: Direct cookie value extraction
- **Developer Tools**: F12 or Ctrl + Shift + I
- **Multiple Browsers**: Firefox, Chrome, Edge, Brave, Safari support

### 2. Cookie Storage Mechanisms
- **Environment Variables**: PERPLEXITY_COOKIES, cookie
- **Configuration Files**: config.conf, .env
- **Session Management**: Auto-refresh, expiry handling
- **Multi-Account**: Cookie pools, sticky routing

### 3. Authentication Strategies
- **Direct Cookies**: Browser cookie extraction
- **OAuth**: Token-based auth with refresh
- **Hybrid**: Cookie + OAuth fallback
- **Anonymous Mode**: No cookies required (limited features)

### 4. Health & Reliability
- **Cooldown Mechanisms**: Per-account cooldown
- **Token Refresh**: Automatic with concurrency lock
- **Health Monitoring**: Per-account status tracking
- **Failover**: Automatic account switching
- **Rate Limiting**: Per-IP limits, timing-safe validation

## Updated Key Learnings for Ti Router

### 1. Browser Cookie Extraction
- Automatic cookie extraction from default browser
- Support for multiple browsers (Firefox, Chrome, Edge, Brave, Safari)
- Cookie file parsing and normalization

### 2. Cookie Management
- Configuration-based cookie storage
- Auto-refresh mechanisms
- Session management
- Cookie expiry handling

### 3. Proxy Support
- HTTP proxy configuration
- Smart proxy integration (Crawlbase)
- Proxy rotation for rate limiting

### 4. Multi-Provider Support
- Unified API interface for multiple providers
- Provider switching
- Fallback mechanisms
- Load balancing

## Key Learnings for Ti Router

### Cookie Extraction
- Implement automatic browser cookie extraction
- Support multiple browsers
- Parse cookie files correctly

### Session Management
- Implement session refresh
- Handle cookie expiry
- Auto-renewal mechanisms

### Proxy Integration
- Add proxy support for rate limiting
- Smart proxy rotation
- Fallback to direct connection

### Multi-Provider Architecture
- Unified interface for cookie providers
- Provider switching logic
- Health monitoring
- Fallback mechanisms

## Implementation Recommendations

### 1. Browser Cookie Extraction
```go
// Extract cookies from browser
cookies := extractFromBrowser("firefox")
// Normalize cookies
normalized := normalizeCookies(cookies)
```

### 2. Auto-Refresh
```go
// Auto-refresh cookies before expiry
if time.Until(expiresAt) < 1*time.Hour {
    refreshCookies()
}
```

### 3. Proxy Support
```go
// Add proxy for rate limiting
client := &http.Client{
    Transport: &http.Transport{
        Proxy: http.ProxyURL(proxyURL),
    },
}
```

### 4. Health Monitoring
```go
// Monitor cookie health
if validateCookie(cookie) {
    provider.MarkHealthy()
} else {
    provider.MarkUnhealthy()
    refreshCookie()
}
```

## Priority
High - Apply these patterns to improve Ti Router's cookie provider implementation.
