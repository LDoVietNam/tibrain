package sandbox

import "testing"

func TestValidateCommandBlocksDangerousCommands(t *testing.T) {
	p := DefaultPolicy()
	if err := ValidateCommand("rm -rf .", p); err == nil {
		t.Fatalf("expected rm -rf . to be denied")
	}
}

func TestValidateCommandAllowsNormalCommand(t *testing.T) {
	p := DefaultPolicy()
	if err := ValidateCommand("go test ./...", p); err != nil {
		t.Fatalf("expected normal command to be allowed: %v", err)
	}
}

func TestSanitizedEnv(t *testing.T) {
	got := SanitizedEnv([]string{"PATH=/bin", "SECRET=x"}, []string{"PATH"})
	if len(got) != 1 || got[0] != "PATH=/bin" {
		t.Fatalf("unexpected sanitized env: %#v", got)
	}
}
