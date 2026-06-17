package status

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"
)

// Detector performs status detection for agent execution
type Detector struct {
	mu                    sync.RWMutex
	patterns              map[string]*regexp.Regexp
	completedPatterns     map[string]*regexp.Regexp
	falsePositivePatterns map[string]*regexp.Regexp
	cache                 map[string]*DetectionResult
	cacheTTL              time.Duration
}

// DetectionResult represents the result of status detection
type DetectionResult struct {
	Status      string
	IsCompleted bool
	IsIdle      bool
	Confidence  float64
	Metadata    map[string]string
	Timestamp   time.Time
}

// Config holds detector configuration
type Config struct {
	CacheEnabled                  bool
	CacheTTL                      time.Duration
	EnableFalsePositivePrevention bool
}

// NewDetector creates a new status detector
func NewDetector(config *Config) *Detector {
	if config == nil {
		config = &Config{
			CacheEnabled:                  true,
			CacheTTL:                      5 * time.Second,
			EnableFalsePositivePrevention: true,
		}
	}

	detector := &Detector{
		patterns:              make(map[string]*regexp.Regexp),
		completedPatterns:     make(map[string]*regexp.Regexp),
		falsePositivePatterns: make(map[string]*regexp.Regexp),
		cache:                 make(map[string]*DetectionResult),
		cacheTTL:              config.CacheTTL,
	}

	detector.initPatterns()
	return detector
}

// initPatterns initializes regex patterns for status detection
func (d *Detector) initPatterns() {
	// Status patterns - detecting current status
	d.patterns["idle"] = regexp.MustCompile(`(?i)(idle|waiting|ready|listening)`)
	d.patterns["busy"] = regexp.MustCompile(`(?i)(busy|working|processing|executing|running)`)
	d.patterns["error"] = regexp.MustCompile(`(?i)(error|failed|exception|crashed)`)
	d.patterns["completed"] = regexp.MustCompile(`(?i)(completed|done|finished|success)`)

	// Combined status detection patterns
	d.patterns["status_bar"] = regexp.MustCompile(`\[(.*?)\]`) // Matches [status] patterns
	d.patterns["progress"] = regexp.MustCompile(`\d+/%d|\d+%`) // Matches progress indicators

	// Completed detection patterns (stateless)
	d.completedPatterns["done"] = regexp.MustCompile(`(?i)^\s*done\s*$`)
	d.completedPatterns["finished"] = regexp.MustCompile(`(?i)^\s*finished\s*$`)
	d.completedPatterns["completed"] = regexp.MustCompile(`(?i)^\s*completed\s*$`)
	d.completedPatterns["success"] = regexp.MustCompile(`(?i)^\s*success\s*$`)
	d.completedPatterns["prompt"] = regexp.MustCompile(`(?i)^\s*>\s*$`) // Prompt ready

	// False positive prevention patterns
	d.falsePositivePatterns["loading"] = regexp.MustCompile(`(?i)(loading|initializing|starting)`)
	d.falsePositivePatterns["info"] = regexp.MustCompile(`(?i)(info|notice|warning)`)
}

// DetectStatus detects the current status from output
func (d *Detector) DetectStatus(ctx context.Context, output string, sessionID string) (*DetectionResult, error) {
	// Check cache first
	if d.cache != nil {
		if cached, ok := d.getFromCache(sessionID); ok {
			return cached, nil
		}
	}

	result := &DetectionResult{
		Status:    "unknown",
		Metadata:  make(map[string]string),
		Timestamp: time.Now(),
	}

	// Detect status from output
	lines := strings.Split(output, "\n")
	lastLine := ""
	if len(lines) > 0 {
		lastLine = strings.TrimSpace(lines[len(lines)-1])
	}

	// Check status bar
	if matches := d.patterns["status_bar"].FindAllStringSubmatch(output, -1); len(matches) > 0 {
		for _, match := range matches {
			if len(match) > 1 {
				result.Metadata["status_bar"] = match[1]
				result.Status = d.detectStatusFromString(match[1])
			}
		}
	}

	// Check for explicit status keywords
	for statusName, pattern := range d.patterns {
		if statusName == "status_bar" || statusName == "progress" {
			continue
		}
		if pattern.MatchString(lastLine) {
			result.Status = statusName
			break
		}
	}

	// Check if completed
	result.IsCompleted = d.detectCompleted(output, lastLine)

	// Check if idle
	result.IsIdle = (result.Status == "idle" || result.Status == "completed") && result.IsCompleted

	// Calculate confidence
	result.Confidence = d.calculateConfidence(result, output)

	// Cache result
	d.addToCache(sessionID, result)

	return result, nil
}

