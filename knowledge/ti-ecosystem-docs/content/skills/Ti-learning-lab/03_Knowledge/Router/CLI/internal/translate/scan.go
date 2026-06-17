package translate

import (
	"io/fs"
	"path/filepath"
	"sort"
	"strings"
)

var skipDirs = map[string]bool{
	".git": true, ".hg": true, ".svn": true,
	"node_modules": true, "vendor": true, "dist": true, "build": true,
	"bin": true, "target": true, ".next": true, ".turbo": true,
	"coverage": true, ".cache": true, ".venv": true, "__pycache__": true,
}

func Scan(root string, opts ScanOptions) ([]Finding, error) {
	if root == "" {
		root = "."
	}
	root = filepath.Clean(root)
	if opts.MaxDepth <= 0 {
		opts.MaxDepth = 10
	}
	var out []Finding
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		base := d.Name()
		if d.IsDir() {
			if path != root && skipDirs[base] {
				return filepath.SkipDir
			}
			if depth(root, path) > opts.MaxDepth {
				return filepath.SkipDir
			}
			return nil
		}
		if d.Type()&fs.ModeSymlink != 0 {
			return nil
		}
		f, ok := Detect(root, path)
		if !ok {
			return nil
		}
		if opts.From != "" && opts.From != FormatAuto && f.Format != opts.From {
			return nil
		}
		out = append(out, f)
		return nil
	})
	sort.SliceStable(out, func(i, j int) bool { return out[i].Path < out[j].Path })
	return out, err
}

