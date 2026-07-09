package main

import (
	"context"
	"fmt"
	"strings"
)

// QualityGate evaluates feedback to prevent data poisoning and low-quality learning
type QualityGate struct {
	alm *AutoLearningMechanism
}

// NewQualityGate creates a new QualityGate
func NewQualityGate(alm *AutoLearningMechanism) *QualityGate {
	return &QualityGate{
		alm: alm,
	}
}

// ProcessFeedbackWithQualityCheck intercepts feedback, validates it, and then passes it to ALM
func (qg *QualityGate) ProcessFeedbackWithQualityCheck(ctx context.Context, queryID string, rating int, correction string, comment string) error {
	// 1. Basic validation
	if rating < 1 || rating > 5 {
		return fmt.Errorf("invalid rating: %d, must be between 1 and 5", rating)
	}

	// 2. Data Poisoning Detection
	// Check for obvious spam or malicious corrections
	if isMalicious(correction) || isMalicious(comment) {
		fmt.Printf("🛡️ QualityGate: Blocked malicious feedback for query %s\n", queryID)
		return fmt.Errorf("feedback rejected due to malicious content")
	}

	// 3. Low Effort Correction Detection
	if rating <= 2 && strings.TrimSpace(correction) == "" && len(strings.TrimSpace(comment)) < 5 {
		fmt.Printf("🛡️ QualityGate: Ignored low-effort negative feedback for query %s\n", queryID)
		// We can return nil to silently drop it, avoiding error logs
		return nil
	}

	// 4. Pass to ALM
	return qg.alm.ProcessFeedback(queryID, rating, correction, comment)
}

func isMalicious(text string) bool {
	lowerText := strings.ToLower(text)
	// Simple heuristics for data poisoning prevention
	maliciousKeywords := []string{"drop table", "1=1", "<script>", "ignore all previous instructions", "rm -rf"}
	for _, kw := range maliciousKeywords {
		if strings.Contains(lowerText, kw) {
			return true
		}
	}
	return false
}
