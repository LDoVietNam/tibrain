# Ti Crew - Implementation Plan

> **Version**: 1.0.0
> **Last Updated**: 2026-05-05
> **Purpose**: Chi tiết kế hoạch implement Ti Crew (Agent orchestration framework)
> **Timeline**: 8 weeks (2 months)
> **Team Size**: 2.25 FTE

---

## 📋 Executive Summary

**Objective**: Xây dựng Ti Crew (agent orchestration framework) tương tự OpenClaw/Hermers để quản lý 104+ agents trong Ti ecosystem

**Approach**: Phased development với 4 phases
- Phase 1: Foundation (Week 1-2)
- Phase 2: Core Features (Week 3-4)
- Phase 3: Advanced Features (Week 5-6)
- Phase 4: UI & Tools (Week 7-8)

**Success Criteria**:
- ✅ Load và index 104 agents
- ✅ Route tasks đến đúng agent với 80%+ accuracy
- ✅ Execute multi-agent workflows với 5+ steps
- ✅ Track metrics cho tất cả agents
- ✅ Discover và rate agents

---

## 🗓️ Phase 1: Foundation (Week 1-2)

### Overview
Xây dựng foundation components: Agent Registry, Task Router (simple scoring), Agent Lifecycle

### Deliverables
1. Agent Registry với 104 agents loaded
2. Task Router với simple scoring logic
3. Agent Lifecycle management
4. Unit tests cho tất cả components
5. Documentation

### Tasks

#### Week 1: Agent Registry + Task Router

| Task ID | Task | Owner | Duration | Dependencies | Status |
|--------|------|-------|----------|--------------|--------|
| P1-T1 | Design Agent Registry data model | Senior Dev | 0.5 day | - | 📋 Pending |
| P1-T2 | Implement Agent Registry core (registry.go) | Senior Dev | 2 days | P1-T1 | 📋 Pending |
| P1-T3 | Implement YAML frontmatter parser | Senior Dev | 1 day | P1-T2 | 📋 Pending |
| P1-T4 | Implement agent discovery from filesystem | Senior Dev | 1 day | P1-T3 | 📋 Pending |
| P1-T5 | Load 104 agents into registry | Senior Dev | 0.5 day | P1-T4 | 📋 Pending |
| P1-T6 | Design Task Router data model | Senior Dev | 0.5 day | P1-T5 | 📋 Pending |
| P1-T7 | Implement Task Router simple scoring | Senior Dev | 2 days | P1-T6 | 📋 Pending |
| P1-T8 | Implement agent filtering logic | Senior Dev | 1 day | P1-T7 | 📋 Pending |
| P1-T9 | Write unit tests for Registry | Senior Dev | 0.5 day | P1-T5 | 📋 Pending |
| P1-T10 | Write unit tests for Router | Senior Dev | 0.5 day | P1-T8 | 📋 Pending |

#### Week 2: Agent Lifecycle + Integration

| Task ID | Task | Owner | Duration | Dependencies | Status |
|--------|------|-------|----------|--------------|--------|
| P1-T11 | Design Agent Lifecycle data model | Senior Dev | 0.5 day | - | 📋 Pending |
| P1-T12 | Implement Agent Lifecycle core (lifecycle.go) | Senior Dev | 2 days | P1-T11 | 📋 Pending |
| P1-T13 | Implement instance management | Senior Dev | 1 day | P1-T12 | 📋 Pending |
| P1-T14 | Implement scaling logic | Senior Dev | 1 day | P1-T13 | 📋 Pending |
| P1-T15 | Write unit tests for Lifecycle | Senior Dev | 0.5 day | P1-T14 | 📋 Pending |
| P1-T16 | Integrate Registry + Router + Lifecycle | Senior Dev | 1 day | P1-T10, P1-T15 | 📋 Pending |
| P1-T17 | End-to-end testing (load 104 agents, route task) | Senior Dev | 0.5 day | P1-T16 | 📋 Pending |
| P1-T18 | Write documentation (ARCHITECTURE.md) | Senior Dev | 0.5 day | P1-T17 | 📋 Pending |

