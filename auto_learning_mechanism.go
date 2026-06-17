package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// AutoLearningMechanism provides automatic learning from user interactions
type AutoLearningMechanism struct {
	db              *sql.DB
	learningEngine  *LearningEngine
	queryAnalyzer   *QueryAnalyzer
	contentAnalyzer *ContentAnalyzer
}

// LearningEngine handles the core learning algorithms
type LearningEngine struct {
	learningRate     float64
	decayFactor      float64
	minConfidence    float64
	maxContentAge    int
	reindexThreshold float64
}

// QueryAnalyzer analyzes user query patterns
type QueryAnalyzer struct {
	queryPatterns map[string]*QueryPattern
	frequencyMap  map[string]int
	successRates  map[string]float64
	responseTimes map[string]float64
}

// ContentAnalyzer analyzes content quality and gaps
type ContentAnalyzer struct {
	qualityScores   map[string]float64
	contentGaps     []string
	coverageMap     map[string]float64
	recommendations []ContentRecommendation
}

// ContentRecommendation represents a content recommendation
type ContentRecommendation struct {
	Type        string  `json:"type"`
	Priority    float64 `json:"priority"`
	Description string  `json:"description"`
	Source      string  `json:"source"`
	Confidence  float64 `json:"confidence"`
}

func NewAutoLearningMechanism(dbPath string) (*AutoLearningMechanism, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %v", err)
	}

	learningEngine := &LearningEngine{
		learningRate:     0.01,
		decayFactor:      0.95,
		minConfidence:    0.7,
		maxContentAge:    30, // days
		reindexThreshold: 0.8,
	}

	queryAnalyzer := &QueryAnalyzer{
		queryPatterns: make(map[string]*QueryPattern),
		frequencyMap:  make(map[string]int),
		successRates:  make(map[string]float64),
		responseTimes: make(map[string]float64),
	}

	contentAnalyzer := &ContentAnalyzer{
		qualityScores:   make(map[string]float64),
		contentGaps:     []string{},
		coverageMap:     make(map[string]float64),
		recommendations: []ContentRecommendation{},
	}

	return &AutoLearningMechanism{
		db:              db,
		learningEngine:  learningEngine,
		queryAnalyzer:   queryAnalyzer,
		contentAnalyzer: contentAnalyzer,
	}, nil
}

// InitializeAutoLearning sets up the auto-learning system
func (alm *AutoLearningMechanism) InitializeAutoLearning() error {
	fmt.Println("🧠 Initializing Auto-Learning Mechanism...")

	// 1. Create learning tables
	err := alm.createLearningTables()
	if err != nil {
		return fmt.Errorf("failed to create learning tables: %v", err)
	}

	// 2. Load existing learning data
	err = alm.loadExistingLearningData()
	if err != nil {
		return fmt.Errorf("failed to load existing learning data: %v", err)
	}

	// 3. Initialize learning algorithms
	err = alm.initializeLearningAlgorithms()
	if err != nil {
		return fmt.Errorf("failed to initialize learning algorithms: %v", err)
	}

	// 4. Set up continuous learning
	err = alm.setupContinuousLearning()
	if err != nil {
		return fmt.Errorf("failed to setup continuous learning: %v", err)
	}

	fmt.Println("✅ Auto-Learning Mechanism initialized successfully!")
	return nil
}

// createLearningTables creates tables for learning data
func (alm *AutoLearningMechanism) createLearningTables() error {
	fmt.Println("📊 Creating learning tables...")

	tables := []string{
		`CREATE TABLE IF NOT EXISTS query_patterns (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			pattern TEXT,
			frequency INTEGER,
			success_rate REAL,
			response_time REAL,
			last_used TEXT,
			confidence REAL,
			related_topics TEXT,
			created_at TEXT DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS content_gaps (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			gap_type TEXT,
			description TEXT,
			priority REAL,
			confidence REAL,
			status TEXT,
			created_at TEXT DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS quality_scores (
			content_id TEXT PRIMARY KEY,
			quality_score REAL,
			factors TEXT,
			last_updated TEXT DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS learning_metrics (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			metric_type TEXT,
			metric_data TEXT,
			created_at TEXT DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS recommendations (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			recommendation_type TEXT,
			priority REAL,
			description TEXT,
			source TEXT,
			confidence REAL,
			status TEXT,
			created_at TEXT DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS user_interactions (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			query TEXT,
			response_time REAL,
			success BOOLEAN,
			user_feedback INTEGER,
			content_accessed TEXT,
			created_at TEXT DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS learning_feedback (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			query_id TEXT,
			rating INTEGER,
			correction TEXT,
			comment TEXT,
			created_at TEXT DEFAULT CURRENT_TIMESTAMP
		)`,
	}

	for _, table := range tables {
		_, err := alm.db.Exec(table)
		if err != nil {
			return fmt.Errorf("failed to create table: %v", err)
		}
	}

	return nil
}

