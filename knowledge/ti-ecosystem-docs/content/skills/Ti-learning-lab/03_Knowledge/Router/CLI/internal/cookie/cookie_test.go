package cookie_test

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/ti/cli/internal/cookie"
)

func TestNewManager(t *testing.T) {
	m := cookie.NewManager()
	if m == nil || m.Len() != 0 {
		t.Fatal("expected initialized empty Manager")
	}
}

func TestLoadChromeJSON(t *testing.T) {
	// User's real cookie export
	chromeJSON := `[
		{
			"domain": "chat.sharedchat.fun",
			"expirationDate": 1776057245.362134,
			"name": "gfsessionid",
			"path": "/",
			"value": "9vf9j001ah2l8edhqx4ena45r0504l63"
		},
		{
			"domain": "chat.sharedchat.fun",
			"expirationDate": 1807507393,
			"name": "_dd_s",
			"path": "/",
			"sameSite": "strict",
			"value": "aid=36e21b42"
		}
	]`

	m := cookie.NewManager()
	err := m.LoadChromeJSON([]byte(chromeJSON), "sharedchat")
	if err != nil {
		t.Fatal(err)
	}

	if m.Len() != 1 {
		t.Fatalf("expected 1 profile, got %d", m.Len())
	}

	sid, ok := m.GetSessionID("sharedchat")
	if !ok {
		t.Fatal("expected gfsessionid to exist")
	}
	if sid != "9vf9j001ah2l8edhqx4ena45r0504l63" {
		t.Fatalf("wrong gfsessionid: %s", sid)
	}
}

func TestGetHeader(t *testing.T) {
	m := cookie.NewManager()
	err := m.LoadChromeJSON([]byte(`[
		{"domain": "example.com", "name": "sid", "value": "abc123"}
	]`), "test")
	if err != nil {
		t.Fatal(err)
	}

	header := m.GetHeader("test")
	if header != "sid=abc123" {
		t.Fatalf("expected 'sid=abc123', got %q", header)
	}
}

func TestGetSessionIDMissing(t *testing.T) {
	m := cookie.NewManager()
	err := m.LoadChromeJSON([]byte(`[
		{"domain": "example.com", "name": "other", "value": "val"}
	]`), "test")
	if err != nil {
		t.Fatal(err)
	}
	_, ok := m.GetSessionID("test")
	if ok {
		t.Fatal("expected no gfsessionid")
	}
}

func TestIsExpired(t *testing.T) {
	m := cookie.NewManager()
	// Expired cookie (epoch time in past)
	err := m.LoadChromeJSON([]byte(`[
		{"domain": "example.com", "name": "sid", "value": "abc", "expirationDate": 1000000000}
	]`), "expired")
	if err != nil {
		t.Fatal(err)
	}
	if !m.IsExpired("expired") {
		t.Fatal("expected cookie to be expired")
	}
}

func TestNotExpired(t *testing.T) {
	m := cookie.NewManager()
	// Far future expiration
	err := m.LoadChromeJSON([]byte(`[
		{"domain": "example.com", "name": "sid", "value": "abc", "expirationDate": 1999999999}
	]`), "fresh")
	if err != nil {
		t.Fatal(err)
	}
	if m.IsExpired("fresh") {
		t.Fatal("expected cookie NOT to be expired")
	}
}

func TestRemoveExpired(t *testing.T) {
	m := cookie.NewManager()
	err := m.LoadChromeJSON([]byte(`[
		{"domain": "example.com", "name": "expired", "value": "a", "expirationDate": 1000000000},
		{"domain": "example.com", "name": "fresh", "value": "b", "expirationDate": 1999999999}
	]`), "test")
	if err != nil {
		t.Fatal(err)
	}

	m.RemoveExpired("test")
	cks, _ := m.GetHTTPCookies("test")
	if len(cks) != 1 {
		t.Fatalf("expected 1 cookie after cleanup, got %d", len(cks))
	}
	if cks[0].Name != "fresh" {
		t.Fatalf("expected 'fresh' cookie, got %s", cks[0].Name)
	}
}

