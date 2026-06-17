# Go Security Best Practices

> **Ngày tạo**: 2026-04-29  
> **Nguồn**: "Best Practices for Secure Error Handling in Go" (JetBrains GoLand Blog)  
> **Mục đích**: Document security best practices cho Go development, đặc biệt là error handling

---

## Tổng Quan

Security trong Go development không chỉ về authentication/authorization, mà còn về:
1. **Error Handling** - Không rò rỉ thông tin nhạy cảm
2. **Input Validation** - Validate và sanitize tất cả input
3. **Secrets Management** - Không hardcode secrets
4. **Output Encoding** - Encode output đúng cách để tránh injection
5. **Resource Limits** - Giới hạn tài nguyên để tránh DoS

---

## Secure Error Handling (Split Brain Pattern)

### Tại sao quan trọng?

Go errors có thể rò rỉ thông tin nhạy cảm:
- Đường dẫn file hệ thống
- Câu truy vấn SQL
- Credentials và token
- Identifier nội bộ
- Stack trace

### Split Brain Pattern

Tách biệt thông tin nội bộ (unsafe) và thông tin công khai (safe):

```go
type SafeError struct {
    Code      string            // Error code cho classification
    UserMsg   string            // Public-safe message cho users
    Internal  error             // Internal error với full details (unsafe)
    Metadata  map[string]string // Safe metadata cho logging
}

func (e *SafeError) Error() string {
    return e.UserMsg  // Chỉ trả message an toàn cho client
}

func (e *SafeError) LogString() string {
    return fmt.Sprintf("Code: %s | Msg: %s | Cause: %v | Meta: %v",
        e.Code, e.UserMsg, e.Internal, e.Metadata)
}
```

**Key Points:**
- `Error()` chỉ trả về UserMsg - an toàn cho client
- `LogString()` chứa đầy đủ chi tiết - cho internal logging
- Metadata chỉ chứa thông tin an toàn (route_id, không password)

### Contextual Sanitization

Thay vì log toàn bộ struct, chỉ log các trường an toàn:

```go
// ❌ Sai - log toàn bộ struct (có thể rò rỉ password)
log.Printf("login failed: %+v", authRequest)

// ✅ Đúng - chỉ log các trường an toàn
log.Printf("login failed: username=%s, ip=%s", 
    req.Username, req.RemoteIP)
```

### Opaque Wrapping

Wrap error để bảo vệ khỏi introspection từ thư viện bên thứ ba:

```go
func GetUserProfile(id string) (*Profile, error) {
    user, err := db.QueryUser(id)
    if err != nil {
        return nil, &SafeError{
            Code:      "FETCH_ERROR",
            UserMsg:   "Unable to retrieve user profile.",
            Internal:  err,  // Chỉ lưu trong internal
        }
    }
    return user, nil
}
```

### Propagate Error An toàn Qua Trust Boundaries

**Qua ranh giới subsystem:** Wrap database error thành domain-specific error

**Qua ranh giới API (service-to-service):** Chuyển đổi thành protocol error chuẩn (gRPC codes)

**Qua ranh giới public:** Chỉ trả message tĩnh, không bao giờ trả error gốc cho client

```go
func translateAndRespond(w http.ResponseWriter, err error) {
    var status int
    var publicMsg string
    switch {
    case errors.Is(err, domain.ErrInvalidInput):
        status = http.StatusBadRequest
        publicMsg = "The provided order details are invalid."
    case errors.Is(err, domain.ErrConflict):
        status = http.StatusConflict
        publicMsg = "This order has already been processed."
    default:
        status = http.StatusInternalServerError
        publicMsg = "An internal error occurred. Please contact support."
    }
    http.Error(w, publicMsg, status)
}
```

### Structured Logging với Sanitization

Sử dụng structured logging (log/slog) và Redactor interface cho dữ liệu nhạy cảm:

```go
type Redactor interface {
    Redact() any
}

func (r LoginRequest) Redact() any {
    return struct {
        Username string `json:"username"`
        Password string `json:"password"`
    }{
        Username: r.Username,
        Password: "***REDACTED***",
    }
}

logger.Info("login attempt", "req", req.Redact())
```

### Security Audit Checklist

Khi review code xử lý lỗi:
- Caller có phải là external không? → Nếu có, chỉ trả message tĩnh
- Error có chứa dữ liệu nhạy cảm không? → Nếu có, dùng SafeError
- Error có đi qua trust boundary không? → Nếu có, wrap lại với message an toàn
- Log có chứa thông tin nhạy cảm không? → Nếu có, implement Redactor
- Error có expose chi tiết thư viện bên thứ ba không? → Nếu có, dùng opaque wrapping

---

## Input Validation

### Validate All Input

