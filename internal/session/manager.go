package session

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"google.golang.org/genai"
)

// Manager orchestrates the session lifecycle.
// Lock ordering: Manager.mu is Level 1 (highest priority).
type Manager struct {
	mu    sync.Mutex
	state State

	sessionID    string
	personaName  string
	matchedVoice string
	languageCode string
	createdAt    time.Time
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

// SessionID returns the session identifier.
func (m *Manager) SessionID() string {
	return m.sessionID
}

// SetPersona stores the matched persona details on the manager.
func (m *Manager) SetPersona(name, voice, langCode string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.personaName = name
	m.matchedVoice = voice
	m.languageCode = langCode
	slog.Info("persona_set", "session", m.sessionID, "name", name, "voice", voice)
}

// PersonaName returns the stored persona name.
func (m *Manager) PersonaName() string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.personaName
}

// MatchedVoice returns the stored matched voice.
func (m *Manager) MatchedVoice() string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.matchedVoice
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

// BuildOnboardingConfig creates the LiveConnectConfig for the onboarding phase.
// System voice: Aoede (missless host), warm guide in Korean.
func (m *Manager) BuildOnboardingConfig() *genai.LiveConnectConfig {
	return &genai.LiveConnectConfig{
		ResponseModalities: []genai.Modality{genai.ModalityAudio, genai.ModalityText},
		SpeechConfig: &genai.SpeechConfig{
			VoiceConfig: &genai.VoiceConfig{
				PrebuiltVoiceConfig: &genai.PrebuiltVoiceConfig{
					VoiceName: "Aoede",
				},
			},
		},
		SystemInstruction: &genai.Content{
			Parts: []*genai.Part{
				genai.NewPartFromText(`You are a warm, empathetic AI guide for missless - a virtual reunion experience.
You help users reconnect with people they miss through AI-powered conversations.

During onboarding:
1. Greet the user warmly: "안녕하세요, missless에 오신 걸 환영해요"
2. Ask who they'd like to reconnect with (name and relationship)
3. Guide them to select YouTube videos of that person
4. Share progress while analyzing: "영상을 분석하고 있어요, 잠시만 기다려주세요"
5. Confirm persona creation when ready

Be gentle, understanding, and supportive. This is an emotional experience.
Speak naturally in Korean unless the user prefers another language.
Keep responses concise for voice — avoid long monologues.`),
			},
		},
		Tools: []*genai.Tool{
			{
				FunctionDeclarations: []*genai.FunctionDeclaration{
					{
						Name:        "generate_scene",
						Description: "Generate a high-quality background scene image with 2-stage progressive rendering",
						Parameters: &genai.Schema{
							Type: genai.TypeObject,
							Properties: map[string]*genai.Schema{
								"prompt":     {Type: genai.TypeString, Description: "Scene description prompt"},
								"mood":       {Type: genai.TypeString, Description: "Emotional mood (warm, nostalgic, joyful, etc.)"},
								"characters": {Type: genai.TypeString, Description: "Character descriptions for the scene"},
							},
							Required: []string{"prompt", "mood"},
						},
					},
					{
						Name:        "generate_fast_scene",
						Description: "Generate a quick preview-only scene image for rapid visual feedback",
						Parameters: &genai.Schema{
							Type: genai.TypeObject,
							Properties: map[string]*genai.Schema{
								"prompt": {Type: genai.TypeString, Description: "Scene description prompt"},
								"mood":   {Type: genai.TypeString, Description: "Emotional mood (warm, nostalgic, joyful, etc.)"},
							},
							Required: []string{"prompt", "mood"},
						},
					},
					{
						Name:        "change_atmosphere",
						Description: "Change the background music to match the conversation mood",
						Parameters: &genai.Schema{
							Type: genai.TypeObject,
							Properties: map[string]*genai.Schema{
								"mood": {Type: genai.TypeString, Description: "Target mood for BGM (warm, nostalgic, joyful, bittersweet)"},
							},
							Required: []string{"mood"},
						},
					},
					{
						Name:        "recall_memory",
						Description: "Search shared memories relevant to the current conversation topic",
						Parameters: &genai.Schema{
							Type: genai.TypeObject,
							Properties: map[string]*genai.Schema{
								"query": {Type: genai.TypeString, Description: "Search query for relevant memories"},
							},
							Required: []string{"query"},
						},
					},
					{
						Name:        "analyze_user",
						Description: "Analyze the user's emotional state and engagement level",
						Parameters: &genai.Schema{
							Type: genai.TypeObject,
							Properties: map[string]*genai.Schema{
								"aspect": {Type: genai.TypeString, Description: "Aspect to analyze (emotion, engagement, comfort)"},
							},
							Required: []string{"aspect"},
						},
					},
					{
						Name:        "end_reunion",
						Description: "End the reunion session and generate a memory album",
						Parameters: &genai.Schema{
							Type: genai.TypeObject,
							Properties: map[string]*genai.Schema{
								"reason": {Type: genai.TypeString, Description: "Reason for ending (user_request, natural_end, timeout)"},
							},
							Required: []string{"reason"},
						},
					},
				},
			},
		},
		SessionResumption: &genai.SessionResumptionConfig{
			Transparent: true,
		},
	}
}

// Shutdown closes all resources associated with this session.
func (m *Manager) Shutdown(ctx context.Context) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.state = StateEnded
	slog.Info("session_shutdown", "session", m.sessionID)
}
