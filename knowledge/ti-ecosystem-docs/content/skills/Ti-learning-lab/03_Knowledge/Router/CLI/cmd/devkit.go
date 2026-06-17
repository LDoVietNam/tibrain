package cmd

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/ti/cli/internal/devkitassets"
)

var (
	devkitDocsDir   string
	devkitAll       bool
	devkitPhasesCSV string
	devkitOverwrite bool
	devkitJSON      bool
)

var devkitCmd = &cobra.Command{
	Use:   "devkit",
	Short: "AI DevKit workflow templates and checks",
	Long:  "Install and inspect AI DevKit-style phase templates, reusable commands, and project readiness checks.",
}

var devkitInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize AI DevKit files in the current project",
	RunE: func(cmd *cobra.Command, args []string) error {
		phases := selectedDevkitPhases()
		if !devkitAll && devkitPhasesCSV == "" {
			phases = []string{"requirements", "design", "planning", "implementation", "testing"}
		}
		if err := os.MkdirAll(devkitDocsDir, 0755); err != nil {
			return err
		}
		created := []string{}
		for _, phase := range phases {
			path, err := copyDevkitPhase(phase, devkitDocsDir, devkitOverwrite)
			if err != nil {
				return err
			}
			created = append(created, path)
		}
		cfg := map[string]any{
			"version":    "go123",
			"docs_dir":   devkitDocsDir,
			"phases":     phases,
			"created_at": time.Now().Format(time.RFC3339),
			"source":     "ti-cli devkit",
		}
		data, _ := json.MarshalIndent(cfg, "", "  ")
		if err := writeFileIfAllowed(".ai-devkit.json", data, devkitOverwrite); err != nil {
			return err
		}
		fmt.Printf("AI DevKit initialized in %s\n", devkitDocsDir)
		for _, item := range created {
			fmt.Printf("  - %s\n", item)
		}
		fmt.Println("Config: .ai-devkit.json")
		return nil
	},
}

var devkitPhaseCmd = &cobra.Command{
	Use:   "phase [name]",
	Short: "List, show, or add phase templates",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return listDevkitFiles("phases")
		}
		path, err := copyDevkitPhase(args[0], devkitDocsDir, devkitOverwrite)
		if err != nil {
			return err
		}
		fmt.Printf("Created %s\n", path)
		return nil
	},
}

var devkitPhaseShowCmd = &cobra.Command{
	Use:   "show [name]",
	Short: "Print a phase template",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return showDevkitFile("phases", args[0])
	},
}

var devkitCommandCmd = &cobra.Command{
	Use:   "command",
	Short: "List or show reusable AI command prompts",
	RunE: func(cmd *cobra.Command, args []string) error {
		return listDevkitFiles("commands")
	},
}

var devkitCommandShowCmd = &cobra.Command{
	Use:   "show [name]",
	Short: "Print a reusable command prompt",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return showDevkitFile("commands", args[0])
	},
}

var devkitLintCmd = &cobra.Command{
	Use:   "lint",
	Short: "Validate AI DevKit project readiness",
	RunE: func(cmd *cobra.Command, args []string) error {
		result := lintDevkit(devkitDocsDir)
		if devkitJSON {
			data, _ := json.MarshalIndent(result, "", "  ")
			fmt.Println(string(data))
			if !result.OK {
				return fmt.Errorf("devkit lint failed")
			}
			return nil
		}
		for _, item := range result.Checks {
			mark := "✅"
			if !item.OK {
				mark = "❌"
			}
			fmt.Printf("%s %s - %s\n", mark, item.Name, item.Message)
		}
		if !result.OK {
			return fmt.Errorf("devkit lint failed")
		}
		return nil
	},
}

type devkitLintCheck struct {
	Name    string `json:"name"`
	OK      bool   `json:"ok"`
	Message string `json:"message"`
}

type devkitLintResult struct {
	OK     bool              `json:"ok"`
	Checks []devkitLintCheck `json:"checks"`
}

func selectedDevkitPhases() []string {
	if devkitAll {
		return []string{"requirements", "design", "planning", "implementation", "testing", "deployment", "monitoring"}
	}
	if devkitPhasesCSV == "" {
		return nil
	}
	parts := strings.Split(devkitPhasesCSV, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			out = append(out, part)
		}
	}
	return out
}

func copyDevkitPhase(name, docsDir string, overwrite bool) (string, error) {
	name = normalizeTemplateName(name)
	data, err := readDevkitTemplate("phases", name)
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(docsDir, 0755); err != nil {
		return "", err
	}
	dst := filepath.Join(docsDir, name+".md")
	return dst, writeFileIfAllowed(dst, data, overwrite)
}

func listDevkitFiles(kind string) error {
	dir, err := resolveDevkitDir(kind)
	if err != nil {
		names, assetErr := devkitassets.List(kind)
		if assetErr != nil {
			return err
		}
		for _, name := range names {
			fmt.Println(name)
		}
		return nil
	}
	var names []string
	err = filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(d.Name(), ".md") {
			return err
		}
		names = append(names, strings.TrimSuffix(d.Name(), ".md"))
		return nil
	})
	if err != nil {
		return err
	}
	sort.Strings(names)
	for _, name := range names {
		fmt.Println(name)
	}
	return nil
}

