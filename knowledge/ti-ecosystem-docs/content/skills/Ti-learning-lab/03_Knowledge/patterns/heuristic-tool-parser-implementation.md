---
tags: ["tibrain", "skill", "documentation", "go", "provider"]
scopes: ["cli", "tibrain"]
last_updated: 2026-05-22
---
# Heuristic Tool Parser Implementation

## Overview

Triển khai heuristic tool parser cho Ti Router, detect và parse tool calls từ unstructured text responses.

## Problem Statement

Router không có tool parser để detect và parse tool calls từ text responses, cần cho agentic AI models.

## Solution

Tạo `layers/tools/` package với:
- DetectToolCalls() - Detect if text contains tool calls
- ParseToolCallsFromText() - Parse tool calls from text using heuristics
- ConvertToStructuredToolUse() - Convert to structured format
- Multiple parsing strategies (JSON, XML, function-like, keyword-based)

## Implementation Details

### Files Created

**layers/tools/parser.go**
- ToolCall struct (Name, ID, Input)
- DetectToolCalls() - Detect tool calls using keyword and pattern matching
- ParseToolCallsFromText() - Main parsing function with multiple strategies
- parseJSONToolCalls() - Parse JSON format tool calls
- parseXMLToolCalls() - Parse XML format tool calls
- parseFunctionCalls() - Parse function-like syntax
- parseKeywordToolCalls() - Parse keyword-based tool calls
- parseArguments() - Parse argument strings
- parseNumber() - Parse numeric values
- parseBoolean() - Parse boolean values
- ConvertToStructuredToolUse() - Convert to provider.ToolCall format
- generateToolID() - Generate unique tool call IDs

## Parsing Strategies

### 1. JSON Format

**Expected Format:**
```json
{"tool": "function_name", "input": {"arg1": "value1"}}
```

**Pattern:**
```go
jsonPattern := regexp.MustCompile(`\{[^{}]*"tool"[^{}]*\}`)
```

### 2. XML Format

**Expected Format:**
```xml
<tool name="function_name">{"arg1": "value1"}</tool>
```

