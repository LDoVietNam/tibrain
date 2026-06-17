# Option 3: CLIProxyAPI Pattern Implementation for Ti Router

## Overview

Option 3 adopts the CLIProxyAPI pattern for advanced provider configuration in Ti Router, enabling:
- Model alias mapping
- API key rotation
- Environment variable expansion
- Extensible advanced configuration

## Implementation Date

2026-05-16

## Problem Solved

The Devin provider required both `DEVIN_API_KEY` and `DEVIN_ORG_ID` environment variables to be loaded correctly. The existing config system only loaded from environment variables, not from YAML configuration files.

## Solution Architecture

### 1. YAML Configuration Enhancement

Added `advanced` section to provider configuration:

```yaml
devin:
  name: devin
  base_url: https://api.devin.ai/v3
  api_key_env: DEVIN_API_KEY
  api_keys:
    - "${DEVIN_API_KEY}"
    - "${DEVIN_ORG_ID}"
  models:
    - devin-auto
    - swe-1-5
    - swe-1-6
    - swe-1-6-fast
    - swe
    - swe-fast
    - opus
    - sonnet
    - gpt
    - gpt-4
    - gpt-4-turbo
    - codex
    - gemini
    - gemini-pro
    - kimi
    - glm
    - glm-4
  format: devin
  timeout_sec: 120
  priority: 5
  weight: 2
  cost_per_1k: 0.0
  notes: "Devin AI (Cognition Labs) - Session-based autonomous agent"
  # CLIProxyAPI-style advanced config
  advanced:
    model_aliases:
      swe-1-6: "swe-1-6:latest"
      swe-1-5: "swe-1-5:latest"
    api_key_rotation: true
```

### 2. Code Changes

#### 2.1 Config Struct Updates

**File**: `layers/provider/config.go`

Added `Advanced` field to both `Config` and `YAMLProviderConfig` structs:

```go
type Config struct {
    // ... existing fields ...
    Advanced map[string]any `json:"advanced,omitempty"` // CLIProxyAPI-style advanced config
}

type YAMLProviderConfig struct {
    // ... existing fields ...
    Advanced map[string]any `yaml:"advanced,omitempty"` // CLIProxyAPI-style advanced config
}
```

#### 2.2 Helper Functions

**File**: `layers/provider/config.go`

Added helper functions to access advanced config:

```go
// GetProviderModelAliases returns model aliases from advanced config for a provider.
func GetProviderModelAliases(providerName string) map[string]string {
    cfg := GetYAMLProviderConfig(providerName)
    if cfg == nil || cfg.Advanced == nil {
        return nil
    }

    aliases, ok := cfg.Advanced["model_aliases"].(map[string]any)
    if !ok {
        return nil
    }

    result := make(map[string]string)
    for k, v := range aliases {
        if alias, ok := v.(string); ok {
            result[k] = alias
        }
    }

    return result
}

// GetAPIKeyRotation returns whether API key rotation is enabled for a provider.
func GetAPIKeyRotation(providerName string) bool {
    cfg := GetYAMLProviderConfig(providerName)
    if cfg == nil || cfg.Advanced == nil {
        return false
    }

    rotation, ok := cfg.Advanced["api_key_rotation"].(bool)
    return rotation && ok
}
```

#### 2.3 Environment Variable Expansion

**File**: `layers/config/config_methods.go`

Added env var expansion support to `loadAPIKeysFromYAML`:

```go
// expandEnvVars expands environment variables in a string.
// Supports ${ENV_VAR} pattern.
func expandEnvVars(s string) string {
    result := s
    re := regexp.MustCompile(`\$\{([^}]+)\}`)
    matches := re.FindAllStringSubmatch(s, -1)
    for _, match := range matches {
        if len(match) > 1 {
            envVar := match[1]
            envValue := os.Getenv(envVar)
            if envValue != "" {
                result = strings.ReplaceAll(result, match[0], envValue)
            }
        }
    }
    return result
}
```

Special handling for devin provider (second API key → devin_org_id):

```go
case []any:
    // Array of API keys - use the first non-empty one
    // For devin provider, also set devin_org_id from second key
    keyIndex := 0
    for _, key := range v {
        if keyStr, ok := key.(string); ok && keyStr != "" {
            expandedKey := expandEnvVars(keyStr)
            if expandedKey != "" {
                if keyIndex == 0 {
                    result[providerName] = expandedKey
                } else if providerName == "devin" && keyIndex == 1 {
                    result[providerName+"_org_id"] = expandedKey
                }
                keyIndex++
            }
        }
    }
```

### 3. Test Results

