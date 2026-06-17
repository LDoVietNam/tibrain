package account

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/ti/cli/internal/browser"
	"github.com/ti/cli/internal/config"
)

type CLI struct {
	browserMgr *browser.BrowserManager
	dataDir    string
}

func NewCLI(dataDir string) (*CLI, error) {
	bm, err := browser.NewBrowserManager(dataDir)
	if err != nil {
		return nil, err
	}
	return &CLI{
		browserMgr: bm,
		dataDir:    dataDir,
	}, nil
}

func (c *CLI) AddAccount(email, name string) error {
	acc, err := c.browserMgr.AddAccount(email, name)
	if err != nil {
		return err
	}
	fmt.Printf("Added account: %s (%s)\n", acc.Email, acc.ID)
	return nil
}

func (c *CLI) RemoveAccount(id string) error {
	acc := c.browserMgr.GetAccount(id)
	if acc == nil {
		return fmt.Errorf("account not found: %s", id)
	}
	if err := c.browserMgr.RemoveAccount(id); err != nil {
		return err
	}
	fmt.Printf("Removed account: %s\n", acc.Email)
	return nil
}

func (c *CLI) ListAccounts() error {
	accounts := c.browserMgr.ListAccounts()
	if len(accounts) == 0 {
		fmt.Println("No accounts")
		return nil
	}
	fmt.Println("Accounts:")
	for _, acc := range accounts {
		active := ""
		if acc.Active {
			active = " (active)"
		}
		fmt.Printf("  %s  %s  %s%s\n", acc.ID, acc.Email, acc.Name, active)
	}
	return nil
}

func (c *CLI) SetActiveAccount(id string) error {
	acc := c.browserMgr.GetAccount(id)
	if acc == nil {
		return fmt.Errorf("account not found: %s", id)
	}
	if err := c.browserMgr.SetActiveAccount(id); err != nil {
		return err
	}
	fmt.Printf("Active account: %s\n", acc.Email)
	return nil
}

func (c *CLI) AddProfile(name string, browserType browser.BrowserType, userDataDir string) error {
	activeAcc := c.browserMgr.GetActiveAccount()
	accountID := ""
	if activeAcc != nil {
		accountID = activeAcc.ID
	}

	prof, err := c.browserMgr.AddProfile(name, browserType, userDataDir, accountID)
	if err != nil {
		return err
	}
	fmt.Printf("Added profile: %s (%s) for %s\n", prof.Name, prof.ID, prof.Browser)
	return nil
}

func (c *CLI) RemoveProfile(id string) error {
	prof := c.browserMgr.GetProfile(id)
	if prof == nil {
		return fmt.Errorf("profile not found: %s", id)
	}
	if err := c.browserMgr.RemoveProfile(id); err != nil {
		return err
	}
	fmt.Printf("Removed profile: %s\n", prof.Name)
	return nil
}

func (c *CLI) ListProfiles() error {
	profiles := c.browserMgr.ListProfiles()
	if len(profiles) == 0 {
		fmt.Println("No profiles")
		return nil
	}
	fmt.Println("Profiles:")
	for _, prof := range profiles {
		active := ""
		if prof.Active {
			active = " (active)"
		}
		accountInfo := ""
		if prof.AccountID != "" {
			accountInfo = fmt.Sprintf(" [account: %s]", prof.AccountID)
		}
		fmt.Printf("  %s  %s  %s%s%s\n", prof.ID, prof.Name, prof.Browser, active, accountInfo)
	}
	return nil
}

func (c *CLI) SetActiveProfile(id string) error {
	prof := c.browserMgr.GetProfile(id)
	if prof == nil {
		return fmt.Errorf("profile not found: %s", id)
	}
	if err := c.browserMgr.SetActiveProfile(id); err != nil {
		return err
	}
	fmt.Printf("Active profile: %s\n", prof.Name)
	return nil
}

func (c *CLI) LinkAccount(profileID, accountID string) error {
	if err := c.browserMgr.LinkAccountToProfile(profileID, accountID); err != nil {
		return err
	}
	fmt.Printf("Linked profile %s to account %s\n", profileID, accountID)
	return nil
}

func (c *CLI) OpenBrowser(profileID string, headless bool) (string, error) {
	session, err := c.browserMgr.NewSession(profileID, headless)
	if err != nil {
		return "", err
	}
	fmt.Printf("Opened browser session: %s\n", session.ID)
	return session.ID, nil
}

func (c *CLI) CloseSession(sessionID string) error {
	if err := c.browserMgr.CloseSession(sessionID); err != nil {
		return err
	}
	fmt.Printf("Closed session: %s\n", sessionID)
	return nil
}

