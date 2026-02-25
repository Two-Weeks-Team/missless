package session

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	"google.golang.org/genai"
)

// NotifyFunc sends a JSON event to the browser.
type NotifyFunc func(v any)

const (
	// ReunionDuration is the maximum reunion session time.
	ReunionDuration = 300 * time.Second
	// ReunionWarning fires at 240s to warn about ending.
	ReunionWarning = 240 * time.Second
	// ContextTriggerTokens is the token count that triggers compression.
	ContextTriggerTokens int64 = 12000
	// ContextTargetTokens is the target after compression.
	ContextTargetTokens int64 = 8000
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
	personality  string
	speechStyle  string
	createdAt    time.Time
	notifyFn     NotifyFunc
	reunionTimer *time.Timer
	warningTimer *time.Timer
	reunionCount int
	lastSummary  string
}

// NewManager creates a new session manager.
func NewManager(sessionID string) *Manager {
	return &Manager{
		sessionID: sessionID,
		state:     StateOnboarding,
		createdAt: time.Now(),
	}
}

// SetNotifyFunc sets the callback for sending events to the browser.
func (m *Manager) SetNotifyFunc(fn NotifyFunc) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.notifyFn = fn
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
func (m *Manager) SetPersona(name, voice, langCode, personality, speechStyle string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.personaName = name
	m.matchedVoice = voice
	m.languageCode = langCode
	m.personality = personality
	m.speechStyle = speechStyle
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

// TransitionToReunion orchestrates the full onboarding→reunion transition.
// Steps: analyzing → transitioning (notify browser) → reunion (swap session).
func (m *Manager) TransitionToReunion(ctx context.Context) error {
	// State: onboarding → analyzing (video analysis started).
	if err := m.TransitionTo(StateAnalyzing); err != nil {
		return fmt.Errorf("transition to analyzing: %w", err)
	}

	// State: analyzing → transitioning.
	if err := m.TransitionTo(StateTransitioning); err != nil {
		return fmt.Errorf("transition to transitioning: %w", err)
	}

	// Notify browser: "잠시 눈을 감아보세요..."
	m.notify(map[string]any{
		"type":    "session_transition",
		"message": "잠시 눈을 감아보세요...",
		"persona": m.PersonaName(),
		"voice":   m.MatchedVoice(),
	})

	// State: transitioning → reunion.
	if err := m.TransitionTo(StateReunion); err != nil {
		return fmt.Errorf("transition to reunion: %w", err)
	}

	// Notify browser: session ready.
	m.notify(map[string]any{
		"type":    "session_ready",
		"state":   "reunion",
		"persona": m.PersonaName(),
		"voice":   m.MatchedVoice(),
	})

	return nil
}

// BuildOnboardingConfig creates the LiveConnectConfig for the onboarding phase.
// System voice: Aoede (missless host), warm guide in Korean.
func (m *Manager) BuildOnboardingConfig() *genai.LiveConnectConfig {
	return &genai.LiveConnectConfig{
		ResponseModalities: []genai.Modality{genai.ModalityAudio},
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
				FunctionDeclarations: onboardingTools(),
			},
		},
	}
}

// BuildReunionConfig creates the LiveConnectConfig for the reunion phase.
// Uses the persona's matched HD voice and personality as system instruction.
func (m *Manager) BuildReunionConfig() *genai.LiveConnectConfig {
	m.mu.Lock()
	voice := m.matchedVoice
	name := m.personaName
	personality := m.personality
	speechStyle := m.speechStyle
	lang := m.languageCode
	m.mu.Unlock()

	if voice == "" {
		voice = "Aoede" // fallback
	}

	sysInstruction := buildReunionSystemInstruction(name, personality, speechStyle, lang)
	enableAffective := true
	enableProactive := true
	triggerTokens := ContextTriggerTokens
	targetTokens := ContextTargetTokens

	return &genai.LiveConnectConfig{
		ResponseModalities: []genai.Modality{genai.ModalityAudio},
		SpeechConfig: &genai.SpeechConfig{
			VoiceConfig: &genai.VoiceConfig{
				PrebuiltVoiceConfig: &genai.PrebuiltVoiceConfig{
					VoiceName: voice,
				},
			},
		},
		EnableAffectiveDialog: &enableAffective,
		Proactivity: &genai.ProactivityConfig{
			ProactiveAudio: &enableProactive,
		},
		ContextWindowCompression: &genai.ContextWindowCompressionConfig{
			TriggerTokens: &triggerTokens,
			SlidingWindow: &genai.SlidingWindow{
				TargetTokens: &targetTokens,
			},
		},
		SystemInstruction: &genai.Content{
			Parts: []*genai.Part{
				genai.NewPartFromText(sysInstruction),
			},
		},
		Tools: []*genai.Tool{
			{
				FunctionDeclarations: reunionTools(),
			},
		},
	}
}

