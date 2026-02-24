# Phase 1: 인프라 + Live API 기반 (T01~T06)

> D-20 ~ D-15 | 6일 | 핵심 마일스톤: 브라우저→Go→Live API 양방향 음성 + 이미지 Fluid Stream

---

## T01. Go scaffolding + 의존성 설정

**일수**: 0.5 | **난이도**: 하 | **의존성**: PRE-01~PRE-12 (사전준비 완료 필수)

### 작업 내용

1. Go 프로젝트 초기화 (`go mod init`)
2. 모든 의존성 설치 및 빌드 확인
3. 환경변수 검증 (config.go)
4. HTTP 서버 + 헬스체크 엔드포인트

> ⚠️ GCP 프로젝트, API 키, OAuth, Firestore, Storage 등은 `01-PREREQUISITES.md`에서 이미 완료된 상태여야 합니다.

### Go 프로젝트 구조

```
missless/
├── cmd/server/main.go            // 엔트리포인트
├── internal/
│   ├── config/config.go          // 환경변수 + 검증
│   ├── session/manager.go        // SessionManager (핵심)
│   ├── session/state.go          // 세션 상태 머신
│   ├── live/proxy.go             // Live API WebSocket 프록시
│   ├── live/tools.go             // Tool 핸들러
│   ├── live/reconnect.go         // GoAway + Session Resumption
│   ├── onboarding/pipeline.go      // Sequential Agent 온보딩 (2단계)
│   ├── onboarding/analyzer.go      // Stage 1: VideoAnalyzer (YouTube URL 분석)
│   ├── onboarding/voice_matcher.go // Stage 2: VoiceMatcher (30개 HD 프리셋 매핑)
│   ├── scene/generator.go        // 2단계 이미지 생성
│   ├── scene/anchor.go           // CharacterAnchor 관리
│   ├── scene/album.go            // 앨범 생성
│   ├── media/youtube.go          // YouTube Data API 클라이언트
│   ├── media/privacy.go          // 영상 공개상태 분류
│   ├── media/upload.go           // 갤러리 Fallback 업로드
│   ├── auth/oauth.go             // Google OAuth 2.0
│   ├── auth/session.go           // 사용자 세션
│   ├── store/firestore.go        // Firestore 세션 스토어 (OAuth, 페르소나, 대화)
│   ├── memory/store.go           // Firestore 추억 CRUD
│   ├── handler/websocket.go      // 브라우저 WebSocket 핸들러
│   ├── handler/upload.go         // 미디어 업로드 API
│   ├── handler/health.go         // 헬스체크
│   ├── handler/oauth_callback.go // OAuth 콜백
│   ├── middleware/recovery.go    // panic recovery
│   ├── middleware/logging.go     // slog 구조화 로깅
│   ├── util/safego.go            // SafeGo goroutine 래퍼
│   └── retry/backoff.go          // Exponential Backoff + Jitter
├── web/                           // Next.js 15 (static export → Go FileServer로 서빙)
├── deploy/
│   ├── Dockerfile
│   ├── cloudbuild.yaml
│   └── terraform/
├── go.mod
├── go.sum
└── README.md
```

### Go 모듈 의존성 (V7 검증 완료)

```go
// go.mod
module github.com/Two-Weeks-Team/missless

go 1.22

require (
    google.golang.org/genai v1.47.0   // V7 검증: 2026-02-19 릴리스
    google.golang.org/adk v0.5.0      // V7 검증: 최신 버전
    google.golang.org/api v0.214.0     // YouTube Data API v3
    github.com/gorilla/websocket v1.5.3
    cloud.google.com/go/firestore v1.17.0
    cloud.google.com/go/storage v1.47.0
    golang.org/x/oauth2 v0.25.0
    golang.org/x/sync v0.10.0          // errgroup
)
```

### config.go — 환경변수 검증

