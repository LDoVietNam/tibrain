package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/ti/cli/internal/providers"
	"github.com/ti/cli/internal/sandbox"
)

// DefaultBashTimeout là thời gian tối đa cho một lệnh bash.
const DefaultBashTimeout = 30 * time.Second

// TodoItem đại diện một task trong todo list của AI.
type TodoItem struct {
	ID         string `json:"id"`
	Content    string `json:"content"`              // imperative form: "Fix auth bug"
	ActiveForm string `json:"activeForm,omitempty"` // present continuous: "Fixing auth bug"
	Status     string `json:"status"`               // pending | in_progress | completed
}

var (
	todoMu    sync.Mutex
	todoStore []TodoItem
)

// AllDefs returns the tool definitions to send to the AI provider.
func AllDefs() []providers.ToolDef {
	return []providers.ToolDef{
		{
			Name:        "read_file",
			Description: "Read a file from the local filesystem. Results include line numbers (cat -n format). By default reads up to 2000 lines. Use offset/limit for large files. Use grep/glob for discovery — only read a file when you know the path.",
			InputSchema: providers.ToolInputSchema{
				Type: "object",
				Properties: map[string]providers.ToolPropSchema{
					"path":   {Type: "string", Description: "Absolute or relative file path to read"},
					"offset": {Type: "integer", Description: "Start reading from this line number (1-based, optional)"},
					"limit":  {Type: "integer", Description: "Maximum number of lines to read (default 2000, optional)"},
				},
				Required: []string{"path"},
			},
		},
		{
			Name:        "write_file",
			Description: "Write content to a file (creates or overwrites). PREFER edit_file for modifying existing files — it only sends the diff. Use write_file ONLY for new files or complete rewrites.",
			InputSchema: providers.ToolInputSchema{
				Type: "object",
				Properties: map[string]providers.ToolPropSchema{
					"path":    {Type: "string", Description: "File path to write"},
					"content": {Type: "string", Description: "Content to write"},
				},
				Required: []string{"path", "content"},
			},
		},
		{
			Name:        "edit_file",
			Description: "Replace occurrences of old_str with new_str in a file. IMPORTANT: You MUST call read_file at least once before editing. The edit will fail if old_str is not unique unless replace_all=true. Use replace_all=true to rename variables across the file.",
			InputSchema: providers.ToolInputSchema{
				Type: "object",
				Properties: map[string]providers.ToolPropSchema{
					"path":        {Type: "string", Description: "File path to edit"},
					"old_str":     {Type: "string", Description: "Exact string to find (must match file content exactly, including whitespace)"},
					"new_str":     {Type: "string", Description: "Replacement string"},
					"replace_all": {Type: "boolean", Description: "Replace ALL occurrences (useful for renaming variables). Default: false (fails if not unique)"},
				},
				Required: []string{"path", "old_str", "new_str"},
			},
		},
		{
			Name:        "list_dir",
			Description: "List files and directories at a path.",
			InputSchema: providers.ToolInputSchema{
				Type: "object",
				Properties: map[string]providers.ToolPropSchema{
					"path": {Type: "string", Description: "Directory path to list"},
				},
				Required: []string{"path"},
			},
		},
		{
			Name:        "glob",
			Description: "Find files matching a glob pattern. Results sorted by modification time (newest first). Use when you need to find files by name pattern. For open-ended searches requiring multiple rounds of globbing and grepping, use bash instead.",
			InputSchema: providers.ToolInputSchema{
				Type: "object",
				Properties: map[string]providers.ToolPropSchema{
					"pattern": {Type: "string", Description: "Glob pattern, e.g. src/**/*.go or **/*.ts"},
				},
				Required: []string{"pattern"},
			},
		},
		{
			Name:        "bash",
			Description: "Run a shell command and return stdout+stderr. Use for git, go build, tests, etc. Timeout: 30s. Git safety: NEVER run destructive git commands (push --force, reset --hard, clean -f) unless explicitly requested. NEVER skip hooks (--no-verify). NEVER commit unless explicitly asked.",
			InputSchema: providers.ToolInputSchema{
				Type: "object",
				Properties: map[string]providers.ToolPropSchema{
					"command": {Type: "string", Description: "Shell command to execute"},
					"timeout": {Type: "integer", Description: "Timeout in seconds (default 30, max 120)"},
					"cwd":     {Type: "string", Description: "Working directory for the command (optional, defaults to current dir)"},
					"sandbox": {Type: "boolean", Description: "Run this shell command through Ti sandbox policy."},
				},
				Required: []string{"command"},
			},
		},
		{
			Name:        "grep",
			Description: "Search for a pattern in files using ripgrep. ALWAYS use this tool for search tasks — NEVER invoke grep or rg as a bash command. Supports full regex (e.g. 'log.*Error', 'func\\s+\\w+'). Literal braces need escaping (use 'interface\\{\\}' to find 'interface{}' in Go). output_mode: 'content' (default, shows matching lines with line numbers) | 'files_with_matches' (file paths only) | 'count'.",
			InputSchema: providers.ToolInputSchema{
				Type: "object",
				Properties: map[string]providers.ToolPropSchema{
					"pattern":     {Type: "string", Description: "String or regex to search for (ripgrep syntax)"},
					"path":        {Type: "string", Description: "Directory or file to search in"},
					"glob":        {Type: "string", Description: "File glob filter, e.g. *.go or **/*.ts (optional)"},
					"output_mode": {Type: "string", Description: "content (default) | files_with_matches | count"},
					"multiline":   {Type: "boolean", Description: "Enable multiline matching for cross-line patterns (default false)"},
				},
				Required: []string{"pattern", "path"},
			},
		},
		{
			Name:        "todo_write",
			Description: "Update the task todo list for the current session. Use proactively for multi-step tasks (3+ steps). Mark tasks in_progress BEFORE starting work. Only ONE task in_progress at a time. Mark completed IMMEDIATELY after finishing — don't batch. ONLY mark completed when FULLY done (no failing tests, no partial impl).",
			InputSchema: providers.ToolInputSchema{
				Type: "object",
				Properties: map[string]providers.ToolPropSchema{
					"todos": {Type: "string", Description: "JSON array: [{\"id\":\"1\",\"content\":\"Fix auth bug\",\"activeForm\":\"Fixing auth bug\",\"status\":\"pending|in_progress|completed\"}]. Always provide both content (imperative) and activeForm (present continuous)."},
				},
				Required: []string{"todos"},
			},
		},
		{
			Name:        "todo_read",
			Description: "Read the current task todo list for this session.",
			InputSchema: providers.ToolInputSchema{
				Type:       "object",
				Properties: map[string]providers.ToolPropSchema{},
				Required:   []string{},
			},
		},
		{
			Name:        "web_fetch",
			Description: "Fetch content from a URL and return it as text. HTTP URLs are upgraded to HTTPS. Results may be truncated for large pages. Use this to read documentation, check APIs, or retrieve web content.",
			InputSchema: providers.ToolInputSchema{
				Type: "object",
				Properties: map[string]providers.ToolPropSchema{
					"url": {Type: "string", Description: "Fully-formed URL to fetch (http:// or https://)"},
				},
				Required: []string{"url"},
			},
		},
		{
			Name:        "move_file",
			Description: "Move or rename a file or directory. Works cross-platform (Windows + Unix). Use this instead of bash mv/move commands.",
			InputSchema: providers.ToolInputSchema{
				Type: "object",
				Properties: map[string]providers.ToolPropSchema{
					"source":      {Type: "string", Description: "Source path (file or directory)"},
					"destination": {Type: "string", Description: "Destination path"},
				},
				Required: []string{"source", "destination"},
			},
		},
		{
			Name:        "delete_file",
			Description: "Delete a file or directory. Use this instead of bash rm/del commands. Set recursive=true to delete non-empty directories.",
			InputSchema: providers.ToolInputSchema{
				Type: "object",
				Properties: map[string]providers.ToolPropSchema{
					"path":      {Type: "string", Description: "Path to file or directory to delete"},
					"recursive": {Type: "boolean", Description: "Delete directory and all contents recursively (default false)"},
				},
				Required: []string{"path"},
			},
		},
		{
			Name:        "session_create",
			Description: "Create a persistent shell session. The shell stays alive between calls — cd, env vars, and variables are preserved. Returns a session_id to use with session_run. PREFER this over bash when you need multiple related commands (e.g. cd into dir then run commands).",
			InputSchema: providers.ToolInputSchema{
				Type: "object",
				Properties: map[string]providers.ToolPropSchema{
					"cwd": {Type: "string", Description: "Initial working directory for the session (optional)"},
				},
				Required: []string{},
			},
		},
		{
			Name:        "session_run",
			Description: "Run a command inside an existing shell session created by session_create. State (cwd, env vars, variables) is preserved from previous commands in the same session.",
			InputSchema: providers.ToolInputSchema{
				Type: "object",
				Properties: map[string]providers.ToolPropSchema{
					"session_id": {Type: "string", Description: "Session ID returned by session_create"},
					"command":    {Type: "string", Description: "Shell command to run"},
					"timeout":    {Type: "number", Description: "Timeout in seconds (max 120, default 30)"},
				},
				Required: []string{"session_id", "command"},
			},
		},
		{
			Name:        "session_close",
			Description: "Close and terminate a shell session created by session_create. Always close sessions when done to free resources.",
			InputSchema: providers.ToolInputSchema{
				Type: "object",
				Properties: map[string]providers.ToolPropSchema{
					"session_id": {Type: "string", Description: "Session ID to close"},
				},
				Required: []string{"session_id"},
			},
		},
		{
			Name:        "session_list",
			Description: "List all active shell sessions with their IDs and creation time.",
			InputSchema: providers.ToolInputSchema{
				Type:       "object",
				Properties: map[string]providers.ToolPropSchema{},
				Required:   []string{},
			},
		},
	}
}

