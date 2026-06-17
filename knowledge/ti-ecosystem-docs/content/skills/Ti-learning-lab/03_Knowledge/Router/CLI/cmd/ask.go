package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/ti/cli/internal/config"
	"github.com/ti/cli/internal/providers"
	"github.com/ti/cli/internal/ui"
)

var askCmd = &cobra.Command{
	Use:   "ask",
	Short: "Start interactive AI chat TUI",
	Long:  `Launch an interactive terminal UI to chat with AI providers.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadAuto()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		reg, issues := providers.BuildRegistryFromConfig(cfg)
		if len(issues) > 0 {
			for _, issue := range issues {
				fmt.Printf("Warning: %s: %s\n", issue.Provider, issue.Message)
			}
		}

		if len(reg.List()) == 0 {
			fmt.Println("No providers available.")
			fmt.Println()
			fmt.Println("Set one of these environment variables:")
			fmt.Println("  TI_ROUTER_URL        - Ti Router URL (default: http://localhost:1806)")
			fmt.Println("  OPENAI_API_KEY       - OpenAI API key")
			fmt.Println("  ANTHROPIC_API_KEY    - Anthropic API key")
			fmt.Println("  OPENROUTER_API_KEY   - OpenRouter API key")
			fmt.Println("Or configure auth files in ti.json")
			return nil
		}

		return ui.StartChatTUI(reg)
	},
}

func init() {
	rootCmd.AddCommand(askCmd)
}
