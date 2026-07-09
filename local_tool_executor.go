package main

import (
	"bufio"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

func (s *Server) executeLocalTool(ctx context.Context, name string, params map[string]interface{}) (map[string]interface{}, error) {
	switch name {
	case "read_file":
		return executeReadFile(params)
	case "write_file":
		return executeWriteFile(params)
	case "edit_file":
		return executeEditFile(params)
	case "delete_file":
		return executeDeleteFile(params)
	case "list_files":
		return executeListFiles(params)
	case "list_directory":
		return executeListFiles(params)
	case "file_info":
		return executeFileInfo(params)
	case "get_file_info":
		return executeFileInfo(params)
	case "create_directory":
		return executeCreateDirectory(params)
	case "file_exists":
		return executeFileExists(params)
	case "walk_directory":
		return executeWalkDirectory(params)
	case "copy_file":
		return executeCopyFile(params)
	case "move_file":
		return executeMoveFile(params)
	case "grep_search":
		return executeGrepSearch(params)
	case "find_files":
		return executeFindFiles(params)
	case "run_command":
		return executeRunCommand(ctx, params)
	case "exec":
		return executeExec(ctx, params)
	case "exec_status":
		return executeExecStatus(params)
	case "exec_cancel":
		return executeExecCancel(params)
	case "git_status":
		return executeGitStatus(ctx, params)
	case "process_list":
		return executeProcessList(ctx, params)
	case "process_kill":
		return executeProcessKill(ctx, params)
	case "go_build", "go_build_cmd":
		return executeGoCommand(ctx, "build", params)
	case "go_test":
		return executeGoCommand(ctx, "test", params)
	case "npm_install":
		return executeNpmInstall(ctx, params)
	default:
		return nil, fmt.Errorf("tool %q is not handled by the local executor", name)
	}
}

func executeReadFile(params map[string]interface{}) (map[string]interface{}, error) {
	path, err := resolveToolPath(toolStringParam(params, "file_path", "path", "file"))
	if err != nil {
		return nil, err
	}

	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}

	result := map[string]interface{}{
		"path":         path,
		"bytes_read":   len(raw),
		"content":      string(raw),
		"content_hash": hashBytes(raw),
	}

	lineOffset := toolIntParam(params, 0, "offset", "line_offset")
	lineLimit := toolIntParam(params, 0, "limit", "line_limit")
	if lineOffset > 0 || lineLimit > 0 {
		lines := splitLines(string(raw))
		start := lineOffset
		if start > 0 {
			start--
		}
		if start < 0 {
			start = 0
		}
		if start > len(lines) {
			start = len(lines)
		}
		end := len(lines)
		if lineLimit > 0 && start+lineLimit < end {
			end = start + lineLimit
		}
		result["content"] = strings.Join(lines[start:end], "\n")
		result["line_offset"] = start
		result["line_limit"] = lineLimit
		result["total_lines"] = len(lines)
	}

	return result, nil
}

func executeWriteFile(params map[string]interface{}) (map[string]interface{}, error) {
	path, err := resolveToolPath(toolStringParam(params, "file_path", "path", "file"))
	if err != nil {
		return nil, err
	}

	content := toolStringParam(params, "content", "text", "data")
	if err := ensureParentDir(path); err != nil {
		return nil, err
	}

	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return nil, fmt.Errorf("write file: %w", err)
	}

	data := []byte(content)
	return map[string]interface{}{
		"path":          path,
		"bytes_written": len(data),
		"content_hash":  hashBytes(data),
		"stdout":        "",
	}, nil
}