### Acceptance Criteria
- ✅ 104 agents loaded vào registry
- ✅ Task router routes tasks với simple scoring
- ✅ Agent lifecycle starts/stops instances
- ✅ Unit test coverage ≥ 80%
- ✅ Documentation complete

---

## 🗓️ Phase 2: Core Features (Week 3-4)

### Overview
Xây dựng core features: Agent Orchestrator, Performance Monitor, Beads integration

### Deliverables
1. Agent Orchestrator với dependency graph
2. Performance Monitor với metrics tracking
3. Beads integration
4. Knowledge Graph Memory integration
5. Integration tests

### Tasks

#### Week 3: Agent Orchestrator

| Task ID | Task | Owner | Duration | Dependencies | Status |
|--------|------|-------|----------|--------------|--------|
| P2-T1 | Design Orchestrator workflow model | Senior Dev | 0.5 day | - | 📋 Pending |
| P2-T2 | Implement dependency graph builder | Senior Dev | 2 days | P2-T1 | 📋 Pending |
| P2-T3 | Implement topological sort | Senior Dev | 1 day | P2-T2 | 📋 Pending |
| P2-T4 | Implement sequential execution | Senior Dev | 1 day | P2-T3 | 📋 Pending |
| P2-T5 | Implement parallel execution | Senior Dev | 2 days | P2-T4 | 📋 Pending |
| P2-T6 | Implement error handling & recovery | Senior Dev | 1 day | P2-T5 | 📋 Pending |
| P2-T7 | Write unit tests for Orchestrator | Senior Dev | 1 day | P2-T6 | 📋 Pending |

#### Week 4: Performance Monitor + Integration

| Task ID | Task | Owner | Duration | Dependencies | Status |
|--------|------|-------|----------|--------------|--------|
| P2-T8 | Design Performance Monitor data model | Senior Dev | 0.5 day | - | 📋 Pending |
| P2-T9 | Implement metrics collection | Senior Dev | 1 day | P2-T8 | 📋 Pending |
| P2-T10 | Implement metrics aggregation | Senior Dev | 1 day | P2-T9 | 📋 Pending |
| P2-T11 | Implement leaderboard ranking | Senior Dev | 1 day | P2-T10 | 📋 Pending |
| P2-T12 | Write unit tests for Monitor | Senior Dev | 0.5 day | P2-T11 | 📋 Pending |
| P2-T13 | Integrate with Beads system | Senior Dev | 1 day | P2-T7, P2-T12 | 📋 Pending |
| P2-T14 | Integrate with Knowledge Graph Memory | Senior Dev | 1 day | P2-T13 | 📋 Pending |
| P2-T15 | Write integration tests | Senior Dev | 1 day | P2-T14 | 📋 Pending |
| P2-T16 | End-to-end testing (orchestrate 5-step workflow) | Senior Dev | 0.5 day | P2-T15 | 📋 Pending |

### Acceptance Criteria
- ✅ Orchestrator executes 5+ step workflows
- ✅ Performance Monitor tracks metrics cho 104 agents
- ✅ Beads integration working
- ✅ Knowledge Graph Memory integration working
- ✅ Integration test coverage ≥ 70%

---

## 🗓️ Phase 3: Advanced Features (Week 5-6)

### Overview
Xây dựng advanced features: Agent Marketplace, Agent Learning, LLM integration

### Deliverables
1. Agent Marketplace với rating/review
2. Agent Learning với pattern extraction
3. LLM integration cho routing
4. LLM integration cho learning
5. Advanced tests

### Tasks

#### Week 5: Agent Marketplace

| Task ID | Task | Owner | Duration | Dependencies | Status |
|--------|------|-------|----------|--------------|--------|
| P3-T1 | Design Marketplace data model | Senior Dev | 0.5 day | - | 📋 Pending |
| P3-T2 | Implement rating/review system | Senior Dev | 1 day | P3-T1 | 📋 Pending |
| P3-T3 | Implement search/filter | Senior Dev | 1 day | P3-T2 | 📋 Pending |
| P3-T4 | Implement trending algorithm | Senior Dev | 1 day | P3-T3 | 📋 Pending |
| P3-T5 | Write unit tests for Marketplace | Senior Dev | 0.5 day | P3-T4 | 📋 Pending |

