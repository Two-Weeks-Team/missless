# Phase 4: 배포 + 데모 + 제출 (T20~T24)

> D-4 ~ D-0 | 4일 | 핵심 마일스톤: Cloud Run 라이브 + 데모 영상 3:50 + DevPost 제출

---

## T20. Cloud Run 배포 + 도메인 + SSL

**일수**: 0.5 | **난이도**: 중 | **의존성**: T19

### Dockerfile (멀티스테이지)

```dockerfile
# Stage 1: Go 빌드
FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /server ./cmd/server

# Stage 2: Next.js Static Export 빌드
FROM node:20-alpine AS frontend
WORKDIR /app
COPY web/ .
RUN npm ci && npm run build
# output: 'export' → web/out/ 디렉터리에 정적 파일 생성

# Stage 3: 실행 (Go가 정적 파일도 서빙)
FROM alpine:3.20
RUN apk add --no-cache ca-certificates
COPY --from=builder /server /server
COPY --from=frontend /app/out /web/out
EXPOSE 8080
CMD ["/server"]
```

### 배포 명령

```bash
# 수동 배포 (Terraform은 시간 여유 시)
gcloud run deploy missless \
  --source . \
  --region asia-northeast3 \
  --allow-unauthenticated \
  --min-instances 1 \
  --max-instances 3 \
  --memory 512Mi \
  --timeout 300 \
  --set-env-vars "GEMINI_API_KEY=$GEMINI_API_KEY,GCP_PROJECT_ID=$PROJECT_ID"

# 도메인 매핑
gcloud run domain-mappings create \
  --service missless \
  --domain missless.co \
  --region asia-northeast3
```

### 체크포인트

- [ ] `missless.co` HTTPS 접속 성공
- [ ] WebSocket `wss://missless.co/ws` 연결 성공
- [ ] Cloud Run 콘솔 스크린샷 (배포 증거)
- [ ] min-instances=1 → cold start 없음 확인
- [ ] 환경변수 (API 키 등) 안전하게 설정

---

## T21. 데모 시나리오 준비 (자체 영상 + 캐싱)

**일수**: 1 | **난이도**: 중 | **의존성**: T19, T20

### 사전 준비 전략

1. **YouTube 영상 자체 제작 (D-7 이전)**
   - 지인과 3-5분 일상 대화 영상 촬영
   - 장난, 감성적 순간, 특징적 표현 포함
   - YouTube에 일부공개(Unlisted)로 업로드
   - 출연자 동의서 확보

2. **Go 백엔드 사전 캐싱**

```go
// 데모 전 실행: 분석 결과를 Firestore에 캐싱
func PrewarmDemo(ctx context.Context, pipeline *onboarding.Pipeline, store *memory.Store) error {
    videos := []onboarding.VideoInput{
        {YouTubeURL: "https://youtube.com/watch?v=DEMO_VIDEO_ID"},
    }

    persona, err := pipeline.Run(ctx, videos, "민수", func(step string, pct int, hl string) {
        log.Printf("[prewarm] %s %d%% %s", step, pct, hl)
    })
    if err != nil {
        return err
    }

    // Firestore에 페르소나 저장
    return store.SavePersona(ctx, "demo-persona", persona)
}
```

3. **데모 시연 시 캐시 히트**
   - 영상 선택 → 분석 UI는 보여주되 실제로는 Firestore에서 캐시 로드
   - 분석 대기 시간 10초로 단축 (자연스러운 UX 유지)

### 데모 시나리오 스크립트 (3:50)

```
[0:00~0:15] BLACK + "마지막으로 그 사람과 대화한 게 언제예요?"
[0:15~0:25] 스마트폰 들고 있는 장면 → missless.co 로고
[0:25~0:55] 온보딩 시연
  - Google 로그인 → YouTube 그리드 (공개✅ 표시)
  - 영상 터치 → "Gemini API가 YouTube URL 직접 분석 (다운로드 ZERO!)"
  - 인물 선택 → 분석 진행 → 하이라이트 발견
  - 텍스트 입력 없이 전부 음성 + 터치
[0:55~2:30] 가상 재회 (핵심 — interleaved output 데모 구간)
  - AI가 먼저 인사 (페르소나 HD 음성)
  - 장면 이미지 실시간 생성 (preview→final 크로스페이드)
  - ⭐ **나레이션 캡션**: "Audio + Image + BGM generated simultaneously in one stream
    — this is Gemini's interleaved output in action"
  - 자연스러운 대화 + 추억 회상
  - 사용자 감정 반응 (Affective Dialog)
  - BGM 자동 전환
  - ⭐ 이 구간에서 음성/이미지/BGM이 **동시에** 출력되는 장면을 **최소 2회** 명확히 보여줄 것
[2:30~3:05] 기술 설명
  - 아키텍처 다이어그램 (Go 백엔드 프록시 + 5모델)
  - ⭐ "Gemini's interleaved output + Tool orchestration = fluid multimodal stream"
  - "YouTube URL → Gemini 직접 분석 (다운로드 불필요!)"
  - "Go 백엔드가 모든 API 호출 관리 — 보안 + 안정성"
[3:05~3:30] 시나리오 몽타주 (할머니, 친구, 가족)
[3:30~3:50] 클로징 — QR + "그리움을 줄여드립니다"
```