func TestGetCookieJar(t *testing.T) {
	m := cookie.NewManager()
	err := m.LoadChromeJSON([]byte(`[
		{"domain": "example.com", "name": "sid", "value": "abc"}
	]`), "test")
	if err != nil {
		t.Fatal(err)
	}

	jar, err := m.GetCookieJar("test")
	if err != nil {
		t.Fatal(err)
	}
	if jar == nil {
		t.Fatal("expected cookie jar")
	}
}

func TestHTTPCookies(t *testing.T) {
	m := cookie.NewManager()
	err := m.LoadChromeJSON([]byte(`[
		{"domain": "example.com", "name": "sid", "value": "abc123", "path": "/"}
	]`), "test")
	if err != nil {
		t.Fatal(err)
	}

	cks, ok := m.GetHTTPCookies("test")
	if !ok {
		t.Fatal("expected profile to exist")
	}
	if len(cks) != 1 {
		t.Fatalf("expected 1 cookie, got %d", len(cks))
	}
	if cks[0].Name != "sid" || cks[0].Value != "abc123" {
		t.Fatalf("wrong cookie: %+v", cks[0])
	}
}

func TestEncryptDecrypt(t *testing.T) {
	key := make([]byte, 32) // 256-bit key
	for i := range key {
		key[i] = byte(i)
	}

	plaintext := "hello world"
	encrypted, err := cookie.EncryptBytes([]byte(plaintext), key)
	if err != nil {
		t.Fatal(err)
	}

	decrypted, err := cookie.DecryptBytes(encrypted, key)
	if err != nil {
		t.Fatal(err)
	}
	if decrypted != plaintext {
		t.Fatalf("expected %q, got %q", plaintext, decrypted)
	}
}

func TestSaveAndLoad(t *testing.T) {
	m := cookie.NewManager()
	err := m.LoadChromeJSON([]byte(`[
		{"domain": "example.com", "name": "sid", "value": "abc123"}
	]`), "test")
	if err != nil {
		t.Fatal(err)
	}

	dir := t.TempDir()
	file := dir + "/cookies.json"

	err = m.SaveFile("test", file)
	if err != nil {
		t.Fatal(err)
	}

	m2 := cookie.NewManager()
	err = m2.LoadFile(file, "test2")
	if err != nil {
		t.Fatal(err)
	}

	// Verify cookies loaded correctly
	cks, ok := m2.GetHTTPCookies("test2")
	if !ok || len(cks) == 0 {
		t.Fatal("expected cookies after round-trip")
	}
	if cks[0].Name != "sid" || cks[0].Value != "abc123" {
		t.Fatalf("expected sid=abc123, got %s=%s", cks[0].Name, cks[0].Value)
	}
}

func TestProfileNames(t *testing.T) {
	m := cookie.NewManager()
	m.LoadChromeJSON([]byte(`[{"domain": "a.com", "name": "k", "value": "v"}]`), "a")
	m.LoadChromeJSON([]byte(`[{"domain": "b.com", "name": "k", "value": "v"}]`), "b")
	names := m.ProfileNames()
	if len(names) != 2 {
		t.Fatalf("expected 2 names, got %d", len(names))
	}
}

func TestRemove(t *testing.T) {
	m := cookie.NewManager()
	m.LoadChromeJSON([]byte(`[{"domain": "a.com", "name": "k", "value": "v"}]`), "del")
	m.Remove("del")
	if m.Len() != 0 {
		t.Fatal("expected 0 profiles after remove")
	}
}

// =============================================================================
// Encryption tests
// =============================================================================

func TestEncryptDecryptRoundTrip(t *testing.T) {
	key := make([]byte, 32)
	for i := range key {
		key[i] = byte(i)
	}

	tests := []string{
		"hello world",
		"",
		"a",
		"test with special chars: !@#$%^&*()",
		"unicode: \u4e2d\u6587\u6d4b\u8bd5",
	}

	for _, tc := range tests {
		enc, err := cookie.EncryptGCM([]byte(tc), key)
		if err != nil {
			t.Fatalf("encrypt failed for %q: %v", tc, err)
		}
		dec, err := cookie.DecryptGCM(enc, key)
		if err != nil {
			t.Fatalf("decrypt failed for %q: %v", tc, err)
		}
		if dec != tc {
			t.Fatalf("round-trip failed for %q: got %q", tc, dec)
		}
	}
}