#### Week 6: Agent Learning + LLM Integration

| Task ID | Task | Owner | Duration | Dependencies | Status |
|--------|------|-------|----------|--------------|--------|
| P3-T6 | Design Agent Learning data model | LLM Engineer | 0.5 day | - | 📋 Pending |
| P3-T7 | Implement pattern extraction with LLM | LLM Engineer | 2 days | P3-T6 | 📋 Pending |
| P3-T8 | Implement pattern database | LLM Engineer | 1 day | P3-T7 | 📋 Pending |
| P3-T9 | Implement pattern sharing | LLM Engineer | 1 day | P3-T8 | 📋 Pending |
| P3-T10 | Integrate LLM for intelligent routing | LLM Engineer | 1 day | P3-T9 | 📋 Pending |
| P3-T11 | Implement fallback logic for LLM | LLM Engineer | 0.5 day | P3-T10 | 📋 Pending |
| P3-T12 | Write unit tests for Learning | LLM Engineer | 0.5 day | P3-T11 | 📋 Pending |
| P3-T13 | Write integration tests for LLM routing | LLM Engineer | 1 day | P3-T12 | 📋 Pending |
| P3-T14 | End-to-end testing (learn from 10+ tasks) | LLM Engineer | 0.5 day | P3-T13 | 📋 Pending |

### Acceptance Criteria
- ✅ Marketplace allows rating/review agents
- ✅ Learning extracts patterns từ tasks
- ✅ LLM integration improves routing accuracy to 80%+
- ✅ Fallback logic works when LLM unavailable
- ✅ Advanced test coverage ≥ 60%

---

## 🗓️ Phase 4: UI & Tools (Week 7-8)

### Overview
Xây dựng UI và tools: CLI tool, Web UI, Monitoring dashboard

### Deliverables
1. CLI tool cho framework management
2. Web UI cho agent marketplace
3. Monitoring dashboard
4. Analytics và reporting
5. Documentation

### Tasks

#### Week 7: CLI Tool

| Task ID | Task | Owner | Duration | Dependencies | Status |
|--------|------|-------|----------|--------------|--------|
| P4-T1 | Design CLI tool architecture | Senior Dev | 0.5 day | - | 📋 Pending |
| P4-T2 | Implement registry commands (list, search, get) | Senior Dev | 1 day | P4-T1 | 📋 Pending |
| P4-T3 | Implement routing commands (route, test) | Senior Dev | 1 day | P4-T2 | 📋 Pending |
| P4-T4 | Implement orchestrator commands (run, status) | Senior Dev | 1 day | P4-T3 | 📋 Pending |
| P4-T5 | Implement lifecycle commands (start, stop, scale) | Senior Dev | 1 day | P4-T4 | 📋 Pending |
| P4-T6 | Write CLI documentation | Senior Dev | 0.5 day | P4-T5 | 📋 Pending |

#### Week 8: Web UI + Dashboard

| Task ID | Task | Owner | Duration | Dependencies | Status |
|--------|------|-------|----------|--------------|--------|
| P4-T7 | Design Web UI architecture | Frontend Dev | 0.5 day | - | 📋 Pending |
| P4-T8 | Implement marketplace UI (list, search, rate) | Frontend Dev | 2 days | P4-T7 | 📋 Pending |
| P4-T9 | Implement monitoring dashboard | Frontend Dev | 1.5 days | P4-T8 | 📋 Pending |
| P4-T10 | Implement analytics & reporting | Frontend Dev | 1 day | P4-T9 | 📋 Pending |
| P4-T11 | Write UI documentation | Frontend Dev | 0.5 day | P4-T10 | 📋 Pending |
| P4-T12 | End-to-end testing (CLI + Web) | Senior Dev | 0.5 day | P4-T11 | 📋 Pending |
| P4-T13 | Final documentation & handoff | Senior Dev | 0.5 day | P4-T12 | 📋 Pending |

### Acceptance Criteria
- ✅ CLI tool manages registry, routing, orchestrator, lifecycle
- ✅ Web UI allows marketplace discovery
- ✅ Monitoring dashboard shows real-time metrics
- ✅ Analytics & reporting working
- ✅ Documentation complete

---

## 👥 Resource Allocation

