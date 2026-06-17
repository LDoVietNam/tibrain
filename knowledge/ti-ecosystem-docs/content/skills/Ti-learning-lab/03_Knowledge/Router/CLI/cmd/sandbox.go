package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/ti/cli/internal/blocks"
	"github.com/ti/cli/internal/sandbox"
)

var sandboxPolicyFile string

var sandboxCmd = &cobra.Command{
	Use:   "sandbox",
	Short: "Run commands inside Ti's sandbox runtime",
	Long: `Run shell commands with Ti's sandbox policy.

The default portable backend is cross-platform guardrails: timeout, output cap,
environment allowlist, denylist checks, and audit log. On Linux, Ti can use
bubblewrap automatically when bwrap is installed. Docker is available as an
optional backend for stronger process isolation. Sandbox runs are recorded as
Ti Blocks by default so they can be searched, exported, rerun, or explained.`,
}

var sandboxInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Create the default sandbox policy",
	RunE: func(cmd *cobra.Command, args []string) error {
		overwrite, _ := cmd.Flags().GetBool("force")
		_, created, err := sandbox.EnsurePolicy(sandboxPolicyFile, overwrite)
		if err != nil {
			return err
		}
		path := sandboxPolicyFile
		if path == "" {
			path = sandbox.DefaultPolicyPath()
		}
		if created {
			fmt.Printf("created sandbox policy: %s\n", sandbox.ExpandPath(path))
		} else {
			fmt.Printf("sandbox policy already exists: %s\n", sandbox.ExpandPath(path))
		}
		return nil
	},
}

var sandboxStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show sandbox status and selected backend",
	RunE: func(cmd *cobra.Command, args []string) error {
		p, err := sandbox.LoadPolicy(sandboxPolicyFile)
		if err != nil {
			return err
		}
		ctx := cmd.Context()
		selected := sandbox.ResolveBackend(ctx, p.Backend, p.Fallback)
		path := sandboxPolicyFile
		if path == "" {
			path = sandbox.DefaultPolicyPath()
		}
		fmt.Printf("policy:   %s\n", sandbox.ExpandPath(path))
		fmt.Printf("summary:  %s\n", p.Summary())
		fmt.Printf("selected: %s\n", selected)
		fmt.Printf("available: %s\n", strings.Join(sandbox.AvailableBackends(ctx), ", "))
		return nil
	},
}

var sandboxDoctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Check sandbox backends and platform support",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		p, err := sandbox.LoadPolicy(sandboxPolicyFile)
		if err != nil {
			return err
		}
		fmt.Printf("os/arch:   %s/%s\n", runtime.GOOS, runtime.GOARCH)
		fmt.Printf("policy:    %s\n", p.Summary())
		fmt.Printf("portable:  available\n")
		fmt.Printf("bubblewrap: %s\n", availability(ctx, sandbox.BackendBubblewrap))
		fmt.Printf("docker:     %s\n", availability(ctx, sandbox.BackendDocker))
		fmt.Printf("selected:   %s\n", sandbox.ResolveBackend(ctx, p.Backend, p.Fallback))
		if runtime.GOOS != "linux" {
			fmt.Println("note: bubblewrap is Linux-only; portable backend will be used on this OS")
		}
		return nil
	},
}

var sandboxPolicyCmd = &cobra.Command{
	Use:   "policy",
	Short: "Print the active sandbox policy JSON",
	RunE: func(cmd *cobra.Command, args []string) error {
		p, err := sandbox.LoadPolicy(sandboxPolicyFile)
		if err != nil {
			return err
		}
		b, _ := json.MarshalIndent(p, "", "  ")
		fmt.Println(string(b))
		return nil
	},
}