func TestEncryptDifferentNonceEachTime(t *testing.T) {
	key := make([]byte, 32)
	for i := range key {
		key[i] = byte(i)
	}

	enc1, _ := cookie.EncryptGCM([]byte("test"), key)
	enc2, _ := cookie.EncryptGCM([]byte("test"), key)

	if enc1 == enc2 {
		t.Fatal("expected different ciphertext due to random nonce")
	}
}

func TestDecryptWrongKey(t *testing.T) {
	key1 := make([]byte, 32)
	key2 := make([]byte, 32)
	for i := range key1 {
		key1[i] = byte(i)
		key2[i] = byte(i + 1)
	}

	enc, _ := cookie.EncryptGCM([]byte("secret"), key1)
	_, err := cookie.DecryptGCM(enc, key2)
	if err == nil {
		t.Fatal("expected error when decrypting with wrong key")
	}
}

func TestDecryptTamperedData(t *testing.T) {
	key := make([]byte, 32)
	for i := range key {
		key[i] = byte(i)
	}

	enc, _ := cookie.EncryptGCM([]byte("data"), key)
	// Tamper with the last byte
	tampered := enc[:len(enc)-1] + "X"
	_, err := cookie.DecryptGCM(tampered, key)
	if err == nil {
		t.Fatal("expected error when decrypting tampered data")
	}
}

func TestDecryptInvalidBase64(t *testing.T) {
	key := make([]byte, 32)
	_, err := cookie.DecryptGCM("not-valid-base64!!!", key)
	if err == nil {
		t.Fatal("expected error for invalid base64")
	}
}

func TestEncryptEmptyKey(t *testing.T) {
	_, err := cookie.EncryptGCM([]byte("data"), []byte{})
	if err == nil {
		t.Fatal("expected error for empty key")
	}
}

func TestSaveAndLoadEncrypted(t *testing.T) {
	key := make([]byte, 32)
	for i := range key {
		key[i] = byte(i)
	}

	m := cookie.NewManager().WithKey(key)
	err := m.LoadChromeJSON([]byte(`[
		{"domain": "example.com", "name": "secret", "value": "top-secret"}
	]`), "secure")
	if err != nil {
		t.Fatal(err)
	}

	dir := t.TempDir()
	file := dir + "/cookies.enc"

	err = m.SaveEncrypted("secure", file)
	if err != nil {
		t.Fatal(err)
	}

	m2 := cookie.NewManager().WithKey(key)
	err = m2.LoadEncrypted(file, "secure2")
	if err != nil {
		t.Fatal(err)
	}

	sid, ok := m2.GetSessionID("secure2")
	// Check at least the profile was loaded
	cks, ok := m2.GetHTTPCookies("secure2")
	if !ok || len(cks) == 0 {
		t.Fatal("expected cookies after encrypted round-trip")
	}
	if cks[0].Name != "secret" || cks[0].Value != "top-secret" {
		t.Fatalf("expected secret=top-secret, got %s=%s", cks[0].Name, cks[0].Value)
	}
	_ = sid
}

func TestSaveEncryptedNoKey(t *testing.T) {
	m := cookie.NewManager()
	m.LoadChromeJSON([]byte(`[{"domain": "a.com", "name": "k", "value": "v"}]`), "test")
	err := m.SaveEncrypted("test", "/tmp/test.enc")
	if err == nil {
		t.Fatal("expected error when saving without encryption key")
	}
}

func TestLoadEncryptedNoKey(t *testing.T) {
	m := cookie.NewManager()
	err := m.LoadEncrypted("/tmp/nonexistent.enc", "test")
	if err == nil {
		t.Fatal("expected error when loading without encryption key")
	}
}

