package tokenopt

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

type CompressionResult struct {
	Command            string   `json:"command,omitempty"`
	Kind               string   `json:"kind"`
	OriginalBytes      int      `json:"original_bytes"`
	CompressedBytes    int      `json:"compressed_bytes"`
	EstimatedTokensIn  int      `json:"estimated_tokens_in"`
	EstimatedTokensOut int      `json:"estimated_tokens_out"`
	SavedTokens        int      `json:"saved_tokens"`
	SavedPercent       float64  `json:"saved_percent"`
	Output             string   `json:"output"`
	Notes              []string `json:"notes,omitempty"`
}

type Options struct {
	Command  string
	MaxLines int
	MaxBytes int
}

var ansiRe = regexp.MustCompile(`\x1b\[[0-9;?]*[ -/]*[@-~]`)
var pathLineRe = regexp.MustCompile(`(?i)(panic|fatal|error|failed|failure|exception|traceback|assert|expected|actual|diff|--- FAIL|FAIL|FAILED|✗|×)`)

func EstimateTokens(s string) int {
	if strings.TrimSpace(s) == "" {
		return 0
	}
	// Practical approximation for CLI accounting. It intentionally avoids model-specific tokenizers.
	runes := len([]rune(s))
	words := len(strings.Fields(s))
	byChars := (runes + 3) / 4
	byWords := (words*4 + 2) / 3
	if byWords > byChars {
		return byWords
	}
	return byChars
}

func Compress(input string, opts Options) CompressionResult {
	input = ansiRe.ReplaceAllString(input, "")
	input = strings.ReplaceAll(input, "\r\n", "\n")
	input = strings.ReplaceAll(input, "\r", "\n")
	if opts.MaxLines <= 0 {
		opts.MaxLines = 160
	}
	if opts.MaxBytes <= 0 {
		opts.MaxBytes = 24000
	}
	kind := DetectKind(opts.Command, input)
	var output string
	var notes []string
	switch kind {
	case "git-diff":
		output, notes = compressGitDiff(input, opts)
	case "test-log":
		output, notes = compressTestLog(input, opts)
	case "install-log":
		output, notes = compressInstallLog(input, opts)
	case "json-log":
		output, notes = compressGeneric(input, opts)
		notes = append(notes, "JSON/log-like output compressed with head/tail and error-line retention")
	default:
		output, notes = compressGeneric(input, opts)
	}
	if len(output) > opts.MaxBytes {
		output = limitBytes(output, opts.MaxBytes)
		notes = append(notes, fmt.Sprintf("output capped at %d bytes", opts.MaxBytes))
	}
	inTok := EstimateTokens(input)
	outTok := EstimateTokens(output)
	saved := inTok - outTok
	if saved < 0 {
		saved = 0
	}
	pct := 0.0
	if inTok > 0 {
		pct = float64(saved) * 100 / float64(inTok)
	}
	return CompressionResult{
		Command: opts.Command, Kind: kind, OriginalBytes: len(input), CompressedBytes: len(output),
		EstimatedTokensIn: inTok, EstimatedTokensOut: outTok, SavedTokens: saved, SavedPercent: pct,
		Output: output, Notes: uniqueStrings(notes),
	}
}

func DetectKind(command, input string) string {
	c := strings.ToLower(command)
	if strings.Contains(c, "git diff") || strings.HasPrefix(strings.TrimSpace(input), "diff --git") {
		return "git-diff"
	}
	if strings.Contains(c, "go test") || strings.Contains(c, "pytest") || strings.Contains(c, "cargo test") || strings.Contains(c, "npm test") || strings.Contains(c, "pnpm test") || strings.Contains(input, "--- FAIL:") || strings.Contains(input, "Traceback (most recent call last)") {
		return "test-log"
	}
	if strings.Contains(c, "npm install") || strings.Contains(c, "pnpm install") || strings.Contains(c, "yarn install") || strings.Contains(c, "go mod download") || strings.Contains(c, "pip install") {
		return "install-log"
	}
	trim := strings.TrimSpace(input)
	if strings.HasPrefix(trim, "{") || strings.HasPrefix(trim, "[") {
		return "json-log"
	}
	return "generic"
}

func compressGeneric(input string, opts Options) (string, []string) {
	lines := splitLines(input)
	if len(lines) <= opts.MaxLines {
		return collapseRepeated(lines), []string{"kept all lines; collapsed adjacent repeated lines"}
	}
	headN := opts.MaxLines / 4
	tailN := opts.MaxLines / 4
	if headN < 20 {
		headN = 20
	}
	if tailN < 20 {
		tailN = 20
	}
	important := make([]string, 0, opts.MaxLines/2)
	for _, l := range lines {
		if pathLineRe.MatchString(l) {
			important = append(important, l)
		}
		if len(important) >= opts.MaxLines/2 {
			break
		}
	}
	var out []string
	out = append(out, fmt.Sprintf("[ti-tokenopt] generic output compressed: %d lines -> target %d lines", len(lines), opts.MaxLines))
	out = append(out, "", "[head]")
	out = append(out, lines[:min(headN, len(lines))]...)
	if len(important) > 0 {
		out = append(out, "", "[important/error lines]")
		out = append(out, important...)
	}
	out = append(out, "", fmt.Sprintf("[... omitted %d middle lines ...]", max(0, len(lines)-headN-tailN)))
	out = append(out, "", "[tail]")
	out = append(out, lines[max(0, len(lines)-tailN):]...)
	return collapseRepeated(out), []string{"head/tail compression", "retained error-like lines"}
}

