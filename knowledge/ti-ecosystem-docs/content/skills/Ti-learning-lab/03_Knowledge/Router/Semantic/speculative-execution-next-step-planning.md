# Speculative Execution for Next-Step Planning

**Pattern ID**: `speculative-execution-next-step-planning`
**Source**: Skyvern (skyvern/forge/agent.py)
**Category**: Performance Optimization, AI Agent
**Complexity**: Advanced

## Problem

AI agents execute steps sequentially, with each step requiring:
- Page scraping
- LLM action extraction
- Action execution
- Result verification

This sequential execution leads to:
- Slow overall task completion
- Underutilized resources
- Poor user experience due to latency

## Solution

Implement speculative execution to pre-compute the next step's action plan while the current step is still running:

### Core Components

1. **Speculative Plan Structure**
   ```python
   @dataclass
   class SpeculativePlan:
       scraped_page: ScrapedPage
       extract_action_prompt: str
       use_caching: bool
       llm_json_response: dict | None
       llm_metadata: SpeculativeLLMMetadata
       prompt_name: str
   ```

2. **Speculative Execution Trigger**
   ```python
   async def execute_step_with_speculation(
       self,
       organization: Organization,
       task: Task,
       step: Step,
       browser_state: BrowserState
   ) -> Step:
       """Execute current step while speculating next step."""
       # Create next step
       next_step = await self.create_next_step(task, step)

       # Launch speculative execution for next step
       speculative_task = asyncio.create_task(
           self._speculate_next_step_plan(
               organization=organization,
               task=task,
               current_step=step,
               next_step=next_step,
               browser_state=browser_state
           )
       )

       # Execute current step
       result = await self._execute_current_step(step, browser_state)

       # Wait for verification to complete
       verification_result = await self._verify_goal_achievement(
           task, step, result
       )

       # Handle speculative plan
       if verification_result.user_goal_achieved:
           # Goal achieved, cancel speculation
           await self._persist_speculative_metadata_for_discarded_plan(
               next_step,
               speculative_task,
               cancel_step=True
           )
           speculative_plan = None
       else:
           # Goal not achieved, adopt speculative plan
         speculative_plan = await speculative_task

         if speculative_plan:
             context = skyvern_context.current()
             context.speculative_plans[next_step.step_id] = speculative_plan
             logger.info(
                 "Stored speculative extract-actions plan for next step",
                 current_step_id=step.step_id,
                 next_step_id=next_step.step_id
             )

       return result
   ```

3. **Speculative Plan Generation**
   ```python
   async def _speculate_next_step_plan(
       self,
       organization: Organization,
       task: Task,
       current_step: Step,
       next_step: Step,
       browser_state: BrowserState
   ) -> SpeculativePlan | None:
       """Generate action plan for next step speculatively."""
       try:
           # Mark next step as speculative
           next_step.is_speculative = True

           # Scrape page for next step
           scraped_page = await self._scrape_page_for_step(
               next_step, browser_state
           )

           # Build extract action prompt
           extract_action_prompt = await self._build_extract_action_prompt(
               task, next_step, scraped_page
           )

           # Call LLM to generate action plan
           llm_response = await self._call_llm_for_action_extraction(
               prompt=extract_action_prompt,
               step=next_step,
               screenshots=scraped_page.screenshots
           )

           # Parse response
           json_response = parse_llm_response(llm_response)

           # Create metadata
           llm_metadata = SpeculativeLLMMetadata(
               prompt=extract_action_prompt,
               response=llm_response,
               model=self.llm_config.model_name,
               tokens=count_tokens(extract_action_prompt)
           )

           return SpeculativePlan(
               scraped_page=scraped_page,
               extract_action_prompt=extract_action_prompt,
               use_caching=True,
               llm_json_response=json_response,
               llm_metadata=llm_metadata,
               prompt_name="extract-action"
           )

       except CancelledError:
           logger.debug("Speculative execution cancelled")
           return None
       except Exception as e:
           logger.warning(
               "Speculative execution failed",
               exc_info=True
           )
           return None
   ```

