# Ti UI Architecture - Terminal UI & Component Generation

> **Created**: 2026-05-07
> **Status**: ✅ Active
> **Verified**: true
> **Last Used**: 2026-05-07
> **Related Projects**: [Ti Platform]

---

## 📚 Overview

Ti có **2 hệ thống UI chính**:
1. **Ti TUI** - Terminal UI riêng biệt với 9 modules để quản lý Ti ecosystem
2. **UIGenerator** - Component generator trong convert module để tạo UI code

### Location
- **Ti TUI**: `apps/ui/tui/`
- **UIGenerator**: `apps/core/cli/internal/convert/generators/ui_generator.go`

---

## 🎯 Ti CLI Internal UI Capabilities

### Overview

Ti CLI có **TUI nội bộ** được tích hợp trực tiếp vào CLI binary, cung cấp các interactive terminal interfaces cho các tasks cụ thể. Đây là khác biệt với Ti TUI (tui.exe) là ứng dụng riêng biệt.

### Location

`apps/core/cli/internal/ui/`

### Framework

- **Bubble Tea**: `github.com/charmbracelet/bubbletea` - Elm architecture cho terminal UI
- **Lipgloss**: `github.com/charmbracelet/lipgloss` - Styling cho terminal UI
- **Bubbles**: `github.com/charmbracelet/bubbles` - Reusable UI components (textinput, textarea, etc.)

### CLI UI Modules

#### 1. Dashboard TUI (`dashboard.go`)

**Purpose**: Mission control menu hiển thị các Ti commands

**Features**:
- Interactive menu với keyboard navigation
- Filter commands bằng `/`
- Mouse support cho selection
- Grouped commands (Core, Engines, Tools)
- Safety indicators cho mỗi command
- Hiển thị command khi chọn (không execute trực tiếp - safety first)

**Keyboard Shortcuts**:
- `↑`/`↓` hoặc `j`/`k` - Navigate
- `/` - Toggle filter
- `Enter` - Show command
- `Esc` - Quit/close filter
- `q` - Quit
- `?` - Show help

**Command**: `ti tui` hoặc `ti dashboard` hoặc `ti ui` hoặc `ti menu`

**Implementation**: `apps/core/cli/cmd/tui.go`, `apps/core/cli/internal/ui/dashboard.go`

**Visual**:
```
✦ Ti Mission Control
Không cần nhớ command: chọn việc cần làm, Ti sẽ gợi ý lệnh đúng. Enter = xem lệnh, q = thoát

  Review project              ti review --run              read-only
  Fix build/test              ti fix --run                 asks before write
  Plan work                   ti plan "..." --run           read-only
  Optimize context            ti optimize context --run     read-only
  ...
```

#### 2. Chat TUI (`chat.go`)

**Purpose**: Interactive chat interface với configured provider/router

**Features**:
- Real-time chat với LLM provider
- Provider management
- Slash commands:
  - `/help` - Show help
  - `/providers` - List providers
  - `/provider <name>` - Show provider details
  - `/model <model>` - Switch model
  - `/add-provider` - Add new provider (opens wizard)
  - `/clear` - Clear chat history
  - `/exit` - Exit chat

**Command**: `ti ask`

**Implementation**: `apps/core/cli/internal/ui/chat.go` (refactored from tui.go)

**Key Components**:
```go
type ChatModel struct {
    width               int
    height              int
    messages            []ChatMessage
    input               textinput.Model
    provider            providers.Provider
    registry            *providers.Registry
    pluginRegistry      any
    selectedIdx         int
    providers           []string
    showSuggestions     bool
    suggestions         []string
    showProviderForm    bool
    providerFormFocus   int
    // ...
}
```

#### 3. Provider Wizard (`provider_wizard.go`)

**Purpose**: Multi-step form wizard để add new provider

**Features**:
- 5-step wizard: Base URL → Model → Models → Auth → Headers
- Mouse support
- Tab navigation giữa fields
- Auth choice: Environment variable / Direct API key / No auth
- Custom headers support
- Summary review trước khi save

