// Package opencode provides OpenCode sidecar integration for Ti CLI.
package opencode

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Config holds OpenCode sidecar configuration.
type Config struct {
	Enabled         bool     `json:"enabled"`
	Path            string   `json:"path"`             // path to opencode binary (default: "opencode")
	Port            int      `json:"port"`             // sidecar server port (default: 4096)
	Model           string   `json:"model"`            // e.g. "anthropic/claude-sonnet-4-5"
	Agent           string   `json:"agent"`            // e.g. "coder"
	Plugins         []string `json:"plugins"`          // plugin specs
	AutoServe       bool     `json:"auto_serve"`       // auto-start sidecar on run
	SkipPermissions bool     `json:"skip_permissions"` // --dangerously-skip-permissions
}

// DefaultConfig returns sensible defaults.
func DefaultConfig() Config {
	return Config{
		Enabled:         false,
		Path:            "opencode",
		Port:            4096,
		SkipPermissions: true,
	}
}

// tiJSON is a minimal subset of ti.json for reading the opencode block.
type tiJSON struct {
	Opencode *Config `json:"opencode"`
}

// LoadConfig reads opencode config from ti.json in the current directory,
// falling back to defaults if the file or key is absent.
func LoadConfig() Config {
	cfg := DefaultConfig()

	data, err := os.ReadFile(filepath.Join(".", "ti.json"))
	if err != nil {
		return cfg
	}

	var tj tiJSON
	if err := json.Unmarshal(data, &tj); err != nil {
		return cfg
	}

	if tj.Opencode == nil {
		return cfg
	}

	merged := *tj.Opencode
	if merged.Path == "" {
		merged.Path = cfg.Path
	}
	if merged.Port == 0 {
		merged.Port = cfg.Port
	}
	return merged
}
