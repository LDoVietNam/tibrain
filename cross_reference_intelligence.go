package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// CrossReferenceIntelligence provides intelligent cross-referencing between documents
type CrossReferenceIntelligence struct {
	db               *sql.DB
	linkAnalyzer     *LinkAnalyzer
	versionTracker   *VersionTracker
	dependencyMapper *DependencyMapper
	impactAnalyzer   *ImpactAnalyzer
}

// LinkAnalyzer analyzes links between documents
type LinkAnalyzer struct {
	internalLinks map[string][]string
	externalLinks map[string][]string
	brokenLinks   map[string][]string
	linkTypes     map[string]map[string]int
	linkStrength  map[string]map[string]float64
}

// VersionTracker tracks document versions
type VersionTracker struct {
	versions        map[string][]DocumentVersion
	currentVersions map[string]string
	versionHistory  map[string][]VersionChange
	branchVersions  map[string]map[string]string
}

// DependencyMapper maps dependencies between documents
type DependencyMapper struct {
	dependencies     map[string][]string
	dependents       map[string][]string
	circularDeps     []string
	dependencyLevels map[string]int
	dependencyGraph  *DependencyGraph
}

// ImpactAnalyzer analyzes impact of changes
type ImpactAnalyzer struct {
	changeImpact   map[string][]string
	impactScore    map[string]float64
	criticalPaths  []string
	riskAssessment map[string]RiskLevel
}

// DocumentVersion represents a document version
type DocumentVersion struct {
	ID        string    `json:"id"`
	Version   string    `json:"version"`
	Timestamp time.Time `json:"timestamp"`
	Author    string    `json:"author"`
	Changes   []string  `json:"changes"`
	Checksum  string    `json:"checksum"`
	Size      int64     `json:"size"`
}

// VersionChange represents a version change
type VersionChange struct {
	Type        string    `json:"type"`
	Description string    `json:"description"`
	Timestamp   time.Time `json:"timestamp"`
	Impact      string    `json:"impact"`
}

// DependencyGraph represents the dependency graph
type DependencyGraph struct {
	Nodes []DependencyNode `json:"nodes"`
	Edges []DependencyEdge `json:"edges"`
}

// DependencyNode represents a node in the dependency graph
type DependencyNode struct {
	ID       string                 `json:"id"`
	Name     string                 `json:"name"`
	Type     string                 `json:"type"`
	Level    int                    `json:"level"`
	Metadata map[string]interface{} `json:"metadata"`
}

// DependencyEdge represents an edge in the dependency graph
type DependencyEdge struct {
	Source   string            `json:"source"`
	Target   string            `json:"target"`
	Type     string            `json:"type"`
	Strength float64           `json:"strength"`
	Metadata map[string]string `json:"metadata"`
}

// RiskLevel represents risk level for impact analysis
type RiskLevel struct {
	Level       string  `json:"level"`
	Score       float64 `json:"score"`
	Description string  `json:"description"`
	Mitigation  string  `json:"mitigation"`
}

// CrossReferenceReport represents comprehensive cross-reference report
type CrossReferenceReport struct {
	Timestamp         time.Time              `json:"timestamp"`
	LinkAnalysis      map[string]interface{} `json:"link_analysis"`
	VersionTracking   map[string]interface{} `json:"version_tracking"`
	DependencyMapping map[string]interface{} `json:"dependency_mapping"`
	ImpactAnalysis    map[string]interface{} `json:"impact_analysis"`
	Recommendations   []string               `json:"recommendations"`
	QualityMetrics    map[string]float64     `json:"quality_metrics"`
}

