package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// DeepDataStructureAnalysis performs comprehensive analysis of indexed data
type DeepDataStructureAnalysis struct {
	db        *sql.DB
	analytics map[string]interface{}
}

// ContentCluster represents a group of similar content
type ContentCluster struct {
	ID         string                 `json:"id"`
	Name       string                 `json:"name"`
	Size       int                    `json:"size"`
	Files      []string               `json:"files"`
	Categories map[string]int         `json:"categories"`
	Metadata   map[string]interface{} `json:"metadata"`
	Similarity float64                `json:"similarity"`
}

// SemanticMapping represents semantic relationships between documents
type SemanticMapping struct {
	Source     string            `json:"source"`
	Target     string            `json:"target"`
	Relation   string            `json:"relation"`
	Confidence float64           `json:"confidence"`
	Metadata   map[string]string `json:"metadata"`
}

// KnowledgeGraph represents the knowledge graph structure
type KnowledgeGraph struct {
	Nodes []KnowledgeNode `json:"nodes"`
	Edges []KnowledgeEdge `json:"edges"`
}

// KnowledgeNode represents a node in the knowledge graph
type KnowledgeNode struct {
	ID       string                 `json:"id"`
	Label    string                 `json:"label"`
	Type     string                 `json:"type"`
	Weight   float64                `json:"weight"`
	Metadata map[string]interface{} `json:"metadata"`
}

// KnowledgeEdge represents an edge in the knowledge graph
type KnowledgeEdge struct {
	Source   string            `json:"source"`
	Target   string            `json:"target"`
	Weight   float64           `json:"weight"`
	Relation string            `json:"relation"`
	Metadata map[string]string `json:"metadata"`
}

func NewDeepDataStructureAnalysis(dbPath string) (*DeepDataStructureAnalysis, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %v", err)
	}

	analytics := make(map[string]interface{})

	return &DeepDataStructureAnalysis{
		db:        db,
		analytics: analytics,
	}, nil
}

// AnalyzeContentStructure performs comprehensive content structure analysis
func (dsa *DeepDataStructureAnalysis) AnalyzeContentStructure() error {
	fmt.Println("🔍 Starting Deep Data Structure Analysis...")

	// 1. Content Clustering
	clusters, err := dsa.performContentClustering()
	if err != nil {
		return fmt.Errorf("content clustering failed: %v", err)
	}

	// 2. Semantic Mapping
	mappings, err := dsa.performSemanticMapping()
	if err != nil {
		return fmt.Errorf("semantic mapping failed: %v", err)
	}

	// 3. Knowledge Graph Construction
	knowledgeGraph, err := dsa.buildKnowledgeGraph()
	if err != nil {
		return fmt.Errorf("knowledge graph construction failed: %v", err)
	}

	// 4. Duplicate Detection
	duplicates, err := dsa.detectDuplicates()
	if err != nil {
		return fmt.Errorf("duplicate detection failed: %v", err)
	}

	// 5. Generate comprehensive report
	report := map[string]interface{}{
		"timestamp":         time.Now(),
		"total_files":       dsa.getTotalFiles(),
		"content_clusters":  clusters,
		"semantic_mappings": mappings,
		"knowledge_graph":   knowledgeGraph,
		"duplicates":        duplicates,
		"analytics":         dsa.analytics,
	}

	// Save report
	reportPath := filepath.Join("Z:\\01_PROJECTS\\tibrain\\reports", "deep_structure_analysis.json")
	os.MkdirAll(filepath.Dir(reportPath), 0755)

	reportData, _ := json.MarshalIndent(report, "", "  ")
	err = os.WriteFile(reportPath, reportData, 0644)
	if err != nil {
		return fmt.Errorf("failed to save report: %v", err)
	}

	fmt.Printf("✅ Deep Data Structure Analysis completed!\n")
	fmt.Printf("📊 Report saved to: %s\n", reportPath)
	fmt.Printf("📈 Found %d content clusters\n", len(clusters))
	fmt.Printf("🔗 Created %d semantic mappings\n", len(mappings))
	fmt.Printf("🕸️ Built knowledge graph with %d nodes and %d edges\n",
		len(knowledgeGraph.Nodes), len(knowledgeGraph.Edges))
	fmt.Printf("🔄 Found %d potential duplicates\n", len(duplicates))

	return nil
}

