# Phase 4: 배포 + 데모 + 제출 (T20~T24)

> D-4 ~ D-0 | 4일 | 핵심 마일스톤: Cloud Run 라이브 + 데모 영상 3:50 + DevPost 제출

---

## T20. Cloud Run 배포 + 도메인 + SSL

**일수**: 0.5 | **난이도**: 중 | **의존성**: T19

### Dockerfile (멀티스테이지)

```dockerfile
# Stage 1: Go 빌드
FROM golang:1.26-alpine AS builder
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

### 데모 시나리오 스크립트

```
[0:00~0:15] BLACK + "마지막으로 그 사람과 대화한 게 언제예요?"
[0:15~0:25] 스마트폰 들고 있는 장면 → missless.co 로고
[0:25~0:55] 온보딩 시연
  - Google 로그인 → YouTube 그리드 (공개✅ 표시)
  - 영상 터치 → "Gemini API가 YouTube URL 직접 분석 (다운로드 ZERO!)"
  - 인물 선택 → 분석 진행 → 하이라이트 발견
  - 텍스트 입력 없이 전부 음성 + 터치
[0:55~2:30] 가상 재회 (핵심)
  - AI가 먼저 인사 (페르소나 HD 음성)
  - 장면 이미지 실시간 생성 (preview→final 크로스페이드)
  - 자연스러운 대화 + 추억 회상
  - 사용자 감정 반응 (Affective Dialog)
  - BGM 자동 전환
[2:30~3:05] 기술 설명
  - 아키텍처 다이어그램 (Go 백엔드 프록시 + 5모델)
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

### 체크포인트

- [ ] 영상 길이 3:50 이내
- [ ] 실제 동작하는 앱 시연 (목업 아님)
- [ ] 실제 감정 반응 포함
- [ ] 아키텍처 다이어그램 포함
- [ ] 해상도 1080p 이상

---

## T23. README + 아키텍처 다이어그램

**일수**: 0.5 | **난이도**: 하 | **의존성**: T20

### README 필수 항목

```markdown
# missless.co — 그리움을 줄여드립니다

## Quick Start (spin-up guide)

### Prerequisites
- Go 1.26+
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
1. **VideoAnalyzer**: Gemini 3.1 Pro가 YouTube URL로 영상 분석 → 인물 메타데이터 추출
2. **VoiceMatcher**: 분석 결과 → 30개 HD 프리셋 음성 중 최적 매핑

## Tech Stack
- Go 1.26 + google.golang.org/genai v1.29+ + ADK v0.4.0
- Next.js 15 (Static Export → Go FileServer 서빙)
- Gemini API (5모델 + YouTube URL 직접 분석 + Lyria RealTime BGM)
- Cloud Run + Firestore (세션 스토어) + Cloud Storage
- 재회 세션: 300초 제한 + 계속하기 지원
```

### 체크포인트

- [ ] GitHub 공개 레포지토리 생성
- [ ] README에 spin-up 가이드 포함
- [ ] 아키텍처 다이어그램 이미지 포함
- [ ] Terraform 파일 포함 (시간 여유 시)

---

## T24. DevPost 제출

**일수**: 0.5 | **난이도**: 하 | **의존성**: T22, T23

### 제출 체크리스트

**필수 항목**:
- [ ] 데모 영상 4분 미만 (YouTube 또는 직접 업로드)
- [ ] GitHub 레포 URL (공개)
- [ ] 라이브 URL (missless.co)
- [ ] 프로젝트 설명 (영문)
- [ ] 사용한 Google 기술 목록
- [ ] 팀 정보

**기술 스택 명시**:
- [ ] Gemini API (5개 모델 + YouTube URL 직접 분석)
- [ ] Gemini Live API (gemini-live-2.5-flash-native-audio, 30개 HD 프리셋 음성)
- [ ] Gemini Lyria RealTime (BGM 생성)
- [ ] google.golang.org/genai SDK
- [ ] ADK v0.4.0
- [ ] YouTube Data API v3
- [ ] Google OAuth 2.0
- [ ] Cloud Run + Firestore (세션 스토어) + Cloud Storage

### 체크포인트

- [ ] DevPost 제출 페이지 모든 필드 작성
- [ ] **KST 3/17 09:00 이전 제출 완료**
