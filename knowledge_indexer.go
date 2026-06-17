package main

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const defaultMaxIndexFileBytes int64 = 512 * 1024

type KnowledgeIndexSource struct {
	Path        string `json:"path"`
	Category    string `json:"category"`
	Description string `json:"description"`
	// ShallowOnly limits indexing to files directly inside Path (no recursion into subdirectories).
	// Useful for app root dirs where you want READMEs/docs but not all source code.
	ShallowOnly bool `json:"shallow_only,omitempty"`
}

type KnowledgeIndexOptions struct {
	Sources      []KnowledgeIndexSource `json:"sources"`
	Force        bool                   `json:"force"`
	MaxFileBytes int64                  `json:"max_file_bytes"`
}

type KnowledgeIndexResult struct {
	Indexed int      `json:"indexed"`
	Skipped int      `json:"skipped"`
	Errors  []string `json:"errors"`
	Sources int      `json:"sources"`
}

type KnowledgeIndexer struct {
	hub        *Hub
	ragManager *RAGSystemManager
	root       string
}

func NewKnowledgeIndexer(hub *Hub) *KnowledgeIndexer {
	ragManager := NewRAGSystemManager(hub)
	ragManager.deferKnowledgeBaseStats = true
	return &KnowledgeIndexer{
		hub:        hub,
		ragManager: ragManager,
		root:       workspaceRoot(),
	}
}

func DefaultKnowledgeSources() []KnowledgeIndexSource {
	root := workspaceRoot()
	appsDir := filepath.Join(root, "apps")
	return []KnowledgeIndexSource{
		// ── Workspace root ────────────────────────────────────────────────────
		{Path: filepath.Join(root, "AGENTS.md"), Category: "central-instructions", Description: "Agent operating instructions and Ti Brain architecture"},
		{Path: filepath.Join(root, "README.md"), Category: "project-overview", Description: "Workspace project overview"},

		// ── Curated apps-docs (legacy, kept for backward compat) ─────────────
		{Path: filepath.Join(root, "apps-docs"), Category: "apps-docs", Description: "Curated application documentation"},

		// ── docs/ subdirs ─────────────────────────────────────────────────────
		{Path: filepath.Join(root, "docs", "00-index"), Category: "docs-index", Description: "Documentation indexes"},
		{Path: filepath.Join(root, "docs", "01-architecture"), Category: "architecture", Description: "Architecture documentation"},
		{Path: filepath.Join(root, "docs", "03-components"), Category: "components", Description: "Component documentation"},
		{Path: filepath.Join(root, "docs", "06-rag"), Category: "rag", Description: "RAG documentation"},
		{Path: filepath.Join(root, "docs", "router-agent-architecture.md"), Category: "router-agent", Description: "Router Agent architecture"},

		// ── apps/ — root-level docs (README, Makefile, UNIFIED_ARCHITECTURE) ─
		// ShallowOnly: only files at the top of apps/, not all source code.
		{Path: appsDir, Category: "apps-overview", Description: "Apps directory overview and Makefile", ShallowOnly: true},

		// ── Per-app: root-level docs (README, ARCHITECTURE, AGENTS, etc.) ────
		// ShallowOnly keeps us out of Go/Python source trees.
		{Path: filepath.Join(appsDir, "cli"), Category: "app-cli", Description: "Ti CLI root docs", ShallowOnly: true},
		{Path: filepath.Join(appsDir, "router"), Category: "app-router", Description: "Ti Router root docs", ShallowOnly: true},
		{Path: filepath.Join(appsDir, "tibrain"), Category: "app-tibrain", Description: "Ti Brain root docs", ShallowOnly: true},
		{Path: filepath.Join(appsDir, "ticrew"), Category: "app-ticrew", Description: "TiCrew root docs", ShallowOnly: true},
		{Path: filepath.Join(appsDir, "unified-loadbalancer"), Category: "app-unified-lb", Description: "Unified Load Balancer root docs", ShallowOnly: true},
		{Path: filepath.Join(appsDir, "load_balancer"), Category: "app-load-balancer-deprecated", Description: "Load Balancer (deprecated) root docs", ShallowOnly: true},
		{Path: filepath.Join(appsDir, "intelligent-agent"), Category: "app-intelligent-agent-superseded", Description: "Intelligent Agent (superseded) root docs", ShallowOnly: true},
		{Path: filepath.Join(appsDir, "email-multi-checker"), Category: "app-email-checker", Description: "Email Multi-Checker root docs", ShallowOnly: true},
		{Path: filepath.Join(appsDir, "mini-browser"), Category: "app-mini-browser", Description: "Mini Browser root docs", ShallowOnly: true},
		{Path: filepath.Join(appsDir, "auto_reg"), Category: "app-auto-reg", Description: "Auto Registration root docs", ShallowOnly: true},
		{Path: filepath.Join(appsDir, "mcp"), Category: "app-mcp-stub", Description: "MCP stub/redirect docs", ShallowOnly: true},

		// ── Per-app: docs/ subdirs (full recursion where rich docs exist) ─────
		{Path: filepath.Join(appsDir, "cli", "docs"), Category: "app-cli-docs", Description: "Ti CLI full documentation"},
		{Path: filepath.Join(appsDir, "router", "docs"), Category: "app-router-docs", Description: "Ti Router full documentation (635+ files)"},
		{Path: filepath.Join(appsDir, "ticrew", "docs"), Category: "app-ticrew-docs", Description: "TiCrew full documentation"},

		// ── Router Agent Brain (kept from original) ───────────────────────────
		{Path: filepath.Join(appsDir, "router", "brain"), Category: "router-brain", Description: "Router Agent Brain implementation and docs"},

		// ── integrations/ MCP docs ────────────────────────────────────────────
		{Path: filepath.Join(appsDir, "integrations", "mcp"), Category: "integrations-mcp", Description: "MCP server implementations and configs"},
	}
}

