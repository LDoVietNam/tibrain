# AutoReg + TempMail OTP Integration Guide

## Overview
This guide explains how to integrate the TempMail + OTP automation system with the existing AutoReg system in `apps/cli/internal/autoreg/autoreg.go`.

## Current AutoReg System Analysis

### Key Components
- **AutoregManager**: Manages account creation and pool management
- **PlatformConfig**: Configuration for each platform (requires_email, requires_phone, etc.)
- **Account**: Account structure with email, password, cookie, status
- **RegistrarFactory**: Platform-specific registration implementations

### Current Email Handling
- Uses real email domains (outlook.com, hotmail.com, live.com)
- Generates random email addresses
- No OTP verification mechanism
- Phone verification is optional but not implemented

## Integration Requirements

### 1. Add TempMail Email Generation
Replace real email generation with tempmail email generation.

**Location**: `apps/cli/internal/autoreg/autoreg.go`

**Modification**: Add TempMail client to generate temporary emails.

```go
// Add to imports
import (
    "github.com/ti/cli/internal/tempmail"
)

// Modify generateRandomEmail method
func (m *AutoregManager) generateRandomEmail() (string, error) {
    // Use TempMail client instead of real domains
    tmClient := tempmail.NewClient()
    email, err := tmClient.GenerateEmail()
    if err != nil {
        return "", err
    }
    return email, nil
}
```

### 2. Add OTP Verification to Account Structure
Add OTP-related fields to the Account struct.

**Location**: `apps/cli/internal/autoreg/autoreg.go` (line 49)

**Modification**:
```go
type Account struct {
    ID          string            `json:"id"`
    Platform    string            `json:"platform"`
    Email       string            `json:"email"`
    Password    string            `json:"password"`
    Cookie      string            `json:"cookie"`
    CookieJSON  map[string]string `json:"cookie_json"`
    Status      AccountStatus     `json:"status"`
    CreatedAt   time.Time         `json:"created_at"`
    LastUsedAt  time.Time         `json:"last_used_at"`
    UsageCount  int               `json:"usage_count"`
    SuccessRate float64           `json:"success_rate"`
    Quota       int               `json:"quota"`
    QuotaUsed   int               `json:"quota_used"`
    Metadata    map[string]string `json:"metadata"`
    
    // New fields for OTP
    OTPCode         string    `json:"otp_code,omitempty"`
    OTPReceivedAt   time.Time `json:"otp_received_at,omitempty"`
    OTPVerified     bool      `json:"otp_verified,omitempty"`
    OTPProvider     string    `json:"otp_provider,omitempty"`
}
```

### 3. Add OTP Verification to Registration Flow
Add OTP verification step after registration form submission.

**Location**: `apps/cli/internal/autoreg/autoreg.go` (line 364 - CreateAccount method)

**Modification**:
```go
func (m *AutoregManager) CreateAccount(platform string) (*Account, error) {
    // ... existing code ...
    
    // Generate tempmail email
    tmClient := tempmail.NewClient()
    email, err := tmClient.GenerateEmail()
    if err != nil {
        return nil, fmt.Errorf("failed to generate tempmail: %w", err)
    }
    
    account.Email = email
    account.OTPProvider = tmClient.GetProvider()
    
    // Execute registration via browser automation
    factory := NewRegistrarFactory()
    registrar, err := factory.CreateRegistrar(platform)
    if err == nil {
        var result *RegistrationResult
        result, err = registrar.Register(email, password)
        
        if err == nil && result.Success {
            // Check if OTP verification is needed
            if result.NeedsOTPVerify {
                // Wait for OTP email
                otp, err := tmClient.WaitForOTP(email, 5*time.Minute)
                if err != nil {
                    account.Status = AccountStatusPending
                    account.Metadata["otp_error"] = err.Error()
                } else {
                    account.OTPCode = otp
                    account.OTPReceivedAt = time.Now()
                    
                    // Submit OTP to complete registration
                    verifyErr := registrar.VerifyOTP(otp)
                    if verifyErr == nil {
                        account.OTPVerified = true
                        account.Status = AccountStatusActive
                    } else {
                        account.Status = AccountStatusPending
                        account.Metadata["otp_verify_error"] = verifyErr.Error()
                    }
                }
            } else {
                account.Status = AccountStatusActive
            }
            
            // Save cookie
            account.Cookie = result.CookieHeader
        }
    }
    
    // ... rest of existing code ...
}
```

