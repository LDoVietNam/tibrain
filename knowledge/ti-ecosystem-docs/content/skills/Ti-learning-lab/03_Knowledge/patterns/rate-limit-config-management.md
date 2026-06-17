# Quản Lý Cấu Hình Rate Limit trong Go

> **Ngày tạo**: 2026-04-28  
> **Task**: TR-001 - Remove Hardcoded Rate Limit Values  
> **Project**: Ti Router  
> **Trạng thái**: ✅ Hoàn thành

---

## 📋 Tổng Quan

Bài học này mô tả cách quản lý cấu hình rate limit trong Go project, thay thế hardcoded values bằng configuration management pattern.

## 🎯 Vấn Đề Ban Đầu

**Problem**: Hardcoded rate limit values (100 req/min) trong code

**Locations**:
- `Z:\Ti\router\cmd\routerd\router.go` line 61: `r.rateLimiter.AddProvider(name, 100, 100)`
- `Z:\Ti\router\cmd\routerd\main.go` line 608: `rateLimiter.AddProvider(name, 100, 100)`

**Issues**:
- Không thể thay đổi rate limit mà không rebuild
- Không có flexibility cho different environments
- Không thể override với environment variables
- Hardcoded values khó maintain

## ✅ Giải Pháp

### 1. Config Structure

```go
// RateLimitConfig represents rate limit configuration
type RateLimitConfig struct {
    DefaultRPM  int `yaml:"default-rpm"`  // requests per minute
    DefaultBurst int `yaml:"default-burst"` // burst capacity
}
```

### 2. Router Integration

```go
type Router struct {
    // ... existing fields
    rateLimitConfig RateLimitConfig
}

func NewRouter() *Router {
    return &Router{
        // ... existing initialization
        rateLimitConfig: RateLimitConfig{
            DefaultRPM:  100,  // default fallback
            DefaultBurst: 100,
        },
    }
}
```

### 3. Config Loading

```go
func (r *Router) loadRateLimitConfig(configPath string) error {
    // Read YAML file
    data, err := os.ReadFile(configPath)
    if err != nil {
        return fmt.Errorf("read config file: %w", err)
    }

    // Parse YAML
    var config struct {
        Server struct {
            RateLimit RateLimitConfig `yaml:"rate-limit"`
        } `yaml:"server"`
    }

    if err := yaml.Unmarshal(data, &config); err != nil {
        return fmt.Errorf("parse config: %w", err)
    }

    // Apply config with environment variable override
    r.rateLimitConfig = config.Server.RateLimit

    // Override with environment variables if set
    if rpm := os.Getenv("ROUTER_RATE_LIMIT_RPM"); rpm != "" {
        if val, err := strconv.Atoi(rpm); err == nil && val > 0 {
            r.rateLimitConfig.DefaultRPM = val
        }
    }

    if burst := os.Getenv("ROUTER_RATE_LIMIT_BURST"); burst != "" {
        if val, err := strconv.Atoi(burst); err == nil && val > 0 {
            r.rateLimitConfig.DefaultBurst = val
        }
    }

    // Ensure minimum values
    if r.rateLimitConfig.DefaultRPM <= 0 {
        r.rateLimitConfig.DefaultRPM = 100
    }
    if r.rateLimitConfig.DefaultBurst <= 0 {
        r.rateLimitConfig.DefaultBurst = 100
    }

    return nil
}
```

### 4. YAML Config

```yaml
server:
  rate-limit:
    default-rpm: 100
    default-burst: 100
```

### 5. Environment Variable Override

```bash
# Override rate limit with environment variables
export ROUTER_RATE_LIMIT_RPM=200
export ROUTER_RATE_LIMIT_BURST=200
```

### 6. Usage in Code

```go
// Initialize rate limiter with config values
for name := range providerConfigs {
    r.rateLimiter.AddProvider(name, r.rateLimitConfig.DefaultRPM, r.rateLimitConfig.DefaultBurst)
}
```

