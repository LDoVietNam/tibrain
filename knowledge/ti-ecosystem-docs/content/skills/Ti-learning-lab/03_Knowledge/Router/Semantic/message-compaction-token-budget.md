# Message Compaction for Token Budget Management

**Pattern ID**: `message-compaction-token-budget`
**Source**: browser-use (browser_use/agent/service.py)
**Category**: Cost Optimization, AI Agent
**Complexity**: Intermediate

## Problem

AI agents accumulate message history across steps, leading to:
- Excessive token usage and cost
- Context window exceeded errors
- Slow response times
- Diminishing returns from old context

## Solution

Implement message compaction to manage token budget by removing or summarizing old messages:

### Core Components

1. **Token Budget Configuration**
   ```python
   @dataclass
   class AgentConfig:
       max_tokens: int = 200_000
       token_budget_warning_threshold: float = 0.8  # 80%
       compact_threshold: int = 50_000  # Compact when over this
       compact_ratio: float = 0.5  # Remove 50% of old messages
   ```

2. **Message Compaction Strategy**
   ```python
   class MessageManager:
       def __init__(self, config: AgentConfig):
           self.config = config
           self.messages: list[dict] = []
           self.total_tokens: int = 0

       async def compact_messages(self) -> None:
           """Compact message history to stay within token budget."""
           current_tokens = self._count_tokens()

           if current_tokens < self.config.compact_threshold:
               return

           logger.info(
               f"Compacting messages: {current_tokens} tokens "
               f"(threshold: {self.config.compact_threshold})"
           )

           # Remove oldest messages while keeping recent context
           target_tokens = int(current_tokens * self.config.compact_ratio)
           compacted = self._remove_oldest_messages(target_tokens)

           logger.info(
               f"Compacted {compacted['removed_count']} messages, "
               f"removed {compacted['removed_tokens']} tokens, "
               f"remaining {compacted['remaining_tokens']} tokens"
           )

       def _remove_oldest_messages(
       self,
       target_tokens: int
   ) -> dict:
       """Remove oldest messages until under target token count."""
       removed_count = 0
       removed_tokens = 0

       # Keep system message and last N user/assistant pairs
       keep_count = 5  # Keep last 5 exchanges
       if len(self.messages) <= keep_count * 2 + 1:  # +1 for system
           return {
               'removed_count': 0,
               'removed_tokens': 0,
               'remaining_tokens': self._count_tokens()
           }

       # Remove messages from middle (keep system and recent)
       system_msg = self.messages[0] if self.messages[0]['role'] == 'system' else None
       recent_messages = self.messages[-(keep_count * 2):]  # Last N exchanges
       middle_messages = self.messages[1:-(keep_count * 2)]

       # Calculate tokens in middle messages
       for msg in middle_messages:
           removed_tokens += self._count_message_tokens(msg)
           removed_count += 1

       # Reconstruct with system + recent
       new_messages = []
       if system_msg:
           new_messages.append(system_msg)
       new_messages.extend(recent_messages)

       self.messages = new_messages

       return {
           'removed_count': removed_count,
           'removed_tokens': removed_tokens,
           'remaining_tokens': self._count_tokens()
       }

   def _count_tokens(self) -> int:
       """Count total tokens in message history."""
       return sum(self._count_message_tokens(msg) for msg in self.messages)

   def _count_message_tokens(self, message: dict) -> int:
       """Count tokens in a single message."""
       import tiktoken
       encoding = tiktoken.encoding_for_model("gpt-4")

       content = message.get('content', '')
       if isinstance(content, str):
           return len(encoding.encode(content))
       elif isinstance(content, list):
           # Handle multimodal content (text + images)
           text_content = '\n'.join(
               item.get('text', '') for item in content
               if item.get('type') == 'text'
           )
           return len(encoding.encode(text_content))
       return 0
   ```

