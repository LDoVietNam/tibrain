package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/ti/cli/internal/blocks"
	"github.com/ti/cli/internal/tools"
)

var toolCmd = &cobra.Command{Use: "tool", Short: "Inspect and run Ti agent tools", Long: "List tool definitions and run local tools through the same registry used by Ti agents."}

var toolListCmd = &cobra.Command{Use: "list", Short: "List available agent tools", RunE: func(cmd *cobra.Command, args []string) error {
	asJSON, _ := cmd.Flags().GetBool("json")
	defs := tools.AllDefs()
	if asJSON {
		data, _ := json.MarshalIndent(defs, "", "  ")
		fmt.Println(string(data))
		return nil
	}
	for _, d := range defs {
		fmt.Printf("%-18s %s\n", d.Name, d.Description)
	}
	return nil
}}

var toolRunCmd = &cobra.Command{Use: "run <name> [json-or-key=value...]", Short: "Run a tool with JSON or key=value arguments", Args: cobra.MinimumNArgs(1), RunE: func(cmd *cobra.Command, args []string) error {
	input := map[string]any{}
	if len(args) >= 2 {
		joined := strings.Join(args[1:], " ")
		if strings.HasPrefix(strings.TrimSpace(joined), "{") {
			if err := json.Unmarshal([]byte(joined), &input); err != nil {
				return fmt.Errorf("invalid JSON input: %w", err)
			}
		} else {
			for _, item := range args[1:] {
				k, v, ok := strings.Cut(item, "=")
				if !ok {
					return fmt.Errorf("argument %q must be key=value or JSON", item)
				}
				input[k] = v
			}
		}
	}
	if stdin, _ := cmd.Flags().GetBool("stdin"); stdin {
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			return err
		}
		if err := json.Unmarshal(data, &input); err != nil {
			return err
		}
	}
	if useSandbox, _ := cmd.Flags().GetBool("sandbox"); useSandbox {
		input["sandbox"] = true
	}
	noBlock, _ := cmd.Flags().GetBool("no-block")
	timeoutSec, _ := cmd.Flags().GetInt("timeout")
	if timeoutSec <= 0 {
		timeoutSec = 30
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeoutSec)*time.Second)
	defer cancel()
	started := time.Now()
	out, err := tools.Dispatch(ctx, args[0], input)
	exitCode := 0
	if err != nil {
		exitCode = 1
	}
	if !noBlock {
		metadata := map[string]string{"tool": args[0]}
		if input["sandbox"] == true {
			metadata["sandbox"] = "true"
		}
		b, recErr := blocks.NewStore(blockStorePath).Record(blocks.RecordRequest{
			Type:      blocks.TypeTool,
			Title:     "tool " + args[0],
			Command:   "ti-cli tool run " + strings.Join(args, " "),
			StartedAt: started,
			EndedAt:   time.Now(),
			ExitCode:  exitCode,
			Stdout:    out,
			Metadata:  metadata,
		})
		if recErr == nil && b != nil {
			fmt.Fprintf(os.Stderr, "[block:%s]\n", b.ID)
		}
	}
	if err != nil {
		return err
	}
	fmt.Println(out)
	return nil
}}

func init() {
	rootCmd.AddCommand(toolCmd)
	toolListCmd.Flags().Bool("json", false, "print JSON output")
	toolRunCmd.Flags().Bool("stdin", false, "read JSON input from stdin")
	toolRunCmd.Flags().Int("timeout", 30, "tool timeout in seconds")
	toolRunCmd.Flags().Bool("sandbox", false, "run bash tool calls through Ti sandbox")
	toolRunCmd.Flags().Bool("no-block", false, "do not record this tool run as a Ti block")
	toolCmd.AddCommand(toolListCmd, toolRunCmd)
}
