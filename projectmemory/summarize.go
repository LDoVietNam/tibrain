package projectmemory

import (
	"bufio"
	"os"
	"strings"
)

func summarize(path, kind string) FileSummary {
	b, err := os.ReadFile(path)
	if err != nil {
		return FileSummary{Path: path, Kind: kind}
	}
	return FileSummary{
		Path:     path,
		Kind:     kind,
		Exists:   true,
		Bytes:    len(b),
		Headings: headings(string(b), 16),
	}
}

func headings(text string, limit int) []string {
	var out []string
	s := bufio.NewScanner(strings.NewReader(text))
	for s.Scan() {
		line := strings.TrimSpace(s.Text())
		if strings.HasPrefix(line, "#") {
			out = append(out, line)
			if len(out) >= limit {
				break
			}
		}
	}
	return out
}
