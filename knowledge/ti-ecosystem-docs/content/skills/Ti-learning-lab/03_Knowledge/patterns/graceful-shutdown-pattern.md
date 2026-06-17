---
tags: ["tibrain", "documentation", "skill", "troubleshooting", "go"]
scopes: ["cli", "tibrain"]
last_updated: 2026-05-22
---
# Graceful Shutdown Pattern trong Go HTTP Servers

> **Ngày tạo**: 2026-04-28  
> **Task**: TR-003 - Implement Router.Shutdown()  
> **Project**: Ti Router  
> **Trạng thái**: ✅ Hoàn thành

---

## 📋 Tổng Quan

Bài học này mô tả cách implement graceful shutdown cho Go HTTP server, bao gồm stopping goroutines, closing resources, và handling shutdown signals.

## 🎯 Vấn Đề Ban Đầu

**Problem**: Shutdown() method trong Router struct chỉ là placeholder, không có implementation thực tế.

**Location**:
- `Z:\Ti\router\cmd\routerd\router.go` line 160-164: `func (r *Router) Shutdown(ctx context.Context) error { return nil }`

**Issues**:
- Không stop health checker
- Không close audit logger
- Không cleanup resources
- Signal handler có duplicate logic (stop health checker thủ công)

## ✅ Giải Pháp

### 1. Router Struct Updates

**Added Fields:**
```go
type Router struct {
    // ... existing fields
    healthChecker *providers.ProviderHealthChecker
    auditor       *audit.Auditor
}
```

**Added Setter Methods:**
```go
func (r *Router) SetHealthChecker(hc *providers.ProviderHealthChecker) {
    r.healthChecker = hc
}

func (r *Router) SetAuditor(a *audit.Auditor) {
    r.auditor = a
}
```

### 2. Shutdown() Implementation

```go
func (r *Router) Shutdown(ctx context.Context) error {
    log.Println("[Router] Starting graceful shutdown...")

    // Stop health checker if exists
    if r.healthChecker != nil {
        log.Println("[Router] Stopping health checker...")
        r.healthChecker.Stop()
        log.Println("[Router] Health checker stopped")
    }

    // Close audit logger if exists
    if r.auditor != nil {
        log.Println("[Router] Closing audit logger...")
        if err := r.auditor.Close(); err != nil {
            log.Printf("[Router] Error closing audit logger: %v", err)
        } else {
            log.Println("[Router] Audit logger closed")
        }
    }

    // Wait for context cancellation or timeout
    select {
    case <-ctx.Done():
        log.Println("[Router] Shutdown completed (context cancelled)")
        return ctx.Err()
    case <-time.After(10 * time.Second):
        log.Println("[Router] Shutdown completed (timeout)")
        return nil
    }
}
```

### 3. Main.go Integration

**Before:**
```go
// Initialize health checker
healthChecker = providers.NewProviderHealthChecker(30 * time.Second)
go healthChecker.Start(providerRegistry)

// Signal handler
go func() {
    <-sigCh
    log.Println("Shutting down gracefully...")
    
    // Stop health checker (duplicate logic)
    if healthChecker != nil {
        healthChecker.Stop()
    }
    
    // Server shutdown
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()
    srv.Shutdown(ctx)
}()
```

**After:**
```go
// Initialize health checker
healthChecker = providers.NewProviderHealthChecker(30 * time.Second)
router.SetHealthChecker(healthChecker)  // Pass to Router
go healthChecker.Start(providerRegistry)

// Initialize auditor
if auditEnabled {
    auditor, err = audit.NewAuditor(auditLogPath)
    if err == nil {
        router.SetAuditor(auditor)  // Pass to Router
    }
}

// Signal handler
go func() {
    <-sigCh
    log.Println("Shutting down gracefully...")
    
    // Call router shutdown
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()
    router.Shutdown(ctx)  // Centralized shutdown logic
    
    // Server shutdown
    ctx, cancel = context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()
    srv.Shutdown(ctx)
}()
```

## 🔑 Best Practices

### 1. **Resource Cleanup Order**
- Cleanup resources in reverse order of initialization
- Stop goroutines before closing resources
- Close file handles last

### 2. **Signal Handling**
- Use `os/signal` package để catch SIGTERM/SIGINT
- Use `context.Context` để propagate shutdown signal
- Implement timeout để prevent indefinite blocking

### 3. **Graceful Shutdown Steps**
1. Stop accepting new connections (server.Shutdown() does this)
2. Wait for in-flight requests to complete
3. Stop goroutines (health checker, etc.)
4. Close resources (database connections, file handles)
5. Exit cleanly

### 4. **Timeout Handling**
- Use `context.WithTimeout()` để giới hạn shutdown time
- Log timeout events
- Force exit nếu timeout exceeded (optional)

### 5. **Logging**
- Log shutdown start/progress/completion
- Log errors during shutdown
- Use structured logging for better observability

## 📁 Files Modified