### 체크포인트

- [ ] 데모용 YouTube 영상 업로드 완료
- [ ] Firestore 사전 캐싱 완료
- [ ] 캐시 히트로 분석 10초 이내
- [ ] 리허설 3회 이상 성공

---

## T22. 데모 영상 촬영 (3:50)

**일수**: 1.5 | **난이도**: 상 | **의존성**: T21

### 촬영 가이드

- 실제 스마트폰 화면 녹화 (스크린 레코딩)
- 외부 카메라로 사용자 반응 촬영 (PIP용)
- 리허설 5회 후 최적 테이크 선택
- pro-image 결과가 좋은 테이크 우선
- 편집: 온보딩은 빠르게(30초) / 재회는 길게(90초) / 기술은 간결(35초)

### DevPost 영상 규칙 (V6 공식 규칙 반영)

- **4분 미만** (3:50 목표 → 10초 여유, **4분 초과 시 이후 부분 심사 안 함**)
- 실제 동작하는 소프트웨어 시연 필수 (목업/모션 그래픽 불가)
- YouTube 또는 Vimeo에 **공개** 업로드 (비공개/일부공개 시 심사 불가)
- 영문 또는 영문 자막 포함 (영어 심사위원 대상)
- **문제 정의(problem definition)** + **솔루션 가치 피치(value pitch)** 반드시 포함
- 심사위원은 앱을 직접 테스트하지 않을 수 있음 → **영상만으로 판단 가능해야 함**

### 데모 영상 필수 포함 요소 (채점 대응)

| 포함 요소 | 시간 | 대응 채점 기준 |
|-----------|:---:|---------------|
| Problem definition (그리움 문제 제기) | 0:00~0:15 | Demo 30% — 문제/솔루션 명확성 |
| Solution value pitch (missless 가치) | 0:15~0:25 | Demo 30% — 문제/솔루션 명확성 |
| 실제 동작 시연 (온보딩+재회) | 0:25~2:30 | Demo 30% — 실제 동작 소프트웨어 |
| interleaved output 강조 (음성+이미지+BGM 동시) | 0:55~2:30 | Innovation 40% — multimodal UX |
| 아키텍처 다이어그램 | 2:30~3:05 | Demo 30% — 다이어그램 명확성 |
| Cloud 배포 증거 (Cloud Run 콘솔) | 2:30~3:05 | Tech 30% — GCP 배포 |
| Google 기술 스택 나열 | 3:05~3:15 | Tech 30% — Google Cloud 통합 |

### 체크포인트

- [ ] 영상 길이 3:50 이내 (4분 미만 규칙)
- [ ] 실제 동작하는 앱 시연 (목업 아님)
- [ ] **문제 정의 + 솔루션 피치** 포함 (첫 25초)
- [ ] **interleaved output** (음성+이미지+BGM) 동시 생성 장면 **최소 2회** 명확히 포함
- [ ] "interleaved output" 키워드가 나레이션/캡션에 **명시적으로** 등장
- [ ] 실제 감정 반응 포함
- [ ] 아키텍처 다이어그램 포함
- [ ] Cloud Run 배포 증거 화면 포함
- [ ] 해상도 1080p 이상
- [ ] 영문 자막 또는 영문 나레이션

### ⚠️ Creative Storyteller 트랙 심사 키워드 전략

> 심사위원은 "interleaved/mixed output capabilities" 사용 여부를 핵심 지표로 판단.
> 데모 영상에서 이 단어를 **자막/나레이션으로 직접 사용**해야 함.
>
> 핵심 메시지: "Gemini's interleaved output capabilities + Tool orchestration
> = audio, images, and music flowing together in one cohesive reunion experience"
>
> 이 문구는 데모 영상 (0:55~2:30 구간) + 기술 설명 (2:30~3:05 구간) 모두에서 언급.

