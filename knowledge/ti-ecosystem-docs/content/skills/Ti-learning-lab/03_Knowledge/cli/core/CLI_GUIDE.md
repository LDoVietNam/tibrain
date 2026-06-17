# CLI Guide - Official vs Legacy

## ⚠️ Critical: CLI Location Confusion

### Official CLI (Use This) ✅
- **Location**: `apps/cli/` in Z:\10_WORKPLACE\Ti monorepo
- **Build Output**: `bin/ti`
- **Status**: Active development
- **Features**:
  - Advanced converter with Go code generation
  - IR (Intermediate Representation) layer
  - Learning system with beads integration
  - Policy inference and risk analysis
  - Multi-format support (Claude Code, Codex, OpenCode, AmpCode, etc.)

### Legacy CLI (Deprecated) ❌
- **Location**: Previously at `Z:\Ti\CLI\`
- **Current Status**: Moved to `Z:\Ti\CLI.backup-*`
- **Features**:
  - Simple converter only (CLI config formats)
  - No Go code generation
  - No IR layer
  - No learning system

## Build Instructions

### Build Official CLI
```bash
cd apps/cli
go mod tidy
go build -o ../../bin/ti .
```

### Run Official CLI
```bash
cd /z/10_WORKPLACE/Ti
./bin/ti tui
```

## Advanced Converter Features

The official CLI includes a powerful converter with Go code generation:

### Commands
```bash
# Inspect source and analyze
ti convert inspect -i <source.json>

# Generate Ti IR (Intermediate Representation)
ti convert ir -i <source.json> -o ir.json

# Generate Go code from any CLI/IDE config
ti convert to-go -i <source.json> --out ./generated-plugin --with-tests

# View learning statistics
ti convert learn stats
```

### Pipeline
```
Source (Claude Code, Codex, etc.)
  ↓
Ti Pack (portable format)
  ↓
Ti IR (rich intermediate representation)
  ↓
Go Plugin/Project (generated code)
  ↓
Build/Test
  ↓
Memory/Beads Learning
```

### Supported Source Formats
- Claude Code
- Codex
- OpenCode
- Kilo
- AmpCode
- Cursor
- MCP configs
- Workflow definitions
- Skill definitions
- Agent definitions
- Ti configs

### Generated Go Project Structure
```
generated-plugin/
├── go.mod              # Go module definition
├── main.go             # Standalone plugin with embedded IR
├── ir.json             # Durable IR contract
├── plugin_generated_test.go  # Auto-generated tests
└── README.md           # Documentation with risk analysis
```

## Migration from Legacy CLI

If you were using `Z:\Ti\CLI\`:

1. **Stop using** `Z:\Ti\CLI\ti.exe`
2. **Build** official CLI from `apps/cli/`
3. **Update** any scripts/aliases to point to `bin/ti`
4. **Test** advanced converter features
5. **Delete** `Z:\Ti\CLI.backup-*` after verification

## Common Issues

### "command not found: ti"
**Solution**: Ensure `bin/ti` is built and in your PATH, or use full path `/z/10_WORKPLACE/Ti/bin/ti`

### "convert: command not found"
**Solution**: You're using legacy CLI. Rebuild from `apps/cli/`

### "no such file or directory: Z:\Ti\CLI"
**Solution**: Legacy CLI has been moved. Use `apps/cli/` instead

## Feature Comparison

| Feature | Official CLI | Legacy CLI |
|---------|-------------|------------|
| CLI Config Conversion | ✅ | ✅ |
| Go Code Generation | ✅ | ❌ |
| IR Layer | ✅ | ❌ |
| Learning System | ✅ | ❌ |
| Policy Inference | ✅ | ❌ |
| Risk Analysis | ✅ | ❌ |
| Beads Integration | ✅ | ❌ |
| Build/Test Integration | ✅ | ❌ |

## Transcode Tool (Code → Go)

Transcode is a standalone tool that converts source code from 9 languages to Go:
- Python, JavaScript, TypeScript, Rust, Java, C, C++, Ruby, PHP

### Status
- **Location**: `Z:\10_WORKPLACE\transcode\`
- **Integration**: Standalone tool (direct integration on hold due to build issues)
- **Documentation**: `docs/TRANSCODE_INTEGRATION.md`

### Usage (Standalone)
```bash
cd Z:\10_WORKPLACE\transcode
go build -o ../../bin/transcode ./cmd/transcode

# Convert single file
./bin/transcode main.py -o main.go

# Convert directory
./bin/transcode ./src --dir

# With LLM refinement
./bin/transcode app.js --llm-key $ANTHROPIC_KEY --llm-model claude-3-5-haiku-20241022
```

### Integration with Ti CLI
```bash
# 1. Transcode source files
transcode main.py -o main.go

# 2. Use Ti CLI for code review
ti review main.go

# 3. Optimize generated code
ti optimize main.go
```

### Future Integration
Direct integration into Ti CLI and MCP server is planned once build environment issues are resolved. See `docs/TRANSCODE_INTEGRATION.md` for details.

## TiBrain Integration (Alternative Approach)

Transcode has been integrated into TiBrain as an external tool:
- **Status**: Tool registered in TiBrain database ✅
- **Endpoint**: Available via TiBrain API
- **Documentation**: `docs/TRANSCODE_TIBRAIN_INTEGRATION.md`

### Benefits
- Centralized tool management in TiBrain
- Usage tracking and skill recommendations
- Integration with Ti ecosystem
- Quality metrics and learning

### Usage (When Build Issues Resolved)
```bash
# Register tool (already done)
bash scripts/register_transcode_tool.sh