// detectStatusFromString detects status from status bar string
func (d *Detector) detectStatusFromString(statusBar string) string {
	statusBar = strings.ToLower(statusBar)

	if strings.Contains(statusBar, "busy") || strings.Contains(statusBar, "working") {
		return "busy"
	}
	if strings.Contains(statusBar, "idle") || strings.Contains(statusBar, "waiting") {
		return "idle"
	}
	if strings.Contains(statusBar, "error") {
		return "error"
	}
	if strings.Contains(statusBar, "done") || strings.Contains(statusBar, "completed") {
		return "completed"
	}

	return "unknown"
}

// detectCompleted detects if the task is completed (stateless)
func (d *Detector) detectCompleted(output, lastLine string) bool {
	// Check completed patterns in last line
	for _, pattern := range d.completedPatterns {
		if pattern.MatchString(lastLine) {
			return true
		}
	}

	// Check for prompt character (indicates ready for next command)
	if strings.HasSuffix(strings.TrimSpace(lastLine), ">") || strings.HasSuffix(strings.TrimSpace(lastLine), "$") {
		return true
	}

	// Check for completion in full output
	for _, pattern := range d.completedPatterns {
		if pattern.MatchString(output) {
			return true
		}
	}

	return false
}

// calculateConfidence calculates confidence score for detection
func (d *Detector) calculateConfidence(result *DetectionResult, output string) float64 {
	confidence := 0.5 // Base confidence

	// Increase confidence if we have explicit status
	if result.Status != "unknown" {
		confidence += 0.3
	}

	// Increase confidence if completed state is clear
	if result.IsCompleted {
		confidence += 0.1
	}

	// Decrease confidence if output is empty
	if len(strings.TrimSpace(output)) == 0 {
		confidence -= 0.3
	}

	// Apply false positive prevention
	if d.shouldPreventFalsePositive(result, output) {
		confidence -= 0.2
	}

	// Ensure confidence is in [0, 1] range
	if confidence < 0 {
		confidence = 0
	}
	if confidence > 1 {
		confidence = 1
	}

	return confidence
}

// shouldPreventFalsePositive checks if this might be a false positive
func (d *Detector) shouldPreventFalsePositive(result *DetectionResult, output string) bool {
	if !d.falsePositivePatterns["loading"].MatchString(output) {
		// Still loading, might not be actually completed
		return true
	}

	if result.IsCompleted && d.falsePositivePatterns["info"].MatchString(output) {
		// Info message might look like completion
		return true
	}

	return false
}

// getFromCache retrieves cached result
func (d *Detector) getFromCache(sessionID string) (*DetectionResult, bool) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	cached, ok := d.cache[sessionID]
	if !ok {
		return nil, false
	}

	// Check if cache is expired
	if time.Since(cached.Timestamp) > d.cacheTTL {
		delete(d.cache, sessionID)
		return nil, false
	}

	return cached, true
}

// addToCache adds result to cache
func (d *Detector) addToCache(sessionID string, result *DetectionResult) {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.cache[sessionID] = result
}

// ClearCache clears the detection cache
func (d *Detector) ClearCache() {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.cache = make(map[string]*DetectionResult)
}

// DetectCombinedStatus performs combined status detection (prompt + status bar)
func (d *Detector) DetectCombinedStatus(ctx context.Context, prompt, statusBar, sessionID string) (*DetectionResult, error) {
	combinedOutput := fmt.Sprintf("%s\n%s", prompt, statusBar)

	result, err := d.DetectStatus(ctx, combinedOutput, sessionID)
	if err != nil {
		return nil, err
	}

	// Add combined detection metadata
	result.Metadata["detection_mode"] = "combined"
	result.Metadata["has_prompt"] = fmt.Sprintf("%v", prompt != "")
	result.Metadata["has_status_bar"] = fmt.Sprintf("%v", statusBar != "")

	return result, nil
}