## 🔑 Best Practices

### 1. **Config Hierarchy**
1. Default values (code)
2. YAML config file
3. Environment variables (highest priority)

### 2. **Validation**
- Validate config values before using
- Ensure minimum/maximum values
- Handle parse errors gracefully

### 3. **Fallback**
- Always have default values
- Use fallback if config loading fails
- Never crash on missing config

### 4. **Documentation**
- Document config structure in YAML
- Document environment variables
- Provide examples in README

## 📁 Files Modified

1. **Z:\Ti\router\cmd\routerd\router.go**
   - Added `RateLimitConfig` struct
   - Added `rateLimitConfig` field to Router
   - Added `loadRateLimitConfig()` method
   - Updated `Init()` to use config values
   - Added imports: `fmt`, `strconv`, `yaml.v3`

2. **Z:\Ti\router\cmd\routerd\main.go**
   - Updated config reload to use `router.rateLimitConfig`

3. **Z:\Ti\router\configs\Tiserverrouter.yaml**
   - Added `server.rate-limit` section

## 🧪 Testing

### Test Scenarios

1. **Default Config**
   - Expected: Use 100, 100 from config file
   - Command: `./routerd.exe -config configs/Tiserverrouter.yaml`

2. **Custom Config**
   - Modify YAML: `default-rpm: 200`
   - Expected: Use 200, 100 from config

3. **Environment Override**
   - Set env: `ROUTER_RATE_LIMIT_RPM=300`
   - Expected: Use 300, 100 (env overrides YAML)

4. **Missing Config**
   - Remove `server.rate-limit` from YAML
   - Expected: Use defaults 100, 100

5. **Invalid Config**
   - Set invalid value: `default-rpm: -1`
   - Expected: Use fallback 100, 100

## 🎓 Lessons Learned

### Config Management Patterns

1. **YAML Structure**
   - Use nested structure for organization
   - Use kebab-case for YAML keys
   - Group related configs together

2. **Environment Variables**
   - Use uppercase with underscores
   - Prefix with app name (ROUTER_)
   - Document in README

3. **Validation**
   - Validate type (strconv.Atoi)
   - Validate range (> 0)
   - Provide fallback values

4. **Error Handling**
   - Log warnings, don't crash
   - Use default values on error
   - Continue operation if possible

### Go-Specific Patterns

1. **Struct Tags**
   - Use `yaml:"key"` for mapping
   - Use same naming convention
   - Document in comments

2. **Error Wrapping**
   - Use `fmt.Errorf("context: %w", err)`
   - Preserve error context
   - Handle at appropriate level

3. **Default Values**
   - Set in struct literal
   - Override in Init()
   - Validate before use

## 🔮 Future Improvements

1. **Per-Provider Rate Limits**
   ```yaml
   server:
     rate-limit:
       providers:
         openai:
           rpm: 200
           burst: 200
         claude:
           rpm: 100
           burst: 100
   ```

2. **Dynamic Rate Limiting**
   - Adjust based on load
   - Time-based limits (peak/off-peak)
   - User-based limits

3. **Config Validation Library**
   - Use go-playground/validator
   - Define validation rules
   - Validate on startup

4. **Hot Reload**
   - Watch config file changes
   - Reload without restart
   - Validate before applying

## 📚 References

- **Viper**: https://github.com/spf13/viper (popular Go config library)
- **YAML in Go**: gopkg.in/yaml.v3
- **Environment Variables**: os.Getenv()
- **Config Patterns**: 12-Factor App (https://12factor.net/config)

## ✅ Verification

- [x] Build passes without errors
- [x] Config loads from YAML
- [x] Environment variables override config
- [x] Default values used when config missing
- [x] Invalid values handled gracefully
- [x] Config reload uses new values

---

**Kết luận**: Config management pattern giúp loại bỏ hardcoded values, tăng flexibility, và improve maintainability. Environment variable override cho phép different configurations per environment without code changes.