func NewCrossReferenceIntelligence(dbPath string) (*CrossReferenceIntelligence, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %v", err)
	}

	linkAnalyzer := &LinkAnalyzer{
		internalLinks: make(map[string][]string),
		externalLinks: make(map[string][]string),
		brokenLinks:   make(map[string][]string),
		linkTypes:     make(map[string]map[string]int),
		linkStrength:  make(map[string]map[string]float64),
	}

	versionTracker := &VersionTracker{
		versions:        make(map[string][]DocumentVersion),
		currentVersions: make(map[string]string),
		versionHistory:  make(map[string][]VersionChange),
		branchVersions:  make(map[string]map[string]string),
	}

	dependencyMapper := &DependencyMapper{
		dependencies:     make(map[string][]string),
		dependents:       make(map[string][]string),
		circularDeps:     []string{},
		dependencyLevels: make(map[string]int),
		dependencyGraph:  &DependencyGraph{},
	}

	impactAnalyzer := &ImpactAnalyzer{
		changeImpact:   make(map[string][]string),
		impactScore:    make(map[string]float64),
		criticalPaths:  []string{},
		riskAssessment: make(map[string]RiskLevel),
	}

	return &CrossReferenceIntelligence{
		db:               db,
		linkAnalyzer:     linkAnalyzer,
		versionTracker:   versionTracker,
		dependencyMapper: dependencyMapper,
		impactAnalyzer:   impactAnalyzer,
	}, nil
}

// InitializeCrossReferenceIntelligence sets up the cross-reference intelligence system
func (cri *CrossReferenceIntelligence) InitializeCrossReferenceIntelligence() error {
	fmt.Println("🔗 Initializing Cross-Reference Intelligence...")

	// 1. Create cross-reference tables
	err := cri.createCrossReferenceTables()
	if err != nil {
		return fmt.Errorf("failed to create cross-reference tables: %v", err)
	}

	// 2. Analyze existing links
	err = cri.analyzeExistingLinks()
	if err != nil {
		return fmt.Errorf("failed to analyze existing links: %v", err)
	}

	// 3. Initialize version tracking
	err = cri.initializeVersionTracking()
	if err != nil {
		return fmt.Errorf("failed to initialize version tracking: %v", err)
	}

	// 4. Map dependencies
	err = cri.mapDependencies()
	if err != nil {
		return fmt.Errorf("failed to map dependencies: %v", err)
	}

	// 5. Analyze impact
	err = cri.analyzeImpact()
	if err != nil {
		return fmt.Errorf("failed to analyze impact: %v", err)
	}

	fmt.Println("✅ Cross-Reference Intelligence initialized successfully!")
	return nil
}

// createCrossReferenceTables creates tables for cross-reference data
func (cri *CrossReferenceIntelligence) createCrossReferenceTables() error {
	fmt.Println("📊 Creating cross-reference tables...")

	tables := []string{
		`CREATE TABLE IF NOT EXISTS document_links (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			source_doc TEXT,
			target_doc TEXT,
			link_type TEXT,
			link_strength REAL,
			link_text TEXT,
			is_broken BOOLEAN,
			created_at TEXT DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS document_versions (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			doc_id TEXT,
			version TEXT,
			author TEXT,
			changes TEXT,
			checksum TEXT,
			size INTEGER,
			created_at TEXT DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS dependencies (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			source_doc TEXT,
			target_doc TEXT,
			dependency_type TEXT,
			strength REAL,
			level INTEGER,
			is_circular BOOLEAN,
			created_at TEXT DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS impact_analysis (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			doc_id TEXT,
			impact_score REAL,
			affected_docs TEXT,
			risk_level TEXT,
			critical_path BOOLEAN,
			created_at TEXT DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS cross_reference_metrics (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			metric_type TEXT,
			metric_value REAL,
			metadata TEXT,
			created_at TEXT DEFAULT CURRENT_TIMESTAMP
		)`,
	}

	for _, table := range tables {
		_, err := cri.db.Exec(table)
		if err != nil {
			return fmt.Errorf("failed to create table: %v", err)
		}
	}

	return nil
}

