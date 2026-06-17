package sandbox

import (
	"fmt"
	"path/filepath"
	"runtime"
	"strings"
)

func ValidateCommand(command string, p *Policy) error {
	if strings.TrimSpace(command) == "" {
		return fmt.Errorf("empty command")
	}
	lower := strings.ToLower(command)
	for _, denied := range p.CommandDeny {
		d := strings.ToLower(strings.TrimSpace(denied))
		if d == "" {
			continue
		}
		if strings.Contains(lower, d) {
			return fmt.Errorf("command denied by sandbox policy: %q", denied)
		}
	}
	for _, deniedPath := range p.PathDeny {
		if pathMentioned(command, deniedPath) {
			return fmt.Errorf("command mentions denied path pattern: %s", deniedPath)
		}
	}
	return nil
}

func pathMentioned(command, pattern string) bool {
	pattern = strings.TrimSpace(pattern)
	if pattern == "" {
		return false
	}
	expanded := ExpandPath(pattern)
	compactPattern := strings.Trim(pattern, "'\"")
	candidates := []string{compactPattern, expanded}
	if strings.Contains(pattern, "**") || strings.Contains(pattern, "*") || strings.Contains(pattern, "?") {
		base := strings.Split(pattern, "*")[0]
		base = strings.TrimRight(base, "/\\")
		if base != "" && strings.Contains(command, base) {
			return true
		}
		return false
	}
	for _, c := range candidates {
		if c == "" || c == "." || c == string(filepath.Separator) {
			continue
		}
		if runtime.GOOS == "windows" {
			if strings.Contains(strings.ToLower(command), strings.ToLower(c)) {
				return true
			}
		} else if strings.Contains(command, c) {
			return true
		}
	}
	return false
}

func SanitizedEnv(env []string, allow []string) []string {
	if len(env) == 0 {
		env = []string{}
	}
	allowed := map[string]bool{}
	for _, k := range allow {
		if k = strings.TrimSpace(k); k != "" {
			allowed[k] = true
		}
	}
	out := make([]string, 0, len(env))
	for _, kv := range env {
		k, _, ok := strings.Cut(kv, "=")
		if !ok {
			continue
		}
		if allowed[k] {
			out = append(out, kv)
		}
	}
	return out
}
