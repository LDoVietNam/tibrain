---
tags: ["tibrain", "router", "provider", "documentation", "cli"]
scopes: ["cli", "tibrain"]
last_updated: 2026-05-22
---
# OpenAI-compatible Providers

Ti CLI now supports chatgpt-cli-style configurable providers through `compatible_providers`.

## API-key provider

```json
{
  "compatible_providers": [
    {
      "name": "local-openai",
      "type": "openai_compatible",
      "base_url": "http://localhost:11434/v1",
      "chat_path": "/chat/completions",
      "api_key_env": "LOCAL_OPENAI_API_KEY",
      "auth_header": "Authorization",
      "auth_token_prefix": "Bearer",
      "default_model": "llama3.1",
      "models": ["llama3.1"],
      "force_https": false
    }
  ]
}
```

Run:

```bash
ti-cli provider list
ti-cli prompt run code-review "review this change" --provider local-openai
```

## Supported fields

- `name`: provider name used by `--provider`.
- `type`: `openai_compatible` or `cookie_http`.
- `base_url`: base API URL, for example `https://api.example.com/v1`.
- `chat_path`: path for chat completions; default `/chat/completions`.
- `api_key` or `api_key_env`: token source.
- `auth_header`: default `Authorization`.
- `auth_token_prefix`: default `Bearer`.
- `custom_headers`: extra static headers.
- `default_model` and `models`: model metadata for routing/status.
- `force_https`: reject non-HTTPS base URLs unless local development.