3. **URL Shortening**
   ```python
   def shorten_urls_in_messages(self) -> None:
       """Shorten URLs in messages to save tokens."""
       for message in self.messages:
           content = message.get('content', '')
           if isinstance(content, str):
               message['content'] = self._shorten_urls(content)
           elif isinstance(content, list):
               for item in content:
                   if item.get('type') == 'text':
                       item['text'] = self._shorten_urls(item['text'])

   def _shorten_urls(self, text: str) -> str:
       """Shorten URLs in text."""
       import re

       url_pattern = r'https?://[^\s<>"{}|\\^`\[\]]+'

       def replace_url(match):
           url = match.group(0)
           # Keep domain, truncate path
           parsed = urllib.parse.urlparse(url)
           if len(parsed.path) > 50:
               short_path = parsed.path[:50] + '...'
               new_url = f"{parsed.scheme}://{parsed.netloc}{short_path}"
               if parsed.query:
                   new_url += '?...'
               return new_url
           return url

       return re.sub(url_pattern, replace_url, text)
   ```

4. **Context Summarization**
   ```python
   async def summarize_old_context(self) -> None:
       """Summarize old context instead of removing it."""
       if len(self.messages) < 10:
           return

       # Split into old and new
       old_messages = self.messages[:5]
       new_messages = self.messages[-5:]

       # Generate summary of old messages
       summary = await self._generate_summary(old_messages)

       # Replace old messages with summary
       system_msg = self.messages[0] if self.messages[0]['role'] == 'system' else None

       self.messages = []
       if system_msg:
           self.messages.append(system_msg)
       self.messages.append({
           'role': 'system',
           'content': f'Previous context summary: {summary}'
       })
       self.messages.extend(new_messages)

   async def _generate_summary(self, messages: list[dict]) -> str:
       """Generate summary of messages using LLM."""
       summary_prompt = (
           'Summarize the following conversation history in 2-3 sentences. '
           'Focus on key information and progress toward the goal.\n\n'
           + '\n'.join(f"{msg['role']}: {msg['content']}" for msg in messages)
       )

       response = await self.llm.ainvoke(summary_prompt)
       return response.content
   ```

### Implementation Details

**Token Budget Monitoring**:
```python
class Agent:
    async def step(self) -> None:
        """Execute agent step with token budget management."""
        # Check token budget
        current_tokens = self.message_manager._count_tokens()
        budget_ratio = current_tokens / self.config.max_tokens

        if budget_ratio > self.config.token_budget_warning_threshold:
            logger.warning(
                f"Token budget at {budget_ratio:.1%} "
                f"({current_tokens}/{self.config.max_tokens} tokens)"
            )

        if current_tokens > self.config.compact_threshold:
            await self.message_manager.compact_messages()

        # Execute step
        await self._execute_step()
```

**Selective Compaction**:
```python
def compact_messages_selectively(self) -> None:
    """Compact messages while preserving critical information."""
    # Identify critical message types to preserve
    critical_types = ['system', 'user_request', 'error']

    # Group messages by type
    critical_messages = [
        msg for msg in self.messages
        if msg.get('type') in critical_types
    ]
    other_messages = [
        msg for msg in self.messages
        if msg.get('type') not in critical_types
    ]

    # Compact non-critical messages first
    if len(other_messages) > 10:
        # Keep last 5 non-critical messages
        other_messages = other_messages[-5:]

    # Reconstruct
    self.messages = critical_messages + other_messages
```

### Benefits

1. **Cost Reduction**: Reduces token usage by 40-60%
2. **Performance**: Smaller context improves response time
3. **Reliability**: Prevents context window exceeded errors
4. **Flexibility**: Multiple compaction strategies (removal, summarization)
5. **Monitoring**: Track token usage patterns

### Trade-offs

1. **Information Loss**: Removing old messages may lose important context
2. **Summarization Cost**: Summarization requires additional LLM calls
3. **Complexity**: Adds logic to message management
4. **Tuning**: Thresholds require tuning per use case

### Variations

**browser-use Approach**:
- Remove oldest messages while keeping recent context
- URL shortening to save tokens
- Selective compaction preserving critical messages
- Token budget monitoring with warnings

**Skyvern Approach**:
- Prompt ceiling enforcement with key dropping
- Element tree truncation (economy mode)
- Less message history compaction
- More focused on prompt-level optimization

### When to Use

- Building AI agents with long-running conversations
- When token cost is a concern
- Tasks with growing context over time
- When context window limits are tight
- Building production agents with cost constraints

### When Not to Use

- Short conversations (< 10 messages)
- When all context is critical
- When using models with very large context windows
- Quick prototypes without cost concerns

### Related Patterns

- `prompt-ceiling-enforcement`
- `multi-template-prompt-system`
- `vision-based-action-extraction`

## Implementation Checklist

- [ ] Define token budget configuration
- [ ] Implement token counting with tiktoken
- [ ] Add message compaction logic (remove oldest)
- [ ] Implement URL shortening
- [ ] Add context summarization option
- [ ] Implement selective compaction
- [ ] Add token budget monitoring
- [ ] Test with various conversation lengths
- [ ] Benchmark token savings
- [ ] Monitor compaction frequency in production

## References

- browser-use: `browser_use/agent/service.py` (message compaction)
- tiktoken: https://github.com/openai/tiktoken
