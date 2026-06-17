# Skyvern Workflow Engine Architecture

**Repository**: Skyvern  
**Location**: `skyvern/forge/sdk/workflow/`  
**Key Files**:
- `models/block.py` (8,195 lines) - 27 block types
- `service.py` (6,513 lines) - Workflow execution logic
- `models/workflow.py` (301 lines) - Workflow definition
- `context_manager.py` - Workflow run context management

---

## Overview

Skyvern's Workflow Engine is a sophisticated DAG-based execution system that orchestrates browser automation tasks through modular blocks. It supports both sequential and DAG execution patterns, with advanced features like conditional branching, loop blocks, and progressive script caching.

---

## Core Architecture

### 1. Block System (27 Block Types)

**Location**: `models/block.py` (lines 780-7717)

Skyvern uses a block-based architecture where each block represents a discrete operation. The 27 block types are:

#### Task Blocks
- **BaseTaskBlock** (line 780) - Base class for all task-oriented blocks
- **TaskBlock** (line 1448) - Basic task block
- **TaskV2Block** (line 5590) - Enhanced task block with child blocks
- **NavigationBlock** (line 5562) - Navigation-only task
- **ExtractionBlock** (line 5571) - Data extraction task
- **LoginBlock** (line 5577) - Authentication task
- **FileDownloadBlock** (line 5583) - File download task
- **UrlBlock** (line 5590) - URL handling task
- **ActionBlock** (line 5554) - Generic action task
- **HumanInteractionBlock** (line 5271) - Human-in-the-loop task
- **ValidationBlock** (line 5509) - Validation task

#### Loop Blocks
- **ForLoopBlock** (line 1608) - Iterative execution over a collection
- **WhileLoopBlock** (line 2567) - Conditional loop execution
- **LoopBlockExecutedResult** (line 1454) - Loop execution result tracking

#### Control Flow Blocks
- **ConditionalBlock** (line 7161) - Conditional branching logic
- **WorkflowTriggerBlock** (line 7717) - Workflow trigger

#### Data Processing Blocks
- **CodeBlock** (line 3231) - Custom Python code execution
- **TextPromptBlock** (line 3524) - LLM text prompt
- **FileParserBlock** (line 4536) - File parsing
- **PDFParserBlock** (line 5111) - PDF-specific parsing
- **HttpRequestBlock** (line 5815) - HTTP request execution
- **PrintPageBlock** (line 6294) - Page printing

#### Integration Blocks
- **DownloadToS3Block** (line 3750) - S3 download
- **UploadToS3Block** (line 3838) - S3 upload
- **FileUploadBlock** (line 3956) - File upload
- **SendEmailBlock** (line 4207) - Email sending

#### Utility Blocks
- **WaitBlock** (line 5224) - Delay execution

### 2. Block Execution Pattern

**Location**: `service.py` (lines 2164-2824)

```python
async def _execute_single_block(
    self,
    *,
    workflow: Workflow,
    block: BlockTypeVar,
    block_idx: int,
    blocks_cnt: int,
    workflow_run: WorkflowRun,
    organization: Organization,
    workflow_run_id: str,
    browser_session_id: str | None,
    script_blocks_by_label: dict[str, Any],
    loaded_script_module: Any,
    is_script_run: bool,
    blocks_to_update: set[str],
    parent_workflow_run_block_id: str | None = None,
) -> tuple[WorkflowRun, set[str], BlockResult | None, bool, dict[str, Any] | None]
```

**Key Features**:
- **Script vs Agent Execution**: Blocks can execute via cached scripts or AI agent
- **Parent Block Tracking**: Supports nested block execution (conditional branches, loops)
- **Block Result Tracking**: Returns execution status, outputs, and metadata
- **Progressive Caching**: Only executed blocks are cached for future runs

### 3. DAG Execution Engine

**Location**: `service.py` (lines 2003-2162)

**Sequential Execution** (`_execute_workflow_blocks`, line 1559):
- Simple linear execution of blocks
- Used for workflows without complex branching
- Faster execution with less overhead

**DAG Execution** (`_execute_workflow_blocks_dag`, line 2003):
```python
async def _execute_workflow_blocks_dag(
    self,
    *,
    workflow: Workflow,
    workflow_run: WorkflowRun,
    organization: Organization,
    browser_session_id: str | None,
    script_blocks_by_label: dict[str, Any],
    loaded_script_module: Any,
    is_script_run: bool,
    blocks_to_update: set[str],
) -> tuple[WorkflowRun, set[str]]:
```

**Key Features**:
- **Graph Building**: `_build_workflow_graph()` (line 3120) builds execution graph
- **Cycle Detection**: Prevents infinite loops (line 2144)
- **Conditional Scope Computation**: `compute_conditional_scopes()` (line 1540) tracks block hierarchy
- **Next Block Resolution**: Dynamic next block selection based on execution results
- **Branch Metadata**: Tracks which branch was taken in conditionals

