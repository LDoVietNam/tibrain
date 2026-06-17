package permission

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// Permission represents a permission type
type Permission string

const (
	PermissionRead    Permission = "read"
	PermissionWrite   Permission = "write"
	PermissionExecute Permission = "execute"
	PermissionNetwork Permission = "network"
	PermissionSystem  Permission = "system"
)

// PermissionMode represents the permission mode
type PermissionMode string

const (
	ModeStrict     PermissionMode = "strict"     // Explicit allow only
	ModePermissive PermissionMode = "permissive" // Allow by default, deny explicit
	ModeBalanced   PermissionMode = "balanced"   // Balanced approach
)

// PermissionResult represents the result of a permission check
type PermissionResult struct {
	Allowed bool
	Reason  string
	Mode    PermissionMode
}

// Manager manages permissions
type Manager struct {
	mu           sync.RWMutex
	mode         PermissionMode
	rules        []PermissionRule
	allowedPaths map[string][]Permission
	deniedPaths  map[string][]Permission
}

// PermissionRule represents a permission rule
type PermissionRule struct {
	Action      Permission
	Resource    string
	Pattern     string
	Allowed     bool
	Description string
}

// Config holds permission manager configuration
type Config struct {
	Mode         PermissionMode
	Rules        []PermissionRule
	AllowedPaths map[string][]Permission
	DeniedPaths  map[string][]Permission
}

// NewManager creates a new permission manager
func NewManager(config *Config) *Manager {
	if config == nil {
		config = &Config{
			Mode:         ModeBalanced,
			Rules:        []PermissionRule{},
			AllowedPaths: make(map[string][]Permission),
			DeniedPaths:  make(map[string][]Permission),
		}
	}

	manager := &Manager{
		mode:         config.Mode,
		rules:        config.Rules,
		allowedPaths: config.AllowedPaths,
		deniedPaths:  config.DeniedPaths,
	}

	// Initialize default rules
	manager.initDefaultRules()

	return manager
}

// initDefaultRules initializes default permission rules
func (m *Manager) initDefaultRules() {
	// Default safe rules
	m.rules = append(m.rules, PermissionRule{
		Action:      PermissionRead,
		Resource:    "workspace",
		Pattern:     "*",
		Allowed:     true,
		Description: "Allow reading from workspace",
	})

	m.rules = append(m.rules, PermissionRule{
		Action:      PermissionWrite,
		Resource:    "workspace",
		Pattern:     "*.go",
		Allowed:     true,
		Description: "Allow writing Go files in workspace",
	})

	m.rules = append(m.rules, PermissionRule{
		Action:      PermissionSystem,
		Resource:    "system",
		Pattern:     "*",
		Allowed:     false,
		Description: "Deny system operations by default",
	})
}

// Check checks if an action is allowed on a resource
func (m *Manager) Check(ctx context.Context, action Permission, resource string) (*PermissionResult, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := &PermissionResult{
		Allowed: false,
		Reason:  "",
		Mode:    m.mode,
	}

	// Check denied paths first
	if m.isDenied(action, resource) {
		result.Reason = fmt.Sprintf("Action %s on %s is denied by path rules", action, resource)
		return result, nil
	}

	// Check allowed paths
	if m.isAllowed(action, resource) {
		result.Allowed = true
		result.Reason = fmt.Sprintf("Action %s on %s is allowed by path rules", action, resource)
		return result, nil
	}

	// Check rules
	for _, rule := range m.rules {
		if m.matchesRule(action, resource, rule) {
			result.Allowed = rule.Allowed
			result.Reason = rule.Description
			return result, nil
		}
	}

	// Default behavior based on mode
	switch m.mode {
	case ModeStrict:
		result.Reason = "Strict mode: denied by default (no matching rule)"
		return result, nil
	case ModePermissive:
		result.Allowed = true
		result.Reason = "Permissive mode: allowed by default (no matching rule)"
		return result, nil
	case ModeBalanced:
		// Safe operations allowed by default
		if action == PermissionRead {
			result.Allowed = true
			result.Reason = "Balanced mode: read operations allowed by default"
		} else {
			result.Reason = "Balanced mode: write operations denied by default"
		}
		return result, nil
	default:
		result.Reason = "Unknown mode"
		return result, nil
	}
}

