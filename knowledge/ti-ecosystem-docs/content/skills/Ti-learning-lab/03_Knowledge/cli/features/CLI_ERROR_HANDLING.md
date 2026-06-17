# CLI Error Handling Patterns

> **Component**: `internal/cli/errors.go`
> **Status**: ✅ Implemented & Tested

## Overview

The Error Handling System provides structured, user-friendly error handling for CLI commands. It replaces simple string errors with typed errors that include exit codes, suggestions, and detailed context.

## Architecture

### Core Components

#### ErrorCode
```go
type ErrorCode string
```

Machine-readable error codes for programmatic error handling.

#### CLIError
```go
type CLIError struct {
    Code       ErrorCode  // Machine-readable error code
    Message    string     // Human-readable error message
    Details    string     // Additional error details
    Suggestion string     // User-friendly suggestion
    Wrapped    error      // Underlying error (for chaining)
    ExitCode   int        // Process exit code
}
```

## Error Codes

### General Errors

| Code | Description | Use Case |
|------|-------------|----------|
| `UNKNOWN` | Unknown error | Fallback for unclassified errors |
| `INVALID_INPUT` | Invalid input provided | User input validation failures |
| `NOT_FOUND` | Resource not found | Missing files, configs, etc. |
| `PERMISSION_DENIED` | Permission denied | File access, API auth failures |
| `VALIDATION_ERROR` | Validation failed | Data validation failures |
| `CONFLICT` | Conflict error | Resource conflicts, state mismatches |
| `TIMEOUT` | Operation timeout | Network timeouts, long operations |
| `RATE_LIMIT` | Rate limit exceeded | API rate limiting |

### CLI-Specific Errors

| Code | Description | Use Case |
|------|-------------|----------|
| `COMMAND_NOT_FOUND` | Command not found | Invalid command names |
| `FLAG_INVALID` | Invalid flag value | Flag validation failures |
| `CONFIG_INVALID` | Configuration invalid | Config file parsing errors |
| `DEPENDENCY_ERROR` | Dependency error | Missing dependencies |
| `EXECUTION_ERROR` | Execution failed | Command execution failures |

## Usage

### Basic Error Creation

```go
err := NewCLIError(ErrCodeValidation, "validation failed")
```

### Error with Details and Suggestions

```go
err := NewCLIError(ErrCodeConfigInvalid, "configuration error").
    WithDetails("Missing required field: api_key").
    WithSuggestion("Add api_key to your config file or set TI_API_KEY environment variable")
```

### Error Chaining

```go
originalErr := fmt.Errorf("file not found")
err := NewCLIError(ErrCodeNotFound, "config file missing").
    Wrap(originalErr)
```

### Error with Exit Code

```go
err := NewCLIError(ErrCodePermission, "access denied").
    WithExitCode(126)  // Custom exit code
```

## Error Constructors

### Command Not Found

```go
err := NewCommandNotFoundError("test-cmd")
// Output: "command 'test-cmd' not found (code: COMMAND_NOT_FOUND)"
// Suggestion: "Use --help to see available commands"
```

### Flag Invalid

```go
err := NewFlagInvalidError("--port", "must be a number")
// Output: "flag '--port' is invalid: must be a number (code: FLAG_INVALID)"
// Suggestion: "Use --help to see flag usage"
```

### Config Invalid

```go
err := NewConfigInvalidError("invalid YAML syntax")
// Output: "configuration error: invalid YAML syntax (code: CONFIG_INVALID)"
// Suggestion: "Check your configuration file or environment variables"
```

### Execution Error

```go
originalErr := fmt.Errorf("connection timeout")
err := NewExecutionError("connect to server", originalErr)
// Output: "execution failed: connect to server: connection timeout (code: EXECUTION_ERROR)"
```

### Validation Error

```go
err := NewValidationError("email", "invalid format")
// Output: "validation error: email (code: VALIDATION_ERROR)"
// Details: "invalid format"
```

### Not Found

```go
err := NewNotFoundError("config.yaml")
// Output: "resource not found: config.yaml (code: NOT_FOUND)"
```

### Permission Error

```go
err := NewPermissionError("delete file")
// Output: "permission denied: delete file (code: PERMISSION_DENIED)"
// Suggestion: "Check your permissions or use --force flag"
```

### Timeout Error

```go
err := NewTimeoutError("database connection")
// Output: "operation timed out: database connection (code: TIMEOUT)"
// Suggestion: "Try increasing timeout with --timeout flag"
```

## Integration with Cobra

### Manual Error Handling

```go
var myCmd = &cobra.Command{
    RunE: func(cmd *cobra.Command, args []string) error {
        if err := doSomething(); err != nil {
            return NewCLIError(ErrCodeExecutionError, "operation failed").Wrap(err)
        }
        return nil
    },
}
```

### With CommandWrapper

```go
wrapper := NewCommandWrapper(cmd).
    WithErrorHandler(func(err error) error {
        if err != nil {
            return DefaultErrorHandler(err)
        }
        return nil
    })
```

### Error Display

```go
cmd.RunE = func(cmd *cobra.Command, args []string) error {
    err := doSomething()
    if err != nil {
        exitCode := HandleCommandError(cmd, err)
        os.Exit(exitCode)
    }
    return nil
}
```

Output:
```
Error: configuration error
Details: Missing required field: api_key
Suggestion: Add api_key to your config file or set TI_API_KEY environment variable
```

## Best Practices

### 1. Use Specific Error Codes

```go
// Good
err := NewCLIError(ErrCodePermission, "access denied")

// Avoid
err := NewCLIError(ErrCodeUnknown, "something went wrong")
```

### 2. Provide Helpful Suggestions

