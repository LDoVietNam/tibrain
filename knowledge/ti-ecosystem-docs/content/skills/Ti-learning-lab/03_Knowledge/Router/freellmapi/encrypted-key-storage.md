# Encrypted Key Storage (AES-256-GCM)

> **Feature**: Lưu trữ API keys được mã hóa bằng AES-256-GCM
> **Nguồn cảm hứng**: FreeLLMAPI - `server/src/lib/crypto.ts`
> **Trạng thái**: ✅ Hoàn thành

## Tổng Quan

Ti Router hiện hỗ trợ mã hóa API keys và sensitive data sử dụng AES-256-GCM, tương tự như FreeLLMAPI. Điều này đảm bảo rằng keys được lưu trữ an toàn tại rest (at-rest encryption).

## Kiến Trúc

### File triển khai
- `Z:\Ti\router\layers\authentication\encryption.go` - Core encryption logic
- `Z:\Ti\router\layers\authentication\encryption_test.go` - Unit tests

### Components

#### 1. InitEncryptionKey
Khởi tạo encryption key từ 3 nguồn theo thứ tự:

```go
func InitEncryptionKey(dbConn db.Database) error
```

**Thứ tự ưu tiên**:
1. **Environment variable**: `ENCRYPTION_KEY` (64-char hex string)
2. **Database**: Lưu trong `key_value` table với namespace="system", key="encryption_key"
3. **Auto-generate**: Tạo random key nếu không tìm thấy ở đâu

**Lưu ý**: Key được cache in-memory để tránh truy cập DB liên tục.

#### 2. Encrypt
Mã hóa plaintext sử dụng AES-256-GCM:

```go
func Encrypt(plaintext string) (*EncryptResult, error)
```

**Quy trình**:
1. Lấy encryption key từ cache
2. Tạo AES cipher block
3. Tạo GCM mode (Galois/Counter Mode)
4. Generate random nonce (12 bytes cho GCM)
5. Seal (encrypt) - GCM tự động append auth tag
6. Return base64-encoded ciphertext + nonce

**EncryptResult structure**:
```go
type EncryptResult struct {
    Encrypted string `json:"encrypted"` // base64 ciphertext (bao gồm nonce)
    IV        string `json:"iv"`        // base64 nonce (để tương thích)
}
```

#### 3. Decrypt
Giải mã ciphertext:

```go
func Decrypt(encryptedB64, ivB64 string) (string, error)
```

**Quy trình**:
1. Lấy encryption key từ cache
2. Decode base64 ciphertext
3. Tạo AES cipher block + GCM mode
4. Extract nonce từ đầu ciphertext (12 bytes)
5. Open (decrypt) - GCM tự động verify auth tag
6. Return plaintext

**Lưu ý**: IV parameter hiện tại không được sử dụng vì nonce đã được embed trong ciphertext. Được giữ lại để tương thích với API cũ.

#### 4. MaskKey
Mask API key để hiển thị an toàn:

```go
func MaskKey(key string) string
```

**Logic**:
- Nếu key <= 8 chars: return `"****"`
- Nếu không: return `key[:4] + "..." + key[len(key)-4:]`

**Ví dụ**:
- `"sk-test-api-key-1234567890"` → `"sk-t...7890"`
- `"short"` → `"****"`

#### 5. CryptoManager
Wrapper struct cho dễ sử dụng:

```go
type CryptoManager struct{}

func NewCryptoManager() *CryptoManager
func (c *CryptoManager) Encrypt(plaintext string) (*EncryptResult, error)
func (c *CryptoManager) Decrypt(encrypted *EncryptResult) (string, error)
func (c *CryptoManager) MaskKey(key string) string
```

## So Sánh với FreeLLMAPI

