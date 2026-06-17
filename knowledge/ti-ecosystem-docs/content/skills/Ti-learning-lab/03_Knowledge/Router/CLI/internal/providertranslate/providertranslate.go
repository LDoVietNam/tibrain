package providertranslate

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

type Finding struct {
	Path       string   `json:"path"`
	Format     string   `json:"format"`
	Kind       string   `json:"kind"`
	Confidence float64  `json:"confidence"`
	Notes      []string `json:"notes,omitempty"`
}

type Provider struct {
	Name            string            `json:"name"`
	Type            string            `json:"type,omitempty"`
	BaseURL         string            `json:"base_url"`
	ChatPath        string            `json:"chat_path,omitempty"`
	ModelsPath      string            `json:"models_path,omitempty"`
	APIKey          string            `json:"api_key,omitempty"`
	APIKeyEnv       string            `json:"api_key_env,omitempty"`
	AuthHeader      string            `json:"auth_header,omitempty"`
	AuthTokenPrefix string            `json:"auth_token_prefix,omitempty"`
	CustomHeaders   map[string]string `json:"custom_headers,omitempty"`
	Models          []string          `json:"models,omitempty"`
	DefaultModel    string            `json:"default_model,omitempty"`
	ForceHTTPS      bool              `json:"force_https,omitempty"`
	TimeoutSec      int               `json:"timeout_sec,omitempty"`
	SourcePath      string            `json:"source_path,omitempty"`
	SourceFormat    string            `json:"source_format,omitempty"`
}

type Result struct {
	Providers []Provider `json:"compatible_providers"`
	Warnings  []string   `json:"warnings,omitempty"`
}

var providerFileNames = map[string]string{
	"opencode.json":     "opencode",
	"opencode.jsonc":    "opencode",
	".opencode.json":    "opencode",
	"kilo.json":         "kilo",
	"kilocode.json":     "kilo",
	".kilocode.json":    "kilo",
	"providers.json":    "generic",
	"provider.json":     "generic",
	"ti.providers.json": "ti",
}

func Scan(root string, maxDepth int) ([]Finding, error) {
	if root == "" {
		root = "."
	}
	if maxDepth <= 0 {
		maxDepth = 8
	}
	root = filepath.Clean(root)
	var out []Finding
	skip := map[string]bool{".git": true, "node_modules": true, "vendor": true, "dist": true, "build": true, "bin": true, "target": true, ".cache": true}
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			if path != root && skip[d.Name()] {
				return filepath.SkipDir
			}
			if depth(root, path) > maxDepth {
				return filepath.SkipDir
			}
			return nil
		}
		base := d.Name()
		rel, _ := filepath.Rel(root, path)
		rel = filepath.ToSlash(rel)
		if fmtName, ok := providerFileNames[strings.ToLower(base)]; ok {
			out = append(out, Finding{Path: rel, Format: fmtName, Kind: "provider_config", Confidence: 0.9, Notes: []string{"recognized provider config filename"}})
			return nil
		}
		low := strings.ToLower(rel)
		switch {
		case strings.HasPrefix(low, ".opencode/") && (strings.HasSuffix(low, ".json") || strings.HasSuffix(low, ".jsonc")):
			out = append(out, Finding{Path: rel, Format: "opencode", Kind: "provider_config", Confidence: 0.75, Notes: []string{"opencode directory JSON"}})
		case strings.Contains(low, "kilo") && strings.HasSuffix(low, ".json"):
			out = append(out, Finding{Path: rel, Format: "kilo", Kind: "provider_config", Confidence: 0.7, Notes: []string{"kilo-like JSON filename"}})
		}
		return nil
	})
	sort.Slice(out, func(i, j int) bool { return out[i].Path < out[j].Path })
	return out, err
}

func ConvertFile(path string, forcedFormat string) (Result, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Result{}, err
	}
	clean := stripJSONComments(string(data))
	var root any
	if err := json.Unmarshal([]byte(clean), &root); err != nil {
		return Result{}, fmt.Errorf("parse provider config %s: %w", path, err)
	}
	format := forcedFormat
	if format == "" || format == "auto" {
		format = detectFormat(path)
	}
	res := Result{}
	seen := map[string]bool{}
	add := func(p Provider) {
		p = normalizeProvider(p, path, format)
		if p.BaseURL == "" {
			return
		}
		key := p.Name + "|" + p.BaseURL
		if seen[key] {
			return
		}
		seen[key] = true
		res.Providers = append(res.Providers, p)
	}

	if m, ok := root.(map[string]any); ok {
		for _, p := range extractProvidersFromMap(m, format) {
			add(p)
		}
		if len(res.Providers) == 0 {
			if p := providerFromMap("provider", m, format); p.BaseURL != "" {
				add(p)
			}
		}
	}
	sort.Slice(res.Providers, func(i, j int) bool { return res.Providers[i].Name < res.Providers[j].Name })
	if len(res.Providers) == 0 {
		res.Warnings = append(res.Warnings, "no OpenAI-compatible provider entries were detected; keep the file as a manual reference")
	}
	return res, nil
}

