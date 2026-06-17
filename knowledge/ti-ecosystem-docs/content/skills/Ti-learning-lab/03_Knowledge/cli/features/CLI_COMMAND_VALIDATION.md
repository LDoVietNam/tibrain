# CLI Command Validation

> **Component**: `internal/cli/validator.go`
> **Status**: ✅ Implemented & Tested

## Overview

The Command Validation System provides a flexible, composable way to validate command inputs. It replaces manual validation logic with reusable validator components and a rule-based validation framework.

## Architecture

### Core Components

#### Validator Interface
```go
type Validator interface {
    Validate(value interface{}) error
}
```

Base interface for all validators.

#### ValidationRule
```go
type ValidationRule struct {
    Name      string     // Rule name for identification
    Validator Validator  // Validator instance
    Message   string     // Custom error message
}
```

Named validation rule with optional custom message.

#### CommandValidator
```go
type CommandValidator struct {
    rules map[string][]ValidationRule
}
```

Registry for managing validation rules per field.

## Built-in Validators

### RequiredValidator

Ensures a value is not empty (nil, empty string, empty slice/map).

```go
v := &RequiredValidator{}
v.Validate("")        // Error: value is required
v.Validate(nil)       // Error: value is required
v.Validate("test")    // Success
v.Validate([]int{})   // Error: value is required
```

### MinLengthValidator

Ensures a string has minimum length.

```go
v := &MinLengthValidator{Min: 3}
v.Validate("ab")      // Error: value must be at least 3 characters
v.Validate("abc")     // Success
```

### MaxLengthValidator

Ensures a string has maximum length.

```go
v := &MaxLengthValidator{Max: 10}
v.Validate("abcdefghijk")  // Error: value must be at most 10 characters
v.Validate("abc")          // Success
```

### PatternValidator

Ensures a string matches a regex pattern.

```go
v := &PatternValidator{Pattern: `^[a-zA-Z0-9]+$`}
v.Validate("test123")   // Success
v.Validate("test-123")  // Error: value must match pattern
```

### RangeValidator

Ensures a number is within range.

```go
v := &RangeValidator{Min: 1, Max: 10}
v.Validate(0)   // Error: value must be between 1 and 10
v.Validate(5)   // Success
v.Validate(11)  // Error: value must be between 1 and 10
```

### EnumValidator

Ensures a value is in allowed values.

```go
v := &EnumValidator{Allowed: []interface{}{"a", "b", "c"}}
v.Validate("a")  // Success
v.Validate("d")  // Error: value must be one of [a b c]
```

## Common Validators

Pre-configured validators for common use cases:

| Validator | Description | Pattern/Range |
|-----------|-------------|---------------|
| `Required` | Non-empty value | - |
| `NonEmpty` | Alias for Required | - |
| `MinLength3` | Minimum 3 characters | Min: 3 |
| `MinLength8` | Minimum 8 characters | Min: 8 |
| `MaxLength255` | Maximum 255 characters | Max: 255 |
| `MaxLength1024` | Maximum 1024 characters | Max: 1024 |
| `EmailPattern` | Email format | `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$` |
| `URLPattern` | URL format | `^https?://[^\s/$.?#].[^\s]*$` |
| `PositiveInt` | Positive integer | 0-2147483647 |
| `PortRange` | Valid port number | 1-65535 |

## Usage

### Basic Validation

```go
validator := NewCommandValidator()

validator.AddRule("name", ValidationRule{
    Name:      "required",
    Validator: Required,
    Message:   "Name is required",
})

err := validator.Validate("name", "")
if err != nil {
    // Handle error
}
```

### Multiple Rules

```go
validator.AddRule("email", ValidationRule{
    Name:      "required",
    Validator: Required,
})

validator.AddRule("email", ValidationRule{
    Name:      "email-format",
    Validator: EmailPattern,
    Message:   "Invalid email format",
})

err := validator.Validate("email", "invalid-email")
// Both rules are checked
```

### Batch Validation

```go
fields := map[string]interface{}{
    "name":  "John Doe",
    "email": "john@example.com",
    "age":   30,
}

err := validator.ValidateAll(fields)
if err != nil {
    // First validation error is returned
}
```

### Custom Error Messages

```go
validator.AddRule("password", ValidationRule{
    Name:      "min-length",
    Validator: MinLength8,
    Message:   "Password must be at least 8 characters long",
})

err := validator.Validate("password", "short")
// Error: Password must be at least 8 characters long
```

## Integration with Cobra

### Manual Validation

