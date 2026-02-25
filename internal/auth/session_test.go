package auth

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// ---------------------------------------------------------------------------
// session.go — UserSession tests
// ---------------------------------------------------------------------------

func TestNewUserSession(t *testing.T) {
	token := &oauth2.Token{AccessToken: "abc123"}
	sess := NewUserSession(token)

	if sess == nil {
		t.Fatal("expected non-nil session")
	}
	if sess.Token != token {
		t.Error("session token does not match the provided token")
	}
}

func TestUserSession_GetToken(t *testing.T) {
	token := &oauth2.Token{AccessToken: "tok_get"}
	sess := NewUserSession(token)

	got := sess.GetToken()
	if got == nil {
		t.Fatal("GetToken returned nil")
	}
	if got.AccessToken != "tok_get" {
		t.Errorf("expected AccessToken %q, got %q", "tok_get", got.AccessToken)
	}
}

func TestUserSession_UpdateToken(t *testing.T) {
	original := &oauth2.Token{AccessToken: "original"}
	sess := NewUserSession(original)

	updated := &oauth2.Token{AccessToken: "refreshed"}
	sess.UpdateToken(updated)

	got := sess.GetToken()
	if got.AccessToken != "refreshed" {
		t.Errorf("expected AccessToken %q after update, got %q", "refreshed", got.AccessToken)
	}
}

// ---------------------------------------------------------------------------
// session.go — SessionStore tests
// ---------------------------------------------------------------------------

func TestNewSessionStore(t *testing.T) {
	store := NewSessionStore()

	if store == nil {
		t.Fatal("expected non-nil store")
	}
	if store.sessions == nil {
		t.Error("sessions map should be initialized")
	}
	if store.states == nil {
		t.Error("states map should be initialized")
	}
	if len(store.sessions) != 0 {
		t.Error("sessions map should be empty on creation")
	}
	if len(store.states) != 0 {
		t.Error("states map should be empty on creation")
	}
}

func TestSessionStore_CreateSession(t *testing.T) {
	store := NewSessionStore()
	token := &oauth2.Token{AccessToken: "create_tok"}

	id, err := store.CreateSession(token)
	if err != nil {
		t.Fatalf("CreateSession returned error: %v", err)
	}
	if len(id) != 64 {
		t.Errorf("expected session ID of 64 hex chars, got length %d", len(id))
	}

	// Verify all characters are valid hex.
	for _, c := range id {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
			t.Errorf("session ID contains non-hex character: %c", c)
			break
		}
	}
}

func TestSessionStore_GetSession_Valid(t *testing.T) {
	store := NewSessionStore()
	token := &oauth2.Token{AccessToken: "valid_tok"}

	id, err := store.CreateSession(token)
	if err != nil {
		t.Fatalf("CreateSession error: %v", err)
	}

	sess := store.GetSession(id)
	if sess == nil {
		t.Fatal("GetSession returned nil for a valid ID")
	}
	if sess.GetToken().AccessToken != "valid_tok" {
		t.Errorf("expected AccessToken %q, got %q", "valid_tok", sess.GetToken().AccessToken)
	}
}

func TestSessionStore_GetSession_NotFound(t *testing.T) {
	store := NewSessionStore()

	sess := store.GetSession("nonexistent_id")
	if sess != nil {
		t.Error("expected nil for unknown session ID")
	}
}

func TestSessionStore_GetSession_Expired(t *testing.T) {
	store := NewSessionStore()
	token := &oauth2.Token{AccessToken: "expired_tok"}

	id, err := store.CreateSession(token)
	if err != nil {
		t.Fatalf("CreateSession error: %v", err)
	}

	// Directly set the expiry to the past (possible because test is in the same package).
	store.mu.Lock()
	store.sessions[id].expiry = time.Now().Add(-1 * time.Hour)
	store.mu.Unlock()

	sess := store.GetSession(id)
	if sess != nil {
		t.Error("expected nil for an expired session")
	}
}

// ---------------------------------------------------------------------------
// session.go — State token tests
// ---------------------------------------------------------------------------

func TestSessionStore_StoreState(t *testing.T) {
	store := NewSessionStore()
	store.StoreState("state_abc")

	store.mu.RLock()
	_, ok := store.states["state_abc"]
	store.mu.RUnlock()

	if !ok {
		t.Error("state token was not stored")
	}
}

func TestSessionStore_ValidateState_Valid(t *testing.T) {
	store := NewSessionStore()
	store.StoreState("valid_state")

	if !store.ValidateState("valid_state") {
		t.Error("expected ValidateState to return true for a valid state")
	}
}

func TestSessionStore_ValidateState_Invalid(t *testing.T) {
	store := NewSessionStore()

	if store.ValidateState("unknown_state") {
		t.Error("expected ValidateState to return false for unknown state")
	}
}

