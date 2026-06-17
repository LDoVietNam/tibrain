---
tags: ["authentication", "tibrain", "security", "documentation", "javascript"]
scopes: ["cli", "auth", "tibrain"]
last_updated: 2026-05-22
---
# 9Router Deep Dive: OAuth & Token Refresh

> Files: `src/lib/oauth/providers.js`, `constants/oauth.js`, `services/*.js`, `open-sse/services/tokenRefresh.js`
> Date: 2026-04-28

---

## 1. PKCE Utilities (`utils/pkce.js`)

```javascript
generatePKCE() => {
  codeVerifier:  base64url(randomBytes(32)),   // 43 chars
  codeChallenge: base64url(sha256(verifier)),  // S256 method
  state:         base64url(randomBytes(32))    // CSRF
}
```

S256 method: SHA256 của verifier → base64url. Đây là chuẩn PKCE phổ biến nhất.

---

## 2. OAuthService Base Class (`services/oauth.js`)

Pattern DRY cho tất cả OAuth providers. 4 methods chính:

| Method | Mô tả |
|--------|-------|
| `buildAuthUrl()` | `client_id` + `response_type=code` + `redirect_uri` + `state` + `code_challenge` + `code_challenge_method=S256` + `extraParams` |
| `startAuthFlow()` | Khởi động local server random port, `setInterval(100ms)` poll callback params, timeout 5 phút |
| `exchangeCode()` | POST tokenUrl với `grant_type=authorization_code` + `code` + `code_verifier` + `redirect_uri`. Hỗ trợ cả JSON body và form-urlencoded |
| `authenticate()` | Full flow: `generatePKCE` → `startAuthFlow` → `open(browser)` → `waitCallback` → validate `state` → `exchangeCode` |

**Key**: `state` validation (dòng 145) — callback state PHẢI match generated state, chống CSRF.

---

## 3. 3 Loại OAuth Flow

### 3.1 PKCE Authorization Code Flow (phổ biến nhất)
- Claude, iFlow, Antigravity, Gemini, Codex, Cursor, Cline, KiloCode, GitLab
- Đặc điểm: Không cần client_secret, client secret embedded trong app không an toàn

### 3.2 Standard OAuth2 (có client_secret)
- Gemini, Antigravity (Google OAuth)
- Cần `client_secret` khi exchange code
- Refresh cũng cần `client_secret`

### 3.3 Device Code Flow
- GitHub Copilot, Qwen
- Quy trình:
  1. `getDeviceCode()` → nhận `device_code`, `user_code`, `verification_uri`
  2. In ra terminal: "Visit URL, enter code"
  3. `open(browser)` tự động (nếu được)
  4. `pollAccessToken()` → `while(true)` poll mỗi `interval` (default 5s)
  5. Xử lý các device code status:
     - `authorization_pending` → continue polling
     - `slow_down` → `interval += 5000`ms
     - `expired_token` → throw
     - `access_denied` → throw
     - `access_token` → success

### 3.4 Custom Flows
- **Qoder**: Device Token Flow với `deviceTokenUrl` + `deviceRefreshUrl` + `statusUrl`
- **Kiro**: Custom flow với platform enum (`darwin arm64=2`, `win32=5`)
- **GitHub Copilot**: 2-step (GitHub device code → get Copilot token riêng)

---

## 4. Provider Constants (`constants/oauth.js`)

15+ providers với config riêng. Ví dụ:

```javascript
// Claude — PKCE, JSON body
CLAUDE_CONFIG = {
  clientId: "9d1c250a-e61b-44d9-88ed-5944d1962f5e",
  authorizeUrl: "https://claude.ai/oauth/authorize",
  tokenUrl: "https://api.anthropic.com/v1/oauth/token",
  scopes: ["org:create_api_key", "user:profile", "user:inference"],
  codeChallengeMethod: "S256",
}

// GitHub Copilot — Device Code Flow
GITHUB_CONFIG = {
  clientId: "Iv1.b507a08c87ecfebf",
  deviceCodeUrl: "https://github.com/login/device/code",
  tokenUrl: "https://github.com/login/oauth/access_token",
  copilotTokenUrl: "https://api.github.com/copilot/v1/token",
  scopes: "read:user",
}

// Gemini — Standard OAuth2 + client_secret
GEMINI_CONFIG = {
  clientId: "...apps.googleusercontent.com",
  clientSecret: "GOCSPX-...",
  authorizeUrl: "https://accounts.google.com/o/oauth2/v2/auth",
  tokenUrl: "https://oauth2.googleapis.com/token",
  scopes: ["https://www.googleapis.com/auth/cloud-platform", "..."],
}
```

