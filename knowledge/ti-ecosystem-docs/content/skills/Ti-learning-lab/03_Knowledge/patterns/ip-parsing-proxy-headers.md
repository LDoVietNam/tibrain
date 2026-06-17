# Phân Tích Địa Chỉ IP và Xử Lý Proxy Headers trong Go

> **Ngày tạo**: 2026-04-28  
> **Task**: TR-002 - Parse IP from RemoteAddr  
> **Project**: Ti Router  
> **Trạng thái**: ✅ Hoàn thành

---

## 📋 Tổng Quan

Bài học này mô tả cách phân tích địa chỉ IP từ HTTP request trong Go, bao gồm xử lý proxy headers và IPv6 support.

## 🎯 Vấn Đề Ban Đầu

**Problem**: `r.RemoteAddr` chứa cả IP và port (ví dụ "127.0.0.1:12345"), nhưng rate limiter cần chỉ IP address.

**Location**:
- `Z:\Ti\router\cmd\routerd\handlers_chat.go` line 463: `userIP := r.RemoteAddr`

**Issues**:
- Rate limiter dùng "IP:Port" thay vì chỉ IP
- Không handle proxy scenarios (X-Forwarded-For, X-Real-IP)
- Không validate IP addresses
- Không hỗ trợ IPv6 đúng cách

## ✅ Giải Pháp

### 1. Utility Function

```go
// getClientIP extracts the client IP address from the request
// It handles X-Forwarded-For and X-Real-IP headers for proxy scenarios
// Falls back to RemoteAddr if headers are not present
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header (proxy scenario)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// X-Forwarded-For can contain multiple IPs: "client, proxy1, proxy2"
		// Take the first one (original client)
		ips := strings.Split(xff, ",")
		if len(ips) > 0 {
			ip := strings.TrimSpace(ips[0])
			// Validate IP
			if net.ParseIP(ip) != nil {
				return ip
			}
		}
	}

	// Check X-Real-IP header (nginx proxy)
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		// Validate IP
		if net.ParseIP(xri) != nil {
			return xri
		}
	}

	// Fall back to RemoteAddr
	// RemoteAddr format: "IP:Port" or "[IPv6]:Port"
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		// If SplitHostPort fails, return as-is
		return r.RemoteAddr
	}

	// Validate extracted IP
	if net.ParseIP(ip) != nil {
		return ip
	}

	// If validation fails, return RemoteAddr as fallback
	return r.RemoteAddr
}
```

### 2. Usage in Rate Limiter

**Before:**
```go
userIP := r.RemoteAddr
if rateLimiter != nil && !rateLimiter.AllowUser(userIP, 100, 100) {
```

**After:**
```go
userIP := getClientIP(r)
if rateLimiter != nil && !rateLimiter.AllowUser(userIP, router.rateLimitConfig.DefaultRPM, router.rateLimitConfig.DefaultBurst) {
```

### 3. Go Net Package Functions

**net.SplitHostPort**
```go
// Splits host:port into host and port
// Handles both IPv4 and IPv6
ip, port, err := net.SplitHostPort("127.0.0.1:8080")  // ip="127.0.0.1", port="8080"
ip, port, err := net.SplitHostPort("[::1]:8080")      // ip="::1", port="8080"
```

**net.ParseIP**
```go
// Parses and validates IP address
// Returns nil if invalid
net.ParseIP("127.0.0.1")  // valid IPv4
net.ParseIP("::1")       // valid IPv6
net.ParseIP("invalid")   // nil (invalid)
```

## 🔑 Best Practices

### 1. **Header Priority Order**
1. X-Forwarded-For (proxy chain, take first IP)
2. X-Real-IP (nginx proxy)
3. RemoteAddr (direct connection)

### 2. **Proxy Header Handling**

**X-Forwarded-For Format:**
```
X-Forwarded-For: client, proxy1, proxy2
```
- Chỉ lấy IP đầu tiên (original client)
- Trim whitespace
- Validate IP trước khi sử dụng

**X-Real-IP Format:**
```
X-Real-IP: 192.168.1.1
```
- Single IP address
- Validate trước khi sử dụng

### 3. **IPv6 Support**
- IPv6 addresses wrapped in brackets: `[::1]:8080`
- `net.SplitHostPort()` handles brackets automatically
- `net.ParseIP()` validates both IPv4 and IPv6

### 4. **Validation**
- Luôn validate IP với `net.ParseIP()`
- Fallback nếu validation fails
- Không crash trên invalid input

### 5. **Error Handling**
- `net.SplitHostPort()` có thể fail trên invalid format
- Fallback to RemoteAddr nếu error
- Log warnings cho debugging

## 📁 Files Modified

1. **Z:\Ti\router\cmd\routerd\utils.go**
   - Added `net` import
   - Added `getClientIP()` function
   - Handles X-Forwarded-For, X-Real-IP, RemoteAddr
   - Validates IP addresses

