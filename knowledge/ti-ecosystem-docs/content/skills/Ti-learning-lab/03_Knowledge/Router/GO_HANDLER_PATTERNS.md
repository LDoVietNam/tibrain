# Go HTTP Handler Patterns

> **Ngày tạo**: 2026-04-28  
> **Nguồn**: golang-patterns skill + Code review handlers/routes package + Secure Error Handling (code4func.com)  
> **Mục đích**: Document các patterns best practices cho Go HTTP handler organization

---

## Tổng Quan

Go HTTP handler organization theo các principles:
1. **Simplicity and Clarity** - Code rõ ràng, dễ đọc
2. **Avoid Package-Level State** - Sử dụng dependency injection
3. **Accept Interfaces, Return Structs** - Functions accept interface params, return concrete types
4. **Error Handling with Context** - Wrap errors với context
5. **Secure Error Handling** - Tách biệt internal unsafe message và public safe message (Split Brain pattern)

---

## Project Layout Standard

```
myproject/
├── cmd/
│   └── myapp/
│       └── main.go           # Entry point
├── internal/
│   ├── handler/              # HTTP handlers
│   ├── service/              # Business logic
│   ├── repository/           # Data access
│   └── config/               # Configuration
├── pkg/
│   └── client/               # Public API client
└── go.mod
```

---

## Package Organization

### Package Naming

✅ **Tốt**:
- Short, lowercase, no underscores
- Ví dụ: `http`, `json`, `user`, `routes`

❌ **Tránh**:
- Verbose: `httpHandler`, `json_parser`
- Mixed case: `UserService`
- Redundant suffix: `userService`

### Avoid Package-Level State

❌ **Tránh** - Global mutable state:
```go
var db *sql.DB

func init() {
    db, _ = sql.Open("postgres", os.Getenv("DATABASE_URL"))
}
```

✅ **Nên dùng** - Dependency injection:
```go
type Server struct {
    db *sql.DB
}

func NewServer(db *sql.DB) *Server {
    return &Server{db: db}
}
```

---

## Interface Design

### Accept Interfaces, Return Structs

✅ **Tốt**:
```go
func ProcessData(r io.Reader) (*Result, error) {
    data, err := io.ReadAll(r)
    if err != nil {
        return nil, err
    }
    return &Result{Data: data}, nil
}
```

❌ **Tránh**:
```go
func ProcessData(r io.Reader) (io.Reader, error) {
    // Hides implementation details unnecessarily
}
```

### Define Interfaces Where They're Used

```go
// Trong consumer package, không phải provider package
package service

type UserStore interface {
    GetUser(id string) (*User, error)
    SaveUser(user *User) error
}

type Service struct {
    store UserStore
}
```

---

## Handler Organization Pattern

### Structure từ handlers/routes package

```
handlers/routes/
├── types.go          # Type definitions (Route, Model)
├── storage.go        # Storage interface & implementation
├── handlers.go       # HTTP handler functions
├── helpers.go        # Helper functions (respondJSON, respondError)
└── adapter.go        # Backward compatibility (TODO: remove after refactor)
```

### Type Definitions (types.go)

```go
// Route represents a virtual model endpoint with routing strategy
type Route struct {
    ID          string   `json:"id"`
    Name        string   `json:"name"`
    Slug        string   `json:"slug"`
    Description string   `json:"description"`
    Strategy    string   `json:"strategy"`
    Models      []Model  `json:"models"`
    Fallbacks   []string `json:"fallbacks"`
    Enabled     bool     `json:"enabled"`
}
```

### Storage Interface (storage.go)

```go
// RouteStorage defines the interface for route storage
type RouteStorage interface {
    Get(id string) (*Route, bool)
    List() []*Route
    Create(route *Route) error
    Update(id string, route *Route) error
    Delete(id string) error
    Exists(id string) bool
}

// InMemoryRouteStorage implements RouteStorage with in-memory storage
type InMemoryRouteStorage struct {
    mu    sync.RWMutex
    routes map[string]*Route
    nextID int
}
```

**Key Points**:
- Interface cho phép dễ test (mock storage)
- Thread-safe với mutex
- Dependency injection vào handler

### Handler Functions (handlers.go)

```go
// Handler holds route handler dependencies
type Handler struct {
    storage RouteStorage
}

// NewHandler creates a new route handler
func NewHandler(storage RouteStorage) *Handler {
    return &Handler{
        storage: storage,
    }
}

// HandleRoutes handles route listing and creation
func (h *Handler) HandleRoutes(w http.ResponseWriter, r *http.Request) {
    switch r.Method {
    case "GET":
        h.ListRoutes(w, r)
    case "POST":
        h.CreateRoute(w, r)
    default:
        respondError(w, http.StatusMethodNotAllowed, "Method not allowed")
    }
}
```

**Key Points**:
- Handler struct holds dependencies (storage)
- Dependency injection qua constructor
- Method routing theo HTTP method

### Helper Functions (helpers.go)