```go
package config

import (
    "fmt"
    "os"
)

type Config struct {
    GeminiAPIKey    string
    ProjectID       string
    Port            string
    YouTubeClientID string
    YouTubeSecret   string
    StorageBucket   string
    FirestoreDB     string
}

func Load() (*Config, error) {
    c := &Config{
        GeminiAPIKey:    os.Getenv("GEMINI_API_KEY"),
        ProjectID:       os.Getenv("GCP_PROJECT_ID"),
        Port:            os.Getenv("PORT"),
        YouTubeClientID: os.Getenv("YOUTUBE_CLIENT_ID"),
        YouTubeSecret:   os.Getenv("YOUTUBE_CLIENT_SECRET"),
        StorageBucket:   os.Getenv("STORAGE_BUCKET"),
        FirestoreDB:     os.Getenv("FIRESTORE_DB"),
    }
    if c.Port == "" {
        c.Port = "8080"
    }
    if c.GeminiAPIKey == "" {
        return nil, fmt.Errorf("GEMINI_API_KEY required")
    }
    if c.ProjectID == "" {
        return nil, fmt.Errorf("GCP_PROJECT_ID required")
    }
    return c, nil
}
```

### 체크포인트

- [ ] `go run cmd/server/main.go` — 서버 시작 성공
- [ ] 5개 모델 `GenerateContent` 호출 성공 (단순 텍스트 프롬프트)
- [ ] YouTube Data API `channels.list` 호출 성공
- [ ] Tier 1 활성화 확인 (RPM 한도 150+)

---

## T02. Go WebSocket 프록시 + Live API 연결 (핵심)

**일수**: 1.5 | **난이도**: 상 | **의존성**: T01

### V7 핵심 아키텍처 — Go 백엔드가 Live API 소유

V4에서는 브라우저가 Ephemeral Token으로 Live API에 직접 연결했으나, V5 이후 **Go 백엔드가 Live API 세션을 소유**하고 브라우저는 Go에만 연결합니다.

```
[Browser] ←──WebSocket──→ [Go Backend] ←──WebSocket──→ [Gemini Live API]
  (PCM 오디오 송수신)         (프록시 + Tool 실행)        (양방향 스트리밍)
```

### proxy.go — Live API 프록시

