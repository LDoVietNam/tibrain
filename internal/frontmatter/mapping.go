package frontmatter

import (
	"fmt"
	"strings"
	"time"
)

// Mapper handles frontmatter mapping between Obsidian and Ti Brain formats
type Mapper struct {
	validTags   map[string]bool
	validScopes map[string]bool
}

// NewMapper creates a new frontmatter mapper
func NewMapper(validTags, validScopes []string) *Mapper {
	tagMap := make(map[string]bool)
	for _, tag := range validTags {
		tagMap[strings.ToLower(tag)] = true
	}

	scopeMap := make(map[string]bool)
	for _, scope := range validScopes {
		scopeMap[strings.ToLower(scope)] = true
	}

	return &Mapper{
		validTags:   tagMap,
		validScopes: scopeMap,
	}
}

// ObsidianFrontmatter represents Obsidian frontmatter structure (extended)
type ObsidianFrontmatter struct {
	Frontmatter
	URL    string `yaml:"url,omitempty"`
	Source string `yaml:"source,omitempty"`
}

// MapObsidianToTiBrain maps Obsidian frontmatter to Ti Brain format
func (m *Mapper) MapObsidianToTiBrain(obsidian ObsidianFrontmatter) (Frontmatter, error) {
	validTags := m.validateTags(obsidian.Tags)
	validScopes := m.validateScopes(obsidian.Scopes)

	// Set defaults
	category := obsidian.Category
	if category == "" {
		category = "general"
	}

	tier := obsidian.Tier
	if tier == "" {
		tier = "T3"
	}

	priority := obsidian.Priority
	if priority == "" {
		priority = "P2"
	}

	lastUpdated := obsidian.LastUpdated
	if lastUpdated == "" {
		lastUpdated = time.Now().Format(time.RFC3339)
	}

	tibrain := Frontmatter{
		Title:       obsidian.Title,
		Tags:        validTags,
		Scopes:      validScopes,
		Category:    category,
		Tier:        tier,
		Priority:    priority,
		LastUpdated: lastUpdated,
		Version:     obsidian.Version,
	}

	return tibrain, nil
}

// MapTiBrainToObsidian maps Ti Brain frontmatter to Obsidian format
func (m *Mapper) MapTiBrainToObsidian(tibrain Frontmatter) (ObsidianFrontmatter, error) {
	obsidian := ObsidianFrontmatter{
		Frontmatter: tibrain,
	}

	return obsidian, nil
}

// validateTags validates tags against Ti Brain taxonomy
func (m *Mapper) validateTags(tags []string) []string {
	valid := make([]string, 0, len(tags))
	invalid := make([]string, 0)

	for _, tag := range tags {
		normalized := normalizeTag(tag)
		if m.validTags[normalized] {
			valid = append(valid, normalized)
		} else {
			invalid = append(invalid, tag)
		}
	}

	if len(invalid) > 0 {
		// Log warning (in real implementation, use proper logger)
		fmt.Printf("Warning: Invalid tags (not in taxonomy): %v\n", invalid)
	}

	return valid
}

// validateScopes validates scopes against Ti Brain taxonomy
func (m *Mapper) validateScopes(scopes []string) []string {
	valid := make([]string, 0, len(scopes))
	invalid := make([]string, 0)

	for _, scope := range scopes {
		normalized := strings.ToLower(scope)
		if m.validScopes[normalized] {
			valid = append(valid, normalized)
		} else {
			invalid = append(invalid, scope)
		}
	}

	if len(invalid) > 0 {
		fmt.Printf("Warning: Invalid scopes (not in taxonomy): %v\n", invalid)
	}

	return valid
}

// normalizeTag normalizes tag format for consistency
func normalizeTag(tag string) string {
	// Trim whitespace first
	tag = strings.TrimSpace(tag)
	
	// Remove leading # if present
	tag = strings.TrimPrefix(tag, "#")
	
	// Replace slashes with dashes (hierarchical tags)
	tag = strings.ReplaceAll(tag, "/", "-")
	
	// Convert to lowercase
	tag = strings.ToLower(tag)
	
	return tag
}

// ValidateTiBrainFrontmatter validates Ti Brain frontmatter structure
func ValidateTiBrainFrontmatter(frontmatter Frontmatter) error {
	// Validate tier
	validTiers := map[string]bool{"T1": true, "T2": true, "T3": true}
	if !validTiers[frontmatter.Tier] {
		return fmt.Errorf("invalid tier: %s (must be T1, T2, or T3)", frontmatter.Tier)
	}

	// Validate priority
	validPriorities := map[string]bool{"P0": true, "P1": true, "P2": true, "P3": true}
	if !validPriorities[frontmatter.Priority] {
		return fmt.Errorf("invalid priority: %s (must be P0, P1, P2, or P3)", frontmatter.Priority)
	}

	// Validate last_updated format
	if frontmatter.LastUpdated != "" {
		_, err := time.Parse(time.RFC3339, frontmatter.LastUpdated)
		if err != nil {
			return fmt.Errorf("invalid last_updated format: %s (must be RFC3339)", frontmatter.LastUpdated)
		}
	}

	return nil
}

// MergeFrontmatter merges two frontmatter structures (Ti Brain takes precedence)
func MergeFrontmatter(base, override Frontmatter) Frontmatter {
	merged := base

	if override.Title != "" {
		merged.Title = override.Title
	}

	if len(override.Tags) > 0 {
		merged.Tags = override.Tags
	}

	if len(override.Scopes) > 0 {
		merged.Scopes = override.Scopes
	}

	if override.Category != "" {
		merged.Category = override.Category
	}

	if override.Tier != "" {
		merged.Tier = override.Tier
	}

	if override.Priority != "" {
		merged.Priority = override.Priority
	}

	if override.LastUpdated != "" {
		merged.LastUpdated = override.LastUpdated
	}

	if override.Version != "" {
		merged.Version = override.Version
	}

	return merged
}
