# Prompt Ceiling Enforcement with Progressive Key Dropping

**Pattern ID**: `prompt-ceiling-enforcement`
**Source**: Skyvern (skyvern/forge/sdk/prompting.py)
**Category**: Prompt Engineering, Cost Optimization
**Complexity**: Intermediate

## Problem

LLM prompts can grow unbounded in complex workflows, leading to:
- Context window exceeded errors
- Excessive token costs
- Slow response times
- Rate limit failures

Agents need a systematic way to enforce prompt size limits while preserving the most important information.

## Solution

Implement a prompt ceiling enforcement system with progressive key dropping:

### Core Components

1. **Prompt Ceiling Definition**
   ```python
   PROMPT_CEILING_TOKENS = 180_000  # Hard ceiling
   PROMPT_WARNING_THRESHOLD = 150_000  # Warning threshold
   ```

2. **Token Counting**
   ```python
   import tiktoken

   def count_tokens(text: str, model: str = "gpt-4") -> int:
       """Count tokens in text using tiktoken."""
       encoding = tiktoken.encoding_for_model(model)
       return len(encoding.encode(text))
   ```

3. **Progressive Key Dropping Strategy**
   ```python
   def truncate_prompt_to_ceiling(
       prompt: str,
       ceiling: int,
       keys: list[str],
       key_order: list[str]  # Priority order for dropping
   ) -> str:
       """Truncate prompt by dropping keys in priority order."""
       current_tokens = count_tokens(prompt)

       if current_tokens <= ceiling:
           return prompt

       # Parse prompt to identify key-value pairs
       parsed = parse_prompt_keys(prompt)

       # Drop keys in priority order until under ceiling
       for key in key_order:
           if current_tokens <= ceiling:
               break
           if key in parsed:
               tokens_removed = count_tokens(parsed[key])
               del parsed[key]
               current_tokens -= tokens_removed
               logger.info(f"Dropped key '{key}' to meet ceiling: {tokens_removed} tokens")

       # Reconstruct prompt
       return reconstruct_prompt(parsed)
   ```

4. **Element Tree Truncation (Economy Mode)**
   ```python
   def truncate_element_tree(
       elements: list[dict],
       max_elements: int = 1000
   ) -> list[dict]:
       """Truncate element tree to maximum element count."""
       if len(elements) <= max_elements:
           return elements

       # Progressive truncation: keep 2/3, then 1/2, then 1/3
       ratios = [0.67, 0.5, 0.33]

       for ratio in ratios:
           truncated = elements[:int(len(elements) * ratio)]
           if len(truncated) <= max_elements:
               logger.info(
                   f"Truncated elements to {len(truncated)} "
                   f"(ratio: {ratio})"
               )
               return truncated

       # Final fallback: exact max_elements
       return elements[:max_elements]
   ```

### Implementation Details

**Skyvern's Prompt Engine**:
```python
class PromptEngine:
    def __init__(self):
        self.ceiling = 180_000
        self.warning_threshold = 150_000
        self.key_priority = [
            "workflow_knowledge_base",  # Drop first
            "error_code_mapping",
            "data_extraction_goal",
            "complete_criterion",
            "navigation_goal",  # Keep last
            "elements",  # Keep last
        ]

    async def build_prompt(
        self,
        template_name: str,
        **kwargs
    ) -> str:
        """Build prompt with ceiling enforcement."""
        # Render template with all variables
        prompt = self.render_template(template_name, **kwargs)

        # Check token count
        token_count = count_tokens(prompt)

        if token_count > self.warning_threshold:
            logger.warning(
                f"Prompt approaching ceiling: {token_count} tokens "
                f"(threshold: {self.warning_threshold})"
            )

        if token_count > self.ceiling:
            logger.error(
                f"Prompt exceeds ceiling: {token_count} tokens "
                f"(ceiling: {self.ceiling})"
            )
            # Apply progressive key dropping
            prompt = self.truncate_to_ceiling(prompt, kwargs)
            token_count = count_tokens(prompt)
            logger.info(f"Truncated prompt to {token_count} tokens")

        return prompt

    def truncate_to_ceiling(
        self,
        prompt: str,
        kwargs: dict
    ) -> str:
        """Truncate prompt by dropping keys in priority order."""
        current_tokens = count_tokens(prompt)

        for key in self.key_priority:
            if current_tokens <= self.ceiling:
                break
            if key in kwargs:
                tokens_removed = count_tokens(str(kwargs[key]))
                del kwargs[key]
                current_tokens -= tokens_removed
                logger.info(
                    f"Dropped key '{key}' ({tokens_removed} tokens) "
                    f"to meet ceiling"
                )

        # Re-render with reduced kwargs
        return self.render_template(self.current_template, **kwargs)
```

**Element Tree Economy Mode**:
```python
def build_element_tree(
    elements: list[dict],
    economy_mode: bool = False
) -> str:
    """Build element tree string with optional economy mode."""
    if economy_mode and len(elements) > 1000:
        elements = truncate_element_tree(elements, max_elements=1000)

    # Build tree string
    tree_lines = []
    for element in elements:
        tree_lines.append(format_element(element))

    return "\n".join(tree_lines)
```

### Prompt Structure Optimization

**Key Priority Strategy**:
1. **Drop First** (Low Priority):
   - Workflow knowledge base (large static text)
   - Error code mappings
   - Historical context
   - Data extraction schemas

2. **Keep Last** (High Priority):
   - Current navigation goal
   - Current page elements
   - User details
   - Screenshot context

**Truncation Strategies**:
1. **Key Dropping**: Remove entire key-value pairs
2. **Element Truncation**: Reduce DOM element count
3. **History Truncation**: Limit action history length
4. **Screenshot Reduction**: Reduce screenshot count/quality

### Benefits

1. **Cost Control**: Prevents runaway token usage
2. **Reliability**: Avoids context window errors
3. **Performance**: Smaller prompts improve response time
4. **Graceful Degradation**: Maintains core functionality under constraints
5. **Predictability**: Consistent behavior across different input sizes

### Trade-offs

1. **Information Loss**: Dropping keys may lose important context
2. **Complexity**: Requires careful priority tuning
3. **Token Counting Overhead**: Adds latency to prompt building
4. **Model Dependency**: Token counts vary by model/encoding

### Variations

**browser-use Approach**:
- Uses message compaction instead of hard ceiling
- Removes old messages from history
- URL shortening to save tokens
- Less aggressive truncation

**Skyvern Approach**:
- Hard ceiling with progressive key dropping
- Economy mode for element trees
- Workflow knowledge base as first drop target
- More systematic truncation strategy

### When to Use

- Building production AI agents with cost constraints
- Handling variable-sized inputs (DOM trees, histories)
- Working with context-limited models
- Need predictable token usage
- Building multi-step workflows with growing context

### When Not to Use

- Simple agents with fixed prompt sizes
- When prompt size is predictable and small
- When information loss is unacceptable
- When using models with very large context windows

### Related Patterns

- `multi-template-prompt-system`
- `message-compaction`
- `vision-based-action-extraction`

## Implementation Checklist

- [ ] Define prompt ceiling and warning thresholds
- [ ] Implement token counting with tiktoken
- [ ] Create key priority ordering
- [ ] Implement progressive key dropping logic
- [ ] Add element tree truncation (economy mode)
- [ ] Implement prompt reconstruction
- [ ] Add logging for truncation events
- [ ] Test with various input sizes
- [ ] Benchmark token savings
- [ ] Monitor ceiling hit rate in production

## References

- Skyvern: `skyvern/forge/sdk/prompting.py`
- tiktoken: https://github.com/openai/tiktoken
