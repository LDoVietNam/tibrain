# Skyvern Script Generation & Caching System

**Repository**: Skyvern  
**Location**: `skyvern/core/script_generations/`  
**Key Files**:
- `generate_script.py` (3,908 lines) - Python code generation
- `transform_workflow_run.py` (410 lines) - Workflow run transformation
- `generate_workflow_parameters.py` - Parameter schema generation
- `parameter_reference_guard.py` - Parameter validation
- `deterministic_field_naming.py` - Field naming logic

**Service Files**:
- `skyvern/services/workflow_script_service.py` - Script storage and caching
- `skyvern/forge/sdk/workflow/service.py` - Regeneration decision logic

---

## Overview

Skyvern's Script Generation system converts workflow runs into executable Python code that can be cached and reused. This enables "run with code" mode where workflows execute via cached scripts instead of the AI agent, providing **10-50x performance improvement** and **zero LLM costs** for cached blocks.

---

## Core Architecture

### 1. Script Generation Pipeline

**Location**: `generate_script.py` (3,908 lines)

**Pipeline Stages**:

1. **Transform Workflow Run** → `transform_workflow_run.py`
   - Convert DB workflow run into code generation input
   - Batch fetch tasks and actions
   - Process ForLoop child blocks recursively
   - Merge execution data into definition blocks

2. **Generate Workflow Parameters** → `generate_workflow_parameters.py`
   - Create parameter schema from actions
   - Assign deterministic field names
   - Handle parameter reference guard

3. **Generate Python Code** → `generate_script.py`
   - Generate executable Python code
   - Create script block metadata
   - Handle all 27 block types

4. **Store Script** → `workflow_script_service.py`
   - Upload script files to storage
   - Store script blocks in database
   - Track script revisions

### 2. Code Generation Input

**Location**: `transform_workflow_run.py` (line 17)

```python
@dataclass
class CodeGenInput:
    file_name: str
    workflow_run: dict[str, Any]
    workflow: dict[str, Any]
    workflow_blocks: list[dict[str, Any]]
    actions_by_task: dict[str, list[dict[str, Any]]]
    task_v2_child_blocks: dict[str, list[dict[str, Any]]]
```

### 3. Batch Query Optimization

**Location**: `transform_workflow_run.py` (lines 261-295)

**Previous Pattern** (N+1 queries):
```python
# Old approach: N+1 queries
for block in blocks:
    task = await get_task(block.task_id)  # 1 query per block
    actions = await get_task_actions(task.task_id)  # 1 query per block
```

**Current Pattern** (2 queries):
```python
# New approach: Batch fetch upfront
all_task_ids: set[str] = set()
for rb in workflow_run_blocks:
    if rb.block_type in SCRIPT_TASK_BLOCKS and rb.task_id:
        all_task_ids.add(rb.task_id)

# Single query for all tasks
tasks = await app.DATABASE.tasks.get_tasks_by_ids(task_ids=task_ids_list)
tasks_by_id = {task.task_id: task for task in tasks}

# Single query for all actions (returns desc order for timeline; reverse for chronological)
all_actions = await app.DATABASE.tasks.get_tasks_actions(task_ids=task_ids_list)
all_actions.reverse()
for action in all_actions:
    actions_by_task_id[action.task_id].append(action)
```

**Impact**: Reduces from **2N queries to 2 queries** for workflows with N task blocks.

**For workflows with 20 blocks**:
- Old: 40 DB queries
- New: 2 DB queries
- **20x reduction**

---

## ForLoop Child Block Processing

**Location**: `transform_workflow_run.py` (lines 74-223)

### Challenge

When a ForLoop iterates N times, there are N child run blocks per label. We need to pick the best candidate per label for code generation.

### Solution: Best Candidate Selection

**Selection Priority** (lines 97-133):

1. **Task Blocks**:
   - Prefer block with `task_id` over block without
   - When both have `task_id`, prefer more actions
   - On action tie, prefer `completed` status over failed/other

2. **Nested ForLoops**:
   - Prefer iteration that produced grandchildren
   - On tie, break by total deep-descendant action count
   - Ensures usable actions win even at 3+ nesting levels

