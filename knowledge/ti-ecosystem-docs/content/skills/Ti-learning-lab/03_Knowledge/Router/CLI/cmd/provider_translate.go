package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/ti/cli/internal/providertranslate"
)

var (
	providerImportFrom     string
	providerImportOut      string
	providerImportDryRun   bool
	providerImportMaxDepth int
	providerImportJSON     bool
)

var providerImportCmd = &cobra.Command{
	Use:   "import [file-or-dir]",
	Short: "Translate Kilo/OpenCode/generic provider configs into Ti compatible providers",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		target := "."
		if len(args) > 0 && strings.TrimSpace(args[0]) != "" {
			target = args[0]
		}
		info, err := os.Stat(target)
		if err != nil {
			return err
		}
		var res providertranslate.Result
		if info.IsDir() {
			findings, err := providertranslate.Scan(target, providerImportMaxDepth)
			if err != nil {
				return err
			}
			res, err = providertranslate.ConvertPaths(target, findings, providerImportFrom)
			if err != nil {
				return err
			}
		} else {
			res, err = providertranslate.ConvertFile(target, providerImportFrom)
			if err != nil {
				return err
			}
		}
		if providerImportJSON || providerImportDryRun {
			s, err := providertranslate.RenderTiConfigSnippet(res)
			if err != nil {
				return err
			}
			fmt.Print(s)
			return nil
		}
		out := providerImportOut
		if out == "" {
			out = filepath.Join(".ti", "providers.generated.json")
		}
		if err := providertranslate.Write(out, res); err != nil {
			return err
		}
		fmt.Printf("wrote provider config snippet: %s\n", out)
		fmt.Printf("providers: %d\n", len(res.Providers))
		if len(res.Warnings) > 0 {
			fmt.Println("warnings:")
			for _, w := range res.Warnings {
				fmt.Printf("- %s\n", w)
			}
		}
		return nil
	},
}

var providerScanCmd = &cobra.Command{
	Use:   "scan [path]",
	Short: "Find provider configs from Ti, Kilo, OpenCode, and generic OpenAI-compatible tools",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		root := "."
		if len(args) > 0 && strings.TrimSpace(args[0]) != "" {
			root = args[0]
		}
		findings, err := providertranslate.Scan(root, providerImportMaxDepth)
		if err != nil {
			return err
		}
		if providerImportJSON {
			b, _ := json.MarshalIndent(findings, "", "  ")
			fmt.Println(string(b))
			return nil
		}
		fmt.Printf("Detected %d provider config file(s) under %s\n", len(findings), root)
		for _, f := range findings {
			fmt.Printf("- %-10s %-16s %s\n", f.Format, f.Kind, f.Path)
		}
		return nil
	},
}

func init() {
	providerCmd.PersistentFlags().BoolVar(&providerImportJSON, "json", false, "print JSON output")
	providerCmd.PersistentFlags().IntVar(&providerImportMaxDepth, "max-depth", 8, "maximum directory depth to scan")
	providerImportCmd.Flags().StringVar(&providerImportFrom, "from", "auto", "source format: auto, opencode, kilo, generic, ti")
	providerImportCmd.Flags().StringVar(&providerImportOut, "out", "", "output JSON snippet path (default .ti/providers.generated.json)")
	providerImportCmd.Flags().BoolVar(&providerImportDryRun, "dry-run", false, "print converted provider snippet without writing")
	providerCmd.AddCommand(providerScanCmd, providerImportCmd)
}