**Advanced Config Loading**:
```
Devin model aliases:
  swe-1-6 -> swe-1-6:latest
  swe-1-5 -> swe-1-5:latest
Devin API key rotation: true
Total providers loaded: 54
```

**API Endpoint Verification**:
- `/v1/models?provider=devin` → **17 models** (bao gồm swe-1-5)

## Comparison with CLIProxyAPI

### CLIProxyAPI Pattern

```yaml
openai-compatibility:
  - name: "openrouter"
    prefix: "test"
    base-url: "https://openrouter.ai/api/v1"
    api-key-entries:
      - api-key: "sk-or-v1-...b780"
        proxy-url: "socks5://proxy.example.com:1080"
    models:
      - name: "moonshotai/kimi-k2:free"
        alias: "kimi-k2"
```

### Ti Router Implementation

```yaml
devin:
  name: devin
  base_url: https://api.devin.ai/v3
  api_keys:
    - "${DEVIN_API_KEY}"
    - "${DEVIN_ORG_ID}"
  models:
    - swe-1-5
    - swe-1-6
  advanced:
    model_aliases:
      swe-1-6: "swe-1-6:latest"
    api_key_rotation: true
```

**Key Differences**:
- CLIProxyAPI uses separate `openai-compatibility` section
- Ti Router integrates advanced config directly into provider config
- CLIProxyAPI supports proxy URLs per API key (future enhancement)
- Ti Router uses env var expansion for security

## Benefits

1. **Extensibility**: Easy to add new advanced features
2. **Backward Compatibility**: Existing configs continue to work
3. **Security**: Env var expansion keeps secrets out of YAML
4. **Flexibility**: Provider-specific advanced config
5. **Foundation**: Ready for future CLIProxyAPI-style features

## Future Enhancements

### 1. Model Alias Resolution in Routing

Currently, model aliases are loaded but not yet used in routing. Implementation plan:

```go
// In routing logic
func resolveModelWithAliases(providerName, modelName string) string {
    aliases := GetProviderModelAliases(providerName)
    if aliases != nil {
        if resolved, ok := aliases[modelName]; ok {
            return resolved
        }
    }
    return modelName
}
```

### 2. API Key Rotation

Currently, `GetAPIKeyRotation` returns the flag but rotation logic is not implemented. Implementation plan:

```go
func getRotatedAPIKey(providerName string) string {
    if !GetAPIKeyRotation(providerName) {
        return getDefaultAPIKey(providerName)
    }

    // Implement rotation logic (round-robin, weighted, etc.)
    keys := getAPIKeysForProvider(providerName)
    return rotateKeys(keys)
}
```

### 3. Advanced Config Schema

Define a structured schema for advanced config:

```go
type AdvancedConfig struct {
    ModelAliases      map[string]string `yaml:"model_aliases"`
    APIKeyRotation    bool              `yaml:"api_key_rotation"`
    MultipleAPIKeys   []APIKeyEntry     `yaml:"api_key_entries"`
    Prefix            string            `yaml:"prefix"`
    // ... more fields
}

type APIKeyEntry struct {
    APIKey   string `yaml:"api_key"`
    ProxyURL string `yaml:"proxy_url,omitempty"`
    Weight   int    `yaml:"weight,omitempty"`
}
```

## Files Modified

1. `layers/provider/config.go` - Added Advanced field and helper functions
2. `layers/config/config_methods.go` - Added env var expansion support
3. `configs/providers.yaml` - Added advanced section to devin provider
4. `cmd/routerd/handlers/models/handlers.go` - Added "devin" to valid providers
5. `cmd/routerd/main.go` - Added router.env path to env loading

## Related Documentation

- `OPTION3_CLIProxyAPI_PATTERN_IMPLEMENTATION.md` - Detailed implementation guide
- `DEVIN_PROVIDER_RESEARCH.md` - Devin provider research
- `configs/providers.yaml` - Provider configuration file

## Lessons Learned

1. **Env Var Expansion**: Using `${ENV_VAR}` pattern in YAML provides security and flexibility
2. **Provider-Specific Logic**: Some providers (like devin) need special handling for multiple keys
3. **Backward Compatibility**: Advanced config should be optional to maintain compatibility
4. **Foundation First**: Implement basic structure before adding complex features
5. **Testing**: Create test scripts to verify advanced config loading

## Conclusion

Option 3 provides a robust foundation for advanced provider configuration following CLIProxyAPI patterns while maintaining backward compatibility and security best practices. The implementation successfully loads Devin provider with 17 models including swe-1-5, and establishes patterns for future enhancements.
