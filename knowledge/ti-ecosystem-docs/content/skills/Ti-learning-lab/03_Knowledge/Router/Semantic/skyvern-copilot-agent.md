# Skyvern Copilot Agent System

**Repository**: Skyvern  
**Location**: `skyvern/forge/sdk/copilot/`  
**Key Files**:
- `agent.py` (727 lines) - Multi-turn tool-use agent
- `context.py` (255 lines) - Structured context for cross-turn memory
- `tools.py` (2,444 lines) - Copilot agent tools
- `enforcement.py` (1,199 lines) - Enforcement wrapper with nudges
- `runtime.py` - Agent runtime context
- `narration.py` - Narration state management
- `loop_detection.py` - Tool loop detection
- `failure_tracking.py` - Failure state tracking
- `mcp_adapter.py` - Model Context Protocol integration

---

## Overview

Skyvern's Copilot Agent is a sophisticated multi-turn tool-use agent for workflow building. It uses the OpenAI Agents SDK with LiteLLM for multi-provider LLM support, featuring enforcement nudges, structured context memory, failure tracking, and progressive workflow testing.

---

## Core Architecture

### 1. Agent Loop Structure

**Location**: `agent.py` (lines 1-727)

**Agent Type**: Multi-turn tool-use agent using OpenAI Agents SDK

**Key Components**:
- **System Prompt**: Built from workflow knowledge base + tool usage guide + security rules
- **User Context**: Workflow YAML + chat history + global LLM context + debug run info
- **Tool Set**: 15+ tools for workflow building and testing
- **Enforcement**: Nudge agent when it skips required steps
- **Context Memory**: Structured context across turns

### 2. Context System

**Location**: `context.py` (255 lines)

#### StructuredContext

**Purpose**: Cross-turn memory for tracking user interactions

**Fields**:
```python
class StructuredContext(BaseModel):
    user_goal: str = ""
    urls_visited: list[UrlVisit] = Field(default_factory=list)
    fields_filled: list[FieldFilled] = Field(default_factory=list)
    credentials_checked: list[CredentialCheck] = Field(default_factory=list)
    decisions_made: list[str] = Field(default_factory=list)
    workflow_state: str = ""
```

**Merge Logic** (line 61):
```python
def merge_turn_summary(self, tool_activity: list[dict]) -> None:
    """Merge tool activity into structured context."""
    for entry in tool_activity:
        tool = entry.get("tool", "")
        summary = entry.get("summary", "")
        
        if tool == "navigate_browser":
            url = summary.removeprefix("Navigated to ").strip()
            if url and not any(v.url == url for v in self.urls_visited):
                self.urls_visited.append(UrlVisit(url=url, summary=""))
        
        elif tool == "list_credentials":
            match = re.search(r"Found (\d+)", summary)
            found = int(match.group(1)) > 0 if match else False
            self.credentials_checked.append(CredentialCheck(credential_name=summary, found=found))
        
        elif tool == "type_text":
            parts = summary.split("into ")
            selector = parts[-1].strip("'\"") if len(parts) > 1 else ""
            # Intentionally omit value: typed text may contain PII / credentials.
            self.fields_filled.append(FieldFilled(selector=selector, label=selector))
        
        elif tool == "update_workflow":
            self.workflow_state = summary
        
        elif tool in ("click", "evaluate", "run_blocks_and_collect_debug", "get_run_results"):
            self.decisions_made.append(f"{tool}: {summary}")
```

**Context Limits** (lines 99-106):
```python
if len(self.decisions_made) > 20:
    self.decisions_made = self.decisions_made[-15:]
if len(self.urls_visited) > 50:
    self.urls_visited = self.urls_visited[-40:]
if len(self.fields_filled) > 50:
    self.fields_filled = self.fields_filled[-40:]
if len(self.credentials_checked) > 50:
    self.credentials_checked = self.credentials_checked[-40:]
```

#### CopilotContext

**Purpose**: Unified context for the copilot agent run

