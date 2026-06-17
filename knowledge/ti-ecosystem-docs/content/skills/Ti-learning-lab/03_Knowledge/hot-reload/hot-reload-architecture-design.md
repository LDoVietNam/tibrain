---
tags: ["tibrain", "documentation", "skill", "http", "api"]
scopes: ["integration", "tibrain"]
last_updated: 2026-05-22
---
# Hot Reload Architecture Design - Ti Router

> **Ngày tạo**: 2026-04-29  
> **Ngôn ngữ**: Tiếng Việt  
> **Mục đích**: Design architecture cho hot reload trong Ti Router

## Research Summary

### GitHub Analysis

#### 1. edwingeng/hotswap (426 stars)
- **Approach**: Plugin-based hot reload
- **Features**:
  - Reload code mà không restart server
  - Live functions, live types, live data
  - Plugin isolation
  - Export/Import giữa plugins
- **Limitations**:
  - Không support Windows plugin reloading
  - Cần CGO_ENABLED=1
  - Types trong different versions không giống nhau
- **Use case**: Dynamic code reloading cho plugins

#### 2. jpillora/overseer (2.4k stars)
- **Approach**: Graceful restart & self-upgrade
- **Features**:
  - Zero-downtime restarts
  - Socket file exchange
  - Self-upgrading binaries
  - Works với process managers (systemd, supervisor)
  - Signal forwarding
  - Multiple fetchers (HTTP, S3, Github, File)
- **How it works**:
  - Main process checks for upgrades
  - Child process runs actual program
  - Main process passes listener files to child
  - All signals forwarded from main to child
  - Fetcher checks for updates in goroutine
- **Limitations**:
  - Config cannot be changed via upgrade
  - Addresses can only be changed by restarting main process
  - Uses `mv` command for file moves
  - `init()` functions run twice
- **Use case**: Production-grade graceful restart

#### 3. cloudflare/tableflip (3.2k stars)
- **Approach**: Graceful process restarts
- **Features**:
  - Socket passing
  - Zero-downtime
  - Simple API
- **Status**: Actively maintained (updated 5 days ago)

#### 4. facebookarchive/grace (4.9k stars)
- **Approach**: Graceful restart & zero downtime
- **Status**: Archived

#### 5. Production Examples
- **Caddy**: Uses systemd service với graceful reload
  - `systemctl reload caddy`
  - Config reload without restart
  - Zero-downtime deployment

## Decision Matrix

| Approach | Pros | Cons | Ti Router Fit |
|----------|------|------|---------------|
| **hotswap (plugin-based)** | True hot reload, no restart | Windows limitation, complex | Low - Windows env |
| **overseer (graceful restart)** | Zero-downtime, production-ready, Windows support | Still restarts (but graceful) | **High** |
| **tableflip (socket passing)** | Simple, zero-downtime | Less features | Medium |
| **Config watcher only** | Simple, already implemented | Only config, not code | Medium (current) |

## Recommended Architecture: Hybrid Approach

```
┌─────────────────────────────────────────────────────────────────┐
│                        Ti Router                                │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  ┌───────────────────────────────────────────────────────────┐  │
│  │  Overseer (Main Process)                                │  │
│  │  - Check for code changes                               │  │
│  │  - Trigger build on .go file changes                   │  │
│  │  - Verify binary                                        │  │
│  │  - Graceful restart child process                       │  │
│  │  - Pass listener sockets to child                       │  │
│  │  - Forward signals                                      │  │
│  └───────────────────────────────────────────────────────────┘  │
│                              │                                 │
│                              ▼                                 │
│  ┌───────────────────────────────────────────────────────────┐  │
│  │  Router (Child Process)                                 │  │
│  │  - Actual router logic                                  │  │
│  │  - HTTP server with listener from overseer              │  │
│  │  - Provider registry                                    │  │
│  │  - Config watcher (reload providers.yaml)              │  │
│  │  - All business logic                                   │  │
│  └───────────────────────────────────────────────────────────┘  │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

## Implementation Plan

### Phase 1: Graceful Restart với overseer

**Goal**: Enable zero-downtime binary reload

**Steps**:
1. Add overseer dependency
   ```go
   go get github.com/jpillora/overseer
   ```

2. Refactor main() in cmd/routerd/main.go
   ```go
   func main() {
       overseer.Run(overseer.Config{
           Program: routerProg,
           Address: ":1807",
           // Optional: Add fetcher for auto-upgrade
           // Fetcher: &fetcher.File{...},
       })
   }
   
   func routerProg(state overseer.State) {
       // Move current router logic here
       // Use state.Listener instead of net.Listen
       router := &Router{}
       router.Init("configs/Tiserverrouter.yaml")
       router.Serve(state.Listener)
   }
   ```

3. Update Router struct
   - Add Serve(listener net.Listener) method
   - Remove Listen() from existing code

4. Test graceful restart
   - Send SIGUSR2 to trigger restart
   - Verify zero-downtime (no connection drops)

### Phase 2: Auto-Build & Restart

**Goal**: Watch code changes → Build → Restart automatically

**Steps**:
1. Create code watcher in cmd/routerd/code_watcher.go
   - Watch .go files with fsnotify
   - On change, trigger build
   - On build success, trigger overseer restart

2. Add build trigger
   ```go
   func triggerBuild() error {
       cmd := exec.Command("go", "build", "-o", "bin/routerd.exe", "./cmd/routerd")
       return cmd.Run()
   }
   ```

3. Add restart trigger
   ```go
   func triggerRestart() error {
       // Send SIGUSR2 to main process
       // Or use overseer's internal restart mechanism
   }
   ```

4. Integrate với overseer
   - Add custom fetcher or use overseer.File
   - Or implement custom upgrade logic

### Phase 3: Windows Compatibility

**Goal**: Ensure works on MINGW64/Windows

**Steps**:
1. Test overseer on Windows
   - Socket passing works on Windows
   - Signal handling may differ

2. Adjust for Windows
   - Use appropriate signal (SIGUSR2 may not work)
   - May need to use HTTP endpoint for restart trigger

3. Alternative for Windows
   - Add HTTP endpoint `/api/restart`
   - Call endpoint to trigger restart
   - Fallback if signals don't work

### Phase 4: Integration với Config Watcher

**Goal**: Unified hot reload system

**Current State**:
- Config watcher watches providers.yaml
- Reloads config without restart

**Integration**:
- Keep config watcher as-is (works well)
- Add code watcher for .go files
- Both work independently:
  - Config change → Config reload (fast)
  - Code change → Build → Graceful restart (slower but zero-downtime)

## Architecture Diagram

```
┌─────────────────────────────────────────────────────────────────┐
│                     Development Workflow                         │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  1. Edit code (.go files)                                       │
│     │                                                           │
│     ▼                                                           │
│  2. Code Watcher detects change (fsnotify)                      │
│     │                                                           │
│     ▼                                                           │
│  3. Trigger build (go build)                                   │
│     │                                                           │
│     ▼                                                           │
│  4. Build success?                                              │
│     ├─ No → Log error, continue                                │
│     └─ Yes → Continue                                          │
│     │                                                           │
│     ▼                                                           │
│  5. Trigger overseer restart (SIGUSR2 or HTTP)                  │
│     │                                                           │
│     ▼                                                           │
│  6. Overseer graceful restart                                   │
│     │                                                           │
│     ▼                                                           │
│  7. New binary running with zero downtime                        │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

