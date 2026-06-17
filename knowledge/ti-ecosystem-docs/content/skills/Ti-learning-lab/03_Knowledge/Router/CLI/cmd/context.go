package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var contextCmd = &cobra.Command{Use: "context", Short: "Project context and AGENTS.md learning", Long: "Scan, show, and create AGENTS.md files so Ti can learn project-specific rules."}

var contextScanCmd = &cobra.Command{Use: "scan [path]", Short: "Find AGENTS.md files from a path up to the repo root", Args: cobra.MaximumNArgs(1), RunE: func(cmd *cobra.Command, args []string) error {
	start := "."; if len(args) == 1 { start = args[0] }
	files, err := findAgentFiles(start); if err != nil { return err }
	if len(files) == 0 { fmt.Println("No AGENTS.md found."); return nil }
	for _, f := range files { fmt.Println(f) }
	return nil
}}

var contextShowCmd = &cobra.Command{Use: "show [path]", Short: "Print merged AGENTS.md context", Args: cobra.MaximumNArgs(1), RunE: func(cmd *cobra.Command, args []string) error {
	start := "."; if len(args) == 1 { start = args[0] }
	files, err := findAgentFiles(start); if err != nil { return err }
	if len(files) == 0 { fmt.Println("No AGENTS.md found."); return nil }
	for _, f := range files {
		data, err := os.ReadFile(f); if err != nil { return err }
		fmt.Printf("\n# %s\n\n%s\n", f, strings.TrimSpace(string(data)))
	}
	return nil
}}

var contextInitCmd = &cobra.Command{Use: "init", Short: "Create AGENTS.md in the current directory", RunE: func(cmd *cobra.Command, args []string) error {
	path := "AGENTS.md"
	if _, err := os.Stat(path); err == nil { return fmt.Errorf("%s already exists", path) }
	content := `# AGENTS.md

## Project rules
- Keep changes small and reviewable.
- Prefer existing patterns before adding new abstractions.
- Run tests or explain why they could not be run.

## Safe tool policy
- Ask before destructive shell commands.
- Do not print secrets or commit auth material.
- Do not modify generated/vendor files unless explicitly requested.

## Build
- Go version: 1.23.x
- Preferred build: make build-small
`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil { return err }
	fmt.Println("Created AGENTS.md")
	return nil
}}

func findAgentFiles(start string) ([]string, error) {
	abs, err := filepath.Abs(start); if err != nil { return nil, err }
	info, err := os.Stat(abs); if err != nil { return nil, err }
	if !info.IsDir() { abs = filepath.Dir(abs) }
	var reversed []string
	for {
		candidate := filepath.Join(abs, "AGENTS.md")
		if _, err := os.Stat(candidate); err == nil { reversed = append(reversed, candidate) }
		parent := filepath.Dir(abs)
		if parent == abs { break }
		if _, err := os.Stat(filepath.Join(abs, ".git")); err == nil && len(reversed) > 0 { break }
		abs = parent
	}
	files := make([]string, 0, len(reversed))
	for i := len(reversed)-1; i >= 0; i-- { files = append(files, reversed[i]) }
	return files, nil
}

func init() {
	rootCmd.AddCommand(contextCmd)
	contextCmd.AddCommand(contextScanCmd, contextShowCmd, contextInitCmd)
}
