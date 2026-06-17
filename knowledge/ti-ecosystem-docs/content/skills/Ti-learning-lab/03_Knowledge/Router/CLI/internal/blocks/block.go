package blocks

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const (
	TypeCommand   = "command"
	TypeSandbox   = "sandbox"
	TypeTool      = "tool"
	TypeTranslate = "translate"
	TypeAgent     = "agent"
	TypeError     = "error"
)

type Block struct {
	ID         string            `json:"id"`
	SessionID  string            `json:"session_id,omitempty"`
	Type       string            `json:"type"`
	Title      string            `json:"title,omitempty"`
	Command    string            `json:"command,omitempty"`
	CWD        string            `json:"cwd,omitempty"`
	StartedAt  time.Time         `json:"started_at"`
	EndedAt    time.Time         `json:"ended_at"`
	DurationMS int64             `json:"duration_ms"`
	ExitCode   int               `json:"exit_code"`
	StdoutRef  string            `json:"stdout_ref,omitempty"`
	StderrRef  string            `json:"stderr_ref,omitempty"`
	OutputHash string            `json:"output_sha256,omitempty"`
	Metadata   map[string]string `json:"metadata,omitempty"`
}

type RecordRequest struct {
	SessionID string
	Type      string
	Title     string
	Command   string
	CWD       string
	StartedAt time.Time
	EndedAt   time.Time
	ExitCode  int
	Stdout    string
	Stderr    string
	Metadata  map[string]string
}

type Store struct {
	Root string
}

func DefaultRoot() string {
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return filepath.Join(".ti", "data", "blocks")
	}
	return filepath.Join(home, ".ti", "data", "blocks")
}

func NewStore(root string) *Store {
	if strings.TrimSpace(root) == "" {
		root = DefaultRoot()
	}
	return &Store{Root: ExpandPath(root)}
}

func ExpandPath(p string) string {
	if p == "" {
		return p
	}
	if p == "~" || strings.HasPrefix(p, "~/") || strings.HasPrefix(p, "~"+string(filepath.Separator)) {
		if home, err := os.UserHomeDir(); err == nil && home != "" {
			if p == "~" {
				return home
			}
			return filepath.Join(home, p[2:])
		}
	}
	return os.ExpandEnv(p)
}

func (s *Store) IndexPath() string    { return filepath.Join(s.Root, "blocks.jsonl") }
func (s *Store) ArtifactsDir() string { return filepath.Join(s.Root, "artifacts") }

func (s *Store) Ensure() error {
	if err := os.MkdirAll(s.ArtifactsDir(), 0o700); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(s.IndexPath()), 0o700); err != nil {
		return err
	}
	if _, err := os.Stat(s.IndexPath()); errors.Is(err, os.ErrNotExist) {
		return os.WriteFile(s.IndexPath(), nil, 0o600)
	}
	return nil
}

func (s *Store) Record(req RecordRequest) (*Block, error) {
	if err := s.Ensure(); err != nil {
		return nil, err
	}
	now := time.Now()
	if req.StartedAt.IsZero() {
		req.StartedAt = now
	}
	if req.EndedAt.IsZero() {
		req.EndedAt = now
	}
	if req.Type == "" {
		req.Type = TypeCommand
	}
	if req.CWD == "" {
		req.CWD, _ = os.Getwd()
	}
	if abs, err := filepath.Abs(req.CWD); err == nil {
		req.CWD = abs
	}
	id := newID(req.Type)
	b := &Block{
		ID:         id,
		SessionID:  req.SessionID,
		Type:       req.Type,
		Title:      req.Title,
		Command:    req.Command,
		CWD:        req.CWD,
		StartedAt:  req.StartedAt,
		EndedAt:    req.EndedAt,
		DurationMS: req.EndedAt.Sub(req.StartedAt).Milliseconds(),
		ExitCode:   req.ExitCode,
		Metadata:   req.Metadata,
	}
	if req.Stdout != "" {
		rel := filepath.Join("artifacts", id+".out")
		if err := os.WriteFile(filepath.Join(s.Root, rel), []byte(req.Stdout), 0o600); err != nil {
			return nil, err
		}
		b.StdoutRef = rel
	}
	if req.Stderr != "" {
		rel := filepath.Join("artifacts", id+".err")
		if err := os.WriteFile(filepath.Join(s.Root, rel), []byte(req.Stderr), 0o600); err != nil {
			return nil, err
		}
		b.StderrRef = rel
	}
	sum := sha256.Sum256([]byte(req.Stdout + "\n" + req.Stderr))
	b.OutputHash = hex.EncodeToString(sum[:])
	f, err := os.OpenFile(s.IndexPath(), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	line, err := json.Marshal(b)
	if err != nil {
		return nil, err
	}
	if _, err := f.Write(append(line, '\n')); err != nil {
		return nil, err
	}
	return b, nil
}

func (s *Store) List(limit int) ([]Block, error) {
	if err := s.Ensure(); err != nil {
		return nil, err
	}
	f, err := os.Open(s.IndexPath())
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	defer f.Close()
	var blocks []Block
	sc := bufio.NewScanner(f)
	// allow reasonably large metadata lines
	sc.Buffer(make([]byte, 0, 64*1024), 2*1024*1024)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}
		var b Block
		if json.Unmarshal([]byte(line), &b) == nil {
			blocks = append(blocks, b)
		}
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	sort.SliceStable(blocks, func(i, j int) bool { return blocks[i].StartedAt.After(blocks[j].StartedAt) })
	if limit > 0 && len(blocks) > limit {
		blocks = blocks[:limit]
	}
	return blocks, nil
}

