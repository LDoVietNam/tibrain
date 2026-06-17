# OAuth Session Pool & MCP Tool Routing

> **Folder**: OAuth Session Pool & MCP Tool Routing Knowledge  
> **Date**: 2026-04-28  
> **Purpose**: Tổng hợp kiến thức về OAuth session pool, usage tracking, MCP integration cho Ti Router

---

## 📁 Files

| File | Description | Status |
|------|-------------|--------|
| `analysis.md` | Comprehensive analysis với gap details, implementation plan, cost comparison | ✅ Complete |

---

## 📚 Related Knowledge (Ti-learning-lab)

### 9Router (JavaScript Reference)
- `../9router/01-account-selection-fallback.md` - Account selection & fallback
- `../9router/03-oauth-token-refresh.md` - OAuth token refresh

### FreeLLMAPI (TypeScript Reference)
- `../freellmapi/per-key-rate-tracking.md` - Per-key rate tracking (đã implement trong Ti Router)
- `../freellmapi/sticky-sessions.md` - Sticky sessions (đã implement trong Ti Router)
- `../freellmapi/enhanced-analytics.md` - Enhanced analytics (đã implement core, cần DB migration)

---

## 🎯 Implementation Plan

### Phase 1: OAuth Session Pool (Priority 1)
1. Extend OAuthCredential struct với usage tracking fields
2. Create SessionPool manager với selection logic
3. Implement model-level locking
4. Implement exponential backoff
5. Integration với router

### Phase 2: MCP Tool Routing (Priority 2)
1. MCP client management
2. MCP tool routing
3. Integration với OAuth sessions

---

## 💡 Cost Analysis

| Option | Description | Cost (Year 1) | Complexity |
|--------|-------------|---------------|------------|
| **Option 1** | Add to Ti Router | $584-880 | Medium ✅ Recommended |
| **Option 2** | Replace Router | $1,168-1,760 | High |
| **Option 3** | Separate Service | $880-1,320 | High |
| **Option 4** | Plugin System | $440-660 | Medium-High |

---

## 📝 References

### Ti Router Implementation
- `Z:\Ti\router\layers\authentication\` - OAuth infrastructure
- `Z:\Ti\router\layers\authentication\rate_tracker.go` - Per-key rate tracking
- `Z:\Ti\router\layers\authentication\sticky_session.go` - Sticky sessions
- `Z:\Ti\router\layers\authentication\analytics.go` - Enhanced analytics
- `Z:\Ti\router\layers\http\mcp\mcp.go` - MCP handler (placeholder)

---

**Last Updated**: 2026-04-28  
**Agent**: Claude Code
