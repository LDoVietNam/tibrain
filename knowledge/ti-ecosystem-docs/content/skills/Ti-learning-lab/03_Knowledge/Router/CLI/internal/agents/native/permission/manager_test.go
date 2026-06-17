package permission

import (
	"context"
	"testing"
)

func TestNewManager(t *testing.T) {
	config := &Config{
		Mode: ModeBalanced,
	}

	manager := NewManager(config)
	if manager == nil {
		t.Fatal("Expected non-nil manager")
	}

	if manager.GetMode() != ModeBalanced {
		t.Errorf("Expected mode Balanced, got %s", manager.GetMode())
	}
}

func TestNewManagerDefaultConfig(t *testing.T) {
	manager := NewManager(nil)
	if manager == nil {
		t.Fatal("Expected non-nil manager")
	}

	if manager.GetMode() != ModeBalanced {
		t.Errorf("Expected default mode Balanced, got %s", manager.GetMode())
	}
}

func TestSetMode(t *testing.T) {
	manager := NewManager(nil)

	manager.SetMode(ModeStrict)

	if manager.GetMode() != ModeStrict {
		t.Errorf("Expected mode Strict, got %s", manager.GetMode())
	}
}

func TestCheckStrictMode(t *testing.T) {
	manager := NewManager(&Config{Mode: ModeStrict})

	// Remove default rules for strict mode test
	manager.rules = []PermissionRule{}

	result, err := manager.Check(context.Background(), PermissionRead, "/test/file.txt")
	if err != nil {
		t.Fatalf("Check failed: %v", err)
	}

	if result.Allowed {
		t.Error("Expected strict mode to deny by default when no rules")
	}
}

func TestCheckPermissiveMode(t *testing.T) {
	manager := NewManager(&Config{Mode: ModePermissive})

	result, err := manager.Check(context.Background(), PermissionRead, "/test/file.txt")
	if err != nil {
		t.Fatalf("Check failed: %v", err)
	}

	if !result.Allowed {
		t.Error("Expected permissive mode to allow by default")
	}
}

func TestCheckBalancedMode(t *testing.T) {
	manager := NewManager(&Config{Mode: ModeBalanced})

	// Read should be allowed
	result, err := manager.Check(context.Background(), PermissionRead, "/test/file.txt")
	if err != nil {
		t.Fatalf("Check failed: %v", err)
	}

	if !result.Allowed {
		t.Error("Expected balanced mode to allow read by default")
	}

	// Write should be denied
	result, err = manager.Check(context.Background(), PermissionWrite, "/test/file.txt")
	if err != nil {
		t.Fatalf("Check failed: %v", err)
	}

	if result.Allowed {
		t.Error("Expected balanced mode to deny write by default")
	}
}

func TestAllowPath(t *testing.T) {
	manager := NewManager(nil)

	perms := []Permission{PermissionRead, PermissionWrite}
	manager.AllowPath("/safe/path", perms)

	result, err := manager.Check(context.Background(), PermissionRead, "/safe/path/file.txt")
	if err != nil {
		t.Fatalf("Check failed: %v", err)
	}

	if !result.Allowed {
		t.Error("Expected allowed path to grant permission")
	}
}

func TestDenyPath(t *testing.T) {
	manager := NewManager(nil)

	perms := []Permission{PermissionWrite, PermissionExecute}
	manager.DenyPath("/dangerous/path", perms)

	result, err := manager.Check(context.Background(), PermissionWrite, "/dangerous/path/file.txt")
	if err != nil {
		t.Fatalf("Check failed: %v", err)
	}

	if result.Allowed {
		t.Error("Expected denied path to deny permission")
	}
}

func TestAddRule(t *testing.T) {
	manager := NewManager(nil)

	rule := PermissionRule{
		Action:      PermissionRead,
		Resource:    "test",
		Pattern:     "*",
		Allowed:     true,
		Description: "Allow all test reads",
	}

	manager.AddRule(rule)

	rules := manager.GetRules()
	if len(rules) == 0 {
		t.Error("Expected rule to be added")
	}
}

func TestRemoveRule(t *testing.T) {
	manager := NewManager(nil)

	rule := PermissionRule{
		Action:      PermissionRead,
		Resource:    "test",
		Pattern:     "*",
		Allowed:     true,
		Description: "Test rule to remove",
	}

	manager.AddRule(rule)

	removed := manager.RemoveRule("Test rule to remove")
	if !removed {
		t.Error("Expected rule to be removed")
	}

	// Try to remove again
	removed = manager.RemoveRule("Test rule to remove")
	if removed {
		t.Error("Expected rule to not be found on second removal")
	}
}

func TestCheckFilePermission(t *testing.T) {
	manager := NewManager(nil)

	// Create a temp file
	tmpFile := "/tmp/test-permission-check.txt"

	result, err := manager.CheckFilePermission(context.Background(), PermissionRead, tmpFile)
	if err != nil {
		t.Fatalf("CheckFilePermission failed: %v", err)
	}

	// Should be allowed (balanced mode, read operation)
	if !result.Allowed {
		t.Error("Expected file read to be allowed")
	}
}

func TestCheckNetworkPermission(t *testing.T) {
	manager := NewManager(nil)

	result, err := manager.CheckNetworkPermission(context.Background(), "localhost", 8080)
	if err != nil {
		t.Fatalf("CheckNetworkPermission failed: %v", err)
	}

	// Should be denied by default (network operations)
	if result.Allowed {
		t.Error("Expected network operation to be denied by default")
	}
}

func TestPathMatching(t *testing.T) {
	manager := NewManager(nil)

	// Test wildcard matching
	if !manager.pathMatches("/test/path", "/test/*") {
		t.Error("Expected wildcard match")
	}

	if !manager.pathMatches("/test/path", "*") {
		t.Error("Expected wildcard match")
	}

	// Test exact match
	if !manager.pathMatches("/test/path", "/test/path") {
		t.Error("Expected exact match")
	}

	// Test no match
	if manager.pathMatches("/test/path", "/other/path") {
		t.Error("Expected no match")
	}
}

func TestMatchesRule(t *testing.T) {
	manager := NewManager(nil)

	rule := PermissionRule{
		Action:  PermissionRead,
		Pattern: "/test",
		Allowed: true,
	}

	// Should match
	if !manager.matchesRule(PermissionRead, "/test/file", rule) {
		t.Error("Expected rule to match")
	}

	// Should not match (different action)
	if manager.matchesRule(PermissionWrite, "/test/file", rule) {
		t.Error("Expected rule to not match different action")
	}

	// Should match wildcard
	ruleWildcard := PermissionRule{
		Action:  PermissionRead,
		Pattern: "*",
		Allowed: true,
	}

	if !manager.matchesRule(PermissionRead, "/test/file", ruleWildcard) {
		t.Error("Expected wildcard rule to match")
	}
}
