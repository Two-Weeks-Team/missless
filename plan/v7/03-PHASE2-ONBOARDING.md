# Phase 2: 온보딩 파이프라인 (T07~T12)

> D-14 ~ D-10 | 6일 | 핵심 마일스톤: YouTube URL 분석 → 페르소나 생성 → 세션 전환 (음성 변경)

---

## T07. Google OAuth + YouTube 영상 목록 + 공개상태

**일수**: 1 | **난이도**: 중 | **의존성**: T01

### auth/oauth.go — Go 백엔드 OAuth 2.0

```go
package auth

import (
    "context"
    "encoding/json"
    "net/http"

    "golang.org/x/oauth2"
    "golang.org/x/oauth2/google"
    "google.golang.org/api/youtube/v3"
)

type OAuthHandler struct {
    config *oauth2.Config
    store  SessionStore
}

func NewOAuthHandler(clientID, clientSecret, redirectURL string, store SessionStore) *OAuthHandler {
    return &OAuthHandler{
        config: &oauth2.Config{
            ClientID:     clientID,
            ClientSecret: clientSecret,
            RedirectURL:  redirectURL,
            Scopes:       []string{youtube.YoutubeReadonlyScope},
            Endpoint:     google.Endpoint,
        },
        store: store,
    }
}

func (h *OAuthHandler) HandleLogin(w http.ResponseWriter, r *http.Request) {
    state := generateState()
    h.store.SetState(r.Context(), state)
    url := h.config.AuthCodeURL(state, oauth2.AccessTypeOffline)
    http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func (h *OAuthHandler) HandleCallback(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()

    if !h.store.ValidateState(ctx, r.URL.Query().Get("state")) {
        http.Error(w, "invalid state", http.StatusBadRequest)
        return
    }

    token, err := h.config.Exchange(ctx, r.URL.Query().Get("code"))
    if err != nil {
        http.Error(w, "exchange failed", http.StatusInternalServerError)
        return
    }

    // 세션에 OAuth 토큰 저장
    sessionID := h.store.CreateSession(ctx, token)
    http.SetCookie(w, &http.Cookie{
        Name: "session_id", Value: sessionID,
        HttpOnly: true, Secure: true, SameSite: http.SameSiteStrictMode,
    })
    http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}
```

### media/youtube.go — YouTube 영상 목록 + 공개상태

```go
package media

import (
    "context"
    "log/slog"

    "golang.org/x/oauth2"
    "google.golang.org/api/option"
    "google.golang.org/api/youtube/v3"
)

type YouTubeVideo struct {
    ID            string `json:"id"`
    Title         string `json:"title"`
    ThumbnailURL  string `json:"thumbnailUrl"`
    Duration      string `json:"duration"`
    PrivacyStatus string `json:"privacyStatus"` // "public", "unlisted", "private"
    Analyzable    bool   `json:"analyzable"`     // public/unlisted = true
}

type YouTubeClient struct{}

// ListUserVideos는 사용자의 YouTube 업로드 영상 목록을 반환한다.
// privacyStatus를 포함하여 공개/비공개 상태를 분류한다 (일부공개는 분석 불가).
func (c *YouTubeClient) ListUserVideos(ctx context.Context, token *oauth2.Token) ([]YouTubeVideo, error) {
    svc, err := youtube.NewService(ctx, option.WithTokenSource(
        oauth2.StaticTokenSource(token),
    ))
    if err != nil {
        return nil, err
    }

    // 1) 채널의 업로드 플레이리스트 ID 조회
    chResp, err := svc.Channels.List([]string{"contentDetails"}).Mine(true).Do()
    if err != nil {
        return nil, err
    }
    if len(chResp.Items) == 0 {
        return nil, nil
    }
    uploadsID := chResp.Items[0].ContentDetails.RelatedPlaylists.Uploads

    // 2) 플레이리스트 아이템 조회
    plResp, err := svc.PlaylistItems.List([]string{"snippet", "status"}).
        PlaylistId(uploadsID).MaxResults(50).Do()
    if err != nil {
        return nil, err
    }

    videoIDs := make([]string, 0, len(plResp.Items))
    for _, item := range plResp.Items {
        videoIDs = append(videoIDs, item.Snippet.ResourceId.VideoId)
    }

    // 3) 각 영상의 상세 정보 (duration, privacyStatus) 조회
    vResp, err := svc.Videos.List([]string{"snippet", "contentDetails", "status"}).
        Id(videoIDs...).Do()
    if err != nil {
        return nil, err
    }

    videos := make([]YouTubeVideo, 0, len(vResp.Items))
    for _, v := range vResp.Items {
        privacy := v.Status.PrivacyStatus
        videos = append(videos, YouTubeVideo{
            ID:            v.Id,
            Title:         v.Snippet.Title,
            ThumbnailURL:  v.Snippet.Thumbnails.Medium.Url,
            Duration:      v.ContentDetails.Duration,
            PrivacyStatus: privacy,
            Analyzable:    privacy == "public",
        })
    }

    slog.Info("youtube_videos_listed", "count", len(videos))
    return videos, nil
}
```