// analyzeExistingLinks analyzes links between documents
func (cri *CrossReferenceIntelligence) analyzeExistingLinks() error {
	fmt.Println("🔍 Analyzing existing links...")

	// Get all documents
	rows, err := cri.db.Query("SELECT id, content FROM rag_documents WHERE content IS NOT NULL AND content != ''")
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var id, content string
		err := rows.Scan(&id, &content)
		if err != nil {
			continue
		}

		// Extract links from content
		links := cri.extractLinks(content)

		for _, link := range links {
			// Determine link type
			linkType := cri.determineLinkType(link)

			// Calculate link strength
			strength := cri.calculateLinkStrength(link, content)

			// Store link
			_, err := cri.db.Exec(`
				INSERT INTO document_links (source_doc, target_doc, link_type, link_strength, link_text, is_broken)
				VALUES (?, ?, ?, ?, ?, ?)
			`, id, link.Target, linkType, strength, link.Text, link.IsBroken)
			if err != nil {
				continue
			}

			// Update internal tracking
			if link.IsInternal {
				cri.linkAnalyzer.internalLinks[id] = append(cri.linkAnalyzer.internalLinks[id], link.Target)
			} else {
				cri.linkAnalyzer.externalLinks[id] = append(cri.linkAnalyzer.externalLinks[id], link.Target)
			}

			if link.IsBroken {
				cri.linkAnalyzer.brokenLinks[id] = append(cri.linkAnalyzer.brokenLinks[id], link.Target)
			}

			// Update link types
			if cri.linkAnalyzer.linkTypes[id] == nil {
				cri.linkAnalyzer.linkTypes[id] = make(map[string]int)
			}
			cri.linkAnalyzer.linkTypes[id][linkType]++

			// Update link strength
			if cri.linkAnalyzer.linkStrength[id] == nil {
				cri.linkAnalyzer.linkStrength[id] = make(map[string]float64)
			}
			cri.linkAnalyzer.linkStrength[id][link.Target] = strength
		}
	}

	fmt.Println("✅ Link analysis completed")
	return nil
}

// Link represents a link in the content
type Link struct {
	Text       string
	Target     string
	IsInternal bool
	IsBroken   bool
	Type       string
}

// extractLinks extracts links from content
func (cri *CrossReferenceIntelligence) extractLinks(content string) []Link {
	var links []Link

	// Markdown links: [text](url)
	markdownLinks := regexp.MustCompile(`\[([^\]]+)\]\(([^)]+)\)`)
	matches := markdownLinks.FindAllStringSubmatch(content, -1)

	for _, match := range matches {
		if len(match) >= 3 {
			text := match[1]
			target := match[2]

			link := Link{
				Text:       text,
				Target:     target,
				IsInternal: cri.isInternalLink(target),
				IsBroken:   false, // Will be checked later
				Type:       "markdown",
			}

			links = append(links, link)
		}
	}

	// HTTP links: http:// or https://
	httpLinks := regexp.MustCompile(`https?://[^\s]+`)
	httpMatches := httpLinks.FindAllString(content, -1)

	for _, match := range httpMatches {
		link := Link{
			Text:       match,
			Target:     match,
			IsInternal: false,
			IsBroken:   false, // Will be checked later
			Type:       "http",
		}

		links = append(links, link)
	}

	// Reference links: @ref_file or @ref_snippet
	refLinks := regexp.MustCompile(`@ref_(file|snippet)\s+([^\s]+)`)
	matches = refLinks.FindAllStringSubmatch(content, -1)

	for _, match := range matches {
		if len(match) >= 3 {
			refType := match[1]
			target := match[2]

			link := Link{
				Text:       fmt.Sprintf("@ref_%s %s", refType, target),
				Target:     target,
				IsInternal: true,
				IsBroken:   false, // Will be checked later
				Type:       fmt.Sprintf("ref_%s", refType),
			}

			links = append(links, link)
		}
	}

	return links
}

// isInternalLink checks if link is internal
func (cri *CrossReferenceIntelligence) isInternalLink(target string) bool {
	// Check if it's a relative path
	return !strings.HasPrefix(target, "http://") && !strings.HasPrefix(target, "https://")
}