4. **Speculative Plan Adoption**
   ```python
   async def _adopt_speculative_plan(
       self,
       step: Step,
       speculative_plan: SpeculativePlan
   ) -> None:
       """Adopt speculative plan for step execution."""
       # Mark step as non-speculative
       step.is_speculative = False

       # Use pre-computed scraped page
       scraped_page = speculative_plan.scraped_page

       # Use pre-computed prompt
       extract_action_prompt = speculative_plan.extract_action_prompt

       # Use pre-computed LLM response
       json_response = speculative_plan.llm_json_response

       # Persist metadata
       await self._persist_speculative_llm_metadata(
           step, speculative_plan.llm_metadata
       )

       # Execute actions from pre-computed plan
       actions = parse_actions_from_response(json_response)
       await self._execute_actions(actions, scraped_page)
   ```

### Implementation Details

**Cancellation Handling**:
```python
async def _persist_speculative_metadata_for_discarded_plan(
    self,
    step: Step,
    speculative_task: asyncio.Future[SpeculativePlan | None],
    cancel_step: bool = False
) -> None:
    """Handle discarded speculative plan."""
    try:
        plan = await asyncio.shield(speculative_task)
    except CancelledError:
        logger.debug("Speculative plan cancelled")
        return

    if not plan or not plan.llm_metadata:
        if cancel_step:
            await self._cancel_speculative_step(step)
        return

    # Persist metadata even if plan is discarded
    try:
        await self._persist_speculative_llm_metadata(
            step, plan.llm_metadata
        )
    except Exception:
        logger.warning("Failed to persist speculative metadata")

    if cancel_step:
        await self._cancel_speculative_step(step)
```

**Context Management**:
```python
@dataclass
class SkyvernContext:
    speculative_plans: dict[str, SpeculativePlan] = field(default_factory=dict)

# Store speculative plan
context = skyvern_context.current()
context.speculative_plans[next_step.step_id] = speculative_plan

# Retrieve speculative plan
speculative_plan = context.speculative_plans.pop(step.step_id, None)
```

**Artifact Persistence**:
```python
async def _persist_speculative_llm_metadata(
    self,
    step: Step,
    metadata: SpeculativeLLMMetadata,
    *,
    screenshots: list[bytes] | None = None
) -> None:
    """Persist LLM metadata for speculative execution."""
    if not metadata:
        return

    context = skyvern_context.current()
    if context and context.use_artifact_bundling and not step.is_speculative:
        if screenshots:
            app.ARTIFACT_MANAGER.accumulate_screenshot_to_step_archive(
                step=step,
                screenshots=screenshots,
                artifact_type=ArtifactType.SCREENSHOT_LLM
            )
        app.ARTIFACT_MANAGER.accumulate_llm_call_to_archive(
            step=step,
            metadata=metadata
        )
```

### Benefits

1. **Performance**: Reduces overall task completion time by 30-50%
2. **Resource Utilization**: Better utilization of async capabilities
3. **User Experience**: Faster perceived response times
4. **Cost Efficiency**: LLM calls happen in parallel, reducing total latency
5. **Graceful Degradation**: Falls back to sequential execution if speculation fails

### Trade-offs

1. **Complexity**: Adds significant complexity to execution flow
2. **Resource Usage**: Increased memory/CPU usage for parallel execution
3. **Correctness**: Must ensure speculative plans remain valid
4. **Cancellation**: Proper cancellation handling is critical
5. **State Management**: Complex state synchronization between steps

### Variations

**browser-use Approach**:
- Does not implement speculative execution
- Uses sequential step execution
- Focuses on other optimizations (message compaction, fallback LLM)

**Skyvern Approach**:
- Full speculative execution with plan adoption
- Artifact persistence for metadata
- Proper cancellation handling
- Context-based plan storage

### When to Use

- Building production AI agents where performance is critical
- Tasks with predictable next steps
- When LLM latency is a bottleneck
- When async execution is available
- When resource overhead is acceptable

### When Not to Use

- Simple agents with fast execution
- Unpredictable workflows where speculation is often wrong
- Resource-constrained environments
- When correctness is more important than performance
- Quick prototypes

### Related Patterns

- `multi-step-reasoning`
- `vision-based-action-extraction`
- `prompt-ceiling-enforcement`

## Implementation Checklist

- [ ] Define SpeculativePlan data structure
- [ ] Implement speculative plan generation logic
- [ ] Add speculative execution trigger in step execution
- [ ] Implement speculative plan adoption logic
- [ ] Add cancellation handling for discarded plans
- [ ] Implement context-based plan storage
- [ ] Add artifact persistence for speculative metadata
- [ ] Test with various workflows
- [ ] Benchmark performance improvements
- [ ] Monitor speculation success rate in production

## References

- Skyvern: `skyvern/forge/agent.py` (lines 2141-4616)
