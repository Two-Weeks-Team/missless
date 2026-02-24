package auth

import (
	"sync"

	"golang.org/x/oauth2"
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
