package executor

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	coreauth "github.com/router-for-me/CLIProxyAPI/v7/sdk/cliproxy/auth"
	clipexec "github.com/router-for-me/CLIProxyAPI/v7/sdk/cliproxy/executor"
)

const (
	cookieProviderAttrBaseURL    = "base_url"
	cookieProviderAttrCookieEnv  = "cookie_env"
	cookieProviderAttrCookieFile = "cookie_file"
	cookieProviderAttrHeaders    = "headers_json"
	cookieProviderAttrPath       = "path"
)

// CookieProviderRuntime carries secrets loaded from config without persisting them
// into the auth storage JSON. The YAML config remains the source of truth.
type CookieProviderRuntime struct {
	Cookie   string
	Headers  map[string]string
	ModelMap map[string]string
}

// CookieProviderExecutor adapts a cookie-authenticated OpenAI-compatible upstream.
// It intentionally does not read browser cookie databases, automate login, or refresh
// cookies. Operators must supply cookies explicitly through config/env/file.
type CookieProviderExecutor struct {
	provider string
}

func NewCookieProviderExecutor(provider string) *CookieProviderExecutor {
	return &CookieProviderExecutor{provider: strings.TrimSpace(provider)}
}

func (e *CookieProviderExecutor) Identifier() string { return e.provider }

func (e *CookieProviderExecutor) Execute(ctx context.Context, auth *coreauth.Auth, req clipexec.Request, opts clipexec.Options) (clipexec.Response, error) {
	endpoint, err := e.endpoint(auth)
	if err != nil {
		return clipexec.Response{}, err
	}

	payload := rewritePayloadModel(req.Payload, auth, req.Model)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(payload))
	if err != nil {
		return clipexec.Response{}, err
	}
	copyClientHeaders(httpReq.Header, opts.Headers)
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")

	resp, err := e.HttpRequest(ctx, auth, httpReq)
	if err != nil {
		return clipexec.Response{}, err
	}
	defer resp.Body.Close()

	payload, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return clipexec.Response{}, readErr
	}
	out := clipexec.Response{Payload: payload, Headers: resp.Header}
	if resp.StatusCode >= http.StatusBadRequest {
		return out, &coreauth.Error{HTTPStatus: resp.StatusCode, Message: fmt.Sprintf("cookie provider upstream returned HTTP %d", resp.StatusCode)}
	}
	return out, nil
}

