# Skyvern LLM Integration

**Repository**: Skyvern  
**Location**: `skyvern/forge/sdk/api/llm/`  
**Key Files**:
- `api_handler_factory.py` (2,421 lines) - LLM handler factory with LiteLLM
- `api_handler.py` (49 lines) - LLM handler protocol
- `config_registry.py` - LLM configuration registry
- `exceptions.py` - LLM exceptions
- `models.py` - LLM configuration models
- `utils.py` - LLM utilities

---

## Overview

Skyvern's LLM Integration uses LiteLLM for multi-provider support with 20+ specialized LLM handlers for different use cases. It supports OpenAI, Anthropic, Google, Groq, and other providers with unified API, streaming support, and comprehensive error handling.

---

## LLM Handler Protocol

**Location**: `api_handler.py` (lines 9-28)

```python
class LLMAPIHandler(Protocol):
    def __call__(
        self,
        prompt: str,
        prompt_name: str,
        step: Step | None = None,
        task_v2: TaskV2 | None = None,
        thought: Thought | None = None,
        ai_suggestion: AISuggestion | None = None,
        workflow_run_block_id: str | None = None,
        screenshots: list[bytes] | None = None,
        parameters: dict[str, Any] | None = None,
        organization_id: str | None = None,
        tools: list | None = None,
        use_message_history: bool = False,
        raw_response: bool = False,
        window_dimension: Resolution | None = None,
        force_dict: bool = True,
        system_prompt: str | None = None,
    ) -> Awaitable[dict[str, Any] | Any]: ...
```

**Parameters**:
- `prompt` - The prompt to send to LLM
- `prompt_name` - Name of the prompt for tracking
- `step` - Step context (optional)
- `task_v2` - Task v2 context (optional)
- `screenshots` - Screenshots for vision models
- `parameters` - Additional parameters
- `organization_id` - Organization for rate limiting
- `tools` - Tools for function calling
- `use_message_history` - Whether to use message history
- `system_prompt` - Override system prompt

---

## LLM Handler Factory

**Location**: `api_handler_factory.py` (2,421 lines)

### LiteLLM Integration

**Purpose**: Multi-provider LLM support

**Supported Providers**:
- OpenAI
- Anthropic (Claude)
- Google (Gemini)
- Groq
- Azure OpenAI
- AWS Bedrock
- VolcEngine
- Custom providers

### LLMCaller

**Purpose**: Main LLM caller with retry logic and error handling

**Features**:
- Automatic retry with exponential backoff
- Rate limit handling
- Token usage tracking
- Cost calculation
- Streaming support
- Cache integration

### LLMCallerManager

**Purpose**: Manage multiple LLM callers with routing

**Features**:
- Load balancing across providers
- Fallback on failure
- Cost optimization
- Performance optimization

---

## Configuration Registry

**Location**: `config_registry.py`

**Purpose**: Registry for LLM configurations

**Features**:
- Configuration validation
- Default values
- Environment variable mapping
- Model aliases

---

## LLM Exceptions

**Location**: `exceptions.py`

### Exception Types

```python
class LLMProviderError(Exception):
    """Base LLM provider error"""

class LLMProviderErrorRetryableTask(LLMProviderError):
    """Retryable LLM error for specific task types"""

class InvalidLLMConfigError(Exception):
    """Invalid LLM configuration"""

class DuplicateCustomLLMProviderError(Exception):
    """Duplicate custom LLM provider"""
```

### Retryable Task Types

**Location**: `api_handler_factory.py`

```python
LLM_PROVIDER_ERROR_RETRYABLE_TASK_TYPE = {
    "extract-actions",  # Retryable for action extraction
    # ... other retryable task types
}
```

---

## LLM Models

**Location**: `models.py`

### LLMConfig

```python
class LLMConfig(BaseModel):
    model: str
    api_key: str | None
    api_base: str | None
    temperature: float | None
    max_tokens: int | None
    # ... other config
```

### LLMRouterConfig

```python
class LLMRouterConfig(BaseModel):
    model_list: list[LLMConfig]
    routing_strategy: str | None
    allowed_fails_policy: AllowedFailsPolicy | None
```

---

## Screenshot Handling

**Location**: `api_handler_factory.py` (line 53)

