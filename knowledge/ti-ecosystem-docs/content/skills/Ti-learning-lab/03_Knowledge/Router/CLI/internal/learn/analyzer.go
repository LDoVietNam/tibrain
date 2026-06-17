package learn

import (
	"crypto/sha256"
	"encoding/hex"
	"regexp"
	"strings"
)

// ClassifyPattern maps regex → task type.
type ClassifyPattern struct {
	TaskType string
	Pattern  *regexp.Regexp
}

// TaskClassifier determines task type/domain from prompt.
type TaskClassifier struct {
	patterns       []ClassifyPattern
	domainKeywords map[string][]string
}

// TaskClassifier classifies prompts into task types and domains.
var taskClassifier = &TaskClassifier{
	patterns: []ClassifyPattern{
		// Order matters: most specific first
		{"debugging", regexp.MustCompile(`(?i)\b(debug|fix.*bug|why.*not.*work|error.*message|crash|stack.*trace|panic|loi|lỗi|troubleshoot)\b`)},
		{"testing", regexp.MustCompile(`(?i)\b(test|unit.*test|write.*test|test.*case|coverage|assert|mock|vitest|pytest)\b`)},
		{"refactoring", regexp.MustCompile(`(?i)\b(refactor|improve.*code|optimize|clean.*up|restructure|simplify|reorganize)\b`)},
		{"review", regexp.MustCompile(`(?i)\b(review|audit|inspect|evaluate|lint|code.*review|security.*audit)\b`)},
		{"coding", regexp.MustCompile(`(?i)\b(code|implement|write.*function|create.*script|build|compile|sua|sửa|develop)\b`)},
		{"documenting", regexp.MustCompile(`(?i)\b(document|write.*doc|README|api.*doc|comment|explain.*code)\b`)},
		{"researching", regexp.MustCompile(`(?i)\b(research|investigate|find.*information|analyze|study|compare)\b`)},
		{"planning", regexp.MustCompile(`(?i)\b(plan|design.*out|architect|roadmap|strategy|pipeline)\b`)},
	},
	domainKeywords: map[string][]string{
		"backend":  {"api", "server", "database", "sql", "query", "redis", "postgres", "mysql", "backend", "service", "handler", "middleware"},
		"frontend": {"react", "vue", "angular", "component", "ui", "css", "html", "dom", "frontend", "client", "browser"},
		"database": {"sql", "query", "index", "migration", "schema", "database", "join", "nosql", "mongodb", "postgres", "mysql"},
		"devops":   {"docker", "kubernetes", "deploy", "ci/cd", "pipeline", "terraform", "cloud", "aws", "gcp", "azure"},
		"ml":       {"machine.*learning", "model", "train", "dataset", "pytorch", "tensorflow", "neural", "ai"},
		"mobile":   {"android", "ios", "swift", "kotlin", "react.*native", "flutter", "mobile"},
		"security": {"auth", "authentication", "authorization", "jwt", "oauth", "security", "encryption", "vulnerability"},
	},
}

// NewTaskClassifier returns a new classifier.
func NewTaskClassifier() *TaskClassifier {
	return taskClassifier
}

// Classify analyzes a prompt and returns task type + domain.
func (c *TaskClassifier) Classify(prompt string) TaskAnalysis {
	promptLower := strings.ToLower(prompt)

	// 1. Task type (first match wins - ordered by specificity)
	taskType := "general"
	for _, p := range c.patterns {
		if p.Pattern.MatchString(promptLower) {
			taskType = p.TaskType
			break
		}
	}

	// 2. Domain (count keyword matches)
	domain := "general"
	maxMatches := 0
	for d, keywords := range c.domainKeywords {
		matches := 0
		for _, kw := range keywords {
			if strings.Contains(promptLower, kw) {
				matches++
			}
		}
		if matches > maxMatches {
			maxMatches = matches
			domain = d
		}
	}
	if maxMatches == 0 {
		domain = "general"
	}

	// 3. Complexity (1-10)
	complexity := c.estimateComplexity(prompt, taskType, domain)

	// 4. Keywords extraction
	keywords := c.extractKeywords(prompt)

	// 5. Hash
	promptHash := hashPrompt(prompt)

	return TaskAnalysis{
		TaskType:        taskType,
		TaskDomain:      domain,
		Complexity:      complexity,
		Keywords:        keywords,
		PromptHash:      promptHash,
		PromptLength:    len(prompt),
		EstimatedTokens: len(prompt) / 4, // Rough estimate: 4 chars/token
	}
}

// estimateComplexity heuristically scores task difficulty 1-10.
func (c *TaskClassifier) estimateComplexity(prompt, taskType, domain string) float64 {
	score := 3.0 // Base: medium

	// Length factor: longer prompts often more complex
	words := len(strings.Fields(prompt))
	if words > 50 {
		score += 2
	} else if words > 20 {
		score += 1
	}

	// Task type modifiers
	complexityByType := map[string]float64{
		"coding":      6.0,
		"debugging":   5.0,
		"refactoring": 7.0,
		"testing":     4.0,
		"documenting": 2.0,
		"researching": 5.0,
		"planning":    6.0,
		"review":      5.0,
		"general":     3.0,
	}
	if base, ok := complexityByType[taskType]; ok {
		score = (score + base) / 2
	}

	// Domain modifiers
	complexityByDomain := map[string]float64{
		"database": 6.0,
		"devops":   7.0,
		"ml":       9.0,
		"security": 8.0,
		"backend":  5.0,
		"frontend": 4.0,
		"mobile":   6.0,
		"general":  3.0,
	}
	if base, ok := complexityByDomain[domain]; ok {
		score = (score + base) / 2
	}

	// Clamp 1-10
	if score < 1 {
		score = 1
	} else if score > 10 {
		score = 10
	}

	return score
}

// extractKeywords pulls important keywords from prompt.
func (c *TaskClassifier) extractKeywords(prompt string) []string {
	// Remove stop words
	stopWords := map[string]bool{
		"the": true, "a": true, "an": true, "for": true, "to": true,
		"and": true, "or": true, "but": true, "in": true, "on": true,
		"at": true, "by": true, "with": true, "from": true,
		"write": true, "create": true, "make": true, "get": true,
		"set": true, "use": true, "using": true,
	}

	words := strings.Fields(strings.ToLower(prompt))
	keywords := make([]string, 0, len(words)/2)

	for _, word := range words {
		// Clean punctuation
		word = strings.Trim(word, ".,;:!?()[]{}<>\"'")
		if len(word) > 2 && !stopWords[word] {
			keywords = append(keywords, word)
		}
	}

	// Deduplicate
	seen := make(map[string]bool)
	unique := make([]string, 0, len(keywords))
	for _, kw := range keywords {
		if !seen[kw] {
			seen[kw] = true
			unique = append(unique, kw)
		}
	}

	return unique
}

// hashPrompt creates a SHA256 hash of the prompt for deduplication.
func hashPrompt(prompt string) string {
	h := sha256.New()
	h.Write([]byte(prompt))
	return hex.EncodeToString(h.Sum(nil))
}

// AnalyzePrompt is a convenience function.
func AnalyzePrompt(prompt string) TaskAnalysis {
	return taskClassifier.Classify(prompt)
}
