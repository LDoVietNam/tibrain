# auth2api Pattern Analysis

## Repository: auth2api
**URL:** https://github.com/AmazingAng/auth2api
**Location:** Z:\10_WORKPLACE\Ti\Ti-learning-lab\07_Repositories\router\auth2api

## Architecture Patterns

### 1. OAuth PKCE Flow Pattern
**Pattern:** PKCE (Proof Key for Code Exchange) implementation
```typescript
export function generateAuthURL(state: string, pkce: PKCECodes): string {
  const params = new URLSearchParams({
    code: "true",
    client_id: CLIENT_ID,
    response_type: "code",
    redirect_uri: REDIRECT_URI,
    code_challenge: pkce.codeChallenge,
    code_challenge_method: "S256",
    state,
  });
  
  // Append scope with colons unencoded
  const scopeEncoded = SCOPE.split(" ")
    .map((s) => encodeURIComponent(s).replace(/%3A/gi, ":"))
    .join("+");
  
  return `${AUTH_URL}?${params.toString()}&scope=${scopeEncoded}`;
}
```

**Key Learnings:**
- PKCE code challenge generation
- State parameter for CSRF protection
- Scope encoding with unencoded colons
- S256 hash method for code challenge

### 2. Token Exchange Pattern
**Pattern:** Authorization code to token exchange
```typescript
export async function exchangeCodeForTokens(
  code: string,
  returnedState: string,
  expectedState: string,
  pkce: PKCECodes,
): Promise<TokenData> {
  if (returnedState !== expectedState) {
    throw new Error("OAuth state mismatch — possible CSRF attack");
  }

  const resp = await fetch(TOKEN_URL, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({
      code,
      grant_type: "authorization_code",
      client_id: CLIENT_ID,
      redirect_uri: REDIRECT_URI,
      code_verifier: pkce.codeVerifier,
      state: expectedState,
    }),
  });

  const data: any = await resp.json();
  const expiresAt = new Date(Date.now() + data.expires_in * 1000).toISOString();

  return {
    accessToken: data.access_token,
    refreshToken: data.refresh_token,
    email: data.account?.email_address || "unknown",
    expiresAt,
    accountUuid: data.account?.uuid || "",
  };
}
```

**Key Learnings:**
- State validation for CSRF protection
- Code verifier for PKCE validation
- Expiration time calculation
- Account metadata extraction

### 3. Token Refresh Pattern
**Pattern:** Refresh token with retry logic
```typescript
export async function refreshTokens(refreshToken: string): Promise<TokenData> {
  const resp = await fetch(TOKEN_URL, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({
      client_id: CLIENT_ID,
      grant_type: "refresh_token",
      refresh_token: refreshToken,
    }),
  });

  if (!resp.ok) {
    const text = await resp.text();
    const reason = detectExhaustedReason(text);
    if (reason) {
      throw new RefreshTokenExhaustedError(reason, resp.status, text);
    }
    throw new Error(`Token refresh failed (${resp.status}): ${text}`);
  }

  const data: any = await resp.json();
  const expiresAt = new Date(Date.now() + data.expires_in * 1000).toISOString();

  return {
    accessToken: data.access_token,
    refreshToken: data.refresh_token,
    email: data.account?.email_address || "unknown",
    expiresAt,
    accountUuid: data.account?.uuid || "",
  };
}
```

**Key Learnings:**
- Refresh token exchange
- Exhausted token detection
- Expiration time recalculation
- Error handling for token exhaustion

### 4. Retry with Backoff Pattern
**Pattern:** Exponential backoff for token refresh
```typescript
export async function refreshTokensWithRetry(
  refreshToken: string,
  maxRetries = 3,
): Promise<TokenData> {
  for (let attempt = 1; ; attempt++) {
    try {
      return await refreshTokens(refreshToken);
    } catch (err) {
      // Refresh token revoked/expired/reused — do not retry
      if (err instanceof RefreshTokenExhaustedError) throw err;
      if (attempt >= maxRetries) throw err;
      await timeout(attempt * 1000);  // Exponential backoff
    }
  }
}
```

**Key Learnings:**
- Exponential backoff strategy
- Non-retryable error detection
- Max retry limit
- Timeout between retries