// Dispatch executes a tool by name with the given input map.
func Dispatch(ctx context.Context, name string, input map[string]any) (string, error) {
	str := func(key string) string {
		v, _ := input[key].(string)
		return v
	}
	switch name {
	case "read_file":
		offset := 0
		limit := 2000
		if v, ok := input["offset"].(float64); ok && v > 0 {
			offset = int(v)
		}
		if v, ok := input["limit"].(float64); ok && v > 0 {
			limit = int(v)
		}
		r, err := Read(ctx, str("path"))
		if err != nil {
			return "", err
		}
		if !r.Exists {
			return fmt.Sprintf("file not found: %s", r.Path), nil
		}
		// Apply offset/limit với line numbers (cat -n format)
		lines := strings.Split(r.Content, "\n")
		total := len(lines)
		start := 0
		if offset > 0 {
			start = offset - 1 // 1-based → 0-based
		}
		if start >= total {
			return fmt.Sprintf("(offset %d exceeds file length %d lines)", offset, total), nil
		}
		end := start + limit
		if end > total {
			end = total
		}
		var sb strings.Builder
		for i := start; i < end; i++ {
			fmt.Fprintf(&sb, "%6d\t%s\n", i+1, lines[i])
		}
		if end < total {
			fmt.Fprintf(&sb, "\n[showing lines %d-%d of %d total]", start+1, end, total)
		}
		return sb.String(), nil
	case "write_file":
		r, err := Write(ctx, str("path"), str("content"))
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("wrote %d bytes to %s", r.Bytes, r.Path), nil
	case "edit_file":
		replaceAll, _ := input["replace_all"].(bool)
		r, err := Edit(ctx, str("path"), str("old_str"), str("new_str"), replaceAll)
		if err != nil {
			return "", err
		}
		if r.Replaced == 0 {
			return "no occurrences found", nil
		}
		return fmt.Sprintf("replaced %d occurrence(s) in %s", r.Replaced, r.Path), nil
	case "list_dir":
		r, err := List(ctx, str("path"))
		if err != nil {
			return "", err
		}
		names := make([]string, 0, len(r.Files))
		for _, f := range r.Files {
			names = append(names, f.Name())
		}
		return strings.Join(names, "\n"), nil
	case "glob":
		r, err := Glob(ctx, str("pattern"))
		if err != nil {
			return "", err
		}
		if len(r.Matches) == 0 {
			return "(no files matched)", nil
		}
		return strings.Join(r.Matches, "\n"), nil
	case "grep":
		outputMode := str("output_mode")
		if outputMode == "" {
			outputMode = "content"
		}
		return Grep(ctx, str("pattern"), str("path"), str("glob"), outputMode, input["multiline"] == true)
	case "todo_write":
		var items []TodoItem
		raw := str("todos")
		if err := json.Unmarshal([]byte(raw), &items); err != nil {
			return "", fmt.Errorf("todo_write: invalid JSON: %w", err)
		}
		todoMu.Lock()
		todoStore = items
		todoMu.Unlock()
		return fmt.Sprintf("todo list updated (%d items)", len(items)), nil
	case "todo_read":
		todoMu.Lock()
		items := make([]TodoItem, len(todoStore))
		copy(items, todoStore)
		todoMu.Unlock()
		if len(items) == 0 {
			return "(no todos)", nil
		}
		b, _ := json.MarshalIndent(items, "", "  ")
		return string(b), nil
	case "bash":
		timeout := DefaultBashTimeout
		if v, ok := input["timeout"].(float64); ok && v > 0 {
			secs := v
			if secs > 120 {
				secs = 120
			}
			timeout = time.Duration(secs) * time.Second
		}
		useSandbox, _ := input["sandbox"].(bool)
		if useSandbox || os.Getenv("TI_TOOL_SANDBOX") == "1" {
			policy, err := sandbox.LoadPolicy("")
			if err != nil {
				return "", err
			}
			res, err := sandbox.Run(ctx, sandbox.RunRequest{
				Command: str("command"),
				CWD:     str("cwd"),
				Timeout: timeout,
				Policy:  policy,
			})
			if res != nil {
				return res.Output, err
			}
			return "", err
		}
		return Bash(ctx, str("command"), str("cwd"), timeout)
	case "web_fetch":
		return WebFetch(ctx, str("url"))
	case "move_file":
		return MoveFile(str("source"), str("destination"))
	case "delete_file":
		recursive, _ := input["recursive"].(bool)
		return DeleteFile(str("path"), recursive)
	case "session_create":
		return SessionCreate(str("cwd"))
	case "session_run":
		var timeoutSecs float64
		if v, ok := input["timeout"].(float64); ok {
			timeoutSecs = v
		}
		return SessionRun(ctx, str("session_id"), str("command"), timeoutSecs)
	case "session_close":
		return SessionClose(str("session_id"))
	case "session_list":
		return SessionList()
	default:
		return "", fmt.Errorf("unknown tool: %s", name)
	}
}