```go
var addCmd = &cobra.Command{
    Use: "add <name>",
    Args: cobra.ExactArgs(1),
    RunE: func(cmd *cobra.Command, args []string) error {
        validator := NewCommandValidator()

        validator.AddRule("name", ValidationRule{
            Name:      "required",
            Validator: Required,
        })

        validator.AddRule("name", ValidationRule{
            Name:      "min-length",
            Validator: MinLength3,
        })

        if err := validator.Validate("name", args[0]); err != nil {
            return err
        }

        // Proceed with command...
        return nil
    },
}
```

### With CommandWrapper

```go
wrapper := NewCommandWrapper(cmd).
    WithValidation(func(cmd *cobra.Command, args []string) error {
        validator := NewCommandValidator()

        validator.AddRule("name", ValidationRule{
            Name:      "required",
            Validator: Required,
        })

        return validator.Validate("name", args[0])
    })
```

## Examples

### Example 1: User Registration

```go
func validateUserRegistration(name, email, password string) error {
    validator := NewCommandValidator()

    // Name validation
    validator.AddRule("name", ValidationRule{
        Name:      "required",
        Validator: Required,
        Message:   "Name is required",
    })

    validator.AddRule("name", ValidationRule{
        Name:      "min-length",
        Validator: MinLength3,
        Message:   "Name must be at least 3 characters",
    })

    validator.AddRule("name", ValidationRule{
        Name:      "max-length",
        Validator: MaxLength255,
        Message:   "Name must be less than 255 characters",
    })

    // Email validation
    validator.AddRule("email", ValidationRule{
        Name:      "required",
        Validator: Required,
        Message:   "Email is required",
    })

    validator.AddRule("email", ValidationRule{
        Name:      "email-format",
        Validator: EmailPattern,
        Message:   "Invalid email format",
    })

    // Password validation
    validator.AddRule("password", ValidationRule{
        Name:      "required",
        Validator: Required,
        Message:   "Password is required",
    })

    validator.AddRule("password", ValidationRule{
        Name:      "min-length",
        Validator: MinLength8,
        Message:   "Password must be at least 8 characters",
    })

    // Validate all fields
    fields := map[string]interface{}{
        "name":     name,
        "email":    email,
        "password": password,
    }

    return validator.ValidateAll(fields)
}
```

### Example 2: Server Configuration

```go
func validateServerConfig(host string, port int, timeout int) error {
    validator := NewCommandValidator()

    // Host validation
    validator.AddRule("host", ValidationRule{
        Name:      "required",
        Validator: Required,
    })

    validator.AddRule("host", ValidationRule{
        Name:      "format",
        Validator: URLPattern,
        Message:   "Host must be a valid URL",
    })

    // Port validation
    validator.AddRule("port", ValidationRule{
        Name:      "required",
        Validator: Required,
    })

    validator.AddRule("port", ValidationRule{
        Name:      "range",
        Validator: PortRange,
        Message:   "Port must be between 1 and 65535",
    })

    // Timeout validation
    validator.AddRule("timeout", ValidationRule{
        Name:      "required",
        Validator: Required,
    })

    validator.AddRule("timeout", ValidationRule{
        Name:      "range",
        Validator: &RangeValidator{Min: 1, Max: 300},
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

### Example 3: MCP Server Configuration

```go
func validateMCPConfig(name, transport string, command []string) error {
    validator := NewCommandValidator()

    // Name validation
    validator.AddRule("name", ValidationRule{
        Name:      "required",
        Validator: Required,
    })

    validator.AddRule("name", ValidationRule{
        Name:      "pattern",
        Validator: &PatternValidator{Pattern: `^[a-zA-Z0-9_-]+$`},
        Message:   "Name must contain only alphanumeric characters, hyphens, and underscores",
    })

    // Transport validation
    validator.AddRule("transport", ValidationRule{
        Name:      "required",
        Validator: Required,
    })

    validator.AddRule("transport", ValidationRule{
        Name:      "enum",
        Validator: &EnumValidator{Allowed: []interface{}{"stdio", "http", "websocket"}},
        Message:   "Transport must be one of: stdio, http, websocket",
    })

    // Command validation (for stdio transport)
    if transport == "stdio" {
        validator.AddRule("command", ValidationRule{
            Name:      "required",
            Validator: Required,
        })

        validator.AddRule("command", ValidationRule{
            Name:      "min-length",
            Validator: MinLength3,
            Message:   "Command must have at least 3 characters",
        })
    }

    fields := map[string]interface{}{
        "name":     name,
        "transport": transport,
        "command":  command,
    }

    return validator.ValidateAll(fields)
}
```

### Example 4: Custom Validator

```go
// Custom validator for checking if a file exists
type FileExistsValidator struct {
    BaseDir string
}

