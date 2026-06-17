# Loop Detection and Replan Nudges

**Pattern ID**: `loop-detection-replan-nudges`
**Source**: browser-use (browser_use/agent/service.py)
**Category**: AI Agent, Reasoning
**Complexity**: Intermediate

## Problem

AI agents can get stuck in:
- Infinite loops (repeating the same actions)
- Stagnation (staying on the same page without progress)
- Repetitive failures (trying the same approach multiple times)

Without detection, agents waste resources and never complete tasks.

## Solution

Implement loop detection with escalating nudges to force replanning:

### Core Components

1. **Loop Detector State**
   ```python
   @dataclass
   class LoopDetector:
       max_repetition_count: int = 0
       consecutive_stagnant_pages: int = 0
       last_page_hashes: list[str] = field(default_factory=list)
       last_actions: list[dict] = field(default_factory=list)

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
   ```

2. **Action Repetition Detection**
   ```python
   def _are_actions_repetitive(
       self,
       actions: list[dict],
       threshold: int = 3
   ) -> bool:
       """Check if actions are repetitive."""
       if not actions:
           return False

       # Add current actions to history
       self.last_actions.append(actions)
       if len(self.last_actions) > threshold:
           self.last_actions.pop(0)

       # Check if last N action sets are identical
       if len(self.last_actions) >= threshold:
           recent_actions = self.last_actions[-threshold:]
           return all(
               self._actions_equal(recent_actions[0], other)
               for other in recent_actions[1:]
           )

       return False

   def _actions_equal(self, actions1: list[dict], actions2: list[dict]) -> bool:
       """Check if two action lists are equal."""
       if len(actions1) != len(actions2):
           return False

       for a1, a2 in zip(actions1, actions2):
           # Compare action types and parameters
           if a1.keys() != a2.keys():
               return False
           for key in a1.keys():
               if key == 'text' and a1[key] != a2[key]:
                   # Allow text differences if they're similar
                   continue
               if a1[key] != a2[key]:
                   return False

       return True
   ```

3. **Page Stagnation Detection**
   ```python
   def update_page_state(self, page_hash: str) -> None:
       """Update loop detector with current page state."""
       if page_hash in self.last_page_hashes:
           self.consecutive_stagnant_pages += 1
       else:
           self.consecutive_stagnant_pages = 0
           self.last_page_hashes.append(page_hash)

           # Keep only last 5 page hashes
           if len(self.last_page_hashes) > 5:
               self.last_page_hashes.pop(0)

   def compute_page_hash(self, page_content: str) -> str:
       """Compute hash of page content for stagnation detection."""
       import hashlib
       return hashlib.md5(page_content.encode()).hexdigest()
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

### Implementation Details

**Loop Detector Integration**:
```python
class Agent:
    def __init__(self, config: AgentConfig):
        self.state = AgentState()
        self.loop_detector = LoopDetector()
        self.settings = config

    async def step(self) -> None:
        """Execute agent step with loop detection."""
        # Execute actions
        actions = await self._get_actions_from_llm()
        await self._execute_actions(actions)

        # Update loop detector
        self.loop_detector.update(actions, self.current_page_hash)

        # Check for loop detection
        nudge = self.loop_detector.get_nudge_message()
        if nudge:
            logger.warning(f"Loop detection nudge: {nudge}")
            self._message_manager._add_context_message(UserMessage(content=nudge))

        # Inject replan nudge if needed
        self._inject_replan_nudge()

        # Inject exploration nudge if needed
        self._inject_exploration_nudge()
```

**Escalating Nudge Strategy**:
```python
class LoopDetector:
    def get_escalating_nudge(self) -> str | None:
        """Get escalating nudge message based on severity."""
        if self.max_repetition_count >= 5:
            return (
                'CRITICAL LOOP: You have repeated the same actions '
                f'{self.max_repetition_count} times. '
                'STOP and try a completely different approach, or call `done` immediately.'
            )
        elif self.max_repetition_count >= 3:
            return (
                f'LOOP DETECTED: You have repeated the same actions '
                f'{self.max_repetition_count} times. '
                'Try a different approach or call `done` with your findings.'
            )
        elif self.consecutive_stagnant_pages >= 5:
            return (
                'CRITICAL STAGNATION: You have been stuck on the same page '
                f'{self.consecutive_stagnant_pages} steps. '
                'Navigate away or call `done` immediately.'
            )
        elif self.consecutive_stagnant_pages >= 3:
            return (
                f'STAGNATION DETECTED: You have been on the same page '
                f'{self.consecutive_stagnant_pages} steps without progress. '
                'Try navigating to a different page or call `done`.'
            )
        return None
```

**Configuration**:
```python
@dataclass
class AgentConfig:
    enable_planning: bool = True
    planning_replan_on_stall: int = 3  # Replan after N consecutive failures
    planning_exploration_limit: int = 10  # Require plan after N steps
    loop_detection_enabled: bool = True
    loop_repetition_threshold: int = 3
    loop_stagnation_threshold: int = 3
```

### Benefits

1. **Resource Efficiency**: Prevents wasted computation on loops
2. **Task Completion**: Forces agent to try new approaches
3. **User Experience**: Faster failure detection and recovery
4. **Monitoring**: Track loop patterns for analytics
5. **Adaptive Behavior**: Agent self-corrects when stuck

### Trade-offs

1. **False Positives**: May detect loops that are intentional
2. **Over-Nudging**: May interrupt valid repetitive actions
3. **Complexity**: Adds state tracking and detection logic
4. **Tuning**: Thresholds require careful tuning per use case

### Variations

**browser-use Approach**:
- Action repetition detection (exact match)
- Page stagnation detection (hash-based)
- Replan nudges on consecutive failures
- Exploration nudges for long-running tasks without plans
- Escalating nudge messages

**Skyvern Approach**:
- Less explicit loop detection
- Focus on task-level loop detection
- LLM error classification for failure patterns
- Speculative execution to avoid loops

### When to Use

- Building AI agents for complex, unpredictable tasks
- When preventing infinite loops is critical
- Tasks with potential for repetitive actions
- When resource efficiency is important
- Building agents that need to self-correct

### When Not to Use

- Simple tasks with predictable behavior
- When repetitive actions are intentional
- Quick prototypes without production requirements
- When loop detection adds unacceptable overhead

### Related Patterns

- `multi-step-reasoning-planning-memory`
- `llm-error-classification`
- `fallback-llm-system`

## Implementation Checklist

- [ ] Define loop detector state structure
- [ ] Implement action repetition detection
- [ ] Add page stagnation detection with hashing
- [ ] Implement replan nudge logic
- [ ] Add exploration nudge logic
- [ ] Create escalating nudge strategy
- [ ] Add configuration thresholds
- [ ] Integrate loop detector into agent step
- [ ] Test with various loop scenarios
- [ ] Monitor loop detection rate in production

## References

- browser-use: `browser_use/agent/service.py` (lines 1484-1499)
