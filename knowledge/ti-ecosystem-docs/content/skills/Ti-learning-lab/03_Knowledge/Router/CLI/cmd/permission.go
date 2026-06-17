package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/ti/cli/internal/permission"
)

var (
	permissionFile string
	permissionJSON bool
)

var permissionCmd = &cobra.Command{Use: "permission", Short: "Manage Ti tool permission policy", Long: "Inspect and edit the local permission policy used before Ti runs shell, file, and destructive tools."}

func defaultPermissionPath() string {
	home, err := os.UserHomeDir()
	if err != nil { return ".ti/permissions.json" }
	return filepath.Join(home, ".ti", "permissions.json")
}

func permissionPath() string {
	if strings.TrimSpace(permissionFile) != "" { return permissionFile }
	return defaultPermissionPath()
}

func loadPermissionPolicy() (*permission.Policy, error) { return permission.LoadPolicy(permissionPath()) }

var permissionPathCmd = &cobra.Command{Use: "path", Short: "Print permission policy path", Run: func(cmd *cobra.Command, args []string) { fmt.Println(permissionPath()) }}

var permissionListCmd = &cobra.Command{Use: "list", Short: "Show permission policy", RunE: func(cmd *cobra.Command, args []string) error {
	p, err := loadPermissionPolicy(); if err != nil { return err }
	if permissionJSON { data, _ := json.MarshalIndent(p, "", "  "); fmt.Println(string(data)); return nil }
	fmt.Printf("Mode: %s\n", p.Mode)
	fmt.Println("Allow:"); for _, r := range p.Allow { fmt.Println("  - " + r) }
	fmt.Println("Ask:"); for _, r := range p.Ask { fmt.Println("  - " + r) }
	fmt.Println("Deny:"); for _, r := range p.Deny { fmt.Println("  - " + r) }
	return nil
}}

var permissionInitCmd = &cobra.Command{Use: "init", Short: "Create a default permission policy", RunE: func(cmd *cobra.Command, args []string) error {
	p := permission.DefaultPolicy()
	if err := p.Save(permissionPath()); err != nil { return err }
	fmt.Println("Created " + permissionPath())
	return nil
}}

var permissionModeCmd = &cobra.Command{Use: "mode <ask|auto|yes|deny|plan|smart>", Short: "Set permission mode", Args: cobra.ExactArgs(1), RunE: func(cmd *cobra.Command, args []string) error {
	mode := permission.PermissionMode(args[0])
	if !mode.IsValid() { return fmt.Errorf("invalid mode: %s", args[0]) }
	p, err := loadPermissionPolicy(); if err != nil { return err }
	p.Mode = mode
	if err := p.Save(permissionPath()); err != nil { return err }
	fmt.Printf("Permission mode set to %s\n", mode)
	return nil
}}

func addPermissionRule(behavior permission.PermissionBehavior, rule string) error {
	p, err := loadPermissionPolicy(); if err != nil { return err }
	p.AddRule(behavior, rule)
	return p.Save(permissionPath())
}

var permissionAllowCmd = &cobra.Command{Use: "allow <tool:pattern>", Short: "Add an allow rule", Args: cobra.ExactArgs(1), RunE: func(cmd *cobra.Command, args []string) error { return addPermissionRule(permission.BehaviorAllow, args[0]) }}
var permissionDenyCmd = &cobra.Command{Use: "deny <tool:pattern>", Short: "Add a deny rule", Args: cobra.ExactArgs(1), RunE: func(cmd *cobra.Command, args []string) error { return addPermissionRule(permission.BehaviorDeny, args[0]) }}
var permissionAskCmd = &cobra.Command{Use: "ask <tool:pattern>", Short: "Add an ask rule", Args: cobra.ExactArgs(1), RunE: func(cmd *cobra.Command, args []string) error { return addPermissionRule(permission.BehaviorAsk, args[0]) }}

var permissionCheckCmd = &cobra.Command{Use: "check <tool> [path-or-command]", Short: "Evaluate a tool call against policy", Args: cobra.RangeArgs(1, 2), RunE: func(cmd *cobra.Command, args []string) error {
	p, err := loadPermissionPolicy(); if err != nil { return err }
	pathOrCommand := ""; if len(args) == 2 { pathOrCommand = args[1] }
	filePath := pathOrCommand
	command := ""
	if args[0] == "bash" { command = pathOrCommand; filePath = "" }
	decision := p.Evaluate(args[0], filePath, command)
	if permissionJSON { data, _ := json.MarshalIndent(decision, "", "  "); fmt.Println(string(data)); return nil }
	fmt.Printf("Decision: %s\nReason: %s\n", decision.Behavior, decision.Reason)
	return nil
}}

func init() {
	rootCmd.AddCommand(permissionCmd)
	permissionCmd.PersistentFlags().StringVar(&permissionFile, "file", "", "permission policy file (default ~/.ti/permissions.json)")
	permissionCmd.PersistentFlags().BoolVar(&permissionJSON, "json", false, "print JSON output")
	permissionCmd.AddCommand(permissionPathCmd, permissionListCmd, permissionInitCmd, permissionModeCmd, permissionAllowCmd, permissionDenyCmd, permissionAskCmd, permissionCheckCmd)
}