**Fields** (lines 134-199):
```python
@dataclass
class CopilotContext(AgentContext):
    # Enforcement state
    navigate_called: bool = False
    observation_after_navigate: bool = False
    navigate_enforcement_done: bool = False
    update_workflow_called: bool = False
    test_after_update_done: bool = False
    post_update_nudge_count: int = 0
    coverage_nudge_count: int = 0
    format_nudge_count: int = 0
    copilot_total_timeout_exceeded: bool = False
    user_message: str = ""
    
    # Tool tracking
    consecutive_tool_tracker: list[str] = field(default_factory=list)
    tool_activity: list[dict[str, Any]] = field(default_factory=list)
    
    # Token usage
    total_tokens_used: int | None = None
    input_tokens_used: int | None = None
    output_tokens_used: int | None = None
    
    # Workflow state
    last_workflow: Workflow | None = None
    last_workflow_yaml: str | None = None
    workflow_persisted: bool = False
    last_update_block_count: int | None = None
    last_test_ok: bool | None = None
    last_test_failure_reason: str | None = None
    failed_test_nudge_count: int = 0
    explore_without_workflow_nudge_count: int = 0
    last_failed_workflow_yaml: str | None = None
    pending_reconciliation_run_id: str | None = None
    null_data_streak_count: int = 0
```

### 3. Enforcement System

**Location**: `enforcement.py` (1,199 lines)

#### Nudge Types

**Post-Update Nudge** (line 78):
```python
POST_UPDATE_NUDGE = (
    "You updated the workflow but did not test it. "
    "You MUST call run_blocks_and_collect_debug (or update_and_run_blocks next time) "
    "to test at least the first block before responding to the user. "
    "This verifies the workflow actually works."
)
```

**Post-Navigate Nudge** (line 85):
```python
POST_NAVIGATE_NUDGE = (
    "You navigated to a page but did not observe its content. "
    "You MUST use evaluate, get_browser_screenshot, click, type_text, "
    "scroll, select_option, press_key, or console_messages "
    "to inspect the page before responding. Do NOT answer from memory."
)
```

**Intermediate Success Nudge** (line 92):
```python
POST_INTERMEDIATE_SUCCESS_NUDGE = (
    "STOP — do NOT respond to the user yet. "
    "Your workflow only covers a subset of what the user asked for. "
    "You MUST add the next block now: call update_and_run_blocks with the current "
    "block chain. The tool preserves verified prefix state and reruns only the "
    "invalidated frontier, so passing the full chain is cheap. "
    "Only respond to the user when every distinct action they requested is covered "
    "by a workflow block, or you have clear evidence that continuing is infeasible."
)
```

#### Nudge Limits

**Location**: `enforcement.py` (lines 30-50)

```python
MAX_POST_UPDATE_NUDGES = 2
MAX_INTERMEDIATE_NUDGES = 8
MAX_FAILED_TEST_NUDGES = 2
MAX_FORMAT_NUDGES = 2
MAX_EXPLORE_WITHOUT_WORKFLOW_NUDGES = 2
NULL_DATA_STREAK_ESCALATE_AT = 2
REPEATED_FRONTIER_STREAK_ESCALATE_AT = 2
REPEATED_FRONTIER_STREAK_STOP_AT = 3
PROBABLE_SITE_BLOCK_STREAK_STOP_AT = 2
MIN_BLOCKS_FOR_AUTO_COMPLETE = 10
TOTAL_TIMEOUT_SECONDS = 600
MAX_ITERATIONS = 50
```

#### Token Budget Management

**Location**: `enforcement.py` (lines 52-76)

```python
TOKEN_BUDGET = 90_000
TOKENS_PER_RESIZED_IMAGE = 765  # OpenAI detail=high cost per resized image

KEEP_RECENT_TOOL_OUTPUTS = 3
_RECENT_TOOL_OUTPUT_CHAR_CAP = 2000
_TOOL_OUTPUT_SUMMARIZE_THRESHOLD = 300
_TOOL_OUTPUT_TRUNCATION_SUFFIX = "\n... [older tool output truncated]"
_TOOL_OUTPUT_HEAD_TRUNCATION_SUFFIX = "\n... [truncated]"
```

**Screenshot Dropping** (line 55):
```python
SCREENSHOT_DROPPED_NUDGE = (
    "Your previous screenshot was dropped from context to recover from a token-budget overflow. "
    "Do NOT reason about the page from memory. Re-take the screenshot "
    "(get_browser_screenshot) or call evaluate before deciding your next step."
)
```