var sandboxRunCmd = &cobra.Command{
	Use:   "run -- <command>",
	Short: "Run a command through the sandbox",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		p, err := sandbox.LoadPolicy(sandboxPolicyFile)
		if err != nil {
			return err
		}
		backend, _ := cmd.Flags().GetString("backend")
		cwd, _ := cmd.Flags().GetString("cwd")
		timeoutSec, _ := cmd.Flags().GetInt("timeout")
		allowNetwork, _ := cmd.Flags().GetBool("allow-network")
		asJSON, _ := cmd.Flags().GetBool("json")
		noBlock, _ := cmd.Flags().GetBool("no-block")
		command := strings.Join(args, " ")
		ctx := cmd.Context()
		started := time.Now()
		if timeoutSec > 0 {
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(ctx, time.Duration(timeoutSec+5)*time.Second)
			defer cancel()
		}
		res, runErr := sandbox.Run(ctx, sandbox.RunRequest{
			Command:      command,
			CWD:          cwd,
			Backend:      backend,
			Timeout:      time.Duration(timeoutSec) * time.Second,
			AllowNetwork: allowNetwork,
			Policy:       p,
		})
		if !noBlock && res != nil {
			metadata := map[string]string{"backend": res.Backend, "sandbox_id": res.ID}
			if allowNetwork {
				metadata["network"] = "allow"
			}
			b, recErr := blocks.NewStore(blockStorePath).Record(blocks.RecordRequest{
				Type:      blocks.TypeSandbox,
				Command:   command,
				CWD:       res.CWD,
				StartedAt: started,
				EndedAt:   time.Now(),
				ExitCode:  res.ExitCode,
				Stdout:    res.Output,
				Metadata:  metadata,
			})
			if recErr == nil && b != nil {
				fmt.Fprintf(os.Stderr, "[block:%s]\n", b.ID)
			}
		}
		if asJSON {
			b, _ := json.MarshalIndent(res, "", "  ")
			fmt.Println(string(b))
		} else if res != nil {
			if res.Output != "" {
				fmt.Print(res.Output)
				if !strings.HasSuffix(res.Output, "\n") {
					fmt.Println()
				}
			}
			fmt.Fprintf(os.Stderr, "[sandbox:%s id=%s exit=%d duration=%s]\n", res.Backend, res.ID, res.ExitCode, res.Duration.Truncate(time.Millisecond))
		}
		return runErr
	},
}

var sandboxLogsCmd = &cobra.Command{
	Use:   "logs",
	Short: "Show recent sandbox audit events",
	RunE: func(cmd *cobra.Command, args []string) error {
		p, err := sandbox.LoadPolicy(sandboxPolicyFile)
		if err != nil {
			return err
		}
		limit, _ := cmd.Flags().GetInt("limit")
		asJSON, _ := cmd.Flags().GetBool("json")
		events, err := sandbox.ReadAuditTail(p.AuditFile, limit)
		if err != nil {
			return err
		}
		if asJSON {
			b, _ := json.MarshalIndent(events, "", "  ")
			fmt.Println(string(b))
			return nil
		}
		if len(events) == 0 {
			fmt.Println("(no sandbox audit events)")
			return nil
		}
		for _, ev := range events {
			status := "ok"
			if !ev.Allowed || ev.Error != "" || ev.ExitCode != 0 {
				status = "warn"
			}
			fmt.Printf("%s %-5s %-10s exit=%d %s\n", ev.Time.Format(time.RFC3339), status, ev.Backend, ev.ExitCode, ev.Command)
		}
		return nil
	},
}

func availability(ctx context.Context, backend string) string {
	selected := sandbox.ResolveBackend(ctx, backend, "")
	if selected == backend {
		return "available"
	}
	return "not found"
}

func init() {
	sandboxCmd.PersistentFlags().StringVar(&sandboxPolicyFile, "policy", "", "sandbox policy file (default ~/.ti/sandbox/policy.json)")
	sandboxInitCmd.Flags().Bool("force", false, "overwrite existing policy")
	sandboxRunCmd.Flags().String("backend", "", "backend: auto, portable, bubblewrap, docker")
	sandboxRunCmd.Flags().String("cwd", "", "working directory (default current directory)")
	sandboxRunCmd.Flags().Int("timeout", 0, "timeout in seconds (default from policy)")
	sandboxRunCmd.Flags().Bool("allow-network", false, "allow network for this run")
	sandboxRunCmd.Flags().Bool("json", false, "print structured result as JSON")
	sandboxRunCmd.Flags().Bool("no-block", false, "do not record this run as a Ti block")
	sandboxLogsCmd.Flags().Int("limit", 20, "number of audit events to show")
	sandboxLogsCmd.Flags().Bool("json", false, "print JSON")
	sandboxCmd.AddCommand(sandboxInitCmd, sandboxStatusCmd, sandboxDoctorCmd, sandboxPolicyCmd, sandboxRunCmd, sandboxLogsCmd)
	rootCmd.AddCommand(sandboxCmd)
}