```go
err := NewCLIError(ErrCodeConfigInvalid, "missing api_key").
    WithSuggestion("Set TI_API_KEY environment variable or add to config.yaml")
```

### 3. Chain Errors for Context

```go
originalErr := os.Open("config.yaml")
if originalErr != nil {
    return NewCLIError(ErrCodeNotFound, "config file missing").
        Wrap(originalErr)
}
```

### 4. Use Appropriate Exit Codes

```go
// Standard exit codes
err.WithExitCode(1)  // General error
err.WithExitCode(2)  // Misuse of shell command
err.WithExitCode(126) // Command invoked cannot execute
err.WithExitCode(127) // Command not found
err.WithExitCode(130) // SIGINT (Ctrl+C)
```

### 5. Include Details for Debugging

```go
err := NewCLIError(ErrCodeValidation, "validation failed").
    WithDetails(fmt.Sprintf("Field '%s' failed validation: %s", field, reason))
```

## Examples

### Example 1: File Operation Error

```go
func readFile(path string) error {
    data, err := os.ReadFile(path)
    if err != nil {
        if os.IsNotExist(err) {
            return NewNotFoundError(path).
                WithSuggestion("Check the file path or create the file")
        }
        if os.IsPermission(err) {
            return NewPermissionError(fmt.Sprintf("read %s", path)).
                WithSuggestion("Check file permissions or run with elevated privileges")
        }
        return NewCLIError(ErrCodeExecutionError, "failed to read file").Wrap(err)
    }
    // Process data...
    return nil
}
```

### Example 2: Configuration Validation

```go
func validateConfig(cfg *Config) error {
    if cfg.APIKey == "" {
        return NewValidationError("api_key", "required field is empty").
            WithSuggestion("Set TI_API_KEY environment variable or add to config.yaml")
    }

    if cfg.Port < 1 || cfg.Port > 65535 {
        return NewValidationError("port", "must be between 1 and 65535").
            WithDetails(fmt.Sprintf("Provided value: %d", cfg.Port))
    }

    if cfg.Timeout < 0 {
        return NewValidationError("timeout", "must be positive").
            WithDetails(fmt.Sprintf("Provided value: %d", cfg.Timeout))
    }

    return nil
}
```

### Example 3: API Call Error

```go
func callAPI(endpoint string) error {
    resp, err := http.Get(endpoint)
    if err != nil {
        if strings.Contains(err.Error(), "timeout") {
            return NewTimeoutError("API request").
                WithSuggestion("Increase timeout with --timeout flag or check network connectivity")
        }
        return NewCLIError(ErrCodeExecutionError, "API request failed").Wrap(err)
    }
    defer resp.Body.Close()

    if resp.StatusCode == 401 {
        return NewPermissionError("API authentication").
            WithSuggestion("Check your API credentials")
    }

    if resp.StatusCode == 429 {
        return NewCLIError(ErrCodeRateLimit, "API rate limit exceeded").
            WithSuggestion("Wait before retrying or upgrade your plan")
    }

    return nil
}
```

### Example 4: Command Validation

```go
var addCmd = &cobra.Command{
    Use:   "add <name>",
    Args:  cobra.ExactArgs(1),
    RunE: func(cmd *cobra.Command, args []string) error {
        name := args[0]

        if name == "" {
            return NewValidationError("name", "cannot be empty")
        }

        if len(name) < 3 {
            return NewValidationError("name", "must be at least 3 characters").
                WithDetails(fmt.Sprintf("Provided: %s", name))
        }

        // Proceed with command...
        return nil
    },
}
```

## Error Recovery

### Check Exit Code

```go
if cliErr, ok := err.(*CLIError); ok {
    if cliErr.IsExitCode(126) {
        // Handle permission error specifically
    }
}
```

### Get Exit Code

```go
exitCode := GetExitCode(err)
os.Exit(exitCode)
```

### Unwrap for Underlying Error

```go
if cliErr, ok := err.(*CLIError); ok {
    underlying := cliErr.Unwrap()
    // Handle underlying error
}
```

## Testing

```go
func TestCLIError(t *testing.T) {
    err := NewCLIError(ErrCodeValidation, "validation failed")

    if err.Code != ErrCodeValidation {
        t.Errorf("Expected code %s, got %s", ErrCodeValidation, err.Code)
    }

    if err.Message != "validation failed" {
        t.Errorf("Expected message 'validation failed', got '%s'", err.Message)
    }

    if err.ExitCode != 1 {
        t.Errorf("Expected exit code 1, got %d", err.ExitCode)
    }
}
```

## Migration from fmt.Errorf

### Before (Old Pattern)

```go
if err != nil {
    return fmt.Errorf("failed to read config: %w", err)
}
```

### After (New Pattern)

```go
if err != nil {
    return NewCLIError(ErrCodeExecutionError, "failed to read config").
        Wrap(err)
}
```

### Before (Old Pattern with Context)

```go
if port < 1 || port > 65535 {
    return fmt.Errorf("invalid port: %d (must be 1-65535)", port)
}
```

### After (New Pattern)

```go
if port < 1 || port > 65535 {
    return NewValidationError("port", "must be between 1 and 65535").
        WithDetails(fmt.Sprintf("Provided: %d", port))
}
```

## Benefits

1. **Structured Errors**: Consistent error format across all commands
2. **Exit Codes**: Proper process exit codes for scripting
3. **Suggestions**: User-friendly guidance for error recovery
4. **Details**: Additional context for debugging
5. **Chaining**: Error wrapping for context preservation
6. **Type Safety**: Machine-readable error codes
7. **Better UX**: Clear, actionable error messages

## See Also

- [Flag Management](CLI_FLAG_MANAGEMENT.md)
- [Command Validation](CLI_COMMAND_VALIDATION.md)
- [CLI Guide](../CLI_GUIDE.md)
