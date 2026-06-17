package devin

import (
	"context"
	"fmt"
)

// Adapter adapts Ti context to Devin context and vice versa
type Adapter struct {
	config *Config
}

// NewAdapter creates a new Devin adapter
func NewAdapter(config *Config) *Adapter {
	return &Adapter{
		config: config,
	}
}

// TiContext represents Ti CLI context
type TiContext struct {
	Workspace   string
	SessionID   string
	Environment map[string]string
	UserIntent  string
	Metadata    map[string]interface{}
}

// DevinContext represents Devin CLI context
type DevinContext struct {
	Prompt    string
	Workspace string
	SessionID string
	Tools     []string
	Options   map[string]interface{}
	Metadata  map[string]string
}

// TaskType represents the type of task
type TaskType string

const (
	TaskTypeCodeEdit      TaskType = "code_edit"
	TaskTypeFileOperation TaskType = "file_operation"
	TaskTypeShellCommand  TaskType = "shell_command"
	TaskTypeAnalysis      TaskType = "analysis"
	TaskTypeRefactoring   TaskType = "refactoring"
	TaskTypeDebugging     TaskType = "debugging"
	TaskTypeGeneral       TaskType = "general"
)

// TiToDevin translates Ti context to Devin context
func (a *Adapter) TiToDevin(ctx context.Context, tiCtx *TiContext, taskType TaskType) (*DevinContext, error) {
	devinCtx := &DevinContext{
		Workspace: tiCtx.Workspace,
		SessionID: tiCtx.SessionID,
		Tools:     a.getToolsForTask(taskType),
		Options:   a.translateOptions(tiCtx.Environment),
		Metadata:  a.translateMetadata(tiCtx.Metadata),
		Prompt:    a.translatePrompt(tiCtx.UserIntent, taskType),
	}

	return devinCtx, nil
}

// DevinToTi translates Devin output to Ti format
func (a *Adapter) DevinToTi(ctx context.Context, devinOutput map[string]interface{}, tiCtx *TiContext) (map[string]interface{}, error) {
	tiOutput := map[string]interface{}{
		"success":    false,
		"result":     "",
		"metadata":   map[string]interface{}{},
		"errors":     []interface{}{},
		"session_id": tiCtx.SessionID,
	}

	if success, ok := devinOutput["success"].(bool); ok {
		tiOutput["success"] = success
	}

	if result, ok := devinOutput["result"].(string); ok {
		tiOutput["result"] = a.translateResult(result)
	}

	if metadata, ok := devinOutput["metadata"].(map[string]interface{}); ok {
		tiOutput["metadata"] = a.translateOutputMetadata(metadata)
	}

	if errors, ok := devinOutput["errors"].([]interface{}); ok {
		tiOutput["errors"] = errors
	}

	return tiOutput, nil
}

// getToolsForTask returns the list of tools needed for a task type
func (a *Adapter) getToolsForTask(taskType TaskType) []string {
	toolMap := map[TaskType][]string{
		TaskTypeCodeEdit:      {"edit", "read", "write", "search"},
		TaskTypeFileOperation: {"read", "write", "search"},
		TaskTypeShellCommand:  {"run_shell_command"},
		TaskTypeAnalysis:      {"read", "search", "git"},
		TaskTypeRefactoring:   {"edit", "read", "write", "search", "git"},
		TaskTypeDebugging:     {"read", "search", "run_shell_command", "git"},
		TaskTypeGeneral:       {"read", "write", "search", "edit"},
	}

	if tools, ok := toolMap[taskType]; ok {
		return tools
	}

	return []string{"read", "write", "search"}
}

// translatePrompt translates user intent to Devin-compatible prompt
func (a *Adapter) translatePrompt(userIntent string, taskType TaskType) string {
	taskPrefixes := map[TaskType]string{
		TaskTypeCodeEdit:      "Please edit the code to: ",
		TaskTypeFileOperation: "Please perform file operation: ",
		TaskTypeShellCommand:  "Please execute shell command: ",
		TaskTypeAnalysis:      "Please analyze: ",
		TaskTypeRefactoring:   "Please refactor: ",
		TaskTypeDebugging:     "Please debug: ",
		TaskTypeGeneral:       "Please help with: ",
	}

	prefix := taskPrefixes[taskType]
	if prefix == "" {
		prefix = "Please help with: "
	}

	return fmt.Sprintf("%s%s", prefix, userIntent)
}

// translateOptions translates environment variables to Devin options
func (a *Adapter) translateOptions(environment map[string]string) map[string]interface{} {
	options := make(map[string]interface{})

	envMapping := map[string]string{
		"EDITOR": "editor",
		"SHELL":  "shell",
		"LANG":   "language",
		"TZ":     "timezone",
	}

	for envKey, envValue := range environment {
		optionKey := envMapping[envKey]
		if optionKey == "" {
			optionKey = envKey
		}
		options[optionKey] = envValue
	}

	return options
}

// translateMetadata translates metadata to string format
func (a *Adapter) translateMetadata(metadata map[string]interface{}) map[string]string {
	stringMetadata := make(map[string]string)

	for key, value := range metadata {
		if strValue, ok := value.(string); ok {
			stringMetadata[key] = strValue
		} else {
			stringMetadata[key] = fmt.Sprintf("%v", value)
		}
	}

	return stringMetadata
}

// translateResult translates result string
func (a *Adapter) translateResult(result string) string {
	return result
}

// translateOutputMetadata translates output metadata
func (a *Adapter) translateOutputMetadata(metadata map[string]interface{}) map[string]interface{} {
	return metadata
}

// InjectSkillPrompt injects skill-specific prompt into original prompt
func (a *Adapter) InjectSkillPrompt(originalPrompt string, skill string) string {
	skillPrompts := map[string]string{
		"code_review":   "Please review the code with focus on:\n- Code quality\n- Performance\n- Security\n- Best practices",
		"refactoring":   "Please refactor the code to:\n- Improve readability\n- Reduce complexity\n- Follow SOLID principles\n- Maintain functionality",
		"debugging":     "Please debug the code by:\n- Identifying the root cause\n- Proposing fixes\n- Explaining the issue",
		"documentation": "Please add documentation:\n- Function descriptions\n- Parameter explanations\n- Return value documentation\n- Usage examples",
	}

	if skillPrompt, ok := skillPrompts[skill]; ok {
		return fmt.Sprintf("%s\n\n%s", originalPrompt, skillPrompt)
	}

	return originalPrompt
}
