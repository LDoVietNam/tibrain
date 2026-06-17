package repoindex

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type FileInfo struct {
	Path        string   `json:"path"`
	Size        int64    `json:"size"`
	Ext         string   `json:"ext"`
	Kind        string   `json:"kind"`
	Signals     []string `json:"signals,omitempty"`
	Dir         string   `json:"dir,omitempty"`
	IsTest      bool     `json:"is_test"`
	IsGenerated bool     `json:"is_generated"`
}

type Summary struct {
	Root            string         `json:"root"`
	Module          string         `json:"module,omitempty"`
	FileCount       int            `json:"file_count"`
	DirCount        int            `json:"dir_count"`
	TestFileCount   int            `json:"test_file_count"`
	Languages       map[string]int `json:"languages"`
	TopDirectories  []string       `json:"top_directories,omitempty"`
	EntryPoints     []string       `json:"entry_points,omitempty"`
	ConfigFiles     []string       `json:"config_files,omitempty"`
	Docs            []string       `json:"docs,omitempty"`
	Files           []FileInfo     `json:"files,omitempty"`
	IgnoredPatterns []string       `json:"ignored_patterns,omitempty"`
}

var defaultIgnores = []string{
	".git", ".hg", ".svn", "node_modules", "vendor", "dist", "build", "bin", ".idea", ".vscode", ".next", ".turbo",
}

