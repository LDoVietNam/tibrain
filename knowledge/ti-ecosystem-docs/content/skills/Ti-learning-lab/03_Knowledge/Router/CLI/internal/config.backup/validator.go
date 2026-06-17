package config

import (
	"fmt"
	"os"
	"path/filepath"
)

func Validate(cfg *Config) error {
	// Validate log level
	validLogLevels := map[string]bool{
		"debug": true,
		"info":  true,
		"warn":  true,
		"error": true,
	}
	if !validLogLevels[cfg.LogLevel] {
		return fmt.Errorf("invalid log level: %s", cfg.LogLevel)
	}

	// Validate plugin directory
	if cfg.PluginDir == "" {
		return fmt.Errorf("plugin directory cannot be empty")
	}
	if !filepath.IsAbs(cfg.PluginDir) {
		// Convert to absolute path
		abs, err := filepath.Abs(cfg.PluginDir)
		if err != nil {
			return fmt.Errorf("error converting plugin directory to absolute path: %w", err)
		}
		cfg.PluginDir = abs
	}

	// Validate GRPC config
	if cfg.GRPC.Port < 1 || cfg.GRPC.Port > 65535 {
		return fmt.Errorf("invalid gRPC port: %d", cfg.GRPC.Port)
	}
	if cfg.GRPC.Host == "" {
		return fmt.Errorf("gRPC host cannot be empty")
	}

	// Validate TLS config
	if cfg.GRPC.EnableTLS || cfg.GRPC.EnableMTLS {
		if cfg.GRPC.CertFile == "" {
			return fmt.Errorf("cert file required when TLS is enabled")
		}
		if !fileExists(cfg.GRPC.CertFile) {
			return fmt.Errorf("cert file not found: %s", cfg.GRPC.CertFile)
		}
	}

	if cfg.GRPC.EnableMTLS {
		if cfg.GRPC.KeyFile == "" {
			return fmt.Errorf("key file required when mTLS is enabled")
		}
		if !fileExists(cfg.GRPC.KeyFile) {
			return fmt.Errorf("key file not found: %s", cfg.GRPC.KeyFile)
		}
		if cfg.GRPC.CAFile == "" {
			return fmt.Errorf("CA file required when mTLS is enabled")
		}
		if !fileExists(cfg.GRPC.CAFile) {
			return fmt.Errorf("CA file not found: %s", cfg.GRPC.CAFile)
		}
	}

	// Validate resource limits
	if cfg.Security.ResourceLimits.MaxMemoryMB < 64 {
		return fmt.Errorf("max memory must be at least 64MB")
	}
	if cfg.Security.ResourceLimits.MaxCPU < 1 {
		return fmt.Errorf("max CPU must be at least 1")
	}
	if cfg.Security.ResourceLimits.MaxTimeout < 10 {
		return fmt.Errorf("max timeout must be at least 10 seconds")
	}

	return nil
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
