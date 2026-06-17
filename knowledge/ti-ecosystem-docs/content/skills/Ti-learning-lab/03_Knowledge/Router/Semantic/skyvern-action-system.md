# Skyvern Action System

**Repository**: Skyvern  
**Location**: `skyvern/webeye/actions/`  
**Key Files**:
- `action_types.py` (57 lines) - Action type definitions
- `actions.py` (402 lines) - Action models and schemas
- `handler.py` (4,931 lines) - Action execution handler
- `handler_utils.py` - Handler utilities
- `parse_actions.py` - Action parsing
- `responses.py` - Action response types
- `caching.py` - Action caching

---

## Overview

Skyvern's Action System provides 22 distinct action types for browser automation, ranging from basic interactions (click, type) to complex operations (CAPTCHA solving, file operations). Actions are executed through a sophisticated handler with element tree integration, error handling, and caching.

---

## Action Types

**Location**: `action_types.py` (57 lines)

### Web Actions (7 types)

```python
class ActionType(StrEnum):
    CLICK = "click"
    INPUT_TEXT = "input_text"
    UPLOAD_FILE = "upload_file"
    DOWNLOAD_FILE = "download_file"  # Not used in current implementation
    SELECT_OPTION = "select_option"
    CHECKBOX = "checkbox"
    HOVER = "hover"
```

**Purpose**: Direct browser interactions with page elements

### Control Actions (7 types)

```python
    WAIT = "wait"
    SOLVE_CAPTCHA = "solve_captcha"
    TERMINATE = "terminate"
    COMPLETE = "complete"
    RELOAD_PAGE = "reload_page"
    CLOSE_PAGE = "close_page"
    NULL_ACTION = "null_action"
```

**Purpose**: Control flow and task lifecycle management

### Data Actions (2 types)

```python
    EXTRACT = "extract"
    VERIFICATION_CODE = "verification_code"
```

**Purpose**: Data extraction and verification

### Navigation Actions (5 types)

```python
    GOTO_URL = "goto_url"
    SCROLL = "scroll"
    KEYPRESS = "keypress"
    MOVE = "move"
    DRAG = "drag"
    LEFT_MOUSE = "left_mouse"
```

**Purpose**: Page navigation and cursor manipulation

### Post-Action Execution

**Location**: `action_types.py` (lines 45-57)

```python
POST_ACTION_EXECUTION_ACTION_TYPES = [
    ActionType.CLICK,
    ActionType.HOVER,
    ActionType.INPUT_TEXT,
    ActionType.UPLOAD_FILE,
    ActionType.DOWNLOAD_FILE,
    ActionType.SELECT_OPTION,
    ActionType.WAIT,
    ActionType.SOLVE_CAPTCHA,
    ActionType.EXTRACT,
    ActionType.KEYPRESS,
    ActionType.SCROLL,
]
```

**Purpose**: Actions that trigger post-execution processing (e.g., screenshot, element tree refresh)

---

## Action Models

**Location**: `actions.py` (402 lines)

### Action Status

```python
class ActionStatus(StrEnum):
    pending = "pending"
    skipped = "skipped"
    failed = "failed"
    completed = "completed"
```

### Verification Status

```python
class VerificationStatus(StrEnum):
    complete = "complete"  # Goal achieved successfully
    terminate = "terminate"  # Goal cannot be achieved, stop trying
    continue_step = "continue"  # Goal not yet achieved, continue with more steps
```

### Complete Verify Result

```python
class CompleteVerifyResult(BaseModel):
    # New field: explicit status with three options
    status: VerificationStatus | None = None
    
    # Legacy fields: for backward compatibility
    user_goal_achieved: bool = False
    should_terminate: bool = False
    
    thoughts: str
    page_info: str | None = None
    failure_categories: list[dict] = []
    
    @property
    def is_complete(self) -> bool:
        """True if goal was achieved (supports both new and legacy formats)."""
        if self.status:
            return self.status == VerificationStatus.complete
        return self.user_goal_achieved
    
    @property
    def is_terminate(self) -> bool:
        """True if task should terminate (supports both new and legacy formats)."""
        if self.status:
            return self.status == VerificationStatus.terminate
        return self.should_terminate
    
    @property
    def is_continue(self) -> bool:
        """True if task should continue (supports both new and legacy formats)."""
        if self.status:
            return self.status == VerificationStatus.continue_step
        return not self.user_goal_achieved and not self.should_terminate
```

### Input Or Select Context

