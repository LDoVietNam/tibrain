# CLI Flag Management System

> **Component**: `internal/cli/flags.go`
> **Inspired by**: OpenCode-Dev flag.ts
> **Status**: ✅ Implemented & Tested

## Overview

The Flag Management System provides a centralized, type-safe way to define and manage CLI flags. It replaces package-level variables with a registry-based approach, enabling better organization, validation, and environment variable support.

## Architecture

### Core Components

#### Flag Struct
```go
type Flag struct {
    Name        string      // Flag name (e.g., "port")
    Description string      // Human-readable description
    Type        FlagType    // Data type (string, int, bool, string_array)
    Default     interface{} // Default value
    Required    bool        // Whether flag is required
    EnvVar      string      // Environment variable name (e.g., "TI_PORT")
    Short       string      // Short flag name (e.g., "p")
    Category    string      // Category for grouping (e.g., "server")
}
```

#### FlagRegistry
```go
type FlagRegistry struct {
    flags map[string]*Flag
}
```

Thread-safe registry for managing flag definitions with CRUD operations.

### Flag Types

| Type | Description | Example |
|------|-------------|---------|
| `FlagTypeString` | String value | `"localhost"` |
| `FlagTypeInt` | Integer value | `8080` |
| `FlagTypeBool` | Boolean value | `true` |
| `FlagTypeStringArray` | Comma-separated strings | `["a", "b", "c"]` |

## Usage

### Basic Registration

```go
registry := NewFlagRegistry()

flag := &Flag{
    Name:        "port",
    Description: "Server port",
    Type:        FlagTypeInt,
    Default:     8080,
    Required:    false,
    EnvVar:      "TI_PORT",
    Short:       "p",
    Category:    "server",
}

err := registry.Register(flag)
if err != nil {
    log.Fatal(err)
}
```

### Environment Variable Support

```go
// Get value from environment variable
value, err := registry.GetEnvValue(flag)
// If TI_PORT=3000 is set, returns 3000
// Otherwise, returns default value (8080)
```

### Validation

```go
// Validate flag value against type
err := registry.Validate(flag, "8080")
if err != nil {
    // Handle validation error
}
```

### Retrieval

```go
// Get specific flag
flag, exists := registry.Get("port")

// List all flags
flags := registry.List()

// List by category
serverFlags := registry.ListByCategory("server")
```

## Error Handling

The system provides specific error types:

| Error | Description |
|-------|-------------|
| `ErrFlagNameRequired` | Flag name cannot be empty |
| `ErrFlagValueRequired` | Required flag has no value |
| `ErrFlagInvalidType` | Value doesn't match flag type |
| `ErrFlagAlreadyExists` | Flag with same name already registered |
| `ErrFlagNotFound` | Flag not found in registry |

## Best Practices

### 1. Use Descriptive Names
```go
// Good
flag.Name = "max-connections"

// Avoid
flag.Name = "mc"
```

### 2. Set Environment Variables
```go
flag.EnvVar = "TI_" + strings.ToUpper(strings.ReplaceAll(flag.Name, "-", "_"))
// Example: "TI_MAX_CONNECTIONS"
```

### 3. Use Categories for Grouping
```go
flag.Category = "server"   // Server-related flags
flag.Category = "auth"     // Authentication flags
flag.Category = "logging"  // Logging flags
```

### 4. Provide Clear Descriptions
```go
flag.Description = "Maximum number of concurrent connections (default: 100)"
```

### 5. Set Appropriate Defaults
```go
flag.Default = 100  // Sensible default
flag.Required = false  // Make optional if reasonable default exists
```

## Integration with Cobra

### Manual Integration

```go
var registry *FlagRegistry

func init() {
    registry = NewFlagRegistry()

    // Register flags
    registry.Register(&Flag{
        Name:        "port",
        Type:        FlagTypeInt,
        Default:     8080,
        EnvVar:      "TI_PORT",
    })

    // Bind to Cobra
    cmd.Flags().IntP("port", "p", 8080, "Server port")
}
```

