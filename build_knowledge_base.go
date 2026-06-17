package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// KnowledgeBaseBuilder builds unified knowledge base
type KnowledgeBaseBuilder struct {
	sourceDir      string
	targetDir      string
	totalFiles     int
	processedFiles int
	errors         []string
	startTime      time.Time
}

// NewKnowledgeBaseBuilder creates a new builder
func NewKnowledgeBaseBuilder(sourceDir, targetDir string) *KnowledgeBaseBuilder {
	return &KnowledgeBaseBuilder{
		sourceDir: sourceDir,
		targetDir: targetDir,
		startTime: time.Now(),
	}
}

// Build builds the knowledge base
func (kbb *KnowledgeBaseBuilder) Build() error {
	log.Println("=== 🚀 BUILDING UNIFIED KNOWLEDGE BASE ===")
	log.Printf("Source: %s", kbb.sourceDir)
	log.Printf("Target: %s", kbb.targetDir)

	// Create target directories
	if err := kbb.createDirectories(); err != nil {
		return fmt.Errorf("create directories: %w", err)
	}

	// Scan and collect files
	if err := kbb.scanAndCollect(); err != nil {
		return fmt.Errorf("scan and collect: %w", err)
	}

	// Generate index
	if err := kbb.generateIndex(); err != nil {
		return fmt.Errorf("generate index: %w", err)
	}

	// Print summary
	kbb.printSummary()

	return nil
}

// createDirectories creates target directories
func (kbb *KnowledgeBaseBuilder) createDirectories() error {
	dirs := []string{
		"content",
		"content/skills",
		"content/agents",
		"content/rules",
		"content/workflows",
		"apps-docs",
		"packages",
		"docs",
	}

	for _, dir := range dirs {
		fullPath := filepath.Join(kbb.targetDir, dir)
		if err := os.MkdirAll(fullPath, 0755); err != nil {
			return fmt.Errorf("create dir %s: %w", fullPath, err)
		}
	}

	log.Println("✅ Directory structure created")
	return nil
}

// scanAndCollect scans and collects files
func (kbb *KnowledgeBaseBuilder) scanAndCollect() error {
	log.Println("🔍 Scanning and collecting files...")

	// Walk through source directory
	err := filepath.Walk(kbb.sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			kbb.errors = append(kbb.errors, fmt.Sprintf("Error accessing %s: %v", path, err))
			return nil
		}

		// Skip directories and hidden files
		if info.IsDir() || strings.HasPrefix(filepath.Base(path), ".") {
			return nil
		}

		// Process only specific file types
		ext := strings.ToLower(filepath.Ext(path))
		if !kbb.isTargetFile(ext) {
			return nil
		}

		kbb.totalFiles++

		// Process file
		if err := kbb.processFile(path); err != nil {
			kbb.errors = append(kbb.errors, fmt.Sprintf("Error processing %s: %v", path, err))
			return nil
		}

		kbb.processedFiles++

		// Progress update
		if kbb.processedFiles%100 == 0 {
			log.Printf("Processed %d/%d files (%.1f%%)",
				kbb.processedFiles, kbb.totalFiles,
				float64(kbb.processedFiles)/float64(kbb.totalFiles)*100)
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("walk directory: %w", err)
	}

	log.Printf("✅ Scanning complete: %d files processed", kbb.processedFiles)
	return nil
}

// isTargetFile checks if file should be processed
func (kbb *KnowledgeBaseBuilder) isTargetFile(ext string) bool {
	targetExts := map[string]bool{
		".md":   true,
		".txt":  true,
		".json": true,
		".yaml": true,
		".yml":  true,
		".go":   true,
		".js":   true,
		".ts":   true,
		".py":   true,
		".sh":   true,
		".bat":  true,
	}

	return targetExts[ext]
}

// processFile processes a single file
func (kbb *KnowledgeBaseBuilder) processFile(sourcePath string) error {
	// Determine target category
	category := kbb.categorizeFile(sourcePath)

	// Create target path
	relPath, err := filepath.Rel(kbb.sourceDir, sourcePath)
	if err != nil {
		return fmt.Errorf("get relative path: %w", err)
	}

	targetPath := filepath.Join(kbb.targetDir, category, relPath)

	// Create target directory
	if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
		return fmt.Errorf("create target dir: %w", err)
	}

	// Copy file
	if err := kbb.copyFile(sourcePath, targetPath); err != nil {
		return fmt.Errorf("copy file: %w", err)
	}

	return nil
}