### 체크포인트

- [ ] Google 로그인 → OAuth 토큰 획득 성공
- [ ] YouTube Data API 영상 목록 조회 성공
- [ ] 영상별 privacyStatus (public✅/private🔒) 분류
- [ ] 브라우저에 `youtube_videos` 이벤트 전송 → 그리드 표시

---

## T08. YouTube URL 직접 분석 + 갤러리 Fallback

**일수**: 1.5 | **난이도**: 상 | **의존성**: T01, T07

### onboarding/analyzer.go — 영상 분석 (제로 다운로드)

```go
package onboarding

import (
    "context"
    "encoding/json"
    "fmt"
    "log/slog"
    "time"

    "github.com/Two-Weeks-Team/missless/internal/retry"
    "google.golang.org/genai"
)

type VideoAnalysis struct {
    SpeechPatterns    []string    `json:"speechPatterns"`
    Expressions       []string    `json:"expressions"`
    PersonalityTraits []string    `json:"personalityTraits"`
    Highlights        []Highlight `json:"highlights"`
    VoiceCharacteristics string   `json:"voiceCharacteristics"`
}

type Highlight struct {
    Timestamp   string `json:"timestamp"`
    Description string `json:"description"`
    Expression  string `json:"expression"`
}

type Analyzer struct {
    client *genai.Client
}

// AnalyzeYouTubeURL은 YouTube URL을 Gemini API에 직접 전달하여 분석한다.
// 다운로드/임시저장 없이 URL만으로 분석 완료 (제로 다운로드).
// V7 모델: gemini-2.5-pro
func (a *Analyzer) AnalyzeYouTubeURL(ctx context.Context, videoURL, targetPerson string, progressFn func(step string, percent int)) (*VideoAnalysis, error) {
    slog.Info("youtube_url_analysis_start", "url", videoURL, "target", targetPerson)

    progressFn("영상 분석 시작", 10)

    var result *VideoAnalysis
    err := retry.WithBackoff(ctx, 3, func() error {
        resp, err := a.client.Models.GenerateContent(ctx,
            "gemini-2.5-pro",
            genai.FileData{
                FileURI:  videoURL,
                MIMEType: "video/mp4",
            },
            genai.Text(buildAnalysisPrompt(targetPerson)),
            &genai.GenerateContentConfig{
                Temperature:      genai.Ptr(float32(0.2)),
                ResponseMIMEType: "application/json",
                ResponseSchema:   videoAnalysisSchema(),
            },
        )
        if err != nil {
            return err
        }
        result = parseAnalysis(resp)
        return nil
    })

    if err != nil {
        return nil, fmt.Errorf("youtube analysis failed: %w", err)
    }

    progressFn("영상 분석 완료", 50)
    return result, nil
}

// AnalyzeUploadedVideo는 갤러리 Fallback 경로이다.
// Cloud Storage 임시 → File API 업로드 → 분석 → 삭제
func (a *Analyzer) AnalyzeUploadedVideo(ctx context.Context, fileURI, targetPerson string, progressFn func(string, int)) (*VideoAnalysis, error) {
    slog.Info("fallback_analysis_start", "file_uri", fileURI)

    progressFn("업로드 영상 분석 시작", 10)

    var result *VideoAnalysis
    err := retry.WithBackoff(ctx, 3, func() error {
        resp, err := a.client.Models.GenerateContent(ctx,
            "gemini-2.5-pro",
            genai.FileData{FileURI: fileURI, MIMEType: "video/mp4"},
            genai.Text(buildAnalysisPrompt(targetPerson)),
            &genai.GenerateContentConfig{
                Temperature:      genai.Ptr(float32(0.2)),
                ResponseMIMEType: "application/json",
            },
        )
        if err != nil {
            return err
        }
        result = parseAnalysis(resp)
        return nil
    })

    if err != nil {
        return nil, err
    }

    progressFn("영상 분석 완료", 50)
    return result, nil
}

func buildAnalysisPrompt(targetPerson string) string {
    return fmt.Sprintf(`이 영상에서 "%s" 인물을 분석하라.