### With CommandWrapper (Recommended)

```go
wrapper := NewCommandWrapper(cmd)
    .WithErrorHandler(DefaultErrorHandler)
    .WithValidation(func(cmd *cobra.Command, args []string) error {
        // Validate flags using registry
        return nil
    })
```

## Examples

### Example 1: Server Configuration

```go
registry := NewFlagRegistry()

flags := []*Flag{
    {
        Name:        "host",
        Description: "Server host address",
        Type:        FlagTypeString,
        Default:     "localhost",
        EnvVar:      "TI_HOST",
        Category:    "server",
    },
    {
        Name:        "port",
        Description: "Server port",
        Type:        FlagTypeInt,
        Default:     8080,
        EnvVar:      "TI_PORT",
        Category:    "server",
    },
    {
        Name:        "debug",
        Description: "Enable debug mode",
        Type:        FlagTypeBool,
        Default:     false,
        EnvVar:      "TI_DEBUG",
        Category:    "logging",
    },
}

for _, flag := range flags {
    if err := registry.Register(flag); err != nil {
        log.Fatal(err)
    }
}
```

### Example 2: MCP Server Configuration

```go
registry := NewFlagRegistry()

flag := &Flag{
    Name:        "transport",
    Description: "MCP transport type (stdio, http, websocket)",
    Type:        FlagTypeString,
    Default:     "stdio",
    Required:    true,
    EnvVar:      "TI_MCP_TRANSPORT",
    Category:    "mcp",
}

registry.Register(flag)
```

### Example 3: Array Flags

```go
flag := &Flag{
    Name:        "allowed-origins",
    Description: "CORS allowed origins",
    Type:        FlagTypeStringArray,
    Default:     []string{"http://localhost:3000"},
    EnvVar:      "TI_ALLOWED_ORIGINS",
    Category:    "security",
}

// Environment variable: TI_ALLOWED_ORIGINS=http://localhost:3000,https://example.com
// Result: []string{"http://localhost:3000", "https://example.com"}
```

## Testing

```go
func TestFlagRegistry(t *testing.T) {
    registry := NewFlagRegistry()

    flag := &Flag{
        Name:  "test",
        Type:  FlagTypeString,
    }

    err := registry.Register(flag)
    if err != nil {
        t.Fatal(err)
    }

    retrieved, exists := registry.Get("test")
    if !exists {
        t.Fatal("Flag not found")
    }

    if retrieved.Name != flag.Name {
        t.Errorf("Expected %s, got %s", flag.Name, retrieved.Name)
    }
}
```

## Migration from Package Variables

### Before (Old Pattern)

```go
var (
    port int
    host string
)

func init() {
    cmd.Flags().IntVarP(&port, "port", "p", 8080, "Server port")
    cmd.Flags().StringVarP(&host, "host", "H", "localhost", "Server host")
}
```

### After (New Pattern)

```go
var registry *FlagRegistry

func init() {
    registry = NewFlagRegistry()

    registry.Register(&Flag{
        Name:    "port",
        Type:    FlagTypeInt,
        Default: 8080,
        EnvVar:  "TI_PORT",
    })

    registry.Register(&Flag{
        Name:    "host",
        Type:    FlagTypeString,
        Default: "localhost",
        EnvVar:  "TI_HOST",
    })
}
```

## Benefits

1. **Centralized Management**: All flag definitions in one place
2. **Type Safety**: Automatic type validation
3. **Environment Variables**: Built-in support for configuration via env vars
4. **Documentation**: Structured metadata for each flag
5. **Validation**: Type checking and required field validation
6. **Categories**: Logical grouping of related flags
7. **Thread-Safe**: Safe for concurrent access

## See Also

- [Error Handling Patterns](CLI_ERROR_HANDLING.md)
- [Command Validation](CLI_COMMAND_VALIDATION.md)
- [CLI Guide](../CLI_GUIDE.md)