**Algorithm**:
```python
child_run_blocks_by_label: dict[str, Any] = {}
for b in child_run_blocks:
    existing = child_run_blocks_by_label.get(b.label)
    if existing is None:
        child_run_blocks_by_label[b.label] = b
    elif b.block_type in SCRIPT_TASK_BLOCKS:
        # Prefer iteration with richest execution evidence
        if b.task_id and not existing.task_id:
            child_run_blocks_by_label[b.label] = b
        elif b.task_id and existing.task_id:
            b_actions = len(actions_by_task_id.get(b.task_id, []))
            existing_actions = len(actions_by_task_id.get(existing.task_id, []))
            if b_actions > existing_actions:
                child_run_blocks_by_label[b.label] = b
            elif b_actions == existing_actions:
                # Break tie by status: completed > everything else
                b_completed = str(b.status) == "completed"
                existing_completed = str(existing.status) == "completed"
                if b_completed and not existing_completed:
                    child_run_blocks_by_label[b.label] = b
    elif b.block_type == BlockType.FOR_LOOP:
        # Prefer nested for-loop iteration that produced grandchildren
        existing_children = children_by_parent.get(existing.workflow_run_block_id, [])
        b_children_list = children_by_parent.get(b.workflow_run_block_id, [])
        if len(b_children_list) > len(existing_children):
            child_run_blocks_by_label[b.label] = b
```

### Recursive Processing

**Nested ForLoops** (lines 191-204):
```python
# Recursively process nested for-loops so their inner blocks
# also get task_id and actions merged
if child_run_block and child_run_block.block_type == BlockType.FOR_LOOP:
    nested_forloop_count += 1
    inner_loop_blocks = loop_block_dump.get("loop_blocks", [])
    if inner_loop_blocks:
        loop_block_dump["loop_blocks"] = _process_forloop_children(
            forloop_run_block=child_run_block,
            loop_blocks_def=inner_loop_blocks,
            children_by_parent=children_by_parent,
            tasks_by_id=tasks_by_id,
            actions_by_task_id=actions_by_task_id,
            actions_by_task=actions_by_task,
        )
```

**Purpose**: Handles deeply nested blocks (e.g., extraction inside a double-nested for-loop).

---

## Block-Level Script Generation

**Location**: `service.py` (line 1943)

### Previous Pattern

Generated scripts after each action (~10-50x per workflow run):
```python
# Old approach: Generate after each action
for action in actions:
    await generate_or_update_pending_workflow_script()  # Called 10-50x per run
```

### Current Pattern

Generate scripts at block completion:
```python
async def _generate_pending_script_for_block(
    self,
    workflow_run_block_id: str,
    workflow_run_id: str,
    organization_id: str,
) -> None:
    """Generate script for a single block after completion."""
```

**Called From**:
- `_execute_workflow_blocks()` (line 1316)
- `_execute_workflow_blocks_dag()` (line 2090)

**Impact**: Reduces script generation frequency by **10-50x** while maintaining progressive updates.

---

## Parameter Reference Guard

**Location**: `parameter_reference_guard.py`

### Purpose

Prevent LLM from hallucinating parameter references that don't exist in the workflow definition.

### Implementation

**Valid Keys Collection** (lines 76-94):
```python
def _collect_declared_param_keys(workflow: dict[str, Any]) -> frozenset[str]:
    """Return the set of all parameter keys from the workflow definition.

    Includes workflow, output, and context parameters — any parameter type whose
    key can legally appear as `context.parameters['key']` at runtime.
    """
    keys: set[str] = set()
    defn = workflow.get("workflow_definition") or {}
    for param in defn.get("parameters") or []:
        if not isinstance(param, dict):
            continue
        key = param.get("key")
        if key:
            keys.add(key)
    return frozenset(keys)
```

**Secret Parameter Filtering** (lines 97-117):
```python
def _collect_secret_param_keys(workflow: dict[str, Any]) -> frozenset[str]:
    """Return the set of parameter keys that carry credential/secret data.

    Routes each declared parameter through ``is_sensitive_workflow_parameter`` —
    the canonical filter that combines ``ParameterType.is_secret_or_credential``
    (aws_secret, bitwarden_*, onepassword, azure_*, credential) with the
    ``workflow_parameter_type=credential_id`` sub-check.
    """
```

