# Error Mapping Patterns trong Go - Kinh Nghiệm từ GitHub

## Mục tiêu
Triển khai Error Mapping cho Ti Router, dịch lỗi từ provider → router → HTTP responses, không để lộ chi tiết nội bộ.

## Các mẫu thiết kế từ GitHub

### 1. Error Boundaries (atharvapandey.com)
- **Nguyên tắc**: Mỗi layer boundary có hàm `translate` để chuyển đổi error vocabulary
- **Flow**: DB errors → Domain errors → Transport errors (HTTP/gRPC)
- **Không để lộ**: Internal details không được thoát ra ngoài
- **Ví dụ**:
  ```go
  func translateDBError(err error, op string) error {
      var pgErr *pgconn.PgError
      if errors.As(err, &pgErr) {
          switch pgErr.Code {
          case "23505": // unique_violation
              return &AppError{Kind: KindConflict, Message: "resource already exists"}
          }
      }
      return fmt.Errorf("internal: %w", err)
  }
  ```

### 2. goerr Package (github.com/tyrenix/goerr)
- **Khái niệm**: Business error với `Code` và `Kind`
- **Cách dùng**:
  ```go
  NotFound = goerr.NewWithSpec("not_found", "not_found", goerr.KindNotFound)
  
  // Wrap với technical cause
  return fmt.Errorf("get user: %w: %w", ErrNotFound, err)
  ```
- **Extract info ở transport layer**:
  ```go
  code, _ := goerr.CodeOf(err)
  kind, _ := goerr.KindOf(err)
  ```

### 3. REST Error Mapping (alesr.github.io)
- **Error Handler**: Chứa `errorMap map[error]*RESTErr`
- **Flow**: Service trả về mapped error → Handler kiểm tra `errors.Is()` → Trả về JSON response
- **Validation**: Kiểm tra error map có đầy đủ không lúc runtime

### 4. Goa DSL Error Mapping
- **API level**: Define error một lần với default mapping
- **Service level**: Make error returnable bởi bất kỳ method nào
- **Method level**: Override mapping nếu cần
- **Transport mapping**: `Response("error_name", func() { Code(-32001) })`

## Áp dụng cho Ti Router

### Hiện trạng
- Router trả về raw errors từ providers (OpenAI, Claude, etc.)
- Không có error mapping layer
- HTTP handlers trả về lỗi trực tiếp, không chuẩn hóa

### Thiết kế Error Mapping

#### 1. Định nghĩa Router Error Types
```go
type RouterError struct {
    Code    string // "provider_error", "rate_limit", "auth_failed", etc.
    Status  int    // HTTP status code
    Message string // User-facing message
    Cause   error  // Original error (logged, not sent to client)
}
```

#### 2. Provider Error Translation
Mỗi provider có hàm `translateProviderError()` để chuyển đổi lỗi đặc thù:
- OpenAI: `rate_limit_exceeded` → `RouterError{Code: "rate_limit", Status: 429}`
- Claude: `invalid_api_key` → `RouterError{Code: "auth_failed", Status: 401}`
- Generic: `context.DeadlineExceeded` → `RouterError{Code: "timeout", Status: 504}`

#### 3. HTTP Response Format
```json
{
  "error": {
    "code": "rate_limit",
    "message": "Rate limit exceeded. Please retry after 60 seconds.",
    "request_id": "req_abc123"
  }
}
```

#### 4. Error Mapping Table
| Provider Error | Router Code | HTTP Status |
|---------------|-------------|-------------|
| rate_limit_exceeded | rate_limit | 429 |
| invalid_api_key | auth_failed | 401 |
| insufficient_quota | quota_exceeded | 403 |
| context.DeadlineExceeded | timeout | 504 |
| unexpected_status 5xx | upstream_error | 502 |
| unexpected_status 4xx | bad_request | 400 |

## Tài liệu tham khảo
- [Error Boundaries Across Layers](https://www.atharvapandey.com/post/go/go-errors-boundaries/)
- [goerr package](https://pkg.go.dev/github.com/tyrenix/goerr/v3)
- [Effective RESTful Error Handling](https://alesr.github.io/posts/rest-errors/)
- [Goa Error Mapping](https://goa.design/docs/4-concepts/7-jsonrpc/6-error-mapping)
- [Error Translation in Go](https://rednafi.com/go/error-translation/)