// isDenied checks if action is denied by path rules
func (m *Manager) isDenied(action Permission, resource string) bool {
	// Check denied paths
	for path, deniedPerms := range m.deniedPaths {
		if m.pathMatches(resource, path) {
			for _, perm := range deniedPerms {
				if perm == action || perm == "*" {
					return true
				}
			}
		}
	}
	return false
}

// isAllowed checks if action is allowed by path rules
func (m *Manager) isAllowed(action Permission, resource string) bool {
	// Check allowed paths
	for path, allowedPerms := range m.allowedPaths {
		if m.pathMatches(resource, path) {
			for _, perm := range allowedPerms {
				if perm == action || perm == "*" {
					return true
				}
			}
		}
	}
	return false
}

// matchesRule checks if action and resource match a rule
func (m *Manager) matchesRule(action Permission, resource string, rule PermissionRule) bool {
	if rule.Action != action && rule.Action != "*" {
		return false
	}

	if rule.Pattern == "*" {
		return true
	}

	return strings.Contains(resource, rule.Pattern)
}

// pathMatches checks if resource matches path pattern
func (m *Manager) pathMatches(resource, pattern string) bool {
	// Simple wildcard matching
	if pattern == "*" {
		return true
	}

	if strings.HasSuffix(pattern, "*") {
		prefix := strings.TrimSuffix(pattern, "*")
		return strings.HasPrefix(resource, prefix)
	}

	if strings.HasPrefix(pattern, "*") {
		suffix := strings.TrimPrefix(pattern, "*")
		return strings.HasSuffix(resource, suffix)
	}

	return resource == pattern
}

// AllowPath adds an allowed path with permissions
func (m *Manager) AllowPath(path string, permissions []Permission) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.allowedPaths == nil {
		m.allowedPaths = make(map[string][]Permission)
	}

	m.allowedPaths[path] = permissions
}

// DenyPath adds a denied path with permissions
func (m *Manager) DenyPath(path string, permissions []Permission) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.deniedPaths == nil {
		m.deniedPaths = make(map[string][]Permission)
	}

	m.deniedPaths[path] = permissions
}

// AddRule adds a permission rule
func (m *Manager) AddRule(rule PermissionRule) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.rules = append(m.rules, rule)
}

// RemoveRule removes a permission rule by description
func (m *Manager) RemoveRule(description string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	for i, rule := range m.rules {
		if rule.Description == description {
			m.rules = append(m.rules[:i], m.rules[i+1:]...)
			return true
		}
	}

	return false
}

// SetMode sets the permission mode
func (m *Manager) SetMode(mode PermissionMode) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.mode = mode
}

// GetMode returns the current permission mode
func (m *Manager) GetMode() PermissionMode {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.mode
}

// GetRules returns all permission rules
func (m *Manager) GetRules() []PermissionRule {
	m.mu.RLock()
	defer m.mu.RUnlock()

	rules := make([]PermissionRule, len(m.rules))
	copy(rules, m.rules)
	return rules
}

// CheckFilePermission checks permission for a file operation
func (m *Manager) CheckFilePermission(ctx context.Context, action Permission, filePath string) (*PermissionResult, error) {
	// Resolve absolute path
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve path: %w", err)
	}

	// Check if file exists
	if _, err := os.Stat(absPath); err != nil {
		if os.IsNotExist(err) {
			// File doesn't exist, check parent directory
			parentDir := filepath.Dir(absPath)
			return m.Check(ctx, action, parentDir)
		}
		return nil, fmt.Errorf("failed to stat file: %w", err)
	}

	return m.Check(ctx, action, absPath)
}

// CheckDirectoryPermission checks permission for a directory operation
func (m *Manager) CheckDirectoryPermission(ctx context.Context, action Permission, dirPath string) (*PermissionResult, error) {
	absPath, err := filepath.Abs(dirPath)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve path: %w", err)
	}

	return m.Check(ctx, action, absPath)
}

// CheckNetworkPermission checks permission for network operation
func (m *Manager) CheckNetworkPermission(ctx context.Context, host string, port int) (*PermissionResult, error) {
	resource := fmt.Sprintf("%s:%d", host, port)
	return m.Check(ctx, PermissionNetwork, resource)
}