// Bash runs a shell command and returns combined stdout+stderr (max 8KB).
// Timeout defaults to DefaultBashTimeout nếu không truyền.
// cwd: working directory (empty = current dir).
// bashMaxOutput là giới hạn output của bash (32KB — đủ cho hầu hết output).
const bashMaxOutput = 32 * 1024

// truncateOutput giữ phần đầu và phần cuối, bỏ phần giữa.
// Giống Claude Code EndTruncatingAccumulator — ưu tiên giữ phần cuối (error messages, exit status).
func truncateOutput(s string, maxBytes int) string {
	if len(s) <= maxBytes {
		return s
	}
	// Giữ 60% đầu + 40% cuối, bỏ phần giữa
	head := maxBytes * 6 / 10
	tail := maxBytes - head
	return s[:head] + fmt.Sprintf("\n\n[... %d bytes truncated ...]\n\n", len(s)-head-tail) + s[len(s)-tail:]
}

func Bash(ctx context.Context, command string, cwd string, timeout time.Duration) (string, error) {
	if timeout <= 0 {
		timeout = DefaultBashTimeout
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.CommandContext(ctx, "cmd", "/C", command)
	} else {
		cmd = exec.CommandContext(ctx, "sh", "-c", command)
	}
	if cwd != "" {
		cmd.Dir = cwd
	}
	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf
	err := cmd.Run()

	out := truncateOutput(buf.String(), bashMaxOutput)

	// Context timeout — báo rõ cho AI biết
	if ctx.Err() == context.DeadlineExceeded {
		return out + "\n[timeout: command exceeded " + timeout.String() + "]", nil
	}

	// Bỏ qua exit code — trả output về cho AI tự phán đoán
	_ = err
	return out, nil
}

