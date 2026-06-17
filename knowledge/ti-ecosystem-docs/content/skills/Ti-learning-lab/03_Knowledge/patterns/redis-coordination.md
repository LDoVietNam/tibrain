# Redis-Based Coordination

> **Version**: 1.0.0  
> **Last Updated**: 2026-04-28  
> **Category**: Distributed Systems  
> **Language**: Tiếng Việt

---

## 📋 Tổng Quan

Redis-Based Coordination là hệ thống coordination phân tán sử dụng Redis theo Google race-condition pattern, cung cấp distributed locking, leader election, và shared state management.

## 🎯 Mục Tiêu

1. **Distributed Locking** - Lock phân tán cho resource access
2. **Leader Election** - Leader election cho high availability
3. **Shared State** - Shared state management
4. **Pub/Sub** - Message passing giữa nodes

## 🏗️ Architecture

```
RedisCoordination
├── Distributed Locks
│   ├── Acquire
│   ├── Release
│   └── Renew
├── Leader Election
│   ├── Try Become Leader
│   ├── Heartbeat
│   └── Leader Detection
├── Shared State
│   ├── Set State
│   ├── Get State
│   └── Delete State
└── Pub/Sub
    ├── Publish
    ├── Subscribe
    └── Unsubscribe
```

## 🔒 Distributed Locking

### Acquire Lock

Sử dụng Redis SET NX để acquire lock:

```go
lock, err := rc.AcquireLock(ctx, "resource", 30*time.Second)
if err != nil {
    // Lock acquisition failed
}
```

### Release Lock

Sử dụng Lua script để release lock atomically:

```go
err := rc.ReleaseLock(ctx, "resource")
if err != nil {
    // Lock release failed
}
```

### Renew Lock

Sử dụng Lua script để renew lock atomically:

```go
err := rc.RenewLock(ctx, "resource")
if err != nil {
    // Lock renewal failed
}
```

### Lock Properties

| Property | Description |
|----------|-------------|
| Key | `ti:lock:{resource}` |
| Value | Node ID |
| TTL | Lock expiration time |
| Atomic | Lua script đảm bảo atomicity |

## 👑 Leader Election

### Try Become Leader

```go
err := rc.TryBecomeLeader(ctx)
if err != nil {
    // Leader election failed
}
```

### Heartbeat

Maintain leadership bằng cách renew leader key:

```go
func (rc *RedisCoordination) leaderHeartbeat(ctx context.Context) {
    ticker := time.NewTicker(rc.heartbeatInterval)
    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            err := rc.redisClient.Expire(ctx, rc.leaderKey, rc.heartbeatInterval*2)
            if err != nil {
                // Lost leadership
                return
            }
        }
    }
}
```

### Leader Detection

```go
leader, err := rc.GetLeader(ctx)
if err != nil {
    // Failed to get leader
}
```

### Leader Properties

| Property | Description |
|----------|-------------|
| Key | `ti:router:leader` |
| Value | Leader node ID |
| TTL | 2x heartbeat interval |
| Heartbeat | Renew key tại intervals |

## 💾 Shared State

### Set State

```go
state := map[string]interface{}{
    "key": "value",
}

err := rc.SetSharedState(ctx, "state_key", state, 1*time.Hour)
if err != nil {
    // Failed to set state
}
```

### Get State

```go
var state map[string]interface{}
err := rc.GetSharedState(ctx, "state_key", &state)
if err != nil {
    // Failed to get state
}
```

### Delete State

```go
err := rc.DeleteSharedState(ctx, "state_key")
if err != nil {
    // Failed to delete state
}
```

### State Properties

| Property | Description |
|----------|-------------|
| Key | `ti:state:{key}` |
| Value | JSON-encoded state |
| TTL | State expiration time |

## 📢 Pub/Sub

### Publish

```go
message := map[string]interface{}{
    "type": "update",
    "data": "value",
}

err := rc.Publish(ctx, "channel", message)
if err != nil {
    // Failed to publish
}
```

### Subscribe

```go
ch, err := rc.Subscribe(ctx, "channel")
if err != nil {
    // Failed to subscribe
}

// Receive messages
for msg := range ch {
    // Process message
}
```

### Unsubscribe

```go
err := rc.Unsubscribe(ctx, "channel")
if err != nil {
    // Failed to unsubscribe
}
```

### Pub/Sub Properties

| Property | Description |
|----------|-------------|
| Channel | `ti:pubsub:{channel}` |
| Message | JSON-encoded message |
| Pattern | Topic-based pub/sub |

## 🚧 Barrier

Distributed barrier để synchronize nodes:

```go
err := rc.Barrier(ctx, "barrier_name", expectedCount)
if err != nil {
    // Barrier failed
}
```

### Barrier Logic

