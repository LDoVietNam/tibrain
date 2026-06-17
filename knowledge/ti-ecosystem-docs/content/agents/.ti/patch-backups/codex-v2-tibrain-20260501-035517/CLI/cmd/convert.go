package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	convertpkg "github.com/ti/cli/internal/convert"
)

var convertFrom string
var convertTo string
var convertInput string
var convertOutput string

var convertInspectJSON bool
var convertGoOut string
var convertGoPackage string
var convertGoModule string
var convertGoKind string
var convertGoSDKImport string
var convertGoWithTests bool
var convertGoBuiltIn bool
var convertGoLearn bool
var convertGoBeads bool
var convertGoBuild bool
var convertLearnRoot string

var convertCmd = &cobra.Command{
	Use:   "convert",
	Short: "Convert skills, workflows, agents, plugins, and CLI/IDE configs",
	Long: `Convert Devin, Claude Code, Codex, OpenCode, Kilo, AmpCode, Cursor,
MCP, workflow, skill, agent, and Ti-style config documents.

The powerful path is:
  source -> Ti Pack -> Ti IR -> Go plugin/project -> build/test -> memory/beads learning`,
	RunE: runConvertLegacy,
}

var convertInspectCmd = &cobra.Command{
	Use:   "inspect",
	Short: "Inspect a source pack and show what Ti can convert",
	RunE: func(cmd *cobra.Command, args []string) error {
		data, err := readConvertInput(convertInput)
		if err != nil {
			return err
		}
		ir, pack, err := convertpkg.BuildIR(data, convertpkg.Options{From: convertFrom, To: convertTo})
		if err != nil {
			return err
		}
		if convertInspectJSON {
			b, _ := json.MarshalIndent(map[string]any{"ir": ir, "pack": pack}, "", "  ")
			fmt.Println(string(b))
			return nil
		}
		fmt.Printf("Source: %s  fingerprint=%s  bytes=%d\n", ir.Source.Format, ir.Source.Fingerprint, ir.Source.Bytes)
		fmt.Printf("IR:     %s  kind=%s  name=%s\n", ir.ID, ir.Kind, ir.Name)
		fmt.Printf("Items:  commands=%d workflows=%d tools=%d agents=%d skills=%d providers=%d policies=%d\n", len(ir.Commands), len(ir.Workflows), len(ir.Tools), len(ir.Agents), len(ir.Skills), len(ir.Providers), len(ir.Policies))
		fmt.Printf("Brain:  model_role=%s risk=%d complexity=%d\n", ir.Learning.RecommendedModel, ir.Learning.RiskScore, ir.Learning.ComplexityScore)
		if len(ir.Learning.Warnings) > 0 {
			fmt.Println("Warnings:")
			for _, w := range ir.Learning.Warnings {
				fmt.Println("  - " + w)
			}
		}
		fmt.Println("\nNext:")
		fmt.Println("  ti convert ir -i <path>")
		fmt.Println("  ti convert to-go -i <path> --out ./Plugins/ti-plugin-generated --with-tests --learn --beads")
		return nil
	},
}

var convertIRCmd = &cobra.Command{
	Use:   "ir",
	Short: "Render normalized Ti IR for brain/memory/codegen",
	RunE: func(cmd *cobra.Command, args []string) error {
		data, err := readConvertInput(convertInput)
		if err != nil {
			return err
		}
		ir, _, err := convertpkg.BuildIR(data, convertpkg.Options{From: convertFrom, To: convertTo})
		if err != nil {
			return err
		}
		out, err := convertpkg.RenderIR(ir)
		if err != nil {
			return err
		}
		return writeConvertOutput(convertOutput, out)
	},
}

