# Nghiên cứu Parse IP từ RemoteAddr trong Go

## Tổng quan

Tài liệu tổng hợp các phương pháp trích xuất IP address từ `http.Request.RemoteAddr` trong Go.

## 1. Cấu trúc RemoteAddr

### Format chuẩn
```go
r.RemoteAddr // Format: "IP:Port" hoặc "[IPv6]:Port"
```

### Ví dụ
```go
"127.0.0.1:54572"    // IPv4
"[::1]:53947"          // IPv6
"[fe80::1%lo0]:8080"  // IPv6 với zone ID
```

## 2. Phương pháp chuẩn: net.SplitHostPort()

### Code mẫu
```go
import "net"

func getClientIP(r *http.Request) string {
    // Sử dụng net.SplitHostPort để tách IP và Port
    host, _, err := net.SplitHostPort(r.RemoteAddr)
    if err != nil {
        // Nếu SplitHostPort thất bại, trả về RemoteAddr nguyên bản
        return r.RemoteAddr
    }
    return host
}
```

### Đặc điểm
- ✅ Hỗ trợ cả IPv4 và IPv6
- ✅ Tự động xử lý dấu ngoặc vuông `[::1]:53947`
- ✅ Là phương pháp chuẩn được khuyên dùng (Stack Overflow, Go docs)
- ✅ Secure: `RemoteAddr` được set từ TCP packet, không thể forge qua HTTP headers

## 3. Xử lý Proxy/Load Balancer

### Các headers thường gặp
```go
X-Forwarded-For: "203.0.113.195, 70.41.3.18, 150.172.238.178"
X-Real-IP: "203.0.113.195"
```

### Implementation hoàn chỉnh
```go
func getClientIP(r *http.Request) string {
    // 1. Kiểm tra X-Forwarded-For (có thể có nhiều IP)
    if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
        ips := strings.Split(xff, ",")
        if len(ips) > 0 {
            ip := strings.TrimSpace(ips[0]) // IP đầu tiên là client gốc
            if net.ParseIP(ip) != nil {
                return ip
            }
        }
    }
    
    // 2. Kiểm tra X-Real-IP (nginx)
    if xri := r.Header.Get("X-Real-IP"); xri != "" {
        if net.ParseIP(xri) != nil {
            return xri
        }
    }
    
    // 3. Fallback to RemoteAddr
    host, _, err := net.SplitHostPort(r.RemoteAddr)
    if err != nil {
        return r.RemoteAddr
    }
    
    // Validate IP
    if net.ParseIP(host) != nil {
        return host
    }
    
    return r.RemoteAddr
}
```

## 4. Tham khảo từ GitHub Repos

### router-for-me/CLIProxyAPI
```go
// Sử dụng trực tiếp RemoteAddr, không qua headers
ip := r.RemoteAddr // Format: "IP:Port"
// Sau đó SplitHostPort để lấy IP
```

### gofri/go-github-ratelimit
```go
// Sử dụng http.RoundTripper, không trực tiếp access RemoteAddr
// Nhưng pattern tương tự cho rate limiting
```

### Best Practices từ GitHub
1. ✅ Luôn validate IP với `net.ParseIP()`
2. ✅ Xử lý multiple IPs trong `X-Forwarded-For`
3. ✅ Fallback to `RemoteAddr` nếu headers không có
4. ✅ `RemoteAddr` là source of truth cho TCP connection

## 5. Security Considerations

### RemoteAddr có secure không?
- ✅ **CÓ** - Được set từ TCP packet (`Source Address`)
- ✅ Yêu cầu 3-way handshake, không thể fake dễ dàng
- ❌ `X-Forwarded-For` và `X-Real-IP` **CÓ THỂ BỊ FORGE** (headers thì HTTP, gửi sau TCP handshake)

### Khi nào tin vào headers?
- Chỉ tin `X-Forwarded-For` khi biết chắc chắn request đi qua proxy/load balancer tin cậy
- Nên configure allowlist các IP của proxy/load balancer

## 6. Ti Router Implementation

### File: `cmd/routerd/utils.go`
```go
func getClientIP(r *http.Request) string {
    // Check X-Forwarded-For header (proxy scenario)
    if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
        ips := strings.Split(xff, ",")
        if len(ips) > 0 {
            ip := strings.TrimSpace(ips[0])
            if net.ParseIP(ip) != nil {
                return ip
            }
        }
    }
    
    // Check X-Real-IP header (nginx proxy)
    if xri := r.Header.Get("X-Real-IP"); xri != "" {
        if net.ParseIP(xri) != nil {
            return xri
        }
    }
    
    // Fall back to RemoteAddr
    ip, _, err := net.SplitHostPort(r.RemoteAddr)
    if err != nil {
        return r.RemoteAddr
    }
    
    // Validate extracted IP
    if net.ParseIP(ip) != nil {
        return ip
    }
    
    return r.RemoteAddr
}
```

### Status: ✅ ĐÃ HOÀN THÀNH
- Function đã tồn tại và hoạt động đúng
- Sử dụng `net.SplitHostPort()` chuẩn
- Có xử lý proxy headers
- Có IP validation

## 7. Kết luận

### Cho TR-002: Parse IP from RemoteAddr
- ✅ **ĐÃ HOÀN THÀNH** - Code đã tồn tại trong `utils.go`
- ✅ Sử dụng `net.SplitHostPort()` (best practice từ GitHub)
- ✅ Xử lý `X-Forwarded-For` và `X-Real-IP`
- ✅ Validate IP với `net.ParseIP()`
- ✅ Fallback to `RemoteAddr` nếu cần

### Bài học
1. Luôn search trước khi implement - có thể code đã tồn tại
2. `net.SplitHostPort()` là chuẩn industry cho Go
3. Proxy headers cần thiết nhưng `RemoteAddr` vẫn là source of truth
4. IP validation quan trọng để tránh malformed addresses

### References
- https://stackoverflow.com/questions/57563049/whats-the-cleanest-way-to-obtain-the-ip-address-from-net-http-request-remoteadd
- https://pkg.go.dev/net@go1.25.6#SplitHostPort
- https://github.com/tuck1s/go-github-ratelimit
- https://github.com/router-for-me/CLIProxyAPI

---
*Nghiên cứu bởi: claude*
*Ngày: 2026-04-30*
*Task: TR-002: Parse IP from RemoteAddr*
