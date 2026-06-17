# browser-use Agent Step Execution Pattern

**Pattern ID**: `browser-use-agent-step-execution`
**Source**: browser-use (browser_use/agent/service.py)
**Category**: AI Agent, Execution Flow
**Complexity**: Advanced

## Problem

AI agents need a robust, repeatable execution pattern that:
- Handles browser state changes reliably
- Manages LLM interactions with retry logic
- Executes actions safely with error recovery
- Tracks progress across steps
- Prevents infinite loops
- Manages token budget efficiently

## Solution

Implement a 3-phase step execution pattern with comprehensive error handling and state management:

### Phase 1: Prepare Context

**Purpose**: Gather browser state, update action models, prepare LLM context

**Implementation**:
```python
async def step(self, step_info: AgentStepInfo | None = None) -> None:
    """Execute one step of the task"""
    try:
        # Phase 0: CAPTCHA wait (if enabled)
        if self.browser_session:
            captcha_wait = await self.browser_session.wait_if_captcha_solving()
            if captcha_wait and captcha_wait.waited:
                # Inject CAPTCHA result into agent context
                captcha_result = ActionResult(long_term_memory=msg)
                self.state.last_result = [captcha_result]

        # Phase 1: Prepare context
        browser_state_summary = await self._prepare_context(step_info)

        # Clear previous step state
        self.state.last_model_output = None
        self.state.last_result = None

        # Phase 2: Get model output and execute actions
        await self._get_next_action(browser_state_summary)
        await self._execute_actions()

        # Phase 3: Post-processing
        await self._post_process()

    except Exception as e:
        await self._handle_step_error(e)

    finally:
        await self._finalize(browser_state_summary)
```

**Context Preparation Details**:
```python
async def _prepare_context(self, step_info: AgentStepInfo | None = None) -> BrowserStateSummary:
    """Prepare context for the step"""
    # 1. Get browser state with screenshot
    browser_state_summary = await self.browser_session.get_browser_state_summary(
        include_screenshot=True,  # Always capture for cloud sync
        include_recent_events=self.include_recent_events,
    )

    # 2. Check for downloads
    await self._check_and_update_downloads('after getting browser state')

    # 3. Update action models with page-specific actions
    await self._update_action_models_for_page(browser_state_summary.url)

    # 4. Get page-specific filtered actions
    page_filtered_actions = self.tools.registry.get_prompt_description(
        browser_state_summary.url
    )

    # 5. Render plan description
    plan_description = self._render_plan_description()

    # 6. Prepare step state in message manager
    self._message_manager.prepare_step_state(
        browser_state_summary=browser_state_summary,
        model_output=self.state.last_model_output,
        result=self.state.last_result,
        step_info=step_info,
        sensitive_data=self.sensitive_data,
    )

    # 7. Maybe compact messages
    await self._maybe_compact_messages(step_info)

    # 8. Create state messages for LLM
    self._message_manager.create_state_messages(
        browser_state_summary=browser_state_summary,
        model_output=self.state.last_model_output,
        result=self.state.last_result,
        step_info=step_info,
        use_vision=self.settings.use_vision,
        page_filtered_actions=page_filtered_actions,
        sensitive_data=self.sensitive_data,
        available_file_paths=self.available_file_paths,
        plan_description=plan_description,
    )

    # 9. Inject nudges
    self._inject_replan_nudge()
    self._inject_exploration_nudge()
    self._inject_loop_detection_nudge()

    return browser_state_summary
```

### Phase 2: Get Action and Execute

**Purpose**: Call LLM to get next action, then execute actions