func executeEditFile(params map[string]interface{}) (map[string]interface{}, error) {
	path, err := resolveToolPath(toolStringParam(params, "file_path", "path", "file"))
	if err != nil {
		return nil, err
	}

	oldString := toolStringParam(params, "old_string", "old", "search")
	newString := toolStringParam(params, "new_string", "new", "replace")
	if oldString == "" {
		return nil, fmt.Errorf("old_string is required")
	}

	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read file before edit: %w", err)
	}

	content := string(raw)
	if !strings.Contains(content, oldString) {
		return nil, fmt.Errorf("old_string not found in file")
	}

	updated := strings.Replace(content, oldString, newString, 1)
	if err := os.WriteFile(path, []byte(updated), 0o644); err != nil {
		return nil, fmt.Errorf("write edited file: %w", err)
	}

	return map[string]interface{}{
		"path":          path,
		"replacements":  1,
		"bytes_written": len(updated),
		"content_hash":  hashBytes([]byte(updated)),
	}, nil
}

func executeDeleteFile(params map[string]interface{}) (map[string]interface{}, error) {
	path, err := resolveToolPath(toolStringParam(params, "file_path", "path", "file"))
	if err != nil {
		return nil, err
	}

	if err := os.Remove(path); err != nil {
		return nil, fmt.Errorf("delete file: %w", err)
	}

	return map[string]interface{}{
		"path":    path,
		"deleted": true,
	}, nil
}

func executeCreateDirectory(params map[string]interface{}) (map[string]interface{}, error) {
	dir, err := resolveToolPath(toolStringParam(params, "dir_path", "dir", "path"))
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("create directory: %w", err)
	}
	return map[string]interface{}{
		"path":    dir,
		"created": true,
		"exists":  true,
	}, nil
}

func executeFileExists(params map[string]interface{}) (map[string]interface{}, error) {
	path, err := resolveToolPath(toolStringParam(params, "file_path", "path", "file"))
	if err != nil {
		return nil, err
	}
	if _, statErr := os.Stat(path); statErr == nil {
		return map[string]interface{}{
			"path":   path,
			"exists": true,
		}, nil
	} else if os.IsNotExist(statErr) {
		return map[string]interface{}{
			"path":   path,
			"exists": false,
		}, nil
	} else {
		return nil, fmt.Errorf("file exists check: %w", statErr)
	}
}

func executeWalkDirectory(params map[string]interface{}) (map[string]interface{}, error) {
	root, err := resolveToolPath(toolStringParam(params, "dir_path", "dir", "path"))
	if err != nil {
		return nil, err
	}

	maxDepth := toolIntParam(params, 0, "max_depth", "depth")
	maxEntries := toolIntParam(params, 0, "max_entries", "limit")

	entries := make([]map[string]interface{}, 0)
	truncated := false
	err = filepath.WalkDir(root, func(currentPath string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		rel, relErr := filepath.Rel(root, currentPath)
		if relErr != nil {
			return relErr
		}
		if rel == "." {
			return nil
		}

		depth := 1
		if rel != "." {
			depth = strings.Count(rel, string(filepath.Separator)) + 1
		}
		if maxDepth > 0 && depth > maxDepth {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		item := map[string]interface{}{
			"path":    currentPath,
			"name":    d.Name(),
			"is_dir":  d.IsDir(),
			"depth":   depth,
			"symlink": d.Type()&fs.ModeSymlink != 0,
		}
		if info, infoErr := d.Info(); infoErr == nil {
			item["size"] = info.Size()
			item["mod_time"] = info.ModTime().UTC().Format(timeLayout)
		}
		entries = append(entries, item)

		if maxEntries > 0 && len(entries) >= maxEntries {
			truncated = true
			return filepath.SkipAll
		}
		return nil
	})
	if err != nil && !errorsIsSkipAll(err) {
		return nil, fmt.Errorf("walk directory: %w", err)
	}

	return map[string]interface{}{
		"root":      root,
		"entries":   entries,
		"count":     len(entries),
		"truncated": truncated,
	}, nil
}

func executeListFiles(params map[string]interface{}) (map[string]interface{}, error) {
	dir, err := resolveToolPath(toolStringParam(params, "dir_path", "dir", "path"))
	if err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("list directory: %w", err)
	}

	sort.Slice(entries, func(i, j int) bool {
		return strings.ToLower(entries[i].Name()) < strings.ToLower(entries[j].Name())
	})

	resultEntries := make([]map[string]interface{}, 0, len(entries))
	for _, entry := range entries {
		item := map[string]interface{}{
			"name":    entry.Name(),
			"is_dir":  entry.IsDir(),
			"path":    filepath.Join(dir, entry.Name()),
			"mode":    entry.Type().String(),
			"symlink": entry.Type()&fs.ModeSymlink != 0,
		}
		if info, statErr := entry.Info(); statErr == nil {
			item["size"] = info.Size()
			item["mod_time"] = info.ModTime().UTC().Format(timeLayout)
		}
		resultEntries = append(resultEntries, item)
	}

	return map[string]interface{}{
		"dir":     dir,
		"entries": resultEntries,
		"count":   len(resultEntries),
	}, nil
}

