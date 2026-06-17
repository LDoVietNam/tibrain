// Package permission implements a production-ready permission system for Ti.
// Inspired by Claude Code (5 modes), Codex (4 policies), OpenCode (pattern-based rules).
package permission

import (
	"fmt"
	"path/filepath"
	"strings"
)

// PermissionMode defines how Ti handles tool permission checks.
type PermissionMode string

const (
	// ModeAsk: Ask user for permission before each tool execution (default).
	ModeAsk PermissionMode = "ask"
	// ModeAuto: Auto-approve read-only tools, ask for write/destructive.
	ModeAuto PermissionMode = "auto"
	// ModeYes: Auto-approve all tool executions (dangerous!).
	ModeYes PermissionMode = "yes"
	// ModeDeny: Deny all tool executions (safe mode).
	ModeDeny PermissionMode = "deny"
	// ModePlan: Read-only mode for planning. Only allow writes to .plan files.
	ModePlan PermissionMode = "plan"
	// ModeSmart: AI decides based on risk assessment + user preferences.
	ModeSmart PermissionMode = "smart"
)

// ValidModes returns all valid permission modes.
func ValidModes() []PermissionMode {
	return []PermissionMode{ModeAsk, ModeAuto, ModeYes, ModeDeny, ModePlan, ModeSmart}
}

// IsValid checks if the mode is valid.
func (m PermissionMode) IsValid() bool {
	for _, valid := range ValidModes() {
		if m == valid {
			return true
		}
	}
	return false
}

// PermissionBehavior is the decision result for a tool execution.
type PermissionBehavior string

const (
	BehaviorAllow PermissionBehavior = "allow"
	BehaviorDeny  PermissionBehavior = "deny"
	BehaviorAsk   PermissionBehavior = "ask"
)

// PermissionRuleSource defines where a permission rule originated from.
type PermissionRuleSource string

const (
	SourceUserSettings    PermissionRuleSource = "userSettings"
	SourceProjectSettings PermissionRuleSource = "projectSettings"
	SourceLocalSettings   PermissionRuleSource = "localSettings"
	SourceCLIArg          PermissionRuleSource = "cliArg"
	SourceSession         PermissionRuleSource = "session"
)

// PermissionRuleValue represents the target of a permission rule.
type PermissionRuleValue struct {
	ToolName    string // Tool name (e.g., "bash", "file_write")
	Pattern     string // Glob pattern (e.g., "src/**/*.ts", "*.env")
	Description string // Human-readable description
}

// PermissionRule is a single rule with its source and behavior.
type PermissionRule struct {
	Source   PermissionRuleSource
	Behavior PermissionBehavior
	Value    PermissionRuleValue
}

// ToolPermissionContext holds the complete permission state.
type ToolPermissionContext struct {
	Mode                  PermissionMode
	AlwaysAllowRules      map[PermissionRuleSource][]string // Pattern strings
	AlwaysDenyRules       map[PermissionRuleSource][]string
	AlwaysAskRules        map[PermissionRuleSource][]string
	AdditionalWorkingDirs []string
	IsBypassAvailable     bool
	ShouldAvoidPrompts    bool
}

// NewPermissionContext creates a new permission context with defaults.
func NewPermissionContext(mode PermissionMode) ToolPermissionContext {
	if !mode.IsValid() {
		mode = ModeAsk // Default to ask mode
	}

	return ToolPermissionContext{
		Mode:             mode,
		AlwaysAllowRules: make(map[PermissionRuleSource][]string),
		AlwaysDenyRules:  make(map[PermissionRuleSource][]string),
		AlwaysAskRules:   make(map[PermissionRuleSource][]string),
	}
}

// SetYolo enables unrestricted tool execution mode.
func (ctx *ToolPermissionContext) SetYolo(enabled bool) {
	if enabled {
		ctx.Mode = ModeYes
	}
}

// PermissionDecision is the result of a permission check.
type PermissionDecision struct {
	Behavior    PermissionBehavior
	Reason      string
	Rule        *PermissionRule // Nil if decided by mode
	ToolName    string
	ToolInput   map[string]any
	Suggestions []PermissionSuggestion // Hints for user
}

// PermissionSuggestion provides actionable hints for user.
type PermissionSuggestion struct {
	Type        string // "allow_once", "allow_always", "deny_always", "add_rule"
	Description string
}

