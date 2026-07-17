package logger

import (
	"fmt"
	"log"
	"sync"
	"time"
)

// LogLevel represents a log level.
type LogLevel int

const (
	LogLevelDebug LogLevel = iota
	LogLevelInfo
	LogLevelWarn
	LogLevelError
)

// Logger is the central structured logger.
type Logger struct {
	mu     sync.Mutex
	level  LogLevel
	prefix string
}

// NewLogger creates a new Logger.
func NewLogger(level LogLevel, prefix string) *Logger {
	return &Logger{level: level, prefix: prefix}
}

func (l *Logger) log(level LogLevel, format string, args ...interface{}) {
	if l == nil {
		log.Printf(format, args...)
		return
	}
	if level < l.level {
		return
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	timestamp := time.Now().Format("2006-01-02 15:04:05.000")
	levelStr := ""
	switch level {
	case LogLevelDebug:
		levelStr = "DEBUG"
	case LogLevelInfo:
		levelStr = "INFO"
	case LogLevelWarn:
		levelStr = "WARN"
	case LogLevelError:
		levelStr = "ERROR"
	}
	message := fmt.Sprintf(format, args...)
	log.Printf("[%s] [%s] [%s] %s", timestamp, levelStr, l.prefix, message)
}

func (l *Logger) Debug(format string, args ...interface{}) {
	l.log(LogLevelDebug, format, args...)
}

func (l *Logger) Info(format string, args ...interface{}) {
	l.log(LogLevelInfo, format, args...)
}

func (l *Logger) Warn(format string, args ...interface{}) {
	l.log(LogLevelWarn, format, args...)
}

func (l *Logger) Error(format string, args ...interface{}) {
	l.log(LogLevelError, format, args...)
}
