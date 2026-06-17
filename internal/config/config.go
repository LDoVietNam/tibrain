// Package config provides configuration management for TiBrain.
package config

import (
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config holds all TiBrain configuration.
type Config struct {
	Port               int
	DataDir            string
	RouterBrainURL     string
	IndexKnowledge     bool
	IndexOnly          bool
	KnowledgeSources   []string
	CLIRegistry        bool
	HandoffTrack       bool
	SkillSync          bool
	MCPHubEnabled      bool
	MCPHubURL          string
	MCPHubAutoSync     bool
	MCPHubSyncInterval string
	MCPHubRegistryFile string
}

// DefaultConfig returns the default configuration.
func DefaultConfig() *Config {
	return &Config{
		Port:               1810,
		DataDir:            filepath.Join(getTiBrainDir(), "memory"),
		RouterBrainURL:     "http://localhost:1816",
		KnowledgeSources:   []string{},
		CLIRegistry:        true,
		HandoffTrack:       true,
		SkillSync:          false,
		MCPHubEnabled:      false,
		MCPHubURL:          "http://localhost:3000",
		MCPHubAutoSync:     true,
		MCPHubSyncInterval: "5m",
		MCPHubRegistryFile: filepath.Join(getTiBrainDir(), "mcp_registry.json"),
	}
}

// LoadConfig loads configuration from file and environment.
func LoadConfig() *Config {
	config := DefaultConfig()

	configFile := filepath.Join(getTiBrainDir(), "config.yaml")
	if data, err := os.ReadFile(configFile); err == nil {
		var yamlConfig struct {
			TiBrain struct {
				Port           int    `yaml:"port"`
				DataDir        string `yaml:"data_dir"`
				RouterBrainURL string `yaml:"router_brain_url"`
			} `yaml:"tibrain"`
			RouterBrain struct {
				URL string `yaml:"url"`
			} `yaml:"router_brain"`
			MCPHub struct {
				Enabled      bool   `yaml:"enabled"`
				HubURL       string `yaml:"hub_url"`
				AutoSync     bool   `yaml:"auto_sync"`
				SyncInterval string `yaml:"sync_interval"`
				RegistryFile string `yaml:"registry_file"`
			} `yaml:"mcp_hub"`
			CLIRegistry struct {
				Enabled bool `yaml:"enabled"`
			} `yaml:"cli_registry"`
			HandoffTracking struct {
				Enabled bool `yaml:"enabled"`
			} `yaml:"handoff_tracking"`
			SkillSync struct {
				Enabled bool `yaml:"enabled"`
			} `yaml:"skill_sync"`
		}

		if err := yaml.Unmarshal(data, &yamlConfig); err == nil {
			if yamlConfig.TiBrain.Port > 0 {
				config.Port = yamlConfig.TiBrain.Port
			}
			if yamlConfig.TiBrain.DataDir != "" {
				config.DataDir = yamlConfig.TiBrain.DataDir
			}
			if yamlConfig.TiBrain.RouterBrainURL != "" {
				config.RouterBrainURL = yamlConfig.TiBrain.RouterBrainURL
			}
			if yamlConfig.RouterBrain.URL != "" {
				config.RouterBrainURL = yamlConfig.RouterBrain.URL
			}
			config.MCPHubEnabled = yamlConfig.MCPHub.Enabled
			if yamlConfig.MCPHub.HubURL != "" {
				config.MCPHubURL = yamlConfig.MCPHub.HubURL
			}
			config.MCPHubAutoSync = yamlConfig.MCPHub.AutoSync
			if yamlConfig.MCPHub.SyncInterval != "" {
				config.MCPHubSyncInterval = yamlConfig.MCPHub.SyncInterval
			}
			if yamlConfig.MCPHub.RegistryFile != "" {
				config.MCPHubRegistryFile = yamlConfig.MCPHub.RegistryFile
			}
			config.CLIRegistry = yamlConfig.CLIRegistry.Enabled
			config.HandoffTrack = yamlConfig.HandoffTracking.Enabled
			config.SkillSync = yamlConfig.SkillSync.Enabled
		}
	}

	if envURL := os.Getenv("TIBRAIN_ROUTER_BRAIN_URL"); envURL != "" {
		config.RouterBrainURL = envURL
	}
	if os.Getenv("TIBRAIN_INDEX_KNOWLEDGE") == "1" || strings.EqualFold(os.Getenv("TIBRAIN_INDEX_KNOWLEDGE"), "true") {
		config.IndexKnowledge = true
	}

	return config
}

func getTiBrainDir() string {
	if wd, err := os.Getwd(); err == nil {
		return wd
	}
	if exe, err := os.Executable(); err == nil {
		return filepath.Dir(exe)
	}
	return "Z:\\Ti\\router\\tibrain"
}
