# CLI Tool System

> **Component**: `internal/tool/`
> **Inspired by**: OpenCode-Dev tool architecture
> **Status**: ✅ Implemented & Tested

## Overview

The Tool System provides a pluggable, extensible architecture for implementing tools that can be executed by the Ti CLI. It replaces monolithic command logic with composable, reusable tool components that can be registered, discovered, and executed dynamically.

## Architecture

### Core Components

#### Tool Interface
```go
type Tool interface {
    Name() string
    Description() string
    Schema() map[string]interface{}
    Execute(ctx context.Context, params map[string]interface{}) (*Result, error)
}
```

Base interface that all tools must implement.

#### Result Struct
```go
type Result struct {
    Success bool                   `json:"success"`
    Data    interface{}            `json:"data,omitempty"`
    Error   string                 `json:"error,omitempty"`
    Metadata map[string]interface{} `json:"metadata,omitempty"`
}
```

Standardized result format for tool execution.

#### ToolSpec Struct
```go
type ToolSpec struct {
    Name        string                 `json:"name"`
    Description string                 `json:"description"`
    Schema      map[string]interface{} `json:"schema"`
    Category    string                 `json:"category,omitempty"`
}
```

Tool specification for discovery and validation.

#### Registry
```go
type Registry struct {
    tools map[string]Tool
    mu    sync.RWMutex
}
```

Thread-safe registry for managing tool registration and discovery.

## Built-in Tools

### ReadTool
Reads a file from the filesystem.

**Parameters:**
- `file_path` (string, required): Absolute path to the file
- `offset` (int, optional): Line number to start reading from (1-based)
- `limit` (int, optional): Number of lines to read

**Example:**
```go
result, err := readTool.Execute(ctx, map[string]interface{}{
    "file_path": "/path/to/file.txt",
    "offset":    10,
    "limit":     20,
})
```

### WriteTool
Writes content to a file (overwrites if exists).

**Parameters:**
- `file_path` (string, required): Absolute path to the file
- `content` (string, required): Content to write

**Example:**
```go
result, err := writeTool.Execute(ctx, map[string]interface{}{
    "file_path": "/path/to/file.txt",
    "content":   "Hello, World!",
})
```

### GrepTool
Searches for patterns in files using regex.

**Parameters:**
- `pattern` (string, required): Regex pattern to search
- `path` (string, optional): Directory to search (default: current)
- `glob_pattern` (string, optional): Glob pattern for file filtering
- `output_mode` (string, optional): Output mode (content, files_with_matches, count)
- `case_insensitive` (bool, optional): Case-insensitive search
- `max_results` (int, optional): Maximum results to return
- `context_lines` (int, optional): Context lines before/after matches

**Example:**
```go
result, err := grepTool.Execute(ctx, map[string]interface{}{
    "pattern":        "func.*Tool",
    "path":           "/path/to/code",
    "output_mode":    "content",
    "case_insensitive": true,
})
```

### ExecTool
Executes shell commands.

**Parameters:**
- `command` (string, required): Shell command to execute
- `shell_id` (string, optional): Shell session ID for persistent sessions
- `run_in_background` (bool, optional): Run in background mode
- `timeout` (int, optional): Timeout in milliseconds

**Example:**
```go
result, err := execTool.Execute(ctx, map[string]interface{}{
    "command": "ls -la",
    "timeout": 5000,
})
```

### EditTool
Performs exact string replacements in files.

**Parameters:**
- `file_path` (string, required): Absolute path to the file
- `old_string` (string, required): Text to replace
- `new_string` (string, required): Replacement text
- `replace_all` (bool, optional): Replace all occurrences

**Example:**
```go
result, err := editTool.Execute(ctx, map[string]interface{}{
    "file_path":  "/path/to/file.txt",
    "old_string": "old text",
    "new_string": "new text",
    "replace_all": true,
})
```

### MultiEditTool
Performs multiple edits in a single operation.

**Parameters:**
- `file_path` (string, required): Absolute path to the file
- `edits` (array, required): Array of edit operations

**Example:**
```go
result, err := multiEditTool.Execute(ctx, map[string]interface{}{
    "file_path": "/path/to/file.txt",
    "edits": []map[string]interface{}{
        {"old_string": "old1", "new_string": "new1"},
        {"old_string": "old2", "new_string": "new2"},
    },
})
```

## Usage

### Creating a Custom Tool