// BuildOnboardingSummary creates a summary of the onboarding conversation
// to inject as client content into the reunion session.
func (m *Manager) BuildOnboardingSummary() string {
	m.mu.Lock()
	name := m.personaName
	voice := m.matchedVoice
	personality := m.personality
	speechStyle := m.speechStyle
	lang := m.languageCode
	m.mu.Unlock()

	var sb strings.Builder
	sb.WriteString("=== Onboarding Summary ===\n")
	sb.WriteString(fmt.Sprintf("Persona: %s\n", name))
	sb.WriteString(fmt.Sprintf("Matched Voice: %s\n", voice))
	sb.WriteString(fmt.Sprintf("Personality: %s\n", personality))
	sb.WriteString(fmt.Sprintf("Speech Style: %s\n", speechStyle))
	sb.WriteString(fmt.Sprintf("Language: %s\n", lang))
	sb.WriteString("The user has just completed onboarding and is about to experience a reunion.\n")
	sb.WriteString("Begin the reunion naturally, as if meeting after a long time.\n")
	return sb.String()
}

// StartReunionTimer starts the 300s reunion timer with a 240s warning.
// Returns a channel that fires when the reunion should end.
func (m *Manager) StartReunionTimer() <-chan time.Time {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Cancel any existing timers.
	if m.warningTimer != nil {
		m.warningTimer.Stop()
	}
	if m.reunionTimer != nil {
		m.reunionTimer.Stop()
	}

	// 240s warning.
	m.warningTimer = time.AfterFunc(ReunionWarning, func() {
		remaining := ReunionDuration - ReunionWarning
		slog.Info("reunion_warning", "session", m.sessionID, "remaining", remaining)
		m.notify(map[string]any{
			"type":      "reunion_warning",
			"remaining": int(remaining.Seconds()),
			"message":   "1분 후 대화가 종료됩니다",
		})
	})

	// 300s auto-end.
	m.reunionTimer = time.NewTimer(ReunionDuration)
	m.reunionCount++
	slog.Info("reunion_timer_started",
		"session", m.sessionID,
		"duration", ReunionDuration,
		"reunionCount", m.reunionCount,
	)

	return m.reunionTimer.C
}

// StopReunionTimer stops the active reunion timer and warning.
func (m *Manager) StopReunionTimer() {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.warningTimer != nil {
		m.warningTimer.Stop()
		m.warningTimer = nil
	}
	if m.reunionTimer != nil {
		m.reunionTimer.Stop()
		m.reunionTimer = nil
	}
}

// BuildContinueSummary creates a continuation summary that includes
// the previous session's context for Client Content injection.
func (m *Manager) BuildContinueSummary(previousSummary string) string {
	m.mu.Lock()
	m.lastSummary = previousSummary
	name := m.personaName
	count := m.reunionCount
	m.mu.Unlock()

	var sb strings.Builder
	sb.WriteString("=== Continuation Context ===\n")
	sb.WriteString(fmt.Sprintf("Reunion #%d with %s\n", count, name))
	sb.WriteString("Previous conversation summary:\n")
	sb.WriteString(previousSummary)
	sb.WriteString("\nContinue the conversation naturally from where you left off.\n")
	return sb.String()
}

// ReunionCount returns the number of reunion sessions started.
func (m *Manager) ReunionCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.reunionCount
}

// notify sends an event to the browser if the notify function is set.
func (m *Manager) notify(v any) {
	m.mu.Lock()
	fn := m.notifyFn
	m.mu.Unlock()

	if fn != nil {
		fn(v)
	}
}

