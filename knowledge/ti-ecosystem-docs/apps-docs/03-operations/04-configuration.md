# Configuration Guide

## Configuration File

The router configuration is defined in `Tiserverrouter.yaml`.

## Key Settings

### Server Settings
```yaml
host: ''              # Bind to all interfaces (use "127.0.0.1" for localhost only)
port: 1807           # Server port
```

### Authentication
```yaml
api-keys:
  - 'sk-jarvis-dev'  # API keys for client authentication
```

### Management API
```yaml
remote-management:
  allow-remote: true              # Allow non-localhost access
  secret-key: '$2a$10$...'       # Management panel authentication
  disable-control-panel: false    # Enable/disable management UI
```

### Debug & Logging
```yaml
debug: true                      # Enable debug logging
pprof:
  enable: true                   # Enable pprof profiling
  addr: "127.0.0.1:8316"        # Pprof address
usage-statistics-enabled: true   # Enable usage statistics
logging-to-file: true            # Log to file instead of stdout
logs-max-total-size-mb: 100      # Max log size before rotation
```

### OAuth Token Directory
```yaml
auth-dir: 'auth'                 # Directory for OAuth tokens
oauth-token-dir: 'Z:\06_AUTH'   # Source directory for OAuth tokens
```

## Provider Configuration

### Gemini API Keys
```yaml
gemini-api-key:
  - api-key: 'AIza...'
    models:
      - name: gemini-2.5-flash
        alias: Gemini 2.5 Flash
```

### OpenAI-Compatible Providers
```yaml
openai-compatibility:
  - name: 'groq'
    base-url: 'https://api.groq.com/openai/v1'
    api-key-entries:
      - api-key: 'gsk_...'
    models:
      - name: 'llama-3.3-70b-versatile'
        alias: 'llama-3-3-70b'
```

## Model Aliases

Model aliases provide friendly names for model IDs:
```yaml
models:
  - name: gemini-2.5-flash
    alias: Gemini 2.5 Flash  # User-friendly name
```

Use aliases in API requests:
```json
{
  "model": "Gemini 2.5 Flash",
  "messages": [...]
}
```

## OAuth Model Aliases

Configure OAuth-specific model aliases:
```yaml
oauth-model-alias:
  github-copilot:
    - name: claude-sonnet-4-6
      alias: claude-sonnet-4-6
      fork: true
```

## Excluded Models

Exclude models from specific OAuth providers:
```yaml
oauth-excluded-models:
  antigravity:
    - '*'  # Exclude all antigravity models
```

## Reloading Configuration

The router automatically reloads configuration when:
- `Tiserverrouter.yaml` is modified
- Files in `auth/` directory change
- OAuth tokens are updated

No manual restart required.

## Best Practices

1. **API Keys**: Use strong, random API keys
2. **Secret Key**: Generate a strong secret key for management panel
3. **Logging**: Enable file logging in production
4. **Debug**: Disable debug logging in production
5. **OAuth**: Keep OAuth tokens in secure directory (Z:\06_AUTH)

## Troubleshooting

### Router won't start
- Check port 1807 is not in use: `netstat -ano | findstr :1807`
- Verify YAML syntax is correct
- Check file permissions

### API returns 401
- Verify API key is correct
- Check `api-keys` in config
- Ensure Authorization header format: `Bearer sk-jarvis-dev`

### Management panel not accessible
- Check `secret-key` is set in config
- Verify `disable-control-panel` is false
- Use correct header: `X-Management-Key`

### Models not available
- Verify provider API keys are valid
- Check model names match provider specifications
- Review provider-specific configuration
