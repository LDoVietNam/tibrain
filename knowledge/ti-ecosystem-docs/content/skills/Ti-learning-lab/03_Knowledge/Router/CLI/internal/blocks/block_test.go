package blocks

import (
	"strings"
	"testing"
)

func TestRecordListOutput(t *testing.T) {
	store := NewStore(t.TempDir())
	b, err := store.Record(RecordRequest{Type: TypeCommand, Command: "echo hello", ExitCode: 0, Stdout: "hello\n"})
	if err != nil {
		t.Fatal(err)
	}
	if b.ID == "" || b.StdoutRef == "" {
		t.Fatalf("unexpected block: %+v", b)
	}
	blocks, err := store.List(10)
	if err != nil {
		t.Fatal(err)
	}
	if len(blocks) != 1 {
		t.Fatalf("expected 1 block, got %d", len(blocks))
	}
	out, err := store.Output(&blocks[0])
	if err != nil {
		t.Fatal(err)
	}
	if out != "hello\n" {
		t.Fatalf("unexpected output: %q", out)
	}
}

func TestExportMarkdown(t *testing.T) {
	store := NewStore(t.TempDir())
	b, err := store.Record(RecordRequest{Type: TypeSandbox, Command: "go test ./...", ExitCode: 1, Stdout: "FAIL\n"})
	if err != nil {
		t.Fatal(err)
	}
	md, err := store.ExportMarkdown(b)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(md, "go test ./...") || !strings.Contains(md, "FAIL") {
		t.Fatalf("markdown missing content: %s", md)
	}
}
