---
tags: ["tibrain", "authentication", "security", "testing", "mcp"]
scopes: ["integration", "code", "auth", "tibrain"]
last_updated: 2026-05-22
---
# MCP OAuth Integration

> **Component**: `internal/mcp/oauth.go`
> **Status**: ✅ Implemented & Tested

## Overview

The MCP OAuth Integration provides a secure, standardized way to handle OAuth 2.0 authentication for MCP (Model Context Protocol) servers. It implements a callback handler with state management, token storage, and HTTP server for OAuth flows.

## Architecture

### Core Components

#### OAuthState
```go
type OAuthState struct {
    StateToken    string    // Unique state token for CSRF protection
    ServerName    string    // MCP server name
    RedirectURI   string    // OAuth redirect URI
    CreatedAt    time.Time // State creation time
    ExpiresAt    time.Time // State expiration time (10 minutes)
    AccessToken  string    // OAuth access token (after callback)
    RefreshToken string    // OAuth refresh token (after callback)
    TokenType    string    // Token type (e.g., "Bearer")
    ExpiresIn    int       // Token expiration in seconds
}
```

Represents an OAuth flow state with CSRF protection and token storage.

#### OAuthConfig
```go
type OAuthConfig struct {
    ClientID     string   // OAuth client ID
    ClientSecret string   // OAuth client secret
    AuthURL      string   // Authorization URL
    TokenURL     string   // Token URL
    Scopes       []string // OAuth scopes
    RedirectURI  string   // OAuth redirect URI
}
```

OAuth 2.0 configuration for MCP servers.

#### OAuthCallbackHandler
```go
type OAuthCallbackHandler struct {
    states       map[string]*OAuthState
    statesMu     sync.RWMutex
    callbackPort int
    logger       Logger
    server       *http.Server
}
```

HTTP server for handling OAuth callbacks with thread-safe state management.

### HTTP Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/oauth/callback` | GET | OAuth callback endpoint |
| `/oauth/health` | GET | Health check endpoint |

## Usage

### Initializing OAuth Handler

```go
import (
    "github.com/ti/cli/internal/mcp"
)

// Create OAuth callback handler
handler := mcp.NewOAuthCallbackHandler(8080, logger)

// Start the HTTP server
err := handler.Start(ctx)
if err != nil {
    log.Fatal(err)
}

// Stop the server when done
defer handler.Stop(ctx)
```

### Initiating OAuth Flow

```go
// Create OAuth configuration
oauthConfig := &mcp.OAuthConfig{
    ClientID:     "your-client-id",
    ClientSecret: "your-client-secret",
    AuthURL:      "https://auth.example.com/authorize",
    TokenURL:     "https://auth.example.com/token",
    Scopes:       []string{"read", "write"},
    RedirectURI:  "http://localhost:8080/oauth/callback",
}

// Add OAuth config to MCP server
serverConfig := &mcp.ServerConfig{
    Name:        "my-mcp-server",
    Transport:   mcp.TransportHTTP,
    URL:         "https://api.example.com",
    OAuth:       oauthConfig,
}

// Initiate OAuth flow
manager := mcp.NewManager()
authURL, stateToken, err := manager.InitiateOAuth("my-mcp-server")
if err != nil {
    log.Fatal(err)
}

// Redirect user to authURL
fmt.Printf("Visit: %s\n", authURL)
```

### Handling OAuth Callback

The OAuth callback handler automatically handles the callback:

1. User is redirected to authorization URL
2. User grants permission
3. OAuth provider redirects to `/oauth/callback`
4. Handler validates state token
5. Handler exchanges authorization code for tokens
6. Tokens are stored in the OAuth state

### Retrieving OAuth Tokens

```go
// Get OAuth tokens for a server
tokens, err := manager.GetOAuthTokens("my-mcp-server")
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Access Token: %s\n", tokens.AccessToken)
fmt.Printf("Refresh Token: %s\n", tokens.RefreshToken)
fmt.Printf("Token Type: %s\n", tokens.TokenType)
fmt.Printf("Expires In: %d\n", tokens.ExpiresIn)
```

### Using OAuth Tokens with MCP Connections

```go
// Connect to MCP server with OAuth
client, err := manager.Connect("my-mcp-server")
if err != nil {
    log.Fatal(err)
}

// OAuth tokens are automatically used for authentication
// during MCP protocol handshake
```

## State Management

### Creating OAuth State

```go
// Create a new OAuth state
stateToken, err := handler.CreateState("my-mcp-server", "http://localhost:8080/oauth/callback")
if err != nil {
    log.Fatal(err)
}

// State token expires in 10 minutes
```

### Validating OAuth State

```go
// Validate state token exists and is not expired
state, exists := handler.GetState(stateToken)
if !exists {
    log.Fatal("Invalid state token")
}

if time.Now().After(state.ExpiresAt) {
    log.Fatal("State token expired")
}
```

