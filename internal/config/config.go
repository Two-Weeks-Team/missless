package config

import (
	"fmt"
	"os"
)

// Config holds all environment-driven configuration.
type Config struct {
	GeminiAPIKey    string
	ProjectID       string
	Port            string
	Domain          string
	Environment     string
	YouTubeClientID string
	YouTubeSecret   string
	OAuthRedirect   string
	StorageBucket   string
	FirestoreDB     string
	SessionSecret   string
}

// Load reads environment variables and validates required fields.
func Load() (*Config, error) {
	c := &Config{
		GeminiAPIKey:    os.Getenv("GEMINI_API_KEY"),
		ProjectID:       os.Getenv("GCP_PROJECT_ID"),
		Port:            os.Getenv("PORT"),
		Domain:          os.Getenv("DOMAIN"),
		Environment:     os.Getenv("ENVIRONMENT"),
		YouTubeClientID: os.Getenv("YOUTUBE_CLIENT_ID"),
		YouTubeSecret:   os.Getenv("YOUTUBE_CLIENT_SECRET"),
		OAuthRedirect:   os.Getenv("OAUTH_REDIRECT_URL"),
		StorageBucket:   os.Getenv("STORAGE_BUCKET"),
		FirestoreDB:     os.Getenv("FIRESTORE_DB"),
		SessionSecret:   os.Getenv("SESSION_SECRET"),
	}

	if c.Port == "" {
		c.Port = "8080"
	}
	if c.Domain == "" {
		c.Domain = "localhost"
	}
	if c.Environment == "" {
		c.Environment = "development"
	}
	if c.FirestoreDB == "" {
		c.FirestoreDB = "(default)"
	}

	if c.GeminiAPIKey == "" {
		return nil, fmt.Errorf("GEMINI_API_KEY is required")
	}
	if c.ProjectID == "" {
		return nil, fmt.Errorf("GCP_PROJECT_ID is required")
	}

	return c, nil
}

// IsProd returns true if running in production.
func (c *Config) IsProd() bool {
	return c.Environment == "production"
}
