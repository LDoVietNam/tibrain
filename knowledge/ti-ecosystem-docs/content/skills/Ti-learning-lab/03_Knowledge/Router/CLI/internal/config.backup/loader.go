package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

type Config struct {
	LogLevel  string            `mapstructure:"log_level"`
	PluginDir string            `mapstructure:"plugin_dir"`
	Plugins   map[string]string `mapstructure:"plugins"`
	GRPC      GRPCConfig        `mapstructure:"grpc"`
	Security  SecurityConfig    `mapstructure:"security"`
}

type GRPCConfig struct {
	Port           int    `mapstructure:"port"`
	Host           string `mapstructure:"host"`
	EnableTLS      bool   `mapstructure:"enable_tls"`
	EnableMTLS     bool   `mapstructure:"enable_mtls"`
	CertFile       string `mapstructure:"cert_file"`
	KeyFile        string `mapstructure:"key_file"`
	CAFile         string `mapstructure:"ca_file"`
	MaxConnections int    `mapstructure:"max_connections"`
	EnablePooling  bool   `mapstructure:"enable_pooling"`
}

type SecurityConfig struct {
	EnableSandbox  bool           `mapstructure:"enable_sandbox"`
	SandboxProfile string         `mapstructure:"sandbox_profile"`
	ResourceLimits ResourceLimits `mapstructure:"resource_limits"`
	AllowedPlugins []string       `mapstructure:"allowed_plugins"`
}

type ResourceLimits struct {
	MaxMemoryMB int `mapstructure:"max_memory_mb"`
	MaxCPU      int `mapstructure:"max_cpu"`
	MaxTimeout  int `mapstructure:"max_timeout"`
}

var cfg *Config

func Load(configPath string) (*Config, error) {
	v := viper.New()

	// Set defaults
	setDefaults(v)

	// Config file
	if configPath != "" {
		v.SetConfigFile(configPath)
	} else {
		// Search config in home directory
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("error getting home directory: %w", err)
		}
		v.AddConfigPath(filepath.Join(home, ".ti-cli"))
		v.AddConfigPath(".")
		v.SetConfigName("config")
		v.SetConfigType("yaml")
	}

	// Environment variables
	v.SetEnvPrefix("TI_CLI")
	v.AutomaticEnv()

	// Read config file
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
		// Config file not found, use defaults
	}

	// Unmarshal config
	var config Config
	if err := v.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}

	cfg = &config
	return cfg, nil
}

func setDefaults(v *viper.Viper) {
	v.SetDefault("log_level", "info")
	v.SetDefault("plugin_dir", "./plugins")
	v.SetDefault("grpc.port", 50051)
	v.SetDefault("grpc.host", "localhost")
	v.SetDefault("grpc.enable_tls", false)
	v.SetDefault("grpc.enable_mtls", false)
	v.SetDefault("grpc.max_connections", 10)
	v.SetDefault("grpc.enable_pooling", true)
	v.SetDefault("security.enable_sandbox", true)
	v.SetDefault("security.sandbox_profile", "default")
	v.SetDefault("security.resource_limits.max_memory_mb", 512)
	v.SetDefault("security.resource_limits.max_cpu", 2)
	v.SetDefault("security.resource_limits.max_timeout", 300)
}

func Get() *Config {
	if cfg == nil {
		// Load with default path
		cfg, _ = Load("")
	}
	return cfg
}
