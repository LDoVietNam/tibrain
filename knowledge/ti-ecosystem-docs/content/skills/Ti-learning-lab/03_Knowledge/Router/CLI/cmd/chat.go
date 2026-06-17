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
	chatAgent   string
	chatSession string
	chatStream  bool
)

var chatCmd = &cobra.Command{
	Use:   "chat [prompt]",
	Short: "Chat with an AI agent",
	Long:  `Chat with an AI agent using the plugin system. Routes to the appropriate agent based on capabilities.`,
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		prompt := args[0]

		// Initialize platform
		loader := plugins.NewDefaultLoader()
		registry := core.NewPluginRegistry()
		platform := core.NewDefaultPlatform(loader, registry)

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()

		// Load plugin
		pluginName := "devin"
		if chatAgent != "" {
			pluginName = chatAgent
		}

		if err := platform.LoadPlugin(ctx, pluginName); err != nil {
			return fmt.Errorf("failed to load plugin %s: %w", pluginName, err)
		}

		// Execute chat
		input := map[string]interface{}{
			"prompt":  prompt,
			"session": chatSession,
			"stream":  chatStream,
		}

		result, err := platform.ExecutePlugin(ctx, pluginName, "execute", input)
		if err != nil {
			return fmt.Errorf("failed to execute chat: %w", err)
		}

		// Print result
		if result["success"] == true {
			fmt.Println(result["result"])
		} else {
			return fmt.Errorf("chat failed")
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(chatCmd)

	chatCmd.Flags().StringVarP(&chatAgent, "agent", "a", "devin", "agent to use (devin, native)")
	chatCmd.Flags().StringVarP(&chatSession, "session", "s", "", "session ID for conversation context")
	chatCmd.Flags().BoolVarP(&chatStream, "stream", "", false, "stream responses in real-time")
}
