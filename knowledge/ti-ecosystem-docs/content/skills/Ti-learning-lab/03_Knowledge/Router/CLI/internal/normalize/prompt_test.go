package normalize

import (
	"strings"
	"testing"
)

func TestNewNormalizerDefaults(t *testing.T) {
	n := NewNormalizer()
	if n.maxFileLines != 500 {
		t.Errorf("expected maxFileLines=500, got %d", n.maxFileLines)
	}
	if n.maxLineLen != 2000 {
		t.Errorf("expected maxLineLen=2000, got %d", n.maxLineLen)
	}
	if n.maxPromptLen != 100000 {
		t.Errorf("expected maxPromptLen=100000, got %d", n.maxPromptLen)
	}
	if !n.stripFunctions {
		t.Error("expected stripFunctions=true by default")
	}
	if !n.stripComments {
		t.Error("expected stripComments=true by default")
	}
	if !n.stripWhitespace {
		t.Error("expected stripWhitespace=true by default")
	}
}

func TestBuilderPattern(t *testing.T) {
	n := NewNormalizer().
		WithMaxFileLines(100).
		WithMaxLineLen(500).
		WithMaxPromptLen(50000).
		WithStripFunctions(false).
		WithStripComments(false).
		WithStripWhitespace(false)

	if n.maxFileLines != 100 {
		t.Errorf("expected maxFileLines=100, got %d", n.maxFileLines)
	}
	if n.maxLineLen != 500 {
		t.Errorf("expected maxLineLen=500, got %d", n.maxLineLen)
	}
	if n.maxPromptLen != 50000 {
		t.Errorf("expected maxPromptLen=50000, got %d", n.maxPromptLen)
	}
	if n.stripFunctions {
		t.Error("expected stripFunctions=false")
	}
	if n.stripComments {
		t.Error("expected stripComments=false")
	}
	if n.stripWhitespace {
		t.Error("expected stripWhitespace=false")
	}
}

func TestEmptyPrompt(t *testing.T) {
	n := NewNormalizer()
	result := n.Normalize("")
	if result.Original != "" {
		t.Errorf("expected empty original, got %q", result.Original)
	}
	if result.Normalized != "" {
		t.Errorf("expected empty normalized, got %q", result.Normalized)
	}
	if result.OriginalLen != 0 {
		t.Errorf("expected originalLen=0, got %d", result.OriginalLen)
	}
	if result.NormalizedLen != 0 {
		t.Errorf("expected normalizedLen=0, got %d", result.NormalizedLen)
	}
	if result.Savings != 0.0 {
		t.Errorf("expected savings=0.0, got %f", result.Savings)
	}
}

func TestStripBlockComments(t *testing.T) {
	input := "Hello\n/* this is a\nmulti-line comment */\nWorld"
	n := NewNormalizer()
	result := n.Normalize(input)

	if strings.Contains(result.Normalized, "multi-line comment") {
		t.Error("block comment should be stripped")
	}
	if !containsStr(result.TasksRemoved, "block comments") {
		t.Errorf("expected 'block comments' in TasksRemoved, got %v", result.TasksRemoved)
	}
}

func TestStripSingleLineComments(t *testing.T) {
	input := "code line\n// this is a comment\nmore code"
	n := NewNormalizer()
	result := n.Normalize(input)

	if strings.Contains(result.Normalized, "// this is a comment") {
		t.Error("// comment line should be stripped")
	}
	if !containsStr(result.TasksRemoved, "// line comments") {
		t.Errorf("expected '// line comments' in TasksRemoved, got %v", result.TasksRemoved)
	}
}

func TestStripHTMLComments(t *testing.T) {
	input := "<!-- HTML comment -->\nActual content"
	n := NewNormalizer()
	result := n.Normalize(input)

	if strings.Contains(result.Normalized, "HTML comment") {
		t.Error("HTML comment should be stripped")
	}
	if !containsStr(result.TasksRemoved, "HTML comments") {
		t.Errorf("expected 'HTML comments' in TasksRemoved, got %v", result.TasksRemoved)
	}
}

func TestStripHashComments(t *testing.T) {
	input := "# Python comment\ncode here\n# Another comment"
	n := NewNormalizer()
	result := n.Normalize(input)

	if strings.Contains(result.Normalized, "Python comment") {
		t.Error("# comment should be stripped")
	}
	if !containsStr(result.TasksRemoved, "# comments") {
		t.Errorf("expected '# comments' in TasksRemoved, got %v", result.TasksRemoved)
	}
}

func TestStripFunctionDefinitions(t *testing.T) {
	input := "setup code\nfunc main() {\n\tfmt.Println(\"hi\")\n}\nmore code"
	n := NewNormalizer()
	result := n.Normalize(input)

	if strings.Contains(result.Normalized, "func main()") {
		t.Error("Go func definition should be stripped")
	}
}

func TestStripPythonDef(t *testing.T) {
	input := "import os\ndef hello():\n\treturn 'hi'\n\nclass Foo:\n\tpass"
	n := NewNormalizer()
	result := n.Normalize(input)

	if strings.Contains(result.Normalized, "def hello()") {
		t.Error("Python def should be stripped")
	}
	if strings.Contains(result.Normalized, "class Foo:") {
		t.Error("Python class should be stripped")
	}
}