func executeFileInfo(params map[string]interface{}) (map[string]interface{}, error) {
	path, err := resolveToolPath(toolStringParam(params, "file_path", "path", "file"))
	if err != nil {
		return nil, err
	}

	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("file info: %w", err)
	}

	return map[string]interface{}{
		"path":      path,
		"name":      info.Name(),
		"size":      info.Size(),
		"is_dir":    info.IsDir(),
		"mode":      info.Mode().String(),
		"mod_time":  info.ModTime().UTC().Format(timeLayout),
		"exists":    true,
		"read_only": info.Mode().Perm()&0o222 == 0,
	}, nil
}

func executeCopyFile(params map[string]interface{}) (map[string]interface{}, error) {
	src, err := resolveToolPath(toolStringParam(params, "src", "source"))
	if err != nil {
		return nil, err
	}
	dst, err := resolveToolPath(toolStringParam(params, "dst", "dest", "destination"))
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(src)
	if err != nil {
		return nil, fmt.Errorf("copy source: %w", err)
	}
	if err := ensureParentDir(dst); err != nil {
		return nil, err
	}
	if err := os.WriteFile(dst, data, 0o644); err != nil {
		return nil, fmt.Errorf("copy destination: %w", err)
	}

	return map[string]interface{}{
		"src":           src,
		"dst":           dst,
		"bytes_copied":  len(data),
		"content_hash":  hashBytes(data),
		"overwrite":     true,
		"preserve_mode": false,
	}, nil
}

func executeMoveFile(params map[string]interface{}) (map[string]interface{}, error) {
	src, err := resolveToolPath(toolStringParam(params, "src", "source"))
	if err != nil {
		return nil, err
	}
	dst, err := resolveToolPath(toolStringParam(params, "dst", "dest", "destination"))
	if err != nil {
		return nil, err
	}

	if err := ensureParentDir(dst); err != nil {
		return nil, err
	}
	if err := os.Rename(src, dst); err == nil {
		return map[string]interface{}{
			"src":    src,
			"dst":    dst,
			"moved":  true,
			"method": "rename",
		}, nil
	}

	data, err := os.ReadFile(src)
	if err != nil {
		return nil, fmt.Errorf("move source: %w", err)
	}
	if err := os.WriteFile(dst, data, 0o644); err != nil {
		return nil, fmt.Errorf("move destination: %w", err)
	}
	if err := os.Remove(src); err != nil {
		return nil, fmt.Errorf("remove source after copy: %w", err)
	}

	return map[string]interface{}{
		"src":          src,
		"dst":          dst,
		"moved":        true,
		"method":       "copy-delete",
		"bytes_moved":  len(data),
		"content_hash": hashBytes(data),
	}, nil
}

