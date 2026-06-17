package translate

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type WriteOptions struct {
	OutDir string
	DryRun bool
}

type ExportOptions struct {
	To     string
	OutDir string
	DryRun bool
}

func WritePack(pack *Pack, lock LockFile, opts WriteOptions) (string, error) {
	if opts.OutDir == "" {
		opts.OutDir = filepath.Join(".ti", "packs", pack.Metadata.Name)
	}
	yaml := RenderYAML(pack)
	if opts.DryRun {
		return yaml, nil
	}
	if err := os.MkdirAll(opts.OutDir, 0o755); err != nil {
		return "", err
	}
	packPath := filepath.Join(opts.OutDir, "ti.pack.yaml")
	if err := os.WriteFile(packPath, []byte(yaml), 0o644); err != nil {
		return "", err
	}
	lockPath := filepath.Join(opts.OutDir, "compat.lock.json")
	data, err := json.MarshalIndent(lock, "", "  ")
	if err != nil {
		return "", err
	}
	data = append(data, '\n')
	if err := os.WriteFile(lockPath, data, 0o644); err != nil {
		return "", err
	}
	return packPath, nil
}

func RenderYAML(p *Pack) string {
	var b strings.Builder
	w := func(format string, args ...any) { fmt.Fprintf(&b, format, args...) }
	w("apiVersion: %s\n", p.APIVersion)
	w("kind: %s\n", p.Kind)
	w("metadata:\n")
	w("  name: %s\n", quote(p.Metadata.Name))
	if p.Metadata.Description != "" {
		w("  description: %s\n", quote(p.Metadata.Description))
	}
	w("  generatedAt: %s\n", quote(p.Metadata.GeneratedAt.Format("2006-01-02T15:04:05Z07:00")))
	w("  source: %s\n", quote(p.Metadata.Source))
	if p.Metadata.Version != "" {
		w("  version: %s\n", quote(p.Metadata.Version))
	}
	if len(p.Instructions) > 0 {
		w("instructions:\n")
		for _, x := range p.Instructions {
			w("  - id: %s\n    name: %s\n    scope: %s\n", quote(x.ID), quote(x.Name), quote(x.Scope))
			if len(x.Globs) > 0 {
				renderList(&b, "globs", x.Globs, 4)
			}
			renderSource(&b, x.Source, 4)
			renderStringMap(&b, "meta", x.Meta, 4)
			renderBlock(&b, "content", x.Content, 4)
		}
	}
	if len(p.Commands) > 0 {
		w("commands:\n")
		for _, x := range p.Commands {
			w("  - id: %s\n    name: %s\n", quote(x.ID), quote(x.Name))
			if x.Description != "" {
				w("    description: %s\n", quote(x.Description))
			}
			if len(x.Args) > 0 {
				w("    args:\n")
				for _, a := range x.Args {
					w("      - name: %s\n", quote(a.Name))
					if a.Default != "" {
						w("        default: %s\n", quote(a.Default))
					}
					if a.Description != "" {
						w("        description: %s\n", quote(a.Description))
					}
					if a.Required {
						w("        required: true\n")
					}
				}
			}
			if len(x.Tools) > 0 {
				renderList(&b, "tools", x.Tools, 4)
			}
			if x.Sandbox {
				w("    sandbox: true\n")
			}
			renderSource(&b, x.Source, 4)
			renderStringMap(&b, "metadata", x.Metadata, 4)
			renderBlock(&b, "prompt", x.Prompt, 4)
		}
	}
	if len(p.Agents) > 0 {
		w("agents:\n")
		for _, x := range p.Agents {
			w("  - id: %s\n    name: %s\n", quote(x.ID), quote(x.Name))
			if x.Description != "" {
				w("    description: %s\n", quote(x.Description))
			}
			if x.Model != "" {
				w("    model: %s\n", quote(x.Model))
			}
			if x.Mode != "" {
				w("    mode: %s\n", quote(x.Mode))
			}
			if len(x.Tools) > 0 {
				renderList(&b, "tools", x.Tools, 4)
			}
			if x.Sandbox {
				w("    sandbox: true\n")
			}
			renderSource(&b, x.Source, 4)
			renderStringMap(&b, "metadata", x.Metadata, 4)
			renderBlock(&b, "prompt", x.Prompt, 4)
		}
	}
	if len(p.Skills) > 0 {
		w("skills:\n")
		for _, x := range p.Skills {
			w("  - id: %s\n    name: %s\n", quote(x.ID), quote(x.Name))
			if x.Description != "" {
				w("    description: %s\n", quote(x.Description))
			}
			if len(x.Tags) > 0 {
				renderList(&b, "tags", x.Tags, 4)
			}
			renderSource(&b, x.Source, 4)
			renderStringMap(&b, "metadata", x.Metadata, 4)
			renderBlock(&b, "content", x.Content, 4)
		}
	}
	if len(p.Workflows) > 0 {
		w("workflows:\n")
		for _, x := range p.Workflows {
			w("  - id: %s\n    name: %s\n", quote(x.ID), quote(x.Name))
			if x.Description != "" {
				w("    description: %s\n", quote(x.Description))
			}
			renderSource(&b, x.Source, 4)
			w("    steps:\n")
			for _, st := range x.Steps {
				w("      - run: %s\n", quote(st.Run))
				if st.Name != "" {
					w("        name: %s\n", quote(st.Name))
				}
				if st.Sandbox {
					w("        sandbox: true\n")
				}
			}
		}
	}
	if len(p.Hooks) > 0 {
		w("hooks:\n")
		for _, x := range p.Hooks {
			w("  - event: %s\n    run: %s\n", quote(x.Event), quote(x.Run))
			if x.Matcher != "" {
				w("    matcher: %s\n", quote(x.Matcher))
			}
			if x.Sandbox {
				w("    sandbox: true\n")
			}
			renderSource(&b, x.Source, 4)
			renderStringMap(&b, "meta", x.Meta, 4)
		}
	}
	if len(p.Plugins) > 0 {
		w("plugins:\n")
		for _, x := range p.Plugins {
			w("  - name: %s\n    runtime: %s\n", quote(x.Name), quote(x.Runtime))
			if x.Command != "" {
				w("    command: %s\n", quote(x.Command))
			}
			if len(x.Args) > 0 {
				renderList(&b, "args", x.Args, 4)
			}
			renderStringMap(&b, "env", x.Env, 4)
			if x.Sandbox {
				w("    sandbox: true\n")
			}
			if x.Description != "" {
				w("    description: %s\n", quote(x.Description))
			}
			renderSource(&b, x.Source, 4)
		}
	}
	if len(p.MCPServers) > 0 {
		w("mcpServers:\n")
		for _, x := range p.MCPServers {
			w("  - name: %s\n", quote(x.Name))
			if x.Command != "" {
				w("    command: %s\n", quote(x.Command))
			}
			if len(x.Args) > 0 {
				renderList(&b, "args", x.Args, 4)
			}
			renderStringMap(&b, "env", x.Env, 4)
			if x.URL != "" {
				w("    url: %s\n", quote(x.URL))
			}
			if x.Sandbox {
				w("    sandbox: true\n")
			}
			renderSource(&b, x.Source, 4)
		}
	}
	if len(p.Providers) > 0 {
		w("providers:\n")
		for _, x := range p.Providers {
			w("  - name: %s\n", quote(x.Name))
			if x.Type != "" {
				w("    type: %s\n", quote(x.Type))
			}
			if x.BaseURL != "" {
				w("    base_url: %s\n", quote(x.BaseURL))
			}
			if x.APIKeyEnv != "" {
				w("    api_key_env: %s\n", quote(x.APIKeyEnv))
			}
			if x.DefaultModel != "" {
				w("    default_model: %s\n", quote(x.DefaultModel))
			}
			if len(x.Models) > 0 {
				renderList(&b, "models", x.Models, 4)
			}
			renderStringMap(&b, "headers", x.Headers, 4)
			renderSource(&b, x.Source, 4)
		}
	}
	if len(p.Permissions) > 0 {
		w("permissions:\n")
		for _, x := range p.Permissions {
			w("  - subject: %s\n    action: %s\n    mode: %s\n", quote(x.Subject), quote(x.Action), quote(x.Mode))
			if len(x.Paths) > 0 {
				renderList(&b, "paths", x.Paths, 4)
			}
			renderSource(&b, x.Source, 4)
		}
	}
	if len(p.Conflicts) > 0 {
		w("conflicts:\n")
		for _, x := range p.Conflicts {
			w("  - id: %s\n    level: %s\n    topic: %s\n    message: %s\n", quote(x.ID), quote(x.Level), quote(x.Topic), quote(x.Message))
			if len(x.Sources) > 0 {
				w("    sources:\n")
				for _, s := range x.Sources {
					renderSource(&b, s, 6)
				}
			}
		}
	}
	if len(p.Risks) > 0 {
		w("risks:\n")
		for _, x := range p.Risks {
			w("  - level: %s\n    reason: %s\n    action: %s\n", quote(x.Level), quote(x.Reason), quote(x.Action))
			renderSource(&b, x.Source, 4)
		}
	}
	return b.String()
}

