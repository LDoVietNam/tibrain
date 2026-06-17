package authtoken

import "testing"

func TestStoreImport(t *testing.T) {
	dir := t.TempDir()
	s := NewStore(dir)
	p, err := s.Import("windsurf", "windsurf", "abc1234567890", "test")
	if err != nil {
		t.Fatal(err)
	}
	if p.Preview == "abc1234567890" {
		t.Fatal("token not redacted")
	}
	tok, _, err := s.GetSecret("windsurf")
	if err != nil {
		t.Fatal(err)
	}
	if tok != "abc1234567890" {
		t.Fatal("wrong token")
	}
}