// determineLinkType determines the type of link
func (cri *CrossReferenceIntelligence) determineLinkType(link Link) string {
	if link.Type != "" {
		return link.Type
	}

	// Determine based on target
	if strings.HasSuffix(link.Target, ".md") {
		return "markdown_file"
	}
	if strings.HasSuffix(link.Target, ".go") {
		return "code_file"
	}
	if strings.HasSuffix(link.Target, ".json") {
		return "config_file"
	}
	if strings.Contains(link.Target, "github.com") {
		return "github"
	}
	if strings.Contains(link.Target, "docs.") {
		return "documentation"
	}

	return "unknown"
}

// calculateLinkStrength calculates the strength of a link
func (cri *CrossReferenceIntelligence) calculateLinkStrength(link Link, content string) float64 {
	strength := 0.5 // Base strength

	// Increase strength based on context
	if strings.Contains(content, "important") && strings.Contains(content, link.Text) {
		strength += 0.2
	}
	if strings.Contains(content, "critical") && strings.Contains(content, link.Text) {
		strength += 0.3
	}
	if strings.Contains(content, "see also") && strings.Contains(content, link.Text) {
		strength += 0.1
	}
	if strings.Contains(content, "reference") && strings.Contains(content, link.Text) {
		strength += 0.15
	}

	// Adjust based on link type
	switch link.Type {
	case "ref_file", "ref_snippet":
		strength += 0.2
	case "markdown_file":
		strength += 0.1
	case "code_file":
		strength += 0.15
	}

	// Cap at 1.0
	if strength > 1.0 {
		strength = 1.0
	}

	return strength
}

// initializeVersionTracking initializes version tracking
func (cri *CrossReferenceIntelligence) initializeVersionTracking() error {
	fmt.Println("📝 Initializing version tracking...")

	// Get all documents
	rows, err := cri.db.Query("SELECT id, content, created_at FROM rag_documents")
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var id, content, createdAt string
		err := rows.Scan(&id, &content, &createdAt)
		if err != nil {
			continue
		}

		// Create initial version
		version := DocumentVersion{
			ID:        fmt.Sprintf("%s_v1", id),
			Version:   "1.0.0",
			Timestamp: time.Now(),
			Author:    "system",
			Changes:   []string{"Initial version"},
			Checksum:  cri.calculateChecksum(content),
			Size:      int64(len(content)),
		}

		// Store version
		cri.versionTracker.versions[id] = append(cri.versionTracker.versions[id], version)
		cri.versionTracker.currentVersions[id] = version.Version

		// Save to database
		changesJSON, _ := json.Marshal(version.Changes)
		_, err = cri.db.Exec(`
			INSERT INTO document_versions (doc_id, version, author, changes, checksum, size)
			VALUES (?, ?, ?, ?, ?, ?)
		`, id, version.Version, version.Author, string(changesJSON), version.Checksum, version.Size)
		if err != nil {
			continue
		}
	}

	fmt.Println("✅ Version tracking initialized")
	return nil
}

// calculateChecksum calculates checksum for content
func (cri *CrossReferenceIntelligence) calculateChecksum(content string) string {
	// Simple checksum calculation (in production, use proper hash)
	return fmt.Sprintf("%x", len(content)*12345)
}

// mapDependencies maps dependencies between documents
func (cri *CrossReferenceIntelligence) mapDependencies() error {
	fmt.Println("🗺️ Mapping dependencies...")

	// Build dependency graph from links
	for source, targets := range cri.linkAnalyzer.internalLinks {
		for _, target := range targets {
			// Add dependency
			cri.dependencyMapper.dependencies[source] = append(cri.dependencyMapper.dependencies[source], target)
			cri.dependencyMapper.dependents[target] = append(cri.dependencyMapper.dependents[target], source)

			// Determine dependency type
			depType := cri.determineDependencyType(source, target)

			// Calculate strength
			strength := cri.linkAnalyzer.linkStrength[source][target]

			// Save to database
			_, err := cri.db.Exec(`
				INSERT INTO dependencies (source_doc, target_doc, dependency_type, strength, level, is_circular)
				VALUES (?, ?, ?, ?, ?, ?)
			`, source, target, depType, strength, 0, false)
			if err != nil {
				continue
			}
		}
	}

	// Calculate dependency levels
	cri.calculateDependencyLevels()

	// Detect circular dependencies
	cri.detectCircularDependencies()

	// Build dependency graph
	cri.buildDependencyGraph()

	fmt.Println("✅ Dependency mapping completed")
	return nil
}

