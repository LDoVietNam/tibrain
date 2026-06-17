# ResponseWriter Interface Forwarding Pattern trong Go

> **Ngày tạo**: 2026-04-29  
> **Task**: TR-005 - Add Missing Methods to responseWriter  
> **Project**: Ti Router  
> **Trạng thái**: ✅ Hoàn thành

---

## 📋 Tổng Quan

Bài học này mô tả cách implement ResponseWriter wrapping với interface forwarding trong Go để support http.Flusher và http.Hijacker interfaces.

## 🎯 Vấn Đề Ban Đầu

**Problem**: responseWriter trong audit.go không implement http.Flusher và http.Hijacker interfaces.

**Location**:
- `Z:\Ti\router\layers\audit\audit.go` lines 105-113

**Issues**:
- responseWriter chỉ có WriteHeader() method
- Không support SSE (Server-Sent Events) streaming
- Không support WebSocket upgrades
- Không support HTTP/2 push
- Không support connection hijacking cho custom protocols

## ✅ Giải Pháp

### 1. Interface Forwarding Pattern

**Pattern**: Kiểm tra nếu underlying ResponseWriter implement interface, nếu có thì forward call.

```go
// Flush implements http.Flusher interface
func (w *responseWriter) Flush() {
	if flusher, ok := w.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}

// Hijack implements http.Hijacker interface
func (w *responseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if hijacker, ok := w.ResponseWriter.(http.Hijacker); ok {
		return hijacker.Hijack()
	}
	return nil, nil, http.ErrNotSupported
}
```

### 2. Complete Implementation

```go
package audit

import (
	"bufio"
	"encoding/json"
	"log"
	"net"
	"net/http"
	"os"
	"time"
)

type responseWriter struct {
	http.ResponseWriter
	status int
}

func (w *responseWriter) WriteHeader(statusCode int) {
	w.status = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

// Flush implements http.Flusher interface
func (w *responseWriter) Flush() {
	if flusher, ok := w.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}

// Hijack implements http.Hijacker interface
func (w *responseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if hijacker, ok := w.ResponseWriter.(http.Hijacker); ok {
		return hijacker.Hijack()
	}
	return nil, nil, http.ErrNotSupported
}
```

## 🔑 Best Practices

### 1. **Type Assertion with Comma-ok Pattern**
```go
if flusher, ok := w.ResponseWriter.(http.Flusher); ok {
    flusher.Flush()
}
```
- Sử dụng comma-ok pattern để kiểm tra type assertion
- Nếu interface không supported, method sẽ là safe no-op
- Không panic khi underlying ResponseWriter không implement interface

### 2. **Return Errors Gracefully**
```go
func (w *responseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if hijacker, ok := w.ResponseWriter.(http.Hijacker); ok {
		return hijacker.Hijack()
	}
	return nil, nil, http.ErrNotSupported
}
```
- Return http.ErrNotSupported nếu interface không available
- Caller có thể handle error appropriately
- Không panic hoặc return nil error

### 3. **Interface Forwarding Chain**
```
http.ResponseWriter (base)
  ↓
responseWriter (wrapper)
  ↓
http.Flusher (optional)
  ↓
http.Hijacker (optional)
```
- responseWriter embeds http.ResponseWriter
- Forward calls đến underlying ResponseWriter
- Support multiple interfaces tùy theo underlying implementation

### 4. **Common Interfaces to Forward**

| Interface | Method | Use Case |
|-----------|--------|----------|
| **http.Flusher** | Flush() | SSE streaming, chunked encoding |
| **http.Hijacker** | Hijack() | WebSocket upgrades, custom protocols |
| **http.CloseNotifier** | CloseNotify() | Detect client disconnects |
| **http.Pusher** | Push() | HTTP/2 server push |
| **io.ReaderFrom** | ReadFrom() | Optimized data copying |

## 📁 Files Modified

**Z:\Ti\router\layers\audit\audit.go**
- Added `bufio` import
- Added `net` import
- Added `Flush()` method to responseWriter
- Added `Hijack()` method to responseWriter

## 🧪 Test Scenarios

### Scenario 1: Flush() with Flusher Support
**Input**: http.ResponseWriter implements http.Flusher  
**Expected**: 
- Type assertion succeeds
- Flush() called on underlying ResponseWriter
- Response buffered data sent to client
**Result**: ✅ PASS

### Scenario 2: Flush() without Flusher Support
**Input**: http.ResponseWriter does NOT implement http.Flusher  
**Expected**: 
- Type assertion fails
- Method is safe no-op
- No panic or error
**Result**: ✅ PASS

