package management

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/ti/cli/internal/config"
)

func TestBootstrapScansAuthLayout(t *testing.T) {
	root := t.TempDir()
	cfg := config.Default()
	cfg.AuthDir = root
	cfg.AuthFiles = nil // isolate from real machine auth files
	paths := deriveAuthPaths(cfg)

	for _, dir := range []string{paths.Cookies, paths.Auth, paths.OAuth} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", dir, err)
		}
	}
	if err := os.WriteFile(filepath.Join(paths.Cookies, "iflow.json"), []byte(`[{"name":"gfsessionid","value":"abc","domain":"example.com","path":"/"}]`), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(paths.OAuth, "gemini-cli.oauth"), []byte(`{"state":"connected","account":"demo@example.com"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(paths.Auth, "openrouter.token"), []byte(`{"value":"sk-demo"}`), 0o600); err != nil {
		t.Fatal(err)
	}

	boot := NewService(cfg).Bootstrap()
	if !boot.Dashboard.AuthPathsValid {
		t.Fatalf("expected valid auth paths")
	}
	if len(boot.Cookies) != 1 {
		t.Fatalf("cookies = %d", len(boot.Cookies))
	}
	if len(boot.OAuth) != 1 {
		t.Fatalf("oauth = %d", len(boot.OAuth))
	}
	if len(boot.AuthFiles) < 3 {
		t.Fatalf("authFiles = %d", len(boot.AuthFiles))
	}
	if len(boot.APIKeys) != 1 {
		t.Fatalf("apiKeys = %d", len(boot.APIKeys))
	}
	if boot.OAuth[0].Account != "demo@example.com" {
		t.Fatalf("account = %q", boot.OAuth[0].Account)
	}
}

func TestBootstrapJSONShape(t *testing.T) {
	payload, err := json.Marshal(NewService(config.Default()).Bootstrap())
	if err != nil {
		t.Fatal(err)
	}
	var raw map[string]any
	if err := json.Unmarshal(payload, &raw); err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"dashboard", "authPaths", "providers", "apiKeys", "authFiles", "cookies", "oauth", "oauthStarters", "cookieStarters", "config", "system"} {
		if _, ok := raw[key]; !ok {
			t.Fatalf("missing %s", key)
		}
	}
}
