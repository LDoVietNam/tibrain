package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type OmniRouteKnowledgePack struct {
	Meta      map[string]interface{}     `json:"_meta"`
	Providers []OmniRouteKnowledgeRecord `json:"providers"`
}

type OmniRouteKnowledgeRecord struct {
	ID          string                           `json:"id"`
	Name        string                           `json:"name"`
	Alias       string                           `json:"alias"`
	Auth        string                           `json:"auth"`
	ModelsCount int                              `json:"modelsCount"`
	Models      []OmniRouteKnowledgeProviderModel `json:"models"`
}

type OmniRouteKnowledgeProviderModel struct {
	ID        string `json:"id"`
	Name      string `json:"name,omitempty"`
	Ctx       int    `json:"ctx,omitempty"`
	Vision    bool   `json:"vision,omitempty"`
	Tools     bool   `json:"tools,omitempty"`
	Reasoning bool   `json:"reasoning,omitempty"`
}

type OmniRouteKnowledgeImporter struct {
	ragManager *RAGSystemManager
}

func NewOmniRouteKnowledgeImporter(ragManager *RAGSystemManager) *OmniRouteKnowledgeImporter {
	return &OmniRouteKnowledgeImporter{ragManager: ragManager}
}

func (i *OmniRouteKnowledgeImporter) ImportJSONFile(filePath string) (int, error) {
	raw, err := os.ReadFile(filePath)
	if err != nil {
		return 0, fmt.Errorf("read knowledge pack: %w", err)
	}

	var pack OmniRouteKnowledgePack
	if err := json.Unmarshal(raw, &pack); err != nil {
		return 0, fmt.Errorf("parse knowledge pack: %w", err)
	}

	baseName := strings.TrimSuffix(filepath.Base(filePath), filepath.Ext(filePath))
	imported := 0

	for _, provider := range pack.Providers {
		doc := i.providerToDocument(baseName, filePath, provider, pack.Meta)
		if err := i.ragManager.AddDocument(doc); err != nil {
			logger.Warn("Failed to import OmniRoute provider %s: %v", provider.ID, err)
			continue
		}
		imported++
	}

	if len(pack.Providers) > 0 {
		summaryDoc := i.summaryToDocument(baseName, filePath, pack)
		if err := i.ragManager.AddDocument(summaryDoc); err != nil {
			logger.Warn("Failed to import OmniRoute summary doc: %v", err)
		} else {
			imported++
		}
	}

	logger.Info("Imported %d OmniRoute knowledge documents from %s", imported, filePath)
	return imported, nil
}

func (i *OmniRouteKnowledgeImporter) ImportRepoDocuments(rootPath string, maxDocs int) (int, error) {
	if maxDocs <= 0 {
		maxDocs = 12
	}

	candidatePaths := []string{
		"README.md",
		"CLAUDE.md",
		"RTK_INTEGRATION.md",
		"IMPLEMENTATION_CHECKLIST.md",
		"SPEC.md",
		"PLAN.md",
		"TODO.md",
		"TEST_SPECIFICATION.md",
		"CHANGELOG.md",
	}

	var docs []string
	for _, rel := range candidatePaths {
		fullPath := filepath.Join(rootPath, rel)
		if fileExists(fullPath) {
			docs = append(docs, fullPath)
		}
	}

	docsDir := filepath.Join(rootPath, "docs")
	if fileExists(docsDir) {
		_ = filepath.Walk(docsDir, func(path string, info os.FileInfo, err error) error {
			if err != nil || info == nil || info.IsDir() {
				return nil
			}
			ext := strings.ToLower(filepath.Ext(path))
			if ext == ".md" || ext == ".txt" || ext == ".json" {
				docs = append(docs, path)
			}
			return nil
		})
	}

	docs = uniqueImportStrings(docs)
	sort.Strings(docs)
	if len(docs) > maxDocs {
		docs = docs[:maxDocs]
	}

	imported := 0
	for _, docPath := range docs {
		if err := i.importSingleRepoDocument(rootPath, docPath); err != nil {
			logger.Warn("Failed to import OmniRoute repo doc %s: %v", docPath, err)
			continue
		}
		imported++
	}

	logger.Info("Imported %d OmniRoute repo documents from %s", imported, rootPath)
	return imported, nil
}

func (i *OmniRouteKnowledgeImporter) importSingleRepoDocument(rootPath, docPath string) error {
	raw, err := os.ReadFile(docPath)
	if err != nil {
		return fmt.Errorf("read repo doc: %w", err)
	}

	relPath, _ := filepath.Rel(rootPath, docPath)
	title := strings.TrimSuffix(filepath.Base(docPath), filepath.Ext(docPath))
	content := strings.TrimSpace(string(raw))
	if content == "" {
		return fmt.Errorf("empty repo doc")
	}

	category := "router_knowledge"
	tags := []string{"omniroute", "repo-doc", "llm-router"}
	lowerRel := strings.ToLower(relPath)
	if strings.Contains(lowerRel, "changelog") {
		tags = append(tags, "release-notes")
	}
	if strings.Contains(lowerRel, "rtk") {
		tags = append(tags, "rtk")
	}
	if strings.Contains(lowerRel, "docs") {
		tags = append(tags, "documentation")
	}

	metadata, _ := json.Marshal(map[string]interface{}{
		"source_system":  "OmniRoute",
		"source_type":    "repo_document",
		"repo_root":      rootPath,
		"relative_path":  relPath,
		"original_path":  docPath,
		"document_title": title,
	})

	now := time.Now()
	doc := RAGDocument{
		ID:        generateContentHash(docPath + "::repo-doc"),
		Title:     fmt.Sprintf("OmniRoute Doc %s", title),
		Content:   content,
		Path:      docPath,
		Category:  category,
		Tags:      strings.Join(tags, ","),
		Status:    "active",
		CreatedAt: now,
		UpdatedAt: now,
		Metadata:  string(metadata),
	}

	return i.ragManager.AddDocument(doc)
}

