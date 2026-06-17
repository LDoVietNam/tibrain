# Skyvern MCP Integration

**Repository**: Skyvern  
**Location**: `skyvern/forge/sdk/copilot/mcp_adapter.py`  
**Related Files**:
- `copilot/tools.py` - Tool registration
- `copilot/runtime.py` - Agent runtime

---

## Overview

Skyvern's MCP (Model Context Protocol) Integration enables the copilot agent to use external tools and services through the MCP standard. It provides a schema overlay system for parameter mapping and tool invocation.

---

## MCP Adapter

**Location**: `forge/sdk/copilot/mcp_adapter.py`

**Purpose**: Integrate MCP tools into copilot

**Features**:
- Tool registration
- Schema overlay
- Parameter mapping
- Tool invocation

### SchemaOverlay

```python
class SchemaOverlay(BaseModel):
    """Overlay schema for MCP tool parameters."""
```

**Purpose**: Map MCP tool parameters to copilot tool parameters

---

## Tool Registration

**Location**: `copilot/tools.py`

**Purpose**: Register MCP tools with copilot

**Features**:
- Tool discovery
- Tool registration
- Tool invocation
- Error handling

---

## Key Patterns

### 1. Schema Overlay

**Pattern**: Overlay schema for parameter mapping

**Benefits**:
- Flexible parameter mapping
- Type safety
- Validation

### 2. Tool Registration

**Pattern**: Register MCP tools dynamically

**Benefits**:
- Extensible tool set
- Dynamic discovery
- Easy integration

---

## Testing Considerations

### Test Scenarios

1. **Tool registration** - Verify tool discovery
2. **Schema overlay** - Verify parameter mapping
3. **Tool invocation** - Verify tool execution
4. **Error handling** - Verify error recovery

### Test Commands

```bash
# Run MCP tests
python -m pytest tests/unit/test_mcp.py -v
```

---

## References

- **MCP Adapter**: `forge/sdk/copilot/mcp_adapter.py`
- **Tool Registration**: `copilot/tools.py`
- **Agent Runtime**: `copilot/runtime.py`
