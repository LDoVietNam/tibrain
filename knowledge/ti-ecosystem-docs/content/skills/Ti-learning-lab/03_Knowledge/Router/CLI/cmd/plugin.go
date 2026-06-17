package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/ti/cli/internal/core"
	"github.com/ti/cli/internal/plugins"
)

var pluginCmd = &cobra.Command{
	Use:   "plugin",
	Short: "Manage plugins",
	Long:  `Manage and interact with the plugin system.`,
}

var pluginListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all available plugins",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Initialize registry
		registry := core.NewPluginRegistry()

		// List all plugins
		metadataList := registry.ListPlugins()

		fmt.Println("Available Plugins:")
		for _, metadata := range metadataList {
			fmt.Printf("  - %s (v%s, %s): %s\n", metadata.Name, metadata.Version, metadata.Type, metadata.Description)
			fmt.Printf("    Capabilities: %v\n", metadata.Capabilities)
		}

		return nil
	},
}

var pluginLoadCmd = &cobra.Command{
	Use:   "load [name]",
	Short: "Load a plugin",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		pluginName := args[0]

		// Initialize platform
		loader := plugins.NewDefaultLoader()
		registry := core.NewPluginRegistry()
		platform := core.NewDefaultPlatform(loader, registry)

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// Load plugin
		if err := platform.LoadPlugin(ctx, pluginName); err != nil {
			return fmt.Errorf("failed to load plugin %s: %w", pluginName, err)
		}

		fmt.Printf("Plugin %s loaded successfully\n", pluginName)

		return nil
	},
}

var pluginUnloadCmd = &cobra.Command{
	Use:   "unload [name]",
	Short: "Unload a plugin",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		pluginName := args[0]

		// Initialize platform
		loader := plugins.NewDefaultLoader()
		registry := core.NewPluginRegistry()
		platform := core.NewDefaultPlatform(loader, registry)

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// Unload plugin
		if err := platform.UnloadPlugin(ctx, pluginName); err != nil {
			return fmt.Errorf("failed to unload plugin %s: %w", pluginName, err)
		}

		fmt.Printf("Plugin %s unloaded successfully\n", pluginName)

		return nil
	},
}

var pluginStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show plugin status",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Initialize platform
		loader := plugins.NewDefaultLoader()
		registry := core.NewPluginRegistry()
		platform := core.NewDefaultPlatform(loader, registry)

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		// Load all plugins
		metadataList := registry.ListPlugins()
		for _, metadata := range metadataList {
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

		return nil
	},
}

var pluginExecuteCmd = &cobra.Command{
	Use:   "execute [name] [task]",
	Short: "Execute a task on a plugin",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		pluginName := args[0]
		task := args[1]

		// Initialize platform
		loader := plugins.NewDefaultLoader()
		registry := core.NewPluginRegistry()
		platform := core.NewDefaultPlatform(loader, registry)

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()

		// Load plugin
		if err := platform.LoadPlugin(ctx, pluginName); err != nil {
			return fmt.Errorf("failed to load plugin %s: %w", pluginName, err)
		}

		// Execute task
		input := map[string]interface{}{}

		result, err := platform.ExecutePlugin(ctx, pluginName, task, input)
		if err != nil {
			return fmt.Errorf("failed to execute task: %w", err)
		}

		// Print result
		fmt.Printf("Result: %+v\n", result)

		return nil
	},
}

func init() {
	rootCmd.AddCommand(pluginCmd)
	pluginCmd.AddCommand(pluginListCmd)
	pluginCmd.AddCommand(pluginLoadCmd)
	pluginCmd.AddCommand(pluginUnloadCmd)
	pluginCmd.AddCommand(pluginStatusCmd)
	pluginCmd.AddCommand(pluginExecuteCmd)
}
