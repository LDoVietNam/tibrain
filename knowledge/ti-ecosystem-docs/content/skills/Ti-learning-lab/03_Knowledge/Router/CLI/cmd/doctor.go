package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
	"github.com/ti/cli/internal/config"
	"github.com/ti/cli/internal/providers"
)

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Check local CLI prerequisites and provider configuration",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, report, err := config.LoadAutoWithReport()
		if err != nil {
			return fmt.Errorf("load config: %w", err)
		}

		fmt.Println("Ti CLI Doctor")
		fmt.Println(strings.Repeat("-", 48))
		fmt.Printf("Go runtime: %s (%s/%s)\n", runtime.Version(), runtime.GOOS, runtime.GOARCH)
		fmt.Printf("GOTOOLCHAIN: %s\n", envOrDefault("GOTOOLCHAIN", "auto"))
		checkCommand("go")
		checkOptionalCommand("protoc")
		checkOptionalCommand("golangci-lint")
		fmt.Println()

		fmt.Println("Config sources:")
		for _, src := range report.Sources {
			state := "missing"
			if src.Found && src.Applied {
				state = "applied"
			} else if src.Found {
				state = "found"
			}
			path := src.Path
			if path == "" {
				path = "-"
			}
			fmt.Printf("  %-10s %-8s %s\n", src.Name, state, path)
		}

		if len(report.Issues) > 0 {
			fmt.Println("\nConfig issues:")
			for _, issue := range report.Issues {
				fmt.Printf("  [%s] %s: %s\n", issue.Level, issue.Field, issue.Message)
			}
		}

		fmt.Println("\nProvider bootstrap:")
		reg, issues := providers.BuildRegistryFromConfig(cfg)
		if names := reg.List(); len(names) > 0 {
			for _, name := range names {
				p, _ := reg.Get(name)
				status := "unhealthy"
				if p.IsHealthy() {
					status = "healthy"
				}
				fmt.Printf("  %-14s %s default=%s\n", name, status, p.DefaultModel())
			}
		} else {
			fmt.Println("  no providers registered")
		}
		for _, issue := range issues {
			fmt.Printf("  warning %-8s %s\n", issue.Provider, issue.Message)
		}

		fmt.Println("\nEnvironment hints:")
		hintEnv("TI_ROUTER_URL")
		hintEnv("OPENAI_API_KEY")
		hintEnv("ANTHROPIC_API_KEY")
		hintEnv("OPENROUTER_API_KEY")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(doctorCmd)
}

func checkCommand(name string) {
	if path, err := exec.LookPath(name); err == nil {
		fmt.Printf("  ok       %-14s %s\n", name, path)
		return
	}
	fmt.Printf("  missing  %-14s required for source builds\n", name)
}

func checkOptionalCommand(name string) {
	if path, err := exec.LookPath(name); err == nil {
		fmt.Printf("  ok       %-14s %s\n", name, path)
		return
	}
	fmt.Printf("  optional %-14s install when using lint/proto targets\n", name)
}

func hintEnv(name string) {
	if os.Getenv(name) == "" {
		fmt.Printf("  unset    %s\n", name)
		return
	}
	fmt.Printf("  set      %s\n", name)
}

func envOrDefault(name, fallback string) string {
	if v := os.Getenv(name); v != "" {
		return v
	}
	return fallback
}
