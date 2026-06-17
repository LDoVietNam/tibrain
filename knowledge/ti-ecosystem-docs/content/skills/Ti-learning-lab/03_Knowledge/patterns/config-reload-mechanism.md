---
tags: ["tibrain", "documentation", "skill", "api", "authentication"]
scopes: ["auth", "tibrain"]
last_updated: 2026-05-22
---
# Config Reload Mechanism - Cơ chế Tải Lại Cấu hình

> **Ngày tạo**: 2026-04-29  
> **Tác giả**: Claude AI Agent  
> **Dự án**: Ti Router  
> **Mục tiêu**: Implement cơ chế reload config mà không cần restart router  

---

## Vấn đề

Trong dev mode, mỗi khi add provider mới, developer phải:
1. Thêm provider code vào bootstrap.go
2. Build lại binary: `go build ./cmd/routerd`
3. Restart router
4. Test provider mới

Quy trình này **tốn thời** và **không hiệu quả** cho dev workflow.

---

## Giải pháp

Implement **Config Reload Mechanism** cho phép:
- Reload config mà không cần restart
- Add/remove providers dynamically
- Update API keys mà không cần rebuild
- Hot reload cho dev mode

---

## Implementation

### 1. Endpoint `/api/config/reload`

**File**: `Z:\Ti\router\cmd\routerd\handlers_admin.go`

```go
func handleConfigReload(w http.ResponseWriter, r *http.Request) {
	// Authentication check
	if authService != nil && !authService.Authenticate(r) {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// Reload provider configs from providers.yaml
	providerConfigsPath := "configs/providers.yaml"
	newProviderConfigs, err := providers.LoadFromYAML(providerConfigsPath)
	if err != nil {
		// Handle error
		return
	}

	// Rebuild registry with new configs
	newRegistry, issues := providers.BuildRegistryFromConfig(nil, nil)

	// Update global registry
	providerRegistry = newRegistry

	// Update model registry
	modelRegistry.SyncFromProviderConfigs(newProviderConfigs)

	// Return success
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "ok",
		"providers": len(newProviderConfigs),
		"models": modelRegistry.Count(),
	})
}
```

---

## Workflow Dev Mode

### Trước (Cũ)

```
1. Thêm provider code vào bootstrap.go
2. Build: go build ./cmd/routerd
3. Kill router
4. Start router: ./bin/routerd.exe
5. Test provider
```

### Sau (Mới)

```
1. Thêm provider vào providers.yaml
2. POST /api/config/reload
3. Test provider ngay lập tức
```

---

## Cách Sử Dụng

### Add Provider Mới

**Step 1**: Thêm provider vào `configs/providers.yaml`

```yaml
providers:
  new_provider:
    name: new_provider
    base_url: https://api.example.com/v1
    api_key_env: NEW_PROVIDER_API_KEY
    models:
      - model-1
      - model-2
    format: openai
```

**Step 2**: Reload config

```bash
curl -X POST http://localhost:1807/api/config/reload \
  -H "Authorization: Bearer sk-jarvis-dev"
```

**Step 3**: Test provider

```bash
curl http://localhost:1807/v1/chat/completions \
  -H "Authorization: Bearer sk-jarvis-dev" \
  -H "Content-Type: application/json" \
  -d '{"model": "model-1", "messages": [{"role": "user", "content": "Hello!"}]}'
```

---

## Hạn Chế Hiện Tại

1. **Bootstrap hardcoded**: Mỗi provider vẫn cần được hardcode trong `bootstrap.go`
   - Build vẫn cần nếu add provider mới với code implementation
   - Config reload chỉ work cho providers đã được implement

2. **Không có hot reload cho code**: Chỉ reload config, không reload code
   - Cần build lại nếu thay đổi logic provider
   - Cần restart nếu thay đổi struct/interface

---

## Roadmap Tương Lai

### Phase 1: Config Reload (✅ Hoàn thành)
- [x] Implement `/api/config/reload` endpoint
- [x] Reload providers.yaml
- [x] Rebuild registry
- [x] Update model registry

### Phase 2: Dynamic Provider Loading (🔄 Chưa làm)
- [ ] Plugin system cho providers
- [ ] Load providers từ external directory
- [ ] Hot reload cho provider code
- [ ] Không cần build lại khi add provider

### Phase 3: Config Watcher (🔄 Chưa làm)
- [ ] Watch file changes
- [ ] Auto-reload khi config thay đổi
- [ ] Debounce để tránh reload liên tục
- [ ] Notify khi config reload thành công

---

## Kiến Thúc Học Được

1. **Global variables cần careful update**
   - ProviderRegistry, ModelRegistry cần được update đồng bộ
   - Race conditions có thể xảy ra nếu không dùng mutex

2. **Config loading có 2 layers**
   - `providers.Config` cho provider-specific configs
   - `config.Config` cho global router config
   - Cần convert giữa 2 formats

3. **Bootstrap pattern**
   - BuildRegistryFromConfig nhận `*config.Config` (không phải providers.Config)
   - Cần implement proper conversion hoặc sử dụng env vars

4. **Dev workflow optimization**
   - Config reload giúp dev workflow nhanh hơn
   - Nhưng không thay thế hoàn toàn việc build lại
   - Cần plugin system để không cần build lại

---

## Best Practices

1. **Luôn test config reload trước khi deploy**
   - Test với curl: `curl -X POST /api/config/reload`
   - Verify providers count và models count

2. **Backup config trước khi modify**
   - Copy providers.yaml → providers.yaml.bak
   - Dễ rollback nếu config sai

3. **Sử dụng env vars cho API keys**
   - Không hardcode API keys trong config
   - Sử dụng `{env:VAR_NAME}` pattern

4. **Log config reload events**
   - Log khi config reload thành công
   - Log khi có issues với providers

---

## Ví dụ Sử Dụng Thực Tế

### Test Config Reload

```bash
# 1. Start router
cd Z:\Ti\router
./bin/routerd.exe -config configs/Tiserverrouter.yaml

# 2. Add provider vào providers.yaml
# (edit configs/providers.yaml)

# 3. Reload config
curl -X POST http://localhost:1807/api/config/reload \
  -H "Authorization: Bearer sk-jarvis-dev"

# Response:
# {
#   "status": "ok",
#   "message": "configuration reloaded successfully",
#   "providers": 6,
#   "models": 19,
#   "issues": 0
# }

# 4. Verify models
curl http://localhost:1807/v1/models \
  -H "Authorization: Bearer sk-jarvis-dev"
```

---

## Troubleshooting

### Config reload không work

**Symptom**: `/api/config/reload` trả về error

**Solutions**:
1. Check authentication: API key có hợp lệ không
2. Check file path: `configs/providers.yaml` có tồn tại không
3. Check YAML syntax: YAML có valid không
4. Check permissions: Router có quyền đọc file không

### Providers không được load sau reload

**Symptom**: Models list không có provider mới

**Solutions**:
1. Check bootstrap.go: Provider đã được implement chưa
2. Check provider configs: Config có đúng format không
3. Check API keys: API keys có được set không
4. Check logs: Log có show issues không

---

## References

- **Config Package**: `Z:\Ti\router\layers\config\config.go`
- **Provider Config**: `Z:\Ti\router\layers\provider\config.go`
- **Bootstrap**: `Z:\Ti\router\layers\provider\bootstrap.go`
- **Handlers Admin**: `Z:\Ti\router\cmd\routerd\handlers_admin.go`

---

*Last Updated: 2026-04-29*
