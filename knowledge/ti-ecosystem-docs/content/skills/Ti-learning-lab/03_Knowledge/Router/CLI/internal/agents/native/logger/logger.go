package logger

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"
	"time"
)

// Level represents log level
type Level int

const (
	LevelDebug Level = iota
	LevelInfo
	LevelWarn
	LevelError
	LevelFatal
)

// String returns string representation of log level
func (l Level) String() string {
	return [...]string{"DEBUG", "INFO", "WARN", "ERROR", "FATAL"}[l]
}

// Logger provides structured logging capabilities
type Logger struct {
	mu            sync.RWMutex
	level         Level
	prefix        string
	outputFile    *os.File
	patterns      map[string]*LogPattern
	enablePattern bool
}

// Config holds logger configuration
type Config struct {
	Level          Level
	Prefix         string
	OutputFile     string
	EnablePatterns bool
}

// LogPattern represents a log pattern for detection
type LogPattern struct {
	Name     string
	Pattern  string
	Level    Level
	Callback func(string)
}

// LogEntry represents a log entry
type LogEntry struct {
	Timestamp time.Time
	Level     Level
	Message   string
	Fields    map[string]interface{}
}

// NewLogger creates a new logger
func NewLogger(config *Config) (*Logger, error) {
	if config == nil {
		config = &Config{
			Level:          LevelInfo,
			Prefix:         "ti-cli",
			OutputFile:     "",
			EnablePatterns: true,
		}
	}

	logger := &Logger{
		level:         config.Level,
		prefix:        config.Prefix,
		patterns:      make(map[string]*LogPattern),
		enablePattern: config.EnablePatterns,
	}

	// Setup output file if specified
	if config.OutputFile != "" {
		file, err := os.OpenFile(config.OutputFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return nil, fmt.Errorf("failed to open output file: %w", err)
		}
		logger.outputFile = file
		log.SetOutput(file)
	}

	// Initialize default patterns
	logger.initDefaultPatterns()

	return logger, nil
}

// initDefaultPatterns initializes default log patterns
func (l *Logger) initDefaultPatterns() {
	l.patterns["error"] = &LogPattern{
		Name:    "error",
		Pattern: `(?i)error|exception|failed`,
		Level:   LevelError,
	}

	l.patterns["warning"] = &LogPattern{
		Name:    "warning",
		Pattern: `(?i)warning|warn`,
		Level:   LevelWarn,
	}

	l.patterns["info"] = &LogPattern{
		Name:    "info",
		Pattern: `(?i)info|notice`,
		Level:   LevelInfo,
	}
}

// Debug logs a debug message
func (l *Logger) Debug(ctx context.Context, message string, fields map[string]interface{}) {
	l.log(ctx, LevelDebug, message, fields)
}

// Info logs an info message
func (l *Logger) Info(ctx context.Context, message string, fields map[string]interface{}) {
	l.log(ctx, LevelInfo, message, fields)
}

// Warn logs a warning message
func (l *Logger) Warn(ctx context.Context, message string, fields map[string]interface{}) {
	l.log(ctx, LevelWarn, message, fields)
}

// Error logs an error message
func (l *Logger) Error(ctx context.Context, message string, fields map[string]interface{}) {
	l.log(ctx, LevelError, message, fields)
}

// Fatal logs a fatal message and exits
func (l *Logger) Fatal(ctx context.Context, message string, fields map[string]interface{}) {
	l.log(ctx, LevelFatal, message, fields)
	os.Exit(1)
}

// log is the internal logging method
func (l *Logger) log(ctx context.Context, level Level, message string, fields map[string]interface{}) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	// Check if level is enabled
	if level < l.level {
		return
	}

	entry := LogEntry{
		Timestamp: time.Now(),
		Level:     level,
		Message:   message,
		Fields:    fields,
	}

	// Format log message
	formatted := l.formatEntry(entry)

	// Log to standard logger
	log.Println(formatted)

	// Check patterns if enabled
	if l.enablePattern {
		l.checkPatterns(entry)
	}
}

// formatEntry formats a log entry
func (l *Logger) formatEntry(entry LogEntry) string {
	timestamp := entry.Timestamp.Format("2006-01-02 15:04:05")
	level := entry.Level.String()

	formatted := fmt.Sprintf("%s [%s] %s: %s", timestamp, level, l.prefix, entry.Message)

	// Add fields
	if len(entry.Fields) > 0 {
		formatted += " |"
		for key, value := range entry.Fields {
			formatted += fmt.Sprintf(" %s=%v", key, value)
		}
	}

	return formatted
}

// checkPatterns checks if message matches any patterns
func (l *Logger) checkPatterns(entry LogEntry) {
	// Simple pattern matching (in production, use regex)
	message := entry.Message
	for _, pattern := range l.patterns {
		// This is simplified - in production use proper regex matching
		if contains(message, pattern.Pattern) {
			if pattern.Callback != nil {
				pattern.Callback(message)
			}
		}
	}
}

// AddPattern adds a log pattern
func (l *Logger) AddPattern(name, pattern string, level Level, callback func(string)) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.patterns[name] = &LogPattern{
		Name:     name,
		Pattern:  pattern,
		Level:    level,
		Callback: callback,
	}
}

// RemovePattern removes a log pattern
func (l *Logger) RemovePattern(name string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	delete(l.patterns, name)
}

// SetLevel sets the log level
func (l *Logger) SetLevel(level Level) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.level = level
}

// GetLevel returns the current log level
func (l *Logger) GetLevel() Level {
	l.mu.RLock()
	defer l.mu.RUnlock()

	return l.level
}

// Close closes the logger and releases resources
func (l *Logger) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.outputFile != nil {
		return l.outputFile.Close()
	}

	return nil
}

// WithFields adds fields to a log entry
func (l *Logger) WithFields(fields map[string]interface{}) map[string]interface{} {
	return fields
}

// DetectPatterns detects log patterns in output
func (l *Logger) DetectPatterns(output string) []string {
	l.mu.RLock()
	defer l.mu.RUnlock()

	var detected []string
	for _, pattern := range l.patterns {
		if contains(output, pattern.Pattern) {
			detected = append(detected, pattern.Name)
		}
	}

	return detected
}

// contains checks if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