// Grep searches for pattern in path (file or directory) and returns matching lines.
// Exit code 1 (no matches) không phải lỗi — giống commandSemantics của claude-code.
// Grep searches for pattern in path using Go native regexp — cross-platform, no shell dependency.
// output_mode: "content" (default) | "files_with_matches" | "count"
func Grep(ctx context.Context, pattern, searchPath, globFilter, outputMode string, multiline bool) (string, error) {
	if multiline {
		pattern = "(?s)" + pattern
	}
	re, err := regexp.Compile(pattern)
	if err != nil {
		return "", fmt.Errorf("grep: invalid pattern %q: %w", pattern, err)
	}

	type match struct {
		file  string
		line  int
		text  string
		count int
	}

	var results []match
	totalBytes := 0
	const maxBytes = 8192

	walkErr := filepath.WalkDir(searchPath, func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		// glob filter
		if globFilter != "" {
			matched, _ := filepath.Match(globFilter, filepath.Base(p))
			if !matched {
				return nil
			}
		}
		data, readErr := os.ReadFile(p)
		if readErr != nil {
			return nil
		}
		lines := strings.Split(string(data), "\n")
		fileCount := 0
		for i, line := range lines {
			if re.MatchString(line) {
				fileCount++
				results = append(results, match{file: p, line: i + 1, text: line, count: fileCount})
			}
		}
		return nil
	})
	if walkErr != nil && walkErr != ctx.Err() {
		// single file case
		data, readErr := os.ReadFile(searchPath)
		if readErr != nil {
			return "", fmt.Errorf("grep: %w", readErr)
		}
		lines := strings.Split(string(data), "\n")
		for i, line := range lines {
			if re.MatchString(line) {
				results = append(results, match{file: searchPath, line: i + 1, text: line})
			}
		}
	}

	if len(results) == 0 {
		return "(no matches found)", nil
	}

	var sb strings.Builder
	seen := map[string]int{}
	for _, m := range results {
		seen[m.file]++
	}

	switch outputMode {
	case "files_with_matches":
		for f := range seen {
			sb.WriteString(f + "\n")
			totalBytes += len(f) + 1
			if totalBytes > maxBytes {
				sb.WriteString("[truncated]")
				break
			}
		}
	case "count":
		for f, c := range seen {
			line := fmt.Sprintf("%s:%d\n", f, c)
			sb.WriteString(line)
			totalBytes += len(line)
			if totalBytes > maxBytes {
				sb.WriteString("[truncated]")
				break
			}
		}
	default: // content
		for _, m := range results {
			line := fmt.Sprintf("%s:%d:%s\n", m.file, m.line, m.text)
			sb.WriteString(line)
			totalBytes += len(line)
			if totalBytes > maxBytes {
				sb.WriteString("[truncated]")
				break
			}
		}
	}

	return sb.String(), nil
}

