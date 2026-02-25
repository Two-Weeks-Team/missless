# PROJECT KNOWLEDGE BASE

**Generated:** 2026-02-25
**Branch:** master

## OVERVIEW

**missless** — AI-powered virtual reunion experience for the Gemini Live Agent Challenge 2026 (Creative Storyteller track). YouTube video analysis → persona generation → real-time voice + image + BGM storytelling. Go backend with WebSocket proxy to Gemini Live API, Next.js 15 PWA frontend.

## STRUCTURE

```
missless/
├── cmd/server/main.go                 # Entry point — HTTP server + graceful shutdown
├── internal/                          # Go backend packages (13 modules)
│   ├── config/                        # Environment config + HTTP client factory
│   │   ├── config.go                  # Env loading (default port 18080)
│   │   ├── http.go                    # Shared HTTP client with timeouts
│   │   └── config_test.go
│   ├── live/                          # Live API proxy layer (core)
│   │   ├── proxy.go                   # WebSocket <> Live API bidirectional proxy
│   │   ├── tools.go                   # 6 server-side tool handlers
│   │   ├── bgm.go                     # Preset BGM URL mapper
│   │   ├── reconnect.go              # Session resumption logic
│   │   ├── proxy_test.go             # Proxy unit tests
│   │   └── tools_test.go
│   ├── session/                       # State machine
│   │   ├── manager.go                # SessionManager (state transitions)
│   │   ├── state.go                  # State enum + transitions
│   │   └── manager_test.go
│   ├── onboarding/                    # Sequential Agent pipeline
│   │   ├── analyzer.go               # YouTube video analysis (Stage 1)
│   │   ├── pipeline.go               # Onboarding orchestrator
│   │   ├── voice_matcher.go          # 30 preset voice mapping (Stage 2)
│   │   ├── analyzer_test.go
│   │   └── pipeline_test.go
│   ├── scene/                         # Image generation + album
│   │   ├── generator.go              # Progressive 2-stage (Flash -> Imagen 4)
│   │   ├── anchor.go                 # CharacterAnchor (consistency)
│   │   ├── album.go                  # Album compilation + sharing
│   │   ├── generator_test.go
│   │   └── album_test.go
│   ├── handler/                       # HTTP/WebSocket handlers
│   │   ├── websocket.go              # WS upgrade + session init
│   │   ├── upload.go                 # File upload handler
│   │   ├── health.go                 # Health check endpoint
│   │   ├── oauth_callback.go         # OAuth redirect handler
│   │   ├── websocket_test.go
│   │   ├── oauth_callback_test.go
│   │   ├── health_test.go
│   │   └── integration_test.go
│   ├── media/                         # External media services
│   │   ├── youtube.go                # YouTube Data API client
│   │   ├── privacy.go               # Video privacy checker
│   │   ├── upload.go                 # Cloud Storage upload
│   │   └── youtube_test.go
│   ├── auth/                          # Authentication
│   │   ├── oauth.go                  # Google OAuth 2.0 flow
│   │   └── session.go               # Session cookie management
│   ├── store/                         # Persistence
│   │   ├── firestore.go              # Firestore session store
│   │   └── firestore_test.go
│   ├── memory/                        # Recall memory for reunion
│   │   ├── store.go                  # Memory CRUD (search by persona)
│   │   └── store_test.go
│   ├── middleware/                     # HTTP middleware
│   │   ├── recovery.go               # Panic recovery
│   │   ├── logging.go               # Request logging
│   │   ├── timeout.go               # Request timeout
│   │   ├── recovery_test.go
│   │   ├── logging_test.go
│   │   └── timeout_test.go
│   ├── retry/                         # Resilience
│   │   ├── backoff.go               # Exponential backoff
│   │   └── backoff_test.go
│   └── util/                          # Shared utilities
│       ├── safego.go                 # SafeGo goroutine wrapper (panic recovery)
│       └── safego_test.go
├── web/                               # Next.js 15 PWA frontend
│   ├── app/
│   │   ├── page.tsx                  # Main page (voice-first reunion UI)
│   │   ├── layout.tsx                # Root layout
│   │   └── album/page.tsx            # Album sharing page
│   ├── components/
│   │   ├── OnboardingFlow.tsx        # YouTube URL input flow
│   │   ├── SceneDisplay.tsx          # Progressive image display
│   │   ├── BGMPlayer.tsx             # Background music player
│   │   ├── YouTubeGrid.tsx           # Video selection grid
│   │   ├── HighlightCard.tsx         # Highlight moment card
│   │   ├── ProgressBar.tsx           # Analysis progress indicator
│   │   ├── SessionTransition.tsx     # State transition animation
│   │   └── PrivateVideoPopup.tsx     # Privacy warning popup
│   ├── hooks/
│   │   ├── useWebSocket.ts           # WS connection management
│   │   └── useAudio.ts              # Audio capture/playback
│   ├── styles/                       # CSS styles
│   ├── public/                       # Static assets
│   ├── next.config.js                # Next.js configuration
│   ├── tsconfig.json                 # TypeScript config
│   ├── eslint.config.mjs             # ESLint config
│   └── package.json                  # Node dependencies
├── deploy/                            # Infrastructure
│   ├── Dockerfile                    # Multi-stage Go build
│   └── cloudbuild.yaml               # Cloud Build pipeline
├── plan/                              # Planning documents (Korean)
│   └── v7/                           # Latest — canonical implementation spec
│       └── 00-INDEX.md               # Start here for plan context
├── .github/workflows/ci.yml          # GitHub Actions CI
├── go.mod / go.sum                   # Go module (Go 1.24)
└── .env.example                      # Environment variable template
```