**DAG Execution Flow**:
1. Build workflow graph with `label_to_block` and `default_next_map`
2. Compute conditional scopes for timeline nesting
3. Traverse graph starting from start_label
4. Execute each block via `_execute_single_block()`
5. Determine next block based on block type and result
6. Handle conditional branches specially (line 2103-2110)
7. Detect cycles and visited nodes
8. Execute finally_block if configured

### 4. Conditional Blocks & Progressive Branch Caching

**Location**: `models/block.py` (lines 1540-1606, 7161-7716)

**Conditional Scope Computation**:
```python
def compute_conditional_scopes(
    label_to_block: dict[str, Any],
    default_next_map: dict[str, str | None],
) -> dict[str, str]:
    """Map each block label to the conditional block label whose scope it belongs to."""
```

**Progressive Branch Caching Pattern**:
- **Conditional blocks are NOT cached** - they always run via agent to evaluate conditions at runtime
- **Cacheable blocks inside conditional branches ARE cached** when they execute
- **Multi-run branch coverage**: Different runs can take different branches, caching each progressively
- **Workflow definition vs execution**: Definition has all blocks, execution only executes some

**Example**:
```
Run 1: Takes branch A → caches blocks from A
Run 2: Takes branch B → caches blocks from B (preserves A's cache)
Run 3: Takes branch A again → uses cached blocks from A
```

### 5. Loop Blocks

**Location**: `models/block.py` (lines 1608-2567)

**ForLoopBlock** (line 1608):
```python
class ForLoopBlock(Block):
    loop_blocks: list[BlockTypeVar]
    loop_over: str  # Parameter reference or array
    max_iterations: int | None = None
    continue_on_failure: bool = False
    next_loop_on_failure: bool = False
```

**WhileLoopBlock** (line 2567):
```python
class WhileLoopBlock(Block):
    loop_blocks: list[BlockTypeVar]
    loop_condition: str  # Jinja2 template
    max_iterations: int | None = None
    continue_on_failure: bool = False
    next_loop_on_failure: bool = False
```

**Loop Execution Result** (line 1454):
```python
class LoopBlockExecutedResult(BaseModel):
    outputs_with_loop_values: list[list[dict[str, Any]]]
    block_outputs: list[BlockResult]
    last_block: BlockTypeVar | None
    natural_completion: bool = False  # True only when loop exhausted naturally
```

**Loop Status Resolution** (line 1509):
- **is_canceled()**: Block was canceled
- **is_completed()**: Block completed successfully or swallowed failure
- **is_terminated()**: Block was terminated
- **is_synthetic_loop_failure()**: Structural/safety limit failure
- **resolve_status()**: Decides overall status with parent swallow flags

**Key Features**:
- **Max Iterations**: Safety limit to prevent infinite loops
- **Failure Swallowing**: `continue_on_failure` and `next_loop_on_failure` flags
- **Natural Completion Tracking**: Distinguishes between natural exhaustion and early exit
- **Synthetic Failures**: Structural failures (max iterations, cancel) vs child failures

### 6. Parameter System

**Location**: `models/parameter.py`

**Parameter Types**:
- **WorkflowParameter** - Workflow-level input parameters
- **OutputParameter** - Block output parameters
- **ContextParameter** - Runtime context parameters
- **AWSSecretParameter** - AWS secret references
- **BitwardenLoginCredentialParameter** - Bitwarden credentials
- **AzureVaultCredentialParameter** - Azure vault credentials
- **OnePasswordCredentialParameter** - 1Password credentials

**Parameter Resolution** (line 804):
```python
def get_all_parameters(self, workflow_run_id: str) -> list[PARAMETER_TYPE]:
    """Collect all parameters for a block, including URL references."""
```

**Template Formatting** (line 817):
```python
def format_potential_template_parameters(self, workflow_run_context: WorkflowRunContext) -> None:
    """Format all Jinja2 template parameters with runtime values."""
```

### 7. Workflow Definition Validation

**Location**: `models/workflow.py` (lines 61-86)

```python
class WorkflowDefinition(BaseModel):
    version: int = 1
    parameters: list[PARAMETER_TYPE]
    blocks: List[BlockTypeVar]
    finally_block_label: str | None = None
    error_code_mapping: dict[str, str] | None = None
    workflow_system_prompt: str | None = None

    def validate(self) -> None:
        """Validate workflow definition for duplicate labels and finally block constraints."""
```

**Validation Rules**:
1. **Duplicate Label Detection**: All block labels must be unique
2. **Finally Block Validation**: Finally block must be a top-level block and must be terminal
3. **Graph Validation**: Workflow graph must be acyclic and connected
4. **Parameter Validation**: All parameter references must be valid

---

## Execution Modes

### 1. Script Execution Mode

**Location**: `service.py` (lines 1576-1658)

**When Used**:
- All top-level blocks have cached script blocks
- All script blocks have non-null `run_signature`
- `run_with: code` is specified

**Benefits**:
- **10-50x faster** than agent execution
- **Deterministic** behavior
- **No LLM costs** for cached blocks
- **Predictable latency**

**Script Loading**:
```python
script_blocks = await app.DATABASE.scripts.get_script_blocks_by_script_revision_id(
    script_revision_id=script.script_revision_id,
    organization_id=organization_id,
)
```

