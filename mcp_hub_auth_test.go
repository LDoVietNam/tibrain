package main

import (
	"context"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewMCPHubClient_UsesProxyConfigAPIKeyFallback(t *testing.T) {
	t.Setenv("MCP_HUB_API_KEY", "")
	t.Setenv("MCPPROXY_API_KEY", "")
	t.Setenv("MCP_HUB_API_KEY_FILE", "")
	t.Setenv("MCP_HUB_ENV_FILE", "")
	t.Setenv("MCP_HUB_CONFIG", "")
	t.Setenv("MCPPROXY_CONFIG", "")

	tempDir := t.TempDir()
	proxyBin := filepath.Join(tempDir, "mcpproxy.exe")
	configDir := filepath.Join(tempDir, ".mcpproxy")
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatalf("create config dir: %v", err)
	}

	const expectedKey = "proxy-secret-123"
	configPath := filepath.Join(configDir, "mcp_config.json")
	configBody := []byte(`{"api_key":"proxy-secret-123"}`)
	if err := os.WriteFile(configPath, configBody, 0o600); err != nil {
		t.Fatalf("write config file: %v", err)
	}

	t.Setenv("MCP_PROXY_BIN", proxyBin)

	emptyWD := filepath.Join(tempDir, "empty-workdir")
	if err := os.MkdirAll(emptyWD, 0o755); err != nil {
		t.Fatalf("create empty workdir: %v", err)
	}
	originalWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("get working directory: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(originalWD)
	})
	if err := os.Chdir(emptyWD); err != nil {
		t.Fatalf("change working directory: %v", err)
	}

	client := NewMCPHubClient("http://localhost:1840", "")
	if client.apiKey != expectedKey {
		t.Fatalf("expected API key %q from proxy config, got %q", expectedKey, client.apiKey)
	}
}

func TestNewMCPHubClient_UsesRuntimeSecretFallback(t *testing.T) {
	t.Setenv("MCP_HUB_API_KEY", "")
	t.Setenv("MCPPROXY_API_KEY", "")
	t.Setenv("MCP_HUB_API_KEY_FILE", "")
	t.Setenv("MCP_HUB_ENV_FILE", "")
	t.Setenv("MCP_HUB_CONFIG", "")
	t.Setenv("MCPPROXY_CONFIG", "")
	t.Setenv("MCP_PROXY_BIN", "")

	runtimeRoot := t.TempDir()
	if err := os.MkdirAll(filepath.Join(runtimeRoot, "config"), 0o755); err != nil {
		t.Fatalf("create config dir: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(runtimeRoot, "secrets"), 0o755); err != nil {
		t.Fatalf("create secrets dir: %v", err)
	}

	const expectedKey = "runtime-secret-456"
	if err := os.WriteFile(filepath.Join(runtimeRoot, "secrets", "api-key.txt"), []byte(expectedKey), 0o600); err != nil {
		t.Fatalf("write secret file: %v", err)
	}

	t.Setenv("MCP_RUNTIME_ROOT", runtimeRoot)

	client := NewMCPHubClient("http://localhost:1840", "")
	if client.apiKey != expectedKey {
		t.Fatalf("expected API key %q from runtime secret, got %q", expectedKey, client.apiKey)
	}
}

func TestNewMCPHubClient_UsesMCPRuntimeConfigFallback(t *testing.T) {
	t.Setenv("MCP_HUB_API_KEY", "")
	t.Setenv("MCPPROXY_API_KEY", "")
	t.Setenv("MCP_HUB_API_KEY_FILE", "")
	t.Setenv("MCP_HUB_ENV_FILE", "")
	t.Setenv("MCP_HUB_CONFIG", "")
	t.Setenv("MCPPROXY_CONFIG", "")
	t.Setenv("MCP_RUNTIME_ROOT", "")
	t.Setenv("MCP_PROXY_BIN", "")

	tempDir := t.TempDir()
	configDir := filepath.Join(tempDir, "mcp", ".runtime", "config")
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatalf("create config dir: %v", err)
	}

	const expectedKey = "mcp-runtime-secret-123"
	if err := os.WriteFile(filepath.Join(configDir, "mcp_config.json"), []byte(`{"api_key":"mcp-runtime-secret-123"}`), 0o600); err != nil {
		t.Fatalf("write config file: %v", err)
	}

	originalWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("get working directory: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(originalWD)
	})
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("change working directory: %v", err)
	}

	client := NewMCPHubClient("http://localhost:1840", "")
	if client.apiKey != expectedKey {
		t.Fatalf("expected API key %q from mcp runtime config, got %q", expectedKey, client.apiKey)
	}
}