type ReadResult struct {
	Path    string
	Content string
	Lines   int
	Exists  bool
}

func Read(ctx context.Context, path string) (*ReadResult, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &ReadResult{Path: path, Exists: false}, nil
		}
		return nil, err
	}
	lines := strings.Count(string(content), "\n") + 1
	return &ReadResult{
		Path:    path,
		Content: string(content),
		Lines:   lines,
		Exists:  true,
	}, nil
}

type WriteResult struct {
	Path      string
	Bytes     int
	Created   bool
	Truncated bool
}

func Write(ctx context.Context, path, content string) (*WriteResult, error) {
	existing := true
	if _, err := os.Stat(path); os.IsNotExist(err) {
		existing = false
	} else if err != nil {
		return nil, err
	}

	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return nil, err
	}

	return &WriteResult{
		Path:      path,
		Bytes:     len(content),
		Created:   !existing,
		Truncated: existing,
	}, nil
}

type EditResult struct {
	Path      string
	Content   string
	Replaced  int
	BytesDiff int
}

// normalizeQuotes chuyển curly quotes → straight quotes để match dễ hơn.
// Giống Claude Code FileEditTool/utils.ts normalizeQuotes().
func normalizeQuotes(s string) string {
	r := strings.NewReplacer(
		"\u2018", "'", // '
		"\u2019", "'", // '
		"\u201C", "\"", // "
		"\u201D", "\"", // "
	)
	return r.Replace(s)
}

