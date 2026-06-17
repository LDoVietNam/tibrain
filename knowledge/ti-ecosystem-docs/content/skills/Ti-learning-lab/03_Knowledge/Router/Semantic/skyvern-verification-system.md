# Skyvern Verification System

**Repository**: Skyvern  
**Location**: `skyvern/forge/sdk/copilot/`  
**Key Files**:
- `feasibility_gate.py` - Feasibility gate
- `agent.py` - Agent verification logic
- `actions/actions.py` - Verification result models

---

## Overview

Skyvern's Verification System validates workflow correctness through multiple verification stages including feasibility gates, judge systems, and LLM verification. It ensures workflows are valid before execution and provides actionable feedback for debugging.

---

## Feasibility Gate

**Location**: `forge/sdk/copilot/feasibility_gate.py`

**Purpose**: Fast-path validation for workflow feasibility

**Features**:
- Quick validation checks
- Early rejection of invalid workflows
- User-friendly feedback
- Performance optimization

### Feasibility Checks

1. **URL Validation** - Verify URLs are valid
2. **Parameter Validation** - Verify parameters are defined
3. **Block Validation** - Verify blocks are valid
4. **Cycle Detection** - Detect workflow cycles

---

## Judge System

**Purpose**: LLM-based verification of workflow correctness

**Features**:
- LLM verification
- Rule-based checks
- Hybrid verification
- Confidence scoring

---

## Verification Result Models

**Location**: `webeye/actions/actions.py`

### CompleteVerifyResult

```python
class CompleteVerifyResult(BaseModel):
    status: VerificationStatus | None = None
    user_goal_achieved: bool = False
    should_terminate: bool = False
    thoughts: str
    page_info: str | None = None
    failure_categories: list[dict] = []
```

### VerificationStatus

```python
class VerificationStatus(StrEnum):
    complete = "complete"
    terminate = "terminate"
    continue_step = "continue"
```

---

## Key Patterns

### 1. Feasibility Gate

**Pattern**: Fast-path validation for quick rejection

**Benefits**:
- Faster feedback
- Reduced LLM calls
- Better UX

### 2. LLM Verification

**Pattern**: Use LLM for complex verification

**Benefits**:
- Flexible validation
- Context-aware
- Intelligent feedback

### 3. Hybrid Verification

**Pattern**: Combine rule-based and LLM verification

**Benefits**:
- Best of both worlds
- Fast for simple cases
- Smart for complex cases

---

## Testing Considerations

### Test Scenarios

1. **Feasibility gate** - Verify quick validation
2. **LLM verification** - Verify LLM-based checks
3. **Hybrid verification** - Verify combined approach
4. **Error messages** - Verify user-friendly feedback

### Test Commands

```bash
# Run verification tests
python -m pytest tests/unit/test_verification.py -v

# Run feasibility gate tests
python -m pytest tests/unit/test_feasibility_gate.py -v
```

---

## References

- **Feasibility Gate**: `forge/sdk/copilot/feasibility_gate.py`
- **Verification Models**: `webeye/actions/actions.py`
- **Agent Verification**: `forge/sdk/copilot/agent.py`