```go
// respondJSON sends JSON response
func respondJSON(w http.ResponseWriter, status int, data interface{}) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    if err := json.NewEncoder(w).Encode(data); err != nil {
        log.Printf("Failed to encode JSON response: %v", err)
    }
}

// respondError sends error response
func respondError(w http.ResponseWriter, status int, message string) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    if err := json.NewEncoder(w).Encode(map[string]string{"error": message}); err != nil {
        log.Printf("Failed to encode error response: %v", err)
    }
}
```

**Key Points**:
- Luôn handle error từ JSON encoding
- Log error để debug
- Consistent response format

---

## Common Pitfalls & Solutions

### 1. ID Generation Bug

❌ **Sai** - Modulo arithmetic gây collision:
```go
func GenerateID(slug string, counter int) string {
    return slug + "-" + string(rune('0'+counter%10))  // Chỉ 0-9
}
```

✅ **Đúng** - Sử dụng strconv.Itoa:
```go
func GenerateID(slug string, counter int) string {
    return slug + "-" + strconv.Itoa(counter)
}
```

### 2. Missing Error Handling

❌ **Sai** - Ignore error:
```go
json.NewEncoder(w).Encode(data)  // Error ignored
```

✅ **Đúng** - Handle error:
```go
if err := json.NewEncoder(w).Encode(data); err != nil {
    log.Printf("Failed to encode JSON response: %v", err)
}
```

### 3. 204 Response with Body

❌ **Sai** - 204 không nên có body:
```go
respondJSON(w, http.StatusNoContent, nil)  // Sai
```

✅ **Đúng** - Chỉ set status code:
```go
w.WriteHeader(http.StatusNoContent)
```

### 4. Modifying Storage Directly

❌ **Sai** - Modify retrieved pointer:
```go
route, _ := storage.Get(id)
route.Name = newName  // Modify storage directly
storage.Update(id, route)
```

✅ **Đúng** - Create copy trước khi modify:
```go
route, _ := storage.Get(id)
updated := *route  // Create copy
updated.Name = newName
storage.Update(id, &updated)
```

### 5. Error Comparison

❌ **Sai** - Pointer comparison:
```go
var ErrNotFound = &RouteError{Message: "not found"}
if err == ErrNotFound { ... }  // Sai với wrapped errors
```

✅ **Đúng** - Use errors.Is:
```go
var ErrNotFound = errors.New("not found")
if errors.Is(err, ErrNotFound) { ... }
```

---

## Error Handling Patterns

### Error Wrapping with Context

```go
func LoadConfig(path string) (*Config, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, fmt.Errorf("load config %s: %w", path, err)
    }
    // ...
}
```

### Sentinel Errors

```go
var (
    ErrNotFound     = errors.New("resource not found")
    ErrUnauthorized = errors.New("unauthorized")
    ErrInvalidInput = errors.New("invalid input")
)
```

### Error Checking with errors.Is/errors.As

```go
func HandleError(err error) {
    if errors.Is(err, sql.ErrNoRows) {
        log.Println("No records found")
        return
    }

    var validationErr *ValidationError
    if errors.As(err, &validationErr) {
        log.Printf("Validation error: %s", validationErr.Message)
        return
    }
}
```

---

## Thread Safety

### Mutex Usage

```go
type InMemoryRouteStorage struct {
    mu    sync.RWMutex
    routes map[string]*Route
}

func (s *InMemoryRouteStorage) Get(id string) (*Route, bool) {
    s.mu.RLock()  // Read lock
    defer s.mu.RUnlock()
    route, ok := s.routes[id]
    return route, ok
}

func (s *InMemoryRouteStorage) Create(route *Route) error {
    s.mu.Lock()  // Write lock
    defer s.mu.Unlock()
    // ...
}
```

**Key Points**:
- RWMutex cho read-heavy workloads
- Always defer unlock
- Lock minimal critical section

---

## Backward Compatibility

### Adapter Pattern

```go
// adapter.go - Temporary backward compatibility
var defaultStorage = NewInMemoryRouteStorage()
var defaultHandler = NewHandler(defaultStorage)

func HandleRoutes(w http.ResponseWriter, r *http.Request) {
    defaultHandler.HandleRoutes(w, r)
}

// TODO: Remove after main.go refactor
```

---

## Lessons Learned từ handlers/routes Refactor

1. **Dependency Analysis Critical** - Phân tích dependencies trước refactor để identify safe order
2. **Zero Coupling Files First** - Refactor files với zero coupling trước (handlers_routes.go)
3. **Interface Design** - Define storage interface để dễ test và mock
4. **Thread Safety** - Luôn dùng mutex cho shared state
5. **Error Handling** - Never ignore errors, luôn log để debug
6. **Code Review Essential** - Sub-agent review tìm thấy bugs không obvious

---

## Secure Error Handling Patterns (Split Brain)

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

## References

- golang-patterns skill: C:\Users\MIN\.claude\skills\best\golang-patterns\SKILL.md
- handlers/routes package: Z:\Ti\router\cmd\routerd\handlers\routes\
- DEPENDENCY_ANALYSIS.md: Z:\Ti\router\DEPENDENCY_ANALYSIS.md
- REFACTOR_LOG.md: Z:\Ti\router\REFACTOR_LOG.md

---

*Last Updated: 2026-04-28*