### 4. Add OTP Verification to Login Flow
Add OTP verification step after login.

**Location**: `apps/cli/internal/autoreg/autoreg.go`

**Add new method**:
```go
func (m *AutoregManager) LoginWithOTP(platform, email, password string) (*Account, error) {
    platformConfig, exists := m.config.Platforms[platform]
    if !exists {
        return nil, fmt.Errorf("platform %s not configured", platform)
    }
    
    // Find account
    m.mu.RLock()
    var account *Account
    for _, acc := range m.accounts[platform] {
        if acc.Email == email {
            account = acc
            break
        }
    }
    m.mu.RUnlock()
    
    if account == nil {
        return nil, fmt.Errorf("account not found")
    }
    
    // Execute login via browser automation
    factory := NewRegistrarFactory()
    registrar, err := factory.CreateRegistrar(platform)
    if err != nil {
        return nil, err
    }
    
    // Submit login
    result, err := registrar.Login(email, password)
    if err != nil {
        return nil, err
    }
    
    // Check if OTP is needed
    if result.NeedsOTPVerify {
        // Wait for OTP email
        tmClient := tempmail.NewClient()
        otp, err := tmClient.WaitForOTP(email, 5*time.Minute)
        if err != nil {
            return nil, fmt.Errorf("failed to receive OTP: %w", err)
        }
        
        // Submit OTP
        verifyErr := registrar.VerifyOTP(otp)
        if verifyErr != nil {
            return nil, fmt.Errorf("failed to verify OTP: %w", verifyErr)
        }
        
        account.OTPCode = otp
        account.OTPReceivedAt = time.Now()
        account.OTPVerified = true
    }
    
    // Update cookie
    account.Cookie = result.CookieHeader
    account.LastUsedAt = time.Now()
    account.Status = AccountStatusActive
    
    // Save
    m.saveToStorage()
    
    return account, nil
}
```

### 5. Create TempMail Client
Create a new tempmail client package.

**Location**: `apps/cli/internal/tempmail/client.go`

```go
package tempmail

import (
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "time"
)

type TempMailClient struct {
    provider string
    client   *http.Client
}

type Email struct {
    Address   string
    Provider  string
    CreatedAt time.Time
}

func NewClient() *TempMailClient {
    return &TempMailClient{
        provider: "guerrillamail", // default
        client: &http.Client{
            Timeout: 30 * time.Second,
        },
    }
}

func (c *TempMailClient) SetProvider(provider string) {
    c.provider = provider
}

func (c *TempMailClient) GenerateEmail() (string, error) {
    switch c.provider {
    case "guerrillamail":
        return c.generateGuerrillaMail()
    case "tempmail":
        return c.generateTempMail()
    default:
        return c.generateGuerrillaMail()
    }
}

func (c *TempMailClient) generateGuerrillaMail() (string, error) {
    resp, err := c.client.Get("https://api.guerrillamail.com/cgi-bin/ajax_email_address_generate.php")
    if err != nil {
        return "", err
    }
    defer resp.Body.Close()
    
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return "", err
    }
    
    var result map[string]interface{}
    if err := json.Unmarshal(body, &result); err != nil {
        return "", err
    }
    
    email, ok := result["email_addr"].(string)
    if !ok {
        return "", fmt.Errorf("failed to extract email from response")
    }
    
    return email, nil
}

func (c *TempMailClient) generateTempMail() (string, error) {
    // Implement temp-mail.org API
    return "", fmt.Errorf("not implemented")
}

func (c *TempMailClient) WaitForOTP(email string, timeout time.Duration) (string, error) {
    deadline := time.Now().Add(timeout)
    
    for time.Now().Before(deadline) {
        // Check inbox for OTP
        otp, err := c.checkForOTP(email)
        if err == nil && otp != "" {
            return otp, nil
        }
        
        // Wait before next check
        time.Sleep(5 * time.Second)
    }
    
    return "", fmt.Errorf("timeout waiting for OTP")
}

func (c *TempMailClient) checkForOTP(email string) (string, error) {
    // Implement OTP extraction from email
    return "", fmt.Errorf("not implemented")
}

func (c *TempMailClient) GetProvider() string {
    return c.provider
}
```