### 5. Multi-Provider Architecture Pattern
**Pattern:** Provider registry with routing
```typescript
// providers/registry.ts
export const PROVIDERS: Record<string, Provider> = {
  anthropic: anthropicProvider,
  codex: codexProvider,
  cursor: cursorProvider,
};

export function getProvider(name: string): Provider {
  const provider = PROVIDERS[name];
  if (!provider) {
    throw new Error(`Unknown provider: ${name}`);
  }
  return provider;
}
```

**Key Learnings:**
- Provider registry pattern
- Dynamic provider lookup
- Error handling for unknown providers
- Extensible provider system

### 6. Account Management Pattern
**Pattern:** Multi-account pooling with sticky routing
```typescript
// accounts/manager.ts
export class AccountManager {
  private accounts: Map<string, Account> = new Map();
  private stickyRouting: Map<string, string> = new Map();  // session -> account
  
  async getAccount(sessionId: string): Promise<Account> {
    // Sticky routing: use same account for session
    if (this.stickyRouting.has(sessionId)) {
      const accountId = this.stickyRouting.get(sessionId);
      return this.accounts.get(accountId);
    }
    
    // Load balancing: select account based on health
    const healthyAccounts = this.getHealthyAccounts();
    const account = this.selectAccount(healthyAccounts);
    this.stickyRouting.set(sessionId, account.id);
    return account;
  }
  
  markUnhealthy(accountId: string, reason: string) {
    const account = this.accounts.get(accountId);
    account.status = 'unhealthy';
    account.lastError = reason;
    account.cooldownUntil = Date.now() + COOLDOWN_MS;
  }
}
```

**Key Learnings:**
- Session-based sticky routing
- Health-based account selection
- Cooldown mechanism for unhealthy accounts
- Load balancing across accounts

### 7. Stats Recording Pattern
**Pattern:** Per-account usage tracking
```typescript
// stats/recorder.ts
export class StatsRecorder {
  private stats: Map<string, AccountStats> = new Map();
  
  recordUsage(accountId: string, model: string, tokens: number) {
    const stats = this.stats.get(accountId) || this.initStats(accountId);
    stats.requestCount++;
    stats.tokenCount += tokens;
    stats.modelUsage[model] = (stats.modelUsage[model] || 0) + tokens;
    stats.lastUsed = Date.now();
  }
  
  getStats(accountId: string): AccountStats {
    return this.stats.get(accountId);
  }
}
```

**Key Learnings:**
- Per-account usage tracking
- Model-specific token counting
- Last used timestamp
- Request counting

### 8. Concurrency Lock Pattern
**Pattern:** Token refresh with concurrency control
```typescript
export class TokenStorage {
  private refreshLock: Map<string, Promise<TokenData>> = new Map();
  
  async getValidToken(accountId: string): Promise<string> {
    const account = this.getAccount(accountId);
    
    // Check if token is still valid
    if (this.isTokenValid(account.expiresAt)) {
      return account.accessToken;
    }
    
    // Check if refresh is already in progress
    if (this.refreshLock.has(accountId)) {
      return this.refreshLock.get(accountId).then(data => data.accessToken);
    }
    
    // Start refresh with lock
    const refreshPromise = this.refreshTokensWithRetry(account.refreshToken);
    this.refreshLock.set(accountId, refreshPromise);
    
    try {
      const data = await refreshPromise;
      this.updateAccount(accountId, data);
      return data.accessToken;
    } finally {
      this.refreshLock.delete(accountId);
    }
  }
}
```

**Key Learnings:**
- Concurrency lock for token refresh
- Prevents duplicate refresh requests
- Wait for in-progress refresh
- Lock cleanup after completion

## Implementation Recommendations for Ti Router

### 1. OAuth PKCE Flow
- Implement PKCE for OAuth flows
- Add state parameter for CSRF protection
- Handle scope encoding properly
- Validate code verifier

### 2. Token Management
- Implement token refresh with retry
- Add exponential backoff
- Detect exhausted tokens
- Calculate expiration times

### 3. Multi-Provider Architecture
- Implement provider registry
- Dynamic provider lookup
- Extensible provider system
- Error handling for unknown providers

### 4. Account Management
- Implement multi-account pooling
- Add sticky routing for sessions
- Health-based account selection
- Cooldown mechanism

### 5. Usage Tracking
- Implement per-account stats
- Track token usage by model
- Record request counts
- Last used timestamps

### 6. Concurrency Control
- Implement concurrency locks
- Prevent duplicate operations
- Wait for in-progress operations
- Lock cleanup

## Priority
High - Apply these patterns for OAuth authentication and multi-account management.
