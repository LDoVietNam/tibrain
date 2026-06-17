package frontmatter

import (
	"testing"
	"time"
)

func TestNewMapper(t *testing.T) {
	validTags := []string{"go", "typescript", "testing"}
	validScopes := []string{"auth", "providers", "web"}

	mapper := NewMapper(validTags, validScopes)

	if mapper == nil {
		t.Fatal("NewMapper returned nil")
	}

	if len(mapper.validTags) != len(validTags) {
		t.Errorf("Expected %d valid tags, got %d", len(validTags), len(mapper.validTags))
	}

	if len(mapper.validScopes) != len(validScopes) {
		t.Errorf("Expected %d valid scopes, got %d", len(validScopes), len(mapper.validScopes))
	}
}

func TestMapObsidianToTiBrain(t *testing.T) {
	validTags := []string{"go", "typescript", "testing", "provider-antigravity"}
	validScopes := []string{"auth", "providers", "web", "code"}
	mapper := NewMapper(validTags, validScopes)

	obsidian := ObsidianFrontmatter{
		Frontmatter: Frontmatter{
			Title:       "Test Note",
			Category:    "core",
			Tier:        "T1",
			Tags:        []string{"#go", "typescript", "invalid-tag"},
			Scopes:      []string{"providers", "invalid-scope"},
			Priority:    "P1",
			LastUpdated: time.Now().Format(time.RFC3339),
			Version:     "1.0.0",
		},
		URL:    "https://example.com",
		Source: "obsidian",
	}

	result, err := mapper.MapObsidianToTiBrain(obsidian)
	if err != nil {
		t.Fatalf("MapObsidianToTiBrain failed: %v", err)
	}

	if result.Title != "Test Note" {
		t.Errorf("Expected title 'Test Note', got '%s'", result.Title)
	}

	if result.Category != "core" {
		t.Errorf("Expected category 'core', got '%s'", result.Category)
	}

	if result.Tier != "T1" {
		t.Errorf("Expected tier 'T1', got '%s'", result.Tier)
	}

	if result.Priority != "P1" {
		t.Errorf("Expected priority 'P1', got '%s'", result.Priority)
	}

	// Check that invalid tags were filtered out
	expectedValidTags := 2
	if len(result.Tags) != expectedValidTags {
		t.Errorf("Expected %d valid tags, got %d", expectedValidTags, len(result.Tags))
	}

	// Check that invalid scopes were filtered out
	expectedValidScopes := 1
	if len(result.Scopes) != expectedValidScopes {
		t.Errorf("Expected %d valid scopes, got %d", expectedValidScopes, len(result.Scopes))
	}
}

func TestMapObsidianToTiBrainDefaults(t *testing.T) {
	validTags := []string{"go"}
	validScopes := []string{"code"}
	mapper := NewMapper(validTags, validScopes)

	obsidian := ObsidianFrontmatter{
		Frontmatter: Frontmatter{
			Title: "Test Note",
			// Other fields are empty to test defaults
		},
	}

	result, err := mapper.MapObsidianToTiBrain(obsidian)
	if err != nil {
		t.Fatalf("MapObsidianToTiBrain failed: %v", err)
	}

	if result.Category != "general" {
		t.Errorf("Expected default category 'general', got '%s'", result.Category)
	}

	if result.Tier != "T3" {
		t.Errorf("Expected default tier 'T3', got '%s'", result.Tier)
	}

	if result.Priority != "P2" {
		t.Errorf("Expected default priority 'P2', got '%s'", result.Priority)
	}

	if result.LastUpdated == "" {
		t.Error("Expected default LastUpdated to be set")
	}
}

