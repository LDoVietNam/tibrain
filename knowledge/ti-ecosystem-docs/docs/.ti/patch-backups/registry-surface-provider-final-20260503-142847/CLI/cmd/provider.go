package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/ti/cli/internal/config"
	"github.com/ti/cli/internal/providers"
	"github.com/ti/cli/internal/ui"
)

var providerCmd = &cobra.Command{
	Use:   "provider",
	Short: "Manage and query AI providers",
	Long:  `List, inspect, and manage AI backend providers registered in Ti CLI.`,
}

var providerAddBaseURL string
var providerAddModel string
var providerAddModels string
var providerAddAPIKey string
var providerAddAPIKeyEnv string
var providerAddHeaders []string
var providerAddNoKey bool
var providerAddInteractive bool
var providerListInteractive bool

var providerListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all registered providers",
	RunE: func(cmd *cobra.Command, args []string) error {
		if providerListInteractive {
			fmt.Println("🖱️  Provider List Interactive Mode")
			fmt.Println("   (Full TUI coming soon - using current list view)")
			fmt.Println()
		}

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
			fmt.Println("  3. For cookie provider specs: export TI_COOKIE_PROVIDER='profile|base_url|cookie'")
			fmt.Println("  4. Run 'ti provider cookie-profiles' to view presets")

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
		fmt.Println(strings.Repeat("-", 72))

		for _, snap := range snapshots {
			status := "❌"
			if snap.Healthy {
				status = "✅"
			}

			fmt.Printf("%s %-24s | Model: %-24s | Models: %d\n",
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

var providerAddCmd = &cobra.Command{
	Use:   "add [name]",
	Short: "Add or update a custom OpenAI-compatible provider",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if providerAddInteractive {
			return runInteractiveProviderAdd(args[0])
		}

		headers, err := parseProviderHeaders(providerAddHeaders)
		if err != nil {
			return err
		}
		models := splitCSV(providerAddModels)
		if providerAddModel != "" {
			models = append(models, providerAddModel)
		}
		secret := strings.TrimSpace(providerAddAPIKey)
		if strings.TrimSpace(providerAddAPIKeyEnv) != "" {
			secret = "env:" + strings.TrimSpace(providerAddAPIKeyEnv)
		}
		path, err := providers.SaveCustomCompatibleProvider(providers.SavedCompatibleProvider{
			Name:      args[0],
			BaseURL:   providerAddBaseURL,
			Model:     providerAddModel,
			Models:    models,
			APIKey:    secret,
			Headers:   headers,
			NoKey:     providerAddNoKey,
		})
		if err != nil {
			return err
		}
		fmt.Printf("Custom provider saved: %s\n", args[0])
		fmt.Printf("Store: %s\n", path)
		if providerAddAPIKey != "" && providerAddAPIKeyEnv == "" {
			fmt.Println("Warning: API key was saved in the provider store. Prefer --api-key-env for long-term use.")
		}
		fmt.Println("Run `ti provider list` to verify, or `ti ask` to use it in the chat TUI.")
		return nil
	},
}

var providerCustomListCmd = &cobra.Command{
	Use:   "custom-list",
	Short: "List saved custom providers",
	RunE: func(cmd *cobra.Command, args []string) error {
		items, path, err := providers.ListSavedCustomProviders()
		if err != nil {
			return err
		}
		fmt.Printf("Custom provider store: %s\n", path)
		if len(items) == 0 {
			fmt.Println("No custom providers saved.")
			return nil
		}
		fmt.Println(strings.Repeat("-", 88))
		for _, item := range items {
			keyMode := "missing-key"
			if item.NoKey {
				keyMode = "no-key/local"
			} else if item.APIKeyEnv != "" {
				keyMode = "env:" + item.APIKeyEnv
			} else if item.APIKey != "" {
				keyMode = "saved-key"
			}
			fmt.Printf("%-18s model=%-24s auth=%-14s base=%s\n", item.Name, item.Model, keyMode, item.BaseURL)
		}
		return nil
	},
}

var providerRemoveCmd = &cobra.Command{
	Use:   "remove [name]",
	Short: "Remove a saved custom provider",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		path, deleted, err := providers.DeleteCustomCompatibleProvider(args[0])
		if err != nil {
			return err
		}
		if !deleted {
			return fmt.Errorf("custom provider not found: %s", args[0])
		}
		fmt.Printf("Removed custom provider %s from %s\n", args[0], path)
		return nil
	},
}

func parseProviderHeaders(values []string) (map[string]string, error) {
	if len(values) == 0 {
		return nil, nil
	}
	out := map[string]string{}
	for _, raw := range values {
		idx := strings.Index(raw, "=")
		if idx <= 0 {
			return nil, fmt.Errorf("invalid header %q; expected Name=Value", raw)
		}
		out[strings.TrimSpace(raw[:idx])] = strings.TrimSpace(raw[idx+1:])
	}
	return out, nil
}

func splitCSV(raw string) []string {
	var out []string
	for _, part := range strings.Split(raw, ",") {
		part = strings.TrimSpace(part)
		if part != "" {
			out = append(out, part)
		}
	}
	return out
}

func runInteractiveProviderAdd(name string) error {
	fmt.Println("🖱️  Starting Provider Wizard with Mouse Support...")
	fmt.Println("   Use mouse to click options or keyboard to navigate")
	fmt.Println()

	provider, err := ui.StartProviderWizard(name)
	if err != nil {
		if err.Error() == "cancelled by user" {
			fmt.Println("❌ Cancelled by user")
			return nil
		}
		return fmt.Errorf("wizard error: %w", err)
	}

	path, err := providers.SaveCustomCompatibleProvider(provider)
	if err != nil {
		return fmt.Errorf("failed to save provider: %w", err)
	}

	fmt.Printf("\n✅ Provider saved successfully!\n")
	fmt.Printf("Store: %s\n", path)
	if provider.APIKey != "" && !strings.HasPrefix(provider.APIKey, "env:") {
		fmt.Println("⚠️  Warning: API key was saved in the provider store. Consider using environment variables for long-term use.")
	}
	fmt.Println("\nNext steps:")
	fmt.Println("  - Run 'ti provider list' to verify")
	fmt.Println("  - Run 'ti ask' to use it in the chat TUI")

	return nil
}

var providerCookieProfilesCmd = &cobra.Command{
	Use:   "cookie-profiles",
	Short: "List cookie provider profiles",
	Run: func(cmd *cobra.Command, args []string) {
		profiles := providers.KnownCookieProfiles()
		fmt.Println("Cookie provider profiles:")
		fmt.Println(strings.Repeat("-", 72))
		for _, name := range providers.CookieProviderProfileNames() {
			p := profiles[name]
			fmt.Printf("%-14s base=%-32s style=%s\n", name, p.BaseURL, p.PayloadStyle)
		}
		fmt.Println()
		fmt.Println("Spec format:")
		fmt.Println("  profile|base_url|cookie")
		fmt.Println("  name|base_url|path=/v1/chat/completions|model=auto|style=openai|cookie=a=b; c=d")
		fmt.Println()
		fmt.Println("Example:")
		fmt.Println("  export TI_COOKIE_PROVIDER='sharedchat|https://chat.sharedchat.fun|gfsessionid=xxx; _dd_s=yyy'")
	},
}

var providerCookieSpecCmd = &cobra.Command{
	Use:   "cookie-spec [profile] [base-url] [cookie]",
	Short: "Print a TI_COOKIE_PROVIDER spec with cookie at the end",
	Args:  cobra.RangeArgs(2, 3),
	Run: func(cmd *cobra.Command, args []string) {
		profile := args[0]
		baseURL := args[1]
		cookie := "PASTE_COOKIE_HERE"
		if len(args) > 2 {
			cookie = args[2]
		}
		fmt.Printf("export TI_COOKIE_PROVIDER='%s|%s|%s'\n", profile, baseURL, cookie)
	},
}

var providerCookieTestCmd = &cobra.Command{
	Use:   "cookie-test [spec] [prompt]",
	Short: "Test a cookie provider spec with a prompt",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		p, err := providers.NewCookieProviderFromSpec(args[0])
		if err != nil {
			return err
		}
		ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
		defer cancel()
		resp, err := p.Chat(ctx, providers.ChatRequest{
			Messages: []providers.Message{{Role: "user", Content: args[1]}},
		})
		if err != nil {
			return err
		}
		fmt.Println(resp.Content)
		return nil
	},
}