# Use via wrapper
bash scripts/transcode_wrapper.sh main.py -o main.go

# Verify integration
curl http://localhost:1810/v1/brain/brains
```

## CLI Architecture (OpenCode-Dev Integration)

The Ti CLI has been enhanced with architecture patterns inspired by OpenCode-Dev, implemented in Go idioms. This provides a robust foundation for CLI development with better error handling, validation, and tool management.

### Architecture Components

#### 1. Flag Management System
**Location**: `internal/cli/flags.go`
**Documentation**: `docs/CLI_FLAG_MANAGEMENT.md`

Centralized, type-safe flag management with:
- Config-based flag definitions (replaces package-level variables)
- Environment variable support
- Type validation (string, int, bool, string array)
- Flag categories for grouping
- Thread-safe registry

```go
registry := cli.NewFlagRegistry()
registry.Register(&cli.Flag{
    Name:        "port",
    Type:        cli.FlagTypeInt,
    Default:     8080,
    EnvVar:      "TI_PORT",
    Category:    "server",
})
```

#### 2. Error Handling Patterns
**Location**: `internal/cli/errors.go`
**Documentation**: `docs/CLI_ERROR_HANDLING.md`

Structured error handling with:
- Machine-readable error codes
- User-friendly error messages
- Exit code support
- Error chaining
- Suggestions for error recovery

```go
err := cli.NewCLIError(cli.ErrCodeValidation, "validation failed").
    WithDetails("Missing required field").
    WithSuggestion("Add field to config")
```

#### 3. Command Validation
**Location**: `internal/cli/validator.go`
**Documentation**: `docs/CLI_COMMAND_VALIDATION.md`

Flexible validation system with:
- Reusable validator components
- Built-in validators (required, min/max length, pattern, range, enum)
- Common validators (email, URL, port range)
- Rule-based validation framework
- Batch validation support

```go
validator := cli.NewCommandValidator()
validator.AddRule("email", cli.ValidationRule{
    Name:      "email-format",
    Validator: cli.EmailPattern,
})
```

#### 4. Tool System
**Location**: `internal/tool/`
**Documentation**: `docs/CLI_TOOL_SYSTEM.md`

Pluggable tool architecture with:
- Tool interface for extensibility
- Tool registry for discovery
- JSON Schema for parameter validation
- Built-in tools (read, write, grep, exec, edit, multiedit)
- Integration with AI-driven execution

```go
registry := tool.NewRegistry()
registry.Register(builtin.NewReadTool())
registry.Register(builtin.NewWriteTool())
```

#### 5. MCP OAuth Integration
**Location**: `internal/mcp/oauth.go`
**Documentation**: `docs/CLI_MCP_OAUTH.md`

OAuth 2.0 support for MCP servers with:
- OAuth callback handler with HTTP server
- State management with CSRF protection
- Token storage in SQLite
- Support for any OAuth 2.0 provider
- Thread-safe operations

```go
handler := mcp.NewOAuthCallbackHandler(8080, logger)
handler.Start(ctx)
authURL, stateToken, _ := manager.InitiateOAuth("my-mcp-server")
```

### Architecture Benefits

1. **Centralized Management**: Flags, errors, and tools managed centrally
2. **Type Safety**: Automatic type checking and validation
3. **Extensibility**: Easy to add new tools and validators
4. **User Experience**: Better error messages and suggestions
5. **Security**: CSRF protection and secure token storage
6. **Thread-Safety**: Safe for concurrent operations
7. **Standardization**: Consistent patterns across commands

### Integration with Existing Commands

The new architecture is designed for gradual adoption:

```go
// Wrap existing commands
wrapper := cli.NewCommandWrapper(cmd).
    WithErrorHandler(cli.DefaultErrorHandler).
    WithValidation(func(cmd *cobra.Command, args []string) error {
        // Validate inputs
        return nil
    })
```

### Documentation

- **Flag Management**: `docs/CLI_FLAG_MANAGEMENT.md`
- **Error Handling**: `docs/CLI_ERROR_HANDLING.md`
- **Command Validation**: `docs/CLI_COMMAND_VALIDATION.md`
- **Tool System**: `docs/CLI_TOOL_SYSTEM.md`
- **MCP OAuth**: `docs/CLI_MCP_OAUTH.md`

### Implementation Status

| Component | Status | Tests |
|-----------|--------|-------|
| Flag Management | ✅ Complete | ✅ Passing |
| Error Handling | ✅ Complete | ✅ Passing |
| Command Validation | ✅ Complete | ✅ Passing |
| Tool System | ✅ Complete | ✅ Passing |
| MCP OAuth | ✅ Complete | ✅ Passing |

## Documentation

- **Architecture**: `Z:\07_DOCS\ti`
- **This Guide**: `docs/CLI_GUIDE.md`
- **Converter Implementation**: `apps/cli/internal/convert/`
  - `convert.go` - Basic Pack conversion
  - `ir.go` - Intermediate Representation
  - `go_codegen.go` - Go code generation
  - `learning.go` - Learning system
  - `codexv2.go` - Codex v2 parser
- **CLI Architecture**: `docs/CLI_FLAG_MANAGEMENT.md`, `docs/CLI_ERROR_HANDLING.md`, `docs/CLI_COMMAND_VALIDATION.md`, `docs/CLI_TOOL_SYSTEM.md`, `docs/CLI_MCP_OAUTH.md`

## Support

For issues or questions about the CLI:
1. Check this guide first
2. Review converter implementation in `apps/cli/internal/convert/`
3. Test with `ti convert inspect` to analyze your source
4. Check learning stats with `ti convert learn stats`
5. Review CLI architecture documentation for new patterns