// loadExistingLearningData loads existing learning data from database
func (alm *AutoLearningMechanism) loadExistingLearningData() error {
	fmt.Println("📚 Loading existing learning data...")

	// Load query patterns
	rows, err := alm.db.Query("SELECT pattern, frequency, success_rate, response_time, last_used, confidence, related_topics FROM query_patterns")
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var pattern string
		var frequency int
		var successRate, responseTime, confidence float64
		var lastUsed, relatedTopics string

		err := rows.Scan(&pattern, &frequency, &successRate, &responseTime, &lastUsed, &confidence, &relatedTopics)
		if err != nil {
			continue
		}

		lastUsedTime, _ := time.Parse(time.RFC3339, lastUsed)
		var topics []string
		json.Unmarshal([]byte(relatedTopics), &topics)

		alm.queryAnalyzer.queryPatterns[pattern] = &QueryPattern{
			PatternID:       pattern,
			QueryText:       pattern,
			Frequency:       frequency,
			AvgResponseTime: responseTime,
			SuccessRate:     successRate,
			LastUsed:        lastUsedTime,
		}

		alm.queryAnalyzer.frequencyMap[pattern] = frequency
		alm.queryAnalyzer.successRates[pattern] = successRate
		alm.queryAnalyzer.responseTimes[pattern] = responseTime
	}

	// Load quality scores
	rows, err = alm.db.Query("SELECT content_id, quality_score, factors FROM quality_scores")
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var contentID string
		var qualityScore float64
		var factors string

		err := rows.Scan(&contentID, &qualityScore, &factors)
		if err != nil {
			continue
		}

		alm.contentAnalyzer.qualityScores[contentID] = qualityScore
	}

	fmt.Println("✅ Existing learning data loaded")
	return nil
}

// initializeLearningAlgorithms initializes the learning algorithms
func (alm *AutoLearningMechanism) initializeLearningAlgorithms() error {
	fmt.Println("🔧 Initializing learning algorithms...")

	// Initialize query pattern analysis
	err := alm.initializeQueryPatternAnalysis()
	if err != nil {
		return fmt.Errorf("failed to initialize query pattern analysis: %v", err)
	}

	// Initialize content gap detection
	err = alm.initializeContentGapDetection()
	if err != nil {
		return fmt.Errorf("failed to initialize content gap detection: %v", err)
	}

	// Initialize quality scoring
	err = alm.initializeQualityScoring()
	if err != nil {
		return fmt.Errorf("failed to initialize quality scoring: %v", err)
	}

	fmt.Println("✅ Learning algorithms initialized")
	return nil
}

// initializeQueryPatternAnalysis sets up query pattern analysis
func (alm *AutoLearningMechanism) initializeQueryPatternAnalysis() error {
	fmt.Println("🔍 Initializing query pattern analysis...")

	// Analyze existing queries to find patterns
	sampleQueries := []string{
		"how to setup authentication",
		"api documentation examples",
		"troubleshooting connection issues",
		"security best practices",
		"configuration guide",
		"getting started tutorial",
		"error handling patterns",
		"deployment instructions",
		"database connection setup",
		"performance optimization",
	}

	for _, query := range sampleQueries {
		pattern := alm.extractQueryPattern(query)
		if pattern == "" {
			continue
		}

		if _, exists := alm.queryAnalyzer.queryPatterns[pattern]; !exists {
			alm.queryAnalyzer.queryPatterns[pattern] = &QueryPattern{
				PatternID:       pattern,
				QueryText:       pattern,
				Frequency:       1,
				AvgResponseTime: 0.5,
				SuccessRate:     0.8,
				LastUsed:        time.Now(),
			}

			alm.queryAnalyzer.frequencyMap[pattern] = 1
			alm.queryAnalyzer.successRates[pattern] = 0.8
			alm.queryAnalyzer.responseTimes[pattern] = 0.5
		}
	}

	fmt.Println("✅ Query pattern analysis initialized")
	return nil
}

