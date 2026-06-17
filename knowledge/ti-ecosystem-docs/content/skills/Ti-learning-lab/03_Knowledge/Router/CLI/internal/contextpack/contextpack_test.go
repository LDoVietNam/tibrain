package contextpack

import "testing"

func TestPreferredPromptShapeAffectsBundle(t *testing.T) {
	b := Build(Input{Goal: "fix router", TaskType: "code", CandidateFiles: []string{"a.go", "b.go", "c.go", "d.go", "e.go"}, PreferredPromptShape: "concise-execution"})
	if b.PreferredPromptShape != "concise-execution" {
		t.Fatalf("got %q", b.PreferredPromptShape)
	}
	if len(b.CandidateFiles) > 4 {
		t.Fatalf("expected concise preference to reduce files, got %d", len(b.CandidateFiles))
	}
}
