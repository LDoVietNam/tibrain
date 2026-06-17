package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/ti/cli/internal/config"
	cookiestore "github.com/ti/cli/internal/cookie"
)

var (
	authCookieFile  string
	authCookieState string
)

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Manage local auth material",
	Long:  "Manage local auth material such as browser-exported cookie profiles. Secrets are stored outside project config and are never printed in full.",
}

var authImportCookiesCmd = &cobra.Command{
	Use:   "import-cookies [profile]",
	Short: "Import a browser cookie JSON export into the Ti cookie state",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadAuto()
		if err != nil {
			return fmt.Errorf("load config: %w", err)
		}
		profile := cfg.CookieProfile
		if len(args) == 1 && strings.TrimSpace(args[0]) != "" {
			profile = strings.TrimSpace(args[0])
		}
		if profile == "" {
			profile = "sharedchat"
		}
		if authCookieFile == "" {
			return fmt.Errorf("--file is required; pass a local browser cookie JSON export")
		}
		stateFile := authCookieState
		if stateFile == "" {
			stateFile = cfg.CookieFile
		}
		stateFile = config.ExpandHome(stateFile)

		mgr := cookiestore.NewManager()
		if _, err := os.Stat(stateFile); err == nil {
			_ = mgr.LoadStateFile(stateFile)
		}
		if err := mgr.LoadFile(config.ExpandHome(authCookieFile), profile); err != nil {
			return err
		}
		if err := mgr.SaveStateFile(stateFile); err != nil {
			return err
		}
		_ = os.Chmod(stateFile, 0600)

		summary := mgr.Summaries()
		fmt.Printf("Imported cookie profile %q into %s\n", profile, config.ContractHome(stateFile))
		for _, item := range summary {
			if item.Name == profile {
				fmt.Printf("Profile: %s\nDomain: %s\nCookies: %d\nSession: %s\n", item.Name, item.Domain, item.CookieCount, yesNo(item.HasSessionID))
				if item.SessionPreview != "" {
					fmt.Printf("Session preview: %s\n", item.SessionPreview)
				}
			}
		}
		return nil
	},
}

var authCookiesCmd = &cobra.Command{
	Use:   "cookies",
	Short: "List imported cookie profiles",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadAuto()
		if err != nil {
			return fmt.Errorf("load config: %w", err)
		}
		stateFile := authCookieState
		if stateFile == "" {
			stateFile = cfg.CookieFile
		}
		mgr, err := loadCookieState(config.ExpandHome(stateFile))
		if err != nil {
			return err
		}
		summaries := mgr.Summaries()
		if len(summaries) == 0 {
			fmt.Println("No cookie profiles imported.")
			return nil
		}
		fmt.Printf("Cookie state: %s\n\n", config.ContractHome(config.ExpandHome(stateFile)))
		for _, s := range summaries {
			fmt.Printf("- %s\n  domain: %s\n  cookies: %d\n  session: %s", s.Name, s.Domain, s.CookieCount, yesNo(s.HasSessionID))
			if s.SessionPreview != "" {
				fmt.Printf(" (%s)", s.SessionPreview)
			}
			if s.ContainsMetadata {
				fmt.Print("\n  metadata: yes")
			}
			fmt.Println()
		}
		return nil
	},
}

var authStatusCmd = &cobra.Command{
	Use:   "status [profile]",
	Short: "Check an imported cookie profile without revealing values",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadAuto()
		if err != nil {
			return fmt.Errorf("load config: %w", err)
		}
		profile := cfg.CookieProfile
		if len(args) == 1 {
			profile = args[0]
		}
		stateFile := authCookieState
		if stateFile == "" {
			stateFile = cfg.CookieFile
		}
		mgr, err := loadCookieState(config.ExpandHome(stateFile))
		if err != nil {
			return err
		}
		var found bool
		for _, s := range mgr.Summaries() {
			if s.Name != profile {
				continue
			}
			found = true
			fmt.Printf("Profile: %s\nDomain: %s\nCookies: %d\nHas session: %s\nExpired: %s\nUpdated: %s\n", s.Name, s.Domain, s.CookieCount, yesNo(s.HasSessionID), yesNo(mgr.IsExpired(profile)), s.UpdatedAt.Format("2006-01-02 15:04:05"))
			if s.SessionPreview != "" {
				fmt.Printf("Session preview: %s\n", s.SessionPreview)
			}
		}
		if !found {
			return fmt.Errorf("cookie profile %q not found", profile)
		}
		return nil
	},
}

var authRemoveCookiesCmd = &cobra.Command{
	Use:   "remove-cookies [profile]",
	Short: "Remove an imported cookie profile",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadAuto()
		if err != nil {
			return fmt.Errorf("load config: %w", err)
		}
		stateFile := authCookieState
		if stateFile == "" {
			stateFile = cfg.CookieFile
		}
		stateFile = config.ExpandHome(stateFile)
		mgr, err := loadCookieState(stateFile)
		if err != nil {
			return err
		}
		mgr.Remove(args[0])
		if err := mgr.SaveStateFile(stateFile); err != nil {
			return err
		}
		fmt.Printf("Removed cookie profile %q from %s\n", args[0], config.ContractHome(stateFile))
		return nil
	},
}

var authPathCmd = &cobra.Command{
	Use:   "path",
	Short: "Print auth storage paths",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadAuto()
		if err != nil {
			return fmt.Errorf("load config: %w", err)
		}
		out := map[string]string{
			"auth_dir":    config.ContractHome(cfg.AuthDir),
			"cookie_file": config.ContractHome(cfg.CookieFile),
		}
		data, _ := json.MarshalIndent(out, "", "  ")
		fmt.Println(string(data))
		return nil
	},
}

func loadCookieState(path string) (*cookiestore.Manager, error) {
	mgr := cookiestore.NewManager()
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return mgr, nil
		}
		return nil, err
	}
	if err := mgr.LoadStateFile(path); err != nil {
		return nil, err
	}
	return mgr, nil
}

func yesNo(v bool) string {
	if v {
		return "yes"
	}
	return "no"
}

func init() {
	rootCmd.AddCommand(authCmd)
	authCmd.PersistentFlags().StringVar(&authCookieState, "state", "", "cookie state file (default from config cookie_file)")
	authImportCookiesCmd.Flags().StringVarP(&authCookieFile, "file", "f", "", "browser cookie JSON export to import")
	authCmd.AddCommand(authImportCookiesCmd)
	authCmd.AddCommand(authCookiesCmd)
	authCmd.AddCommand(authStatusCmd)
	authCmd.AddCommand(authRemoveCookiesCmd)
	authCmd.AddCommand(authPathCmd)

}
