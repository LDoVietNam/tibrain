# Multi-Step Reasoning with Planning and Memory

**Pattern ID**: `multi-step-reasoning-planning-memory`
**Source**: browser-use (browser_use/agent/service.py), Skyvern (skyvern/forge/agent.py)
**Category**: AI Agent, Reasoning
**Complexity**: Advanced

## Problem

AI agents need to:
- Break down complex tasks into manageable steps
- Track progress across multiple steps
- Remember context from previous steps
- Adapt plans when obstacles arise
- Avoid infinite loops and repetitive failures

## Solution

Implement a multi-step reasoning system with planning, memory, and adaptive replanning:

### Core Components

1. **Plan State Management**
   ```python
   @dataclass
   class PlanItem:
       text: str
       status: str  # 'done', 'current', 'pending', 'skipped'

   @dataclass
   class AgentState:
       plan: list[PlanItem] | None = None
       current_plan_item_index: int = 0
       plan_generation_step: int = 0
       consecutive_failures: int = 0
   ```

2. **Plan Creation and Updates**
   ```python
   def _update_plan_from_model_output(
       self,
       model_output: AgentOutput
   ) -> None:
       """Update plan state from model output."""
       if not self.settings.enable_planning:
           return

       # If model provided a new plan, replace current plan
       if model_output.plan_update is not None:
           self.state.plan = [
               PlanItem(text=step_text)
               for step_text in model_output.plan_update
           ]
           self.state.current_plan_item_index = 0
           self.state.plan_generation_step = self.state.n_steps
           if self.state.plan:
               self.state.plan[0].status = 'current'
           logger.info(
               f'Plan {"updated" if self.state.plan_generation_step else "created"} '
               f'with {len(self.state.plan)} steps'
           )
           return

       # If model provided step index update, advance plan
       if model_output.current_plan_item is not None and self.state.plan is not None:
           new_idx = model_output.current_plan_item
           new_idx = max(0, min(new_idx, len(self.state.plan) - 1))
           old_idx = self.state.current_plan_item_index

           # Mark steps between old and new as done
           for i in range(old_idx, new_idx):
               if i < len(self.state.plan) and \
                  self.state.plan[i].status in ('current', 'pending'):
                   self.state.plan[i].status = 'done'

           # Mark new step as current
           if new_idx < len(self.state.plan):
               self.state.plan[new_idx].status = 'current'

           self.state.current_plan_item_index = new_idx
   ```

3. **Plan Rendering for Context**
   ```python
   def _render_plan_description(self) -> str | None:
       """Render current plan as text description."""
       if not self.settings.enable_planning or self.state.plan is None:
           return None

       markers = {
           'done': '[x]',
           'current': '[>]',
           'pending': '[ ]',
           'skipped': '[-]'
       }
       lines = []
       for i, step in enumerate(self.state.plan):
           marker = markers.get(step.status, '[ ]')
           lines.append(f'{marker} {i}: {step.text}')
       return '\n'.join(lines)
   ```

4. **Replan Nudges**
   ```python
   def _inject_replan_nudge(self) -> None:
       """Inject replan nudge when stall detection threshold is met."""
       if not self.settings.enable_planning or self.state.plan is None:
           return
       if self.settings.planning_replan_on_stall <= 0:
           return
       if self.state.consecutive_failures >= self.settings.planning_replan_on_stall:
           msg = (
               'REPLAN SUGGESTED: You have failed '
               f'{self.state.consecutive_failures} consecutive times. '
               'Your current plan may need revision. '
               'Output a new `plan_update` with revised steps to recover.'
           )
           logger.info(
               f'Replan nudge injected after '
               f'{self.state.consecutive_failures} consecutive failures'
           )
           self._message_manager._add_context_message(UserMessage(content=msg))
   ```

5. **Exploration Nudges**
   ```python
   def _inject_exploration_nudge(self) -> None:
       """Nudge agent to create plan after exploring without one."""
       if not self.settings.enable_planning or self.state.plan is not None:
           return
       if self.settings.planning_exploration_limit <= 0:
           return
       if self.state.n_steps >= self.settings.planning_exploration_limit:
           msg = (
               'PLANNING NUDGE: You have taken '
               f'{self.state.n_steps} steps without creating a plan. '
               'If the task is complex, output a `plan_update` with clear todo items now. '
               'If the task is already done or nearly done, call `done` instead.'
           )
           logger.info(
               f'Exploration nudge injected after {self.state.n_steps} steps without a plan'
           )
           self._message_manager._add_context_message(UserMessage(content=msg))
   ```