func ExportPack(pack *Pack, opts ExportOptions) (string, error) {
	to := opts.To
	if to == "" || to == FormatAuto {
		to = FormatClaudeCode
	}
	out := opts.OutDir
	if out == "" {
		out = exportDefaultDir(to)
	}
	var written []string
	write := func(path, content string) error {
		written = append(written, path)
		if opts.DryRun {
			return nil
		}
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return err
		}
		return os.WriteFile(path, []byte(content), 0o644)
	}
	switch to {
	case FormatClaudeCode:
		var claude strings.Builder
		for _, ins := range pack.Instructions {
			if ins.Scope == "project" || ins.Source.Format == FormatClaudeCode || ins.Source.Format == FormatCodex {
				fmt.Fprintf(&claude, "\n## %s\n\n%s\n", ins.Name, ins.Content)
			}
		}
		if strings.TrimSpace(claude.String()) != "" {
			if err := write(filepath.Join(out, "CLAUDE.md"), strings.TrimSpace(claude.String())+"\n"); err != nil {
				return "", err
			}
		}
		for _, c := range pack.Commands {
			content := renderFrontMatter(map[string]string{"description": c.Description, "argument-hint": argHint(c.Args), "allowed-tools": strings.Join(c.Tools, ", ")}) + denormalizePrompt(c.Prompt, FormatClaudeCode) + "\n"
			if err := write(filepath.Join(out, ".claude", "commands", c.Name+".md"), content); err != nil {
				return "", err
			}
		}
		for _, a := range pack.Agents {
			content := renderFrontMatter(map[string]string{"description": a.Description, "model": a.Model, "tools": strings.Join(a.Tools, ", ")}) + a.Prompt + "\n"
			if err := write(filepath.Join(out, ".claude", "agents", a.Name+".md"), content); err != nil {
				return "", err
			}
		}
	case FormatCodex:
		var b strings.Builder
		for _, ins := range pack.Instructions {
			fmt.Fprintf(&b, "\n## %s\n\n%s\n", ins.Name, ins.Content)
		}
		if err := write(filepath.Join(out, "AGENTS.md"), strings.TrimSpace(b.String())+"\n"); err != nil {
			return "", err
		}
	case FormatOpenCode:
		for _, c := range pack.Commands {
			content := renderFrontMatter(map[string]string{"description": c.Description, "tools": strings.Join(c.Tools, ", ")}) + c.Prompt + "\n"
			if err := write(filepath.Join(out, ".opencode", "commands", c.Name+".md"), content); err != nil {
				return "", err
			}
		}
		for _, a := range pack.Agents {
			content := renderFrontMatter(map[string]string{"description": a.Description, "model": a.Model, "tools": strings.Join(a.Tools, ", ")}) + a.Prompt + "\n"
			if err := write(filepath.Join(out, ".opencode", "agents", a.Name+".md"), content); err != nil {
				return "", err
			}
		}
		for _, s := range pack.Skills {
			content := renderFrontMatter(map[string]string{"description": s.Description, "tags": strings.Join(s.Tags, ", ")}) + s.Content + "\n"
			if err := write(filepath.Join(out, ".opencode", "skills", s.Name+".md"), content); err != nil {
				return "", err
			}
		}
	default:
		return "", fmt.Errorf("unsupported export target %q", to)
	}
	if opts.DryRun {
		return "would write:\n- " + strings.Join(written, "\n- ") + "\n", nil
	}
	return strings.Join(written, "\n"), nil
}

