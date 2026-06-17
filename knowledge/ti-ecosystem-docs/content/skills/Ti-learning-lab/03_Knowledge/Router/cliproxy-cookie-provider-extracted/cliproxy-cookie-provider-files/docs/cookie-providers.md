---
tags: ["router", "tibrain", "provider", "documentation", "cli"]
scopes: ["cli", "tibrain"]
last_updated: 2026-05-22
---
# Cookie providers

`cookie-providers` adds a generic cookie-authenticated OpenAI-compatible upstream. It forwards OpenAI `/chat/completions` JSON and streaming responses to an upstream that accepts a `Cookie` header.

This feature does not read browser cookie databases, automate login, reset trials, or refresh cookies. Provide only cookies you are allowed to use.

## Example

```yaml
cookie-providers:
  - name: "my-web-upstream"
    prefix: "web"
    base-url: "https://example.com/api/v1"
    cookie-env: "MY_WEB_UPSTREAM_COOKIE"
    # Or use cookie-file: "~/.cli-proxy-api/my-web-upstream.cookie"
    # Avoid literal cookie: in shared config files.
    headers:
      X-Client: "CLIProxyAPI"
    models:
      - name: "upstream-model-id"
        alias: "web-model"
```

Then call the provider through an OpenAI-compatible client:

```bash
curl http://127.0.0.1:8317/v1/chat/completions \
  -H 'Authorization: Bearer your-api-key-1' \
  -H 'Content-Type: application/json' \
  -d '{"model":"web/web-model","messages":[{"role":"user","content":"hello"}]}'
```

## Fields

- `name`: required; used to create an internal provider key like `cookie-my-web-upstream`.
- `prefix`: optional client-facing model prefix.
- `base-url`: required; upstream OpenAI-compatible base URL, for example `https://example.com/api/v1`.
- `path`: optional; defaults to `/chat/completions`.
- `cookie-env`: environment variable containing the cookie string.
- `cookie-file`: file containing the cookie string.
- `cookie`: literal cookie string; supported, but not recommended for shared config files.
- `headers`: additional static headers. `Cookie` and `Authorization` are ignored here.
- `proxy-url`: optional per-provider proxy. Supports `direct` or `none`.
- `models`: upstream model names and optional client aliases.