func (c *CLI) Navigate(sessionID, url string) error {
	session := c.browserMgr.GetSession(sessionID)
	if session == nil {
		return fmt.Errorf("session not found: %s", sessionID)
	}
	if err := session.Navigate(url); err != nil {
		return err
	}
	fmt.Printf("Navigated to: %s\n", url)
	return nil
}

func (c *CLI) RunOAuthFlow(accountID, url string) error {
	bm := c.browserMgr
	activeProf := bm.GetActiveProfile()
	if activeProf == nil {
		return fmt.Errorf("no active profile")
	}

	session, err := bm.NewSession(activeProf.ID, false)
	if err != nil {
		return err
	}
	defer session.Close()

	if err := session.Navigate(url); err != nil {
		return err
	}

	fmt.Println("Browser opened. Complete OAuth authorization in the browser.")
	fmt.Println("Press Enter when done...")
	fmt.Scanln()

	return nil
}

func AccountCommand(args []string, cfg *config.Config) error {
	cli, err := NewCLI(cfg.DataDir)
	if err != nil {
		return err
	}

	if len(args) == 0 {
		printAccountHelp()
		return nil
	}

	switch args[0] {
	case "add":
		if len(args) < 3 {
			return fmt.Errorf("usage: ti account add <email> <name>")
		}
		return cli.AddAccount(args[1], args[2])

	case "remove", "rm", "delete":
		if len(args) < 2 {
			return fmt.Errorf("usage: ti account remove <id>")
		}
		return cli.RemoveAccount(args[1])

	case "list", "ls":
		return cli.ListAccounts()

	case "active":
		if len(args) < 2 {
			active := cli.browserMgr.GetActiveAccount()
			if active != nil {
				fmt.Printf("Active: %s (%s)\n", active.Email, active.ID)
			}
			return nil
		}
		return cli.SetActiveAccount(args[1])

	case "profile":
		return profileSubcommand(args[1:], cli)

	case "oauth":
		return oauthSubcommand(args[1:], cli)

	case "export":
		if len(args) < 2 {
			return fmt.Errorf("usage: ti account export <file>")
		}
		return cli.ExportAccounts(args[1])

	case "import":
		if len(args) < 2 {
			return fmt.Errorf("usage: ti account import <file>")
		}
		return cli.ImportAccounts(args[1])

	default:
		return fmt.Errorf("unknown subcommand: %s (try 'ti account help')", args[0])
	}
}

func profileSubcommand(args []string, cli *CLI) error {
	if len(args) == 0 {
		return cli.ListProfiles()
	}

	switch args[0] {
	case "add":
		if len(args) < 2 {
			return fmt.Errorf("usage: ti account profile add <name> [browser] [user-data-dir]")
		}
		name := args[1]
		browserType := browser.Chrome
		userDataDir := ""

		if len(args) > 2 {
			browserType = browser.BrowserType(args[2])
		}
		if len(args) > 3 {
			userDataDir = args[3]
		}

		return cli.AddProfile(name, browserType, userDataDir)

	case "remove", "rm":
		if len(args) < 2 {
			return fmt.Errorf("usage: ti account profile remove <id>")
		}
		return cli.RemoveProfile(args[1])

	case "list", "ls":
		return cli.ListProfiles()

	case "active":
		if len(args) < 2 {
			active := cli.browserMgr.GetActiveProfile()
			if active != nil {
				fmt.Printf("Active: %s (%s)\n", active.Name, active.ID)
			}
			return nil
		}
		return cli.SetActiveProfile(args[1])

	case "link":
		if len(args) < 3 {
			return fmt.Errorf("usage: ti account profile link <profile-id> <account-id>")
		}
		return cli.LinkAccount(args[1], args[2])

	case "open":
		if len(args) < 2 {
			return fmt.Errorf("usage: ti account profile open <profile-id> [--headless]")
		}

		profileID := args[1]
		headless := false

		for i := 2; i < len(args); i++ {
			if args[i] == "--headless" {
				headless = true
			}
		}

		_, err := cli.OpenBrowser(profileID, headless)
		return err

	case "close":
		if len(args) < 2 {
			return fmt.Errorf("usage: ti account profile close <session-id>")
		}
		return cli.CloseSession(args[1])

	case "navigate", "goto":
		if len(args) < 3 {
			return fmt.Errorf("usage: ti account profile navigate <session-id> <url>")
		}
		return cli.Navigate(args[1], args[2])

	default:
		return fmt.Errorf("unknown subcommand: %s (try 'ti account profile help')", args[0])
	}
}