### 4. Tool System

**Location**: `tools.py` (2,444 lines)

#### Safety Ceilings

**Location**: `tools.py` (lines 86-99)

```python
# Absolute upper bound on a single run_blocks tool invocation
RUN_BLOCKS_SAFETY_CEILING_SECONDS = 1200  # 20 min

# Primary exit condition: seconds of no observed progress
RUN_BLOCKS_STAGNATION_WINDOW_SECONDS = 90

# 5 s balances responsiveness against false positives
RUN_BLOCKS_HEARTBEAT_INTERVAL_SECONDS = 5
```

#### Tool Categories

**Browser Interaction Tools**:
- `navigate_browser` - Navigate to URL
- `get_browser_screenshot` - Capture screenshot
- `evaluate` - Evaluate JavaScript
- `click` - Click element
- `type_text` - Type text into element
- `scroll` - Scroll page
- `select_option` - Select dropdown option
- `press_key` - Press keyboard key
- `console_messages` - Get console messages

**Workflow Building Tools**:
- `update_workflow` - Update workflow definition
- `update_and_run_blocks` - Update and test workflow
- `run_blocks_and_collect_debug` - Test workflow blocks
- `get_run_results` - Get workflow run results

**Credential Tools**:
- `list_credentials` - List available credentials

**Utility Tools**:
- `ask_question` - Ask user a question

#### Tool Loop Detection

**Location**: `loop_detection.py`

**Purpose**: Detect when agent is stuck in a tool loop

**Pattern**:
```python
def detect_tool_loop(consecutive_tool_tracker: list[str]) -> tuple[bool, str]:
    """Detect if agent is stuck in a tool loop."""
    if len(consecutive_tool_tracker) < 3:
        return False, ""
    
    # Check for repeated tool pattern
    if len(set(consecutive_tool_tracker[-3:])) == 1:
        return True, f"Agent is stuck in a loop calling {consecutive_tool_tracker[-1]}"
    
    return False, ""
```

### 5. Failure Tracking

**Location**: `failure_tracking.py`

#### Failure Normalization

```python
def normalize_failure_reason(failure_reason: str | None) -> str:
    """Normalize failure reason for tracking."""
```

#### Repeated Failure Detection

```python
def update_repeated_failure_state(
    frontier: str,
    failure_reason: str,
    state: dict[str, Any],
) -> tuple[bool, str]:
    """Track repeated failures at the same frontier."""
```

#### Action Sequence Fingerprint

```python
def compute_action_sequence_fingerprint(actions: list[dict]) -> str:
    """Compute fingerprint of action sequence for comparison."""
```

### 6. MCP Integration

**Location**: `mcp_adapter.py`

**Purpose**: Integrate Model Context Protocol tools into copilot

**SchemaOverlay**:
```python
class SchemaOverlay(BaseModel):
    """Overlay schema for MCP tool parameters."""
```

### 7. Narration System

**Location**: `narration.py`

**Purpose**: Provide real-time narration of agent actions

**NarratorState**:
```python
class NarratorState:
    """State for narration generation."""
```

**TransitionKind**:
```python
class TransitionKind(StrEnum):
    """Types of transitions for narration."""
```

---

## Agent Execution Flow

### 1. Session Initialization

**Location**: `session_factory.py`

**Steps**:
1. Resolve live browser session ID (if provided)
2. Build system prompt from workflow knowledge base
3. Build user context with workflow YAML and chat history
4. Initialize CopilotContext with enforcement state
5. Register tools with OpenAI Agents SDK

### 2. Turn Execution

**Location**: `agent.py` (lines 200-400)

**Steps**:
1. Parse user message and chat history
2. Build system prompt with tool usage guide
3. Build user context with workflow YAML
4. Run agent loop with enforcement
5. Track tool activity and token usage
6. Merge turn summary into structured context
7. Return AgentResult with updated workflow

### 3. Enforcement Loop

**Location**: `enforcement.py` (lines 200-800)

**Steps**:
1. Check if required steps were skipped
2. Apply appropriate nudge if needed
3. Track nudge counts
4. Escalate if limits exceeded
5. Manage token budget
6. Drop screenshots if needed
7. Truncate old tool outputs

