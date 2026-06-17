package main

import (
	"regexp"
	"strings"
)

type QueryIntent string

const (
	IntentFactual    QueryIntent = "factual"
	IntentHowTo      QueryIntent = "how_to"
	IntentComparison QueryIntent = "comparison"
	IntentStatus     QueryIntent = "status"
	IntentHistorical QueryIntent = "historical"
	IntentCode       QueryIntent = "code"
	IntentImage      QueryIntent = "image"
	IntentMultiModal QueryIntent = "multimodal"
	IntentGeneral    QueryIntent = "general"
)

type QueryEnhancement struct {
	OriginalQuery string
	ExpandedQuery string
	Intent        QueryIntent
	Keywords      []string
	Categories    []string
	Modalities    []string
}

// MultiModalResult contains results from multi-modal queries
type MultiModalResult struct {
	TextResults  []string
	ImageResults []string
	CodeResults  []string
	Modalities   []string
	Combined     string
}

// QueryClassifier classifies query intent and expands keywords
func QueryClassifier(query string) QueryEnhancement {
	intent := classifyIntent(query)
	keywords := extractKeywords(query)
	categories := suggestCategories(query, intent)
	expanded := expandQuery(query, intent)
	modalities := detectModalities(query, intent)

	return QueryEnhancement{
		OriginalQuery: query,
		ExpandedQuery: expanded,
		Intent:        intent,
		Keywords:      keywords,
		Categories:    categories,
		Modalities:    modalities,
	}
}

func classifyIntent(query string) QueryIntent {
	q := strings.ToLower(query)

	// Check for multi-modal first (contains both image and text/code keywords)
	hasImage := matchAny(q, []string{"image", "picture", "photo", "screenshot", "diagram", "chart", "graph", "visual"})
	hasTextCode := matchAny(q, []string{"text", "code", "function", "api", "example", "syntax"})
	if hasImage && hasTextCode {
		return IntentMultiModal
	}

	if matchAny(q, []string{"how to", "how do i", "how can i", "tutorial", "guide", "steps"}) {
		return IntentHowTo
	}
	if matchAny(q, []string{"compare", "difference between", "vs", "versus", "better than"}) {
		return IntentComparison
	}
	if matchAny(q, []string{"status", "running", "active", "current", "now", "live", "real-time"}) {
		return IntentStatus
	}
	if matchAny(q, []string{"history", "past", "previous", "archive", "log", "when did", "timeline"}) {
		return IntentHistorical
	}
	if matchAny(q, []string{"code", "function", "api", "endpoint", "implementation", "syntax", "error", "bug", "class", "struct"}) {
		return IntentCode
	}
	if matchAny(q, []string{"image", "picture", "photo", "screenshot", "diagram", "chart", "graph", "visual"}) {
		return IntentImage
	}
	if matchAny(q, []string{"what is", "who is", "when", "where", "why", "explain", "define"}) {
		return IntentFactual
	}
	return IntentGeneral
}

func extractKeywords(query string) []string {
	// Simple tokenization - remove punctuation, split by spaces
	re := regexp.MustCompile(`[^\w\s]`)
	clean := re.ReplaceAllString(query, " ")
	words := strings.Fields(clean)

	stopwords := map[string]bool{
		"the": true, "a": true, "an": true, "is": true, "are": true, "was": true, "were": true,
		"be": true, "been": true, "being": true, "have": true, "has": true, "had": true,
		"do": true, "does": true, "did": true, "will": true, "would": true, "could": true,
		"should": true, "may": true, "might": true, "must": true, "shall": true,
		"can": true, "need": true, "dare": true, "ought": true, "used": true,
		"to": true, "of": true, "in": true, "for": true, "on": true, "with": true,
		"at": true, "by": true, "from": true, "as": true, "into": true, "through": true,
		"during": true, "before": true, "after": true, "above": true, "below": true,
		"between": true, "under": true, "again": true, "further": true, "then": true,
		"once": true, "here": true, "there": true, "when": true, "where": true,
		"why": true, "how": true, "all": true, "each": true, "few": true, "more": true,
		"most": true, "other": true, "some": true, "such": true, "no": true, "nor": true,
		"not": true, "only": true, "own": true, "same": true, "so": true, "than": true,
		"too": true, "very": true, "just": true, "and": true, "but": true, "or": true,
		"yet": true, "if": true, "because": true, "while": true, "about": true,
		"against": true, "until": true, "since": true, "up": true, "down": true,
		"out": true, "off": true, "over": true, "this": true, "that": true,
		"these": true, "those": true, "i": true, "me": true, "my": true, "myself": true,
		"we": true, "our": true, "ours": true, "ourselves": true, "you": true, "your": true,
		"yours": true, "yourself": true, "yourselves": true, "he": true, "him": true,
		"his": true, "himself": true, "she": true, "her": true, "hers": true,
		"herself": true, "it": true, "its": true, "itself": true, "they": true,
		"them": true, "their": true, "theirs": true, "themselves": true,
	}

	var keywords []string
	for _, w := range words {
		w = strings.ToLower(w)
		if len(w) > 2 && !stopwords[w] {
			keywords = append(keywords, w)
		}
	}
	return keywords
}

