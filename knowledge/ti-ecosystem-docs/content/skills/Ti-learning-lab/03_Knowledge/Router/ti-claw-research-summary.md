# Ti Claw Research & Planning Summary

> **Date**: 2026-05-04
> **Status**: Research & Planning Complete ✅
> **Next**: Implementation Phase 1 (Core Framework)

---

## Executive Summary

Đã hoàn thành research và planning cho Ti Claw (rebrand từ GoClaw) - framework agent cho Ti ecosystem.

### Quyết Định Chính

1. **Framework nền tảng**: GoClaw (rebrand thành Ti Claw)
2. **Approach**: Học patterns, không áp dụng framework trực tiếp
3. **Integration**: Plugin system với Ti CLI
4. **Agent đầu tiên**: Notion Agent (sync knowledge)

---

## Research Results

### Go Agent Frameworks So Sánh

| Framework | Stars | Pros | Cons | Verdict |
|-----------|-------|------|------|---------|
| **LangChain Go** | 9.2k | LLM library phổ biến | Chỉ là LLM library, không có agent loop | ❌ Không phù hợp |
| **Blades** | 763 | Framework đầy đủ, middleware ecosystem | Quá complex, không phù hợp CLI | ❌ Không phù hợp |
| **Orloj** | 84 | Full-stack orchestration, governance | Quá nặng cho Ti use case | ❌ Không phù hợp |
| **Cordum** | 470 | Governance-focused, safety kernel | Không phải agent framework | ❌ Không phù hợp |
| **Autopus-ADK** | 97 | Multi-agent orchestration | Chưa mature | ❌ Không phù hợp |
| **GoClaw** | - | Đầy đủ patterns, đã có trong learning lab | Enterprise-grade, cần simplify | ✅ **CHỌN** |

### Tại Sao Chọn GoClaw?

1. **Đầy đủ patterns**: Agent loop, store, tools, providers, memory, config
2. **Đã có trong learning lab**: Dễ research và extract patterns
3. **Mature**: Được production-use
4. **Extensible**: Dễ customize cho Ti use case
5. **Documentation**: CLAUDE.md chi tiết

---

## Patterns Đã Extract

### 1. Core Architecture Patterns

- **Agent Loop Pattern** (Think-Act-Observe)
- **Store Layer Pattern** (Repository Pattern)
- **Tool Registry Pattern**
- **Provider Pattern** (Adapter Pattern)
- **Event Bus Pattern**

### 2. Configuration Patterns

- **Config Loading** (JSON5 + Env Overlay)
- **Agent Types** (open/predefined)
- **Context Files** (agent-level + per-user)

### 3. Scheduling Patterns

- **Lane-based Concurrency**
- **Cron Scheduling**

### 4. Memory Patterns

- **Memory System** (pgvector)
- **Session Management**
- **Auto-summarization**

### 5. Security Patterns

- **Input Guard**
- **RBAC** (Role-Based Access Control)

### 6. Integration Patterns

- **MCP** (Model Context Protocol)
- **Channel Integration**

### 7. Bootstrap Patterns

- **Context File Seeding**
- **Skill Loading** (BM25 search)

### 8. Internationalization (i18n)

- **Backend**: Message catalog
- **Web UI**: i18next
- **Supported**: en, vi, zh

### 9. Desktop Edition (Lite) Patterns

- **Build tag**: `//go:build sqliteonly`
- **Edition system**: Standard vs Lite
- **Secrets**: OS keyring

### 10. Mobile UI/UX Patterns

- **Viewport height**: `h-dvh`
- **Input font-size**: 16px on mobile
- **Safe areas**: Notched devices
- **Touch targets**: ≥44px

### 11. Migration Patterns

- **Migration files**: `migrations/`
- **Version tracking**: `internal/upgrade/version.go`

### 12. Testing Patterns

- **Integration tests**: `tests/integration/`
- **Race detector**: `go test -race`
- **Post-implementation checklist**

### 13. Go Conventions

