package commitflow

import "testing"

func TestParsePorcelain(t *testing.T) {
	s := ParsePorcelain(" M a.go\nA  b.go\n")
	if s.Count != 2 {
		t.Fatalf("count=%d", s.Count)
	}
	if got := GenerateMessage("Implement", "code", s); got == "" {
		t.Fatal("empty message")
	}
}
