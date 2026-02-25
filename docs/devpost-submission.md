# missless — DevPost Submission Text

> For the Gemini Live Agent Challenge 2026 (Creative Storyteller Track)

---

## Inspiration

We all carry someone in our hearts — a grandparent who passed away before we could say goodbye, a parent we lost touch with, a friend who moved to the other side of the world. We built **missless** because we believe AI can do more than answer questions — it can help people heal.

The idea came from a simple but powerful question: *What if you could hear your loved one's voice again, not from a recording, but in a real conversation?* Not a chatbot pretending to be them, but an AI that truly understood how they spoke, what they cared about, and the warmth they carried.

## What it does

**missless** creates a virtual reunion experience. Users provide a YouTube video of someone they miss — a family vlog, an interview, a memorial video — and our AI analyzes the person's speech patterns, personality traits, and emotional expressions. It then builds a realistic persona and initiates a real-time voice conversation where users can talk to their loved one again.

During the reunion, the AI generates contextual scene images (a kitchen where you used to cook together, the park where you played as a child) and plays background music that matches the emotional tone of the conversation. The entire experience culminates in a shareable album — a keepsake of the reunion.

### Key Features
- **Voice-first interaction** — 100% voice-based, no text input required
- **YouTube video analysis** — Gemini 2.5 Pro analyzes personality directly from video URLs (zero-download)
- **30 HD preset voices** — Automatic voice matching based on analyzed characteristics
- **Native interleaved output** — Text narration + illustration generated together from a single Gemini call, creating seamless story pages
- **Progressive image generation** — Flash preview in 1-3 seconds, Imagen 4 HD final in 8-12 seconds
- **Real-time BGM** — Background music that adapts to conversation mood
- **Shareable album** — Reunion scenes compiled into a keepsake

## How we built it

### Architecture
The browser (Next.js 15 PWA) serves as a pure renderer — all AI orchestration happens server-side in Go on Cloud Run. A persistent WebSocket carries bidirectional PCM audio between the browser and our Go backend, which proxies to Gemini's Live API.

### Gemini Models (4 models)
1. **Live API** (`gemini-2.5-flash-native-audio`) — Real-time voice conversation with native audio streaming
2. **Gemini 2.5 Pro** — YouTube video analysis for personality extraction
3. **Gemini 2.5 Flash Image** — Quick scene previews (1-3s)
4. **Imagen 4** — High-quality final scene images (8-12s)

### Server-Side Tools (6 tools)
The Live API session uses Tool Calling to trigger server-side actions:
- `generate_scene` — 2-stage progressive image generation
- `generate_story_page` — **Native interleaved output** (text narration + illustration from a single Gemini call)
- `change_atmosphere` — BGM selection and crossfade
- `recall_memory` — Search persona memories in Firestore for grounded conversation
- `analyze_user` — Flash Vision analysis of user input
- `end_reunion` — Compile scenes into a shareable album

### Tech Stack
- **Backend**: Go 1.25+ on Cloud Run
- **Frontend**: Next.js 15 (PWA, static export)
- **AI SDK**: google.golang.org/genai v1.47.0
- **Database**: Cloud Firestore (sessions, personas, memories)
- **Storage**: Cloud Storage (BGM presets, generated assets)
- **Auth**: Google OAuth 2.0 (YouTube access)

## Challenges we ran into

1. **Live API bidirectional streaming in Go** — The Gemini Live API's WebSocket protocol required careful goroutine management. We implemented a dual-channel proxy pattern where the browser WebSocket and Gemini WebSocket run independently with coordinated shutdown.

2. **Progressive image generation** — Users can't wait 8-12 seconds for an image during a live conversation. We solved this with a 2-stage pipeline: a Flash model generates a quick preview (1-3s) while Imagen 4 produces the final HD version in the background.

3. **GoAway signals and session resumption** — Live API sessions can be interrupted by GoAway signals. We implemented automatic reconnection with session resumption tokens to maintain conversation continuity.

4. **Goroutine safety** — With multiple concurrent operations (audio streaming, image generation, memory search), we enforced strict safety patterns: `SafeGo()` wrappers for panic recovery, 6-level lock ordering to prevent deadlocks, and mandatory race detector testing.

5. **Voice matching accuracy** — Mapping analyzed personality traits to the right voice from 30 presets required careful calibration of age, gender, tone, and emotional warmth parameters.

## Accomplishments that we're proud of

- **Zero-download video analysis** — Gemini 2.5 Pro analyzes YouTube videos directly from URLs without downloading, making the onboarding flow instant
- **Sub-3-second first image** — Progressive rendering ensures users see scene illustrations within seconds, not minutes
- **Production Go backend** — Strict safety patterns (SafeGo, lock ordering, race detector) make the concurrent system reliable under real-world conditions
- **Emotional impact** — Early testers described the experience as "genuinely moving" — our goal was healing through technology, and it works

## What we learned

- Gemini's Live API is remarkably capable for real-time voice applications, but managing the bidirectional streaming lifecycle (especially GoAway signals) requires robust engineering
- Tool Calling through the Live API enables powerful server-side orchestration — the AI naturally decides when to generate images, change music, or recall memories
- Progressive rendering is essential for real-time AI experiences — perceived speed matters more than actual speed
- The interleaved output capability (text + image from a single model call) creates uniquely cohesive narratives

## What's next for missless

- **Voice cloning** — Replace preset voices with actual voice synthesis from video analysis
- **Multi-person reunion** — Support conversations with multiple personas simultaneously
- **Lyria BGM** — Replace preset BGM files with real-time AI-generated background music when Go SDK support becomes available
- **Extended memory** — Cross-session memory so the AI remembers previous reunions
- **Mobile app** — Full native app with offline album viewing