var providerSyncFromRouterCmd = &cobra.Command{
	Use:   "sync-from-router",
	Short: "Sync providers from Ti Router",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get router URL from environment or default
		routerURL := os.Getenv("TI_ROUTER_AGENT_URL")
		if routerURL == "" {
			routerURL = "http://localhost:1807"
		}

		fmt.Printf("Syncing providers from Router: %s\n", routerURL)

		// Fetch providers from router
		client := &http.Client{Timeout: 30 * time.Second}
		resp, err := client.Get(routerURL + "/api/providers/export")
		if err != nil {
			return fmt.Errorf("failed to fetch providers from router: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("router returned status %d", resp.StatusCode)
		}

		var result struct {
			Providers  map[string]interface{} `json:"providers"`
			Count      int                    `json:"count"`
			ExportedAt string                 `json:"exported_at"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}

		fmt.Printf("Fetched %d providers from Router\n", result.Count)
		fmt.Printf("Exported at: %s\n", result.ExportedAt)

		// TODO: Import providers into CLI registry
		// This would require converting router provider configs to CLI provider configs

		fmt.Println("\nProviders available from Router:")
		for name := range result.Providers {
			fmt.Printf("  - %s\n", name)
		}

		fmt.Println("\nNote: Full provider sync to CLI registry not yet implemented.")
		fmt.Println("Router providers can be used via the 'router' provider.")

		return nil
	},
}

var providerSyncToRouterCmd = &cobra.Command{
	Use:   "sync-to-router",
	Short: "Sync local providers to Ti Router",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get router URL from environment or default
		routerURL := os.Getenv("TI_ROUTER_AGENT_URL")
		if routerURL == "" {
			routerURL = "http://localhost:1807"
		}

		fmt.Printf("Syncing providers to Router: %s\n", routerURL)

		// Load CLI providers
		cfg, err := config.LoadAuto()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		reg, _ := providers.BuildRegistryFromConfig(cfg)
		snapshots := reg.Snapshots()

		if len(snapshots) == 0 {
			fmt.Println("No local providers to sync.")
			return nil
		}

		fmt.Printf("Found %d local providers\n", len(snapshots))

		// TODO: Convert CLI providers to Router format and send to /api/providers/import
		// This would require mapping CLI provider configs to Router provider configs

		fmt.Println("\nLocal providers:")
		for _, snap := range snapshots {
			fmt.Printf("  - %s (model: %s)\n", snap.Name, snap.DefaultModel)
		}

		fmt.Println("\nNote: Provider sync to Router not yet implemented.")

		return nil
	},
}

var providerCompatProfilesCmd = &cobra.Command{
	Use:   "compat-profiles",
	Short: "List OpenAI-compatible provider profiles",
	Run: func(cmd *cobra.Command, args []string) {
		profiles := providers.KnownCompatibleProfiles()
		fmt.Println("OpenAI-compatible provider profiles:")
		fmt.Println(strings.Repeat("-", 88))
		for _, name := range providers.CompatibleProfileNames() {
			p := profiles[name]
			keyMode := "api-key"
			if p.HealthyNoKey {
				keyMode = "no-key/local"
			}
			fmt.Printf("%-14s provider=%-18s base=%-38s auth=%s\n", name, p.Name, p.BaseURL, keyMode)
		}
		fmt.Println()
		fmt.Println("Spec format:")
		fmt.Println("  profile")
		fmt.Println("  profile|base_url|model|api_key")
		fmt.Println("  name|base_url|model=auto|models=a,b|key=sk-...|header:Name=Value")
		fmt.Println()
		fmt.Println("Env examples:")
		fmt.Println("  export TI_COMPAT_PROVIDER='nvidia_nim'")
		fmt.Println("  export TI_COMPAT_PROVIDERS='lmstudio;;ollama;;custom|http://localhost:8088/v1|model=auto|nokey=true'")
	},
}

var providerCompatSpecCmd = &cobra.Command{
	Use:   "compat-spec [profile] [base-url] [model] [api-key]",
	Short: "Print a TI_COMPAT_PROVIDER spec",
	Args:  cobra.RangeArgs(1, 4),
	Run: func(cmd *cobra.Command, args []string) {
		profile := args[0]
		baseURL := ""
		model := "auto"
		apiKey := "PASTE_API_KEY_HERE"
		if len(args) > 1 {
			baseURL = args[1]
		}
		if len(args) > 2 {
			model = args[2]
		}
		if len(args) > 3 {
			apiKey = args[3]
		}
		if baseURL == "" {
			fmt.Printf("export TI_COMPAT_PROVIDER='%s'\n", profile)
			return
		}
		fmt.Printf("export TI_COMPAT_PROVIDER='%s|%s|model=%s|key=%s'\n", profile, baseURL, model, apiKey)
	},
}

func init() {
	// Add flags to providerAddCmd
	providerAddCmd.Flags().StringVar(&providerAddBaseURL, "base-url", "", "Base URL of the provider (required)")
	providerAddCmd.Flags().StringVar(&providerAddModel, "model", "", "Default model")
	providerAddCmd.Flags().StringVar(&providerAddModels, "models", "", "Comma-separated list of available models")
	providerAddCmd.Flags().StringVar(&providerAddAPIKey, "api-key", "", "API key (not recommended for long-term use)")
	providerAddCmd.Flags().StringVar(&providerAddAPIKeyEnv, "api-key-env", "", "Environment variable name for API key (recommended)")
	providerAddCmd.Flags().StringArrayVar(&providerAddHeaders, "header", []string{}, "Custom headers in format Name=Value")
	providerAddCmd.Flags().BoolVar(&providerAddNoKey, "no-key", false, "Provider does not require an API key (e.g., local LM Studio)")
	providerAddCmd.Flags().BoolVar(&providerAddInteractive, "interactive", false, "Interactive wizard mode for guided provider setup")
	
	// Add flags to providerListCmd
	providerListCmd.Flags().BoolVar(&providerListInteractive, "interactive", false, "Interactive list mode with mouse support")

	// Register commands
	rootCmd.AddCommand(providerCmd)
	providerCmd.AddCommand(providerListCmd)
	providerCmd.AddCommand(providerStatusCmd)
	providerCmd.AddCommand(providerAddCmd)
	providerCmd.AddCommand(providerCustomListCmd)
	providerCmd.AddCommand(providerRemoveCmd)
	providerCmd.AddCommand(providerCookieProfilesCmd)
	providerCmd.AddCommand(providerCompatProfilesCmd)
	providerCmd.AddCommand(providerCompatSpecCmd)
	providerCmd.AddCommand(providerCookieSpecCmd)
	providerCmd.AddCommand(providerCookieTestCmd)
	providerCmd.AddCommand(providerSyncFromRouterCmd)
	providerCmd.AddCommand(providerSyncToRouterCmd)
}