// determineDependencyType determines the type of dependency
func (cri *CrossReferenceIntelligence) determineDependencyType(source, target string) string {
	// Simple logic based on file extensions
	if strings.HasSuffix(target, ".md") {
		return "documentation"
	}
	if strings.HasSuffix(target, ".go") {
		return "code"
	}
	if strings.HasSuffix(target, ".json") {
		return "configuration"
	}
	if strings.Contains(target, "config") {
		return "configuration"
	}
	if strings.Contains(target, "setup") {
		return "setup"
	}

	return "general"
}

// calculateDependencyLevels calculates dependency levels
func (cri *CrossReferenceIntelligence) calculateDependencyLevels() {
	// Simple level calculation based on dependency depth
	visited := make(map[string]bool)

	var calculateLevel func(string) int
	calculateLevel = func(doc string) int {
		if visited[doc] {
			return 0 // Avoid infinite recursion
		}
		visited[doc] = true

		maxLevel := 0
		for _, dep := range cri.dependencyMapper.dependencies[doc] {
			level := calculateLevel(dep) + 1
			if level > maxLevel {
				maxLevel = level
			}
		}

		cri.dependencyMapper.dependencyLevels[doc] = maxLevel
		return maxLevel
	}

	// Calculate levels for all documents
	for doc := range cri.dependencyMapper.dependencies {
		calculateLevel(doc)
	}
}

// detectCircularDependencies detects circular dependencies
func (cri *CrossReferenceIntelligence) detectCircularDependencies() {
	visited := make(map[string]bool)
	recStack := make(map[string]bool)

	var dfs func(string) bool
	dfs = func(doc string) bool {
		visited[doc] = true
		recStack[doc] = true

		for _, dep := range cri.dependencyMapper.dependencies[doc] {
			if !visited[dep] && dfs(dep) {
				return true
			} else if recStack[dep] {
				// Found circular dependency
				circular := fmt.Sprintf("%s -> %s", doc, dep)
				cri.dependencyMapper.circularDeps = append(cri.dependencyMapper.circularDeps, circular)
				return true
			}
		}

		recStack[doc] = false
		return false
	}

	// Check all documents
	for doc := range cri.dependencyMapper.dependencies {
		if !visited[doc] {
			dfs(doc)
		}
	}
}

// buildDependencyGraph builds the dependency graph
func (cri *CrossReferenceIntelligence) buildDependencyGraph() {
	graph := &DependencyGraph{
		Nodes: []DependencyNode{},
		Edges: []DependencyEdge{},
	}

	// Add nodes
	for doc := range cri.dependencyMapper.dependencies {
		node := DependencyNode{
			ID:    doc,
			Name:  doc,
			Type:  cri.determineDependencyType(doc, ""),
			Level: cri.dependencyMapper.dependencyLevels[doc],
			Metadata: map[string]interface{}{
				"dependencies_count": len(cri.dependencyMapper.dependencies[doc]),
				"dependents_count":   len(cri.dependencyMapper.dependents[doc]),
			},
		}
		graph.Nodes = append(graph.Nodes, node)
	}

	// Add edges
	for source, targets := range cri.dependencyMapper.dependencies {
		for _, target := range targets {
			edge := DependencyEdge{
				Source:   source,
				Target:   target,
				Type:     cri.determineDependencyType(source, target),
				Strength: cri.linkAnalyzer.linkStrength[source][target],
				Metadata: map[string]string{
					"level": fmt.Sprintf("%d", cri.dependencyMapper.dependencyLevels[source]),
				},
			}
			graph.Edges = append(graph.Edges, edge)
		}
	}

	cri.dependencyMapper.dependencyGraph = graph
}

