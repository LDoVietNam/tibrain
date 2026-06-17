package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/ti/cli/internal/session"
)

var (
	sessionDBPath string
	sessionLimit  int
	sessionJSON   bool
)

var sessionCmd = &cobra.Command{Use: "session", Short: "Manage persistent Ti sessions", Long: "Create, list, inspect, fork, archive, import, and export persistent Ti sessions."}

func openSessionStore() (*session.Store, error) { return session.NewStore(sessionDBPath) }

var sessionPathCmd = &cobra.Command{Use: "path", Short: "Print the session database path", RunE: func(cmd *cobra.Command, args []string) error {
	if sessionDBPath != "" {
		fmt.Println(sessionDBPath)
		return nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	fmt.Println(home + "/.ti/data/sessions.db")
	return nil
}}

var sessionListCmd = &cobra.Command{Use: "list", Short: "List recent sessions", RunE: func(cmd *cobra.Command, args []string) error {
	store, err := openSessionStore()
	if err != nil {
		return err
	}
	defer store.Close()
	sessions, err := store.List(sessionLimit)
	if err != nil {
		return err
	}
	if sessionJSON {
		data, _ := json.MarshalIndent(sessions, "", "  ")
		fmt.Println(string(data))
		return nil
	}
	if len(sessions) == 0 {
		fmt.Println("No sessions yet.")
		return nil
	}
	fmt.Printf("%-38s  %-19s  %-12s  %-18s  %s\n", "ID", "UPDATED", "PROVIDER", "MODEL", "TITLE")
	for _, s := range sessions {
		updated := time.Unix(s.UpdatedAt, 0).Format("2006-01-02 15:04:05")
		title := s.Title
		if title == "" && len(s.Messages) > 0 {
			title = trimOneLine(s.Messages[0].Content, 60)
		}
		fmt.Printf("%-38s  %-19s  %-12s  %-18s  %s\n", s.ID, updated, s.Provider, s.Model, title)
	}
	return nil
}}

var sessionStatsCmd = &cobra.Command{Use: "stats", Short: "Show session store statistics", RunE: func(cmd *cobra.Command, args []string) error {
	store, err := openSessionStore()
	if err != nil {
		return err
	}
	defer store.Close()
	stats, err := store.Stats()
	if err != nil {
		return err
	}
	if sessionJSON {
		data, _ := json.MarshalIndent(stats, "", "  ")
		fmt.Println(string(data))
		return nil
	}
	fmt.Printf("Total: %d\nActive: %d\nArchived: %d\nDistinct phases: %d\n", stats.Total, stats.Active, stats.Archived, stats.DistinctPhases)
	return nil
}}

var sessionShowCmd = &cobra.Command{Use: "show <id>", Short: "Show a session", Args: cobra.ExactArgs(1), RunE: func(cmd *cobra.Command, args []string) error {
	store, err := openSessionStore()
	if err != nil {
		return err
	}
	defer store.Close()
	s, err := store.Get(args[0])
	if err != nil {
		return err
	}
	if s == nil {
		return fmt.Errorf("session not found: %s", args[0])
	}
	data, _ := json.MarshalIndent(s, "", "  ")
	fmt.Println(string(data))
	return nil
}}

var sessionCreateCmd = &cobra.Command{Use: "create [title]", Short: "Create a new empty session", Args: cobra.MaximumNArgs(1), RunE: func(cmd *cobra.Command, args []string) error {
	provider, _ := cmd.Flags().GetString("provider")
	model, _ := cmd.Flags().GetString("model")
	phase, _ := cmd.Flags().GetString("phase")
	project, _ := cmd.Flags().GetString("project")
	title := ""
	if len(args) == 1 {
		title = args[0]
	}
	cwd, _ := os.Getwd()
	store, err := openSessionStore()
	if err != nil {
		return err
	}
	defer store.Close()
	s := &session.Session{Provider: provider, Model: model, Phase: phase, Project: project, Title: title, Dir: cwd}
	if err := store.Create(s); err != nil {
		return err
	}
	fmt.Println(s.ID)
	return nil
}}

var sessionAppendCmd = &cobra.Command{Use: "append <id> <role> <content>", Short: "Append a message to a session", Args: cobra.ExactArgs(3), RunE: func(cmd *cobra.Command, args []string) error {
	role := strings.TrimSpace(args[1])
	switch role {
	case "system", "user", "assistant", "tool":
	default:
		return fmt.Errorf("role must be system, user, assistant, or tool")
	}
	store, err := openSessionStore()
	if err != nil {
		return err
	}
	defer store.Close()
	return store.AppendMessage(args[0], session.Message{Role: role, Content: args[2]}, 0)
}}

var sessionExportCmd = &cobra.Command{Use: "export <id> [file]", Short: "Export a session as JSON", Args: cobra.RangeArgs(1, 2), RunE: func(cmd *cobra.Command, args []string) error {
	store, err := openSessionStore()
	if err != nil {
		return err
	}
	defer store.Close()
	data, err := store.Export(args[0])
	if err != nil {
		return err
	}
	if len(args) == 2 {
		return os.WriteFile(args[1], data, 0600)
	}
	fmt.Println(string(data))
	return nil
}}

var sessionImportCmd = &cobra.Command{Use: "import <file>", Short: "Import a session JSON file", Args: cobra.ExactArgs(1), RunE: func(cmd *cobra.Command, args []string) error {
	data, err := os.ReadFile(args[0])
	if err != nil {
		return err
	}
	store, err := openSessionStore()
	if err != nil {
		return err
	}
	defer store.Close()
	s, err := store.Import(data)
	if err != nil {
		return err
	}
	fmt.Println(s.ID)
	return nil
}}

var sessionForkCmd = &cobra.Command{Use: "fork <id> [title]", Short: "Fork a session into a new branch", Args: cobra.RangeArgs(1, 2), RunE: func(cmd *cobra.Command, args []string) error {
	title := ""
	if len(args) == 2 {
		title = args[1]
	}
	store, err := openSessionStore()
	if err != nil {
		return err
	}
	defer store.Close()
	s, err := store.Fork(args[0], title)
	if err != nil {
		return err
	}
	fmt.Println(s.ID)
	return nil
}}

var sessionArchiveCmd = &cobra.Command{Use: "archive <id>", Short: "Archive a session", Args: cobra.ExactArgs(1), RunE: func(cmd *cobra.Command, args []string) error {
	store, err := openSessionStore()
	if err != nil {
		return err
	}
	defer store.Close()
	return store.Archive(args[0], true)
}}

var sessionDeleteCmd = &cobra.Command{Use: "delete <id>", Short: "Delete a session", Args: cobra.ExactArgs(1), RunE: func(cmd *cobra.Command, args []string) error {
	store, err := openSessionStore()
	if err != nil {
		return err
	}
	defer store.Close()
	return store.Delete(args[0])
}}

func trimOneLine(s string, n int) string {
	s = strings.Join(strings.Fields(s), " ")
	if len(s) <= n {
		return s
	}
	if n < 4 {
		return s[:n]
	}
	return s[:n-3] + "..."
}

func init() {
	rootCmd.AddCommand(sessionCmd)
	sessionCmd.PersistentFlags().StringVar(&sessionDBPath, "db", "", "session database path (default ~/.ti/data/sessions.db)")
	sessionCmd.PersistentFlags().IntVar(&sessionLimit, "limit", 20, "maximum sessions to list")
	sessionCmd.PersistentFlags().BoolVar(&sessionJSON, "json", false, "print JSON output")
	sessionCreateCmd.Flags().String("provider", "", "provider name")
	sessionCreateCmd.Flags().String("model", "", "model name")
	sessionCreateCmd.Flags().String("phase", "", "phase label")
	sessionCreateCmd.Flags().String("project", "", "project label")
	sessionCmd.AddCommand(sessionPathCmd, sessionListCmd, sessionStatsCmd, sessionShowCmd, sessionCreateCmd, sessionAppendCmd, sessionExportCmd, sessionImportCmd, sessionForkCmd, sessionArchiveCmd, sessionDeleteCmd)
}