func TestSessionStore_ValidateState_Expired(t *testing.T) {
	store := NewSessionStore()
	store.StoreState("expiring_state")

	// Directly set the expiry to the past.
	store.mu.Lock()
	store.states["expiring_state"] = time.Now().Add(-1 * time.Hour)
	store.mu.Unlock()

	if store.ValidateState("expiring_state") {
		t.Error("expected ValidateState to return false for expired state")
	}
}

func TestSessionStore_ValidateState_OneTimeUse(t *testing.T) {
	store := NewSessionStore()
	store.StoreState("onetime_state")

	if !store.ValidateState("onetime_state") {
		t.Fatal("first ValidateState call should return true")
	}
	if store.ValidateState("onetime_state") {
		t.Error("second ValidateState call should return false (one-time use)")
	}
}

// ---------------------------------------------------------------------------
// session.go — Cookie / HTTP helper tests
// ---------------------------------------------------------------------------

func TestGetSessionFromRequest(t *testing.T) {
	store := NewSessionStore()
	token := &oauth2.Token{AccessToken: "cookie_tok"}

	id, err := store.CreateSession(token)
	if err != nil {
		t.Fatalf("CreateSession error: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: SessionCookieName, Value: id})

	sess := store.GetSessionFromRequest(req)
	if sess == nil {
		t.Fatal("expected session from request cookie")
	}
	if sess.GetToken().AccessToken != "cookie_tok" {
		t.Errorf("expected AccessToken %q, got %q", "cookie_tok", sess.GetToken().AccessToken)
	}
}

func TestGetSessionFromRequest_NoCookie(t *testing.T) {
	store := NewSessionStore()

	req := httptest.NewRequest(http.MethodGet, "/", nil)

	sess := store.GetSessionFromRequest(req)
	if sess != nil {
		t.Error("expected nil when no session cookie is present")
	}
}

func TestSetSessionCookie(t *testing.T) {
	rr := httptest.NewRecorder()
	SetSessionCookie(rr, "sess_id_123", false)

	resp := rr.Result()
	defer resp.Body.Close()

	cookies := resp.Cookies()
	if len(cookies) == 0 {
		t.Fatal("expected at least one Set-Cookie header")
	}

	var found *http.Cookie
	for _, c := range cookies {
		if c.Name == SessionCookieName {
			found = c
			break
		}
	}
	if found == nil {
		t.Fatalf("cookie %q not found in response", SessionCookieName)
	}
	if found.Value != "sess_id_123" {
		t.Errorf("expected cookie value %q, got %q", "sess_id_123", found.Value)
	}
	if !found.HttpOnly {
		t.Error("expected HttpOnly to be true")
	}
	if found.SameSite != http.SameSiteStrictMode {
		t.Errorf("expected SameSite=Strict, got %v", found.SameSite)
	}
	if found.Path != "/" {
		t.Errorf("expected Path %q, got %q", "/", found.Path)
	}
	if found.Secure {
		t.Error("expected Secure=false when isProd is false")
	}
}

func TestSetStateCookie(t *testing.T) {
	rr := httptest.NewRecorder()
	SetStateCookie(rr, "state_xyz", true)

	resp := rr.Result()
	defer resp.Body.Close()

	cookies := resp.Cookies()
	var found *http.Cookie
	for _, c := range cookies {
		if c.Name == StateCookieName {
			found = c
			break
		}
	}
	if found == nil {
		t.Fatalf("cookie %q not found in response", StateCookieName)
	}
	if found.Value != "state_xyz" {
		t.Errorf("expected cookie value %q, got %q", "state_xyz", found.Value)
	}
	if !found.HttpOnly {
		t.Error("expected HttpOnly to be true")
	}
	if found.SameSite != http.SameSiteLaxMode {
		t.Errorf("expected SameSite=Lax, got %v", found.SameSite)
	}
	if found.Path != "/auth/callback" {
		t.Errorf("expected Path %q, got %q", "/auth/callback", found.Path)
	}
	if !found.Secure {
		t.Error("expected Secure=true when isProd is true")
	}
}

func TestGetStateFromCookie(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/auth/callback", nil)
	req.AddCookie(&http.Cookie{Name: StateCookieName, Value: "returned_state"})

	got := GetStateFromCookie(req)
	if got != "returned_state" {
		t.Errorf("expected state %q, got %q", "returned_state", got)
	}
}

func TestGetStateFromCookie_Missing(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/auth/callback", nil)

	got := GetStateFromCookie(req)
	if got != "" {
		t.Errorf("expected empty string when cookie is missing, got %q", got)
	}
}

// ---------------------------------------------------------------------------
// oauth.go tests
// ---------------------------------------------------------------------------

func TestNewOAuthConfig(t *testing.T) {
	cfg := NewOAuthConfig("client-id", "client-secret", "https://example.com/callback")

	if cfg.ClientID != "client-id" {
		t.Errorf("expected ClientID %q, got %q", "client-id", cfg.ClientID)
	}
	if cfg.ClientSecret != "client-secret" {
		t.Errorf("expected ClientSecret %q, got %q", "client-secret", cfg.ClientSecret)
	}
	if cfg.RedirectURL != "https://example.com/callback" {
		t.Errorf("expected RedirectURL %q, got %q", "https://example.com/callback", cfg.RedirectURL)
	}
	if len(cfg.Scopes) != 1 || cfg.Scopes[0] != "https://www.googleapis.com/auth/youtube.readonly" {
		t.Errorf("unexpected Scopes: %v", cfg.Scopes)
	}
	if cfg.Endpoint != google.Endpoint {
		t.Error("expected Endpoint to be google.Endpoint")
	}
}

func TestGenerateState(t *testing.T) {
	state, err := GenerateState()
	if err != nil {
		t.Fatalf("GenerateState returned error: %v", err)
	}
	if len(state) != 32 {
		t.Errorf("expected state of 32 hex chars (16 bytes), got length %d", len(state))
	}

	// Verify all characters are valid hex.
	for _, c := range state {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
			t.Errorf("state contains non-hex character: %c", c)
			break
		}
	}
}

func TestGenerateState_Unique(t *testing.T) {
	state1, err := GenerateState()
	if err != nil {
		t.Fatalf("first GenerateState error: %v", err)
	}
	state2, err := GenerateState()
	if err != nil {
		t.Fatalf("second GenerateState error: %v", err)
	}
	if state1 == state2 {
		t.Error("two consecutive GenerateState calls should produce different values")
	}
}

// ---------------------------------------------------------------------------
// Concurrency safety tests (race detector validation)
// ---------------------------------------------------------------------------

func TestSessionStore_ConcurrentCreateAndGet(t *testing.T) {
	store := NewSessionStore()
	const goroutines = 50

	var wg sync.WaitGroup
	ids := make([]string, goroutines)
	errs := make([]error, goroutines)

	// Concurrently create sessions.
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			token := &oauth2.Token{AccessToken: fmt.Sprintf("tok_%d", idx)}
			id, err := store.CreateSession(token)
			ids[idx] = id
			errs[idx] = err
		}(i)
	}
	wg.Wait()

	// Verify all sessions were created successfully.
	for i, err := range errs {
		if err != nil {
			t.Fatalf("goroutine %d: CreateSession error: %v", i, err)
		}
	}

	// Concurrently read sessions.
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			sess := store.GetSession(ids[idx])
			if sess == nil {
				t.Errorf("goroutine %d: GetSession returned nil for valid ID", idx)
				return
			}
			expected := fmt.Sprintf("tok_%d", idx)
			if sess.GetToken().AccessToken != expected {
				t.Errorf("goroutine %d: expected %q, got %q", idx, expected, sess.GetToken().AccessToken)
			}
		}(i)
	}
	wg.Wait()
}