func renderSource(b *strings.Builder, src SourceRef, indent int) {
	pad := strings.Repeat(" ", indent)
	fmt.Fprintf(b, "%ssource:\n%s  format: %s\n%s  path: %s\n%s  kind: %s\n", pad, pad, quote(src.Format), pad, quote(src.Path), pad, quote(src.Kind))
}

func renderList(b *strings.Builder, key string, values []string, indent int) {
	if len(values) == 0 {
		return
	}
	pad := strings.Repeat(" ", indent)
	fmt.Fprintf(b, "%s%s:\n", pad, key)
	for _, v := range values {
		fmt.Fprintf(b, "%s  - %s\n", pad, quote(v))
	}
}

func renderStringMap(b *strings.Builder, key string, values map[string]string, indent int) {
	if len(values) == 0 {
		return
	}
	pad := strings.Repeat(" ", indent)
	fmt.Fprintf(b, "%s%s:\n", pad, key)
	keys := make([]string, 0, len(values))
	for k := range values {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		fmt.Fprintf(b, "%s  %s: %s\n", pad, quote(k), quote(values[k]))
	}
}

func renderBlock(b *strings.Builder, key, value string, indent int) {
	pad := strings.Repeat(" ", indent)
	fmt.Fprintf(b, "%s%s: |\n", pad, key)
	if strings.TrimSpace(value) == "" {
		fmt.Fprintf(b, "%s  \n", pad)
		return
	}
	for _, line := range strings.Split(value, "\n") {
		fmt.Fprintf(b, "%s  %s\n", pad, line)
	}
}

