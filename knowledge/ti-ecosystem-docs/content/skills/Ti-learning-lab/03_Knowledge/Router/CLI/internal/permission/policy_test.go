package permission

import "testing"

func TestPolicyEvaluate(t *testing.T) {
	p := DefaultPolicy()
	d := p.Evaluate("bash", "", "rm -rf dist")
	if d.Behavior == BehaviorAllow {
		t.Fatal("expected non-allow for destructive bash")
	}
	d = p.Evaluate("file_read", "README.md", "")
	if d.Behavior == BehaviorDeny {
		t.Fatal("read should not be denied by default")
	}
}
