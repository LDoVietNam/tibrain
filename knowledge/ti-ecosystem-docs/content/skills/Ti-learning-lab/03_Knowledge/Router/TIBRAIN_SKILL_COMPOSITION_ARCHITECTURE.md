# TiBrain Skill Composition System Architecture

> **Version**: 1.0.0
> **Date**: 2026-04-29
> **Status**: Design Phase

## Overview

TiBrain Skill Composition System là một kiến trúc để kết nối các skills với nhau, cho phép:
1. **Skill Recommendation** - Khi dùng skill A, tự động recommend skill B liên quan
2. **Skill Dependency Graph** - Build dependency graph giữa skills để agent hiểu khi nào cần skill nào
3. **Skill Composition Engine** - Tự động kết hợp multiple skills để tạo workflow mạnh hơn (A + B = x10)

## Problem Statement

**Hiện trạng:**
- Agents biết skill A nhưng không biết skill B
- Dù kết hợp A + B thì chất lượng tăng 10x
- Không có hệ thống để discover skill relationships
- Skills hoạt động độc lập, không có orchestration

**Ví dụ:**
- Agent biết "python-testing" skill nhưng không biết "django-tdd" skill
- Khi làm Django project, nên dùng cả 2 skills để có TDD workflow tối ưu cho Django
- Nhưng agent chỉ dùng python-testing generic, không tận dụng Django-specific patterns

## Architecture Components

### 1. Skill Metadata Enhancement

**Current TiBrain Schema (26 columns):**
```sql
-- Core (11)
id, name, description, parameters, handler, category, permissions, enabled
quality_score, security_score, created_at, updated_at

-- Metadata (15)
best_practices_score, source, family, tags, version
skill_level, quality_tier, security_tier
security_status, validation_status
variant_id, variant_label, source_type, root_path
```

**New Columns for Composition:**
```sql
-- Relationships (5)
dependencies TEXT        -- JSON array of skill IDs this skill depends on
related_skills TEXT      -- JSON array of skill IDs related to this skill
composition_score REAL   -- Score 0-1: how well this skill composes with others
usage_patterns TEXT      -- JSON: common skill combinations observed
compatibility_matrix TEXT -- JSON: compatibility scores with other skills

-- Orchestration (3)
workflow_type TEXT       -- 'standalone', 'sequential', 'parallel', 'conditional'
orchestration_hints TEXT -- JSON: hints for combining with other skills
composition_templates TEXT -- JSON: pre-defined skill combinations
```

### 2. Skill Dependency Graph

**Graph Structure:**
```yaml
nodes:
  - id: "python-testing"
    type: "skill"
    metadata: {...}
  - id: "django-tdd"
    type: "skill"
    metadata: {...}

edges:
  - from: "python-testing"
    to: "django-tdd"
    type: "enhances"
    weight: 0.9
    evidence: "Django projects benefit from Django-specific TDD patterns"
  - from: "django-tdd"
    to: "python-testing"
    type: "depends_on"
    weight: 1.0
    evidence: "Django TDD builds on Python TDD fundamentals"
```

**Edge Types:**
- `depends_on` - Skill B required for skill A
- `enhances` - Skill B enhances skill A
- `conflicts_with` - Skills should not be used together
- `alternative_to` - Skill B is alternative to skill A
- `complements` - Skills work well together
- `prerequisite_for` - Skill A is prerequisite for skill B

### 3. Skill Recommendation Engine

**Recommendation Algorithm:**

```python
def recommend_skills(current_skill: str, context: dict) -> list[SkillRecommendation]:
    """
    Recommend related skills based on:
    1. Direct dependencies (from graph)
    2. Category similarity
    3. Tag overlap
    4. Usage patterns (from continuous-learning)
    5. Project context (tech stack)
    """

    # 1. Direct dependencies
    deps = get_dependencies(current_skill)

    # 2. Category similarity
    same_category = get_skills_by_category(current_skill.category)

    # 3. Tag overlap
    related_tags = get_skills_by_tag_overlap(current_skill.tags)

    # 4. Usage patterns (from continuous-learning)
    common_combos = get_common_combinations(current_skill)

    # 5. Project context
    project_relevant = filter_by_project_context(all_candidates, context)

    # Score and rank
    scored = score_and_rank(deps + same_category + related_tags + common_combos + project_relevant)

    return scored[:top_k]
```