// analyzeImpact analyzes impact of changes
func (cri *CrossReferenceIntelligence) analyzeImpact() error {
	fmt.Println("💥 Analyzing impact...")

	// Analyze impact for each document
	for doc := range cri.dependencyMapper.dependencies {
		impact := cri.calculateImpact(doc)

		// Store impact analysis
		affectedDocsJSON, _ := json.Marshal(impact.AffectedDocs)
		_, err := cri.db.Exec(`
			INSERT INTO impact_analysis (doc_id, impact_score, affected_docs, risk_level, critical_path)
			VALUES (?, ?, ?, ?, ?)
		`, doc, impact.Score, string(affectedDocsJSON), impact.RiskLevel.Level, impact.IsCriticalPath)
		if err != nil {
			continue
		}

		cri.impactAnalyzer.changeImpact[doc] = impact.AffectedDocs
		cri.impactAnalyzer.impactScore[doc] = impact.Score
		cri.impactAnalyzer.riskAssessment[doc] = impact.RiskLevel

		if impact.IsCriticalPath {
			cri.impactAnalyzer.criticalPaths = append(cri.impactAnalyzer.criticalPaths, doc)
		}
	}

	fmt.Println("✅ Impact analysis completed")
	return nil
}

// Impact represents impact analysis result
type Impact struct {
	Score          float64   `json:"score"`
	AffectedDocs   []string  `json:"affected_docs"`
	RiskLevel      RiskLevel `json:"risk_level"`
	IsCriticalPath bool      `json:"is_critical_path"`
}

// calculateImpact calculates impact score for a document
func (cri *CrossReferenceIntelligence) calculateImpact(doc string) Impact {
	impact := Impact{
		Score:          0.0,
		AffectedDocs:   []string{},
		RiskLevel:      RiskLevel{Level: "low", Score: 0.1, Description: "Low impact", Mitigation: "Standard review"},
		IsCriticalPath: false,
	}

	// Get all dependents
	dependents := cri.dependencyMapper.dependents[doc]
	impact.AffectedDocs = append(impact.AffectedDocs, dependents...)

	// Calculate impact score based on:
	// 1. Number of dependents
	// 2. Dependency levels
	// 3. Link strengths
	// 4. Critical path involvement

	dependentCount := len(dependents)
	impact.Score += float64(dependentCount) * 0.1

	// Add score based on dependency levels
	for _, dependent := range dependents {
		level := cri.dependencyMapper.dependencyLevels[dependent]
		impact.Score += float64(level) * 0.05
	}

	// Add score based on link strengths
	for _, dependent := range dependents {
		strength := cri.linkAnalyzer.linkStrength[doc][dependent]
		impact.Score += strength * 0.2
	}

	// Determine risk level
	if impact.Score > 0.8 {
		impact.RiskLevel = RiskLevel{
			Level:       "critical",
			Score:       impact.Score,
			Description: "Critical impact - affects many components",
			Mitigation:  "Requires thorough testing and review",
		}
		impact.IsCriticalPath = true
	} else if impact.Score > 0.5 {
		impact.RiskLevel = RiskLevel{
			Level:       "high",
			Score:       impact.Score,
			Description: "High impact - affects important components",
			Mitigation:  "Requires careful review and testing",
		}
	} else if impact.Score > 0.3 {
		impact.RiskLevel = RiskLevel{
			Level:       "medium",
			Score:       impact.Score,
			Description: "Medium impact - affects some components",
			Mitigation:  "Requires standard review",
		}
	} else {
		impact.RiskLevel = RiskLevel{
			Level:       "low",
			Score:       impact.Score,
			Description: "Low impact - minimal effects",
			Mitigation:  "Standard review sufficient",
		}
	}

	return impact
}