func (i *OmniRouteKnowledgeImporter) providerToDocument(baseName, filePath string, provider OmniRouteKnowledgeRecord, meta map[string]interface{}) RAGDocument {
	now := time.Now()
	tags := []string{"omniroute", "provider-routing", "llm-router"}
	if provider.Auth != "" {
		tags = append(tags, provider.Auth)
	}

	metadata := map[string]interface{}{
		"source_system": "OmniRoute",
		"source_type":   "provider_knowledge_pack",
		"provider_id":   provider.ID,
		"provider_alias": provider.Alias,
		"models_count":  provider.ModelsCount,
		"file_path":     filePath,
		"pack_name":     baseName,
		"pack_meta":     meta,
	}
	metadataJSON, _ := json.Marshal(metadata)

	var sections []string
	sections = append(sections, fmt.Sprintf("Provider: %s", safeDisplay(provider.Name, provider.ID)))
	sections = append(sections, fmt.Sprintf("Provider ID: %s", provider.ID))
	if provider.Alias != "" {
		sections = append(sections, fmt.Sprintf("Alias: %s", provider.Alias))
	}
	if provider.Auth != "" {
		sections = append(sections, fmt.Sprintf("Auth: %s", provider.Auth))
	}
	sections = append(sections, fmt.Sprintf("Models count: %d", provider.ModelsCount))
	sections = append(sections, "")
	sections = append(sections, "Supported models:")

	for _, model := range provider.Models {
		capabilities := make([]string, 0, 3)
		if model.Vision {
			capabilities = append(capabilities, "vision")
		}
		if model.Tools {
			capabilities = append(capabilities, "tools")
		}
		if model.Reasoning {
			capabilities = append(capabilities, "reasoning")
		}

		line := fmt.Sprintf("- %s", model.ID)
		if model.Name != "" && model.Name != model.ID {
			line += fmt.Sprintf(" (%s)", model.Name)
		}
		if model.Ctx > 0 {
			line += fmt.Sprintf(" | ctx=%d", model.Ctx)
		}
		if len(capabilities) > 0 {
			line += fmt.Sprintf(" | caps=%s", strings.Join(capabilities, ","))
		}
		sections = append(sections, line)
	}

	return RAGDocument{
		ID:        generateContentHash(filePath + "::provider::" + provider.ID),
		Title:     fmt.Sprintf("OmniRoute Provider %s", safeDisplay(provider.Name, provider.ID)),
		Content:   strings.Join(sections, "\n"),
		Path:      filePath,
		Category:  "router_knowledge",
		Tags:      strings.Join(tags, ","),
		Status:    "active",
		CreatedAt: now,
		UpdatedAt: now,
		Metadata:  string(metadataJSON),
	}
}

func (i *OmniRouteKnowledgeImporter) summaryToDocument(baseName, filePath string, pack OmniRouteKnowledgePack) RAGDocument {
	now := time.Now()
	metaJSON, _ := json.Marshal(map[string]interface{}{
		"source_system": "OmniRoute",
		"source_type":   "provider_knowledge_summary",
		"file_path":     filePath,
		"pack_name":     baseName,
		"providers_count": len(pack.Providers),
		"pack_meta":     pack.Meta,
	})

	lines := []string{
		"OmniRoute provider knowledge summary",
		fmt.Sprintf("Providers count: %d", len(pack.Providers)),
		"",
		"Providers:",
	}

	for _, provider := range pack.Providers {
		lines = append(lines, fmt.Sprintf("- %s | id=%s | alias=%s | auth=%s | models=%d",
			safeDisplay(provider.Name, provider.ID), provider.ID, provider.Alias, provider.Auth, provider.ModelsCount))
	}

	return RAGDocument{
		ID:        generateContentHash(filePath + "::summary"),
		Title:     fmt.Sprintf("OmniRoute Knowledge Summary %s", baseName),
		Content:   strings.Join(lines, "\n"),
		Path:      filePath,
		Category:  "router_knowledge",
		Tags:      "omniroute,summary,provider-routing,llm-router",
		Status:    "active",
		CreatedAt: now,
		UpdatedAt: now,
		Metadata:  string(metaJSON),
	}
}

func safeDisplay(primary, fallback string) string {
	if strings.TrimSpace(primary) != "" {
		return primary
	}
	return fallback
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func uniqueImportStrings(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	out := make([]string, 0, len(values))
	for _, value := range values {
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	return out
}
