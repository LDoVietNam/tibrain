package browser

import (
	"fmt"
	"sync"
)

// BrowserType represents a browser engine
type BrowserType string

const (
	Chrome  BrowserType = "chrome"
	Firefox BrowserType = "firefox"
	Edge    BrowserType = "edge"
	Safari  BrowserType = "safari"
)

// Account represents a browser account
type Account struct {
	ID     string
	Email  string
	Name   string
	Active bool
}

// Profile represents a browser profile
type Profile struct {
	ID        string
	Name      string
	Browser   BrowserType
	Active    bool
	AccountID string
}

// Session represents an active browser session
type Session struct {
	ID        string
	ProfileID string
	Headless  bool
}

// Navigate navigates the session to a URL
func (s *Session) Navigate(url string) error {
	return fmt.Errorf("browser automation not implemented")
}

// Close closes the browser session
func (s *Session) Close() error {
	return nil
}

// BrowserManager manages browser accounts, profiles, and sessions
type BrowserManager struct {
	dataDir  string
	accounts map[string]*Account
	profiles map[string]*Profile
	sessions map[string]*Session
	mu       sync.RWMutex
}

// NewBrowserManager creates a new BrowserManager
func NewBrowserManager(dataDir string) (*BrowserManager, error) {
	return &BrowserManager{
		dataDir:  dataDir,
		accounts: make(map[string]*Account),
		profiles: make(map[string]*Profile),
		sessions: make(map[string]*Session),
	}, nil
}

// AddAccount adds a new account
func (bm *BrowserManager) AddAccount(email, name string) (*Account, error) {
	bm.mu.Lock()
	defer bm.mu.Unlock()
	id := fmt.Sprintf("acc_%d", len(bm.accounts)+1)
	acc := &Account{ID: id, Email: email, Name: name}
	bm.accounts[id] = acc
	return acc, nil
}

// GetAccount retrieves an account by ID
func (bm *BrowserManager) GetAccount(id string) *Account {
	bm.mu.RLock()
	defer bm.mu.RUnlock()
	return bm.accounts[id]
}

// RemoveAccount removes an account
func (bm *BrowserManager) RemoveAccount(id string) error {
	bm.mu.Lock()
	defer bm.mu.Unlock()
	delete(bm.accounts, id)
	return nil
}

// ListAccounts lists all accounts
func (bm *BrowserManager) ListAccounts() []Account {
	bm.mu.RLock()
	defer bm.mu.RUnlock()
	var list []Account
	for _, acc := range bm.accounts {
		list = append(list, *acc)
	}
	return list
}

// SetActiveAccount sets the active account
func (bm *BrowserManager) SetActiveAccount(id string) error {
	bm.mu.Lock()
	defer bm.mu.Unlock()
	for _, acc := range bm.accounts {
		acc.Active = acc.ID == id
	}
	return nil
}

// GetActiveAccount returns the active account
func (bm *BrowserManager) GetActiveAccount() *Account {
	bm.mu.RLock()
	defer bm.mu.RUnlock()
	for _, acc := range bm.accounts {
		if acc.Active {
			return acc
		}
	}
	return nil
}

// AddProfile adds a new profile
func (bm *BrowserManager) AddProfile(name string, browserType BrowserType, userDataDir, accountID string) (*Profile, error) {
	bm.mu.Lock()
	defer bm.mu.Unlock()
	id := fmt.Sprintf("prof_%d", len(bm.profiles)+1)
	prof := &Profile{ID: id, Name: name, Browser: browserType, AccountID: accountID}
	bm.profiles[id] = prof
	return prof, nil
}

// GetProfile retrieves a profile by ID
func (bm *BrowserManager) GetProfile(id string) *Profile {
	bm.mu.RLock()
	defer bm.mu.RUnlock()
	return bm.profiles[id]
}

// RemoveProfile removes a profile
func (bm *BrowserManager) RemoveProfile(id string) error {
	bm.mu.Lock()
	defer bm.mu.Unlock()
	delete(bm.profiles, id)
	return nil
}

// ListProfiles lists all profiles
func (bm *BrowserManager) ListProfiles() []Profile {
	bm.mu.RLock()
	defer bm.mu.RUnlock()
	var list []Profile
	for _, prof := range bm.profiles {
		list = append(list, *prof)
	}
	return list
}

// SetActiveProfile sets the active profile
func (bm *BrowserManager) SetActiveProfile(id string) error {
	bm.mu.Lock()
	defer bm.mu.Unlock()
	for _, prof := range bm.profiles {
		prof.Active = prof.ID == id
	}
	return nil
}

// GetActiveProfile returns the active profile
func (bm *BrowserManager) GetActiveProfile() *Profile {
	bm.mu.RLock()
	defer bm.mu.RUnlock()
	for _, prof := range bm.profiles {
		if prof.Active {
			return prof
		}
	}
	return nil
}

// LinkAccountToProfile links an account to a profile
func (bm *BrowserManager) LinkAccountToProfile(profileID, accountID string) error {
	bm.mu.Lock()
	defer bm.mu.Unlock()
	prof := bm.profiles[profileID]
	if prof == nil {
		return fmt.Errorf("profile not found")
	}
	prof.AccountID = accountID
	return nil
}

// NewSession creates a new browser session
func (bm *BrowserManager) NewSession(profileID string, headless bool) (*Session, error) {
	bm.mu.Lock()
	defer bm.mu.Unlock()
	id := fmt.Sprintf("sess_%d", len(bm.sessions)+1)
	sess := &Session{ID: id, ProfileID: profileID, Headless: headless}
	bm.sessions[id] = sess
	return sess, nil
}

// CloseSession closes a session
func (bm *BrowserManager) CloseSession(sessionID string) error {
	bm.mu.Lock()
	defer bm.mu.Unlock()
	delete(bm.sessions, sessionID)
	return nil
}

// GetSession retrieves a session by ID
func (bm *BrowserManager) GetSession(sessionID string) *Session {
	bm.mu.RLock()
	defer bm.mu.RUnlock()
	return bm.sessions[sessionID]
}

// GetByAccount returns profiles linked to an account
func (bm *BrowserManager) GetByAccount(accountID string) []Profile {
	bm.mu.RLock()
	defer bm.mu.RUnlock()
	var list []Profile
	for _, prof := range bm.profiles {
		if prof.AccountID == accountID {
			list = append(list, *prof)
		}
	}
	return list
}
