package authtoken

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type TokenProfile struct {
	Name      string    `json:"name"`
	Provider  string    `json:"provider"`
	TokenHash string    `json:"token_sha256"`
	Preview   string    `json:"preview"`
	Source    string    `json:"source,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type tokenSecret struct {
	Name     string `json:"name"`
	Provider string `json:"provider"`
	Token    string `json:"token"`
}

type Store struct{ Root string }

func DefaultRoot() string {
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return filepath.Join(".ti", "auth", "tokens")
	}
	return filepath.Join(home, ".ti", "auth", "tokens")
}

func NewStore(root string) *Store {
	if strings.TrimSpace(root) == "" {
		root = DefaultRoot()
	}
	return &Store{Root: expand(root)}
}

func expand(p string) string {
	if p == "~" || strings.HasPrefix(p, "~/") {
		if home, err := os.UserHomeDir(); err == nil && home != "" {
			if p == "~" {
				return home
			}
			return filepath.Join(home, p[2:])
		}
	}
	return os.ExpandEnv(p)
}

func (s *Store) Ensure() error     { return os.MkdirAll(s.Root, 0700) }
func (s *Store) indexPath() string { return filepath.Join(s.Root, "tokens.jsonl") }
func (s *Store) secretPath(name string) string {
	return filepath.Join(s.Root, safeName(name)+".secret.json")
}

func (s *Store) Import(name, provider, token, source string) (*TokenProfile, error) {
	name = strings.TrimSpace(name)
	provider = strings.TrimSpace(provider)
	token = strings.TrimSpace(token)
	if name == "" {
		return nil, fmt.Errorf("token profile name is required")
	}
	if provider == "" {
		provider = "generic"
	}
	if token == "" {
		return nil, fmt.Errorf("token value is empty")
	}
	if err := s.Ensure(); err != nil {
		return nil, err
	}
	now := time.Now()
	sum := sha256.Sum256([]byte(token))
	p := &TokenProfile{Name: name, Provider: provider, TokenHash: hex.EncodeToString(sum[:]), Preview: Redact(token), Source: source, CreatedAt: now, UpdatedAt: now}
	if old, ok, _ := s.GetProfile(name); ok {
		p.CreatedAt = old.CreatedAt
	}
	secret := tokenSecret{Name: name, Provider: provider, Token: token}
	b, _ := json.MarshalIndent(secret, "", "  ")
	if err := os.WriteFile(s.secretPath(name), append(b, '\n'), 0600); err != nil {
		return nil, err
	}
	if err := s.rewriteIndexWith(*p); err != nil {
		return nil, err
	}
	return p, nil
}

func (s *Store) GetSecret(name string) (string, *TokenProfile, error) {
	p, ok, err := s.GetProfile(name)
	if err != nil {
		return "", nil, err
	}
	if !ok {
		return "", nil, os.ErrNotExist
	}
	data, err := os.ReadFile(s.secretPath(name))
	if err != nil {
		return "", &p, err
	}
	var sec tokenSecret
	if err := json.Unmarshal(data, &sec); err != nil {
		return "", &p, err
	}
	return sec.Token, &p, nil
}

func (s *Store) GetProfile(name string) (TokenProfile, bool, error) {
	items, err := s.List()
	if err != nil {
		return TokenProfile{}, false, err
	}
	for _, p := range items {
		if p.Name == name {
			return p, true, nil
		}
	}
	return TokenProfile{}, false, nil
}

func (s *Store) List() ([]TokenProfile, error) {
	if err := s.Ensure(); err != nil {
		return nil, err
	}
	f, err := os.Open(s.indexPath())
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	defer f.Close()
	latest := map[string]TokenProfile{}
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}
		var p TokenProfile
		if json.Unmarshal([]byte(line), &p) == nil {
			latest[p.Name] = p
		}
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	out := make([]TokenProfile, 0, len(latest))
	for _, p := range latest {
		out = append(out, p)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out, nil
}

func (s *Store) Remove(name string) error {
	items, err := s.List()
	if err != nil {
		return err
	}
	var kept []TokenProfile
	for _, p := range items {
		if p.Name != name {
			kept = append(kept, p)
		}
	}
	if err := s.Ensure(); err != nil {
		return err
	}
	f, err := os.OpenFile(s.indexPath(), os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	enc := json.NewEncoder(f)
	for _, p := range kept {
		_ = enc.Encode(p)
	}
	_ = f.Close()
	if err := os.Remove(s.secretPath(name)); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return nil
}

func (s *Store) rewriteIndexWith(p TokenProfile) error {
	items, err := s.List()
	if err != nil {
		return err
	}
	written := false
	for i := range items {
		if items[i].Name == p.Name {
			items[i] = p
			written = true
		}
	}
	if !written {
		items = append(items, p)
	}
	sort.Slice(items, func(i, j int) bool { return items[i].Name < items[j].Name })
	f, err := os.OpenFile(s.indexPath(), os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	for _, it := range items {
		if err := enc.Encode(it); err != nil {
			return err
		}
	}
	return nil
}

func Redact(token string) string {
	token = strings.TrimSpace(token)
	if token == "" {
		return ""
	}
	if len(token) <= 10 {
		return strings.Repeat("*", len(token))
	}
	return token[:4] + strings.Repeat("*", min(8, len(token)-8)) + token[len(token)-4:]
}

func safeName(name string) string {
	name = strings.TrimSpace(strings.ToLower(name))
	var b strings.Builder
	for _, r := range name {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' || r == '_' || r == '.' {
			b.WriteRune(r)
		} else {
			b.WriteByte('_')
		}
	}
	if b.Len() == 0 {
		return "default"
	}
	return b.String()
}
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
