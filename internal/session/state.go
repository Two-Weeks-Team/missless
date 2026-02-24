package session

// State represents the session state machine.
type State string

const (
	StateOnboarding    State = "onboarding"
	StateAnalyzing     State = "analyzing"
	StateTransitioning State = "transitioning"
	StateReunion       State = "reunion"
	StateAlbum         State = "album"
	StateEnded         State = "ended"
)

// ValidTransitions defines allowed state transitions.
var ValidTransitions = map[State][]State{
	StateOnboarding:    {StateAnalyzing},
	StateAnalyzing:     {StateTransitioning},
	StateTransitioning: {StateReunion},
	StateReunion:       {StateAlbum},
	StateAlbum:         {StateEnded},
}

// CanTransitionTo checks if the transition is valid.
func (s State) CanTransitionTo(target State) bool {
	allowed, ok := ValidTransitions[s]
	if !ok {
		return false
	}
	for _, a := range allowed {
		if a == target {
			return true
		}
	}
	return false
}
