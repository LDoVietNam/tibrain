package memory

import (
	"path/filepath"
	"testing"
)

func TestPromoteVerifiedStepUsesRules(t *testing.T) {
	s, err := NewGovernedStore(filepath.Join(t.TempDir(), "mem"))
	if err != nil {
		t.Fatal(err)
	}
	ok, err := s.Promote("verified_step", "ti", "verify", VerificationConfidence(true, 2, 0), map[string]any{
		"plan_id": "p1", "step_id": "S1", "step_title": "Validate", "verify_commands": 2, "verify_passed": true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("expected promotion success")
	}
}

func TestPromoteVerifiedStepRejectsFailedVerify(t *testing.T) {
	s, err := NewGovernedStore(filepath.Join(t.TempDir(), "mem"))
	if err != nil {
		t.Fatal(err)
	}
	ok, err := s.Promote("verified_step", "ti", "verify", VerificationConfidence(false, 1, 1), map[string]any{
		"plan_id": "p1", "step_id": "S1", "step_title": "Validate", "verify_commands": 1, "verify_passed": false,
	})
	if err != nil {
		t.Fatal(err)
	}
	if ok {
		t.Fatal("expected promotion rejection")
	}
}
