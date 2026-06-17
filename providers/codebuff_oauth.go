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

const CodebuffBaseURL = "https://codebuff.com"

type CodebuffLoginCodeReq struct {
	FingerprintId string `json:"fingerprintId"`
}

type CodebuffLoginCodeRes struct {
	LoginUrl        string `json:"loginUrl"`
	FingerprintHash string `json:"fingerprintHash"`
	ExpiresAt       string `json:"expiresAt"`
}

type CodebuffUser struct {
	Id            string `json:"id"`
	Name          string `json:"name"`
	Email         string `json:"email"`
	AuthToken     string `json:"authToken"`
	FingerprintId string `json:"fingerprintId,omitempty"`
}

type CodebuffLoginStatusRes struct {
	User *CodebuffUser `json:"user,omitempty"`
}

type CodebuffOAuth struct {
	Client *http.Client
}

func NewCodebuffOAuth() *CodebuffOAuth {
	return &CodebuffOAuth{
		Client: &http.Client{Timeout: 10 * time.Second},
	}
}

func (c *CodebuffOAuth) GetLoginCode(fingerprintId string) (*CodebuffLoginCodeRes, error) {
	url := CodebuffBaseURL + "/api/auth/cli/code"
	reqBody, _ := json.Marshal(CodebuffLoginCodeReq{FingerprintId: fingerprintId})

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	res, err := c.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return nil, fmt.Errorf("failed to get login code: status %d", res.StatusCode)
	}

	var data CodebuffLoginCodeRes
	if err := json.NewDecoder(res.Body).Decode(&data); err != nil {
		return nil, err
	}

	return &data, nil
}

func (c *CodebuffOAuth) PollLoginStatus(fingerprintId, fingerprintHash, expiresAt string) (*CodebuffUser, error) {
	url := fmt.Sprintf("%s/api/auth/cli/status?fingerprintId=%s&fingerprintHash=%s&expiresAt=%s",
		CodebuffBaseURL, fingerprintId, fingerprintHash, expiresAt)

	// Poll every 5 seconds for up to 5 minutes
	maxAttempts := 60
	for i := 0; i < maxAttempts; i++ {
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return nil, err
		}

		res, err := c.Client.Do(req)
		if err != nil {
			time.Sleep(5 * time.Second)
			continue
		}

		if res.StatusCode == 200 {
			var data CodebuffLoginStatusRes
			if err := json.NewDecoder(res.Body).Decode(&data); err == nil && data.User != nil {
				res.Body.Close()
				return data.User, nil
			}
		}
		res.Body.Close()
		time.Sleep(5 * time.Second)
	}

	return nil, fmt.Errorf("login polling timed out")
}

// Credentials Management

type CredentialsFile struct {
	Default *CodebuffUser `json:"default,omitempty"`
}

func GetConfigDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		home = "."
	}
	return filepath.Join(home, ".config", "ti_supreme")
}

func GetCredentialsPath() string {
	return filepath.Join(GetConfigDir(), "credentials.json")
}

func SaveUserCredentials(user *CodebuffUser) error {
	dir := GetConfigDir()
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	// Encrypt the token on the fly
	userCopy := *user
	if userCopy.AuthToken != "" {
		enc, err := EncryptToken([]byte(userCopy.AuthToken))
		if err == nil {
			userCopy.AuthToken = enc
		}
	}

	path := GetCredentialsPath()
	creds := CredentialsFile{Default: &userCopy}

	data, err := json.MarshalIndent(creds, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0600)
}

func GetUserCredentials() (*CodebuffUser, error) {
	path := GetCredentialsPath()
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var creds CredentialsFile
	if err := json.Unmarshal(data, &creds); err != nil {
		return nil, err
	}

	if creds.Default == nil {
		return nil, fmt.Errorf("no default credentials found")
	}

	// Decrypt the token on the fly
	if creds.Default.AuthToken != "" {
		dec, err := DecryptToken(creds.Default.AuthToken)
		if err == nil {
			creds.Default.AuthToken = string(dec)
		}
	}

	return creds.Default, nil
}