**LLM Call with Retry Logic**:
```python
async def _get_next_action(self, browser_state_summary: BrowserStateSummary) -> None:
    """Execute LLM interaction with retry logic"""
    input_messages = self._message_manager.get_messages()

    try:
        # Call LLM with timeout
        model_output = await asyncio.wait_for(
            self._get_model_output_with_retry(input_messages),
            timeout=self.settings.llm_timeout
        )
    except TimeoutError:
        raise TimeoutError(
            f'LLM call timed out after {self.settings.llm_timeout} seconds. '
            'Keep your thinking and output short.'
        )

    self.state.last_model_output = model_output

    # Handle callbacks and conversation saving
    await self._handle_post_llm_processing(browser_state_summary, input_messages)
```

**Retry Logic with Fallback LLM**:
```python
async def _get_model_output_with_retry(
    self,
    input_messages: list[BaseMessage]
) -> AgentOutput:
    """Get model output with retry logic and fallback LLM"""
    max_retries = 3

    for attempt in range(max_retries):
        try:
            # Call current LLM (primary or fallback)
            return await self.llm.ainvoke(
                messages=input_messages,
                output_format=self._get_output_format()
            )
        except (ModelRateLimitError, ModelProviderError) as e:
            # Check if fallback is available
            if e.status_code in [401, 402, 429, 500, 502, 503, 504]:
                if self._fallback_llm and not self._using_fallback_llm:
                    self._switch_to_fallback_llm()
                    continue  # Retry with fallback
            raise  # No fallback or other error
        except Exception as e:
            if attempt == max_retries - 1:
                raise
            # Exponential backoff
            await asyncio.sleep(2 ** attempt)

def _switch_to_fallback_llm(self) -> None:
    """Switch to fallback LLM"""
    self.llm = self._fallback_llm
    self._using_fallback_llm = True
    self.logger.warning(f'Switched to fallback LLM: {self.llm.model}')
```

**Action Execution**:
```python
async def _execute_actions(self) -> None:
    """Execute the actions from model output"""
    if self.state.last_model_output is None:
        raise ValueError('No model output to execute actions from')

    result = await self.multi_act(self.state.last_model_output.action)
    self.state.last_result = result
```

### Phase 3: Post-Processing

**Purpose**: Handle downloads, update plans, track loops, log results

**Implementation**:
```python
async def _post_process(self) -> None:
    """Handle post-action processing"""
    # 1. Check for new downloads
    await self._check_and_update_downloads('after executing actions')

    # 2. Update plan state from model output
    if self.state.last_model_output is not None:
        self._update_plan_from_model_output(self.state.last_model_output)

    # 3. Record actions for loop detection
    self._update_loop_detector_actions()

    # 4. Track consecutive failures
    if self.state.last_result and len(self.state.last_result) == 1:
        if self.state.last_result[-1].error:
            self.state.consecutive_failures += 1
            return

    # Reset consecutive failures on success
    if self.state.consecutive_failures > 0:
        self.state.consecutive_failures = 0

    # 5. Log completion results
    if self.state.last_result and self.state.last_result[-1].is_done:
        success = self.state.last_result[-1].success
        extracted_content = self.state.last_result[-1].extracted_content
        self.logger.info(f'Final Result: {extracted_content}')
```

### Error Handling

**Comprehensive Error Handler**:
```python
async def _handle_step_error(self, error: Exception) -> None:
    """Handle all types of errors during a step"""
    # Handle interruption (not an error)
    if isinstance(error, InterruptedError):
        self.logger.warning('The agent was interrupted mid-step')
        return

    # Handle connection errors with reconnection
    if self._is_connection_like_error(error):
        if self.browser_session.is_reconnecting:
            await self.browser_session._reconnect_event.wait()
            if self.browser_session.is_cdp_connected:
                self.logger.info('Reconnection succeeded, retrying step...')
                self.state.last_result = [ActionResult(
                    error=f'Connection lost and recovered: {error}'
                )]
                return

    # Handle browser closed
    if self._is_browser_closed_error(error):
        self.logger.warning(f'Browser closed or disconnected: {error}')
        self.state.stopped = True
        return

    # Handle all other exceptions
    max_total_failures = self.settings.max_failures + int(
        self.settings.final_response_after_failure
    )
    self.state.consecutive_failures += 1

    # Log error
    error_msg = AgentError.format_error(error)
    is_final_failure = self.state.consecutive_failures >= max_total_failures
    log_level = logging.ERROR if is_final_failure else logging.WARNING

    self.logger.log(log_level, f'Result failed: {error_msg}')

    # Set error result
    self.state.last_result = [ActionResult(error=error_msg)]
```