반드시 다음 항목을 추출하라:
1. speechPatterns: 말투 패턴, 자주 쓰는 표현, 어미 습관 (최소 5개)
2. expressions: 표정/행동 패턴, 리액션, 제스처 (최소 3개)
3. personalityTraits: 성격 특성 (외향/내향, 유머, 다정함 등) (최소 4개)
4. highlights: 인상적인 장면 타임스탬프 + 설명 + 감정 (최소 3개)
5. voiceCharacteristics: 음성 특징 (톤, 속도, 높낮이)

복수 인물이 있으면 "%s"에 해당하는 인물만 분석하라.`, targetPerson, targetPerson)
}
```

### 체크포인트

- [ ] YouTube 공개 영상 URL → `genai.FileData{FileURI}` → 분석 성공
- [ ] 비공개/일부공개 영상 → 에러 감지 → 갤러리 Fallback 안내
- [ ] 갤러리 업로드 → Cloud Storage → File API → 분석 → 삭제
- [ ] 분석 중 progressFn → 브라우저에 실시간 진행률 전송

---

## T09. Sequential Agent 온보딩 (VideoAnalyzer→VoiceMatcher)

**일수**: 1 | **난이도**: 중 | **의존성**: T08

### 아키텍처: 2단계 파이프라인

- **VideoAnalyzer**: Gemini 2.5 Pro가 YouTube 영상(URL)을 보고 듣고, 대상 인물의 성별·나이대·말투·억양·감정 톤을 텍스트 메타데이터로 추출
- **VoiceMatcher**: 분석 데이터 기반으로 **30개 HD 프리셋 음성** 중 가장 유사한 목소리를 선택(매핑)

> 📌 현 단계에서는 30개 HD 프리셋 음성 매핑으로 구현. (보이스클로닝은 07-FUTURE-DEV.md 참조)

### onboarding/pipeline.go — 2단계 파이프라인