```go
package live

import (
    "context"
    "io"
    "log/slog"
    "sync"

    "github.com/gorilla/websocket"
    "google.golang.org/genai"
)

// Proxy는 브라우저↔Live API 사이의 오디오/이벤트 프록시이다.
// Go 백엔드가 Live API 세션을 소유하므로 모든 Tool 실행이 서버 내부에서 처리된다.
type Proxy struct {
    browserConn *websocket.Conn   // 브라우저 WebSocket
    liveSession *genai.Session    // Gemini Live API 세션
    toolHandler *ToolHandler      // 서버사이드 Tool 실행기

    mu          sync.Mutex        // liveSession 교체 시 동시성 보호
    done        chan struct{}     // graceful shutdown
}

// NewProxy는 브라우저 연결을 받아 Live API 프록시를 생성한다.
func NewProxy(browserConn *websocket.Conn, toolHandler *ToolHandler) *Proxy {
    return &Proxy{
        browserConn: browserConn,
        toolHandler: toolHandler,
        done:        make(chan struct{}),
    }
}

// StartSession은 Live API 세션을 생성하고 양방향 프록시를 시작한다.
// V7 모델명: Developer API → gemini-2.5-flash-native-audio-preview-12-2025
//            Vertex AI    → gemini-live-2.5-flash-native-audio
func (p *Proxy) StartSession(ctx context.Context, client *genai.Client, model string, config *genai.LiveConnectConfig) error {
    session, err := client.Live.Connect(ctx, model, config)
    if err != nil {
        return err
    }

    p.mu.Lock()
    p.liveSession = session
    p.mu.Unlock()

    // 브라우저→Live API (오디오 포워딩)
    go p.forwardBrowserToLive(ctx)

    // Live API→브라우저 (응답 수신 + Tool 실행)
    go p.forwardLiveToBrowser(ctx)

    return nil
}

// forwardBrowserToLive는 브라우저 오디오 PCM을 Live API에 포워딩한다.
func (p *Proxy) forwardBrowserToLive(ctx context.Context) {
    defer func() {
        if r := recover(); r != nil {
            slog.Error("browser_to_live_panic", "recover", r)
        }
    }()

    for {
        select {
        case <-ctx.Done():
            return
        case <-p.done:
            return
        default:
        }

        msgType, data, err := p.browserConn.ReadMessage()
        if err != nil {
            if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
                slog.Info("browser_disconnected")
                return
            }
            slog.Error("browser_read_error", "error", err)
            return
        }

        p.mu.Lock()
        session := p.liveSession
        p.mu.Unlock()

        if session == nil {
            continue // 세션 전환 중
        }

        switch msgType {
        case websocket.BinaryMessage:
            // PCM 오디오 데이터 → Live API
            if err := session.SendRealtimeInput(genai.LiveRealtimeInput{
                Audio: &genai.Blob{
                    MIMEType: "audio/pcm;rate=16000",
                    Data:     data,
                },
            }); err != nil {
                slog.Error("live_send_audio_error", "error", err)
            }

        case websocket.TextMessage:
            // JSON 이벤트 (audioStreamEnd 등) → Live API
            p.handleBrowserEvent(ctx, data)
        }
    }
}

// forwardLiveToBrowser는 Live API 응답을 브라우저에 전달하고 Tool Call을 처리한다.
func (p *Proxy) forwardLiveToBrowser(ctx context.Context) {
    defer func() {
        if r := recover(); r != nil {
            slog.Error("live_to_browser_panic", "recover", r)
        }
    }()

    for {
        p.mu.Lock()
        session := p.liveSession
        p.mu.Unlock()

        if session == nil {
            continue
        }

        msg, err := session.Receive()
        if err != nil {
            if err == io.EOF {
                slog.Info("live_session_ended")
                return
            }
            slog.Error("live_receive_error", "error", err)
            return
        }

        // 1) 오디오 응답 → 브라우저로 포워딩
        if msg.ServerContent != nil {
            for _, part := range msg.ServerContent.ModelTurn.Parts {
                if part.InlineData != nil {
                    // PCM 오디오 바이너리 → 브라우저
                    p.browserConn.WriteMessage(websocket.BinaryMessage, part.InlineData.Data)
                }
            }
        }

        // 2) Tool Call → 서버에서 직접 실행
        if msg.ToolCall != nil {
            go p.executeToolCall(ctx, msg.ToolCall)
        }

        // 3) GoAway 시그널 → 자동 재접속
        if msg.GoAway != nil {
            slog.Info("goaway_received")
            go p.handleGoAway(ctx)
        }

        // 4) Session Resumption 토큰 갱신
        if msg.SessionResumption != nil && msg.SessionResumption.Handle != "" {
            p.toolHandler.UpdateResumptionToken(msg.SessionResumption.Handle)
        }
    }
}

// executeToolCall은 Tool Call을 서버 내부에서 실행하고 결과를 Live API에 반환한다.
func (p *Proxy) executeToolCall(ctx context.Context, toolCall *genai.ToolCall) {
    for _, fc := range toolCall.FunctionCalls {
        response := p.toolHandler.Handle(ctx, fc, p.browserConn)

        p.mu.Lock()
        session := p.liveSession
        p.mu.Unlock()

        if session != nil {
            if err := session.SendToolResponse(genai.LiveToolResponseInput{FunctionResponses: []*genai.FunctionResponse{response}}); err != nil {
                slog.Error("tool_response_error", "tool", fc.Name, "error", err)
            }
        }
    }
}

// SwapSession은 세션을 교체한다 (온보딩→재회 전환 시).
// 브라우저 WebSocket은 유지하고 Live API 세션만 교체.
func (p *Proxy) SwapSession(ctx context.Context, client *genai.Client, model string, config *genai.LiveConnectConfig) error {
    p.mu.Lock()
    oldSession := p.liveSession
    p.liveSession = nil // 전환 중 표시
    p.mu.Unlock()

    // 이전 세션 종료
    if oldSession != nil {
        oldSession.Close()
    }

    // 새 세션 시작
    newSession, err := client.Live.Connect(ctx, model, config)
    if err != nil {
        return err
    }

    p.mu.Lock()
    p.liveSession = newSession
    p.mu.Unlock()

    slog.Info("session_swapped")
    return nil
}

// Close는 모든 연결을 정리한다.
func (p *Proxy) Close() {
    close(p.done)
    p.mu.Lock()
    if p.liveSession != nil {
        p.liveSession.Close()
    }
    p.mu.Unlock()
}
```

