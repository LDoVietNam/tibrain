package main

import (
	"fmt"
	"strings"
	"sync"
)

// LogicalTarget maps a human alias to an internal MCP server name.
type LogicalTarget struct {
	Alias      string // e.g. "filesystem"
	ServerName string // e.g. "1mcp-local-bridge"
}

// TargetRegistry resolves logical target aliases to internal server names.
type TargetRegistry struct {
	mu      sync.RWMutex
	targets map[string]LogicalTarget
}

var globalTargetRegistry = newTargetRegistry()

func newTargetRegistry() *TargetRegistry {
	r := &TargetRegistry{
		targets: make(map[string]LogicalTarget),
	}
	r.RegisterDefaults()
	return r
}

// RegisterDefaults registers the hardcoded default target mappings.
func (r *TargetRegistry) RegisterDefaults() {
	defaults := []LogicalTarget{
		{Alias: "filesystem", ServerName: "1mcp-local-bridge"},
		{Alias: "fs", ServerName: "1mcp-local-bridge"},
		{Alias: "files", ServerName: "1mcp-local-bridge"},
		{Alias: "router", ServerName: "omniroute-router"},
		{Alias: "omniroute", ServerName: "omniroute-router"},
		{Alias: "browser", ServerName: "browser-extension"},
		{Alias: "obsidian", ServerName: "obsidian-mcp-server"},
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, t := range defaults {
		r.targets[strings.ToLower(t.Alias)] = t
	}
}

// Register adds or overwrites a target mapping.
func (r *TargetRegistry) Register(alias, serverName string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	alias = strings.ToLower(strings.TrimSpace(alias))
	if alias != "" && serverName != "" {
		r.targets[alias] = LogicalTarget{Alias: alias, ServerName: serverName}
	}
}

// Resolve returns the LogicalTarget for a given alias.
// If the alias is not found, it is treated as a direct server name.
func (r *TargetRegistry) Resolve(alias string) (LogicalTarget, error) {
	key := strings.ToLower(strings.TrimSpace(alias))
	r.mu.RLock()
	t, ok := r.targets[key]
	r.mu.RUnlock()
	if ok {
		return t, nil
	}
	// Fall back: treat alias as direct server name
	if key != "" {
		return LogicalTarget{Alias: alias, ServerName: alias}, nil
	}
	return LogicalTarget{}, fmt.Errorf("target %q is not registered and cannot be inferred", alias)
}

// List returns all registered targets.
func (r *TargetRegistry) List() []LogicalTarget {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]LogicalTarget, 0, len(r.targets))
	for _, t := range r.targets {
		out = append(out, t)
	}
	return out
}

// LoadFromConfig loads target mappings from config entries.
// map key = alias, value = server name.
func (r *TargetRegistry) LoadFromConfig(targets map[string]string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for alias, serverName := range targets {
		alias = strings.ToLower(strings.TrimSpace(alias))
		if alias != "" && serverName != "" {
			r.targets[alias] = LogicalTarget{Alias: alias, ServerName: serverName}
		}
	}
}