```go
package onboarding

import (
    "context"
    "fmt"
    "log/slog"
)

type Persona struct {
    Name              string   `json:"name"`
    LanguageCode      string   `json:"languageCode"`
    Personality       string   `json:"personality"`
    SpeechStyle       string   `json:"speechStyle"`
    FrequentPhrases   []string `json:"frequentPhrases"`
    EmotionalPatterns string   `json:"emotionalPatterns"`
    MatchedVoice      string   `json:"matchedVoice"`       // 30개 HD 프리셋 중 선택
    VoiceReason       string   `json:"voiceReason"`         // 선택 이유
    VoiceMeta         *VoiceMetadata `json:"voiceMeta"`     // 분석된 음성 메타데이터
    RefImages         [][]byte `json:"-"`                    // 인물 크롭 이미지
}

// VoiceMetadata는 VideoAnalyzer가 추출한 음성 특성이다.
type VoiceMetadata struct {
    Gender        string `json:"gender"`       // "male" | "female"
    AgeRange      string `json:"ageRange"`     // "20s", "30s", "40s" 등
    SpeechTone    string `json:"speechTone"`   // "warm", "energetic", "calm" 등
    PitchLevel    string `json:"pitchLevel"`   // "low", "mid", "high"
    SpeechSpeed   string `json:"speechSpeed"`  // "slow", "normal", "fast"
    Accent        string `json:"accent"`       // "standard", "regional" 등
    EmotionalTone string `json:"emotionalTone"` // "cheerful", "gentle", "serious" 등
}

type Pipeline struct {
    analyzer     *Analyzer       // Stage 1: VideoAnalyzer
    voiceMatcher *VoiceMatcher   // Stage 2: VoiceMatcher
}

type ProgressFunc func(step string, percent int, highlight string)

// Run은 온보딩 2단계 파이프라인을 순차 실행한다.
func (p *Pipeline) Run(ctx context.Context, videos []VideoInput, targetPerson string, progressFn ProgressFunc) (*Persona, error) {
    // Stage 1: VideoAnalyzer — YouTube 영상 분석
    slog.Info("pipeline_stage1_video_analyzer", "videos", len(videos))
    progressFn("영상 분석 중...", 10, "")

    var analyses []*VideoAnalysis
    for i, v := range videos {
        var analysis *VideoAnalysis
        var err error

        if v.YouTubeURL != "" {
            analysis, err = p.analyzer.AnalyzeYouTubeURL(ctx, v.YouTubeURL, targetPerson,
                func(step string, pct int) {
                    progressFn(step, 10+pct*40/100, "")
                })
        } else {
            analysis, err = p.analyzer.AnalyzeUploadedVideo(ctx, v.FileURI, targetPerson,
                func(step string, pct int) {
                    progressFn(step, 10+pct*40/100, "")
                })
        }

        if err != nil {
            slog.Error("analysis_failed", "video_index", i, "error", err)
            continue
        }

        analyses = append(analyses, analysis)

        for _, h := range analysis.Highlights {
            progressFn("발견!", 20+i*10, h.Description)
        }
    }

    if len(analyses) == 0 {
        return nil, fmt.Errorf("no videos analyzed successfully")
    }

    // Stage 2: VoiceMatcher — 30개 HD 프리셋 음성 매핑
    slog.Info("pipeline_stage2_voice_matcher")
    progressFn("음성 매핑 중...", 70, "")

    persona, err := p.voiceMatcher.MatchFromAnalyses(ctx, analyses, targetPerson)
    if err != nil {
        return nil, fmt.Errorf("voice match failed: %w", err)
    }

    progressFn("준비 완료!", 100, "")
    slog.Info("pipeline_complete", "persona", persona.Name, "voice", persona.MatchedVoice)

    return persona, nil
}

type VideoInput struct {
    YouTubeURL string // 공개만: URL 직접 (일부공개 불가)
    FileURI    string // Fallback: File API URI
}
```

### onboarding/voice_matcher.go — 30개 HD 프리셋 음성 매핑