## WHERE TO LOOK

| Task | Location | Notes |
|------|----------|-------|
| **Entry point** | `cmd/server/main.go` | HTTP routes, graceful shutdown |
| **WebSocket proxy (core)** | `internal/live/proxy.go` | Bidirectional audio + event streaming |
| **Tool execution** | `internal/live/tools.go` | 6 tools: generate_scene, change_atmosphere, etc. |
| **State machine** | `internal/session/manager.go` + `state.go` | Onboarding -> Reunion -> Album -> Ended |
| **Image generation** | `internal/scene/generator.go` | 2-stage progressive (Flash -> Imagen 4) |
| **YouTube analysis** | `internal/onboarding/analyzer.go` | Gemini 2.5 Pro video analysis |
| **Voice mapping** | `internal/onboarding/voice_matcher.go` | 30 preset voice -> persona matching |
| **Frontend main** | `web/app/page.tsx` | Voice-first UI entry |
| **WS hook** | `web/hooks/useWebSocket.ts` | Client WebSocket management |
| **Implementation plan** | `plan/v7/00-INDEX.md` | V7 spec (canonical) |
| **Go safety rules** | `plan/v7/06-GO-SAFETY.md` | Lock ordering, SafeGo, race detector |
| **Deploy** | `deploy/` | Dockerfile, cloudbuild.yaml |

## CONVENTIONS

- **Language**: Code in English, planning docs in Korean
- **Go safety**: All goroutines via `util.SafeGo()` — bare `go func()` is forbidden
- **Lock ordering**: Manager(L1) -> Proxy(L2) -> ToolHandler(L3) -> CharacterAnchor(L4) -> AlbumGenerator(L5) -> MemoryStore(L6)
- **No I/O under locks**: Perform I/O outside critical sections
- **Buffered channels**: Always `make(chan T, 1)` minimum for goroutine results
- **Version naming**: v1->v7 in plan/, V7 is canonical
- **Task IDs**: `PRE-XX` for prerequisites, `T-XX` for dev tasks
- **Test files**: Co-located `*_test.go` next to source

## KEY DECISIONS