2. **Z:\Ti\router\cmd\routerd\handlers_chat.go**
   - Changed `userIP := r.RemoteAddr` to `userIP := getClientIP(r)`
   - Changed hardcoded `100, 100` to `router.rateLimitConfig.DefaultRPM, router.rateLimitConfig.DefaultBurst`

## 🧪 Test Scenarios

### Scenario 1: IPv4 Direct Connection
**Input**: `RemoteAddr = "127.0.0.1:12345"`  
**Expected**: `"127.0.0.1"`  
**Result**: ✅ PASS

### Scenario 2: IPv6 Direct Connection
**Input**: `RemoteAddr = "[::1]:12345"`  
**Expected**: `"::1"`  
**Result**: ✅ PASS

### Scenario 3: X-Forwarded-For Header
**Input**: `X-Forwarded-For = "192.168.1.1, 10.0.0.1"`  
**Expected**: `"192.168.1.1"` (first IP)  
**Result**: ✅ PASS

### Scenario 4: X-Real-IP Header
**Input**: `X-Real-IP = "192.168.1.1"`  
**Expected**: `"192.168.1.1"`  
**Result**: ✅ PASS

### Scenario 5: Invalid IP in Header
**Input**: `X-Forwarded-For = "invalid-ip"`  
**Expected**: Fallback to RemoteAddr  
**Result**: ✅ PASS

### Scenario 6: Invalid RemoteAddr Format
**Input**: `RemoteAddr = "invalid-format"`  
**Expected**: Return as-is (fallback)  
**Result**: ✅ PASS

## 🎓 Lessons Learned

### IP Parsing Patterns

1. **Use net.SplitHostPort()**
   - Better than `strings.LastIndex()`
   - Handles IPv6 brackets automatically
   - Standard library function

2. **Validate IP Addresses**
   - Use `net.ParseIP()` for validation
   - Reject invalid IPs
   - Fallback gracefully

3. **Proxy Headers Priority**
   - X-Forwarded-For first (standard proxy header)
   - X-Real-IP second (nginx-specific)
   - RemoteAddr last (direct connection)

4. **X-Forwarded-For Format**
   - Can contain multiple IPs (proxy chain)
   - First IP is original client
   - Comma-separated, may have spaces

### Security Considerations

1. **IP Spoofing**
   - X-Forwarded-For có thể bị spoofed
   - Chỉ trust headers từ trusted proxies
   - Consider using CIDR whitelist cho proxy IPs

2. **Validation**
   - Luôn validate IP trước khi sử dụng
   - Reject invalid/malformed IPs
   - Log warnings cho suspicious IPs

3. **Rate Limiting**
   - Use extracted IP cho rate limiting
   - Consider per-user rate limiting với auth
   - Implement IP-based rate limiting với caution

### Go-Specific Patterns

1. **Error Handling**
   - Check error từ `net.SplitHostPort()`
   - Fallback gracefully
   - Don't crash on invalid input

2. **String Manipulation**
   - Use `strings.TrimSpace()` cho header values
   - Use `strings.Split()` cho comma-separated values
   - Trim whitespace properly

## 🔮 Future Improvements

1. **Trusted Proxy Configuration**
   ```go
   type Config struct {
       TrustedProxies []string `yaml:"trusted-proxies"`
   }
   ```
   - Chỉ trust X-Forwarded-For từ trusted proxies
   - Validate proxy IP against whitelist

2. **IP-based Access Control**
   ```go
   type Config struct {
       AllowedIPs []string `yaml:"allowed-ips"`
       BlockedIPs []string `yaml:"blocked-ips"`
   }
   ```
   - Whitelist/blacklist IPs
   - CIDR notation support

3. **GeoIP Lookup**
   - Lookup country/city từ IP
   - Rate limiting per region
   - Compliance với data residency

4. **IPv6 Transition**
   - IPv4-mapped IPv6 addresses
   - Dual-stack support
   - IPv6-specific rate limiting

## 📚 References

- **Go net package**: https://pkg.go.dev/net
- **X-Forwarded-For spec**: https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/X-Forwarded-For
- **X-Real-IP (nginx)**: http://nginx.org/en/docs/http/ngx_http_proxy_module.html#proxy_set_header
- **IPv6 addressing**: https://datatracker.ietf.org/doc/html/rfc4291

## ✅ Verification

- [x] getClientIP function implemented
- [x] net.SplitHostPort used for parsing
- [x] X-Forwarded-For header handled
- [x] X-Real-IP header handled
- [x] IP validation with net.ParseIP
- [x] IPv6 support with brackets
- [x] Rate limiter updated to use getClientIP
- [x] Rate limiter updated to use config values
- [x] Error handling with fallback

---

**Kết luận**: IP parsing với proxy header support giúp rate limiting hoạt động đúng trong cả direct connection và proxy scenarios. net.SplitHostPort và net.ParseIP là standard library functions nên reliable và efficient.
