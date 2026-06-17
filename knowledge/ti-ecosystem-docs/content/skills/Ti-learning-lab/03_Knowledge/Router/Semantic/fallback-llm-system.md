# Fallback LLM System for Rate Limits

**Pattern ID**: `fallback-llm-system`
**Source**: browser-use (browser_use/agent/service.py)
**Category**: Error Handling, AI Agent
**Complexity**: Intermediate

## Problem

AI agents using LLM APIs encounter:
- Rate limit errors (429)
- Server errors (500, 502, 503, 504)
- Provider outages
- Authentication failures (401, 402)

Without a fallback mechanism, these errors cause task failures and poor user experience.

## Solution

Implement a fallback LLM system that automatically switches to a backup provider when the primary fails:

### Core Components

1. **Fallback LLM Configuration**
   ```python
   @dataclass
   class AgentConfig:
       llm: BaseChatModel
       fallback_llm: BaseChatModel | None = None
       max_retries: int = 3
       retry_delay: float = 1.0

   class Agent:
       def __init__(
           self,
           task: str,
           llm: BaseChatModel,
           fallback_llm: BaseChatModel | None = None
       ):
           self.task = task
           self._original_llm = llm
           self._fallback_llm = fallback_llm
           self._current_llm = llm
           self._using_fallback_llm = False
   ```

2. **LLM Error Detection**
   ```python
   class ModelRateLimitError(Exception):
       """Raised when LLM API returns rate limit error."""
       def __init__(self, message: str, status_code: int, model: str):
           self.message = message
           self.status_code = status_code
           self.model = model
           super().__init__(message)

   class ModelProviderError(Exception):
       """Raised when LLM API returns server error."""
       def __init__(self, message: str, status_code: int, model: str):
           self.message = message
           self.status_code = status_code
           self.model = model
           super().__init__(message)

   FALLBACK_STATUS_CODES = [401, 402, 429, 500, 502, 503, 504]
   ```

3. **Fallback Trigger Logic**
   ```python
   async def step(self) -> None:
       """Execute agent step with fallback LLM."""
       try:
           await self._execute_step_with_current_llm()
       except (ModelRateLimitError, ModelProviderError) as e:
           if e.status_code in FALLBACK_STATUS_CODES and \
              self._fallback_llm is not None and \
              not self._using_fallback_llm:
               logger.warning(
                   f"Primary LLM failed with status {e.status_code}, "
                   f"switching to fallback LLM"
               )
               self._switch_to_fallback_llm()
               # Retry step with fallback
               await self._execute_step_with_current_llm()
           else:
               raise

   def _switch_to_fallback_llm(self) -> None:
       """Switch from primary to fallback LLM."""
       self._current_llm = self._fallback_llm
       self._using_fallback_llm = True
       logger.info(
           f"Switched to fallback LLM: {self._fallback_llm.model}"
       )
   ```

4. **Retry Logic with Exponential Backoff**
   ```python
   async def _execute_step_with_current_llm(self) -> None:
       """Execute step with retry logic."""
       max_retries = self.config.max_retries
       retry_delay = self.config.retry_delay

       for attempt in range(max_retries):
           try:
               return await self._current_llm.ainvoke(
                   messages=self._build_messages(),
                   output_format=self._get_output_format()
               )
           except (ModelRateLimitError, ModelProviderError) as e:
               if attempt == max_retries - 1:
                   raise  # Last attempt failed

               # Exponential backoff
               wait_time = retry_delay * (2 ** attempt)
               logger.warning(
                   f"LLM call failed (attempt {attempt + 1}/{max_retries}), "
                   f"retrying in {wait_time}s"
               )
               await asyncio.sleep(wait_time)
   ```

### Implementation Details