| Decision | Choice | Rationale |
|----------|--------|-----------|
| Track | Creative Storyteller (not Live Agent) | Lower competition, better multimodal showcase |
| Backend | Go + Cloud Run (not Python ADK) | Performance + Google Cloud scoring |
| Live API | Go backend proxy (not browser direct) | Tool execution security + latency |
| BGM | Preset files from Cloud Storage (not Lyria) | Lyria WebSocket-only, no Go SDK |
| Image gen | 2-stage: Flash preview(1-3s) -> Imagen 4(8-12s) | Quality + perceived speed |
| Voice | 30 preset voices (not voice cloning) | Cloning deferred to post-MVP |
| YouTube | URL-direct via Gemini (not download) | Zero-download architecture |
| API platform | Vertex AI (not Developer API) | Cloud deployment scoring |

## ANTI-PATTERNS

- **NEVER use Imagen 3** — shut down 2025-11-10. Use `imagen-4.0-generate-001`
- **NEVER use Lyria in Go** — Go SDK has no Lyria types. Preset BGM only
- **NEVER call `session.Receive(ctx)`** — `Session.Receive()` does not accept a ctx parameter. Use `session.Receive()`
- **NEVER analyze unlisted YouTube videos** — Gemini FileData can't access unlisted URLs
- **NEVER use `SessionResumptionConfig.Transparent` with Developer API** — Vertex AI only
- **NEVER use bare `go func()`** — always `util.SafeGo()` for panic recovery

## MODEL MATRIX (verified 2026-02-24)

| Purpose | Model ID | Notes |
|---------|----------|-------|
| Live API (voice) | `gemini-2.5-flash-native-audio` (Vertex) | Real-time bidirectional |
| Image preview | `gemini-2.5-flash-image` | 1-3s |
| Image final | `imagen-4.0-generate-001` | Imagen 4, 8-12s |
| Video analysis | `gemini-2.5-pro` | GA |
| Persona generation | `gemini-2.5-pro` | JSON structured output |

## COMMANDS

```bash
# Backend
go run cmd/server/main.go              # Run locally (port 18080)
go test -race ./...                    # Run all tests with race detector
go vet ./...                           # Static analysis
go build ./...                         # Compile check

# Frontend
cd web && npm install                  # Install deps
cd web && npm run build                # Build static export
cd web && npm run lint                 # ESLint
cd web && npx tsc --noEmit             # Type check

# Deploy
gcloud builds submit --config deploy/cloudbuild.yaml
docker build -f deploy/Dockerfile -t missless .
docker run -p 18080:18080 --env-file .env missless
```

## KNOWN ISSUES (as of 2026-02-25)

| # | Severity | Summary |
|---|----------|---------|
| #51 | CRITICAL | Container Registry vs Artifact Registry image path mismatch |
| #52 | CRITICAL | go.mod Go 1.24 vs CI/Dockerfile Go 1.22 version mismatch |
| #53 | HIGH | proxy.go Close() bare goroutine — SafeGo not used |

| #55 | MEDIUM | V7 anti-pattern `session.Receive(ctx)` description inaccurate |
| #56 | MEDIUM | PRE-01~PRE-12 prerequisite tasks have no GitHub issues |
| #57 | HIGH | Dockerfile healthcheck port (8080) vs default port (18080) mismatch |

## SPRINT

| Phase | Dates | Milestone |
|-------|-------|-----------|
| PRE | 2/24 | GCP setup, API keys, all prerequisites |
| P1 | 2/25-3/2 | Go scaffolding -> WebSocket proxy -> Live API -> PWA -> Progressive image |
| P2 | 3/3-3/8 | OAuth -> YouTube analysis -> Sequential Agent -> SessionManager |
| P3 | 3/9-3/12 | Reunion session -> Character consistency -> BGM -> Album -> E2E |
| P4 | 3/13-3/16 | Cloud Run deploy -> Demo video -> DevPost submit |
| **Deadline** | **3/16 17:00 PDT** | **(KST 3/17 09:00)** |

## NOTES

- **Demo video > working app** — Judges may evaluate solely from video + description
- **51 Go files** (26 source + 25 tests), **13 frontend files** (3 pages + 8 components + 2 hooks)
- **Bonus points**: dev.to 4-post series (+0.6), cloudbuild.yaml (+0.2), GDG membership (+0.2)
- **Credit deadline**: 2026-03-13 12:00 PM PT