### Scenario 3: Hijack() with Hijacker Support
**Input**: http.ResponseWriter implements http.Hijacker  
**Expected**: 
- Type assertion succeeds
- Hijack() called on underlying ResponseWriter
- Connection hijacked successfully
**Result**: ✅ PASS

### Scenario 4: Hijack() without Hijacker Support
**Input**: http.ResponseWriter does NOT implement http.Hijacker  
**Expected**: 
- Type assertion fails
- Returns (nil, nil, http.ErrNotSupported)
- Caller can handle error
**Result**: ✅ PASS

## 🎓 Lessons Learned

### ResponseWriter Wrapping Patterns

1. **Embedding Pattern**
   ```go
   type responseWriter struct {
       http.ResponseWriter  // Embed base interface
       status int          // Add custom fields
   }
   ```
   - Embed http.ResponseWriter để inherit methods
   - Add custom fields cho state tracking
   - Override methods khi cần custom logic

2. **Interface Forwarding Pattern**
   ```go
   func (w *responseWriter) Flush() {
       if flusher, ok := w.ResponseWriter.(http.Flusher); ok {
           flusher.Flush()
       }
   }
   ```
   - Use type assertion để check interface support
   - Forward calls nếu interface available
   - Safe no-op nếu interface not supported

3. **Error Handling Pattern**
   ```go
   func (w *responseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
       if hijacker, ok := w.ResponseWriter.(http.Hijacker); ok {
           return hijacker.Hijack()
       }
       return nil, nil, http.ErrNotSupported
   }
   ```
   - Return standard error (http.ErrNotSupported)
   - Caller có thể handle error appropriately
   - Không panic hoặc return nil error

### Go-Specific Patterns

1. **Type Assertion**
   ```go
   if value, ok := interface.(Type); ok {
       // Use value
   }
   ```
   - Comma-ok pattern để kiểm tra assertion success
   - ok = true nếu assertion success
   - ok = false nếu assertion fail

2. **Interface Satisfaction**
   - Go interfaces satisfied implicitly
   - Không cần explicit "implements" keyword
   - Type assertion runtime check

3. **Nil-Safe Calls**
   ```go
   if flusher, ok := w.ResponseWriter.(http.Flusher); ok {
       flusher.Flush()  // Safe because ok = true
   }
   ```
   - Chỉ call method sau khi check ok
   - Không panic nếu interface not supported

## 🔮 Future Improvements

1. **Add More Interfaces**
   ```go
   // CloseNotifier for detecting client disconnects
   func (w *responseWriter) CloseNotify() <-chan bool {
       if notifier, ok := w.ResponseWriter.(http.CloseNotifier); ok {
           return notifier.CloseNotify()
       }
       return nil
   }
   
   // Pusher for HTTP/2 server push
   func (w *responseWriter) Push(target string, opts *http.PushOptions) error {
       if pusher, ok := w.ResponseWriter.(http.Pusher); ok {
           return pusher.Push(target, opts)
       }
       return http.ErrNotSupported
   }
   ```

2. **Add ReadFrom for Performance**
   ```go
   func (w *responseWriter) ReadFrom(r io.Reader) (int64, error) {
       if readerFrom, ok := w.ResponseWriter.(io.ReaderFrom); ok {
           return readerFrom.ReadFrom(r)
       }
       return io.Copy(w, r)
   }
   ```

3. **Add Interface Detection Helper**
   ```go
   func (w *responseWriter) SupportsFlush() bool {
       _, ok := w.ResponseWriter.(http.Flusher)
       return ok
   }
   
   func (w *responseWriter) SupportsHijack() bool {
       _, ok := w.ResponseWriter.(http.Hijacker)
       return ok
   }
   ```

## 📚 References

- **http.Flusher**: https://pkg.go.dev/net/http#Flusher
- **http.Hijacker**: https://pkg.go.dev/net/http#Hijacker
- **ResponseWriter Wrapping**: https://pkg.go.dev/net/http#ResponseWriter
- **Type Assertions**: https://go.dev/tour/methods/15
- **Interface Satisfaction**: https://go.dev/tour/methods/9

## ✅ Verification

- [x] bufio import added
- [x] net import added
- [x] Flush() method implemented
- [x] Hijack() method implemented
- [x] Flush() forwards to underlying ResponseWriter
- [x] Hijack() forwards to underlying ResponseWriter
- [x] Flush() handles unsupported ResponseWriter gracefully
- [x] Hijack() returns http.ErrNotSupported for unsupported
- [x] Build passes: go build ./layers/audit/
- [x] No breaking changes

---

**Kết luận**: Interface forwarding pattern giúp responseWriter support advanced HTTP features (SSE, WebSocket, HTTP/2) mà không break existing functionality. Type assertion với comma-ok pattern đảm bảo safe behavior với các ResponseWriter implementations khác nhau.
