package auth

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// NewOAuthConfig creates an OAuth2 config for YouTube access.
func NewOAuthConfig(clientID, clientSecret, redirectURL string) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
		Scopes: []string{
			"https://www.googleapis.com/auth/youtube.readonly",
		},
		Endpoint: google.Endpoint,
	}
}

// GenerateState creates a crypto-random state token for CSRF protection.
func GenerateState() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate state: %w", err)
	}
	return hex.EncodeToString(b), nil
}