| Feature | FreeLLMAPI (TS) | Ti Router (Go) |
|---------|-----------------|----------------|
| Algorithm | AES-256-GCM | AES-256-GCM ✅ |
| Key sources | env, DB, auto-generate | env, DB, auto-generate ✅ |
| Nonce size | 12 bytes | 12 bytes (GCM default) ✅ |
| Auth tag handling | Manual (extract/combine) | Automatic (GCM built-in) ✅ |
| Base64 encoding | Yes | Yes ✅ |
| Mask function | Yes | Yes ✅ |

**Khác biệt chính**:
- FreeLLMAPI manually extract và combine auth tag
- Ti Router sử dụng GCM's built-in auth tag (tự động append/verify)
- Cả hai đều an toàn, nhưng Ti Router's approach là idiomatic cho Go

## Usage Example

```go
// Initialize encryption key (thường làm ở startup)
err := authentication.InitEncryptionKey(db.Get())
if err != nil {
    log.Fatal(err)
}

// Create crypto manager
crypto := authentication.NewCryptoManager()

// Encrypt API key
encrypted, err := crypto.Encrypt("sk-actual-api-key-here")
if err != nil {
    log.Fatal(err)
}

// Save to DB (encrypted.Encrypted, encrypted.IV)
// ...

// Later: decrypt
decrypted, err := crypto.Decrypt(encrypted)
if err != nil {
    log.Fatal(err)
}
// decrypted == "sk-actual-api-key-here"

// Mask for display
masked := crypto.MaskKey("sk-actual-api-key-here")
// masked == "sk-a...ere"
```

## Testing

Tất cả tests pass:

```bash
cd Z:\Ti\router
go test ./layers/authentication/ -v
```

**Test coverage**:
- `TestEncryptDecrypt` - Mã hóa và giải mã thành công
- `TestMaskKey` - Mask function hoạt động đúng
- `TestEncryptDecryptWithDifferentManagers` - Key cache hoạt động đúng

## Security Considerations

### ✅ Best Practices
- AES-256-GCM là authenticated encryption (confidentiality + integrity)
- Random nonce cho mỗi encryption (không reuse)
- Key được cache in-memory (không log ra file/stdout)
- Auth tag tự động verify trong GCM mode

### ⚠️ Lưu ý
- **ENCRYPTION_KEY env var**: Nếu set, phải là 64-char hex string
- **DB persistence**: Key được lưu trong `key_value` table (không encrypted)
- **Key rotation**: Hiện tại không support rotation (future enhancement)

### 🔐 Recommendations cho Production
1. Set `ENCRYPTION_KEY` env var (không dùng auto-generate)
2. Backup DB thường xuyên (nếu mất key = mất tất cả encrypted data)
3. Implement key rotation mechanism
4. Log encryption/decryption failures (cho security monitoring)

## Integration với Existing Code

### Cần update để sử dụng encryption:

1. **APIKeyManager** (`keys.go`):
   - Thay vì lưu `Key` plaintext, lưu `EncryptedKey` (EncryptResult)
   - Update `GenerateKey`, `GetKey`, `ValidateKey`

2. **CredentialStore** (`credential_store.go`):
   - Encrypt `AccessToken`, `RefreshToken` trước khi lưu
   - Decrypt khi cần sử dụng

3. **OAuth handlers** (`oauth_handlers.go`, `tokens_handlers.go`):
   - Mask keys trong API responses
   - Decrypt tokens trước khi gửi đến providers

## Future Enhancements

1. **Key Rotation**: Hỗ trợ rotate encryption keys
2. **Key Versioning**: Track key versions trong DB
3. **Key Derivation**: Sử dụng KDF (Key Derivation Function) từ master key
4. **Hardware Security Module (HSM)**: Support HSM cho production
5. **Audit Logging**: Log tất cả encryption/decryption operations

## References

- FreeLLMAPI: `Z:\Ti\Ti-learning-lab\05_Repositories\router\freellmapi-main\server\src\lib\crypto.ts`
- Go crypto package: https://pkg.go.dev/crypto/aes
- GCM specification: NIST SP 800-38D

---

**Ngày tạo**: 2026-04-28
**Agent**: Claude Code
**Project**: Ti Router