1. Mỗi node increment counter
2. Node cuối cùng delete barrier key
4. Các node khác chờ barrier key bị delete
5. Tất cả nodes tiếp tục execution

## 🚀 Usage

### Basic Setup

```go
// Create Redis client
redisClient := NewMockRedisClient()

// Create coordination
rc := NewRedisCoordination(redisClient, "node-1", 10*time.Second)
```

### Distributed Locking

```go
// Acquire lock
lock, err := rc.AcquireLock(ctx, "resource", 30*time.Second)
if err != nil {
    // Handle error
}

// Use resource
// ...

// Release lock
err = rc.ReleaseLock(ctx, "resource")
if err != nil {
    // Handle error
}
```

### Leader Election

```go
// Try to become leader
err := rc.TryBecomeLeader(ctx)
if err != nil {
    // Handle error
}

// Check if leader
if rc.IsLeader() {
    // Leader logic
} else {
    // Follower logic
}
```

### Shared State

```go
// Set state
state := map[string]interface{}{
    "config": "value",
}
err := rc.SetSharedState(ctx, "config", state, 1*time.Hour)

// Get state
var config map[string]interface{}
err = rc.GetSharedState(ctx, "config", &config)
```

### Pub/Sub

```go
// Publish
message := map[string]interface{}{
    "type": "update",
}
err := rc.Publish(ctx, "updates", message)

// Subscribe
ch, err := rc.Subscribe(ctx, "updates")
for msg := range ch {
    // Process message
}
```

## ⚙️ Configuration

### Heartbeat Interval

```go
rc := NewRedisCoordination(client, nodeID, heartbeatInterval)
```

- `heartbeatInterval`: Interval giữa các heartbeat (default: 10s)
- Leader TTL: 2x heartbeat interval

### Lock TTL

```go
lock, err := rc.AcquireLock(ctx, key, ttl)
```

- `ttl`: Lock expiration time
- Nên set > expected operation time

### State TTL

```go
err := rc.SetSharedState(ctx, key, state, ttl)
```

- `ttl`: State expiration time
- Nên set dựa trên state freshness requirement

## 🎓 Best Practices

### Distributed Locking

1. **Set appropriate TTL** - TTL > expected operation time
2. **Renew locks periodically** - Renew cho long-running operations
3. **Handle lock acquisition failures** - Fallback hoặc retry
4. **Release locks in finally block** - Đảm bảo locks được release

### Leader Election

1. **Use heartbeat** - Heartbeat để maintain leadership
2. **Handle leadership loss** - Graceful handover
3. **Monitor leader health** - Detect leader failures
4. **Use fencing tokens** - Prevent split-brain

### Shared State

1. **Use appropriate TTL** - TTL dựa trên state freshness
2. **Version state** - Versioning để detect conflicts
3. **Use transactions** - Multi-key operations
4. **Handle state conflicts** - Conflict resolution strategies

### Pub/Sub

1. **Use structured messages** - Structured messages cho parsing
2. **Handle message loss** - Acknowledgment hoặc retry
3. **Use message IDs** - Deduplication
4. **Monitor channel health** - Detect channel failures

## 🔍 Troubleshooting

### Lock Issues

**Issue**: Lock acquisition fails
- **Solution**: Retry với exponential backoff, check lock TTL

**Issue**: Lock not released
- **Solution**: Use finally block, set reasonable TTL, monitor lock leaks

**Issue**: Lock renewal fails
- **Solution**: Check network connectivity, increase TTL, implement fallback

### Leader Election Issues

**Issue**: Leader election fails
- **Solution**: Check Redis connectivity, increase heartbeat interval, implement retry

**Issue**: Leader lost unexpectedly
- **Solution**: Increase heartbeat interval, check network stability, implement fencing

**Issue**: Split-brain scenario
- **Solution**: Use fencing tokens, quorum-based election, lease mechanism

### Shared State Issues

**Issue**: State not found
- **Solution**: Check TTL, verify key existence, implement fallback

**Issue**: State conflicts
- **Solution**: Use versioning, implement conflict resolution, use transactions

**Issue**: State corruption
- **Solution**: Validate state format, use schema validation, implement checksums

### Pub/Sub Issues

**Issue**: Messages not received
- **Solution**: Check subscription, verify channel name, monitor connection

**Issue**: Message loss
- **Solution**: Use acknowledgment, implement retry, use durable storage

**Issue**: High latency
- **Solution**: Use batching, compress messages, optimize serialization

## 📚 References

- Redis Distributed Locks: https://redis.io/docs/manual/patterns/distributed-locks/
- Leader Election: https://redis.io/docs/manual/patterns/distributed-locks/
- Pub/Sub: https://redis.io/docs/manual/pubsub/
- Google Race-Condition Pattern: https://cloud.google.com/architecture/race-conditions

---

*Last Updated: 2026-04-28*
