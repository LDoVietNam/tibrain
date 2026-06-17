# Skyvern Prompt Engineering

**Repository**: Skyvern  
**Location**: `skyvern/forge/prompts/skyvern/`  
**Total Templates**: 70+ Jinja2 templates

---

## Overview

Skyvern uses sophisticated prompt engineering with 70+ Jinja2 templates for different LLM interactions. Templates use static/dynamic split for caching optimization, structured JSON outputs, and detailed reasoning instructions.

---

## Prompt Categories

### Action Extraction Prompts

**extract-action.j2** (97 lines) - Main action extraction prompt

**Key Features**:
- Structured JSON output with action plan
- User detail query/answer pattern for data separation
- Confidence scoring for each action
- Click context for single/multi-choice detection
- Context extraction for input/select operations

**Output Schema**:
```json
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
    "captcha_type": str,
    "text": str,
    "key": str,
    "direction": str,
    "file_url": str,
    "download": bool,
    "option": {...},
    "click_context": {...},
    "context": {...}
  }]
}
```

### Task V2 Prompts

**task_v2.j2** (78 lines) - Task planning with 3 task types

**Task Types**:
- **navigate** - Mini goal for page interaction
- **extract** - Information extraction from page
- **loop** - Iterate through values for parallel tasks

**Key Features**:
- Conservative termination criteria
- Explicit termination guidelines
- Loop usage examples
- Task history integration

**Output Schema**:
```json
{
  "page_info": str,
  "extraction_thought": str,
  "require_extraction": bool,
  "task_history_information": str,
  "information_extracted": bool,
  "thoughts": str,
  "user_goal_achieved": bool,
  "should_terminate": bool,
  "termination_reason": str,
  "plan": str,
  "task_type": str,
  "loop_values": list[str],
  "is_loop_value_link": bool
}
```

### Single Action Prompts

- **single-click-action.j2** - Single click detection
- **single-input-action.j2** - Single input detection
- **single-select-action.j2** - Single select detection
- **single-upload-action.j2** - Single upload detection
- **single-locate-element.j2** - Element location

### Context Parsing Prompts

- **parse-input-or-select-context.j2** - Input/select context extraction
- **auto-completion-choose-option.j2** - Auto-completion selection
- **auto-completion-potential-answers.j2** - Auto-completion answers
- **auto-completion-tweak-value.j2** - Auto-completion value tweaking

### Verification Prompts

- **check-user-goal.j2** - User goal verification
- **check-user-goal-with-termination.j2** - Verification with termination
- **check-date-format.j2** - Date format validation
- **check-phone-number-format.j2** - Phone format validation
- **decisive-criterion-validate.j2** - Criterion validation

### Extraction Prompts

- **extract-information.j2** - Data extraction
- **extract-action-static.j2** - Static template for caching
- **extract-action-dynamic.j2** - Dynamic template with variables
- **extract-information-from-file-text.j2** - File text extraction
- **extract-text-from-image.j2** - Image text extraction

### Script Generation Prompts

- **script-generation-input-text-generatiion.j2** - Input text generation
- **script-generation-file-url-generation.j2** - File URL generation
- **script-reviewer.j2** - Script review
- **script-reviewer-form-filling.j2** - Form filling review
- **script-reviewer-extraction.j2** - Extraction review
- **script-reviewer-conditional.j2** - Conditional review
- **script-failure-triage.j2** - Failure triage

### Copilot Prompts

- **workflow-copilot-agent.j2** - Copilot agent system prompt
- **workflow-copilot-system.j2** - Copilot system prompt
- **workflow-copilot-user.j2** - Copilot user prompt
- **feasibility-gate.j2** - Feasibility gate validation

### Recording Prompts

- **recording-action-block-prompt.j2** - Recording action block
- **recording-action-block-prompt-input-text.j2** - Recording input text
- **recording-go-to-url-block-prompt.j2** - Recording URL navigation
- **recording-wait-block-prompt.j2** - Recording wait block

### Specialized Prompts

- **handle-dialog.j2** - Dialog handling
- **solve-captcha.j2** - CAPTCHA solving
- **custom-select.j2** - Custom selection
- **select-from-group.j2** - Group selection
- **normal-select.j2** - Normal selection
- **confirm-multi-selection-finish.j2** - Multi-selection confirmation