- `errors.Is(err, sentinel)` instead of `err == sentinel`
- `switch/case` instead of `if/else if`
- `append(dst, src...)` instead of loop-based append
- Always handle errors

---

## Components Mapping

### Keep (Giữ nguyên)

1. **Agent Loop** - Think-Act-Observe pattern
2. **Store Layer** - Interface-based với SQLite
3. **Tool Registry** - Tool management
4. **Provider Pattern** - LLM provider abstraction
5. **Event Bus** - Event system
6. **Session Management** - History + summary
7. **Memory System** - SQLite-based (simplified)
8. **Config Loading** - JSON5 + env overlay
9. **Bootstrap Pattern** - Context file seeding
10. **Skill Loading** - BM25 search

### Remove (Bỏ)

1. **Multi-tenant** - Ti là single-user CLI
2. **PostgreSQL** - Dùng SQLite only
3. **Web UI** - Ti là CLI-first
4. **Desktop UI** - Không cần cho CLI
5. **Channels** - Không cần cho CLI
6. **RBAC** - Single-user, không permissions
7. **WebSocket** - CLI dùng stdin/stdout
8. **Cron Scheduling** - Không cần cho CLI
9. **Lane-based Scheduler** - Simplify thành single-threaded
10. **OAuth** - Không cần cho CLI

### Tweak (Tùy chỉnh)

1. **Agent Types** - Chỉ giữ "predefined" (Ti có fixed agent definitions)
2. **Context Files** - Adapt cho Ti's content/ structure
3. **Skills** - Integrate với Ti's .devin/skills/
4. **Memory** - Dùng Ti's Knowledge Graph Memory (server-memory)
5. **Providers** - Giữ Anthropic, OpenAI, thêm Ti Router
6. **Tools** - Adapt cho Ti's MCP servers
7. **i18n** - Giữ Vietnamese (Ti là VN-focused)
8. **Security** - Simplify (không rate limiting, CORS, SSRF)

### New (Mới)

1. **Ti CLI Integration** - Plugin system cho apps/cli/
2. **Ti Router Provider** - Dùng Ti Router như LLM provider
3. **Knowledge Graph Memory** - Integrate với server-memory MCP
4. **Notion Agent** - Agent đầu tiên cho knowledge sync
5. **Ti Skill System** - Integrate với .devin/skills/
6. **OmniRoute Patterns** - Integrate routing patterns
7. **Ti Config** - Dùng Ti's config system

---

## Integration Plan

### Phase 1: Core Framework (Week 1-2)

**Goal**: Implement core Ti Claw patterns

**Tasks**:
1. Create `apps/ticlaw/` repository
2. Implement Agent Loop
3. Implement Store Layer (SQLite)
4. Implement Tool Registry
5. Implement Provider Pattern
6. Implement Memory System
7. Implement Config Loading
8. Implement Bootstrap Pattern

**Deliverables**:
- Core Ti Claw framework
- Unit tests
- Documentation

---

### Phase 2: CLI Integration (Week 3)

**Goal**: Integrate Ti Claw với Ti CLI

**Tasks**:
1. Add Ti Claw Commands to Ti CLI
2. Implement Plugin System
3. Integrate with Ti Router
4. Integrate with Knowledge Graph Memory
5. Integrate with Ti Skills

**Deliverables**:
- Ti CLI với Ti Claw integration
- Plugin system working
- Integration tests

---

### Phase 3: Notion Agent (Week 4)

**Goal**: Build first agent - Notion Agent

**Tasks**:
1. Create Notion Agent Definition
2. Implement Notion MCP Integration
3. Build Notion Agent
4. Test Notion Agent

**Deliverables**:
- Notion Agent working
- Knowledge sync automated
- Documentation

---

### Phase 4: Additional Agents (Week 5-6)

**Goal**: Build additional agents

**Tasks**:
1. Knowledge Sync Agent
2. Code Review Agent
3. Test Agent
4. Documentation Agent

**Deliverables**:
- 4+ agents working
- Agent templates
- Best practices

---

