package tokenopt

import "testing"

func TestCompressSavesTokens(t *testing.T) {
	in := ""
	for i := 0; i < 400; i++ {
		in += "ok line\n"
	}
	in += "ERROR failed assertion\n"
	res := Compress(in, Options{Command: "go test ./...", MaxLines: 80, MaxBytes: 8000})
	if res.EstimatedTokensOut >= res.EstimatedTokensIn {
		t.Fatalf("expected compression: in=%d out=%d", res.EstimatedTokensIn, res.EstimatedTokensOut)
	}
	if res.Output == "" {
		t.Fatal("empty output")
	}
}