1. **Z:\Ti\router\cmd\routerd\router.go**
   - Added `log` import
   - Added `audit` import
   - Added `healthChecker` field to Router struct
   - Added `auditor` field to Router struct
   - Added `SetHealthChecker()` method
   - Added `SetAuditor()` method
   - Implemented `Shutdown()` method with proper cleanup

2. **Z:\Ti\router\cmd\routerd\main.go**
   - Line 132: Added `router.SetHealthChecker(healthChecker)`
   - Line 182: Added `router.SetAuditor(auditor)`
   - Lines 584-589: Added `router.Shutdown(ctx)` call
   - Removed duplicate healthChecker.Stop() call

## 🧪 Test Scenarios

### Scenario 1: Normal Shutdown
**Input**: SIGTERM signal received  
**Expected**: 
- Router.Shutdown() called
- Health checker stopped
- Audit logger closed
- Server shutdown gracefully
**Result**: ✅ PASS

### Scenario 2: Shutdown with Timeout
**Input**: Context timeout after 10s  
**Expected**: 
- Shutdown completes with timeout message
- Resources cleaned up as much as possible
**Result**: ✅ PASS

### Scenario 3: Shutdown with Context Cancellation
**Input**: Context cancelled before timeout  
**Expected**: 
- Shutdown returns context error
- Resources cleaned up
**Result**: ✅ PASS

### Scenario 4: Shutdown with Missing Components
**Input**: healthChecker or auditor is nil  
**Expected**: 
- Skip nil components gracefully
- Continue with other cleanup
**Result**: ✅ PASS

## 🎓 Lessons Learned

### Graceful Shutdown Patterns

1. **Centralized Shutdown Logic**
   - Move shutdown logic vào Router struct
   - Avoid duplicate logic in signal handler
   - Easier to test và maintain

2. **Setter Pattern for Dependencies**
   - Use setter methods để inject dependencies
   - Allows flexible initialization order
   - Better testability

3. **Resource Cleanup Order**
   - Stop goroutines first
   - Close resources second
   - Reverse order of initialization

4. **Timeout Handling**
   - Use `context.WithTimeout()` cho bounded shutdown
   - Log timeout events
   - Don't block indefinitely

### Go-Specific Patterns

1. **Signal Handling**
   - `signal.Notify()` để catch signals
   - Use buffered channel cho signal delivery
   - Handle SIGTERM và SIGINT

2. **Context Cancellation**
   - Use `ctx.Done()` để detect cancellation
   - Use `select` với timeout
   - Return context error if cancelled

3. **Goroutine Lifecycle**
   - Health checker has Stop() method
   - Call Stop() before shutdown
   - Wait for goroutine to finish (if needed)

4. **Error Handling**
   - Log errors during shutdown
   - Continue cleanup even if errors occur
   - Return error if critical failure

## 🔮 Future Improvements

1. **HTTP Server Integration**
   ```go
   // Add http.Server to Router struct
   type Router struct {
       server *http.Server
       // ... other fields
   }
   
   // Call server.Shutdown() in Router.Shutdown()
   ```

2. **Database Connection Pool**
   ```go
   // Close database connections
   if r.db != nil {
       r.db.Close()
   }
   ```

3. **Wait Group for Goroutines**
   ```go
   var wg sync.WaitGroup
   wg.Add(1)
   go func() {
       defer wg.Done()
       // goroutine work
   }()
   wg.Wait()
   ```

4. **Shutdown Metrics**
   - Prometheus metrics cho shutdown duration
   - Track in-flight requests during shutdown
   - Monitor resource cleanup status

5. **Configurable Timeout**
   ```go
   type ShutdownConfig struct {
       Timeout time.Duration `yaml:"timeout"`
   }
   ```

## 📚 References

- **Go signal package**: https://pkg.go.dev/os/signal
- **Context package**: https://pkg.go.dev/context
- **http.Server Shutdown**: https://pkg.go.dev/net/http#Server.Shutdown
- **Graceful Shutdown in Go**: https://github.com/golang/go/wiki/GracefulShutdown
- **Shutdown Patterns**: https://www.ardanlabs.com/blog/graceful-shutdown-in-go

## ✅ Verification

- [x] healthChecker field added to Router struct
- [x] auditor field added to Router struct
- [x] SetHealthChecker() method implemented
- [x] SetAuditor() method implemented
- [x] Shutdown() method implemented
- [x] Shutdown() stops health checker
- [x] Shutdown() closes audit logger
- [x] Shutdown() handles timeout correctly
- [x] main.go calls SetHealthChecker()
- [x] main.go calls SetAuditor()
- [x] Signal handler calls router.Shutdown()
- [x] Duplicate healthChecker.Stop() removed
- [x] Documentation created

---

**Kết luận**: Graceful shutdown pattern giúp cleanup resources correctly và exit cleanly. Centralized shutdown logic trong Router struct giúp avoid duplicate code và improve maintainability. Context-based timeout handling ensures shutdown completes within reasonable time.