**Keyboard Shortcuts**:
- `Enter` - Next step / Save
- `Tab` - Switch fields / options
- `Esc` / `q` - Cancel
- `Y`/`N` - Confirm on summary

**Implementation**: `apps/core/cli/internal/ui/provider_wizard.go` (429 lines)

**Steps**:
```
Step 1/5: Base URL
https://api.openai.com/v1

Step 2/5: Default Model
gpt-4

Step 3/5: Additional Models
gpt-3.5-turbo,gpt-4

Step 4/5: Authentication
[1] Environment variable (recommended)
[2] Direct API key
[3] No authentication

Step 5/5: Custom Headers (optional)
Name:  Authorization
Value: Bearer xxx
```

#### 4. Config TUI (`config_tui.go`)

**Purpose**: Interactive config management

**Features**:
- Browse config keys
- Edit config values
- Reload config
- Mouse support
- Scrollable list
- Detail view cho selected key

**Keyboard Shortcuts**:
- `↑`/`↓` hoặc `j`/`k` - Navigate
- `Enter` hoặc `e` - Edit
- `r` - Reload config
- `Esc` / `q` - Quit

**Implementation**: `apps/core/cli/internal/ui/config_tui.go` (355 lines)

#### 5. Memory TUI (`memory_tui.go`)

**Purpose**: Browse & search project memory

**Implementation**: `apps/core/cli/internal/ui/memory_tui.go`

#### 6. Plugin TUI (`plugin_tui.go`)

**Purpose**: Plugin management interface

**Implementation**: `apps/core/cli/internal/ui/plugin_tui.go`

#### 7. Logo & Utilities (`logo.go`)

**Purpose**: Logo rendering và common utility functions

**Features**:
- Logo rendering với automatic compact mode
- Utility functions: maxInt, minInt, clampInt, truncateText, wrapText
- Badge rendering
- Color palette constants

**Implementation**: `apps/core/cli/internal/ui/logo.go`

#### 8. Constants (`constants.go`)

**Purpose**: Centralized UI constants cho maintainability

**Features**:
- Input field limits (char limits, widths)
- Layout dimensions (widths, margins, padding)
- UI element sizes (filter dimensions)
- Mouse click zones (offsets, heights)
- Form field indices (name, baseurl, apikey)
- Scroll behavior
- Color numbers

**Implementation**: `apps/core/cli/internal/ui/constants.go` (NEW - created 2026-05-07)

### UI Patterns Used

#### Pattern 1: Menu Selection (Dashboard)

```go
type dashboardModel struct {
    items         []DashboardItem
    selected      int
    filter        textinput.Model
    showFilter    bool
    // ...
}

func (m dashboardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch msg.String() {
        case "up", "k":
            if m.selected > 0 {
                m.selected--
            }
        case "enter":
            item := items[m.selected]
            m.message = fmt.Sprintf("Run: %s", item.Command)
        }
    }
    return m, cmd
}
```

#### Pattern 2: Multi-step Wizard (Provider Wizard)

```go
type ProviderWizardModel struct {
    step         int
    baseURL      textinput.Model
    model        textinput.Model
    apiKeyEnv    textinput.Model
    apiKey       textinput.Model
    authChoice   int
    // ...
}

func (m ProviderWizardModel) handleEnter() (tea.Model, tea.Cmd) {
    switch m.step {
    case 1:
        m.step = 2
        m.model.Focus()
        m.baseURL.Blur()
    case 2:
        m.step = 3
        // ...
    }
    return m, nil
}
```

#### Pattern 3: Form with Input Fields (Config TUI)

```go
type ConfigTUIModel struct {
    config       *kernel.Config
    configKeys   []string
    selected     int
    showEdit     bool
    editValue    textinput.Model
    // ...
}

func (m ConfigTUIModel) View() string {
    if m.showEdit {
        return m.editView()
    }
    // Show list view
}
```

### Bubble Tea Components Used

- **textinput**: Text input fields
- **textarea**: Multi-line text input
- **viewport**: Scrollable content areas
- **list**: List selection
- **spinner**: Loading indicators

### Styling với Lipgloss