// CheckPermission checks if a tool execution should be allowed.
func CheckPermission(ctx ToolPermissionContext, toolName string, toolInput map[string]any, filePath string) PermissionDecision {
	decision := PermissionDecision{
		ToolName:  toolName,
		ToolInput: toolInput,
	}

	// Step 1: Check deny rules (highest priority)
	if rule, matched := checkRules(ctx.AlwaysDenyRules, toolName, filePath); matched {
		decision.Behavior = BehaviorDeny
		decision.Reason = fmt.Sprintf("Denied by rule from %s: %s", rule.Source, rule.Value.Description)
		decision.Rule = &rule
		return decision
	}

	// Step 2: Check ask rules (force prompt)
	if rule, matched := checkRules(ctx.AlwaysAskRules, toolName, filePath); matched {
		decision.Behavior = BehaviorAsk
		decision.Reason = fmt.Sprintf("Requires approval per rule from %s: %s", rule.Source, rule.Value.Description)
		decision.Rule = &rule
		decision.Suggestions = buildSuggestions(toolName, filePath)
		return decision
	}

	// Step 3: Check allow rules
	if _, matched := checkRules(ctx.AlwaysAllowRules, toolName, filePath); matched {
		decision.Behavior = BehaviorAllow
		decision.Reason = "Allowed by rule"
		return decision
	}

	// Step 4: Apply mode-based logic
	return checkByMode(ctx, toolName, filePath)
}

// checkRules checks if any rule matches the tool and file.
func checkRules(rulesBySource map[PermissionRuleSource][]string, toolName string, filePath string) (PermissionRule, bool) {
	for source, patterns := range rulesBySource {
		for _, pattern := range patterns {
			if matchesPermissionRule(pattern, toolName, filePath) {
				rule := PermissionRule{
					Source:   source,
					Behavior: BehaviorAllow, // Will be overridden by caller
					Value: PermissionRuleValue{
						ToolName: toolName,
						Pattern:  pattern,
					},
				}
				return rule, true
			}
		}
	}
	return PermissionRule{}, false
}

// matchesPermissionRule checks if a pattern matches the tool/file.
func matchesPermissionRule(pattern string, toolName string, filePath string) bool {
	// Format: "tool:pattern" or just "pattern"
	parts := strings.SplitN(pattern, ":", 2)
	if len(parts) == 2 {
		// Has tool specifier
		ruleTool := parts[0]
		rulePattern := parts[1]

		if ruleTool != "*" && ruleTool != toolName {
			return false
		}

		// Check pattern against file path
		if filePath == "" {
			return rulePattern == "*"
		}

		matched, _ := filepath.Match(rulePattern, filepath.Base(filePath))
		if matched {
			return true
		}

		// Try full path match
		matched, _ = filepath.Match(rulePattern, filePath)
		return matched
	}

	// No tool specifier, matches all tools
	p := parts[0]
	if p == "*" {
		return true
	}

	if filePath == "" {
		return false
	}

	matched, _ := filepath.Match(p, filepath.Base(filePath))
	if matched {
		return true
	}

	matched, _ = filepath.Match(p, filePath)
	return matched
}

// checkByMode applies mode-specific permission logic.
func checkByMode(ctx ToolPermissionContext, toolName string, filePath string) PermissionDecision {
	decision := PermissionDecision{
		ToolName:  toolName,
		ToolInput: map[string]any{"mode": string(ctx.Mode), "file_path": filePath},
	}

	switch ctx.Mode {
	case ModeYes:
		decision.Behavior = BehaviorAllow
		decision.Reason = "Auto-approved by 'yes' mode"

	case ModeDeny:
		decision.Behavior = BehaviorDeny
		decision.Reason = "Denied by 'deny mode'"

	case ModePlan:
		if isReadOnlyTool(toolName) {
			decision.Behavior = BehaviorAllow
			decision.Reason = "Auto-approved: Read-only tool in Plan Mode"
		} else {
			// Allow writes only to plan files
			if filePath != "" && (strings.HasSuffix(filePath, ".plan") || strings.Contains(filePath, "current_plan")) {
				decision.Behavior = BehaviorAllow
				decision.Reason = "Auto-approved: Write to plan file in Plan Mode"
			} else {
				decision.Behavior = BehaviorDeny
				decision.Reason = "Denied: Write tools are restricted in Plan Mode (except for .plan files)"
			}
		}

	case ModeAsk:
		decision.Behavior = BehaviorAsk
		decision.Reason = "Requires approval in 'ask' mode"
		decision.Suggestions = buildSuggestions(toolName, filePath)

	case ModeAuto:
		// Auto-approve read-only tools, ask for write/destructive
		if isReadOnlyTool(toolName) {
			decision.Behavior = BehaviorAllow
			decision.Reason = fmt.Sprintf("Auto-approved: %s is read-only", toolName)
		} else if isSafeWriteTool(toolName) {
			decision.Behavior = BehaviorAllow
			decision.Reason = fmt.Sprintf("Auto-approved: %s is safe", toolName)
		} else {
			decision.Behavior = BehaviorAsk
			decision.Reason = fmt.Sprintf("Requires approval: %s may modify system", toolName)
			decision.Suggestions = buildSuggestions(toolName, filePath)
		}

	case ModeSmart:
		// AI decides based on risk assessment
		risk := assessToolRisk(toolName, filePath)
		switch risk {
		case "low":
			decision.Behavior = BehaviorAllow
			decision.Reason = fmt.Sprintf("Smart mode: %s is low risk", toolName)
		case "medium":
			decision.Behavior = BehaviorAsk
			decision.Reason = fmt.Sprintf("Smart mode: %s is medium risk", toolName)
			decision.Suggestions = buildSuggestions(toolName, filePath)
		case "high":
			decision.Behavior = BehaviorAsk
			decision.Reason = fmt.Sprintf("Smart mode: %s is HIGH RISK - review carefully", toolName)
			decision.Suggestions = buildSuggestions(toolName, filePath)
		}

	default:
		decision.Behavior = BehaviorAsk
		decision.Reason = fmt.Sprintf("Unknown mode '%s', defaulting to ask", ctx.Mode)
	}

	return decision
}

