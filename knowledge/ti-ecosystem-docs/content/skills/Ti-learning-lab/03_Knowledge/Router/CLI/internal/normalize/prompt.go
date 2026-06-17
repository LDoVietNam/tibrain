// Package normalize provides prompt normalization for the Ti platform.
// It cleans and compresses prompt text by removing comments, functions,
// and excessive whitespace while tracking savings metrics.
package normalize

import (
	"regexp"
	"strings"
)

// NormalizedPrompt holds the result of normalizing a prompt.
type NormalizedPrompt struct {
	Original      string   // Original prompt text
	Normalized    string   // Cleaned prompt text
	OriginalLen   int      // Original character count
	NormalizedLen int      // Normalized character count
	Savings       float64  // Percentage saved: (1 - normalized/original) * 100
	TasksRemoved  []string // What was removed (for debugging)
}

// Normalizer configures prompt normalization behavior.
type Normalizer struct {
	maxFileLines    int
	maxLineLen      int
	maxPromptLen    int
	stripFunctions  bool
	stripComments   bool
	stripWhitespace bool
}

const (
	defaultMaxFileLines = 500
	defaultMaxLineLen   = 2000
	defaultMaxPromptLen = 100000
)

// NewNormalizer creates a Normalizer with sensible defaults:
// maxFileLines=500, maxLineLen=2000, maxPromptLen=100000,
// stripFunctions=true, stripComments=true, stripWhitespace=true.
func NewNormalizer() *Normalizer {
	return &Normalizer{
		maxFileLines:    defaultMaxFileLines,
		maxLineLen:      defaultMaxLineLen,
		maxPromptLen:    defaultMaxPromptLen,
		stripFunctions:  true,
		stripComments:   true,
		stripWhitespace: true,
	}
}

// WithMaxFileLines sets the maximum lines per @file mention.
func (n *Normalizer) WithMaxFileLines(lines int) *Normalizer {
	n.maxFileLines = lines
	return n
}

// WithMaxLineLen sets the maximum characters per line.
func (n *Normalizer) WithMaxLineLen(nLen int) *Normalizer {
	n.maxLineLen = nLen
	return n
}

// WithMaxPromptLen sets the maximum total prompt length.
func (n *Normalizer) WithMaxPromptLen(nLen int) *Normalizer {
	n.maxPromptLen = nLen
	return n
}

// WithStripFunctions enables or disables function definition removal.
func (n *Normalizer) WithStripFunctions(v bool) *Normalizer {
	n.stripFunctions = v
	return n
}

// WithStripComments enables or disables code comment removal.
func (n *Normalizer) WithStripComments(v bool) *Normalizer {
	n.stripComments = v
	return n
}

// WithStripWhitespace enables or disables excessive whitespace collapsing.
func (n *Normalizer) WithStripWhitespace(v bool) *Normalizer {
	n.stripWhitespace = v
	return n
}

// Normalize processes the input prompt through all configured normalization steps.
// Steps execute in order: strip comments, strip functions, collapse whitespace,
// trim, truncate, then calculate metrics.
func (n *Normalizer) Normalize(prompt string) *NormalizedPrompt {
	original := prompt
	originalLen := len(prompt)
	removed := make([]string, 0, 8)

	result := prompt

	// Step 1: Strip code comments
	if n.stripComments {
		result, removed = stripComments(result, removed)
	}

	// Step 2: Strip function definitions
	if n.stripFunctions {
		result, removed = stripFunctions(result, removed)
	}

	// Step 3: Collapse excessive whitespace
	if n.stripWhitespace {
		result = collapseWhitespace(result)
	}

	// Step 4: Trim leading/trailing whitespace
	result = strings.TrimSpace(result)

	// Step 5: Truncate lines per @file mentions
	result = truncateFileLines(result, n.maxFileLines)

	// Step 6: Truncate line length
	result = truncateLineLengths(result, n.maxLineLen)

	// Step 7: Truncate total length
	if len(result) > n.maxPromptLen {
		result = result[:n.maxPromptLen]
		removed = append(removed, "truncated to maxPromptLen")
	}

	normalizedLen := len(result)
	savings := 0.0
	if originalLen > 0 {
		savings = (1.0 - float64(normalizedLen)/float64(originalLen)) * 100.0
	}

	return &NormalizedPrompt{
		Original:      original,
		Normalized:    result,
		OriginalLen:   originalLen,
		NormalizedLen: normalizedLen,
		Savings:       savings,
		TasksRemoved:  removed,
	}
}

// Precompiled comment patterns.
var (
	// Multi-line C-style comments: /* ... */
	reBlockComment = regexp.MustCompile(`(?s)/\*.*?\*/`)
	// Multi-line HTML comments: <!-- ... -->
	reHTMLComment = regexp.MustCompile(`(?s)<!--.*?-->`)
	// Multi-line Python docstrings: """ ... """ or ''' ... '''
	reTripleDoubleQuote = regexp.MustCompile(`(?s)""".*?"""`)
	reTripleSingleQuote = regexp.MustCompile(`(?s)'''[^']*'''`)
)

