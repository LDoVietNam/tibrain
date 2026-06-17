package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"time"
)

const (
	oauthAuthURL         = "https://accounts.google.com/o/oauth2/v2/auth"
	oauthTokenURL        = "https://oauth2.googleapis.com/token"
	userInfoURL          = "https://www.googleapis.com/oauth2/v1/userinfo"
	antigravClientID     = "YOUR_GOOGLE_CLIENT_ID"
	antigravClientSecret = "YOUR_GOOGLE_CLIENT_SECRET"
	antigravRedirectURI  = "http://localhost:51121/callback"
)

type oauthToken struct {
	Refresh string
	Access  string
	Expiry  int64  // epoch sec
	Project string // optional, can be empty
	Email   string
}

const authFile = "Z:\\06_AUTH\\auth\\antigravity.oauth"

func AntigravityAuthURL(w http.ResponseWriter, r *http.Request) {
	values := url.Values{
		"response_type": {"code"},
		"client_id":     {antigravClientID},
		"redirect_uri":  {antigravRedirectURI},
		"scope":         {"openid email profile"},
		"prompt":        {"consent"},
	}
	w.Header().Set("Content-Type", "text/plain")
	fmt.Fprint(w, fmt.Sprintf("%s?%s", oauthAuthURL, values.Encode()))
}

func AntigravityCallback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "missing code", http.StatusBadRequest)
		return
	}

	form := url.Values{
		"code":          {code},
		"client_id":     {antigravClientID},
		"client_secret": {antigravClientSecret},
		"redirect_uri":  {antigravRedirectURI},
		"grant_type":    {"authorization_code"},
	}
	resp, err := http.PostForm(oauthTokenURL, form)
	if err != nil || resp.StatusCode != http.StatusOK {
		http.Error(w, "token exchange failed", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	var token struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int    `json:"expires_in"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&token); err != nil {
		http.Error(w, "parse token failed", http.StatusInternalServerError)
		return
	}

	userResp, _ := http.Get(userInfoURL + "?access_token=" + token.AccessToken)
	var u struct {
		Email string `json:"email"`
	}
	if err := json.NewDecoder(userResp.Body).Decode(&u); err != nil {
		u.Email = ""
	}

	if err := fileWrite(authFile, oauthToken{
		Refresh: token.RefreshToken,
		Access:  token.AccessToken,
		Expiry:  time.Now().Add(time.Duration(token.ExpiresIn) * time.Second).Unix(),
		Email:   u.Email,
	}); err != nil {
		http.Error(w, "save token failed", http.StatusInternalServerError)
		return
	}
	fmt.Fprint(w, "✅ Antigravity login OK. Token saved to "+authFile)
}

func AntigravityHealth(w http.ResponseWriter, r *http.Request) {
	if _, err := http.Get(oauthAuthURL); err != nil {
		http.Error(w, "unreachable auth server", http.StatusServiceUnavailable)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func fileWrite(path string, data interface{}) error {
	b, _ := json.MarshalIndent(data, "", "  ")
	return os.WriteFile(path, b, 0o644)
}