// =============================================================================
// Multi-profile tests
// =============================================================================

func TestMultiProfile(t *testing.T) {
	m := cookie.NewManager()
	m.LoadChromeJSON([]byte(`[{"domain": "a.com", "name": "k", "value": "v1"}]`), "profile-a")
	m.LoadChromeJSON([]byte(`[{"domain": "b.com", "name": "k", "value": "v2"}]`), "profile-b")

	cksA, _ := m.GetHTTPCookies("profile-a")
	cksB, _ := m.GetHTTPCookies("profile-b")

	if cksA[0].Value != "v1" {
		t.Fatalf("expected v1 for profile-a, got %s", cksA[0].Value)
	}
	if cksB[0].Value != "v2" {
		t.Fatalf("expected v2 for profile-b, got %s", cksB[0].Value)
	}
}

func TestGetHeaderMissingProfile(t *testing.T) {
	m := cookie.NewManager()
	if got := m.GetHeader("nonexistent"); got != "" {
		t.Fatalf("expected empty header for missing profile, got %q", got)
	}
}

func TestGetCookieJarMissingProfile(t *testing.T) {
	m := cookie.NewManager()
	_, err := m.GetCookieJar("nonexistent")
	if err == nil {
		t.Fatal("expected error for missing profile")
	}
}

func TestIsExpiredMissingProfile(t *testing.T) {
	m := cookie.NewManager()
	if !m.IsExpired("nonexistent") {
		t.Fatal("expected true for missing profile")
	}
}

func TestRemoveExpiredMissingProfile(t *testing.T) {
	m := cookie.NewManager()
	m.RemoveExpired("nonexistent") // Should not panic
}

func TestUpdateCookies(t *testing.T) {
	m := cookie.NewManager()
	m.LoadChromeJSON([]byte(`[{"domain": "a.com", "name": "old", "value": "v"}]`), "test")

	newCookies := []*cookie.Cookie{
		{Name: "new1", Value: "v1", Domain: "a.com", Path: "/"},
		{Name: "new2", Value: "v2", Domain: "a.com", Path: "/"},
	}
	m.UpdateCookies("test", newCookies)

	cks, _ := m.GetHTTPCookies("test")
	if len(cks) != 2 {
		t.Fatalf("expected 2 cookies, got %d", len(cks))
	}
}

func TestUpdateCookiesMissingProfile(t *testing.T) {
	m := cookie.NewManager()
	m.UpdateCookies("nonexistent", nil) // Should not panic
}

func TestGetHTTPCookiesMissingProfile(t *testing.T) {
	m := cookie.NewManager()
	_, ok := m.GetHTTPCookies("nonexistent")
	if ok {
		t.Fatal("expected false for missing profile")
	}
}

func TestGetSessionIDMissingProfile(t *testing.T) {
	m := cookie.NewManager()
	_, ok := m.GetSessionID("nonexistent")
	if ok {
		t.Fatal("expected false for missing profile")
	}
}

// =============================================================================
// HTTPClient tests
// =============================================================================

func TestNewHTTPClient(t *testing.T) {
	m := cookie.NewManager()
	m.LoadChromeJSON([]byte(`[{"domain": "example.com", "name": "sid", "value": "abc"}]`), "test")

	hc := cookie.NewHTTPClient(m, "test")
	if hc == nil {
		t.Fatal("expected non-nil HTTPClient")
	}
}

func TestHTTPClientGet(t *testing.T) {
	m := cookie.NewManager()
	m.LoadChromeJSON([]byte(`[{"domain": "127.0.0.1", "name": "sid", "value": "abc"}]`), "test")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check that cookie was sent
		ck, err := r.Cookie("sid")
		if err != nil {
			t.Fatal("expected cookie 'sid' in request")
		}
		if ck.Value != "abc" {
			t.Fatalf("expected cookie value 'abc', got %q", ck.Value)
		}
		w.Write([]byte("ok"))
	}))
	defer server.Close()

	hc := cookie.NewHTTPClient(m, "test").WithMaxRetries(0, 0)
	resp, err := hc.Get(server.URL)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
}