func TestSessionStore_ConcurrentStateValidation(t *testing.T) {
	store := NewSessionStore()
	const goroutines = 50

	// Store unique states.
	states := make([]string, goroutines)
	for i := 0; i < goroutines; i++ {
		states[i] = fmt.Sprintf("state_%d", i)
		store.StoreState(states[i])
	}

	// Concurrently validate states (each should succeed exactly once).
	results := make([]bool, goroutines)
	var wg sync.WaitGroup
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			results[idx] = store.ValidateState(states[idx])
		}(i)
	}
	wg.Wait()

	for i, ok := range results {
		if !ok {
			t.Errorf("goroutine %d: ValidateState returned false for valid state", i)
		}
	}

	// Verify one-time use: second attempt should all fail.
	for i := 0; i < goroutines; i++ {
		if store.ValidateState(states[i]) {
			t.Errorf("state %d: ValidateState succeeded on second call (should be one-time)", i)
		}
	}
}

func TestSessionStore_ConcurrentMixedOperations(t *testing.T) {
	store := NewSessionStore()
	const goroutines = 30

	var wg sync.WaitGroup

	// Mix of CreateSession, GetSession, StoreState, and ValidateState.
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			token := &oauth2.Token{AccessToken: fmt.Sprintf("mixed_%d", idx)}
			id, err := store.CreateSession(token)
			if err != nil {
				t.Errorf("goroutine %d: CreateSession error: %v", idx, err)
				return
			}
			_ = store.GetSession(id)
			_ = store.GetSession("nonexistent")

			state := fmt.Sprintf("mixed_state_%d", idx)
			store.StoreState(state)
			_ = store.ValidateState(state)
		}(i)
	}
	wg.Wait()
}
