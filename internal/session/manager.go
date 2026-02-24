package session

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"
)

// Manager orchestrates the session lifecycle.
// Lock ordering: Manager.mu is Level 1 (highest priority).
type Manager struct {
	mu    sync.Mutex
	state State

	sessionID string
	createdAt time.Time
}

// NewManager creates a new session manager.
func NewManager(sessionID string) *Manager {
	return &Manager{
		sessionID: sessionID,
		state:     StateOnboarding,
		createdAt: time.Now(),
	}
}

// State returns the current session state.
func (m *Manager) State() State {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.state
}

// TransitionTo transitions to a new state with validation.
func (m *Manager) TransitionTo(target State) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.state.CanTransitionTo(target) {
		return fmt.Errorf("invalid state transition: %s → %s", m.state, target)
	}

	old := m.state
	m.state = target
	slog.Info("state_transition",
		"session", m.sessionID,
		"from", string(old),
		"to", string(target),
	)
	return nil
}

// Shutdown closes all resources associated with this session.
func (m *Manager) Shutdown(ctx context.Context) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.state = StateEnded
	slog.Info("session_shutdown", "session", m.sessionID)
}