func TestHTTPClientPost(t *testing.T) {
	m := cookie.NewManager()
	m.LoadChromeJSON([]byte(`[{"domain": "127.0.0.1", "name": "sid", "value": "abc"}]`), "test")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", r.Method)
		}
		w.Write([]byte("posted"))
	}))
	defer server.Close()

	hc := cookie.NewHTTPClient(m, "test").WithMaxRetries(0, 0)
	resp, err := hc.Post(server.URL, "application/json", []byte(`{"key":"val"}`))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
}

func TestHTTPClientResponseCookies(t *testing.T) {
	m := cookie.NewManager()
	m.LoadChromeJSON([]byte(`[{"domain": "127.0.0.1", "name": "sid", "value": "abc"}]`), "test")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.SetCookie(w, &http.Cookie{
			Name:  "newcookie",
			Value: "newval",
		})
		w.Write([]byte("ok"))
	}))
	defer server.Close()

	hc := cookie.NewHTTPClient(m, "test").WithMaxRetries(0, 0)
	resp, err := hc.Get(server.URL)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	// Check that response cookie was saved
	cks, _ := m.GetHTTPCookies("test")
	found := false
	for _, ck := range cks {
		if ck.Name == "newcookie" && ck.Value == "newval" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected 'newcookie' to be saved, got cookies: %v", cks)
	}
}

func TestHTTPClientCaching(t *testing.T) {
	m := cookie.NewManager()
	m.LoadChromeJSON([]byte(`[{"domain": "127.0.0.1", "name": "sid", "value": "abc"}]`), "test")

	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.Write([]byte("response"))
	}))
	defer server.Close()

	hc := cookie.NewHTTPClient(m, "test").WithMaxRetries(0, 0).WithCacheTTL(1 * time.Second)

	// First call
	resp1, err := hc.Get(server.URL)
	if err != nil {
		t.Fatal(err)
	}
	resp1.Body.Close()

	// Second call (should be cached)
	resp2, err := hc.Get(server.URL)
	if err != nil {
		t.Fatal(err)
	}
	resp2.Body.Close()

	if callCount != 1 {
		t.Fatalf("expected 1 server call (cached), got %d", callCount)
	}

	if hc.CacheLen() != 1 {
		t.Fatalf("expected cache size 1, got %d", hc.CacheLen())
	}
}

func TestHTTPClientClearCache(t *testing.T) {
	m := cookie.NewManager()
	m.LoadChromeJSON([]byte(`[{"domain": "127.0.0.1", "name": "sid", "value": "abc"}]`), "test")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	}))
	defer server.Close()

	hc := cookie.NewHTTPClient(m, "test").WithMaxRetries(0, 0).WithCacheTTL(1 * time.Second)
	hc.Get(server.URL)

	if hc.CacheLen() != 1 {
		t.Fatalf("expected cache size 1")
	}

	hc.ClearCache()
	if hc.CacheLen() != 0 {
		t.Fatalf("expected cache size 0 after clear")
	}
}

func TestHTTPClientCacheExpires(t *testing.T) {
	m := cookie.NewManager()
	m.LoadChromeJSON([]byte(`[{"domain": "127.0.0.1", "name": "sid", "value": "abc"}]`), "test")

	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.Write([]byte("ok"))
	}))
	defer server.Close()

	hc := cookie.NewHTTPClient(m, "test").WithMaxRetries(0, 0).WithCacheTTL(50 * time.Millisecond)

	// First call
	resp, _ := hc.Get(server.URL)
	resp.Body.Close()

	// Wait for cache to expire
	time.Sleep(100 * time.Millisecond)

	// Second call (cache expired, should hit server)
	resp, _ = hc.Get(server.URL)
	resp.Body.Close()

	if callCount != 2 {
		t.Fatalf("expected 2 server calls after cache expiry, got %d", callCount)
	}
}