---

## 5. Token Refresh (`open-sse/services/tokenRefresh.js`)

### 5.1 General pattern
```javascript
// Refresh khi: expiresAt - now < TOKEN_EXPIRY_BUFFER_MS (5 phút)
if (!refreshToken || !config.refreshUrl) return null;

fetch(refreshUrl, {
  method: "POST",
  body: URLSearchParams({ grant_type: "refresh_token", refresh_token, client_id, client_secret })
})
// Fallback: giữ refreshToken cũ nếu provider không trả về mới
return { accessToken, refreshToken: tokens.refresh_token || refreshToken, expiresIn }
```

### 5.2 Per-provider specializations

| Provider | Đặc điểm refresh |
|----------|-----------------|
| Claude | JSON body (không form-urlencoded), không cần client_secret |
| Google (Gemini, Antigravity) | Cần `client_secret` |
| Qwen | Form-urlencoded |
| GitHub | Device flow → không có refresh token (token hết hạn thì re-auth) |
| Copilot | Dùng GitHub access_token để lấy Copilot token (không phải refresh_token) |

### 5.3 Refresh lead time
```javascript
REFRESH_LEAD_MS[provider] || TOKEN_EXPIRY_BUFFER_MS  // 5 phút default
```
Một số provider có thể cần buffer lớn hơn (ví dụ: token expire nhanh).

---

## 6. Account Identification

```javascript
function decodeJwtPayload(jwt) {
  const parts = jwt.split(".");
  const base64 = parts[1].replace(/-/g, "+").replace(/_/g, "/");
  const padded = base64 + "=".repeat((4 - (base64.length % 4)) % 4);
  return JSON.parse(Buffer.from(padded, "base64").toString("utf8"));
}

function extractEmailFromAccessToken(accessToken) {
  const payload = decodeJwtPayload(accessToken);
  return payload.email || payload.preferred_username || payload.sub;
}
```

Cách lấy account identity từ JWT access_token:
1. Split 3 phần (header.payload.signature)
2. Base64 decode phần 2 (handle url-safe base64: `-`→`+`, `_`→`/`)
3. Pad `=` đúng chuẩn
4. Parse JSON → lấy `email` → `preferred_username` → `sub`

---

## 7. Lessons cho Ti

1. **PKCE là chuẩn** cho CLI tools — không cần client_secret, server local callback, browser tự mở.
2. **Device Code Flow** là giải pháp cho headless/server không có browser (GitHub Copilot).
3. **State validation bắt buộc** — callback state PHẢI match, chống CSRF.
4. **Per-provider refresh specialization** — mỗi provider có API format khác nhau (Claude JSON, Google cần client_secret, Qwen form-urlencoded).
5. **Giữ refreshToken cũ nếu không có mới** — một số provider chỉ trả access_token mới, không trả refresh_token mới.
6. **JWT decode từ access_token** để lấy email/identity — không cần gọi userInfo API riêng nếu token đã có thông tin.
7. **GitHub Copilot 2-step**: GitHub device code → get Copilot token → dùng Copilot token cho API. Copilot token có expiry riêng.

---

## 8. So sánh với Ti

Ti router hiện tại có vẻ chỉ hỗ trợ **API key** (`ANTHROPIC_API_KEY`, `OPENROUTER_API_KEY`, etc.) — chưa có OAuth flow.

Nếu Ti muốn hỗ trợ:
- **Claude Code free tier** (via Claude OAuth)
- **Antigravity** (Google OAuth)
- **GitHub Copilot** (Device Code Flow)
- **Gemini CLI** (Google OAuth)

→ Cần implement OAuth system tương tự. Hoặc delegate cho external tool (như 9router dashboard).

**Recommendation**: Ti có thể **không cần OAuth** nếu chỉ dùng paid API keys. Nhưng nếu muốn free tier → cần học từ 9router.
