package convert

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// IR is Ti's high-signal intermediate representation for code generation.
// It is intentionally richer than Pack so the brain/memory layer can learn
// stable patterns from conversions rather than brittle raw files.
type IR struct {
	Version     string         `json:"version"`
	ID          string         `json:"id"`
	Name        string         `json:"name"`
	Kind        string         `json:"kind"`
	Description string         `json:"description,omitempty"`
	Source      SourceMeta     `json:"source"`
	GeneratedAt time.Time      `json:"generated_at"`
	Commands    []IRCommand    `json:"commands,omitempty"`
	Tools       []IRTool       `json:"tools,omitempty"`
	Workflows   []IRWorkflow   `json:"workflows,omitempty"`
	Agents      []IRAgent      `json:"agents,omitempty"`
	Skills      []IRSkill      `json:"skills,omitempty"`
	Providers   []IRProvider   `json:"providers,omitempty"`
	Policies    []IRPolicy     `json:"policies,omitempty"`
	Files       []IRFile       `json:"files,omitempty"`
	Learning    LearningMeta   `json:"learning"`
	Meta        map[string]any `json:"meta,omitempty"`
}

type SourceMeta struct {
	Format      string `json:"format"`
	Fingerprint string `json:"fingerprint"`
	Bytes       int    `json:"bytes"`
}

type IRCommand struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Aliases     []string          `json:"aliases,omitempty"`
	Description string            `json:"description,omitempty"`
	Input       string            `json:"input,omitempty"`
	Policy      string            `json:"policy,omitempty"`
	Meta        map[string]string `json:"meta,omitempty"`
}

type IRTool struct {
	ID           string            `json:"id"`
	Name         string            `json:"name"`
	Description  string            `json:"description,omitempty"`
	Type         string            `json:"type,omitempty"`
	Command      string            `json:"command,omitempty"`
	Endpoint     string            `json:"endpoint,omitempty"`
	Capabilities []string          `json:"capabilities,omitempty"`
	Policy       string            `json:"policy,omitempty"`
	Meta         map[string]string `json:"meta,omitempty"`
}

type IRWorkflow struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description,omitempty"`
	Steps       []string `json:"steps,omitempty"`
	Command     string   `json:"command,omitempty"`
	Triggers    []string `json:"triggers,omitempty"`
	Policy      string   `json:"policy,omitempty"`
}

type IRAgent struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	Model        string   `json:"model,omitempty"`
	Role         string   `json:"role,omitempty"`
	Instructions string   `json:"instructions,omitempty"`
	Tools        []string `json:"tools,omitempty"`
	ModelRole    string   `json:"model_role,omitempty"`
}

type IRSkill struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description,omitempty"`
	Content     string   `json:"content,omitempty"`
	Tags        []string `json:"tags,omitempty"`
}

type IRProvider struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Endpoint string `json:"endpoint,omitempty"`
	Model    string `json:"model,omitempty"`
	Role     string `json:"role,omitempty"`
}

type IRPolicy struct {
	Name            string `json:"name"`
	AllowWrite      bool   `json:"allow_write,omitempty"`
	AllowShell      bool   `json:"allow_shell,omitempty"`
	AllowNetwork    bool   `json:"allow_network,omitempty"`
	AllowSecrets    bool   `json:"allow_secrets,omitempty"`
	RequireSnapshot bool   `json:"require_snapshot,omitempty"`
	RequireConfirm  bool   `json:"require_confirm,omitempty"`
	Reason          string `json:"reason,omitempty"`
}

type IRFile struct {
	Path    string `json:"path"`
	Kind    string `json:"kind,omitempty"`
	Content string `json:"content,omitempty"`
}

type LearningMeta struct {
	ComplexityScore  int      `json:"complexity_score"`
	RiskScore        int      `json:"risk_score"`
	RecommendedKind  string   `json:"recommended_kind"`
	RecommendedModel string   `json:"recommended_model"`
	Warnings         []string `json:"warnings,omitempty"`
	PatternKeys      []string `json:"pattern_keys,omitempty"`
}

// BuildIR converts any supported source document into Ti IR.
func BuildIR(data []byte, opts Options) (*IR, *Pack, error) {
	pack, err := Convert(data, opts)
	if err != nil {
		return nil, nil, err
	}
	ir := PackToIR(pack, data, opts)
	return ir, pack, nil
}