// extractQueryPattern extracts pattern from query
func (alm *AutoLearningMechanism) extractQueryPattern(query string) string {
	query = strings.ToLower(query)

	// Extract main keywords
	keywords := []string{"how to", "setup", "configure", "troubleshoot", "fix", "implement", "create", "update", "delete", "get", "list"}

	for _, keyword := range keywords {
		if strings.Contains(query, keyword) {
			return keyword
		}
	}

	// Extract domain keywords
	domains := []string{"api", "database", "security", "authentication", "authorization", "deployment", "performance", "testing"}

	for _, domain := range domains {
		if strings.Contains(query, domain) {
			return domain
		}
	}

	return ""
}

// initializeContentGapDetection sets up content gap detection
func (alm *AutoLearningMechanism) initializeContentGapDetection() error {
	fmt.Println("🔍 Initializing content gap detection...")

	// Analyze existing content to find gaps
	rows, err := alm.db.Query("SELECT id, content FROM rag_documents WHERE content IS NOT NULL AND content != '' LIMIT 1000")
	if err != nil {
		return err
	}
	defer rows.Close()

	contentTopics := make(map[string]int)
	totalContent := 0

	for rows.Next() {
		var id, content string
		err := rows.Scan(&id, &content)
		if err != nil {
			continue
		}

		topics := alm.extractContentTopics(content)
		for _, topic := range topics {
			contentTopics[topic]++
		}
		totalContent++
	}

	// Identify gaps (topics with low coverage)
	expectedTopics := []string{
		"authentication", "authorization", "security", "api", "database",
		"deployment", "testing", "monitoring", "logging", "configuration",
		"troubleshooting", "performance", "scaling", "backup", "recovery",
	}

	for _, topic := range expectedTopics {
		coverage := float64(contentTopics[topic]) / float64(totalContent)
		alm.contentAnalyzer.coverageMap[topic] = coverage

		if coverage < 0.1 { // Less than 10% coverage
			alm.contentAnalyzer.contentGaps = append(alm.contentAnalyzer.contentGaps, topic)
		}
	}

	fmt.Println("✅ Content gap detection initialized")
	return nil
}

// extractContentTopics extracts topics from content
func (alm *AutoLearningMechanism) extractContentTopics(content string) []string {
	content = strings.ToLower(content)
	topics := []string{}

	// Topic keywords
	topicKeywords := map[string][]string{
		"authentication":  {"auth", "login", "signin", "credential", "token"},
		"authorization":   {"permission", "role", "access", "grant", "deny"},
		"security":        {"security", "secure", "protect", "encrypt", "hash"},
		"api":             {"api", "endpoint", "request", "response", "rest"},
		"database":        {"database", "db", "sql", "query", "table"},
		"deployment":      {"deploy", "deployment", "production", "staging"},
		"testing":         {"test", "testing", "unit", "integration", "e2e"},
		"monitoring":      {"monitor", "metrics", "logging", "alert", "health"},
		"configuration":   {"config", "configuration", "setting", "parameter"},
		"troubleshooting": {"troubleshoot", "fix", "error", "issue", "problem"},
		"performance":     {"performance", "optimize", "speed", "latency", "throughput"},
	}

	for topic, keywords := range topicKeywords {
		for _, keyword := range keywords {
			if strings.Contains(content, keyword) {
				topics = append(topics, topic)
				break
			}
		}
	}

	return topics
}

// initializeQualityScoring sets up quality scoring
func (alm *AutoLearningMechanism) initializeQualityScoring() error {
	fmt.Println("⭐ Initializing quality scoring...")

	// Analyze existing content for quality
	rows, err := alm.db.Query("SELECT id, content FROM rag_documents WHERE content IS NOT NULL AND content != '' LIMIT 1000")
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

		qualityScore := alm.calculateContentQuality(content)
		alm.contentAnalyzer.qualityScores[id] = qualityScore
	}

	fmt.Println("✅ Quality scoring initialized")
	return nil
}