---

## Prompt Engineering Techniques

### 1. Static/Dynamic Split

**Pattern**: Split templates into static (cacheable) and dynamic (runtime) parts

**Files**:
- `extract-action.j2` - Complete template
- `extract-action-static.j2` - Cacheable prefix
- `extract-action-dynamic.j2` - Dynamic suffix

**Benefits**:
- **Caching optimization**: Cache static parts
- **Reduced LLM costs**: Only generate dynamic parts
- **Faster execution**: Skip static generation

### 2. User Detail Query/Answer Pattern

**Pattern**: Separate user detail query from answer

**Purpose**: Enable data separation and user context flexibility

**Example**:
```json
{
  "user_detail_query": "What product ID should I input?",
  "user_detail_answer": "Product ID from user details"
}
```

### 3. Confidence Scoring

**Pattern**: Request confidence score for each action

**Purpose**: Enable action filtering and quality control

**Range**: 0.0 (no confidence) to 1.0 (full confidence)

### 4. Click Context Detection

**Pattern**: Detect single vs multi-choice scenarios

**Purpose**: Enable intelligent action planning

**Fields**:
- `single_option_click` - True if only one valid option
- `thought` - Reasoning for classification

### 5. Context Extraction

**Pattern**: Extract detailed context for input/select operations

**Fields**:
- `field` - Which field is being filled
- `is_required` - Whether field is required
- `is_search_bar` - Whether element is search bar
- `is_location_input` - Whether asking for location
- `is_date_related` - Whether date input
- `date_format` - Date format if applicable
- `is_text_captcha` - Whether text CAPTCHA

### 6. Conservative Termination

**Pattern**: Explicit termination guidelines with conservative approach

**Guidelines**:
- Terminate ONLY with explicit evidence
- Do NOT terminate on transient errors
- Do NOT terminate when alternative approaches exist
- Quote EXACT error message for termination

### 7. Loop Task Guidance

**Pattern**: Provide clear guidance on when to use loop tasks

**Examples**:
- Breadth-first search situations
- Parallel tasks with same goal
- Multiple values to iterate through

### 8. Verification Code Detection

**Pattern**: Detect and handle verification codes

**Fields**:
- `verification_code_reasoning` - Reasoning about verification code
- `place_to_enter_verification_code` - Whether place exists
- `should_enter_verification_code` - Whether to enter code
- `should_verify_by_magic_link` - Whether magic link verification

---

## Prompt Template System

**Location**: `forge/prompts/`

**Template Engine**: Jinja2

**Loading**:
```python
from skyvern.forge.prompts import prompt_engine

prompt = prompt_engine.load_prompt(
    template="extract-action",
    navigation_goal="...",
    elements="...",
    # ... other variables
)
```

---

## Key Patterns

### 1. Structured JSON Output

**Pattern**: Request valid JSON with strict formatting

**Instructions**:
```
MAKE SURE YOU OUTPUT VALID JSON. No text before or after JSON, 
no trailing commas, no comments (//), no unnecessary quotes, etc.
```

### 2. Step-by-Step Reasoning

**Pattern**: Request step-by-step thinking

**Examples**:
- "Think step by step. Describe..."
- "Let's think step by step. Explain..."

### 3. User Information Agnostic Queries

**Pattern**: Ask generic questions without specific user data

**Purpose**: Enable context reuse across different users

### 4. Error Code Mapping Integration

**Pattern**: Include user-defined error codes in prompts

**Purpose**: Enable user-facing error messages

### 5. Action History Integration

**Pattern**: Include action history from previous steps

**Purpose**: Enable learning from past actions

---

## Testing Considerations

### Test Scenarios

1. **Template rendering** - Verify all templates render correctly
2. **JSON parsing** - Verify JSON output is valid
3. **Static/dynamic split** - Verify caching optimization
4. **Context extraction** - Verify context detection accuracy
5. **Termination logic** - Verify conservative termination

---

## References

- **Prompts Directory**: `forge/prompts/skyvern/`
- **Prompt Engine**: `forge/prompts/__init__.py`
- **Extract Action**: `forge/prompts/skyvern/extract-action.j2`
- **Task V2**: `forge/prompts/skyvern/task_v2.j2`
