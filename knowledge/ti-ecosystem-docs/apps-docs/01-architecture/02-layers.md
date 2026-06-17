# Layer Architecture

## Layer Stack

```
┌─────────────────────────────────────────┐
│  HTTP Transport Layer (http/)           │
│  — SSE streaming, CORS, middleware    │
│  — Browser impersonation, nested SSE  │
│  — WebSocket streaming, delta parsing  │
├─────────────────────────────────────────┤
│  Authentication Layer (auth/)           │
│  — OAuth, API keys, sessions          │
│  — OAuth PKCE, JWT, CLI token (AuthGuard)│
│  — Cookie extraction, filtering       │
├─────────────────────────────────────────┤
│  Routing Layer (routing/)               │
│  — autocombo, policy engine, cache     │
│  — UsageTracker integration            │
├─────────────────────────────────────────┤
│  Provider Layer (provider/)             │
│  — OpenAI, Anthropic, OpenRouter, ...  │
│  — Cookie-based authentication        │
├─────────────────────────────────────────┤
│  Monitoring Layer (monitoring/)         │
│  — logs, metrics, PII sanitizer        │
├─────────────────────────────────────────┤
│  Stats Layer (stats/)                  │
│  — UsageTracker, provider/model metrics│
├─────────────────────────────────────────┤
│  Specialized Agents (content/agents/)   │
│  — Cloudflare, OAuth, File Upload      │
│  — Storage, Conversation agents       │
└─────────────────────────────────────────┘
```

## Request Flow

1. **HTTP** receives request → CORS, method validation, AuthGuard protection
2. **Auth** validates credentials → API key, OAuth PKCE, or cookie-based auth
3. **Routing** selects provider/model → autocombo or explicit (with UsageTracker insights)
4. **Provider** forwards to upstream → builds URL, injects key/cookies
5. **Monitoring** records usage → UsageTracker, async log, metrics
6. **Agents** handle specialized tasks → Cloudflare tunnel, OAuth flows, file uploads

## New Components

### HTTP Transport Layer Enhancements
- **Browser Impersonation**: Standardized headers mimicking real browsers
- **Nested SSE Parsing**: Support for complex streaming responses (perplexity-ai pattern)
- **WebSocket Streaming**: Delimiter-based message splitting (sydney.py pattern)
- **Streaming Delta**: Incremental token delivery (sydney.py pattern)

### Authentication Layer Enhancements
- **OAuth PKCE**: Full PKCE implementation with retry logic (auth2api pattern)
- **Cookie Extraction**: Cross-platform browser cookie extraction (WebAI-to-API pattern)
- **Cookie Filtering**: Provider-specific cookie filtering (Gemini-API pattern)
- **AuthGuard**: HTTP handler protection with JWT and CLI token support (9router pattern)

### Stats Layer
- **UsageTracker**: Real-time statistics tracking for providers and models (CLIProxyAPI pattern)
- **Integration with LearningService**: High-confidence insights for adaptive routing

### Specialized Agents
Instead of implementing all functionality in Go, specialized agents handle complex tasks:
- **Cloudflare Agent**: Tunnel/Tailscale access control, DDoS protection
- **OAuth Agent**: OAuth PKCE flows for multiple providers
- **File Upload Agent**: Two-stage upload patterns
- **Storage Agent**: Multi-storage backend management
- **Conversation Agent**: Conversation lifecycle management

## Provider Registry

| Provider | Base URL | Auth Type |
|----------|----------|-----------|
| openrouter | https://openrouter.ai/api/v1 | Bearer |
| groq | https://api.groq.com/openai/v1 | Bearer |
| cerebras | https://api.cerebras.ai/v1 | Bearer |
| deepseek | https://api.deepseek.com/v1 | Bearer |
| google | https://generativelanguage.googleapis.com/v1beta/openai | Bearer/Cookie |
| openai | https://api.openai.com/v1 | Bearer |
| anthropic | https://api.anthropic.com/v1 | Bearer/OAuth/Cookie |
| claude | https://claude.ai | OAuth PKCE |
| gemini | https://generativelanguage.googleapis.com/v1beta | Cookie |
| perplexity | https://www.perplexity.ai | Cookie |

## Local Generation

| Type | Provider | Base URL |
|------|----------|----------|
| Video | ComfyUI | http://localhost:8188 |
| Video | SD WebUI | http://localhost:7860 |
| Music | ComfyUI | http://localhost:8188 |

## Configuration Files

Configuration files for new components:
- `configs/base/cloudflare.yaml` - Cloudflare agent configuration
- `configs/base/oauth.yaml` - OAuth agent configuration
- `configs/base/file-upload.yaml` - File upload agent configuration
- `configs/base/storage.yaml` - Storage agent configuration
- `configs/base/conversation.yaml` - Conversation agent configuration
- `configs/base/monitoring.yaml` - Monitoring and alerting configuration