func Edit(ctx context.Context, path, oldStr, newStr string, replaceAll bool) (*EditResult, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	oldContent := string(content)

	// Thử exact match trước
	count := strings.Count(oldContent, oldStr)
	actualOldStr := oldStr

	// Nếu không match exact, thử normalize curly quotes
	if count == 0 {
		normOld := normalizeQuotes(oldStr)
		normContent := normalizeQuotes(oldContent)
		count = strings.Count(normContent, normOld)
		if count > 0 {
			// Tìm actual string trong file (có curly quotes)
			// Dùng normalized để tìm vị trí, rồi extract từ original
			idx := strings.Index(normContent, normOld)
			if idx >= 0 {
				actualOldStr = oldContent[idx : idx+len(oldStr)]
			}
		}
	}

	if count == 0 {
		return &EditResult{Path: path, Content: oldContent, Replaced: 0}, nil
	}

	// Nếu không phải replace_all, yêu cầu old_str phải unique
	if !replaceAll && count > 1 {
		return nil, fmt.Errorf("old_str matches %d occurrences — use replace_all=true to replace all, or provide more context to make it unique", count)
	}

	// Replace dùng actualOldStr (giữ nguyên quote style của file)
	newContent := strings.ReplaceAll(oldContent, actualOldStr, newStr)
	if err := os.WriteFile(path, []byte(newContent), 0644); err != nil {
		return nil, err
	}

	return &EditResult{
		Path:      path,
		Content:   newContent,
		Replaced:  count,
		BytesDiff: len(newContent) - len(oldContent),
	}, nil
}

type GlobResult struct {
	Path    string
	Matches []string
}

func Glob(ctx context.Context, pattern string) (*GlobResult, error) {
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return nil, err
	}

	var results []string
	for _, m := range matches {
		if info, err := os.Stat(m); err == nil && !info.IsDir() {
			results = append(results, m)
		}
	}

	return &GlobResult{
		Path:    pattern,
		Matches: results,
	}, nil
}

type ListResult struct {
	Path  string
	Files []fs.FileInfo
	Count int
}

func List(ctx context.Context, path string) (*ListResult, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	var files []fs.FileInfo
	for _, e := range entries {
		info, err := e.Info()
		if err != nil {
			continue
		}
		files = append(files, info)
	}

	return &ListResult{
		Path:  path,
		Files: files,
		Count: len(files),
	}, nil
}