6. **Loop Detection**
   ```python
   @dataclass
   class LoopDetector:
       max_repetition_count: int = 0
       consecutive_stagnant_pages: int = 0
       last_page_hashes: list[str] = field(default_factory=list)

       def get_nudge_message(self) -> str | None:
           """Get nudge message if loop detected."""
           if self.max_repetition_count >= 3:
               return (
                   f'LOOP DETECTED: You have repeated the same actions '
                   f'{self.max_repetition_count} times. '
                   'Try a different approach or call `done` with your findings.'
               )
           if self.consecutive_stagnant_pages >= 3:
               return (
                   f'STAGNATION DETECTED: You have been on the same page '
                   f'{self.consecutive_stagnant_pages} steps without progress. '
                   'Try navigating to a different page or call `done`.'
               )
           return None

       def update(self, actions: list[dict], page_hash: str) -> None:
           """Update loop detector with current state."""
           # Check for action repetition
           if self._are_actions_repetitive(actions):
               self.max_repetition_count += 1
           else:
               self.max_repetition_count = 0

           # Check for page stagnation
           if page_hash in self.last_page_hashes:
               self.consecutive_stagnant_pages += 1
           else:
               self.consecutive_stagnant_pages = 0
               self.last_page_hashes.append(page_hash)
               if len(self.last_page_hashes) > 5:
                   self.last_page_hashes.pop(0)
   ```

### Implementation Details

**Structured Reasoning Output**:
```python
@dataclass
class AgentOutput:
    thinking: str  # Structured reasoning block
    evaluation_previous_goal: str  # Success/failure analysis
    memory: str  # Progress tracking
    next_goal: str  # Immediate next action
    current_plan_item: int | None  # Current plan index
    plan_update: list[str] | None  # New plan steps
    action: list[dict]  # Actions to execute
```

**Prompt Template with Planning**:
```markdown
<planning>
Decide whether to plan based on task complexity:
- Simple task (1-3 actions): Act directly. Do NOT output `plan_update`.
- Complex but clear task: Output `plan_update` immediately with 3-10 todo items.
- Complex and unclear task: Explore for a few steps first, then output `plan_update`.

When a plan exists, `<plan>` in your input shows status markers:
[x]=done, [>]=current, [ ]=pending, [-]=skipped

Output `current_plan_item` to indicate which item you are working on.
Output `plan_update` only to revise the plan after unexpected obstacles.
</planning>

<reasoning_rules>
You must reason explicitly and systematically at every step in your `thinking` block:
- Reason about <agent_history> to track progress toward <user_request>
- Analyze the most recent "Next Goal" and "Action Result"
- Explicitly judge success/failure/uncertainty of the last action
- Analyze whether you are stuck (repeating actions without progress)
- Decide what concise context should be stored in memory
- Before done, verify all requirements against original <user_request>
</reasoning_rules>
```

**Skyvern's Task History Approach**:
```python
@dataclass
class Task:
    task_id: str
    navigation_goal: str
    task_history: list[dict]  # History of completed tasks
    status: TaskStatus

# Task history includes:
# - task type (navigate, extract, loop)
# - completion status
# - extracted information
# - failure reasons
```

### Benefits

1. **Task Decomposition**: Breaks complex tasks into manageable steps
2. **Progress Tracking**: Clear visibility into task completion
3. **Adaptive Replanning**: Adjusts plans when obstacles arise
4. **Loop Prevention**: Detects and prevents infinite loops
5. **Memory Management**: Maintains context across steps

### Trade-offs

1. **Complexity**: Adds significant complexity to agent logic
2. **Overhead**: Plan management adds computational overhead
3. **Planning Errors**: Incorrect plans can lead to wasted effort
4. **Nudge Sensitivity**: May trigger replanning unnecessarily

### Variations

**browser-use Approach**:
- Explicit plan with status markers ([x], [>], [ ])
- Replan nudges on consecutive failures
- Exploration nudges for long-running tasks without plans
- Loop detection with action repetition and page stagnation
- Structured reasoning output (thinking, evaluation, memory, next_goal)

**Skyvern Approach**:
- Task history with completion status
- Task types (navigate, extract, loop)
- Loop tasks for parallel execution
- Less explicit plan structure
- More focused on task-level planning

### When to Use

- Building AI agents for complex multi-step tasks
- Tasks requiring progress tracking and adaptation
- When preventing loops and stagnation is critical
- When clear task decomposition improves success rates
- Building agents that need to handle obstacles gracefully

### When Not to Use

- Simple single-step tasks
- When task structure is unpredictable
- Quick prototypes without production requirements
- When planning overhead is unacceptable

### Related Patterns

- `speculative-execution-next-step-planning`
- `llm-error-classification`
- `fallback-llm-system`

## Implementation Checklist

- [ ] Define plan state data structures
- [ ] Implement plan creation and update logic
- [ ] Add plan rendering for context injection
- [ ] Implement replan nudges on failures
- [ ] Add exploration nudges for long-running tasks
- [ ] Implement loop detection
- [ ] Add structured reasoning output
- [ ] Create planning prompt templates
- [ ] Test with various task complexities
- [ ] Monitor plan success rate and replan frequency

## References

- browser-use: `browser_use/agent/service.py` (lines 1405-1499)
- Skyvern: `skyvern/forge/agent.py` (task history, loop tasks)
