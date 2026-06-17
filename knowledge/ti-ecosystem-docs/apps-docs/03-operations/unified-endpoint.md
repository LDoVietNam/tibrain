# Unified Endpoint Pattern

Ti Router provides a unified OpenAI-compatible endpoint for all tools and clients.

## Base URL

```
http://localhost:1807/v1
```

## API Key

Default API key: `sk-jarvis-dev`

Or set via environment variable: `TI_API_KEY`

## OpenAI-Compatible Endpoints

Ti Router implements the OpenAI API specification:

### List Models
```bash
curl http://localhost:1807/v1/models \
  -H "Authorization: Bearer sk-jarvis-dev"
```

### Chat Completions
```bash
curl http://localhost:1807/v1/chat/completions \
  -H "Authorization: Bearer sk-jarvis-dev" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "llama-3-3-70b",
    "messages": [{"role": "user", "content": "Hello"}]
  }'
```

### Completions
```bash
curl http://localhost:1807/v1/completions \
  -H "Authorization: Bearer sk-jarvis-dev" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "llama-3-3-70b",
    "prompt": "Hello"
  }'
```

## Supported Tools

All tools that support OpenAI-compatible APIs can use the unified endpoint:

| Tool | Config Location | Status |
|------|-----------------|--------|
| Claude Code | `~/.claude/settings.json` | ✅ Supported |
| OpenAI Codex CLI | `~/.codex/config.toml` | ✅ Supported |
| OpenCode | `~/.config/opencode/opencode.json` | ✅ Supported |
| Open Claw | `~/.openclaw/openclaw.json` | ✅ Supported |
| Cursor | Settings > Models | ✅ Supported |
| Cline | Settings panel | ✅ Supported |
| Continue | `~/.continue/config.json` | ✅ Supported |

See [Tool Guides](../05-integration/tool-guides/) for detailed setup instructions.

## Model Naming Convention

Ti Router uses a simple model naming convention:

```
<provider>/<model-name>
```

Examples:
- `groq/llama-3-3-70b` - Groq provider, Llama 3.3 70B model
- `openrouter/google/gemini-2.0-flash-exp:free` - OpenRouter provider, Gemini 2.0 Flash
- `gemini/gemini-2.5-flash` - Gemini provider, Gemini 2.5 Flash

For tools that don't support provider prefix, you can use the model name directly:
- `llama-3-3-70b` - Auto-routed to best provider
- `gemini-2.5-flash` - Auto-routed to Gemini provider

## Auto-Routing

Ti Router supports intelligent model routing:

1. **Provider Selection**: Auto-selects the best provider based on:
   - Health status
   - Latency
   - Cost
   - Rate limits

2. **Fallback**: Automatically switches to backup providers if primary fails

3. **Load Balancing**: Distributes requests across multiple providers

## Usage Dashboard

Monitor API usage via the admin endpoint:

```bash
curl http://localhost:1807/admin/usage \
  -H "Authorization: Bearer sk-jarvis-dev"
```

Returns:
- Total requests
- Success/failure rates
- Token usage (in/out)
- Cost tracking
- Per-provider statistics

## Configuration

### Port Configuration

Default port: `1807`

To change port, edit `configs/Tiserverrouter.yaml`:

```yaml
host: ''
port: 1807
```

### API Key Configuration

Default API keys in `configs/Tiserverrouter.yaml`:

```yaml
api-keys:
  - 'sk-jarvis-dev'
  - '$TI_API_KEY'
```

### Provider Configuration

Providers configured in `configs/Tiserverrouter.yaml`:

```yaml
openai-compatibility:
  - name: 'groq'
    base-url: 'https://api.groq.com/openai/v1'
    api-key-entries:
      - api-key: 'your-groq-key'
    models:
      - name: 'llama-3.3-70b-versatile'
        alias: 'llama-3-3-70b'
```

## Troubleshooting

### Connection Refused

**Problem**: `curl: (7) Failed to connect to localhost port 1807`

**Solution**: Start Ti Router:
```powershell
cd Z:\Ti\router
.\routerd.exe -config configs\Tiserverrouter.yaml
```

### Unauthorized

**Problem**: `401 Unauthorized`

**Solution**: Check API key is correct:
```bash
curl http://localhost:1807/v1/models \
  -H "Authorization: Bearer sk-jarvis-dev"
```

### Model Not Found

**Problem**: `404 Model not found`

**Solution**: Check model name in `configs/Tiserverrouter.yaml`:
```bash
curl http://localhost:1807/v1/models \
  -H "Authorization: Bearer sk-jarvis-dev"
```

## Additional Resources

- [Tool Guides](../05-integration/tool-guides/)
- [API Reference](../02-api/02-v1-reference.md)
- [Configuration Guide](./04-configuration.md)
- [Troubleshooting](./03-troubleshooting.md)