func TestMapTiBrainToObsidian(t *testing.T) {
	validTags := []string{"go"}
	validScopes := []string{"code"}
	mapper := NewMapper(validTags, validScopes)

	tibrain := Frontmatter{
		Title:       "Test Note",
		Category:    "core",
		Tier:        "T1",
		Tags:        []string{"go"},
		Scopes:      []string{"code"},
		Priority:    "P1",
		LastUpdated: time.Now().Format(time.RFC3339),
		Version:     "1.0.0",
	}

	result, err := mapper.MapTiBrainToObsidian(tibrain)
	if err != nil {
		t.Fatalf("MapTiBrainToObsidian failed: %v", err)
	}

	if result.Title != tibrain.Title {
		t.Errorf("Expected title '%s', got '%s'", tibrain.Title, result.Title)
	}

	if result.Category != tibrain.Category {
		t.Errorf("Expected category '%s', got '%s'", tibrain.Category, result.Category)
	}
}

func TestNormalizeTag(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"#go", "go"},
		{"#typescript", "typescript"},
		{"go", "go"},
		{"#web/development", "web-development"},
		{"#Web/Development", "web-development"},
		{"  #go  ", "go"},
	}

	for _, tt := range tests {
		result := normalizeTag(tt.input)
		if result != tt.expected {
			t.Errorf("normalizeTag(%q) = %q, expected %q", tt.input, result, tt.expected)
		}
	}
}

func TestValidateTags(t *testing.T) {
	validTags := []string{"go", "typescript", "testing"}
	validScopes := []string{"code"}
	mapper := NewMapper(validTags, validScopes)

	inputTags := []string{"#go", "typescript", "invalid-tag"}
	result := mapper.validateTags(inputTags)

	if len(result) != 2 {
		t.Errorf("Expected 2 valid tags, got %d", len(result))
	}
}

func TestValidateScopes(t *testing.T) {
	validTags := []string{"go"}
	validScopes := []string{"auth", "providers", "web"}
	mapper := NewMapper(validTags, validScopes)

	inputScopes := []string{"providers", "invalid-scope"}
	result := mapper.validateScopes(inputScopes)

	if len(result) != 1 {
		t.Errorf("Expected 1 valid scope, got %d", len(result))
	}
}

func TestValidateTiBrainFrontmatter(t *testing.T) {
	tests := []struct {
		name        string
		frontmatter Frontmatter
		expectError bool
	}{
		{
			name: "Valid frontmatter",
			frontmatter: Frontmatter{
				Tier:        "T1",
				Priority:    "P1",
				LastUpdated: time.Now().Format(time.RFC3339),
			},
			expectError: false,
		},
		{
			name: "Invalid tier",
			frontmatter: Frontmatter{
				Tier:        "T4",
				Priority:    "P1",
				LastUpdated: time.Now().Format(time.RFC3339),
			},
			expectError: true,
		},
		{
			name: "Invalid priority",
			frontmatter: Frontmatter{
				Tier:        "T1",
				Priority:    "P4",
				LastUpdated: time.Now().Format(time.RFC3339),
			},
			expectError: true,
		},
		{
			name: "Invalid date format",
			frontmatter: Frontmatter{
				Tier:        "T1",
				Priority:    "P1",
				LastUpdated: "invalid-date",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateTiBrainFrontmatter(tt.frontmatter)
			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}

func TestMergeFrontmatter(t *testing.T) {
	base := Frontmatter{
		Title:       "Base Title",
		Category:    "core",
		Tier:        "T2",
		Tags:        []string{"go"},
		Scopes:      []string{"code"},
		Priority:    "P2",
		LastUpdated: "2024-01-01T00:00:00Z",
		Version:     "1.0.0",
	}

	override := Frontmatter{
		Title:    "Override Title",
		Category: "quality",
		Tags:     []string{"typescript"},
	}

	result := MergeFrontmatter(base, override)

	if result.Title != "Override Title" {
		t.Errorf("Expected title 'Override Title', got '%s'", result.Title)
	}

	if result.Category != "quality" {
		t.Errorf("Expected category 'quality', got '%s'", result.Category)
	}

	if result.Tags[0] != "typescript" {
		t.Errorf("Expected tag 'typescript', got '%s'", result.Tags[0])
	}

	// Check that non-overridden fields remain from base
	if result.Tier != "T2" {
		t.Errorf("Expected tier 'T2' from base, got '%s'", result.Tier)
	}

	if result.Priority != "P2" {
		t.Errorf("Expected priority 'P2' from base, got '%s'", result.Priority)
	}
}
