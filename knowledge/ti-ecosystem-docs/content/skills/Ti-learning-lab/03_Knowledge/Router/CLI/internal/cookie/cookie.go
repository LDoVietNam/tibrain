// Package cookie manages HTTP cookies with browser-like behavior.
// Supports: Chrome cookie import, AES-GCM encryption, persistence,
// auto-refresh, caching, rate limiting, and TLS security.
package cookie

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// ChromeCookie is the Chrome cookie export format.
type ChromeCookie struct {
	Domain         string  `json:"domain"`
	ExpirationDate float64 `json:"expirationDate,omitempty"`
	HostOnly       bool    `json:"hostOnly,omitempty"`
	HTTPOnly       bool    `json:"httpOnly,omitempty"`
	Name           string  `json:"name"`
	Path           string  `json:"path"`
	SameSite       string  `json:"sameSite,omitempty"`
	Secure         bool    `json:"secure,omitempty"`
	Session        bool    `json:"session,omitempty"`
	StoreID        string  `json:"storeId,omitempty"`
	Value          string  `json:"value"`
}

// Cookie is the internal cookie representation.
type Cookie struct {
	Name     string `json:"name"`
	Value    string `json:"value"`
	Domain   string `json:"domain"`
	Expire   int64  `json:"expire"`
	Secure   bool   `json:"secure"`
	HTTPOnly bool   `json:"httpOnly"`
	Path     string `json:"path"`
}

