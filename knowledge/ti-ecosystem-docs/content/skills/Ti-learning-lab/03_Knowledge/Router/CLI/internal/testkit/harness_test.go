package testkit

import "testing"

func TestHarnessConfig(t *testing.T) {
	h := New(t)
	cfg := h.Config()
	if cfg.DataDir == "" || cfg.AuthDir == "" {
		t.Fatal("expected harness dirs")
	}
	if cfg.Model == "" || cfg.DefaultPhase == "" {
		t.Fatal("expected harness config defaults")
	}
}
