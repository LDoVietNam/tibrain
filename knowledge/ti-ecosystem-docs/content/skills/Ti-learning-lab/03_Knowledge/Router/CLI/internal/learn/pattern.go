package learn

import (
	"regexp"
	"strings"
	"sync"
	"time"
)

// Pattern represents a normalized prompt pattern.
type Pattern struct {
	ID           string    `json:"id"`
	Template     string    `json:"template"`
	TaskType     string    `json:"task_type"`
	TaskDomain   string    `json:"task_domain"`
	SuccessCount int       `json:"success_count"`
	FailureCount int       `json:"failure_count"`
	AvgQuality   float64   `json:"avg_quality"`
	BestModel    string    `json:"best_model"`
	FirstSeen    time.Time `json:"first_seen"`
	LastUsed     time.Time `json:"last_used"`
}

// PatternExtractor normalizes prompts and extracts patterns.
type PatternExtractor struct {
	mu             sync.RWMutex
	patterns       map[string]*Pattern // normalized_template → Pattern
	taskClassifier *TaskClassifier
	storage        *Storage
}

// NewPatternExtractor creates a new extractor.
func NewPatternExtractor(storage *Storage) *PatternExtractor {
	return &PatternExtractor{
		patterns:       make(map[string]*Pattern),
		taskClassifier: NewTaskClassifier(),
		storage:        storage,
	}
}

// Normalize converts a prompt to a generalized template.
// Example: "Write unit test for UserService.CreateUser" → "write unit test for {Type}.{method}"
func (pe *PatternExtractor) Normalize(prompt string) string {
	p := strings.ToLower(prompt)

	// Replace specific identifiers (CamelCase, snake_case) with {Type}
	// Matches: FooBar, foo_bar, Foo123
	reType := regexp.MustCompile(`([A-Z][a-z0-9]+|[a-z]+(?:_[a-z0-9]+)+)`)
	p = reType.ReplaceAllString(p, "{Type}")

	// Replace specific numbers with {num}
	reNum := regexp.MustCompile(`\b\d+\b`)
	p = reNum.ReplaceAllString(p, "{num}")

	// Replace filenames with {file}
	reFile := regexp.MustCompile(`\b[\w./-]+\.(go|py|js|ts|java|cpp|c|h|rs|rb)\b`)
	p = reFile.ReplaceAllString(p, "{file}")

	// Replace API keys, tokens (redact)
	reToken := regexp.MustCompile(`(?:sk|gh|api|token)[-_][\w\d]{20,}`)
	p = reToken.ReplaceAllString(p, "{token}")

	// Replace URLs
	reURL := regexp.MustCompile(`https?://[^\s]+`)
	p = reURL.ReplaceAllString(p, "{url}")

	// Replace emails
	reEmail := regexp.MustCompile(`[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`)
	p = reEmail.ReplaceAllString(p, "{email}")

	// Collapse whitespace
	p = strings.Join(strings.Fields(p), " ")

	// Remove trailing punctuation
	p = strings.TrimRight(p, ".,;:!?")

	return p
}

// Process extracts/updates pattern from a log entry.
func (pe *PatternExtractor) Process(entry *LogInput) error {
	template := pe.Normalize(entry.Prompt)

	pe.mu.Lock()
	defer pe.mu.Unlock()

	pattern, exists := pe.patterns[template]
	if !exists {
		pattern = &Pattern{
			ID:         hashPrompt(template)[:12], // Short ID
			Template:   template,
			TaskType:   entry.TaskType,
			TaskDomain: entry.TaskDomain,
			FirstSeen:  time.Now(),
			LastUsed:   time.Now(),
		}
		pe.patterns[template] = pattern
	}

	// Update counters
	switch entry.Outcome {
	case "success":
		pattern.SuccessCount++
	case "failed", "error":
		pattern.FailureCount++
	}
	// Quality is 0-10
	pattern.AvgQuality = movingAverage(pattern.AvgQuality, entry.QualityCombined)
	pattern.LastUsed = time.Now()

	// If quality high and model specified, update best_model
	if entry.QualityCombined >= 8.0 && entry.Model != "" {
		pattern.BestModel = entry.Model
	}

	// Periodically persist to DB (via background worker)
	return nil
}