// performContentClustering groups similar content together
func (dsa *DeepDataStructureAnalysis) performContentClustering() ([]ContentCluster, error) {
	fmt.Println("📊 Performing Content Clustering...")

	rows, err := dsa.db.Query(`
		SELECT id, content, metadata FROM rag_documents 
		WHERE content IS NOT NULL AND content != ''
		LIMIT 10000
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	clusters := make(map[string]*ContentCluster)

	for rows.Next() {
		var id, content string
		var metadataJSON sql.NullString

		err := rows.Scan(&id, &content, &metadataJSON)
		if err != nil {
			continue
		}

		// Extract features for clustering
		features := dsa.extractFeatures(content)
		clusterID := dsa.generateClusterID(features)

		if cluster, exists := clusters[clusterID]; exists {
			cluster.Size++
			cluster.Files = append(cluster.Files, id)
			cluster.Categories[features.Category]++
		} else {
			var metadata map[string]interface{}
			if metadataJSON.Valid {
				json.Unmarshal([]byte(metadataJSON.String), &metadata)
			}

			clusters[clusterID] = &ContentCluster{
				ID:         clusterID,
				Name:       features.Name,
				Size:       1,
				Files:      []string{id},
				Categories: map[string]int{features.Category: 1},
				Metadata:   metadata,
				Similarity: 0.0,
			}
		}
	}

	// Convert to slice
	result := make([]ContentCluster, 0, len(clusters))
	for _, cluster := range clusters {
		// Calculate similarity score
		cluster.Similarity = dsa.calculateClusterSimilarity(cluster)
		result = append(result, *cluster)
	}

	// Update analytics
	dsa.analytics["content_clusters"] = map[string]interface{}{
		"total":    len(result),
		"avg_size": dsa.calculateAverageClusterSize(result),
		"largest":  dsa.findLargestCluster(result),
		"smallest": dsa.findSmallestCluster(result),
	}

	return result, nil
}

// ContentFeatures represents extracted features from content
type ContentFeatures struct {
	Name     string
	Category string
	Type     string
	Size     int
	Tokens   []string
}

// extractFeatures extracts features from content for clustering
func (dsa *DeepDataStructureAnalysis) extractFeatures(content string) ContentFeatures {
	features := ContentFeatures{
		Tokens: strings.Fields(strings.ToLower(content)),
		Size:   len(content),
	}

	// Detect content type
	if strings.Contains(content, "func ") || strings.Contains(content, "function") {
		features.Type = "code"
		features.Category = "programming"
	} else if strings.Contains(content, "# ") || strings.Contains(content, "## ") {
		features.Type = "markdown"
		features.Category = "documentation"
	} else if strings.Contains(content, "{") && strings.Contains(content, "}") {
		features.Type = "json"
		features.Category = "configuration"
	} else {
		features.Type = "text"
		features.Category = "general"
	}

	// Generate name from first line or tokens
	lines := strings.Split(content, "\n")
	if len(lines) > 0 && len(strings.TrimSpace(lines[0])) > 0 {
		features.Name = strings.TrimSpace(lines[0])
		if len(features.Name) > 50 {
			features.Name = features.Name[:47] + "..."
		}
	} else {
		features.Name = fmt.Sprintf("%s_%d", features.Type, len(content))
	}

	return features
}

// generateClusterID generates a cluster ID based on features
func (dsa *DeepDataStructureAnalysis) generateClusterID(features ContentFeatures) string {
	// Simple clustering based on type and size category
	sizeCategory := "small"
	if features.Size > 1000 {
		sizeCategory = "medium"
	}
	if features.Size > 5000 {
		sizeCategory = "large"
	}

	return fmt.Sprintf("%s_%s_%s", features.Type, features.Category, sizeCategory)
}

// calculateClusterSimilarity calculates similarity score for a cluster
func (dsa *DeepDataStructureAnalysis) calculateClusterSimilarity(cluster *ContentCluster) float64 {
	if cluster.Size <= 1 {
		return 1.0
	}

	// Simple similarity based on size and category distribution
	categoryScore := 0.0
	for _, count := range cluster.Categories {
		ratio := float64(count) / float64(cluster.Size)
		categoryScore += ratio * ratio
	}

	sizeScore := math.Min(float64(cluster.Size)/10.0, 1.0)

	return (categoryScore + sizeScore) / 2.0
}

// performSemanticMapping creates semantic relationships between documents
func (dsa *DeepDataStructureAnalysis) performSemanticMapping() ([]SemanticMapping, error) {
	fmt.Println("🔗 Performing Semantic Mapping...")

	var mappings []SemanticMapping

	// Get sample documents for analysis
	rows, err := dsa.db.Query(`
		SELECT id, content, metadata FROM rag_documents 
		WHERE content IS NOT NULL AND content != ''
		LIMIT 1000
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	documents := make(map[string]string)
	for rows.Next() {
		var id, content string
		var metadataJSON sql.NullString

		err := rows.Scan(&id, &content, &metadataJSON)
		if err != nil {
			continue
		}

		documents[id] = content
	}

	// Create semantic mappings based on content similarity
	for id1, content1 := range documents {
		for id2, content2 := range documents {
			if id1 >= id2 {
				continue
			}

			similarity := dsa.calculateContentSimilarity(content1, content2)
			if similarity > 0.3 { // Threshold for semantic relationship
				relation := dsa.determineRelation(content1, content2)

				mapping := SemanticMapping{
					Source:     id1,
					Target:     id2,
					Relation:   relation,
					Confidence: similarity,
					Metadata: map[string]string{
						"similarity_type": "content_based",
						"created_at":      time.Now().Format(time.RFC3339),
					},
				}

				mappings = append(mappings, mapping)
			}
		}
	}

	// Update analytics
	dsa.analytics["semantic_mappings"] = map[string]interface{}{
		"total":          len(mappings),
		"avg_confidence": dsa.calculateAverageConfidence(mappings),
		"relations":      dsa.countRelationTypes(mappings),
	}

	return mappings, nil
}

// calculateContentSimilarity calculates similarity between two content pieces
func (dsa *DeepDataStructureAnalysis) calculateContentSimilarity(content1, content2 string) float64 {
	// Simple Jaccard similarity on word tokens
	words1 := strings.Fields(strings.ToLower(content1))
	words2 := strings.Fields(strings.ToLower(content2))

	set1 := make(map[string]bool)
	set2 := make(map[string]bool)

	for _, word := range words1 {
		set1[word] = true
	}

	for _, word := range words2 {
		set2[word] = true
	}

	intersection := 0
	for word := range set1 {
		if set2[word] {
			intersection++
		}
	}

	union := len(set1) + len(set2) - intersection

	if union == 0 {
		return 0.0
	}

	return float64(intersection) / float64(union)
}

// determineRelation determines the type of relationship between documents
func (dsa *DeepDataStructureAnalysis) determineRelation(content1, content2 string) string {
	// Simple relation detection based on content patterns
	if strings.Contains(content1, "import") && strings.Contains(content2, "export") {
		return "dependency"
	}

	if strings.Contains(content1, "extends") || strings.Contains(content2, "extends") {
		return "inheritance"
	}

	if strings.Contains(content1, "implements") || strings.Contains(content2, "implements") {
		return "implementation"
	}

	if strings.Contains(content1, "refers to") || strings.Contains(content2, "refers to") {
		return "reference"
	}

	return "similar"
}

// buildKnowledgeGraph constructs the knowledge graph
func (dsa *DeepDataStructureAnalysis) buildKnowledgeGraph() (*KnowledgeGraph, error) {
	fmt.Println("🕸️ Building Knowledge Graph...")

	graph := &KnowledgeGraph{
		Nodes: []KnowledgeNode{},
		Edges: []KnowledgeEdge{},
	}

	// Create nodes from documents
	rows, err := dsa.db.Query(`
		SELECT id, content, metadata FROM rag_documents 
		WHERE content IS NOT NULL AND content != ''
		LIMIT 5000
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	nodeMap := make(map[string]KnowledgeNode)

	for rows.Next() {
		var id, content string
		var metadataJSON sql.NullString

		err := rows.Scan(&id, &content, &metadataJSON)
		if err != nil {
			continue
		}

		features := dsa.extractFeatures(content)

		var metadata map[string]interface{}
		if metadataJSON.Valid {
			json.Unmarshal([]byte(metadataJSON.String), &metadata)
		}

		node := KnowledgeNode{
			ID:     id,
			Label:  features.Name,
			Type:   features.Type,
			Weight: float64(len(content)),
			Metadata: map[string]interface{}{
				"category": features.Category,
				"size":     features.Size,
				"tokens":   len(features.Tokens),
				"metadata": metadata,
			},
		}

		graph.Nodes = append(graph.Nodes, node)
		nodeMap[id] = node
	}

	// Create edges based on semantic mappings
	mappings, err := dsa.performSemanticMapping()
	if err != nil {
		return nil, err
	}

	for _, mapping := range mappings {
		edge := KnowledgeEdge{
			Source:   mapping.Source,
			Target:   mapping.Target,
			Weight:   mapping.Confidence,
			Relation: mapping.Relation,
			Metadata: mapping.Metadata,
		}

		graph.Edges = append(graph.Edges, edge)
	}

	// Update analytics
	dsa.analytics["knowledge_graph"] = map[string]interface{}{
		"nodes":      len(graph.Nodes),
		"edges":      len(graph.Edges),
		"density":    dsa.calculateGraphDensity(graph),
		"avg_degree": dsa.calculateAverageDegree(graph),
	}

	return graph, nil
}

// detectDuplicates finds potential duplicate content
func (dsa *DeepDataStructureAnalysis) detectDuplicates() ([]map[string]interface{}, error) {
	fmt.Println("🔄 Detecting Duplicates...")

	var duplicates []map[string]interface{}

	// Get all documents for duplicate detection
	rows, err := dsa.db.Query(`
		SELECT id, content, metadata FROM rag_documents 
		WHERE content IS NOT NULL AND content != ''
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	documents := make(map[string]string)
	for rows.Next() {
		var id, content string
		var metadataJSON sql.NullString

		err := rows.Scan(&id, &content, &metadataJSON)
		if err != nil {
			continue
		}

		documents[id] = content
	}

	// Find duplicates using content similarity
	processed := make(map[string]bool)

	for id1, content1 := range documents {
		if processed[id1] {
			continue
		}

		var duplicateGroup []string
		duplicateGroup = append(duplicateGroup, id1)

		for id2, content2 := range documents {
			if id1 == id2 || processed[id2] {
				continue
			}

			similarity := dsa.calculateContentSimilarity(content1, content2)
			if similarity > 0.8 { // High similarity threshold for duplicates
				duplicateGroup = append(duplicateGroup, id2)
				processed[id2] = true
			}
		}

		if len(duplicateGroup) > 1 {
			duplicate := map[string]interface{}{
				"group_id":   fmt.Sprintf("dup_%d", len(duplicates)),
				"files":      duplicateGroup,
				"similarity": dsa.calculateGroupSimilarity(documents, duplicateGroup),
				"size":       len(duplicateGroup),
			}
			duplicates = append(duplicates, duplicate)
		}

		processed[id1] = true
	}

	// Update analytics
	dsa.analytics["duplicates"] = map[string]interface{}{
		"total_groups":   len(duplicates),
		"total_files":    dsa.countDuplicateFiles(duplicates),
		"avg_group_size": dsa.calculateAverageDuplicateGroupSize(duplicates),
	}

	return duplicates, nil
}

// Helper methods
func (dsa *DeepDataStructureAnalysis) getTotalFiles() int {
	var count int
	dsa.db.QueryRow("SELECT COUNT(*) FROM rag_documents").Scan(&count)
	return count
}

func (dsa *DeepDataStructureAnalysis) calculateAverageClusterSize(clusters []ContentCluster) float64 {
	if len(clusters) == 0 {
		return 0
	}

	total := 0
	for _, cluster := range clusters {
		total += cluster.Size
	}

	return float64(total) / float64(len(clusters))
}

func (dsa *DeepDataStructureAnalysis) findLargestCluster(clusters []ContentCluster) ContentCluster {
	if len(clusters) == 0 {
		return ContentCluster{}
	}

	largest := clusters[0]
	for _, cluster := range clusters {
		if cluster.Size > largest.Size {
			largest = cluster
		}
	}

	return largest
}

func (dsa *DeepDataStructureAnalysis) findSmallestCluster(clusters []ContentCluster) ContentCluster {
	if len(clusters) == 0 {
		return ContentCluster{}
	}

	smallest := clusters[0]
	for _, cluster := range clusters {
		if cluster.Size < smallest.Size {
			smallest = cluster
		}
	}

	return smallest
}

func (dsa *DeepDataStructureAnalysis) calculateAverageConfidence(mappings []SemanticMapping) float64 {
	if len(mappings) == 0 {
		return 0
	}

	total := 0.0
	for _, mapping := range mappings {
		total += mapping.Confidence
	}

	return total / float64(len(mappings))
}

func (dsa *DeepDataStructureAnalysis) countRelationTypes(mappings []SemanticMapping) map[string]int {
	relationTypes := make(map[string]int)

	for _, mapping := range mappings {
		relationTypes[mapping.Relation]++
	}

	return relationTypes
}

func (dsa *DeepDataStructureAnalysis) calculateGraphDensity(graph *KnowledgeGraph) float64 {
	if len(graph.Nodes) == 0 {
		return 0
	}

	maxEdges := len(graph.Nodes) * (len(graph.Nodes) - 1) / 2
	if maxEdges == 0 {
		return 0
	}

	return float64(len(graph.Edges)) / float64(maxEdges)
}

func (dsa *DeepDataStructureAnalysis) calculateAverageDegree(graph *KnowledgeGraph) float64 {
	if len(graph.Nodes) == 0 {
		return 0
	}

	degrees := make(map[string]int)

	for _, edge := range graph.Edges {
		degrees[edge.Source]++
		degrees[edge.Target]++
	}

	total := 0
	for _, degree := range degrees {
		total += degree
	}

	return float64(total) / float64(len(graph.Nodes))
}

func (dsa *DeepDataStructureAnalysis) calculateGroupSimilarity(documents map[string]string, group []string) float64 {
	if len(group) < 2 {
		return 1.0
	}

	totalSimilarity := 0.0
	comparisons := 0

	for i := 0; i < len(group); i++ {
		for j := i + 1; j < len(group); j++ {
			similarity := dsa.calculateContentSimilarity(documents[group[i]], documents[group[j]])
			totalSimilarity += similarity
			comparisons++
		}
	}

	if comparisons == 0 {
		return 1.0
	}

	return totalSimilarity / float64(comparisons)
}

func (dsa *DeepDataStructureAnalysis) countDuplicateFiles(duplicates []map[string]interface{}) int {
	total := 0
	for _, group := range duplicates {
		total += group["size"].(int)
	}
	return total
}

func (dsa *DeepDataStructureAnalysis) calculateAverageDuplicateGroupSize(duplicates []map[string]interface{}) float64 {
	if len(duplicates) == 0 {
		return 0
	}

	total := 0
	for _, group := range duplicates {
		total += group["size"].(int)
	}

	return float64(total) / float64(len(duplicates))
}
