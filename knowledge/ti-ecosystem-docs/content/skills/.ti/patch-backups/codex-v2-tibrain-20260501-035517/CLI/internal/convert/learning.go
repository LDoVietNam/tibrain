package convert

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type LearnOptions struct {
	Root       string
	Beads      bool
	BeadsBin   string
	SourcePath string
	OutDir     string
	BuildLog   string
	TestOK     bool
	BuildOK    bool
}

type LearnRecord struct {
	Time       time.Time    `json:"time"`
	IRID       string       `json:"ir_id"`
	Name       string       `json:"name"`
	Kind       string       `json:"kind"`
	Source     SourceMeta   `json:"source"`
	Learning   LearningMeta `json:"learning"`
	SourcePath string       `json:"source_path,omitempty"`
	OutDir     string       `json:"out_dir,omitempty"`
	BuildOK    bool         `json:"build_ok"`
	TestOK     bool         `json:"test_ok"`
	BuildLog   string       `json:"build_log,omitempty"`
}

func SaveLearning(ir *IR, opts LearnOptions) (*LearnRecord, error) {
	if ir == nil {
		return nil, fmt.Errorf("nil IR")
	}
	root := strings.TrimSpace(opts.Root)
	if root == "" {
		root = "."
	}
	rec := &LearnRecord{Time: time.Now().UTC(), IRID: ir.ID, Name: ir.Name, Kind: ir.Kind, Source: ir.Source, Learning: ir.Learning, SourcePath: opts.SourcePath, OutDir: opts.OutDir, BuildOK: opts.BuildOK, TestOK: opts.TestOK, BuildLog: opts.BuildLog}
	dir := filepath.Join(root, ".ti", "memory", "convert")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}
	if err := appendJSONL(filepath.Join(dir, "conversions.jsonl"), rec); err != nil {
		return nil, err
	}
	pattern := map[string]any{"time": rec.Time, "ir_id": rec.IRID, "keys": rec.Learning.PatternKeys, "kind": rec.Kind, "risk": rec.Learning.RiskScore, "model_role": rec.Learning.RecommendedModel, "success": rec.BuildOK || rec.TestOK}
	if err := appendJSONL(filepath.Join(dir, "patterns.jsonl"), pattern); err != nil {
		return nil, err
	}
	if opts.BuildLog != "" && (!opts.BuildOK || !opts.TestOK) {
		failure := map[string]any{"time": rec.Time, "ir_id": rec.IRID, "name": rec.Name, "build_log": rec.BuildLog, "hint": RepairHint(opts.BuildLog)}
		_ = appendJSONL(filepath.Join(dir, "failures.jsonl"), failure)
	}
	if opts.Beads {
		_ = logLearningToBeads(rec, opts)
	}
	return rec, nil
}

func LearningStats(root string) (map[string]any, error) {
	if strings.TrimSpace(root) == "" {
		root = "."
	}
	path := filepath.Join(root, ".ti", "memory", "convert", "conversions.jsonl")
	data, err := os.ReadFile(path)
	if err != nil {
		return map[string]any{"records": 0, "path": path}, nil
	}
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	byKind := map[string]int{}
	byModel := map[string]int{}
	success := 0
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		var rec LearnRecord
		if json.Unmarshal([]byte(line), &rec) == nil {
			byKind[rec.Kind]++
			byModel[rec.Learning.RecommendedModel]++
			if rec.BuildOK || rec.TestOK {
				success++
			}
		}
	}
	return map[string]any{"records": len(lines), "success": success, "by_kind": byKind, "by_model_role": byModel, "path": path}, nil
}

func RepairHint(log string) string {
	q := strings.ToLower(log)
	switch {
	case strings.Contains(q, "undefined:"):
		return "undefined symbol: check SDK import, generated package name, or pluginapi version"
	case strings.Contains(q, "no required module provides package"):
		return "missing module: run go mod tidy or generate standalone plugin with go.mod"
	case strings.Contains(q, "imported and not used"):
		return "unused import: regenerate or run gofmt/goimports"
	case strings.Contains(q, "expected 'package'"):
		return "malformed generated Go file: inspect template output"
	default:
		return "inspect build log and add a repair hint after fixing once"
	}
}

func appendJSONL(path string, v any) error {
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.Write(append(b, '\n'))
	return err
}

func logLearningToBeads(rec *LearnRecord, opts LearnOptions) error {
	bin := strings.TrimSpace(opts.BeadsBin)
	if bin == "" {
		bin = "bd"
	}
	if _, err := exec.LookPath(bin); err != nil {
		return err
	}
	title := fmt.Sprintf("[ti/convert] %s -> %s", rec.Name, rec.Kind)
	args := []string{"create", title, "-t", "task", "-p", "2", "-l", "ti,convert,learning", "--json"}
	cmd := exec.Command(bin, args...)
	cmd.Dir = opts.Root
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("bd create failed: %w: %s", err, strings.TrimSpace(string(out)))
	}
	comment := fmt.Sprintf("IR=%s source=%s risk=%d model=%s build_ok=%v test_ok=%v patterns=%s", rec.IRID, rec.Source.Format, rec.Learning.RiskScore, rec.Learning.RecommendedModel, rec.BuildOK, rec.TestOK, strings.Join(rec.Learning.PatternKeys, ","))
	_ = exec.Command(bin, "comment", "add", rec.IRID, comment).Run()
	return nil
}