```go
accent := lipgloss.Color("63")
muted := lipgloss.Color("245")

title := lipgloss.NewStyle().
    Bold(true).
    Foreground(lipgloss.Color("230")).
    Background(accent).
    Padding(0, 2).
    Width(w).
    Render("✦ Ti Mission Control")
```

### Safety Design

**Dashboard TUI không execute commands trực tiếp**:
- Chỉ hiển thị command string
- User phải copy/paste command vào shell
- Prevent accidental destructive actions
- "Ti keeps this TUI safe: it shows the command instead of executing"

---

## 🎯 Ti TUI - Terminal UI (Separate Application)

### Architecture

```
Ti Platform
├── Ti CLI (Interface 1) - Command-line interface
├── Ti TUI (Interface 2) - Terminal UI interface ← NEW Component
├── Ti Router (Service) - Model routing service (port 1807)
├── MCP Hub (Service) - MCP servers (24+ tools)
├── Ti Brain (Component) - AI reasoning
├── Automation (Component) - Automation workflows
├── Taskboard (Component) - Task management
└── Core (Shared Library) - Shared contracts (JSON schemas)
```

### Structure

```
apps/ui/tui/
├── cmd/tui/                    # TUI entry point
│   └── main.go
├── internal/
│   ├── ui/                     # Bubble Tea TUI
│   │   ├── launcher.go         # Launcher với 9 modules
│   │   ├── cli_tui.go          # CLI tab
│   │   ├── agent_hub.go        # Agent Hub tab
│   │   ├── memory_tui.go       # Memory tab
│   │   ├── skills_tui.go       # Skills tab
│   │   ├── brain_tui.go        # TiBrain Hub tab
│   │   ├── router_tui.go       # Router tab
│   │   ├── mcp_tui.go          # MCP tab
│   │   ├── automation_tui.go   # Automation tab
│   │   ├── taskboard_tui.go    # Taskboard tab
│   │   └── dashboard_tui.go    # 4-panel split dashboard
│   ├── integration/            # Ti platform integration
│   │   ├── tibrain/            # TiBrain integration (qua MCP)
│   │   ├── router/             # Router integration (qua HTTP API)
│   │   ├── mcp/                # MCP Hub integration (qua MCP client)
│   │   ├── automation/         # Automation integration
│   │   ├── taskboard/          # Taskboard integration
│   │   └── manager.go          # Integration manager
│   ├── agent/                  # Multi-agent orchestration
│   ├── persistence/            # SQLite persistence
│   ├── skills/                 # Skills integration
│   ├── tools/                  # Tools integration
│   ├── circuit/                # Circuit breakers
│   ├── api/                    # REST API
│   ├── cli/                    # CLI integration
│   ├── config/                 # Config management
│   └── windows/                # Windows executor
├── go.mod                      # Module: github.com/ti/agent-tui
└── bin/ti-tui.exe              # TUI binary
```

### Framework

- **Bubble Tea**: `github.com/charmbracelet/bubbletea` - Elm architecture for terminal UI
- **Lipgloss**: `github.com/charmbracelet/lipgloss` - Styling for terminal UI

---

## 🖥️ TUI Modules

### 1. Launcher

**Features**:
- Matrix-style hacker aesthetic với boot sequence
- 9 module icons với navigation
- Boot sequence animation
- Status indicators

**Keyboard Shortcuts**:
- `↑`/`↓` - Navigate modules
- `1`-`9` - Select module
- `Enter` - Launch selected module
- `Ctrl+C` / `Q` - Quit

**Visual**:
```
╔══════════════════════════════════════════════════════════════╗
║  ████████╗██╗     ╔══════════════════════════════════════════╗
║     ██╔══╝██║     ║  ░░  A G E N T   C O N T R O L  ░░  ║
║     ██║   ██║     ║    ─────────────────────────────  ║
║     ██║   ██║     ║    T E R M I N A L   S Y S T E M  ║
║                     ║                                  ║
║     ██║   ██║     ║    v1.0.0  ·  Neural Interface   ║
║     ██║   ██║     ║                                  ║
║     ╚═╝   ╚══════╝  ║  ● ONLINE    ⬡ 8 MODULES        ║
╚═══════════════════════╣══════════════════════════════════╣
║ ⊞  Dashboard            ║>_ CLI                    ║
║ ⬡  Agent Hub            ║◉ Memory                 ║
║ ⚡  Skills               ║🧠 Brain                  ║
║ 🌐  Router              ║⚙️ Automation             ║
║ 📋  Taskboard           ║                          ║
╚═════════════════════════╩══════════════════════════════════╝
```

