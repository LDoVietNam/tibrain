# Skyvern Task System

**Repository**: Skyvern  
**Location**: `skyvern/forge/sdk/schemas/tasks.py` (472 lines)  
**Related Files**:
- `forge/agent.py` (5,581 lines) - Task execution logic
- `forge/sdk/db/enums.py` - Task type enums
- `services/task_service.py` - Task service layer

---

## Overview

Skyvern's Task System manages browser automation tasks with three main types: general, validation, and action. Tasks execute through the agent with action sequences, status transitions, and verification logic.

---

## Task Types

**Location**: `forge/sdk/db/enums.py` (lines 12-16)

```python
class TaskType(StrEnum):
    general = "general"
    validation = "validation"
    action = "action"
```

### General Task

**Purpose**: Standard browser automation task with navigation and extraction

**Features**:
- Navigation goal
- Data extraction goal
- Navigation payload
- Error code mapping
- Complete/terminate criteria

### Validation Task

**Purpose**: Validate page state or data

**Features**:
- Verification logic
- Page state validation
- Data validation

### Action Task

**Purpose**: Execute specific actions without full task lifecycle

**Features**:
- Action-only execution
- Minimal overhead
- Direct action execution

---

## Task Schema

**Location**: `schemas/tasks.py` (lines 25-130)

### TaskBase

```python
class TaskBase(BaseModel):
    title: str | None = None
    url: str
    webhook_callback_url: str | None = None
    webhook_failure_reason: str | None = None
    totp_verification_url: str | None = None
    totp_identifier: str | None = None
    navigation_goal: str | None = None
    data_extraction_goal: str | None = None
    navigation_payload: dict[str, Any] | list | str | None = None
    error_code_mapping: dict[str, str] | None = None
    workflow_system_prompt: str | None = None
    proxy_location: ProxyLocationInput = None
    extracted_information_schema: dict[str, Any] | list | str | None = None
    extra_http_headers: dict[str, str] | None = None
    complete_criterion: str | None = None
    terminate_criterion: str | None = None
    task_type: TaskType | None = TaskType.general
    application: str | None = None
    include_action_history_in_verification: bool | None = False
    max_screenshot_scrolls: int | None = None
    browser_address: str | None = None
    download_timeout: float | None = None
    include_extracted_text: bool = True
```

### TaskRequest

```python
class TaskRequest(TaskBase):
    url: str  # Required
    webhook_callback_url: str | None = None
    totp_verification_url: str | None = None
    browser_session_id: str | None = None
    model: dict[str, Any] | None = None
    
    @model_validator(mode="after")
    def validate_url(self) -> Self:
        """Validate URL format."""
```

### PromptedTaskRequest

```python
class PromptedTaskRequest(TaskRequest):
    ai_fallback: bool | None = False
    publish_workflow: bool | None = False
    run_with: str | None = None  # "code" or "agent"
    user_prompt: str  # Required
```

---

## Task Status

**Location**: `schemas/tasks.py` (lines 195-261)

### Status Values

```python
class TaskStatus(StrEnum):
    created = "created"
    queued = "queued"
    running = "running"
    timed_out = "timed_out"
    failed = "failed"
    terminated = "terminated"
    completed = "completed"
    canceled = "canceled"
```

### Status Transition Logic

```python
def can_update_to(self, new_status: TaskStatus) -> bool:
    allowed_transitions: dict[TaskStatus, set[TaskStatus]] = {
        TaskStatus.created: {
            TaskStatus.queued,
            TaskStatus.running,
            TaskStatus.timed_out,
            TaskStatus.failed,
            TaskStatus.canceled,
        },
        TaskStatus.queued: {
            TaskStatus.running,
            TaskStatus.timed_out,
            TaskStatus.failed,
            TaskStatus.canceled,
        },
        TaskStatus.running: {
            TaskStatus.completed,
            TaskStatus.failed,
            TaskStatus.terminated,
            TaskStatus.timed_out,
            TaskStatus.canceled,
        },
        TaskStatus.failed: set(),
        TaskStatus.terminated: set(),
        TaskStatus.completed: set(),
        TaskStatus.timed_out: set(),
        TaskStatus.canceled: {TaskStatus.completed},
    }
    return new_status in allowed_transitions[self]
```

**Transition Diagram**:
```
created → queued → running → completed
          ↓         ↓         ↓
        timed_out  failed  terminated
          ↓         ↓         ↓
        canceled ←───────────┘
```

### Status Helper Methods

```python
def is_final(self) -> bool:
    """Check if status is final (no further transitions)."""
    return self in {
        TaskStatus.failed,
        TaskStatus.terminated,
        TaskStatus.completed,
        TaskStatus.timed_out,
        TaskStatus.canceled,
    }

def requires_extracted_info(self) -> bool:
    """Check if status requires extracted information."""
    return self in {TaskStatus.completed}

def cant_have_extracted_info(self) -> bool:
    """Check if status cannot have extracted information."""
    return self in {
        TaskStatus.created,
        TaskStatus.queued,
        TaskStatus.running,
        TaskStatus.failed,
        TaskStatus.terminated,
    }

def requires_failure_reason(self) -> bool:
    """Check if status requires failure reason."""
    return self in {TaskStatus.failed, TaskStatus.terminated}
```

