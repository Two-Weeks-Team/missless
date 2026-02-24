package config

import (
	"testing"
)

func TestConfigLoad_MissingAPIKey(t *testing.T) {
	t.Setenv("GEMINI_API_KEY", "")
	t.Setenv("GCP_PROJECT_ID", "test-project")

	_, err := Load()
	if err == nil {
		t.Fatal("expected error for missing GEMINI_API_KEY, got nil")
	}
	if err.Error() != "GEMINI_API_KEY is required" {
		t.Fatalf("unexpected error: %s", err)
	}
}

func TestConfigLoad_MissingProjectID(t *testing.T) {
	t.Setenv("GEMINI_API_KEY", "test-key")
	t.Setenv("GCP_PROJECT_ID", "")

	_, err := Load()
	if err == nil {
		t.Fatal("expected error for missing GCP_PROJECT_ID, got nil")
	}
	if err.Error() != "GCP_PROJECT_ID is required" {
		t.Fatalf("unexpected error: %s", err)
	}
}

func TestConfigLoad_DefaultPort(t *testing.T) {
	t.Setenv("GEMINI_API_KEY", "test-key")
	t.Setenv("GCP_PROJECT_ID", "test-project")
	t.Setenv("PORT", "")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if cfg.Port != "18080" {
		t.Fatalf("expected default port 18080, got %s", cfg.Port)
	}
}

func TestConfigLoad_CustomPort(t *testing.T) {
	t.Setenv("GEMINI_API_KEY", "test-key")
	t.Setenv("GCP_PROJECT_ID", "test-project")
	t.Setenv("PORT", "9090")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if cfg.Port != "9090" {
		t.Fatalf("expected port 9090, got %s", cfg.Port)
	}
}

func TestConfigLoad_Defaults(t *testing.T) {
	t.Setenv("GEMINI_API_KEY", "test-key")
	t.Setenv("GCP_PROJECT_ID", "test-project")
	t.Setenv("PORT", "")
	t.Setenv("DOMAIN", "")
	t.Setenv("ENVIRONMENT", "")
	t.Setenv("FIRESTORE_DB", "")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if cfg.Domain != "localhost" {
		t.Fatalf("expected default domain localhost, got %s", cfg.Domain)
	}
	if cfg.Environment != "development" {
		t.Fatalf("expected default environment development, got %s", cfg.Environment)
	}
	if cfg.FirestoreDB != "(default)" {
		t.Fatalf("expected default firestore db (default), got %s", cfg.FirestoreDB)
	}
}

func TestConfigLoad_AllFields(t *testing.T) {
	t.Setenv("GEMINI_API_KEY", "key-123")
	t.Setenv("GCP_PROJECT_ID", "proj-456")
	t.Setenv("PORT", "3000")
	t.Setenv("DOMAIN", "example.com")
	t.Setenv("ENVIRONMENT", "production")
	t.Setenv("YOUTUBE_CLIENT_ID", "yt-client")
	t.Setenv("YOUTUBE_CLIENT_SECRET", "yt-secret")
	t.Setenv("OAUTH_REDIRECT_URL", "http://localhost/callback")
	t.Setenv("STORAGE_BUCKET", "my-bucket")
	t.Setenv("FIRESTORE_DB", "custom-db")
	t.Setenv("SESSION_SECRET", "secret-abc")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	checks := []struct {
		name, got, want string
	}{
		{"GeminiAPIKey", cfg.GeminiAPIKey, "key-123"},
		{"ProjectID", cfg.ProjectID, "proj-456"},
		{"Port", cfg.Port, "3000"},
		{"Domain", cfg.Domain, "example.com"},
		{"Environment", cfg.Environment, "production"},
		{"YouTubeClientID", cfg.YouTubeClientID, "yt-client"},
		{"YouTubeSecret", cfg.YouTubeSecret, "yt-secret"},
		{"OAuthRedirect", cfg.OAuthRedirect, "http://localhost/callback"},
		{"StorageBucket", cfg.StorageBucket, "my-bucket"},
		{"FirestoreDB", cfg.FirestoreDB, "custom-db"},
		{"SessionSecret", cfg.SessionSecret, "secret-abc"},
	}

	for _, c := range checks {
		if c.got != c.want {
			t.Errorf("%s: got %q, want %q", c.name, c.got, c.want)
		}
	}
}

func TestConfig_IsProd(t *testing.T) {
	tests := []struct {
		env  string
		want bool
	}{
		{"production", true},
		{"development", false},
		{"staging", false},
		{"", false},
	}

	for _, tt := range tests {
		c := &Config{Environment: tt.env}
		if got := c.IsProd(); got != tt.want {
			t.Errorf("IsProd() with env=%q: got %v, want %v", tt.env, got, tt.want)
		}
	}
}
