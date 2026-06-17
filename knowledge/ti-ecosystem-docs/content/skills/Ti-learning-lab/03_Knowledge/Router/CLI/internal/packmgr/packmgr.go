package packmgr

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type PackInfo struct {
	Name     string    `json:"name"`
	Path     string    `json:"path"`
	Enabled  bool      `json:"enabled"`
	Modified time.Time `json:"modified"`
	Summary  Summary   `json:"summary"`
}

type Summary struct {
	Instructions int `json:"instructions"`
	Commands     int `json:"commands"`
	Agents       int `json:"agents"`
	Skills       int `json:"skills"`
	Workflows    int `json:"workflows"`
	Hooks        int `json:"hooks"`
	MCPServers   int `json:"mcpServers"`
	Risks        int `json:"risks"`
}

func DefaultRoot() string {
	if home, err := os.UserHomeDir(); err == nil && home != "" {
		return filepath.Join(home, ".ti", "packs")
	}
	return filepath.Join(".ti", "packs")
}

func ProjectRoot() string { return filepath.Join(".ti", "packs") }

func List(root string) ([]PackInfo, error) {
	if strings.TrimSpace(root) == "" {
		root = ProjectRoot()
	}
	root = expand(root)
	var out []PackInfo
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			return nil
		}
		if d.Name() != "ti.pack.yaml" {
			return nil
		}
		info, err := Inspect(path)
		if err == nil {
			out = append(out, info)
		}
		return nil
	})
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out, err
}

func Inspect(path string) (PackInfo, error) {
	path = expand(path)
	if fi, err := os.Stat(path); err == nil && fi.IsDir() {
		path = filepath.Join(path, "ti.pack.yaml")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return PackInfo{}, err
	}
	fi, _ := os.Stat(path)
	dir := filepath.Dir(path)
	info := PackInfo{Path: path, Enabled: !fileExists(filepath.Join(dir, ".disabled")), Modified: time.Now()}
	if fi != nil {
		info.Modified = fi.ModTime()
	}
	info.Name = readYAMLScalar(data, "name")
	if info.Name == "" {
		info.Name = filepath.Base(dir)
	}
	info.Summary = countSections(data)
	return info, nil
}

func Enable(path string, enable bool) error {
	path = expand(path)
	if fi, err := os.Stat(path); err == nil && !fi.IsDir() {
		path = filepath.Dir(path)
	}
	if err := os.MkdirAll(path, 0o755); err != nil {
		return err
	}
	marker := filepath.Join(path, ".disabled")
	if enable {
		if err := os.Remove(marker); err != nil && !errors.Is(err, os.ErrNotExist) {
			return err
		}
		return nil
	}
	return os.WriteFile(marker, []byte("disabled by ti pack\n"), 0o644)
}

func RenderList(items []PackInfo) string {
	var b strings.Builder
	fmt.Fprintf(&b, "%-24s %-8s %-4s %-4s %-4s %-4s %-5s %s\n", "NAME", "ENABLED", "CMD", "AGT", "SKL", "MCP", "RISK", "PATH")
	for _, p := range items {
		fmt.Fprintf(&b, "%-24s %-8v %-4d %-4d %-4d %-4d %-5d %s\n", p.Name, p.Enabled, p.Summary.Commands, p.Summary.Agents, p.Summary.Skills, p.Summary.MCPServers, p.Summary.Risks, p.Path)
	}
	return b.String()
}

func JSON(v any) string { b, _ := json.MarshalIndent(v, "", "  "); return string(b) }

func readYAMLScalar(data []byte, key string) string {
	sc := bufio.NewScanner(strings.NewReader(string(data)))
	inMetadata := false
	for sc.Scan() {
		line := sc.Text()
		trimmed := strings.TrimSpace(line)
		if trimmed == "metadata:" {
			inMetadata = true
			continue
		}
		if inMetadata && !strings.HasPrefix(line, " ") && trimmed != "" {
			inMetadata = false
		}
		if !inMetadata {
			continue
		}
		if strings.HasPrefix(trimmed, key+":") {
			v := strings.TrimSpace(strings.TrimPrefix(trimmed, key+":"))
			v = strings.Trim(v, "\"")
			return v
		}
	}
	return ""
}

func countSections(data []byte) Summary {
	var s Summary
	sc := bufio.NewScanner(strings.NewReader(string(data)))
	section := ""
	for sc.Scan() {
		line := sc.Text()
		if strings.TrimSpace(line) == "" {
			continue
		}
		if !strings.HasPrefix(line, " ") && strings.HasSuffix(strings.TrimSpace(line), ":") {
			section = strings.TrimSuffix(strings.TrimSpace(line), ":")
			continue
		}
		if strings.HasPrefix(strings.TrimSpace(line), "- ") {
			switch section {
			case "instructions":
				s.Instructions++
			case "commands":
				s.Commands++
			case "agents":
				s.Agents++
			case "skills":
				s.Skills++
			case "workflows":
				s.Workflows++
			case "hooks":
				s.Hooks++
			case "mcpServers":
				s.MCPServers++
			case "risks":
				s.Risks++
			}
		}
	}
	return s
}

func expand(p string) string {
	if p == "~" || strings.HasPrefix(p, "~/") {
		if home, err := os.UserHomeDir(); err == nil && home != "" {
			if p == "~" {
				return home
			}
			return filepath.Join(home, p[2:])
		}
	}
	return os.ExpandEnv(p)
}
func fileExists(path string) bool { _, err := os.Stat(path); return err == nil }