func ConvertPaths(root string, findings []Finding, forcedFormat string) (Result, error) {
	out := Result{}
	seen := map[string]bool{}
	for _, f := range findings {
		path := f.Path
		if root != "" {
			path = filepath.Join(root, filepath.FromSlash(f.Path))
		}
		res, err := ConvertFile(path, firstNonEmpty(forcedFormat, f.Format))
		if err != nil {
			out.Warnings = append(out.Warnings, err.Error())
			continue
		}
		out.Warnings = append(out.Warnings, res.Warnings...)
		for _, p := range res.Providers {
			key := p.Name + "|" + p.BaseURL
			if seen[key] {
				continue
			}
			seen[key] = true
			out.Providers = append(out.Providers, p)
		}
	}
	sort.Slice(out.Providers, func(i, j int) bool { return out.Providers[i].Name < out.Providers[j].Name })
	return out, nil
}

func RenderTiConfigSnippet(res Result) (string, error) {
	b, err := json.MarshalIndent(res, "", "  ")
	if err != nil {
		return "", err
	}
	return string(append(b, '\n')), nil
}

func Write(path string, res Result) error {
	if path == "" {
		path = "ti.providers.generated.json"
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil && filepath.Dir(path) != "." {
		return err
	}
	s, err := RenderTiConfigSnippet(res)
	if err != nil {
		return err
	}
	return os.WriteFile(path, []byte(s), 0o600)
}

func extractProvidersFromMap(m map[string]any, format string) []Provider {
	var out []Provider
	keys := []string{"compatible_providers", "providers", "provider", "models"}
	for _, k := range keys {
		v, ok := getAny(m, k)
		if !ok {
			continue
		}
		switch x := v.(type) {
		case []any:
			for i, item := range x {
				if im, ok := item.(map[string]any); ok {
					out = append(out, providerFromMap(fmt.Sprintf("provider-%d", i+1), im, format))
				}
			}
		case map[string]any:
			// OpenCode often keeps provider-like sections as maps. Treat each object child as a provider.
			for name, item := range x {
				if im, ok := item.(map[string]any); ok {
					out = append(out, providerFromMap(name, im, format))
				}
			}
			if p := providerFromMap("provider", x, format); p.BaseURL != "" {
				out = append(out, p)
			}
		}
	}
	return out
}

func providerFromMap(name string, m map[string]any, format string) Provider {
	p := Provider{Name: cleanName(firstNonEmpty(str(m, "name"), str(m, "id"), name)), Type: "openai_compatible", SourceFormat: format}
	p.BaseURL = firstNonEmpty(str(m, "base_url"), str(m, "baseURL"), str(m, "baseUrl"), str(m, "api_base"), str(m, "apiBase"), str(m, "url"), str(m, "endpoint"))
	p.ChatPath = firstNonEmpty(str(m, "chat_path"), str(m, "chatPath"), str(m, "chat_completions_path"), str(m, "chatCompletionsPath"))
	p.ModelsPath = firstNonEmpty(str(m, "models_path"), str(m, "modelsPath"))
	p.APIKeyEnv = firstNonEmpty(str(m, "api_key_env"), str(m, "apiKeyEnv"), str(m, "env"), str(m, "keyEnv"))
	if p.APIKeyEnv == "" {
		if s := str(m, "apiKey"); isEnvRef(s) {
			p.APIKeyEnv = trimEnvRef(s)
		}
		if s := str(m, "api_key"); isEnvRef(s) {
			p.APIKeyEnv = trimEnvRef(s)
		}
	}
	p.AuthHeader = firstNonEmpty(str(m, "auth_header"), str(m, "authHeader"))
	p.AuthTokenPrefix = firstNonEmpty(str(m, "auth_token_prefix"), str(m, "authTokenPrefix"))
	p.DefaultModel = firstNonEmpty(str(m, "default_model"), str(m, "defaultModel"), str(m, "model"))
	p.Models = stringSlice(firstAny(m, "models", "model_ids", "modelIDs"))
	if len(p.Models) == 0 && p.DefaultModel != "" {
		p.Models = []string{p.DefaultModel}
	}
	p.CustomHeaders = mapString(firstAny(m, "headers", "custom_headers", "customHeaders"))
	p.ForceHTTPS = true
	p.TimeoutSec = intNumber(firstAny(m, "timeout_sec", "timeoutSec", "timeout"))
	return p
}

func normalizeProvider(p Provider, sourcePath, format string) Provider {
	p.Name = cleanName(p.Name)
	if p.Name == "" || p.Name == "provider" {
		p.Name = cleanName(format + "-provider")
	}
	p.Type = firstNonEmpty(p.Type, "openai_compatible")
	p.BaseURL = strings.TrimRight(strings.TrimSpace(p.BaseURL), "/")
	if p.ChatPath == "" {
		p.ChatPath = "/chat/completions"
	}
	if !strings.HasPrefix(p.ChatPath, "/") {
		p.ChatPath = "/" + p.ChatPath
	}
	if p.AuthHeader == "" {
		p.AuthHeader = "Authorization"
	}
	if p.AuthTokenPrefix == "" {
		p.AuthTokenPrefix = "Bearer"
	}
	if p.DefaultModel == "" && len(p.Models) > 0 {
		p.DefaultModel = p.Models[0]
	}
	if p.TimeoutSec == 0 {
		p.TimeoutSec = 120
	}
	p.SourcePath = filepath.ToSlash(sourcePath)
	if p.SourceFormat == "" {
		p.SourceFormat = format
	}
	return p
}

func detectFormat(path string) string {
	low := strings.ToLower(filepath.ToSlash(path))
	switch {
	case strings.Contains(low, "opencode"):
		return "opencode"
	case strings.Contains(low, "kilo"):
		return "kilo"
	case strings.Contains(low, "ti.providers"):
		return "ti"
	default:
		return "generic"
	}
}

func stripJSONComments(s string) string {
	reLine := regexp.MustCompile(`(?m)//.*$`)
	s = reLine.ReplaceAllString(s, "")
	reBlock := regexp.MustCompile(`(?s)/\*.*?\*/`)
	return reBlock.ReplaceAllString(s, "")
}

func getAny(m map[string]any, key string) (any, bool) {
	if v, ok := m[key]; ok {
		return v, true
	}
	for k, v := range m {
		if strings.EqualFold(k, key) {
			return v, true
		}
	}
	return nil, false
}
func firstAny(m map[string]any, keys ...string) any {
	for _, k := range keys {
		if v, ok := getAny(m, k); ok {
			return v
		}
	}
	return nil
}
func str(m map[string]any, key string) string {
	if v, ok := getAny(m, key); ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}
func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}
func cleanName(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = strings.NewReplacer("_", "-", " ", "-", "/", "-", ":", "-").Replace(s)
	return strings.Trim(s, "-")
}
func isEnvRef(s string) bool {
	return strings.HasPrefix(s, "$") || strings.HasPrefix(s, "env:") || strings.HasPrefix(s, "{env:")
}
func trimEnvRef(s string) string {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "$")
	s = strings.TrimPrefix(s, "env:")
	s = strings.TrimPrefix(s, "{env:")
	s = strings.TrimSuffix(s, "}")
	return s
}

func stringSlice(v any) []string {
	var out []string
	switch x := v.(type) {
	case []any:
		for _, it := range x {
			if s, ok := it.(string); ok && strings.TrimSpace(s) != "" {
				out = append(out, strings.TrimSpace(s))
			}
		}
	case []string:
		out = append(out, x...)
	case string:
		if strings.TrimSpace(x) != "" {
			out = []string{strings.TrimSpace(x)}
		}
	case map[string]any:
		for k := range x {
			out = append(out, k)
		}
	}
	sort.Strings(out)
	return out
}

func mapString(v any) map[string]string {
	out := map[string]string{}
	if m, ok := v.(map[string]any); ok {
		for k, val := range m {
			if s, ok := val.(string); ok {
				out[k] = s
			}
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func intNumber(v any) int {
	switch x := v.(type) {
	case float64:
		return int(x)
	case int:
		return x
	case json.Number:
		i, _ := x.Int64()
		return int(i)
	}
	return 0
}

func depth(root, path string) int {
	rel, err := filepath.Rel(root, path)
	if err != nil || rel == "." {
		return 0
	}
	return len(strings.Split(filepath.ToSlash(rel), "/"))
}
