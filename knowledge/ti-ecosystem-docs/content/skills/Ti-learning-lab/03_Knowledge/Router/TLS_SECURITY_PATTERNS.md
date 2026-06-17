# TLS Security Patterns trong Ti Router

## Tổng Quan

Document này ghi nhận các patterns xử lý TLS certificate verification trong Ti Router, bao gồm best practices và security considerations.

## Patterns Đã Tìm Thấy

### 1. Secure Pattern (Khuyên Dùng)
**File:** `layers/security/tls.go`
```go
config.InsecureSkipVerify = false
config.PreferServerCipherSuites = true
```
- InsecureSkipVerify: false (mặc định)
- Verify server certificate
- Sử dụng TLS 1.2+ minimum
- Curve preferences: P256, P384, P521

### 2. Cookie Client Pattern
**File:** `layers/provider/cookie/client.go`
```go
TLSClientConfig: &tls.Config{
    MinVersion: tls.VersionTLS12,
    // InsecureSkipVerify: false // Verify server cert (default)
}
```
- Mặc định InsecureSkipVerify: false
- Hỗ trợ custom CA certificate
- Hỗ trợ client certificate (mTLS)

### 3. Provider-Specific TLS Handling
**Files:** 
- `layers/provider/windsurf.go` - Windsurf provider
- `layers/authentication/windsurfoauth.go` - Windsurf OAuth
- `layers/provider/executor.go` - BaseExecutor

## Security Issues Đã Phát Hiện

### Vấn Đề 1: BaseExecutor Global InsecureSkipVerify
**File:** `layers/provider/executor.go` (đã fix)
```go
// ❌ KHÔNG KHUYẾN - affects ALL providers
TLSClientConfig: &tls.Config{
    InsecureSkipVerify: true,
}
```

**Vấn đề:** 
- BaseExecutor được dùng bởi TẤT CẢ providers
- InsecureSkipVerify: true ảnh hưởng toàn bộ router
- Man-in-the-middle attacks có thể xảy ra

### Giải Pháp Đã Áp Dụng
```go
// ✅ KHUYẾN - conditional insecure mode
transport := &http.Transport{
    MaxIdleConns:        100,
    MaxIdleConnsPerHost: 10,
    IdleConnTimeout:     90 * time.Second,
}

// Only skip TLS verification if explicitly enabled via environment variable
if os.Getenv("TI_TLS_INSECURE_MODE") == "true" {
    transport.TLSClientConfig = &tls.Config{
        InsecureSkipVerify: true,
    }
}
```

**Benefits:**
- Mặc định secure (InsecureSkipVerify: false)
- Chỉ enable khi cần thiết qua environment variable
- Dễ kiểm soát trong production vs development

## Best Practices

### 1. Sử dụng Environment Variable Cho Insecure Mode
```bash
# Development
TI_TLS_INSECURE_MODE=true ./routerd

# Production (mặc định)
./routerd
```

### 2. Provider-Specific TLS Config
Nên implement per-provider TLS config thay vì global:
```yaml
providers:
  windsurf:
    tls:
      insecure: true  # chỉ cho windsurf
      ca_file: /path/to/windsurf-ca.pem
  openai:
    tls:
      insecure: false  # mặc định secure
```

### 3. Custom CA Certificate cho Development
```go
// Load custom CA certificate
caCert, err := os.ReadFile(caFile)
caCertPool := x509.NewCertPool()
caCertPool.AppendCertsFromPEM(caCert)
config.RootCAs = caCertPool
```

### 4. Certificate Pinning (Production)
```go
// Pin certificate hash
certPool := x509.NewCertPool()
cert, err := tls.LoadX509KeyPair(certFile, keyFile)
config.RootCAs = certPool
```

## Windsurf Provider Case Study

### Vấn Đề
Windsurf API (`server.windsurf.com`) có certificate verification error:
```
x509: certificate signed by unknown authority
```

### Giải Pháp Hiện Tại
1. **WindsurfProvider:** InsecureSkipVerify: true (provider-specific)
2. **BaseExecutor:** Conditional InsecureSkipVerify via env var
3. **WindsurfOAuth:** InsecureSkipVerify: true (OAuth token validation)

### Lưu Ý
- WindsurfProvider và WindsurfOAuth vẫn có InsecureSkipVerify: true
- BaseExecutor giờ conditional qua environment variable
- Production nên dùng custom CA certificate hoặc fix certificate của Windsurf

## Recommendations

### Ngắn Hạn
1. ✅ BaseExecutor conditional TLS mode (đã implement)
2. ✅ Documentation security implications
3. ✅ Environment variable flag

### Dài Hạn (Future Work)
1. Implement per-provider TLS config in providers.yaml
2. Add certificate pinning for production
3. Implement custom CA bundle loader
4. Add TLS health checks
5. Monitor TLS errors and alert

## Security Checklist

- [x] BaseExecutor mặc định secure (InsecureSkipVerify: false)
- [x] Environment variable flag cho insecure mode
- [x] Documentation security implications
- [ ] Per-provider TLS config in YAML
- [ ] Custom CA certificate support
- [ ] Certificate pinning for production
- [ ] TLS health monitoring
- [ ] Security audit for all providers

## References

- Go TLS Documentation: https://pkg.go.dev/crypto/tls
- OWASP TLS Best Practices
- Ti Router security patterns in `layers/security/tls.go`
