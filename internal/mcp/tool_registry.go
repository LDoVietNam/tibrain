package mcp

import (
	"sync"
)

// ValidatePath checks if a path is within allowed roots.
func ValidatePath(allowedRoots []string, path string) bool {
	if len(allowedRoots) == 0 {
		return true
	}
	for _, root := range allowedRoots {
		if len(path) >= len(root) && path[:len(root)] == root {
			return true
		}
	}
	return false
}

// ToolRegistry manages tool registration and discovery.
type ToolRegistry struct {
	mu    sync.RWMutex
	tools map[string]*ToolMetadata
}

// ToolMetadata contains metadata for a registered tool.
type ToolMetadata struct {
	Name        string
	Description string
	Category    string
	SafeLevel   SafeLevel
}

// SafeLevel indicates the safety level of a tool.
type SafeLevel int

const (
	SafeRead SafeLevel = iota
	SafeWrite
	SafeDestructive
)

// Global tool registry instance.
var Registry = &ToolRegistry{
	tools: make(map[string]*ToolMetadata),
}

// Register registers a tool with metadata.
func (r *ToolRegistry) Register(name, desc, category string, safeLevel SafeLevel) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.tools[name] = &ToolMetadata{
		Name:        name,
		Description: desc,
		Category:    category,
		SafeLevel:   safeLevel,
	}
}

// PolicyEngine evaluates tool access based on policies.
type PolicyEngine struct {
	allowedRoots []string
}

// NewPolicyEngine creates a new policy engine.
func NewPolicyEngine(allowedRoots []string) *PolicyEngine {
	return &PolicyEngine{allowedRoots: allowedRoots}
}

// ValidatePath checks if a path is within allowed roots.
func (p *PolicyEngine) ValidatePath(path string) bool {
	return ValidatePath(p.allowedRoots, path)
}

// RegisterFileTools registers tools (stub - used by mcp/server.go)
func RegisterFileTools() {
	Registry.Register("read_file", "Read file contents", "filesystem", SafeRead)
	Registry.Register("write_file", "Write file contents", "filesystem", SafeWrite)
	Registry.Register("list_files", "List directory contents", "filesystem", SafeRead)
	Registry.Register("delete_file", "Delete file", "filesystem", SafeDestructive)
	Registry.Register("file_info", "Get file info", "filesystem", SafeRead)
}