// GenerateCrossReferenceReport generates comprehensive cross-reference report
func (cri *CrossReferenceIntelligence) GenerateCrossReferenceReport() error {
	fmt.Println("📋 Generating cross-reference report...")

	report := CrossReferenceReport{
		Timestamp: time.Now(),
		LinkAnalysis: map[string]interface{}{
			"total_links":    cri.getTotalLinks(),
			"internal_links": len(cri.linkAnalyzer.internalLinks),
			"external_links": len(cri.linkAnalyzer.externalLinks),
			"broken_links":   len(cri.linkAnalyzer.brokenLinks),
			"link_types":     cri.getLinkTypeStats(),
			"link_strengths": cri.getLinkStrengthStats(),
		},
		VersionTracking: map[string]interface{}{
			"total_versions":   cri.getTotalVersions(),
			"current_versions": len(cri.versionTracker.currentVersions),
			"version_history":  cri.getVersionHistoryStats(),
		},
		DependencyMapping: map[string]interface{}{
			"total_dependencies":    cri.getTotalDependencies(),
			"circular_dependencies": len(cri.dependencyMapper.circularDeps),
			"dependency_levels":     cri.getDependencyLevelStats(),
			"dependency_graph":      cri.dependencyMapper.dependencyGraph,
		},
		ImpactAnalysis: map[string]interface{}{
			"total_impacts":   len(cri.impactAnalyzer.changeImpact),
			"critical_paths":  len(cri.impactAnalyzer.criticalPaths),
			"risk_assessment": cri.getRiskAssessmentStats(),
			"impact_scores":   cri.getImpactScoreStats(),
		},
		Recommendations: cri.generateRecommendations(),
		QualityMetrics:  cri.calculateQualityMetrics(),
	}

	// Save report
	reportPath := filepath.Join("Z:\\01_PROJECTS\\tibrain\\reports", "cross_reference_report.json")
	os.MkdirAll(filepath.Dir(reportPath), 0755)

	reportData, _ := json.MarshalIndent(report, "", "  ")
	err := os.WriteFile(reportPath, reportData, 0644)
	if err != nil {
		return fmt.Errorf("failed to save report: %v", err)
	}

	fmt.Printf("✅ Cross-reference report saved to: %s\n", reportPath)
	return nil
}

// Helper methods for statistics
func (cri *CrossReferenceIntelligence) getTotalLinks() int {
	var count int
	cri.db.QueryRow("SELECT COUNT(*) FROM document_links").Scan(&count)
	return count
}

func (cri *CrossReferenceIntelligence) getLinkTypeStats() map[string]int {
	stats := make(map[string]int)

	rows, err := cri.db.Query("SELECT link_type, COUNT(*) FROM document_links GROUP BY link_type")
	if err != nil {
		return stats
	}
	defer rows.Close()

	for rows.Next() {
		var linkType string
		var count int
		rows.Scan(&linkType, &count)
		stats[linkType] = count
	}

	return stats
}

func (cri *CrossReferenceIntelligence) getLinkStrengthStats() map[string]interface{} {
	stats := make(map[string]interface{})

	var avgStrength, maxStrength, minStrength float64
	cri.db.QueryRow("SELECT AVG(link_strength), MAX(link_strength), MIN(link_strength) FROM document_links").Scan(&avgStrength, &maxStrength, &minStrength)

	stats["average"] = avgStrength
	stats["maximum"] = maxStrength
	stats["minimum"] = minStrength

	return stats
}

func (cri *CrossReferenceIntelligence) getTotalVersions() int {
	var count int
	cri.db.QueryRow("SELECT COUNT(*) FROM document_versions").Scan(&count)
	return count
}

func (cri *CrossReferenceIntelligence) getVersionHistoryStats() map[string]interface{} {
	stats := make(map[string]interface{})

	rows, err := cri.db.Query("SELECT doc_id, COUNT(*) FROM document_versions GROUP BY doc_id")
	if err != nil {
		return stats
	}
	defer rows.Close()

	versionCounts := make(map[string]int)
	for rows.Next() {
		var docID string
		var count int
		rows.Scan(&docID, &count)
		versionCounts[docID] = count
	}

	stats["version_counts"] = versionCounts
	stats["avg_versions_per_doc"] = float64(len(versionCounts)) / float64(len(versionCounts))

	return stats
}

func (cri *CrossReferenceIntelligence) getTotalDependencies() int {
	var count int
	cri.db.QueryRow("SELECT COUNT(*) FROM dependencies").Scan(&count)
	return count
}

