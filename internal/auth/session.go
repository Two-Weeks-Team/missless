package auth

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"sync"
	"time"

	"golang.org/x/oauth2"
)

const (
	// SessionCookieName is the name of the session cookie.
	SessionCookieName = "missless_session"
	// StateCookieName is the name of the OAuth state cookie.
	StateCookieName = "missless_oauth_state"

	sessionTTL = 24 * time.Hour
	stateTTL   = 10 * time.Minute
)

// UserSession stores OAuth tokens for the user.
type UserSession struct {
	mu    sync.RWMutex
	Token *oauth2.Token
}

// NewUserSession creates a new user session.
func NewUserSession(token *oauth2.Token) *UserSession {
	return &UserSession{Token: token}
}

// GetToken returns the current OAuth token.
func (s *UserSession) GetToken() *oauth2.Token {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.Token
}

// UpdateToken refreshes the OAuth token.
func (s *UserSession) UpdateToken(token *oauth2.Token) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Token = token
}

// sessionEntry stores session data with expiry.
type sessionEntry struct {
	session *UserSession
	expiry  time.Time
}

// SessionStore manages user sessions and OAuth state tokens.
type SessionStore struct {
	mu       sync.RWMutex
	sessions map[string]*sessionEntry
	states   map[string]time.Time
}

// NewSessionStore creates a new session store.
func NewSessionStore() *SessionStore {
	return &SessionStore{
		sessions: make(map[string]*sessionEntry),
		states:   make(map[string]time.Time),
	}
}

// CreateSession stores a new user session and returns the session ID.
func (s *SessionStore) CreateSession(token *oauth2.Token) (string, error) {
	id, err := generateSessionID()
	if err != nil {
		return "", err
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.sessions[id] = &sessionEntry{
		session: NewUserSession(token),
		expiry:  time.Now().Add(sessionTTL),
	}
	return id, nil
}

// GetSession retrieves a session by ID, returning nil if not found or expired.
func (s *SessionStore) GetSession(id string) *UserSession {
	s.mu.RLock()
	defer s.mu.RUnlock()

	entry, ok := s.sessions[id]
	if !ok || time.Now().After(entry.expiry) {
		return nil
	}
	return entry.session
}

// StoreState saves an OAuth state token with expiry.
func (s *SessionStore) StoreState(state string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.states[state] = time.Now().Add(stateTTL)
}

// ValidateState checks and consumes a state token (one-time use).
func (s *SessionStore) ValidateState(state string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	expiry, ok := s.states[state]
	if !ok || time.Now().After(expiry) {
		return false
	}
	delete(s.states, state)
	return true
}

// GetSessionFromRequest extracts the session from the request cookie.
func (s *SessionStore) GetSessionFromRequest(r *http.Request) *UserSession {
	cookie, err := r.Cookie(SessionCookieName)
	if err != nil {
		return nil
	}
	return s.GetSession(cookie.Value)
}

// SetSessionCookie writes the session cookie to the response.
func SetSessionCookie(w http.ResponseWriter, sessionID string, isProd bool) {
	http.SetCookie(w, &http.Cookie{
		Name:     SessionCookieName,
		Value:    sessionID,
		Path:     "/",
		MaxAge:   int(sessionTTL.Seconds()),
		HttpOnly: true,
		Secure:   isProd,
		SameSite: http.SameSiteStrictMode,
	})
}

// SetStateCookie writes the OAuth state cookie.
// Uses SameSite=Lax to allow the cookie on OAuth redirect back from Google.
func SetStateCookie(w http.ResponseWriter, state string, isProd bool) {
	http.SetCookie(w, &http.Cookie{
		Name:     StateCookieName,
		Value:    state,
		Path:     "/auth/callback",
		MaxAge:   int(stateTTL.Seconds()),
		HttpOnly: true,
		Secure:   isProd,
		SameSite: http.SameSiteLaxMode,
	})
}

// GetStateFromCookie reads the state value from the cookie.
func GetStateFromCookie(r *http.Request) string {
	cookie, err := r.Cookie(StateCookieName)
	if err != nil {
		return ""
	}
	return cookie.Value
}

func generateSessionID() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate session id: %w", err)
	}
	return hex.EncodeToString(b), nil
}
