package patch

import (
	"encoding/json"
	"errors"
	"regexp"
	"strings"
)

var fencedJSONPattern = regexp.MustCompile("(?s)```(?:json)?\\s*(\\{.*?\\})\\s*```")

func ExtractOpsFile(text string) (*OpsFile, error) {
	text = strings.TrimSpace(text)
	if text == "" {
		return nil, errors.New("empty patch content")
	}
	candidates := []string{}
	if matches := fencedJSONPattern.FindAllStringSubmatch(text, -1); len(matches) > 0 {
		for _, m := range matches {
			if len(m) > 1 {
				candidates = append(candidates, strings.TrimSpace(m[1]))
			}
		}
	}
	candidates = append(candidates, strings.TrimSpace(text))
	if start := strings.Index(text, "{"); start >= 0 {
		if end := strings.LastIndex(text, "}"); end > start {
			candidates = append(candidates, strings.TrimSpace(text[start:end+1]))
		}
	}
	var lastErr error
	for _, c := range candidates {
		var ops OpsFile
		if err := json.Unmarshal([]byte(c), &ops); err == nil && len(ops.Ops) > 0 {
			return &ops, nil
		} else if err != nil {
			lastErr = err
		}
	}
	if lastErr == nil {
		lastErr = errors.New("no patch ops found in content")
	}
	return nil, lastErr
}