func TestCollapseExcessiveWhitespace(t *testing.T) {
	input := "line1\n\n\n\n\nline2"
	n := NewNormalizer()
	result := n.Normalize(input)

	// 5 newlines between line1 and line2 should collapse to 2
	if strings.Count(result.Normalized, "\n\n\n") > 0 {
		t.Errorf("excessive blank lines should be collapsed, got %q", result.Normalized)
	}
}

func TestTrimWhitespace(t *testing.T) {
	input := "   \n\n  hello world  \n\n   "
	n := NewNormalizer()
	result := n.Normalize(input)

	if result.Normalized != "hello world" {
		t.Errorf("expected 'hello world', got %q", result.Normalized)
	}
}

func TestTruncateMaxPromptLen(t *testing.T) {
	input := strings.Repeat("a", 200)
	n := NewNormalizer().WithMaxPromptLen(50)
	result := n.Normalize(input)

	if result.NormalizedLen != 50 {
		t.Errorf("expected normalizedLen=50, got %d", result.NormalizedLen)
	}
	if !containsStr(result.TasksRemoved, "truncated to maxPromptLen") {
		t.Errorf("expected 'truncated to maxPromptLen' in TasksRemoved, got %v", result.TasksRemoved)
	}
}

func TestSavingsCalculation(t *testing.T) {
	input := "hello world /* remove this large comment block that takes up lots of space */ done"
	n := NewNormalizer()
	result := n.Normalize(input)

	if result.Savings <= 0 {
		t.Errorf("expected positive savings, got %f", result.Savings)
	}
	if result.Savings > 100 {
		t.Errorf("expected savings <= 100, got %f", result.Savings)
	}
}

func TestNoCommentsDisabled(t *testing.T) {
	input := "code // comment\n/* block */"
	n := NewNormalizer().WithStripComments(false)
	result := n.Normalize(input)

	if result.Normalized != input {
		t.Errorf("with comments disabled, expected unchanged output, got %q", result.Normalized)
	}
	if len(result.TasksRemoved) > 0 {
		t.Errorf("with comments disabled, expected no removals, got %v", result.TasksRemoved)
	}
}

func TestNoFunctionsDisabled(t *testing.T) {
	input := "func main() {}"
	n := NewNormalizer().WithStripFunctions(false)
	result := n.Normalize(input)

	if result.Normalized != input {
		t.Errorf("with functions disabled, expected unchanged output, got %q", result.Normalized)
	}
}

func TestMaxLineLen(t *testing.T) {
	input := "short\n" + strings.Repeat("x", 300) + "\nalso short"
	n := NewNormalizer().WithMaxLineLen(50)
	result := n.Normalize(input)

	lines := strings.Split(result.Normalized, "\n")
	for i, line := range lines {
		if len(line) > 50 {
			t.Errorf("line %d has length %d, expected <= 50", i, len(line))
		}
	}
}

func TestMaxFileLines(t *testing.T) {
	input := "@file config.go\nline1\nline2\nline3\n\nother code"
	n := NewNormalizer().WithMaxFileLines(2)
	result := n.Normalize(input)

	// Should keep @file mention + 2 lines, then blank line ends the block
	if !strings.Contains(result.Normalized, "@file config.go") {
		t.Error("@file mention should be preserved")
	}
	if strings.Contains(result.Normalized, "line3") {
		t.Error("line3 should be truncated (maxFileLines=2)")
	}
}

func TestArrowFunctionStripping(t *testing.T) {
	input := "const handler = (req, res) => {\n\tconsole.log(req)\n}\nother code"
	n := NewNormalizer()
	result := n.Normalize(input)

	if strings.Contains(result.Normalized, "const handler") {
		t.Error("arrow function should be stripped")
	}
}

func TestComprehensiveNormalization(t *testing.T) {
	input := `/*
 * Large header comment block
 * With multiple lines
 */

// Single line comment
import "fmt"

func main() {
	fmt.Println("hello")
}



# Python style comment

def helper():
	pass


more regular code here
`
	n := NewNormalizer()
	result := n.Normalize(input)

	// Verify all stripping happened
	if strings.Contains(result.Normalized, "Large header comment") {
		t.Error("block comments should be stripped")
	}
	if strings.Contains(result.Normalized, "// Single line comment") {
		t.Error("// comments should be stripped")
	}
	if strings.Contains(result.Normalized, "func main()") {
		t.Error("func definitions should be stripped")
	}
	if strings.Contains(result.Normalized, "def helper()") {
		t.Error("def definitions should be stripped")
	}
	if strings.Contains(result.Normalized, "# Python style") {
		t.Error("# comments should be stripped")
	}
	// Should have trimmed/collapsed
	if result.Normalized == input {
		t.Error("normalized should equal original when stripping is enabled")
	}
	if result.OriginalLen <= result.NormalizedLen {
		t.Errorf("expected OriginalLen > NormalizedLen, got %d <= %d", result.OriginalLen, result.NormalizedLen)
	}
}

func TestOriginalPreserved(t *testing.T) {
	input := "hello /* world */"
	n := NewNormalizer()
	result := n.Normalize(input)

	if result.Original != input {
		t.Errorf("Original should equal input, got %q", result.Original)
	}
	if result.OriginalLen != len(input) {
		t.Errorf("OriginalLen mismatch, expected %d, got %d", len(input), result.OriginalLen)
	}
}

// containsStr checks if a slice contains a target string.
func containsStr(slice []string, target string) bool {
	for _, s := range slice {
		if s == target {
			return true
		}
	}
	return false
}
