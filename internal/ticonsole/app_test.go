package ticonsole

import (
	"strings"
	"testing"
	"time"
)

func TestRenderSnapshotShowsLiveStateAndGuardrail(t *testing.T) {
	snapshot := Snapshot{
		GeneratedAt: time.Date(2026, 6, 28, 10, 0, 0, 0, time.Local),
		Services: []ServiceSnapshot{
			{Name: "TiBrain", Role: "control-plane", BaseURL: "http://127.0.0.1:1810", State: StateHealthy, Latency: 12 * time.Millisecond},
			{Name: "Router@1817", Role: "model-router-candidate", BaseURL: "http://127.0.0.1:1817", State: StateDown, Detail: "connection refused"},
		},
		Summary: Summary{Healthy: 1, Down: 1},
		Notices: []string{"Router endpoints are candidates only."},
	}

	frame := RenderSnapshot(TabDashboard, snapshot, false)
	for _, expected := range []string{"Ti Console", "TiBrain", "Router@1817", "DOWN", "Router endpoints are candidates only."} {
		if !strings.Contains(frame, expected) {
			t.Fatalf("expected frame to contain %q\n%s", expected, frame)
		}
	}
}