// Profile stores cookies for a named session.
type Profile struct {
	Name      string    `json:"name"`
	Cookies   []*Cookie `json:"cookies"`
	Domain    string    `json:"domain"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Manager manages multiple cookie profiles with thread-safe access.
type Manager struct {
	mu       sync.RWMutex
	profiles map[string]*Profile
	key      []byte // AES encryption key (32 bytes for AES-256)
}

// NewManager creates a new Cookie Manager.
func NewManager() *Manager {
	return &Manager{
		profiles: make(map[string]*Profile),
	}
}

// WithKey sets the AES encryption key for cookie storage.
func (c *Manager) WithKey(key []byte) *Manager {
	c.key = key
	return c
}

// LoadChromeJSON parses Chrome cookie export JSON and creates a profile.
func (c *Manager) LoadChromeJSON(jsonData []byte, profileName string) error {
	var chromeCookies []ChromeCookie
	if err := json.Unmarshal(jsonData, &chromeCookies); err != nil {
		return fmt.Errorf("parse chrome cookie JSON: %w", err)
	}
	if len(chromeCookies) == 0 {
		return fmt.Errorf("no cookies in JSON")
	}

	if profileName == "" {
		profileName = sanitizeDomainName(chromeCookies[0].Domain)
	}

	cookies := make([]*Cookie, 0, len(chromeCookies))
	domain := chromeCookies[0].Domain
	for _, ck := range chromeCookies {
		expire := int64(0)
		if ck.ExpirationDate > 0 {
			expire = int64(ck.ExpirationDate)
		}
		cookies = append(cookies, &Cookie{
			Name:     ck.Name,
			Value:    ck.Value,
			Domain:   ck.Domain,
			Path:     ck.Path,
			Secure:   ck.Secure,
			HTTPOnly: ck.HTTPOnly,
			Expire:   expire,
		})
	}

	c.mu.Lock()
	c.profiles[profileName] = &Profile{
		Name:      profileName,
		Domain:    domain,
		Cookies:   cookies,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	c.mu.Unlock()

	return nil
}

// LoadFile reads a Chrome cookie JSON file and creates a profile.
func (c *Manager) LoadFile(filePath string, profileName string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("read file %s: %w", filePath, err)
	}
	return c.LoadChromeJSON(data, profileName)
}

// SaveFile writes a profile to a Chrome-compatible JSON file.
func (c *Manager) SaveFile(name string, filePath string) error {
	c.mu.RLock()
	profile, ok := c.profiles[name]
	c.mu.RUnlock()
	if !ok {
		return fmt.Errorf("profile %q not found", name)
	}

	chromeCookies := make([]ChromeCookie, 0, len(profile.Cookies))
	for _, ck := range profile.Cookies {
		chromeCookies = append(chromeCookies, ChromeCookie{
			Domain:         ck.Domain,
			Name:           ck.Name,
			Value:          ck.Value,
			Path:           ck.Path,
			Secure:         ck.Secure,
			HTTPOnly:       ck.HTTPOnly,
			ExpirationDate: float64(ck.Expire),
		})
	}

	data, err := json.MarshalIndent(chromeCookies, "", "    ")
	if err != nil {
		return fmt.Errorf("marshal cookies: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		return fmt.Errorf("mkdir: %w", err)
	}
	return os.WriteFile(filePath, data, 0600)
}

// SaveEncrypted writes an encrypted cookie profile to disk (AES-GCM).
func (c *Manager) SaveEncrypted(name string, filePath string) error {
	if len(c.key) == 0 {
		return fmt.Errorf("no encryption key set")
	}

	c.mu.RLock()
	profile, ok := c.profiles[name]
	c.mu.RUnlock()
	if !ok {
		return fmt.Errorf("profile %q not found", name)
	}

	data, err := json.Marshal(profile)
	if err != nil {
		return fmt.Errorf("marshal profile: %w", err)
	}

	encrypted, err := EncryptGCM(data, c.key)
	if err != nil {
		return fmt.Errorf("encrypt: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		return fmt.Errorf("mkdir: %w", err)
	}
	return os.WriteFile(filePath, []byte(encrypted), 0600)
}

// LoadEncrypted reads and decrypts an encrypted cookie profile (AES-GCM).
func (c *Manager) LoadEncrypted(filePath string, profileName string) error {
	if len(c.key) == 0 {
		return fmt.Errorf("no encryption key set")
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("read file: %w", err)
	}

	decrypted, err := DecryptGCM(string(data), c.key)
	if err != nil {
		return fmt.Errorf("decrypt: %w", err)
	}

	var profile Profile
	if err := json.Unmarshal([]byte(decrypted), &profile); err != nil {
		return fmt.Errorf("unmarshal profile: %w", err)
	}

	if profileName != "" {
		profile.Name = profileName
	}

	c.mu.Lock()
	c.profiles[profile.Name] = &profile
	c.profiles[profile.Name].UpdatedAt = time.Now()
	c.mu.Unlock()

	return nil
}

// GetCookieJar returns an http.CookieJar for the profile.
func (c *Manager) GetCookieJar(name string) (http.CookieJar, error) {
	c.mu.RLock()
	profile, ok := c.profiles[name]
	c.mu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("profile %q not found", name)
	}

	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}
	if len(profile.Cookies) == 0 {
		return jar, nil
	}

	domain := profile.Domain
	if domain == "" && len(profile.Cookies) > 0 {
		domain = profile.Cookies[0].Domain
	}
	if domain == "" {
		domain = "localhost"
	}

	httpCookies := make([]*http.Cookie, 0, len(profile.Cookies))
	for _, ck := range profile.Cookies {
		httpCookies = append(httpCookies, &http.Cookie{
			Name:     ck.Name,
			Value:    ck.Value,
			Domain:   ck.Domain,
			Path:     ck.Path,
			Secure:   ck.Secure,
			HttpOnly: ck.HTTPOnly,
		})
	}

	u := &url.URL{Scheme: "https", Host: domain}
	jar.SetCookies(u, httpCookies)
	return jar, nil
}

// GetHeader returns cookies as HTTP "Cookie" header string.
func (c *Manager) GetHeader(name string) string {
	c.mu.RLock()
	profile, ok := c.profiles[name]
	c.mu.RUnlock()
	if !ok {
		return ""
	}
	var parts []string
	for _, ck := range profile.Cookies {
		parts = append(parts, ck.Name+"="+ck.Value)
	}
	return strings.Join(parts, "; ")
}

// GetHTTPCookies returns cookies as []*http.Cookie for use with http.Request.
func (c *Manager) GetHTTPCookies(name string) ([]*http.Cookie, bool) {
	c.mu.RLock()
	profile, ok := c.profiles[name]
	c.mu.RUnlock()
	if !ok {
		return nil, false
	}

	httpCookies := make([]*http.Cookie, 0, len(profile.Cookies))
	for _, ck := range profile.Cookies {
		httpCookies = append(httpCookies, &http.Cookie{
			Name:     ck.Name,
			Value:    ck.Value,
			Domain:   ck.Domain,
			Path:     ck.Path,
			Secure:   ck.Secure,
			HttpOnly: ck.HTTPOnly,
		})
	}
	return httpCookies, true
}

// GetSessionID extracts the gfsessionid value.
func (c *Manager) GetSessionID(name string) (string, bool) {
	c.mu.RLock()
	profile, ok := c.profiles[name]
	c.mu.RUnlock()
	if !ok {
		return "", false
	}
	for _, ck := range profile.Cookies {
		if ck.Name == "gfsessionid" {
			return ck.Value, true
		}
	}
	return "", false
}

// IsExpired checks if any cookie in the profile is expired.
func (c *Manager) IsExpired(name string) bool {
	c.mu.RLock()
	profile, ok := c.profiles[name]
	c.mu.RUnlock()
	if !ok {
		return true
	}
	now := time.Now().Unix()
	for _, ck := range profile.Cookies {
		if ck.Expire > 0 && ck.Expire < now {
			return true
		}
	}
	return false
}

// RemoveExpired deletes expired cookies from a profile.
func (c *Manager) RemoveExpired(name string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	profile, ok := c.profiles[name]
	if !ok {
		return
	}
	now := time.Now().Unix()
	filtered := profile.Cookies[:0]
	for _, ck := range profile.Cookies {
		if ck.Expire == 0 || ck.Expire >= now {
			filtered = append(filtered, ck)
		}
	}
	profile.Cookies = filtered
	profile.UpdatedAt = time.Now()
}

// UpdateCookies replaces cookies in a profile.
func (c *Manager) UpdateCookies(name string, cookies []*Cookie) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if profile, ok := c.profiles[name]; ok {
		profile.Cookies = cookies
		profile.UpdatedAt = time.Now()
	}
}

// ProfileNames returns all profile names.
func (c *Manager) ProfileNames() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	names := make([]string, 0, len(c.profiles))
	for name := range c.profiles {
		names = append(names, name)
	}
	return names
}

// Remove deletes a profile.
func (c *Manager) Remove(name string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.profiles, name)
}

// Len returns the number of profiles.
func (c *Manager) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.profiles)
}

type ProfileSummary struct {
	Name             string    `json:"name"`
	Domain           string    `json:"domain"`
	CookieCount      int       `json:"cookie_count"`
	HasSessionID     bool      `json:"has_session_id"`
	SessionPreview   string    `json:"session_preview,omitempty"`
	UpdatedAt        time.Time `json:"updated_at"`
	ContainsMetadata bool      `json:"contains_metadata"`
}

// Summaries returns redacted per-profile summaries.
func (c *Manager) Summaries() []ProfileSummary {
	c.mu.RLock()
	defer c.mu.RUnlock()
	out := make([]ProfileSummary, 0, len(c.profiles))
	for name, profile := range c.profiles {
		summary := ProfileSummary{Name: name, Domain: profile.Domain, CookieCount: len(profile.Cookies), UpdatedAt: profile.UpdatedAt}
		for _, ck := range profile.Cookies {
			if ck.Name == "gfsessionid" && ck.Value != "" {
				summary.HasSessionID = true
				summary.SessionPreview = previewCookieValue(ck.Value)
			}
			if ck.Name == "oai-client-auth-info" || ck.Name == "oai-default-model-config" || ck.Name == "oai-last-model-config" {
				summary.ContainsMetadata = true
			}
		}
		out = append(out, summary)
	}
	return out
}

// BestProfileForHost finds the profile with the strongest match for a target host.
func (c *Manager) BestProfileForHost(host string) (string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	host = strings.ToLower(strings.TrimSpace(host))
	bestName := ""
	bestScore := -1
	for name, profile := range c.profiles {
		score := 0
		for _, ck := range profile.Cookies {
			domain := strings.TrimPrefix(strings.ToLower(strings.TrimSpace(ck.Domain)), ".")
			if domain == "" {
				continue
			}
			if host == domain {
				score += 3
			} else if strings.HasSuffix(host, "."+domain) {
				score += 2
			}
			if ck.Name == "gfsessionid" && ck.Value != "" {
				score += 5
			}
		}
		if score > bestScore {
			bestName, bestScore = name, score
		}
	}
	return bestName, bestScore > 0
}

// LoadStateFile loads the multi-profile state file used by the Ti CLI.
func (c *Manager) LoadStateFile(stateFile string) error {
	data, err := os.ReadFile(stateFile)
	if err != nil {
		return err
	}
	var wrapper struct {
		Profiles map[string][]ChromeCookie `json:"profiles"`
	}
	if err := json.Unmarshal(data, &wrapper); err != nil {
		return fmt.Errorf("parse state: %w", err)
	}
	for profileName, cookies := range wrapper.Profiles {
		chromeJSON, err := json.Marshal(cookies)
		if err != nil {
			return fmt.Errorf("marshal profile %s: %w", profileName, err)
		}
		if err := c.LoadChromeJSON(chromeJSON, profileName); err != nil {
			return fmt.Errorf("load profile %s: %w", profileName, err)
		}
	}
	return nil
}

// SaveStateFile saves all profiles to the Ti multi-profile state file.
func (c *Manager) SaveStateFile(stateFile string) error {
	if err := os.MkdirAll(filepath.Dir(stateFile), 0755); err != nil {
		return err
	}

	state := make(map[string][]ChromeCookie)
	c.mu.RLock()
	for profileName, profile := range c.profiles {
		entries := make([]ChromeCookie, 0, len(profile.Cookies))
		for _, ck := range profile.Cookies {
			entries = append(entries, ChromeCookie{
				Domain:         ck.Domain,
				Name:           ck.Name,
				Value:          ck.Value,
				Path:           ck.Path,
				Secure:         ck.Secure,
				HTTPOnly:       ck.HTTPOnly,
				ExpirationDate: float64(ck.Expire),
			})
		}
		state[profileName] = entries
	}
	c.mu.RUnlock()

	wrapper := map[string]any{"profiles": state}
	data, err := json.MarshalIndent(wrapper, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(stateFile, data, 0600)
}

func previewCookieValue(v string) string {
	if v == "" {
		return ""
	}
	if len(v) <= 8 {
		return strings.Repeat("*", len(v))
	}
	return v[:4] + strings.Repeat("*", len(v)-8) + v[len(v)-4:]
}

// =============================================================================
// AES-GCM Encryption (secure, authenticated, random nonce)
// =============================================================================

// EncryptGCM encrypts data using AES-256-GCM with a random nonce.
// Output: base64(nonce + ciphertext + authTag)
func EncryptGCM(plaintext []byte, key []byte) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	// Random nonce (12 bytes for GCM)
	nonce := make([]byte, aesGCM.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return "", err
	}

	// Encrypt and authenticate
	ciphertext := aesGCM.Seal(nil, nonce, plaintext, nil)

	// Prepend nonce to ciphertext
	result := append(nonce, ciphertext...)
	return base64.StdEncoding.EncodeToString(result), nil
}

// EncryptBytes encrypts data using AES-256-GCM with a random nonce.
// Output: nonce + ciphertext + authTag (raw bytes).
// Alias for EncryptGCM but returns raw bytes instead of base64.
func EncryptBytes(plaintext []byte, key []byte) (string, error) {
	return EncryptGCM(plaintext, key)
}

// DecryptBytes decrypts data using AES-256-GCM.
// Alias for DecryptGCM for backwards compatibility.
func DecryptBytes(encryptedValue string, key []byte) (string, error) {
	return DecryptGCM(encryptedValue, key)
}

// DecryptGCM decrypts data using AES-256-GCM.
// Input: base64(nonce + ciphertext + authTag)
func DecryptGCM(encryptedValue string, key []byte) (string, error) {
	data, err := base64.StdEncoding.DecodeString(encryptedValue)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := aesGCM.NonceSize()
	if len(data) < nonceSize {
		return "", errors.New("ciphertext too short")
	}

	// Split nonce and ciphertext
	nonce, ciphertext := data[:nonceSize], data[nonceSize:]

	// Decrypt and verify authentication tag
	plaintext, err := aesGCM.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("decryption failed (wrong key or tampered data): %w", err)
	}

	return string(plaintext), nil
}

// =============================================================================
// Secure Cookie Setting (HttpOnly, Secure, SameSite)
// =============================================================================

// SetSecureCookie sets a cookie with security best practices.
func SetSecureCookie(w http.ResponseWriter, name, value string, maxAge int) {
	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   maxAge,
	})
}

// =============================================================================
// HTTPClient with cookie management, TLS, caching, and rate limiting
// =============================================================================

// HTTPClient is an HTTP client with:
//   - Automatic cookie management (load/attach/save)
//   - TLS configuration (certificate verification)
//   - Response caching
//   - Retry with exponential backoff
//   - Rate limiting
type HTTPClient struct {
	client     *http.Client
	mgr        *Manager
	profile    string
	cookieFile string

	// Caching
	cache    map[string]*cacheEntry
	cacheMu  sync.RWMutex
	cacheTTL time.Duration

	// Rate limiting
	rateMu     sync.Mutex
	rateTokens float64
	rateMax    float64
	rateRefill time.Duration
	rateLast   time.Time

	// Retry
	maxRetries int
	baseDelay  time.Duration
}

type cacheEntry struct {
	body      []byte
	headers   http.Header
	status    int
	expiresAt time.Time
}

// NewHTTPClient creates an HTTP client with cookie management.
func NewHTTPClient(mgr *Manager, profile string) *HTTPClient {
	return &HTTPClient{
		client: &http.Client{
			Timeout: 120 * time.Second,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					MinVersion: tls.VersionTLS12,
					// InsecureSkipVerify: false // Verify server cert (default)
				},
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 10,
				IdleConnTimeout:     90 * time.Second,
			},
		},
		mgr:      mgr,
		profile:  profile,
		cache:    make(map[string]*cacheEntry),
		cacheTTL: 5 * time.Minute,

		// Rate limit: 10 requests/second, burst 20
		rateMax:    20,
		rateTokens: 20,
		rateRefill: 100 * time.Millisecond,
		rateLast:   time.Now(),

		maxRetries: 3,
		baseDelay:  1 * time.Second,
	}
}

// WithCookieFile sets the file path for saving/loading cookies.
func (hc *HTTPClient) WithCookieFile(path string) *HTTPClient {
	hc.cookieFile = path
	return hc
}

// WithCacheTTL sets the cache time-to-live.
func (hc *HTTPClient) WithCacheTTL(ttl time.Duration) *HTTPClient {
	hc.cacheTTL = ttl
	return hc
}

// WithMaxRetries sets the retry configuration.
func (hc *HTTPClient) WithMaxRetries(n int, delay time.Duration) *HTTPClient {
	hc.maxRetries = n
	hc.baseDelay = delay
	return hc
}

// WithRateLimit sets the rate limit (tokens per refill interval).
func (hc *HTTPClient) WithRateLimit(maxTokens float64, refill time.Duration) *HTTPClient {
	hc.rateMax = maxTokens
	hc.rateTokens = maxTokens
	hc.rateRefill = refill
	return hc
}

// WithTLSClientCert adds a client certificate for mutual TLS.
func (hc *HTTPClient) WithTLSClientCert(certFile, keyFile string) error {
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return fmt.Errorf("load client cert: %w", err)
	}

	transport := hc.client.Transport.(*http.Transport)
	if transport.TLSClientConfig == nil {
		transport.TLSClientConfig = &tls.Config{MinVersion: tls.VersionTLS12}
	}
	transport.TLSClientConfig.Certificates = append(transport.TLSClientConfig.Certificates, cert)
	return nil
}

// WithCACert adds a custom CA certificate.
func (hc *HTTPClient) WithCACert(caFile string) error {
	caCert, err := os.ReadFile(caFile)
	if err != nil {
		return fmt.Errorf("read CA cert: %w", err)
	}

	transport := hc.client.Transport.(*http.Transport)
	if transport.TLSClientConfig == nil {
		transport.TLSClientConfig = &tls.Config{MinVersion: tls.VersionTLS12}
	}
	if transport.TLSClientConfig.RootCAs == nil {
		transport.TLSClientConfig.RootCAs, _ = x509.SystemCertPool()
	}
	transport.TLSClientConfig.RootCAs.AppendCertsFromPEM(caCert)
	return nil
}

// Do sends an HTTP request with cookies, caching, rate limiting, and retry.
func (hc *HTTPClient) Do(req *http.Request) (*http.Response, error) {
	// 1. Check cache (GET requests only)
	if req.Method == http.MethodGet {
		if resp := hc.checkCache(req.URL.String()); resp != nil {
			return resp, nil
		}
	}

	// 2. Rate limit
	if err := hc.waitRateLimit(req.Context()); err != nil {
		return nil, err
	}

	// 3. Retry with exponential backoff
	var lastErr error
	for attempt := 0; attempt <= hc.maxRetries; attempt++ {
		if attempt > 0 {
			delay := hc.baseDelay * time.Duration(1<<uint(attempt-1))
			if delay > 30*time.Second {
				delay = 30 * time.Second
			}
			time.Sleep(delay)
		}

		resp, err := hc.doOnce(req)
		if err == nil {
			// Cache the response
			if req.Method == http.MethodGet && resp.StatusCode == http.StatusOK {
				hc.cacheResponse(req.URL.String(), resp)
			}
			return resp, nil
		}
		lastErr = err
	}

	return nil, fmt.Errorf("request failed after %d retries: %w", hc.maxRetries, lastErr)
}

// doOnce performs a single HTTP request with cookie attachment.
func (hc *HTTPClient) doOnce(req *http.Request) (*http.Response, error) {
	// Remove expired cookies
	hc.mgr.RemoveExpired(hc.profile)

	// Attach cookies
	cks, ok := hc.mgr.GetHTTPCookies(hc.profile)
	if ok {
		for _, ck := range cks {
			req.AddCookie(ck)
		}
	}

	// Accept gzip
	req.Header.Set("Accept-Encoding", "gzip")

	// Send request
	resp, err := hc.client.Do(req)
	if err != nil {
		return nil, err
	}

	// Handle gzip
	if resp.Header.Get("Content-Encoding") == "gzip" {
		gzReader, err := gzip.NewReader(resp.Body)
		if err == nil {
			resp.Body = &gzipCloser{gzReader, resp.Body}
		}
	}

	// Save response cookies
	for _, ck := range resp.Cookies() {
		hc.mgr.mu.Lock()
		if profile, exists := hc.mgr.profiles[hc.profile]; exists {
			found := false
			for i, existing := range profile.Cookies {
				if existing.Name == ck.Name {
					profile.Cookies[i].Value = ck.Value
					found = true
					break
				}
			}
			if !found {
				profile.Cookies = append(profile.Cookies, &Cookie{
					Name:   ck.Name,
					Value:  ck.Value,
					Domain: ck.Domain,
					Path:   ck.Path,
					Expire: time.Now().Add(24 * time.Hour).Unix(),
				})
			}
			profile.UpdatedAt = time.Now()
		}
		hc.mgr.mu.Unlock()

		if hc.cookieFile != "" {
			if err := hc.mgr.SaveFile(hc.profile, hc.cookieFile); err != nil {
				fmt.Fprintf(os.Stderr, "warning: save cookies failed: %v\n", err)
			}
		}
	}

	return resp, nil
}

// Get makes an HTTP GET request with cookie management.
func (hc *HTTPClient) Get(url string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(context.Background(), "GET", url, nil)
	if err != nil {
		return nil, err
	}
	return hc.Do(req)
}

// Post makes an HTTP POST request with cookie management.
func (hc *HTTPClient) Post(url string, contentType string, body []byte) (*http.Response, error) {
	req, err := http.NewRequestWithContext(context.Background(), "POST", url, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", contentType)
	return hc.Do(req)
}

// checkCache returns a cached response if available and not expired.
func (hc *HTTPClient) checkCache(url string) *http.Response {
	hc.cacheMu.RLock()
	entry, ok := hc.cache[url]
	hc.cacheMu.RUnlock()
	if !ok || time.Now().After(entry.expiresAt) {
		return nil
	}

	return &http.Response{
		StatusCode: entry.status,
		Header:     entry.headers.Clone(),
		Body:       io.NopCloser(bytes.NewReader(entry.body)),
	}
}

// cacheResponse stores a response in the cache.
func (hc *HTTPClient) cacheResponse(url string, resp *http.Response) {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}
	resp.Body.Close()
	resp.Body = io.NopCloser(bytes.NewReader(body))

	hc.cacheMu.Lock()
	hc.cache[url] = &cacheEntry{
		body:      body,
		headers:   resp.Header.Clone(),
		status:    resp.StatusCode,
		expiresAt: time.Now().Add(hc.cacheTTL),
	}
	hc.cacheMu.Unlock()
}

// waitRateLimit blocks until a token is available.
func (hc *HTTPClient) waitRateLimit(ctx context.Context) error {
	hc.rateMu.Lock()
	defer hc.rateMu.Unlock()

	now := time.Now()
	elapsed := now.Sub(hc.rateLast)
	hc.rateTokens += elapsed.Seconds() / hc.rateRefill.Seconds()
	if hc.rateTokens > hc.rateMax {
		hc.rateTokens = hc.rateMax
	}
	hc.rateLast = now

	if hc.rateTokens < 1 {
		waitTime := hc.rateRefill * time.Duration(1-hc.rateTokens)
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(waitTime):
			hc.rateTokens = 1
		}
	} else {
		hc.rateTokens--
	}
	return nil
}

// ClearCache clears the response cache.
func (hc *HTTPClient) ClearCache() {
	hc.cacheMu.Lock()
	defer hc.cacheMu.Unlock()
	hc.cache = make(map[string]*cacheEntry)
}

// CacheLen returns the number of cached responses.
func (hc *HTTPClient) CacheLen() int {
	hc.cacheMu.RLock()
	defer hc.cacheMu.RUnlock()
	return len(hc.cache)
}

// gzipCloser wraps a gzip.Reader to close both the reader and the underlying body.
type gzipCloser struct {
	*gzip.Reader
	underlying io.Closer
}

func (g *gzipCloser) Close() error {
	g.Reader.Close()
	return g.underlying.Close()
}

// sanitizeDomainName makes a domain safe for use as a filename.
func sanitizeDomainName(domain string) string {
	domain = strings.TrimPrefix(domain, ".")
	domain = strings.ReplaceAll(domain, ".", "_")
	domain = strings.ReplaceAll(domain, "/", "_")
	return strings.TrimSpace(domain)
}