**Module Loading**:
```python
spec = importlib.util.spec_from_file_location("user_script", script_path)
loaded_script_module = importlib.util.module_from_spec(spec)
spec.loader.exec_module(loaded_script_module)
```

### 2. Agent Execution Mode

**When Used**:
- Blocks don't have cached scripts
- `run_with: agent` is specified
- Blocks require AI decision-making

**Benefits**:
- **Flexible** - handles dynamic scenarios
- **Adaptive** - can handle unexpected page states
- **Intelligent** - uses LLM for reasoning

### 3. Hybrid Mode

**Combination**:
- Some blocks execute via script
- Some blocks execute via agent
- Automatic fallback when script fails

---

## Error Handling

### Block Result Status

**Location**: `schemas/workflows.py`

```python
class BlockStatus(StrEnum):
    completed = "completed"
    failed = "failed"
    canceled = "canceled"
    terminated = "terminated"
    skipped = "skipped"
```

### Failure Swallowing

**Loop Blocks**:
- `continue_on_failure`: Continue loop iterations on child failure
- `next_loop_on_failure`: Move to next loop iteration on child failure

**Conditional Blocks**:
- `continue_on_failure`: Continue to next block on branch failure

### Error Code Mapping

**Location**: `models/block.py` (lines 867-879)

```python
# Inherit workflow-level error_code_mapping; block-level entries override on key conflicts
merged_mapping = dict(workflow_error_code_mapping or {})
merged_mapping.update(self.error_code_mapping or {})
```

**Purpose**: Provide user-friendly error messages for LLM error classification

---

## Performance Optimizations

### 1. Block-Level Script Generation

**Location**: `service.py` (line 1943)

**Previous Pattern**: Generated scripts after each action (~10-50x per workflow run)

**Current Pattern**: Generate scripts at block completion
```python
async def _generate_pending_script_for_block(
    self,
    workflow_run_block_id: str,
    workflow_run_id: str,
    organization_id: str,
) -> None:
```

**Impact**: Reduces script generation frequency by 10-50x

### 2. Progressive Caching

**Pattern**: Only cache blocks that actually execute

**Benefits**:
- **Faster first run**: Don't generate scripts for unused blocks
- **Smaller cache**: Only store executed blocks
- **Branch coverage**: Cache different branches progressively

### 3. DAG Execution Optimization

**Features**:
- **Cycle detection**: Prevents infinite loops
- **Visited tracking**: Avoids re-executing blocks
- **Conditional scope computation**: Efficient timeline nesting
- **Next block caching**: Pre-computed default next map

---

## Integration Points

### 1. Browser Session Management

**Location**: `service.py` (line 1044)

```python
async def auto_create_browser_session_if_needed(
    self,
    workflow: Workflow,
    workflow_run: WorkflowRun,
    organization: Organization,
) -> str | None:
```

### 2. Artifact Management

**Screenshot Storage**:
- Screenshots bundled with LLM prompts
- Stored in artifact system
- Referenced in workflow run timeline

### 3. Telemetry & Observability

**OpenTelemetry Integration**:
- Traced function decorators
- Structured logging with structlog
- Performance metrics collection

---

## Key Patterns

### 1. Template Parameter Formatting

**Pattern**: All block parameters support Jinja2 templates

```python
self.url = self.format_block_parameter_template_from_workflow_run_context(
    self.url, workflow_run_context
)
```

### 2. Error Code Mapping Inheritance

**Pattern**: Block-level error codes override workflow-level

```python
merged_mapping = dict(workflow_error_code_mapping or {})
merged_mapping.update(self.error_code_mapping or {})
```

### 3. Conditional Scope Tracking

**Pattern**: Track which conditional block each block belongs to

```python
conditional_scopes = compute_conditional_scopes(label_to_block, default_next_map)
parent_wrb_id = conditional_wrb_ids.get(cond_label)
```

### 4. Loop Failure Swallowing

**Pattern**: Distinguish between structural failures and child failures

```python
if self.natural_completion and not self.is_synthetic_loop_failure():
    # Swallow child failure
```

---

## Testing Considerations

### Test Scenarios

1. **Sequential vs DAG Execution**
2. **Conditional Branch Coverage**
3. **Loop Max Iterations**
4. **Failure Swallowing**
5. **Progressive Caching**
6. **Script vs Agent Execution**
7. **Cycle Detection**
8. **Parameter Resolution**

### Test Commands

```bash
# Run workflow tests
python -m pytest tests/unit/test_workflow.py -v

# Run conditional tests
python -m pytest tests/unit/test_conditional_blocks.py -v

# Run loop tests
python -m pytest tests/unit/test_loop_blocks.py -v
```

---

## References

- **Block Definitions**: `models/block.py` (8,195 lines)
- **Workflow Service**: `service.py` (6,513 lines)
- **Workflow Definition**: `models/workflow.py` (301 lines)
- **Context Manager**: `context_manager.py`
- **Script Generation**: `core/script_generations/generate_script.py` (3,908 lines)
