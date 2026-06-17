package cmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/ti/cli/internal/config"
	"github.com/ti/cli/internal/providers"
)

var (
	promptFile     string
	promptProvider string
	promptModel    string
	promptStream   bool
)

var promptCmd = &cobra.Command{
	Use:   "prompt",
	Short: "Run reusable prompt files through a configured provider",
	Long:  "List, show, and run reusable prompt files. Built-in AI DevKit command prompts are embedded into the binary and also shipped under templates/devkit/commands.",
}

var promptListCmd = &cobra.Command{
	Use:   "list",
	Short: "List built-in prompt templates",
	RunE: func(cmd *cobra.Command, args []string) error {
		return listDevkitFiles("commands")
	},
}

var promptShowCmd = &cobra.Command{
	Use:   "show [name-or-file]",
	Short: "Show a built-in prompt or local prompt file",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		data, err := readPromptTemplate(args[0])
		if err != nil {
			return err
		}
		fmt.Print(string(data))
		return nil
	},
}

var promptRunCmd = &cobra.Command{
	Use:   "run [name-or-file] [input]",
	Short: "Run a prompt template with optional stdin/input",
	Args:  cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		template, err := readPromptTemplate(args[0])
		if err != nil {
			return err
		}
		input := ""
		if len(args) == 2 {
			input = args[1]
		} else if stdinHasData() {
			buf, _ := io.ReadAll(os.Stdin)
			input = string(buf)
		}
		content := strings.TrimSpace(string(template))
		if strings.TrimSpace(input) != "" {
			content += "\n\n---\n\nUser input/context:\n" + strings.TrimSpace(input)
		}

		cfg, err := config.LoadAuto()
		if err != nil {
			return fmt.Errorf("load config: %w", err)
		}
		reg, issues := providers.BuildRegistryFromConfig(cfg)
		for _, issue := range issues {
			fmt.Fprintf(os.Stderr, "warning: %s: %s\n", issue.Provider, issue.Message)
		}
		if len(reg.List()) == 0 {
			return fmt.Errorf("no providers available; configure OPENAI_API_KEY, OPENROUTER_API_KEY, TI_ROUTER_URL, or compatible_providers")
		}
		providerName := promptProvider
		if providerName == "" {
			providerName = cfg.Provider
		}
		model := promptModel
		if model == "" {
			model = cfg.Model
		}
		ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout())
		defer cancel()
		req := providers.ChatRequest{Model: model, Messages: []providers.Message{{Role: "user", Content: content}}}
		if promptStream {
			p, ok := reg.Get(providerName)
			if !ok {
				return fmt.Errorf("provider %q not found", providerName)
			}
			_, err := p.ChatStream(ctx, req, func(chunk providers.StreamChunk) {
				if chunk.Content != "" {
					fmt.Print(chunk.Content)
				}
			})
			if err == nil {
				fmt.Println()
			}
			return err
		}
		resp, usedProvider, err := reg.ChatWithFallback(ctx, req, providerName)
		if err != nil {
			return err
		}
		fmt.Printf("[%s]\n%s\n", usedProvider, resp.Content)
		return nil
	},
}

func readPromptTemplate(name string) ([]byte, error) {
	if promptFile != "" {
		return os.ReadFile(promptFile)
	}
	if _, err := os.Stat(name); err == nil {
		return os.ReadFile(name)
	}
	if strings.ContainsRune(name, filepath.Separator) {
		return nil, fmt.Errorf("prompt file %q not found", name)
	}
	return readDevkitTemplate("commands", name)
}

func stdinHasData() bool {
	info, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return info.Mode()&os.ModeCharDevice == 0
}

func init() {
	rootCmd.AddCommand(promptCmd)
	promptRunCmd.Flags().StringVar(&promptFile, "file", "", "local prompt file to run")
	promptRunCmd.Flags().StringVar(&promptProvider, "provider", "", "provider name to use")
	promptRunCmd.Flags().StringVar(&promptModel, "model", "", "model override")
	promptRunCmd.Flags().BoolVar(&promptStream, "stream", false, "stream response chunks")
	promptCmd.AddCommand(promptListCmd)
	promptCmd.AddCommand(promptShowCmd)
	promptCmd.AddCommand(promptRunCmd)
}