func quote(s string) string { data, _ := json.Marshal(s); return string(data) }

func SHA256File(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func ValidatePackFile(path string) []string {
	data, err := os.ReadFile(path)
	if err != nil {
		return []string{"cannot read pack: " + err.Error()}
	}
	s := string(data)
	return ValidatePackText(s)
}

func ValidatePackText(s string) []string {
	var issues []string
	if !strings.Contains(s, "apiVersion: ti.dev/v1") {
		issues = append(issues, "missing apiVersion: ti.dev/v1")
	}
	if !strings.Contains(s, "kind: TiPack") {
		issues = append(issues, "missing kind: TiPack")
	}
	if !strings.Contains(s, "metadata:") {
		issues = append(issues, "missing metadata")
	}
	if strings.Contains(s, "run:") && !strings.Contains(s, "sandbox: true") {
		issues = append(issues, "pack has executable hooks/steps without explicit sandbox")
	}
	if strings.Contains(s, "command:") && strings.Contains(s, "plugins:") && !strings.Contains(s, "sandbox: true") {
		issues = append(issues, "plugin command without sandbox")
	}
	return issues
}

func renderFrontMatter(m map[string]string) string {
	var b strings.Builder
	b.WriteString("---\n")
	keys := make([]string, 0, len(m))
	for k, v := range m {
		if strings.TrimSpace(v) != "" {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)
	for _, k := range keys {
		fmt.Fprintf(&b, "%s: %s\n", k, m[k])
	}
	b.WriteString("---\n")
	return b.String()
}
func argHint(args []Arg) string {
	var parts []string
	for _, a := range args {
		if a.Name == "arguments" {
			continue
		}
		if a.Required {
			parts = append(parts, "<"+a.Name+">")
		} else {
			parts = append(parts, "["+a.Name+"]")
		}
	}
	if len(parts) == 0 {
		for _, a := range args {
			if a.Name == "arguments" {
				return "[arguments]"
			}
		}
	}
	return strings.Join(parts, " ")
}
func denormalizePrompt(s, to string) string {
	if to == FormatClaudeCode {
		return strings.ReplaceAll(s, "{{args.arguments}}", "$ARGUMENTS")
	}
	return s
}
func exportDefaultDir(to string) string {
	switch to {
	case FormatClaudeCode:
		return "."
	case FormatCodex:
		return "."
	case FormatOpenCode:
		return "."
	default:
		return "."
	}
}