### Team Composition

| Role | FTE | Weeks | Total Effort | Phases |
|------|-----|-------|--------------|--------|
| **Senior Go Developer** | 1.0 | 8 | 8 weeks | All |
| **LLM/ML Engineer** | 0.5 | 8 | 4 weeks | Phase 3 |
| **DevOps Engineer** | 0.25 | 8 | 2 weeks | Phase 4 |
| **Frontend Developer** | 0.5 | 4 | 2 weeks | Phase 4 |
| **Total** | **2.25** | **8** | **16 weeks** | **All** |

### Weekly Resource Distribution

| Week | Senior Dev | LLM Engineer | DevOps | Frontend | Total |
|------|------------|--------------|--------|----------|-------|
| Week 1 | 1.0 | 0 | 0 | 0 | 1.0 |
| Week 2 | 1.0 | 0 | 0 | 0 | 1.0 |
| Week 3 | 1.0 | 0 | 0 | 0 | 1.0 |
| Week 4 | 1.0 | 0 | 0 | 0 | 1.0 |
| Week 5 | 0.5 | 0.5 | 0 | 0 | 1.0 |
| Week 6 | 0.5 | 0.5 | 0 | 0 | 1.0 |
| Week 7 | 1.0 | 0 | 0.25 | 0.5 | 1.75 |
| Week 8 | 1.0 | 0 | 0.25 | 0.5 | 1.75 |

---

## 🔗 Dependencies

### External Dependencies

| Dependency | Version | Purpose | Status |
|------------|---------|---------|--------|
| Go | 1.21+ | Language | ✅ Available |
| gopkg.in/yaml.v3 | 3.0.1 | YAML parsing | ✅ Available |
| PostgreSQL | 15+ | Database | ✅ Available |
| Gemini API | 2.5 Flash | LLM integration | ⚠️ Need API key |
| Beads System | - | Task management | ✅ Available |
| Knowledge Graph Memory | - | Persistent memory | ✅ Available |

### Internal Dependencies

| Component | Depends On | Phase |
|-----------|------------|-------|
| Task Router | Agent Registry | Phase 1 |
| Agent Orchestrator | Agent Registry, Task Router | Phase 2 |
| Performance Monitor | Agent Registry | Phase 2 |
| Agent Marketplace | Agent Registry | Phase 3 |
| Agent Learning | Agent Registry, Knowledge Graph | Phase 3 |
| Agent Communicator | Agent Registry, LLM | Phase 3 |
| Agent Lifecycle | Agent Registry | Phase 1 |
| CLI Tool | All components | Phase 4 |
| Web UI | All components | Phase 4 |

---

## ⚠️ Risk Management

### Risk Register

| Risk ID | Risk | Probability | Impact | Mitigation Strategy | Owner |
|---------|------|-------------|--------|---------------------|--------|
| R001 | LLM API rate limits | MEDIUM | MEDIUM | Implement caching, fallback logic, use cheaper models | LLM Engineer |
| R002 | LLM response quality | MEDIUM | HIGH | Use high-quality models, add validation, human review | LLM Engineer |
| R003 | Performance at scale (104 agents) | LOW | MEDIUM | Implement caching, indexing, load testing | Senior Dev |
| R004 | Dependency graph complexity | MEDIUM | HIGH | Use proven libraries, add tests, start simple | Senior Dev |
| R005 | State consistency in orchestrator | MEDIUM | HIGH | Use database transactions, add retries | Senior Dev |
| R006 | Scope creep | HIGH | MEDIUM | Strict MVP scope, phased approach, weekly reviews | Senior Dev |
| R007 | Timeline overruns | MEDIUM | MEDIUM | Buffer time, prioritize features | Senior Dev |
| R008 | Resource constraints | LOW | HIGH | Hire contractors, extend timeline | Senior Dev |

### Risk Response Plan

#### R001: LLM API Rate Limits
- **Prevention**: Implement request queuing, caching, rate limiting
- **Monitoring**: Track API usage, set alerts at 80% quota
- **Response**: Fallback to simple scoring, upgrade plan if needed

