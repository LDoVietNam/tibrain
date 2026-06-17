# Multi-Template Prompt System with Jinja2

**Pattern ID**: `multi-template-prompt-system`
**Source**: Skyvern (skyvern/forge/prompts)
**Category**: Prompt Engineering, AI Agent
**Complexity**: Intermediate

## Problem

AI agents need different prompt templates for various tasks (action extraction, task planning, data extraction, etc.). Hardcoding prompts makes maintenance difficult and prevents reuse across different agent types.

## Solution

Implement a Jinja2-based multi-template prompt system with:

### Core Components

1. **Template Registry**
   - Organize templates by purpose (action extraction, task planning, verification)
   - Use `.j2` extension for Jinja2 templates
   - Support template inheritance and composition

2. **Prompt Engine**
   ```python
   from jinja2 import Environment, FileSystemLoader, Template

   class PromptEngine:
       def __init__(self, template_dir: str):
           self.env = Environment(
               loader=FileSystemLoader(template_dir),
               autoescape=False
           )

       def render_template(
           self,
           template_name: str,
           **kwargs
       ) -> str:
           template = self.env.get_template(template_name)
           return template.render(**kwargs)
   ```

3. **Template Structure**
   ```
   prompts/
   ├── skyvern/
   │   ├── extract-action.j2              # Complete template
   │   ├── extract-action-static.j2      # Cacheable prefix
   │   ├── extract-action-dynamic.j2     # Dynamic suffix
   │   ├── task_v2.j2                     # Task planning
   │   ├── single-input-action.j2         # Input actions
   │   ├── single-click-action.j2         # Click actions
   │   ├── script-reviewer.j2             # Script review
   │   └── workflow-copilot-agent.j2      # Workflow copilot
   └── browser_use/
       ├── system_prompt.md               # System prompt
       ├── system_prompt_browser_use.md   # Browser-specific
       └── system_prompt_flash.md         # Flash model
   ```

### Template Examples

**Action Extraction Template** (`extract-action.j2`):
```jinja2
Identify actions to help user progress towards the user goal using the DOM elements given in the list and the screenshot of the website.
Include only the elements that are relevant to the user goal, without altering or imagining new elements.
Accurately interpret and understand the functional significance of SVG elements based on their shapes and context within the webpage.

Reply in JSON format with the following keys:
{
    "user_goal_stage": str,
    "user_goal_achieved": bool,
    "action_plan": str,
    "actions": [{
        "reasoning": str,
        "user_detail_query": str,
        "user_detail_answer": str,
        "confidence_float": float,
        "action_type": str,
        "id": str,
        {% if parse_select_feature_enabled %}
        "context": {
            "thought": str,
            "field": str,
            "is_required": bool,
            "is_search_bar": bool,
            "is_location_input": bool,
            "is_date_related": bool,
            "date_format": str,
            "is_text_captcha": bool
        }
        {% endif %}
    }]
}

User goal:
```
{{ navigation_goal }}
```

User details:
```
{{ navigation_payload_str }}
```

Clickable elements from `{{ current_url }}`:
```
{{ elements }}
```

Current datetime, ISO format:
```
{{ local_datetime }}
```
```

**Task Planning Template** (`task_v2.j2`):
```jinja2
You're to assist the user to achieve the user goal in the web, given the DOM elements in the list, the screenshots of the website and the task history list. Plan the next task the user needs to do towards the goal.

You have access to the following task types:
- navigate: set up a mini goal to achieve in the web which most likely results in navigating
- extract: extract information users would like to output from the page
- loop: generate a list of planning sessions for parallel tasks

Reply in JSON format with the following keys:
{
  "page_info": str,
  "extraction_thought": str,
  "require_extraction": bool,
  "task_history_information": str,
  "information_extracted": optional[bool],
  "thoughts": str,
  "user_goal_achieved": bool,
  "should_terminate": bool,
  "termination_reason": str,
  "plan": str,
  "task_type": str,  // navigate, extract, loop
  "loop_values": list[str],
  "is_loop_value_link": bool
}

The URL of the page you're on right now is `{{ current_url }}`.

Clickable elements from the page:
```
{{ elements }}
```

User goal:
```
{{ user_goal }}
```

Task history:
```
{{ task_history }}
```
```

### Template Caching Optimization

