---
tags: ["tibrain", "authentication", "security", "testing", "documentation"]
scopes: ["code", "auth", "tibrain"]
last_updated: 2026-05-22
---
# Windsurf OAuth Integration

## Overview

Windsurf OAuth integration cho phép người dùng import OTT (One-Time Token) từ `windsurf.com/show-auth-token` để sử dụng Windsurf API.

## Architecture

### Windsurf OAuth Pattern

Khác với OAuth chuẩn (authorization code flow), Windsurf sử dụng **OTT token pattern**:

1. **User truy cập** `https://windsurf.com/show-auth-token` để lấy token
2. **Token format**: `ott$...` (One-Time Token)
3. **Validation**: Token được validate bằng cách gọi Windsurf API `/v1/models`
4. **Storage**: Token được lưu trong CredentialStore
5. **Refresh**: OTT tokens không thể refresh, user cần lấy token mới khi hết hạn

### Components

#### 1. WindsurfOAuthProvider (`layers/authentication/windsurfoauth.go`)

```go
type WindsurfOAuthProvider struct{}

func (p *WindsurfOAuthProvider) Name() string { return "windsurf" }
func (p *WindsurfOAuthProvider) AuthorizeURL(state, redirectURI string) string {
    return "https://windsurf.com/show-auth-token"
}
func (p *WindsurfOAuthProvider) Exchange(ctx context.Context, token, redirectURI string) (*OAuthToken, error)
func (p *WindsurfOAuthProvider) Refresh(ctx context.Context, refreshToken string) (*OAuthToken, error) {
    return nil, fmt.Errorf("Windsurf OTT tokens cannot be refresh")
}
```

**Key points:**
- `AuthorizeURL` trả về URL để user lấy token (không phải OAuth authorize URL)
- `Exchange` validate OTT token bằng cách gọi API
- `Refresh` không hỗ trợ (OTT tokens không thể refresh)

#### 2. WindsurfImportHandler (`layers/authentication/oauth_handlers.go`)

```go
func (h *OAuthHandlers) WindsurfImportHandler(w http.ResponseWriter, r *http.Request)
```

**Flow:**
1. Nhận OTT token từ request body: `{"token": "ott$..."}`
2. Validate token bằng `ValidateWindsurfToken()`
3. Lưu token vào CredentialStore
4. Trả về token response

**Endpoint:** `POST /oauth/windsurf/import`

#### 3. Registration (`cmd/routerd/main.go`)

```go
// Windsurf OAuth (always available)
windsurfProvider := &authentication.WindsurfOAuthProvider{}
oauthRegistry.Register(windsurfProvider)
log.Printf("OAuth provider registered: windsurf")
```

Windsurf OAuth provider luôn available, không cần environment variables.

## Usage

### 1. Get Windsurf Token

Truy cập `https://windsurf.com/show-auth-token` trong browser và copy token.

Token format: `ott$xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx`

### 2. Import Token via API

```bash
curl -X POST http://localhost:1807/oauth/windsurf/import \
  -H "Content-Type: application/json" \
  -d '{"token": "ott$your-token-here"}'
```

**Response:**
```json
{
  "access_token": "ott$your-token-here",
  "token_type": "Bearer",
  "expires_in": 86400
}
```

### 3. Use Windsurf Provider

Token được lưu trong CredentialStore và WindsurfProvider sẽ tự động sử dụng token này cho API calls.

```bash
curl http://localhost:1807/v1/chat/completions \
  -H "Authorization: Bearer sk-jarvis-dev" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "claude-4.5-sonnet-thinking",
    "messages": [{"role": "user", "content": "Hello"}]
  }'
```

## Implementation Details

### Changes Made

1. **OAuth Handlers Registration** (`cmd/routerd/main.go`)
   - Thay đổi từ custom routing sang `RegisterRoutes()` pattern
   - Sử dụng `InMemoryCredentialStore` thay vì `MemorySessionStore`
   - Fix type mismatch với auditor middleware

2. **WindsurfImportHandler** (`layers/authentication/oauth_handlers.go`)
   - Thêm handler mới cho `/oauth/windsurf/import`
   - Validate token via `ValidateWindsurfToken()`
   - Store token in CredentialStore

3. **Build Fixes**
   - Fix `RoutingError` duplicate definition (selector.go vs adaptive_router.go)
   - Fix `ProviderCandidate` undefined (import autocombo package)
   - Fix `CredentialStore` type mismatch (use InMemoryCredentialStore)

### Files Modified

- `cmd/routerd/main.go` - OAuth handlers registration
- `layers/authentication/oauth_handlers.go` - WindsurfImportHandler
- `layers/routing/selector.go` - Remove duplicate RoutingError
- `layers/routing/adaptive_router.go` - Import autocombo, fix ProviderCandidate
- `layers/routing/reaction_engine.go` - Import autocombo, fix ProviderCandidate

### Files Existing (No Changes)

- `layers/authentication/windsurfoauth.go` - WindsurfOAuthProvider implementation
- `layers/provider/windsurf.go` - WindsurfProvider (uses OTT tokens)

## Security Considerations

1. **Token Validation**: OTT tokens được validate trước khi lưu
2. **Token Storage**: Tokens được lưu trong CredentialStore (in-memory for now)
3. **No Refresh**: OTT tokens không thể refresh, user cần lấy token mới khi hết hạn
4. **HTTPS**: Production nên sử dụng HTTPS để bảo vệ tokens

## Testing

### Manual Test

```bash
# 1. Start router
cd Z:\Ti\router
.\bin\routerd.exe -config configs\Tiserverrouter.yaml

# 2. Get Windsurf token from browser
# https://windsurf.com/show-auth-token

# 3. Import token
curl -X POST http://localhost:1807/oauth/windsurf/import \
  -H "Content-Type: application/json" \
  -d '{"token": "ott$your-token"}'

# 4. Test Windsurf API
curl http://localhost:1807/v1/chat/completions \
  -H "Authorization: Bearer sk-jarvis-dev" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "claude-4.5-sonnet-thinking",
    "messages": [{"role": "user", "content": "Hello"}]
  }'
```

## Future Improvements

1. **Persistent CredentialStore**: Sử dụng database thay vì in-memory store
2. **Token Health Check**: Auto-validate tokens periodically
3. **Token Expiry Alert**: Notify user khi token sắp hết hạn
4. **Multi-token Support**: Hỗ trợ nhiều Windsurf tokens
5. **Token Rotation**: Auto-rotate tokens nếu có multiple tokens

## References

- Windsurf API: https://server.windsurf.com/v1
- Windsurf Auth: https://windsurf.com/show-auth-token
- OmniRoute Issue #1679: Windsurf integration
- OAuth Pattern: `layers/authentication/oauth.go`
