package learn

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPatternExtractor_Integration(t *testing.T) {
	storage := setupTestStorage(t)
	defer storage.Close()

	pe := NewPatternExtractor(storage)

	// Process some entries
	entries := []*LogInput{
		{
			Timestamp:       time.Now(),
			SessionID:       "s1",
			TaskType:        "testing",
			TaskDomain:      "backend",
			Prompt:          "Write unit test for UserService.CreateUser",
			QualityCombined: 9.0,
			Outcome:         "success",
		},
		{
			Timestamp:       time.Now(),
			SessionID:       "s2",
			TaskType:        "testing",
			TaskDomain:      "backend",
			Prompt:          "Create unit test for ProductService.Update",
			QualityCombined: 8.5,
			Outcome:         "success",
		},
		{
			Timestamp:       time.Now(),
			SessionID:       "s3",
			TaskType:        "testing",
			TaskDomain:      "backend",
			Prompt:          "Fix bug in auth middleware",
			QualityCombined: 6.0,
			Outcome:         "failed",
		},
	}

	for _, e := range entries {
		_ = pe.Process(e)
	}

	// Should have some patterns
	patterns := pe.GetAllPatterns()
	assert.Greater(t, len(patterns), 0)

	// Find similar
	similar := pe.FindSimilar("Write unit test for OrderService", "testing", 3)
	assert.Greater(t, len(similar), 0, "Should find similar testing patterns")
}

func TestPatternExtractor_BestPractices(t *testing.T) {
	pe := NewPatternExtractor(nil)

	// Get best practices for known task types
	p := pe.GetBestPractices("testing")
	assert.Contains(t, p, "table-driven")

	p = pe.GetBestPractices("debugging")
	assert.Contains(t, p, "root cause")

	p = pe.GetBestPractices("unknown")
	assert.Empty(t, p)
}

func TestPatternExtractor_SimilarityCalculation(t *testing.T) {
	pe := NewPatternExtractor(nil)

	// Identical prompts → similarity 1.0
	sim := levenshteinSimilarity("hello world", "hello world")
	assert.InEpsilon(t, 1.0, sim, 0.01)

	// Completely different → similarity 0
	sim = levenshteinSimilarity("hello", "world")
	assert.InEpsilon(t, 0.0, sim, 0.01)

	// Partial overlap
	sim = levenshteinSimilarity("write test for foo", "write test for bar")
	assert.Greater(t, sim, 0.3)
	assert.Less(t, sim, 0.8)
}