### 체크포인트

- [ ] 브라우저→Go WebSocket 연결 성공
- [ ] Go→Live API WebSocket 연결 성공
- [ ] 브라우저 마이크 PCM → Go → Live API → AI 응답 PCM → Go → 브라우저 스피커
- [ ] 양방향 음성 대화 1분 이상 유지
- [ ] GoAway 시그널 수신 → 자동 재접속 → 대화 끊김 없음

---

## T03. Tool 등록 + 서버사이드 실행 기반

**일수**: 1 | **난이도**: 중 | **의존성**: T02

### tools.go — 서버사이드 Tool 핸들러

```go
package live

import (
    "context"
    "log/slog"
    "time"

    "github.com/gorilla/websocket"
    "google.golang.org/genai"
)

type ToolHandler struct {
    genaiClient     *genai.Client
    sceneGen        *scene.Generator
    memoryStore     *memory.Store
    resumptionToken string
    sessionID       string

    mu sync.RWMutex // resumptionToken 접근 보호
}

// Handle은 FunctionCall을 받아 실행하고 FunctionResponse를 반환한다.
// 모든 API 호출은 Go 서버 내부에서 처리되므로 API 키가 브라우저에 노출되지 않는다.
func (th *ToolHandler) Handle(ctx context.Context, fc *genai.FunctionCall, browserWs *websocket.Conn) *genai.FunctionResponse {
    start := time.Now()

    defer func() {
        if r := recover(); r != nil {
            slog.Error("tool_panic", "tool", fc.Name, "recover", r)
        }
        slog.Info("tool_executed",
            "tool", fc.Name,
            "latency_ms", time.Since(start).Milliseconds(),
        )
    }()

    switch fc.Name {
    case "generate_scene":
        return th.handleGenerateScene(ctx, fc, browserWs)
    case "generate_fast_scene":
        return th.handleGenerateFastScene(ctx, fc, browserWs)
    case "change_atmosphere":
        return th.handleChangeAtmosphere(ctx, fc, browserWs)
    case "recall_memory":
        return th.handleRecallMemory(ctx, fc)
    case "analyze_user":
        return th.handleAnalyzeUser(ctx, fc)
    case "end_reunion":
        return th.handleEndReunion(ctx, fc, browserWs)
    default:
        slog.Warn("unknown_tool", "name", fc.Name)
        return &genai.FunctionResponse{
            Name:     fc.Name,
            Response: map[string]any{"error": "unknown tool"},
        }
    }
}

// handleGenerateScene — 2단계 Progressive Rendering
// Stage 1: flash-image (1-3초) → scene_preview 전송
// Stage 2: pro-image (8-12초) → scene_final 전송 (크로스페이드)
func (th *ToolHandler) handleGenerateScene(ctx context.Context, fc *genai.FunctionCall, ws *websocket.Conn) *genai.FunctionResponse {
    prompt, _ := fc.Args["prompt"].(string)
    mood, _ := fc.Args["mood"].(string)
    characters, _ := fc.Args["characters"].(string)

    // Stage 1: flash-image 즉시 생성 (비동기)
    go func() {
        ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
        defer cancel()

        img, err := th.sceneGen.GenerateFlash(ctx, prompt, mood, characters)
        if err != nil {
            slog.Error("flash_scene_failed", "error", err)
            writeJSON(ws, map[string]any{
                "type": "tool_error", "tool": "generate_scene",
                "message": "이미지 생성에 문제가 생겼습니다.",
            })
            return
        }
        writeJSON(ws, map[string]any{
            "type": "scene_preview", "image": img,
        })
    }()

    // Stage 2: pro-image 고품질 생성 (비동기)
    go func() {
        ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
        defer cancel()

        img, err := th.sceneGen.GeneratePro(ctx, prompt, mood, characters)
        if err != nil {
            slog.Warn("pro_scene_failed_using_flash_only", "error", err)
            return // flash만 사용
        }
        writeJSON(ws, map[string]any{
            "type": "scene_final", "image": img,
        })
    }()

    // 즉시 Dummy 응답 → Live API 음성 재개
    return &genai.FunctionResponse{
        Name:     "generate_scene",
        Response: map[string]any{"status": "generating", "description": prompt},
    }
}

func writeJSON(ws *websocket.Conn, v any) {
    if err := ws.WriteJSON(v); err != nil {
        slog.Error("ws_write_error", "error", err)
    }
}
```