### 6. Update PlatformConfig
Remove phone requirement and add OTP requirement.

**Location**: `apps/cli/internal/autoreg/autoreg.go` (line 67)

**Modification**:
```go
type PlatformConfig struct {
    Name            string `json:"name"`
    BaseURL         string `json:"base_url"`
    SignupURL       string `json:"signup_url"`
    LoginURL        string `json:"login_url"`
    RequiresEmail   bool   `json:"requires_email"`
    RequiresOTP     bool   `json:"requires_otp"`   // NEW: Requires OTP verification
    RequiresPhone   bool   `json:"requires_phone"`   // DEPRECATED: No longer needed
    RequiresCaptcha bool   `json:"requires_captcha"`
    HasFreeTier     bool   `json:"has_free_tier"`
    DailyQuota      int    `json:"daily_quota"`
    MaxAccounts     int    `json:"max_accounts"`
    Enabled         bool   `json:"enabled"`
}
```

### 7. Update Default Platform Configs
Update platform configs to use OTP instead of phone.

**Location**: `apps/cli/internal/autoreg/autoreg.go` (line 148)

**Modification**:
```go
func DefaultPlatformConfigs() map[string]*PlatformConfig {
    return map[string]*PlatformConfig{
        "sharedchat": {
            Name:            "sharedchat",
            BaseURL:         "https://chat.sharedchat.fun",
            SignupURL:       "https://chat.sharedchat.fun/signup",
            LoginURL:        "https://chat.sharedchat.fun/login",
            RequiresEmail:   true,
            RequiresOTP:     true,   // Changed from RequiresPhone
            RequiresPhone:   false,  // Disabled
            RequiresCaptcha: false,
            HasFreeTier:     true,
            DailyQuota:      50,
            MaxAccounts:     10,
            Enabled:         true,
        },
        "poe-free": {
            Name:            "poe-free",
            BaseURL:         "https://poe.com",
            SignupURL:       "https://poe.com/signup",
            LoginURL:        "https://poe.com/login",
            RequiresEmail:   true,
            RequiresOTP:     true,   // Changed from RequiresPhone
            RequiresPhone:   false,  // Disabled
            RequiresCaptcha: true,
            HasFreeTier:     true,
            DailyQuota:      30,
            MaxAccounts:     5,
            Enabled:         true,  // Can now enable
        },
    }
}
```

### 8. Update Registrar Interface
Add OTP verification method to registrar interface.

**Location**: `apps/cli/internal/autoreg/registrar.go` (create new file)

```go
package autoreg

import "time"

// RegistrationResult - Result of registration attempt
type RegistrationResult struct {
    Success        bool      `json:"success"`
    CookieHeader   string    `json:"cookie_header"`
    Session        *Session  `json:"session,omitempty"`
    NeedsEmailVerify bool    `json:"needs_email_verify"`
    NeedsOTPVerify   bool    `json:"needs_otp_verify"`
    NeedsCaptcha     bool    `json:"needs_captcha"`
    Error          string    `json:"error,omitempty"`
}

// Session - Browser session
type Session struct {
    Cookies []Cookie `json:"cookies"`
}

// Cookie - HTTP cookie
type Cookie struct {
    Name   string `json:"name"`
    Value  string `json:"value"`
    Domain string `json:"domain"`
}

// Registrar - Interface for platform-specific registration
type Registrar interface {
    Register(email, password string) (*RegistrationResult, error)
    Login(email, password string) (*RegistrationResult, error)
    VerifyOTP(otp string) error
}

// RegistrarFactory - Factory for creating registrars
type RegistrarFactory struct{}

func NewRegistrarFactory() *RegistrarFactory {
    return &RegistrarFactory{}
}

func (f *RegistrarFactory) CreateRegistrar(platform string) (Registrar, error) {
    switch platform {
    case "sharedchat":
        return NewSharedChatRegistrar(), nil
    case "poe-free":
        return NewPoeRegistrar(), nil
    default:
        return nil, fmt.Errorf("no registrar for platform: %s", platform)
    }
}
```

