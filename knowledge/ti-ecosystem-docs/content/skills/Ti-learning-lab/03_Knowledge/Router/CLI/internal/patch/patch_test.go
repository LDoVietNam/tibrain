package patch

import (
	"os"
	"path/filepath"
	"testing"
)

func TestApplyAndRollback(t *testing.T) {
	dir := t.TempDir()
	mgr, err := NewManager(filepath.Join(dir, ".ti", "patches"))
	if err != nil {
		t.Fatal(err)
	}
	target := filepath.Join(dir, "hello.txt")
	if err := os.WriteFile(target, []byte("hello world"), 0o644); err != nil {
		t.Fatal(err)
	}
	report, err := mgr.ApplyReplaceOps([]ReplaceOp{{Path: target, Old: "world", New: "ti"}}, "replace world")
	if err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(target)
	if string(data) != "hello ti" {
		t.Fatalf("got %q", string(data))
	}
	if _, err := mgr.Rollback(report.SnapshotID); err != nil {
		t.Fatal(err)
	}
	data, _ = os.ReadFile(target)
	if string(data) != "hello world" {
		t.Fatalf("rollback got %q", string(data))
	}
}