### LiveConnectConfig — Tool 등록

```go
func buildLiveConfig(persona *Persona) *genai.LiveConnectConfig {
    return &genai.LiveConnectConfig{
        SystemInstruction: &genai.Content{
            Parts: []genai.Part{genai.Text(buildSystemPrompt(persona))},
        },
        ResponseModalities: []genai.Modality{genai.ModalityAudio},
        EnableAffectiveDialog: true,
        ProactiveAudio:        true,
        SpeechConfig: &genai.SpeechConfig{
            LanguageCode: persona.LanguageCode,
            VoiceConfig: &genai.VoiceConfig{
                PrebuiltVoiceConfig: &genai.PrebuiltVoiceConfig{
                    VoiceName: persona.MatchedVoice,
                },
            },
        },
        InputAudioTranscription:  &genai.InputAudioTranscription{},
        OutputAudioTranscription: &genai.OutputAudioTranscription{},
        Tools: []*genai.Tool{
            {GoogleSearch: &genai.GoogleSearch{}},
            {FunctionDeclarations: toolDeclarations()},
        },
        ContextWindowCompression: &genai.ContextWindowCompressionConfig{
            SlidingWindow: &genai.SlidingWindow{
                TargetTokenCount:  8192,
                TriggerTokenCount: 12288,
            },
        },
        SessionResumption: &genai.SessionResumptionConfig{},
        // ⚠️ Transparent 필드는 Vertex AI 전용. Developer API에서는 미지원.
    }
}
```

### 체크포인트

- [ ] Live API가 Tool Call 발생 → Go 핸들러에서 수신 성공
- [ ] FunctionResponse 반환 → Live API 음성 스트리밍 재개
- [ ] 비동기 이미지 생성 goroutine이 브라우저에 이미지 전송
- [ ] Tool 실행 중 panic → recovery로 서버 안정 유지

---

## T04. Next.js PWA 순수 렌더러 (Static Export + Go FileServer)

**일수**: 1 | **난이도**: 중 | **의존성**: T01

### V7 프론트엔드 역할 — 순수 렌더러

V4에서 프론트가 Live API 직접 연결 + Tool 중계를 담당했으나, V5 이후 **순수 렌더러**:

- Go 백엔드와 단일 WebSocket 연결만 유지
- 오디오 PCM 송수신
- 이미지/이벤트 수신 및 렌더링
- UI 상태 관리 (온보딩/재회/앨범)

### Next.js Static Export + Go FileServer 서빙

**빌드**: `next.config.js`에 `output: 'export'` → `web/out/` 디렉터리에 정적 파일 생성
**서빙**: Go 서버가 `http.FileServer`로 정적 파일 직접 서빙 (동일 오리진 → CORS 불필요)

```go
// cmd/server/main.go — 정적 파일 서빙
mux := http.NewServeMux()

// API + WebSocket 라우트
mux.HandleFunc("/ws", handler.WebSocket)
mux.HandleFunc("/api/", handler.API)
mux.HandleFunc("/auth/", handler.OAuth)

// Next.js Static Export 서빙 (SPA fallback)
staticFS := http.FileServer(http.Dir("web/out"))
mux.Handle("/", staticFS)
```

```javascript
// web/next.config.js
/** @type {import('next').NextConfig} */
const nextConfig = {
  output: 'export',       // Static HTML Export
  trailingSlash: true,    // /about → /about/index.html
  images: { unoptimized: true },  // Static export에서 Image Optimization 비활성화
};
module.exports = nextConfig;
```

### 프론트 메시지 프로토콜

