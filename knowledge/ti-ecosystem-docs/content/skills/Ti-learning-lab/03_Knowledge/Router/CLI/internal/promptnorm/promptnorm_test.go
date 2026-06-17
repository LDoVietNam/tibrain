package promptnorm

import "testing"

func TestPreferredPromptShapeOverridesDefault(t *testing.T) {
	res := Transform(Request{Goal: "fix", TaskType: "code", Route: "cliproxyapi", PreferredPromptShape: "concise-execution"})
	if res.PromptShape != "concise-execution" {
		t.Fatalf("got %q", res.PromptShape)
	}
}