```go
package custom

import (
    "context"
    "github.com/ti/cli/internal/tool"
)

type MyTool struct{}

func (t *MyTool) Name() string {
    return "my_tool"
}

func (t *MyTool) Description() string {
    return "My custom tool description"
}

func (t *MyTool) Schema() map[string]interface{} {
    return map[string]interface{}{
        "type": "object",
        "properties": map[string]interface{}{
            "input": map[string]interface{}{
                "type":        "string",
                "description": "Input parameter",
            },
        },
        "required": []string{"input"},
    }
}

func (t *MyTool) Execute(ctx context.Context, params map[string]interface{}) (*tool.Result, error) {
    input, ok := params["input"].(string)
    if !ok {
        return &tool.Result{
            Success: false,
            Error:   "input must be a string",
        }, nil
    }

    // Tool logic here
    output := "processed: " + input

    return &tool.Result{
        Success: true,
        Data:    output,
    }, nil
}
```

### Registering Tools

```go
import (
    "github.com/ti/cli/internal/tool"
    "github.com/ti/cli/internal/tool/builtin"
)

func main() {
    registry := tool.NewRegistry()

    // Register built-in tools
    registry.Register(builtin.NewReadTool())
    registry.Register(builtin.NewWriteTool())
    registry.Register(builtin.NewGrepTool())
    registry.Register(builtin.NewExecTool())
    registry.Register(builtin.NewEditTool())
    registry.Register(builtin.NewMultiEditTool())

    // Register custom tools
    registry.Register(&custom.MyTool())
}
```

### Executing Tools

```go
// Get tool from registry
tool, exists := registry.Get("read")
if !exists {
    log.Fatal("Tool not found")
}

// Execute tool
result, err := tool.Execute(ctx, map[string]interface{}{
    "file_path": "/path/to/file.txt",
})

if err != nil {
    log.Fatal(err)
}

if !result.Success {
    log.Fatal(result.Error)
}

// Use result
data := result.Data
```

### Tool Discovery

```go
// List all tool names
names := registry.List()

// Get tool specifications
specs := registry.Specs()

// Check if tool exists
exists := registry.Exists("read")

// Get tool count
count := registry.Count()
```

## Integration with Executor

The tool system integrates with the copilot executor for AI-driven tool execution.

```go
import (
    "github.com/ti/cli/internal/copilot"
)

func setupExecutor() *copilot.Executor {
    registry := tool.NewRegistry()

    // Register tools
    registry.Register(builtin.NewReadTool())
    registry.Register(builtin.NewWriteTool())
    // ... more tools

    // Create executor with tool registry
    executor := copilot.NewExecutor()
    executor.SetToolRegistry(registry)

    return executor
}
```

## JSON Schema Format

Tools use JSON Schema for parameter validation:

```go
func (t *MyTool) Schema() map[string]interface{} {
    return map[string]interface{}{
        "type": "object",
        "properties": map[string]interface{}{
            "param1": map[string]interface{}{
                "type":        "string",
                "description": "Parameter description",
                "default":     "default_value",
            },
            "param2": map[string]interface{}{
                "type":        "integer",
                "description": "Integer parameter",
                "minimum":     0,
                "maximum":     100,
            },
            "param3": map[string]interface{}{
                "type":        "boolean",
                "description": "Boolean parameter",
            },
            "param4": map[string]interface{}{
                "type":        "array",
                "description": "Array parameter",
                "items": map[string]interface{}{
                    "type": "string",
                },
            },
        },
        "required": []string{"param1", "param2"},
    }
}
```

## Best Practices

### 1. Use Context for Cancellation

```go
func (t *MyTool) Execute(ctx context.Context, params map[string]interface{}) (*tool.Result, error) {
    select {
    case <-ctx.Done():
        return &tool.Result{
            Success: false,
            Error:   "operation cancelled",
        }, ctx.Err()
    default:
        // Proceed with execution
    }
}
```

### 2. Validate Parameters

```go
func (t *MyTool) Execute(ctx context.Context, params map[string]interface{}) (*tool.Result, error) {
    input, ok := params["input"].(string)
    if !ok {
        return &tool.Result{
            Success: false,
            Error:   "input must be a string",
        }, nil
    }

    if input == "" {
        return &tool.Result{
            Success: false,
            Error:   "input cannot be empty",
        }, nil
    }

    // Proceed with logic
}
```

### 3. Return Structured Results

```go
return &tool.Result{
    Success: true,
    Data: map[string]interface{}{
        "output": processedData,
        "stats":  statistics,
    },
    Metadata: map[string]interface{}{
        "execution_time": time.Since(start),
        "version":       "1.0.0",
    },
}
```

### 4. Handle Errors Gracefully