```go
package onboarding

import (
    "context"
    "fmt"
    "log/slog"

    "google.golang.org/genai"
    "github.com/Two-Weeks-Team/missless/internal/retry"
)

// 30개 HD 프리셋 음성 목록 (V7 검증 완료)
// 전체 목록은 00-INDEX.md 참조
var presetVoices = []PresetVoice{
    // 여성 음성 (14개)
    {Name: "Achernar", Gender: "female", Tone: "soft", AgeHint: "mature"},
    {Name: "Aoede", Gender: "female", Tone: "breezy", AgeHint: "young"},
    {Name: "Autonoe", Gender: "female", Tone: "bright", AgeHint: "young"},
    {Name: "Callirrhoe", Gender: "female", Tone: "easy-going", AgeHint: "young"},
    {Name: "Despina", Gender: "female", Tone: "smooth", AgeHint: "mature"},
    {Name: "Erinome", Gender: "female", Tone: "clear", AgeHint: "young"},
    {Name: "Gacrux", Gender: "female", Tone: "mature", AgeHint: "mature"},
    {Name: "Kore", Gender: "female", Tone: "firm", AgeHint: "young"},
    {Name: "Laomedeia", Gender: "female", Tone: "upbeat", AgeHint: "young"},
    {Name: "Leda", Gender: "female", Tone: "youthful", AgeHint: "young"},
    {Name: "Pulcherrima", Gender: "female", Tone: "forward", AgeHint: "young"},
    {Name: "Sulafat", Gender: "female", Tone: "warm", AgeHint: "mature"},
    {Name: "Vindemiatrix", Gender: "female", Tone: "gentle", AgeHint: "mature"},
    {Name: "Zephyr", Gender: "female", Tone: "bright", AgeHint: "young"},
    // 남성 음성 (16개)
    {Name: "Achird", Gender: "male", Tone: "friendly", AgeHint: "young"},
    {Name: "Algenib", Gender: "male", Tone: "gravelly", AgeHint: "mature"},
    {Name: "Algieba", Gender: "male", Tone: "smooth", AgeHint: "mature"},
    {Name: "Alnilam", Gender: "male", Tone: "firm", AgeHint: "mature"},
    {Name: "Charon", Gender: "male", Tone: "informative", AgeHint: "mature"},
    {Name: "Enceladus", Gender: "male", Tone: "breathy", AgeHint: "young"},
    {Name: "Fenrir", Gender: "male", Tone: "excitable", AgeHint: "young"},
    {Name: "Iapetus", Gender: "male", Tone: "clear", AgeHint: "young"},
    {Name: "Orus", Gender: "male", Tone: "firm", AgeHint: "mature"},
    {Name: "Puck", Gender: "male", Tone: "upbeat", AgeHint: "young"},
    {Name: "Rasalgethi", Gender: "male", Tone: "informative", AgeHint: "mature"},
    {Name: "Sadachbia", Gender: "male", Tone: "lively", AgeHint: "young"},
    {Name: "Sadaltager", Gender: "male", Tone: "knowledgeable", AgeHint: "mature"},
    {Name: "Schedar", Gender: "male", Tone: "even", AgeHint: "mature"},
    {Name: "Umbriel", Gender: "male", Tone: "easy-going", AgeHint: "young"},
    {Name: "Zubenelgenubi", Gender: "male", Tone: "casual", AgeHint: "young"},
}

type PresetVoice struct {
    Name    string `json:"name"`
    Gender  string `json:"gender"`
    Tone    string `json:"tone"`
    AgeHint string `json:"ageHint"`
}

type VoiceMatcher struct {
    client *genai.Client
}

// MatchFromAnalyses는 영상 분석 결과에서 페르소나를 생성하고
// 30개 HD 프리셋 중 최적 음성을 매핑한다.
// Gemini가 음성 메타데이터와 프리셋 목록을 비교하여 선택.
// V7 모델: gemini-2.5-pro
func (vm *VoiceMatcher) MatchFromAnalyses(ctx context.Context, analyses []*VideoAnalysis, targetPerson string) (*Persona, error) {
    var persona *Persona
    err := retry.WithBackoff(ctx, 3, func() error {
        resp, err := vm.client.Models.GenerateContent(ctx,
            "gemini-2.5-pro",
            genai.Text(buildVoiceMatchPrompt(analyses, targetPerson)),
            &genai.GenerateContentConfig{
                Temperature:      genai.Ptr(float32(0.3)),
                ResponseMIMEType: "application/json",
                ResponseSchema:   personaWithVoiceSchema(),
            },
        )
        if err != nil {
            return err
        }
        persona = parsePersonaResponse(resp)
        return nil
    })

    if err != nil {
        return nil, err
    }

    slog.Info("voice_matched",
        "persona", persona.Name,
        "voice", persona.MatchedVoice,
        "reason", persona.VoiceReason,
    )
    return persona, nil
}

func buildVoiceMatchPrompt(analyses []*VideoAnalysis, targetPerson string) string {
    return fmt.Sprintf(`다음은 "%s"에 대한 영상 분석 결과입니다:

%s

