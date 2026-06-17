# JWT Authentication Pattern

> **Category**: Security  
> **Language**: Tiếng Việt  
> **Last Updated**: 2026-04-30

---

## Tổng Quan

JWT (JSON Web Token) là standard cho stateless authentication. Token được signed với secret key, chứa claims về user, và được validate bằng signature verification.

---

## Khi Nào Sử Dụng

- Stateless authentication (không cần session storage)
- Distributed systems (multiple servers)
- Mobile apps (token stored client-side)
- Microservices (cross-service authentication)
- API authentication

---

## Cấu Trúc JWT

### JWT Format

```
HEADER.PAYLOAD.SIGNATURE
```

**Example**:
```
eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c
```

### Header

```json
{
  "alg": "HS256",  // Algorithm (HS256, RS256, etc.)
  "typ": "JWT"     // Type
}
```

### Payload

```json
{
  "iss": "issuer",          // Issuer
  "sub": "user_id",        // Subject (user ID)
  "aud": "audience",       // Audience
  "exp": 1234567890,       // Expiration time
  "nbf": 1234567890,       // Not before
  "iat": 1234567890,       // Issued at
  "jti": "token_id",       // JWT ID
  "custom_claims": {...}   // Custom claims
}
```

### Signature

```
HMACSHA256(
  base64UrlEncode(header) + "." + base64UrlEncode(payload),
  secret_key
)
```

---

## Algorithms

### 1. HS256 (HMAC-SHA256)
- Symmetric encryption (same secret cho signing và verification)
- Dễ implement
- Phù hợp cho single-server applications
- **Risk**: Nếu secret bị compromise, attacker có thể forge tokens

### 2. RS256 (RSA-SHA256)
- Asymmetric encryption (private key signing, public key verification)
- Phù hợp cho distributed systems
- **Best Practice**: Cho production

### 3. ES256 (ECDSA-SHA256)
- Asymmetric với elliptic curve
- Smaller signatures
- More efficient

---

## Implementation Pattern

### 1. Create JWT Manager

```go
manager := NewJWTManager(
    secret_key,  // Secret key hoặc private key
    issuer,      // Issuer identifier
)
```

### 2. Generate Token

```go
token, err := manager.GenerateToken(
    user_id,           // Subject
    expiration,        // Token expiration (e.g., 1 hour)
    custom_claims,     // Custom claims (optional)
)
```

### 3. Validate Token

```go
payload, err := manager.ValidateToken(token)
if err != nil {
    // Invalid token
}
user_id := payload.Sub
```

### 4. Refresh Token

```go
new_token, err := manager.RefreshToken(token, new_expiration)
```

---

## Best Practices

### 1. Use Strong Secret Keys
- Minimum 256 bits (32 bytes) cho HMAC
- Use cryptographically secure random generator
- Rotate keys periodically
- Never commit secrets to git

### 2. Set Appropriate Expiration
- Short expiration cho sensitive operations (5-15 minutes)
- Longer expiration cho less sensitive (1-24 hours)
- Use refresh tokens cho long-lived sessions
- Never use permanent tokens

### 3. Include Relevant Claims
- Always include `sub` (subject/user ID)
- Include `exp` (expiration)
- Include `iat` (issued at)
- Include `jti` (token ID) cho revocation
- Include custom claims cho business logic

### 4. Validate All Claims
- Validate signature
- Validate expiration
- Validate issuer
- Validate audience
- Validate custom claims

### 5. Implement Token Revocation
- Maintain blacklist của revoked token IDs
- Use short expiration + refresh tokens
- Implement logout endpoint (add to blacklist)

---

## Common Mistakes

### 1. Weak Secret Keys
- **Problem**: Secret key quá weak hoặc predictable
- **Solution**: Use cryptographically secure random key, minimum 256 bits

### 2. Long Expiration
- **Problem**: Token expiration quá dài (months/years)
- **Solution**: Use short expiration (1-24 hours) với refresh tokens

### 3. No Revocation
- **Problem**: Không có cách revoke tokens
- **Solution**: Implement token blacklist hoặc use short expiration

### 4. Storing Sensitive Data in Token
- **Problem**: Store passwords, secrets trong JWT payload
- **Solution**: JWT payload không encrypted, chỉ store non-sensitive data

### 5. Not Validating All Claims
- **Problem**: Chỉ validate signature, không validate expiration/issuer
- **Solution**: Validate tất cả claims (signature, exp, iss, aud, nbf)

---

## Advanced Patterns

### 1. Refresh Tokens
- Short-lived access token (15-60 minutes)
- Long-lived refresh token (days/weeks)
- Access token expired → use refresh token để get new access token
- Useful cho mobile apps và web apps

### 2. Token Blacklist
- Maintain blacklist của revoked token IDs
- Check blacklist khi validate token
- Use Redis cho distributed blacklist
- Useful cho immediate revocation

### 3. Key Rotation
- Rotate keys periodically (e.g., every 90 days)
- Support multiple keys (old và new)
- Graceful transition period
- Useful cho security

### 4. JWKS (JSON Web Key Set)
- Public keys stored in JWKS endpoint
- Automatic key rotation
- Useful cho distributed systems

---

## Tools & Libraries

### Go Libraries
- **github.com/golang-jwt/jwt**: Popular JWT library cho Go
- **github.com/dgrijalva/jwt-go**: Older JWT library (deprecated)
- **github.com/lestrrat-go/jwx**: Advanced JWT library

### Other Languages
- **JavaScript**: jsonwebtoken (npm)
- **Python**: PyJWT (pip)
- **Java**: java-jwt (Maven)
- **Ruby**: ruby-jwt (gem)

### Tools
- **jwt.io**: JWT debugger và validator
- **jwt.ms**: JWT debugger

---

## Security Considerations

### 1. Secret Management
- Store secrets in environment variables
- Use secret management service (AWS Secrets Manager, HashiCorp Vault)
- Never commit secrets to git
- Rotate secrets regularly

### 2. HTTPS Required
- Always use HTTPS cho JWT transmission
- Prevent man-in-the-middle attacks
- Enforce TLS 1.2+

### 3. Token Storage
- Store tokens securely (HttpOnly cookies, secure storage)
- Avoid localStorage cho sensitive tokens
- Use short expiration

### 4. Algorithm Confusion
- Validate algorithm in header
- Prevent algorithm switching attacks
- Only allow specific algorithms

---

## Ti Router Implementation

**File**: `layers/authentication/jwt.go`

**Features**:
- JWT struct với Header, Payload, Signature
- JWTManager với secret key và issuer
- GenerateToken() method với custom claims
- ValidateToken() method với signature, expiration, issuer validation
- RefreshToken() method
- HMAC-SHA256 signing
- Base64 URL encoding

**Usage**:
```go
manager := NewJWTManager(secret_key, "ti-router")
token, err := manager.GenerateToken(user_id, 1*time.Hour, custom_claims)
payload, err := manager.ValidateToken(token)
new_token, err := manager.RefreshToken(token, 2*time.Hour)
```

---

## References

- [JWT.io](https://jwt.io/)
- [JWT Best Practices](https://tools.ietf.org/html/rfc8725)
- [OWASP JWT Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/JSON_Web_Token_for_Java_Cheat_Sheet.html)

---

## Next Steps

1. Implement refresh tokens
2. Add token blacklist với Redis
3. Implement key rotation
4. Add JWKS support
5. Integrate với OAuth2 providers
