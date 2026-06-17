package plugins

import (
	"log"
	"os"
)

var pluginLogger *log.Logger

func InitLogger(logLevel string) error {
	// Create logger with standard library
	pluginLogger = log.New(os.Stdout, "[TI-CLI] ", log.Ldate|log.Ltime|log.Lshortfile)
	return nil
}

func GetLogger() *log.Logger {
	if pluginLogger == nil {
		// Initialize with default
		_ = InitLogger("info")
	}
	return pluginLogger
}

func Sync() {
	// Standard library logger doesn't need sync
}

// Info logs an info message
func Info(msg string, args ...interface{}) {
	if pluginLogger != nil {
		pluginLogger.Printf("[INFO] "+msg, args...)
	}
}

// Warn logs a warning message
func Warn(msg string, args ...interface{}) {
	if pluginLogger != nil {
		pluginLogger.Printf("[WARN] "+msg, args...)
	}
}

// Error logs an error message
func Error(msg string, args ...interface{}) {
	if pluginLogger != nil {
		pluginLogger.Printf("[ERROR] "+msg, args...)
	}
}

// Debug logs a debug message
func Debug(msg string, args ...interface{}) {
	if pluginLogger != nil {
		pluginLogger.Printf("[DEBUG] "+msg, args...)
	}
}
