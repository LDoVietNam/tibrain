package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/ti/cli/internal/authtoken"
)

var (
	authTokenProvider string
	authTokenFile     string
	authTokenStdin    bool
	authTokenRoot     string
	authTokenJSON     bool
)

var authTokenCmd = &cobra.Command{
	Use:   "token",
	Short: "Manage provider auth tokens without printing secret values",
	Long: `Manage provider auth tokens in a local 0600 store.

Ti does not scrape protected application credentials. For Windsurf and similar tools,
open the provider's official token page while logged in, copy your own token, then
import it via --file or --stdin. Never commit generated token files.`,
}

var authTokenImportCmd = &cobra.Command{
	Use:   "import [profile]",
	Short: "Import an auth token from a local file or stdin",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		profile := "default"
		if len(args) == 1 && strings.TrimSpace(args[0]) != "" {
			profile = strings.TrimSpace(args[0])
		}
		token, source, err := readTokenInput()
		if err != nil {
			return err
		}
		store := authtoken.NewStore(authTokenRoot)
		p, err := store.Import(profile, authTokenProvider, token, source)
		if err != nil {
			return err
		}
		if authTokenJSON {
			b, _ := json.MarshalIndent(p, "", "  ")
			fmt.Println(string(b))
			return nil
		}
		fmt.Printf("Imported token profile %q (%s) into %s\n", p.Name, p.Provider, store.Root)
		fmt.Printf("Preview: %s\n", p.Preview)
		return nil
	},
}

var authTokenListCmd = &cobra.Command{
	Use:   "list",
	Short: "List imported auth token profiles",
	RunE: func(cmd *cobra.Command, args []string) error {
		store := authtoken.NewStore(authTokenRoot)
		items, err := store.List()
		if err != nil {
			return err
		}
		if authTokenJSON {
			b, _ := json.MarshalIndent(items, "", "  ")
			fmt.Println(string(b))
			return nil
		}
		if len(items) == 0 {
			fmt.Println("No token profiles imported.")
			return nil
		}
		for _, p := range items {
			fmt.Printf("- %s\n  provider: %s\n  preview: %s\n  updated: %s\n", p.Name, p.Provider, p.Preview, p.UpdatedAt.Format("2006-01-02 15:04:05"))
		}
		return nil
	},
}

var authTokenShowCmd = &cobra.Command{
	Use:   "show <profile>",
	Short: "Show token metadata and redacted preview",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		store := authtoken.NewStore(authTokenRoot)
		p, ok, err := store.GetProfile(args[0])
		if err != nil {
			return err
		}
		if !ok {
			return fmt.Errorf("token profile %q not found", args[0])
		}
		b, _ := json.MarshalIndent(p, "", "  ")
		fmt.Println(string(b))
		return nil
	},
}

var authTokenRemoveCmd = &cobra.Command{
	Use:   "remove <profile>",
	Short: "Remove an imported auth token profile",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		store := authtoken.NewStore(authTokenRoot)
		if err := store.Remove(args[0]); err != nil {
			return err
		}
		fmt.Printf("Removed token profile %q from %s\n", args[0], store.Root)
		return nil
	},
}

var authTokenPathCmd = &cobra.Command{
	Use:   "path",
	Short: "Print the auth token store path",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println(authtoken.NewStore(authTokenRoot).Root)
		return nil
	},
}

var authWindsurfCmd = &cobra.Command{
	Use:   "windsurf",
	Short: "Print safe Windsurf token import instructions",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("Windsurf token import flow:")
		fmt.Println("1. Open this URL in your browser while logged into Windsurf:")
		fmt.Println("   https://windsurf.com/editor/show-auth-token")
		fmt.Println("2. Copy your own token into a local file that is not committed, for example ~/.ti/auth/windsurf.token")
		fmt.Println("3. Import it:")
		fmt.Println("   ti-cli auth token import windsurf --provider windsurf --file ~/.ti/auth/windsurf.token")
		fmt.Println("Ti will store the token in a 0600 local profile and only print redacted previews.")
		return nil
	},
}

func readTokenInput() (token string, source string, err error) {
	if authTokenFile != "" {
		data, err := os.ReadFile(authTokenFile)
		if err != nil {
			return "", "", err
		}
		return strings.TrimSpace(string(data)), authTokenFile, nil
	}
	if authTokenStdin {
		data, err := io.ReadAll(bufio.NewReader(os.Stdin))
		if err != nil {
			return "", "", err
		}
		return strings.TrimSpace(string(data)), "stdin", nil
	}
	return "", "", fmt.Errorf("provide --file or --stdin; direct token arguments are intentionally not supported to avoid shell history leaks")
}

func init() {
	authTokenCmd.PersistentFlags().StringVar(&authTokenRoot, "store", "", "auth token store root (default ~/.ti/auth/tokens)")
	authTokenCmd.PersistentFlags().BoolVar(&authTokenJSON, "json", false, "print JSON output")
	authTokenImportCmd.Flags().StringVar(&authTokenProvider, "provider", "generic", "provider name, for example windsurf, opencode, openrouter")
	authTokenImportCmd.Flags().StringVarP(&authTokenFile, "file", "f", "", "local file containing the token")
	authTokenImportCmd.Flags().BoolVar(&authTokenStdin, "stdin", false, "read token from stdin")
	authTokenCmd.AddCommand(authTokenImportCmd, authTokenListCmd, authTokenShowCmd, authTokenRemoveCmd, authTokenPathCmd)
	authCmd.AddCommand(authTokenCmd, authWindsurfCmd)
}
