package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"
)

// CloudflareTokenDBRecord stores encrypted Cloudflare token metadata in the hub DB.
type CloudflareTokenDBRecord struct {
	ID             string `json:"id"`
	Label          string `json:"label"`
	EncryptedToken string `json:"-"`
	AccountID      string `json:"account_id,omitempty"`
	AccountName    string `json:"account_name,omitempty"`
	ScopeSummary   string `json:"scope_summary,omitempty"`
	Enabled        bool   `json:"enabled"`
	TokenStatus    string `json:"token_status"`
	ValidationJSON string `json:"validation_json,omitempty"`
	LastValidated  int64  `json:"last_validated,omitempty"`
	LastSynced     int64  `json:"last_synced,omitempty"`
	CreatedAt      int64  `json:"created_at,omitempty"`
	UpdatedAt      int64  `json:"updated_at,omitempty"`
}

func initCloudflareSchema(db *sql.DB) error {
	schema := `
	CREATE TABLE IF NOT EXISTS cloudflare_tokens (
		id TEXT PRIMARY KEY,
		label TEXT NOT NULL,
		encrypted_token TEXT NOT NULL,
		account_id TEXT NOT NULL DEFAULT '',
		account_name TEXT NOT NULL DEFAULT '',
		scope_summary TEXT NOT NULL DEFAULT '',
		enabled INTEGER NOT NULL DEFAULT 1,
		token_status TEXT NOT NULL DEFAULT 'unknown',
		validation_json TEXT NOT NULL DEFAULT '',
		last_validated INTEGER,
		last_synced INTEGER,
		created_at INTEGER NOT NULL,
		updated_at INTEGER NOT NULL
	);
	CREATE INDEX IF NOT EXISTS idx_cloudflare_tokens_enabled ON cloudflare_tokens(enabled);
	CREATE INDEX IF NOT EXISTS idx_cloudflare_tokens_account ON cloudflare_tokens(account_id);
	`
	_, err := db.Exec(schema)
	if err != nil {
		return fmt.Errorf("init cloudflare schema: %w", err)
	}
	return nil
}

func (h *Hub) RegisterCloudflareToken(record CloudflareTokenDBRecord) error {
	timestamp := time.Now().Unix()
	if record.CreatedAt == 0 {
		record.CreatedAt = timestamp
	}
	record.UpdatedAt = timestamp
	_, err := h.db.Exec(`
		INSERT OR REPLACE INTO cloudflare_tokens
		(id, label, encrypted_token, account_id, account_name, scope_summary, enabled, token_status, validation_json, last_validated, last_synced, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, record.ID, record.Label, record.EncryptedToken, record.AccountID, record.AccountName, record.ScopeSummary, cloudflareBoolToInt(record.Enabled), record.TokenStatus, record.ValidationJSON, record.LastValidated, record.LastSynced, record.CreatedAt, record.UpdatedAt)
	if err != nil {
		return fmt.Errorf("register cloudflare token: %w", err)
	}
	return nil
}

func (h *Hub) ListCloudflareTokens() ([]CloudflareTokenDBRecord, error) {
	rows, err := h.db.Query(`
		SELECT id, label, encrypted_token, account_id, account_name, scope_summary, enabled, token_status, validation_json, last_validated, last_synced, created_at, updated_at
		FROM cloudflare_tokens
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("list cloudflare tokens: %w", err)
	}
	defer rows.Close()

	var out []CloudflareTokenDBRecord
	for rows.Next() {
		var r CloudflareTokenDBRecord
		var enabled int
		if err := rows.Scan(&r.ID, &r.Label, &r.EncryptedToken, &r.AccountID, &r.AccountName, &r.ScopeSummary, &enabled, &r.TokenStatus, &r.ValidationJSON, &r.LastValidated, &r.LastSynced, &r.CreatedAt, &r.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan cloudflare token: %w", err)
		}
		r.Enabled = enabled == 1
		out = append(out, r)
	}
	return out, nil
}

func (h *Hub) GetCloudflareToken(id string) (*CloudflareTokenDBRecord, error) {
	var r CloudflareTokenDBRecord
	var enabled int
	err := h.db.QueryRow(`
		SELECT id, label, encrypted_token, account_id, account_name, scope_summary, enabled, token_status, validation_json, last_validated, last_synced, created_at, updated_at
		FROM cloudflare_tokens WHERE id = ?
	`, id).Scan(&r.ID, &r.Label, &r.EncryptedToken, &r.AccountID, &r.AccountName, &r.ScopeSummary, &enabled, &r.TokenStatus, &r.ValidationJSON, &r.LastValidated, &r.LastSynced, &r.CreatedAt, &r.UpdatedAt)
	if err != nil {
		return nil, err
	}
	r.Enabled = enabled == 1
	return &r, nil
}

func (h *Hub) UpdateCloudflareTokenValidation(id string, enabled bool, tokenStatus string, validationJSON string, accountID string, accountName string, scopeSummary string, lastValidated int64) error {
	_, err := h.db.Exec(`
		UPDATE cloudflare_tokens
		SET enabled = ?, token_status = ?, validation_json = ?, account_id = ?, account_name = ?, scope_summary = ?, last_validated = ?, updated_at = ?
		WHERE id = ?
	`, cloudflareBoolToInt(enabled), tokenStatus, validationJSON, accountID, accountName, scopeSummary, lastValidated, time.Now().Unix(), id)
	if err != nil {
		return fmt.Errorf("update cloudflare token validation: %w", err)
	}
	return nil
}

func (h *Hub) SetCloudflareTokenSecret(id string, encryptedToken string) error {
	_, err := h.db.Exec(`UPDATE cloudflare_tokens SET encrypted_token = ?, updated_at = ? WHERE id = ?`, encryptedToken, time.Now().Unix(), id)
	if err != nil {
		return fmt.Errorf("set cloudflare token secret: %w", err)
	}
	return nil
}

func cloudflareBoolToInt(v bool) int {
	if v {
		return 1
	}
	return 0
}

func jsonString(v any) string {
	b, _ := json.Marshal(v)
	return string(b)
}
