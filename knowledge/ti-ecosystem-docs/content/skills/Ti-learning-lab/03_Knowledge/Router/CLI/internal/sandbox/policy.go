package sandbox

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

const (
	BackendAuto       = "auto"
	BackendPortable   = "portable"
	BackendBubblewrap = "bubblewrap"
	BackendDocker     = "docker"
)

// Policy controls how Ti runs shell commands for tools and agents.
// The portable backend is a guardrail runner, not a kernel-level sandbox.
// Bubblewrap/Docker backends provide stronger isolation when available.
type Policy struct {
	Enabled        bool     `json:"enabled"`
	Backend        string   `json:"backend"`
	Fallback       string   `json:"fallback"`
	Network        string   `json:"network"` // deny | allow
	TimeoutSeconds int      `json:"timeout_seconds"`
	MaxOutputBytes int      `json:"max_output_bytes"`
	ReadAllow      []string `json:"read_allow"`
	WriteAllow     []string `json:"write_allow"`
	PathDeny       []string `json:"path_deny"`
	EnvAllow       []string `json:"env_allow"`
	CommandDeny    []string `json:"command_deny"`
	AuditFile      string   `json:"audit_file"`
}

func DefaultPolicy() *Policy {
	return &Policy{
		Enabled:        true,
		Backend:        BackendAuto,
		Fallback:       BackendPortable,
		Network:        "deny",
		TimeoutSeconds: 120,
		MaxOutputBytes: 200000,
		ReadAllow:      []string{"./"},
		WriteAllow:     []string{"./", "./tmp"},
		PathDeny: []string{
			".env", ".env.*", "**/.env", "**/.env.*",
			"~/.ssh/**", "~/.config/**", "~/.aws/**", "~/.gnupg/**",
			".git/**", "node_modules/**",
		},
		EnvAllow: []string{
			"PATH", "HOME", "USER", "USERNAME", "SHELL",
			"TMPDIR", "TEMP", "TMP",
			"GOCACHE", "GOMODCACHE", "GOTOOLCHAIN", "GOWORK",
			"HTTP_PROXY", "HTTPS_PROXY", "NO_PROXY",
		},
		CommandDeny: []string{
			"rm -rf /", "rm -rf .", "rm -rf *", "del /s", "format ",
			"mkfs", "dd if=", "shutdown", "reboot", "poweroff",
			":(){", "fork bomb", "chmod -R 777 /", "chown -R",
		},
		AuditFile: filepath.Join("~", ".ti", "data", "sandbox", "audit.jsonl"),
	}
}

func DefaultPolicyPath() string {
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return filepath.Join(".ti", "sandbox", "policy.json")
	}
	return filepath.Join(home, ".ti", "sandbox", "policy.json")
}

func ExpandPath(p string) string {
	if p == "" {
		return p
	}
	if p == "~" || strings.HasPrefix(p, "~/") || strings.HasPrefix(p, "~"+string(filepath.Separator)) {
		if home, err := os.UserHomeDir(); err == nil && home != "" {
			if p == "~" {
				return home
			}
			return filepath.Join(home, p[2:])
		}
	}
	return os.ExpandEnv(p)
}

func LoadPolicy(path string) (*Policy, error) {
	if path == "" {
		path = DefaultPolicyPath()
	}
	data, err := os.ReadFile(ExpandPath(path))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return DefaultPolicy(), nil
		}
		return nil, err
	}
	p := DefaultPolicy()
	if err := json.Unmarshal(data, p); err != nil {
		return nil, err
	}
	p.Normalize()
	return p, nil
}

func SavePolicy(path string, p *Policy) error {
	if path == "" {
		path = DefaultPolicyPath()
	}
	p.Normalize()
	target := ExpandPath(path)
	if err := os.MkdirAll(filepath.Dir(target), 0o700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(target, data, 0o600)
}

func EnsurePolicy(path string, overwrite bool) (*Policy, bool, error) {
	if path == "" {
		path = DefaultPolicyPath()
	}
	target := ExpandPath(path)
	if !overwrite {
		if _, err := os.Stat(target); err == nil {
			p, loadErr := LoadPolicy(target)
			return p, false, loadErr
		}
	}
	p := DefaultPolicy()
	if err := SavePolicy(target, p); err != nil {
		return nil, false, err
	}
	return p, true, nil
}

func (p *Policy) Normalize() {
	if p.Backend == "" {
		p.Backend = BackendAuto
	}
	if p.Fallback == "" {
		p.Fallback = BackendPortable
	}
	if p.Network == "" {
		p.Network = "deny"
	}
	if p.TimeoutSeconds <= 0 {
		p.TimeoutSeconds = 120
	}
	if p.TimeoutSeconds > 3600 {
		p.TimeoutSeconds = 3600
	}
	if p.MaxOutputBytes <= 0 {
		p.MaxOutputBytes = 200000
	}
	if p.AuditFile == "" {
		p.AuditFile = DefaultPolicy().AuditFile
	}
	if len(p.EnvAllow) == 0 {
		p.EnvAllow = DefaultPolicy().EnvAllow
	}
	if len(p.CommandDeny) == 0 {
		p.CommandDeny = DefaultPolicy().CommandDeny
	}
	if len(p.PathDeny) == 0 {
		p.PathDeny = DefaultPolicy().PathDeny
	}
	if runtime.GOOS == "windows" && p.Backend == BackendBubblewrap {
		p.Backend = BackendPortable
	}
}

func (p *Policy) Timeout() time.Duration {
	p.Normalize()
	return time.Duration(p.TimeoutSeconds) * time.Second
}

func (p *Policy) AllowNetwork() bool {
	return strings.EqualFold(p.Network, "allow") || strings.EqualFold(p.Network, "on") || strings.EqualFold(p.Network, "true")
}

func (p *Policy) Summary() string {
	p.Normalize()
	return fmt.Sprintf("enabled=%v backend=%s fallback=%s network=%s timeout=%ds max_output=%d", p.Enabled, p.Backend, p.Fallback, p.Network, p.TimeoutSeconds, p.MaxOutputBytes)
}