**Static/Dynamic Split**:
```jinja2
# extract-action-static.j2 (cacheable prefix)
Identify actions to help user progress towards the user goal using the DOM elements given in the list and the screenshot of the website.
Include only the elements that are relevant to the user goal, without altering or imagining new elements.
Accurately interpret and understand the functional significance of SVG elements based on their shapes and context within the webpage.

Reply in JSON format with the following keys:
{
    "user_goal_stage": str,
    "user_goal_achieved": bool,
    "action_plan": str,
    "actions": [{
        "reasoning": str,
        "user_detail_query": str,
        "user_detail_answer": str,
        "confidence_float": float,
        "action_type": str,
        "id": str,
        {% if parse_select_feature_enabled %}
        "context": {
            "thought": str,
            "field": str,
            "is_required": bool,
            "is_search_bar": bool,
            "is_location_input": bool,
            "is_date_related": bool,
            "date_format": str,
            "is_text_captcha": bool
        }
        {% endif %}
    }]
}

# extract-action-dynamic.j2 (dynamic suffix with runtime variables)
User goal:
```
{{ navigation_goal }}
```

User details:
```
{{ navigation_payload_str }}
```

Clickable elements from `{{ current_url }}`:
```
{{ elements }}
```

Current datetime, ISO format:
```
{{ local_datetime }}
```
```

**Usage**:
```python
def render_extract_action_prompt(
    static_template: str,
    dynamic_template: str,
    **kwargs
) -> str:
    static = prompt_engine.render_template(static_template, **kwargs)
    dynamic = prompt_engine.render_template(dynamic_template, **kwargs)
    return static + dynamic
```

### Conditional Template Logic

**Feature Flags**:
```jinja2
{% if parse_select_feature_enabled %}
"context": {
    "thought": str,
    "field": str,
    "is_required": bool,
    "is_search_bar": bool,
    "is_location_input": bool,
    "is_date_related": bool,
    "date_format": str,
    "is_text_captcha": bool
}{% endif %}
```

**Verification Code Handling**:
```jinja2
{% if verification_code_check %}
"verification_code_reasoning": str,
"place_to_enter_verification_code": bool,
"should_enter_verification_code": bool{% endif %}
```

### Template Selection Strategy

```python
def select_template(
    action_type: str,
    use_case: str,
    feature_flags: dict
) -> str:
    """Select appropriate template based on context."""
    if action_type == "INPUT_TEXT":
        return "single-input-action.j2"
    elif action_type == "CLICK":
        return "single-click-action.j2"
    elif action_type == "SELECT_OPTION":
        return "single-select-action.j2"
    elif use_case == "task_planning":
        return "task_v2.j2"
    elif use_case == "verification":
        return "task_v2_check_completion.j2"
    else:
        return "extract-action.j2"
```

### Benefits

1. **Maintainability**: Centralized template management
2. **Reusability**: Templates shared across different agent types
3. **Flexibility**: Conditional rendering based on feature flags
4. **Caching**: Static/dynamic split enables prompt caching
5. **Versioning**: Easy to track template changes
6. **Testing**: Templates can be tested independently

### Trade-offs

1. **Complexity**: Jinja2 syntax adds learning curve
2. **Performance**: Template rendering adds overhead (mitigated by caching)
3. **Debugging**: Template errors can be harder to debug
4. **File Proliferation**: Many template files to manage

### Variations

**browser-use Approach**:
- Uses Markdown files for system prompts
- Multiple system prompt variants for different models
- Inline prompt construction in Python code
- Less template-heavy, more code-driven

**Skyvern Approach**:
- Extensive Jinja2 template library (80+ templates)
- Strict separation of static/dynamic parts
- Feature flag-driven conditional rendering
- Template-specific for each action type

### When to Use

- Building AI agents with multiple interaction modes
- Need for flexible prompt composition
- Prompt caching is important for cost optimization
- Team needs to maintain prompts without code changes
- Supporting multiple LLM providers with different formats

### When Not to Use

- Simple agents with single prompt template
- When prompt changes are rare
- Performance-critical paths where rendering overhead matters
- Team unfamiliar with Jinja2

### Related Patterns

- `prompt-ceiling-enforcement`
- `vision-based-action-extraction`
- `structured-action-output`

## Implementation Checklist

- [ ] Set up Jinja2 environment with template loader
- [ ] Create template directory structure
- [ ] Implement prompt engine with render method
- [ ] Create core templates (action extraction, task planning)
- [ ] Add conditional rendering with feature flags
- [ ] Implement static/dynamic template split for caching
- [ ] Add template selection logic
- [ ] Create template testing framework
- [ ] Document template variables and usage
- [ ] Set up template versioning

## References

- Skyvern: `skyvern/forge/prompts/skyvern/`
- Jinja2 Documentation: https://jinja.palletsprojects.com/