---

## Task Execution

**Location**: `forge/agent.py` (5,581 lines)

### Execution Flow

1. **Task Creation** - Create task with request parameters
2. **Queue Task** - Add to execution queue
3. **Initialize Browser** - Create browser session
4. **Execute Actions** - Run action sequence through handler
5. **Verify Completion** - Check complete/terminate criteria
6. **Extract Information** - Extract data if extraction goal specified
7. **Update Status** - Transition to final status
8. **Persist Results** - Save extracted information and artifacts
9. **Send Webhook** - Call webhook if configured

### Action Sequence

**Pattern**: Task executes through sequence of actions

```python
actions = [
    ClickAction(...),
    InputTextAction(...),
    SelectOptionAction(...),
    ExtractAction(...),
    ...
]
```

### Verification Logic

**Location**: `forge/agent.py`

**Complete Criterion**: Check if task goal is achieved

```python
if complete_criterion:
    # Evaluate complete criterion
    is_complete = await evaluate_criterion(complete_criterion)
    if is_complete:
        status = TaskStatus.completed
```

**Terminate Criterion**: Check if task should terminate

```python
if terminate_criterion:
    # Evaluate terminate criterion
    should_terminate = await evaluate_criterion(terminate_criterion)
    if should_terminate:
        status = TaskStatus.terminated
```

---

## Task Output

**Location**: `schemas/tasks.py` (lines 263-294)

### Task Model

```python
class Task(TaskBase):
    created_at: datetime
    modified_at: datetime
    task_id: str
    status: TaskStatus
    extracted_information: dict[str, Any] | list | str | None = None
    failure_reason: str | None = None
    organization_id: str
    workflow_run_id: str | None = None
    workflow_permanent_id: str | None = None
    browser_session_id: str | None = None
    order: int | None = None
    retry: int | None = None
    max_steps_per_run: int | None = None
```

### TaskOutput

```python
class TaskOutput(BaseModel):
    task_id: str
    status: TaskStatus
    extracted_information: dict[str, Any] | list | str | None = None
    failure_reason: str | None = None
    artifact_ids: list[str] | None = None
```

---

## Error Handling

### Error Code Mapping

**Purpose**: Provide user-friendly error messages for LLM classification

**Pattern**:
```python
error_code_mapping = {
    "out_of_stock": "Return this error when the product is out of stock",
    "not_found": "Return this error when the product is not found",
}
```

### Failure Reason

**Location**: `schemas/tasks.py` (line 284)

```python
failure_reason: str | None = None
```

**Purpose**: Detailed explanation of task failure

---

## Task Types in Context

### Navigate Task

**Purpose**: Navigate to URL and verify page state

**Parameters**:
- `url` - Target URL
- `navigation_goal` - Goal description
- `complete_criterion` - Completion condition

### Extract Task

**Purpose**: Extract data from page

**Parameters**:
- `url` - Target URL
- `data_extraction_goal` - Extraction goal
- `extracted_information_schema` - Expected schema

### Loop Task

**Purpose**: Execute actions in loop

**Parameters**:
- `url` - Target URL
- `navigation_goal` - Goal description
- Loop configuration in block

---

## Key Patterns

### 1. Status Transition Validation

**Pattern**: Validate status transitions before updating

**Benefits**:
- Prevents invalid state transitions
- Ensures task lifecycle integrity
- Provides clear transition rules

### 2. Criterion-Based Completion

**Pattern**: Use criteria for flexible completion logic

**Benefits**:
- Supports complex completion conditions
- Enables early termination
- Flexible goal verification

### 3. Error Code Mapping

**Pattern**: Map error codes to user-friendly messages

**Benefits**:
- Improves LLM error classification
- Provides actionable error messages
- Separates technical errors from user-facing messages

### 4. Extracted Information Schema

**Pattern**: Define expected output schema

**Benefits**:
- Ensures structured output
- Enables validation
- Improves LLM extraction accuracy

---

## Performance Optimizations

### 1. Task Queueing

**Pattern**: Queue tasks for efficient execution

**Impact**: Better resource utilization

### 2. Browser Session Reuse

**Pattern**: Reuse browser sessions across tasks

**Impact**: Faster task execution, reduced overhead

### 3. Action Caching

**Pattern**: Cache action results

**Impact**: Reduces redundant browser operations

---

## Testing Considerations

### Test Scenarios

1. **Status transitions** - Verify all valid transitions work
2. **Invalid transitions** - Verify invalid transitions are blocked
3. **Complete criterion** - Verify completion logic
4. **Terminate criterion** - Verify termination logic
5. **Error code mapping** - Verify error classification
6. **Extracted information** - Verify data extraction

### Test Commands

```bash
# Run task tests
python -m pytest tests/unit/test_tasks.py -v

# Run task status tests
python -m pytest tests/unit/test_task_status.py -v
```

---

## References

- **Task Schema**: `forge/sdk/schemas/tasks.py` (472 lines)
- **Task Types**: `forge/sdk/db/enums.py` (lines 12-16)
- **Task Execution**: `forge/agent.py` (5,581 lines)
- **Task Service**: `services/task_service.py`
