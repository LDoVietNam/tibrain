package projectmemory

import (
	"os"
	"path/filepath"
	"strings"
)

var allowedExt = map[string]bool{
	".md": true, ".txt": true, ".json": true, ".jsonl": true,
	".yaml": true, ".yml": true, ".toml": true,
}

func Discover(root string) ([]FileSummary, error) {
	root, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}
	var out []FileSummary
	err = filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		name := d.Name()
		if d.IsDir() {
			if shouldSkipDir(name) {
				return filepath.SkipDir
			}
			return nil
		}
		if !allowedExt[strings.ToLower(filepath.Ext(name))] {
			return nil
		}
		rel, _ := filepath.Rel(root, path)
		kind := classify(rel)
		if kind == "ignore" {
			return nil
		}
		out = append(out, summarize(path, kind))
		return nil
	})
	return out, err
}

func shouldSkipDir(name string) bool {
	s := strings.ToLower(name)
	switch s {
	case ".git", "node_modules", "bin", "dist", "build", ".next", "vendor", "__pycache__":
		return true
	}
	return false
}

func classify(rel string) string {
	r := filepath.ToSlash(strings.ToLower(rel))
	switch {
	case r == "readme.md" || r == "agents.md":
		return "root_doc"
	case strings.HasPrefix(r, "docs/"):
		return "project_doc"
	case strings.HasPrefix(r, ".ti/"):
		return "project_memory"
	case strings.Contains(r, "/brain/"):
		return "member_brain"
	case strings.Contains(r, "/memory/"):
		return "member_memory"
	case strings.Contains(r, "/profile/"):
		return "member_profile"
	case strings.Contains(r, "/docs/"):
		return "member_doc"
	case strings.Contains(r, "/logs/"):
		return "ignore"
	}
	return "ignore"
}
