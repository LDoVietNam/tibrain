package providers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

const (
	GithubClientID = "Iv1.b507a3b2077f9685" // Standard VS Code Client ID for GitHub Copilot
)

type GithubDeviceCodeReq struct {
	ClientID string `json:"client_id"`
	Scope    string `json:"scope,omitempty"`
}

type GithubDeviceCodeRes struct {
	DeviceCode      string `json:"device_code"`
	UserCode        string `json:"user_code"`
	VerificationURI string `json:"verification_uri"`
	ExpiresIn       int    `json:"expires_in"`
	Interval        int    `json:"interval"`
}

type GithubTokenReq struct {
	ClientID   string `json:"client_id"`
	DeviceCode string `json:"device_code"`
	GrantType  string `json:"grant_type"`
}

type GithubTokenRes struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	Scope       string `json:"scope"`
	Error       string `json:"error,omitempty"`
}

type GithubOAuth struct {
	Client *http.Client
}

func NewGithubOAuth() *GithubOAuth {
	return &GithubOAuth{
		Client: &http.Client{Timeout: 10 * time.Second},
	}
}

func (g *GithubOAuth) GetDeviceCode() (*GithubDeviceCodeRes, error) {
	url := "https://github.com/login/device/code"
	reqBody, _ := json.Marshal(GithubDeviceCodeReq{
		ClientID: GithubClientID,
		Scope:    "read:user",
	})

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	res, err := g.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return nil, fmt.Errorf("failed to get device code: status %d", res.StatusCode)
	}

	var data GithubDeviceCodeRes
	if err := json.NewDecoder(res.Body).Decode(&data); err != nil {
		return nil, err
	}

	return &data, nil
}

func (g *GithubOAuth) PollLoginStatus(deviceCode string, interval int) (string, error) {
	url := "https://github.com/login/oauth/access_token"
	
	if interval <= 0 {
		interval = 5
	}
	
	// Poll for up to 5 minutes (60 attempts * 5s)
	maxAttempts := 60
	for i := 0; i < maxAttempts; i++ {
		reqBody, _ := json.Marshal(GithubTokenReq{
			ClientID:   GithubClientID,
			DeviceCode: deviceCode,
			GrantType:  "urn:ietf:params:oauth:grant-type:device_code",
		})

		req, err := http.NewRequest("POST", url, bytes.NewBuffer(reqBody))
		if err != nil {
			return "", err
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")

		res, err := g.Client.Do(req)
		if err != nil {
			time.Sleep(time.Duration(interval) * time.Second)
			continue
		}

		if res.StatusCode == 200 {
			var data GithubTokenRes
			if err := json.NewDecoder(res.Body).Decode(&data); err == nil {
				res.Body.Close()
				if data.AccessToken != "" {
					return data.AccessToken, nil
				}
				if data.Error != "" && data.Error != "authorization_pending" {
					return "", fmt.Errorf("github oauth error: %s", data.Error)
				}
			}
		}
		if res.Body != nil {
			res.Body.Close()
		}
		time.Sleep(time.Duration(interval) * time.Second)
	}

	return "", fmt.Errorf("github device login polling timed out")
}

func SaveGithubToken(token string) error {
	authDir := "Z:\\06_AUTH"
	if err := os.MkdirAll(authDir, 0755); err != nil {
		return err
	}

	// Encrypt the token on the fly
	encToken, err := EncryptToken([]byte(token))
	if err != nil {
		encToken = token // fallback
	}

	payload := map[string]string{
		"access_token": encToken,
	}

	jsonData, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return err
	}

	filePath := filepath.Join(authDir, "github-copilot.oauth")
	return os.WriteFile(filePath, jsonData, 0600)
}