func TestHTTPClientRateLimiting(t *testing.T) {
	m := cookie.NewManager()
	m.LoadChromeJSON([]byte(`[{"domain": "127.0.0.1", "name": "sid", "value": "abc"}]`), "test")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	}))
	defer server.Close()

	hc := cookie.NewHTTPClient(m, "test").WithMaxRetries(0, 0).WithRateLimit(5, 100*time.Millisecond)

	// Make 5 requests (within burst limit)
	for i := 0; i < 5; i++ {
		resp, err := hc.Get(server.URL)
		if err != nil {
			t.Fatalf("request %d failed: %v", i, err)
		}
		resp.Body.Close()
	}
}

func TestHTTPClientWithCookieFile(t *testing.T) {
	m := cookie.NewManager()
	m.LoadChromeJSON([]byte(`[{"domain": "127.0.0.1", "name": "sid", "value": "abc"}]`), "test")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	}))
	defer server.Close()

	dir := t.TempDir()
	file := dir + "/cookies.json"

	hc := cookie.NewHTTPClient(m, "test").WithMaxRetries(0, 0).WithCookieFile(file)
	resp, err := hc.Get(server.URL)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
}

// =============================================================================
// Concurrency tests
// =============================================================================

func TestConcurrentProfileAccess(t *testing.T) {
	m := cookie.NewManager()

	// Create profiles
	for i := 0; i < 10; i++ {
		m.LoadChromeJSON([]byte(`[{"domain": "a.com", "name": "k", "value": "v"}]`), "profile")
	}

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(4)
		go func() {
			defer wg.Done()
			m.GetHeader("profile")
		}()
		go func() {
			defer wg.Done()
			m.GetSessionID("profile")
		}()
		go func() {
			defer wg.Done()
			m.IsExpired("profile")
		}()
		go func() {
			defer wg.Done()
			m.ProfileNames()
		}()
	}
	wg.Wait()
}

func TestConcurrentProfileCreateAndDelete(t *testing.T) {
	m := cookie.NewManager()

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		name := "p" + string(rune('a'+i%26))
		wg.Add(2)
		go func(n string) {
			defer wg.Done()
			m.LoadChromeJSON([]byte(`[{"domain": "a.com", "name": "k", "value": "v"}]`), n)
		}(name)
		go func(n string) {
			defer wg.Done()
			m.Remove(n)
		}(name)
	}
	wg.Wait()
	// Should not panic or deadlock
}

// =============================================================================
// Edge cases
// =============================================================================

func TestLoadChromeJSONEmpty(t *testing.T) {
	m := cookie.NewManager()
	err := m.LoadChromeJSON([]byte(`[]`), "empty")
	if err == nil {
		t.Fatal("expected error for empty cookie array")
	}
}

