# Skyvern Data Extraction

**Repository**: Skyvern  
**Location**: `skyvern/forge/sdk/`  
**Key Files**:
- `schemas/tasks.py` - Extraction goal and schema
- `api/llm/schema_validator.py` - Schema validation
- `workflow/models/block.py` - Extraction block

---

## Overview

Skyvern's Data Extraction system extracts structured data from web pages using LLM-powered extraction with schema validation. It supports complex nested schemas, dynamic parameters, and extraction caching for performance.

---

## Extraction Goal

**Location**: `schemas/tasks.py` (lines 52-56)

```python
data_extraction_goal: str | None = Field(
    default=None,
    description="The user's goal for data extraction.",
    examples=["Extract the quote price"],
)
```

**Purpose**: Describe what data to extract

---

## Extraction Schema

**Location**: `schemas/tasks.py` (lines 81-84)

```python
extracted_information_schema: dict[str, Any] | list | str | None = Field(
    default=None,
    description="The requested schema of the extracted information.",
)
```

**Purpose**: Define expected output structure

### Schema Formats

**JSON Schema**:
```json
{
  "price": "string",
  "product_name": "string",
  "availability": "boolean"
}
```

**List Schema**:
```json
["string", "string", "boolean"]
```

**String Schema**:
```json
"Extract the price as a string"
```

---

## Schema Validation

**Location**: `forge/sdk/api/llm/schema_validator.py`

**Purpose**: Validate extracted data against schema

**Features**:
- Schema validation
- Type checking
- Required field validation
- Nested schema validation

### validate_and_fill_extraction_result

```python
def validate_and_fill_extraction_result(
    extraction_result: dict[str, Any],
    schema: dict[str, Any],
) -> dict[str, Any]:
    """Validate and fill extraction result against schema."""
```

---

## Extraction Block

**Location**: `workflow/models/block.py` (lines 5571-5577)

```python
class ExtractionBlock(BaseTaskBlock):
    block_type: Literal[BlockType.EXTRACTION] = BlockType.EXTRACTION
```

**Purpose**: Dedicated block for data extraction

**Features**:
- Extraction goal
- Extraction schema
- Output parameters
- Validation logic

---

## Extraction Caching

**Location**: `forge/sdk/cache/extraction_cache.py`

**Purpose**: Cache extraction results for performance

**Features**:
- Result caching
- Cache invalidation
- Cache warming
- Shadow caching

---

## Key Patterns

### 1. Schema-Based Extraction

**Pattern**: Use schema to guide extraction

**Benefits**:
- Structured output
- Type safety
- Validation

### 2. Schema Validation

**Pattern**: Validate extracted data

**Benefits**:
- Data quality
- Error detection
- Type consistency

### 3. Extraction Caching

**Pattern**: Cache extraction results

**Benefits**:
- Performance
- Reduced LLM calls
- Cost savings

---

## Performance Optimizations

### 1. Extraction Caching

**Impact**: Reduced LLM calls

### 2. Schema Validation

**Impact**: Fast validation

### 3. Shadow Caching

**Impact**: Cache warming

---

## Testing Considerations

### Test Scenarios

1. **Schema validation** - Verify validation logic
2. **Extraction accuracy** - Verify extraction quality
3. **Nested schemas** - Verify complex schemas
4. **Extraction caching** - Verify cache hit/miss

### Test Commands

```bash
# Run extraction tests
python -m pytest tests/unit/test_extraction.py -v

# Run schema validation tests
python -m pytest tests/unit/test_schema_validation.py -v
```

---

## References

- **Extraction Schema**: `schemas/tasks.py`
- **Schema Validator**: `forge/sdk/api/llm/schema_validator.py`
- **Extraction Block**: `workflow/models/block.py`
- **Extraction Cache**: `forge/sdk/cache/extraction_cache.py`
