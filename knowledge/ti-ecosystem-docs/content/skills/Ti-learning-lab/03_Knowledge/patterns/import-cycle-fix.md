# Import Cycle Fix - Khắc Phục Vòng Lặp Import

> **Ngày tạo**: 2026-04-29  
> **Tác giả**: Claude AI Agent  
> **Dự án**: Ti Router  
> **Mục tiêu**: Khắc phục vòng lặp import giữa packages  

---

## Vấn Đề

Trong session trước, có import cycle giữa:
- `cookie` package imports `provider` package
- `provider` package imports `cookie` package

Go compiler không cho phép circular imports.

---

## Giải Pháp

### 1. Xóa Cookie Package

**Tình trạng**: Cookie package đã bị xóa trong session trước

**File đã xóa**:
- `Z:\Ti\router\layers\cookie\generic.go`
- `Z:\Ti\router\layers\cookie\` (toàn bộ package)

**Lý do**: Cookie providers đã được implement inline trong bootstrap.go, không cần package riêng.

---

### 2. Inline Provider Implementations

**Pattern**: Implement providers inline trong bootstrap.go thay vì sub-packages

**Ví dụ**: Local providers (lmstudio, llamacpp, ollama)

**Trước (Sub-package)**:
```
layers/provider/lmstudio/
  ├── provider.go
  ├── request.go
  └── stream.go
```

**Sau (Inline trong bootstrap.go)**:
```go
// buildLMStudioProvider creates LM Studio provider
func buildLMStudioProvider(cfg *config.Config) (Provider, error) {
    // Implementation inline
}
```

---

## Kiến Thúc Học Được

1. **Go import cycle là common issue**
   - Go compiler không cho phép circular imports
   - Cần refactor architecture để tránh vòng lặp

2. **Sub-packages không phải lúc nào cũng cần thiết**
   - Nếu provider logic đơn giản, có thể implement inline
   - Sub-packages chỉ cần cho complex logic cần tái sử dụng

3. **Bootstrap pattern tốt cho provider registration**
   - Centralized provider creation trong bootstrap.go
   - Dễ quản lý và debug
   - Tránh import cycle

4. **Type alias có thể giúp tránh import cycle**
   - Nếu cần type từ package khác, dùng type alias thay vì import
   - Ví dụ: `type ProviderStatus = string`

---

## Best Practices

1. **Tránh circular dependencies**
   - Design architecture với clear dependency direction
   - Core packages không nên import from feature packages
   - Feature packages có thể import from core packages

2. **Sử dụng dependency injection**
   - Pass dependencies qua constructor thay vì import trực tiếp
   - Interface-based design giúp giảm coupling

3. **Centralized registration**
   - Bootstrap pattern centralizes provider registration
   - Dễ quản lý lifecycle của providers
   - Tránh scattered initialization code

---

## Troubleshooting

### Detect Import Cycle

**Symptom**: Build error: `import cycle not allowed`

**Solutions**:
1. Xác định cycle: `go build -gcflags="-all=-l"` để xem import graph
2. Refactor architecture: Tách common logic vào separate package
3. Use interfaces: Define interfaces để reduce coupling
4. Type aliases: Sử dụng type alias thay vì import

---

## References

- **Go Import Cycle**: https://golang.org/ref/mod#build-constraints
- **Bootstrap Pattern**: `Z:\Ti\router\layers\provider\bootstrap.go`
- **Provider Registry**: `Z:\Ti\router\layers\provider\registry.go`

---

*Last Updated: 2026-04-29*