// MoveFile moves or renames a file/directory cross-platform.
func MoveFile(source, destination string) (string, error) {
	if source == "" {
		return "", fmt.Errorf("move_file: source is required")
	}
	if destination == "" {
		return "", fmt.Errorf("move_file: destination is required")
	}
	if err := os.Rename(source, destination); err != nil {
		return "", fmt.Errorf("move_file: %w", err)
	}
	return fmt.Sprintf("moved %s → %s", source, destination), nil
}

// DeleteFile deletes a file or directory.
// recursive=true dùng os.RemoveAll để xóa non-empty directories.
func DeleteFile(path string, recursive bool) (string, error) {
	if path == "" {
		return "", fmt.Errorf("delete_file: path is required")
	}
	if _, err := os.Stat(path); err != nil {
		return "", fmt.Errorf("delete_file: %w", err)
	}
	if recursive {
		if err := os.RemoveAll(path); err != nil {
			return "", fmt.Errorf("delete_file: %w", err)
		}
	} else {
		if err := os.Remove(path); err != nil {
			return "", fmt.Errorf("delete_file: directory not empty — use recursive=true: %w", err)
		}
	}
	return fmt.Sprintf("deleted %s", path), nil
}

// WebFetch fetches a URL and returns the body as text (max 100KB).
// HTTP URLs are upgraded to HTTPS automatically.
func WebFetch(ctx context.Context, rawURL string) (string, error) {
	if strings.HasPrefix(rawURL, "http://") {
		rawURL = "https://" + rawURL[7:]
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return "", fmt.Errorf("web_fetch: invalid URL: %w", err)
	}
	req.Header.Set("User-Agent", "Ti-CLI/1.0")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("web_fetch: %w", err)
	}
	defer resp.Body.Close()
	const maxBytes = 100 * 1024
	body, err := io.ReadAll(io.LimitReader(resp.Body, maxBytes))
	if err != nil {
		return "", fmt.Errorf("web_fetch: read body: %w", err)
	}
	return fmt.Sprintf("HTTP %d\n\n%s", resp.StatusCode, string(body)), nil
}

// ---------------------------------------------------------------------------
// Stateful Bash Sessions
// Giữ một shell process sống giữa các lần gọi — cd, env vars, variables
// được preserve. Mỗi session có ID riêng.
// ---------------------------------------------------------------------------

// syncBuffer wraps bytes.Buffer với mutex để thread-safe.
type syncBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (sb *syncBuffer) Write(p []byte) (n int, err error) {
	sb.mu.Lock()
	defer sb.mu.Unlock()
	return sb.buf.Write(p)
}

func (sb *syncBuffer) String() string {
	sb.mu.Lock()
	defer sb.mu.Unlock()
	return sb.buf.String()
}

func (sb *syncBuffer) Reset() {
	sb.mu.Lock()
	defer sb.mu.Unlock()
	sb.buf.Reset()
}

// bashSession giữ một shell process đang chạy.
type bashSession struct {
	id        string
	cmd       *exec.Cmd
	stdin     io.WriteCloser
	stdout    *syncBuffer
	mu        sync.Mutex
	lastUsed  time.Time
	createdAt time.Time
}

var (
	sessionsMu sync.Mutex
	sessions   = map[string]*bashSession{}
)

// SessionCreate tạo một shell session mới, trả về session ID.
func SessionCreate(cwd string) (string, error) {
	id := fmt.Sprintf("sess-%d", time.Now().UnixNano())

	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/Q", "/K", "prompt $G")
	} else {
		cmd = exec.Command("bash", "--norc", "--noprofile")
	}
	if cwd != "" {
		cmd.Dir = cwd
	}

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return "", fmt.Errorf("session_create: stdin pipe: %w", err)
	}
	var buf syncBuffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf

	if err := cmd.Start(); err != nil {
		return "", fmt.Errorf("session_create: start shell: %w", err)
	}

	now := time.Now()
	s := &bashSession{id: id, cmd: cmd, stdin: stdin, stdout: &buf, createdAt: now, lastUsed: now}
	sessionsMu.Lock()
	sessions[id] = s
	sessionsMu.Unlock()

	return id, nil
}