func showDevkitFile(kind, name string) error {
	data, err := readDevkitTemplate(kind, name)
	if err != nil {
		return err
	}
	fmt.Print(string(data))
	return nil
}

func readDevkitTemplate(kind, name string) ([]byte, error) {
	path, err := resolveDevkitTemplate(kind, name)
	if err == nil {
		return os.ReadFile(path)
	}
	data, assetErr := devkitassets.Read(kind, name)
	if assetErr == nil {
		return data, nil
	}
	return nil, err
}

func resolveDevkitTemplate(kind, name string) (string, error) {
	name = normalizeTemplateName(name)
	dir, err := resolveDevkitDir(kind)
	if err != nil {
		return "", err
	}
	path := filepath.Join(dir, name+".md")
	if _, err := os.Stat(path); err != nil {
		return "", fmt.Errorf("%s template %q not found", kind, name)
	}
	return path, nil
}

func resolveDevkitDir(kind string) (string, error) {
	candidates := []string{}
	if env := os.Getenv("TI_DEVKIT_TEMPLATE_DIR"); env != "" {
		candidates = append(candidates, filepath.Join(env, kind))
	}
	candidates = append(candidates, filepath.Join("templates", "devkit", kind))
	if exe, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exe)
		candidates = append(candidates,
			filepath.Join(exeDir, "templates", "devkit", kind),
			filepath.Join(exeDir, "..", "templates", "devkit", kind),
		)
	}
	for _, candidate := range candidates {
		if st, err := os.Stat(candidate); err == nil && st.IsDir() {
			return candidate, nil
		}
	}
	return "", fmt.Errorf("devkit templates/%s not found; keep templates/devkit next to the source or set TI_DEVKIT_TEMPLATE_DIR", kind)
}

func writeFileIfAllowed(path string, data []byte, overwrite bool) error {
	if _, err := os.Stat(path); err == nil && !overwrite {
		return fmt.Errorf("%s already exists; pass --overwrite to replace it", path)
	}
	return os.WriteFile(path, data, 0644)
}

func normalizeTemplateName(name string) string {
	name = strings.TrimSpace(strings.ToLower(name))
	name = strings.TrimSuffix(name, ".md")
	return name
}

func lintDevkit(docsDir string) devkitLintResult {
	checks := []devkitLintCheck{}
	add := func(name string, ok bool, msg string) {
		checks = append(checks, devkitLintCheck{Name: name, OK: ok, Message: msg})
	}
	if _, err := os.Stat(".ai-devkit.json"); err == nil {
		add("config", true, ".ai-devkit.json found")
	} else {
		add("config", false, ".ai-devkit.json missing; run ti-cli devkit init")
	}
	if st, err := os.Stat(docsDir); err == nil && st.IsDir() {
		add("docs_dir", true, docsDir+" found")
	} else {
		add("docs_dir", false, docsDir+" missing")
	}
	for _, phase := range []string{"requirements", "design", "planning", "implementation", "testing"} {
		path := filepath.Join(docsDir, phase+".md")
		if info, err := os.Stat(path); err == nil && info.Size() > 0 {
			add("phase:"+phase, true, path+" ready")
		} else {
			add("phase:"+phase, false, path+" missing or empty")
		}
	}
	ok := true
	for _, check := range checks {
		if !check.OK {
			ok = false
			break
		}
	}
	return devkitLintResult{OK: ok, Checks: checks}
}

func init() {
	rootCmd.AddCommand(devkitCmd)
	devkitInitCmd.Flags().BoolVar(&devkitAll, "all", false, "initialize all phase templates")
	devkitInitCmd.Flags().StringVar(&devkitPhasesCSV, "phases", "", "comma-separated phases to initialize")
	devkitInitCmd.Flags().StringVar(&devkitDocsDir, "docs-dir", "docs/ai", "AI documentation directory")
	devkitInitCmd.Flags().BoolVar(&devkitOverwrite, "overwrite", false, "overwrite existing files")
	devkitPhaseCmd.Flags().StringVar(&devkitDocsDir, "docs-dir", "docs/ai", "AI documentation directory")
	devkitPhaseCmd.Flags().BoolVar(&devkitOverwrite, "overwrite", false, "overwrite existing phase file")
	devkitLintCmd.Flags().StringVar(&devkitDocsDir, "docs-dir", "docs/ai", "AI documentation directory")
	devkitLintCmd.Flags().BoolVar(&devkitJSON, "json", false, "output JSON")

	devkitCmd.AddCommand(devkitInitCmd)
	devkitCmd.AddCommand(devkitPhaseCmd)
	devkitPhaseCmd.AddCommand(devkitPhaseShowCmd)
	devkitCmd.AddCommand(devkitCommandCmd)
	devkitCommandCmd.AddCommand(devkitCommandShowCmd)
	devkitCmd.AddCommand(devkitLintCmd)
}