func Detect(root, path string) (Finding, bool) {
	rel, err := filepath.Rel(root, path)
	if err != nil {
		rel = path
	}
	relSlash := filepath.ToSlash(rel)
	base := filepath.Base(path)
	lowRel := strings.ToLower(relSlash)
	lowBase := strings.ToLower(base)
	name := strings.TrimSuffix(base, filepath.Ext(base))
	f := Finding{Path: relSlash, Name: slug(name), Confidence: 0.85}

	switch {
	case base == "AGENTS.md" || base == "AGENTS.override.md":
		f.Format, f.Kind, f.Name = FormatCodex, "instruction", strings.TrimSuffix(base, ".md")
		f.Notes = []string{"Codex/OpenAI-style project instruction file"}
		return f, true
	case base == "CLAUDE.md":
		f.Format, f.Kind, f.Name = FormatClaudeCode, "instruction", "claude"
		f.Notes = []string{"Claude Code project instruction"}
		return f, true
	case strings.HasPrefix(relSlash, ".claude/commands/") && strings.HasSuffix(lowRel, ".md"):
		f.Format, f.Kind = FormatClaudeCode, "command"
		f.Notes = []string{"Claude Code slash command markdown"}
		return f, true
	case strings.HasPrefix(relSlash, ".claude/agents/") && strings.HasSuffix(lowRel, ".md"):
		f.Format, f.Kind = FormatClaudeCode, "agent"
		f.Notes = []string{"Claude Code agent markdown"}
		return f, true
	case relSlash == ".claude/settings.json" || relSlash == ".claude/settings.local.json":
		f.Format, f.Kind, f.Name = FormatClaudeCode, "settings", strings.TrimSuffix(base, ".json")
		f.Risks = []string{"settings may contain hooks, permissions, env, or tool policies; review before apply"}
		return f, true
	case strings.HasPrefix(relSlash, ".claude/hooks/") && (strings.HasSuffix(lowRel, ".json") || strings.HasSuffix(lowRel, ".jsonc")):
		f.Format, f.Kind = FormatClaudeCode, "hooks"
		f.Risks = []string{"hooks run commands; sandbox required"}
		return f, true

	case strings.HasPrefix(relSlash, ".cursor/rules/") && (strings.HasSuffix(lowRel, ".mdc") || strings.HasSuffix(lowRel, ".md")):
		f.Format, f.Kind = FormatCursor, "rule"
		return f, true
	case lowBase == ".cursorrules":
		f.Format, f.Kind, f.Name = FormatCursor, "rule", "cursorrules"
		return f, true
	case strings.HasPrefix(relSlash, ".windsurf/rules/") && (strings.HasSuffix(lowRel, ".md") || strings.HasSuffix(lowRel, ".mdc")):
		f.Format, f.Kind = FormatWindsurf, "rule"
		return f, true
	case lowBase == ".windsurfrules":
		f.Format, f.Kind, f.Name = FormatWindsurf, "rule", "windsurfrules"
		return f, true

	case strings.HasPrefix(relSlash, ".opencode/commands/") && strings.HasSuffix(lowRel, ".md"):
		f.Format, f.Kind = FormatOpenCode, "command"
		return f, true
	case strings.HasPrefix(relSlash, ".opencode/agents/") && strings.HasSuffix(lowRel, ".md"):
		f.Format, f.Kind = FormatOpenCode, "agent"
		return f, true
	case strings.HasPrefix(relSlash, ".opencode/skills/") && strings.HasSuffix(lowRel, ".md"):
		f.Format, f.Kind = FormatOpenCode, "skill"
		return f, true
	case strings.HasPrefix(relSlash, ".opencode/plugins/") && (strings.HasSuffix(lowRel, ".ts") || strings.HasSuffix(lowRel, ".js") || strings.HasSuffix(lowRel, ".mjs")):
		f.Format, f.Kind = FormatOpenCode, "plugin"
		f.Risks = []string{"external JS/TS plugin; run only through sandbox/bridge"}
		return f, true
	case lowRel == "opencode.json" || lowRel == "opencode.jsonc" || lowRel == ".opencode.json":
		f.Format, f.Kind, f.Name = FormatOpenCode, "config", "opencode"
		f.Risks = []string{"config may define providers, plugins, commands, or permissions"}
		return f, true

	case relSlash == "mcp.json" || relSlash == ".mcp.json" || strings.HasSuffix(lowRel, "/mcp.json"):
		f.Format, f.Kind, f.Name = FormatMCP, "mcp_server_config", "mcp"
		f.Risks = []string{"MCP servers run external processes or remote tools; sandbox required"}
		return f, true
	case lowBase == ".aider.conf.yml" || lowBase == ".aider.conf.yaml" || lowBase == "aider.conf.yml" || lowBase == "aider.conf.yaml":
		f.Format, f.Kind, f.Name = FormatAider, "config", "aider"
		return f, true
	case lowBase == "conventions.md" || lowBase == ".aider.conventions.md":
		f.Format, f.Kind, f.Name = FormatAider, "instruction", strings.TrimSuffix(base, filepath.Ext(base))
		return f, true
	case strings.HasPrefix(relSlash, ".roo/") && (strings.HasSuffix(lowRel, ".md") || strings.HasSuffix(lowRel, ".json") || strings.HasSuffix(lowRel, ".yaml") || strings.HasSuffix(lowRel, ".yml")):
		f.Format, f.Kind = FormatRoo, inferRooKind(lowRel)
		return f, true
	case strings.Contains(lowRel, "devin") && (strings.HasSuffix(lowRel, ".yaml") || strings.HasSuffix(lowRel, ".yml") || strings.HasSuffix(lowRel, ".json")):
		f.Format, f.Kind = FormatDevin, "workflow"
		f.Risks = []string{"Devin-like workflow may reference remote sessions/secrets; review before apply"}
		return f, true
	case strings.Contains(lowRel, "kilo") && (strings.HasSuffix(lowRel, ".json") || strings.HasSuffix(lowRel, ".yaml") || strings.HasSuffix(lowRel, ".yml")):
		f.Format, f.Kind = FormatKilo, "config"
		f.Risks = []string{"Kilo-like config may reference providers/plugins; review before apply"}
		return f, true
	}
	return Finding{}, false
}

func inferRooKind(lowRel string) string {
	if strings.Contains(lowRel, "mode") {
		return "agent"
	}
	if strings.Contains(lowRel, "command") || strings.Contains(lowRel, "prompt") {
		return "command"
	}
	return "instruction"
}

func depth(root, path string) int {
	rel, err := filepath.Rel(root, path)
	if err != nil || rel == "." {
		return 0
	}
	return len(strings.Split(filepath.ToSlash(rel), "/"))
}
