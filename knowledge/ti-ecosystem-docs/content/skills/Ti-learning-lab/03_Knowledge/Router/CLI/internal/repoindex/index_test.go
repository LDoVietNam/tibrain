package repoindex

import (
	"os"
	"path/filepath"
	"testing"
)

func TestAnalyzeAndSearch(t *testing.T) {
	dir := t.TempDir()
	mustWrite := func(path, body string) {
		full := filepath.Join(dir, path)
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	mustWrite("go.mod", "module example.com/test\n")
	mustWrite("cmd/app/main.go", "package main\n")
	mustWrite("internal/router/router.go", "package router\n")
	mustWrite("internal/router/router_test.go", "package router\n")
	summary, err := Analyze(dir)
	if err != nil {
		t.Fatal(err)
	}
	if summary.Module != "example.com/test" {
		t.Fatalf("module=%q", summary.Module)
	}
	files := SearchRelevantFiles(summary, "fix router", 3)
	if len(files) == 0 {
		t.Fatal("expected relevant files")
	}
}
