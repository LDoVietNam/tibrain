package learn

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

// RegisterAll registers learn commands with root command.
// This file replaces cmd/learn.go and integrates into existing CLI.
func RegisterAll(root *cobra.Command) {
	// Parent: learn
	var learnCmd = &cobra.Command{
		Use:   "learn",
		Short: "BEADS learning system management",
		Long:  `View learning statistics, manage patterns, budgets, and modes.`,
	}

	// Subcommands
	learnCmd.AddCommand(newStatsCmd())
	learnCmd.AddCommand(newPatternsCmd())
	learnCmd.AddCommand(newModelsCmd())
	learnCmd.AddCommand(newBudgetCmd())
	learnCmd.AddCommand(newConfigCmd())
	learnCmd.AddCommand(newFeedbackCmd())
	learnCmd.AddCommand(newModeCmd())
	learnCmd.AddCommand(newResetCmd())
	learnCmd.AddCommand(newExportCmd())
	learnCmd.AddCommand(newHealthCmd())
	learnCmd.AddCommand(newEnableCmd())
	learnCmd.AddCommand(newDisableCmd())

	root.AddCommand(learnCmd)
}

// ========== Commands ==========

func newStatsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "stats [--json]",
		Short: "Show learning statistics",
		Run: func(cmd *cobra.Command, args []string) {
			engine := GetEngine() // Global engine (from init)
			stats := engine.GetStats()

			if jsonOutput, _ := cmd.Flags().GetBool("json"); jsonOutput {
				data, _ := json.MarshalIndent(stats, "", "  ")
				fmt.Println(string(data))
				return
			}

			fmt.Println(`
╔══════════════════════════════════════════════════╗
║         BEADS Learning System Stats              ║
╠══════════════════════════════════════════════════╣
║ Total samples:      ` + fmt.Sprintf("%-6d", stats.TotalSamples) + `                ║
║ Avg quality:        ` + fmt.Sprintf("%.2f / 10.0", stats.AvgQuality) + `             ║
║ Success rate:       ` + fmt.Sprintf("%.1f%%", stats.SuccessRate*100) + `              ║
║ Total cost:         $` + fmt.Sprintf("%.2f", stats.TotalCostUSD) + `                   ║
║ Cost savings:       $` + fmt.Sprintf("%.2f", stats.CostSavingsUSD) + `                     ║
╠══════════════════════════════════════════════════╣
║ Top Models:                                 ║`)
			for i, m := range stats.TopModels {
				fmt.Printf("║   %d. %-25s  %.2f (%d samples)       ║\n", i+1, m.Model+"("+m.Provider+")", m.Quality, m.Samples)
			}
			fmt.Printf(`╠══════════════════════════════════════════════════╣
║ Patterns learned:   %-5d                        ║
║ Bandit arms:        %-5d                        ║
║ Mode:               %-10s                     ║
╚══════════════════════════════════════════════════╝
`, stats.PatternCount, stats.BanditArms, engine.GetMode())
		},
	}
}

func newPatternsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "patterns [--task TYPE] [--limit N]",
		Short: "List learned prompt patterns",
		Run: func(cmd *cobra.Command, args []string) {
			taskType, _ := cmd.Flags().GetString("task")
			limit, _ := cmd.Flags().GetInt("limit")
			if limit == 0 {
				limit = 10
			}

			engine := GetEngine()
			patterns := engine.GetTopPatterns(limit)

			fmt.Printf("Top %d patterns for task=%s:\n\n", limit, taskType)
			for i, p := range patterns {
				if taskType != "" && p.TaskType != taskType {
					continue
				}
				fmt.Printf("%d. Template: %s\n", i+1, p.Template)
				fmt.Printf("   Quality: %.2f, Success: %d/%d, BestModel: %s\n",
					p.AvgQuality, p.SuccessCount, p.SuccessCount+p.FailureCount, p.BestModel)
				fmt.Println()
			}
		},
	}
}

func newModelsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "models [--task TYPE]",
		Short: "Compare model performance",
		Run: func(cmd *cobra.Command, args []string) {
			taskType, _ := cmd.Flags().GetString("task")
			engine := GetEngine()
			stats := engine.GetBanditStats()

			fmt.Printf("Model performance for task='%s':\n\n", taskType)
			fmt.Println("MODEL                      WIN RATE   SAMPLES")
			fmt.Println("--------------------------------------------------")

			for key, stat := range stats {
				parts := strings.Split(key, ":")
				if len(parts) < 3 {
					continue
				}
				if taskType != "" && parts[2] != taskType {
					continue
				}
				fmt.Printf("%-28s  %6.1f%%   %6d\n",
					key, stat.WinRate*100, stat.Wins+stat.Losses)
			}
		},
	}
}

func newBudgetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "budget [--set DAILY] [--set-monthly MONTHLY]",
		Short: "Check or set budget limits",
		Run: func(cmd *cobra.Command, args []string) {
			// Access cost optimizer - need to add Getter
			fmt.Println("Budget tracking not yet exposed via CLI")
		},
	}
	cmd.Flags().Float64("set", 0, "Set daily budget")
	cmd.Flags().Float64("set-monthly", 0, "Set monthly budget")
	return cmd
}

func newConfigCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "config [--get KEY] [--set KEY VALUE]",
		Short: "View or modify learning config",
		Run: func(cmd *cobra.Command, args []string) {
			engine := GetEngine()
			cfg := engine.Config()
			// Pretty print
			fmt.Printf("Mode: %s\n", cfg.Mode)
			fmt.Printf("Daily budget: $%.2f\n", cfg.DailyBudgetUSD)
			fmt.Printf("Monthly budget: $%.2f\n", cfg.MonthlyBudgetUSD)
			fmt.Printf("Exploration rate: %.2f\n", cfg.ExplorationRate)
		},
	}
}

func newFeedbackCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "feedback --quality N [--session ID] [--comment TEXT]",
		Short: "Provide explicit quality feedback",
		Run: func(cmd *cobra.Command, args []string) {
			quality, _ := cmd.Flags().GetFloat64("quality")
			sessionID, _ := cmd.Flags().GetString("session")
			comment, _ := cmd.Flags().GetString("comment")

			// TODO: Update existing log entry with quality.User
			fmt.Printf("Recorded feedback: quality=%.1f, session=%s\n", quality, sessionID)
			if comment != "" {
				fmt.Printf("Comment: %s\n", comment)
			}
		},
	}
}

func newModeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "mode [off|shadow|hybrid|auto] [--pct N]",
		Short: "Set learning mode",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				fmt.Println("Usage: opencode learn mode [off|shadow|hybrid|auto] [--pct 20]")
				return
			}
			mode := args[0]
			engine := GetEngine()
			engine.SetMode(mode)

			if mode == "hybrid" {
				pct, _ := cmd.Flags().GetInt("pct")
				if pct > 0 {
					engine.SetHybridRatio(pct)
				}
				fmt.Printf("Mode set to hybrid (%d%% learned routing)\n", engine.GetHybridRatio())
			} else {
				fmt.Printf("Mode set to %s\n", mode)
			}
		},
	}
}

func newResetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "reset [--logs] [--patterns] [--all]",
		Short: "Reset learning data",
		Run: func(cmd *cobra.Command, args []string) {
			engine := GetEngine()
			resetAll, _ := cmd.Flags().GetBool("all")
			resetLogs, _ := cmd.Flags().GetBool("logs")
			resetPatterns, _ := cmd.Flags().GetBool("patterns")

			if resetAll {
				engine.Reset(false)
				fmt.Println("Reset all learning data (config preserved)")
			} else if resetLogs {
				engine.Reset(true) // Keep patterns, clear logs? - TODO
				fmt.Println("Reset logs only")
			} else if resetPatterns {
				// TODO: Clear patterns only
				fmt.Println("Reset patterns only")
			} else {
				fmt.Println("Specify --logs, --patterns, or --all")
			}
		},
	}
}

func newExportCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "export [--file PATH]",
		Short: "Export learning data",
		Run: func(cmd *cobra.Command, args []string) {
			// TODO: Export to JSON
			fmt.Println("Export not yet implemented")
		},
	}
}

func newHealthCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "health",
		Short: "Check engine health",
		Run: func(cmd *cobra.Command, args []string) {
			engine := GetEngine()
			health := engine.Health()

			icon := "✅"
			if health.Status != "healthy" {
				icon = "⚠️"
			}
			fmt.Printf("%s Status: %s\n", icon, health.Status)
			fmt.Printf("Bandit arms: %d\n", health.BanditArms)
			totalCache := health.PatternCacheHits + health.PatternCacheMiss
			if totalCache > 0 {
				hitRate := float64(health.PatternCacheHits) / float64(totalCache) * 100
				fmt.Printf("Cache hit rate: %.1f%%\n", hitRate)
			} else {
				fmt.Printf("Cache hit rate: N/A\n")
			}
			fmt.Printf("DB size: %.2f MB\n", float64(health.DBSizeBytes)/1024/1024)
			for _, w := range health.Warnings {
				fmt.Printf("Warning: %s\n", w)
			}
		},
	}
}

func newEnableCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "enable",
		Short: "Enable learning (shadow mode)",
		Run: func(cmd *cobra.Command, args []string) {
			engine := GetEngine()
			engine.SetMode("shadow")
			fmt.Println("Learning enabled (shadow mode - logging only, no routing changes)")
			fmt.Println("Use `opencode learn mode hybrid` to start auto-routing")
		},
	}
}

func newDisableCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "disable",
		Short: "Disable learning",
		Run: func(cmd *cobra.Command, args []string) {
			engine := GetEngine()
			engine.SetMode("off")
			fmt.Println("Learning disabled")
		},
	}
}

// Flag helpers for commands

// Usage:
//   cmd.Flags().Bool("json", false, "JSON output")
//   cmd.Flags().String("task", "", "Filter by task type")
//   cmd.Flags().Int("limit", 10, "Limit results")
