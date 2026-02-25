# missless.co V7 — 최종 구현 계획서 (Phase 분할)

> Gemini Live Agent Challenge 2026 | Creative Storyteller 트랙
> 작성일: 2026-02-24 | 마감일: 2026-03-16 17:00 PDT (KST 3/17 09:00)
> GitHub: https://github.com/Two-Weeks-Team/missless (public)
> V6 → V7 핵심 변경: 모델명 GA 업데이트 + Lyria→프리셋 BGM 전략 + YouTube 공개정책 수정 + Go SDK 시그니처 수정 + 상금 구조 반영

---

## V6 → V7 변경 이력

| 항목 | V6 | V7 | 이유 / 출처 |
|------|----|----|-------------|
| Flash Image 모델명 | `gemini-2.5-flash-preview-image-generation` | **`gemini-2.5-flash-image`** (stable) | [공식 deprecation 2026-01-15](https://ai.google.dev/gemini-api/docs/image-generation) |
| Pro 모델명 | `gemini-2.5-pro-preview-03-25` | **`gemini-2.5-pro`** (GA) | [Gemini 모델 페이지](https://ai.google.dev/gemini-api/docs/models) — GA 별칭 사용 권장 |
| BGM 전략 | Lyria `GenerateContent` 호출 | **프리셋 BGM** (Cloud Storage) | Lyria는 WebSocket 전용 (`client.aio.live.music.connect()`) — [Go SDK에 Lyria 미지원](https://pkg.go.dev/google.golang.org/genai) |
| YouTube 분석 범위 | 공개+일부공개 | **공개(public)만** | [YouTube Data API](https://developers.google.com/youtube/v3/docs/videos) — unlisted URL은 Gemini FileData로 분석 불가 |
| Go SDK 시그니처 | `session.Receive(ctx)` 등 | `session.Receive()` 등 | [genai v1.47.0 GoDoc](https://pkg.go.dev/google.golang.org/genai#Session) — `Session.Receive()`는 처음부터 ctx 파라미터 없음 |
| Grand Prize 상금 | $50,000 | **$25,000** + 크레딧/티켓/여행비 | [DevPost 공식 규칙](https://geminiliveagentchallenge.devpost.com/rules) |
| Category Winner | 미기재 | **$10,000** (트랙당 1팀) | DevPost 공식 규칙 |
| Honorable Mentions | 미기재 | **5×$2,000** | DevPost 공식 규칙 |
| Imagen 모델 | `imagen-3.0-generate-002` (shut down) | **`imagen-4.0-generate-001`** (Imagen 4) | [Imagen 3 shut down 2025-11-10](https://ai.google.dev/gemini-api/docs/deprecations) |
| Subcategory Winner | 미기재 | **$5,000** (3개 부문) | [DevPost 공식 규칙](https://geminiliveagentchallenge.devpost.com/rules) |

---

## V5 → V6 변경 이력

| 항목 | V5 | V6 | 이유 |
|------|----|----|------|
| 문서 구조 | 바로 개발 시작 | **사전준비(01-PREREQUISITES) 분리** | GCP/API 온보딩 누락 방지 |
| Live API 모델명 | `gemini-live-2.5-flash-native-audio` | **Vertex AI: `gemini-live-2.5-flash-native-audio`** / **Developer API: `gemini-2.5-flash-native-audio-preview-12-2025`** | 플랫폼별 모델명 차이 확인 |
| Image 모델명(flash) | `gemini-2.5-flash-image` (미검증) | **`gemini-2.5-flash-image`** (stable, 2026-01-21 GA) | 실제 모델명 검증 완료 |
| Image 모델명(pro) | `gemini-3-pro-image-preview` (미검증) | **`imagen-3.0-generate-002`** (V6 당시 유효) | V6 작성 시점에는 Imagen 3 유효 → V7에서 Imagen 4로 재수정 |
| 비디오 분석 모델 | `gemini-3.1-pro-preview` (미검증) | **`gemini-2.5-pro`** (GA) | 실제 모델명 검증 완료 |
| Lyria BGM 모델 | `lyria-realtime-v1` (오류) | **`models/lyria-realtime-exp`** | 실제 모델명 확인 |
| Go GenAI SDK | `v1.29.0` | **`v1.47.0`** (2026-02-19 릴리스) | 최신 버전 확인 |
| ADK Go | `v0.4.0` | **미사용 (GenAI SDK만 사용)** | go.mod에 ADK 미포함 |
| HD 프리셋 음성 | 일부만 기재 (15개) | **전체 30개 목록 + 성별/톤 특성 완비** | 공식 문서 기반 완전 목록 |
| 보이스클로닝 | 미정 | **추후개발로 분리 (MiniMax/ElevenLabs)** | 제출본 완성 후 재판단 |
| 제출 요구사항 | 간략 기재 | **DevPost 공식 규칙 전문 반영** | 누락 방지 |
| 보너스 점수 | 미기재 | **콘텐츠 제작(+0.6), 자동 배포(+0.2), GDG(+0.2)** | 가점 확보 |

---

## 전체 시스템 아키텍처 (V7)

```
브라우저 (Next.js 15 PWA — 순수 렌더러)
    │ WebSocket (WS-01: 오디오 PCM 양방향 + 이미지/이벤트 수신)
    │ HTTPS (업로드, 헬스체크)
    ▼
┌──────────────────────────────────────────────────────────┐
│                Go Backend (Cloud Run)                       │
│                                                            │
│  ┌─────────────────────────────────────────────────────┐  │
│  │              SessionManager (핵심)                    │  │
│  │                                                       │  │
│  │  Phase 1: 온보딩 Live API 세션                        │  │
│  │    - 시스템 음성 (Aoede — missless 호스트)             │  │
│  │    - 사용자 안내 + YouTube 영상 선택                   │  │
│  │                                                       │  │
│  │  Phase 2: 재회 Live API 세션                           │  │
│  │    - 페르소나 HD 음성 (30개 프리셋 중 선택)            │  │
│  │    - 전체 Tool 등록 (장면 생성, BGM, 추억 등)          │  │
│  │    - Affective Dialog + Proactive Audio                │  │
│  │                                                       │  │
│  │  브라우저 WebSocket은 유지 — Live API만 교체            │  │
│  └─────────────────────────────────────────────────────┘  │
│                                                            │
│  ┌──────────────────────────────────────────────────────┐ │
│  │         Tool Executor (서버 사이드)                    │ │
│  │                                                        │ │
│  │  generate_scene → 2단계 Progressive Rendering          │ │
│  │    Stage 1: flash-image (1-3s) → scene_preview 전송   │ │
│  │    Stage 2: pro-image (8-12s) → scene_final 전송      │ │
│  │                                                        │ │
│  │  change_atmosphere → 프리셋 BGM 선택 + 크로스페이드     │ │
│  │  recall_memory → Firestore 검색                       │ │
│  │  analyze_user → Flash Vision 분석                     │ │
│  │  end_reunion → 앨범 생성                               │ │
│  └──────────────────────────────────────────────────────┘ │
│                                                            │
│  ┌───────────────┐  ┌────────────┐  ┌────────────┐       │
│  │ Onboarding    │  │ Album      │  │ Tool       │       │
│  │ Pipeline      │  │ Generator  │  │ Monitor    │       │
│  │ (Sequential)  │  │            │  │ (slog)     │       │
│  └───────────────┘  └────────────┘  └────────────┘       │
└─────┬────────────────┬──────────────────┬────────────────┘
      ▼                ▼                  ▼
┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐
│ Gemini   │  │ YouTube  │  │ Cloud    │  │ Cloud    │
│ API      │  │ Data API │  │ Storage  │  │ Firestore│
│(5모델    │  │ v3       │  │(Fallback │  │(세션스토어│
│+URL분석) │  │(영상목록)│  │+생성물)  │  │+페르소나)│
└──────────┘  └──────────┘  └──────────┘  └──────────┘
```

---

## 사용 모델 매트릭스 (V7 검증 완료)

| 용도 | 모델 ID (Developer API) | 모델 ID (Vertex AI) | 비고 |
|------|------------------------|--------------------|----|
| Live API (온보딩) | `gemini-2.5-flash-native-audio-preview-12-2025` | `gemini-live-2.5-flash-native-audio` | 시스템 음성 Aoede |
| Live API (재회) | `gemini-2.5-flash-native-audio-preview-12-2025` | `gemini-live-2.5-flash-native-audio` | 페르소나 HD 음성 |
| 이미지 생성 (flash) | `gemini-2.5-flash-image` | 동일 | 1-3초 프리뷰 |
| 이미지 생성 (pro) | `imagen-4.0-generate-001` | 동일 | 고품질 (Imagen 4) |
| 비디오/URL 분석 | `gemini-2.5-pro` | 동일 | YouTube URL 직접 분석 |
| 페르소나 생성 | `gemini-2.5-pro` | 동일 | JSON 구조화 출력 |
| BGM 생성 | `models/lyria-realtime-exp` | 동일 | WebSocket 전용 (Go SDK 미지원) → **프리셋 BGM 사용** |

> **API 선택 결정**: Vertex AI 사용 (Cloud Run 배포 + IAM 인증 + GA 모델 사용 가능)
> Developer API는 API 키 방식이지만, Vertex AI가 Cloud 배포 심사 점수에 유리

---

## 30개 HD 프리셋 음성 목록 (완전판)

### 여성 음성 (14개)

| # | 음성 이름 | 톤/성격 | 용도 매핑 |
|---|-----------|---------|-----------|
| 1 | **Achernar** | Soft (부드러운) | 조용한 성격, 내성적 |
| 2 | **Aoede** | Breezy (산뜻한) | **시스템 온보딩 음성 (기본)** |
| 3 | **Autonoe** | Bright (밝은) | 활발한 성격 |
| 4 | **Callirrhoe** | Easy-going (편안한) | 친근한 성격 |
| 5 | **Despina** | Smooth (매끄러운) | 우아한 성격 |
| 6 | **Erinome** | Clear (맑은) | 깔끔한 성격 |
| 7 | **Gacrux** | Mature (성숙한) | 연장자, 어머니 |
| 8 | **Kore** | Firm (단호한) | 강한 성격 |
| 9 | **Laomedeia** | Upbeat (활기찬) | 에너지 넘치는 성격 |
| 10 | **Leda** | Youthful (젊은) | 10대~20대 |
| 11 | **Pulcherrima** | Forward (적극적) | 외향적 성격 |
| 12 | **Sulafat** | Warm (따뜻한) | 따뜻한 성격, 어머니 |
| 13 | **Vindemiatrix** | Gentle (온화한) | 다정한 성격 |
| 14 | **Zephyr** | Bright (밝은) | 밝고 경쾌한 성격 |

### 남성 음성 (16개)

| # | 음성 이름 | 톤/성격 | 용도 매핑 |
|---|-----------|---------|-----------|
| 15 | **Achird** | Friendly (친근한) | 친구, 동료 |
| 16 | **Algenib** | Gravelly (거친) | 거친 매력, 중년 남성 |
| 17 | **Algieba** | Smooth (매끄러운) | 세련된 성격 |
| 18 | **Alnilam** | Firm (단호한) | 강인한 성격 |
| 19 | **Charon** | Informative (정보전달) | 차분한 설명 톤 |
| 20 | **Enceladus** | Breathy (숨결있는) | 부드러운 남성 |
| 21 | **Fenrir** | Excitable (들뜬) | 에너지 넘치는 젊은 남성 |
| 22 | **Iapetus** | Clear (맑은) | 깔끔한 톤 |
| 23 | **Orus** | Firm (단호한) | 권위있는 톤 |
| 24 | **Puck** | Upbeat (활기찬) | **기본 음성**, 밝은 톤 |
| 25 | **Rasalgethi** | Informative (정보전달) | 지적인 톤 |
| 26 | **Sadachbia** | Lively (활발한) | 활기찬 성격 |
| 27 | **Sadaltager** | Knowledgeable (박식한) | 지식인 타입 |
| 28 | **Schedar** | Even (균일한) | 안정적인 톤 |
| 29 | **Umbriel** | Easy-going (편안한) | 편안한 친구 |
| 30 | **Zubenelgenubi** | Casual (캐주얼) | 격식없는 톤 |

---

## 문서 구성 (V7)

| 파일 | 내용 | 태스크 |
|------|------|--------|
| **`00-INDEX.md`** | 본 문서. 전체 구조, 아키텍처, 모델 매트릭스, Phase 개요 | — |
| **`01-PREREQUISITES.md`** | ⚠️ **개발 전 필수 사전준비** — GCP, API, 도메인, 환경변수 | PRE-01~PRE-12 |
| **`02-PHASE1-INFRA-LIVE.md`** | 인프라 + Live API 기반 + WebSocket 프록시 | T01~T06 |
| **`03-PHASE2-ONBOARDING.md`** | YouTube 분석 + 페르소나 생성 + 온보딩 UX | T07~T12 |
| **`04-PHASE3-REUNION.md`** | 재회 경험 엔진 + 이미지/BGM/앨범 | T13~T19 |
| **`05-PHASE4-DEPLOY-DEMO.md`** | 배포 + 데모 영상 + DevPost 제출 | T20~T24 |
| **`06-GO-SAFETY.md`** | Go 위험요소 분류 (Panic/Race/Bottleneck/Leak) | 전체 |
| **`07-FUTURE-DEV.md`** | 추후개발 항목 (보이스클로닝, 확장 기능) + dev.to 시리즈 전략 | — |
| **`08-TRACK-ASSESSMENT.md`** | 트랙 적합도 평가 + 수상 전략 + 점수 예측 | — |

---

## 태스크 총괄 (사전준비 12개 + 개발 24개 = 36개)

### PRE: 사전준비 (개발 시작 전 완료 필수)

| ID | 태스크 | 소요시간 | 체크포인트 |
|----|--------|:---:|------------|
| PRE-01 | GCP 프로젝트 생성 + Billing 연결 | 15분 | 프로젝트 ID 확보 |
| PRE-02 | Gemini API 활성화 + Tier 1 승격 신청 | 30분 | API 키 발급 + RPM 150+ |
| PRE-03 | YouTube Data API v3 활성화 | 10분 | API 콘솔에서 활성화 확인 |
| PRE-04 | OAuth 2.0 클라이언트 ID 생성 | 20분 | Client ID + Secret 확보 |
| PRE-05 | Cloud Firestore 데이터베이스 생성 | 10분 | Native 모드, asia-northeast3 |
| PRE-06 | Cloud Storage 버킷 생성 | 10분 | missless-assets 버킷 |
| PRE-07 | 도메인 missless.co DNS 설정 | 30분 | Cloud Run 매핑 준비 |
| PRE-08 | GitHub 레포지토리 생성 (public) | 5분 | Two-Weeks-Team/missless |
| PRE-09 | 로컬 개발환경 Go + Node.js 설치 | 20분 | go 1.22+ / node 20+ |
| PRE-10 | .env.example 작성 + 환경변수 설정 | 15분 | 모든 키 로컬 설정 완료 |
| PRE-11 | 서비스 계정 키 생성 (Vertex AI용) | 15분 | JSON 키 파일 확보 |
| PRE-12 | 모든 API 호출 검증 테스트 | 30분 | 5개 모델 + YouTube API 성공 |

### Phase 1: 인프라 + Live API 기반 (D-20 ~ D-15)

| ID | 태스크 | 난이도 | 일수 | 체크포인트 |
|----|--------|:---:|:---:|------------|
| T01 | Go scaffolding + 의존성 설정 | 하 | 0.5 | 서버 시작 + 헬스체크 |
| T02 | Go WebSocket 프록시 + Live API 연결 | 상 | 1.5 | 음성 양방향 성공 |
| T03 | Tool 등록 + 서버사이드 실행 기반 | 중 | 1 | Tool Call → 핸들러 실행 |
| T04 | Next.js PWA 순수 렌더러 | 중 | 1 | 오디오 재생 + 이미지 표시 |
| T05 | 2단계 이미지 생성 (flash→pro Progressive) | 상 | 1 | preview → final 크로스페이드 |
| T06 | Rate Limit 방어 + 에러 핸들링 기반 | 중 | 0.5 | retryWithBackoff 동작 |

### Phase 2: 온보딩 파이프라인 (D-14 ~ D-10)

| ID | 태스크 | 난이도 | 일수 | 체크포인트 |
|----|--------|:---:|:---:|------------|
| T07 | Google OAuth + YouTube 영상 목록 + 공개상태 | 중 | 1 | 영상 목록+공개상태 표시 |
| T08 | YouTube URL 직접 분석 + 갤러리 Fallback | 상 | 1.5 | URL→분석 성공 + Fallback |
| T09 | Sequential Agent 온보딩 (VideoAnalyzer→VoiceMatcher) | 중 | 1 | 30개 HD 프리셋 매핑 완료 |
| T10 | 온보딩 Live API 세션 (시스템 음성) | 중 | 1 | AI 음성 안내 동작 |
| T11 | SessionManager — 온보딩→재회 전환 | 상 | 1 | 세션 교체 2초 이내 |
| T12 | 온보딩 UX (진행률 + 하이라이트 + 인물 선택) | 중 | 1 | 실시간 카드 업데이트 |

### Phase 3: 재회 경험 엔진 (D-9 ~ D-5)

| ID | 태스크 | 난이도 | 일수 | 체크포인트 |
|----|--------|:---:|:---:|------------|
| T13 | 재회 Live API 세션 + 페르소나 System Instruction | 중 | 1 | 페르소나 성격으로 대화 |
| T14 | 캐릭터 일관성 (CharacterAnchor + 실루엣) | 중 | 1 | 3장면 일관성 유지 |
| T15 | BGM 시스템 + change_atmosphere | 중 | 0.5 | 프리셋 BGM 선택 → 브라우저 재생 |
| T16 | recall_memory + Firestore 추억 검색 | 중 | 0.5 | 관련 추억 → 대화 반영 |
| T17 | Affective Dialog + Proactive Audio 최적화 | 중 | 0.5 | 감정 인지 + 선택적 응답 |
| T18 | 앨범 생성 + 공유 카드 | 중 | 1 | 장면→앨범→OG 카드 |
| T19 | E2E 통합 테스트 | 상 | 1 | 온보딩→재회→앨범 완주 |

### Phase 4: 배포 + 데모 + 제출 (D-4 ~ D-0)

| ID | 태스크 | 난이도 | 일수 | 체크포인트 |
|----|--------|:---:|:---:|------------|
| T20 | Cloud Run 배포 + 도메인 + SSL | 중 | 0.5 | missless.co 라이브 |
| T21 | 데모 시나리오 준비 (자체 영상 + 캐싱) | 중 | 1 | YouTube 영상 + 캐시 |
| T22 | 데모 영상 촬영 (3:50) | 상 | 1.5 | 실제 감정 반응 포함 |
| T23 | README + 아키텍처 다이어그램 | 하 | 0.5 | spin-up 가이드 |
| T24 | DevPost 제출 | 하 | 0.5 | ✅ 제출 완료 |

---

## 스프린트 일정 (V7)

| 일자 | 요일 | Phase | 태스크 | 마일스톤 |
|------|:---:|:---:|--------|----------|
| 2/24 | 월 | PRE | PRE-01~PRE-12 + GDG 가입 + 크레딧 신청 | **사전준비 완료** |
| 2/25 | 화 | P1 | T01 | Go scaffolding + 의존성 |
| 2/26 | 수 | P1 | T02 | WebSocket 프록시 PoC |
| 2/27 | 목 | P1 | T02+T03 | Live API 양방향 + Tool 실행 |
| 2/28 | 금 | P1 | T04+T06 | PWA + 에러 핸들링 |
| 3/1 | 토 | P1 | T05 | 2단계 이미지 Progressive |
| 3/2 | 일 | P1 | T05 + 📝 dev.to #1 작성 | **P1 완성 — Fluid Stream 동작** + Blog #1 게시 |
| 3/3 | 월 | P2 | T07 | OAuth + YouTube 그리드 |
| 3/4 | 화 | P2 | T08 | YouTube URL 직접 분석 PoC |
| 3/5 | 수 | P2 | T09 | Sequential Agent |
| 3/6 | 목 | P2 | T10+T11 | 온보딩 세션 + SessionManager |
| 3/7 | 금 | P2 | T12 | 온보딩 UX 완성 |
| 3/8 | 토 | P2 | (버퍼) + 📝 dev.to #2 작성 | **P2 완성 — 온보딩→재회 전환** + Blog #2 게시 |
| 3/9 | 일 | P3 | T13+T14 | 재회 세션 + 캐릭터 일관성 |
| 3/10 | 월 | P3 | T15+T16 | BGM + 추억 검색 |
| 3/11 | 화 | P3 | T17+T18 | Affective Dialog + 앨범 |
| 3/12 | 수 | P3 | T19 + 📝 dev.to #3 작성 | **P3 완성 — E2E 통합 성공** + Blog #3 게시 |
| 3/13 | 목 | P4 | T20+T21 | Cloud Run 배포 + 데모 준비 |
| 3/14 | 금 | P4 | T22 | 데모 영상 촬영 |
| 3/15 | 토 | P4 | T22+T23 + 📝 dev.to #4 작성 (메인) | 영상 편집 + README + Blog #4 메인 게시(+0.6) |
| 3/16 | 일 | P4 | T24 | **✅ 최종 제출 (PDT 17:00 / KST 3/17 09:00)** |

---

## 채점 기준 대응 매트릭스 (V7 — 공식 규칙 정밀 반영)

### Stage 2: 심사 기준 (1~5점, 가중합산)

#### 1. Innovation & Multimodal UX (40%)

| 세부 기준 | V7 대응 | 점수 근거 | Phase |
|-----------|---------|-----------|:---:|
| "text box" 패러다임 탈피 | 전체 경험 텍스트 입력 ZERO. 온보딩도 음성+터치만 | 채팅 UI 없음, 완전 음성 기반 | P2 |
| 자연스럽고 몰입적 상호작용 | 이미지+음성+BGM이 실시간 동시 생성 (Progressive Rendering) | 텍스트/이미지/오디오 interleaved output | P1,P3 |
| **Creative Storyteller 트랙 실행** | YouTube 영상 → 인물 분석 → 재회 스토리 (이미지+음성+BGM 동시 생성) | **Gemini의 interleaved/mixed output** 직접 활용 | P2,P3 |
| 독특한 성격/목소리 | YouTube 분석 → 30개 HD 프리셋 음성 매핑 (메타데이터 기반) | 인물별 최적 음성 자동 선택 | P2 |
| Live & context-aware | Affective Dialog + Proactive Audio + Vision + recall_memory | 감정 인지 + 추억 기반 대화 | P3 |

#### 2. Technical Implementation & Agent Architecture (30%)

| 세부 기준 | V7 대응 | 점수 근거 | Phase |
|-----------|---------|-----------|:---:|
| Google Cloud 통합 | Cloud Run + Firestore + Storage + OAuth + YouTube API | 6개 Google Cloud 서비스 활용 | P1,P4 |
| GCP 백엔드 견고성 | Go 백엔드 프록시 (SafeGo, Lock ordering, graceful shutdown) | 06-GO-SAFETY.md 전체 패턴 적용 | 전체 |
| Agent 로직 건전성 | Sequential Agent (VideoAnalyzer→VoiceMatcher) via GenAI SDK | 2단계 파이프라인 + 5개 Tool 실행 | P2,P3 |
| 에러 핸들링 | retryWithBackoff + Rate Limit defense + Fallback 체계 | 모든 API 호출에 3중 방어 | P1 |
| 할루시네이션 방지 | YouTube URL 직접 분석 기반 그라운딩 + recall_memory | 실제 영상 데이터 기반 대화 | P2,P3 |

#### 3. Demo & Presentation (30%)

| 세부 기준 | V7 대응 | 점수 근거 | Phase |
|-----------|---------|-----------|:---:|
| **문제/솔루션 명확성** (problem definition + value pitch) | 15초 감성 인트로 "마지막으로 그 사람과 대화한 게 언제예요?" | 즉시 이해 가능한 문제 정의 | P4 |
| 아키텍처 다이어그램 명확성 | Go 프록시 → 5모델 → Cloud 서비스 연결도 | 35초 기술 설명 구간 | P4 |
| Cloud 배포 증거 | 데모 영상과 **별도** Cloud Run 콘솔 스크린 레코딩 | 독립 증거물 제출 | P4 |
| 실제 동작 소프트웨어 | 자체 촬영 영상 + 실제 감정 반응 (목업/모션 그래픽 불가) | 사전 캐싱으로 안정적 시연 | P4 |

### Stage 3: 보너스 기여 (최대 +1.0점)

| 보너스 | 점수 | V7 전략 | 필수 조건 |
|--------|:---:|---------|-----------|
| 콘텐츠 제작 | +0.6 | **dev.to 4편 시리즈** (Phase별 작성) | **"created for purposes of entering this hackathon"** 문구 필수 + `#GeminiLiveAgentChallenge` 해시태그 |
| 자동 배포 | +0.2 | `cloudbuild.yaml` + Terraform 파일을 공개 repo에 포함 | repo에 IaC 코드 존재 |
| GDG 멤버십 | +0.2 | GDG 가입 → 공개 프로필 링크 제출 | 프로필 URL 필요 |

> **점수 범위**: Stage 2 심사 1~5점 + Stage 3 보너스 최대 +1.0점 = **최종 1~6점**

### Subcategory 상 대응 전략

| Subcategory 상 | 상금 | missless 경쟁력 | 핵심 어필 포인트 |
|----------------|:---:|:---:|----------------|
| **Best Multimodal Integration & UX** | $5,000 | ⭐⭐⭐ 최강 | 음성+이미지+BGM interleaved output, 텍스트 입력 ZERO |
| **Best Technical Execution & Architecture** | $5,000 | ⭐⭐ 강 | Go 프록시 아키텍처, 5모델 통합, SafeGo 패턴 |
| **Best Innovation & Thought Leadership** | $5,000 | ⭐⭐⭐ 최강 | "가상 재회" 컨셉 자체가 혁신적 |

---

## Creative Storyteller 트랙 필수 요구사항 검증

> **공식 요구**: "Must use Gemini's interleaved/mixed output capabilities"
> "Leverage Gemini's native interleaved output to generate rich, mixed-media responses"

| 트랙 요구사항 | V7 구현 | 검증 |
|---------------|---------|:---:|
| interleaved/mixed output | Live API 음성 + Tool Call로 이미지/BGM 동시 생성 | ✅ |
| rich, mixed-media responses | 음성(HD 프리셋) + 이미지(Progressive) + BGM(프리셋) 실시간 | ✅ |
| 텍스트+이미지+오디오+비디오 seamless | 대화 중 장면 이미지 자동 생성 + 크로스페이드 + BGM 전환 | ✅ |
| 예시: Interactive storybooks | "가상 재회" = 추억 기반 인터랙티브 스토리텔링 | ✅ |

### 트랙 선택 검증 결과 (2026-02-24 검증)

> **결론: ✅ Creative Storyteller가 최적 트랙. Live Agents 대비 경쟁 우위 확보.**

| 비교 항목 | Creative Storyteller | Live Agents | 판정 |
|-----------|:---:|:---:|:---:|
| 필수 기술 일치도 | ⭐ interleaved output **정확히** 구현 | Live API 사용하지만 barge-in은 핵심 아님 | CS 우세 |
| Innovation 40% 채점 | "media interleaving" 직접 대응 | "interruption handling" 대응 | CS 정확 일치 |
| 경쟁 밀도 (추정) | 낮음 (기술 난이도 높음) | 높음 (단순 챗봇/번역기 다수) | CS 우세 |
| 데모 임팩트 | 시각+청각+음악 = 감성적 | 음성 대화만 | CS 우세 |
| Subcategory 이중 기회 | "Best Multimodal UX" 자동 어필 | 제한적 | CS 우세 |

**⚠️ interleaved output 해석 주의사항**:
missless.co는 "단일 모델의 interleaved output"이 아닌 "Live API 음성 + Tool Call 이미지/BGM"으로
다중 모델을 orchestrate하여 interleaved stream 구현. 데모 영상과 텍스트 설명에서
**"Gemini's interleaved output capabilities + Tool orchestration"**이라고 명시적으로 서술할 것.

---

## 제출 필수 체크리스트 (DevPost 공식 규칙 완전판)

### 필수 제출물

- [ ] 4분 미만 데모 영상 (YouTube/Vimeo **공개** — 비공개/일부공개 시 심사 불가)
- [ ] 영문 또는 영문 자막 포함
- [ ] 실제 동작 소프트웨어 시연 (목업/모션 그래픽 불가)
- [ ] **문제 정의 + 솔루션 가치 피치** 포함 (데모 영상 내)
- [ ] 공개 소스 코드 레포지토리 (GitHub) — 2026-02-16 이후 생성
- [ ] step-by-step spin-up 가이드 포함 README.md (재현 가능)
- [ ] 아키텍처 다이어그램 (이미지 캐러셀 또는 파일 업로드)
- [ ] Cloud 배포 증거 (데모 영상과 **별도** 스크린 레코딩 또는 코드 링크)
- [ ] 프로젝트 텍스트 설명 (영문: 기능, 기술, 데이터 소스, 학습 내용)
- [ ] **카테고리 선택**: Creative Storyteller
- [ ] **테스팅 접근**: 라이브 URL (missless.co) + 필요 시 테스트 크리덴셜
- [ ] **사용한 Google 기술 명시** (제출 폼 내)
- [ ] **서드파티 통합 선언** (gorilla/websocket, Next.js 등 — 제출 설명에 명시)
- [ ] Gemini 모델 사용 필수
- [ ] GenAI SDK 사용 필수 (google.golang.org/genai)
- [ ] 최소 1개 Google Cloud 서비스 사용 필수
- [ ] Google Cloud 이용약관(AUP) 준수

### 보너스 제출물

- [ ] **dev.to 블로그 시리즈 URL (4편)** — 모든 포스트에 **"created for purposes of entering this hackathon"** 문구 포함 (+0.6)
  - [ ] Post #1 URL (3/2): Real-Time Audio Proxy
  - [ ] Post #2 URL (3/8): YouTube → AI Persona
  - [ ] Post #3 URL (3/12): Interleaved Reunion
  - [ ] Post #4 URL (3/14-15): Complete Architecture (메인)
- [ ] `cloudbuild.yaml` / Terraform — 공개 repo에 포함 (+0.2)
- [ ] GDG 공개 프로필 링크 (+0.2)

---

## 제한사항 및 예외상황 정리

### API 제한사항
| 항목 | 제한 | 대응 |
|------|------|------|
| Gemini API Tier 1 RPM | 150 RPM | retryWithBackoff + semaphore |
| Live API 동시 세션 | Tier별 50~1,000 | Cloud Run 인스턴스당 1세션 |
| Live API 세션 시간 | 무제한 (GoAway 발생 시 재접속) | Session Resumption |
| YouTube URL 분석 | 공개(public) 영상만 (일부공개 불가) | 비공개/일부공개 → 갤러리 Fallback |
| Lyria RealTime | WebSocket 전용 API (`client.aio.live.music.connect()`) — Go SDK 미지원 | **프리셋 BGM 기본 전략** (Cloud Storage 서빙) |
| Imagen 4 이미지 | 안전 필터링 적용 | 실루엣/수채화 스타일로 우회 |
| Cloud Run 메모리 | 512Mi 기본 | 이미지 오프로드 → Cloud Storage |
| Cloud Run graceful shutdown | SIGTERM 후 10초 | 병렬 shutdown 8초 이내 |

### 예외상황 대응
| 상황 | 대응 |
|------|------|
| YouTube 비공개/일부공개 영상 선택 | → 공개 변경 안내 또는 갤러리 업로드 |
| 영상 분석 실패 | → 재시도 3회 → 실패 시 갤러리 Fallback |
| Live API GoAway 시그널 | → Session Resumption으로 자동 재접속 |
| 이미지 생성 실패 | → tool_error 이벤트 → AI가 자연스럽게 안내 |
| BGM 재생 실패 | → 로컬 프리셋 BGM Fallback (/bgm/*.mp3) |
| 300초 재회 시간 초과 | → 240초 경고 → 300초 자동 종료 → 앨범 생성 |
| Context Window 초과 | → SlidingWindow 압축 (8192 target) |
| 사용자 마이크 권한 거부 | → 안내 팝업 + 텍스트 입력 미지원 안내 |
| 모바일 AudioContext 제한 | → "시작하기" 터치 이벤트에서 초기화 |
| OAuth 토큰 만료 | → Refresh Token으로 자동 갱신 |
| Firestore 쿼리 한계 (전문 검색 미지원) | → 키워드 매칭 + 추후 벡터 검색 |
| Cloud Run cold start | → min-instances=1 설정 |

### 지리적 제한 (참가 불가 지역)
이탈리아, 퀘벡(캐나다), 크리미아, 쿠바, 이란, 시리아, 북한, 수단, 벨라루스, 러시아, 미국 제재 대상국

---

## 외부 참조 목록

| 리소스 | URL |
|--------|-----|
| DevPost 공식 페이지 | https://geminiliveagentchallenge.devpost.com/ |
| DevPost 공식 규칙 | https://geminiliveagentchallenge.devpost.com/rules |
| DevPost 리소스 | https://geminiliveagentchallenge.devpost.com/resources |
| ADK 공식 문서 (참고용) | https://google.github.io/adk-docs/ |
| Gemini Live API 문서 | https://ai.google.dev/gemini-api/docs/live |
| Gemini Live API 가이드 | https://ai.google.dev/gemini-api/docs/live-guide |
| Gemini TTS 음성 목록 | https://ai.google.dev/gemini-api/docs/speech-generation |
| Way Back Home 코드랩 | https://codelabs.developers.google.com/way-back-home-level-3 |
| Lyria (Music Gen) 문서 | https://ai.google.dev/gemini-api/docs/music-generation |
| Go GenAI SDK | https://pkg.go.dev/google.golang.org/genai |
| Go ADK (미사용, 참고용) | https://pkg.go.dev/google.golang.org/adk |
| 크레딧 신청 마감 | 2026-03-13 12:00 PM PT |
| Imagen 공식 문서 | https://ai.google.dev/gemini-api/docs/imagen |
| Imagen deprecation | https://ai.google.dev/gemini-api/docs/deprecations |
| 음성 설정 가이드 | https://docs.cloud.google.com/vertex-ai/generative-ai/docs/live-api/configure-language-voice |

---

## V7 기술 주장 공식문서 교차검증 (2026-02-24)

> 아래 표는 V7 문서의 모든 핵심 기술 주장을 공식 문서와 대조 검증한 결과입니다.

| # | 기술 주장 | V7 값 | 검증 결과 | 공식 출처 |
|---|----------|-------|:---------:|-----------|
| 1 | Flash Image 모델명 | `gemini-2.5-flash-image` | ✅ 확인 | [Image Generation 문서](https://ai.google.dev/gemini-api/docs/image-generation) |
| 2 | Pro 모델명 (GA) | `gemini-2.5-pro` | ✅ 확인 | [Models 문서](https://ai.google.dev/gemini-api/docs/models) |
| 3 | Live API 모델 (Developer) | `gemini-2.5-flash-native-audio-preview-12-2025` | ✅ 확인 | [Live API 문서](https://ai.google.dev/gemini-api/docs/live) |
| 4 | Live API 모델 (Vertex) | `gemini-live-2.5-flash-native-audio` | ✅ 확인 | [Vertex AI Live API](https://docs.cloud.google.com/vertex-ai/generative-ai/docs/live-api) |
| 5 | Imagen 모델 | `imagen-4.0-generate-001` (Imagen 4) | ✅ 확인 | [Imagen 문서](https://ai.google.dev/gemini-api/docs/imagen) — Imagen 3은 shut down |
| 6 | 30개 프리셋 음성 | 30 voices (Zephyr~Sulafat) | ✅ 확인 | [음성 설정 가이드](https://docs.cloud.google.com/vertex-ai/generative-ai/docs/live-api/configure-language-voice) — 공식 문서에 "HD" 라벨 없음, "voice options"으로 표기 |
| 7 | Lyria RealTime | WebSocket 전용, Go SDK 미지원 | ✅ 확인 | [Go SDK](https://pkg.go.dev/google.golang.org/genai) — Lyria 관련 타입 없음 |
| 8 | Go SDK v1.47.0 시그니처 | `Receive()`, `SendRealtimeInput(LiveRealtimeInput{})` 등 (no ctx) | ✅ 확인 | [Session GoDoc](https://pkg.go.dev/google.golang.org/genai#Session) |
| 9 | ContextWindowCompression | `ContextWindowCompressionConfig` + `SlidingWindow` | ✅ 확인 | [GoDoc](https://pkg.go.dev/google.golang.org/genai) |
| 10 | SessionResumption | `SessionResumptionConfig` | ✅ 확인 | [GoDoc](https://pkg.go.dev/google.golang.org/genai) — ⚠️ `Transparent` 필드는 Vertex AI 전용 |
| 11 | Affective Dialog | `enable_affective_dialog` 설정으로 활성화 | ✅ 확인 | [Live Guide](https://ai.google.dev/gemini-api/docs/live-guide) |
| 12 | Proactive Audio | `proactivity` 설정으로 활성화 | ✅ 확인 | [Live Guide](https://ai.google.dev/gemini-api/docs/live-guide) |
| 13 | YouTube 공개만 분석 | unlisted URL → Gemini FileData 분석 불가 | ✅ 확인 | [YouTube Data API](https://developers.google.com/youtube/v3/docs/videos) |
| 14 | Grand Prize | $25,000 + 크레딧/티켓/여행비 | ✅ 확인 | [DevPost Rules](https://geminiliveagentchallenge.devpost.com/rules) |
| 15 | Category Winner | $10,000 (3트랙) | ✅ 확인 | DevPost Rules |
| 16 | Subcategory Winner | $5,000 (3부문) | ✅ 확인 | DevPost Rules |
| 17 | Honorable Mentions | 5×$2,000 | ✅ 확인 | DevPost Rules |

**검증 방법**: 각 항목의 공식 URL을 직접 fetch하여 현재(2026-02-24) 시점의 내용과 대조.
**V6 대비 추가 발견**:
- Imagen 3 (`imagen-3.0-generate-002`) 2025-11-10 shut down → Imagen 4 (`imagen-4.0-generate-001`)로 대체 완료
- "HD" 음성 라벨: 공식 문서에는 "HD" 표기 없음 → 마케팅 용도로만 사용, 기술 문서에서는 "프리셋 음성"으로 통일
- `SessionResumptionConfig.Transparent`: Developer API 미지원 (Vertex AI 전용)
- Lyria 문서 URL: `/docs/lyria` → 404, 정확한 URL은 `/docs/music-generation`
