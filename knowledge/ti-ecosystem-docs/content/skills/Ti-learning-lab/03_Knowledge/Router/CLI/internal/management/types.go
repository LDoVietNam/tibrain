package management

import "time"

type AuthPaths struct {
	Root    string `json:"root"`
	Cookies string `json:"cookies"`
	Auth    string `json:"auth"`
	OAuth   string `json:"oauth"`
}

type ProviderRow struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Tier     string `json:"tier"`
	Auth     string `json:"auth"`
	Protocol string `json:"protocol"`
	Group    string `json:"group"`
	Enabled  bool   `json:"enabled"`
}

type APIKeyRow struct {
	Name   string `json:"name"`
	Prefix string `json:"prefix"`
	Scope  string `json:"scope"`
	Status string `json:"status"`
}

type AuthFileRow struct {
	File     string `json:"file"`
	Provider string `json:"provider"`
	Status   string `json:"status"`
	Updated  string `json:"updated"`
	Kind     string `json:"kind,omitempty"`
	Path     string `json:"path,omitempty"`
}

type CookieRow struct {
	File     string `json:"file"`
	Provider string `json:"provider"`
	Status   string `json:"status"`
	Updated  string `json:"updated"`
	Path     string `json:"path,omitempty"`
}

type OAuthConnectionRow struct {
	Name    string `json:"name"`
	State   string `json:"state"`
	Account string `json:"account"`
	Updated string `json:"updated,omitempty"`
	Path    string `json:"path,omitempty"`
}

type OAuthStarter struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	Flow          string `json:"flow"`
	AuthType      string `json:"authType"`
	Artifact      string `json:"artifact"`
	Storage       string `json:"storage"`
	ProviderScope string `json:"providerScope"`
	Note          string `json:"note"`
	Status        string `json:"status"`
}

type CookieStarter struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	Source        string `json:"source"`
	Storage       string `json:"storage"`
	ProviderScope string `json:"providerScope"`
	Note          string `json:"note"`
	Status        string `json:"status"`
}

type Dashboard struct {
	Providers      int    `json:"providers"`
	AuthFiles      int    `json:"authFiles"`
	CookieJars     int    `json:"cookieJars"`
	OAuthLanes     int    `json:"oauthLanes"`
	APIKeys        int    `json:"apiKeys"`
	AuthPathsValid bool   `json:"authPathsValid"`
	AuthRoot       string `json:"authRoot"`
}

type SystemInfo struct {
	Version   string    `json:"version"`
	BuildDate string    `json:"buildDate"`
	Now       time.Time `json:"now"`
	ModelPath string    `json:"modelPath"`
}

type ConfigView struct {
	Host            string    `json:"host"`
	Port            int       `json:"port"`
	RateLimit       int       `json:"rateLimit"`
	Provider        string    `json:"provider"`
	Model           string    `json:"model"`
	AuthPaths       AuthPaths `json:"authPaths"`
	RoutingStrategy string    `json:"routingStrategy"`
}

type Bootstrap struct {
	Dashboard      Dashboard            `json:"dashboard"`
	AuthPaths      AuthPaths            `json:"authPaths"`
	Providers      []ProviderRow        `json:"providers"`
	APIKeys        []APIKeyRow          `json:"apiKeys"`
	AuthFiles      []AuthFileRow        `json:"authFiles"`
	Cookies        []CookieRow          `json:"cookies"`
	OAuth          []OAuthConnectionRow `json:"oauth"`
	OAuthStarters  []OAuthStarter       `json:"oauthStarters"`
	CookieStarters []CookieStarter      `json:"cookieStarters"`
	Config         ConfigView           `json:"config"`
	System         SystemInfo           `json:"system"`
}

// --- POST request bodies ---

type AddAPIKeyRequest struct {
	Provider string `json:"provider"` // e.g. "openrouter"
	Key      string `json:"key"`      // raw API key value
}

type AddAuthFileRequest struct {
	Provider string `json:"provider"` // e.g. "anthropic"
	Kind     string `json:"kind"`     // "token" | "oauth" | "cookie"
	Content  string `json:"content"`  // file content (JSON or plain text)
}

type AddCookieRequest struct {
	Provider string `json:"provider"` // e.g. "iflow"
	Content  string `json:"content"`  // JSON cookie jar content
}

type AddOAuthRequest struct {
	Provider string         `json:"provider"` // e.g. "claude-code"
	Data     map[string]any `json:"data"`     // arbitrary OAuth payload
}

type WriteResult struct {
	OK   bool   `json:"ok"`
	Path string `json:"path,omitempty"`
	Msg  string `json:"msg,omitempty"`
}