func (ki *KnowledgeIndexer) Index(options KnowledgeIndexOptions) (*KnowledgeIndexResult, error) {
	if ki == nil || ki.hub == nil || ki.ragManager == nil {
		return nil, fmt.Errorf("knowledge indexer is not initialized")
	}
	if len(options.Sources) == 0 {
		options.Sources = DefaultKnowledgeSources()
	}
	if options.MaxFileBytes <= 0 {
		options.MaxFileBytes = defaultMaxIndexFileBytes
	}

	result := &KnowledgeIndexResult{Sources: len(options.Sources)}
	for _, source := range options.Sources {
		if source.Category == "" {
			source.Category = categoryFromPath(source.Path)
		}
		if err := ki.ensureKnowledgeBase(source); err != nil {
			result.Errors = append(result.Errors, err.Error())
			continue
		}

		info, err := os.Stat(source.Path)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("%s: %v", source.Path, err))
			continue
		}
		if info.IsDir() {
			ki.indexDirectory(source, options, result)
		} else {
			ki.indexFile(source.Path, source, options, result)
		}
	}

	ki.ragManager.flushKnowledgeBaseStats()

	return result, nil
}

func (ki *KnowledgeIndexer) ensureKnowledgeBase(source KnowledgeIndexSource) error {
	return ki.ragManager.CreateKnowledgeBase(RAGKnowledgeBase{
		ID:          source.Category,
		Name:        source.Category,
		Description: source.Description,
		Path:        source.Path,
		Status:      "active",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	})
}

func (ki *KnowledgeIndexer) indexDirectory(source KnowledgeIndexSource, options KnowledgeIndexOptions, result *KnowledgeIndexResult) {
	if source.ShallowOnly {
		// Only index files directly inside the directory, skip all subdirectories.
		entries, err := os.ReadDir(source.Path)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("%s: %v", source.Path, err))
			return
		}
		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			ki.indexFile(filepath.Join(source.Path, entry.Name()), source, options, result)
		}
		return
	}

	err := filepath.WalkDir(source.Path, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("%s: %v", path, walkErr))
			return nil
		}
		if entry.IsDir() {
			if shouldSkipDir(entry.Name()) {
				return filepath.SkipDir
			}
			return nil
		}
		ki.indexFile(path, source, options, result)
		return nil
	})
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("%s: %v", source.Path, err))
	}
}

func (ki *KnowledgeIndexer) indexFile(path string, source KnowledgeIndexSource, options KnowledgeIndexOptions, result *KnowledgeIndexResult) {
	if !isIndexableFile(path) {
		result.Skipped++
		return
	}

	info, err := os.Stat(path)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("%s: %v", path, err))
		return
	}
	if info.Size() == 0 || info.Size() > options.MaxFileBytes {
		result.Skipped++
		return
	}

	contentBytes, err := os.ReadFile(path)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("%s: %v", path, err))
		return
	}
	content := strings.TrimSpace(string(contentBytes))
	if content == "" {
		result.Skipped++
		return
	}

	relPath := path
	if rel, err := filepath.Rel(ki.root, path); err == nil {
		relPath = rel
	}
	metadata, _ := json.Marshal(map[string]interface{}{
		"source_category": source.Category,
		"source_root":     source.Path,
		"indexed_at":      time.Now().Format(time.RFC3339),
		"file_size":       info.Size(),
	})

	doc := RAGDocument{
		ID:       "doc-" + generateContentHash(relPath),
		Title:    titleFromPath(path),
		Content:  content,
		Path:     relPath,
		Category: source.Category,
		Tags:     tagsFromPath(relPath, source.Category),
		Status:   "active",
		Metadata: string(metadata),
	}

	if err := ki.ragManager.AddDocument(doc); err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("%s: %v", path, err))
		return
	}
	result.Indexed++
}

func workspaceRoot() string {
	wd, err := os.Getwd()
	if err != nil {
		return "."
	}
	clean := filepath.Clean(wd)
	if filepath.Base(clean) == "tibrain" && filepath.Base(filepath.Dir(clean)) == "apps" {
		return filepath.Dir(filepath.Dir(clean))
	}
	return clean
}

func isIndexableFile(path string) bool {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".md", ".txt", ".yaml", ".yml", ".json", ".jsonc", ".go", ".toml":
		return true
	default:
		return false
	}
}

func shouldSkipDir(name string) bool {
	switch strings.ToLower(name) {
	case ".git", "node_modules", "dist", "build", "vendor", "venv", ".venv", "__pycache__", "memory", "archive", "99-archive":
		return true
	default:
		return false
	}
}

func categoryFromPath(path string) string {
	base := strings.ToLower(filepath.Base(filepath.Clean(path)))
	base = strings.TrimSuffix(base, filepath.Ext(base))
	base = strings.ReplaceAll(base, " ", "-")
	if base == "" || base == "." {
		return "general"
	}
	return base
}

func titleFromPath(path string) string {
	base := filepath.Base(path)
	ext := filepath.Ext(base)
	return strings.TrimSuffix(base, ext)
}

func tagsFromPath(path, category string) string {
	parts := []string{category}
	for _, part := range strings.Split(filepath.ToSlash(path), "/") {
		part = strings.TrimSpace(strings.ToLower(part))
		if part == "" || strings.Contains(part, ".") {
			continue
		}
		parts = append(parts, part)
	}
	return strings.Join(uniqueStrings(parts), ",")
}

func uniqueStrings(values []string) []string {
	seen := make(map[string]bool)
	result := []string{}
	for _, value := range values {
		if seen[value] {
			continue
		}
		seen[value] = true
		result = append(result, value)
	}
	return result
}
