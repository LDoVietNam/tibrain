package sandbox

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

type AuditEvent struct {
	ID          string    `json:"id"`
	Time        time.Time `json:"time"`
	Backend     string    `json:"backend"`
	Command     string    `json:"command"`
	CWD         string    `json:"cwd"`
	Network     string    `json:"network"`
	Allowed     bool      `json:"allowed"`
	ExitCode    int       `json:"exit_code"`
	DurationMS  int64     `json:"duration_ms"`
	OutputBytes int       `json:"output_bytes"`
	Error       string    `json:"error,omitempty"`
}

func WriteAudit(path string, ev AuditEvent) error {
	if path == "" {
		path = DefaultPolicy().AuditFile
	}
	target := ExpandPath(path)
	if err := os.MkdirAll(filepath.Dir(target), 0o700); err != nil {
		return err
	}
	f, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil {
		return err
	}
	defer f.Close()
	b, err := json.Marshal(ev)
	if err != nil {
		return err
	}
	_, err = f.Write(append(b, '\n'))
	return err
}

func ReadAuditTail(path string, limit int) ([]AuditEvent, error) {
	if path == "" {
		path = DefaultPolicy().AuditFile
	}
	if limit <= 0 {
		limit = 20
	}
	data, err := os.ReadFile(ExpandPath(path))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	lines := splitLines(data)
	if len(lines) > limit {
		lines = lines[len(lines)-limit:]
	}
	out := make([]AuditEvent, 0, len(lines))
	for _, line := range lines {
		if len(line) == 0 {
			continue
		}
		var ev AuditEvent
		if json.Unmarshal(line, &ev) == nil {
			out = append(out, ev)
		}
	}
	return out, nil
}

func splitLines(data []byte) [][]byte {
	var lines [][]byte
	start := 0
	for i, b := range data {
		if b == '\n' {
			lines = append(lines, data[start:i])
			start = i + 1
		}
	}
	if start < len(data) {
		lines = append(lines, data[start:])
	}
	return lines
}
