package main

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

// resolveMCPHubKeyFromFiles resolves the MCP proxy credential without
// embedding it in source or logs. Explicit process environment variables
// remain the primary source in NewMCPHubClient; this function only provides
// file-based fallbacks for service launchers that cannot inject environment.
func resolveMCPHubKeyFromFiles() string {
	if path := strings.TrimSpace(os.Getenv("MCP_HUB_API_KEY_FILE")); path != "" {
		if value := readTrimmedSecretFile(path); value != "" {
			return value
		}
	}

	if value := readMCPHubKeyFromConfigFiles(); value != "" {
		return value
	}

	candidates := make([]string, 0, 3)
	if path := strings.TrimSpace(os.Getenv("MCP_HUB_ENV_FILE")); path != "" {
		candidates = append(candidates, path)
	}
	if cwd, err := os.Getwd(); err == nil {
		candidates = append(candidates, filepath.Join(cwd, ".env"))
	}
	if executable, err := os.Executable(); err == nil {
		candidates = append(candidates, filepath.Join(filepath.Dir(executable), ".env"))
	}

	seen := make(map[string]struct{}, len(candidates))
	for _, path := range candidates {
		cleaned := filepath.Clean(path)
		if _, exists := seen[cleaned]; exists {
			continue
		}
		seen[cleaned] = struct{}{}
		if value := readNamedEnvValue(cleaned, "MCP_HUB_API_KEY", "MCPPROXY_API_KEY"); value != "" {
			return value
		}
	}

	return ""
}

func readMCPHubKeyFromConfigFiles() string {
	candidates := make([]string, 0, 8)
	add := func(path string) {
		path = strings.TrimSpace(path)
		if path == "" {
			return
		}
		candidates = append(candidates, filepath.Clean(path))
	}

	add(os.Getenv("MCP_HUB_CONFIG"))
	add(os.Getenv("MCPPROXY_CONFIG"))

	if runtimeRoot := strings.TrimSpace(os.Getenv("MCP_RUNTIME_ROOT")); runtimeRoot != "" {
		add(filepath.Join(runtimeRoot, "config", "mcp_config.json"))
		add(filepath.Join(runtimeRoot, "secrets", "api-key.txt"))
	}

	// Prefer the MCP runtime colocated with the running TiBrain checkout before
	// MCP_PROXY_BIN. This prevents stale launcher variables from pointing auth
	// back to the old Z:\01_PROJECTS\apps\mcp tree after MCP was moved under
	// apps\tibrain.
	if cwd, err := os.Getwd(); err == nil {
		add(filepath.Join(cwd, "mcp", ".runtime", "config", "mcp_config.json"))
		add(filepath.Join(cwd, "mcp", ".runtime", "secrets", "api-key.txt"))
		add(filepath.Join(cwd, "mcp", ".mcpproxy", "mcp_config.json"))
		add(filepath.Join(cwd, "data", "1mcp", "mcp_config.json"))
		add(filepath.Join(cwd, ".runtime", "config", "mcp_config.json"))
		add(filepath.Join(cwd, ".runtime", "secrets", "api-key.txt"))
		add(filepath.Join(cwd, ".runtime", "mcpproxy", "mcp_config.json"))
		add(filepath.Join(cwd, ".mcpproxy", "mcp_config.json"))
		add(filepath.Join(filepath.Dir(cwd), "mcp", ".runtime", "config", "mcp_config.json"))
		add(filepath.Join(filepath.Dir(cwd), "mcp", ".runtime", "secrets", "api-key.txt"))
		add(filepath.Join(filepath.Dir(cwd), "mcp", ".mcpproxy", "mcp_config.json"))
	}
	if executable, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(executable)
		add(filepath.Join(exeDir, "mcp", ".runtime", "config", "mcp_config.json"))
		add(filepath.Join(exeDir, "mcp", ".runtime", "secrets", "api-key.txt"))
		add(filepath.Join(exeDir, "mcp", ".mcpproxy", "mcp_config.json"))
		add(filepath.Join(exeDir, "data", "1mcp", "mcp_config.json"))
		add(filepath.Join(exeDir, ".runtime", "config", "mcp_config.json"))
		add(filepath.Join(exeDir, ".runtime", "secrets", "api-key.txt"))
		add(filepath.Join(exeDir, ".runtime", "mcpproxy", "mcp_config.json"))
		add(filepath.Join(exeDir, ".mcpproxy", "mcp_config.json"))
		add(filepath.Join(filepath.Dir(exeDir), "mcp", ".runtime", "config", "mcp_config.json"))
		add(filepath.Join(filepath.Dir(exeDir), "mcp", ".runtime", "secrets", "api-key.txt"))
		add(filepath.Join(filepath.Dir(exeDir), "mcp", ".mcpproxy", "mcp_config.json"))
	}

	if proxyBin := strings.TrimSpace(os.Getenv("MCP_PROXY_BIN")); proxyBin != "" {
		add(filepath.Join(filepath.Dir(proxyBin), ".mcpproxy", "mcp_config.json"))
	}
	if homeDir, err := os.UserHomeDir(); err == nil && strings.TrimSpace(homeDir) != "" {
		add(filepath.Join(homeDir, ".runtime", "config", "mcp_config.json"))
		add(filepath.Join(homeDir, ".runtime", "secrets", "api-key.txt"))
		add(filepath.Join(homeDir, ".mcpproxy", "mcp_config.json"))
	}

	seen := make(map[string]struct{}, len(candidates))
	for _, path := range candidates {
		if _, exists := seen[path]; exists {
			continue
		}
		seen[path] = struct{}{}
		if value := readMCPHubAPIKeyFromCandidate(path); value != "" {
			return value
		}
	}

	return ""
}

func readMCPHubAPIKeyFromCandidate(path string) string {
	base := strings.ToLower(filepath.Base(path))
	if base == "api-key.txt" {
		return readTrimmedSecretFile(path)
	}
	return readMCPHubAPIKeyFromConfigFile(path)
}

func readMCPHubAPIKeyFromConfigFile(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}

	var cfg struct {
		APIKey string `json:"api_key"`
	}
	if err := json.Unmarshal(data, &cfg); err != nil {
		return ""
	}

	return strings.TrimSpace(cfg.APIKey)
}

func readTrimmedSecretFile(path string) string {
	data, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}

func readNamedEnvValue(path string, names ...string) string {
	file, err := os.Open(filepath.Clean(path))
	if err != nil {
		return ""
	}
	defer file.Close()

	wanted := make(map[string]struct{}, len(names))
	for _, name := range names {
		wanted[name] = struct{}{}
	}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		line = strings.TrimSpace(strings.TrimPrefix(line, "export "))
		name, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		name = strings.TrimSpace(name)
		if _, ok := wanted[name]; !ok {
			continue
		}
		value = strings.TrimSpace(value)
		if len(value) >= 2 {
			if (value[0] == '"' && value[len(value)-1] == '"') || (value[0] == '\'' && value[len(value)-1] == '\'') {
				value = value[1 : len(value)-1]
			}
		}
		if value != "" {
			return value
		}
	}
	return ""
}
