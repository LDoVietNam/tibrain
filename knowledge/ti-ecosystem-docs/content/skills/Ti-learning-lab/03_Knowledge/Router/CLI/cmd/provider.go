package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/ti/cli/internal/config"
	"github.com/ti/cli/internal/providers"
)

var providerCmd = &cobra.Command{
	Use:   "provider",
	Short: "Manage and query AI providers",
	Long:  `List, inspect, and manage AI backend providers registered in Ti CLI.`,
}

var providerListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all registered providers",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Load configuration
		cfg, err := config.LoadAuto()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		// Build registry from config
		reg, issues := providers.BuildRegistryFromConfig(cfg)

		// List providers
		snapshots := reg.Snapshots()

		if len(snapshots) == 0 {
			fmt.Println("No providers registered.")
			fmt.Println()
			fmt.Println("To register a provider:")
			fmt.Println("  1. Set up authentication (cookie, API key, or OAuth)")
			fmt.Println("  2. Configure in ti.json or environment variables")
			fmt.Println("  3. Run 'ti doctor' to verify setup")

			if len(issues) > 0 {
				fmt.Println()
				fmt.Println("Bootstrap issues:")
				for _, issue := range issues {
					fmt.Printf("  - %s: %s\n", issue.Provider, issue.Message)
				}
			}
			return nil
		}

		fmt.Printf("Registered Providers (%d):\n", len(snapshots))
		fmt.Println(strings.Repeat("-", 60))

		for _, snap := range snapshots {
			status := "❌"
			if snap.Healthy {
				status = "✅"
			}

			fmt.Printf("%s %-20s | Model: %-25s | Models: %d\n",
				status, snap.Name, snap.DefaultModel, len(snap.Models))

			if len(snap.Models) > 0 {
				fmt.Printf("   Available models: %s\n", strings.Join(snap.Models, ", "))
			}
		}

		if len(issues) > 0 {
			fmt.Println()
			fmt.Println("Bootstrap warnings:")
			for _, issue := range issues {
				fmt.Printf("  ⚠️  %s: %s\n", issue.Provider, issue.Message)
			}
		}

		return nil
	},
}

var providerStatusCmd = &cobra.Command{
	Use:   "status [name]",
	Short: "Check provider health status",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		providerName := args[0]

		// Load configuration
		cfg, err := config.LoadAuto()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		// Build registry from config
		reg, _ := providers.BuildRegistryFromConfig(cfg)

		// Get provider
		p, ok := reg.Get(providerName)
		if !ok {
			return fmt.Errorf("provider '%s' not found. Run 'ti provider list' to see available providers", providerName)
		}

		// Display status
		fmt.Printf("Provider: %s\n", p.Name())
		fmt.Printf("Default Model: %s\n", p.DefaultModel())
		fmt.Printf("Status: ")
		if p.IsHealthy() {
			fmt.Println("✅ Healthy")
		} else {
			fmt.Println("❌ Unhealthy")
		}
		fmt.Printf("Available Models: %s\n", strings.Join(p.Models(), ", "))

		return nil
	},
}

func init() {
	rootCmd.AddCommand(providerCmd)
	providerCmd.AddCommand(providerListCmd)
	providerCmd.AddCommand(providerStatusCmd)
}
