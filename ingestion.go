package main

import (
	"os"
	"path/filepath"
	"strings"
	"time"
)

// DocumentIngester walks knowledge directories and indexes documents
type DocumentIngester struct {
	ragManager *RAGSystemManager
	chunkSize  int
	overlap    int
}

func NewDocumentIngester(ragManager *RAGSystemManager) *DocumentIngester {
	return &DocumentIngester{
		ragManager: ragManager,
		chunkSize:  512,
		overlap:    128,
	}
}

// IngestDirectory walks a directory and indexes all supported files
func (di *DocumentIngester) IngestDirectory(dir string, category string) (int, error) {
	indexed := 0
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))
		if ext != ".md" && ext != ".txt" && ext != ".go" && ext != ".json" {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return nil
		}

		title := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
		doc := RAGDocument{
			ID:        generateID(),
			Title:     title,
			Content:   string(content),
			Path:      path,
			Category:  category,
			Tags:      "",
			Status:    "active",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		if err := di.ragManager.AddDocument(doc); err != nil {
			logger.Warn("Failed to index %s: %v", path, err)
		} else {
			indexed++
		}
		return nil
	})

	logger.Info("Ingested %d documents from %s", indexed, dir)
	return indexed, err
}

// AutoIngest discovers and indexes standard knowledge directories
func (di *DocumentIngester) AutoIngest(basePath string) map[string]int {
	results := make(map[string]int)

	dirs := map[string]string{
		"docs":     filepath.Join(basePath, "docs"),
		"apps":     filepath.Join(basePath, "apps-docs"),
		"guides":   filepath.Join(basePath, "guides"),
		"examples": filepath.Join(basePath, "examples"),
	}

	for category, dir := range dirs {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			continue
		}
		count, _ := di.IngestDirectory(dir, category)
		results[category] = count
	}

	return results
}

// IngestObsidianVault indexes .md files from an Obsidian vault directory into Ti Brain.
// Obsidian vaults are markdown-first with wikilinks, but we treat each .md as a standalone document.
// Set vaultPath = "" to skip (no Obsidian vault configured).
func (di *DocumentIngester) IngestObsidianVault(vaultPath string) (int, error) {
	if vaultPath == "" {
		return 0, nil
	}
	if _, err := os.Stat(vaultPath); os.IsNotExist(err) {
		logger.Warn("Obsidian vault not found at %s — skipping", vaultPath)
		return 0, nil
	}
	logger.Info("Ingesting Obsidian vault from %s", vaultPath)
	return di.IngestDirectory(vaultPath, "obsidian")
}