// SessionRun chạy một command trong session đã tạo, trả về output.
// Dùng sentinel để biết khi nào command kết thúc.
func SessionRun(ctx context.Context, sessionID, command string, timeoutSecs float64) (string, error) {
	sessionsMu.Lock()
	s, ok := sessions[sessionID]
	sessionsMu.Unlock()
	if !ok {
		return "", fmt.Errorf("session_run: session %q not found — call session_create first", sessionID)
	}

	timeout := 30 * time.Second
	if timeoutSecs > 0 && timeoutSecs <= 120 {
		timeout = time.Duration(timeoutSecs) * time.Second
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Sentinel: print unique marker sau khi command xong
	sentinel := fmt.Sprintf("__TI_DONE_%d__", time.Now().UnixNano())
	var fullCmd string
	if runtime.GOOS == "windows" {
		fullCmd = fmt.Sprintf("%s\r\necho %s\r\n", command, sentinel)
	} else {
		fullCmd = fmt.Sprintf("%s\necho %s\n", command, sentinel)
	}

	// Reset buffer trước khi chạy
	s.stdout.Reset()

	if _, err := io.WriteString(s.stdin, fullCmd); err != nil {
		return "", fmt.Errorf("session_run: write command: %w", err)
	}

	// Poll cho đến khi thấy sentinel hoặc timeout
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		select {
		case <-ctx.Done():
			return s.stdout.String(), ctx.Err()
		default:
		}
		out := s.stdout.String()
		if strings.Contains(out, sentinel) {
			// Bỏ sentinel khỏi output
			out = strings.ReplaceAll(out, sentinel, "")
			out = strings.TrimRight(out, "\r\n ")
			if len(out) > 8192 {
				out = out[:8192] + "\n[truncated]"
			}
			return out, nil
		}
		time.Sleep(50 * time.Millisecond)
	}

	out := s.stdout.String()
	if len(out) > 8192 {
		out = out[:8192] + "\n[truncated]"
	}
	return out + "\n[timeout]", nil
}

// SessionClose đóng và xóa một session.
func SessionClose(sessionID string) (string, error) {
	sessionsMu.Lock()
	s, ok := sessions[sessionID]
	if ok {
		delete(sessions, sessionID)
	}
	sessionsMu.Unlock()

	if !ok {
		return "", fmt.Errorf("session_close: session %q not found", sessionID)
	}

	_ = s.stdin.Close()
	_ = s.cmd.Process.Kill()
	return fmt.Sprintf("session %s closed", sessionID), nil
}

// SessionList trả về danh sách các session đang mở kèm thông tin.
func SessionList() (string, error) {
	sessionsMu.Lock()
	defer sessionsMu.Unlock()
	if len(sessions) == 0 {
		return "(no active sessions)", nil
	}
	var sb strings.Builder
	for id, s := range sessions {
		idle := time.Since(s.lastUsed).Round(time.Second)
		fmt.Fprintf(&sb, "%s  created=%s  idle=%s\n", id, s.createdAt.Format("15:04:05"), idle)
	}
	return strings.TrimRight(sb.String(), "\n"), nil
}

// sessionTTL là thời gian idle tối đa trước khi session bị tự động đóng.
const sessionTTL = 30 * time.Minute

// init khởi động TTL cleanup goroutine — xóa sessions idle quá 30 phút.
func init() {
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			sessionsMu.Lock()
			for id, s := range sessions {
				if time.Since(s.lastUsed) > sessionTTL {
					_ = s.stdin.Close()
					_ = s.cmd.Process.Kill()
					delete(sessions, id)
				}
			}
			sessionsMu.Unlock()
		}
	}()
}