// categorizeFile determines file category
func (kbb *KnowledgeBaseBuilder) categorizeFile(path string) string {
	path = strings.ToLower(path)

	// Check for specific patterns
	if strings.Contains(path, "skill") || strings.Contains(path, "learning") {
		return "content/skills"
	}
	if strings.Contains(path, "agent") || strings.Contains(path, "ai") {
		return "content/agents"
	}
	if strings.Contains(path, "rule") || strings.Contains(path, "config") {
		return "content/rules"
	}
	if strings.Contains(path, "workflow") || strings.Contains(path, "process") {
		return "content/workflows"
	}
	if strings.Contains(path, "app") || strings.Contains(path, "apps") {
		return "apps-docs"
	}
	if strings.Contains(path, "package") || strings.Contains(path, "lib") {
		return "packages"
	}

	// Default based on location
	if strings.Contains(path, "docs") || strings.Contains(path, "readme") {
		return "docs"
	}
	if strings.Contains(path, "content") {
		return "content"
	}

	// Default to docs
	return "docs"
}

// copyFile copies a file
func (kbb *KnowledgeBaseBuilder) copyFile(src, dst string) error {
	// Open source file
	sourceFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("open source: %w", err)
	}
	defer sourceFile.Close()

	// Create destination file
	destFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("create dest: %w", err)
	}
	defer destFile.Close()

	// Copy content
	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return fmt.Errorf("copy content: %w", err)
	}

	return nil
}

// generateIndex generates knowledge base index
func (kbb *KnowledgeBaseBuilder) generateIndex() error {
	log.Println("📋 Generating knowledge base index...")

	indexPath := filepath.Join(kbb.targetDir, "KNOWLEDGE_INDEX.md")

	file, err := os.Create(indexPath)
	if err != nil {
		return fmt.Errorf("create index: %w", err)
	}
	defer file.Close()

	// Write index content
	writer := bufio.NewWriter(file)

	fmt.Fprintf(writer, "# Ti Ecosystem Knowledge Base Index\n\n")
	fmt.Fprintf(writer, "**Generated**: %s\n", time.Now().Format("2006-01-02 15:04:05"))
	fmt.Fprintf(writer, "**Total Files**: %d\n", kbb.processedFiles)
	fmt.Fprintf(writer, "**Build Time**: %s\n\n", time.Since(kbb.startTime))

	// Write category sections
	categories := []string{
		"content/skills",
		"content/agents",
		"content/rules",
		"content/workflows",
		"apps-docs",
		"packages",
		"docs",
	}

	for _, category := range categories {
		fmt.Fprintf(writer, "## %s\n\n", strings.ToUpper(strings.Replace(category, "/", " - ", -1)))

		categoryPath := filepath.Join(kbb.targetDir, category)
		if err := kbb.writeCategoryIndex(writer, categoryPath, category); err != nil {
			log.Printf("Error writing category %s: %v", category, err)
		}
	}

	// Write errors if any
	if len(kbb.errors) > 0 {
		fmt.Fprintf(writer, "## Errors\n\n")
		for _, err := range kbb.errors {
			fmt.Fprintf(writer, "- %s\n", err)
		}
	}

	if err := writer.Flush(); err != nil {
		return fmt.Errorf("flush writer: %w", err)
	}

	log.Println("✅ Index generated")
	return nil
}

// writeCategoryIndex writes index for a category
func (kbb *KnowledgeBaseBuilder) writeCategoryIndex(writer *bufio.Writer, categoryPath, categoryName string) error {
	return filepath.Walk(categoryPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		if info.IsDir() {
			return nil
		}

		relPath, err := filepath.Rel(kbb.targetDir, path)
		if err != nil {
			return nil
		}

		fmt.Fprintf(writer, "- [%s](%s)\n", filepath.Base(path), relPath)
		return nil
	})
}

// printSummary prints build summary
func (kbb *KnowledgeBaseBuilder) printSummary() {
	log.Println("=== 📊 BUILD SUMMARY ===")
	log.Printf("Total Files Found: %d", kbb.totalFiles)
	log.Printf("Files Processed: %d", kbb.processedFiles)
	log.Printf("Errors: %d", len(kbb.errors))
	log.Printf("Build Time: %s", time.Since(kbb.startTime))

	if len(kbb.errors) > 0 {
		log.Println("⚠️  Errors occurred during build:")
		for _, err := range kbb.errors[:min(10, len(kbb.errors))] {
			log.Printf("  - %s", err)
		}
		if len(kbb.errors) > 10 {
			log.Printf("  ... and %d more errors", len(kbb.errors)-10)
		}
	}

	log.Println("🎉 Knowledge base build completed!")
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
