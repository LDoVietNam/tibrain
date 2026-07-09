package frontmatter

import (
	"fmt"
	"regexp"
	"strings"
)

// Frontmatter represents parsed YAML frontmatter from markdown files
type Frontmatter struct {
	Title       string   `yaml:"title"`
	Category    string   `yaml:"category"`
	Tier        string   `yaml:"tier"`
	Tags        []string `yaml:"tags"`
	Scopes      []string `yaml:"scopes"`
	Priority    string   `yaml:"priority"`
	LastUpdated string   `yaml:"last_updated"`
	Version     string   `yaml:"version"`
}

// Parser handles frontmatter parsing and validation
type Parser struct {
	validTags   map[string]bool
	validScopes map[string]bool
}

// NewParser creates a new frontmatter parser with validation rules
func NewParser() *Parser {
	return &Parser{
		validTags:   loadValidTags(),
		validScopes: loadValidScopes(),
	}
}

// ParseFrontmatter extracts and parses YAML frontmatter from markdown content
func (p *Parser) ParseFrontmatter(content string) (*Frontmatter, error) {
	// Extract frontmatter between --- delimiters
	fmRegex := regexp.MustCompile(`(?s)^---\n(.*?)\n---`)
	matches := fmRegex.FindStringSubmatch(content)

	if len(matches) < 2 {
		return nil, fmt.Errorf("no frontmatter found in content")
	}

	fmContent := matches[1]

	// Parse key-value pairs from YAML-like format
	fm := &Frontmatter{}
	lines := strings.Split(fmContent, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, ":", 2)
		if len(parts) < 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// Remove quotes if present
		value = strings.Trim(value, `"`)

		switch key {
		case "title":
			fm.Title = value
		case "category":
			fm.Category = value
		case "tier":
			fm.Tier = value
		case "tags":
			fm.Tags = parseArray(value)
		case "scopes":
			fm.Scopes = parseArray(value)
		case "priority":
			fm.Priority = value
		case "last_updated":
			fm.LastUpdated = value
		case "version":
			fm.Version = value
		}
	}

	return fm, nil
}

// ValidateTags checks if all tags are valid according to TAGS.md taxonomy
func (p *Parser) ValidateTags(tags []string) error {
	for _, tag := range tags {
		if !p.validTags[tag] {
			return fmt.Errorf("invalid tag: %s", tag)
		}
	}
	return nil
}

// ValidateScopes checks if all scopes are valid according to SCOPES.md
func (p *Parser) ValidateScopes(scopes []string) error {
	for _, scope := range scopes {
		if !p.validScopes[scope] {
			return fmt.Errorf("invalid scope: %s", scope)
		}
	}
	return nil
}

// ValidateCategory checks if category is valid
func (p *Parser) ValidateCategory(category string) error {
	validCategories := []string{"core", "quality", "qa", "pattern", "safety", "communication", "infrastructure", "domain", "meta"}

	for _, valid := range validCategories {
		if category == valid {
			return nil
		}
	}

	return fmt.Errorf("invalid category: %s", category)
}

// ValidateTier checks if tier is valid
func (p *Parser) ValidateTier(tier string) error {
	validTiers := []string{"hot", "warm", "cold"}

	for _, valid := range validTiers {
		if tier == valid {
			return nil
		}
	}

	return fmt.Errorf("invalid tier: %s", tier)
}

// parseArray parses a JSON-like array string into a slice
func parseArray(value string) []string {
	// Remove brackets and quotes
	value = strings.Trim(value, "[]")
	value = strings.ReplaceAll(value, `"`, "")

	if value == "" {
		return []string{}
	}

	// Split by comma
	items := strings.Split(value, ",")
	result := make([]string, 0, len(items))

	for _, item := range items {
		item = strings.TrimSpace(item)
		if item != "" {
			result = append(result, item)
		}
	}

	return result
}

// loadValidTags loads valid tags from TAGS.md taxonomy
func loadValidTags() map[string]bool {
	// Core tags
	tags := []string{
		"authentication", "authorization", "rate-limiting", "caching", "translation",
		"retry", "monitoring", "deployment", "troubleshooting", "security",
		"performance", "testing", "documentation",

		// Provider tags
		"provider-antigravity", "provider-openai", "provider-claude", "provider-gemini",
		"provider-deepseek", "provider-groq", "provider-openrouter", "provider-windsurf",
		"provider-notion",

		// Technology tags
		"go", "typescript", "javascript", "python", "rust", "java", "mcp", "oauth",
		"oidc", "jwt", "sse", "grpc", "http", "websocket", "graphql", "rest",
		"sql", "nosql", "docker", "kubernetes", "terraform",

		// Component tags
		"router", "cli", "tibrain", "ticrew", "mcp-server", "provider", "plugin",
		"skill", "workflow", "agent", "dashboard", "api", "database", "cache",

		// Pattern tags
		"pattern-auth-pkce", "pattern-auth-device-code", "pattern-auth-api-key",
		"pattern-auth-cookie", "pattern-retry-exponential", "pattern-retry-circuit-breaker",
		"pattern-cache-content", "pattern-cache-session", "pattern-translation-hub-spoke",
		"pattern-translation-native-passthrough", "pattern-monitoring-metrics",
		"pattern-monitoring-logging", "pattern-monitoring-tracing",
		"pattern-deployment-blue-green", "pattern-deployment-canary",
		"pattern-testing-tdd", "pattern-testing-bdd",

		// Domain tags
		"cli-tools", "web-automation", "browser-automation", "code-generation",
		"code-analysis", "integration", "messaging", "storage", "networking",

		// Scopes
		"auth", "resilience", "integration", "observability", "infrastructure",
		"providers", "code", "web",
	}

	valid := make(map[string]bool)
	for _, tag := range tags {
		valid[tag] = true
	}

	return valid
}

// loadValidScopes loads valid scopes from SCOPES.md
func loadValidScopes() map[string]bool {
	scopes := []string{
		"auth", "resilience", "integration", "observability", "infrastructure",
		"providers", "cli", "tibrain", "ticrew", "web", "code",
	}

	valid := make(map[string]bool)
	for _, scope := range scopes {
		valid[scope] = true
	}

	return valid
}

// ExtractContent extracts content without frontmatter
func (p *Parser) ExtractContent(content string) string {
	fmRegex := regexp.MustCompile(`(?s)^---\n.*?\n---\n`)
	return fmRegex.ReplaceAllString(content, "")
}