// GetBestPractices returns guidance for a task type.
func (pe *PatternExtractor) GetBestPractices(taskType string) string {
	// Map task type → best practices
	practices := map[string]string{
		"coding": `
Best practices for coding:
1. Include error handling
2. Add unit tests for key paths
3. Follow language idioms (Go: idiomatic Go)
4. Document public functions
5. Consider edge cases`,
		"debugging": `
Debugging checklist:
1. Reproduce the issue consistently
2. Check logs/error messages
3. Verify assumptions (print/log intermediate state)
4. Isolate the failing component
5. Write a test that fails before fix, passes after`,
		"testing": `
Testing guidelines:
1. Use table-driven tests
2. Cover edge cases and error paths
3. Mock external dependencies
4. Test behavior, not implementation
5. Aim for 80%+ coverage`,
		"refactoring": `
Refactoring principles:
1. Write tests first (safety net)
2. Make small, incremental changes
3. Run tests after each change
4. Don't mix refactoring with feature changes
5. Use IDE refactoring tools when available`,
		"documenting": `
Documentation tips:
1. Start with "why" before "how"
2. Include code examples
3. Keep it concise but complete
4. Update docs when code changes
5. Use clear, simple language`,
	}

	if p, ok := practices[taskType]; ok {
		return strings.TrimSpace(p)
	}
	return ""
}

// FindSimilar finds top-N similar prompt patterns.
func (pe *PatternExtractor) FindSimilar(prompt string, limit int) []PatternMatch {
	template := pe.Normalize(prompt)
	pe.mu.RLock()
	defer pe.mu.RUnlock()

	var matches []PatternMatch
	for _, p := range pe.patterns {
		if p.Template == template {
			// Exact match - highest score
			matches = append(matches, PatternMatch{
				Pattern:    p,
				Similarity: 1.0,
				Score:      p.AvgQuality,
			})
			continue
		}

		// Fuzzy: Levenshtein distance (simplified: word overlap)
		sim := levenshteinSimilarity(template, p.Template)
		if sim > 0.3 { // Threshold
			matches = append(matches, PatternMatch{
				Pattern:    p,
				Similarity: sim,
				Score:      p.AvgQuality * sim, // Weight by quality
			})
		}
	}

	// Sort by Score desc, take top N
	sortPatternMatches(matches)
	if len(matches) > limit {
		matches = matches[:limit]
	}

	return matches
}

// PatternMatch is a similar pattern result.
type PatternMatch struct {
	Pattern    *Pattern
	Similarity float64 // 0-1
	Score      float64 // quality * similarity
}

// levenshteinSimilarity computes similarity between two strings (0-1).
// Simplified: Jaccard index on word sets.
func levenshteinSimilarity(a, b string) float64 {
	wordsA := strings.Fields(a)
	wordsB := strings.Fields(b)

	if len(wordsA) == 0 && len(wordsB) == 0 {
		return 1.0
	}

	setA := make(map[string]bool)
	for _, w := range wordsA {
		setA[w] = true
	}

	overlap := 0
	for _, w := range wordsB {
		if setA[w] {
			overlap++
		}
	}

	total := len(wordsA) + len(wordsB) - overlap
	if total == 0 {
		return 0.0
	}
	return float64(overlap) / float64(total)
}

// sortPatternMatches sorts by Score desc.
func sortPatternMatches(matches []PatternMatch) {
	for i := 0; i < len(matches); i++ {
		for j := i + 1; j < len(matches); j++ {
			if matches[j].Score > matches[i].Score {
				matches[i], matches[j] = matches[j], matches[i]
			}
		}
	}
}

// GetAllPatterns returns all learned patterns (for stats).
func (pe *PatternExtractor) GetAllPatterns() []Pattern {
	pe.mu.RLock()
	defer pe.mu.RUnlock()

	patterns := make([]Pattern, 0, len(pe.patterns))
	for _, p := range pe.patterns {
		patterns = append(patterns, *p)
	}
	return patterns
}

// LoadFromDB loads patterns from storage.
func (pe *PatternExtractor) LoadFromDB() error {
	pe.mu.Lock()
	defer pe.mu.Unlock()

	rows, err := pe.storage.db.Query(`
		SELECT pattern_hash, pattern_template, task_type, task_domain,
		       success_count, failure_count, avg_quality, best_model,
		       first_seen, last_used
		FROM prompt_patterns
		ORDER BY success_count + failure_count DESC
	`)
	if err != nil {
		return err
	}
	defer rows.Close()

	pe.patterns = make(map[string]*Pattern)
	for rows.Next() {
		var p Pattern
		var hash string
		err := rows.Scan(
			&hash, &p.Template, &p.TaskType, &p.TaskDomain,
			&p.SuccessCount, &p.FailureCount, &p.AvgQuality, &p.BestModel,
			&p.FirstSeen, &p.LastUsed,
		)
		if err != nil {
			return err
		}
		p.ID = hash[:12]
		pe.patterns[p.Template] = &p
	}

	return nil
}
