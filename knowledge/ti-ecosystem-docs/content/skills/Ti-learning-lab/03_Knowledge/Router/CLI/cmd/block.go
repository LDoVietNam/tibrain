package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/ti/cli/internal/blocks"
	"github.com/ti/cli/internal/sandbox"
)

var (
	blockStorePath string
	blockLimit     int
	blockJSON      bool
)

var blockCmd = &cobra.Command{
	Use:   "block",
	Short: "Inspect and reuse Ti command/output blocks",
	Long: `Ti Blocks stores command, tool, sandbox, translate, and agent output as structured blocks.
Blocks make long-running CLI and agent work easier to search, export, rerun, and explain.`,
}

func openBlockStore() *blocks.Store { return blocks.NewStore(blockStorePath) }

var blockPathCmd = &cobra.Command{
	Use:   "path",
	Short: "Print the block store path",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println(openBlockStore().Root)
		return nil
	},
}

var blockListCmd = &cobra.Command{
	Use:   "list",
	Short: "List recent blocks",
	RunE: func(cmd *cobra.Command, args []string) error {
		store := openBlockStore()
		items, err := store.List(blockLimit)
		if err != nil {
			return err
		}
		if blockJSON {
			b, _ := json.MarshalIndent(items, "", "  ")
			fmt.Println(string(b))
			return nil
		}
		if len(items) == 0 {
			fmt.Println("No blocks yet.")
			return nil
		}
		fmt.Printf("%-31s  %-10s  %-19s  %-6s  %-8s  %s\n", "ID", "TYPE", "STARTED", "EXIT", "DUR", "COMMAND/TITLE")
		for _, b := range items {
			title := b.Command
			if title == "" {
				title = b.Title
			}
			title = trimOneLine(title, 80)
			fmt.Printf("%-31s  %-10s  %-19s  %-6d  %-8s  %s\n", b.ID, b.Type, b.StartedAt.Format("2006-01-02 15:04:05"), b.ExitCode, time.Duration(b.DurationMS)*time.Millisecond, title)
		}
		return nil
	},
}

var blockShowCmd = &cobra.Command{
	Use:   "show <id|last>",
	Short: "Show block metadata",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		store := openBlockStore()
		b, err := store.Get(args[0])
		if err != nil {
			return err
		}
		data, _ := json.MarshalIndent(b, "", "  ")
		fmt.Println(string(data))
		return nil
	},
}

var blockOutputCmd = &cobra.Command{
	Use:   "output <id|last>",
	Short: "Print block output",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		store := openBlockStore()
		b, err := store.Get(args[0])
		if err != nil {
			return err
		}
		out, err := store.Output(b)
		if err != nil {
			return err
		}
		fmt.Print(out)
		if out != "" && !strings.HasSuffix(out, "\n") {
			fmt.Println()
		}
		return nil
	},
}

var blockExportCmd = &cobra.Command{
	Use:   "export <id|last> [file]",
	Short: "Export a block as Markdown",
	Args:  cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		store := openBlockStore()
		b, err := store.Get(args[0])
		if err != nil {
			return err
		}
		md, err := store.ExportMarkdown(b)
		if err != nil {
			return err
		}
		if len(args) == 2 {
			return os.WriteFile(args[1], []byte(md), 0o600)
		}
		fmt.Print(md)
		return nil
	},
}

var blockExplainCmd = &cobra.Command{
	Use:   "explain <id|last>",
	Short: "Explain a block using deterministic Ti heuristics",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		store := openBlockStore()
		b, err := store.Get(args[0])
		if err != nil {
			return err
		}
		out, err := store.Output(b)
		if err != nil {
			return err
		}
		fmt.Print(blocks.ExplainHeuristic(b, out))
		return nil
	},
}

var blockRerunCmd = &cobra.Command{
	Use:   "rerun <id|last>",
	Short: "Rerun a command block through the sandbox",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		store := openBlockStore()
		b, err := store.Get(args[0])
		if err != nil {
			return err
		}
		if strings.TrimSpace(b.Command) == "" {
			return fmt.Errorf("block %s has no command to rerun", b.ID)
		}
		backend, _ := cmd.Flags().GetString("backend")
		timeoutSec, _ := cmd.Flags().GetInt("timeout")
		policy, err := sandbox.LoadPolicy(sandboxPolicyFile)
		if err != nil {
			return err
		}
		ctx := context.Background()
		started := time.Now()
		res, runErr := sandbox.Run(ctx, sandbox.RunRequest{
			Command: b.Command,
			CWD:     b.CWD,
			Backend: backend,
			Timeout: time.Duration(timeoutSec) * time.Second,
			Policy:  policy,
		})
		ended := time.Now()
		exit := 1
		output := ""
		if res != nil {
			exit = res.ExitCode
			output = res.Output
		}
		nb, recErr := store.Record(blocks.RecordRequest{
			Type:      blocks.TypeSandbox,
			Title:     "rerun " + b.ID,
			Command:   b.Command,
			CWD:       b.CWD,
			StartedAt: started,
			EndedAt:   ended,
			ExitCode:  exit,
			Stdout:    output,
			Metadata: map[string]string{
				"rerun_of": b.ID,
				"backend":  backend,
			},
		})
		if recErr == nil && nb != nil {
			fmt.Fprintf(os.Stderr, "[block:%s]\n", nb.ID)
		}
		if output != "" {
			fmt.Print(output)
			if !strings.HasSuffix(output, "\n") {
				fmt.Println()
			}
		}
		return runErr
	},
}

var blockSearchCmd = &cobra.Command{
	Use:   "search <query>",
	Short: "Search block metadata and output",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		items, err := openBlockStore().Search(args[0], blockLimit)
		if err != nil {
			return err
		}
		if blockJSON {
			data, _ := json.MarshalIndent(items, "", "  ")
			fmt.Println(string(data))
			return nil
		}
		if len(items) == 0 {
			fmt.Println("No matching blocks.")
			return nil
		}
		fmt.Printf("%-31s  %-10s  %-6s  %s\n", "ID", "TYPE", "EXIT", "COMMAND/TITLE")
		for _, b := range items {
			title := b.Command
			if title == "" {
				title = b.Title
			}
			fmt.Printf("%-31s  %-10s  %-6d  %s\n", b.ID, b.Type, b.ExitCode, trimOneLine(title, 90))
		}
		return nil
	},
}

var blockStatsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Show block store stats",
	RunE: func(cmd *cobra.Command, args []string) error {
		stats, err := openBlockStore().Stats()
		if err != nil {
			return err
		}
		data, _ := json.MarshalIndent(stats, "", "  ")
		fmt.Println(string(data))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(blockCmd)
	blockCmd.PersistentFlags().StringVar(&blockStorePath, "store", "", "block store root (default ~/.ti/data/blocks)")
	blockCmd.PersistentFlags().IntVar(&blockLimit, "limit", 20, "maximum blocks to list")
	blockCmd.PersistentFlags().BoolVar(&blockJSON, "json", false, "print JSON output")
	blockRerunCmd.Flags().String("backend", "", "sandbox backend: auto, portable, bubblewrap, docker")
	blockRerunCmd.Flags().Int("timeout", 0, "timeout in seconds (default from sandbox policy)")
	blockCmd.AddCommand(blockPathCmd, blockListCmd, blockShowCmd, blockOutputCmd, blockRerunCmd, blockExportCmd, blockExplainCmd, blockSearchCmd, blockStatsCmd)
}