### 2. CLI Module

**Purpose**: Direct terminal chat với Ti CLI

**Features**:
- Single-pane chat interface
- User messages on the right, Ti responses on the left
- History navigation with ↑/↓ arrows
- Streaming output with animated cursor
- No agent management overhead (lean REPL)

**Implementation**: `internal/ui/cli_tui.go`

### 3. Agent Hub Module

**Purpose**: Multi-agent management dashboard

**Features**:
- Agent orchestration
- Cost tracking
- Circuit breakers
- Agent status dashboard
- Multi-agent communication

**Tabs**:
- CHAT - Agent chat interface
- TASKS - Task management
- AGENTS - Agent list
- COST - Cost tracking

**Implementation**: `internal/ui/agents/`

### 4. Memory Module

**Purpose**: Browse & search agent memory bank

**Features**:
- Memory visualization
- Memory management
- Vector search
- Entry browsing

**Tabs**:
- SEARCH - Search memory
- ENTRIES - Browse entries
- EMBED - Embed management

**Implementation**: `internal/ui/memory_tui.go`

### 5. Skills Module

**Purpose**: Manage & invoke agent skill library

**Features**:
- Skill discovery
- Skill execution
- Skill library management

**Implementation**: `internal/ui/skills_tui.go`

### 6. Brain Module

**Purpose**: TiBrain Hub management

**Features**:
- View TiBrain Hub status
- List registered CLIs
- View handoffs between CLIs
- Search skills in TiBrain registry

**Tabs**:
- STATUS - TiBrain Hub status
- CLIs - Registered CLIs
- HANDOFFS - Handoffs between CLIs
- SKILLS - Search skills

**Implementation**: `internal/ui/brain_tui.go`

### 7. Router Module

**Purpose**: Ti Router management

**Features**:
- View Router health status
- List available models
- Chat completion interface

**Tabs**:
- HEALTH - Router health status
- MODELS - Available models
- CHAT - Chat completion interface