var convertToGoCmd = &cobra.Command{
	Use:   "to-go",
	Short: "Generate Go code from Ti IR as a plugin/project",
	Long: `Generate Go code from any supported CLI/IDE/plugin/workflow source.

Default mode emits a standalone external plugin project with its own go.mod so it
will not break the parent Ti build. Use --built-in when the SDK/pluginapi pack is
already applied and you want a built-in plugin package.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		data, err := readConvertInput(convertInput)
		if err != nil {
			return err
		}
		ir, _, err := convertpkg.BuildIR(data, convertpkg.Options{From: convertFrom, To: "go"})
		if err != nil {
			return err
		}
		if strings.TrimSpace(convertGoOut) == "" {
			return fmt.Errorf("missing --out")
		}
		_, err = convertpkg.WriteGoProject(ir, convertpkg.GoGenOptions{OutDir: convertGoOut, Package: convertGoPackage, Module: convertGoModule, Kind: convertGoKind, WithTests: convertGoWithTests, SDKImport: convertGoSDKImport, Standalone: !convertGoBuiltIn})
		if err != nil {
			return err
		}
		fmt.Printf("Generated Go project: %s\n", convertGoOut)
		fmt.Printf("IR: %s  kind=%s  model_role=%s  risk=%d\n", ir.ID, ir.Kind, ir.Learning.RecommendedModel, ir.Learning.RiskScore)
		buildLog := ""
		buildOK := false
		testOK := false
		if convertGoBuild {
			buildLog, testOK, buildOK = runGeneratedGoChecks(convertGoOut)
			fmt.Println(strings.TrimSpace(buildLog))
		}
		if convertGoLearn {
			root, _ := os.Getwd()
			rec, err := convertpkg.SaveLearning(ir, convertpkg.LearnOptions{Root: root, Beads: convertGoBeads, SourcePath: convertInput, OutDir: convertGoOut, BuildLog: buildLog, TestOK: testOK, BuildOK: buildOK})
			if err != nil {
				return err
			}
			fmt.Printf("Learning saved: .ti/memory/convert/conversions.jsonl (%s)\n", rec.IRID)
		}
		return nil
	},
}

var convertLearnCmd = &cobra.Command{
	Use:   "learn",
	Short: "Inspect convert learning memory",
}

var convertLearnStatsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Show convert learning statistics",
	RunE: func(cmd *cobra.Command, args []string) error {
		stats, err := convertpkg.LearningStats(convertLearnRoot)
		if err != nil {
			return err
		}
		b, _ := json.MarshalIndent(stats, "", "  ")
		fmt.Println(string(b))
		return nil
	},
}

func runConvertLegacy(cmd *cobra.Command, args []string) error {
	data, err := readConvertInput(convertInput)
	if err != nil {
		return err
	}
	pack, err := convertpkg.Convert(data, convertpkg.Options{From: convertFrom, To: convertTo})
	if err != nil {
		return err
	}
	out, err := convertpkg.Render(pack, convertTo)
	if err != nil {
		return err
	}
	return writeConvertOutput(convertOutput, out)
}

func readConvertInput(path string) ([]byte, error) {
	if path == "" || path == "-" {
		return io.ReadAll(os.Stdin)
	}
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		return os.ReadFile(path)
	}
	var b strings.Builder
	b.WriteString("{\n")
	first := true
	err = filepath.WalkDir(path, func(p string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			return nil
		}
		ext := strings.ToLower(filepath.Ext(p))
		if ext != ".json" && ext != ".md" && ext != ".txt" && ext != ".yaml" && ext != ".yml" && ext != ".toml" && ext != ".js" && ext != ".ts" && ext != ".go" && ext != ".py" {
			return nil
		}
		data, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(path, p)
		if !first {
			b.WriteString(",\n")
		}
		first = false
		b.WriteString(fmt.Sprintf("%q: %q", rel, string(data)))
		return nil
	})
	if err != nil {
		return nil, err
	}
	b.WriteString("\n}")
	return []byte(b.String()), nil
}

func writeConvertOutput(path string, data []byte) error {
	if path == "" || path == "-" {
		fmt.Println(string(data))
		return nil
	}
	dir := filepath.Dir(path)
	if dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}
	return os.WriteFile(path, data, 0644)
}

func runGeneratedGoChecks(dir string) (string, bool, bool) {
	var out strings.Builder
	testOK := false
	buildOK := false
	cmd := exec.Command("go", "test", "./...")
	cmd.Dir = dir
	b, err := cmd.CombinedOutput()
	out.WriteString("$ go test ./...\n")
	out.Write(b)
	if err == nil {
		testOK = true
	} else {
		out.WriteString(err.Error() + "\n")
	}
	cmd = exec.Command("go", "build", ".")
	cmd.Dir = dir
	b, err = cmd.CombinedOutput()
	out.WriteString("$ go build .\n")
	out.Write(b)
	if err == nil {
		buildOK = true
	} else {
		out.WriteString(err.Error() + "\n")
	}
	return out.String(), testOK, buildOK
}

func init() {
	convertCmd.Flags().StringVar(&convertFrom, "from", "auto", "source CLI/IDE: auto, devin, claude-code, codex, opencode, kilo, cursor, ampcode, ti")
	convertCmd.Flags().StringVar(&convertTo, "to", "ti", "target format: ti, ir, go, claude-code, codex, opencode, ampcode, devin, free-claude-code")
	convertCmd.Flags().StringVarP(&convertInput, "input", "i", "-", "input file/dir, or - for stdin")
	convertCmd.Flags().StringVarP(&convertOutput, "output", "o", "-", "output file, or - for stdout")

	for _, c := range []*cobra.Command{convertInspectCmd, convertIRCmd, convertToGoCmd} {
		c.Flags().StringVar(&convertFrom, "from", "auto", "source CLI/IDE: auto, devin, claude-code, codex, opencode, kilo, cursor, ampcode, ti")
		c.Flags().StringVarP(&convertInput, "input", "i", "-", "input file/dir, or - for stdin")
	}
	convertInspectCmd.Flags().BoolVar(&convertInspectJSON, "json", false, "print full pack and IR JSON")
	convertIRCmd.Flags().StringVarP(&convertOutput, "output", "o", "-", "output file, or - for stdout")

	convertToGoCmd.Flags().StringVar(&convertGoOut, "out", "", "output directory for generated Go project")
	convertToGoCmd.Flags().StringVar(&convertGoPackage, "package", "", "Go package name")
	convertToGoCmd.Flags().StringVar(&convertGoModule, "module", "", "Go module path for standalone output")
	convertToGoCmd.Flags().StringVar(&convertGoKind, "kind", "", "generated kind: external-plugin, built-in-plugin, workflow-plugin, tool-plugin, composite-plugin")
	convertToGoCmd.Flags().StringVar(&convertGoSDKImport, "sdk-import", "github.com/ti/pluginapi", "pluginapi import path for --built-in mode")
	convertToGoCmd.Flags().BoolVar(&convertGoWithTests, "with-tests", true, "generate tests")
	convertToGoCmd.Flags().BoolVar(&convertGoBuiltIn, "built-in", false, "generate built-in plugin package instead of standalone external project")
	convertToGoCmd.Flags().BoolVar(&convertGoLearn, "learn", false, "write conversion result into .ti/memory/convert")
	convertToGoCmd.Flags().BoolVar(&convertGoBeads, "beads", false, "also log conversion learning through bd/beads when available")
	convertToGoCmd.Flags().BoolVar(&convertGoBuild, "build", false, "run go test and go build in generated output")

	convertLearnCmd.PersistentFlags().StringVar(&convertLearnRoot, "root", ".", "workspace root containing .ti/memory/convert")
	convertLearnCmd.AddCommand(convertLearnStatsCmd)

	convertCmd.AddCommand(convertInspectCmd, convertIRCmd, convertToGoCmd, convertLearnCmd)
	rootCmd.AddCommand(convertCmd)
}
