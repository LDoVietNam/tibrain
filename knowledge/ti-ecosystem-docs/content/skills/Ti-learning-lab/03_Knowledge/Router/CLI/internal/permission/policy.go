package permission

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Policy struct {
	Mode      PermissionMode `json:"mode"`
	Allow     []string       `json:"allow,omitempty"`
	Deny      []string       `json:"deny,omitempty"`
	Ask       []string       `json:"ask,omitempty"`
	Workspace string         `json:"workspace,omitempty"`
	UpdatedAt time.Time      `json:"updated_at,omitempty"`
}

func DefaultPolicy() *Policy {
	return &Policy{Mode: ModeAsk, Deny: append([]string(nil), DefaultDenyPatterns...), UpdatedAt: time.Now()}
}

func LoadPolicy(path string) (*Policy, error) {
	policy := DefaultPolicy()
	if path == "" {
		return policy, nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return policy, nil
		}
		return nil, err
	}
	if err := json.Unmarshal(data, policy); err != nil {
		return nil, err
	}
	if !policy.Mode.IsValid() {
		policy.Mode = ModeAsk
	}
	policy.Allow = dedupe(policy.Allow)
	policy.Deny = dedupe(append(policy.Deny, DefaultDenyPatterns...))
	policy.Ask = dedupe(policy.Ask)
	return policy, nil
}

func (p *Policy) Save(path string) error {
	if p == nil {
		return fmt.Errorf("nil policy")
	}
	if path == "" {
		return fmt.Errorf("empty path")
	}
	p.UpdatedAt = time.Now()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o600)
}

func (p *Policy) Context() ToolPermissionContext {
	if p == nil {
		p = DefaultPolicy()
	}
	ctx := NewPermissionContext(p.Mode)
	for _, rule := range p.Allow {
		ctx.AlwaysAllowRules[SourceProjectSettings] = append(ctx.AlwaysAllowRules[SourceProjectSettings], rule)
	}
	for _, rule := range p.Deny {
		ctx.AlwaysDenyRules[SourceProjectSettings] = append(ctx.AlwaysDenyRules[SourceProjectSettings], rule)
	}
	for _, rule := range p.Ask {
		ctx.AlwaysAskRules[SourceProjectSettings] = append(ctx.AlwaysAskRules[SourceProjectSettings], rule)
	}
	if p.Workspace != "" {
		ctx.AddDirectory(p.Workspace)
	}
	return ctx
}

func (p *Policy) Evaluate(toolName, filePath, command string) PermissionDecision {
	ctx := p.Context()
	input := map[string]any{"command": command, "path": filePath}
	decision := CheckPermission(ctx, toolName, input, filePath)
	if toolName == "bash" && command != "" && looksDestructive(command) {
		decision.Behavior = BehaviorAsk
		if p != nil && p.Mode == ModeDeny {
			decision.Behavior = BehaviorDeny
		}
		decision.Reason = "Command looks destructive or privilege-elevating"
	}
	return decision
}

func (p *Policy) AddRule(behavior PermissionBehavior, rule string) {
	if p == nil || strings.TrimSpace(rule) == "" {
		return
	}
	switch behavior {
	case BehaviorAllow:
		p.Allow = dedupe(append(p.Allow, rule))
	case BehaviorDeny:
		p.Deny = dedupe(append(p.Deny, rule))
	case BehaviorAsk:
		p.Ask = dedupe(append(p.Ask, rule))
	}
}

func looksDestructive(cmd string) bool {
	lower := strings.ToLower(strings.TrimSpace(cmd))
	for _, pat := range DestructiveCommands {
		token := strings.ToLower(strings.ReplaceAll(strings.ReplaceAll(pat, ".*", ""), "|", " "))
		for _, part := range strings.Fields(token) {
			if part != "" && strings.Contains(lower, part) {
				return true
			}
		}
	}
	return strings.Contains(lower, "sudo ") || strings.Contains(lower, " rm ") || strings.HasPrefix(lower, "rm ")
}

func dedupe(values []string) []string {
	seen := map[string]bool{}
	out := make([]string, 0, len(values))
	for _, v := range values {
		v = strings.TrimSpace(v)
		if v == "" || seen[v] {
			continue
		}
		seen[v] = true
		out = append(out, v)
	}
	return out
}