#### R002: LLM Response Quality
- **Prevention**: Use Gemini 2.5 Flash (high quality), add validation logic
- **Monitoring**: Track routing accuracy, sample review
- **Response**: Fine-tune prompts, switch to better model if needed

#### R003: Performance at Scale
- **Prevention**: Implement caching from start, add database indexes
- **Monitoring**: Load testing at 50, 100, 200 agents
- **Response**: Optimize queries, add sharding if needed

#### R004: Dependency Graph Complexity
- **Prevention**: Use proven library (github.com/awalterschulze/gographviz), start with simple workflows
- **Monitoring**: Code review, complexity metrics
- **Response**: Simplify workflows, add more tests

#### R005: State Consistency
- **Prevention**: Use database transactions, implement idempotency
- **Monitoring**: Log state changes, add alerts for inconsistencies
- **Response**: Add retry logic, implement state repair

#### R006: Scope Creep
- **Prevention**: Strict MVP scope, weekly reviews, prioritize features
- **Monitoring**: Track task completion vs planned
- **Response**: Defer non-critical features to later phases

#### R007: Timeline Overruns
- **Prevention**: Buffer time (20%), prioritize critical path
- **Monitoring**: Weekly progress reviews
- **Response**: Extend timeline, reduce scope

#### R008: Resource Constraints
- **Prevention**: Cross-train team, use contractors for spikes
- **Monitoring**: Track resource utilization
- **Response**: Hire contractors, extend timeline

---

## 📊 Success Metrics

### Technical Metrics

| Metric | Target | Measurement | Frequency |
|--------|--------|-------------|-----------|
| Agent Registry Load Time | < 1s | Benchmark | Weekly |
| Task Routing Latency | < 100ms | Benchmark | Weekly |
| Routing Accuracy | ≥ 80% | Sample review | Weekly |
| Orchestrator Success Rate | ≥ 90% | Success/failure count | Weekly |
| Performance Monitor Coverage | 100% | Agent count | Weekly |
| Test Coverage | ≥ 80% | Code coverage | Weekly |
| Uptime | ≥ 99% | Monitoring | Continuous |

### Business Metrics

| Metric | Target | Measurement | Frequency |
|--------|--------|-------------|-----------|
| Task Efficiency Gain | ≥ 30% | Time saved per task | Monthly |
| Agent Selection Quality | ≥ 20% | User feedback | Monthly |
| User Satisfaction | ≥ 80% | Survey | Quarterly |
| Adoption Rate | ≥ 70% | Active users | Quarterly |

### Milestones

| Milestone | Date | Deliverables | Status |
|-----------|------|--------------|--------|
| M1: Foundation Complete | Week 2 | Agent Registry, Task Router, Lifecycle | 📋 Pending |
| M2: Core Features Complete | Week 4 | Orchestrator, Performance Monitor, Beads | 📋 Pending |
| M3: Advanced Features Complete | Week 6 | Marketplace, Learning, LLM Integration | 📋 Pending |
| M4: UI & Tools Complete | Week 8 | CLI, Web UI, Dashboard | 📋 Pending |

---

## 📝 Documentation Plan

### Documentation Deliverables

| Document | Purpose | Owner | Due Date |
|----------|---------|-------|----------|
| ARCHITECTURE.md | High-level architecture | Senior Dev | Week 2 |
| API.md | API documentation | Senior Dev | Week 4 |
| CLI_GUIDE.md | CLI tool guide | Senior Dev | Week 7 |
| WEB_UI_GUIDE.md | Web UI guide | Frontend Dev | Week 8 |
| DEPLOYMENT.md | Deployment guide | DevOps | Week 8 |
| CONTRIBUTING.md | Contribution guide | Senior Dev | Week 8 |

### Code Documentation

- **GoDoc**: All public functions documented
- **Comments**: Inline comments cho complex logic
- **Examples**: Usage examples cho key functions

---

## 🚀 Deployment Plan

### Environments

| Environment | Purpose | Resources | Status |
|-------------|---------|-----------|--------|
| Development | Local development | Local machine | ✅ Ready |
| Staging | Pre-production testing | t3.medium x 2 | 📋 Pending |
| Production | Live system | t3.medium x 2 | 📋 Pending |

### Deployment Process

