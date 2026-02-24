package auth

import (
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