**Mock LLM for Testing**:
```python
def create_mock_llm(
    model_name: str = 'mock-llm',
    should_fail: bool = False,
    fail_with: type[Exception] | None = None,
    fail_status_code: int = 429,
    fail_message: str = 'Rate limit exceeded'
) -> BaseChatModel:
    """Create a mock LLM for testing fallback logic."""
    llm = AsyncMock(spec=BaseChatModel)
    llm.model = model_name
    llm.provider = 'mock'
    llm.name = model_name
    llm.model_name = model_name

    async def mock_ainvoke(*args, **kwargs):
        if should_fail:
            if fail_with == ModelRateLimitError:
                raise ModelRateLimitError(
                    message=fail_message,
                    status_code=fail_status_code,
                    model=model_name
                )
            elif fail_with == ModelProviderError:
                raise ModelProviderError(
                    message=fail_message,
                    status_code=fail_status_code,
                    model=model_name
                )
            else:
                raise Exception(fail_message)

        return ChatInvokeCompletion(
            completion=default_response,
            usage=None
        )

    llm.ainvoke.side_effect = mock_ainvoke
    return llm
```

**Fallback State Tracking**:
```python
@dataclass
class AgentState:
    _original_llm: BaseChatModel
    _fallback_llm: BaseChatModel | None
    _current_llm: BaseChatModel
    _using_fallback_llm: bool

    def reset_to_original_llm(self) -> None:
        """Reset to original LLM (e.g., after task completion)."""
        if self._using_fallback_llm:
            self._current_llm = self._original_llm
            self._using_fallback_llm = False
            logger.info("Reset to original LLM")

    @property
    def current_llm_name(self) -> str:
        """Get current LLM name."""
        return self._current_llm.model_name

    @property
    def is_using_fallback(self) -> bool:
        """Check if currently using fallback LLM."""
        return self._using_fallback_llm
```

**Usage Example**:
```python
# Create agent with fallback
primary_llm = ChatOpenAI(model="gpt-4")
fallback_llm = ChatOpenAI(model="gpt-3.5-turbo")

agent = Agent(
    task="Navigate to example.com and extract title",
    llm=primary_llm,
    fallback_llm=fallback_llm
)

# Run agent (will automatically switch to fallback on rate limits)
await agent.run()

# Check if fallback was used
if agent.is_using_fallback:
    logger.info(f"Task completed using fallback LLM: {agent.current_llm_name}")
```

### Benefits

1. **Reliability**: Automatic recovery from rate limits and server errors
2. **Cost Optimization**: Can use cheaper fallback model when primary is rate-limited
3. **User Experience**: Tasks continue instead of failing
4. **Flexibility**: Easy to configure different primary/fallback pairs
5. **Monitoring**: Track fallback usage for analytics

### Trade-offs

1. **Cost**: Fallback models may be more expensive or less capable
2. **Quality**: Fallback model may produce different/worse outputs
3. **Complexity**: Adds complexity to error handling
4. **Latency**: Retry logic adds delay on failures
5. **Configuration**: Requires managing multiple API keys

### Variations

**browser-use Approach**:
- Automatic switching on specific status codes (401, 402, 429, 500, 502, 503, 504)
- Single fallback LLM per agent
- Retry logic with exponential backoff
- State tracking for fallback usage

**Skyvern Approach**:
- LLM error classification (distinguishes LLM failures from runtime crashes)
- Retry logic at API handler level
- Less explicit fallback LLM configuration
- More comprehensive error handling

### When to Use

- Building production AI agents with reliability requirements
- Using rate-limited LLM APIs
- Tasks where completion is more important than optimal quality
- When fallback model quality is acceptable
- When managing multiple LLM providers

### When Not to Use

- Simple prototypes where failures are acceptable
- When fallback model quality is unacceptable
- When using unlimited rate limit tiers
- When cost is the primary concern

### Related Patterns

- `llm-error-classification`
- `multi-step-reasoning-planning-memory`
- `speculative-execution-next-step-planning`

## Implementation Checklist

- [ ] Define fallback LLM configuration
- [ ] Implement LLM error exceptions (RateLimitError, ProviderError)
- [ ] Add fallback trigger logic
- [ ] Implement retry logic with exponential backoff
- [ ] Add state tracking for fallback usage
- [ ] Create mock LLM for testing
- [ ] Add logging for fallback events
- [ ] Test with various error scenarios
- [ ] Monitor fallback usage in production
- [ ] Document fallback model quality differences

## References

- browser-use: `browser_use/agent/service.py`, `tests/ci/test_fallback_llm.py`
- Skyvern: `skyvern/forge/agent.py` (LLM error classification)