```python
class InputOrSelectContext(BaseModel):
    intention: str | None = None
    field: str | None = None
    is_required: bool | None = None
    is_search_bar: bool | None = None  # don't trigger custom-selection logic
    is_location_input: bool | None = None  # address input requires auto completion
    is_date_related: bool | None = None  # date picker requires special logic
    date_format: str | None = None
    is_text_captcha: bool | None = None
```

**Purpose**: Context for input/select operations to enable intelligent handling

### CAPTCHA Types

```python
class CaptchaType(StrEnum):
    TEXT_CAPTCHA = "text_captcha"
    RECAPTCHA = "recaptcha"
    HCAPTCHA = "hcaptcha"
    MTCAPTCHA = "mtcaptcha"
    FUNCAPTCHA = "funcaptcha"
    CLOUDFLARE = "cloudflare"
    OTHER = "other"
```

---

## Action Handler Architecture

**Location**: `handler.py` (4,931 lines)

### Screenshot Without Cursor

**Location**: `handler.py` (lines 127-141)

```python
async def _screenshot_without_cursor(page: Page, **kwargs: Any) -> bytes:
    """Take a screenshot with cursor overlay hidden so it doesn't interfere with LLM analysis."""
    if SettingsManager.get_settings().BROWSER_CURSOR_VISUALIZATION:
        try:
            await SkyvernFrame.hide_cursor_overlay(page)
        except Exception:
            pass
        try:
            return await page.screenshot(**kwargs)
        finally:
            try:
                await SkyvernFrame.show_cursor_overlay(page)
            except Exception:
                pass
    return await page.screenshot(**kwargs)
```

**Purpose**: Hide cursor visualization during screenshot to prevent interference with LLM analysis

### Custom Single Select Result

**Location**: `handler.py` (lines 144-167)

```python
class CustomSingleSelectResult:
    def __init__(self, skyvern_frame: SkyvernFrame) -> None:
        self.reasoning: str | None = None
        self.action_result: ActionResult | None = None
        self.action_type: ActionType | None = None
        self.value: str | None = None
        self.dropdown_menu: SkyvernElement | None = None
        self.skyvern_frame = skyvern_frame
    
    async def is_done(self) -> bool:
        """Check if dropdown menu is still on the page for multi-level selection."""
        if self.dropdown_menu is None:
            return True
        
        if not isinstance(self.action_result, ActionSuccess):
            return True
        
        if await self.dropdown_menu.get_locator().count() == 0:
            return True
        
        return not await self.skyvern_frame.get_element_visible(
            await self.dropdown_menu.get_element_handler()
        )
```

**Purpose**: Handle multi-level dropdown selection with completion detection

### Element Filter Factories

**Location**: `handler.py` (lines 169-199)

```python
def is_ul_or_listbox_element_factory(
    incremental_scraped: IncrementalScrapePage, task: Task, step: Step
) -> Callable[[dict], Awaitable[bool]]:
    """Factory function to check if element is ul or listbox."""
    async def wrapper(element_dict: dict) -> bool:
        element_id: str = element_dict.get("id", "")
        try:
            element = await SkyvernElement.create_from_incremental(incremental_scraped, element_id)
        except Exception:
            return False
        
        if element.get_tag_name() == "ul":
            return True
        
        if await element.get_attr("role") == "listbox":
            return True
        
        return False
    
    return wrapper
```

**Purpose**: Create reusable element filter functions for action execution

---

## Error Handling

### Exception Types

**Location**: `handler.py` (lines 29-60)

**Element Errors**:
- `MissingElement` - Element not found
- `MultipleElementsFound` - Multiple elements match selector
- `MissingElementDict` - Element dictionary missing
- `MissingElementInCSSMap` - Element not in CSS map
- `NoElementMatchedForTargetOption` - No element matches target option

**Interaction Errors**:
- `FailToClick` - Click action failed
- `FailToHover` - Hover action failed
- `FailToSelectByIndex` - Select by index failed
- `FailToSelectByLabel` - Select by label failed
- `FailToSelectByValue` - Select by value failed

**Input Errors**:
- `InputToInvisibleElement` - Input to invisible element
- `InputToReadonlyElement` - Input to readonly element
- `InvalidElementForTextInput` - Invalid element for text input
- `InteractWithDisabledElement` - Interact with disabled element
- `InteractWithDropdownContainer` - Interact with dropdown container

**Selection Errors**:
- `EmptySelect` - Empty select element
- `ErrFoundSelectableElement` - Found selectable element error
- `NoAvailableOptionFoundForCustomSelection` - No option for custom selection
- `NoIncrementalElementFoundForAutoCompletion` - No element for auto-completion
- `NoIncrementalElementFoundForCustomSelection` - No element for custom selection
- `NoSuitableAutoCompleteOption` - No suitable auto-complete option
- `OptionIndexOutOfBound` - Option index out of bounds
- `NoAutoCompleteOptionMeetCondition` - No auto-complete option meets condition