1. **Development → Staging** (End of each phase)
   - Run all tests
   - Deploy to staging
   - Run integration tests
   - User acceptance testing

2. **Staging → Production** (End of Phase 4)
   - Run all tests
   - Security review
   - Performance testing
   - Deploy to production
   - Monitor closely

### Rollback Plan

- **Database**: Automatic backups before deployment
- **Code**: Git tags for each release
- **Infrastructure**: Terraform state management
- **Process**: Rollback to previous version within 15 minutes

---

## 📋 Weekly Review Process

### Weekly Review Agenda

1. **Progress Review** (15 min)
   - Tasks completed this week
   - Tasks pending
   - Blockers

2. **Risk Review** (10 min)
   - New risks identified
   - Risk mitigation progress
   - Risk status updates

3. **Metrics Review** (10 min)
   - Technical metrics
   - Business metrics
   - Milestone progress

4. **Planning** (25 min)
   - Tasks for next week
   - Resource allocation
   - Dependencies

### Weekly Deliverables

- **Status Report**: Progress, risks, metrics
- **Demo**: Working features from this week
- **Documentation**: Updated docs

---

## 🎯 Phase Gates

### Gate 1: Foundation Complete (End of Week 2)

**Criteria**:
- ✅ 104 agents loaded into registry
- ✅ Task router routes tasks with simple scoring
- ✅ Agent lifecycle starts/stops instances
- ✅ Unit test coverage ≥ 80%
- ✅ Documentation complete

**Decision**: Proceed to Phase 2

### Gate 2: Core Features Complete (End of Week 4)

**Criteria**:
- ✅ Orchestrator executes 5+ step workflows
- ✅ Performance Monitor tracks metrics for 104 agents
- ✅ Beads integration working
- ✅ Knowledge Graph Memory integration working
- ✅ Integration test coverage ≥ 70%

**Decision**: Proceed to Phase 3

### Gate 3: Advanced Features Complete (End of Week 6)

**Criteria**:
- ✅ Marketplace allows rating/review agents
- ✅ Learning extracts patterns from tasks
- ✅ LLM integration improves routing accuracy to 80%+
- ✅ Fallback logic works when LLM unavailable
- ✅ Advanced test coverage ≥ 60%

**Decision**: Proceed to Phase 4

### Gate 4: UI & Tools Complete (End of Week 8)

**Criteria**:
- ✅ CLI tool manages registry, routing, orchestrator, lifecycle
- ✅ Web UI allows marketplace discovery
- ✅ Monitoring dashboard shows real-time metrics
- ✅ Analytics & reporting working
- ✅ Documentation complete

**Decision**: Go live

---

## 📞 Communication Plan

### Stakeholders

| Stakeholder | Role | Frequency | Method |
|-------------|------|-----------|--------|
| Development Team | Implementation | Daily | Standup |
| Project Manager | Oversight | Weekly | Status report |
| Product Owner | Requirements | Weekly | Demo |
| Users | Feedback | Bi-weekly | Survey |

### Communication Channels

- **Daily Standup**: Slack #ti-agent-framework
- **Weekly Review**: Zoom meeting
- **Status Report**: Email
- **Documentation**: Confluence

---

## 🎉 Success Celebration

### Recognition

- **Team Celebration**: Lunch/dinner at project completion
- **Individual Recognition**: Acknowledge contributions
- **Public Announcement**: Blog post, social media

### Lessons Learned

- **Retrospective**: End-of-project retrospective
- **Documentation**: Document lessons learned
- **Knowledge Sharing**: Share with other teams

---

## 📞 Contact

**Project Lead**: Senior Go Developer
**Questions**: dev@ti.cli
**Timeline**: 8 weeks
**Budget**: $50,000-80,000 development + $200/month infrastructure

---

## 📚 References

- **Architecture**: `TI_AGENT_FRAMEWORK_ARCHITECTURE.md`
- **Feasibility**: `TI_AGENT_FRAMEWORK_FEASIBILITY.md`
- **Existing Agents**: `content/agents/agents/`
- **Beads System**: `.devin/workflows/reflective-loop-beads.yaml`
- **Knowledge Graph Memory**: `content/mcp/KNOWLEDGE_GRAPH_USAGE_GUIDE.md`
