# CLI Architecture Examples

> **Purpose**: Practical examples for developers using the new CLI architecture
> **Components**: Flag Management, Error Handling, Command Validation, Tool System, MCP OAuth

## Table of Contents

1. [Complete Command Example](#complete-command-example)
2. [Flag Management Examples](#flag-management-examples)
3. [Error Handling Examples](#error-handling-examples)
4. [Command Validation Examples](#command-validation-examples)
5. [Tool System Examples](#tool-system-examples)
6. [MCP OAuth Examples](#mcp-oauth-examples)
7. [Integration Patterns](#integration-patterns)

---

## Complete Command Example

### Example: Server Management Command

```go
package cmd

import (
    "context"
    "fmt"
    "os"

    "github.com/spf13/cobra"
    "github.com/ti/cli/internal/cli"
)

var (
    registry *cli.FlagRegistry
)

func init() {
    // Initialize flag registry
    registry = cli.NewFlagRegistry()

    // Register flags
    registry.Register(&cli.Flag{
        Name:        "host",
        Description: "Server host address",
        Type:        cli.FlagTypeString,
        Default:     "localhost",
        EnvVar:      "TI_SERVER_HOST",
        Category:    "server",
    })

    registry.Register(&cli.Flag{
        Name:        "port",
        Description: "Server port",
        Type:        cli.FlagTypeInt,
        Default:     8080,
        EnvVar:      "TI_SERVER_PORT",
        Category:    "server",
    })

    registry.Register(&cli.Flag{
        Name:        "debug",
        Description: "Enable debug mode",
        Type:        cli.FlagTypeBool,
        Default:     false,
        EnvVar:      "TI_DEBUG",
        Category:    "logging",
    })
}

var serverCmd = &cobra.Command{
    Use:   "server [action]",
    Short: "Manage Ti server",
    Long:  "Start, stop, and manage the Ti server",
    Args:  cobra.MinimumNArgs(1),
    RunE:  runServerCommand,
}

func runServerCommand(cmd *cobra.Command, args []string) error {
    action := args[0]

    // Create validator
    validator := cli.NewCommandValidator()

    // Add validation rules
    validator.AddRule("action", cli.ValidationRule{
        Name:      "enum",
        Validator: &cli.EnumValidator{Allowed: []interface{}{"start", "stop", "restart", "status"}},
        Message:   "Action must be one of: start, stop, restart, status",
    })

    // Validate action
    if err := validator.Validate("action", action); err != nil {
        return err
    }

    // Get flag values
    host, _ := cmd.Flags().GetString("host")
    port, _ := cmd.Flags().GetInt("port")
    debug, _ := cmd.Flags().GetBool("debug")

    // Execute action
    switch action {
    case "start":
        return startServer(host, port, debug)
    case "stop":
        return stopServer()
    case "restart":
        return restartServer(host, port, debug)
    case "status":
        return getServerStatus()
    default:
        return cli.NewCommandNotFoundError(action)
    }
}

func startServer(host string, port int, debug bool) error {
    // Validation
    if port < 1 || port > 65535 {
        return cli.NewValidationError("port", "must be between 1 and 65535").
            WithDetails(fmt.Sprintf("Provided: %d", port))
    }

    // Start server logic
    fmt.Printf("Starting server on %s:%d (debug: %v)\n", host, port, debug)
    return nil
}

func stopServer() error {
    fmt.Println("Stopping server")
    return nil
}

func restartServer(host string, port int, debug bool) error {
    if err := stopServer(); err != nil {
        return cli.NewCLIError(cli.ErrCodeExecutionError, "failed to stop server").Wrap(err)
    }
    return startServer(host, port, debug)
}

func getServerStatus() error {
    fmt.Println("Server status: running")
    return nil
}
```

---

## Flag Management Examples

### Example 1: Basic Flag Registration

```go
package main

import (
    "github.com/ti/cli/internal/cli"
)

func main() {
    registry := cli.NewFlagRegistry()

    // Register a simple flag
    flag := &cli.Flag{
        Name:        "verbose",
        Description: "Enable verbose output",
        Type:        cli.FlagTypeBool,
        Default:     false,
        EnvVar:      "TI_VERBOSE",
    }

    err := registry.Register(flag)
    if err != nil {
        panic(err)
    }

    // Retrieve flag
    retrieved, exists := registry.Get("verbose")
    if exists {
        fmt.Printf("Flag: %s, Type: %s\n", retrieved.Name, retrieved.Type)
    }
}
```

### Example 2: Complex Flag Configuration

```go
package main

import (
    "github.com/ti/cli/internal/cli"
)

func setupFlags() *cli.FlagRegistry {
    registry := cli.NewFlagRegistry()

    // Server configuration flags
    serverFlags := []*cli.Flag{
        {
            Name:        "host",
            Description: "Server host address",
            Type:        cli.FlagTypeString,
            Default:     "localhost",
            EnvVar:      "TI_HOST",
            Category:    "server",
        },
        {
            Name:        "port",
            Description: "Server port",
            Type:        cli.FlagTypeInt,
            Default:     8080,
            EnvVar:      "TI_PORT",
            Category:    "server",
        },
        {
            Name:        "timeout",
            Description: "Request timeout in seconds",
            Type:        cli.FlagTypeInt,
            Default:     30,
            EnvVar:      "TI_TIMEOUT",
            Category:    "server",
        },
    }

    // Logging configuration flags
    loggingFlags := []*cli.Flag{
        {
            Name:        "log-level",
            Description: "Log level (debug, info, warn, error)",
            Type:        cli.FlagTypeString,
            Default:     "info",
            EnvVar:      "TI_LOG_LEVEL",
            Category:    "logging",
        },
        {
            Name:        "log-file",
            Description: "Log file path",
            Type:        cli.FlagTypeString,
            Default:     "",
            EnvVar:      "TI_LOG_FILE",
            Category:    "logging",
        },
    }

    // Register all flags
    for _, flag := range append(serverFlags, loggingFlags...) {
        if err := registry.Register(flag); err != nil {
            panic(err)
        }
    }

    return registry
}
```

### Example 3: Environment Variable Integration

```go
package main

import (
    "fmt"
    "os"

    "github.com/ti/cli/internal/cli"
)

func main() {
    registry := cli.NewFlagRegistry()

    flag := &cli.Flag{
        Name:        "api-key",
        Description: "API key for authentication",
        Type:        cli.FlagTypeString,
        Default:     "",
        EnvVar:      "TI_API_KEY",
        Required:    true,
    }

    registry.Register(flag)

    // Get value from environment variable
    value, err := registry.GetEnvValue(flag)
    if err != nil {
        panic(err)
    }

    // Set environment variable if not set
    if value == "" {
        fmt.Println("TI_API_KEY not set, please set it")
        os.Exit(1)
    }

    fmt.Printf("API Key: %s\n", value)
}
```

---

## Error Handling Examples

### Example 1: Basic Error Creation

```go
package main

import (
    "github.com/ti/cli/internal/cli"
)

func validateInput(input string) error {
    if input == "" {
        return cli.NewValidationError("input", "cannot be empty")
    }
    return nil
}
```

### Example 2: Error with Context

```go
package main

import (
    "fmt"

    "github.com/ti/cli/internal/cli"
)

func processFile(path string) error {
    // Check if file exists
    if _, err := os.Stat(path); os.IsNotExist(err) {
        return cli.NewNotFoundError(path).
            WithSuggestion("Check the file path or create the file")
    }

    // Check if file is readable
    if err := checkReadPermission(path); err != nil {
        return cli.NewPermissionError(fmt.Sprintf("read %s", path)).
            WithSuggestion("Check file permissions or run with elevated privileges")
    }

    return nil
}
```

### Example 3: Error Chaining

```go
package main

import (
    "fmt"
    "os"

    "github.com/ti/cli/internal/cli"
)

func loadConfig(path string) (*Config, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, cli.NewCLIError(cli.ErrCodeExecutionError, "failed to read config").
            Wrap(err).
            WithDetails(fmt.Sprintf("Config path: %s", path))
    }

    config, err := parseConfig(data)
    if err != nil {
        return nil, cli.NewCLIError(cli.ErrCodeConfigInvalid, "failed to parse config").
            Wrap(err)
    }

    return config, nil
}
```

### Example 4: Error Handling in Cobra Command

```go
package cmd

import (
    "github.com/spf13/cobra"
    "github.com/ti/cli/internal/cli"
)

var myCmd = &cobra.Command{
    Use:   "mycommand",
    Short: "My command",
    RunE: func(cmd *cobra.Command, args []string) error {
        // Try something that might fail
        if err := doSomething(); err != nil {
            return cli.NewCLIError(cli.ErrCodeExecutionError, "operation failed").
                Wrap(err).
                WithSuggestion("Check your inputs and try again")
        }
        return nil
    },
}
```

---

## Command Validation Examples

### Example 1: Simple Validation

```go
package main

import (
    "github.com/ti/cli/internal/cli"
)

func validateUser(name, email string) error {
    validator := cli.NewCommandValidator()

    // Name validation
    validator.AddRule("name", cli.ValidationRule{
        Name:      "required",
        Validator: cli.Required,
    })

    validator.AddRule("name", cli.ValidationRule{
        Name:      "min-length",
        Validator: cli.MinLength3,
    })

    // Email validation
    validator.AddRule("email", cli.ValidationRule{
        Name:      "required",
        Validator: cli.Required,
    })

    validator.AddRule("email", cli.ValidationRule{
        Name:      "email-format",
        Validator: cli.EmailPattern,
    })

    // Validate all fields
    fields := map[string]interface{}{
        "name":  name,
        "email": email,
    }

    return validator.ValidateAll(fields)
}
```

### Example 2: Complex Validation

```go
package main

import (
    "github.com/ti/cli/internal/cli"
)

func validateServerConfig(host string, port int, timeout int) error {
    validator := cli.NewCommandValidator()

    // Host validation
    validator.AddRule("host", cli.ValidationRule{
        Name:      "required",
        Validator: cli.Required,
    })

    validator.AddRule("host", cli.ValidationRule{
        Name:      "url-format",
        Validator: cli.URLPattern,
        Message:   "Host must be a valid URL",
    })

    // Port validation
    validator.AddRule("port", cli.ValidationRule{
        Name:      "required",
        Validator: cli.Required,
    })

    validator.AddRule("port", cli.ValidationRule{
        Name:      "port-range",
        Validator: cli.PortRange,
        Message:   "Port must be between 1 and 65535",
    })

    // Timeout validation
    validator.AddRule("timeout", cli.ValidationRule{
        Name:      "required",
        Validator: cli.Required,
    })

    validator.AddRule("timeout", cli.ValidationRule{
        Name:      "range",
        Validator: &cli.RangeValidator{Min: 1, Max: 300},
        Message:   "Timeout must be between 1 and 300 seconds",
    })

    fields := map[string]interface{}{
        "host":    host,
        "port":    port,
        "timeout": timeout,
    }

    return validator.ValidateAll(fields)
}
```

### Example 3: Custom Validator

```go
package main

import (
    "fmt"
    "regexp"

    "github.com/ti/cli/internal/cli"
)

// Custom validator for username
type UsernameValidator struct {
    MinLength int
    MaxLength int
}

func (v *UsernameValidator) Validate(value interface{}) error {
    username, ok := value.(string)
    if !ok {
        return cli.NewValidationError("username", "must be a string")
    }

    if len(username) < v.MinLength {
        return cli.NewValidationError("username", fmt.Sprintf("must be at least %d characters", v.MinLength))
    }

    if len(username) > v.MaxLength {
        return cli.NewValidationError("username", fmt.Sprintf("must be at most %d characters", v.MaxLength))
    }

    // Check for valid characters (alphanumeric, underscore, hyphen)
    matched, _ := regexp.MatchString(`^[a-zA-Z0-9_-]+$`, username)
    if !matched {
        return cli.NewValidationError("username", "must contain only alphanumeric characters, underscores, and hyphens")
    }

    return nil
}

func validateUsername(username string) error {
    validator := cli.NewCommandValidator()

    validator.AddRule("username", cli.ValidationRule{
        Name:      "username-format",
        Validator: &UsernameValidator{MinLength: 3, MaxLength: 20},
    })

    return validator.Validate("username", username)
}
```

---

## Tool System Examples

### Example 1: Using Built-in Tools

```go
package main

import (
    "context"
    "fmt"

    "github.com/ti/cli/internal/tool"
    "github.com/ti/cli/internal/tool/builtin"
)

func main() {
    registry := tool.NewRegistry()

    // Register built-in tools
    registry.Register(builtin.NewReadTool())
    registry.Register(builtin.NewWriteTool())
    registry.Register(builtin.NewGrepTool())

    // Get read tool
    readTool, exists := registry.Get("read")
    if !exists {
        panic("Read tool not found")
    }

    // Execute read tool
    result, err := readTool.Execute(context.Background(), map[string]interface{}{
        "file_path": "/path/to/file.txt",
    })

    if err != nil {
        panic(err)
    }

    if !result.Success {
        panic(result.Error)
    }

    fmt.Printf("File content: %v\n", result.Data)
}
```

### Example 2: Creating Custom Tool

```go
package custom

import (
    "context"
    "fmt"
    "time"

    "github.com/ti/cli/internal/tool"
)

type TimestampTool struct{}

func (t *TimestampTool) Name() string {
    return "timestamp"
}

func (t *TimestampTool) Description() string {
    return "Get current timestamp"
}

func (t *TimestampTool) Schema() map[string]interface{} {
    return map[string]interface{}{
        "type": "object",
        "properties": map[string]interface{}{
            "format": map[string]interface{}{
                "type":        "string",
                "description": "Timestamp format (rfc3339, unix, or custom)",
                "default":     "rfc3339",
            },
        },
    }
}

func (t *TimestampTool) Execute(ctx context.Context, params map[string]interface{}) (*tool.Result, error) {
    format, ok := params["format"].(string)
    if !ok {
        format = "rfc3339"
    }

    now := time.Now()
    var timestamp string

    switch format {
    case "rfc3339":
        timestamp = now.Format(time.RFC3339)
    case "unix":
        timestamp = fmt.Sprintf("%d", now.Unix())
    default:
        timestamp = now.Format(format)
    }

    return &tool.Result{
        Success: true,
        Data: map[string]interface{}{
            "timestamp": timestamp,
            "format":    format,
        },
        Metadata: map[string]interface{}{
            "execution_time": 0,
        },
    }, nil
}
```

### Example 3: Tool with Error Handling

```go
package custom

import (
    "context"
    "fmt"
    "os"

    "github.com/ti/cli/internal/tool"
    "github.com/ti/cli/internal/cli"
)

type FileSizeTool struct{}

func (t *FileSizeTool) Name() string {
    return "file_size"
}

func (t *FileSizeTool) Description() string {
    return "Get file size in bytes"
}

func (t *FileSizeTool) Schema() map[string]interface{} {
    return map[string]interface{}{
        "type": "object",
        "properties": map[string]interface{}{
            "file_path": map[string]interface{}{
                "type":        "string",
                "description": "Absolute path to the file",
            },
        },
        "required": []string{"file_path"},
    }
}

func (t *FileSizeTool) Execute(ctx context.Context, params map[string]interface{}) (*tool.Result, error) {
    filePath, ok := params["file_path"].(string)
    if !ok {
        return &tool.Result{
            Success: false,
            Error:   "file_path must be a string",
        }, nil
    }

    info, err := os.Stat(filePath)
    if err != nil {
        if os.IsNotExist(err) {
            return &tool.Result{
                Success: false,
                Error:   cli.NewNotFoundError(filePath).Error(),
            }, nil
        }
        return &tool.Result{
            Success: false,
            Error:   cli.NewCLIError(cli.ErrCodeExecutionError, "failed to get file info").Wrap(err).Error(),
        }, nil
    }

    return &tool.Result{
        Success: true,
        Data: map[string]interface{}{
            "file_path": filePath,
            "size":     info.Size(),
            "size_mb":  float64(info.Size()) / 1024 / 1024,
        },
    }, nil
}
```

---

## MCP OAuth Examples

### Example 1: Setting up OAuth for MCP Server

```go
package main

import (
    "context"
    "fmt"
    "os"

    "github.com/ti/cli/internal/mcp"
)

func main() {
    // Create OAuth configuration
    oauthConfig := &mcp.OAuthConfig{
        ClientID:     os.Getenv("GITHUB_CLIENT_ID"),
        ClientSecret: os.Getenv("GITHUB_CLIENT_SECRET"),
        AuthURL:      "https://github.com/login/oauth/authorize",
        TokenURL:     "https://github.com/login/oauth/access_token",
        Scopes:       []string{"repo", "user"},
        RedirectURI:  "http://localhost:8080/oauth/callback",
    }

    // Create MCP server configuration
    serverConfig := &mcp.ServerConfig{
        Name:      "github-mcp",
        Transport: mcp.TransportHTTP,
        URL:       "https://api.github.com",
        OAuth:     oauthConfig,
    }

    // Save configuration
    storage := mcp.NewStorage("mcp.db")
    if err := storage.SaveServerConfig(serverConfig); err != nil {
        panic(err)
    }

    fmt.Println("OAuth configuration saved")
}
```

### Example 2: Initiating OAuth Flow

```go
package main

import (
    "context"
    "fmt"

    "github.com/ti/cli/internal/mcp"
)

func main() {
    manager := mcp.NewManager()

    // Initiate OAuth flow
    authURL, stateToken, err := manager.InitiateOAuth("github-mcp")
    if err != nil {
        panic(err)
    }

    fmt.Printf("Visit this URL to authorize: %s\n", authURL)
    fmt.Printf("State token: %s\n", stateToken)

    // Wait for callback...
}
```

### Example 3: Retrieving OAuth Tokens

```go
package main

import (
    "context"
    "fmt"

    "github.com/ti/cli/internal/mcp"
)

func main() {
    manager := mcp.NewManager()

    // Get OAuth tokens
    tokens, err := manager.GetOAuthTokens("github-mcp")
    if err != nil {
        panic(err)
    }

    fmt.Printf("Access Token: %s\n", tokens.AccessToken)
    fmt.Printf("Refresh Token: %s\n", tokens.RefreshToken)
    fmt.Printf("Token Type: %s\n", tokens.TokenType)
    fmt.Printf("Expires In: %d seconds\n", tokens.ExpiresIn)
}
```

---

## Integration Patterns

### Example 1: Wrapping Existing Commands

```go
package cmd

import (
    "github.com/spf13/cobra"
    "github.com/ti/cli/internal/cli"
)

var oldCmd = &cobra.Command{
    Use:   "oldcommand",
    Short: "Old command",
    RunE: func(cmd *cobra.Command, args []string) error {
        // Existing logic
        return nil
    },
}

// Wrap with new architecture
func init() {
    wrapper := cli.NewCommandWrapper(oldCmd).
        WithErrorHandler(cli.DefaultErrorHandler).
        WithValidation(func(cmd *cobra.Command, args []string) error {
            // Add validation
            if len(args) == 0 {
                return cli.NewValidationError("args", "at least one argument required")
            }
            return nil
        })
}
```

### Example 2: Gradual Migration

```go
package cmd

import (
    "github.com/spf13/cobra"
    "github.com/ti/cli/internal/cli"
)

var myCmd = &cobra.Command{
    Use:   "mycommand",
    Short: "My command",
    RunE: func(cmd *cobra.Command, args []string) error {
        // Use new error handling
        if err := doSomething(); err != nil {
            return cli.NewCLIError(cli.ErrCodeExecutionError, "operation failed").Wrap(err)
        }

        // Keep existing validation for now
        if len(args) == 0 {
            return fmt.Errorf("argument required")
        }

        // Proceed with logic
        return nil
    },
}
```

### Example 3: Full Integration

```go
package cmd

import (
    "github.com/spf13/cobra"
    "github.com/ti/cli/internal/cli"
)

var (
    registry *cli.FlagRegistry
)

func init() {
    registry = cli.NewFlagRegistry()

    // Register flags
    registry.Register(&cli.Flag{
        Name:    "port",
        Type:    cli.FlagTypeInt,
        Default: 8080,
        EnvVar:  "TI_PORT",
    })
}

var myCmd = &cobra.Command{
    Use:   "mycommand",
    Short: "My command",
    RunE: func(cmd *cobra.Command, args []string) error {
        // Validate inputs
        validator := cli.NewCommandValidator()
        validator.AddRule("port", cli.ValidationRule{
            Name:      "port-range",
            Validator: cli.PortRange,
        })

        port, _ := cmd.Flags().GetInt("port")
        if err := validator.Validate("port", port); err != nil {
            return err
        }

        // Execute logic with error handling
        if err := doSomething(port); err != nil {
            return cli.NewCLIError(cli.ErrCodeExecutionError, "operation failed").
                Wrap(err).
                WithSuggestion("Check port availability")
        }

        return nil
    },
}
```

---

## Testing Examples

### Example 1: Testing Flag Registry

```go
package cli_test

import (
    "testing"

    "github.com/ti/cli/internal/cli"
)

func TestFlagRegistry(t *testing.T) {
    registry := cli.NewFlagRegistry()

    flag := &cli.Flag{
        Name:  "test",
        Type:  cli.FlagTypeString,
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

### Example 2: Testing Error Handling

```go
package cli_test

import (
    "testing"

    "github.com/ti/cli/internal/cli"
)

func TestCLIError(t *testing.T) {
    err := cli.NewCLIError(cli.ErrCodeValidation, "validation failed")

    if err.Code != cli.ErrCodeValidation {
        t.Errorf("Expected code %s, got %s", cli.ErrCodeValidation, err.Code)
    }

    if err.Message != "validation failed" {
        t.Errorf("Expected message 'validation failed', got '%s'", err.Message)
    }

    if err.ExitCode != 1 {
        t.Errorf("Expected exit code 1, got %d", err.ExitCode)
    }
}
```

### Example 3: Testing Validators

```go
package cli_test

import (
    "testing"

    "github.com/ti/cli/internal/cli"
)

func TestRequiredValidator(t *testing.T) {
    validator := &cli.RequiredValidator{}

    err := validator.Validate(nil)
    if err == nil {
        t.Error("Expected error for nil value")
    }

    err = validator.Validate("")
    if err == nil {
        t.Error("Expected error for empty string")
    }

    err = validator.Validate("test")
    if err != nil {
        t.Errorf("Unexpected error: %v", err)
    }
}
```

---

## Best Practices Summary

1. **Always validate inputs** before processing
2. **Use structured errors** with exit codes and suggestions
3. **Provide clear error messages** for users
4. **Use environment variables** for configuration
5. **Chain errors** to preserve context
6. **Write tests** for validators and error handlers
7. **Gradually migrate** existing commands
8. **Use common validators** when possible
9. **Document custom tools** with clear descriptions
10. **Handle OAuth tokens** securely

---

## See Also

- [Flag Management Documentation](CLI_FLAG_MANAGEMENT.md)
- [Error Handling Documentation](CLI_ERROR_HANDLING.md)
- [Command Validation Documentation](CLI_COMMAND_VALIDATION.md)
- [Tool System Documentation](CLI_TOOL_SYSTEM.md)
- [MCP OAuth Documentation](CLI_MCP_OAUTH.md)
- [CLI Guide](CLI_GUIDE.md)