func PackToIR(pack *Pack, data []byte, opts Options) *IR {
	name := "converted-pack"
	if len(pack.Workflows) > 0 && strings.TrimSpace(pack.Workflows[0].Name) != "" {
		name = pack.Workflows[0].Name
	} else if len(pack.Plugins) > 0 && strings.TrimSpace(pack.Plugins[0].Name) != "" {
		name = pack.Plugins[0].Name
	} else if len(pack.Skills) > 0 && strings.TrimSpace(pack.Skills[0].Name) != "" {
		name = pack.Skills[0].Name
	} else if len(pack.Agents) > 0 && strings.TrimSpace(pack.Agents[0].Name) != "" {
		name = pack.Agents[0].Name
	}
	format := normalizeName(opts.From)
	if format == "" || format == "auto" || format == "item" {
		format = detectSourceFormat(data)
	}
	ir := &IR{
		Version:     "ti-ir/v1",
		ID:          stableID(format, "ir", name),
		Name:        normalizeDisplayName(name),
		Kind:        inferIRKind(pack),
		Source:      SourceMeta{Format: format, Fingerprint: shortFingerprint(data), Bytes: len(data)},
		GeneratedAt: time.Now().UTC(),
		Meta:        map[string]any{"pack_version": pack.Version, "target": normalizeName(opts.To)},
	}
	for _, s := range pack.Skills {
		ir.Skills = append(ir.Skills, IRSkill{ID: s.ID, Name: s.Name, Description: s.Description, Content: s.Content, Tags: s.Tags})
	}
	for _, w := range pack.Workflows {
		pol := inferPolicyForText(strings.Join(append([]string{w.Command}, w.Steps...), "\n"))
		ir.Workflows = append(ir.Workflows, IRWorkflow{ID: w.ID, Name: safeName(w.Name), Description: w.Description, Steps: w.Steps, Command: w.Command, Triggers: w.Triggers, Policy: pol.Name})
		ir.Commands = append(ir.Commands, IRCommand{ID: stableID("command", w.ID), Name: safeName(w.Name), Description: firstNonEmpty(w.Description, "Generated from workflow"), Input: firstNonEmpty(w.Command, strings.Join(w.Steps, "\n")), Policy: pol.Name})
		ir.Policies = appendPolicy(ir.Policies, pol)
	}
	for _, a := range pack.Agents {
		ir.Agents = append(ir.Agents, IRAgent{ID: a.ID, Name: a.Name, Model: a.Model, Role: a.Role, Instructions: a.Instructions, Tools: a.Tools, ModelRole: inferModelRole(a)})
	}
	for _, p := range pack.Plugins {
		pol := inferPolicyForText(p.Command + "\n" + p.Endpoint + "\n" + strings.Join(p.Capabilities, " "))
		ir.Tools = append(ir.Tools, IRTool{ID: p.ID, Name: safeName(p.Name), Type: p.Type, Command: p.Command, Endpoint: p.Endpoint, Capabilities: p.Capabilities, Policy: pol.Name})
		if p.Endpoint != "" {
			ir.Providers = append(ir.Providers, IRProvider{ID: stableID("provider", p.Name), Name: safeName(p.Name), Endpoint: p.Endpoint, Role: "tool"})
		}
		ir.Policies = appendPolicy(ir.Policies, pol)
	}
	if len(ir.Policies) == 0 {
		ir.Policies = append(ir.Policies, IRPolicy{Name: "read-only", Reason: "default safe policy"})
	}
	ir.Learning = analyzeLearning(ir)
	return ir
}

func RenderIR(ir *IR) ([]byte, error) {
	if ir == nil {
		return nil, fmt.Errorf("nil IR")
	}
	return json.MarshalIndent(ir, "", "  ")
}

func detectSourceFormat(data []byte) string {
	s := strings.ToLower(string(data))
	switch {
	case strings.Contains(s, "mcpservers") || strings.Contains(s, "claude_desktop_config") || strings.Contains(s, "claude"):
		return "claude-code"
	case strings.Contains(s, "opencode") || strings.Contains(s, "opencode.json"):
		return "opencode"
	case strings.Contains(s, "cursor") || strings.Contains(s, ".cursorrules"):
		return "cursor"
	case strings.Contains(s, "ampcode") || strings.Contains(s, "amp"):
		return "ampcode"
	case strings.Contains(s, "codex"):
		return "codex"
	case strings.Contains(s, "kilo"):
		return "kilo"
	case strings.Contains(s, "workflow") || strings.Contains(s, "workflows"):
		return "workflow"
	case strings.Contains(s, "plugin") || strings.Contains(s, "plugins"):
		return "plugin"
	default:
		return "auto"
	}
}

func inferIRKind(pack *Pack) string {
	switch {
	case len(pack.Plugins) > 0 && len(pack.Workflows) == 0:
		return "tool-plugin"
	case len(pack.Workflows) > 0 && len(pack.Plugins) == 0:
		return "workflow-plugin"
	case len(pack.Agents) > 0 && len(pack.Workflows)+len(pack.Plugins) == 0:
		return "agent-plugin"
	case len(pack.Skills) > 0 && len(pack.Workflows)+len(pack.Plugins)+len(pack.Agents) == 0:
		return "skill-plugin"
	default:
		return "composite-plugin"
	}
}