func compressTestLog(input string, opts Options) (string, []string) {
	lines := splitLines(input)
	var important []string
	for i, l := range lines {
		if pathLineRe.MatchString(l) || strings.HasPrefix(strings.TrimSpace(l), "at ") {
			start := max(0, i-2)
			end := min(len(lines), i+4)
			important = append(important, lines[start:end]...)
		}
	}
	if len(important) == 0 {
		return compressGeneric(input, opts)
	}
	var out []string
	out = append(out, fmt.Sprintf("[ti-tokenopt] test output compressed: %d lines", len(lines)), "")
	out = append(out, "[failures / stack traces / assertions]")
	out = append(out, dedupePreserve(important)...)
	out = append(out, "", "[tail summary]")
	tail := lines[max(0, len(lines)-min(60, opts.MaxLines/3)):]
	out = append(out, tail...)
	return collapseRepeated(out), []string{"test-log mode", "retained failures, stack traces, assertions, and tail summary"}
}

func compressInstallLog(input string, opts Options) (string, []string) {
	lines := splitLines(input)
	var important []string
	for _, l := range lines {
		ll := strings.ToLower(l)
		if strings.Contains(ll, "error") || strings.Contains(ll, "warn") || strings.Contains(ll, "deprecated") || strings.Contains(ll, "vulnerab") || strings.Contains(ll, "failed") || strings.Contains(ll, "audit") {
			important = append(important, l)
		}
	}
	var out []string
	out = append(out, fmt.Sprintf("[ti-tokenopt] install output compressed: %d lines", len(lines)), "")
	if len(important) > 0 {
		out = append(out, "[warnings/errors]")
		out = append(out, dedupePreserve(important)...)
	}
	out = append(out, "", "[tail]")
	out = append(out, lines[max(0, len(lines)-min(80, opts.MaxLines)):]...)
	return collapseRepeated(out), []string{"install-log mode", "retained warnings/errors/audit lines and tail"}
}

func compressGitDiff(input string, opts Options) (string, []string) {
	lines := splitLines(input)
	var out []string
	add, del, files := 0, 0, 0
	for _, l := range lines {
		if strings.HasPrefix(l, "diff --git") {
			files++
		}
		if strings.HasPrefix(l, "+") && !strings.HasPrefix(l, "+++") {
			add++
		}
		if strings.HasPrefix(l, "-") && !strings.HasPrefix(l, "---") {
			del++
		}
	}
	out = append(out, fmt.Sprintf("[ti-tokenopt] git diff compressed: files=%d +%d -%d", files, add, del), "")
	kept := 0
	for _, l := range lines {
		trim := strings.TrimSpace(l)
		keep := strings.HasPrefix(l, "diff --git") || strings.HasPrefix(l, "index ") || strings.HasPrefix(l, "--- ") || strings.HasPrefix(l, "+++ ") || strings.HasPrefix(l, "@@") || strings.Contains(trim, "TODO") || strings.Contains(trim, "FIXME") || strings.Contains(trim, "panic") || strings.Contains(strings.ToLower(trim), "error")
		if !keep && (strings.HasPrefix(l, "+") || strings.HasPrefix(l, "-")) && kept < opts.MaxLines {
			keep = true
		}
		if keep {
			out = append(out, l)
			kept++
		}
		if kept >= opts.MaxLines {
			out = append(out, fmt.Sprintf("[... diff truncated after %d retained lines ...]", kept))
			break
		}
	}
	return collapseRepeated(out), []string{"git-diff mode", "retained file headers, hunks, and representative changed lines"}
}

func splitLines(s string) []string {
	s = strings.TrimRight(s, "\n")
	if s == "" {
		return nil
	}
	return strings.Split(s, "\n")
}

func collapseRepeated(lines []string) string {
	if len(lines) == 0 {
		return ""
	}
	var out []string
	last := lines[0]
	count := 1
	flush := func() {
		if count > 3 {
			out = append(out, last)
			out = append(out, fmt.Sprintf("[... previous line repeated %d more times ...]", count-1))
		} else {
			for i := 0; i < count; i++ {
				out = append(out, last)
			}
		}
	}
	for _, l := range lines[1:] {
		if l == last {
			count++
			continue
		}
		flush()
		last = l
		count = 1
	}
	flush()
	return strings.Join(out, "\n") + "\n"
}

func dedupePreserve(lines []string) []string {
	seen := map[string]bool{}
	out := make([]string, 0, len(lines))
	for _, l := range lines {
		key := strings.TrimSpace(l)
		if key == "" || seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, l)
	}
	return out
}

