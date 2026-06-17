package devkitassets

import (
	"embed"
	"fmt"
	"io/fs"
	"path/filepath"
	"sort"
	"strings"
)

// FS contains built-in AI DevKit templates.
//
//go:embed templates/commands/*.md templates/phases/*.md
var FS embed.FS

func List(kind string) ([]string, error) {
	dir := "templates/" + kind
	entries, err := fs.ReadDir(FS, dir)
	if err != nil {
		return nil, err
	}
	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".md") {
			names = append(names, strings.TrimSuffix(entry.Name(), ".md"))
		}
	}
	sort.Strings(names)
	return names, nil
}

func Read(kind, name string) ([]byte, error) {
	name = strings.TrimSuffix(strings.ToLower(strings.TrimSpace(name)), ".md")
	path := filepath.ToSlash(filepath.Join("templates", kind, name+".md"))
	data, err := FS.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("embedded %s template %q not found", kind, name)
	}
	return data, nil
}
