package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/ti/cli/internal/blocks"
	"github.com/ti/cli/internal/translate"
)

var (
	translateFrom     string
	translateTo       string
	translateOut      string
	translatePackName string
	translateJSON     bool
	translateDryRun   bool
	translateMaxDepth int
	translateNoBlock  bool
	translateStrict   bool
)

var translateCmd = &cobra.Command{
	Use:   "translate",
	Short: "Convert skills, workflows, commands, providers, and rules from other AI CLIs",
	Long: `Ti Translate Engine scans formats from Codex, Claude Code, OpenCode,
Cursor, Windsurf, MCP, Aider, Roo, Devin, Kilo, and generic AI CLI configs.
It normalizes them into Ti IR, emits ti.pack.yaml, and can export back to a
small set of external formats. Executable hooks/plugins are marked sandboxed
and high-risk instead of being enabled directly.`,
}

var translateFormatsCmd = &cobra.Command{
	Use:   "formats",
	Short: "List source and target formats supported by Ti Translate",
	RunE: func(cmd *cobra.Command, args []string) error {
		formats := map[string]any{
			"sources": []string{"codex", "claude-code", "opencode", "cursor", "windsurf", "mcp", "aider", "roo", "devin", "kilo", "generic"},
			"targets": []string{"ti-pack", "claude-code", "codex", "opencode"},
			"objects": []string{"instructions", "commands", "agents", "skills", "workflows", "hooks", "plugins", "mcpServers", "providers", "permissions", "conflicts", "risks"},
		}
		if translateJSON {
			return printJSON(formats)
		}
		fmt.Println("Source formats: codex, claude-code, opencode, cursor, windsurf, mcp, aider, roo, devin, kilo, generic")
		fmt.Println("Target formats: ti-pack, claude-code, codex, opencode")
		fmt.Println("Objects: instructions, commands, agents, skills, workflows, hooks, plugins, mcpServers, providers, permissions, conflicts, risks")
		return nil
	},
}

var translateScanCmd = &cobra.Command{
	Use:   "scan [path]",
	Short: "Detect importable AI CLI files",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		root := argPath(args)
		items, err := translate.Scan(root, translate.ScanOptions{From: translateFrom, MaxDepth: translateMaxDepth})
		if err != nil {
			return err
		}
		if translateJSON {
			return printJSON(items)
		}
		fmt.Printf("Detected %d importable file(s) under %s\n", len(items), root)
		for _, item := range items {
			fmt.Printf("- %-12s %-18s %-12s %s\n", item.Format, item.Kind, item.Name, item.Path)
			for _, note := range item.Notes {
				fmt.Printf("  note: %s\n", note)
			}
			for _, risk := range item.Risks {
				fmt.Printf("  risk: %s\n", risk)
			}
		}
		return nil
	},
}

var translateExplainCmd = &cobra.Command{
	Use:   "explain [path]",
	Short: "Show what Ti would create before conversion",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		root := argPath(args)
		items, err := translate.Scan(root, translate.ScanOptions{From: translateFrom, MaxDepth: translateMaxDepth})
		if err != nil {
			return err
		}
		pack, _, err := translate.Convert(root, items, packName(root))
		if err != nil {
			return err
		}
		if translateJSON {
			return printJSON(pack)
		}
		printExplain(root, items, pack)
		return nil
	},
}

var translateConvertCmd = &cobra.Command{
	Use:   "convert [path]",
	Short: "Convert detected formats into Ti pack or export target",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		root := argPath(args)
		started := time.Now()
		items, err := translate.Scan(root, translate.ScanOptions{From: translateFrom, MaxDepth: translateMaxDepth})
		if err != nil {
			return err
		}
		pack, lock, err := translate.Convert(root, items, packName(root))
		if err != nil {
			return err
		}
		out := translateOut
		var result string
		if translateTo == "" || translateTo == translate.FormatTiPack {
			if out == "" {
				out = filepath.Join(root, ".ti", "packs", pack.Metadata.Name)
			}
			result, err = translate.WritePack(pack, lock, translate.WriteOptions{OutDir: out, DryRun: translateDryRun})
		} else {
			if out == "" {
				out = root
			}
			result, err = translate.ExportPack(pack, translate.ExportOptions{To: translateTo, OutDir: out, DryRun: translateDryRun})
		}
		if err != nil {
			return err
		}
		if !translateNoBlock {
			metadata := map[string]string{"from": translateFrom, "to": translateTo, "pack": pack.Metadata.Name}
			if translateDryRun {
				metadata["dry_run"] = "true"
			}
			b, recErr := blocks.NewStore(blockStorePath).Record(blocks.RecordRequest{
				Type:      blocks.TypeTranslate,
				Title:     "translate " + root,
				Command:   "ti-cli translate convert " + root,
				StartedAt: started,
				EndedAt:   time.Now(),
				ExitCode:  0,
				Stdout:    result,
				Metadata:  metadata,
			})
			if recErr == nil && b != nil {
				fmt.Fprintf(os.Stderr, "[block:%s]\n", b.ID)
			}
		}
		if translateDryRun {
			fmt.Print(result)
			return nil
		}
		if translateTo == "" || translateTo == translate.FormatTiPack {
			fmt.Printf("Wrote Ti pack: %s\n", result)
			fmt.Printf("Wrote lockfile: %s\n", filepath.Join(out, "compat.lock.json"))
		} else {
			fmt.Printf("Exported %s files:\n%s\n", translateTo, result)
		}
		fmt.Printf("Converted: %d instruction(s), %d command(s), %d agent(s), %d skill(s), %d workflow(s), %d hook(s), %d plugin(s), %d MCP server(s), %d provider(s), %d conflict(s), %d risk(s)\n",
			len(pack.Instructions), len(pack.Commands), len(pack.Agents), len(pack.Skills), len(pack.Workflows), len(pack.Hooks), len(pack.Plugins), len(pack.MCPServers), len(pack.Providers), len(pack.Conflicts), len(pack.Risks))
		if len(pack.Risks) > 0 {
			fmt.Println("Review risks before enabling executable workflows or plugins. Use `ti sandbox init` first.")
		}
		if translateStrict && (len(pack.Conflicts) > 0 || len(pack.Risks) > 0) {
			return fmt.Errorf("strict mode: conflicts or risks detected")
		}
		return nil
	},
}