```go
func (t *MyTool) Execute(ctx context.Context, params map[string]interface{}) (*tool.Result, error) {
    data, err := doSomething()
    if err != nil {
        return &tool.Result{
            Success: false,
            Error:   fmt.Sprintf("operation failed: %v", err),
        }, nil
    }

    return &tool.Result{
        Success: true,
        Data:    data,
    }, nil
}
```

### 5. Provide Clear Descriptions

```go
func (t *MyTool) Description() string {
    return "Process input data and return transformed output"
}
```

## Examples

### Example 1: File Watcher Tool

```go
type WatcherTool struct{}

func (t *WatcherTool) Name() string {
    return "watch"
}

func (t *WatcherTool) Description() string {
    return "Watch a file for changes"
}

func (t *WatcherTool) Schema() map[string]interface{} {
    return map[string]interface{}{
        "type": "object",
        "properties": map[string]interface{}{
            "file_path": map[string]interface{}{
                "type":        "string",
                "description": "Path to file to watch",
            },
            "timeout": map[string]interface{}{
                "type":        "integer",
                "description": "Timeout in milliseconds",
            },
        },
        "required": []string{"file_path"},
    }
}

func (t *WatcherTool) Execute(ctx context.Context, params map[string]interface{}) (*tool.Result, error) {
    filePath, _ := params["file_path"].(string)
    timeout, _ := params["timeout"].(int)

    // Watch logic...
    return &tool.Result{
        Success: true,
        Data:    "File changed detected",
    }, nil
}
```

### Example 2: HTTP Request Tool

```go
type HTTPTool struct{}

func (t *HTTPTool) Name() string {
    return "http_request"
}

func (t *HTTPTool) Description() string {
    return "Make HTTP requests"
}

func (t *HTTPTool) Schema() map[string]interface{} {
    return map[string]interface{}{
        "type": "object",
        "properties": map[string]interface{}{
            "url": map[string]interface{}{
                "type":        "string",
                "description": "URL to request",
            },
            "method": map[string]interface{}{
                "type":        "string",
                "description": "HTTP method (GET, POST, etc.)",
            },
            "headers": map[string]interface{}{
                "type":        "object",
                "description": "HTTP headers",
            },
            "body": map[string]interface{}{
                "type":        "string",
                "description": "Request body",
            },
        },
        "required": []string{"url", "method"},
    }
}

func (t *HTTPTool) Execute(ctx context.Context, params map[string]interface{}) (*tool.Result, error) {
    url, _ := params["url"].(string)
    method, _ := params["method"].(string)

    // HTTP request logic...
    return &tool.Result{
        Success: true,
        Data:    responseData,
    }, nil
}
```

## Testing

```go
func TestReadTool(t *testing.T) {
    tool := builtin.NewReadTool()

    // Test name
    if tool.Name() != "read" {
        t.Errorf("Expected name 'read', got '%s'", tool.Name())
    }

    // Test schema
    schema := tool.Schema()
    if schema["type"] != "object" {
        t.Error("Schema should be an object")
    }

    // Test execution
    result, err := tool.Execute(context.Background(), map[string]interface{}{
        "file_path": "/tmp/test.txt",
    })

    if err != nil {
        t.Fatalf("Execute failed: %v", err)
    }

    if !result.Success {
        t.Errorf("Expected success, got error: %s", result.Error)
    }
}
```

## Migration from Direct Implementation

### Before (Old Pattern)

```go
func readFile(path string) (string, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        return "", err
    }
    return string(data), nil
}
```

### After (New Pattern)

```go
tool := builtin.NewReadTool()
result, err := tool.Execute(ctx, map[string]interface{}{
    "file_path": path,
})

if err != nil {
    return "", err
}

if !result.Success {
    return "", fmt.Errorf(result.Error)
}

data, ok := result.Data.(string)
if !ok {
    return "", fmt.Errorf("unexpected data type")
}

return data, nil
```

## Benefits

1. **Extensibility**: Easy to add new tools
2. **Reusability**: Tools can be used across commands
3. **Discovery**: Dynamic tool discovery and listing
4. **Validation**: JSON Schema for parameter validation
5. **Standardization**: Consistent interface and result format
6. **Thread-Safe**: Safe for concurrent access
7. **AI Integration**: Seamless integration with AI-driven execution

## See Also

- [Flag Management](CLI_FLAG_MANAGEMENT.md)
- [Error Handling](CLI_ERROR_HANDLING.md)
- [Command Validation](CLI_COMMAND_VALIDATION.md)
- [CLI Guide](../CLI_GUIDE.md)
