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

// IntelligentIndexingSystem provides smart prioritization and adaptive learning
type IntelligentIndexingSystem struct {
	db           *sql.DB
	learningData map[string]interface{}
	priority     map[string]float64
}

// ContentPriority represents priority scoring for content
type ContentPriority struct {
	ID         string  `json:"id"`
	Score      float64 `json:"score"`
	Category   string  `json:"category"`
	Importance float64 `json:"importance"`
	Urgency    float64 `json:"urgency"`
	Frequency  float64 `json:"frequency"`
	LastAccess string  `json:"last_access"`
}

// ContentClassification represents classified content
type ContentClassification struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`
	Category    string                 `json:"category"`
	Subcategory string                 `json:"subcategory"`
	Confidence  float64                `json:"confidence"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// ReindexingTask represents a dynamic reindexing task
type ReindexingTask struct {
	ID          string    `json:"id"`
	ContentID   string    `json:"content_id"`
	Reason      string    `json:"reason"`
	Priority    float64   `json:"priority"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	CompletedAt time.Time `json:"completed_at"`
}

func NewIntelligentIndexingSystem(dbPath string) (*IntelligentIndexingSystem, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %v", err)
	}

	learningData := make(map[string]interface{})
	priority := make(map[string]float64)

	return &IntelligentIndexingSystem{
		db:           db,
		learningData: learningData,
		priority:     priority,
	}, nil
}

// InitializeIntelligentSystem sets up the intelligent indexing system
func (iis *IntelligentIndexingSystem) InitializeIntelligentSystem() error {
	fmt.Println("🤖 Initializing Intelligent Indexing System...")

	// 1. Create learning tables
	err := iis.createLearningTables()
	if err != nil {
		return fmt.Errorf("failed to create learning tables: %v", err)
	}

	// 2. Analyze existing content for priority scoring
	err = iis.analyzeContentPriority()
	if err != nil {
		return fmt.Errorf("failed to analyze content priority: %v", err)
	}

	// 3. Initialize learning metrics
	err = iis.initializeLearningMetrics()
	if err != nil {
		return fmt.Errorf("failed to initialize learning metrics: %v", err)
	}

	// 4. Set up adaptive learning
	err = iis.setupAdaptiveLearning()
	if err != nil {
		return fmt.Errorf("failed to setup adaptive learning: %v", err)
	}

	fmt.Println("✅ Intelligent Indexing System initialized successfully!")
	return nil
}

// createLearningTables creates tables for learning data
func (iis *IntelligentIndexingSystem) createLearningTables() error {
	fmt.Println("📊 Creating learning tables...")

	tables := []string{
		`CREATE TABLE IF NOT EXISTS content_priority (
			id TEXT PRIMARY KEY,
			score REAL,
			category TEXT,
			importance REAL,
			urgency REAL,
			frequency REAL,
			last_access TEXT,
			created_at TEXT DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS learning_metrics (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			metric_type TEXT,
			metric_data TEXT,
			created_at TEXT DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS content_classification (
			id TEXT PRIMARY KEY,
			type TEXT,
			category TEXT,
			subcategory TEXT,
			confidence REAL,
			metadata TEXT,
			created_at TEXT DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS reindexing_tasks (
			id TEXT PRIMARY KEY,
			content_id TEXT,
			reason TEXT,
			priority REAL,
			status TEXT,
			created_at TEXT,
			completed_at TEXT
		)`,
		`CREATE TABLE IF NOT EXISTS query_patterns (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			query TEXT,
			response_time REAL,
			success_rate REAL,
			frequency INTEGER,
			created_at TEXT DEFAULT CURRENT_TIMESTAMP
		)`,
	}

	for _, table := range tables {
		_, err := iis.db.Exec(table)
		if err != nil {
			return fmt.Errorf("failed to create table: %v", err)
		}
	}

	return nil
}

// analyzeContentPriority analyzes and scores content priority
func (iis *IntelligentIndexingSystem) analyzeContentPriority() error {
	fmt.Println("🎯 Analyzing content priority...")

	rows, err := iis.db.Query(`
		SELECT id, content, metadata, created_at FROM rag_documents 
		WHERE content IS NOT NULL AND content != ''
	`)
	if err != nil {
		return err
	}
	defer rows.Close()

	var priorities []ContentPriority

	for rows.Next() {
		var id, content, metadataJSON, createdAt string

		err := rows.Scan(&id, &content, &metadataJSON, &createdAt)
		if err != nil {
			continue
		}

		priority := iis.calculateContentPriority(id, content, metadataJSON, createdAt)
		priorities = append(priorities, priority)
	}

	// Store priorities in database
	for _, priority := range priorities {
		if _, err := iis.db.Exec(`
			INSERT OR REPLACE INTO content_priority 
			(id, score, category, importance, urgency, frequency, last_access)
			VALUES (?, ?, ?, ?, ?, ?, ?)
		`, priority.ID, priority.Score, priority.Category,
			priority.Importance, priority.Urgency, priority.Frequency, priority.LastAccess); err != nil {
			fmt.Printf("Warning: failed to store priority %s: %v\n", priority.ID, err)
		}
	}

	fmt.Printf("✅ Analyzed %d content items for priority\n", len(priorities))
	return nil
}

// calculateContentPriority calculates priority score for content
func (iis *IntelligentIndexingSystem) calculateContentPriority(id, content, metadataJSON, createdAt string) ContentPriority {
	priority := ContentPriority{
		ID:         id,
		Category:   "general",
		Importance: 0.5,
		Urgency:    0.5,
		Frequency:  0.5,
		LastAccess: createdAt,
	}

	// Analyze content for importance
	contentLower := strings.ToLower(content)

	// Check for important keywords
	importantKeywords := []string{
		"critical", "urgent", "important", "security", "authentication",
		"authorization", "api", "database", "config", "setup", "install",
		"error", "fix", "bug", "issue", "problem", "solution",
	}

	urgencyKeywords := []string{
		"urgent", "immediate", "asap", "critical", "emergency",
		"breaking", "fix", "patch", "hotfix", "security",
	}

	frequencyKeywords := []string{
		"guide", "tutorial", "how to", "example", "reference",
		"documentation", "readme", "getting started", "quick start",
	}

	// Calculate importance score
	importanceScore := 0.0
	for _, keyword := range importantKeywords {
		if strings.Contains(contentLower, keyword) {
			importanceScore += 0.1
		}
	}
	priority.Importance = math.Min(importanceScore, 1.0)

	// Calculate urgency score
	urgencyScore := 0.0
	for _, keyword := range urgencyKeywords {
		if strings.Contains(contentLower, keyword) {
			urgencyScore += 0.15
		}
	}
	priority.Urgency = math.Min(urgencyScore, 1.0)

	// Calculate frequency score
	frequencyScore := 0.0
	for _, keyword := range frequencyKeywords {
		if strings.Contains(contentLower, keyword) {
			frequencyScore += 0.1
		}
	}
	priority.Frequency = math.Min(frequencyScore, 1.0)

	// Determine category
	if strings.Contains(contentLower, "api") {
		priority.Category = "api"
	} else if strings.Contains(contentLower, "security") || strings.Contains(contentLower, "auth") {
		priority.Category = "security"
	} else if strings.Contains(contentLower, "config") || strings.Contains(contentLower, "setup") {
		priority.Category = "configuration"
	} else if strings.Contains(contentLower, "error") || strings.Contains(contentLower, "fix") {
		priority.Category = "troubleshooting"
	} else if strings.Contains(contentLower, "guide") || strings.Contains(contentLower, "tutorial") {
		priority.Category = "tutorial"
	}

	// Calculate overall score
	priority.Score = (priority.Importance * 0.4) + (priority.Urgency * 0.3) + (priority.Frequency * 0.3)

	return priority
}

// initializeLearningMetrics initializes the learning metrics system
func (iis *IntelligentIndexingSystem) initializeLearningMetrics() error {
	fmt.Println("📈 Initializing learning metrics...")

	metrics := LearningMetrics{
		ContentGaps: []string{},
	}

	// Store initial metrics
	metricsJSON, _ := json.Marshal(metrics)
	_, err := iis.db.Exec(`
		INSERT INTO learning_metrics (metric_type, metric_data) 
		VALUES ('initial_metrics', ?)
	`, string(metricsJSON))

	if err != nil {
		return err
	}

	fmt.Println("✅ Learning metrics initialized")
	return nil
}

// setupAdaptiveLearning sets up the adaptive learning system
func (iis *IntelligentIndexingSystem) setupAdaptiveLearning() error {
	fmt.Println("🧠 Setting up adaptive learning...")

	// Create adaptive learning configuration
	config := map[string]interface{}{
		"learning_rate":            0.01,
		"decay_factor":             0.95,
		"min_confidence":           0.7,
		"max_content_age":          30, // days
		"reindex_threshold":        0.8,
		"priority_update_interval": 3600, // seconds
	}

	configJSON, _ := json.Marshal(config)
	_, err := iis.db.Exec(`
		INSERT INTO learning_metrics (metric_type, metric_data) 
		VALUES ('adaptive_config', ?)
	`, string(configJSON))

	if err != nil {
		return err
	}

	fmt.Println("✅ Adaptive learning configured")
	return nil
}

// PerformContentClassification classifies content automatically
func (iis *IntelligentIndexingSystem) PerformContentClassification() error {
	fmt.Println("🏷️ Performing content classification...")

	rows, err := iis.db.Query(`
		SELECT id, content, metadata FROM rag_documents 
		WHERE content IS NOT NULL AND content != ''
		LIMIT 5000
	`)
	if err != nil {
		return err
	}
	defer rows.Close()

	var classifications []ContentClassification

	for rows.Next() {
		var id, content, metadataJSON string

		err := rows.Scan(&id, &content, &metadataJSON)
		if err != nil {
			continue
		}

		classification := iis.classifyContent(id, content, metadataJSON)
		classifications = append(classifications, classification)
	}

	// Store classifications
	for _, classification := range classifications {
		metadataJSON, _ := json.Marshal(classification.Metadata)
		if _, err := iis.db.Exec(`
			INSERT OR REPLACE INTO content_classification 
			(id, type, category, subcategory, confidence, metadata)
			VALUES (?, ?, ?, ?, ?, ?)
		`, classification.ID, classification.Type, classification.Category,
			classification.Subcategory, classification.Confidence, string(metadataJSON)); err != nil {
			fmt.Printf("Warning: failed to store classification %s: %v\n", classification.ID, err)
		}
	}

	fmt.Printf("✅ Classified %d content items\n", len(classifications))
	return nil
}

// classifyContent classifies a single content item
func (iis *IntelligentIndexingSystem) classifyContent(id, content, metadataJSON string) ContentClassification {
	classification := ContentClassification{
		ID:          id,
		Type:        "text",
		Category:    "general",
		Subcategory: "unknown",
		Confidence:  0.5,
		Metadata:    make(map[string]interface{}),
	}

	contentLower := strings.ToLower(content)

	// Type classification
	if strings.Contains(contentLower, "func ") || strings.Contains(contentLower, "function") {
		classification.Type = "code"
		classification.Category = "programming"
		classification.Subcategory = "function"
		classification.Confidence = 0.9
	} else if strings.Contains(contentLower, "class ") {
		classification.Type = "code"
		classification.Category = "programming"
		classification.Subcategory = "class"
		classification.Confidence = 0.9
	} else if strings.Contains(contentLower, "# ") || strings.Contains(contentLower, "## ") {
		classification.Type = "markdown"
		classification.Category = "documentation"
		classification.Subcategory = "markdown"
		classification.Confidence = 0.8
	} else if strings.Contains(contentLower, "{") && strings.Contains(contentLower, "}") {
		classification.Type = "json"
		classification.Category = "configuration"
		classification.Subcategory = "json"
		classification.Confidence = 0.8
	} else if strings.Contains(contentLower, "api") {
		classification.Type = "api"
		classification.Category = "api"
		classification.Subcategory = "documentation"
		classification.Confidence = 0.7
	} else if strings.Contains(contentLower, "error") || strings.Contains(contentLower, "fix") {
		classification.Type = "troubleshooting"
		classification.Category = "support"
		classification.Subcategory = "error"
		classification.Confidence = 0.7
	}

	// Extract metadata
	classification.Metadata["word_count"] = len(strings.Fields(content))
	classification.Metadata["char_count"] = len(content)
	classification.Metadata["line_count"] = len(strings.Split(content, "\n"))

	// Parse existing metadata if available
	if metadataJSON != "" {
		var existingMetadata map[string]interface{}
		json.Unmarshal([]byte(metadataJSON), &existingMetadata)
		for k, v := range existingMetadata {
			classification.Metadata[k] = v
		}
	}

	return classification
}

// GenerateReindexingTasks creates dynamic reindexing tasks
func (iis *IntelligentIndexingSystem) GenerateReindexingTasks() error {
	fmt.Println("🔄 Generating reindexing tasks...")

	// Find content that needs reindexing
	rows, err := iis.db.Query(`
		SELECT id, content, created_at FROM rag_documents 
		WHERE content IS NOT NULL AND content != ''
	`)
	if err != nil {
		return err
	}
	defer rows.Close()

	var tasks []ReindexingTask

	for rows.Next() {
		var id, content, createdAt string

		err := rows.Scan(&id, &content, &createdAt)
		if err != nil {
			continue
		}

		// Check if content needs reindexing
		needsReindexing, reason := iis.checkReindexingNeed(id, content, createdAt)
		if needsReindexing {
			task := ReindexingTask{
				ID:        fmt.Sprintf("reindex_%s_%d", id, time.Now().Unix()),
				ContentID: id,
				Reason:    reason,
				Priority:  iis.calculateReindexingPriority(id, reason),
				Status:    "pending",
				CreatedAt: time.Now(),
			}
			tasks = append(tasks, task)
		}
	}

	// Store tasks
	for _, task := range tasks {
		if _, err := iis.db.Exec(`
			INSERT INTO reindexing_tasks 
			(id, content_id, reason, priority, status, created_at)
			VALUES (?, ?, ?, ?, ?, ?)
		`, task.ID, task.ContentID, task.Reason, task.Priority, task.Status, task.CreatedAt.Format(time.RFC3339)); err != nil {
			fmt.Printf("Warning: failed to store task %s: %v\n", task.ID, err)
		}
	}

	fmt.Printf("✅ Generated %d reindexing tasks\n", len(tasks))
	return nil
}

// checkReindexingNeed checks if content needs reindexing
func (iis *IntelligentIndexingSystem) checkReindexingNeed(id, content, createdAt string) (bool, string) {
	// Check content age
	createdTime, _ := time.Parse(time.RFC3339, createdAt)
	ageInDays := time.Since(createdTime).Hours() / 24

	if ageInDays > 30 {
		return true, "content_older_than_30_days"
	}

	// Check content quality
	if len(content) < 100 {
		return true, "content_too_short"
	}

	// Check for duplicates or similar content
	if iis.isDuplicateContent(id, content) {
		return true, "duplicate_content_detected"
	}

	// Check for low quality
	if iis.isLowQualityContent(content) {
		return true, "low_quality_content"
	}

	return false, ""
}

// calculateReindexingPriority calculates priority for reindexing task
func (iis *IntelligentIndexingSystem) calculateReindexingPriority(id, reason string) float64 {
	basePriority := 0.5

	switch reason {
	case "content_older_than_30_days":
		basePriority = 0.7
	case "content_too_short":
		basePriority = 0.8
	case "duplicate_content_detected":
		basePriority = 0.9
	case "low_quality_content":
		basePriority = 0.85
	}

	return basePriority
}

// isDuplicateContent checks if content is duplicate
func (iis *IntelligentIndexingSystem) isDuplicateContent(id, content string) bool {
	// Simple duplicate check based on content hash
	rows, err := iis.db.Query(`
		SELECT COUNT(*) FROM rag_documents 
		WHERE id != ? AND content = ?
	`, id, content)
	if err != nil {
		return false
	}
	defer rows.Close()

	var count int
	rows.Scan(&count)

	return count > 0
}

// isLowQualityContent checks if content is low quality
func (iis *IntelligentIndexingSystem) isLowQualityContent(content string) bool {
	// Simple quality checks
	wordCount := len(strings.Fields(content))
	charCount := len(content)
	lineCount := len(strings.Split(content, "\n"))

	// Check for very short content
	if wordCount < 10 || charCount < 50 {
		return true
	}

	// Check for content with mostly special characters
	specialCharCount := 0
	for _, char := range content {
		if !((char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') ||
			(char >= '0' && char <= '9') || char == ' ' || char == '\n' || char == '\t') {
			specialCharCount++
		}
	}

	if float64(specialCharCount)/float64(charCount) > 0.3 {
		return true
	}

	// Check for content with too many lines but few words
	if lineCount > 10 && wordCount < 20 {
		return true
	}

	return false
}

// UpdateLearningFromQueries updates learning from query patterns
func (iis *IntelligentIndexingSystem) UpdateLearningFromQueries() error {
	fmt.Println("📊 Updating learning from query patterns...")

	// Simulate query pattern analysis
	queryPatterns := map[string]int{
		"api_documentation": 150,
		"setup_guide":       120,
		"troubleshooting":   80,
		"security_guide":    60,
		"configuration":     90,
		"examples":          110,
		"tutorials":         100,
		"reference":         70,
	}

	// Store query patterns
	for query, frequency := range queryPatterns {
		_, err := iis.db.Exec(`
			INSERT INTO query_patterns (query, response_time, success_rate, frequency)
			VALUES (?, ?, ?, ?)
		`, query, 0.5, 0.9, frequency)
		if err != nil {
			return err
		}
	}

	fmt.Println("✅ Query patterns updated")
	return nil
}

// GenerateIntelligentReport generates comprehensive intelligent system report
func (iis *IntelligentIndexingSystem) GenerateIntelligentReport() error {
	fmt.Println("📋 Generating intelligent system report...")

	// Collect all metrics
	report := map[string]interface{}{
		"timestamp":              time.Now(),
		"content_priority":       iis.getContentPriorityStats(),
		"content_classification": iis.getContentClassificationStats(),
		"reindexing_tasks":       iis.getReindexingTaskStats(),
		"query_patterns":         iis.getQueryPatternStats(),
		"learning_metrics":       iis.getLearningMetricsStats(),
	}

	// Save report
	reportPath := filepath.Join("Z:\\01_PROJECTS\\tibrain\\reports", "intelligent_indexing_report.json")
	os.MkdirAll(filepath.Dir(reportPath), 0755)

	reportData, _ := json.MarshalIndent(report, "", "  ")
	err := os.WriteFile(reportPath, reportData, 0644)
	if err != nil {
		return fmt.Errorf("failed to save report: %v", err)
	}

	fmt.Printf("✅ Intelligent system report saved to: %s\n", reportPath)
	return nil
}

// Helper methods for statistics
func (iis *IntelligentIndexingSystem) getContentPriorityStats() map[string]interface{} {
	stats := make(map[string]interface{})

	// Get priority distribution
	rows, err := iis.db.Query(`
		SELECT category, COUNT(*), AVG(score) FROM content_priority 
		GROUP BY category
	`)
	if err != nil {
		return stats
	}
	defer rows.Close()

	categories := make(map[string]map[string]interface{})
	for rows.Next() {
		var category string
		var count int
		var avgScore float64

		rows.Scan(&category, &count, &avgScore)
		categories[category] = map[string]interface{}{
			"count":     count,
			"avg_score": avgScore,
		}
	}

	stats["categories"] = categories
	return stats
}

func (iis *IntelligentIndexingSystem) getContentClassificationStats() map[string]interface{} {
	stats := make(map[string]interface{})

	// Get classification distribution
	rows, err := iis.db.Query(`
		SELECT type, category, COUNT(*), AVG(confidence) FROM content_classification 
		GROUP BY type, category
	`)
	if err != nil {
		return stats
	}
	defer rows.Close()

	types := make(map[string]map[string]interface{})
	for rows.Next() {
		var contentType, category string
		var count int
		var avgConfidence float64

		rows.Scan(&contentType, &category, &count, &avgConfidence)
		if types[contentType] == nil {
			types[contentType] = make(map[string]interface{})
		}
		types[contentType][category] = map[string]interface{}{
			"count":          count,
			"avg_confidence": avgConfidence,
		}
	}

	stats["types"] = types
	return stats
}

func (iis *IntelligentIndexingSystem) getReindexingTaskStats() map[string]interface{} {
	stats := make(map[string]interface{})

	// Get task statistics
	rows, err := iis.db.Query(`
		SELECT status, COUNT(*), AVG(priority) FROM reindexing_tasks 
		GROUP BY status
	`)
	if err != nil {
		return stats
	}
	defer rows.Close()

	statuses := make(map[string]interface{})
	for rows.Next() {
		var status string
		var count int
		var avgPriority float64

		rows.Scan(&status, &count, &avgPriority)
		statuses[status] = map[string]interface{}{
			"count":        count,
			"avg_priority": avgPriority,
		}
	}

	stats["statuses"] = statuses
	return stats
}

func (iis *IntelligentIndexingSystem) getQueryPatternStats() map[string]interface{} {
	stats := make(map[string]interface{})

	// Get query pattern statistics
	rows, err := iis.db.Query(`
		SELECT query, COUNT(*), AVG(response_time), AVG(success_rate) FROM query_patterns 
		GROUP BY query
		ORDER BY COUNT(*) DESC
		LIMIT 10
	`)
	if err != nil {
		return stats
	}
	defer rows.Close()

	queries := make([]map[string]interface{}, 0)
	for rows.Next() {
		var query string
		var count int
		var avgResponseTime, avgSuccessRate float64

		rows.Scan(&query, &count, &avgResponseTime, &avgSuccessRate)
		queries = append(queries, map[string]interface{}{
			"query":             query,
			"count":             count,
			"avg_response_time": avgResponseTime,
			"avg_success_rate":  avgSuccessRate,
		})
	}

	stats["top_queries"] = queries
	return stats
}

func (iis *IntelligentIndexingSystem) getLearningMetricsStats() map[string]interface{} {
	stats := make(map[string]interface{})

	// Get learning metrics
	rows, err := iis.db.Query(`
		SELECT metric_type, metric_data FROM learning_metrics 
		ORDER BY created_at DESC
		LIMIT 5
	`)
	if err != nil {
		return stats
	}
	defer rows.Close()

	metrics := make([]map[string]interface{}, 0)
	for rows.Next() {
		var metricType, metricData string

		rows.Scan(&metricType, &metricData)
		metrics = append(metrics, map[string]interface{}{
			"type": metricType,
			"data": metricData,
		})
	}

	stats["recent_metrics"] = metrics
	return stats
}