### 4. Response Generation

**Location**: `agent.py` (lines 234-249)

**Response Types**:
```python
ResponseType = Literal["REPLY", "ASK_QUESTION", "REPLACE_WORKFLOW"]
```

**Response Building**:
```python
def _build_exit_result(
    ctx: CopilotContext,
    user_response: str,
    global_llm_context: str | None,
    cancelled: bool = False,
) -> AgentResult:
    """AgentResult for agent-loop exits."""
    verified_workflow, verified_yaml = _verified_workflow_or_none(ctx)
    return AgentResult(
        user_response=user_response,
        updated_workflow=verified_workflow,
        global_llm_context=global_llm_context,
        workflow_yaml=verified_yaml,
        workflow_was_persisted=ctx.workflow_persisted,
        total_tokens=ctx.total_tokens_used,
        cancelled=cancelled,
    )
```

---

## Key Patterns

### 1. Structured Context Memory

**Pattern**: Track user interactions across turns with structured data

**Benefits**:
- Preserves context without full chat history
- Enables smart nudges based on past actions
- Supports progressive workflow building

### 2. Enforcement Nudges

**Pattern**: Guide agent behavior with corrective nudges

**Benefits**:
- Ensures required steps are followed
- Prevents shortcuts that break workflows
- Maintains workflow quality

### 3. Token Budget Management

**Pattern**: Manage context size with intelligent truncation

**Benefits**:
- Prevents context overflow
- Maintains recent tool outputs at full detail
- Drops old outputs to compact synopses

### 4. Failure Tracking

**Pattern**: Track repeated failures to detect stuck states

**Benefits**:
- Escalates when agent is stuck
- Provides actionable error messages
- Prevents infinite loops

### 5. Progressive Testing

**Pattern**: Test workflow incrementally as blocks are added

**Benefits**:
- Catches errors early
- Maintains verified prefix state
- Reduces rework

---

## Performance Optimizations

### 1. Tool Output Truncation

**Pattern**: Keep recent outputs at full detail, truncate old ones

**Impact**: Reduces context size while preserving recent information

### 2. Screenshot Dropping

**Pattern**: Drop screenshots when token budget exceeded

**Impact**: Prevents context overflow while maintaining functionality

### 3. Batch Tool Calls

**Pattern**: Group related tool calls together

**Impact**: Reduces round trips to LLM

---

## Error Handling

### Tool Loop Detection

**Detection**: Track consecutive tool calls

**Action**: Break loop with nudge

### Failure Escalation

**Detection**: Track repeated failures at same frontier

**Action**: Escalate to user with actionable message

### Timeout Handling

**Detection**: Monitor for stagnation (no progress for 90s)

**Action**: Cancel run with timeout message

---

## Testing Considerations

### Test Scenarios

1. **Enforcement nudges** - Verify nudges trigger correctly
2. **Token budget** - Verify context management under budget pressure
3. **Tool loops** - Verify loop detection breaks infinite loops
4. **Failure tracking** - Verify repeated failure escalation
5. **Progressive testing** - Verify workflow testing works incrementally
6. **Context memory** - Verify structured context merges correctly

### Test Commands

```bash
# Run copilot tests
python -m pytest tests/unit/test_copilot.py -v

# Run enforcement tests
python -m pytest tests/unit/test_copilot_enforcement.py -v

# Run loop detection tests
python -m pytest tests/unit/test_loop_detection.py -v
```

---

## References

- **Copilot Agent**: `forge/sdk/copilot/agent.py` (727 lines)
- **Context System**: `forge/sdk/copilot/context.py` (255 lines)
- **Tool System**: `forge/sdk/copilot/tools.py` (2,444 lines)
- **Enforcement**: `forge/sdk/copilot/enforcement.py` (1,199 lines)
- **Failure Tracking**: `forge/sdk/copilot/failure_tracking.py`
- **Loop Detection**: `forge/sdk/copilot/loop_detection.py`
- **MCP Adapter**: `forge/sdk/copilot/mcp_adapter.py`
- **Narration**: `forge/sdk/copilot/narration.py`
