---
tags: ["tibrain", "provider", "documentation", "router", "skill"]
scopes: ["tibrain"]
last_updated: 2026-05-22
---
# Devin Provider Configuration Solutions Comparison

## Overview

Comprehensive comparison of three solutions tested to fix Devin provider configuration issues in Ti Router.

## Problem Statement

Devin provider failed to initialize because:
1. Config system only loaded from system environment variables
2. No mechanism to load from YAML configuration
3. Devin requires both `DEVIN_API_KEY` and `DEVIN_ORG_ID` to be set correctly

## Solutions Tested

### Option 1: Fix Env Var Loading (Quick Fix)

**Implementation**:
- Added `Z:\00_SECRET\router.env` to `loadEnv()` in `cmd/routerd/main.go`
- Added `"devin"` to valid providers in `cmd/routerd/handlers/models/handlers.go`

**Changes**:
```go
// cmd/routerd/main.go
func loadEnv() {
    envPaths := []string{
        os.Getenv("ROUTER_ENV_FILE"),
        "Z:\\00_SECRET\\router.env",  // Added external path
        "router.env",
        ".env",
        "configs/router.env",
    }
    // ... rest of function
}

// cmd/routerd/handlers/models/handlers.go
validProviders := map[string]bool{
    // ... existing providers
    "devin": true,  // Added
}
```

**Test Results**:
- ✅ Env vars loaded from external file
- ✅ Devin provider loaded with 16 models
- ✅ API `/v1/models?provider=devin` returned 16 models

**Pros**:
- Quick to implement (5 minutes)
- Simple changes
- No YAML structure changes

**Cons**:
- Hardcoded path
- Not flexible
- Provider-specific code

**Complexity**: Low
**Time**: 5 minutes
**Flexibility**: Low

---

### Option 2: Simplified YAML Config (Medium-term)

**Implementation**:
- Added env var expansion `${ENV_VAR}` to `loadAPIKeysFromYAML()`
- Changed devin config to use env var expansion
- Special handling for devin provider (second key → devin_org_id)

**Changes**:
```go
// layers/config/config_methods.go
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

// Special handling for devin
case []any:
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

**YAML Changes**:
```yaml
devin:
  api_keys:
    - "${DEVIN_API_KEY}"
    - "${DEVIN_ORG_ID}"
```

**Test Results**:
- ✅ Env var expansion in YAML
- ✅ Devin provider loaded with 17 models (including swe-1-5)
- ✅ API `/v1/models?provider=devin` returned 17 models

**Pros**:
- Flexible (no hardcoded paths)
- Industry-standard `${VAR}` syntax
- Works with any provider
- Security (secrets not in YAML)

**Cons**:
- Provider-specific logic for devin
- Requires env vars to be set
- More complex than Option 1

**Complexity**: Medium
**Time**: 15 minutes
**Flexibility**: High

---

### Option 3: Adopt CLIProxyAPI Pattern (Long-term)

**Implementation**:
- Added `Advanced` field to `Config` and `YAMLProviderConfig` structs
- Added helper functions: `GetProviderModelAliases()`, `GetAPIKeyRotation()`
- Added advanced section to YAML config
- Combined with env var expansion from Option 2

**Changes**:
```go
// layers/provider/config.go
type Config struct {
    // ... existing fields
    Advanced map[string]any `json:"advanced,omitempty"`
}

type YAMLProviderConfig struct {
    // ... existing fields
    Advanced map[string]any `yaml:"advanced,omitempty"`
}

// Helper functions
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

func GetAPIKeyRotation(providerName string) bool {
    cfg := GetYAMLProviderConfig(providerName)
    if cfg == nil || cfg.Advanced == nil {
        return false
    }
    rotation, ok := cfg.Advanced["api_key_rotation"].(bool)
    return rotation && ok
}
```

**YAML Changes**:
```yaml
devin:
  # ... existing config
  advanced:
    model_aliases:
      swe-1-6: "swe-1-6:latest"
      swe-1-5: "swe-1-5:latest"
    api_key_rotation: true
