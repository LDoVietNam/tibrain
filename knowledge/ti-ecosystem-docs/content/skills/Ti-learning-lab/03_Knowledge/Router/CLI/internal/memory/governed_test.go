package memory

import (
	"path/filepath"
	"testing"
)

func TestGovernedStoreStats(t *testing.T) {
	s, err := NewGovernedStore(filepath.Join(t.TempDir(), "mem"))
	if err != nil {
		t.Fatal(err)
	}
	_ = s.SaveCanonical("approved_plan", "ti", "planner", map[string]any{"id": "p1"})
	_ = s.RecordEpisodic("act_start", "ti", "ticore", map[string]any{"id": "p1"})
	_ = s.RecordLearned("act_outcome", "ti", "ticore", map[string]any{"id": "p1"})
	st, err := s.Stats()
	if err != nil {
		t.Fatal(err)
	}
	if st.Total != 3 || st.Canonical != 1 || st.Episodic != 1 || st.Learned != 1 {
		t.Fatalf("unexpected stats: %+v", st)
	}
}