```typescript
// Go → Browser 메시지 타입
type ServerMessage =
  | { type: "audio"; data: ArrayBuffer }         // PCM 24kHz 오디오
  | { type: "scene_preview"; image: string }      // flash-image base64
  | { type: "scene_final"; image: string }        // pro-image base64
  | { type: "bgm_change"; mood: string }          // BGM 전환
  | { type: "tool_error"; tool: string; message: string }
  | { type: "session_transition" }                // 온보딩→재회 전환 중
  | { type: "session_ready" }                     // 새 세션 준비 완료
  | { type: "analysis_progress"; step: string; percent: number; highlight?: string }
  | { type: "person_detected"; crops: string[] }  // 인물 크롭 이미지
  | { type: "youtube_videos"; videos: YouTubeVideo[] }
  | { type: "transcript"; role: string; text: string }  // 자막

// Browser → Go 메시지 타입
type ClientMessage =
  | { type: "audio"; data: ArrayBuffer }          // PCM 16kHz 오디오
  | { type: "audio_stream_end" }                  // 무음 감지
  | { type: "select_video"; videoId: string }     // YouTube 영상 선택
  | { type: "select_person"; personIndex: number } // 인물 선택
  | { type: "upload_video"; data: ArrayBuffer }   // 갤러리 Fallback
```

### AudioContext 활성화

```typescript
// Step 1.5: "시작하기" 버튼 터치에서 모든 미디어 초기화
async function handleStartTouch() {
  const audioCtx = new AudioContext({ sampleRate: 24000 });
  await audioCtx.resume();

  const stream = await navigator.mediaDevices.getUserMedia({
    audio: { sampleRate: 16000, channelCount: 1, echoCancellation: true }
  });

  const ws = new WebSocket(`wss://${location.host}/ws`);
  // ... AudioWorklet 설정 + PCM 스트리밍
}
```

### 체크포인트

- [ ] AudioContext.resume() 모바일 Chrome에서 성공
- [ ] Go WebSocket 연결 + PCM 오디오 양방향 스트리밍
- [ ] `scene_preview` 수신 → 전체화면 이미지 표시
- [ ] `scene_final` 수신 → 크로스페이드 애니메이션
- [ ] `session_transition` → 전환 UX 표시 → `session_ready` → 재개

---

## T05. 2단계 이미지 생성 (flash→pro Progressive)

**일수**: 1 | **난이도**: 상 | **의존성**: T02, T03

### scene/generator.go

```go
package scene

import (
    "context"
    "encoding/base64"
    "fmt"
    "log/slog"
    "sync"

    "google.golang.org/genai"
)

type Generator struct {
    client *genai.Client
    anchor *CharacterAnchor
    mu     sync.RWMutex // anchor 접근 보호
}

// GenerateFlash는 flash-image로 빠른 프리뷰 이미지를 생성한다 (1-3초).
// V7 모델: gemini-2.5-flash-image (stable, 2026-01-21 GA)
func (g *Generator) GenerateFlash(ctx context.Context, prompt, mood, characters string) (string, error) {
    parts := g.buildPromptParts(prompt, mood, characters, false)

    resp, err := g.client.Models.GenerateContent(ctx,
        "gemini-2.5-flash-image",
        parts...,
    )
    if err != nil {
        return "", fmt.Errorf("flash generate: %w", err)
    }

    return g.extractImageBase64(resp)
}

// GeneratePro는 pro-image로 고품질 이미지를 생성한다 (8-12초).
// V7 모델: imagen-4.0-generate-001 (Imagen 4) — Imagen 3는 2025-11-10 shut down
// 레퍼런스 이미지 체이닝으로 캐릭터 일관성을 유지한다.
func (g *Generator) GeneratePro(ctx context.Context, prompt, mood, characters string) (string, error) {
    parts := g.buildPromptParts(prompt, mood, characters, true)

    resp, err := g.client.Models.GenerateContent(ctx,
        "imagen-4.0-generate-001",
        parts...,
    )
    if err != nil {
        return "", fmt.Errorf("pro generate: %w", err)
    }

    imgB64, err := g.extractImageBase64(resp)
    if err != nil {
        return "", err
    }

    // 직전 장면 업데이트 (체이닝)
    g.mu.Lock()
    g.anchor.LastSceneB64 = imgB64
    g.mu.Unlock()

    return imgB64, nil
}