**Recommendation Factors:**
- `dependency_weight` = 1.0 (highest priority)
- `category_similarity` = 0.7
- `tag_overlap` = 0.5
- `usage_frequency` = 0.6
- `project_relevance` = 0.8

### 4. Skill Composition Engine

**Composition Patterns:**

```yaml
sequential:
  - skill: "python-testing"
    step: 1
  - skill: "django-tdd"
    step: 2
    depends_on: step 1

parallel:
  - skill: "frontend-patterns"
    lane: ui
  - skill: "backend-patterns"
    lane: api
  - skill: "api-design"
    lane: contract

conditional:
  - if: "tech_stack == 'django'"
    then: "django-tdd"
    else: "python-testing"

composable:
  - skill: "tdd-workflow"
    with:
      - "django-tdd"  # if Django
      - "python-testing"  # otherwise
```

**Composition Verification:**
```python
def verify_composition(skills: list[str]) -> CompositionResult:
    """
    Verify that skill combination is safe and effective:
    1. Check for conflicts
    2. Check dependency satisfaction
    3. Check context compatibility
    4. Estimate quality boost
    """

    conflicts = check_conflicts(skills)
    if conflicts:
        return CompositionResult(valid=False, reason=f"Conflicts: {conflicts}")

    missing_deps = check_dependencies(skills)
    if missing_deps:
        return CompositionResult(valid=False, reason=f"Missing deps: {missing_deps}")

    quality_boost = estimate_quality_boost(skills)
    return CompositionResult(
        valid=True,
        quality_boost=quality_boost,
        confidence=calculate_confidence(skills)
    )
```

### 5. Continuous Learning Integration

**Learning from Sessions:**

```python
# From continuous-learning-v2 hooks
def observe_skill_usage(session: Session):
    """
    Capture skill usage patterns:
    - Which skills are used together?
    - In what order?
    - For what types of tasks?
    - With what success rate?
    """

    for skill_invocation in session.skill_invocations:
        record_usage(
            skill=skill_invocation.skill_id,
            context=skill_invocation.context,
            outcome=skill_invocation.outcome,
            related_skills=get_concurrent_skills(skill_invocation)
        )
```

**Pattern Extraction:**
```python
def extract_skill_patterns(observations: list[Observation]) -> SkillPatterns:
    """
    Extract patterns from observations:
    - Frequent skill combinations
    - Common sequences
    - Successful vs unsuccessful combinations
    """

    combinations = find_frequent_combinations(observations, min_support=0.1)
    sequences = find_common_sequences(observations, min_support=0.05)
    success_rates = calculate_success_rates(combinations)

    return SkillPatterns(
        combinations=combinations,
        sequences=sequences,
        success_rates=success_rates
    )
```

### 6. Skill Compliance Verification

**Integration with skill-comply:**

```python
def verify_skill_composition(composition: SkillComposition) -> ComplianceReport:
    """
    Use skill-comply to verify that agents follow the composition:
    1. Generate spec from composition
    2. Generate scenarios
    3. Run agents
    4. Measure compliance
    """

    spec = generate_spec_from_composition(composition)
    scenarios = generate_scenarios(spec, strictness_levels=['supportive', 'neutral', 'competing'])

    results = []
    for scenario in scenarios:
        result = run_agent_with_scenario(scenario)
        compliance = measure_compliance(result.tool_calls, spec)
        results.append(ComplianceResult(scenario, compliance))

    return ComplianceReport(results)
```

## Implementation Plan

### Phase 1: Database Schema Extension
1. Add composition columns to tool_registry
2. Create skill_relationships table
3. Create skill_usage_patterns table
4. Create migration script