---

## T23. README + 아키텍처 다이어그램

**일수**: 0.5 | **난이도**: 하 | **의존성**: T20

### README 필수 항목

```markdown
# missless.co — 그리움을 줄여드립니다

## Quick Start (spin-up guide)

### Prerequisites
- Go 1.22+
- Node.js 20+
- GCP 프로젝트 + $300 크레딧

### Setup
1. Clone: `git clone https://github.com/Two-Weeks-Team/missless.git`
2. 환경변수: `cp .env.example .env` → API 키 설정
3. Go: `go mod download`
4. Frontend: `cd web && npm install && npm run build` (Static Export)
5. Run: `go run cmd/server/main.go`

### Deploy to Cloud Run
gcloud run deploy missless --source .

## Architecture
[아키텍처 다이어그램 이미지]

### Sequential Agent Flow
1. **VideoAnalyzer**: Gemini 2.5 Pro가 YouTube URL로 영상 분석 → 인물 메타데이터 추출
2. **VoiceMatcher**: 분석 결과 → 30개 HD 프리셋 음성 중 최적 매핑

## Tech Stack
- Go 1.22 + google.golang.org/genai v1.47.0 + ADK v0.5.0
- Next.js 15 (Static Export → Go FileServer 서빙)
- Gemini API (5모델 + YouTube URL 직접 분석 + Lyria RealTime BGM)
- Cloud Run + Firestore (세션 스토어) + Cloud Storage
- 재회 세션: 300초 제한 + 계속하기 지원
```

### 체크포인트

- [ ] GitHub 공개 레포지토리 생성 (Two-Weeks-Team/missless)
- [ ] README에 spin-up 가이드 포함
- [ ] 아키텍처 다이어그램 이미지 포함
- [ ] cloudbuild.yaml 또는 Terraform 파일 포함 (보너스 +0.2)

---

## T24. DevPost 제출

**일수**: 0.5 | **난이도**: 하 | **의존성**: T22, T23

### 제출 체크리스트 (V6 공식 규칙 완전판)

**필수 항목**:
- [ ] 데모 영상 4분 미만 (YouTube 또는 Vimeo **공개** — 비공개 시 심사 불가)
- [ ] 영문 또는 영문 자막 포함
- [ ] 실제 동작 소프트웨어 시연 (목업/모션 그래픽 불가)
- [ ] **문제 정의 + 솔루션 가치 피치** 포함 (데모 영상 내)
- [ ] GitHub 레포 URL (공개, 2026-02-16 이후 생성)
- [ ] 라이브 URL (missless.co) + 테스트 크리덴셜 (필요 시)
- [ ] Cloud 배포 증거 (데모 영상과 **별도** 스크린 레코딩 또는 코드 링크)
- [ ] 아키텍처 다이어그램 (이미지 캐러셀 또는 파일 업로드)
- [ ] 프로젝트 텍스트 설명 (영문 — 아래 템플릿 참고)
- [ ] **카테고리 선택**: Creative Storyteller
- [ ] 사용한 Google 기술 목록 명시
- [ ] **서드파티 통합 선언** (gorilla/websocket, Next.js 등)
- [ ] 팀 정보 (대표자 지정)

**기술 스택 명시**:
- [ ] Gemini API (5개 모델 + YouTube URL 직접 분석)
- [ ] Gemini Live API (30개 HD 프리셋 음성, Affective Dialog, Proactive Audio)
- [ ] Gemini Lyria RealTime (BGM 생성)
- [ ] google.golang.org/genai SDK v1.47.0
- [ ] ADK v0.5.0
- [ ] YouTube Data API v3
- [ ] Google OAuth 2.0
- [ ] Cloud Run + Firestore (세션 스토어) + Cloud Storage

**보너스 점수 항목 (최대 +1.0점)**:
- [ ] **dev.to 블로그 시리즈 (4편)** → **+0.6**
  - [ ] Post #1: "Real-Time Audio Proxy with Go + Gemini Live API" (3/2 게시)
  - [ ] Post #2: "YouTube Video Analysis → AI Persona with Sequential Agents" (3/8 게시)
  - [ ] Post #3: "Immersive Reunion Experiences with Interleaved Output" (3/12 게시)
  - [ ] Post #4: "missless.co — Complete Architecture & Lessons Learned" (3/14-15 게시, 메인)
  - 모든 포스트에 `#GeminiLiveAgentChallenge` 태그
  - 모든 포스트에 **필수 문구**: "created for purposes of entering this hackathon"
  - DevPost에 **4개 URL 모두 등록**
  - 각 포스트 Twitter/LinkedIn 공유
