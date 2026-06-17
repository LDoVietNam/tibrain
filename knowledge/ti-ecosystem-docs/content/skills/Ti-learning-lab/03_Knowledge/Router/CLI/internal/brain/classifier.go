package brain

import (
	"regexp"
	"strings"
)

// taskPatterns maps task types to regex patterns for classification.
// Ordered from most specific to most general.
var taskPatterns = []struct {
	taskType string
	pattern  *regexp.Regexp
}{
	{"debugging", regexp.MustCompile(`(?i)\b(debug|fix.*bug|why.*not.*work|error.*message|crash|stack.*trace|panic|loi|lỗi)\b`)},
	{"testing", regexp.MustCompile(`(?i)\b(test|unit.*test|write.*test|test.*case|coverage|assert|mock)\b`)},
	{"refactoring", regexp.MustCompile(`(?i)\b(refactor|improve.*code|optimize|clean.*up|restructure|simplify)\b`)},
	{"review", regexp.MustCompile(`(?i)\b(review|audit|inspect|evaluate|lint|code.*review)\b`)},
	{"coding", regexp.MustCompile(`(?i)\b(code|fix|bug|error|implement|write.*function|create.*script|program|sua|sửa|build|compile|lỗi)\b`)},
	{"explaining", regexp.MustCompile(`(?i)\b(explain|what.*is|how.*does|describe|understand|meaning|tại.*sao|vì.*sao)\b`)},
	{"translation", regexp.MustCompile(`(?i)(translat|dịch|chuyể|convert.*to)`)},
	{"summarization", regexp.MustCompile(`(?i)\b(summarize|tóm.*tắt|tl;dr|brief|overview)\b`)},
	{"writing", regexp.MustCompile(`(?i)\b(write.*email|write.*message|draft|compose|letter|documentation)\b`)},
	{"spec", regexp.MustCompile(`(?i)\b(spec|specification|requirement|design.*doc|architecture)\b`)},
	{"planning", regexp.MustCompile(`(?i)\b(plan|design.*out|architect|roadmap|strategy)\b`)},
}

// ClassifyTask classifies a user prompt into a task type.
func ClassifyTask(prompt string) string {
	prompt = strings.ToLower(prompt)

	for _, tp := range taskPatterns {
		if tp.pattern.MatchString(prompt) {
			return tp.taskType
		}
	}

	return "general"
}