func uniqueStrings(items []string) []string {
	m := map[string]bool{}
	var out []string
	for _, x := range items {
		if x != "" && !m[x] {
			m[x] = true
			out = append(out, x)
		}
	}
	sort.Strings(out)
	return out
}

func limitBytes(s string, max int) string {
	if max <= 0 || len(s) <= max {
		return s
	}
	head := max * 7 / 10
	tail := max - head
	return s[:head] + fmt.Sprintf("\n[... %d bytes omitted by token optimizer ...]\n", len(s)-head-tail) + s[len(s)-tail:]
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func Hash(command, input string) string {
	h := sha256.Sum256([]byte(command + "\x00" + input))
	return hex.EncodeToString(h[:])
}

type CacheEntry struct {
	Key              string    `json:"key"`
	Command          string    `json:"command,omitempty"`
	Kind             string    `json:"kind"`
	CreatedAt        time.Time `json:"created_at"`
	OriginalTokens   int       `json:"original_tokens"`
	CompressedTokens int       `json:"compressed_tokens"`
	SavedTokens      int       `json:"saved_tokens"`
	SavedPercent     float64   `json:"saved_percent"`
	OutputRef        string    `json:"output_ref"`
}

type Cache struct{ Root string }

func DefaultCacheRoot() string {
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return filepath.Join(".ti", "data", "tokenopt")
	}
	return filepath.Join(home, ".ti", "data", "tokenopt")
}

func NewCache(root string) *Cache {
	if strings.TrimSpace(root) == "" {
		root = DefaultCacheRoot()
	}
	return &Cache{Root: expandPath(root)}
}

func expandPath(p string) string {
	if p == "~" || strings.HasPrefix(p, "~/") {
		if home, err := os.UserHomeDir(); err == nil && home != "" {
			if p == "~" {
				return home
			}
			return filepath.Join(home, p[2:])
		}
	}
	return os.ExpandEnv(p)
}

func (c *Cache) indexPath() string  { return filepath.Join(c.Root, "cache.jsonl") }
func (c *Cache) outputsDir() string { return filepath.Join(c.Root, "outputs") }
func (c *Cache) Ensure() error {
	if err := os.MkdirAll(c.outputsDir(), 0700); err != nil {
		return err
	}
	if _, err := os.Stat(c.indexPath()); os.IsNotExist(err) {
		return os.WriteFile(c.indexPath(), nil, 0600)
	}
	return nil
}
func (c *Cache) Put(key string, res CompressionResult) error {
	if err := c.Ensure(); err != nil {
		return err
	}
	rel := filepath.Join("outputs", key+".txt")
	if err := os.WriteFile(filepath.Join(c.Root, rel), []byte(res.Output), 0600); err != nil {
		return err
	}
	entry := CacheEntry{Key: key, Command: res.Command, Kind: res.Kind, CreatedAt: time.Now(), OriginalTokens: res.EstimatedTokensIn, CompressedTokens: res.EstimatedTokensOut, SavedTokens: res.SavedTokens, SavedPercent: res.SavedPercent, OutputRef: rel}
	f, err := os.OpenFile(c.indexPath(), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		return err
	}
	defer f.Close()
	b, _ := json.Marshal(entry)
	_, err = f.Write(append(b, '\n'))
	return err
}
func (c *Cache) Get(key string) (*CacheEntry, string, bool, error) {
	entries, err := c.List()
	if err != nil {
		return nil, "", false, err
	}
	for i := len(entries) - 1; i >= 0; i-- {
		if entries[i].Key == key {
			data, err := os.ReadFile(filepath.Join(c.Root, entries[i].OutputRef))
			return &entries[i], string(data), err == nil, err
		}
	}
	return nil, "", false, nil
}
func (c *Cache) List() ([]CacheEntry, error) {
	if err := c.Ensure(); err != nil {
		return nil, err
	}
	f, err := os.Open(c.indexPath())
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var entries []CacheEntry
	s := bufio.NewScanner(f)
	s.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for s.Scan() {
		line := strings.TrimSpace(s.Text())
		if line == "" {
			continue
		}
		var e CacheEntry
		if json.Unmarshal([]byte(line), &e) == nil {
			entries = append(entries, e)
		}
	}
	return entries, s.Err()
}
func (c *Cache) Stats() (map[string]any, error) {
	entries, err := c.List()
	if err != nil {
		return nil, err
	}
	var saved, in, out int
	byKind := map[string]int{}
	for _, e := range entries {
		saved += e.SavedTokens
		in += e.OriginalTokens
		out += e.CompressedTokens
		byKind[e.Kind]++
	}
	pct := 0.0
	if in > 0 {
		pct = float64(saved) * 100 / float64(in)
	}
	return map[string]any{"path": c.Root, "entries": len(entries), "tokens_in": in, "tokens_out": out, "tokens_saved": saved, "saved_percent": pct, "by_kind": byKind}, nil
}
func (c *Cache) Clear() error { return os.RemoveAll(c.Root) }