func executeGrepSearch(params map[string]interface{}) (map[string]interface{}, error) {
	pattern := toolStringParam(params, "pattern", "query")
	if pattern == "" {
		return nil, fmt.Errorf("pattern is required")
	}
	dir, err := resolveToolPath(toolStringParam(params, "dir", "path"))
	if err != nil {
		return nil, err
	}

	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, fmt.Errorf("compile pattern: %w", err)
	}

	matches := make([]map[string]interface{}, 0, 64)
	truncated := false
	err = filepath.WalkDir(dir, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return nil
		}
		if len(matches) >= 200 {
			truncated = true
			return filepath.SkipAll
		}
		if d.IsDir() {
			return nil
		}
		file, openErr := os.Open(path)
		if openErr != nil {
			return nil
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		scanner.Buffer(make([]byte, 64*1024), 1024*1024)
		lineNumber := 0
		for scanner.Scan() {
			lineNumber++
			line := scanner.Text()
			if re.MatchString(line) {
				matches = append(matches, map[string]interface{}{
					"path":        path,
					"line_number": lineNumber,
					"line":        line,
				})
				if len(matches) >= 200 {
					truncated = true
					return filepath.SkipAll
				}
			}
		}
		return nil
	})
	if err != nil && !errorsIsSkipAll(err) {
		return nil, err
	}

	return map[string]interface{}{
		"dir":       dir,
		"pattern":   pattern,
		"matches":   matches,
		"count":     len(matches),
		"truncated": truncated,
	}, nil
}

func executeFindFiles(params map[string]interface{}) (map[string]interface{}, error) {
	pattern := toolStringParam(params, "pattern", "glob")
	if pattern == "" {
		return nil, fmt.Errorf("pattern is required")
	}
	dir, err := resolveToolPath(toolStringParam(params, "dir", "path"))
	if err != nil {
		return nil, err
	}

	matches := make([]map[string]interface{}, 0, 64)
	truncated := false
	err = filepath.WalkDir(dir, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return nil
		}
		if len(matches) >= 200 {
			truncated = true
			return filepath.SkipAll
		}

		relPath, relErr := filepath.Rel(dir, path)
		if relErr != nil {
			relPath = path
		}

		isMatch := false
		if ok, _ := filepath.Match(pattern, filepath.Base(path)); ok {
			isMatch = true
		}
		if !isMatch {
			if ok, _ := filepath.Match(pattern, relPath); ok {
				isMatch = true
			}
		}
		if isMatch {
			matches = append(matches, map[string]interface{}{
				"path":     path,
				"rel_path": relPath,
				"is_dir":   d.IsDir(),
			})
		}
		return nil
	})
	if err != nil && !errorsIsSkipAll(err) {
		return nil, err
	}

	return map[string]interface{}{
		"dir":       dir,
		"pattern":   pattern,
		"matches":   matches,
		"count":     len(matches),
		"truncated": truncated,
	}, nil
}

func executeRunCommand(ctx context.Context, params map[string]interface{}) (map[string]interface{}, error) {
	command := toolStringParam(params, "command")
	if command == "" {
		return nil, fmt.Errorf("command is required")
	}
	workingDir := toolStringParam(params, "dir", "cwd", "working_dir")
	if workingDir == "" {
		workingDir = getTiBrainDir()
	} else {
		var err error
		workingDir, err = resolveToolPath(workingDir)
		if err != nil {
			return nil, err
		}
	}

	return runShellCommand(ctx, workingDir, command)
}

func executeExec(ctx context.Context, params map[string]interface{}) (map[string]interface{}, error) {
	command := toolStringParam(params, "command")
	if command == "" {
		return nil, fmt.Errorf("command is required")
	}
	workingDir := toolStringParam(params, "working_dir", "workingDir", "cwd", "dir", "path")
	if workingDir != "" {
		var err error
		workingDir, err = resolveToolPath(workingDir)
		if err != nil {
			return nil, err
		}
	}
	waitForCompletion := true
	if raw, ok := params["wait"]; ok {
		if value, ok := raw.(bool); ok {
			waitForCompletion = value
		} else if value, ok := raw.(string); ok {
			waitForCompletion = strings.EqualFold(strings.TrimSpace(value), "true")
		}
	}
	timeoutMs := toolIntParam(params, 0, "timeout_ms", "timeoutMs")

	job, err := startTrackedShellCommand(command, workingDir, timeoutMs)
	if err != nil {
		return nil, err
	}
	if !waitForCompletion {
		return job.snapshot(), nil
	}

	return waitForTrackedJob(ctx, job)
}

