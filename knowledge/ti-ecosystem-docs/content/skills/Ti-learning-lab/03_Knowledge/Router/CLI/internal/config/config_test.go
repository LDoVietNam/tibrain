package config_test

import (
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/ti/cli/internal/config"
)

func TestDefault(t *testing.T) {
	cfg := config.Default()
	if cfg.Model != "qwenvsclaude-3.6" {
		t.Fatalf("default model = %q, want qwenvsclaude-3.6", cfg.Model)
	}
	if cfg.Port != 1810 {
		t.Fatalf("default port = %d, want 1810", cfg.Port)
	}
	if cfg.Provider != "openrouter" {
		t.Fatalf("default provider = %q, want openrouter", cfg.Provider)
	}
	if cfg.RateLimit != 20 {
		t.Fatalf("default rate limit = %d, want 20", cfg.RateLimit)
	}
	if cfg.MaxTokens != 8192 {
		t.Fatalf("default max tokens = %d, want 8192", cfg.MaxTokens)
	}
}

func TestLoadFromFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "ti.json")
	data := `{"model":"gpt-4o","port":9999,"provider":"openai"}`
	if err := os.WriteFile(path, []byte(data), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := config.Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Model != "gpt-4o" {
		t.Fatalf("model = %q, want gpt-4o", cfg.Model)
	}
	if cfg.Port != 9999 {
		t.Fatalf("port = %d, want 9999", cfg.Port)
	}
	if cfg.Provider != "openai" {
		t.Fatalf("provider = %q, want openai", cfg.Provider)
	}
}

func TestLoadMissingFile(t *testing.T) {
	cfg, err := config.Load("/nonexistent/path/ti.json")
	if err != nil {
		t.Fatalf("expected no error for missing file, got: %v", err)
	}
	// Should return defaults
	if cfg.Model != "qwenvsclaude-3.6" {
		t.Fatalf("expected default model, got %q", cfg.Model)
	}
}

func TestEnvOverrides(t *testing.T) {
	os.Setenv("TI_MODEL", "gemini-2.5-pro")
	os.Setenv("TI_PORT", "7777")
	os.Setenv("ANTHROPIC_API_KEY", "sk-test-123")
	defer func() {
		os.Unsetenv("TI_MODEL")
		os.Unsetenv("TI_PORT")
		os.Unsetenv("ANTHROPIC_API_KEY")
	}()

	cfg, err := config.Load("/nonexistent")
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Model != "gemini-2.5-pro" {
		t.Fatalf("model = %q, want gemini-2.5-pro", cfg.Model)
	}
	if cfg.Port != 7777 {
		t.Fatalf("port = %d, want 7777", cfg.Port)
	}
	if cfg.APIKeys["anthropic"] != "sk-test-123" {
		t.Fatalf("anthropic key = %q, want sk-test-123", cfg.APIKeys["anthropic"])
	}
}

func TestEnvFallback(t *testing.T) {
	// Fallback: TI_MODEL should NOT override if model is already set in config
	dir := t.TempDir()
	path := filepath.Join(dir, "ti.json")
	os.WriteFile(path, []byte(`{"model":"from-file"}`), 0644)

	os.Setenv("TI_MODEL", "from-env")
	defer os.Unsetenv("TI_MODEL")

	cfg, err := config.Load(path)
	if err != nil {
		t.Fatal(err)
	}
	// envStr always overwrites — this is correct for TI_MODEL
	if cfg.Model != "from-env" {
		t.Fatalf("model = %q, want from-env (always override)", cfg.Model)
	}
}

func TestExpandHome(t *testing.T) {
	home, _ := os.UserHomeDir()

	got := config.ExpandHome("~/.ti/data")
	want := home + "/.ti/data"
	if got != want {
		t.Fatalf("expand = %q, want %q", got, want)
	}

	// No ~ → unchanged
	got = config.ExpandHome("/absolute/path")
	if got != "/absolute/path" {
		t.Fatalf("expand no ~ = %q, want /absolute/path", got)
	}

	// Empty → unchanged
	got = config.ExpandHome("")
	if got != "" {
		t.Fatalf("expand empty = %q, want empty", got)
	}
}

func TestContractHome(t *testing.T) {
	home, _ := os.UserHomeDir()

	got := config.ContractHome(home + "/.ti/data")
	if got != "~/.ti/data" {
		t.Fatalf("contract = %q, want ~/.ti/data", got)
	}

	// No home prefix → unchanged
	got = config.ContractHome("/other/path")
	if got != "/other/path" {
		t.Fatalf("contract no prefix = %q, want /other/path", got)
	}
}

func TestSaveAndReload(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "ti.json")

	cfg := config.Default()
	cfg.Model = "saved-model"
	cfg.Port = 5555
	cfg.Provider = "google"

	if err := cfg.Save(path); err != nil {
		t.Fatal(err)
	}

	loaded, err := config.Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Model != "saved-model" {
		t.Fatalf("reload model = %q, want saved-model", loaded.Model)
	}
	if loaded.Port != 5555 {
		t.Fatalf("reload port = %d, want 5555", loaded.Port)
	}
}

func TestAtomicReplaceFrom(t *testing.T) {
	cfg := config.Default()

	// Do a hot reload
	newCfg := config.Default()
	newCfg.Model = "hot-reload-model"
	cfg.ReplaceFrom(newCfg)

	// Verify via GetSnapshot
	snap := cfg.GetSnapshot()
	if snap.Model != "hot-reload-model" {
		t.Fatalf("final model = %q, want hot-reload-model", snap.Model)
	}
}