var translateValidateCmd = &cobra.Command{
	Use:   "validate [pack-file]",
	Short: "Validate a generated Ti pack file",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		path := "ti.pack.yaml"
		if len(args) > 0 {
			path = args[0]
		}
		info, err := os.Stat(path)
		if err == nil && info.IsDir() {
			path = filepath.Join(path, "ti.pack.yaml")
		}
		issues := translate.ValidatePackFile(path)
		if translateJSON {
			return printJSON(map[string]any{"path": path, "issues": issues, "ok": len(issues) == 0})
		}
		if len(issues) == 0 {
			fmt.Printf("OK: %s\n", path)
			return nil
		}
		fmt.Printf("Issues in %s:\n", path)
		for _, issue := range issues {
			fmt.Printf("- %s\n", issue)
		}
		return fmt.Errorf("validation failed")
	},
}

var translateConflictsCmd = &cobra.Command{
	Use:   "conflicts [path]",
	Short: "Detect conflicting rules after translation",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		root := argPath(args)
		items, err := translate.Scan(root, translate.ScanOptions{From: translateFrom, MaxDepth: translateMaxDepth})
		if err != nil {
			return err
		}
		pack, _, err := translate.Convert(root, items, packName(root))
		if err != nil {
			return err
		}
		if translateJSON {
			return printJSON(pack.Conflicts)
		}
		if len(pack.Conflicts) == 0 {
			fmt.Println("No conflicts detected")
			return nil
		}
		for _, c := range pack.Conflicts {
			fmt.Printf("- [%s] %s: %s\n", c.Level, c.Topic, c.Message)
			for _, src := range c.Sources {
				fmt.Printf("  source: %s (%s)\n", src.Path, src.Format)
			}
		}
		if translateStrict {
			return fmt.Errorf("conflicts detected")
		}
		return nil
	},
}

var translateExportCmd = &cobra.Command{
	Use:   "export [path]",
	Short: "Translate current project into another CLI format",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		root := argPath(args)
		items, err := translate.Scan(root, translate.ScanOptions{From: translateFrom, MaxDepth: translateMaxDepth})
		if err != nil {
			return err
		}
		pack, _, err := translate.Convert(root, items, packName(root))
		if err != nil {
			return err
		}
		out := translateOut
		if out == "" {
			out = root
		}
		res, err := translate.ExportPack(pack, translate.ExportOptions{To: translateTo, OutDir: out, DryRun: translateDryRun})
		if err != nil {
			return err
		}
		fmt.Print(res)
		if !strings.HasSuffix(res, "\n") {
			fmt.Println()
		}
		return nil
	},
}

