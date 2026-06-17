# Skyvern Telemetry & Observability

**Repository**: Skyvern  
**Location**: `skyvern/forge/sdk/`  
**Key Files**:
- `trace.py` - Tracing utilities
- `core/skyvern_context.py` - Skyvern context
- `event/factory.py` - Event factory
- `log_artifacts.py` - Log artifact management

---

## Overview

Skyvern's Telemetry & Observability system provides comprehensive monitoring, tracing, and logging using OpenTelemetry integration. It tracks workflow execution, agent behavior, and performance metrics for production monitoring and debugging.

---

## OpenTelemetry Integration

**Location**: `forge/sdk/trace.py`

**Purpose**: Distributed tracing with OpenTelemetry

**Features**:
- Span creation
- Context propagation
- Attribute tracking
- Trace export

### Traced Decorator

```python
def traced(func: Callable) -> Callable:
    """Decorator to add tracing to functions."""
```

### Context Attributes

```python
def apply_context_attrs(span, **attrs) -> None:
    """Apply attributes to span from context."""
```

---

## Skyvern Context

**Location**: `forge/sdk/core/skyvern_context.py`

**Purpose**: Context management for telemetry

**Features**:
- Context creation
- Context propagation
- Context attributes
- Token tracking

### SkyvernContext

```python
class SkyvernContext:
    """Context for Skyvern operations."""
    workflow_run_id: str
    task_id: str
    step_id: str
    organization_id: str
    # ... other attributes
```

---

## Event System

**Location**: `forge/sdk/event/factory.py`

**Purpose**: Event creation and management

**Features**:
- Event creation
- Event streaming
- Event filtering
- Event aggregation

### EventStrategyFactory

```python
class EventStrategyFactory:
    """Factory for event strategies."""
```

---

## Log Artifacts

**Location**: `forge/sdk/log_artifacts.py`

**Purpose**: Manage log artifacts

**Features**:
- Log collection
- Log storage
- Log retrieval
- Log analysis

---

## Key Patterns

### 1. Distributed Tracing

**Pattern**: Use OpenTelemetry for distributed tracing

**Benefits**:
- End-to-end visibility
- Performance monitoring
- Debugging support

### 2. Context Propagation

**Pattern**: Propagate context across operations

**Benefits**:
- Consistent telemetry
- Request correlation
- Distributed debugging

### 3. Event Streaming

**Pattern**: Stream events for real-time monitoring

**Benefits**:
- Real-time visibility
- Event aggregation
- Alerting

---

## Performance Optimizations

### 1. Sampling

**Impact**: Reduced telemetry overhead

### 2. Async Export

**Impact**: Non-blocking telemetry

### 3. Batching

**Impact**: Reduced network overhead

---

## Testing Considerations

### Test Scenarios

1. **Tracing** - Verify span creation
2. **Context propagation** - Verify context flow
3. **Event streaming** - Verify event delivery
4. **Log artifacts** - Verify log collection

### Test Commands

```bash
# Run telemetry tests
python -m pytest tests/unit/test_telemetry.py -v

# Run tracing tests
python -m pytest tests/unit/test_tracing.py -v
```

---

## References

- **Trace**: `forge/sdk/trace.py`
- **Skyvern Context**: `forge/sdk/core/skyvern_context.py`
- **Event Factory**: `forge/sdk/event/factory.py`
- **Log Artifacts**: `forge/sdk/log_artifacts.py`