func (cri *CrossReferenceIntelligence) getDependencyLevelStats() map[string]interface{} {
	stats := make(map[string]interface{})

	rows, err := cri.db.Query("SELECT level, COUNT(*) FROM dependencies GROUP BY level")
	if err != nil {
		return stats
	}
	defer rows.Close()

	levelCounts := make(map[string]int)
	for rows.Next() {
		var level int
		var count int
		rows.Scan(&level, &count)
		levelCounts[fmt.Sprintf("level_%d", level)] = count
	}

	stats["level_distribution"] = levelCounts

	return stats
}

func (cri *CrossReferenceIntelligence) getRiskAssessmentStats() map[string]interface{} {
	stats := make(map[string]interface{})

	rows, err := cri.db.Query("SELECT risk_level, COUNT(*) FROM impact_analysis GROUP BY risk_level")
	if err != nil {
		return stats
	}
	defer rows.Close()

	riskCounts := make(map[string]int)
	for rows.Next() {
		var riskLevel string
		var count int
		rows.Scan(&riskLevel, &count)
		riskCounts[riskLevel] = count
	}

	stats["risk_distribution"] = riskCounts

	return stats
}

func (cri *CrossReferenceIntelligence) getImpactScoreStats() map[string]interface{} {
	stats := make(map[string]interface{})

	var avgImpact, maxImpact, minImpact float64
	cri.db.QueryRow("SELECT AVG(impact_score), MAX(impact_score), MIN(impact_score) FROM impact_analysis").Scan(&avgImpact, &maxImpact, &minImpact)

	stats["average"] = avgImpact
	stats["maximum"] = maxImpact
	stats["minimum"] = minImpact

	return stats
}

func (cri *CrossReferenceIntelligence) generateRecommendations() []string {
	var recommendations []string

	// Link quality recommendations
	if len(cri.linkAnalyzer.brokenLinks) > 0 {
		recommendations = append(recommendations, fmt.Sprintf("Fix %d broken links", len(cri.linkAnalyzer.brokenLinks)))
	}

	// Dependency recommendations
	if len(cri.dependencyMapper.circularDeps) > 0 {
		recommendations = append(recommendations, fmt.Sprintf("Resolve %d circular dependencies", len(cri.dependencyMapper.circularDeps)))
	}

	// Impact recommendations
	if len(cri.impactAnalyzer.criticalPaths) > 0 {
		recommendations = append(recommendations, fmt.Sprintf("Review %d critical path documents", len(cri.impactAnalyzer.criticalPaths)))
	}

	// General recommendations
	recommendations = append(recommendations, "Implement regular link validation")
	recommendations = append(recommendations, "Set up automated dependency monitoring")
	recommendations = append(recommendations, "Create impact assessment workflow")

	return recommendations
}

func (cri *CrossReferenceIntelligence) calculateQualityMetrics() map[string]float64 {
	metrics := make(map[string]float64)

	// Link quality
	totalLinks := cri.getTotalLinks()
	brokenLinks := len(cri.linkAnalyzer.brokenLinks)
	if totalLinks > 0 {
		metrics["link_quality"] = 1.0 - (float64(brokenLinks) / float64(totalLinks))
	} else {
		metrics["link_quality"] = 1.0
	}

	// Dependency quality
	totalDeps := cri.getTotalDependencies()
	circularDeps := len(cri.dependencyMapper.circularDeps)
	if totalDeps > 0 {
		metrics["dependency_quality"] = 1.0 - (float64(circularDeps) / float64(totalDeps))
	} else {
		metrics["dependency_quality"] = 1.0
	}

	// Version quality
	totalVersions := cri.getTotalVersions()
	totalDocs := len(cri.versionTracker.currentVersions)
	if totalDocs > 0 {
		metrics["version_quality"] = float64(totalVersions) / float64(totalDocs)
	} else {
		metrics["version_quality"] = 0.0
	}

	// Overall quality
	metrics["overall_quality"] = (metrics["link_quality"] + metrics["dependency_quality"] + metrics["version_quality"]) / 3.0

	return metrics
}