### Completing OAuth Flow

```go
// Complete OAuth flow with authorization code
err := handler.CompleteOAuth(stateToken, authCode)
if err != nil {
    log.Fatal(err)
}

// Tokens are now stored in the state
state, _ := handler.GetState(stateToken)
fmt.Printf("Access Token: %s\n", state.AccessToken)
```

## Storage Integration

OAuth configurations and tokens are stored in SQLite:

```go
// Storage schema
CREATE TABLE mcp_servers (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT UNIQUE NOT NULL,
    transport TEXT NOT NULL,
    command TEXT,
    args TEXT,  -- JSON array
    url TEXT,
    env TEXT,   -- JSON object
    permissions TEXT,  -- JSON object
    enabled INTEGER DEFAULT 1,
    auto_connect INTEGER DEFAULT 0,
    timeout INTEGER,
    max_retries INTEGER DEFAULT 3,
    retry_delay INTEGER DEFAULT 5000,
    oauth TEXT  -- JSON object (OAuthConfig)
);
```

### Saving OAuth Configuration

```go
storage := mcp.NewStorage("mcp.db")

serverConfig := &mcp.ServerConfig{
    Name:      "my-mcp-server",
    Transport: mcp.TransportHTTP,
    URL:       "https://api.example.com",
    OAuth: &mcp.OAuthConfig{
        ClientID:     "your-client-id",
        ClientSecret: "your-client-secret",
        AuthURL:      "https://auth.example.com/authorize",
        TokenURL:     "https://auth.example.com/token",
        Scopes:       []string{"read", "write"},
        RedirectURI:  "http://localhost:8080/oauth/callback",
    },
}

err := storage.SaveServerConfig(serverConfig)
if err != nil {
    log.Fatal(err)
}
```

### Loading OAuth Configuration

```go
config, err := storage.GetServerConfig("my-mcp-server")
if err != nil {
    log.Fatal(err)
}

if config.OAuth != nil {
    fmt.Printf("OAuth Client ID: %s\n", config.OAuth.ClientID)
    fmt.Printf("OAuth Scopes: %v\n", config.OAuth.Scopes)
}
```

## Security Features

### CSRF Protection

State tokens provide CSRF protection:

```go
// State tokens are cryptographically random
stateToken, _ := generateStateToken()  // 32-byte random token

// State tokens expire after 10 minutes
expiresAt := time.Now().Add(10 * time.Minute)
```

### Token Storage

OAuth tokens are stored securely:

```go
// Tokens are stored in OAuthState
state.AccessToken = "access_token_from_provider"
state.RefreshToken = "refresh_token_from_provider"

// States are managed in-memory with mutex lock
handler.statesMu.Lock()
handler.states[stateToken] = state
handler.statesMu.Unlock()
```

### Secure HTTP

HTTPS is recommended for production:

```go
// Configure TLS for production
server := &http.Server{
    Addr:    ":443",
    Handler: mux,
    TLSConfig: &tls.Config{
        MinVersion: tls.VersionTLS12,
    },
}
```

## Best Practices

### 1. Use HTTPS in Production

```go
// Development
handler := mcp.NewOAuthCallbackHandler(8080, logger)

// Production
handler := mcp.NewOAuthCallbackHandler(443, logger)
// Configure TLS certificates
```

### 2. Validate Redirect URIs

```go
// Ensure redirect URI matches expected
if state.RedirectURI != expectedRedirectURI {
    log.Fatal("Invalid redirect URI")
}
```

### 3. Handle Token Expiration

```go
// Check token expiration
if time.Now().After(tokenExpiresAt) {
    // Use refresh token to get new access token
    newTokens, err := refreshTokens(state.RefreshToken)
    if err != nil {
        log.Fatal("Failed to refresh token")
    }
}
```

### 4. Clean Up Expired States

```go
// Clean up expired states periodically
func (h *OAuthCallbackHandler) CleanupExpiredStates() {
    h.statesMu.Lock()
    defer h.statesMu.Unlock()

    now := time.Now()
    for token, state := range h.states {
        if now.After(state.ExpiresAt) {
            delete(h.states, token)
        }
    }
}
```

### 5. Log OAuth Events

```go
// Log OAuth flow events for debugging
h.logger.Info("OAuth state created: %s for server: %s", stateToken, serverName)
h.logger.Info("OAuth callback received: state=%s", stateToken)
h.logger.Info("OAuth flow completed: state=%s", stateToken)
```

## Examples

### Example 1: GitHub OAuth Integration