### Phase 5: Polish & Release (Week 7-8)

**Goal**: Polish và release

**Tasks**:
1. Bug fixes
2. Performance optimization
3. Final documentation
4. Release v1.0.0

**Deliverables**:
- Production-ready Ti Claw
- Complete documentation
- Release notes

---

## Files Created

1. **`goclaw-patterns-extraction.md`** - Patterns extraction từ GoClaw
   - Location: `Ti-learning-lab/03_Knowledge/Router/goclaw-patterns-extraction.md`
   - Content: 16 sections, 200+ lines
   - Status: Complete ✅

2. **`ti-claw-integration-plan.md`** - Integration plan với Ti CLI
   - Location: `Ti-learning-lab/03_Knowledge/Router/ti-claw-integration-plan.md`
   - Content: 13 sections, 500+ lines
   - Status: Complete ✅

3. **`ti-claw-research-summary.md`** - Research summary (file này)
   - Location: `Ti-learning-lab/03_Knowledge/Router/ti-claw-research-summary.md`
   - Content: Executive summary
   - Status: Complete ✅

---

## Next Steps

### Immediate (Next Session)

1. **Sync knowledge to Notion**
   - Khi Notion Agent sẵn sàng (Phase 3)
   - Sync patterns extraction
   - Sync integration plan
   - Sync research summary

2. **Start Phase 1 Implementation**
   - Create `apps/ticlaw/` repository
   - Implement core patterns
   - Unit tests

### Short-term (Week 1-2)

1. **Implement Phase 1** - Core Framework
2. **Unit tests** - Core components
3. **Documentation** - Architecture guide

### Medium-term (Week 3-4)

1. **Implement Phase 2** - CLI Integration
2. **Implement Phase 3** - Notion Agent
3. **Integration tests** - Full flow

### Long-term (Week 5-8)

1. **Implement Phase 4** - Additional Agents
2. **Implement Phase 5** - Polish & Release
3. **Documentation** - Complete docs

---

## Success Criteria

- [x] Research completed
- [x] Framework selected (GoClaw → Ti Claw)
- [x] Patterns extracted
- [x] Integration plan created
- [ ] Core framework implemented
- [ ] CLI integration working
- [ ] Notion Agent syncing knowledge
- [ ] 4+ agents working
- [ ] Documentation complete
- [ ] Tests passing
- [ ] Performance acceptable
- [ ] Release v1.0.0

---

## Risks & Mitigations

### Risk 1: Complexity Overwhelm
**Mitigation**: Start với core patterns only, add features incrementally

### Risk 2: Integration Issues
**Mitigation**: Extensive testing, rollback plan, feature flags

### Risk 3: Performance Issues
**Mitigation**: Profiling, optimization, caching

### Risk 4: Documentation Gaps
**Mitigation**: Document as we go, code comments, examples

---

## References

- **GoClaw Source**: `Ti-learning-lab/01_Learning/lab/07_Repositories/goclaw-main/`
- **GoClaw CLAUDE.md**: `Ti-learning-lab/01_Learning/lab/07_Repositories/goclaw-main/CLAUDE.md`
- **Ti CLI**: `apps/cli/`
- **Ti Router**: `apps/router/`
- **Ti Skills**: `.devin/skills/`
- **Ti Agents**: `content/agents/`
- **Knowledge Graph Memory**: `content/mcp/KNOWLEDGE_GRAPH_USAGE_GUIDE.md`
- **Patterns Extraction**: `Ti-learning-lab/03_Knowledge/Router/goclaw-patterns-extraction.md`
- **Integration Plan**: `Ti-learning-lab/03_Knowledge/Router/ti-claw-integration-plan.md`

---

## Conclusion

Research và planning đã hoàn thành. Ti Claw (rebrand từ GoClaw) là framework phù hợp nhất cho Ti ecosystem. Plan đã được thiết kế chi tiết với 5 phases trong 8 weeks.

**Status**: Ready for Implementation ✅
**Next**: Start Phase 1 (Core Framework)