```go
// ❌ Sai - không validate input
func HandleLogin(w http.ResponseWriter, r *http.Request) {
    username := r.FormValue("username")
    password := r.FormValue("password")
    // Process login without validation
}

// ✅ Đúng - validate input
func HandleLogin(w http.ResponseWriter, r *http.Request) {
    username := r.FormValue("username")
    password := r.FormValue("password")
    
    // Validate username length and format
    if len(username) < 3 || len(username) > 50 {
        http.Error(w, "Invalid username length", http.StatusBadRequest)
        return
    }
    
    // Validate password complexity
    if len(password) < 8 {
        http.Error(w, "Password too short", http.StatusBadRequest)
        return
    }
    
    // Sanitize input
    username = strings.TrimSpace(username)
    
    // Process login
}
```

### Use MaxBytesReader

Giới hạn kích thước request body để tránh DoS:

```go
maxBodySize := int64(10 * 1024 * 1024) // 10MB
r.Body = http.MaxBytesReader(w, r.Body, maxBodySize)
body, err := io.ReadAll(r.Body)
if err != nil {
    http.Error(w, "request body too large", http.StatusRequestEntityTooLarge)
    return
}
```

---

## Secrets Management

### Không Hardcode Secrets

```go
// ❌ Sai - hardcode secrets
const API_KEY = "sk-1234567890abcdef"

// ✅ Đúng - load từ environment
apiKey := os.Getenv("API_KEY")
if apiKey == "" {
    log.Fatal("API_KEY environment variable not set")
}
```

### Use Secure Storage

- Environment variables (production)
- Secret management systems (HashiCorp Vault, AWS Secrets Manager)
- Encrypted config files

---

## Output Encoding

### Encode JSON Responses

```go
// ✅ Đúng - use json.Encoder với proper error handling
func respondJSON(w http.ResponseWriter, status int, data interface{}) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    if err := json.NewEncoder(w).Encode(data); err != nil {
        log.Printf("Failed to encode JSON response: %v", err)
    }
}
```

### Prevent XSS

```go
// ✅ Đúng - escape HTML output
func renderTemplate(w http.ResponseWriter, tmpl string, data interface{}) {
    w.Header().Set("Content-Type", "text/html; charset=utf-8")
    if err := templates.ExecuteTemplate(w, tmpl, data); err != nil {
        http.Error(w, "Internal server error", http.StatusInternalServerError)
    }
}
```

---

## Resource Limits

### Limit Concurrent Requests

```go
// ✅ Đúng - use semaphore to limit concurrent requests
sem := make(chan struct{}, 100) // Max 100 concurrent requests

func handleRequest(w http.ResponseWriter, r *http.Request) {
    sem <- struct{}{}        // Acquire
    defer func() { <-sem }() // Release
    
    // Process request
}
```

### Timeout Context

```go
// ✅ Đúng - use context with timeout
ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
defer cancel()

resp, err := httpClient.Do(req.WithContext(ctx))
```

---

## Common Security Mistakes

### 1. Logging Sensitive Data

```go
// ❌ Sai - log password
log.Printf("User login: username=%s, password=%s", username, password)

// ✅ Đúng - không log password
log.Printf("User login: username=%s", username)
```

### 2. Returning Error Details to Client

```go
// ❌ Sai - return internal error details
http.Error(w, fmt.Sprintf("Database error: %v", err), http.StatusInternalServerError)

// ✅ Đúng - return generic error
http.Error(w, "Internal server error", http.StatusInternalServerError)
log.Printf("Database error: %v", err) // Log internally
```

### 3. Not Validating Input

```go
// ❌ Sai - không validate input
id := r.URL.Query().Get("id")
user, err := db.GetUser(id)

// ✅ Đúng - validate input
id := r.URL.Query().Get("id")
if !isValidID(id) {
    http.Error(w, "Invalid ID", http.StatusBadRequest)
    return
}
user, err := db.GetUser(id)
```

---

## Security Checklist

### Code Review Checklist

- [ ] Không hardcode secrets
- [ ] Validate tất cả input
- [ ] Giới hạn kích thước request body
- [ ] Không log sensitive data
- [ ] Không return error details cho client
- [ ] Use context với timeout
- [ ] Giới hạn concurrent requests
- [ ] Encode output đúng cách
- [ ] Use SafeError pattern cho error handling
- [ ] Implement Redactor cho structured logging

### Deployment Checklist

- [ ] Environment variables configured
- [ ] Secrets loaded from secure storage
- [ ] TLS/SSL enabled
- [ ] Rate limiting configured
- [ ] Monitoring/logging enabled
- [ ] Security scanning performed

---

## References

- "Best Practices for Secure Error Handling in Go" - JetBrains GoLand Blog
- OWASP Go Security Cheat Sheet
- Go Security Guidelines (golang.org)

---

*Last Updated: 2026-04-29*