var translateOptimizeCmd = &cobra.Command{
	Use:   "optimize [path]",
	Short: "Convert, normalize, deduplicate, and print optimized ti.pack.yaml",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		root := argPath(args)
		items, err := translate.Scan(root, translate.ScanOptions{From: translateFrom, MaxDepth: translateMaxDepth})
		if err != nil {
			return err
		}
		pack, _, err := translate.Convert(root, items, packName(root))
		if err != nil {
			return err
		}
		yaml := translate.RenderYAML(pack)
		issues := translate.ValidatePackText(yaml)
		if translateJSON {
			return printJSON(map[string]any{"pack": pack, "issues": issues})
		}
		fmt.Print(yaml)
		if len(issues) > 0 {
			fmt.Fprintln(os.Stderr, "validation issues:")
			for _, issue := range issues {
				fmt.Fprintln(os.Stderr, "- "+issue)
			}
			if translateStrict {
				return fmt.Errorf("validation failed")
			}
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(translateCmd)
	translateCmd.PersistentFlags().StringVar(&translateFrom, "from", translate.FormatAuto, "source format: auto, codex, claude-code, opencode, cursor, windsurf, mcp, aider, roo, devin, kilo, generic")
	translateCmd.PersistentFlags().IntVar(&translateMaxDepth, "max-depth", 10, "maximum directory depth to scan")
	translateCmd.PersistentFlags().BoolVar(&translateJSON, "json", false, "print JSON output")
	translateCmd.PersistentFlags().BoolVar(&translateStrict, "strict", false, "return non-zero when conflicts or validation issues are detected")
	translateConvertCmd.Flags().StringVar(&translateTo, "to", translate.FormatTiPack, "output format: ti-pack, claude-code, codex, opencode")
	translateConvertCmd.Flags().StringVar(&translateOut, "out", "", "output directory for generated pack/export")
	translateConvertCmd.Flags().StringVar(&translatePackName, "name", "", "generated Ti pack name")
	translateConvertCmd.Flags().BoolVar(&translateDryRun, "dry-run", false, "print generated output without writing files")
	translateConvertCmd.Flags().BoolVar(&translateNoBlock, "no-block", false, "do not record this conversion as a Ti block")
	translateExportCmd.Flags().StringVar(&translateTo, "to", translate.FormatClaudeCode, "target format: claude-code, codex, opencode")
	translateExportCmd.Flags().StringVar(&translateOut, "out", "", "output directory")
	translateExportCmd.Flags().BoolVar(&translateDryRun, "dry-run", false, "show files that would be written")
	translateCmd.AddCommand(translateFormatsCmd)
	translateCmd.AddCommand(translateScanCmd)
	translateCmd.AddCommand(translateExplainCmd)
	translateCmd.AddCommand(translateConvertCmd)
	translateCmd.AddCommand(translateValidateCmd)
	translateCmd.AddCommand(translateConflictsCmd)
	translateCmd.AddCommand(translateExportCmd)
	translateCmd.AddCommand(translateOptimizeCmd)
}

func argPath(args []string) string {
	if len(args) > 0 && strings.TrimSpace(args[0]) != "" {
		return args[0]
	}
	return "."
}

func packName(root string) string {
	if strings.TrimSpace(translatePackName) != "" {
		return translatePackName
	}
	abs, err := filepath.Abs(root)
	if err != nil {
		return "translated-pack"
	}
	base := filepath.Base(abs)
	if base == "." || base == string(filepath.Separator) || base == "" {
		return "translated-pack"
	}
	return "translated-" + base
}

func printJSON(v any) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

func printExplain(root string, items []translate.Finding, pack *translate.Pack) {
	fmt.Printf("Ti Translate plan for %s\n", root)
	fmt.Printf("Detected: %d file(s)\n", len(items))
	counts := map[string]int{}
	for _, item := range items {
		counts[item.Format+":"+item.Kind]++
	}
	for key, n := range counts {
		fmt.Printf("- %s: %d\n", key, n)
	}
	fmt.Println("\nWould create:")
	fmt.Printf("- instructions: %d\n", len(pack.Instructions))
	fmt.Printf("- commands:     %d\n", len(pack.Commands))
	fmt.Printf("- agents:       %d\n", len(pack.Agents))
	fmt.Printf("- skills:       %d\n", len(pack.Skills))
	fmt.Printf("- workflows:    %d\n", len(pack.Workflows))
	fmt.Printf("- hooks:        %d\n", len(pack.Hooks))
	fmt.Printf("- plugins:      %d\n", len(pack.Plugins))
	fmt.Printf("- MCP servers:  %d\n", len(pack.MCPServers))
	fmt.Printf("- providers:    %d\n", len(pack.Providers))
	fmt.Printf("- permissions:  %d\n", len(pack.Permissions))
	fmt.Printf("- conflicts:    %d\n", len(pack.Conflicts))
	fmt.Printf("- risks:        %d\n", len(pack.Risks))
	if len(pack.Conflicts) > 0 {
		fmt.Println("\nConflict summary:")
		for _, c := range pack.Conflicts {
			fmt.Printf("- [%s] %s: %s\n", c.Level, c.Topic, c.Message)
		}
	}
	if len(pack.Risks) > 0 {
		fmt.Println("\nRisk summary:")
		for _, risk := range pack.Risks {
			fmt.Printf("- [%s] %s (%s)\n", risk.Level, risk.Reason, risk.Source.Path)
		}
	}
	fmt.Println("\nNext:")
	fmt.Println("  ti translate convert . --dry-run")
	fmt.Println("  ti translate convert . --out .ti/packs/imported")
	fmt.Println("  ti translate validate .ti/packs/imported")
}