위 분석을 바탕으로:
1. 페르소나 프로파일을 생성하세요 (성격, 말투, 자주 쓰는 표현, 감정 패턴)
2. 아래 30개 HD 프리셋 음성 중 이 인물과 가장 유사한 음성을 선택하세요:

%s

선택 기준:
- 성별 일치 (필수)
- 나이대 + 목소리 톤 유사도
- 말하는 속도와 에너지 수준
- 감정 표현 방식 유사도

matchedVoice 필드에 선택한 음성 이름, voiceReason에 선택 이유를 반드시 작성하세요.`,
        targetPerson,
        formatAnalyses(analyses),
        formatPresetVoices(),
    )
}
```

### 체크포인트

- [ ] Stage 1 (VideoAnalyzer): YouTube URL 분석 → VoiceMetadata + 페르소나 특성 추출
- [ ] Stage 2 (VoiceMatcher): 분석 결과 + 30개 프리셋 → 최적 음성 매핑 (JSON 구조화 출력)
- [ ] 2단계 모두 retryWithBackoff 적용
- [ ] 단계별 진행률 + 하이라이트 실시간 브라우저 전송
- [ ] 성별 불일치 시 fallback 로직 동작

---

## T10. 온보딩 Live API 세션 (시스템 음성)

**일수**: 1 | **난이도**: 중 | **의존성**: T02, T07

### store/firestore.go — Firestore 세션 스토어

모든 세션 데이터를 Firestore에 영속화. OAuth 토큰, 페르소나, 대화 기록, 재회 세션 컨텍스트.

```go
package store

import (
    "context"
    "time"

    "cloud.google.com/go/firestore"
)

type FirestoreSessionStore struct {
    client *firestore.Client
}

type SessionData struct {
    ID            string                 `firestore:"id"`
    OAuthToken    *OAuthTokenData        `firestore:"oauthToken"`
    Persona       *PersonaData           `firestore:"persona,omitempty"`
    State         string                 `firestore:"state"`
    Transcripts   []TranscriptEntry      `firestore:"transcripts,omitempty"`
    ReunionCount  int                    `firestore:"reunionCount"`        // 계속하기 횟수
    LastSummary   string                 `firestore:"lastSummary,omitempty"` // 이전 세션 요약
    CreatedAt     time.Time              `firestore:"createdAt"`
    UpdatedAt     time.Time              `firestore:"updatedAt"`
}

type TranscriptEntry struct {
    Role      string    `firestore:"role"`      // "user" | "ai"
    Text      string    `firestore:"text"`
    Timestamp time.Time `firestore:"timestamp"`
}

func (s *FirestoreSessionStore) SaveSession(ctx context.Context, data *SessionData) error {
    data.UpdatedAt = time.Now()
    _, err := s.client.Collection("sessions").Doc(data.ID).Set(ctx, data)
    return err
}

func (s *FirestoreSessionStore) GetSession(ctx context.Context, sessionID string) (*SessionData, error) {
    doc, err := s.client.Collection("sessions").Doc(sessionID).Get(ctx)
    if err != nil {
        return nil, err
    }
    var data SessionData
    if err := doc.DataTo(&data); err != nil {
        return nil, err
    }
    return &data, nil
}

// SaveTranscripts는 재회 세션의 대화 기록을 저장한다 (계속하기 기능 지원).
func (s *FirestoreSessionStore) SaveTranscripts(ctx context.Context, sessionID string, transcripts []TranscriptEntry, summary string) error {
    _, err := s.client.Collection("sessions").Doc(sessionID).Update(ctx, []firestore.Update{
        {Path: "transcripts", Value: transcripts},
        {Path: "lastSummary", Value: summary},
        {Path: "reunionCount", Value: firestore.Increment(1)},
        {Path: "updatedAt", Value: time.Now()},
    })
    return err
}
```

### session/manager.go — SessionManager