func (v *FileExistsValidator) Validate(value interface{}) error {
    filename, ok := value.(string)
    if !ok {
        return NewValidationError("file", "must be a string")
    }

    path := filepath.Join(v.BaseDir, filename)
    if _, err := os.Stat(path); os.IsNotExist(err) {
        return NewValidationError("file", fmt.Sprintf("file does not exist: %s", path))
    }

    return nil
}

// Usage
validator.AddRule("config-file", ValidationRule{
    Name:      "file-exists",
    Validator: &FileExistsValidator{BaseDir: "/etc/app"},
    Message:   "Configuration file must exist",
})
```

## Best Practices

### 1. Use Descriptive Rule Names

```go
// Good
validator.AddRule("email", ValidationRule{
    Name: "email-format",
    Validator: EmailPattern,
})

// Avoid
validator.AddRule("email", ValidationRule{
    Name: "rule1",
    Validator: EmailPattern,
})
```

### 2. Provide Clear Error Messages

```go
// Good
validator.AddRule("password", ValidationRule{
    Name: "min-length",
    Validator: MinLength8,
    Message: "Password must be at least 8 characters long for security",
})

// Avoid
validator.AddRule("password", ValidationRule{
    Name: "min-length",
    Validator: MinLength8,
    // Default message may be unclear
})
```

### 3. Validate Early

```go
// Good - Validate at command entry point
RunE: func(cmd *cobra.Command, args []string) error {
    if err := validateInputs(args); err != nil {
        return err  // Fail fast
    }
    // Proceed with logic
}

// Avoid - Validate deep in logic
func doSomething() error {
    // ... lots of logic ...
    if err := validate(); err != nil {
        return err  // Too late
    }
}
```

### 4. Use Common Validators

```go
// Good
validator.AddRule("email", ValidationRule{
    Name: "email-format",
    Validator: EmailPattern,
})

// Avoid - Reimplement common validators
validator.AddRule("email", ValidationRule{
    Name: "email-format",
    Validator: &PatternValidator{Pattern: `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`},
})
```

### 5. Group Related Validations

```go
// Good - All email validations together
validator.AddRule("email", ValidationRule{Name: "required", Validator: Required})
validator.AddRule("email", ValidationRule{Name: "format", Validator: EmailPattern})
validator.AddRule("email", ValidationRule{Name: "max-length", Validator: MaxLength255})

// Avoid - Scattered validations
validator.AddRule("email", ValidationRule{Name: "required", Validator: Required})
// ... other field validations ...
validator.AddRule("email", ValidationRule{Name: "format", Validator: EmailPattern})
```

## Testing

```go
func TestCommandValidator(t *testing.T) {
    validator := NewCommandValidator()

    validator.AddRule("name", ValidationRule{
        Name:      "required",
        Validator: Required,
    })

    // Test validation failure
    err := validator.Validate("name", "")
    if err == nil {
        t.Error("Expected validation error")
    }

    // Test validation success
    err = validator.Validate("name", "test")
    if err != nil {
        t.Errorf("Unexpected error: %v", err)
    }

    // Test batch validation
    fields := map[string]interface{}{
        "name": "test",
    }

    err = validator.ValidateAll(fields)
    if err != nil {
        t.Errorf("Unexpected error: %v", err)
    }
}
```

## Migration from Manual Validation

### Before (Old Pattern)

```go
if name == "" {
    return fmt.Errorf("name is required")
}

if len(name) < 3 {
    return fmt.Errorf("name must be at least 3 characters")
}

if len(name) > 255 {
    return fmt.Errorf("name must be less than 255 characters")
}
```

### After (New Pattern)

```go
validator := NewCommandValidator()

validator.AddRule("name", ValidationRule{
    Name:      "required",
    Validator: Required,
})

validator.AddRule("name", ValidationRule{
    Name:      "min-length",
    Validator: MinLength3,
})

validator.AddRule("name", ValidationRule{
    Name:      "max-length",
    Validator: MaxLength255,
})

if err := validator.Validate("name", name); err != nil {
    return err
}
```

## Benefits

1. **Reusability**: Validators can be reused across commands
2. **Composability**: Multiple validators can be combined
3. **Type Safety**: Type checking for validator inputs
4. **Clear Messages**: Custom error messages for better UX
5. **Batch Validation**: Validate multiple fields at once
6. **Extensibility**: Easy to add custom validators
7. **Consistency**: Standardized validation patterns

## See Also

- [Flag Management](CLI_FLAG_MANAGEMENT.md)
- [Error Handling](CLI_ERROR_HANDLING.md)
- [CLI Guide](../CLI_GUIDE.md)
