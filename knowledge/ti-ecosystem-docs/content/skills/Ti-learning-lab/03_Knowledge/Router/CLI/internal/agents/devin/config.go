package devin

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Config holds configuration for the Devin agent
type Config struct {
	// gRPC server configuration
	Address string
	Port    int

	// Connection configuration
	Timeout        time.Duration
	ConnectTimeout time.Duration
	MaxRetries     int

	// TLS configuration
	EnableTLS bool
	CertFile  string
	KeyFile   string

	// Python bridge configuration
	PythonPath       string
	PythonExecutable string

	// Devin CLI configuration
	DevinCLIPath string
	DevinAPIKey  string
	DevinBaseURL string
	DevinModel   string

	// Health check configuration
	HealthCheckInterval time.Duration

	// Lifecycle configuration
	AutoRestart      bool
	AutoRestartDelay time.Duration

	// Logging configuration
	LogLevel string
	LogFile  string
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	return &Config{
		Address:             "localhost",
		Port:                50052,
		Timeout:             30 * time.Second,
		ConnectTimeout:      10 * time.Second,
		MaxRetries:          3,
		EnableTLS:           false,
		PythonExecutable:    "python",
		PythonPath:          "",
		DevinCLIPath:        "",
		DevinAPIKey:         "",
		DevinBaseURL:        "",
		DevinModel:          "",
		HealthCheckInterval: 30 * time.Second,
		AutoRestart:         true,
		AutoRestartDelay:    5 * time.Second,
		LogLevel:            "info",
		LogFile:             "",
	}
}

// LoadConfig loads configuration from environment variables and defaults
func LoadConfig() (*Config, error) {
	config := DefaultConfig()

	// Load from environment variables
	if addr := os.Getenv("DEVIN_ADDRESS"); addr != "" {
		config.Address = addr
	}

	if port := os.Getenv("DEVIN_PORT"); port != "" {
		fmt.Sscanf(port, "%d", &config.Port)
	}

	if timeout := os.Getenv("DEVIN_TIMEOUT"); timeout != "" {
		if duration, err := time.ParseDuration(timeout); err == nil {
			config.Timeout = duration
		}
	}

	if enableTLS := os.Getenv("DEVIN_ENABLE_TLS"); enableTLS == "true" {
		config.EnableTLS = true
	}

	if certFile := os.Getenv("DEVIN_CERT_FILE"); certFile != "" {
		config.CertFile = certFile
	}

	if keyFile := os.Getenv("DEVIN_KEY_FILE"); keyFile != "" {
		config.KeyFile = keyFile
	}

	if pythonExec := os.Getenv("DEVIN_PYTHON_EXEC"); pythonExec != "" {
		config.PythonExecutable = pythonExec
	}

	if pythonPath := os.Getenv("DEVIN_PYTHON_PATH"); pythonPath != "" {
		config.PythonPath = pythonPath
	}

	if devinCLIPath := os.Getenv("DEVIN_CLI_PATH"); devinCLIPath != "" {
		config.DevinCLIPath = devinCLIPath
	}

	if apiKey := os.Getenv("DEVIN_API_KEY"); apiKey != "" {
		config.DevinAPIKey = apiKey
	}

	if baseURL := os.Getenv("DEVIN_BASE_URL"); baseURL != "" {
		config.DevinBaseURL = baseURL
	}

	if model := os.Getenv("DEVIN_MODEL"); model != "" {
		config.DevinModel = model
	}

	if healthInterval := os.Getenv("DEVIN_HEALTH_INTERVAL"); healthInterval != "" {
		if duration, err := time.ParseDuration(healthInterval); err == nil {
			config.HealthCheckInterval = duration
		}
	}

	if autoRestart := os.Getenv("DEVIN_AUTO_RESTART"); autoRestart == "false" {
		config.AutoRestart = false
	}

	if logLevel := os.Getenv("DEVIN_LOG_LEVEL"); logLevel != "" {
		config.LogLevel = logLevel
	}

	if logFile := os.Getenv("DEVIN_LOG_FILE"); logFile != "" {
		config.LogFile = logFile
	}

	// Resolve paths
	if err := config.resolvePaths(); err != nil {
		return nil, fmt.Errorf("failed to resolve paths: %w", err)
	}

	return config, nil
}

// resolvePaths resolves relative paths to absolute paths
func (c *Config) resolvePaths() error {
	// Resolve Python path
	if c.PythonPath != "" && !filepath.IsAbs(c.PythonPath) {
		absPath, err := filepath.Abs(c.PythonPath)
		if err != nil {
			return err
		}
		c.PythonPath = absPath
	}

	// Resolve Devin CLI path
	if c.DevinCLIPath != "" && !filepath.IsAbs(c.DevinCLIPath) {
		absPath, err := filepath.Abs(c.DevinCLIPath)
		if err != nil {
			return err
		}
		c.DevinCLIPath = absPath
	}

	// Resolve cert file
	if c.CertFile != "" && !filepath.IsAbs(c.CertFile) {
		absPath, err := filepath.Abs(c.CertFile)
		if err != nil {
			return err
		}
		c.CertFile = absPath
	}

	// Resolve key file
	if c.KeyFile != "" && !filepath.IsAbs(c.KeyFile) {
		absPath, err := filepath.Abs(c.KeyFile)
		if err != nil {
			return err
		}
		c.KeyFile = absPath
	}

	// Resolve log file
	if c.LogFile != "" && !filepath.IsAbs(c.LogFile) {
		absPath, err := filepath.Abs(c.LogFile)
		if err != nil {
			return err
		}
		c.LogFile = absPath
	}

	return nil
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.Port <= 0 || c.Port > 65535 {
		return fmt.Errorf("invalid port: %d", c.Port)
	}

	if c.Timeout <= 0 {
		return fmt.Errorf("invalid timeout: %v", c.Timeout)
	}

	if c.ConnectTimeout <= 0 {
		return fmt.Errorf("invalid connect timeout: %v", c.ConnectTimeout)
	}

	if c.EnableTLS {
		if c.CertFile == "" {
			return fmt.Errorf("TLS enabled but cert file not specified")
		}
		if c.KeyFile == "" {
			return fmt.Errorf("TLS enabled but key file not specified")
		}
	}

	if c.PythonExecutable == "" {
		return fmt.Errorf("Python executable not specified")
	}

	return nil
}

// GetGRPCAddress returns the full gRPC address
func (c *Config) GetGRPCAddress() string {
	return fmt.Sprintf("%s:%d", c.Address, c.Port)
}