// buildPromptParts는 프롬프트 + 레퍼런스 이미지를 조합한다.
func (g *Generator) buildPromptParts(prompt, mood, characters string, includeRef bool) []genai.Part {
    text := fmt.Sprintf(
        "장면: %s\n분위기: %s\n등장인물: %s\n"+
            "스타일: 따뜻한 수채화 일러스트, 측면/뒷모습 위주, 부드러운 톤\n"+
            "레퍼런스 이미지의 인물 외형을 반드시 유지하세요.",
        prompt, mood, characters,
    )

    parts := []genai.Part{genai.Text(text)}

    if includeRef {
        g.mu.RLock()
        anchor := g.anchor
        g.mu.RUnlock()

        if anchor != nil {
            for _, ref := range anchor.RefImages {
                parts = append(parts, genai.ImageData("jpeg", ref))
            }
            if anchor.LastSceneB64 != "" {
                decoded, _ := base64.StdEncoding.DecodeString(anchor.LastSceneB64)
                parts = append(parts, genai.ImageData("png", decoded))
            }
        }
    }

    return parts
}

func (g *Generator) extractImageBase64(resp *genai.GenerateContentResponse) (string, error) {
    for _, cand := range resp.Candidates {
        for _, part := range cand.Content.Parts {
            if part.InlineData != nil {
                return base64.StdEncoding.EncodeToString(part.InlineData.Data), nil
            }
        }
    }
    return "", fmt.Errorf("no image in response")
}
```

### 체크포인트

- [ ] flash-image 호출 → 1-3초 이내 이미지 반환
- [ ] pro-image 호출 → 8-12초 이내 이미지 반환
- [ ] `scene_preview` 먼저 표시 → `scene_final` 크로스페이드
- [ ] CharacterAnchor 레퍼런스 체이닝으로 3장면 이상 일관성 유지
- [ ] 이미지 생성 실패 시 `tool_error` 이벤트 전송

---

## T06. Rate Limit 방어 + 에러 핸들링 기반

**일수**: 0.5 | **난이도**: 중 | **의존성**: T01

### retry/backoff.go

```go
package retry

import (
    "context"
    "fmt"
    "log/slog"
    "math/rand"
    "time"
)

// WithBackoff는 fn을 최대 maxRetries번 재시도한다.
// 429/5xx 에러 시 Exponential Backoff + Jitter를 적용한다.
func WithBackoff(ctx context.Context, maxRetries int, fn func() error) error {
    for i := 0; i < maxRetries; i++ {
        err := fn()
        if err == nil {
            return nil
        }

        if i >= maxRetries-1 {
            return fmt.Errorf("max retries exceeded: %w", err)
        }

        // Exponential Backoff: 1s, 2s, 4s + Jitter (0~50%)
        base := time.Duration(1<<uint(i)) * time.Second
        jitter := time.Duration(rand.Int63n(int64(base / 2)))
        wait := base + jitter

        slog.Warn("api_retry",
            "attempt", i+1,
            "backoff_ms", wait.Milliseconds(),
            "error", err.Error(),
        )

        select {
        case <-ctx.Done():
            return ctx.Err()
        case <-time.After(wait):
        }
    }
    return nil
}
```

### middleware/recovery.go — Panic Recovery

```go
package middleware

import (
    "log/slog"
    "net/http"
    "runtime/debug"
)

func Recovery(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        defer func() {
            if err := recover(); err != nil {
                slog.Error("http_panic",
                    "error", err,
                    "stack", string(debug.Stack()),
                    "path", r.URL.Path,
                )
                http.Error(w, "Internal Server Error", http.StatusInternalServerError)
            }
        }()
        next.ServeHTTP(w, r)
    })
}
```

### 체크포인트

- [ ] retryWithBackoff: 429 에러 → 재시도 성공
- [ ] panic recovery: HTTP 핸들러 panic → 500 반환, 서버 유지
- [ ] WebSocket goroutine panic → recovery 로그 → 연결 정리
- [ ] context timeout: 10초 초과 시 깔끔한 취소
