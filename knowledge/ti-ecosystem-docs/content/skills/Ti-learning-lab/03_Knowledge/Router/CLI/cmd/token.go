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
	"github.com/ti/cli/internal/sandbox"
	"github.com/ti/cli/internal/tokenopt"
)

var (
	tokenCommand   string
	tokenMaxLines  int
	tokenMaxBytes  int
	tokenJSON      bool
	tokenCacheRoot string
	tokenUseCache  bool
	tokenNoBlock   bool
)

var tokenCmd = &cobra.Command{
	Use:   "token",
	Short: "Estimate, compress, cache, and proxy CLI output to reduce LLM token usage",
	Long: `Token optimization compresses noisy dev command output before it is passed to an LLM.
It is inspired by CLI output proxy patterns such as rtk, but implemented as a Go-native Ti module with sandbox, blocks, and cache integration.`,
}

var tokenEstimateCmd = &cobra.Command{
	Use:   "estimate [file|-]",
	Short: "Estimate tokens in a file or stdin",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		data, err := readMaybeFile(args)
		if err != nil {
			return err
		}
		res := map[string]any{"bytes": len(data), "estimated_tokens": tokenopt.EstimateTokens(string(data))}
		if tokenJSON {
			b, _ := json.MarshalIndent(res, "", "  ")
			fmt.Println(string(b))
		} else {
			fmt.Printf("bytes: %d\nestimated_tokens: %d\n", len(data), res["estimated_tokens"])
		}
		return nil
	},
}

var tokenCompressCmd = &cobra.Command{
	Use:   "compress [file|-]",
	Short: "Compress CLI output using deterministic dev-log filters",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		data, err := readMaybeFile(args)
		if err != nil {
			return err
		}
		res := tokenopt.Compress(string(data), tokenopt.Options{Command: tokenCommand, MaxLines: tokenMaxLines, MaxBytes: tokenMaxBytes})
		if tokenJSON {
			b, _ := json.MarshalIndent(res, "", "  ")
			fmt.Println(string(b))
			return nil
		}
		fmt.Fprintf(os.Stderr, "[tokenopt kind=%s tokens=%d->%d saved=%d %.1f%%]\n", res.Kind, res.EstimatedTokensIn, res.EstimatedTokensOut, res.SavedTokens, res.SavedPercent)
		fmt.Print(res.Output)
		return nil
	},
}

var tokenRunCmd = &cobra.Command{
	Use:   "run -- <command>",
	Short: "Run a command through sandbox, compress output, cache it, and record a block",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		command := strings.Join(args, " ")
		backend, _ := cmd.Flags().GetString("backend")
		cwd, _ := cmd.Flags().GetString("cwd")
		timeoutSec, _ := cmd.Flags().GetInt("timeout")
		p, err := sandbox.LoadPolicy(sandboxPolicyFile)
		if err != nil {
			return err
		}
		ctx := cmd.Context()
		if timeoutSec > 0 {
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(ctx, time.Duration(timeoutSec+5)*time.Second)
			defer cancel()
		}
		started := time.Now()
		res, runErr := sandbox.Run(ctx, sandbox.RunRequest{Command: command, CWD: cwd, Backend: backend, Timeout: time.Duration(timeoutSec) * time.Second, Policy: p})
		ended := time.Now()
		out := ""
		exit := 1
		if res != nil {
			out = res.Output
			exit = res.ExitCode
		}
		key := tokenopt.Hash(command, out)
		cache := tokenopt.NewCache(tokenCacheRoot)
		if tokenUseCache {
			if entry, cached, ok, err := cache.Get(key); err == nil && ok {
				if tokenJSON {
					b, _ := json.MarshalIndent(map[string]any{"cache_hit": true, "entry": entry, "output": cached}, "", "  ")
					fmt.Println(string(b))
				} else {
					fmt.Fprintf(os.Stderr, "[tokenopt cache-hit saved=%d %.1f%%]\n", entry.SavedTokens, entry.SavedPercent)
					fmt.Print(cached)
				}
				return runErr
			}
		}
		comp := tokenopt.Compress(out, tokenopt.Options{Command: command, MaxLines: tokenMaxLines, MaxBytes: tokenMaxBytes})
		if tokenUseCache {
			_ = cache.Put(key, comp)
		}
		if !tokenNoBlock {
			meta := map[string]string{"token_kind": comp.Kind, "tokens_in": fmt.Sprint(comp.EstimatedTokensIn), "tokens_out": fmt.Sprint(comp.EstimatedTokensOut), "tokens_saved": fmt.Sprint(comp.SavedTokens)}
			if res != nil {
				meta["sandbox_id"] = res.ID
				meta["backend"] = res.Backend
			}
			if b, recErr := blocks.NewStore(blockStorePath).Record(blocks.RecordRequest{Type: blocks.TypeCommand, Title: "token run", Command: command, CWD: cwd, StartedAt: started, EndedAt: ended, ExitCode: exit, Stdout: comp.Output, Metadata: meta}); recErr == nil && b != nil {
				fmt.Fprintf(os.Stderr, "[block:%s]\n", b.ID)
			}
		}
		if tokenJSON {
			b, _ := json.MarshalIndent(comp, "", "  ")
			fmt.Println(string(b))
			return runErr
		}
		fmt.Fprintf(os.Stderr, "[tokenopt kind=%s tokens=%d->%d saved=%d %.1f%%]\n", comp.Kind, comp.EstimatedTokensIn, comp.EstimatedTokensOut, comp.SavedTokens, comp.SavedPercent)
		fmt.Print(comp.Output)
		return runErr
	},
}

