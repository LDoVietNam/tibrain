# Skyvern Element Tree Building

**Repository**: Skyvern  
**Location**: `skyvern/webeye/scraper/`  
**Key Files**:
- `scraped_page.py` - Element tree builder and management
- `scraper.py` - Page scraping logic
- `domUtils.js` - JavaScript DOM utilities
- `utils/dom.py` - Python DOM utilities
- `utils/page.py` - Page utilities

---

## Overview

Skyvern's Element Tree Building system constructs a structured representation of the page DOM for LLM consumption. It features economy mode for token efficiency, progressive truncation for large pages, and incremental updates for performance.

---

## Element Tree Builder

**Location**: `webeye/scraper/scraped_page.py`

### ElementTreeBuilder

**Purpose**: Build element tree from page DOM

**Features**:
- Economy mode for token efficiency
- Progressive truncation for large pages
- Incremental updates for performance
- Cleanup functions for LLM consumption

### Element Tree Format

**Purpose**: Define element tree representation format

**Formats**:
- JSON - Structured JSON representation
- HTML - HTML string representation
- Custom - Custom format for specific use cases

---

## Economy Mode

**Purpose**: Reduce token consumption by omitting non-essential elements

**Strategy**:
- Omit invisible elements
- Omit decorative elements
- Omit script/style tags
- Preserve interactive elements
- Preserve text content

---

## Progressive Truncation

**Purpose**: Handle large pages by truncating progressively

**Strategy**:
1. Start with full tree
2. If token budget exceeded, truncate by 20%
3. Repeat until within budget
4. Preserve critical elements (forms, inputs, buttons)

**Location**: `webeye/scraper/scraper.py`

```python
def trim_element_tree(element_tree: dict, max_tokens: int) -> dict:
    """Progressively trim element tree to fit token budget."""
```

---

## Incremental Scraping

**Purpose**: Update only changed elements for performance

**Strategy**:
- Track element hashes
- Compare with previous state
- Update only changed elements
- Preserve stable elements

**Location**: `webeye/scraper/scraper.py`

```python
class IncrementalScrapePage:
    """Incremental page scraping with change detection."""
```

---

## DOM Utilities

**Location**: `webeye/utils/dom.py`

### DomUtil

**Purpose**: Python-side DOM manipulation utilities

**Features**:
- Element location
- Element selection
- Attribute extraction
- Text extraction

### InteractiveElement

**Purpose**: Represent interactive elements

**Features**:
- Element identification
- Interaction metadata
- Accessibility attributes

### SkyvernElement

**Purpose**: Skyvern-specific element wrapper

**Features**:
- Element metadata
- Skyvern ID tracking
- Interaction history

---

## JavaScript DOM Utilities

**Location**: `webeye/scraper/domUtils.js`

**Purpose**: Browser-side DOM manipulation

**Features**:
- Element tree construction
- Element hashing
- Change detection
- Accessibility attributes extraction

---

## Key Patterns

### 1. Economy Mode

**Pattern**: Omit non-essential elements to save tokens

**Benefits**:
- 30-50% token reduction
- Faster LLM processing
- Maintains critical information

### 2. Progressive Truncation

**Pattern**: Truncate incrementally to fit budget

**Benefits**:
- Preserves critical elements
- Handles arbitrary page sizes
- Predictable token usage

### 3. Incremental Updates

**Pattern**: Update only changed elements

**Benefits**:
- Faster updates
- Reduced processing
- Better performance

### 4. Element Hashing

**Pattern**: Hash elements for change detection

**Benefits**:
- Efficient change detection
- Minimal updates
- Stable element references

---

## Performance Optimizations

### 1. Element Hashing

**Impact**: Fast change detection

### 2. Incremental Updates

**Impact**: Reduced processing time

### 3. Economy Mode

**Impact**: 30-50% token reduction

### 4. Progressive Truncation

**Impact**: Predictable token usage

---

## Testing Considerations

### Test Scenarios

1. **Economy mode** - Verify token reduction
2. **Progressive truncation** - Verify critical elements preserved
3. **Incremental updates** - Verify change detection
4. **Element hashing** - Verify hash consistency
5. **Large pages** - Verify truncation logic

### Test Commands

```bash
# Run scraper tests
python -m pytest tests/unit/test_scraper.py -v

# Run element tree tests
python -m pytest tests/unit/test_element_tree.py -v
```

---

## References

- **Element Tree Builder**: `webeye/scraper/scraped_page.py`
- **Scraper**: `webeye/scraper/scraper.py`
- **DOM Utilities**: `webeye/utils/dom.py`
- **JS DOM Utils**: `webeye/scraper/domUtils.js`
- **Page Utils**: `webeye/utils/page.py`
