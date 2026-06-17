package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/ti/cli/internal/optimizepack"
)

var optimizeJSON bool

var optimizeCmd = &cobra.Command{
	Use:   "optimize",
	Short: "Optimization diagnostics and release guidance for Ti CLI",
}

var optimizeDoctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Check local tooling for max-optimized builds",
	RunE: func(cmd *cobra.Command, args []string) error {
		checks := optimizepack.Checks()
		if optimizeJSON {
			b, _ := json.MarshalIndent(checks, "", "  ")
			fmt.Println(string(b))
			return nil
		}
		fmt.Print(optimizepack.Summary(checks))
		return nil
	},
}

var optimizeBuildPlanCmd = &cobra.Command{
	Use:   "build-plan",
	Short: "Print the recommended max-optimized build and smoke-test flow",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Print(optimizepack.BuildPlan())
		return nil
	},
}

func init() {
	optimizeCmd.PersistentFlags().BoolVar(&optimizeJSON, "json", false, "print JSON")
	optimizeCmd.AddCommand(optimizeDoctorCmd, optimizeBuildPlanCmd)
	rootCmd.AddCommand(optimizeCmd)
}
