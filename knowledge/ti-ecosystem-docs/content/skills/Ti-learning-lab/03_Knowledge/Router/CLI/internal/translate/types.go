package translate

import "time"

const (
	FormatAuto       = "auto"
	FormatCodex      = "codex"
	FormatClaudeCode = "claude-code"
	FormatOpenCode   = "opencode"
	FormatCursor     = "cursor"
	FormatWindsurf   = "windsurf"
	FormatMCP        = "mcp"
	FormatDevin      = "devin"
	FormatAider      = "aider"
	FormatRoo        = "roo"
	FormatKilo       = "kilo"
	FormatGeneric    = "generic"
	FormatTiPack     = "ti-pack"
)

type ScanOptions struct {
	MaxDepth int
	From     string
}

type Finding struct {
	Path       string   `json:"path"`
	Format     string   `json:"format"`
	Kind       string   `json:"kind"`
	Name       string   `json:"name"`
	Confidence float64  `json:"confidence"`
	Notes      []string `json:"notes,omitempty"`
	Risks      []string `json:"risks,omitempty"`
}

type FrontMatter map[string]string

type Pack struct {
	APIVersion   string        `json:"apiVersion"`
	Kind         string        `json:"kind"`
	Metadata     Metadata      `json:"metadata"`
	Instructions []Instruction `json:"instructions,omitempty"`
	Commands     []Command     `json:"commands,omitempty"`
	Agents       []Agent       `json:"agents,omitempty"`
	Skills       []Skill       `json:"skills,omitempty"`
	Workflows    []Workflow    `json:"workflows,omitempty"`
	Hooks        []Hook        `json:"hooks,omitempty"`
	Plugins      []Plugin      `json:"plugins,omitempty"`
	MCPServers   []MCPServer   `json:"mcpServers,omitempty"`
	Providers    []ProviderRef `json:"providers,omitempty"`
	Permissions  []Permission  `json:"permissions,omitempty"`
	Conflicts    []Conflict    `json:"conflicts,omitempty"`
	Risks        []Risk        `json:"risks,omitempty"`
}

type Metadata struct {
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	GeneratedAt time.Time `json:"generatedAt"`
	Source      string    `json:"source"`
	Version     string    `json:"version,omitempty"`
}

type SourceRef struct {
	Format string `json:"format"`
	Path   string `json:"path"`
	Kind   string `json:"kind"`
}

type Instruction struct {
	ID      string            `json:"id"`
	Name    string            `json:"name"`
	Scope   string            `json:"scope"`
	Globs   []string          `json:"globs,omitempty"`
	Content string            `json:"content"`
	Meta    map[string]string `json:"meta,omitempty"`
	Source  SourceRef         `json:"source"`
}

type Command struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description,omitempty"`
	Prompt      string            `json:"prompt"`
	Args        []Arg             `json:"args,omitempty"`
	Tools       []string          `json:"tools,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
	Sandbox     bool              `json:"sandbox,omitempty"`
	Source      SourceRef         `json:"source"`
}

type Arg struct {
	Name        string `json:"name"`
	Default     string `json:"default,omitempty"`
	Description string `json:"description,omitempty"`
	Required    bool   `json:"required,omitempty"`
}

type Agent struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description,omitempty"`
	Prompt      string            `json:"prompt"`
	Model       string            `json:"model,omitempty"`
	Tools       []string          `json:"tools,omitempty"`
	Mode        string            `json:"mode,omitempty"`
	Sandbox     bool              `json:"sandbox,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
	Source      SourceRef         `json:"source"`
}

type Skill struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description,omitempty"`
	Content     string            `json:"content"`
	Tags        []string          `json:"tags,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
	Source      SourceRef         `json:"source"`
}

type Workflow struct {
	ID          string         `json:"id"`
	Name        string         `json:"name"`
	Description string         `json:"description,omitempty"`
	Steps       []WorkflowStep `json:"steps"`
	Source      SourceRef      `json:"source"`
}

type WorkflowStep struct {
	Name    string `json:"name,omitempty"`
	Run     string `json:"run"`
	Sandbox bool   `json:"sandbox,omitempty"`
}

type Hook struct {
	Event   string            `json:"event"`
	Run     string            `json:"run"`
	Matcher string            `json:"matcher,omitempty"`
	Sandbox bool              `json:"sandbox,omitempty"`
	Meta    map[string]string `json:"meta,omitempty"`
	Source  SourceRef         `json:"source"`
}

type Plugin struct {
	Name        string            `json:"name"`
	Runtime     string            `json:"runtime"`
	Command     string            `json:"command,omitempty"`
	Args        []string          `json:"args,omitempty"`
	Env         map[string]string `json:"env,omitempty"`
	Sandbox     bool              `json:"sandbox"`
	Description string            `json:"description,omitempty"`
	Source      SourceRef         `json:"source"`
}

type MCPServer struct {
	Name    string            `json:"name"`
	Command string            `json:"command,omitempty"`
	Args    []string          `json:"args,omitempty"`
	Env     map[string]string `json:"env,omitempty"`
	URL     string            `json:"url,omitempty"`
	Sandbox bool              `json:"sandbox"`
	Source  SourceRef         `json:"source"`
}

type ProviderRef struct {
	Name         string            `json:"name"`
	Type         string            `json:"type,omitempty"`
	BaseURL      string            `json:"base_url,omitempty"`
	APIKeyEnv    string            `json:"api_key_env,omitempty"`
	DefaultModel string            `json:"default_model,omitempty"`
	Models       []string          `json:"models,omitempty"`
	Headers      map[string]string `json:"headers,omitempty"`
	Source       SourceRef         `json:"source"`
}

type Permission struct {
	Subject string    `json:"subject"`
	Action  string    `json:"action"`
	Paths   []string  `json:"paths,omitempty"`
	Mode    string    `json:"mode"`
	Source  SourceRef `json:"source"`
}

type Conflict struct {
	ID      string      `json:"id"`
	Level   string      `json:"level"`
	Topic   string      `json:"topic"`
	Message string      `json:"message"`
	Sources []SourceRef `json:"sources"`
}

type Risk struct {
	Level  string    `json:"level"`
	Reason string    `json:"reason"`
	Action string    `json:"action"`
	Source SourceRef `json:"source"`
}

type LockFile struct {
	GeneratedAt time.Time   `json:"generated_at"`
	Version     string      `json:"version"`
	Imports     []LockEntry `json:"imports"`
}

type LockEntry struct {
	SourcePath string    `json:"source_path"`
	Format     string    `json:"format"`
	Kind       string    `json:"kind"`
	ImportedAs string    `json:"imported_as"`
	SHA256     string    `json:"sha256"`
	ImportedAt time.Time `json:"imported_at"`
}