func TestLoadChromeJSONInvalid(t *testing.T) {
	m := cookie.NewManager()
	err := m.LoadChromeJSON([]byte(`[invalid]`), "bad")
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestLoadChromeJSONEmptyProfileName(t *testing.T) {
	m := cookie.NewManager()
	err := m.LoadChromeJSON([]byte(`[{"domain": ".example.com", "name": "k", "value": "v"}]`), "")
	if err != nil {
		t.Fatal(err)
	}
	names := m.ProfileNames()
	if len(names) != 1 {
		t.Fatalf("expected 1 profile, got %d", len(names))
	}
	// Domain should be sanitized: .example.com -> example_com
	if names[0] != "example_com" {
		t.Fatalf("expected profile name 'example_com', got %q", names[0])
	}
}

func TestLoadChromeJSONSessionCookies(t *testing.T) {
	m := cookie.NewManager()
	err := m.LoadChromeJSON([]byte(`[
		{"domain": "a.com", "name": "session", "value": "v", "session": true}
	]`), "test")
	if err != nil {
		t.Fatal(err)
	}
	// Session cookies should have expire=0 (never expire)
	if m.IsExpired("test") {
		t.Fatal("session cookies should not be expired")
	}
}

func TestSanitizeDomain(t *testing.T) {
	// Test via empty profile name LoadChromeJSON
	m := cookie.NewManager()
	m.LoadChromeJSON([]byte(`[{"domain": ".sub.example.com", "name": "k", "value": "v"}]`), "")
	names := m.ProfileNames()
	if names[0] != "sub_example_com" {
		t.Fatalf("expected 'sub_example_com', got %q", names[0])
	}
}

func TestSaveFileMissingProfile(t *testing.T) {
	m := cookie.NewManager()
	err := m.SaveFile("nonexistent", "/tmp/test.json")
	if err == nil {
		t.Fatal("expected error for missing profile")
	}
}

func TestSaveFileCreatesDirectory(t *testing.T) {
	m := cookie.NewManager()
	m.LoadChromeJSON([]byte(`[{"domain": "a.com", "name": "k", "value": "v"}]`), "test")

	dir := t.TempDir()
	file := dir + "/sub/nested/cookies.json"

	err := m.SaveFile("test", file)
	if err != nil {
		t.Fatal(err)
	}
}

func TestWithTLSClientCertMissingFile(t *testing.T) {
	m := cookie.NewManager()
	hc := cookie.NewHTTPClient(m, "test")
	err := hc.WithTLSClientCert("/nonexistent/cert.pem", "/nonexistent/key.pem")
	if err == nil {
		t.Fatal("expected error for missing cert file")
	}
}

func TestWithCACertMissingFile(t *testing.T) {
	m := cookie.NewManager()
	hc := cookie.NewHTTPClient(m, "test")
	err := hc.WithCACert("/nonexistent/ca.pem")
	if err == nil {
		t.Fatal("expected error for missing CA file")
	}
}

func TestCookieJarWithMultipleCookies(t *testing.T) {
	m := cookie.NewManager()
	m.LoadChromeJSON([]byte(`[
		{"domain": "example.com", "name": "a", "value": "1"},
		{"domain": "example.com", "name": "b", "value": "2"}
	]`), "test")

	jar, err := m.GetCookieJar("test")
	if err != nil {
		t.Fatal(err)
	}
	if jar == nil {
		t.Fatal("expected non-nil jar")
	}
}

func TestCookieJarEmptyProfile(t *testing.T) {
	m := cookie.NewManager()
	m.LoadChromeJSON([]byte(`[{"domain": "example.com", "name": "k", "value": "v"}]`), "test")
	m.RemoveExpired("test") // No-op

	jar, err := m.GetCookieJar("test")
	if err != nil {
		t.Fatal(err)
	}
	if jar == nil {
		t.Fatal("expected non-nil jar")
	}
}

// =============================================================================
// Benchmarks
// =============================================================================

func BenchmarkEncryptGCM(b *testing.B) {
	key := make([]byte, 32)
	for i := range key {
		key[i] = byte(i)
	}
	data := []byte("test data for encryption benchmark")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cookie.EncryptGCM(data, key)
	}
}

func BenchmarkDecryptGCM(b *testing.B) {
	key := make([]byte, 32)
	for i := range key {
		key[i] = byte(i)
	}
	data := []byte("test data for decryption benchmark")
	enc, _ := cookie.EncryptGCM(data, key)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cookie.DecryptGCM(enc, key)
	}
}

func BenchmarkGetHeader(b *testing.B) {
	m := cookie.NewManager()
	cookies := make([]map[string]interface{}, 100)
	for i := 0; i < 100; i++ {
		cookies[i] = map[string]interface{}{
			"domain": "example.com",
			"name":   "k" + string(rune('a'+i%26)),
			"value":  "v",
		}
	}
	// Build JSON manually
	jsonStr := `[`
	for i, c := range cookies {
		if i > 0 {
			jsonStr += ","
		}
		jsonStr += `{"domain":"example.com","name":"` + c["name"].(string) + `","value":"v"}`
	}
	jsonStr += `]`

	m.LoadChromeJSON([]byte(jsonStr), "bench")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.GetHeader("bench")
	}
}

func BenchmarkConcurrentRead(b *testing.B) {
	m := cookie.NewManager()
	m.LoadChromeJSON([]byte(`[{"domain": "example.com", "name": "sid", "value": "abc123"}]`), "bench")

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			m.GetHeader("bench")
		}
	})
}
