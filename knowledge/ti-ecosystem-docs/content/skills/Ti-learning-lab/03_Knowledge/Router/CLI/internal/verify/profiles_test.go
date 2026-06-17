package verify

import "testing"

func TestProfileForTaskCode(t *testing.T) {
	p := ProfileForTask("code")
	if p.Timeout <= 0 || len(p.Commands) == 0 {
		t.Fatalf("unexpected profile: %+v", p)
	}
}