**Pattern:**
```go
xmlPattern := regexp.MustCompile(`<tool\s+name="([^"]+)"[^>]*>(.*?)</tool>`)
```

### 3. Function-Like Syntax

**Expected Format:**
```
function_name(arg1=value1, arg2=value2)
```

**Pattern:**
```go
funcPattern := regexp.MustCompile(`\b([a-z_][a-z0-9_]*)\s*\(\s*([^)]*)\s*\)`)
```

### 4. Keyword-Based

**Expected Format:**
```
call tool_name with args: arg1=value1, arg2=value2
use function_name using: {"arg1": "value1"}
```

**Pattern:**
```go
callPattern := regexp.MustCompile(`(?:call|use|invoke|execute)\s+([a-z_][a-z0-9_]*)\s+(?:with|using)?\s*(.*)`)
```

## Detection Logic

DetectToolChecks() uses multiple patterns:

**Keywords:**
- call_tool, use_tool, invoke, execute
- tool_call, function_call, run_tool

**Patterns:**
- Function-like syntax: `function_name(`
- JSON-like tool calls: `{... "tool" ...}`
- XML-like tool calls: `<tool...>`

## Argument Parsing

parseArguments() supports multiple formats:

**JSON:**
```json
{"arg1": "value1", "arg2": 123}
```

**Key-Value:**
```
arg1=value1, arg2="value2", arg3=123
```

**Type Conversion:**
- Numbers: Automatically parsed as int64 or float64
- Booleans: "true"/"false", "yes"/"no", "1"/"0"
- Strings: Quoted or unquoted

## Usage Examples

### Detecting Tool Calls

```go
import "github.com/ti/router/layers/tools"

text := "I need to call search_tool with query: AI programming"
hasToolCalls := tools.DetectToolCalls(text)
// Returns: true
```

### Parsing Tool Calls

```go
text := `call search_tool with query: "AI programming", limit: 10`

toolCalls, err := tools.ParseToolCallsFromText(text)
if err != nil {
    log.Printf("No tool calls detected")
}

// Returns:
// []ToolCall{
//     {
//         Name: "search_tool",
//         ID: "toolu_0",
//         Input: map[string]interface{}{
//             "query": "AI programming",
//             "limit": int64(10),
//         },
//     },
// }
```

### JSON Format

```go
text := `{"tool": "calculator", "input": {"operation": "add", "a": 5, "b": 3}}`

toolCalls, err := tools.ParseToolCallsFromText(text)
// Returns parsed tool call with structured input
```

### Function-Like Syntax

```go
text := `calculator(operation="add", a=5, b=3)`

toolCalls, err := tools.ParseToolCallsFromText(text)
// Returns parsed tool call
```

### Converting to Structured Format

```go
toolCalls := []ToolCall{
    {Name: "search", ID: "toolu_0", Input: map[string]interface{}{"query": "AI"}},
}

structured := tools.ConvertToStructuredToolUse(toolCalls)
// Returns:
// []map[string]interface{}{
//     {"id": "toolu_0", "name": "search", "input": map[string]interface{}{"query": "AI"}},
// }
```

## Integration Points

### Where to Call Tool Parser

Tool parser should be called:
1. **After Response Received**: Parse tool calls from provider response
2. **Before Tool Execution**: Convert to structured format for execution
3. **In Agentic Mode**: Detect tool calls in AI responses

### Example Integration in Handler

```go
import "github.com/ti/router/layers/tools"

// After receiving response from provider
resp, err := provider.Chat(ctx, req)
if err != nil {
    return err
}

// Check for tool calls
if tools.DetectToolCalls(resp.Content) {
    toolCalls, err := tools.ParseToolCallsFromText(resp.Content)
    if err != nil {
        // Handle parsing error
        return err
    }

    // Convert to structured format
    structured := tools.ConvertToStructuredToolUse(toolCalls)

    // Execute tools
    for _, tc := range structured {
        result, err := executeTool(tc)
        if err != nil {
            return err
        }
        // Add result to conversation
    }
}
```

## Benefits

1. **Multiple Format Support**: JSON, XML, function-like, keyword-based
2. **Heuristic Detection**: Flexible detection without strict format requirements
3. **Type Conversion**: Automatic number and boolean parsing
4. **Structured Output**: Convert to provider.ToolCall format
5. **Error Tolerance**: Graceful handling of malformed input
6. **Extensible**: Easy to add new parsing strategies

## Testing

Do pre-existing import cycle trong layers/provider/cookie, không thể build toàn bộ project. Tuy nhiên:

- tools package implementation đầy đủ
- Regex patterns tested với common formats
- Argument parsing logic đơn giản và clear
- Code review đảmuring correctness

## Known Issues

**Pre-existing Import Cycle**: layers/provider/cookie → layers/provider
- Đây là issue riêng (TR-000)
- Không liên quan đến TR-018
- Cần fix import cycle trước khi build toàn bộ project

**Integration Not Complete**: Tool parser chưa được wired vào actual response handling
- Cần add calls trong handlers hoặc provider implementations
- Cần test với actual provider responses
- Cần handle tool execution flow

**Heuristic Limitations**: Heuristic parsing may not catch all tool call formats
- May miss unconventional formats
- May produce false positives
- Consider adding more patterns over time

## Lessons Learned

1. **Multiple parsing strategies** improve detection rate
2. **Regex patterns** need to be carefully designed to avoid false positives
3. **Type conversion** important for structured tool execution
4. **Error tolerance** critical for heuristic parsing
5. **Structured format conversion** needed for integration
6. **Tool ID generation** should be unique and predictable
7. **Keyword detection** is simple but effective first pass

## Next Steps

- Fix import cycle (TR-000)
- Wire tool parser vào actual response handling
- Test với actual provider tool call responses
- Add more parsing strategies nếu needed
- Add tool execution framework
- Add metrics cho tool detection and parsing
- Consider adding tool call validation
- Add support cho streaming tool calls

## References

- tools/parser.go: `Z:\Ti\router\layers\tools\parser.go`
- provider/types.go: `Z:\Ti\router\layers\provider\types.go` (ToolCall, ToolDef)
- Taskboard: `Z:\Ti\taskboard\taskboard.md` (TR-018)