func TestNewMCPHubClient_PrefersColocatedRuntimeOverStaleProxyBin(t *testing.T) {
	t.Setenv("MCP_HUB_API_KEY", "")
	t.Setenv("MCPPROXY_API_KEY", "")
	t.Setenv("MCP_HUB_API_KEY_FILE", "")
	t.Setenv("MCP_HUB_ENV_FILE", "")
	t.Setenv("MCP_HUB_CONFIG", "")
	t.Setenv("MCPPROXY_CONFIG", "")
	t.Setenv("MCP_RUNTIME_ROOT", "")

	tempDir := t.TempDir()
	localConfigDir := filepath.Join(tempDir, "mcp", ".runtime", "config")
	if err := os.MkdirAll(localConfigDir, 0o755); err != nil {
		t.Fatalf("create local config dir: %v", err)
	}
	const expectedKey = "local-runtime-secret-123"
	if err := os.WriteFile(filepath.Join(localConfigDir, "mcp_config.json"), []byte(`{"api_key":"local-runtime-secret-123"}`), 0o600); err != nil {
		t.Fatalf("write local config file: %v", err)
	}

	staleDir := filepath.Join(tempDir, "old-mcp")
	staleConfigDir := filepath.Join(staleDir, ".mcpproxy")
	if err := os.MkdirAll(staleConfigDir, 0o755); err != nil {
		t.Fatalf("create stale config dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(staleConfigDir, "mcp_config.json"), []byte(`{"api_key":"stale-proxy-secret-456"}`), 0o600); err != nil {
		t.Fatalf("write stale config file: %v", err)
	}
	t.Setenv("MCP_PROXY_BIN", filepath.Join(staleDir, "mcpproxy.exe"))

	originalWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("get working directory: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(originalWD)
	})
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("change working directory: %v", err)
	}

	client := NewMCPHubClient("http://localhost:1840", "")
	if client.apiKey != expectedKey {
		t.Fatalf("expected API key %q from colocated runtime config, got %q", expectedKey, client.apiKey)
	}
}

func TestNewMCPHubClient_UsesTiBrainDataConfigFallback(t *testing.T) {
	t.Setenv("MCP_HUB_API_KEY", "")
	t.Setenv("MCPPROXY_API_KEY", "")
	t.Setenv("MCP_HUB_API_KEY_FILE", "")
	t.Setenv("MCP_HUB_ENV_FILE", "")
	t.Setenv("MCP_HUB_CONFIG", "")
	t.Setenv("MCPPROXY_CONFIG", "")
	t.Setenv("MCP_RUNTIME_ROOT", "")
	t.Setenv("MCP_PROXY_BIN", "")

	tempDir := t.TempDir()
	configDir := filepath.Join(tempDir, "tibrain_data", "1mcp")
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatalf("create config dir: %v", err)
	}

	const expectedKey = "tibrain-data-secret-789"
	if err := os.WriteFile(filepath.Join(configDir, "mcp_config.json"), []byte(`{"api_key":"tibrain-data-secret-789"}`), 0o600); err != nil {
		t.Fatalf("write config file: %v", err)
	}

	originalWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("get working directory: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(originalWD)
	})
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("change working directory: %v", err)
	}

	client := NewMCPHubClient("http://localhost:1840", "")
	if client.apiKey != expectedKey {
		t.Fatalf("expected API key %q from tibrain_data config, got %q", expectedKey, client.apiKey)
	}
}

func TestMCPHubClient_ListServersFailsFastWithoutAPIKey(t *testing.T) {
	t.Setenv("MCP_HUB_API_KEY", "")
	t.Setenv("MCPPROXY_API_KEY", "")
	t.Setenv("MCP_HUB_API_KEY_FILE", "")
	t.Setenv("MCP_HUB_ENV_FILE", "")
	t.Setenv("MCP_HUB_CONFIG", "")
	t.Setenv("MCPPROXY_CONFIG", "")
	t.Setenv("MCP_RUNTIME_ROOT", "")
	t.Setenv("MCP_PROXY_BIN", "")

	client := &MCPHubClient{
		hubURL: "http://localhost:1840",
		httpClient: &http.Client{
			Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
				t.Fatal("transport should not be called when API key is missing")
				return nil, nil
			}),
		},
	}

	_, err := client.ListServers(context.Background())
	if err == nil {
		t.Fatal("expected error when API key is missing")
	}
	if !strings.Contains(err.Error(), "not configured") {
		t.Fatalf("expected missing-key error, got %v", err)
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}