func inferPolicyForText(text string) IRPolicy {
	q := strings.ToLower(text)
	p := IRPolicy{Name: "read-only", Reason: "no write/shell/network markers detected"}
	if containsAny(q, "rm ", "del ", "remove-item", "writefile", "write file", "apply patch", "git apply", "mkdir", "new-item", "touch ", "copy ", "move ") {
		p.Name = "write-confirmed"
		p.AllowWrite = true
		p.RequireSnapshot = true
		p.RequireConfirm = true
		p.Reason = "write-like action detected"
	}
	if containsAny(q, "sh ", "bash", "powershell", "cmd.exe", "exec", "shell", "subprocess") {
		p.Name = "shell-confirmed"
		p.AllowShell = true
		p.RequireSnapshot = true
		p.RequireConfirm = true
		p.Reason = "shell execution detected"
	}
	if containsAny(q, "http://", "https://", "api", "endpoint", "openai", "anthropic", "openrouter", "tunnel", "mcp") {
		p.AllowNetwork = true
		if p.Name == "read-only" {
			p.Name = "network-read"
			p.Reason = "network endpoint detected"
		}
	}
	if containsAny(q, "cookie", "token", "secret", "api_key", "apikey", "authorization") {
		p.AllowSecrets = false
		p.RequireConfirm = true
		p.Reason = strings.TrimSpace(p.Reason + "; secret-like term detected")
	}
	return p
}

func appendPolicy(policies []IRPolicy, p IRPolicy) []IRPolicy {
	for _, existing := range policies {
		if existing.Name == p.Name {
			return policies
		}
	}
	return append(policies, p)
}

func inferModelRole(a Agent) string {
	q := strings.ToLower(a.Name + " " + a.Role + " " + a.Instructions + " " + strings.Join(a.Tools, " "))
	switch {
	case containsAny(q, "security", "auth", "cookie", "secret", "policy"):
		return "security"
	case containsAny(q, "review", "critic", "audit"):
		return "reviewer"
	case containsAny(q, "code", "coder", "fix", "patch"):
		return "coder"
	case containsAny(q, "context", "long", "planner", "architect"):
		return "long_context"
	default:
		return "scout"
	}
}

func analyzeLearning(ir *IR) LearningMeta {
	complexity := len(ir.Commands)*2 + len(ir.Tools)*2 + len(ir.Workflows)*3 + len(ir.Agents) + len(ir.Skills)
	risk := 0
	keys := []string{ir.Kind, ir.Source.Format}
	var warnings []string
	for _, p := range ir.Policies {
		if p.AllowWrite {
			risk += 3
			keys = append(keys, "writes")
		}
		if p.AllowShell {
			risk += 4
			keys = append(keys, "shell")
		}
		if p.AllowNetwork {
			risk += 2
			keys = append(keys, "network")
		}
		if p.RequireConfirm {
			risk += 1
		}
	}
	for _, t := range ir.Tools {
		q := strings.ToLower(t.Name + " " + t.Endpoint + " " + t.Command)
		if containsAny(q, "cookie", "token", "secret") {
			risk += 5
			keys = append(keys, "secrets")
			warnings = append(warnings, "secret-like tool/provider detected; generated policy should redact context and require confirmation")
		}
		if strings.Contains(q, "mcp") || strings.Contains(q, "tunnel") {
			keys = append(keys, "mcp")
		}
	}
	model := "scout"
	if complexity > 12 {
		model = "long_context"
	}
	if risk >= 7 {
		model = "security"
	}
	kind := ir.Kind
	if len(ir.Tools) > 0 && len(ir.Workflows) > 0 {
		kind = "composite-plugin"
	}
	return LearningMeta{ComplexityScore: complexity, RiskScore: risk, RecommendedKind: kind, RecommendedModel: model, Warnings: warnings, PatternKeys: uniqueStringsLocal(keys)}
}

func shortFingerprint(data []byte) string {
	return stableID("fingerprint", string(data))
}

func containsAny(s string, subs ...string) bool {
	for _, sub := range subs {
		if strings.Contains(s, sub) {
			return true
		}
	}
	return false
}

func normalizeDisplayName(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return "converted-pack"
	}
	return s
}

func safeName(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return "generated"
	}
	return s
}

func uniqueStringsLocal(in []string) []string {
	seen := map[string]bool{}
	out := make([]string, 0, len(in))
	for _, item := range in {
		item = strings.TrimSpace(item)
		if item == "" || seen[item] {
			continue
		}
		seen[item] = true
		out = append(out, item)
	}
	return out
}
