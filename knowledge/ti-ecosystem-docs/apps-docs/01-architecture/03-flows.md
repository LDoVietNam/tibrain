# Data Flows

## Chat Completions (Streaming)

```
Client → POST /v1/chat/completions
  → Auth check (API key / OAuth)
  → Model resolution (autocombo → explicit → fallback)
  → Provider selection (openrouter, groq, ...)
  → Build upstream request
  → Forward to LLM provider
  ← SSE stream back
  → Record usage (async)
```

## Model Resolution Chain

```
1. Request "model" field (explicit)
2. Router Agent /decide
3. BEADS v2 engine (enabled + confidence > 0.5)
4. Brain v1 (task-type preference)
5. Config default (cfg.Model)
6. Hardcoded: claude-sonnet-4
```

## Local Generation Flow (ComfyUI)

```
Client → POST /v1/videos/generations
  → Parse provider/model (comfyui/animatediff)
  → Build workflow JSON
  → Submit to ComfyUI /prompt
  → Poll /history/{prompt_id}
  → Extract output files
  → Fetch /view?filename=...
  → Base64 encode
  ← Return {created, data: [{b64_json, format}]}
```

## Authentication Flows

### OAuth PKCE (Google, Anthropic)
```
Client → GET /api/management/oauth
  → Generate code_verifier + code_challenge
  → Build auth URL
  ← Redirect to provider

Provider → GET /callback?code=...
  → Exchange code for token
  → Store token securely
  ← Return success
```

### GitHub Device Code
```
Client → POST device authorization
  → Request device_code + user_code
  ← Return {device_code, user_code, verification_uri}

User → Browse verification_uri, enter user_code

Client → Poll /token with device_code
  ← Return access_token when authorized
```