func (e *CookieProviderExecutor) ExecuteStream(ctx context.Context, auth *coreauth.Auth, req clipexec.Request, opts clipexec.Options) (*clipexec.StreamResult, error) {
	endpoint, err := e.endpoint(auth)
	if err != nil {
		return nil, err
	}

	payload := rewritePayloadModel(req.Payload, auth, req.Model)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	copyClientHeaders(httpReq.Header, opts.Headers)
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "text/event-stream")

	resp, err := e.HttpRequest(ctx, auth, httpReq)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= http.StatusBadRequest {
		defer resp.Body.Close()
		payload, _ := io.ReadAll(resp.Body)
		return &clipexec.StreamResult{Headers: resp.Header}, &coreauth.Error{HTTPStatus: resp.StatusCode, Message: fmt.Sprintf("cookie provider upstream returned HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(payload)))}
	}

	chunks := make(chan clipexec.StreamChunk)
	go func() {
		defer close(chunks)
		defer resp.Body.Close()

		reader := bufio.NewReader(resp.Body)
		for {
			line, readErr := reader.ReadBytes('\n')
			if len(line) > 0 {
				chunks <- clipexec.StreamChunk{Payload: line}
			}
			if readErr != nil {
				if !errors.Is(readErr, io.EOF) {
					chunks <- clipexec.StreamChunk{Err: readErr}
				}
				return
			}
		}
	}()

	return &clipexec.StreamResult{Headers: resp.Header, Chunks: chunks}, nil
}

func (e *CookieProviderExecutor) Refresh(ctx context.Context, auth *coreauth.Auth) (*coreauth.Auth, error) {
	return auth, nil
}

func (e *CookieProviderExecutor) CountTokens(ctx context.Context, auth *coreauth.Auth, req clipexec.Request, opts clipexec.Options) (clipexec.Response, error) {
	return clipexec.Response{}, &coreauth.Error{HTTPStatus: http.StatusNotImplemented, Message: "cookie provider does not implement token counting"}
}

func (e *CookieProviderExecutor) HttpRequest(ctx context.Context, auth *coreauth.Auth, req *http.Request) (*http.Response, error) {
	if req == nil {
		return nil, errors.New("nil request")
	}
	if err := e.PrepareRequest(req, auth); err != nil {
		return nil, err
	}
	return httpClientForCookieAuth(auth).Do(req)
}

func (e *CookieProviderExecutor) PrepareRequest(req *http.Request, auth *coreauth.Auth) error {
	cookie, err := resolveCookie(auth)
	if err != nil {
		return err
	}
	if cookie == "" {
		return errors.New("cookie provider requires cookie, cookie-env, or cookie-file")
	}

	for k, v := range resolveHeaders(auth) {
		if strings.EqualFold(k, "cookie") || strings.EqualFold(k, "authorization") {
			continue
		}
		req.Header.Set(k, v)
	}
	req.Header.Set("Cookie", cookie)
	return nil
}

func (e *CookieProviderExecutor) endpoint(auth *coreauth.Auth) (string, error) {
	if auth == nil || auth.Attributes == nil {
		return "", errors.New("missing cookie provider auth attributes")
	}
	baseURL := strings.TrimRight(strings.TrimSpace(auth.Attributes[cookieProviderAttrBaseURL]), "/")
	if baseURL == "" {
		return "", errors.New("cookie provider requires base-url")
	}
	path := strings.TrimSpace(auth.Attributes[cookieProviderAttrPath])
	if path == "" {
		path = "/chat/completions"
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	return baseURL + path, nil
}

func rewritePayloadModel(payload []byte, auth *coreauth.Auth, requested string) []byte {
	upstream := upstreamModelName(auth, requested)
	if upstream == "" || upstream == requested || len(payload) == 0 {
		return payload
	}
	var body map[string]any
	if err := json.Unmarshal(payload, &body); err != nil {
		return payload
	}
	body["model"] = upstream
	rewritten, err := json.Marshal(body)
	if err != nil {
		return payload
	}
	return rewritten
}

func upstreamModelName(auth *coreauth.Auth, requested string) string {
	requested = strings.TrimSpace(requested)
	if requested == "" || auth == nil {
		return requested
	}
	if runtimeCfg, ok := auth.Runtime.(*CookieProviderRuntime); ok && len(runtimeCfg.ModelMap) > 0 {
		if upstream := strings.TrimSpace(runtimeCfg.ModelMap[requested]); upstream != "" {
			return upstream
		}
	}
	return requested
}

func resolveCookie(auth *coreauth.Auth) (string, error) {
	if auth == nil {
		return "", errors.New("missing auth")
	}
	if runtimeCfg, ok := auth.Runtime.(*CookieProviderRuntime); ok && strings.TrimSpace(runtimeCfg.Cookie) != "" {
		return strings.TrimSpace(runtimeCfg.Cookie), nil
	}
	if auth.Attributes == nil {
		return "", nil
	}
	if envName := strings.TrimSpace(auth.Attributes[cookieProviderAttrCookieEnv]); envName != "" {
		return strings.TrimSpace(os.Getenv(envName)), nil
	}
	if fileName := strings.TrimSpace(auth.Attributes[cookieProviderAttrCookieFile]); fileName != "" {
		payload, err := os.ReadFile(expandUser(fileName))
		if err != nil {
			return "", err
		}
		return strings.TrimSpace(string(payload)), nil
	}
	return "", nil
}

func resolveHeaders(auth *coreauth.Auth) map[string]string {
	if auth == nil {
		return nil
	}
	if runtimeCfg, ok := auth.Runtime.(*CookieProviderRuntime); ok && len(runtimeCfg.Headers) > 0 {
		return runtimeCfg.Headers
	}
	if auth.Attributes == nil || strings.TrimSpace(auth.Attributes[cookieProviderAttrHeaders]) == "" {
		return nil
	}
	var headers map[string]string
	_ = json.Unmarshal([]byte(auth.Attributes[cookieProviderAttrHeaders]), &headers)
	return headers
}

func httpClientForCookieAuth(auth *coreauth.Auth) *http.Client {
	proxyURL := ""
	if auth != nil {
		proxyURL = strings.TrimSpace(auth.ProxyURL)
	}
	if proxyURL == "" || strings.EqualFold(proxyURL, "direct") || strings.EqualFold(proxyURL, "none") {
		return http.DefaultClient
	}
	parsed, err := url.Parse(proxyURL)
	if err != nil {
		return http.DefaultClient
	}
	transport, ok := http.DefaultTransport.(*http.Transport)
	if !ok {
		return http.DefaultClient
	}
	clone := transport.Clone()
	clone.Proxy = http.ProxyURL(parsed)
	return &http.Client{Transport: clone}
}

func copyClientHeaders(dst, src http.Header) {
	for k, values := range src {
		if skipForwardedHeader(k) {
			continue
		}
		for _, v := range values {
			dst.Add(k, v)
		}
	}
}

func skipForwardedHeader(name string) bool {
	switch strings.ToLower(strings.TrimSpace(name)) {
	case "authorization", "cookie", "connection", "content-length", "host", "keep-alive", "proxy-authenticate", "proxy-authorization", "te", "trailer", "transfer-encoding", "upgrade":
		return true
	default:
		return false
	}
}

func expandUser(path string) string {
	if path == "~" {
		if home, err := os.UserHomeDir(); err == nil {
			return home
		}
	}
	if strings.HasPrefix(path, "~/") || strings.HasPrefix(path, "~"+string(filepath.Separator)) {
		if home, err := os.UserHomeDir(); err == nil {
			return filepath.Join(home, strings.TrimPrefix(strings.TrimPrefix(path, "~/"), "~"+string(filepath.Separator)))
		}
	}
	return path
}