```go
package session

import (
    "context"
    "fmt"
    "log/slog"
    "sync"

    "google.golang.org/genai"
    "github.com/Two-Weeks-Team/missless/internal/live"
    "github.com/Two-Weeks-Team/missless/internal/onboarding"
    "github.com/Two-Weeks-Team/missless/internal/store"
)

// State는 세션의 현재 상태를 나타낸다.
type State int

const (
    StateIdle State = iota
    StateOnboarding    // 온보딩 진행 중
    StateAnalyzing     // 영상 분석 중 (Live API 유지, AI가 진행 안내)
    StateTransitioning // 세션 전환 중
    StateReunion       // 재회 경험 중
    StateAlbum         // 앨범 생성 중
    StateEnded         // 종료
)

// Manager는 하나의 사용자 세션을 관리한다.
// 브라우저 WebSocket을 유지하면서 Live API 세션을 교체한다.
type Manager struct {
    client      *genai.Client
    proxy       *live.Proxy
    pipeline    *onboarding.Pipeline
    store       *store.FirestoreSessionStore
    state       State
    persona     *onboarding.Persona
    sessionData *store.SessionData

    mu sync.RWMutex
}

// StartOnboarding은 온보딩 Live API 세션을 시작한다.
// missless 시스템 음성(Aoede)으로 사용자를 안내한다.
// V7 모델명: Vertex AI → gemini-live-2.5-flash-native-audio
func (m *Manager) StartOnboarding(ctx context.Context) error {
    m.mu.Lock()
    m.state = StateOnboarding
    m.mu.Unlock()

    config := &genai.LiveConnectConfig{
        SystemInstruction: &genai.Content{
            Parts: []genai.Part{genai.Text(onboardingSystemPrompt)},
        },
        ResponseModalities: []genai.Modality{genai.ModalityAudio},
        SpeechConfig: &genai.SpeechConfig{
            LanguageCode: "ko-KR",
            VoiceConfig: &genai.VoiceConfig{
                PrebuiltVoiceConfig: &genai.PrebuiltVoiceConfig{
                    VoiceName: "Aoede", // 시스템 음성
                },
            },
        },
        InputAudioTranscription:  &genai.InputAudioTranscription{},
        OutputAudioTranscription: &genai.OutputAudioTranscription{},
    }

    return m.proxy.StartSession(ctx, m.client, "gemini-live-2.5-flash-native-audio", config)
}

// TransitionToReunion은 온보딩→재회 세션 전환을 수행한다.
// 브라우저 WebSocket은 유지하고 Live API 세션만 교체한다.
func (m *Manager) TransitionToReunion(ctx context.Context, persona *onboarding.Persona) error {
    m.mu.Lock()
    m.state = StateTransitioning
    m.persona = persona
    m.mu.Unlock()

    slog.Info("session_transition_start", "persona", persona.Name, "voice", persona.MatchedVoice)

    // 브라우저에 전환 알림 → "눈 감아보세요..." UX
    m.proxy.NotifyBrowser(map[string]string{"type": "session_transition"})

    // Live API 세션 교체 (1-2초)
    reunionConfig := buildReunionConfig(persona)
    if err := m.proxy.SwapSession(ctx, m.client, "gemini-live-2.5-flash-native-audio", reunionConfig); err != nil {
        return fmt.Errorf("session swap failed: %w", err)
    }

    m.mu.Lock()
    m.state = StateReunion
    m.mu.Unlock()

    // 브라우저에 준비 완료 알림
    m.proxy.NotifyBrowser(map[string]string{"type": "session_ready"})

    slog.Info("session_transition_complete")
    return nil
}

const onboardingSystemPrompt = `당신은 missless.co의 온보딩 가이드입니다.
사용자가 그리워하는 사람을 만날 수 있도록 안내합니다.

역할:
1. 따뜻하고 친근한 톤으로 사용자를 환영합니다
2. 그리운 사람의 이름과 관계를 물어봅니다
3. YouTube 영상 목록에서 그 사람이 나오는 영상을 선택하도록 안내합니다
4. 영상 분석 중에는 진행 상황을 함께 공유하며 기다립니다
5. 페르소나 프로파일이 완성되면 재회를 준비합니다