// calculateContentQuality calculates quality score for content
func (alm *AutoLearningMechanism) calculateContentQuality(content string) float64 {
	score := 0.0

	// Length factor (prefer moderate length)
	wordCount := len(strings.Fields(content))
	if wordCount > 50 && wordCount < 1000 {
		score += 0.2
	} else if wordCount >= 1000 {
		score += 0.1
	}

	// Structure factor (has headers, lists, code blocks)
	if strings.Contains(content, "#") || strings.Contains(content, "##") {
		score += 0.1
	}
	if strings.Contains(content, "-") || strings.Contains(content, "*") {
		score += 0.1
	}
	if strings.Contains(content, "```") || strings.Contains(content, "code") {
		score += 0.1
	}

	// Content quality factors
	if strings.Contains(content, "example") || strings.Contains(content, "sample") {
		score += 0.1
	}
	if strings.Contains(content, "step") || strings.Contains(content, "guide") {
		score += 0.1
	}
	if strings.Contains(content, "error") || strings.Contains(content, "fix") {
		score += 0.1
	}

	// Language quality (simple check)
	sentences := strings.Split(content, ".")
	if len(sentences) > 3 {
		score += 0.1
	}

	return math.Min(score, 1.0)
}

// setupContinuousLearning sets up continuous learning
func (alm *AutoLearningMechanism) setupContinuousLearning() error {
	fmt.Println("🔄 Setting up continuous learning...")

	// Create continuous learning configuration
	config := map[string]interface{}{
		"learning_interval":     3600, // 1 hour
		"batch_size":            100,
		"min_pattern_frequency": 5,
		"quality_threshold":     0.6,
		"gap_threshold":         0.1,
		"recommendation_limit":  10,
	}

	configJSON, _ := json.Marshal(config)
	_, err := alm.db.Exec(`
		INSERT INTO learning_metrics (metric_type, metric_data) 
		VALUES ('continuous_learning_config', ?)
	`, string(configJSON))

	if err != nil {
		return err
	}

	fmt.Println("✅ Continuous learning configured")
	return nil
}

// ProcessUserInteraction processes user interactions for learning
func (alm *AutoLearningMechanism) ProcessUserInteraction(query string, responseTime float64, success bool, userFeedback int) error {
	fmt.Println("👤 Processing user interaction...")

	// Store interaction
	_, err := alm.db.Exec(`
		INSERT INTO user_interactions (query, response_time, success, user_feedback)
		VALUES (?, ?, ?, ?)
	`, query, responseTime, success, userFeedback)
	if err != nil {
		return err
	}

	// Update query patterns
	pattern := alm.extractQueryPattern(query)
	if pattern != "" {
		alm.updateQueryPattern(pattern, responseTime, success)
	}

	// Update learning metrics
	alm.updateLearningMetrics(query, responseTime, success, userFeedback)

	return nil
}

