# Long-Term Memory - Ti-Learning-Lab

> **Long-term memory và kiến thức tích lũy**

---

## 📚 Knowledge Base

### Research Completed:

#### 1. 9Router Deep Dive (2026-04-27)
**Status**: ✅ Completed
**Files**: `../03_Knowledge/research/9router/`
**Key Findings**:
- Multi-account fallback system
- Model-level locking
- Exponential backoff with config-driven rules
- OAuth token refresh flows
- Combo system comparison (Ti Router đã tốt hơn 9Router)

**Lessons Applied**:
- Added Status() method to 16 providers in Ti Router
- Fixed authentication checks in Ti Router
- Added database integration for usage logging
- Added metrics export with Prometheus endpoint

**Impact**: Ti Router now has better status tracking and observability

---

#### 2. RTK Research (2026-04-28)
**Status**: 🔄 In Progress
**Files**: `../05_Repositories/rtk/`
**Key Findings**:
- Command proxy architecture
- Filter pipeline (smart, grouping, truncation, deduplication)
- SQLite tracking for analytics
- Auto-rewrite hook system
- 60-90% token reduction

**Lessons to Apply**:
- RTK wrapper for Ti Router agent tools
- Filter pipeline for provider responses
- SQLite tracking for usage metrics

**Impact Expected**: 60-90% token reduction for tool-heavy workflows

---

### Patterns Learned:

#### 1. Command Proxy Pattern
**Source**: RTK
**Learned**: 2026-04-28
**Applied**: Not yet
**Status**: 📋 Planned for Ti Router

#### 2. Filter Pipeline Pattern
**Source**: RTK
**Learned**: 2026-04-28
**Applied**: Not yet
**Status**: 📋 Planned for Ti Router

#### 3. Multi-Account Fallback Pattern
**Source**: 9Router
**Learned**: 2026-04-27
**Applied**: Not yet
**Status**: 📋 Planned for Ti Router

#### 4. Model-Level Locking Pattern
**Source**: 9Router
**Learned**: 2026-04-27
**Applied**: Not yet
**Status**: 📋 Planned for Ti Router

---

## 🎯 Projects Tracked

### Ti Router
**Location**: `Z:\Ti\router\`
**Status**: Active development
**Recent Changes**:
- Added authentication checks to admin APIs
- Added database integration for usage logging
- Added metrics export with Prometheus endpoint
- Fixed Status() method for 16 providers
- Fixed config builder (LoadFromEnv)

**Next Steps**:
- Apply RTK integration
- Apply multi-account fallback
- Apply model-level locking
- Consider AI agent router transformation

---

### AI Agent Router Plan
**Location**: `Z:\Ti\router\docs\AI_AGENT_ROUTER_PLAN.md`
**Status**: Plan completed
**Next Steps**:
- Phase 0: Architecture & Foundation
- Phase 1: Quick Wins (RTK, Context-Aware, Semantic Caching, Explainable)
- Phase 2: Medium Effort (Predictive Health, Adaptive Rate Limit, Error Recovery)
- Phase 3: Advanced (Dynamic Cost, Multi-Objective, A/B Testing, Collaborative)

---

## 📊 Metrics Tracked

### Token Savings (RTK):
- Target: 60-90%
- Status: Not yet measured
- Plan: Measure after integration

### Cost Reduction (Ti Router):
- Target: 30-50% (routing) + 60-90% (RTK) = 70-95%
- Status: Not yet measured
- Plan: Measure after implementations

### Downtime Reduction:
- Target: 80-90%
- Status: Not yet measured
- Plan: Measure after predictive health implementation

---

## 🔗 External References

### Repositories:
- RTK: https://github.com/rtk-ai/rtk
- 9Router: (research docs in `../03_Knowledge/research/9router/`)

### Documentation:
- AI_AGENT_ROUTER_PLAN.md: `Z:\Ti\router\docs\AI_AGENT_ROUTER_PLAN.md`
- AGENTS.md (Ti Router): `Z:\Ti\router\AGENTS.md`
- CLAUDE.md (RTK): `../05_Repositories/rtk/CLAUDE.md`

---

## 📝 Guidelines Update History

### 2026-04-28:
- Created `.devin/` folder structure
- Created project context
- Created repo index
- Created patterns documentation
- Created lessons learned
- Created session memory
- Created long-term memory
- Updated AGENTS.md with 05_Repositories section

---

## 🎯 Future Plans

### Q2 2026:
- Complete RTK integration
- Apply multi-account fallback to Ti Router
- Apply model-level locking to Ti Router
- Start AI agent router transformation

### Q3 2026:
- Complete AI agent router transformation
- Apply semantic caching
- Apply swarm intelligence
- Complete A/B testing system

---

**Last Updated**: 2026-04-28
