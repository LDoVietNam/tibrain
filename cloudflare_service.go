package main

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"
)

const (
	cloudflareAPIBaseURL = "https://api.cloudflare.com/client/v4"
)

// CloudflareAccount represents an account visible to TiBrain.
type CloudflareAccount struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at,omitempty"`
}

// CloudflareZone represents a Cloudflare zone visible to TiBrain.
type CloudflareZone struct {
	ID         string    `json:"id"`
	Name       string    `json:"name"`
	Status     string    `json:"status"`
	Paused     bool      `json:"paused"`
	Type       string    `json:"type"`
	ModifiedOn time.Time `json:"modified_on,omitempty"`
}

// CloudflareTokenRecord stores a managed Cloudflare token record.
type CloudflareTokenRecord struct {
	ID            string `json:"id"`
	Label         string `json:"label"`
	AccountID     string `json:"account_id,omitempty"`
	AccountName   string `json:"account_name,omitempty"`
	ScopeSummary  string `json:"scope_summary,omitempty"`
	Enabled       bool   `json:"enabled"`
	TokenStatus   string `json:"token_status"`
	LastValidated int64  `json:"last_validated,omitempty"`
	LastSynced    int64  `json:"last_synced,omitempty"`
	CreatedAt     int64  `json:"created_at,omitempty"`
	UpdatedAt     int64  `json:"updated_at,omitempty"`
}

// CloudflareService manages Cloudflare tokens and account metadata.
type CloudflareService struct {
	hub        *Hub
	httpClient *http.Client
	secretKey  []byte
}

func NewCloudflareService(hub *Hub) *CloudflareService {
	return &CloudflareService{
		hub: hub,
		httpClient: &http.Client{
			Timeout: 20 * time.Second,
		},
		secretKey: deriveCloudflareSecret(),
	}
}

func deriveCloudflareSecret() []byte {
	seed := getEnvOrDefault("TIBRAIN_CLOUDFLARE_SECRET", "")
	if seed == "" {
		seed = getEnvOrDefault("TIBRAIN_SECRET_KEY", "tibrain-cloudflare-secret")
	}
	b := []byte(seed)
	if len(b) >= 32 {
		return b[:32]
	}
	out := make([]byte, 32)
	copy(out, b)
	for i := len(b); i < 32; i++ {
		out[i] = byte(i*17 + 31)
	}
	return out
}

func (s *CloudflareService) ensureReady() error {
	if s == nil || s.hub == nil || s.hub.db == nil {
		return errors.New("cloudflare service not initialized")
	}
	return nil
}

func (s *CloudflareService) EncryptToken(plaintext string) (string, error) {
	if plaintext == "" {
		return "", nil
	}
	return encryptStringAESGCM(plaintext, s.secretKey)
}

func (s *CloudflareService) DecryptToken(ciphertext string) (string, error) {
	if ciphertext == "" {
		return "", nil
	}
	return decryptStringAESGCM(ciphertext, s.secretKey)
}

func (s *CloudflareService) ValidateToken(ctx context.Context, token string) (*CloudflareValidationResult, error) {
	if token == "" {
		return nil, errors.New("token is required")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, cloudflareAPIBaseURL+"/user/tokens/verify", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var payload struct {
		Success bool `json:"success"`
		Result  struct {
			ID      string `json:"id"`
			Status  string `json:"status"`
			Expires string `json:"expires_on"`
		} `json:"result"`
		Errors []struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		} `json:"errors"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, err
	}

	result := &CloudflareValidationResult{
		HTTPStatus: resp.StatusCode,
		Success:    payload.Success && resp.StatusCode >= 200 && resp.StatusCode < 300,
		TokenID:    payload.Result.ID,
		TokenState: payload.Result.Status,
	}
	if payload.Result.Expires != "" {
		if t, err := time.Parse(time.RFC3339, payload.Result.Expires); err == nil {
			result.ExpiresAt = &t
		}
	}
	if len(payload.Errors) > 0 {
		result.ErrorMessage = payload.Errors[0].Message
	}
	return result, nil
}

func (s *CloudflareService) ListAccounts(ctx context.Context, token string) ([]CloudflareAccount, error) {
	if token == "" {
		return nil, errors.New("token is required")
	}

	type cloudflareListResponse[T any] struct {
		Success bool `json:"success"`
		Result  []T  `json:"result"`
		Errors  []struct {
			Message string `json:"message"`
		} `json:"errors"`
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, cloudflareAPIBaseURL+"/accounts", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var payload cloudflareListResponse[struct {
		ID      string `json:"id"`
		Name    string `json:"name"`
		Status  string `json:"status"`
		Created string `json:"created_on"`
	}]
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, err
	}
	if !payload.Success {
		if len(payload.Errors) > 0 {
			return nil, errors.New(payload.Errors[0].Message)
		}
		return nil, errors.New("cloudflare accounts request failed")
	}

	accounts := make([]CloudflareAccount, 0, len(payload.Result))
	for _, item := range payload.Result {
		acct := CloudflareAccount{ID: item.ID, Name: item.Name, Status: item.Status}
		if item.Created != "" {
			if t, err := time.Parse(time.RFC3339, item.Created); err == nil {
				acct.CreatedAt = t
			}
		}
		accounts = append(accounts, acct)
	}
	return accounts, nil
}

func (s *CloudflareService) ListZones(ctx context.Context, token string) ([]CloudflareZone, error) {
	if token == "" {
		return nil, errors.New("token is required")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, cloudflareAPIBaseURL+"/zones?per_page=50", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var payload struct {
		Success bool `json:"success"`
		Result  []struct {
			ID         string `json:"id"`
			Name       string `json:"name"`
			Status     string `json:"status"`
			Paused     bool   `json:"paused"`
			Type       string `json:"type"`
			ModifiedOn string `json:"modified_on"`
		} `json:"result"`
		Errors []struct {
			Message string `json:"message"`
		} `json:"errors"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, err
	}
	if !payload.Success {
		if len(payload.Errors) > 0 {
			return nil, errors.New(payload.Errors[0].Message)
		}
		return nil, errors.New("cloudflare zones request failed")
	}

	zones := make([]CloudflareZone, 0, len(payload.Result))
	for _, item := range payload.Result {
		zone := CloudflareZone{ID: item.ID, Name: item.Name, Status: item.Status, Paused: item.Paused, Type: item.Type}
		if item.ModifiedOn != "" {
			if t, err := time.Parse(time.RFC3339, item.ModifiedOn); err == nil {
				zone.ModifiedOn = t
			}
		}
		zones = append(zones, zone)
	}
	return zones, nil
}

func (s *CloudflareService) NormalizeScopeSummary(scopes []string) string {
	if len(scopes) == 0 {
		return ""
	}
	return strings.Join(scopes, ",")
}

// CloudflareValidationResult captures validation output without exposing the raw token.
type CloudflareValidationResult struct {
	Success      bool       `json:"success"`
	HTTPStatus   int        `json:"http_status"`
	TokenID      string     `json:"token_id,omitempty"`
	TokenState   string     `json:"token_state,omitempty"`
	ExpiresAt    *time.Time `json:"expires_at,omitempty"`
	ErrorMessage string     `json:"error_message,omitempty"`
}
