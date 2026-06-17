# Design Patterns - Ti-Learning-Lab

> **Design patterns đã học từ các repositories để áp dụng vào projects**

---

## 📚 Patterns từ RTK

### 1. Command Proxy Pattern
**Source**: RTK (src/main.rs, src/cmds/*)

**Mô tả**: Intercept command calls, process/filter output, return optimized result

**Use Case**: 
- Giảm token consumption cho LLM
- Filter verbose output
- Compress responses

**Implementation**:
```rust
// RTK example
#[derive(Subcommand)]
enum Commands {
    Ls { args: Vec<String> },
    Git { args: Vec<String> },
    // ...
}

// Route to specialized filter
match command {
    Commands::Ls { args } => ls::ls_cmd(args),
    Commands::Git { args } => git::git_cmd(args),
}
```

**Áp dụng cho Ti Router**:
- Intercept tool calls từ agent
- Filter output trước khi gửi cho agent
- Reduce tokens = reduce cost

---

### 2. Filter Pipeline Pattern
**Source**: RTK (src/filters/*)

**Mô tả**: Chuỗi các filters được áp dụng tuần tự để optimize output

**Strategies**:
1. **Smart Filtering** - Remove noise (comments, whitespace, boilerplate)
2. **Grouping** - Aggregate similar items
3. **Truncation** - Keep relevant context, cut redundancy
4. **Deduplication** - Collapse repeated lines with counts

**Implementation**:
```rust
// RTK filter pipeline
pub fn apply_filters(input: &str, level: FilterLevel) -> String {
    let mut output = input.to_string();
    output = smart_filter(&output);
    output = group_items(&output);
    output = truncate(&output, level);
    output = deduplicate(&output);
    output
}
```

**Áp dụng cho Ti Router**:
- Filter provider responses
- Filter agent tool outputs
- Filter error messages

---

### 3. SQLite Tracking Pattern
**Source**: RTK (src/core/tracking.rs)

**Mô tả**: Track usage metrics với SQLite để analytics và gain tracking

**Use Case**:
- Track token savings
- Analytics cho command usage
- Gain reporting

**Implementation**:
```rust
// RTK tracking
fn track_command(cmd: &str, input_tokens: usize, output_tokens: usize) {
    db.execute(
        "INSERT INTO commands (cmd, input_tokens, output_tokens) VALUES (?, ?, ?)",
        cmd, input_tokens, output_tokens
    );
}
```

**Áp dụng cho Ti Router**:
- Track provider usage
- Track token consumption
- Analytics cho routing decisions

---

### 4. Fallback Pattern
**Source**: RTK (core logic)

**Mô tả**: Nếu filter fails, execute raw command unchanged

**Use Case**:
- Guaranteed compatibility
- Graceful degradation
- Error recovery

**Implementation**:
```rust
// RTK fallback
fn execute_with_fallback(cmd: &str) -> Result<String> {
    match apply_filters(cmd) {
        Ok(filtered) => Ok(filtered),
        Err(_) => execute_raw(cmd),  // Fallback
    }
}
```

**Áp dụng cho Ti Router**:
- Fallback khi filter fails
- Fallback khi provider fails
- Fallback when agent fails

---

## 📚 Patterns từ 9Router

### 5. Multi-Account Fallback Pattern
**Source**: 9Router research (01-account-selection-fallback.md)

**Mô tả**: Multiple accounts per provider với selection strategies

**Strategies**:
- **fill-first**: Tiêu thụ quota account 1 trước
- **round-robin**: Phân bổ đều
- **sticky round-robin**: Dùng N lần rồi đổi

**Implementation**:
```javascript
// 9Router example
function getProviderCredentials(provider, excludeIds, model) {
    const available = connections.filter(c => 
        !excludeIds.has(c.id) && 
        !isModelLockActive(c, model)
    );
    
    // Selection strategy
    if (strategy === "round-robin") {
        return selectRoundRobin(available);
    } else {
        return available[0]; // fill-first
    }
}
```

**Áp dụng cho Ti Router**:
- Multi-account per provider
- Account selection strategies
- Model-level locking

---

### 6. Model-Level Locking Pattern
**Source**: 9Router research

**Mô tả**: Lock từng model riêng biệt thay vì lock toàn provider

**Use Case**:
- Account bị rate limit trên model expensive nhưng vẫn dùng model cheap
- Granular control

**Implementation**:
```javascript
// 9Router example
connection.modelLock_claude-sonnet = "2026-04-28T03:00:00Z"
connection.modelLock___all = "2026-04-28T03:00:00Z"  // account-level lock
```

**Áp dụng cho Ti Router**:
- Model-level circuit breaker
- Granular rate limiting
- Better resource utilization

---

### 7. Exponential Backoff Pattern
**Source**: 9Router research

**Mô tả**: Cooldown tăng gấp đôi sau mỗi failure

**Implementation**:
```javascript
// 9Router example
const cooldown = BACKOFF_CONFIG.base * Math.pow(2, backoffLevel);
return Math.min(cooldown, BACKOFF_CONFIG.max);  // max 4 min
```

**Áp dụng cho Ti Router**:
- Ti đã có ExponentialBackoff trong resilience layer
- Có thể học từ 9Router về config-driven error rules

---

## 📝 Quy Tắc Thêm Pattern Mới

### Khi học được pattern mới:

1. **Document pattern**:
   ```
   - Pattern name
   - Source (repo/file)
   - Mô tả
   - Use case
   - Implementation example
   ```

2. **Áp dụng cho Ti**:
   ```
   - Có thể áp dụng không?
   - Cần modify gì?
   - Có pattern tốt hơn không?
   ```

3. **Validate**:
   ```
   - Test pattern
   - Measure impact
   - Document lessons
   ```

---

## 🎯 Patterns Áp Dụng Cho Ti Router

### Priority 1 (High Value):
1. **Command Proxy Pattern** - Từ RTK
2. **Filter Pipeline Pattern** - Từ RTK
3. **Multi-Account Fallback** - Từ 9Router
4. **Model-Level Locking** - Từ 9Router

### Priority 2 (Medium Value):
5. **SQLite Tracking Pattern** - Từ RTK
6. **Fallback Pattern** - Từ RTK
7. **Exponential Backoff** - Từ 9Router (Ti đã có, có thể improve)

---

**Last Updated**: 2026-04-28
