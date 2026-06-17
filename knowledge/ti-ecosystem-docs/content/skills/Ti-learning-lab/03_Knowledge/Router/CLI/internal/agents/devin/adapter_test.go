package devin

import (
	"context"
	"testing"
)

func TestTiToDevinTranslation(t *testing.T) {
	adapter := NewAdapter(&Config{})

	tiCtx := &TiContext{
		Workspace:   "/path/to/workspace",
		SessionID:   "test-session",
		Environment: map[string]string{"EDITOR": "vim", "SHELL": "bash"},
		UserIntent:  "Fix the bug in router.go",
		Metadata:    map[string]interface{}{"task_type": "debugging"},
	}

	devinCtx, err := adapter.TiToDevin(context.Background(), tiCtx, TaskTypeDebugging)
	if err != nil {
		t.Fatalf("TiToDevin failed: %v", err)
	}

	if devinCtx.Workspace != "/path/to/workspace" {
		t.Errorf("Expected workspace '/path/to/workspace', got '%s'", devinCtx.Workspace)
	}

	if devinCtx.SessionID != "test-session" {
		t.Errorf("Expected session_id 'test-session', got '%s'", devinCtx.SessionID)
	}

	if len(devinCtx.Tools) == 0 {
		t.Error("Expected tools to be populated")
	}

	if devinCtx.Options["editor"] != "vim" {
		t.Errorf("Expected editor 'vim', got '%v'", devinCtx.Options["editor"])
	}
}

func TestDevinToTiTranslation(t *testing.T) {
	adapter := NewAdapter(&Config{})

	tiCtx := &TiContext{
		Workspace:   "/path/to/workspace",
		SessionID:   "test-session",
		Environment: map[string]string{},
		UserIntent:  "Test",
		Metadata:    map[string]interface{}{},
	}

	devinOutput := map[string]interface{}{
		"success":  true,
		"result":   "Task completed successfully",
		"metadata": map[string]interface{}{"execution_time": "1.5s"},
		"errors":   []interface{}{},
	}

	tiOutput, err := adapter.DevinToTi(context.Background(), devinOutput, tiCtx)
	if err != nil {
		t.Fatalf("DevinToTi failed: %v", err)
	}

	if tiOutput["success"] != true {
		t.Errorf("Expected success true, got %v", tiOutput["success"])
	}

	if tiOutput["result"] != "Task completed successfully" {
		t.Errorf("Expected result 'Task completed successfully', got '%v'", tiOutput["result"])
	}

	if tiOutput["session_id"] != "test-session" {
		t.Errorf("Expected session_id 'test-session', got '%v'", tiOutput["session_id"])
	}
}

func TestGetToolsForTask(t *testing.T) {
	adapter := NewAdapter(&Config{})

	toolsCodeEdit := adapter.getToolsForTask(TaskTypeCodeEdit)
	if len(toolsCodeEdit) == 0 {
		t.Error("Expected tools for code edit")
	}

	hasEdit := false
	for _, tool := range toolsCodeEdit {
		if tool == "edit" {
			hasEdit = true
		}
	}
	if !hasEdit {
		t.Error("Expected 'edit' tool in code edit tools")
	}
}

func TestTranslatePrompt(t *testing.T) {
	adapter := NewAdapter(&Config{})

	prompt := adapter.translatePrompt("Fix the bug", TaskTypeDebugging)
	if prompt == "" {
		t.Error("Expected non-empty prompt")
	}

	if len(prompt) < len("Fix the bug") {
		t.Error("Expected prompt to be longer than original intent")
	}
}

func TestInjectSkillPrompt(t *testing.T) {
	adapter := NewAdapter(&Config{})

	original := "Please review this code"
	result := adapter.InjectSkillPrompt(original, "code_review")

	if result == original {
		t.Error("Expected prompt to be modified with skill prompt")
	}

	// Check for known skill content
	if len(result) < len(original) {
		t.Error("Expected result to be longer than original")
	}
}

func TestInjectUnknownSkill(t *testing.T) {
	adapter := NewAdapter(&Config{})

	original := "Test prompt"
	result := adapter.InjectSkillPrompt(original, "unknown_skill")

	if result != original {
		t.Error("Expected original prompt to be returned for unknown skill")
	}
}

func TestConfigDefault(t *testing.T) {
	config := DefaultConfig()

	if config.Port == 0 {
		t.Error("Expected default port to be set")
	}

	if config.Timeout == 0 {
		t.Error("Expected default timeout to be set")
	}

	if config.PythonExecutable == "" {
		t.Error("Expected default Python executable to be set")
	}
}

func TestConfigValidate(t *testing.T) {
	config := DefaultConfig()

	err := config.Validate()
	if err != nil {
		t.Errorf("Default config should be valid: %v", err)
	}

	// Test invalid port
	config.Port = -1
	err = config.Validate()
	if err == nil {
		t.Error("Expected validation error for invalid port")
	}
}