### Nudge Injection

**Replan Nudge**:
```python
def _inject_replan_nudge(self) -> None:
    """Inject replan nudge when stall detected"""
    if not self.settings.enable_planning or self.state.plan is None:
        return
    if self.settings.planning_replan_on_stall <= 0:
        return

    if self.state.consecutive_failures >= self.settings.planning_replan_on_stall:
        msg = (
            f'REPLAN SUGGESTED: You have failed '
            f'{self.state.consecutive_failures} consecutive times. '
            'Your current plan may need revision. '
            'Output a new `plan_update` with revised steps to recover.'
        )
        self._message_manager._add_context_message(UserMessage(content=msg))
```

**Exploration Nudge**:
```python
def _inject_exploration_nudge(self) -> None:
    """Nudge agent to create plan after exploring without one"""
    if not self.settings.enable_planning or self.state.plan is not None:
        return
    if self.settings.planning_exploration_limit <= 0:
        return

    if self.state.n_steps >= self.settings.planning_exploration_limit:
        msg = (
            f'PLANNING NUDGE: You have taken '
            f'{self.state.n_steps} steps without creating a plan. '
            'If the task is complex, output a `plan_update` with clear todo items now. '
            'If the task is already done or nearly done, call `done` instead.'
        )
        self._message_manager._add_context_message(UserMessage(content=msg))
```

**Loop Detection Nudge**:
```python
def _inject_loop_detection_nudge(self) -> None:
    """Inject loop detection nudge"""
    if not self.settings.loop_detection_enabled:
        return

    nudge = self.state.loop_detector.get_nudge_message()
    if nudge:
        self.logger.warning(f'Loop detection nudge: {nudge}')
        self._message_manager._add_context_message(UserMessage(content=nudge))
```

### Benefits

1. **Robustness**: Comprehensive error handling for all failure modes
2. **Recovery**: Automatic fallback LLM on rate limits
3. **Adaptability**: Nudges force agent to try new approaches
4. **Efficiency**: Message compaction manages token budget
5. **Reliability**: Connection error handling with reconnection
6. **Progress Tracking**: Clear step-by-step execution with state management

### Trade-offs

1. **Complexity**: 3-phase pattern adds complexity
2. **Overhead**: Multiple phases add execution overhead
3. **State Management**: Complex state synchronization
4. **Nudge Sensitivity**: May interrupt valid repetitive actions

### When to Use

- Building production AI agents with complex execution flows
- When robust error handling is critical
- Tasks requiring multi-step reasoning
- When preventing loops is important
- Building agents that need to self-correct

### When Not to Use

- Simple single-step tasks
- When execution overhead is unacceptable
- Quick prototypes without production requirements
- When all actions are guaranteed to succeed

### Related Patterns

- `multi-step-reasoning-planning-memory`
- `fallback-llm-system`
- `loop-detection-replan-nudges`
- `message-compaction-token-budget`

## Implementation Checklist

- [ ] Implement 3-phase step execution (prepare, execute, post-process)
- [ ] Add LLM call with retry logic and timeout
- [ ] Implement fallback LLM switching
- [ ] Add comprehensive error handler
- [ ] Implement connection error handling with reconnection
- [ ] Add replan nudge on consecutive failures
- [ ] Add exploration nudge for long-running tasks
- [ ] Implement loop detection and nudges
- [ ] Add message compaction in prepare phase
- [ ] Test with various error scenarios

## References

- browser-use: `browser_use/agent/service.py` (lines 1023-1321)
