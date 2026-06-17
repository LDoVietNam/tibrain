package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/ti/cli/internal/config"
)

var skillsDir string

var skillCmd = &cobra.Command{
	Use:   "skill",
	Short: "Manage and query skills",
	Long:  `List, show, and search skills from the best_source repository.`,
}

var skillListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all available skills",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Determine skills directory
		dir := resolveSkillsDir()
		if dir == "" {
			return fmt.Errorf("skills directory not found. Set --skills-dir or configure skills_dir in ti.json")
		}

		// Read skills directory
		entries, err := os.ReadDir(dir)
		if err != nil {
			return fmt.Errorf("failed to read skills directory '%s': %w", dir, err)
		}

		fmt.Printf("Skills directory: %s\n\n", dir)

		var skills []skillInfo
		for _, entry := range entries {
			if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
				continue
			}

			info := parseSkillInfo(filepath.Join(dir, entry.Name()))
			if info != nil {
				skills = append(skills, *info)
			}
		}

		if len(skills) == 0 {
			fmt.Println("No skills found in directory.")
			return nil
		}

		fmt.Printf("Found %d skills:\n\n", len(skills))
		for _, s := range skills {
			priority := s.priority
			if priority == "" {
				priority = "normal"
			}
			fmt.Printf("  %-35s [%s] %s\n", s.name, priority, s.description)
		}

		return nil
	},
}

var skillShowCmd = &cobra.Command{
	Use:   "show [name]",
	Short: "Display a skill's content",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		// Determine skills directory
		dir := resolveSkillsDir()
		if dir == "" {
			return fmt.Errorf("skills directory not found. Set --skills-dir or configure skills_dir in ti.json")
		}

		// Try exact match first
		path := filepath.Join(dir, name+".md")
		if _, err := os.Stat(path); os.IsNotExist(err) {
			// Try fuzzy match
			entries, _ := os.ReadDir(dir)
			for _, entry := range entries {
				if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
					continue
				}
				baseName := strings.TrimSuffix(entry.Name(), ".md")
				if strings.EqualFold(baseName, name) || strings.Contains(strings.ToLower(baseName), strings.ToLower(name)) {
					path = filepath.Join(dir, entry.Name())
					name = baseName
					break
				}
			}
		}

		// Read skill file
		content, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("skill '%s' not found in %s", name, dir)
		}

		fmt.Printf("# %s\n\n", name)
		fmt.Println(string(content))

		return nil
	},
}

var skillSearchCmd = &cobra.Command{
	Use:   "search [query]",
	Short: "Search skills by keyword",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		query := strings.ToLower(args[0])

		// Determine skills directory
		dir := resolveSkillsDir()
		if dir == "" {
			return fmt.Errorf("skills directory not found. Set --skills-dir or configure skills_dir in ti.json")
		}

		// Read skills directory
		entries, err := os.ReadDir(dir)
		if err != nil {
			return fmt.Errorf("failed to read skills directory '%s': %w", dir, err)
		}

		var results []skillInfo
		for _, entry := range entries {
			if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
				continue
			}

			path := filepath.Join(dir, entry.Name())
			info := parseSkillInfo(path)
			if info == nil {
				continue
			}

			// Search in name, description, tags, and content
			if strings.Contains(strings.ToLower(info.name), query) ||
				strings.Contains(strings.ToLower(info.description), query) ||
				strings.Contains(strings.ToLower(info.content), query) {
				results = append(results, *info)
			}
		}

		fmt.Printf("Found %d skills matching '%s':\n\n", len(results), args[0])
		for _, s := range results {
			priority := s.priority
			if priority == "" {
				priority = "normal"
			}
			fmt.Printf("  %-35s [%s] %s\n", s.name, priority, s.description)
		}

		return nil
	},
}

// skillInfo holds parsed skill metadata
type skillInfo struct {
	name        string
	path        string
	priority    string
	description string
	content     string
}

// resolveSkillsDir determines the skills directory from multiple sources
func resolveSkillsDir() string {
	// 1. Flag override
	if skillsDir != "" {
		return skillsDir
	}

	// 2. Config file
	cfg, err := config.LoadAuto()
	if err == nil && cfg.SkillsDir != "" {
		return cfg.SkillsDir
	}

	// 3. Environment variable
	if env := os.Getenv("TI_SKILLS_DIR"); env != "" {
		return env
	}

	// 4. Default path
	if _, err := os.Stat(`Z:\Ti\best_source\skills`); err == nil {
		return `Z:\Ti\best_source\skills`
	}

	// 5. Relative to CWD
	if _, err := os.Stat("best_source/skills"); err == nil {
		path, _ := filepath.Abs("best_source/skills")
		return path
	}

	return ""
}

// parseSkillInfo reads a markdown skill file and extracts metadata
func parseSkillInfo(path string) *skillInfo {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil
	}

	name := strings.TrimSuffix(filepath.Base(path), ".md")
	info := &skillInfo{
		name:    name,
		path:    path,
		content: string(content),
	}

	// Parse YAML frontmatter if present
	lines := strings.Split(string(content), "\n")
	inFrontmatter := false

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Detect frontmatter start/end
		if trimmed == "---" {
			if !inFrontmatter {
				inFrontmatter = true
				continue
			} else {
				break
			}
		}

		if inFrontmatter {
			// Parse key: value
			if idx := strings.Index(line, ":"); idx > 0 {
				key := strings.TrimSpace(line[:idx])
				value := strings.TrimSpace(line[idx+1:])

				switch strings.ToLower(key) {
				case "priority":
					info.priority = value
				case "description":
					info.description = value
				}
			}
		}

		// Extract description from first heading if no frontmatter
		if !inFrontmatter && i < 10 && info.description == "" {
			if strings.HasPrefix(trimmed, "# ") {
				info.description = strings.TrimPrefix(trimmed, "# ")
				break
			}
		}
	}

	return info
}

// readSessionIDFromFile reads a session ID from various file formats
// This is duplicated from providers/bootstrap.go for local use
func readSessionIDFromFile(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}

	content := string(data)

	// Try KEY=VALUE format (.env style)
	scanner := bufio.NewScanner(strings.NewReader(content))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "#") || line == "" {
			continue
		}
		if strings.HasPrefix(strings.ToUpper(line), "SESSIONID=") ||
			strings.HasPrefix(strings.ToUpper(line), "SESSION_ID=") ||
			strings.HasPrefix(strings.ToUpper(line), "SHAREDCHAT_SESSION=") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				return strings.Trim(strings.TrimSpace(parts[1]), `"'`)
			}
		}
	}

	// Try raw content (just the session ID itself)
	raw := strings.TrimSpace(content)
	raw = strings.Trim(raw, `"'`)
	if len(raw) > 10 && !strings.Contains(raw, "=") && !strings.Contains(raw, "{") {
		return raw
	}

	return ""
}

func init() {
	skillCmd.PersistentFlags().StringVar(&skillsDir, "skills-dir", "", "skills directory path")

	rootCmd.AddCommand(skillCmd)
	skillCmd.AddCommand(skillListCmd)
	skillCmd.AddCommand(skillShowCmd)
	skillCmd.AddCommand(skillSearchCmd)
}