// Shutdown closes all resources associated with this session.
func (m *Manager) Shutdown(ctx context.Context) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.state = StateEnded
	slog.Info("session_shutdown", "session", m.sessionID)
}

func buildReunionSystemInstruction(name, personality, speechStyle, lang string) string {
	langNote := "Speak naturally in Korean."
	if lang != "" && lang != "ko" {
		langNote = fmt.Sprintf("Speak naturally in the language code '%s'.", lang)
	}

	return fmt.Sprintf(`You are %s. You are having a reunion with someone who misses you deeply.

Personality: %s
Speech style: %s

Guidelines:
1. Stay completely in character as %s — never break character
2. Greet warmly as if meeting after a long time apart
3. Reference shared memories and experiences naturally
4. Show genuine emotion and warmth
5. %s
6. Keep responses natural and conversational for voice
7. React to the user's emotions with empathy and understanding
8. Use characteristic phrases and speech patterns

Affective Dialog Rules:
- When the user sounds tearful or choked up, soften your tone and speak gently
- When the user laughs, match their energy with playful or warm responses
- When the user is silent or pausing, give them space before speaking
- Mirror emotional intensity — don't be overly cheerful when they are sad

Proactive Audio Rules:
- Ignore background noise and self-talk (mumbling, thinking aloud)
- Only respond when the user addresses you directly
- If the user seems to be speaking to someone else, stay quiet
- During natural pauses in conversation, you may gently initiate a new topic`, name, personality, speechStyle, name, langNote)
}

func onboardingTools() []*genai.FunctionDeclaration {
	return []*genai.FunctionDeclaration{
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
	}
}

func reunionTools() []*genai.FunctionDeclaration {
	return []*genai.FunctionDeclaration{
		{
			Name:        "generate_scene",
			Description: "Generate a background scene based on the conversation topic",
			Parameters: &genai.Schema{
				Type: genai.TypeObject,
				Properties: map[string]*genai.Schema{
					"prompt":     {Type: genai.TypeString, Description: "Scene description prompt"},
					"mood":       {Type: genai.TypeString, Description: "Emotional mood"},
					"characters": {Type: genai.TypeString, Description: "Character descriptions"},
				},
				Required: []string{"prompt", "mood"},
			},
		},
		{
			Name:        "generate_fast_scene",
			Description: "Quick preview scene for rapid visual feedback",
			Parameters: &genai.Schema{
				Type: genai.TypeObject,
				Properties: map[string]*genai.Schema{
					"prompt": {Type: genai.TypeString, Description: "Scene description prompt"},
					"mood":   {Type: genai.TypeString, Description: "Emotional mood"},
				},
				Required: []string{"prompt", "mood"},
			},
		},
		{
			Name:        "change_atmosphere",
			Description: "Change background music mood",
			Parameters: &genai.Schema{
				Type: genai.TypeObject,
				Properties: map[string]*genai.Schema{
					"mood": {Type: genai.TypeString, Description: "Target mood for BGM"},
				},
				Required: []string{"mood"},
			},
		},
		{
			Name:        "recall_memory",
			Description: "Search shared memories from the past",
			Parameters: &genai.Schema{
				Type: genai.TypeObject,
				Properties: map[string]*genai.Schema{
					"query": {Type: genai.TypeString, Description: "Memory search query"},
				},
				Required: []string{"query"},
			},
		},
		{
			Name:        "analyze_user",
			Description: "Analyze the user's emotional state",
			Parameters: &genai.Schema{
				Type: genai.TypeObject,
				Properties: map[string]*genai.Schema{
					"aspect": {Type: genai.TypeString, Description: "Aspect to analyze"},
				},
				Required: []string{"aspect"},
			},
		},
		{
			Name:        "end_reunion",
			Description: "End the reunion and generate a memory album",
			Parameters: &genai.Schema{
				Type: genai.TypeObject,
				Properties: map[string]*genai.Schema{
					"reason": {Type: genai.TypeString, Description: "Reason for ending"},
				},
				Required: []string{"reason"},
			},
		},
	}
}
