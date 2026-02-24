package onboarding

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"github.com/Two-Weeks-Team/missless/internal/util"
)

// maxConcurrentAnalyses limits parallel video analysis to avoid rate limits.
const maxConcurrentAnalyses = 2

// Persona represents a generated persona profile.
type Persona struct {
	Name              string   `json:"name"`
	Personality       string   `json:"personality"`
	SpeechStyle       string   `json:"speechStyle"`
	MatchedVoice      string   `json:"matchedVoice"`
	FrequentPhrases   []string `json:"frequentPhrases"`
	EmotionalPatterns string   `json:"emotionalPatterns"`
	LanguageCode      string   `json:"languageCode"`
}

// Pipeline runs the Sequential Agent: VideoAnalyzer → VoiceMatcher.
type Pipeline struct {
	analyzer     *Analyzer
	voiceMatcher *VoiceMatcher
}

// NewPipeline creates a new onboarding pipeline.
func NewPipeline(analyzer *Analyzer, voiceMatcher *VoiceMatcher) *Pipeline {
	return &Pipeline{
		analyzer:     analyzer,
		voiceMatcher: voiceMatcher,
	}
}

// Run executes the full onboarding pipeline.
// Stage 1: Analyze YouTube videos → extract personality
// Stage 2: Match personality to optimal HD preset voice
func (p *Pipeline) Run(ctx context.Context, videoURLs []string, targetPerson string, progressFn func(step string, percent int)) (*Persona, error) {
	slog.Info("pipeline_start", "videos", len(videoURLs), "target", targetPerson)

	// Stage 1: Video Analysis
	progressFn("Starting video analysis", 0)
	analyses, err := p.analyzeVideos(ctx, videoURLs, targetPerson, progressFn)
	if err != nil {
		return nil, fmt.Errorf("video analysis failed: %w", err)
	}

	if len(analyses) == 0 {
		return nil, fmt.Errorf("no successful video analyses")
	}

	// Stage 2: Voice Matching
	progressFn("Matching voice profile", 60)
	persona, err := p.voiceMatcher.MatchFromAnalyses(ctx, analyses, targetPerson)
	if err != nil {
		return nil, fmt.Errorf("voice matching failed: %w", err)
	}

	progressFn("Persona ready", 100)
	slog.Info("pipeline_complete",
		"persona", persona.Name,
		"voice", persona.MatchedVoice,
	)

	return persona, nil
}

// analyzeVideos runs video analysis in parallel with a semaphore.
func (p *Pipeline) analyzeVideos(ctx context.Context, urls []string, target string, progressFn func(string, int)) ([]*VideoAnalysis, error) {
	sem := make(chan struct{}, maxConcurrentAnalyses)
	var mu sync.Mutex
	var analyses []*VideoAnalysis
	var firstErr error
	var wg sync.WaitGroup

	for i, url := range urls {
		wg.Add(1)
		idx, videoURL := i, url

		util.SafeGo(func() {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			analysis, err := p.analyzer.AnalyzeYouTubeURL(ctx, videoURL, target, func(step string, pct int) {
				// Scale progress within video analysis phase (0-50%).
				basePct := 50 * idx / len(urls)
				scaled := basePct + (pct * 50 / (100 * len(urls)))
				progressFn(fmt.Sprintf("[Video %d/%d] %s", idx+1, len(urls), step), scaled)
			})

			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				slog.Warn("video_analysis_failed", "url", videoURL, "error", err)
				if firstErr == nil {
					firstErr = err
				}
			} else {
				analyses = append(analyses, analysis)
			}
		})
	}

	wg.Wait()

	if len(analyses) == 0 && firstErr != nil {
		return nil, firstErr
	}

	return analyses, nil
}