**Credential Parameter Filtering** (lines 120-143):
```python
def _collect_credential_param_keys(workflow: dict[str, Any]) -> frozenset[str]:
    """Parameter keys typed `WorkflowParameterType.CREDENTIAL_ID`.

    Narrower than `_collect_secret_param_keys`: only params whose runtime
    value is a `{username, password, totp}` dict expanded by `setup()`.
    """
```

**Validation**:
```python
def validate_context_parameter_refs(
    code: str,
    valid_keys: frozenset[str],
    secret_keys: frozenset[str],
    credential_keys: frozenset[str],
) -> tuple[bool, list[str], list[str]]:
    """Validate that all context.parameter references in code are valid."""
```

---

## Deterministic Field Naming

**Location**: `deterministic_field_naming.py`

### Purpose

Assign consistent field names to workflow parameters across script regenerations to prevent schema mismatches with cached block code.

### Implementation

**Field Assignment** (in `generate_workflow_parameters.py`):
```python
def generate_workflow_parameters_schema(
    blocks: list[dict[str, Any]],
    actions_by_task: dict[str, list[dict[str, Any]]],
    existing_field_assignments: dict[int, str] | None = None,
) -> tuple[str, dict[int, str]]:
    """
    Generate a Pydantic schema for workflow parameters with deterministic field names.
    
    Args:
        blocks: List of block dictionaries
        actions_by_task: Dictionary mapping task IDs to lists of action dictionaries
        existing_field_assignments: Dictionary mapping action index to existing field names (for unchanged blocks)
    
    Returns:
        tuple of (schema_code, field_assignments)
    """
```

**Existing Field Preservation** (in `generate_script.py`, lines 210-291):
```python
def _build_existing_field_assignments(
    blocks: list[dict[str, Any]],
    actions_by_task: dict[str, list[dict[str, Any]]],
    cached_blocks: dict[str, ScriptBlockSource],
    updated_block_labels: set[str],
) -> dict[int, str]:
    """
    Build a mapping of action index (1-based) to existing field names for unchanged blocks.
    
    This is used to tell the LLM which field names must be preserved when regenerating
    the workflow parameters schema, preventing schema mismatches with cached block code.
    """
```

---

## Caching Decision Logic

**Location**: `service.py` (lines 357-449)

### Two Mechanisms for Detecting New Blocks

| Mechanism | Location | What it catches |
|-----------|----------|-----------------|
| Execution tracking | service.py:1316 | Blocks that EXECUTED and aren't cached |
| `missing_labels` check | service.py:3436-3441 | Blocks in DEFINITION that aren't cached |

### For Workflows WITHOUT Conditionals

These mechanisms are equivalent - all blocks in definition execute in a single run.

### For Workflows WITH Conditionals

They differ significantly:
- **Definition**: Has all blocks (all branches)
- **Execution**: Only executes some blocks (one branch per run)

### Progressive Branch Caching

**Pattern**:
```
Run 1: Takes branch A → caches blocks from A
Run 2: Takes branch B → caches blocks from B (preserves A's cache)
Run 3: Takes branch A again → uses cached blocks from A
```

**Key Insight**: Conditional blocks are NOT cached - they always run via agent. But cacheable blocks inside conditional branches ARE cached when they execute.

---

## Script Block Requirements

### For `run_with: code` Mode

For a workflow to execute with cached scripts, ALL top-level blocks must have:

1. **A `script_block` database entry**
2. **A non-null `run_signature` field**

Without these, the system falls back to `run_with: agent`.

### Block Types That Should Be Cached

**Location**: `service.py` (constant)

```python
BLOCK_TYPES_THAT_SHOULD_BE_CACHED = [
    BlockType.TASK,
    BlockType.NAVIGATION,
    BlockType.EXTRACTION,
    BlockType.LOGIN,
    BlockType.FILE_DOWNLOAD,
    BlockType.ACTION,
    BlockType.FOR_LOOP,
    # ... other cacheable types
]
```

**Excluded Types**:
- `BlockType.CONDITIONAL` - Always runs via agent
- `BlockType.WHILE_LOOP` - Always runs via agent
- `BlockType.CODE` - Custom code, not cached
- `BlockType.WAIT` - Utility block, not cached

---

## Adding New Cacheable Block Types

### Steps

1. **Add to `BLOCK_TYPES_THAT_SHOULD_BE_CACHED`** in `service.py`

2. **Add handling in `generate_workflow_script_python_code()`** with BOTH:
   - `create_or_update_script_block()` - stores metadata in database
   - `append_block_code(block_code)` - adds code to generated script output