### Resolution

```python
from skyvern.utils.image_resizer import Resolution, get_resize_target_dimension, resize_screenshots
```

**Supported Resolutions**:
- XGA: 1024x768
- WXGA: 1280x800
- FWXGA: 1366x768

**Purpose**: Resize screenshots for vision models to reduce cost

---

## Thinking Budget

**Location**: `api_handler_factory.py` (lines 66-68)

```python
EXTRACT_ACTION_DEFAULT_THINKING_BUDGET = settings.EXTRACT_ACTION_THINKING_BUDGET
DEFAULT_THINKING_BUDGET = settings.DEFAULT_THINKING_BUDGET
```

**Purpose**: Control reasoning token usage for Claude thinking

---

## OpenTelemetry Integration

**Location**: `api_handler_factory.py` (lines 71-100)

### Span Enrichment

```python
def _enrich_llm_span(
    span: otel_trace.Span,
    *,
    model: str,
    prompt_name: str,
    prompt_tokens: int,
    completion_tokens: int,
    reasoning_tokens: int = 0,
    cached_tokens: int = 0,
    latency_ms: int,
    llm_cost: float = 0.0,
) -> None:
```

**Attributes**:
- `llm_model` - Model name
- `prompt_tokens` - Input tokens
- `completion_tokens` - Output tokens
- `reasoning_tokens` - Claude thinking tokens
- `cached_tokens` - Cached tokens
- `latency_ms` - Request latency
- `llm_cost` - Cost in USD
- `cache_hit` - Whether cache was hit

---

## Message Building

**Location**: `utils.py`

### llm_messages_builder

**Purpose**: Build LLM messages from prompt and screenshots

**Features**:
- Image encoding for vision models
- Text message construction
- Multi-modal message building

### llm_messages_builder_with_history

**Purpose**: Build messages with conversation history

**Features**:
- History integration
- Context preservation
- Token budget management

---

## Key Patterns

### 1. Multi-Provider Support

**Pattern**: Use LiteLLM for unified API across providers

**Benefits**:
- **Provider flexibility**: Easy to switch providers
- **Cost optimization**: Choose cheapest provider
- **Reliability**: Fallback between providers
- **Unified interface**: Same API regardless of provider

### 2. Specialized Handlers

**Pattern**: 20+ specialized LLM handlers for different use cases

**Benefits**:
- **Cost optimization**: Use cheaper models for simple tasks
- **Performance**: Use faster models for latency-sensitive tasks
- **Quality**: Use best models for critical tasks
- **Flexibility**: Easy to tune per use case

### 3. Retry Logic

**Pattern**: Automatic retry with exponential backoff

**Benefits**:
- **Reliability**: Handle transient failures
- **Cost optimization**: Retry only when beneficial
- **Task-specific**: Different retry logic per task type

### 4. Screenshot Resizing

**Pattern**: Resize screenshots for vision models

**Benefits**:
- **Cost reduction**: Smaller images cost less
- **Performance**: Faster processing
- **Consistency**: Standard resolution across requests

### 5. Token Budget Management

**Pattern**: Track and manage token usage

**Benefits**:
- **Cost control**: Monitor and limit costs
- **Performance**: Optimize for token efficiency
- **Observability**: Track usage patterns

### 6. OpenTelemetry Integration

**Pattern**: Enrich spans with LLM metadata

**Benefits**:
- **Observability**: Track LLM performance
- **Cost tracking**: Monitor LLM costs
- **Debugging**: Trace LLM requests

---

## Testing Considerations

### Test Scenarios

1. **Multi-provider support** - Verify all providers work
2. **Retry logic** - Verify retry behavior
3. **Screenshot handling** - Verify image encoding
4. **Token tracking** - Verify accurate token counting
5. **Error handling** - Verify error recovery

---

## References

- **API Handler Factory**: `forge/sdk/api/llm/api_handler_factory.py` (2,421 lines)
- **API Handler**: `forge/sdk/api/llm/api_handler.py`
- **Config Registry**: `forge/sdk/api/llm/config_registry.py`
- **Exceptions**: `forge/sdk/api/llm/exceptions.py`
- **Models**: `forge/sdk/api/llm/models.py`
- **Utils**: `forge/sdk/api/llm/utils.py`