func executeExecStatus(params map[string]interface{}) (map[string]interface{}, error) {
	execID := toolStringParam(params, "exec_id", "execId", "id")
	if execID == "" {
		return nil, fmt.Errorf("exec_id is required")
	}
	snapshot, ok := globalExecJobStore.snapshot(execID)
	if !ok {
		return nil, fmt.Errorf("exec job %q not found", execID)
	}
	return snapshot, nil
}

func executeExecCancel(params map[string]interface{}) (map[string]interface{}, error) {
	execID := toolStringParam(params, "exec_id", "execId", "id")
	if execID == "" {
		return nil, fmt.Errorf("exec_id is required")
	}
	job, ok := globalExecJobStore.get(execID)
	if !ok {
		return nil, fmt.Errorf("exec job %q not found", execID)
	}
	return cancelTrackedJob(job), nil
}

func executeGitStatus(ctx context.Context, params map[string]interface{}) (map[string]interface{}, error) {
	dir := toolStringParam(params, "dir", "path", "cwd")
	if dir == "" {
		dir = getTiBrainDir()
	} else {
		var err error
		dir, err = resolveToolPath(dir)
		if err != nil {
			return nil, err
		}
	}

	return runProcess(ctx, dir, "git", "status", "--short", "--branch")
}

func executeProcessList(ctx context.Context, params map[string]interface{}) (map[string]interface{}, error) {
	filter := strings.TrimSpace(toolStringParam(params, "filter", "query"))
	result, err := runProcess(ctx, "", "tasklist")
	if err != nil {
		return nil, err
	}

	if filter == "" {
		return result, nil
	}

	lines := strings.Split(result["stdout"].(string), "\n")
	filtered := make([]string, 0, len(lines))
	for _, line := range lines {
		if strings.Contains(strings.ToLower(line), strings.ToLower(filter)) {
			filtered = append(filtered, line)
		}
	}
	result["stdout"] = strings.Join(filtered, "\n")
	result["count"] = len(filtered)
	return result, nil
}

func executeProcessKill(ctx context.Context, params map[string]interface{}) (map[string]interface{}, error) {
	pid := toolIntParam(params, 0, "pid")
	if pid <= 0 {
		return nil, fmt.Errorf("pid is required")
	}

	return runProcess(ctx, "", "taskkill", "/PID", strconv.Itoa(pid), "/F")
}

func executeGoCommand(ctx context.Context, subcommand string, params map[string]interface{}) (map[string]interface{}, error) {
	dir := toolStringParam(params, "dir", "path", "cwd")
	if dir == "" {
		dir = getTiBrainDir()
	} else {
		var err error
		dir, err = resolveToolPath(dir)
		if err != nil {
			return nil, err
		}
	}

	task := strings.TrimSpace(toolStringParam(params, "task", "args"))
	args := []string{subcommand}
	if task != "" {
		args = append(args, strings.Fields(task)...)
	} else {
		args = append(args, "./...")
	}

	return runProcess(ctx, dir, "go", args...)
}

func executeNpmInstall(ctx context.Context, params map[string]interface{}) (map[string]interface{}, error) {
	dir := toolStringParam(params, "dir", "path", "cwd")
	if dir == "" {
		dir = getTiBrainDir()
	} else {
		var err error
		dir, err = resolveToolPath(dir)
		if err != nil {
			return nil, err
		}
	}

	return runProcess(ctx, dir, "npm", "install")
}

func runShellCommand(ctx context.Context, workingDir, command string) (map[string]interface{}, error) {
	return runProcess(ctx, workingDir, resolveShellCommand("cmd.exe"), "/C", command)
}