func Analyze(root string) (*Summary, error) {
	root, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}
	s := &Summary{Root: root, Languages: map[string]int{}, IgnoredPatterns: append([]string(nil), defaultIgnores...)}
	topDirs := map[string]int{}
	err = filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		rel, relErr := filepath.Rel(root, path)
		if relErr != nil {
			return nil
		}
		if rel == "." {
			return nil
		}
		if shouldIgnore(rel) {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if d.IsDir() {
			s.DirCount++
			top := strings.Split(rel, string(filepath.Separator))[0]
			if top != "." && top != "" {
				topDirs[top]++
			}
			return nil
		}
		info, statErr := d.Info()
		if statErr != nil {
			return nil
		}
		fi := classifyFile(rel, info.Size())
		s.Files = append(s.Files, fi)
		s.FileCount++
		if fi.IsTest {
			s.TestFileCount++
		}
		if fi.Kind != "" {
			s.Languages[fi.Kind]++
		}
		if isEntryPoint(fi.Path) {
			s.EntryPoints = append(s.EntryPoints, fi.Path)
		}
		if isConfigFile(fi.Path) {
			s.ConfigFiles = append(s.ConfigFiles, fi.Path)
		}
		if isDocFile(fi.Path) {
			s.Docs = append(s.Docs, fi.Path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Slice(s.Files, func(i, j int) bool { return s.Files[i].Path < s.Files[j].Path })
	s.TopDirectories = topDirectoryList(topDirs)
	if module, err := readGoModule(root); err == nil {
		s.Module = module
	}
	return s, nil
}

func SearchRelevantFiles(s *Summary, task string, limit int) []FileInfo {
	if s == nil || len(s.Files) == 0 {
		return nil
	}
	tokens := taskTokens(task)
	type scored struct {
		FileInfo
		score int
	}
	var ranked []scored
	for _, f := range s.Files {
		score := scoreFile(f, tokens)
		if score <= 0 {
			continue
		}
		ranked = append(ranked, scored{FileInfo: f, score: score})
	}
	sort.SliceStable(ranked, func(i, j int) bool {
		if ranked[i].score == ranked[j].score {
			return ranked[i].Path < ranked[j].Path
		}
		return ranked[i].score > ranked[j].score
	})
	if limit <= 0 || limit > len(ranked) {
		limit = len(ranked)
	}
	out := make([]FileInfo, 0, limit)
	for _, r := range ranked[:limit] {
		out = append(out, r.FileInfo)
	}
	return out
}

func RenderSummaryMarkdown(s *Summary) string {
	if s == nil {
		return ""
	}
	var b strings.Builder
	fmt.Fprintf(&b, "# Repo Summary\n\n")
	fmt.Fprintf(&b, "- Root: `%s`\n", s.Root)
	if s.Module != "" {
		fmt.Fprintf(&b, "- Module: `%s`\n", s.Module)
	}
	fmt.Fprintf(&b, "- Files: %d\n", s.FileCount)
	fmt.Fprintf(&b, "- Directories: %d\n", s.DirCount)
	fmt.Fprintf(&b, "- Tests: %d\n", s.TestFileCount)
	if len(s.TopDirectories) > 0 {
		fmt.Fprintf(&b, "- Top directories: %s\n", strings.Join(s.TopDirectories, ", "))
	}
	if len(s.EntryPoints) > 0 {
		fmt.Fprintf(&b, "\n## Likely entry points\n")
		for _, p := range s.EntryPoints[:min(8, len(s.EntryPoints))] {
			fmt.Fprintf(&b, "- `%s`\n", p)
		}
	}
	if len(s.ConfigFiles) > 0 {
		fmt.Fprintf(&b, "\n## Config files\n")
		for _, p := range s.ConfigFiles[:min(8, len(s.ConfigFiles))] {
			fmt.Fprintf(&b, "- `%s`\n", p)
		}
	}
	return b.String()
}

func scoreFile(f FileInfo, tokens []string) int {
	if len(tokens) == 0 {
		if f.IsTest || strings.Contains(strings.ToLower(f.Path), "readme") || strings.Contains(strings.ToLower(f.Path), "main") {
			return 1
		}
		return 0
	}
	score := 0
	lp := strings.ToLower(f.Path)
	for _, token := range tokens {
		if strings.Contains(lp, token) {
			score += 4
		}
		for _, sig := range f.Signals {
			if strings.Contains(strings.ToLower(sig), token) {
				score += 3
			}
		}
	}
	if f.IsTest {
		score++
	}
	if isEntryPoint(f.Path) {
		score++
	}
	return score
}

func taskTokens(task string) []string {
	task = strings.ToLower(task)
	replacer := strings.NewReplacer("-", " ", "_", " ", "/", " ", ".", " ", ",", " ", ":", " ", ";", " ", "(", " ", ")", " ")
	task = replacer.Replace(task)
	stop := map[string]bool{
		"the": true, "and": true, "for": true, "with": true, "from": true, "into": true, "this": true, "that": true,
		"repo": true, "project": true,
	}
	parts := strings.Fields(task)
	seen := map[string]bool{}
	var tokens []string
	for _, p := range parts {
		if len(p) < 2 || stop[p] {
			continue
		}
		if !seen[p] {
			seen[p] = true
			tokens = append(tokens, p)
		}
	}
	return tokens
}

func classifyFile(rel string, size int64) FileInfo {
	ext := strings.ToLower(filepath.Ext(rel))
	base := strings.ToLower(filepath.Base(rel))
	kind := map[string]string{
		".go": "go", ".js": "javascript", ".ts": "typescript", ".tsx": "typescript", ".jsx": "javascript",
		".py": "python", ".rs": "rust", ".java": "java", ".rb": "ruby", ".php": "php", ".sh": "shell", ".sql": "sql",
		".md": "markdown", ".yaml": "yaml", ".yml": "yaml", ".json": "json", ".toml": "toml", ".env": "env",
	}[ext]
	signals := []string{}
	if strings.Contains(base, "test") || strings.HasSuffix(base, "_test.go") {
		signals = append(signals, "test")
	}
	if strings.Contains(base, "readme") || ext == ".md" {
		signals = append(signals, "docs")
	}
	if strings.Contains(rel, "cmd"+string(filepath.Separator)) || base == "main.go" {
		signals = append(signals, "entrypoint")
	}
	if strings.HasPrefix(base, "docker") || base == "dockerfile" {
		signals = append(signals, "infra")
	}
	return FileInfo{
		Path:        rel,
		Size:        size,
		Ext:         ext,
		Kind:        kind,
		Signals:     signals,
		Dir:         filepath.Dir(rel),
		IsTest:      strings.HasSuffix(base, "_test.go") || strings.Contains(base, ".test."),
		IsGenerated: strings.Contains(strings.ToLower(rel), "generated") || strings.Contains(strings.ToLower(rel), "gen.go"),
	}
}

func shouldIgnore(rel string) bool {
	parts := strings.Split(rel, string(filepath.Separator))
	for _, p := range parts {
		for _, ignore := range defaultIgnores {
			if p == ignore {
				return true
			}
		}
	}
	base := filepath.Base(rel)
	return strings.HasPrefix(base, ".") && base != ".ti" && base != ".github"
}

func isEntryPoint(path string) bool {
	base := strings.ToLower(filepath.Base(path))
	return base == "main.go" || strings.HasPrefix(path, "cmd"+string(filepath.Separator)) || strings.HasSuffix(path, ".sh")
}

func isConfigFile(path string) bool {
	base := strings.ToLower(filepath.Base(path))
	switch base {
	case "go.mod", "go.sum", "package.json", "ti.json", "dockerfile", "makefile", ".env":
		return true
	}
	ext := strings.ToLower(filepath.Ext(path))
	return ext == ".yaml" || ext == ".yml" || ext == ".json" || ext == ".toml"
}

func isDocFile(path string) bool {
	base := strings.ToLower(filepath.Base(path))
	return strings.HasSuffix(base, ".md") || strings.Contains(base, "readme")
}

func readGoModule(root string) (string, error) {
	f, err := os.Open(filepath.Join(root, "go.mod"))
	if err != nil {
		return "", err
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if strings.HasPrefix(line, "module ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "module ")), nil
		}
	}
	return "", sc.Err()
}

func topDirectoryList(m map[string]int) []string {
	type kv struct {
		k string
		v int
	}
	items := make([]kv, 0, len(m))
	for k, v := range m {
		items = append(items, kv{k, v})
	}
	sort.Slice(items, func(i, j int) bool {
		if items[i].v == items[j].v {
			return items[i].k < items[j].k
		}
		return items[i].v > items[j].v
	})
	out := make([]string, 0, min(8, len(items)))
	for _, it := range items[:min(8, len(items))] {
		out = append(out, it.k)
	}
	return out
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