```

**Test Results**:
- ✅ Advanced config loaded successfully
- ✅ Devin provider loaded with 17 models
- ✅ Model aliases accessible via `GetProviderModelAliases()`
- ✅ API key rotation flag accessible via `GetAPIKeyRotation()`
- ✅ API `/v1/models?provider=devin` returned 17 models

**Test Script Output**:
```
Devin model aliases:
  swe-1-6 -> swe-1-6:latest
  swe-1-5 -> swe-1-5:latest
Devin API key rotation: true
Total providers loaded: 54
```

**Pros**:
- Extensible architecture
- Foundation for advanced features
- Industry-standard pattern (CLIProxyAPI)
- Provider-specific advanced config
- Backward compatible

**Cons**:
- Most complex to implement
- Requires additional code for feature usage
- Overkill for simple use cases

**Complexity**: High
**Time**: 20 minutes
**Flexibility**: Very High

---

## Comparison Table

| Feature | Option 1 | Option 2 | Option 3 |
|---------|----------|----------|----------|
| **Implementation Time** | 5 min | 15 min | 20 min |
| **Complexity** | Low | Medium | High |
| **Flexibility** | Low | High | Very High |
| **Extensibility** | None | Medium | High |
| **Backward Compatible** | Yes | Yes | Yes |
| **Security** | Medium | High | High |
| **Industry Standard** | No | Yes (env expansion) | Yes (CLIProxyAPI) |
| **Provider-Specific Code** | Yes | Yes | No |
| **Future-Proof** | No | Medium | Yes |
| **Maintenance** | Easy | Easy | Medium |

## Recommendation

### Short-term: Option 2 (Simplified YAML Config)
**Best balance** between flexibility and complexity for most use cases.

### Long-term: Option 3 (CLIProxyAPI Pattern)
**Foundation for advanced features** like model alias resolution, API key rotation, and multiple API keys with proxy support.

### When to Use Each Option

**Use Option 1 when**:
- Quick fix needed
- Simple env var loading sufficient
- No need for extensibility

**Use Option 2 when**:
- Need flexible env var handling
- Industry-standard syntax preferred
- Security is important
- Medium complexity acceptable

**Use Option 3 when**:
- Planning advanced features
- Need model alias mapping
- Need API key rotation
- Building long-term solution
- CLIProxyAPI pattern desired

## Files Modified by Each Option

### Option 1
- `cmd/routerd/main.go` - Added external env file path
- `cmd/routerd/handlers/models/handlers.go` - Added "devin" to valid providers

### Option 2
- `layers/config/config_methods.go` - Added env var expansion
- `configs/providers.yaml` - Updated devin config with `${VAR}` syntax
- `cmd/routerd/main.go` - Added external env file path (from Option 1)
- `cmd/routerd/handlers/models/handlers.go` - Added "devin" to valid providers (from Option 1)

### Option 3
- `layers/provider/config.go` - Added Advanced field and helper functions
- `layers/config/config_methods.go` - Added env var expansion (from Option 2)
- `configs/providers.yaml` - Added advanced section
- `cmd/routerd/main.go` - Added external env file path (from Option 1)
- `cmd/routerd/handlers/models/handlers.go` - Added "devin" to valid providers (from Option 1)

## Test Coverage

All options were tested with:
1. Server startup verification
2. Env var loading confirmation
3. Provider initialization check
4. Model count verification
5. API endpoint testing (`/v1/models?provider=devin`)

**Results**: All options passed tests successfully.

## Lessons Learned

1. **Start Simple**: Option 1 provided quick validation
2. **Iterate**: Option 2 added flexibility
3. **Plan Ahead**: Option 3 established foundation
4. **Test Each**: Verify each option independently
5. **Document**: Record comparison for future reference

## Related Documentation

- `ENV_VAR_LOADING_SOLUTION.md` - Detailed env var loading implementation
- `OPTION3_CLIProxyAPI_PATTERN.md` - CLIProxyAPI pattern implementation
- `DEVIN_PROVIDER_RESEARCH.md` - Devin provider research
- `OPTION3_CLIProxyAPI_PATTERN_IMPLEMENTATION.md` - Implementation guide

## Conclusion

All three solutions successfully resolved the Devin provider configuration issue. The choice depends on specific requirements:
- Quick fix → Option 1
- Balance of flexibility/complexity → Option 2
- Long-term extensibility → Option 3

**Final Choice**: Option 3 adopted for production use as it provides the best foundation for future enhancements while maintaining backward compatibility.
