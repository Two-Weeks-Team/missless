# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**missless.co** is a "Virtual Reunion" AI experience for the Gemini Live Agent Challenge 2026 (Creative Storyteller track). It analyzes YouTube videos to build personality profiles and creates interactive voice conversations with realistic personas, featuring real-time image generation and background music.

- **Hackathon**: Gemini Live Agent Challenge 2026
- **Track**: Creative Storyteller (requires Gemini's interleaved/mixed output)
- **Deadline**: March 16, 2026, 17:00 PDT
- **GitHub**: https://github.com/Two-Weeks-Team/missless (public)
- **Deployment**: missless.co (Google Cloud Run)

## Architecture

```
Browser (Next.js 15 PWA — pure renderer)
    │ WebSocket (audio PCM bidirectional + image/event)
    │ HTTPS (upload, healthcheck)
    ▼
Go Backend (Cloud Run)
    ├── SessionManager (state machine: onboarding → analyzing → reunion → album)
    ├── Live API Proxy (bidirectional streaming to Gemini)
    ├── Tool Executor (5 server-side tools)
    ├── Onboarding Pipeline (Sequential Agent: VideoAnalyzer → VoiceMatcher)
    ├── Album Generator
    └── Tool Monitor (slog)
    ▼
External Services:
    ├── Gemini API (5 models)
    ├── YouTube Data API v3
    ├── Cloud Storage (BGM presets + generated assets)
    ├── Cloud Firestore (sessions + personas + memories)
    └── Vertex AI (auth + streaming)
```

The browser is a **pure renderer** — all AI orchestration happens server-side in Go. The browser WebSocket stays connected while the Go backend swaps Live API sessions (onboarding → reunion) underneath.

### Key Design Patterns

- **Dual-channel proxy**: Browser ↔ Go WebSocket ↔ Gemini Live API (separate connections)
- **Progressive Rendering**: 2-stage image generation (Flash preview 1-3s → Imagen 4 final 8-12s)
- **SafeGo**: All goroutines wrapped with panic recovery (`util.SafeGo()`)
- **Lock Ordering**: Manager(L1) → Proxy(L2) → ToolHandler(L3) → CharacterAnchor(L4) → AlbumGenerator(L5) → MemoryStore(L6)
- **Session Resumption**: Auto-reconnect on Live API GoAway signals

## Project Structure

```
missless/
├── cmd/server/main.go              # Go entry point
├── internal/
│   ├── config/config.go            # Environment variables + validation
│   ├── session/manager.go          # SessionManager (core state machine)
│   ├── session/state.go            # Session state definitions
│   ├── live/proxy.go               # Live API WebSocket proxy
│   ├── live/tools.go               # Tool handlers (5 tools)
│   ├── live/reconnect.go           # GoAway + Session Resumption
│   ├── onboarding/pipeline.go      # Sequential Agent (VideoAnalyzer → VoiceMatcher)
│   ├── scene/generator.go          # 2-stage image generation
│   ├── scene/anchor.go             # CharacterAnchor (visual consistency)
│   ├── scene/album.go              # Album generation
│   ├── media/youtube.go            # YouTube Data API client
│   ├── auth/oauth.go               # Google OAuth 2.0
│   ├── store/firestore.go          # Firestore session store
│   ├── memory/store.go             # Memory CRUD (recall_memory tool)
│   ├── handler/websocket.go        # Browser WebSocket handler
│   ├── middleware/recovery.go      # Panic recovery middleware
│   ├── middleware/logging.go       # slog structured logging
│   ├── util/safego.go              # SafeGo goroutine wrapper
│   └── retry/backoff.go            # Exponential backoff + jitter
├── web/                             # Next.js 15 PWA (static export)
├── deploy/
│   ├── Dockerfile
│   ├── cloudbuild.yaml
│   └── terraform/
├── plan/v7/                         # Implementation plan (8 docs)
├── go.mod
└── go.sum
```

## Tech Stack

| Layer | Technology | Version |
|-------|-----------|---------|
| Backend | Go | 1.22+ |
| Frontend | Next.js (PWA, static export) | 15 |
| AI SDK | google.golang.org/genai | v1.47.0 |
| Agent SDK | google.golang.org/adk | v0.5.0 |
| WebSocket | gorilla/websocket | v1.5.3 |
| Database | Cloud Firestore | v1.17.0 |
| Storage | Cloud Storage | v1.47.0 |
| Auth | golang.org/x/oauth2 | v0.25.0 |
| Deployment | Google Cloud Run | - |

### AI Models

| Purpose | Model ID (Developer API) | Model ID (Vertex AI) |
|---------|-------------------------|---------------------|
| Live Voice (onboarding) | `gemini-2.5-flash-native-audio-preview-12-2025` | `gemini-live-2.5-flash-native-audio` |
| Live Voice (reunion) | same | same |
| Quick Image (flash) | `gemini-2.5-flash-image` | same |
| Pro Image | `imagen-4.0-generate-001` (Imagen 4) | same |
| Video/URL Analysis | `gemini-2.5-pro` | same |
| BGM | Preset files (Lyria Go SDK not available) | - |

## Build & Development Commands

```bash
# Go backend
go run cmd/server/main.go           # Run dev server
go build -o ./bin/server cmd/server/main.go  # Build binary
go test ./...                        # Run all tests
go test -race ./...                  # Run tests with race detector (required for PRs)
go test -race -count=1 ./...         # CI test command

# Linting & analysis
go vet ./...
staticcheck ./...
gofmt -w .

# Frontend (Next.js PWA)
cd web && npm install
cd web && npm run dev                # Dev mode
cd web && npm run build              # Static export

# Docker
docker build -t missless .

# Deployment
gcloud run deploy missless --source . --region asia-northeast3

# Profiling (dev)
go tool pprof http://localhost:8080/debug/pprof/heap
go tool pprof http://localhost:8080/debug/pprof/goroutine
```

## Go Module Path

```
github.com/Two-Weeks-Team/missless
```

## Environment Variables

See `.env.example` for all required variables. Key ones:
- `GCP_PROJECT_ID` — GCP project
- `GEMINI_API_KEY` — Developer API key (or use service account for Vertex AI)
- `GOOGLE_APPLICATION_CREDENTIALS` — Service account JSON path (local dev only)
- `YOUTUBE_CLIENT_ID` / `YOUTUBE_CLIENT_SECRET` — OAuth
- `STORAGE_BUCKET` — Cloud Storage bucket name
- `PORT` — Server port (default 8080)

## Go Safety Rules (CRITICAL)

All Go code must follow the patterns in `plan/v7/06-GO-SAFETY.md`. Key rules:

1. **Never use bare `go func()`** — always use `util.SafeGo()` or `util.SafeGoWithContext()` for panic recovery
2. **Never use `panic()` directly** — except in `config` init validation
3. **Always run `go test -race`** — race detector is mandatory
4. **Lock ordering is strict**: Manager → Proxy → ToolHandler → CharacterAnchor → AlbumGenerator → MemoryStore. Never acquire a higher-level lock while holding a lower-level one
5. **No I/O under locks** — do I/O outside the critical section, then lock only to update state
6. **Buffered channels for goroutine results** — always `make(chan T, 1)` minimum
7. **All channel operations use `select` + `ctx.Done()`** — prevent goroutine leaks
8. **Context propagation** — first parameter is always `ctx context.Context`, derive child contexts with `context.WithTimeout`, always `defer cancel()`
9. **`context.Background()` only in**: `main()` and graceful shutdown
10. **Cloud Run graceful shutdown**: parallel, must complete within 8 seconds (SIGTERM grace period is 10s)

## Server-Side Tools (5 tools registered with Live API)

| Tool | Function | Description |
|------|----------|-------------|
| `generate_scene` | Progressive 2-stage rendering | Flash preview (1-3s) → Imagen 4 final (8-12s) |
| `change_atmosphere` | BGM selection | Preset BGM from Cloud Storage + crossfade |
| `recall_memory` | Firestore search | Retrieve persona memories for grounded conversation |
| `analyze_user` | Flash Vision | Analyze user's visual/audio input |
| `end_reunion` | Album generation | Compile scenes into shareable album |

## Planning Documents

Detailed implementation specs are in `plan/v7/`:
- `00-INDEX.md` — Architecture, model matrix, timeline, scoring criteria
- `01-PREREQUISITES.md` — GCP setup checklist (PRE-01~12)
- `02-PHASE1-INFRA-LIVE.md` — Infrastructure + Live API proxy (T01~T06)
- `03-PHASE2-ONBOARDING.md` — YouTube analysis + persona generation (T07~T12)
- `04-PHASE3-REUNION.md` — Reunion experience engine (T13~T19)
- `05-PHASE4-DEPLOY-DEMO.md` — Deployment + demo + submission (T20~T24)
- `06-GO-SAFETY.md` — Go safety patterns (panic, race, deadlock, leak prevention)
- `07-FUTURE-DEV.md` — Post-MVP features
- `08-TRACK-ASSESSMENT.md` — Track scoring strategy

Always consult the relevant plan document before implementing a task.

## API Constraints

- Gemini Tier 1: 150 RPM — use `retry.WithBackoff` + semaphore
- Live API: GoAway signals require Session Resumption
- YouTube: Only **public** videos can be analyzed via URL (unlisted/private → gallery fallback)
- Imagen 4: Safety filtering applies — use silhouette/watercolor style to avoid blocks
- Cloud Run: 512Mi default memory — offload images to Cloud Storage
- Lyria (music): WebSocket-only API, Go SDK has no support — use preset BGM files

## Firestore Collections

```
sessions/{sessionId}     — OAuth tokens, persona, state machine, transcripts
personas/{personaId}     — Persona profiles with memories subcollection
  └─ memories/{memoryId} — Topic, description, source
albums/{albumId}         — Generated album with image URLs
```