func suggestCategories(query string, intent QueryIntent) []string {
	q := strings.ToLower(query)
	var cats []string

	switch intent {
	case IntentCode:
		cats = append(cats, "development", "api", "examples")
	case IntentImage:
		cats = append(cats, "media", "visual", "diagrams")
	case IntentMultiModal:
		cats = append(cats, "multimodal", "mixed")
	case IntentHowTo:
		cats = append(cats, "guides", "tutorials")
	case IntentStatus:
		cats = append(cats, "operations", "monitoring")
	case IntentHistorical:
		cats = append(cats, "archive", "docs")
	case IntentComparison:
		cats = append(cats, "architecture", "docs")
	}

	if strings.Contains(q, "router") || strings.Contains(q, "routing") {
		cats = append(cats, "router")
	}
	if strings.Contains(q, "agent") || strings.Contains(q, "brain") {
		cats = append(cats, "agents")
	}
	if strings.Contains(q, "mcp") || strings.Contains(q, "server") {
		cats = append(cats, "mcp", "integrations")
	}

	return cats
}

// detectModalities identifies which modalities are referenced in the query
func detectModalities(query string, intent QueryIntent) []string {
	q := strings.ToLower(query)
	var modalities []string

	// Image modalities
	if matchAny(q, []string{"image", "picture", "photo", "screenshot", "diagram", "chart", "graph", "visual", "figure"}) {
		modalities = append(modalities, "image")
	}

	// Code modalities
	if matchAny(q, []string{"code", "function", "class", "struct", "method", "api", "syntax", "snippet"}) {
		modalities = append(modalities, "code")
	}

	// Text modalities (default for most queries)
	if intent != IntentImage {
		modalities = append(modalities, "text")
	}

	return modalities
}

// expandQuery adds context-appropriate terms based on intent
func expandQuery(query string, intent QueryIntent) string {
	// Add synonyms based on intent
	switch intent {
	case IntentHowTo:
		query += " steps tutorial guide"
	case IntentFactual:
		query += " definition explanation"
	case IntentCode:
		query += " implementation example syntax code function"
	case IntentImage:
		query += " visual diagram image picture"
	case IntentMultiModal:
		query += " multimodal image text code combined"
	case IntentStatus:
		query += " active running current state"
	case IntentHistorical:
		query += " history log archive past"
	}
	return query
}

// ProcessMultiModalQuery handles queries that combine multiple modalities
func ProcessMultiModalQuery(query string) MultiModalResult {
	intent := classifyIntent(query)

	result := MultiModalResult{
		Modalities: detectModalities(query, intent),
	}

	// Build combined query for multi-modal search
	var parts []string
	if contains(result.Modalities, "text") {
		parts = append(parts, "text search")
	}
	if contains(result.Modalities, "image") {
		parts = append(parts, "image search")
	}
	if contains(result.Modalities, "code") {
		parts = append(parts, "code search")
	}

	result.Combined = strings.Join(parts, " + ")
	return result
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func matchAny(query string, patterns []string) bool {
	for _, p := range patterns {
		if strings.Contains(query, p) {
			return true
		}
	}
	return false
}
