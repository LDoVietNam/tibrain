---
tags: ["authentication", "tibrain", "security", "documentation", "provider-windsurf"]
scopes: ["auth", "tibrain"]
last_updated: 2026-05-22
---
# Windsurf Provider Authentication Pattern

## Context

Windsurf provider was showing as "unconfigured" and "healthy: false" even though the API key was correctly set in the environment variable file. The root cause was multiple initialization issues preventing the provider from loading the API key.

## Problem

Windsurf provider failed to initialize because:
1. Bootstrap check prevented Windsurf from being initialized when `store != nil`
2. Environment variable fallback was missing in the provider initialization
3. Plugin registry factory didn't check environment variables

## Solution Pattern

### 1. Remove Conditional Store Check in Bootstrap

**File**: `layers/provider/bootstrap.go`

```go
// Before:
if store != nil {
    if p, err := buildWindsurfProvider(cfg, store); err == nil && p != nil {
        reg.Register(p)
    }
}

// After:
if p, err := buildWindsurfProvider(cfg, store); err == nil && p != nil {
    reg.Register(p)
}
```

**Why**: The router was calling `BuildRegistryFromConfig(cfg, nil)` with `store = nil`, which prevented Windsurf from being initialized.

### 2. Add Environment Variable Fallback in Provider

**File**: `layers/provider/windsurf.go`

```go
import (
    // ... other imports
    "os"
    // ... other imports
)

func (p *WindsurfProvider) Init(cfg *config.Config) error {
    key := ""
    if cfg != nil && cfg.APIKeys != nil {
        key = cfg.APIKeys["windsurf"]
    }
    // Also try DB first if available
    if key == "" && db.Get() != nil {
        // This will be handled by bootstrap calling InitWithKey
        // Init is kept for backwards compatibility
    }
    // If still no key, try environment variable directly as fallback
    if key == "" {
        if envKey := os.Getenv("WINDSURF_API_KEY"); envKey != "" {
            key = envKey
        }
    }
    return p.InitWithKey(key)
}
```

**Why**: Adds a fallback to directly check the environment variable if the config doesn't have the API key.

### 3. Add Environment Variable Check in Plugin Registry

**File**: `layers/provider/plugin_registry.go`

```go
import (
    // ... other imports
    "os"
    // ... other imports
)

RegisterFactory("windsurf", func(ctx context.Context, cfg map[string]any) (Provider, error) {
    p := &WindsurfProvider{}
    key, _ := cfg["api_key"].(string)
    // If no key from config, try environment variable
    if key == "" {
        key = os.Getenv("WINDSURF_API_KEY")
    }
    if err := p.InitWithKey(key); err != nil {
        return nil, err
    }
    return p, nil
})
```

**Why**: The plugin registry factory directly calls `InitWithKey` bypassing the `Init` method, so it needs its own environment variable check.

### 4. Ensure Environment File Loading

**File**: `cmd/routerd/appinit/init.go`

Verify that `LoadSecretFile()` loads from the correct environment file:
```go
secretFile := `Z:\00_SECRET\router.env`
```

**File**: `cmd/routerd/main.go`

Ensure the environment file is loaded:
```go
godotenv.Load("Z:\\00_SECRET\\router.env")
```

## Configuration

### Environment Variable

Set the Windsurf API key in `Z:\00_SECRET\router.env`:

```bash
WINDSURF_API_KEY=ott$<your-token>
```

### Provider Config

The `providers.yaml` configuration references the environment variable:

```yaml
windsurf:
  name: windsurf
  base_url: https://server.windsurf.com/v1
  api_key_env: WINDSURF_API_KEY
  api_keys: []
  models:
    - claude-4.5-sonnet-thinking
    - windsurf-sonnet-thinking
  format: openai
  timeout_sec: 60
  priority: 15
  weight: 4
  cost_per_1k: 0.003
```

**Important**: Do not hardcode the API key in `api_keys` array to avoid exposing it in commits.

## Testing

### Verify Provider Health

```bash
curl http://localhost:1807/api/providers/health
```

Expected response should include Windsurf with `healthy: true`:

```json
{
  "name": "windsurf",
  "base_url": "https://server.windsurf.com/v1",
  "healthy": true,
  "models": [
    "claude-4.5-opus-thinking",
    "claude-4.5-sonnet-thinking",
    // ... other models
  ]
}
```

## Key Learnings

1. **Multiple Initialization Paths**: Providers can be initialized through multiple paths (bootstrap, plugin registry, direct Init), and each path needs to handle environment variable loading.

2. **Bootstrap Store Dependency**: Conditional checks based on store availability can prevent providers from being initialized if the router is called with `store = nil`.

3. **Environment Variable Fallback**: Always add environment variable fallbacks in all initialization paths to ensure providers can load credentials from environment variables.

4. **File Path Consistency**: Ensure environment file paths are consistent across all loading mechanisms (LoadSecretFile, godotenv.Load, etc.).

## Files Modified

1. `layers/provider/bootstrap.go` - Removed store check for Windsurf
2. `layers/provider/windsurf.go` - Added os import and env var fallback
3. `layers/provider/plugin_registry.go` - Added os import and env var check
4. `cmd/routerd/appinit/init.go` - Verified secret file path
5. `cmd/routerd/main.go` - Verified environment file loading

## References

- Windsurf OTT Token: Obtain from https://windsurf.com/show-auth-token
- Windsurf API Base URL: https://server.windsurf.com/v1
- Similar implementation: [pi-windsurf-provider](https://github.com/wowyuarm/pi-windsurf-provider)