## Config Changes Workflow

```
┌─────────────────────────────────────────────────────────────────┐
│                   Config Change Workflow                         │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  1. Edit config (providers.yaml)                                │
│     │                                                           │
│     ▼                                                           │
│  2. Config Watcher detects change (fsnotify)                    │
│     │                                                           │
│     ▼                                                           │
│  3. Debounce (2s)                                               │
│     │                                                           │
│     ▼                                                           │
│  4. Trigger config reload callback                               │
│     │                                                           │
│     ▼                                                           │
│  5. Reload providers.yaml                                       │
│  6. Rebuild provider registry                                   │
│  7. Update routing router                                       │
│  8. Update rate limiter                                         │
│  9. Update load balancer                                        │
│     │                                                           │
│     ▼                                                           │
│  10. Config reloaded (no restart needed)                        │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

## Comparison: Code vs Config Reload

| Aspect | Code Reload | Config Reload |
|--------|-------------|---------------|
| **Trigger** | .go file change | providers.yaml change |
| **Action** | Build → Graceful restart | Reload config callback |
| **Downtime** | Zero (graceful) | Zero (no restart) |
| **Speed** | Slower (build time) | Faster (instant) |
| **Scope** | Entire binary | Config only |
| **Implementation** | overseer | fsnotify (already done) |

## Trade-offs

### Pros
- **Zero-downtime**: No connection drops during restart
- **Production-ready**: overseer is battle-tested
- **Windows support**: Works on Windows/MINGW64
- **Simple**: Minimal code changes
- **Flexible**: Can add auto-upgrade later

### Cons
- **Still restarts**: Binary restarts (though graceful)
- **Build time**: Must wait for build on code change
- **Complexity**: Adds overseer dependency
- **Signal handling**: May need Windows-specific handling

## Future Enhancements

### 1. Plugin-Based Hot Reload (Optional)
- Use edwingeng/hotswap for providers
- Enable true code reloading without restart
- Only for Linux/macOS (Windows limitation)

### 2. Auto-Upgrade
- Add overseer fetcher (HTTP/S3/Github)
- Auto-download new binary from remote
- CI/CD integration

### 3. Hot Reload for Specific Components
- Reload HTTP handlers without restart
- Reload middleware without restart
- Reload routing rules without restart

### 4. Metrics & Monitoring
- Track restart events
- Monitor build times
- Alert on frequent restarts

## Implementation Complexity

| Phase | Complexity | Time Estimate |
|-------|------------|---------------|
| Phase 1: Graceful Restart | Medium | 2-3 hours |
| Phase 2: Auto-Build | Medium | 2-3 hours |
| Phase 3: Windows Compatibility | Low-Medium | 1-2 hours |
| Phase 4: Integration | Low | 1 hour |
| **Total** | **Medium** | **6-9 hours** |

## Success Criteria

- [ ] Overseer integrated successfully
- [ ] Graceful restart works (zero downtime)
- [ ] Code watcher triggers build
- [ ] Build triggers restart
- [ ] Works on Windows/MINGW64
- [ ] Config watcher still works
- [ ] No connection drops during restart
- [ ] Logs show restart events
- [ ] Documentation complete (tiếng Việt)

## References

- [jpillora/overseer](https://github.com/jpillora/overseer) - 2.4k stars
- [edwingeng/hotswap](https://github.com/edwingeng/hotswap) - 426 stars
- [cloudflare/tableflip](https://github.com/cloudflare/tableflip) - 3.2k stars
- [Caddy Running Guide](https://caddyserver.com/docs/running)
