# Vision-Based Action Extraction with Screenshot Scaling

**Pattern ID**: `vision-based-action-extraction`
**Source**: Skyvern (skyvern/forge), browser-use
**Category**: AI Agent Integration, Vision LLM
**Complexity**: Advanced

## Problem

AI browser automation agents need to understand visual page state to make accurate action decisions. Raw screenshots are often too large and inefficient for LLM processing, leading to high costs and slow response times.

## Solution

Implement a vision-based action extraction system with intelligent screenshot scaling:

### Core Components

1. **Screenshot Capture Pipeline**
   - Take screenshots after each action using browser automation (Playwright)
   - Support multi-screenshot scrolling for long pages
   - Store screenshots as bytes for processing

2. **Intelligent Screenshot Scaling**
   ```python
   # Target dimensions for different aspect ratios
   MAX_SCALING_TARGETS = {
       "XGA": (1024, 768),      # 4:3
       "WXGA": (1280, 800),     # 16:10
       "FWXGA": (1366, 768),    # ~16:9
   }

   def get_resize_target_dimension(window_size, max_targets):
       """Find the best matching target dimension based on aspect ratio."""
       ratio = window_size["width"] / window_size["height"]
       for dimension in max_targets.values():
           if abs(dimension["width"] / dimension["height"] - ratio) < 0.02:
               if dimension["width"] < window_size["width"]:
                   return dimension
       return window_size

   def resize_screenshots(screenshots, target_dimension):
       """Resize screenshots using LANCZOS resampling."""
       new_screenshots = []
       for screenshot in screenshots:
           img = Image.open(io.BytesIO(screenshot))
           resized_img = img.resize(
               (target_dimension["width"], target_dimension["height"]),
               Image.Resampling.LANCZOS
           )
           img_byte_arr = io.BytesIO()
           resized_img.save(img_byte_arr, format="PNG")
           new_screenshots.append(img_byte_arr.getvalue())
       return new_screenshots
   ```

3. **Vision-Enhanced Prompt Engineering**
   - Include screenshots in LLM prompts as base64-encoded images
   - Use different message formats for Anthropic vs OpenAI models
   - Limit image messages to provider constraints (e.g., 10 for Anthropic)

4. **Action Extraction with Visual Grounding**
   - Prompt LLM to analyze screenshot alongside DOM elements
   - Extract actions with element IDs from DOM tree
   - Include confidence scores for each action
   - Support action types: CLICK, INPUT_TEXT, SELECT_OPTION, UPLOAD_FILE, etc.

### Implementation Details

**Screenshot Integration in LLM Calls**:
```python
async def llm_messages_builder_with_history(
    prompt: str,
    screenshots: list[bytes] | None,
    message_history: list[dict] | None,
    message_pattern: str = "openai"
) -> list[dict]:
    messages = []
    if message_history:
        messages = copy.deepcopy(message_history)

    current_user_messages = [{"type": "text", "text": prompt}]

    if screenshots:
        for screenshot in screenshots:
            encoded_image = base64.b64encode(screenshot).decode("utf-8")
            if message_pattern == "anthropic":
                message = {
                    "type": "image",
                    "source": {
                        "type": "base64",
                        "media_type": "image/png",
                        "data": encoded_image,
                    }
                }
            else:
                message = {
                    "type": "image_url",
                    "image_url": {
                        "url": f"data:image/png;base64,{encoded_image}"
                    }
                }
            current_user_messages.append(message)

    # Enforce image message limits
    if message_pattern == "anthropic":
        image_count = sum(
            1 for msg in messages
            if msg.get("role") == "user"
            and any(block.get("type") == "image" for block in msg.get("content", []))
        )
        if image_count > MAX_IMAGE_MESSAGES:
            # Remove oldest image messages
            messages = _prune_old_image_messages(messages, MAX_IMAGE_MESSAGES)

    messages.append({"role": "user", "content": current_user_messages})
    return messages
```

**Action Extraction Prompt Template**:
```jinja2
Identify actions to help user progress towards the user goal using the DOM elements and screenshot.

Reply in JSON format with the following keys:
{
    "user_goal_stage": str,
    "user_goal_achieved": bool,
    "action_plan": str,
    "actions": [{
        "reasoning": str,
        "user_detail_query": str,
        "user_detail_answer": str,
        "confidence_float": float,  // 0.0 to 1.0
        "action_type": str,  // CLICK, INPUT_TEXT, SELECT_OPTION, etc.
        "id": str,
        "text": str,  // for INPUT_TEXT
        "option": {  // for SELECT_OPTION
            "label": str,
            "index": int,
            "value": str
        }
    }]
}

User goal: {{ navigation_goal }}
User details: {{ navigation_payload_str }}
Clickable elements: {{ elements }}
Current URL: {{ current_url }}
```

### Benefits

1. **Cost Efficiency**: Scaled screenshots reduce token usage by 60-80%
2. **Faster Response**: Smaller images improve LLM processing speed
3. **Better Accuracy**: Visual context helps LLM make better action decisions
4. **Provider Compatibility**: Different message formats support multiple LLM providers
5. **Ground Truth**: Screenshots provide visual verification of DOM state

### Trade-offs

1. **Information Loss**: Scaling may lose fine details; balance between size and quality
2. **Processing Overhead**: Resizing adds latency; consider caching
3. **Memory Usage**: Multiple screenshots in history increase context size
4. **Aspect Ratio Constraints**: Fixed target dimensions may not match all screens

### Variations

**browser-use Approach**:
- Uses bounding boxes around interactive elements in screenshots
- Provides element indexes directly in visual context
- Supports vision detail levels (auto/low/high)
- Optional screenshot usage based on agent decision

**Skyvern Approach**:
- Always includes screenshots in action extraction
- Uses aspect-ratio-aware scaling targets
- Supports multi-screenshot scrolling for long pages
- Integrates with artifact management for persistence

### When to Use

- Building AI browser automation agents
- Implementing vision-based UI interaction
- Creating web scraping with visual understanding
- Building agents that need to verify visual state
- Implementing CAPTCHA-solving workflows

### When Not to Use

- Simple form filling without visual complexity
- API-only interactions without browser
- When screenshots add no value (text-only pages)
- Extremely latency-sensitive applications

### Related Patterns

- `multi-template-prompt-system`
- `prompt-ceiling-enforcement`
- `structured-action-output`
- `llm-error-classification`

## Implementation Checklist

- [ ] Implement screenshot capture with browser automation
- [ ] Add aspect-ratio-aware scaling logic
- [ ] Integrate screenshots into LLM message builders
- [ ] Implement message format switching (Anthropic vs OpenAI)
- [ ] Add image message limit enforcement
- [ ] Create action extraction prompt templates
- [ ] Implement confidence scoring for actions
- [ ] Add screenshot artifact persistence
- [ ] Test with different screen sizes and aspect ratios
- [ ] Benchmark cost savings from scaling

## References

- Skyvern: `skyvern/forge/agent.py`, `skyvern/utils/image_resizer.py`
- browser-use: `browser_use/agent/service.py`, `browser_use/agent/views.py`
- Anthropic Computer Use: https://github.com/anthropics/anthropic-quickstarts