**Integration**: HTTP API (http://localhost:1807)

**Implementation**: `internal/ui/router_tui.go`

### 8. MCP Module

**Purpose**: MCP Servers management

**Features**:
- List MCP servers
- View tools per server
- Execute MCP tools

**Tabs**:
- SERVERS - MCP servers list
- TOOLS - Tools per server
- EXECUTE - Tool execution

**Integration**: MCP Client (stdio + HTTP transport)

**Implementation**: `internal/ui/mcp_tui.go`

### 9. Automation Module

**Purpose**: Ti Automation workflows

**Features**:
- List automation workflows
- Execute workflows
- View execution logs

**Tabs**:
- WORKFLOWS - Workflow list
- EXECUTE - Workflow execution
- LOGS - Execution logs

**Integration**: File system integration

**Implementation**: `internal/ui/automation_tui.go`

### 10. Taskboard Module

**Purpose**: Taskboard/BEADS management

**Features**:
- View all beads
- Filter by agent/status
- View statistics

**Tabs**:
- BEADS - All beads
- FILTER - Filter beads
- STATS - Statistics

**Integration**: beads.md parsing

**Implementation**: `internal/ui/taskboard_tui.go`

### 11. Dashboard Module

**Purpose**: 4-panel split dashboard showing CLI, Hub, Memory, Brain simultaneously

**Layout**:
```
╔═══════════════════════════╦═══════════════════════════╗
║● ● ●  ti — terminal     ║● ● ●  agent-hub           ║
║● LIVE  claude  —/128k    ║● 2 active  3 agents       ║
║● 1 CHAT  2 TASKS  …     ║● 1 CHAT  2 TASKS  …     ║
║ messages...              ║ chat content...           ║
╠═══════════════════════════╬═══════════════════════════╣
║● ● ●  memory            ║● ● ●  tibrain             ║
║● 6 entries  1536-dim    ║● ONLINE  2CLIs 5hoff     ║
║● 1 SEARCH  2 ENTRIES    ║● 1 STATUS 2 CLIs …      ║
║ vector results...         ║ stat cards...             ║
╚═══════════════════════════╩═══════════════════════════╝
```

**Features**:
- Traffic-light dots per panel
- Subtitle stats
- Tab bar per panel
- Content scrolling
- Panel focus with Tab/Shift+Tab

**Keyboard Shortcuts**:
- `Tab`/`Shift+Tab` - Cycle panel focus
- `1`-`4` - Jump to panel
- `←`/`→` - Switch tab within panel
- `↑`/`↓` - Scroll content
- `Ctrl+\` - Back to launcher

**Implementation**: `internal/ui/dashboard_tui.go`

---

## 🔌 Platform Integrations

### TiBrain Integration (qua MCP)

**Location**: `internal/integration/tibrain/tibrain.go`

**Methods**:
- GetStatus - View TiBrain Hub status
- ListCLIs - List registered CLIs
- ListHandoffs - View handoffs between CLIs
- RecallHandoff - Recall a handoff
- SearchSkills - Search skills in TiBrain registry
- ExecuteTool - Execute TiBrain tool
- LogUsage - Log usage metrics

**Status**: ✅ Interface defined (MCP client cần implement)

### Router Integration (qua HTTP API)

**Location**: `internal/integration/router/router.go`

**Methods**:
- GetHealth - Get Router health status
- ListModels - List available models
- ChatCompletion - Chat completion interface

**Status**: ✅ Fully implemented với HTTP client

**Endpoint**: `http://localhost:1807`

### MCP Hub Integration (qua MCP Client)

**Location**: `internal/integration/mcp/mcp.go`

**Methods**:
- ListServers - List MCP servers
- ListTools - List tools per server
- ExecuteTool - Execute MCP tool

**Status**: ✅ Interface defined (MCP client cần implement)

**Transport**: stdio + HTTP

### Automation Integration

**Location**: `internal/integration/automation/automation.go`

**Methods**:
- ListWorkflows - List automation workflows
- ExecuteWorkflow - Execute workflow

**Status**: ✅ Implemented với file system integration

### Taskboard Integration

**Location**: `internal/integration/taskboard/taskboard.go`

**Methods**:
- ReadBeads - Read beads from beads.md
- FilterBeadsByStatus - Filter by status
- FilterBeadsByAgent - Filter by agent

**Status**: ✅ Implemented với beads.md parsing

---

## 🎨 Ti UI Generation Capabilities

### Overview

Ti có **2 cách để tạo UI code**:

1. **UIGenerator (Convert Module)** - Generator cơ bản tạo Go code cho UI components
2. **Frontend-Design Skill** - AI agent skill để tạo production-grade frontend interfaces

### 1. UIGenerator - Basic Component Generation

#### Location

`apps/core/cli/internal/convert/generators/ui_generator.go`

#### Purpose

Tạo Go code cho UI components từ IR (Intermediate Representation)

#### Features

- Generate Go code với UI component library
- Template-based generation
- Support component types:
  - Button
  - Input
  - Card
  - Modal
  - List
  - Generic components

#### Component Types

| Type | Description | Props |
|------|-------------|-------|
| **Button** | Clickable button | className, id, variant, children |
| **Input** | Text input field | className, id, type, placeholder |
| **Card** | Card container | className, id, children |
| **Modal** | Modal dialog | className, id, children |
| **List** | List component | className, id, items |
| **Generic** | Generic component | className, id, children |

#### Implementation

```go
type UIGenerator struct {
    templates *template.Template
}

func (g *UIGenerator) Generate(ir *convert.IR) (*GeneratedCode, error) {
    // Extract UI components from IR tools
    components := extractUIComponents(ir)

    // Generate Go code with component library
    // ...
}
```

#### Generated Code Structure

```go
package claude

import (
    "fmt"
    "html/template"
    "strings"
)

// UIComponentLibrary manages generated components
type UIComponentLibrary struct {
    components map[string]*UIComponent
    templates  *template.Template
}

// GenerateButton creates a button component
func (lib *UIComponentLibrary) GenerateButton(props ButtonProps) (*UIComponent, error) {
    // Generate HTML based on props
    html := fmt.Sprintf(`<button class="%s" id="%s">%s</button>`,
        className, id, children)

    comp := &UIComponent{
        Name:  "button",
        Type:  "button",
        HTML:  template.HTML(html),
    }

    return comp, nil
}
```

#### Usage

```bash
# Convert UI description to Go components
ti convert --from ui --to go \
  -i ui_description.json \
  --out ./Components/ui-library \
  --kind ui-component
```

#### Limitations

- Chỉ tạo Go code với HTML templates
- Không support modern frontend frameworks (React, Vue, Angular)
- Component types cơ bản, không có advanced features
- Không có styling system (CSS, Tailwind, etc.)
- Template-based, không có AI-powered generation

---

### 2. Frontend-Design Skill - AI-Powered UI Generation

#### Location

`.windsurf/skills/frontend-design/SKILL.md`

#### Purpose

Tạo production-grade frontend interfaces với AI agent (Claude/Devin)

#### Features

- Tạo distinctive, production-grade frontend interfaces
- Tránh generic "AI slop" aesthetics
- Support multiple frameworks:
  - HTML/CSS/JS
  - React
  - Vue
  - Angular
  - Any frontend framework

#### Design Philosophy

**Design Thinking**:
- **Purpose**: Problem solving, target audience
- **Tone**: Bold aesthetic direction (minimalist, maximalist, retro-futuristic, organic, luxury, playful, editorial, brutalist, art deco, etc.)
- **Constraints**: Technical requirements (framework, performance, accessibility)
- **Differentiation**: What makes it UNFORGETTABLE?

**Frontend Aesthetics Guidelines**:
- **Typography**: Distinctive fonts, avoid generic (Arial, Inter), pair display + body fonts
- **Color & Theme**: Cohesive aesthetic, CSS variables, dominant colors with sharp accents
- **Motion**: CSS-only animations, Motion library for React, staggered reveals, scroll-triggering
- **Spatial Composition**: Unexpected layouts, asymmetry, overlap, diagonal flow, grid-breaking
- **Backgrounds & Visual Details**: Gradient meshes, noise textures, geometric patterns, layered transparencies

#### Usage

```bash
# Invoke frontend-design skill
ti skill invoke frontend-design

# Hoặc khi user yêu cầu tạo UI
"Create a landing page for my product"
"Build a dashboard with charts"
"Design a login form component"
```

#### Example Output

AI agent sẽ tạo:
- Production-grade HTML/CSS/JS hoặc React/Vue code
- Custom typography và color scheme
- Animations và micro-interactions
- Responsive design
- Accessibility considerations
- Unique aesthetic (không generic)

#### Advantages over UIGenerator

- ✅ AI-powered generation (không template-based)
- ✅ Support modern frontend frameworks (React, Vue, Angular)
- ✅ Production-grade code với best practices
- ✅ Unique aesthetics (tránh generic AI look)
- ✅ Full styling system (CSS, Tailwind, custom)
- ✅ Animations và interactions
- ✅ Responsive và accessible

#### Limitations

- ❌ Phụ thuộc vào AI model quality
- ❌ Có thể inconsistent giữa các lần generate
- ❌ Cần review kỹ code trước khi production
- ❌ Không có deterministic output

---

## 🎨 UIGenerator - Component Generation (Legacy)

### Purpose

Tạo Go code cho UI components từ IR (Intermediate Representation)

### Location

`apps/core/cli/internal/convert/generators/ui_generator.go`

### Features

- Generate React/Go UI components từ descriptions
- Template-based generation
- Support multiple component types:
  - Button
  - Input
  - Card
  - Modal
  - List
  - Generic components

### Component Types

| Type | Description | Props |
|------|-------------|-------|
| **Button** | Clickable button | className, id, variant, children |
| **Input** | Text input field | className, id, type, placeholder |
| **Card** | Card container | className, id, children |
| **Modal** | Modal dialog | className, id, children |
| **List** | List component | className, id, items |
| **Generic** | Generic component | className, id, children |

### Implementation

```go
type UIGenerator struct {
    templates *template.Template
}

func NewUIGenerator() *UIGenerator {
    tmpl := template.New("ui")
    tmpl = tmpl.Funcs(template.FuncMap{
        "title": strings.Title,
        "lower": strings.ToLower,
    })
    
    return &UIGenerator{
        templates: template.Must(tmpl.Parse(uiComponentTemplate)),
    }
}

func (g *UIGenerator) Generate(ir *convert.IR) (*GeneratedCode, error) {
    // Extract UI components from IR tools
    components := extractUIComponents(ir)
    
    // Generate Go code with component library
    // ...
}
```

### Generated Code Structure

```go
package claude

import (
    "fmt"
    "html/template"
    "strings"
)

// UIComponentLibrary manages generated components
type UIComponentLibrary struct {
    components map[string]*UIComponent
    templates  *template.Template
}

// GenerateButton creates a button component
func (lib *UIComponentLibrary) GenerateButton(props ButtonProps) (*UIComponent, error) {
    // Generate HTML based on props
    html := fmt.Sprintf(`<button class="%s" id="%s">%s</button>`,
        className, id, children)
    
    comp := &UIComponent{
        Name:  "button",
        Type:  "button",
        HTML:  template.HTML(html),
    }
    
    return comp, nil
}
```

### Usage

```bash
# Convert UI description to Go components
ti convert --from ui --to go \
  -i ui_description.json \
  --out ./Components/ui-library \
  --kind ui-component
```

---

## 🚀 Usage

### Start Ti TUI

```bash
cd Z:\10_WORKPLACE\Ti\tui
./bin/ti-tui.exe

# Hoặc với custom config
TI_TUI_CLI_PATH=/usr/local/ti ./bin/ti-tui.exe
```

### Environment Variables

```
TI_TUI_CLI_PATH   Override path to Ti CLI directory
                    Default: Z:\10_WORKPLACE\Ti
TI_TUI_CLI_ADDR   Override gRPC address (default: localhost:1807)
TI_TUI_HUB_PORT   Override Agent Hub port (default: 1808)
```

### Add to PATH (Windows)

```powershell
[Environment]::SetEnvironmentVariable('Path', 
    [Environment]::GetEnvironmentVariable('Path', 'User') + 
    ';Z:\10_WORKPLACE\Ti\tui\cmd\tui', 
    'User')
```

---

## 🎨 Architecture Principles

### ✅ Platform Architecture (Đúng - Không Lẫn Lộn Vai Trò)

**Ti Platform (Platform Chính)**
- TUI là component riêng của Ti platform
- Không phụ thuộc vào Ti CLI
- Có binary riêng (ti-tui.exe)
- Integrate với các components khác qua đúng interfaces

**Ti CLI (Interface Riêng)**
- Chỉ là CLI interface
- Không chứa TUI
- Build riêng, deploy riêng

**Ti Router (Service Riêng)**
- Service độc lập với binary riêng (routerd.exe)
- TUI integrate qua HTTP API (http://localhost:1807)
- Không merge code

**MCP Hub (Central Hub)**
- Central MCP servers với 4 servers
- TUI integrate qua MCP client
- Không duplicate MCP functionality

**Ti Core (Shared Contracts)**
- JSON schemas cho shared contracts
- Không phải Go library
- TUI sử dụng schemas cho validation

---

## 📊 Real-time Updates

### Auto-refresh Timers

| Component | Refresh Interval |
|-----------|-----------------|
| Router health | 10s |
| MCP servers | 15s |
| Taskboard beads | 20s |

### Progress Indicators

- Context-aware loading messages
- Color-coded status messages:
  - Green: healthy/running/active/ok
  - Red: errors
  - Yellow: warnings

---

## 🔧 Build Status

```bash
✅ Module: github.com/ti/agent-tui
✅ Build: Successful
✅ Binary: Z:\10_WORKPLACE\Ti\tui\bin\ti-tui.exe
✅ Test: Running successfully
```

### Implemented Features

- ✅ Text input components (github.com/charmbracelet/bubbles/textinput)
- ✅ MCP stdio transport
- ✅ MCP HTTP transport
- ✅ MCP client với JSON-RPC protocol
- ✅ Error handling wrappers
- ✅ Nil checks
- ✅ Auto-refresh timers
- ✅ Context-aware progress indicators
- ✅ Color-coded status messages

---

## 📝 Notes

### Ti CLI Internal UI
- Ti CLI có TUI nội bộ tích hợp trực tiếp vào CLI binary
- 6 UI modules: Dashboard, Chat, Provider Wizard, Config, Memory, Plugin
- Sử dụng Bubble Tea framework cho terminal UI
- Dashboard TUI không execute commands trực tiếp (safety first)
- Provider wizard là multi-step form với 5 steps
- Support keyboard shortcuts và mouse interaction

### Ti TUI (Separate Application)
- Ti TUI là TUI chính của Ti Platform với 9 modules
- Là ứng dụng riêng biệt (ti-tui.exe), không phải phần của CLI
- Integrate với Ti Platform components qua đúng interfaces
- Dashboard mode cho phép 4-panel split view
- Real-time auto-refresh cho tất cả modules
- Robust error handling với color-coded status messages

### UIGenerator
- UIGenerator trong convert module tạo Go code cho UI components
- Template-based generation cho React/Go UI components
- Cơ bản: chỉ tạo Go code với HTML templates, không support modern frameworks

### Frontend-Design Skill
- Frontend-design skill tạo production-grade frontend interfaces với AI agent
- Support modern frameworks (React, Vue, Angular, HTML/CSS/JS)
- Unique aesthetics, full styling system, animations, responsive design

### Comparison: UIGenerator vs Frontend-Design Skill

| Feature | UIGenerator | Frontend-Design Skill |
|---------|-------------|----------------------|
| **Type** | Template-based generator | AI-powered skill |
| **Output** | Go code with HTML templates | HTML/CSS/JS, React, Vue, Angular |
| **Frameworks** | None (only Go + HTML) | All modern frameworks |
| **Styling** | Basic HTML only | Full CSS, Tailwind, custom styles |
| **Animations** | None | CSS animations, Motion library |
| **Aesthetics** | Generic | Unique, distinctive |
| **Quality** | Basic components | Production-grade |
| **Use Case** | Simple Go server UI | Full frontend applications |
| **Deterministic** | Yes (template) | No (AI-dependent) |
| **Learning Curve** | Low | Medium |

---

## 🔗 Related Resources

### Source Code
- **Ti TUI**: `apps/ui/tui/` - Separate TUI application
- **UIGenerator**: `apps/core/cli/internal/convert/generators/ui_generator.go` - UI code generator
- **CLI UI**: `apps/core/cli/internal/ui/` - CLI internal TUI modules
  - `dashboard.go` - Mission control menu
  - `tui.go` - Chat TUI (712 lines)
  - `provider_wizard.go` - Provider setup wizard (429 lines)
  - `config_tui.go` - Config management (355 lines)
  - `memory_tui.go` - Memory browser
  - `plugin_tui.go` - Plugin management
- **Router UI**: `apps/core/router/ui/`

### Documentation
- **TUI README**: `apps/ui/tui/README.md`
- **Bubble Tea**: https://github.com/charmbracelet/bubbletea
- **Lipgloss**: https://github.com/charmbracelet/lipgloss

### Related Projects
- [Ti CLI](../../../../apps/core/cli/) - Main Ti CLI
- [Ti Router](../../../../apps/core/router/) - Ti Router service
- [MCP Hub](../../../../apps/core/mcp/) - MCP servers

---

*Last Updated: 2026-05-07*
*Verified: true*
