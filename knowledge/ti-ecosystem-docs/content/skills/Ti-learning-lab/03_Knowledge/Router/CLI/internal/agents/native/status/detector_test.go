package status

import (
	"context"
	"testing"
	"time"
)

func TestNewDetector(t *testing.T) {
	config := &Config{
		CacheEnabled: true,
		CacheTTL:     5 * time.Second,
	}

	detector := NewDetector(config)
	if detector == nil {
		t.Fatal("Expected non-nil detector")
	}

	if detector.patterns == nil {
		t.Error("Expected patterns to be initialized")
	}
}

func TestDetectStatus(t *testing.T) {
	detector := NewDetector(nil)

	output := "Status: [idle] Ready for commands"
	result, err := detector.DetectStatus(context.Background(), output, "test-session")

	if err != nil {
		t.Fatalf("DetectStatus failed: %v", err)
	}

	if result.Status != "idle" {
		t.Errorf("Expected status 'idle', got '%s'", result.Status)
	}
}

func TestDetectStatusBusy(t *testing.T) {
	detector := NewDetector(nil)

	output := "Status: [busy] Processing task"
	result, err := detector.DetectStatus(context.Background(), output, "test-session")

	if err != nil {
		t.Fatalf("DetectStatus failed: %v", err)
	}

	if result.Status != "busy" {
		t.Errorf("Expected status 'busy', got '%s'", result.Status)
	}
}

func TestDetectCompleted(t *testing.T) {
	detector := NewDetector(nil)

	output := "Done"
	result, err := detector.DetectStatus(context.Background(), output, "test-session")

	if err != nil {
		t.Fatalf("DetectStatus failed: %v", err)
	}

	if !result.IsCompleted {
		t.Error("Expected IsCompleted to be true")
	}
}

func TestDetectCompletedPrompt(t *testing.T) {
	detector := NewDetector(nil)

	output := "Done\n>"
	result, err := detector.DetectStatus(context.Background(), output, "test-session")

	if err != nil {
		t.Fatalf("DetectStatus failed: %v", err)
	}

	if !result.IsCompleted {
		t.Error("Expected IsCompleted to be true for prompt")
	}
}

func TestDetectCombinedStatus(t *testing.T) {
	detector := NewDetector(nil)

	prompt := "Fix the bug"
	statusBar := "[busy] Working on it"

	result, err := detector.DetectCombinedStatus(context.Background(), prompt, statusBar, "test-session")

	if err != nil {
		t.Fatalf("DetectCombinedStatus failed: %v", err)
	}

	if result.Status != "busy" {
		t.Errorf("Expected status 'busy', got '%s'", result.Status)
	}

	if result.Metadata["detection_mode"] != "combined" {
		t.Error("Expected detection_mode to be 'combined'")
	}
}

func TestCache(t *testing.T) {
	config := &Config{
		CacheEnabled: true,
		CacheTTL:     1 * time.Second,
	}
	detector := NewDetector(config)

	output := "Status: [idle]"
	sessionID := "cache-test"

	// First call - should populate cache
	result1, err := detector.DetectStatus(context.Background(), output, sessionID)
	if err != nil {
		t.Fatalf("First DetectStatus failed: %v", err)
	}

	// Second call - should use cache
	result2, err := detector.DetectStatus(context.Background(), output, sessionID)
	if err != nil {
		t.Fatalf("Second DetectStatus failed: %v", err)
	}

	if result1.Status != result2.Status {
		t.Error("Cached result should match")
	}

	// Wait for cache to expire
	time.Sleep(2 * time.Second)

	// Third call - cache expired
	result3, err := detector.DetectStatus(context.Background(), output, sessionID)
	if err != nil {
		t.Fatalf("Third DetectStatus failed: %v", err)
	}

	// Should still work but with new timestamp
	if result3.Timestamp.Equal(result1.Timestamp) {
		t.Error("Expected new timestamp after cache expiration")
	}
}

func TestClearCache(t *testing.T) {
	config := &Config{
		CacheEnabled: true,
	}
	detector := NewDetector(config)

	output := "Status: [idle]"
	sessionID := "cache-clear-test"

	// Populate cache
	detector.DetectStatus(context.Background(), output, sessionID)

	// Clear cache
	detector.ClearCache()

	// Verify cache is cleared
	_, ok := detector.getFromCache(sessionID)
	if ok {
		t.Error("Expected cache to be cleared")
	}
}

func TestConfidenceCalculation(t *testing.T) {
	detector := NewDetector(nil)

	// High confidence case
	output := "Done"
	result, err := detector.DetectStatus(context.Background(), output, "test-session")
	if err != nil {
		t.Fatalf("DetectStatus failed: %v", err)
	}

	if result.Confidence < 0.6 {
		t.Errorf("Expected reasonable confidence, got %.2f", result.Confidence)
	}

	// Low confidence case
	emptyOutput := ""
	result2, err := detector.DetectStatus(context.Background(), emptyOutput, "test-session2")
	if err != nil {
		t.Fatalf("DetectStatus failed: %v", err)
	}

	if result2.Confidence > 0.5 {
		t.Errorf("Expected low confidence for empty output, got %.2f", result2.Confidence)
	}
}