### Phase 2: Dependency Graph Construction
1. Manual annotation of key skill relationships
2. Automated extraction from skill descriptions
3. Tag-based similarity calculation
4. Category-based clustering
5. Graph visualization tool

### Phase 3: Recommendation Engine
1. Implement recommendation algorithm
2. Build scoring system
3. Create API endpoint
4. Integrate with TiBrain CLI
5. A/B testing with users

### Phase 4: Composition Engine
1. Define composition patterns
2. Implement composition verification
3. Build composition templates
4. Create composition builder UI
5. Safety checks and rollback

### Phase 5: Continuous Learning
1. Integrate with continuous-learning-v2 hooks
2. Build pattern extraction pipeline
3. Implement feedback loop
4. Confidence scoring
5. Automatic relationship discovery

### Phase 6: Compliance Verification
1. Integrate with skill-comply
2. Build composition spec generator
3. Automated compliance testing
4. Compliance dashboard
5. Alert system for low compliance

## API Design

### GET /v1/tibrain/skill/recommendations
```json
{
  "skill_id": "python-testing",
  "context": {
    "project_type": "django",
    "tech_stack": ["python", "django", "postgresql"]
  },
  "limit": 5
}
```

Response:
```json
{
  "recommendations": [
    {
      "skill_id": "django-tdd",
      "reason": "Enhances python-testing for Django projects",
      "score": 0.95,
      "type": "enhances"
    },
    {
      "skill_id": "django-patterns",
      "reason": "Complementary Django patterns",
      "score": 0.85,
      "type": "complements"
    }
  ]
}
```

### POST /v1/tibrain/skill/compose
```json
{
  "skills": ["python-testing", "django-tdd"],
  "pattern": "sequential"
}
```

Response:
```json
{
  "composition": {
    "id": "comp-123",
    "skills": ["python-testing", "django-tdd"],
    "pattern": "sequential",
    "valid": true,
    "quality_boost": 2.5,
    "confidence": 0.9,
    "verification_status": "pending"
  }
}
```

### GET /v1/tibrain/skill/graph
```json
{
  "skill_id": "python-testing",
  "depth": 2
}
```

Response:
```json
{
  "nodes": [...],
  "edges": [...],
  "layout": "hierarchical"
}
```

## Integration Points

### With TiBrain CLI
```bash
# Get recommendations
tibrain skill recommend python-testing

# Compose skills
tibrain skill compose python-testing django-tdd --pattern sequential

# View skill graph
tibrain skill graph python-testing --depth 2
```

### With Devin/Claude
```python
# Automatic recommendation when skill is invoked
if skill_invocation.skill_id == "python-testing":
    recommendations = get_recommendations(skill_invocation.skill_id, context)
    if recommendations:
        suggest_to_user(recommendations)
```

### With Continuous Learning
```python
# Auto-update relationships based on usage
if session.success:
    update_skill_relationships(session.skill_invocations, success=True)
```

## Success Metrics

1. **Recommendation Accuracy**
   - Click-through rate: >30%
   - User satisfaction: >4/5
   - Time saved: >20%

2. **Composition Effectiveness**
   - Quality boost: >1.5x on average
   - Success rate: >80%
   - User adoption: >40%

3. **System Health**
   - Graph coverage: >90% of skills
   - Relationship accuracy: >85%
   - Compliance rate: >75%

## Risks and Mitigations

| Risk | Impact | Mitigation |
|------|--------|------------|
| Wrong recommendations | High | Confidence scoring, user feedback loop |
| Skill conflicts | High | Conflict detection, safety checks |
| Performance overhead | Medium | Caching, incremental updates |
| Maintenance burden | Medium | Auto-learning, manual override |
| Context contamination | Low | Project-scoped learning, promotion criteria |

## References

- Continuous Learning v2.1 - Instinct-based learning
- Skill Comply - Compliance measurement
- TiBrain Database Schema - 26 columns
- Awesome Omni Skills - 3,480 skills
