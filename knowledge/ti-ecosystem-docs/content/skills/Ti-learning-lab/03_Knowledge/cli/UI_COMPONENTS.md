# Ti CLI UI Components Documentation

> **Last Updated**: 2026-05-07  
> **Purpose**: Comprehensive documentation for Ti CLI TUI (Terminal User Interface) components  
> **Status**: Active

---

## Overview

Ti CLI provides several TUI components built with [Bubble Tea](https://github.com/charmbracelet/bubbletea) and [Lipgloss](https://github.com/charmbracelet/lipgloss). These components provide interactive terminal interfaces for various CLI operations.

---

## Architecture

### File Structure

```
internal/ui/
├── chat.go              # Chat TUI for AI conversations
├── dashboard.go         # Dashboard TUI for mission control
├── config_tui.go        # Config management TUI
├── memory_tui.go        # Memory management TUI
├── plugin_tui.go        # Plugin management TUI
├── provider_wizard.go   # Provider setup wizard
├── logo.go              # Logo rendering and utilities
└── constants.go         # UI constants and configuration
```

### Key Dependencies

- **Bubble Tea**: Elm-inspired TUI framework for Go
- **Lipgloss**: Style library for terminal UI
- **Bubbles**: Pre-built Bubble Tea components (textinput, etc.)

---

## Components

### 1. Chat TUI (`chat.go`)

**Purpose**: Interactive AI chat interface with multi-provider support.

**Key Features**:
- Multi-provider support (OpenAI, Anthropic, OpenRouter, etc.)
- Slash command system (`/help`, `/providers`, `/model`, etc.)
- Command suggestions and auto-completion
- Provider management (add, switch, list)
- Message history with scrolling
- Mouse support for click interactions

**Entry Points**:
- `StartChatTUI(reg *providers.Registry, pluginReg any)` - Full version with plugin registry
- `StartChatTUILegacy(reg *providers.Registry)` - Legacy version for backward compatibility

**Key Functions**:
- `initialChatModel()` - Initialize chat model with providers and commands
- `updateSuggestions()` - Update command suggestions based on input
- `handleSlashCommand()` - Process slash commands
- `sendMessage()` - Send messages to AI provider

**State Management**:
```go
type ChatModel struct {
    width, height          int
    messages              []ChatMessage
    input                 textinput.Model
    provider              providers.Provider
    registry              *providers.Registry
    pluginRegistry        any
    selectedIdx           int
    showSuggestions       bool
    suggestions           []string
    showProviderForm      bool
    providerFormFocus     int
    // ... more fields
}
```

**Keyboard Shortcuts**:
- `Tab` - Switch provider / show suggestions / switch focus
- `Ctrl+L` - Clear chat history
- `PgUp/PgDn` - Scroll messages
- `Esc` - Quit / cancel forms
- `Enter` - Send message / execute command

---

### 2. Dashboard TUI (`dashboard.go`)

**Purpose**: Mission control interface for Ti CLI operations.

**Key Features**:
- Action menu with categorized commands
- Filter functionality
- Command descriptions and safety levels
- Vietnamese localization
- Mouse support for selection

**Entry Points**:
- `StartDashboardTUI(registry any)` - Full version with plugin registry
- `StartDashboardTUILegacy()` - Legacy version with fallback items

**State Management**:
```go
type dashboardModel struct {
    items          []DashboardItem
    filteredItems  []DashboardItem
    selected       int
    filter         textinput.Model
    showFilter     bool
    filtered       bool
    width, height  int
    message        string
}
```

**Keyboard Shortcuts**:
- `↑/↓` - Navigate items
- `Enter` - Execute selected command
- `/` - Toggle filter
- `q` - Quit

---

### 3. Config TUI (`config_tui.go`)

**Purpose**: Interactive configuration management interface.

**Key Features**:
- View and edit configuration keys
- Search/filter functionality
- Edit mode for key-value pairs
- Validation and error handling
- Sample config support for testing

**Entry Points**:
- `StartConfigTUI()` - Launch config management TUI

**State Management**:
```go
type ConfigTUIModel struct {
    config         *kernel.Config
    keys           []string
    selected       int
    editMode       string // "key" or "value"
    editInput      textinput.Model
    scrollOffset   int
    width, height  int
    err            error
}
```

**Keyboard Shortcuts**:
- `↑/↓` - Navigate keys
- `Enter` - Edit selected key
- `Esc` - Cancel edit / quit
- `e` - Toggle edit mode

---

### 4. Memory TUI (`memory_tui.go`)

**Purpose**: Memory management interface for Ti's memory system.

**Key Features**:
- Browse memory entries
- Search functionality
- View memory details
- Filter by content

**Entry Points**:
- `StartMemoryTUI()` - Launch memory management TUI

**State Management**:
```go
type MemoryTUIModel struct {
    memories       []MemoryEntry
    filteredMemories []MemoryEntry
    selected       int
    searchInput    textinput.Model
    showSearch     bool
    scrollOffset   int
    width, height  int
    showingDetail  bool
}
```

**Keyboard Shortcuts**:
- `↑/↓` - Navigate memories
- `Enter` - View details
- `/` - Toggle search
- `Esc` - Close details / quit

---

### 5. Plugin TUI (`plugin_tui.go`)

**Purpose**: Plugin management interface.

**Key Features**:
- List available plugins
- View plugin details
- Enable/disable plugins

**Entry Points**:
- `StartPluginTUI()` - Launch plugin management TUI

**State Management**:
```go
type PluginTUIModel struct {
    plugins        []PluginEntry
    selected       int
    scrollOffset   int
    width, height  int
}
```

---

### 6. Provider Wizard (`provider_wizard.go`)

**Purpose**: Step-by-step wizard for adding new AI providers.

**Key Features**:
- Multi-step form (Name, Base URL, Model, API Key)
- Validation at each step
- Environment variable support
- No-key provider support

**Entry Points**:
- `StartProviderWizard(name string)` - Launch provider wizard

**State Management**:
```go
type ProviderWizardModel struct {
    step           int // 0=name, 1=baseurl, 2=model, 3=models, 4=apikey, 5=confirm
    name           textinput.Model
    baseURL        textinput.Model
    model          textinput.Model
    models         textinput.Model
    apiKeyEnv      textinput.Model
    apiKey         textinput.Model
    confirmed      bool
    width, height  int
}
```

**Keyboard Shortcuts**:
- `Tab` - Next field
- `Enter` - Confirm / next step
- `Esc` - Cancel

---

### 7. Logo & Utilities (`logo.go`)

**Purpose**: Logo rendering and common utility functions.

**Key Functions**:
- `RenderLogo(width int)` - Render Ti CLI logo
- `PrintLogo(width int)` - Print standalone logo
- `maxInt(a, b int)` - Maximum of two integers
- `minInt(a, b int)` - Minimum of two integers
- `clampInt(v, low, high int)` - Clamp value to range
- `truncateText(s, maxWidth int)` - Truncate text to fit width
- `wrapText(s, width int)` - Wrap text to fit width
- `badge(label, fg, bg)` - Render badge component

**Color Palette**:
```go
var (
    colorInk     = lipgloss.Color("230")  // Light yellow
    colorMuted   = lipgloss.Color("245")  // Gray
    colorSubtle  = lipgloss.Color("240")  // Dark gray
    colorPanel   = lipgloss.Color("236")  // Panel background
    colorPanelHi = lipgloss.Color("238")  // Panel highlight
    colorCyan    = lipgloss.Color("45")   // Cyan
    colorBlue    = lipgloss.Color("39")   // Blue
    colorViolet  = lipgloss.Color("99")   // Purple
    colorPink    = lipgloss.Color("205")  // Pink
    colorAmber   = lipgloss.Color("220")  // Amber
    colorGreen   = lipgloss.Color("42")   // Green
    colorRed     = lipgloss.Color("203")  // Red
)
```

---

### 8. Constants (`constants.go`)

**Purpose**: Centralized UI constants for maintainability.

**Categories**:
- Input field limits (char limits, widths)
- Layout dimensions (widths, margins, padding)
- UI element sizes (filter dimensions)
- Mouse click zones (offsets, heights)
- Form field indices (name, baseurl, apikey)
- Scroll behavior
- Color numbers (accent, title, borders, etc.)

**Usage Example**:
```go
// Instead of magic numbers
ti.CharLimit = 12000
ti.Width = 80

// Use constants
ti.CharLimit = maxInputCharLimit
ti.Width = defaultInputWidth
```

---

## Common Patterns

### TUI Model Structure

All TUI models follow the Bubble Tea pattern:

```go
type MyModel struct {
    // State fields
    width, height  int
    selected       int
    // ... more state
    
    // Input fields
    input          textinput.Model
    
    // Error handling
    err            error
}

// Init implements tea.Model
func (m MyModel) Init() tea.Cmd {
    return textinput.Blink
}

// Update implements tea.Model
func (m MyModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    // Handle messages
    switch msg := msg.(type) {
    case tea.KeyMsg:
        // Handle keyboard input
    case tea.MouseMsg:
        // Handle mouse input
    case tea.WindowSizeMsg:
        // Handle window resize
    }
    return m, nil
}

// View implements tea.Model
func (m MyModel) View() string {
    // Render UI
    return "UI content"
}
```

### Message Handling

Common message types:
- `tea.KeyMsg` - Keyboard input
- `tea.MouseMsg` - Mouse clicks/movement
- `tea.WindowSizeMsg` - Window resize
- Custom messages - Application-specific events

### Scroll Offset Pattern

Used in list-based TUIs (Config, Memory, Plugin):

```go
func (m *MyModel) adjustScrollOffset() {
    maxVisible := m.height - constantOffset
    if maxVisible < 5 {
        maxVisible = 5
    }
    
    if m.selected < m.scrollOffset {
        m.scrollOffset = m.selected
    } else if m.selected >= m.scrollOffset+maxVisible {
        m.scrollOffset = m.selected - maxVisible + 1
    }
}
```

---

## Integration Points

### Command Registration

TUI commands are registered in `cmd/` package:

```go
// Example from cmd/tui.go
var tuiCmd = &cobra.Command{
    Use:   "tui",
    Short: "Launch Ti mission control TUI",
    Run: func(cmd *cobra.Command, args []string) {
        // Launch Dashboard TUI
        if err := ui.StartDashboardTUILegacy(); err != nil {
            log.Fatal(err)
        }
    },
}
```

### Provider Integration

Chat TUI integrates with provider registry:

```go
registry := providers.NewRegistry()
// Register providers...
ui.StartChatTUI(registry, nil)
```

### Plugin Registry Integration

Some TUIs support plugin registry for dynamic commands:

```go
// TODO: Integrate with new plugin registry
// Currently uses surface registry for slash commands
```

---

## Testing

### Manual Testing Checklist

- [ ] Dashboard TUI launches and displays items
- [ ] Chat TUI launches with provider support
- [ ] Config TUI shows configuration keys
- [ ] Memory TUI displays memory entries
- [ ] Plugin TUI lists available plugins
- [ ] Provider wizard completes successfully
- [ ] Keyboard shortcuts work correctly
- [ ] Mouse interactions work
- [ ] Window resize handling works
- [ ] Error handling displays correctly

### Automated Testing

Currently, no automated tests exist for UI components. This is a future improvement area.

---

## Future Improvements

### Planned Enhancements

1. **Unit Tests**: Add comprehensive unit tests for UI components
2. **Accessibility**: Add screen reader support and ARIA labels
3. **i18n**: Add internationalization support for UI strings
4. **Performance**: Implement lazy loading and virtual scrolling
5. **Plugin Registry**: Complete integration with new plugin registry
6. **Theme Support**: Add color theme customization
7. **Keyboard Customization**: Allow custom keybindings
8. **State Persistence**: Save TUI state between sessions

### Technical Debt

- [x] Remove deprecated `tui.go` file (refactored to chat.go - 2026-05-07)
- [ ] Consolidate duplicate `adjustScrollOffset` functions
- [ ] Standardize error handling across TUIs
- [ ] Add comprehensive logging
- [ ] Improve type safety for custom messages

---

## Related Documentation

- [Ti CLI Architecture](./CLI_ARCHITECTURE.md)
- [Ti Learning Lab Workflow](../WORKFLOW.md)
- [Bubble Tea Documentation](https://github.com/charmbracelet/bubbletea)
- [Lipgloss Documentation](https://github.com/charmbracelet/lipgloss)

---

## Contributing

When modifying UI components:

1. **Follow patterns**: Use existing patterns for consistency
2. **Test manually**: Verify TUI functionality after changes
3. **Update docs**: Keep this documentation current
4. **Use constants**: Avoid magic numbers, use `constants.go`
5. **Handle errors**: Provide clear error messages
6. **Support keyboard**: Ensure keyboard-only navigation works
7. **Consider accessibility**: Think about screen reader users

---

*Last Updated: 2026-05-07*
