package modelregistry

import "testing"

func TestResolveAlias(t *testing.T) {
	r := DefaultRegistry()
	m, ok := r.Resolve("planner")
	if !ok || m.Canonical != "deepseek-reasoner" {
		t.Fatalf("resolve planner = %+v ok=%v", m, ok)
	}
}

func TestSelectPrefersRequestedThenPhase(t *testing.T) {
	r := DefaultRegistry()
	st := State{DefaultModel: "claude-sonnet-4", PhaseModels: map[string]string{"plan": "deepseek-reasoner"}}
	res := r.Select(st, "plan", "", "", "claude-haiku-4")
	if res.Model.Canonical != "deepseek-reasoner" || res.Source != "phase-preference" {
		t.Fatalf("unexpected resolution: %+v", res)
	}
	res = r.Select(st, "plan", "gpt-4o", "", "claude-haiku-4")
	if res.Model.Canonical != "gpt-4o" || res.Source != "requested" {
		t.Fatalf("unexpected requested resolution: %+v", res)
	}
}