**File Errors**:
- `ImaginaryFileUrl` - Invalid file URL
- `MissingFileUrl` - Missing file URL
- `WrongElementToUploadFile` - Wrong element for file upload

**Other Errors**:
- `IllegitComplete` - Invalid completion
- `ImaginarySecretValue` - Invalid secret value
- `FailedToFetchSecret` - Failed to fetch secret
- `TOTPExpiredError` - TOTP code expired

---

## Action Execution Flow

### 1. Action Parsing

**Location**: `parse_actions.py`

**Purpose**: Parse LLM output into structured action objects

### 2. Action Validation

**Location**: `handler.py`

**Purpose**: Validate action parameters and preconditions

### 3. Element Location

**Location**: `handler.py` (lines 169-199)

**Purpose**: Locate target element using element tree and selectors

### 4. Action Execution

**Location**: `handler.py` (4,931 lines)

**Purpose**: Execute the action with Playwright

### 5. Post-Action Processing

**Location**: `handler.py`

**Purpose**: Refresh element tree, take screenshot, update state

### 6. Result Generation

**Location**: `responses.py`

**Purpose**: Generate action result with status and metadata

---

## Element Tree Integration

### Element Tree Builder

**Location**: `webeye/scraper/scraped_page.py`

**Purpose**: Build and manage element tree for action execution

### Element Tree Format

**Location**: `webeye/scraper/scraped_page.py`

```python
class ElementTreeFormat(StrEnum):
    """Format for element tree representation."""
```

### Cleanup Functions

**Location**: `webeye/scraper/scraped_page.py`

```python
CleanupElementTreeFunc = Callable[[dict], dict]
```

**Purpose**: Clean up element tree for LLM consumption

---

## Action Caching

**Location**: `caching.py`

**Purpose**: Cache action results for performance optimization

---

## Key Patterns

### 1. Factory Pattern for Element Filters

**Pattern**: Create reusable filter functions with closure

**Benefits**:
- Reusable across actions
- Captures context (incremental_scraped, task, step)
- Type-safe with Callable type hints

### 2. Context-Aware Input/Select

**Pattern**: Provide rich context for intelligent handling

**Benefits**:
- Enables special logic for search bars
- Handles date pickers with format hints
- Detects text CAPTCHAs
- Manages location inputs with auto-completion

### 3. Multi-Level Dropdown Detection

**Pattern**: Track dropdown menu state for completion

**Benefits**:
- Handles complex dropdowns
- Detects multi-level selection
- Prevents premature completion

### 4. Cursor Visualization Management

**Pattern**: Hide cursor during screenshot

**Benefits**:
- Prevents interference with LLM analysis
- Maintains cursor for user debugging
- Graceful fallback on errors

### 5. Verification Status Evolution

**Pattern**: Support both new and legacy formats

**Benefits**:
- Backward compatibility
- Gradual migration path
- Property-based access for clean API

---

## Performance Optimizations

### 1. Action Caching

**Pattern**: Cache action results to avoid re-execution

**Impact**: Reduces redundant browser operations

### 2. Element Tree Incremental Updates

**Pattern**: Update only changed elements

**Impact**: Faster element tree refresh

### 3. Screenshot Optimization

**Pattern**: Hide cursor during screenshot

**Impact**: Cleaner screenshots for LLM analysis

---

## Testing Considerations

### Test Scenarios

1. **All 22 action types** - Verify each action works correctly
2. **Error handling** - Verify all exception types are caught
3. **Element location** - Verify element tree integration
4. **Multi-level dropdowns** - Verify completion detection
5. **Context-aware input** - Verify intelligent handling
6. **Action caching** - Verify cache hit/miss logic

### Test Commands

```bash
# Run action tests
python -m pytest tests/unit/test_actions.py -v

# Run handler tests
python -m pytest tests/unit/test_action_handler.py -v

# Run parsing tests
python -m pytest tests/unit/test_parse_actions.py -v
```

---

## References

- **Action Types**: `webeye/actions/action_types.py` (57 lines)
- **Action Models**: `webeye/actions/actions.py` (402 lines)
- **Action Handler**: `webeye/actions/handler.py` (4,931 lines)
- **Handler Utils**: `webeye/actions/handler_utils.py`
- **Action Parsing**: `webeye/actions/parse_actions.py`
- **Action Responses**: `webeye/actions/responses.py`
- **Action Caching**: `webeye/actions/caching.py`
- **Element Tree**: `webeye/scraper/scraped_page.py`
