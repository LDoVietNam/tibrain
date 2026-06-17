package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/ti/cli/internal/config"
	"github.com/ti/cli/internal/learn"
	"github.com/ti/cli/internal/plugins"
)

var cfgFile string
var logLevel string

var rootCmd = &cobra.Command{
	Use:   "ti-cli",
	Short: "Ti CLI - Microkernel + Plugin Architecture",
	Long: `Ti CLI with Microkernel + Plugin Architecture for integrating
Devin CLI capabilities with enterprise-grade quality.

This CLI supports:
- Plugin system with gRPC communication
- Native Go plugins for critical path features
- Devin plugin for complex automation
- Extensible architecture for future agent plugins`,
	Version: version,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Load configuration
		if err := loadConfig(); err != nil {
			return fmt.Errorf("error loading configuration: %w", err)
		}
		return nil
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	// Global flags
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file (JSON; default discovery uses ti.json or ~/.config/ti/config.json)")
	rootCmd.PersistentFlags().StringVarP(&logLevel, "log-level", "l", "", "log level (debug, info, warn, error) (overrides config)")
	learn.RegisterAll(rootCmd)
}

func loadConfig() error {
	// Override log level from flag if provided
	if logLevel != "" {
		// Temporarily set environment variable for viper
		os.Setenv("TI_LOG_LEVEL", logLevel)
	}

	var cfg *config.Config
	var err error
	if cfgFile != "" {
		cfg, err = config.Load(cfgFile)
	} else {
		cfg, err = config.LoadAuto()
	}
	if err != nil {
		return err
	}

	// Validate configuration. Warnings should not block normal CLI usage.
	for _, issue := range cfg.Validate() {
		if issue.Level == "error" {
			return fmt.Errorf("configuration validation failed: %s: %s", issue.Field, issue.Message)
		}
	}

	// Initialize logger
	if err := plugins.InitLogger(cfg.LogLevel); err != nil {
		return fmt.Errorf("error initializing logger: %w", err)
	}

	return nil
}