// isReadOnlyTool checks if a tool is read-only.
func isReadOnlyTool(toolName string) bool {
	readOnlyTools := map[string]bool{
		"bash":       false, // Depends on command
		"file_read":  true,
		"grep":       true,
		"glob":       true,
		"web_fetch":  true,
		"web_search": true,
		"agent_read": true,
		"mcp_read":   true,
	}
	return readOnlyTools[toolName]
}

// isSafeWriteTool checks if a tool is safe to auto-approve.
func isSafeWriteTool(toolName string) bool {
	safeTools := map[string]bool{
		"file_write":  false, // Write can be destructive
		"file_edit":   false,
		"todo_write":  true,
		"task_create": true,
	}
	return safeTools[toolName]
}

// assessToolRisk assesses the risk level of a tool execution.
func assessToolRisk(toolName string, filePath string) string {
	// High risk tools
	highRiskTools := map[string]bool{
		"bash":          true, // Can run any command
		"file_write":    isDangerousPath(filePath),
		"file_edit":     isDangerousPath(filePath),
		"agent_execute": false,
	}

	if isHighRisk, ok := highRiskTools[toolName]; ok {
		if isHighRisk {
			return "high"
		}
		return "medium"
	}

	// Medium risk tools
	mediumRiskTools := map[string]bool{
		"web_fetch":   false,
		"mcp_execute": false,
	}

	if _, ok := mediumRiskTools[toolName]; ok {
		return "medium"
	}

	// Low risk by default
	return "low"
}

// isDangerousPath checks if a file path is dangerous to modify.
func isDangerousPath(filePath string) bool {
	dangerousPatterns := []string{
		"*.env", "*.key", "*.pem", "*.crt",
		".git/*", ".ssh/*", "*/secrets/*",
		"*/passwords/*", "*/credentials/*",
	}

	for _, pattern := range dangerousPatterns {
		matched, _ := filepath.Match(pattern, filePath)
		if matched {
			return true
		}
	}

	return false
}

// buildSuggestions builds permission suggestions for the user.
func buildSuggestions(toolName string, filePath string) []PermissionSuggestion {
	suggestions := []PermissionSuggestion{
		{
			Type:        "allow_once",
			Description: fmt.Sprintf("Allow %s to run once", toolName),
		},
		{
			Type:        "allow_always",
			Description: fmt.Sprintf("Always allow %s for this pattern", toolName),
		},
		{
			Type:        "deny_always",
			Description: fmt.Sprintf("Always deny %s", toolName),
		},
	}

	if filePath != "" {
		suggestions = append(suggestions, PermissionSuggestion{
			Type:        "add_rule",
			Description: fmt.Sprintf("Add rule for %s on %s", toolName, filePath),
		})
	}

	return suggestions
}

// AddRule adds a permission rule to the context.
func (ctx *ToolPermissionContext) AddRule(source PermissionRuleSource, behavior PermissionBehavior, toolName string, pattern string) {
	ruleStr := toolName + ":" + pattern

	switch behavior {
	case BehaviorAllow:
		ctx.AlwaysAllowRules[source] = append(ctx.AlwaysAllowRules[source], ruleStr)
	case BehaviorDeny:
		ctx.AlwaysDenyRules[source] = append(ctx.AlwaysDenyRules[source], ruleStr)
	case BehaviorAsk:
		ctx.AlwaysAskRules[source] = append(ctx.AlwaysAskRules[source], ruleStr)
	}
}

// AddDirectory adds an additional working directory.
func (ctx *ToolPermissionContext) AddDirectory(path string) {
	for _, existing := range ctx.AdditionalWorkingDirs {
		if existing == path {
			return // Already added
		}
	}
	ctx.AdditionalWorkingDirs = append(ctx.AdditionalWorkingDirs, path)
}