3. **Ensure `run_signature` is set** - the code statement to execute the block

**CRITICAL**: Every block type needs BOTH database entry AND script output. Missing `append_block_code()` causes runtime failures even if database entries exist.

---

## Block Processing Order

**Location**: `generate_script.py` (main generation function)

**Order**:
1. `task_v1_blocks` - Blocks in `SCRIPT_TASK_BLOCKS`
2. `task_v2_blocks` - task_v2 blocks with child blocks
3. `for_loop_blocks` - ForLoop container blocks
4. `__start_block__` - Workflow entry point

---

## Performance Optimizations

### 1. Batch Task and Action Queries

**Impact**: 20x reduction for workflows with 20 blocks (40 queries → 2 queries)

### 2. Block-Level Script Generation

**Impact**: 10-50x reduction in script generation frequency

### 3. Progressive Caching

**Impact**: Only cache blocks that execute, faster first runs

### 4. Deterministic Field Naming

**Impact**: Prevents schema mismatches, reduces regeneration needs

---

## Database Operations per Regeneration

Each regeneration does:
1. **DELETE** old script blocks
2. **CREATE** new script blocks
3. **UPLOAD** script files to storage
4. **INSERT** script metadata

**Cost**: Unnecessary regenerations can flood the database.

**Mitigation**: Use `blocks_to_update` tracking to only regenerate changed blocks.

---

## Error Handling

### Parameter Reference Guard

**HallucinatedParameterError** (line 29):
```python
class HallucinatedParameterError(Exception):
    """Raised when LLM hallucinates a parameter reference that doesn't exist."""
```

**Guard Result Logging** (line 31):
```python
def log_or_raise_guard_result(
    guard_result: tuple[bool, list[str], list[str]],
    workflow_run_id: str,
    raise_on_error: bool = False,
) -> None:
    """Log or raise based on guard result."""
```

### Script Generation Failures

**Fallback**: If script generation fails, workflow falls back to agent execution.

**Logging**: All failures are logged with structured context.

---

## Testing Considerations

### Test Scenarios

1. **Same blocks run twice** - Should NOT regenerate on 2nd run
2. **New block added** - Should regenerate to include new block
3. **Workflow with conditionals** - Different branches should cache progressively
4. **Block type not in `BLOCK_TYPES_THAT_SHOULD_BE_CACHED`** - Should NOT trigger caching
5. **ForLoop with nested blocks** - Should handle recursive processing
6. **Parameter reference guard** - Should catch hallucinated references
7. **Deterministic field naming** - Should preserve field names across regenerations

### Test Commands

```bash
# Run script-related tests
python -m pytest tests/unit/ -k "script" --ignore=tests/unit/test_security.py -v

# Run conditional caching tests specifically
python -m pytest tests/unit/test_conditional_script_caching.py -v

# Run forloop script tests
python -m pytest tests/unit/test_forloop_script_generation.py -v
```

---

## Key Patterns

### 1. Batch Query Pattern

**Pattern**: Collect all IDs first, then batch fetch all data in 1-2 queries.

**Benefits**:
- Reduces database load
- Faster execution
- Better scalability

### 2. Best Candidate Selection

**Pattern**: When multiple candidates exist, select the one with richest execution evidence.

**Benefits**:
- Preserves most data
- Handles edge cases (empty iterations)
- Works with nested structures

### 3. Progressive Caching

**Pattern**: Only cache blocks that execute, not all blocks in definition.

**Benefits**:
- Faster first runs
- Smaller cache
- Supports conditional branching

### 4. Deterministic Field Naming

**Pattern**: Assign consistent field names across regenerations.

**Benefits**:
- Prevents schema mismatches
- Reduces regeneration needs
- Maintains cached block compatibility

---

## References

- **Script Generation**: `core/script_generations/generate_script.py` (3,908 lines)
- **Workflow Transform**: `core/script_generations/transform_workflow_run.py` (410 lines)
- **Parameter Guard**: `core/script_generations/parameter_reference_guard.py`
- **Field Naming**: `core/script_generations/deterministic_field_naming.py`
- **Workflow Service**: `forge/sdk/workflow/service.py` (6,513 lines)
- **Script Service**: `services/workflow_script_service.py`