```go
oauthConfig := &mcp.OAuthConfig{
    ClientID:     os.Getenv("GITHUB_CLIENT_ID"),
    ClientSecret: os.Getenv("GITHUB_CLIENT_SECRET"),
    AuthURL:      "https://github.com/login/oauth/authorize",
    TokenURL:     "https://github.com/login/oauth/access_token",
    Scopes:       []string{"repo", "user"},
    RedirectURI:  "http://localhost:8080/oauth/callback",
}

serverConfig := &mcp.ServerConfig{
    Name:      "github-mcp",
    Transport: mcp.TransportHTTP,
    URL:       "https://api.github.com",
    OAuth:     oauthConfig,
}
```

### Example 2: Google OAuth Integration

```go
oauthConfig := &mcp.OAuthConfig{
    ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
    ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
    AuthURL:      "https://accounts.google.com/o/oauth2/v2/auth",
    TokenURL:     "https://oauth2.googleapis.com/token",
    Scopes:       []string{"https://www.googleapis.com/auth/drive.readonly"},
    RedirectURI:  "http://localhost:8080/oauth/callback",
}

serverConfig := &mcp.ServerConfig{
    Name:      "google-drive-mcp",
    Transport: mcp.TransportHTTP,
    URL:       "https://www.googleapis.com/drive/v3",
    OAuth:     oauthConfig,
}
```

### Example 3: Custom OAuth Provider

```go
oauthConfig := &mcp.OAuthConfig{
    ClientID:     "custom-client-id",
    ClientSecret: "custom-client-secret",
    AuthURL:      "https://custom-auth.example.com/authorize",
    TokenURL:     "https://custom-auth.example.com/token",
    Scopes:       []string{"api.read", "api.write"},
    RedirectURI:  "http://localhost:8080/oauth/callback",
}

serverConfig := &mcp.ServerConfig{
    Name:      "custom-api-mcp",
    Transport: mcp.TransportHTTP,
    URL:       "https://api.example.com",
    OAuth:     oauthConfig,
}
```

## Testing

```go
func TestOAuthHandler(t *testing.T) {
    handler := mcp.NewOAuthCallbackHandler(8080, logger)

    // Test state creation
    stateToken, err := handler.CreateState("test-server", "http://localhost:8080/oauth/callback")
    if err != nil {
        t.Fatal(err)
    }

    // Test state retrieval
    state, exists := handler.GetState(stateToken)
    if !exists {
        t.Fatal("State not found")
    }

    if state.ServerName != "test-server" {
        t.Errorf("Expected server name 'test-server', got '%s'", state.ServerName)
    }

    // Test state expiration
    if time.Now().After(state.ExpiresAt) {
        t.Error("State should not be expired immediately")
    }
}
```

## Error Handling

### Common OAuth Errors

| Error | Description | Solution |
|-------|-------------|----------|
| Invalid state token | State token not found or expired | Re-initiate OAuth flow |
| Invalid authorization code | Authorization code rejected | Check OAuth configuration |
| Token exchange failed | Failed to exchange code for tokens | Check token URL and credentials |
| Redirect URI mismatch | Redirect URI doesn't match configuration | Update OAuth config |

### Error Handling Example

```go
authURL, stateToken, err := manager.InitiateOAuth("my-mcp-server")
if err != nil {
    if strings.Contains(err.Error(), "no OAuth config") {
        log.Fatal("OAuth not configured for this server")
    }
    log.Fatal(err)
}

fmt.Printf("Visit: %s\n", authURL)
```

## Migration from Manual Token Management

### Before (Old Pattern)

```go
// Manual token management
accessToken := os.Getenv("MCP_ACCESS_TOKEN")
if accessToken == "" {
    log.Fatal("Access token required")
}

// Use token in requests
req.Header.Set("Authorization", "Bearer " + accessToken)
```

### After (New Pattern)

```go
// OAuth-based token management
oauthConfig := &mcp.OAuthConfig{
    ClientID:     "client-id",
    ClientSecret: "client-secret",
    AuthURL:      "https://auth.example.com/authorize",
    TokenURL:     "https://auth.example.com/token",
    Scopes:       []string{"read"},
    RedirectURI:  "http://localhost:8080/oauth/callback",
}

// OAuth flow handles token management automatically
authURL, _, _ := manager.InitiateOAuth("my-mcp-server")
// User visits authURL, grants permission
// Tokens are stored and used automatically
```

## Benefits

1. **Security**: CSRF protection with state tokens
2. **Standardization**: OAuth 2.0 compliant implementation
3. **Convenience**: Automatic token management
4. **Persistence**: Token storage in SQLite
5. **Thread-Safe**: Safe for concurrent OAuth flows
6. **Extensibility**: Support for any OAuth 2.0 provider
7. **Integration**: Seamless MCP server integration

## See Also

- [MCP Documentation](../content/mcp/README.md)
- [Tool System](CLI_TOOL_SYSTEM.md)
- [CLI Guide](../CLI_GUIDE.md)