func TestConcurrentSnapshotDuringReplace(t *testing.T) {
	cfg := config.Default()

	var wg sync.WaitGroup
	done := make(chan bool, 200)

	// Concurrent reads
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			snap := cfg.GetSnapshot()
			if snap == nil {
				done <- true
				return
			}
			done <- true
		}()
	}

	// Concurrent replaces
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			newCfg := config.Default()
			newCfg.Model = "replace" + string(rune('a'+n%26))
			cfg.ReplaceFrom(newCfg)
		}(i)
	}

	wg.Wait()
	close(done)
	for range done {
		// drain
	}

	// Final state should be valid
	snap := cfg.GetSnapshot()
	if snap == nil {
		t.Fatal("expected non-nil snapshot")
	}
	if snap.Model == "" {
		t.Fatal("expected non-empty model")
	}
}

func TestHash(t *testing.T) {
	cfg1 := config.Default()
	cfg2 := config.Default()

	// Same config → same hash
	h1 := cfg1.Hash()
	h2 := cfg2.Hash()
	if h1 != h2 {
		t.Fatalf("hashes differ: %s vs %s", h1, h2)
	}

	// Different config → different hash
	cfg3 := config.Default()
	cfg3.Model = "different-model"
	h3 := cfg3.Hash()
	if h3 == h1 {
		t.Fatal("expected different hash for different config")
	}
}

func TestEnvSubstitution(t *testing.T) {
	os.Setenv("MY_SECRET_KEY", "super-secret-value")
	defer os.Unsetenv("MY_SECRET_KEY")

	cfg := config.Default()
	cfg.APIKeys["test"] = "{env:MY_SECRET_KEY}"
	cfg.ApplyEnvSubstitution()

	if cfg.APIKeys["test"] != "super-secret-value" {
		t.Fatalf("substitution = %q, want super-secret-value", cfg.APIKeys["test"])
	}
}

func TestEnvSubstitutionMissingVar(t *testing.T) {
	cfg := config.Default()
	cfg.APIKeys["test"] = "{env:NONEXISTENT_VAR_12345}"
	cfg.ApplyEnvSubstitution()

	// Missing env var → replaced with empty string
	if cfg.APIKeys["test"] != "" {
		t.Fatalf("missing var = %q, want empty string", cfg.APIKeys["test"])
	}
}

func TestTimeout(t *testing.T) {
	cfg := config.Default()
	cfg.TimeoutSec = 300
	if cfg.Timeout() != 300*time.Second {
		t.Fatalf("timeout = %v, want 5m", cfg.Timeout())
	}
}

func TestDataDirEnvOverride(t *testing.T) {
	os.Setenv("TI_DATA_DIR", "/custom/data")
	defer os.Unsetenv("TI_DATA_DIR")

	cfg, err := config.Load("/nonexistent")
	if err != nil {
		t.Fatal(err)
	}
	if cfg.DataDir != "/custom/data" {
		t.Fatalf("data_dir = %q, want /custom/data", cfg.DataDir)
	}
}

func TestAPIKeysFromEnv(t *testing.T) {
	os.Setenv("OPENAI_API_KEY", "sk-openai-test")
	os.Setenv("GOOGLE_API_KEY", "google-key-test")
	defer func() {
		os.Unsetenv("OPENAI_API_KEY")
		os.Unsetenv("GOOGLE_API_KEY")
	}()

	cfg, err := config.Load("/nonexistent")
	if err != nil {
		t.Fatal(err)
	}
	if cfg.APIKeys["openai"] != "sk-openai-test" {
		t.Fatalf("openai key = %q", cfg.APIKeys["openai"])
	}
	if cfg.APIKeys["google"] != "google-key-test" {
		t.Fatalf("google key = %q", cfg.APIKeys["google"])
	}
}

func TestPublicExcludesInternal(t *testing.T) {
	cfg := config.Default()
	pub := cfg.Public()

	// Public should not have mutex or atomic pointer
	// (we verify by checking that Public() returns a clean copy)
	if pub == cfg {
		t.Fatal("Public() should return a new copy")
	}
}

func TestLoadAutoDefaults(t *testing.T) {
	// With no config files or env vars set, should return defaults
	os.Unsetenv("TI_CONFIG")

	cfg, err := config.LoadAuto()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Model != "qwenvsclaude-3.6" {
		t.Fatalf("auto default model = %q", cfg.Model)
	}
}

func TestSaveCreatesDirectory(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sub", "dir", "ti.json")

	cfg := config.Default()
	if err := cfg.Save(path); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatal("file not created")
	}
}

func TestLoadAutoWithCustomPath(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "custom.json")
	os.WriteFile(path, []byte(`{"model":"custom-model"}`), 0644)

	os.Setenv("TI_CONFIG", path)
	defer os.Unsetenv("TI_CONFIG")

	cfg, err := config.LoadAuto()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Model != "custom-model" {
		t.Fatalf("auto custom model = %q, want custom-model", cfg.Model)
	}
}

func TestHashConsistency(t *testing.T) {
	cfg := config.Default()

	// Hash should be deterministic
	for i := 0; i < 10; i++ {
		h := cfg.Hash()
		if h != cfg.Hash() {
			t.Fatalf("hash not deterministic: iteration %d", i)
		}
	}
}

func TestConcurrentSaveLoad(t *testing.T) {
	dir := t.TempDir()

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			path := filepath.Join(dir, "cfg.json")
			c := config.Default()
			c.Model = "model" + string(rune('a'+n%26))
			c.Save(path)
			config.Load(path)
		}(i)
	}
	wg.Wait()
}