중요: 텍스트 입력을 요구하지 마세요. 모든 상호작용은 음성으로 진행합니다.
언어: 한국어로 대화합니다.`
```

### 체크포인트

- [ ] 온보딩 세션 시작 → "안녕하세요, missless에 오신 걸 환영해요" 음성
- [ ] 사용자 음성 → AI 응답 (이름/관계 수집)
- [ ] YouTube 그리드 표시 후 → 영상 선택 이벤트 수신
- [ ] 분석 중 AI가 "분석하고 있어요" 음성 안내 + 진행률 표시

---

## T11. SessionManager — 온보딩→재회 전환

**일수**: 1 | **난이도**: 상 | **의존성**: T09, T10

### 전환 시퀀스

```
1. 온보딩 AI: "좋아요. 민수를 만나러 가볼까요?"
2. Go SessionManager.TransitionToReunion() 호출
3. Browser: "session_transition" 수신 → "눈 감아보세요..." 전환 UX
4. Go: 온보딩 Live API 세션 종료
5. Go: 재회 Live API 세션 생성 (페르소나 System Instruction + HD 음성)
6. Go: 온보딩 대화 요약을 Client Content로 주입
7. Browser: "session_ready" 수신 → 전환 UX 종료
8. 재회 AI (민수 음성): "야, 드디어 왔네."
```

### 온보딩 대화 요약 주입

```go
func buildOnboardingSummary(persona *onboarding.Persona) string {
    return fmt.Sprintf(`[이전 온보딩 대화 요약]
사용자가 그리워하는 사람: %s
관계: 사용자가 직접 설명한 관계
성격: %s
말투: %s
자주 쓰는 표현: %s

당신은 지금 %s입니다. 위의 성격과 말투로 자연스럽게 대화를 시작하세요.
사용자가 방금 당신을 만나러 왔습니다.`,
        persona.Name, persona.Personality, persona.SpeechStyle,
        strings.Join(persona.FrequentPhrases, ", "), persona.Name)
}
```

### 체크포인트

- [ ] 온보딩 세션 종료 → 재회 세션 시작 **2초 이내**
- [ ] 브라우저 WebSocket **끊김 없음**
- [ ] 재회 세션 음성이 페르소나 HD 음성으로 변경됨
- [ ] 재회 AI가 페르소나 성격/말투로 첫 인사
- [ ] 전환 중 브라우저에 "눈 감아보세요" UX 표시

---

## T12. 온보딩 UX (진행률 + 하이라이트 + 인물 선택)

**일수**: 1 | **난이도**: 중 | **의존성**: T04, T08, T09

### 프론트 UX 구성

```
온보딩 UX 상태 머신:

[welcome] → [youtube_grid] → [person_select] → [analyzing] → [transition] → [reunion]

1. welcome: "시작하기" 버튼 (AudioContext 활성화)
2. youtube_grid: YouTube 영상 썸네일 그리드 (공개✅/비공개🔒)
3. person_select: 인물 크롭 이미지 선택
4. analyzing: 분석 진행률 + 하이라이트 카드
5. transition: "눈 감아보세요..." (세션 전환 중)
6. reunion: 전체화면 몰입 (Phase 3)
```

### 분석 진행률 메시지 처리

```typescript
// analysis_progress 이벤트 처리
function handleAnalysisProgress(msg: {
  step: string;
  percent: number;
  highlight?: string;
}) {
  setProgress(msg.percent);
  setCurrentStep(msg.step);

  if (msg.highlight) {
    // 하이라이트 카드 추가 (말투 발견, 표정 감지 등)
    addHighlightCard(msg.highlight);
  }
}
```

### 체크포인트

- [ ] YouTube 그리드에서 영상 터치 → `select_video` 이벤트 전송
- [ ] 비공개/일부공개 영상 터치 → 팝업 안내 (공개 변경 / 갤러리 업로드)
- [ ] 인물 크롭 이미지 그리드 표시 → 터치 선택
- [ ] 분석 진행률 바 실시간 업데이트 (0%→100%)
- [ ] 하이라이트 카드 실시간 추가 ("장난스러운 표정 발견!")