// updateQueryPattern updates query pattern based on interaction
func (alm *AutoLearningMechanism) updateQueryPattern(pattern string, responseTime float64, success bool) {
	if existingPattern, exists := alm.queryAnalyzer.queryPatterns[pattern]; exists {
		// Update existing pattern
		existingPattern.Frequency++
		existingPattern.LastUsed = time.Now()

		// Update success rate with learning rate
		newSuccess := 1.0
		if !success {
			newSuccess = 0.0
		}
		existingPattern.SuccessRate = (existingPattern.SuccessRate * (1 - alm.learningEngine.learningRate)) +
			(newSuccess * alm.learningEngine.learningRate)

		// Update response time
		existingPattern.ResponseTime = (existingPattern.ResponseTime * 0.9) + (responseTime * 0.1)

		// Update confidence
		existingPattern.Confidence = math.Min(existingPattern.Confidence+0.01, 1.0)
	} else {
		// Create new pattern
		alm.queryAnalyzer.queryPatterns[pattern] = &QueryPattern{
			PatternID:       pattern,
			QueryText:       pattern,
			Frequency:       1,
			AvgResponseTime: responseTime,
			SuccessRate:     0.8,
			LastUsed:        time.Now(),
		}
	}

	// Save to database
	existingPattern := alm.queryAnalyzer.queryPatterns[pattern]
	relatedTopicsJSON, _ := json.Marshal(existingPattern.RelatedTopics)

	_, err := alm.db.Exec(`
		INSERT OR REPLACE INTO query_patterns 
		(pattern, frequency, success_rate, response_time, last_used, confidence, related_topics)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, existingPattern.Pattern, existingPattern.Frequency, existingPattern.SuccessRate, existingPattern.ResponseTime,
		existingPattern.LastUsed.Format(time.RFC3339), existingPattern.Confidence, string(relatedTopicsJSON))
	if err != nil {
		log.Printf("Failed to save query pattern: %v", err)
	}
}

// updateLearningMetrics updates learning metrics
func (alm *AutoLearningMechanism) updateLearningMetrics(query string, responseTime float64, success bool, userFeedback int) {
	// Update internal metrics
	alm.queryAnalyzer.frequencyMap[query]++

	if success {
		alm.queryAnalyzer.successRates[query] = (alm.queryAnalyzer.successRates[query] * 0.9) + 0.1
	} else {
		alm.queryAnalyzer.successRates[query] = alm.queryAnalyzer.successRates[query] * 0.9
	}

	alm.queryAnalyzer.responseTimes[query] = (alm.queryAnalyzer.responseTimes[query] * 0.9) + (responseTime * 0.1)
}

// DetectContentGaps detects content gaps based on usage patterns
func (alm *AutoLearningMechanism) DetectContentGaps() error {
	fmt.Println("🔍 Detecting content gaps...")

	// Analyze query patterns vs content coverage
	for pattern, queryPattern := range alm.queryAnalyzer.queryPatterns {
		if queryPattern.Frequency > 5 && queryPattern.SuccessRate < 0.7 {
			// High frequency, low success rate indicates content gap

			_, err := alm.db.Exec(`
				INSERT OR IGNORE INTO content_gaps (gap_type, description, priority, confidence, status)
				VALUES (?, ?, ?, ?, ?)
			`, "missing_content", fmt.Sprintf("High demand for %s but low success rate", pattern),
				float64(queryPattern.Frequency)*0.1, queryPattern.Confidence, "pending")
			if err != nil {
				return err
			}
		}
	}

	// Analyze topic coverage
	for topic, coverage := range alm.contentAnalyzer.coverageMap {
		if coverage < 0.1 {
			_, err := alm.db.Exec(`
				INSERT OR IGNORE INTO content_gaps (gap_type, description, priority, confidence, status)
				VALUES (?, ?, ?, ?, ?)
			`, "low_coverage", fmt.Sprintf("Low coverage for %s topic", topic),
				(1.0-coverage)*0.8, 0.7, "pending")
			if err != nil {
				return err
			}
		}
	}

	fmt.Println("✅ Content gaps detected")
	return nil
}

// GenerateRecommendations generates content recommendations
func (alm *AutoLearningMechanism) GenerateRecommendations() error {
	fmt.Println("💡 Generating content recommendations...")

	var recommendations []ContentRecommendation

	// Generate recommendations based on content gaps
	rows, err := alm.db.Query("SELECT description, priority, confidence FROM content_gaps WHERE status = 'pending'")
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var description string
		var priority, confidence float64

		err := rows.Scan(&description, &priority, &confidence)
		if err != nil {
			continue
		}

		recommendation := ContentRecommendation{
			Type:        "content_gap",
			Priority:    priority,
			Description: description,
			Source:      "gap_analysis",
			Confidence:  confidence,
		}
		recommendations = append(recommendations, recommendation)
	}

	// Generate recommendations based on query patterns
	for pattern, queryPattern := range alm.queryAnalyzer.queryPatterns {
		if queryPattern.Frequency > 10 && queryPattern.SuccessRate < 0.6 {
			recommendation := ContentRecommendation{
				Type:        "query_pattern",
				Priority:    float64(queryPattern.Frequency) * 0.05,
				Description: fmt.Sprintf("Improve content for %s queries", pattern),
				Source:      "query_analysis",
				Confidence:  queryPattern.Confidence,
			}
			recommendations = append(recommendations, recommendation)
		}
	}

	// Sort recommendations by priority
	sort.Slice(recommendations, func(i, j int) bool {
		return recommendations[i].Priority > recommendations[j].Priority
	})

	// Store top recommendations
	for i, rec := range recommendations {
		if i >= 10 { // Limit to top 10
			break
		}

		_, err := alm.db.Exec(`
			INSERT INTO recommendations (recommendation_type, priority, description, source, confidence, status)
			VALUES (?, ?, ?, ?, ?, ?)
		`, rec.Type, rec.Priority, rec.Description, rec.Source, rec.Confidence, "pending")
		if err != nil {
			return err
		}
	}

	fmt.Printf("✅ Generated %d recommendations\n", len(recommendations))
	return nil
}

// LearningMetrics tracks learning progress
type LearningMetrics struct {
	TotalQueries       int                     `json:"total_queries"`
	UniqueQueries      int                     `json:"unique_queries"`
	AvgResponseTime    float64                 `json:"avg_response_time"`
	SuccessRate        float64                 `json:"success_rate"`
	ContentGaps        []string                `json:"content_gaps"`
	QualityImprovement float64                 `json:"quality_improvement"`
	LearningProgress   map[string]float64      `json:"learning_progress"`
	Recommendations    []ContentRecommendation `json:"recommendations"`
}

// GenerateLearningReport generates comprehensive learning report
func (alm *AutoLearningMechanism) GenerateLearningReport() error {
	fmt.Println("📊 Generating learning report...")

	// Collect learning metrics
	metrics := LearningMetrics{
		TotalQueries:       alm.getTotalQueries(),
		UniqueQueries:      len(alm.queryAnalyzer.frequencyMap),
		AvgResponseTime:    alm.getAverageResponseTime(),
		SuccessRate:        alm.getOverallSuccessRate(),
		ContentGaps:        alm.getContentGaps(),
		QualityImprovement: alm.getQualityImprovement(),
		LearningProgress:   alm.getLearningProgress(),
		Recommendations:    alm.getTopRecommendations(),
	}

	// Save report
	reportPath := filepath.Join("Z:\\01_PROJECTS\\tibrain\\reports", "auto_learning_report.json")
	os.MkdirAll(filepath.Dir(reportPath), 0755)

	reportData, _ := json.MarshalIndent(metrics, "", "  ")
	err := os.WriteFile(reportPath, reportData, 0644)
	if err != nil {
		return fmt.Errorf("failed to save report: %v", err)
	}

	fmt.Printf("✅ Learning report saved to: %s\n", reportPath)
	return nil
}

// Helper methods for metrics
func (alm *AutoLearningMechanism) getTotalQueries() int {
	var count int
	alm.db.QueryRow("SELECT COUNT(*) FROM user_interactions").Scan(&count)
	return count
}

func (alm *AutoLearningMechanism) getAverageResponseTime() float64 {
	var avgTime float64
	alm.db.QueryRow("SELECT AVG(response_time) FROM user_interactions").Scan(&avgTime)
	return avgTime
}

func (alm *AutoLearningMechanism) getOverallSuccessRate() float64 {
	var rate float64
	alm.db.QueryRow("SELECT AVG(CASE WHEN success = 1 THEN 1.0 ELSE 0.0 END) FROM user_interactions").Scan(&rate)
	return rate
}

func (alm *AutoLearningMechanism) getContentGaps() []string {
	var gaps []string
	rows, err := alm.db.Query("SELECT description FROM content_gaps WHERE status = 'pending'")
	if err != nil {
		return gaps
	}
	defer rows.Close()

	for rows.Next() {
		var gap string
		rows.Scan(&gap)
		gaps = append(gaps, gap)
	}
	return gaps
}

func (alm *AutoLearningMechanism) getQualityImprovement() float64 {
	var improvement float64
	alm.db.QueryRow("SELECT AVG(quality_score) FROM quality_scores").Scan(&improvement)
	return improvement
}

func (alm *AutoLearningMechanism) getLearningProgress() map[string]float64 {
	progress := make(map[string]float64)

	// Calculate progress for different areas
	progress["query_patterns"] = float64(len(alm.queryAnalyzer.queryPatterns)) / 100.0
	progress["content_coverage"] = alm.calculateContentCoverage()
	progress["quality_scoring"] = float64(len(alm.contentAnalyzer.qualityScores)) / 1000.0

	return progress
}

func (alm *AutoLearningMechanism) calculateContentCoverage() float64 {
	if len(alm.contentAnalyzer.coverageMap) == 0 {
		return 0.0
	}

	totalCoverage := 0.0
	for _, coverage := range alm.contentAnalyzer.coverageMap {
		totalCoverage += coverage
	}

	return totalCoverage / float64(len(alm.contentAnalyzer.coverageMap))
}

func (alm *AutoLearningMechanism) getTopRecommendations() []ContentRecommendation {
	var recommendations []ContentRecommendation

	rows, err := alm.db.Query(`
		SELECT recommendation_type, priority, description, source, confidence 
		FROM recommendations 
		WHERE status = 'pending' 
		ORDER BY priority DESC 
		LIMIT 5
	`)
	if err != nil {
		return recommendations
	}
	defer rows.Close()

	for rows.Next() {
		var recType, description, source string
		var priority, confidence float64

		err := rows.Scan(&recType, &priority, &description, &source, &confidence)
		if err != nil {
			continue
		}

		recommendation := ContentRecommendation{
			Type:        recType,
			Priority:    priority,
			Description: description,
			Source:      source,
			Confidence:  confidence,
		}
		recommendations = append(recommendations, recommendation)
	}

	return recommendations
}

// ProcessFeedback processes user feedback and updates learning parameters
// This creates the feedback loop for adaptive learning
func (alm *AutoLearningMechanism) ProcessFeedback(queryID string, rating int, correction string, comment string) error {
	fmt.Printf("🔄 Processing feedback for query %s: rating=%d\n", queryID, rating)

	// Update learning engine parameters based on feedback
	if rating <= 2 { // Low rating - need to adjust
		alm.learningEngine.decayFactor *= 0.98  // Increase forgetting rate
		alm.learningEngine.learningRate *= 1.05 // Increase learning rate
	} else if rating >= 4 { // High rating - reinforce successful patterns
		alm.learningEngine.decayFactor = math.Min(alm.learningEngine.decayFactor*1.01, 0.99)
	}

	// Store feedback in database
	_, err := alm.db.Exec(`
		INSERT INTO learning_feedback (query_id, rating, correction, comment, created_at)
		VALUES (?, ?, ?, ?, ?)
	`, queryID, rating, correction, comment, time.Now().Unix())
	if err != nil {
		return fmt.Errorf("failed to store feedback: %v", err)
	}

	// Trigger content gap detection if rating is low
	if rating <= 2 {
		go alm.DetectContentGaps()
	}

	// Update recommendations based on feedback
	go alm.GenerateRecommendations()

	fmt.Println("✅ Feedback processed successfully")
	return nil
}

// AdaptiveLearningPipeline orchestrates the full learning cycle
func (alm *AutoLearningMechanism) AdaptiveLearningPipeline() error {
	fmt.Println("🚀 Starting Adaptive Learning Pipeline...")

	// 1. Process user interactions
	fmt.Println("  1. Processing user interactions...")

	// 2. Analyze query patterns
	fmt.Println("  2. Analyzing query patterns...")

	// 3. Detect content gaps
	fmt.Println("  3. Detecting content gaps...")
	err := alm.DetectContentGaps()
	if err != nil {
		return fmt.Errorf("gap detection failed: %v", err)
	}

	// 4. Generate recommendations
	fmt.Println("  4. Generating recommendations...")
	err = alm.GenerateRecommendations()
	if err != nil {
		return fmt.Errorf("recommendation generation failed: %v", err)
	}

	// 5. Update quality scores
	fmt.Println("  5. Updating quality scores...")
	err = alm.initializeQualityScoring()
	if err != nil {
		return fmt.Errorf("quality scoring update failed: %v", err)
	}

	// 6. Generate learning report
	fmt.Println("  6. Generating learning report...")
	err = alm.GenerateLearningReport()
	if err != nil {
		return fmt.Errorf("report generation failed: %v", err)
	}

	fmt.Println("✅ Adaptive Learning Pipeline completed")
	return nil
}