var tokenCacheCmd = &cobra.Command{Use: "cache", Short: "Manage token optimization cache"}
var tokenCachePathCmd = &cobra.Command{Use: "path", Short: "Print token cache path", RunE: func(cmd *cobra.Command, args []string) error {
	fmt.Println(tokenopt.NewCache(tokenCacheRoot).Root)
	return nil
}}
var tokenCacheStatsCmd = &cobra.Command{Use: "stats", Short: "Show token cache stats", RunE: func(cmd *cobra.Command, args []string) error {
	stats, err := tokenopt.NewCache(tokenCacheRoot).Stats()
	if err != nil {
		return err
	}
	b, _ := json.MarshalIndent(stats, "", "  ")
	fmt.Println(string(b))
	return nil
}}
var tokenCacheClearCmd = &cobra.Command{Use: "clear", Short: "Clear token cache", RunE: func(cmd *cobra.Command, args []string) error {
	if err := tokenopt.NewCache(tokenCacheRoot).Clear(); err != nil {
		return err
	}
	fmt.Println("token cache cleared")
	return nil
}}

var proxyCmd = &cobra.Command{Use: "proxy", Short: "Compatibility alias for token-saving command proxy"}
var proxyRunCmd = &cobra.Command{Use: "run -- <command>", Short: "Alias for token run", Args: cobra.MinimumNArgs(1), RunE: func(cmd *cobra.Command, args []string) error { return tokenRunCmd.RunE(cmd, args) }}

func readMaybeFile(args []string) ([]byte, error) {
	if len(args) == 0 || args[0] == "-" {
		return io.ReadAll(os.Stdin)
	}
	return os.ReadFile(args[0])
}

func init() {
	tokenCmd.PersistentFlags().BoolVar(&tokenJSON, "json", false, "print JSON output")
	tokenCmd.PersistentFlags().StringVar(&tokenCacheRoot, "cache", "", "token optimization cache root (default ~/.ti/data/tokenopt)")
	tokenCompressCmd.Flags().StringVar(&tokenCommand, "command", "", "source command for better output classification")
	tokenCompressCmd.Flags().IntVar(&tokenMaxLines, "max-lines", 160, "maximum target lines")
	tokenCompressCmd.Flags().IntVar(&tokenMaxBytes, "max-bytes", 24000, "maximum output bytes")
	tokenRunCmd.Flags().IntVar(&tokenMaxLines, "max-lines", 160, "maximum target lines")
	tokenRunCmd.Flags().IntVar(&tokenMaxBytes, "max-bytes", 24000, "maximum output bytes")
	tokenRunCmd.Flags().BoolVar(&tokenUseCache, "use-cache", true, "reuse and write compressed output cache")
	tokenRunCmd.Flags().BoolVar(&tokenNoBlock, "no-block", false, "do not record a Ti block")
	tokenRunCmd.Flags().String("backend", "", "sandbox backend: auto, portable, bubblewrap, docker")
	tokenRunCmd.Flags().String("cwd", "", "working directory")
	tokenRunCmd.Flags().Int("timeout", 0, "timeout seconds (default from sandbox policy)")
	tokenCacheCmd.AddCommand(tokenCachePathCmd, tokenCacheStatsCmd, tokenCacheClearCmd)
	tokenCmd.AddCommand(tokenEstimateCmd, tokenCompressCmd, tokenRunCmd, tokenCacheCmd)
	rootCmd.AddCommand(tokenCmd)

	proxyRunCmd.Flags().IntVar(&tokenMaxLines, "max-lines", 160, "maximum target lines")
	proxyRunCmd.Flags().IntVar(&tokenMaxBytes, "max-bytes", 24000, "maximum output bytes")
	proxyRunCmd.Flags().BoolVar(&tokenUseCache, "use-cache", true, "reuse and write compressed output cache")
	proxyRunCmd.Flags().BoolVar(&tokenNoBlock, "no-block", false, "do not record a Ti block")
	proxyRunCmd.Flags().String("backend", "", "sandbox backend: auto, portable, bubblewrap, docker")
	proxyRunCmd.Flags().String("cwd", "", "working directory")
	proxyRunCmd.Flags().Int("timeout", 0, "timeout seconds (default from sandbox policy)")
	proxyCmd.AddCommand(proxyRunCmd)
	rootCmd.AddCommand(proxyCmd)
}
