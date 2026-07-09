package main

import (
	"fmt"
	"strings"
)

// ContextAssembler builds the prompt context from retrieved chunks
type ContextAssembler struct {
	maxTokens int
	separator string
}

func NewContextAssembler() *ContextAssembler {
	return &ContextAssembler{
		maxTokens: 4000,
		separator: "\n\n---\n\n",
	}
}

// AssembledContext holds the final context and source mapping
type AssembledContext struct {
	Text       string
	Sources    []ContextSource
	TokenCount int
}

type ContextSource struct {
	DocID   string
	Title   string
	Path    string
	Excerpt string
}

// Assemble builds context from documents respecting token budget
func (ca *ContextAssembler) Assemble(docs []RAGDocument, query string) AssembledContext {
	var parts []string
	var sources []ContextSource
	currentTokens := 0

	for _, doc := range docs {
		excerpt := ca.truncate(doc.Content, 800)
		part := fmt.Sprintf("[Source: %s]\n%s", doc.Title, excerpt)
		tokens := ca.estimateTokens(part)

		if currentTokens+tokens > ca.maxTokens {
			break
		}

		parts = append(parts, part)
		sources = append(sources, ContextSource{
			DocID:   doc.ID,
			Title:   doc.Title,
			Path:    doc.Path,
			Excerpt: excerpt,
		})
		currentTokens += tokens
	}

	return AssembledContext{
		Text:       strings.Join(parts, ca.separator),
		Sources:    sources,
		TokenCount: currentTokens,
	}
}

func (ca *ContextAssembler) truncate(text string, maxChars int) string {
	if len(text) <= maxChars {
		return text
	}
	return text[:maxChars] + "..."
}

func (ca *ContextAssembler) estimateTokens(text string) int {
	// Rough estimate: ~4 chars per token for English/Vietnamese
	return len(text) / 4
}
