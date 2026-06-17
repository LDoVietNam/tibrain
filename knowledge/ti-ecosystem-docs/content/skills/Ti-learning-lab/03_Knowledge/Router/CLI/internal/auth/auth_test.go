package auth

import "testing"

func TestLoadAll(t *testing.T) {
	l := NewLoader("")
	if err := l.LoadAll(); err != nil {
		t.Fatalf("LoadAll failed: %v", err)
	}
}