func (s *Store) Get(id string) (*Block, error) {
	if strings.TrimSpace(id) == "last" {
		blocks, err := s.List(1)
		if err != nil || len(blocks) == 0 {
			return nil, err
		}
		return &blocks[0], nil
	}
	blocks, err := s.List(0)
	if err != nil {
		return nil, err
	}
	for _, b := range blocks {
		if b.ID == id || strings.HasPrefix(b.ID, id) {
			bb := b
			return &bb, nil
		}
	}
	return nil, fmt.Errorf("block not found: %s", id)
}

func (s *Store) Output(b *Block) (string, error) {
	if b == nil {
		return "", errors.New("nil block")
	}
	var parts []string
	if b.StdoutRef != "" {
		data, err := os.ReadFile(filepath.Join(s.Root, b.StdoutRef))
		if err != nil {
			return "", err
		}
		parts = append(parts, string(data))
	}
	if b.StderrRef != "" {
		data, err := os.ReadFile(filepath.Join(s.Root, b.StderrRef))
		if err != nil {
			return "", err
		}
		if len(parts) > 0 && !strings.HasSuffix(parts[len(parts)-1], "\n") {
			parts[len(parts)-1] += "\n"
		}
		parts = append(parts, string(data))
	}
	return strings.Join(parts, ""), nil
}

func (s *Store) ExportMarkdown(b *Block) (string, error) {
	out, err := s.Output(b)
	if err != nil {
		return "", err
	}
	var sb strings.Builder
	fmt.Fprintf(&sb, "# Ti Block %s\n\n", b.ID)
	fmt.Fprintf(&sb, "- Type: `%s`\n", b.Type)
	if b.Command != "" {
		fmt.Fprintf(&sb, "- Command: `%s`\n", b.Command)
	}
	if b.CWD != "" {
		fmt.Fprintf(&sb, "- CWD: `%s`\n", b.CWD)
	}
	fmt.Fprintf(&sb, "- Started: `%s`\n", b.StartedAt.Format(time.RFC3339))
	fmt.Fprintf(&sb, "- Duration: `%dms`\n", b.DurationMS)
	fmt.Fprintf(&sb, "- Exit code: `%d`\n\n", b.ExitCode)
	if out != "" {
		sb.WriteString("## Output\n\n```text\n")
		sb.WriteString(out)
		if !strings.HasSuffix(out, "\n") {
			sb.WriteByte('\n')
		}
		sb.WriteString("```\n")
	}
	return sb.String(), nil
}

func ExplainHeuristic(b *Block, output string) string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "Block %s is a %s block", b.ID, b.Type)
	if b.Command != "" {
		fmt.Fprintf(&sb, " for `%s`", b.Command)
	}
	fmt.Fprintf(&sb, ". It exited with code %d after %dms.\n", b.ExitCode, b.DurationMS)
	lower := strings.ToLower(output)
	switch {
	case b.ExitCode == 0:
		sb.WriteString("Status: successful.\n")
	case strings.Contains(lower, "permission denied"):
		sb.WriteString("Likely issue: permission denied. Check file permissions, sandbox policy, or required credentials.\n")
	case strings.Contains(lower, "not found") || strings.Contains(lower, "command not found"):
		sb.WriteString("Likely issue: missing command, file, or dependency. Check PATH and install steps.\n")
	case strings.Contains(lower, "timeout"):
		sb.WriteString("Likely issue: command exceeded timeout. Increase timeout or narrow the task.\n")
	case strings.Contains(lower, "failed") || strings.Contains(lower, "error"):
		sb.WriteString("Likely issue: command produced errors. Review the tail of the output and rerun with a focused command.\n")
	default:
		sb.WriteString("Status: non-zero exit or warning; inspect output for details.\n")
	}
	tail := tailString(output, 1200)
	if tail != "" {
		sb.WriteString("\nOutput tail:\n```text\n")
		sb.WriteString(tail)
		if !strings.HasSuffix(tail, "\n") {
			sb.WriteByte('\n')
		}
		sb.WriteString("```\n")
	}
	return sb.String()
}

func tailString(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[len(s)-max:]
}

func (s *Store) Search(query string, limit int) ([]Block, error) {
	items, err := s.List(0)
	if err != nil {
		return nil, err
	}
	query = strings.ToLower(strings.TrimSpace(query))
	if query == "" {
		if limit > 0 && len(items) > limit {
			return items[:limit], nil
		}
		return items, nil
	}
	var out []Block
	for _, b := range items {
		hay := strings.ToLower(strings.Join([]string{b.ID, b.Type, b.Title, b.Command, b.CWD}, "\n"))
		if strings.Contains(hay, query) {
			out = append(out, b)
			if limit > 0 && len(out) >= limit {
				return out, nil
			}
			continue
		}
		content, _ := s.Output(&b)
		if strings.Contains(strings.ToLower(content), query) {
			out = append(out, b)
			if limit > 0 && len(out) >= limit {
				return out, nil
			}
		}
	}
	return out, nil
}

func (s *Store) Stats() (map[string]any, error) {
	items, err := s.List(0)
	if err != nil {
		return nil, err
	}
	byType := map[string]int{}
	failures := 0
	for _, b := range items {
		byType[b.Type]++
		if b.ExitCode != 0 {
			failures++
		}
	}
	return map[string]any{"root": s.Root, "count": len(items), "failures": failures, "by_type": byType}, nil
}

func newID(kind string) string {
	clean := strings.ToLower(kind)
	if clean == "" {
		clean = "block"
	}
	clean = strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			return r
		}
		return '-'
	}, clean)
	return fmt.Sprintf("blk-%s-%d", clean, time.Now().UnixNano())
}