## Configuration Updates

### Update AutoReg Config
Add tempmail configuration to autoreg config.

**Location**: `configs/base/autoreg.yaml`

```yaml
autoreg:
  storage_path: ~/.ti/autoreg
  max_accounts_per_platform: 10
  auto_create: true
  health_check_interval: 30m
  
  # TempMail configuration
  tempmail:
    enabled: true
    provider: guerrillamail
    providers:
      guerrillamail:
        api_url: https://api.guerrillamail.com
        generate_endpoint: /cgi-bin/ajax_email_address_generate.php
        check_endpoint: /cgi-bin/ajax_email_check.php
      tempmail:
        api_url: https://api.temp-mail.org
        generate_endpoint: /request/mail/id/format/json/
        check_endpoint: /request/mail/id/format/json/
    
    # OTP extraction
    otp:
      polling_interval: 5s
      timeout: 5m
      patterns:
        - "\\b\\d{6}\\b"
        - "\\b\\d{4}\\b"
        - "(?:code|otp|pin|verify)[:\\s]*(\\d{4,8})"
```

## Testing

### Test Registration with TempMail
```bash
# Test registration with tempmail
ti-cli autoreg register --platform sharedchat

# Should:
# 1. Generate tempmail email
# 2. Fill registration form
# 3. Wait for OTP email
# 4. Extract OTP
# 5. Submit OTP
# 6. Complete registration
# 7. Save account with OTP verified status
```

### Test Login with OTP
```bash
# Test login with OTP
ti-cli autoreg login --platform sharedchat --email <tempmail>

# Should:
# 1. Submit login
# 2. Wait for OTP email
# 3. Extract OTP
# 4. Submit OTP
# 5. Complete login
# 6. Update cookie
```

## Migration Steps

1. **Backup existing autoreg data**
   ```bash
   cp -r ~/.ti/autoreg ~/.ti/autoreg.backup
   ```

2. **Update code**
   - Add tempmail client package
   - Modify Account struct
   - Update CreateAccount method
   - Add LoginWithOTP method
   - Update PlatformConfig
   - Update default platform configs

3. **Build and test**
   ```bash
   cd apps/cli
   go build
   ti-cli autoreg register --platform sharedchat
   ```

4. **Verify**
   - Check that tempmail email is generated
   - Check that OTP is received and extracted
   - Check that account is created with OTP verified status
   - Check that login works with OTP

## Rollback Plan

If integration fails:
1. Restore backup: `cp -r ~/.ti/autoreg.backup ~/.ti/autoreg`
2. Revert code changes
3. Disable tempmail in config
4. Continue using real email domains

## Benefits

1. **No Phone Required**: Eliminates need for phone verification
2. **Automated OTP**: Automatic OTP extraction and submission
3. **Scalable**: Can create multiple accounts without real email/phone
4. **Cost Effective**: No need for real email/phone services
5. **Privacy**: Temporary emails are disposable

## Limitations

1. **TempMail Reliability**: Dependent on tempmail provider uptime
2. **OTP Timeout**: Some services have short OTP expiry
3. **Rate Limiting**: Tempmail providers may rate limit
4. **Detection Risk**: Services may detect tempmail usage
5. **Account Security**: Tempmail accounts are less secure