- [ ] `cloudbuild.yaml` + Terraform → **+0.2** (공개 repo에 포함)
- [ ] GDG 공개 프로필 링크 → **+0.2**

### DevPost 텍스트 설명 템플릿 (영문)

```
## What it does
missless.co lets you virtually reconnect with someone you miss.
Upload YouTube videos of a loved one, and our AI analyzes their personality,
speech patterns, and memorable moments to create an immersive audio-visual
reunion experience — **seamlessly weaving together voice, images, and music
in a single fluid stream using Gemini's interleaved output capabilities**.

Zero text input required. The entire experience flows through real-time
voice conversation, AI-generated scene images, and emotionally adaptive
background music — all delivered simultaneously as one cohesive narrative.

## How we built it (Creative Storyteller Track)
- **Gemini's interleaved/mixed output**: Live API voice + Tool-triggered image generation
  + Lyria BGM create a seamless multimodal storytelling stream
- **Go backend proxy** orchestrating 5 Gemini models via WebSocket
- **Gemini Live API** with 30 HD preset voices for persona-matched conversation
- **YouTube URL direct analysis** (no video download) via Gemini 2.5 Pro
- **Progressive image rendering**: Flash preview (1-3s) → Pro final (8-12s) with crossfade
- **Lyria RealTime BGM** for emotionally adaptive background music
- **Sequential Agent pipeline**: VideoAnalyzer → VoiceMatcher (ADK v0.5.0)
- Deployed on **Cloud Run** with Firestore session store and Cloud Storage

## Google technologies used
- Gemini API (gemini-2.5-flash-native-audio, gemini-2.5-flash-preview-image-generation,
  imagen-3.0-generate-002, gemini-2.5-pro-preview-03-25, models/lyria-realtime-exp)
- Google GenAI SDK (google.golang.org/genai v1.47.0)
- Agent Development Kit (ADK v0.5.0)
- Gemini Live API (Affective Dialog, Proactive Audio, 30 HD voices)
- Cloud Run, Firestore, Cloud Storage
- YouTube Data API v3, Google OAuth 2.0

## Challenges we ran into
[개발 중 겪은 도전과 해결 방법 기술]

## What we learned
[학습한 내용 — Gemini Live API, Go concurrency 등]

## What's next
Voice cloning integration, multi-person reunion, vector-based memory search,
multilingual support, offline album viewing via PWA.

## Third-party integrations
- gorilla/websocket (Go WebSocket library, MIT License)
- Next.js 15 (React framework, MIT License)
```

### GCP 크레딧 신청

> ⚠️ **Google Cloud 크레딧 신청 마감: 2026-03-13 12:00 PM PT (KST 3/14 05:00)**
> - DevPost 리소스 페이지에서 신청 가능
> - 마감 전 반드시 신청 완료

### 마감 시간

**PDT 2026-03-16 17:00 = KST 2026-03-17 09:00**

> ⚠️ Late submissions will NOT be accepted per DevPost official rules.
> 최소 D-1(3/15)까지 제출 초안 완료 → D-0(3/16)에 최종 확인 후 제출 권장.

### 주요 마감 일정 요약

| 마감 | 날짜 (PT) | 날짜 (KST) |
|------|-----------|------------|
| GCP 크레딧 신청 | 2026-03-13 12:00 PM | 2026-03-14 05:00 |
| 제출 마감 | 2026-03-16 17:00 | 2026-03-17 09:00 |
| 심사 기간 | 2026-03-17 ~ 04-03 | — |
| 수상자 선정 | ~2026-04-08 | — |
| 공식 발표 | 2026-04-22~24 (Google NEXT) | — |

### 체크포인트

- [ ] DevPost 제출 페이지 모든 필드 작성
- [ ] 영문 텍스트 설명 완성 (위 템플릿 기반)
- [ ] 배포 증거 별도 스크린 레코딩 업로드
- [ ] 서드파티 통합 선언 작성
- [ ] 카테고리 "Creative Storyteller" 선택
- [ ] 라이브 URL + 테스팅 접근 정보 기입
- [ ] GCP 크레딧 신청 완료 (3/13 전)
- [ ] **KST 3/17 09:00 이전 제출 완료**