func oauthSubcommand(args []string, cli *CLI) error {
	if len(args) == 0 {
		fmt.Println("Usage: ti account oauth <subcommand>")
		fmt.Println("  start <account-id>    Start OAuth flow in browser")
		fmt.Println("  url <account-id>     Get OAuth URL for account")
		return nil
	}

	switch args[0] {
	case "start":
		accountID := ""

		if len(args) > 1 {
			accountID = args[1]
		}

		if accountID == "" {
			active := cli.browserMgr.GetActiveAccount()
			if active != nil {
				accountID = active.ID
			}
		}

		if accountID == "" {
			return fmt.Errorf("no account specified")
		}

		oauthURL := fmt.Sprintf("https://accounts.google.com/o/oauth2/auth?client_id=%s&redirect_uri=http://localhost:1808/oauth/callback&scope=https://www.googleapis.com/auth/gmail.readonly+https://www.googleapis.com/auth/gmail.send&response_type=code&access_type=offline",
			os.Getenv("GOOGLE_CLIENT_ID"))

		return cli.RunOAuthFlow(accountID, oauthURL)

	case "url":
		if len(args) < 2 {
			return fmt.Errorf("usage: ti account oauth url <account-id>")
		}

		clientID := os.Getenv("GOOGLE_CLIENT_ID")
		if clientID == "" {
			clientID = "YOUR_CLIENT_ID"
		}

		url := fmt.Sprintf("https://accounts.google.com/o/oauth2/auth?client_id=%s&redirect_uri=http://localhost:1808/oauth/callback&scope=https://www.googleapis.com/auth/gmail.readonly+https://www.googleapis.com/auth/gmail.send&response_type=code&access_type=offline",
			clientID)

		fmt.Println(url)
		return nil

	default:
		return fmt.Errorf("unknown subcommand: %s", args[0])
	}
}

func printAccountHelp() {
	fmt.Println("Usage: ti account <subcommand>")
	fmt.Println("")
	fmt.Println("Subcommands:")
	fmt.Println("  add <email> <name>           Add a new account")
	fmt.Println("  remove <id>                 Remove an account")
	fmt.Println("  list                        List all accounts")
	fmt.Println("  active [id]                 Show or set active account")
	fmt.Println("  profile <subcommand>        Manage browser profiles")
	fmt.Println("  oauth <subcommand>         OAuth flow commands")
	fmt.Println("  export <file>               Export accounts to JSON")
	fmt.Println("  import <file>               Import accounts from JSON")
	fmt.Println("")
	fmt.Println("Profile subcommands:")
	fmt.Println("  profile add <name> [browser] [user-data-dir]")
	fmt.Println("  profile remove <id>")
	fmt.Println("  profile list")
	fmt.Println("  profile active [id]")
	fmt.Println("  profile link <profile-id> <account-id>")
	fmt.Println("  profile open <profile-id> [--headless]")
	fmt.Println("  profile close <session-id>")
	fmt.Println("  profile navigate <session-id> <url>")
	fmt.Println("")
	fmt.Println("OAuth subcommands:")
	fmt.Println("  oauth start [account-id]")
	fmt.Println("  oauth url <account-id>")
}

func ResolveAutoDataDir() string {
	home, _ := os.UserHomeDir()
	if home != "" {
		return filepath.Join(home, ".ti", "data")
	}
	return "~/.ti/data"
}

func ResolveEnvAccountDataDir() string {
	if dir := os.Getenv("TI_ACCOUNT_DIR"); dir != "" {
		return dir
	}
	return ResolveAutoDataDir()
}

var DefaultAccountDataDir = ResolveEnvAccountDataDir()

type AccountSummary struct {
	ID       string `json:"id"`
	Email    string `json:"email"`
	Name     string `json:"name"`
	Active   bool   `json:"active"`
	Profiles int    `json:"profiles"`
}

func (c *CLI) ExportAccounts(filePath string) error {
	accounts := c.browserMgr.ListAccounts()
	var summaries []AccountSummary

	for _, acc := range accounts {
		profiles := c.browserMgr.GetByAccount(acc.ID)
		summaries = append(summaries, AccountSummary{
			ID:       acc.ID,
			Email:    acc.Email,
			Name:     acc.Name,
			Active:   acc.Active,
			Profiles: len(profiles),
		})
	}

	data, err := json.MarshalIndent(summaries, "", "  ")
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		return err
	}

	if err := os.WriteFile(filePath, data, 0600); err != nil {
		return err
	}

	fmt.Printf("Exported %d accounts to: %s\n", len(summaries), filePath)
	return nil
}

func (c *CLI) ImportAccounts(filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	var summaries []AccountSummary
	if err := json.Unmarshal(data, &summaries); err != nil {
		return err
	}

	for _, acc := range summaries {
		if _, err := c.browserMgr.AddAccount(acc.Email, acc.Name); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: %s - %v\n", acc.Email, err)
		}
	}

	fmt.Printf("Imported %d accounts from: %s\n", len(summaries), filePath)
	return nil
}
