package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/ti/cli/internal/core"
	"github.com/ti/cli/internal/plugins"
)

var (
	agentPlugin string
	agentTask   string
)

var agentCmd = &cobra.Command{
	Use:   "agent",
	Short: "Manage AI agents",
	Long:  `Manage and interact with AI agent plugins.`,
}

var agentListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all available agents",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Initialize registry
		registry := core.NewPluginRegistry()

		// List all plugins
		metadataList := registry.ListPlugins()

		fmt.Println("Available Agents:")
		for _, metadata := range metadataList {
			if metadata.Type == "agent" {
				fmt.Printf("  - %s (v%s): %s\n", metadata.Name, metadata.Version, metadata.Description)
			}
		}

		return nil
	},
}

var agentExecuteCmd = &cobra.Command{
	Use:   "execute [task]",
	Short: "Execute a task on an agent",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		task := args[0]

		// Initialize platform
		loader := plugins.NewDefaultLoader()
		registry := core.NewPluginRegistry()
		platform := core.NewDefaultPlatform(loader, registry)

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()

		// Load plugin
		pluginName := "devin"
		if agentPlugin != "" {
			pluginName = agentPlugin
		}

		if err := platform.LoadPlugin(ctx, pluginName); err != nil {
			return fmt.Errorf("failed to load plugin %s: %w", pluginName, err)
		}

		// Execute task
		input := map[string]interface{}{
			"task": task,
		}

		result, err := platform.ExecutePlugin(ctx, pluginName, "execute", input)
		if err != nil {
			return fmt.Errorf("failed to execute task: %w", err)
		}

		// Print result
		fmt.Printf("Result: %+v\n", result)

		return nil
	},
}

var agentStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show agent status",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Initialize platform
		loader := plugins.NewDefaultLoader()
		registry := core.NewPluginRegistry()
		platform := core.NewDefaultPlatform(loader, registry)

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		// Load all agent plugins
		metadataList := registry.ListPlugins()
		for _, metadata := range metadataList {
			if metadata.Type == "agent" {
				if err := platform.LoadPlugin(ctx, metadata.Name); err != nil {
					fmt.Printf("%s: Failed to load - %v\n", metadata.Name, err)
					continue
				}

				plugin, err := platform.GetPlugin(metadata.Name)
				if err != nil {
					fmt.Printf("%s: Failed to get - %v\n", metadata.Name, err)
					continue
				}

				// Health check
				if err := plugin.HealthCheck(ctx); err != nil {
					fmt.Printf("%s: Unhealthy - %v\n", metadata.Name, err)
				} else {
					fmt.Printf("%s: Healthy (State: %s)\n", metadata.Name, plugin.State())
				}
			}
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(agentCmd)
	agentCmd.AddCommand(agentListCmd)
	agentCmd.AddCommand(agentExecuteCmd)
	agentCmd.AddCommand(agentStatusCmd)

	agentExecuteCmd.Flags().StringVarP(&agentPlugin, "plugin", "p", "devin", "plugin to use")
}