func runProcess(ctx context.Context, workingDir, command string, args ...string) (map[string]interface{}, error) {
	start := time.Now()
	cmd := exec.CommandContext(ctx, command, args...)
	if workingDir != "" {
		cmd.Dir = workingDir
	}

	var stdoutBuf bytes.Buffer
	var stderrBuf bytes.Buffer
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf

	err := cmd.Run()
	stdout := stdoutBuf.String()
	stderr := stderrBuf.String()
	exitCode := 0
	if err != nil {
		exitCode = exitCodeFromErr(err)
	}

	result := map[string]interface{}{
		"command":     command,
		"args":        args,
		"workingDir":  workingDir,
		"working_dir": workingDir,
		"stdout":      stdout,
		"stderr":      stderr,
		"exitCode":    exitCode,
		"exit_code":   exitCode,
		"durationMs":  time.Since(start).Milliseconds(),
		"status":      "completed",
	}
	if err != nil {
		result["status"] = "failed"
		result["error"] = err.Error()
		if strings.TrimSpace(stderr) == "" {
			result["error"] = fmt.Sprintf("%s %s failed: %v", command, strings.Join(args, " "), err)
		} else {
			result["error"] = fmt.Sprintf("%s %s failed: %v: %s", command, strings.Join(args, " "), err, strings.TrimSpace(stderr))
		}
	}

	return result, nil
}

func toolStringParam(params map[string]interface{}, keys ...string) string {
	for _, key := range keys {
		if raw, ok := params[key]; ok {
			switch value := raw.(type) {
			case string:
				return strings.TrimSpace(value)
			case fmt.Stringer:
				return strings.TrimSpace(value.String())
			case []byte:
				return strings.TrimSpace(string(value))
			}
		}
	}
	return ""
}

func toolBoolParam(params map[string]interface{}, defaultValue bool, keys ...string) bool {
	for _, key := range keys {
		raw, ok := params[key]
		if !ok {
			continue
		}
		switch value := raw.(type) {
		case bool:
			return value
		case string:
			switch strings.ToLower(strings.TrimSpace(value)) {
			case "true", "1", "yes", "y", "on":
				return true
			case "false", "0", "no", "n", "off":
				return false
			}
		}
	}
	return defaultValue
}

func toolIntParam(params map[string]interface{}, defaultValue int, keys ...string) int {
	for _, key := range keys {
		raw, ok := params[key]
		if !ok {
			continue
		}
		switch value := raw.(type) {
		case int:
			return value
		case int8:
			return int(value)
		case int16:
			return int(value)
		case int32:
			return int(value)
		case int64:
			return int(value)
		case float32:
			return int(value)
		case float64:
			return int(value)
		case json.Number:
			if parsed, err := strconv.Atoi(value.String()); err == nil {
				return parsed
			}
		case string:
			if parsed, err := strconv.Atoi(strings.TrimSpace(value)); err == nil {
				return parsed
			}
		}
	}
	return defaultValue
}

func resolveToolPath(raw string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", fmt.Errorf("path is required")
	}

	clean := filepath.Clean(trimmed)
	if filepath.IsAbs(clean) {
		if err := validateToolPathAllowed(clean); err != nil {
			return "", err
		}
		return clean, nil
	}

	joined := filepath.Join(getTiBrainDir(), clean)
	if err := validateToolPathAllowed(joined); err != nil {
		return "", err
	}
	return joined, nil
}

func ensureParentDir(path string) error {
	parent := filepath.Dir(path)
	if parent == "." || parent == "" {
		return nil
	}
	if err := validateToolPathAllowed(parent); err != nil {
		return err
	}
	if err := os.MkdirAll(parent, 0o755); err != nil {
		return fmt.Errorf("create parent directory: %w", err)
	}
	return nil
}

func splitLines(content string) []string {
	normalized := strings.ReplaceAll(content, "\r\n", "\n")
	return strings.Split(normalized, "\n")
}

func hashBytes(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func exitCodeFromErr(err error) int {
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		return exitErr.ExitCode()
	}
	return -1
}

func errorsIsSkipAll(err error) bool {
	return err == filepath.SkipAll
}

const timeLayout = "2006-01-02T15:04:05Z07:00"