// stripComments removes code comments and tracks what was removed.
func stripComments(input string, removed []string) (string, []string) {
	// Count and remove block comments /* ... */
	blockMatches := reBlockComment.FindAllString(input, -1)
	if len(blockMatches) > 0 {
		removed = append(removed, "block comments")
	}
	result := reBlockComment.ReplaceAllString(input, "")

	// Count and remove HTML comments <!-- ... -->
	htmlMatches := reHTMLComment.FindAllString(result, -1)
	if len(htmlMatches) > 0 {
		removed = append(removed, "HTML comments")
	}
	result = reHTMLComment.ReplaceAllString(result, "")

	// Count and remove Python triple-double-quoted strings (docstrings)
	tdqMatches := reTripleDoubleQuote.FindAllString(result, -1)
	if len(tdqMatches) > 0 {
		removed = append(removed, "triple-double-quoted strings")
	}
	result = reTripleDoubleQuote.ReplaceAllString(result, "")

	// Count and remove Python triple-single-quoted strings
	tsqMatches := reTripleSingleQuote.FindAllString(result, -1)
	if len(tsqMatches) > 0 {
		removed = append(removed, "triple-single-quoted strings")
	}
	result = reTripleSingleQuote.ReplaceAllString(result, "")

	// Handle single-line comments: // ... and # ...
	lines := strings.Split(result, "\n")
	var cleaned []string
	var foundSingleLine, foundHash bool
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Skip lines that are purely single-line comments
		if strings.HasPrefix(trimmed, "//") {
			foundSingleLine = true
			continue
		}

		// Remove inline // comments but preserve the code before them
		if idx := strings.Index(line, "//"); idx >= 0 {
			// Check it's not inside a string literal (basic heuristic)
			before := line[:idx]
			if !strings.Contains(before, `"`) {
				line = before
			}
		}

		// Skip lines that are purely # comments
		if strings.HasPrefix(trimmed, "#") && !strings.HasPrefix(trimmed, "#!") {
			foundHash = true
			continue
		}

		cleaned = append(cleaned, line)
	}

	if foundSingleLine {
		removed = append(removed, "// line comments")
	}
	if foundHash {
		removed = append(removed, "# comments")
	}

	return strings.Join(cleaned, "\n"), removed
}

// Precompiled function definition patterns.
var (
	reFuncGo    = regexp.MustCompile(`(?m)^func\s+.*\{`)
	reFuncJS    = regexp.MustCompile(`(?m)^function\s+.*\{`)
	reDefPython = regexp.MustCompile(`(?m)^(?:async\s+)?def\s+`)
	reClassDef  = regexp.MustCompile(`(?m)^class\s+`)
	reArrowFunc = regexp.MustCompile(`(?m)^const\s+\w+\s*=\s*(?:async\s+)?\([^)]*\)\s*=>`)
	reMethodDef = regexp.MustCompile(`(?m)^\s+(?:async\s+)?\w+\s*\(.*\)\s*[:{]`)
)

// stripFunctions removes function/class definitions and tracks removals.
func stripFunctions(input string, removed []string) (string, []string) {
	lines := strings.Split(input, "\n")
	var result []string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			result = append(result, line)
			continue
		}

		isFunc := false

		// Check each pattern
		for _, re := range []*regexp.Regexp{
			reFuncGo, reFuncJS, reDefPython, reClassDef, reArrowFunc, reMethodDef,
		} {
			if re.MatchString(line) {
				isFunc = true
				break
			}
		}

		if isFunc {
			// Identify which type
			switch {
			case reFuncGo.MatchString(line):
				removed = append(removed, "Go func definition")
			case reFuncJS.MatchString(line):
				removed = append(removed, "JS function definition")
			case reDefPython.MatchString(line):
				removed = append(removed, "Python def")
			case reClassDef.MatchString(line):
				removed = append(removed, "class definition")
			case reArrowFunc.MatchString(line):
				removed = append(removed, "arrow function")
			case reMethodDef.MatchString(line):
				removed = append(removed, "method definition")
			}
			continue
		}

		result = append(result, line)
	}

	return strings.Join(result, "\n"), removed
}

// collapseWhitespace reduces 3+ consecutive blank lines to 1 blank line,
// and trims trailing whitespace from each line.
func collapseWhitespace(input string) string {
	// More than 2 consecutive newlines (3+ blank lines) -> 2 newlines (1 blank line)
	re := regexp.MustCompile(`\n{4,}`)
	result := re.ReplaceAllString(input, "\n\n")

	// Trim trailing whitespace from each line
	lines := strings.Split(result, "\n")
	for i, line := range lines {
		lines[i] = strings.TrimRight(line, " \t")
	}

	return strings.Join(lines, "\n")
}

// truncateFileLines limits the number of lines shown after @file mentions.
// An @file mention is a line containing "@file" or "@filepath" or similar patterns.
func truncateFileLines(input string, maxLines int) string {
	lines := strings.Split(input, "\n")
	var result []string
	inFileBlock := false
	lineCount := 0

	for _, line := range lines {
		// Detect @file mention start
		if strings.Contains(line, "@file") || strings.Contains(line, "@filepath") {
			inFileBlock = true
			lineCount = 0
			result = append(result, line)
			continue
		}

		if inFileBlock {
			// Detect end of file block (blank line or another directive)
			if strings.TrimSpace(line) == "" {
				inFileBlock = false
				result = append(result, line)
				continue
			}
			// Another @ mention starts a new block
			if strings.Contains(line, "@") {
				inFileBlock = true
				lineCount = 0
				result = append(result, line)
				continue
			}

			lineCount++
			if lineCount <= maxLines {
				result = append(result, line)
			}
			continue
		}

		result = append(result, line)
	}

	return strings.Join(result, "\n")
}

// truncateLineLengths truncates each line to maxLen characters.
func truncateLineLengths(input string, maxLen int) string {
	if maxLen <= 0 {
		return input
	}

	lines := strings.Split(input, "\n")
	for i, line := range lines {
		if len(line) > maxLen {
			lines[i] = line[:maxLen]
		}
	}
	return strings.Join(lines, "\n")
}
