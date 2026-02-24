# missless.co — FINAL 계획서

> Gemini Live Agent Challenge 2026 | Creative Storyteller 트랙
> 작성일: 2026-02-24 | 마감일: 2026-03-16 (D-20)

---

## 1. 챌린지 채점 기준

최종 점수는 **1~6점** 스케일.

| 항목 | 비중 | 핵심 질문 |
|------|:---:|----------|
| Innovation & Multimodal UX | 40% | 텍스트 박스를 탈피했는가? See/Hear/Speak가 유동적인가? 독특한 성격/목소리가 있는가? Live하고 context-aware한가? |
| Technical Implementation | 30% | GenAI SDK/ADK를 효과적으로 사용했는가? Google Cloud에 견고히 배포되었는가? 할루시네이션 방지/그라운딩 증거가 있는가? |
| Demo & Presentation | 30% | 문제-솔루션이 명확한가? 아키텍처 다이어그램이 명확한가? Cloud 배포 증거가 있는가? 실제로 동작하는가? |

### Creative Storyteller 트랙 정의 (공식)

> "크리에이티브 디렉터처럼 사고하고 창작하는 에이전트를 구축하라.
> 텍스트, 이미지, 오디오, 비디오를 **하나의 일관된 흐름(single, coherent flow)**으로 엮어라.
> Gemini의 **네이티브 인터리브드 출력**을 반드시 활용해야 한다."

### 심사위원이 보는 것

> "Judges are **not required to test the Project** and may choose to judge based solely on the text description, images, and **video** provided in the Submission."

→ **데모 영상이 사실상 최종 심사 대상.** 앱 자체보다 영상이 더 중요할 수 있음.

---

## 2. 서비스 정의

### missless.co — 그리움을 줄여드립니다

**한 줄**: 그리운 사람과의 가상 재회를 AI가 만들어주는 음성 기반 몰입형 경험.

**대상 사용자**: 군대 간 연인, 유학 간 자녀, 출장 중인 가족, 돌아가신 분을 그리워하는 사람. **전 세계 YouTube 사용자 대상.**

**핵심 경험**: 사용자가 Google 로그인 후 YouTube 영상을 선택하면, **Gemini API가 YouTube URL을 직접 분석**하여 영상 속 그리운 사람의 성격·말투·표정으로 페르소나를 생성한다. 영상 다운로드 없이 URL만으로 분석이 완료되는 **제로 다운로드** 아키텍처. 이후 음성으로 대화하면 AI가 그 사람의 성격으로 응답하며, 동시에 장면 이미지와 배경음악이 실시간으로 생성되어 하나의 연속적 흐름으로 펼쳐진다.

**데이터 소스 우선순위**:
1. **YouTube 영상 — URL 직접 분석** (메인) — Google 로그인 후 공개/일부공개 영상 선택 → Gemini API에 YouTube URL 직접 전달 → 다운로드 없이 분석. 복수 선택 가능.
2. **디바이스 갤러리 영상** (보조) — 비공개 영상이나 YouTube 외 영상은 디바이스에서 직접 업로드.
3. **사진** (선택) — 추가 시각 레퍼런스로 페르소나 정밀도 향상.
4. **음성 메시지** (선택) — 음성 톤/말투 매칭 정밀도 향상.

### 설계 원칙

1. **ZERO TEXT INPUT** — 사용자는 타이핑하지 않는다. 모든 상호작용은 음성.
2. **AI-FIRST** — AI가 먼저 시작하고, 스스로 이끈다. 사용자는 그 안에 존재한다.
3. **FLUID STREAM** — 이미지·음성·BGM이 끊김 없이 동시에 흘러간다.
4. **CONTEXT-AWARE** — 카메라(표정), 마이크(톤/침묵), 대화 맥락을 실시간으로 읽는다.
5. **EVERY TIME UNIQUE** — 같은 페르소나라도 매번 다른 경험. AI가 즉석에서 시나리오를 생성.

---

## 3. ADK Go 공식 샘플 분석 및 missless 패턴 매핑

### 3.1 공식 샘플 4종 분석

Google ADK Samples 레포지토리(`google/adk-samples/go/agents`)의 4개 에이전트를 분석하여 missless에 적용할 아키텍처 패턴을 도출한다.

#### (1) boat-agent — 단일 에이전트 + Embedded Instruction

```
구조: instruction.md (//go:embed) + main.go (HTTP 서버)
모델: gemini-2.5-flash
특징: 단순 질의응답, 구조화된 System Instruction을 외부 마크다운에서 주입
```

**missless 적용**: 페르소나별 System Instruction을 `prompts/*.md`로 분리 관리. `//go:embed`로 기본 프롬프트를 바이너리에 포함하고, 사용자별 동적 데이터는 런타임에 합성.

#### (2) llm-auditor — Sequential 워크플로우 에이전트

```
구조: auditor/ → critic/ → reviser/ (3단계 순차 파이프라인)
패턴: 한 에이전트의 출력이 다음 에이전트의 입력
특징: Google Search Tool 활용, 단계별 책임 분리
```

**missless 적용**: 온보딩 파이프라인을 Sequential 패턴으로 구성.
`VideoAnalyzer(YouTube 영상 분석)` → `PersonaBuilder(성격 프로파일 구축)` → `VoiceMatcher(HD 음성 매칭)`.
각 단계가 명확한 입출력을 가지며 순차 실행된다.

#### (3) navallist — 풀스택 임베디드 에이전트

```
구조: cmd/server/ + internal/agent/ + internal/data/ + web/
패턴: Go HTTP 서버 안에 에이전트가 내장, PostgreSQL + WebSocket 실시간 동기화
특징: 에이전트가 자연어 의도를 해석하여 DB 조작 Tool을 자동 호출
도구: agent/tools.go에 Tool 정의, agent/instruction.md에 System Instruction
```

**missless 적용**: 이것이 missless의 **메인 아키텍처 레퍼런스**. Go 서버 안에 Live API 에이전트를 내장하고, WebSocket으로 프론트엔드와 실시간 통신하며, Firestore에 세션/페르소나 데이터를 관리한다. 에이전트의 Tool 호출 결과가 WebSocket을 통해 즉시 프론트엔드에 반영되는 구조가 동일하다.

#### (4) sail-researcher — 멀티 에이전트 + Sub-Agent as Tool

```
구조: agent_setup.go (팩토리) + prompts/*.md + tools/*.go + tool_monitor.go
패턴: 메인 에이전트가 서브 에이전트를 Tool로 등록하여 사용
특징:
  - constructAgent() 팩토리 패턴으로 에이전트 생성
  - 각 에이전트별 전용 프롬프트 (discovery_agent.md, voyage_agent.md 등)
  - tool_monitor.go로 Tool 실행 텔레메트리 추적
  - Temperature 0.2~0.4 범위로 창의성/일관성 제어
  - maxOutputTokens: 65536
```

**missless 적용**: Live API 세션의 Tool로 등록되는 각 기능(generate_scene, analyze_user 등)을 **독립 서브 에이전트로 분리**. 특히 `tool_monitor.go` 패턴을 채택하여 모든 Tool 호출의 지연 시간·성공 여부를 slog으로 기록한다.

### 3.2 프로그래밍 필수 영역 vs 에이전트 처리 영역

missless의 전체 기능을 **반드시 코드로 작성해야 하는 결정적 영역**과 **AI 에이전트에게 위임하는 비결정적 영역**으로 분류한다.

#### 프로그래밍 필수 (Deterministic — 코드로 구현)

이 영역은 정확한 프로토콜 처리, 바이너리 데이터 변환, 인프라 통신 등 **결과가 항상 동일해야 하는** 작업이다. AI가 대신할 수 없다.

| 영역 | 구현 내용 | 참조 샘플 |
|------|----------|----------|
| **WebSocket 서버** | 클라이언트↔백엔드 양방향 실시간 통신. 메시지 타입 라우팅 (audio/image/control). | navallist: `internal/server/` |
| **오디오 인코딩/디코딩** | PCM 16bit/16kHz(입력) ↔ PCM 16bit/24kHz(출력) 변환. Web Audio API AudioWorklet. | — (프로토콜 스펙) |
| **이미지 전송 파이프라인** | Gemini 응답에서 이미지 바이트 추출 → Base64/Binary → WebSocket 프레임 전송 → Canvas 렌더링. | — (바이너리 처리) |
| **카메라 프레임 캡처** | MediaStream API → 주기적 JPEG 스냅샷 → 백엔드 전송. | — (브라우저 API) |
| **Firestore CRUD** | 페르소나 프로필 저장/조회, 세션 이력 기록, 추억 데이터 색인. | navallist: `internal/data/` |
| **Cloud Storage 업로드** | 사용자 미디어(영상/사진/음성), 생성 이미지, 앨범 저장. 원본 영상은 임시 저장(분석 후 삭제). | — (Cloud SDK) |
| **Google OAuth 2.0** | Google 로그인 + YouTube Data API v3 인증. `youtube.readonly` scope로 채널 영상 목록 조회. | — (인증 레이어) |
| **YouTube 영상 목록 조회** | `channels.list` → `playlistItems.list`로 사용자 업로드 영상(비공개/일부공개 포함) 목록+썸네일 조회. | navallist: `internal/data/` |
| **YouTube URL 직접 분석** | 사용자 선택 YouTube 영상의 URL을 Gemini API `FileData{FileURI}` 파라미터로 직접 전달 → Video Understanding 분석. 다운로드/임시저장 불필요. 공개/일부공개만 가능. | — (Gemini API) |
| **갤러리 영상 업로드 (Fallback)** | 비공개 영상/YouTube 외 영상: 디바이스 갤러리 → Cloud Storage 임시 저장 → Gemini File API 업로드 → 분석 완료 후 원본 삭제. | — (파일 관리) |
| **인물 감지 및 선택 UI** | 영상 프레임에서 Gemini Vision으로 인물 바운딩 박스 추출 → 얼굴/형태 크롭 → 사용자가 대상 인물 선택. | — (Vision 처리) |
| **BGM 오디오 관리** | Howler.js/Web Audio API로 BGM 재생, 크로스페이드, 볼륨 제어. | — (오디오 엔진) |
| **세션 상태 머신** | `idle → onboarding → analyzing → reunion → album → ended` 상태 전이. | navallist: 상태 관리 |
| **HTTP/Health 엔드포인트** | REST API (업로드, 세션 생성, 앨범 조회). Cloud Run 헬스체크. | boat-agent: `main.go` |
| **IaC & CI/CD** | Terraform (Cloud Run + Firestore + Storage), Cloud Build 파이프라인. | — (인프라 코드) |
| **Tool 텔레메트리** | 모든 Tool 호출의 latency, success/fail, 파라미터를 slog 구조화 로깅. | sail-researcher: `tool_monitor.go` |
| **Ephemeral Token 발급** | 백엔드에서 단기 토큰 생성 → 클라이언트에 전달 → 클라이언트가 Live API WebSocket에 인증. API 키를 브라우저에 노출하지 않는다. | — (보안 레이어) |
| **GoAway 시그널 처리** | Live API의 ~10분 연결 수명 만료 시 GoAway 시그널 수신 → Session Resumption 토큰으로 자동 재접속 → 사용자 경험 끊김 없음. | — (연결 관리) |
| **audioStreamEnd 이벤트** | 사용자 마이크 1초 이상 무음 시 `audioStreamEnd` 전송 → 서버 측 캐시된 오디오 플러시. 응답 지연 방지. | — (오디오 프로토콜) |
| **비동기 Tool 에러 파이프라인** | 이미지 생성 Tool 실패 시 프론트 WebSocket으로 `tool_error` 이벤트 전송 → 프론트가 Live API에 Client Content 주입 → 에이전트가 자연스럽게 안내. | — (에러 핸들링) |
| **Rate Limit 방어 (Backoff+Jitter)** | 모든 Gemini API 호출에 Exponential Backoff + Jitter 적용. Sequential Agent 연속 호출 및 이미지 생성 429 에러 방어. | — (API 안정성) |
| **AudioContext 활성화** | 온보딩 첫 터치 이벤트에서 `AudioContext.resume()` 강제 실행. 모바일 브라우저 보안 정책 대응. 마이크 권한 요청 동시 수행. | — (브라우저 정책) |

#### 에이전트 처리 (Non-deterministic — AI가 결정)

이 영역은 **맥락에 따라 매번 다른 결과를 내야 하는** 창의적·판단적 작업이다. 프로그래밍으로 규칙을 정할 수 없으며, LLM의 추론에 위임한다.

| 영역 | AI가 결정하는 것 | 사용 모델 | 참조 샘플 |
|------|----------------|----------|----------|
| **대화 흐름 주도** | 어떤 말을 할지, 언제 침묵할지, 감정 톤 조절. System Instruction이 페르소나 성격을 규정하되 실제 발화는 AI가 생성. | Live API (native-audio) | boat-agent: instruction.md |
| **장면 생성 타이밍** | 대화 중 언제 `generate_scene` Tool을 호출할지. "새로운 장소로 이동" "감정 전환점" 등의 맥락을 AI가 판단. | Live API (auto tool call) | sail-researcher: 에이전트가 Tool 호출 결정 |
| **장면 이미지 프롬프트** | Tool 호출 시 전달할 prompt/mood/characters 파라미터. 대화 맥락과 페르소나에 맞는 장면을 AI가 구성. | 3 Pro Image / 2.5 Flash Image | — |
| **분위기 전환 판단** | `change_atmosphere` 호출 시점과 mood 값. 사용자 감정 변화를 감지하여 BGM/톤을 자동 조절. | Live API + Affective Dialog | — |
| **추억 회상 시점** | `recall_memory` 호출 시점과 topic. 대화 흐름에서 자연스럽게 과거 이야기를 꺼내는 타이밍. | Live API (auto tool call) | — |
| **사용자 감정 해석** | 카메라 프레임의 표정 분석 결과를 어떻게 대화에 반영할지. "울먹임 감지 → 더 부드럽게" 등. | 3 Flash (Vision) | — |
| **재회 마무리 판단** | 언제 자연스럽게 `end_reunion`을 호출할지. 대화가 충분히 진행되었는지, 사용자가 만족했는지 추론. | Live API (auto tool call) | — |
| **페르소나 프로파일 추출** | YouTube 영상에서 음성 전사·말투 패턴·표정·행동·성격 특성을 종합 분석하여 프로파일 생성. 복수 영상 크로스 분석. | 3.1 Pro + Video Understanding | llm-auditor: 순차 분석 |
| **온보딩 대화 진행** | 사용자에게 어떤 질문을 할지, 어떤 순서로 정보를 수집할지. 영상 분석 대기 중 실시간 진행 상황과 발견된 정보를 공유. | Live API | boat-agent: 자연어 상호작용 |
| **시나리오 즉석 생성** | 매 재회마다 새로운 장소·상황·이벤트를 AI가 독자적으로 생성. 사전 정의된 시나리오 없음. | Live API + System Instruction | sail-researcher: 멀티 에이전트 협업 |

### 3.3 패턴 매핑 종합

```
google/adk-samples 패턴           →  missless.co 적용
─────────────────────────────────────────────────────────────
boat-agent                        →  페르소나별 instruction.md
 (단일 에이전트 + //go:embed)          //go:embed 기본 프롬프트
                                       + 런타임 동적 합성

llm-auditor                       →  온보딩 파이프라인
 (Sequential 워크플로우)               VideoAnalyzer (YouTube 영상)
                                       → PersonaBuilder
                                       → VoiceMatcher

navallist                         →  메인 애플리케이션 구조
 (풀스택 임베디드 에이전트)             Go 서버 + 내장 에이전트
                                       + WebSocket + Firestore

sail-researcher                   →  Live API + Tool 에이전트
 (멀티 에이전트 + Tool 오케스트레이션)    generate_scene 서브 에이전트
                                       analyze_user 서브 에이전트
                                       tool_monitor 텔레메트리
```

---

## 4. 핵심 아키텍처: GenMedia Live 패턴

### 레퍼런스: GenMedia Live (Google Cloud Community, 2026.01)

Google Cloud Community에서 Dr. Wafae Bakkali가 발표한 **GenMedia Live** 패턴을 missless.co의 기반으로 채택한다.

> **핵심 개념**: Gemini Live API를 **음성 컨트롤 레이어**로 사용하고,
> 이미지 생성(Gemini 3 Pro Image)과 비디오 생성(Veo)을
> **Tool(Function Calling)**로 등록하여 Live API가 자동으로 호출하게 한다.

```
[사용자 음성] ──→ Gemini Live API (양방향 실시간)
                       │
                       │ (AI가 대화 흐름에서 자동으로 Tool 호출)
                       │
          ┌────────────┼────────────┐
          ▼            ▼            ▼
   [generate_scene]  [change_bgm]  [recall_memory]
   Gemini 3 Pro Image  BGM Engine   Firestore RAG
   장면 이미지 생성     분위기 전환   추억 회상
```

**이것이 왜 최적인가:**
- 하나의 Live API 세션 안에서 음성 대화 + 이미지 생성 + 기타 도구가 **자동으로 오케스트레이션**됨
- 별도의 오케스트레이터 코드 불필요 — Gemini Live API가 알아서 Tool을 호출
- 음성 대화가 끊기지 않으면서 이미지가 생성됨 → **fluid stream** 달성
- Google이 직접 소개한 패턴이므로 기술 심사에서 높은 점수

### 전체 시스템 아키텍처

```
┌──────────────────────────────────────────────────────────┐
│                    사용자 디바이스                           │
│                Next.js 15 PWA (모바일 퍼스트)               │
│                                                            │
│  [카메라] ──→ 표정/시선 캡처 (MediaStream API, 별도 분석)     │
│  [마이크] ──→ PCM 16bit/16kHz 오디오 스트림                  │
│  [스피커] ←── PCM 16bit/24kHz 오디오 스트림                  │
│  [화면]  ←── 전체화면 장면 이미지 (크로스페이드)              │
│                                                            │
│  WebSocket ←──────────────────────────→ Go Backend          │
└────────────────────────┬─────────────────────────────────┘
                         │
                         ▼
┌──────────────────────────────────────────────────────────┐
│               Go Backend (Cloud Run)                       │
│         google.golang.org/genai v1.29.0+                   │
│         google.golang.org/adk v0.4.0                       │
│         google.golang.org/api/youtube/v3                    │
│         golang.org/x/oauth2/google (OAuth 2.0)              │
│                                                            │
│  ┌──────────────────────────────────────────────────────┐ │
│  │          Gemini Live API Session (핵심)                │ │
│  │          gemini-live-2.5-flash-native-audio            │ │
│  │                                                        │ │
│  │  System Instruction:                                   │ │
│  │    - 페르소나 성격/말투/기억 (//go:embed + 동적 합성)    │ │
│  │    - AI-first 행동 규칙                                 │ │
│  │    - 장면 생성 트리거 규칙                               │ │
│  │                                                        │ │
│  │  Registered Tools:                                     │ │
│  │    ┌────────────────────────────────────────────────┐  │ │
│  │    │ generate_scene(prompt, mood, characters)        │  │ │
│  │    │ → Gemini 3 Pro Image로 장면 이미지 생성          │  │ │
│  │    │                                                 │  │ │
│  │    │ generate_fast_scene(prompt)                     │  │ │
│  │    │ → Gemini 2.5 Flash Image로 빠른 보조 이미지     │  │ │
│  │    │                                                 │  │ │
│  │    │ change_atmosphere(mood, intensity)              │  │ │
│  │    │ → BGM 전환 + 장면 톤 변경                        │  │ │
│  │    │                                                 │  │ │
│  │    │ recall_memory(topic)                            │  │ │
│  │    │ → Firestore에서 관련 추억/대화 검색              │  │ │
│  │    │                                                 │  │ │
│  │    │ analyze_user(image_data)                        │  │ │
│  │    │ → Gemini 3 Flash Vision으로 표정 분석            │  │ │
│  │    │                                                 │  │ │
│  │    │ end_reunion(summary)                            │  │ │
│  │    │ → 재회 마무리 + 앨범 생성                         │  │ │
│  │    └────────────────────────────────────────────────┘  │ │
│  │                                                        │ │
│  │  Live API 기능 활용:                                    │ │
│  │    ✅ Affective Dialog (감정 인지 대화)                  │ │
│  │    ✅ Proactive Audio (관련 있을 때만 응답)              │ │
│  │    ✅ Context Window Compression (장시간 세션)           │ │
│  │    ✅ Session Resumption (끊김 복구, 2시간)              │ │
│  │    ✅ 30 HD Voices (페르소나 음성 매핑)                  │ │
│  │    ✅ Automatic Tool Execution                         │ │
│  └──────────────────────────────────────────────────────┘ │
│                                                            │
│  ┌───────────────┐  ┌────────────┐  ┌────────────┐       │
│  │ Onboarding    │  │ Album      │  │ Tool       │       │
│  │ Pipeline      │  │ Generator  │  │ Monitor    │       │
│  │ (Sequential)  │  │            │  │ (slog)     │       │
│  │               │  │            │  │            │       │
│  │ VideoAnalyzer │  │ 장면 저장   │  │ latency    │       │
│  │ → PersonaBuild│  │ 앨범 생성   │  │ tracking   │       │
│  │ → VoiceMatch  │  │            │  │            │       │
│  └───────────────┘  └────────────┘  └────────────┘       │
└─────┬────────────────┬──────────────────┬────────────────┘
      ▼                ▼                  ▼
┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐
│ Gemini   │  │ YouTube  │  │ Cloud    │  │ Cloud    │
│ API      │  │ Data API │  │ Storage  │  │ Firestore│
│(5모델    │  │ v3       │  │(Fallback │  │ (데이터) │
│+URL분석) │  │(영상목록)│  │+생성물)  │  │          │
└──────────┘  └──────────┘  └──────────┘  └──────────┘

    ⚡ YouTube URL 직접 분석 (제로 다운로드)
    ──────────────────────────────────────
    YouTube 공개/일부공개 URL → Gemini API (genai.FileData{FileURI})
    → 다운로드/임시저장 없이 직접 Video Understanding 분석
    → Cloud Storage는 갤러리 Fallback + 생성 이미지 저장에만 사용
```

---

## 5. Gemini 모델 배치 (2026.02.24 기준)

### 사용 모델 5종

| # | 모델 ID | 용도 | 호출 시점 | 비고 |
|---|---------|------|----------|------|
| 1 | `gemini-live-2.5-flash-native-audio` | **메인 세션** — 양방향 실시간 음성 + Tool Calling + Affective Dialog | 상시 (세션 전체) | Live API WebSocket. 30 HD 음성. 24개 언어. |
| 2 | `gemini-3-pro-image-preview` | 고품질 장면 이미지 생성. Interleaved text+image. 4K. 캐릭터 일관성. | Tool 호출 시 (대화 전환점) | Creative Storyteller 핵심 기술 |
| 3 | `gemini-2.5-flash-image` | 빠른 보조 이미지 생성. 저지연. | Tool 호출 시 (미세 장면 변화) | 3 Pro Image보다 3배 빠름 |
| 4 | `gemini-3-flash-preview` | 시나리오 로직, 표정 분석(Vision), 감정 추론, 빠른 판단 | Tool 호출 시 (중요 순간) | thinking_level 4단계 활용. 카메라 프레임은 Live API 세션 외부에서 별도 API 호출로 분석 (오디오+비디오 2분 제한 회피). |
| 5 | `gemini-3.1-pro-preview` | 페르소나 초기 분석 — 사진/대화/음성에서 성격 프로파일 추출 | 온보딩 1회 | 65K 토큰 출력. 최강 추론. |

### 모델 퇴역 안전성 확인

| 모델 | 상태 | 퇴역일 | 안전 |
|------|------|--------|:---:|
| gemini-2.0-flash | ❌ 사용 안 함 | 2026-03-31 | ✅ |
| gemini-live-2.5-flash-preview-09-2025 | ❌ 사용 안 함 | 2026-03-19 | ✅ |
| gemini-live-2.5-flash-native-audio | ✅ 사용 | 미정 | ✅ |
| gemini-3-pro-image-preview | ✅ 사용 | 미정 | ✅ |
| gemini-3-flash-preview | ✅ 사용 | 미정 | ✅ |
| gemini-3.1-pro-preview | ✅ 사용 | 미정 | ✅ |
| gemini-2.5-flash-image | ✅ 사용 | 미정 | ✅ |

### Live API 고급 기능 활용

| 기능 | 설정 | missless에서의 활용 |
|------|------|-------------------|
| **Affective Dialog** | `enableAffectiveDialog: true` | 사용자 음성 톤에서 감정 인지 → 페르소나가 감정에 맞게 응답. 실험적 기능. |
| **Proactive Audio** | `proactiveAudio: true` | 사용자가 혼잣말/배경 소음일 때는 응답 안 함. 직접 말 걸 때만 반응. |
| **Context Window Compression** | `contextWindowCompression: { slidingWindow: {targetTokenCount: N} }` | 장시간 재회 세션(10분+) 지원. 오래된 대화 자동 요약 압축. |
| **Session Resumption** | `sessionResumption: { handle: "token" }` | 네트워크 끊김(WiFi→5G 전환 등) 시 **2시간** 내 자동 복구. 서버가 주기적으로 WebSocket을 리셋할 때 세션 유지. |
| **LanguageCode (BCP-47)** | `SpeechConfig.LanguageCode: "ko-KR"` | 응답 언어 명시적 힌트. System Instruction과 병행하여 한국어 고정. |
| **Input/Output Audio Transcription** | `InputAudioTranscription: {}`, `OutputAudioTranscription: {}` | 대화 로그 자동 생성. 디버깅 및 앨범 텍스트에 활용. |
| **Thought Signatures** | 자동 반환 → 다음 호출에 포함 | 멀티턴 대화에서 추론 연속성 유지. |
| **audioStreamEnd** | 마이크 무음 1초 시 전송 | 서버 캐시 오디오 플러시. 응답 지연 방지. |
| **GoAway + Auto Reconnect** | ~10분 연결 수명, GoAway 시그널 | Session Resumption 토큰으로 끊김 없는 자동 재접속. |
| **Ephemeral Tokens** | 백엔드 토큰 발급 → 클라이언트 인증 | API 키 브라우저 미노출. WebSocket 보안 인증. |
| **Google Search Grounding** | `googleSearch` Tool 등록 | 실시간 정보(날씨, 뉴스 등)를 대화에 자연 반영. 페르소나 대화 풍부화. |

---

## 6. 사용자 경험 플로우

### Phase 1: 온보딩 (Google 로그인 + YouTube 영상 선택 + 음성 대화)

```
[missless.co 접속 — 모바일 브라우저]

── Step 1: Google 로그인 ──

🔘 [Google 로그인 버튼 — "Sign in with Google"]
   (OAuth 2.0, scope: youtube.readonly)

   → 사용자 인증 완료

── Step 1.5: AudioContext 활성화 (브라우저 보안 정책 대응) ──

🔘 ["그리운 사람을 만나러 갈까요?" 시작 버튼 터치]
   → 이 터치 이벤트에서 AudioContext.resume() 실행
   → 모바일 Chrome/Safari 모두 사용자 상호작용 없이는
     AudioContext가 시작되지 않는 브라우저 보안 정책 대응
   → 마이크 권한 요청 + WebSocket 연결 초기화도 여기서 수행

── Step 2: 음성 대화로 시작 ──

🔊 AI: "안녕하세요. missless에 오신 걸 환영해요."
       "보고싶은 사람이 있으세요?"

🎤 사용자: "네..."

🔊 AI: "그 사람 이름이 뭐예요?"

🎤 사용자: "민수요. 남자친구예요."

🔊 AI: "민수는 어떤 사람이에요? 편하게 얘기해주세요."

🎤 사용자: "장난을 많이 치는데 은근 다정해요.
           지금 군대에 있어서 연락이 잘 안 돼요."

── Step 3: YouTube 영상 선택 ──

🔊 AI: "민수가 나오는 영상이 있으면 보여주세요.
        화면 아래에 유튜브 영상 목록이 보일 거예요."

📺 [화면: 사용자 YouTube 채널 영상 썸네일 그리드 표시]
   (YouTube Data API로 업로드 영상 목록 조회)
   (공개/일부공개 영상: ✅ 분석 가능 표시)
   (비공개 영상: 🔒 "일부공개로 변경하면 분석 가능" 안내)
   (영상별 제목+길이+썸네일+privacyStatus)

   [사용자: 민수가 나오는 영상 1~3개 터치 선택]

   💡 YouTube URL → Gemini API 직접 분석 (다운로드 없음!)
   genai.FileData{FileURI: "https://youtube.com/watch?v=...", MIMEType: "video/mp4"}

   📱 [비공개 영상인 경우 → 안내 팝업]
      "이 영상은 비공개여서 직접 분석할 수 없어요.
       '일부공개'로 변경하시거나, 갤러리에서 직접 업로드해주세요."
      [일부공개 변경 가이드 링크] [갤러리에서 업로드]

── Step 3.5: 영상에서 인물 감지 ──

🔊 AI: "좋아요, 영상을 확인하고 있어요."

   (Gemini API가 YouTube URL을 직접 분석 → 인물 바운딩 박스 추출)

📺 [화면: 감지된 인물 얼굴/형태 크롭 이미지 그리드]
   "이 중에서 민수는 누구예요?"

   [사용자: 인물 크롭 이미지 터치 선택]

── Step 4 (선택): 추가 데이터 ──

🔊 AI: "사진이나 음성 메시지가 더 있으면 더 정확하게
        민수를 만날 수 있어요. 없어도 괜찮아요."

📷 [사용자: 사진 추가 선택 (선택사항)]
🎤 [사용자: 음성메시지 공유 (선택사항)]

── Step 5: YouTube URL 직접 분석 + 실시간 진행 UX ──

   ⚡ YouTube URL을 Gemini API에 직접 전달 — 다운로드/임시저장 불필요!
   (비공개 영상 fallback: 갤러리 업로드 → Cloud Storage → File API)

   ── 분석 Sequential Pipeline + 실시간 피드백 ──

🔊 AI: "민수를 알아가고 있어요..."

📺 [화면: 실시간 진행 상황 카드]
   ┌─────────────────────────────────────┐
   │ 🎬 영상 1/2 분석 중...              │
   │ ▓▓▓▓▓▓▓▓▓░░░░░ 62%                │
   │                                      │
   │ 💬 "야, 이리 와봐" — 말투 감지       │
   │ 😊 장난스러운 표정 발견              │
   │ 🖼️ [2~3초 하이라이트 클립 재생]      │
   │ 📸 [추출된 표정 이미지 표시]          │
   └─────────────────────────────────────┘

   VideoAnalyzer (gemini-3.1-pro + Video Understanding:
     YouTube URL 직접 분석으로 음성 전사, 말투 패턴, 표정, 행동, 성격 추출)
   → PersonaBuilder (성격 프로파일 + 말투 패턴 구축)
   → VoiceMatcher (30 HD 음성 중 최적 매칭)
   → 페르소나 데이터만 영구 보관 (원본 영상은 YouTube에 그대로 존재)

🔊 AI: "장난기 많고, '야'로 말을 시작하고, 다정한 면이 있네요."
       "이 목소리가 민수랑 비슷한가요?"

🔊 [AI: 매칭된 HD 음성으로 샘플 재생]

🎤 사용자: "어, 좀 비슷한 것 같아요"

🔊 AI: "좋아요. 민수를 만나러 가볼까요?"
```

### Phase 2: 가상 재회 (AI-Driven Fluid Stream)

```
🔊 AI: "눈 감아보세요..."

   [3초 침묵 + 부드러운 앰비언트 사운드]

🖼️ [화면: 서서히 밝아지며 — 겨울 저녁 카페 창가]
   (gemini-3-pro-image가 generate_scene Tool로 생성)

🎵 [BGM: 잔잔한 어쿠스틱 기타]

🔊 [민수 음성 — Live API Affective Dialog]:
   "야, 드디어 왔네. 여기 앉아. 뭐 마실래?"

   ── AI가 3초 대기 (사용자 반응 대기) ──

   [사용자가 아무 말 안 하면 → AI가 먼저 이어감]
🔊 [민수]: "왜, 할 말 있어? 표정 보니까 뭔가 있는데."
   (analyze_user Tool → gemini-3-flash Vision이 표정 분석)

   [사용자가 말하면]
🎤 사용자: "아메리카노..."
🔊 [민수]: "추운데 무슨 아메리카노야. 따뜻한 초코 시켜줄게."

🖼️ [화면 전환: 테이블 위 핫초코 두 잔]
   (generate_fast_scene Tool → gemini-2.5-flash-image로 빠른 생성)

🔊 [민수]: "있잖아, 나 요즘 훈련 진짜 힘든데..."
   "근데 네 사진 보면서 버텨."

🔊 [민수]: "이 사진 기억나?"
   (recall_memory Tool → Firestore에서 업로드된 사진 검색)

🖼️ [화면: 둘의 추억 장면 — 업로드 사진 기반 AI 재구성]
   (gemini-3-pro-image가 업로드 사진 스타일로 새 장면 생성)

🎵 [BGM 전환: 더 감성적인 피아노]
   (change_atmosphere Tool)

🎤 사용자: (울먹이며) "보고싶었어..."

🔊 [민수 — Affective Dialog 감지, 부드러운 톤]:
   "나도... 진짜 많이."

🖼️ [화면: 창밖으로 눈이 내리는 장면, 두 사람 나란히 앉은 뒷모습]

   ── 자연스럽게 흘러가다가 ──

🔊 [민수]: "다음에 또 만나자. 기다릴게."

🖼️ [화면: 카페를 나서며 손 흔드는 장면]
🎵 [BGM 페이드아웃]

   (end_reunion Tool → 앨범 자동 생성)

🔊 AI(missless 시스템 음성): "오늘 민수와 8분 23초 함께했어요.
   앨범이 저장되었습니다."
```

### Phase 3: 앨범 & 공유

- 재회 중 생성된 장면 이미지 자동 저장
- "오늘의 재회" 요약 카드 이미지 생성
- Instagram/YouTube Shorts/WhatsApp 등 글로벌 플랫폼 공유 최적화
- 다음 재회 제안: "다음에는 벚꽃 산책 어때요?"

---

## 7. Go 기술 구현

### SDK & 의존성 (2026.02.24 최신)

```
google.golang.org/genai v1.29.0+     // Google Gen AI Go SDK (GA). 개발 시점에 최신 버전 확인 필수.
                                      // 확인: https://pkg.go.dev/google.golang.org/genai
                                      // 릴리스: https://github.com/googleapis/go-genai/releases
google.golang.org/adk v0.4.0        // Agent Development Kit for Go (2026.01.30 릴리스)
                                      // 확인: https://github.com/google/adk-go/releases
github.com/gorilla/websocket v1.5.3  // WebSocket 서버
cloud.google.com/go/firestore v1.17  // Firestore 클라이언트
cloud.google.com/go/storage v1.47    // Cloud Storage 클라이언트
google.golang.org/api/youtube/v3     // YouTube Data API v3 클라이언트
golang.org/x/oauth2                  // OAuth 2.0 (Google 로그인 + YouTube 인증)
golang.org/x/oauth2/google           // Google OAuth 설정
```

> **참고**: `github.com/google/generative-ai-go`는 레거시. `google.golang.org/genai`가 후속 통합 SDK로 전체 기능을 지원한다. 2026년 6월 이후 레거시 SDK에는 신규 기능이 추가되지 않으므로 사용하지 않는다.

### ADK Go 에이전트 정의 (sail-researcher 패턴)

```go
package agent

import (
    _ "embed"

    "google.golang.org/adk/agent"
    "google.golang.org/adk/agent/llmagent"
    "google.golang.org/genai"
)

//go:embed prompts/reunion_host.md
var reunionHostInstruction string

//go:embed prompts/onboarding_guide.md
var onboardingGuideInstruction string

// constructReunionAgent는 메인 재회 에이전트를 생성한다.
// sail-researcher의 constructAgent() 팩토리 패턴을 따른다.
func constructReunionAgent(persona *Persona) agent.Agent {
    return llmagent.New(llmagent.Config{
        Name:        "reunion-host",
        Model:       "gemini-live-2.5-flash-native-audio",
        Instruction: buildPersonaInstruction(reunionHostInstruction, persona),
        Tools: []agent.Tool{
            newGenerateSceneTool(),
            newGenerateFastSceneTool(),
            newChangeAtmosphereTool(),
            newRecallMemoryTool(),
            newAnalyzeUserTool(),
            newEndReunionTool(),
        },
        Temperature:    0.3,
        MaxOutputTokens: 65536,
        // ADK v0.4.0: BeforeModelCallback으로 텔레메트리 주입
        BeforeModelCallback: toolMonitorBefore,
        AfterModelCallback:  toolMonitorAfter,
    })
}

// constructOnboardingPipeline은 온보딩 Sequential 에이전트를 생성한다.
// llm-auditor의 순차 파이프라인 패턴을 따른다.
func constructOnboardingPipeline() agent.Agent {
    // Stage 1: YouTube 영상 분석 — YouTube URL을 Gemini API에 직접 전달하여 분석
    // 핵심: genai.FileData{FileURI: "https://youtube.com/watch?v=..."} 로
    //       다운로드 없이 영상 내용을 직접 분석한다.
    videoAnalyzer := llmagent.New(llmagent.Config{
        Name:  "video-analyzer",
        Model: "gemini-3.1-pro-preview",
        Instruction: `YouTube 영상에서 대상 인물을 분석하라.
            - 음성 전사: 말투 패턴, 자주 쓰는 표현, 어미 습관 추출
            - 표정/행동: 감정 표현 방식, 제스처, 리액션 패턴 추출
            - 성격 특성: 외향/내향, 유머 감각, 다정함 등 성격 요소 추론
            - 하이라이트: 인상적인 장면 타임스탬프와 2~3초 구간 표시
            복수 영상이 있으면 크로스 분석하여 일관된 패턴을 추출하라.`,
        Temperature: 0.2,
    })

    // Stage 2: 페르소나 프로파일 구축
    personaBuilder := llmagent.New(llmagent.Config{
        Name:  "persona-builder",
        Model: "gemini-3.1-pro-preview",
        Instruction: `추출된 영상 분석 결과를 바탕으로 완전한 페르소나 정의를 생성하라.
            성격, 말투, 자주 쓰는 표현, 감정 반응 패턴을 포함하라.
            사진/음성 데이터가 추가로 있으면 통합하여 정밀도를 높여라.`,
        Temperature: 0.2,
    })

    // Stage 3: HD 음성 매칭
    voiceMatcher := llmagent.New(llmagent.Config{
        Name:        "voice-matcher",
        Model:       "gemini-3-flash-preview",
        Instruction: "페르소나의 성별, 나이대, 성격에 가장 적합한 HD 음성을 30개 프리셋 중에서 선택하라.",
        Temperature: 0.1,
    })

    // SequentialAgent: videoAnalyzer → personaBuilder → voiceMatcher
    // **QPS 병목 대응**: 각 단계 사이에 Exponential Backoff + Jitter를 적용하여
    // 연달아 호출되는 모델 요청이 429 Too Many Requests에 걸리지 않도록 한다.
    // ADK BeforeModelCallback에서 rate limiter를 주입하거나,
    // 각 에이전트의 호출 사이에 최소 500ms 간격을 보장한다.
    return agent.NewSequentialAgent(videoAnalyzer, personaBuilder, voiceMatcher)
}

// retryWithBackoff는 Gemini API 호출에 Exponential Backoff + Jitter를 적용한다.
// Sequential Agent의 연속 호출 및 이미지 생성 Tool에서 429 에러를 방어한다.
func retryWithBackoff(ctx context.Context, maxRetries int, fn func() error) error {
    for i := 0; i < maxRetries; i++ {
        if err := fn(); err == nil {
            return nil
        } else if i < maxRetries-1 {
            base := time.Duration(1<<uint(i)) * time.Second // 1s, 2s, 4s
            jitter := time.Duration(rand.Int63n(int64(base / 2)))
            slog.Warn("api_retry", "attempt", i+1, "backoff", base+jitter)
            time.Sleep(base + jitter)
        } else {
            return err
        }
    }
    return nil
}
```

### YouTube URL 직접 분석 (제로 다운로드 아키텍처)

```go
package media

import (
    "context"
    "log/slog"

    "google.golang.org/genai"
)

// analyzeYouTubeVideo는 YouTube URL을 Gemini API에 직접 전달하여 영상을 분석한다.
// 영상 다운로드/임시저장 없이 URL만으로 Video Understanding 분석이 완료된다.
// 공개 및 일부공개(Unlisted) 영상만 지원. 비공개 영상은 File API fallback 사용.
func analyzeYouTubeVideo(ctx context.Context, client *genai.Client, videoURL string, targetPerson string) (*VideoAnalysis, error) {
    slog.Info("youtube_url_analysis", "url", videoURL, "target", targetPerson)

    resp, err := client.Models.GenerateContent(ctx,
        "gemini-3.1-pro-preview",
        // YouTube URL을 FileData로 직접 전달 — 다운로드 불필요!
        genai.FileData{
            FileURI:  videoURL, // e.g. "https://www.youtube.com/watch?v=VIDEO_ID"
            MIMEType: "video/mp4",
        },
        genai.Text(buildVideoAnalysisPrompt(targetPerson)),
        &genai.GenerateContentConfig{
            Temperature: genai.Ptr(float32(0.2)),
        },
    )
    if err != nil {
        slog.Error("youtube_analysis_failed", "error", err, "url", videoURL)
        return nil, err
    }

    return parseVideoAnalysis(resp), nil
}

// analyzeUploadedVideo는 File API를 통해 업로드된 영상을 분석한다.
// YouTube 비공개 영상이나 디바이스 갤러리 영상에 대한 fallback 경로.
func analyzeUploadedVideo(ctx context.Context, client *genai.Client, fileURI string, targetPerson string) (*VideoAnalysis, error) {
    slog.Info("file_api_analysis", "file_uri", fileURI, "target", targetPerson)

    resp, err := client.Models.GenerateContent(ctx,
        "gemini-3.1-pro-preview",
        genai.FileData{
            FileURI:  fileURI, // Gemini File API에서 반환된 URI
            MIMEType: "video/mp4",
        },
        genai.Text(buildVideoAnalysisPrompt(targetPerson)),
        &genai.GenerateContentConfig{
            Temperature: genai.Ptr(float32(0.2)),
        },
    )
    if err != nil {
        return nil, err
    }

    return parseVideoAnalysis(resp), nil
}
```

### Live API + Tool 등록 (genai v1.29.0+)

```go
package live

import (
    "context"
    "os"

    "google.golang.org/genai"
)

func startReunionSession(ctx context.Context, persona *Persona) (*genai.Session, error) {
    client, err := genai.NewClient(ctx, &genai.ClientConfig{
        APIKey:  os.Getenv("GEMINI_API_KEY"),
        Backend: genai.BackendGeminiAPI,
    })
    if err != nil {
        return nil, err
    }

    config := &genai.LiveConnectConfig{
        SystemInstruction: &genai.Content{
            Parts: []genai.Part{genai.Text(buildPersonaPrompt(persona))},
        },

        ResponseModalities: []genai.Modality{genai.ModalityAudio},

        // Affective Dialog: 사용자 음성 톤에서 감정 인지 → 페르소나 응답에 반영
        EnableAffectiveDialog: true,

        // Proactive Audio: 사용자가 직접 말할 때만 응답, 배경 소음 무시
        ProactiveAudio: true,

        SpeechConfig: &genai.SpeechConfig{
            // LanguageCode: BCP-47 코드로 응답 언어 명시적 설정.
            // 자동 언어 감지의 불안정성을 방지하기 위해 System Instruction과 병행.
            LanguageCode: persona.LanguageCode, // e.g. "ko-KR", "en-US"
            VoiceConfig: &genai.VoiceConfig{
                PrebuiltVoiceConfig: &genai.PrebuiltVoiceConfig{
                    VoiceName: persona.MatchedVoice,
                },
            },
        },

        // 입출력 오디오 전사 — 대화 로그 및 디버깅용
        InputAudioTranscription:  &genai.InputAudioTranscription{},
        OutputAudioTranscription: &genai.OutputAudioTranscription{},

        Tools: []*genai.Tool{
            // Google Search Grounding — 실시간 정보(날씨, 뉴스 등) 대화 반영
            {GoogleSearch: &genai.GoogleSearch{}},
            {FunctionDeclarations: []*genai.FunctionDeclaration{
                {
                    Name:        "generate_scene",
                    Description: "새로운 장면 이미지를 생성한다. 대화의 전환점이나 새로운 상황이 시작될 때 호출한다.",
                    Parameters: &genai.Schema{
                        Type: genai.TypeObject,
                        Properties: map[string]*genai.Schema{
                            "prompt":     {Type: genai.TypeString, Description: "장면 묘사"},
                            "mood":       {Type: genai.TypeString, Description: "분위기 (warm, romantic, nostalgic, playful)"},
                            "characters": {Type: genai.TypeString, Description: "등장인물 묘사"},
                        },
                    },
                },
                {
                    Name:        "generate_fast_scene",
                    Description: "작은 변화가 필요한 장면을 빠르게 생성한다. 소품 추가, 날씨 변화 등 미세한 장면 변경 시 호출한다.",
                    Parameters: &genai.Schema{
                        Type: genai.TypeObject,
                        Properties: map[string]*genai.Schema{
                            "prompt": {Type: genai.TypeString, Description: "장면 변경 묘사"},
                        },
                    },
                },
                {
                    Name:        "change_atmosphere",
                    Description: "분위기를 전환한다. 감정적 변화가 감지되거나 대화 톤이 바뀔 때 호출한다.",
                    Parameters: &genai.Schema{
                        Type: genai.TypeObject,
                        Properties: map[string]*genai.Schema{
                            "mood":      {Type: genai.TypeString, Description: "새 분위기"},
                            "intensity": {Type: genai.TypeString, Description: "강도 (subtle, moderate, dramatic)"},
                        },
                    },
                },
                {
                    Name:        "recall_memory",
                    Description: "특정 추억이나 기억을 회상할 때 호출한다. 업로드된 사진이나 대화에서 관련 내용을 검색한다.",
                    Parameters: &genai.Schema{
                        Type: genai.TypeObject,
                        Properties: map[string]*genai.Schema{
                            "topic": {Type: genai.TypeString, Description: "회상할 주제/키워드"},
                        },
                    },
                },
                {
                    Name:        "analyze_user",
                    Description: "사용자의 현재 상태를 분석한다. 주기적으로 또는 중요한 순간에 호출하여 표정과 감정을 파악한다.",
                    Parameters: &genai.Schema{
                        Type: genai.TypeObject,
                        Properties: map[string]*genai.Schema{
                            "reason": {Type: genai.TypeString, Description: "분석 이유"},
                        },
                    },
                },
                {
                    Name:        "end_reunion",
                    Description: "재회를 자연스럽게 마무리한다. 대화가 자연스런 끝에 도달했거나 사용자가 종료 의사를 밝혔을 때 호출한다.",
                    Parameters: &genai.Schema{
                        Type: genai.TypeObject,
                        Properties: map[string]*genai.Schema{
                            "summary": {Type: genai.TypeString, Description: "재회 요약"},
                        },
                    },
                },
            },
        }}, // FunctionDeclarations Tool 끝

        ContextWindowCompression: &genai.ContextWindowCompressionConfig{
            SlidingWindow: &genai.SlidingWindow{
                TargetTokenCount: 8192,   // 압축 후 유지할 토큰 수
                TriggerTokenCount: 12288, // 이 토큰 수 초과 시 압축 시작
            },
            // prefixTurns: System Instruction과 초기 N턴을 압축에서 보호
            // 페르소나 성격/말투 설정이 유실되지 않도록 보장
        },

        SessionResumption: &genai.SessionResumptionConfig{
            // 토큰 유효시간: 2시간. GoAway 시그널 수신 시 자동 재접속에 활용.
        },
    }

    session, err := client.Live.Connect(ctx, "gemini-live-2.5-flash-native-audio", config)
    return session, err
}
```

### Tool 실행 핸들러 (navallist tools.go 패턴)

```go
package live

import (
    "context"
    "log/slog"
    "time"

    "google.golang.org/genai"
)

// handleToolCall은 Live API가 호출한 Tool을 실행하고 결과를 반환한다.
// **비동기 패턴**: 이미지 생성 Tool은 즉시 Dummy 응답을 반환하여 음성 스트리밍을
// 재개하고, 실제 이미지는 별도 WebSocket 채널로 비동기 전송한다.
// Live API는 FunctionResponse가 올 때까지 음성을 일시정지하므로,
// 이미지 생성(1~12초)이 대화를 차단하는 것을 방지하는 핵심 아키텍처이다.
//
// **에러 핸들링 파이프라인**: 비동기 이미지 생성 실패 시 프론트엔드 WebSocket으로
// 에러 이벤트를 전송하고, 프론트가 Live API 세션에 Client Content를 주입하여
// 에이전트가 "이미지 생성에 문제가 생겼네요"라고 자연스럽게 안내한다.
func handleToolCall(ctx context.Context, call *genai.FunctionCall, engines *Engines) *genai.FunctionResponse {
    start := time.Now()
    defer func() {
        slog.Info("tool_executed",
            "tool", call.Name,
            "latency_ms", time.Since(start).Milliseconds(),
            "session_id", engines.sessionID,
        )
    }()

    switch call.Name {
    case "generate_scene":
        prompt := call.Args["prompt"].(string)
        mood := call.Args["mood"].(string)

        // 비동기 이미지 생성: goroutine으로 분리하여 즉시 FunctionResponse 반환
        go func() {
            resp, err := engines.imageClient.Models.GenerateContent(ctx,
                "gemini-2.5-flash-image", // flash-image 메인 (1~3초)
                genai.Text(buildScenePrompt(prompt, mood, engines.persona)),
                &genai.GenerateContentConfig{
                    ResponseModalities: []string{
                        string(genai.ModalityImage),
                        string(genai.ModalityText),
                    },
                },
            )
            if err != nil {
                slog.Error("generate_scene_failed", "error", err)
                // 에러 핸들링: 프론트에 에러 이벤트 전송 → 에이전트가 자연스럽게 안내
                engines.wsConn.WriteJSON(map[string]any{
                    "type":    "tool_error",
                    "tool":    "generate_scene",
                    "message": "이미지 생성에 문제가 생겼습니다.",
                })
                return
            }
            imageData := extractImage(resp)
            engines.wsConn.WriteMessage(websocket.BinaryMessage, imageData)
        }()

        // 즉시 Dummy 응답 반환 → Live API 음성 스트리밍 즉시 재개
        return &genai.FunctionResponse{
            Name:     "generate_scene",
            Response: map[string]any{"status": "generating", "description": prompt},
        }

    case "generate_fast_scene":
        prompt := call.Args["prompt"].(string)

        // 비동기 이미지 생성 (flash-image: 1~3초)
        go func() {
            resp, err := engines.fastImageClient.Models.GenerateContent(ctx,
                "gemini-2.5-flash-image",
                genai.Text(prompt),
                &genai.GenerateContentConfig{
                    ResponseModalities: []string{string(genai.ModalityImage)},
                },
            )
            if err != nil {
                slog.Error("generate_fast_scene_failed", "error", err)
                engines.wsConn.WriteJSON(map[string]any{
                    "type":    "tool_error",
                    "tool":    "generate_fast_scene",
                    "message": "이미지 생성에 문제가 생겼습니다.",
                })
                return
            }
            imageData := extractImage(resp)
            engines.wsConn.WriteMessage(websocket.BinaryMessage, imageData)
        }()

        // 즉시 Dummy 응답 반환
        return &genai.FunctionResponse{
            Name:     "generate_fast_scene",
            Response: map[string]any{"status": "generating"},
        }

    case "change_atmosphere":
        mood := call.Args["mood"].(string)
        intensity := call.Args["intensity"].(string)
        engines.wsConn.WriteJSON(map[string]string{
            "type": "bgm_change", "mood": mood, "intensity": intensity,
        })
        return &genai.FunctionResponse{
            Name:     "change_atmosphere",
            Response: map[string]any{"status": "atmosphere_changed"},
        }

    case "recall_memory":
        topic := call.Args["topic"].(string)
        memories, err := engines.firestore.SearchMemories(ctx, engines.persona.ID, topic)
        if err != nil {
            slog.Error("recall_memory_failed", "error", err)
            return &genai.FunctionResponse{
                Name:     "recall_memory",
                Response: map[string]any{"memories": []string{}},
            }
        }
        return &genai.FunctionResponse{
            Name:     "recall_memory",
            Response: map[string]any{"memories": memories},
        }

    case "analyze_user":
        frame := engines.latestCameraFrame
        if frame == nil {
            return &genai.FunctionResponse{
                Name:     "analyze_user",
                Response: map[string]any{"emotion": "unknown", "reason": "no_camera_frame"},
            }
        }
        resp, err := engines.flashClient.Models.GenerateContent(ctx,
            "gemini-3-flash-preview",
            genai.ImageData("jpeg", frame),
            genai.Text("이 사용자의 표정과 감정 상태를 분석해주세요."),
            &genai.GenerateContentConfig{
                ThinkingConfig: &genai.ThinkingConfig{
                    ThinkingLevel: genai.ThinkingLevelLow,
                },
            },
        )
        if err != nil {
            return &genai.FunctionResponse{
                Name:     "analyze_user",
                Response: map[string]any{"emotion": "unknown"},
            }
        }
        return &genai.FunctionResponse{
            Name:     "analyze_user",
            Response: map[string]any{"emotion": extractText(resp)},
        }

    case "end_reunion":
        summary := call.Args["summary"].(string)
        go engines.albumGen.CreateAlbum(ctx, engines.session, summary)
        return &genai.FunctionResponse{
            Name:     "end_reunion",
            Response: map[string]any{"status": "reunion_ended"},
        }
    }

    return nil
}
```

### Tool Monitor (sail-researcher tool_monitor.go 패턴)

```go
package agent

import (
    "context"
    "log/slog"
    "time"
)

// toolMonitorBefore는 ADK v0.4.0 BeforeModelCallback.
// 모든 모델 호출 전 컨텍스트에 시작 시간을 기록한다.
func toolMonitorBefore(ctx context.Context, req *CallbackRequest) (*CallbackResponse, error) {
    slog.Info("model_request",
        "agent", req.AgentName,
        "tools_available", len(req.Tools),
    )
    return nil, nil // nil = 정상 진행
}

// toolMonitorAfter는 AfterModelCallback.
// 모델 응답 후 Tool 호출 결과를 로깅한다.
func toolMonitorAfter(ctx context.Context, resp *CallbackResponse) (*CallbackResponse, error) {
    if resp.ToolCalls != nil {
        for _, tc := range resp.ToolCalls {
            slog.Info("tool_invoked",
                "tool", tc.Name,
                "args", tc.Args,
                "timestamp", time.Now().Format(time.RFC3339),
            )
        }
    }
    return nil, nil
}
```

### 프로젝트 구조

```
missless/
├── cmd/server/main.go                  // 엔트리포인트
├── internal/
│   ├── agent/
│   │   ├── reunion_agent.go            // 재회 에이전트 팩토리 (sail-researcher 패턴)
│   │   ├── onboarding_pipeline.go      // 온보딩 Sequential (llm-auditor 패턴)
│   │   ├── tool_monitor.go             // 텔레메트리 콜백 (sail-researcher 패턴)
│   │   └── prompts/
│   │       ├── reunion_host.md         // 재회 호스트 System Instruction
│   │       ├── onboarding_guide.md     // 온보딩 가이드 Instruction
│   │       └── persona_template.md     // 페르소나 프로파일 템플릿
│   ├── live/
│   │   ├── session.go                  // Live API 세션 관리
│   │   ├── reconnect.go               // GoAway 감지 + 자동 재접속
│   │   ├── tools.go                    // Tool 핸들러 (navallist 패턴)
│   │   └── stream.go                   // 오디오/이미지 스트림 관리
│   ├── persona/
│   │   ├── builder.go                  // 페르소나 생성
│   │   ├── prompt.go                   // System Instruction 동적 합성
│   │   └── voice_matcher.go            // HD 음성 매칭
│   ├── scene/
│   │   ├── generator.go                // 장면 이미지 생성 (3 Pro Image)
│   │   ├── fast_generator.go           // 빠른 이미지 생성 (2.5 Flash Image)
│   │   └── album.go                    // 앨범 생성기
│   ├── auth/
│   │   ├── oauth.go                    // Google OAuth 2.0 (youtube.readonly scope)
│   │   └── session.go                  // 사용자 세션 관리
│   ├── youtube/
│   │   ├── client.go                   // YouTube Data API v3 클라이언트
│   │   └── playlist.go                 // 채널 영상 목록 조회 (channels→playlistItems)
│   ├── media/
│   │   ├── video_analyzer.go           // YouTube URL 직접 분석 (genai.FileData{FileURI}) + Fallback 갤러리 분석
│   │   ├── person_detector.go          // 영상 프레임 인물 감지 + 바운딩 박스 크롭
│   │   ├── privacy_checker.go          // YouTube 영상 공개상태 확인 (공개✅/비공개🔒 분류)
│   │   ├── upload_pipeline.go          // 갤러리 Fallback: 업로드 → Cloud Storage 임시 → File API → 분석 → 삭제
│   │   └── progress_streamer.go        // 분석 진행률 + 하이라이트 실시간 스트리밍
│   ├── memory/
│   │   └── store.go                    // Firestore 추억 저장/검색
│   ├── handler/
│   │   ├── websocket.go                // WebSocket 서버 (navallist 패턴)
│   │   ├── token.go                    // Ephemeral Token 발급 API
│   │   ├── upload.go                   // 미디어 업로드 API
│   │   └── health.go                   // 헬스체크
│   └── bgm/
│       └── manager.go                  // BGM 관리 (프리셋 + 크로스페이드)
├── web/                                 // Next.js 15 프론트엔드
│   ├── app/
│   │   ├── page.tsx                    // 랜딩
│   │   ├── reunion/[id]/page.tsx       // 재회 경험 (전체화면 몰입)
│   │   └── album/[id]/page.tsx         // 앨범 보기/공유
│   ├── components/
│   │   ├── ImmersiveCanvas.tsx         // 전체화면 이미지 + 크로스페이드
│   │   ├── AudioEngine.tsx             // AudioWorklet 기반 음성 재생
│   │   ├── BGMPlayer.tsx               // BGM 관리
│   │   ├── CameraFeed.tsx              // 카메라 캡처 (표정 분석용)
│   │   ├── YouTubeVideoGrid.tsx        // YouTube 영상 목록 썸네일 그리드
│   │   ├── PersonSelector.tsx          // 감지된 인물 크롭 이미지 선택 UI
│   │   ├── AnalysisProgress.tsx        // 영상 분석 진행률 + 실시간 피드백 카드
│   │   ├── SubtitleOverlay.tsx         // 선택적 자막 (ON/OFF)
│   │   └── ShareCard.tsx               // 공유 카드 생성
│   ├── lib/
│   │   ├── websocket.ts                // WebSocket 클라이언트 + tool_error 핸들링
│   │   ├── ephemeral-token.ts          // Ephemeral Token 획득 + Live API 연결
│   │   ├── audio-worklet.ts            // PCM 오디오 처리 + audioStreamEnd
│   │   └── error-handler.ts            // 비동기 Tool 에러 → Client Content 주입
│   └── public/
│       ├── manifest.json               // PWA
│       ├── sw.js                       // Service Worker
│       └── bgm/                        // BGM 프리셋 파일
├── deploy/
│   ├── Dockerfile                      // 멀티스테이지 빌드
│   ├── cloudbuild.yaml                 // Cloud Build CI/CD
│   └── terraform/
│       ├── main.tf
│       ├── variables.tf
│       └── outputs.tf
├── go.mod
├── go.sum
└── README.md                            // spin-up 가이드 (필수)
```

---

## 8. Google Cloud 서비스

| 서비스 | 용도 | 필수 여부 |
|--------|------|:---:|
| **Cloud Run** | Go 백엔드 + Next.js 호스팅 | ✅ 필수 |
| **Cloud Firestore** | 페르소나 프로필, 대화 기억, 세션 이력 | ✅ |
| **Cloud Storage** | 갤러리 Fallback 영상 (임시), 생성 이미지, 앨범 | ✅ |
| **Gemini API** | 5개 모델 + YouTube URL 직접 분석 (Video Understanding) + File API (Fallback) | ✅ 필수 |
| **Google Identity Platform** | Google OAuth 2.0 로그인 + YouTube Data API v3 인증 | ✅ |
| **YouTube Data API v3** | 사용자 채널 영상 목록 조회 (메타데이터+썸네일) | ✅ |
| **Cloud Build** | CI/CD 자동 빌드/배포 | ✅ |
| **Artifact Registry** | Docker 이미지 저장 | ○ |

> GCP 무료 크레딧 $300 + Gemini API Free Tier 활용.

---

## 9. 데모 영상 전략 (3분 50초)

> "Judges are not required to test the Project and may choose to judge based solely on the video."

**이것이 가장 중요한 제출물이다.**

### 영상 구성

```
0:00 ~ 0:15  [BLACK + 타이포그래피]
  "마지막으로 그 사람과 대화한 게 언제예요?"

0:15 ~ 0:25  [실제 사람이 스마트폰을 들고 있는 장면]
  "만약, 지금 바로 만날 수 있다면?"
  → missless.co 로고 + 태그라인

0:25 ~ 0:55  [온보딩 시연]
  실제 스마트폰 화면 녹화
  Google 로그인 → YouTube 영상 목록 (공개✅ 상태 표시)
  영상 터치 → Gemini API가 YouTube URL 직접 분석 (다운로드 ZERO!)
  인물 감지 → 대상 인물 터치 선택
  AI 분석 중 실시간 하이라이트/진행률 표시
  텍스트 입력 없이 전부 음성 + 터치로 진행

0:55 ~ 2:30  [가상 재회 경험 — 핵심 시연]
  "군대 간 남자친구와 카페에서 만남"
  - AI가 먼저 말을 건다 (AI-first)
  - 장면 이미지가 실시간 생성/전환 (Creative Storyteller 핵심)
  - 사용자와 자연스러운 음성 대화
  - 사용자가 울먹이면 AI가 감정 인지하여 반응 (Affective Dialog)
  - BGM이 분위기에 따라 자동 전환
  → 텍스트 입력 0, 버튼 클릭 0을 강조

2:30 ~ 3:05  [기술 설명]
  아키텍처 다이어그램 (GenMedia Live 패턴 + 제로 다운로드)
  5개 Gemini 모델 + YouTube URL 직접 분석 역할 설명
  "Google Login → YouTube URL을 Gemini API에 직접 전달 (다운로드 불필요!)
   → Video Understanding으로 인물 분석 → 페르소나 자동 생성
   → Live API가 음성 컨트롤 레이어로 동작하며,
   이미지 생성을 Tool로 등록하여 자동 호출"
  올 Google 생태계 (OAuth + YouTube API + Gemini + Cloud Run)

3:05 ~ 3:30  [다양한 시나리오 몽타주]
  "할머니와 추석" / "딸과 생일" / "친구와 졸업식"
  빠르게 3~4개 시나리오 스위칭

3:30 ~ 3:50  [클로징]
  missless.co QR 코드
  "그리움을 줄여드립니다."
  #GeminiLiveAgentChallenge
```

---

## 10. 채점 항목별 대응

### Innovation & Multimodal UX (40%)

| 평가 질문 | missless 대응 |
|----------|--------------|
| 텍스트 박스 탈피 | 전체 경험에서 텍스트 입력 ZERO. 온보딩도 음성 대화. |
| See/Hear/Speak 유동성 | 이미지(See)+음성(Hear)+대화(Speak)가 동시에 끊김 없이 흐름. |
| 독특한 성격/목소리 | 사용자가 학습시킨 실제 사람의 성격+말투+HD 음성 매핑. |
| Live & context-aware | Affective Dialog(감정 톤) + Proactive Audio(선택적 응답) + Vision(표정). |

### Technical Implementation (30%)

| 평가 질문 | missless 대응 |
|----------|--------------|
| GenAI SDK/ADK 활용 | google.golang.org/genai (최신) + ADK v0.4.0. 5개 모델 + YouTube URL 직접 분석 (Video Understanding) + Context Caching. Sequential/LlmAgent 패턴. |
| Google Cloud 배포 | Cloud Run + Firestore + Storage + Identity Platform. Terraform IaC 포함. |
| Google 생태계 통합 | Google OAuth + YouTube Data API v3 + Gemini API + Cloud Run. 올 Google 스택. |
| 에이전트 로직 | GenMedia Live 패턴 — Live API Tool Calling으로 자동 오케스트레이션. ADK 에이전트 구조. |
| 할루시네이션 방지 | YouTube URL 직접 분석 기반 페르소나 System Instruction + recall_memory 그라운딩. |

### Demo & Presentation (30%)

| 평가 질문 | missless 대응 |
|----------|--------------|
| 문제-솔루션 명확 | 15초 감성 인트로 → 즉시 이해 가능. |
| 아키텍처 다이어그램 | GenMedia Live 패턴 + 5개 모델 파이프라인 + ADK 에이전트 구조 시각화. |
| Cloud 배포 증거 | Cloud Run 콘솔 스크린샷 + 라이브 URL + QR. |
| 실제 동작 | 실제 감정 반응이 포함된 리얼 시연. |

---

## 11. 20일 스프린트

### Week 1 (2/25 ~ 3/2): Live API + 이미지 파이프라인

| 일자 | 작업 | 검증 기준 |
|------|------|----------|
| 2/25(화) | GCP 프로젝트 + API 키 5개 모델 발급 + Go 프로젝트 초기화 (genai 최신 + ADK v0.4.0) | 모든 모델 API 호출 성공 |
| 2/26(수) | **Live API PoC** — WebSocket 양방향 음성 + Tool 등록 | 음성 대화 + Tool 자동 호출 확인 |
| 2/27(목) | **gemini-3-pro-image interleaved PoC** — Tool에서 이미지 생성 | Tool 호출 → 이미지 반환 → 프론트 표시 |
| 2/28(금) | Next.js 15 PWA 초기화 + ImmersiveCanvas + AudioEngine | 전체화면 이미지 + 오디오 재생 |
| 3/1(토) | Live API + 이미지 Tool 통합 — **Fluid Stream 구현** | 음성 대화 중 이미지 자동 생성/표시 |
| 3/2(일) | 카메라 피드 + analyze_user Tool + tool_monitor 텔레메트리 | 표정 분석 + slog 로깅 확인 |

### Week 2 (3/3 ~ 3/9): 페르소나 + 경험 완성

| 일자 | 작업 | 검증 기준 |
|------|------|----------|
| 3/3(월) | Google OAuth 2.0 + YouTube Data API v3 연동. 영상 목록 조회 + 공개상태 표시(✅🔒) 썸네일 그리드 UI. | Google 로그인 → 영상 목록 + 공개상태 표시 성공 |
| 3/4(화) | **YouTube URL 직접 분석 PoC** — `genai.FileData{FileURI}` + Video Understanding. 갤러리 Fallback 파이프라인. 인물 감지/크롭 UI. | 공개 영상 URL → Gemini 직접 분석 성공. 비공개 → 갤러리 Fallback 성공 |
| 3/5(수) | ADK Sequential Agent 온보딩 (VideoAnalyzer→PersonaBuilder→VoiceMatcher) + 실시간 진행 UX | YouTube URL 분석 중 하이라이트/진행률 표시 + Memory Store |
| 3/6(목) | BGM 시스템 + change_atmosphere Tool | 분위기 전환 시 BGM 자동 변경 |
| 3/7(금) | Affective Dialog + Proactive Audio 설정 최적화 | 감정 인지 대화 + 선택적 응답 |
| 3/8(토) | Context Window Compression + Session Resumption | 10분+ 세션 안정성 + 끊김 복구 |
| 3/9(일) | E2E 통합 테스트 — 온보딩→재회→앨범 전체 플로우 | 전체 경험 끊김 없이 완주 |

### Week 3 (3/10 ~ 3/16): 배포 + 영상 + 제출

| 일자 | 작업 | 검증 기준 |
|------|------|----------|
| 3/10(월) | Cloud Run 배포 + Terraform IaC + 도메인 SSL | missless.co 라이브 접속 |
| 3/11(화) | UI 폴리싱 — 크로스페이드, 로딩 UX, 모바일 최적화 | 스마트폰에서 완벽 동작 |
| 3/12(수) | 데모 시나리오 리허설 + 영상 대본 확정 | 리허설 3회 이상 |
| 3/13(목) | **데모 영상 촬영** (실제 감정 반응 포함) | 3:50 이내 완성본 |
| 3/14(금) | 아키텍처 다이어그램 + README(spin-up 가이드) | GitHub 공개 |
| 3/15(토) | DevPost 제출 페이지 작성 | 제출 페이지 초안 완성 |
| 3/16(일) | **최종 점검 + 제출** (KST 3/17 09:00 마감) | ✅ 제출 완료 |

---

## 12. 데이터 저장 전략

### 제로 다운로드 데이터 (저장하지 않음)

| 데이터 | 처리 방식 | 비고 |
|--------|----------|------|
| **YouTube 공개/일부공개 영상** | Gemini API에 URL 직접 전달 → 분석. **다운로드/저장 없음.** | 원본은 YouTube에 그대로 존재. 서버 스토리지 미사용. |

### 임시 데이터 (분석 완료 후 삭제 — Fallback 경로만 해당)

| 데이터 | 저장 위치 | 보관 기간 | 삭제 방법 |
|--------|----------|----------|----------|
| 갤러리 업로드 영상 (비공개 fallback) | Cloud Storage `tmp-videos/` 버킷 | 분석 완료 즉시 | lifecycle policy + 명시적 삭제 |
| Gemini File API 업로드 (fallback) | Gemini File API | 48시간 자동 삭제 | 분석 후 수동 삭제 권장 |
| 분석 중간 결과 | 메모리 (Go 서버) | 세션 종료 시 | GC 자동 |

### 영구 데이터 (서비스에 필수)

| 데이터 | 저장 위치 | 내용 |
|--------|----------|------|
| 페르소나 프로파일 | Firestore `personas/` | 성격, 말투, 표현 패턴, 음성 매칭 결과 (JSON) |
| 생성된 장면 이미지 | Cloud Storage `albums/` | 재회 중 AI가 생성한 이미지 |
| 앨범 메타데이터 | Firestore `albums/` | 재회 요약, 생성일, 공유 링크 |
| 추억 데이터 | Firestore `memories/` | 영상 분석에서 추출한 핵심 장면/대사 (텍스트+타임스탬프) |
| 인물 크롭 이미지 | Cloud Storage `personas/faces/` | 인물 선택 시 추출한 얼굴/형태 크롭 (소량) |
| YouTube 영상 URL | Firestore `personas/` | 분석에 사용된 YouTube URL 참조 (재분석 시 활용) |

> **원칙**: YouTube URL 직접 분석으로 원본 영상 다운로드/저장을 원천 차단.
> AI가 추출/생성한 결과물만 영구 보관. 스토리지 비용 최소화 + 프라이버시 보호 + YouTube ToS 100% 준수.

---

## 13. 윤리 & 안전

1. **동의**: 페르소나 대상자의 본인 동의 또는 가족 동의 필수
2. **투명성**: "이것은 AI가 만든 가상 경험입니다" 세션 시작/종료 시 명시
3. **건강한 사용**: 30분 이상 연속 사용 시 알림 ("실제 소통도 중요해요")
4. **데이터 보호**: YouTube 원본 영상은 다운로드/저장 자체를 하지 않음 (제로 다운로드). 갤러리 Fallback 영상만 분석 후 즉시 삭제. 페르소나 데이터와 생성물만 암호화 보관. 사용자 삭제 요청 즉시 처리.
5. **SynthID**: Gemini 생성 이미지에 자동 SynthID 워터마크 포함
6. **YouTube ToS 준수**: YouTube URL을 Gemini API에 직접 전달하여 분석. 영상 바이너리 다운로드 원천 차단. YouTube Data API는 메타데이터 조회 + 공개상태 확인에만 사용.

---

## 14. 제출 체크리스트

### 필수 (감점 방지)

- [ ] Gemini 모델 사용 (5개)
- [ ] google.golang.org/genai SDK (최신) + ADK v0.4.0 + YouTube Data API v3
- [ ] Google Cloud 서비스 5개 (Cloud Run + Firestore + Storage + Identity Platform + YouTube API)
- [ ] Google Cloud 배포 증거 (콘솔 스크린샷 또는 라이브 URL)
- [ ] 아키텍처 다이어그램
- [ ] 데모 영상 4분 미만 (실제 동작, 목업 불가)
- [ ] README에 spin-up 가이드
- [ ] GitHub 공개 레포지토리
- [ ] Terraform IaC + cloudbuild.yaml 공개 레포 포함

### 차별화

- [ ] 전체 경험에서 텍스트 입력 ZERO 확인
- [ ] AI가 먼저 시작하고 주도하는 흐름 확인
- [ ] 이미지+음성+BGM Fluid Stream 끊김 없음 확인
- [ ] Affective Dialog (`enableAffectiveDialog: true`) + Proactive Audio 시연
- [ ] 데모 영상에 실제 감정 반응 포함
- [ ] ADK Sequential Agent 온보딩 파이프라인 시연
- [ ] tool_monitor 텔레메트리 로그 시연
- [ ] Ephemeral Token으로 API 키 미노출 보안 구현
- [ ] GoAway 자동 재접속 — 10분+ 세션에서 끊김 없음 확인
- [ ] audioStreamEnd 전송으로 응답 지연 없음 확인
- [ ] Context Window Compression (triggerTokenCount + prefixTurns) 장시간 세션 안정성
- [ ] Google OAuth + YouTube Data API 연동으로 올 Google 생태계 활용 시연
- [ ] **YouTube URL 직접 분석** (genai.FileData{FileURI}) — 다운로드 ZERO 제로 다운로드 아키텍처 시연
- [ ] 영상 공개상태 분류 (공개✅/비공개🔒) + 갤러리 Fallback 시연
- [ ] 영상 속 인물 감지 → 선택 UI (Google Photos 스타일) 시연
- [ ] 분석 중 실시간 하이라이트/진행률 피드백 UX 시연
- [ ] Context Caching 활용 비용 최적화 언급
- [ ] 비동기 Tool 패턴 — 이미지 생성 중 음성 대화 끊김 없음 확인 (Dummy FunctionResponse 즉시 반환)
- [ ] 비동기 Tool 에러 핸들링 — 이미지 실패 시 에이전트가 자연스럽게 안내 (tool_error → Client Content 주입)
- [ ] Exponential Backoff + Jitter — 429 에러 없이 Sequential Agent 완주 확인
- [ ] AudioContext.resume() — 온보딩 첫 터치에서 오디오 활성화 확인 (모바일 필수)

---

---

## 15. Ephemeral Token 보안 및 연결 관리

### 15.1 Ephemeral Token — 클라이언트 보안 인증

Live API WebSocket 연결 시 **API 키를 브라우저에 노출하지 않기 위해** Ephemeral Token을 사용한다. 공식 문서(ai.google.dev/gemini-api/docs/ephemeral-tokens)에 명시된 필수 보안 패턴이다.

```
[브라우저]                    [Go Backend]                [Gemini API]
    │                              │                           │
    │ POST /api/token              │                           │
    │──────────────────────────→   │                           │
    │                              │  POST /ephemeral-token    │
    │                              │  (API Key 포함)           │
    │                              │──────────────────────────→│
    │                              │     { token: "eph_xxx" }  │
    │                              │←──────────────────────────│
    │   { token: "eph_xxx" }       │                           │
    │←──────────────────────────   │                           │
    │                              │                           │
    │ WebSocket Connect (eph_xxx)                              │
    │─────────────────────────────────────────────────────────→│
    │ ← Live API 양방향 스트림 →                                │
```

```go
// internal/handler/token.go — Ephemeral Token 발급 엔드포인트
package handler

import (
    "context"
    "encoding/json"
    "net/http"

    "google.golang.org/genai"
)

func (h *Handler) HandleTokenRequest(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()

    client, err := genai.NewClient(ctx, &genai.ClientConfig{
        APIKey:  h.apiKey, // 서버에서만 보관
        Backend: genai.BackendGeminiAPI,
    })
    if err != nil {
        http.Error(w, "client error", http.StatusInternalServerError)
        return
    }

    // Ephemeral Token 발급 — 유효기간 제한된 단기 토큰
    token, err := client.Live.CreateEphemeralToken(ctx,
        "gemini-live-2.5-flash-native-audio",
        h.liveConnectConfig, // 미리 설정된 LiveConnectConfig 재사용
    )
    if err != nil {
        http.Error(w, "token error", http.StatusInternalServerError)
        return
    }

    json.NewEncoder(w).Encode(map[string]string{
        "token": token.Value,
    })
}
```

```typescript
// web/lib/ephemeral-token.ts — 프론트엔드 토큰 획득 및 WebSocket 연결
async function getEphemeralToken(): Promise<string> {
  const resp = await fetch('/api/token', { method: 'POST' });
  const data = await resp.json();
  return data.token;
}

async function connectLiveAPI() {
  const token = await getEphemeralToken();
  // Ephemeral Token으로 WebSocket 연결 — API 키 미노출
  const ws = new WebSocket(
    `wss://generativelanguage.googleapis.com/ws/google.ai.generativelanguage.v1beta.GenerativeService.BidiGenerateContent?key=${token}`
  );
  return ws;
}
```

### 15.2 GoAway 시그널 및 자동 재접속

Live API WebSocket은 **~10분**마다 서버가 GoAway 시그널을 보내고 연결을 종료한다. Session Resumption 토큰(유효 2시간)을 활용하여 사용자 경험에 끊김 없이 자동 재접속한다.

```go
// internal/live/reconnect.go — GoAway 감지 및 자동 재접속
package live

import (
    "context"
    "log/slog"
    "time"

    "google.golang.org/genai"
)

type SessionManager struct {
    client          *genai.Client
    config          *genai.LiveConnectConfig
    model           string
    resumptionToken string
    session         *genai.Session
}

// handleGoAway는 GoAway 시그널 수신 시 호출된다.
// Session Resumption 토큰을 사용하여 새 연결을 수립한다.
func (sm *SessionManager) handleGoAway(ctx context.Context) error {
    slog.Info("goaway_received", "resumption_token_exists", sm.resumptionToken != "")

    if sm.resumptionToken == "" {
        return fmt.Errorf("no resumption token available")
    }

    // Session Resumption으로 재접속 — 기존 컨텍스트 유지
    sm.config.SessionResumption = &genai.SessionResumptionConfig{
        Handle: sm.resumptionToken,
    }

    var err error
    for attempt := 0; attempt < 3; attempt++ {
        sm.session, err = sm.client.Live.Connect(ctx, sm.model, sm.config)
        if err == nil {
            slog.Info("session_resumed", "attempt", attempt+1)
            return nil
        }
        time.Sleep(time.Duration(attempt+1) * time.Second) // 지수 백오프
    }
    return fmt.Errorf("reconnection failed after 3 attempts: %w", err)
}

// updateResumptionToken은 서버 응답에서 새 토큰을 추출하여 갱신한다.
func (sm *SessionManager) updateResumptionToken(token string) {
    sm.resumptionToken = token
    slog.Debug("resumption_token_updated")
}
```

### 15.3 audioStreamEnd — 오디오 캐시 플러시

사용자 마이크가 **1초 이상 무음**일 때 `audioStreamEnd` 이벤트를 전송한다. 미전송 시 서버가 캐시된 오디오 데이터를 플러시하지 않아 응답 지연이 발생한다.

```typescript
// web/lib/audio-worklet.ts — audioStreamEnd 전송 로직
class AudioStreamManager {
  private ws: WebSocket;
  private silenceTimer: NodeJS.Timeout | null = null;
  private readonly SILENCE_THRESHOLD_MS = 1000;

  onAudioData(pcmData: ArrayBuffer) {
    // 오디오 데이터 전송
    this.ws.send(pcmData);

    // 무음 타이머 리셋
    if (this.silenceTimer) clearTimeout(this.silenceTimer);
    this.silenceTimer = setTimeout(() => {
      // 1초 무음 → audioStreamEnd 전송
      this.ws.send(JSON.stringify({ audioStreamEnd: true }));
    }, this.SILENCE_THRESHOLD_MS);
  }
}
```

### 15.4 비동기 Tool 에러 핸들링 — Client Content 주입

비동기 이미지 생성이 실패하면, 프론트엔드가 Live API 세션에 **Client Content**를 주입하여 에이전트가 자연스럽게 에러를 안내한다. Live API의 `client_content` 메시지 타입을 활용한다.

```typescript
// web/lib/error-handler.ts — 비동기 Tool 에러 → 에이전트 안내
class ToolErrorHandler {
  private liveWs: WebSocket;    // Live API 직접 연결 (Ephemeral Token)
  private backendWs: WebSocket; // Go 백엔드 WebSocket

  constructor(liveWs: WebSocket, backendWs: WebSocket) {
    this.liveWs = liveWs;
    this.backendWs = backendWs;

    // 백엔드 WebSocket에서 tool_error 이벤트 수신
    this.backendWs.addEventListener('message', (event) => {
      const data = JSON.parse(event.data);
      if (data.type === 'tool_error') {
        this.injectErrorToAgent(data.tool, data.message);
      }
    });
  }

  // Live API 세션에 Client Content 주입 → 에이전트가 에러를 인지하고 자연스럽게 안내
  private injectErrorToAgent(toolName: string, message: string) {
    this.liveWs.send(JSON.stringify({
      client_content: {
        turns: [{
          role: 'user',
          parts: [{ text: `[시스템: ${toolName} 도구 실행 중 오류 발생 - ${message}. 사용자에게 자연스럽게 안내해주세요.]` }]
        }],
        turn_complete: true
      }
    }));
  }
}
```

### 15.5 AudioContext 활성화 (브라우저 보안 정책 대응)

모바일 브라우저(iOS Safari, Android Chrome)는 **사용자 상호작용 없이 AudioContext 시작을 차단**한다. 온보딩 첫 터치 이벤트에서 반드시 활성화해야 한다.

```typescript
// web/components/StartButton.tsx — 첫 터치에서 AudioContext + 마이크 활성화
async function handleStartTouch() {
  // 1. AudioContext 활성화 (브라우저 정책 대응)
  const audioCtx = new AudioContext();
  await audioCtx.resume(); // 반드시 사용자 이벤트 핸들러 내에서 호출

  // 2. 마이크 권한 요청
  const stream = await navigator.mediaDevices.getUserMedia({ audio: true });

  // 3. Ephemeral Token 획득 + Live API WebSocket 연결
  const token = await getEphemeralToken();
  const liveWs = connectLiveAPI(token);

  // 4. 백엔드 WebSocket 연결
  const backendWs = new WebSocket(`wss://${location.host}/ws`);

  // 5. 에러 핸들러 초기화
  new ToolErrorHandler(liveWs, backendWs);
}
```

---

## 16. 공식 문서 출처 및 기술 제한사항

### 16.1 핵심 기술 공식 문서 (개발 시 반드시 참조)

| 기술 | 공식 문서 URL | 비고 |
|------|-------------|------|
| **Gemini Live API 세션 관리** | https://ai.google.dev/gemini-api/docs/live-session | 세션 지속시간, 압축, 재개 |
| **Live API 도구 사용** | https://ai.google.dev/gemini-api/docs/live-tools | Tool/Function Calling in Live |
| **Live API 기능 가이드** | https://ai.google.dev/gemini-api/docs/live-guide | Affective Dialog, Proactive Audio |
| **오디오/비디오 스트림 전송** | https://cloud.google.com/vertex-ai/generative-ai/docs/live-api/send-audio-video-streams | PCM 스펙 상세 |
| **Gemini 3 개발자 가이드** | https://ai.google.dev/gemini-api/docs/gemini-3 | 3 Pro Image, 3 Flash, 3.1 Pro |
| **Gemini 모델 목록** | https://ai.google.dev/gemini-api/docs/models | 전체 모델 ID, 토큰 한도 |
| **Rate Limits** | https://ai.google.dev/gemini-api/docs/rate-limits | RPM, TPM, RPD 한도 |
| **GenAI Go SDK** | https://pkg.go.dev/google.golang.org/genai | API 레퍼런스 |
| **GenAI Go SDK 릴리스** | https://github.com/googleapis/go-genai/releases | 버전별 변경사항 |
| **ADK Go 시작 가이드** | https://google.github.io/adk-docs/get-started/go/ | ADK Go 설치/설정 |
| **ADK Go 릴리스** | https://github.com/google/adk-go/releases | v0.4.0 변경사항 |
| **ADK 콜백** | https://google.github.io/adk-docs/callbacks/ | Before/After ModelCallback |
| **ADK 커스텀 도구** | https://google.github.io/adk-docs/tools-custom/ | Tool 인터페이스 정의 |
| **ADK Sequential Agent** | https://google.github.io/adk-docs/agents/workflow-agents/sequential-agents/ | 순차 워크플로우 |
| **ADK 멀티 에이전트** | https://google.github.io/adk-docs/agents/multi-agents/ | 서브 에이전트 패턴 |
| **ADK Go 샘플** | https://github.com/google/adk-samples/tree/main/go/agents | 4개 레퍼런스 에이전트 |
| **Gemini 3 Pro Image (DeepMind)** | https://deepmind.google/models/gemini-image/pro/ | 캐릭터 일관성, 텍스트 렌더링 |
| **Gemini Native Audio 블로그** | https://blog.google/innovation-and-ai/models-and-research/google-deepmind/gemini-2-5-native-audio/ | HD 음성, Affective Dialog 상세 |
| **Ephemeral Tokens** | https://ai.google.dev/gemini-api/docs/ephemeral-tokens | 클라이언트 보안 인증 토큰 |
| **Gemini Video Understanding** | https://ai.google.dev/gemini-api/docs/video-understanding | 영상 분석, 타임스탬프 쿼리, 토큰 소비 |
| **Gemini File API** | https://ai.google.dev/gemini-api/docs/files | 파일 업로드 5GB, 48시간 자동 삭제 |
| **Gemini Context Caching** | https://ai.google.dev/gemini-api/docs/caching | 반복 분석 비용 90% 절감 |
| **Google OAuth 2.0 (Web)** | https://developers.google.com/youtube/v3/guides/auth/server-side-web-apps | Go 서버 OAuth 흐름 |
| **YouTube Data API v3 PlaylistItems** | https://developers.google.com/youtube/v3/docs/playlistItems/list | 채널 영상 목록 조회 |
| **YouTube Data API v3 Videos** | https://developers.google.com/youtube/v3/docs/videos | 영상 메타데이터/썸네일 |
| **YouTube API Quota 계산기** | https://developers.google.com/youtube/v3/determine_quota_cost | API 호출별 quota unit |
| **YouTube API ToS** | https://developers.google.com/youtube/terms/api-services-terms-of-service | API 사용 약관 |
| **YouTube Developer Policies** | https://developers.google.com/youtube/terms/developer-policies | 개발자 정책 (다운로드 금지 등) |
| **GenAI SDK 마이그레이션** | https://ai.google.dev/gemini-api/docs/migrate | 레거시 SDK → genai 전환 |

### 16.2 검증된 기술 제한사항

개발 중 반드시 인지하고 대응해야 하는 제한사항들이다.

| 항목 | 제한 | 대응 방안 | 출처 |
|------|------|----------|------|
| **세션 지속시간 (오디오 전용)** | 압축 없이 **~15분**. 토큰 누적으로 컨텍스트 초과. | `contextWindowCompression` 필수 적용. targetTokenCount: 8192, triggerTokenCount: 12288. | ai.google.dev/gemini-api/docs/live-session |
| **세션 지속시간 (오디오+비디오)** | 압축 없이 **~2분**. 카메라 프레임이 토큰을 빠르게 소비. | 카메라 프레임 전송 주기를 5~10초로 제한. 상시 전송 대신 중요 순간에만 analyze_user Tool로 캡처. | ai.google.dev/gemini-api/docs/live-session |
| **WebSocket 연결 수명** | **~10분**마다 서버가 GoAway 시그널 전송 후 연결 종료. | Session Resumption 토큰 유지 + GoAway 감지 시 자동 재접속 로직 구현. 사용자 경험에 끊김 없어야 함. | ai.google.dev/gemini-api/docs/live-session |
| **Session Resumption 유효시간** | 토큰 발급 후 **2시간**. (이전 계획의 24시간은 오류) | 2시간 내 재접속 설계. 초과 시 새 세션 생성 + Firestore에서 컨텍스트 복원. | ai.google.dev/gemini-api/docs/live-session |
| **Affective Dialog** | **실험적 기능**. 예상치 못한 결과가 발생할 수 있음. | 중요 시연에서는 System Instruction으로 톤 보정. 데모 영상 촬영 시 여러 번 리허설. | ai.google.dev/gemini-api/docs/live-guide |
| **Proactive Audio** | **실험적 기능**. VAD를 넘는 지능적 응답 판단. | 과묵하거나 과다 응답할 수 있으므로 System Instruction에서 응답 빈도 가이드 제공. | blog.google (Gemini 2.5 native audio) |
| **캐릭터 일관성** | gemini-3-pro-image: 최대 5인, **일관성 보장 안 됨** (개선 중). | 동일 프롬프트에 캐릭터 묘사 상세히 포함. 여러 번 생성 후 최적 선택. 레퍼런스 이미지 활용 (최대 14장). | deepmind.google/models/gemini-image/pro/ |
| **음성 언어 선택** | Live API 음성은 기본 **자동 언어 감지**. `SpeechConfig.LanguageCode`로 BCP-47 코드 설정 가능하나, 100% 강제인지 힌트인지 불명확 (커뮤니티 보고: 때때로 언어 전환 발생). | `LanguageCode: "ko-KR"` 설정 + System Instruction "항상 한국어로 대화" **병행** 적용. 이중 안전장치. | ai.google.dev/gemini-api/docs/live-guide, GitHub #381 |
| **Rate Limits** | Free Tier: 5~15 RPM. Tier 1: 150~300 RPM. **프로젝트 단위** 적용. | 이미지 생성 호출 간격 조절 (최소 2초). 429 에러 시 지수 백오프. 데모 시 Tier 1 이상 사용. | ai.google.dev/gemini-api/docs/rate-limits |
| **오디오 포맷** | 입력: PCM 16bit/16kHz/mono. 출력: PCM 16bit/24kHz. **little-endian**. | AudioWorklet에서 정확한 포맷 처리. 샘플레이트 불일치 시 리샘플링 구현. | cloud.google.com/vertex-ai (audio streams) |
| **thinking_level** | `thinking_budget`은 **deprecated**. `thinking_level`(minimal/low/medium/high) 사용. 둘을 동시에 사용하면 오류. | 3 Flash에서 thinking_level만 사용. thinking_budget 코드 제거. | ai.google.dev/gemini-api/docs/gemini-3 |
| **gemini-3-pro-image 토큰** | 입력: 65,536 토큰. 출력: 32,768 토큰. 레퍼런스 이미지: 최대 14장. | 장면 프롬프트를 간결하게 유지. 레퍼런스 이미지 5장 이내 권장. | cloud.google.com/vertex-ai (3 Pro Image) |
| **레거시 SDK 퇴역** | `cloud.google.com/go/vertexai/genai`는 **2026-06-24 퇴역**. `github.com/google/generative-ai-go`도 레거시. | `google.golang.org/genai`만 사용. 모든 import에서 레거시 패키지 제거. | ai.google.dev/gemini-api/docs/migrate |
| **genai Go SDK 버전** | v1.29.0 확정. 이후 버전은 개발 시점에 확인 필요. | `go get google.golang.org/genai@latest`로 최신 설치. 릴리스 노트 확인 후 Breaking Change 검토. | github.com/googleapis/go-genai/releases |
| **Ephemeral Token** | 단기 토큰 발급 API 필수. API 키를 클라이언트에 절대 노출하지 않는다. | Go 백엔드에서 `/token` 엔드포인트 구현. 토큰 유효기간 내 WebSocket 연결 수립. 만료 전 갱신 로직. | ai.google.dev/gemini-api/docs/ephemeral-tokens |
| **audioStreamEnd** | 마이크 무음 시 반드시 전송. 미전송 시 서버가 오디오 캐시를 플러시하지 않아 응답 지연 발생. | 프론트엔드 AudioWorklet에서 1초 무음 감지 → audioStreamEnd 메시지 전송 로직 구현. | ai.google.dev/gemini-api/docs/live |
| **File API 파일 크기** | 최대 **5GB**/파일. 프로젝트당 총 **50GB** 저장 한도. | 긴 영상은 분할 업로드 검토. 분석 완료 후 즉시 삭제하여 50GB 한도 관리. | ai.google.dev/gemini-api/docs/files |
| **File API 자동 삭제** | 업로드 후 **48시간** 자동 삭제. 파일 접근 시 타이머 리셋. 다운로드 불가. | 48시간 내 분석 완료 필수. 원본 영상은 Cloud Storage 임시 버킷에 별도 백업 후 분석 완료 시 삭제. | ai.google.dev/gemini-api/docs/files |
| **Video Understanding 토큰** | 영상 **~258 토큰/초** (초당 1프레임 샘플링). 10분 영상 ≈ **180,000 토큰**. | 저해상도 설정으로 토큰 절감. Context Caching 활용 시 반복 분석 비용 90% 절감. | ai.google.dev/gemini-api/docs/video-understanding |
| **Video Understanding 길이** | 표준 해상도 최대 **1시간**, 저해상도 최대 **3시간**. 요청당 최대 **10개 영상** (Gemini 2.5+). | 사용자 영상을 3~5개 이내로 제한 권장. 영상별 핵심 구간만 분석하는 최적화 고려. | ai.google.dev/gemini-api/docs/video-understanding |
| **YouTube Data API Quota** | 하루 **10,000 유닛** 기본. 읽기: 1 유닛, 검색: 100 유닛. | 사용자당 영상 목록 조회 2~3 유닛. 해커톤 규모에서 충분. 초과 시 Google Cloud Console에서 상향 신청. | developers.google.com/youtube/v3/determine_quota_cost |
| **YouTube 영상 다운로드** | YouTube Data API로 영상 바이너리 **다운로드 불가**. 서드파티(yt-dlp 등)는 **ToS 위반**. | **Gemini API YouTube URL 직접 분석으로 해결!** `genai.FileData{FileURI: youtubeURL}` 전달. 다운로드 불필요. 공개/일부공개만 지원. 비공개는 갤러리 업로드 fallback. | ai.google.dev/gemini-api/docs/video-understanding |
| **YouTube URL 분석 제한** | Gemini API YouTube URL 직접 분석은 **공개(Public)/일부공개(Unlisted) 영상만** 지원. 비공개(Private) 영상 미지원. 무료 티어: 하루 최대 8시간 분량. 요청당 YouTube URL 1개. | 공개/일부공개 영상 우선 표시. 비공개 영상은 "일부공개 변경" 안내 또는 갤러리 업로드 fallback. Gemini 2.5+: 요청당 최대 10개 영상 (File API 경유). | ai.google.dev/gemini-api/docs/video-understanding |
| **인물 얼굴 클러스터링** | Gemini Vision으로 얼굴 감지 가능하나 **자동 클러스터링은 미지원**. 얼굴 인식(특정인 식별)도 미지원. | Gemini로 바운딩 박스 추출 → 크롭 이미지 생성 → 사용자가 직접 대상 인물 선택. Google Photos 수준의 자동화는 불가. | ai.google.dev/gemini-api/docs/video-understanding |
| **데이터 저장 정책** | 원본 영상은 **임시 저장만** (분석 완료 후 삭제). 페르소나 프로파일과 생성 이미지만 영구 보관. | Cloud Storage lifecycle policy로 임시 버킷 자동 삭제 설정. Firestore에 페르소나 데이터만 영구 저장. | — (서비스 정책) |
| **비동기 Tool 에러 핸들링** | 비동기 패턴에서 이미지 생성 실패 시 Live API(음성 에이전트)가 실패를 인지 못함. 사용자가 이미지를 무한 대기. | 백엔드→프론트 WebSocket `tool_error` 이벤트 전송 → 프론트가 Live API에 Client Content 주입 → 에이전트가 "이미지 생성에 문제가 생겼네요"라고 자연스럽게 안내. | — (비동기 에러 파이프라인) |
| **QPS(초당 쿼리 수) 병목** | Sequential Agent가 3개 모델을 연달아 호출 시 QPS 제한에 걸릴 확률 높음. Tier 1에서도 초당 제한 존재. | Exponential Backoff + Jitter 로직 적용 (`retryWithBackoff` 함수). 429 에러 시 1s→2s→4s 지수 백오프. 데모 시연 중 Rate Limit 방어 필수. | ai.google.dev/gemini-api/docs/rate-limits |
| **모바일 AudioContext 정책** | iOS Safari, Android Chrome 모두 **사용자 상호작용(터치/클릭) 없이 AudioContext 시작 불가**. 브라우저 보안 정책. | 온보딩 Step 1.5에서 "시작하기" 버튼 터치 이벤트로 `AudioContext.resume()` 강제 실행. 마이크 권한 요청도 동시 수행. | Web Audio API 브라우저 정책 |

### 16.3 추적이 필요한 빠른 변화 항목

개발 기간(2/25~3/16) 동안 아래 항목들은 **매주 1회** 공식 문서를 재확인한다.

| 추적 항목 | 확인 URL | 확인 주기 | 확인 내용 |
|----------|---------|----------|----------|
| Gemini 모델 퇴역/신규 | https://ai.google.dev/gemini-api/docs/models | 주 1회 | 사용 중인 5개 모델의 퇴역 일정 변경, 신규 모델 출시 |
| genai Go SDK 릴리스 | https://github.com/googleapis/go-genai/releases | 주 1회 | 신규 버전, Breaking Changes, Live API 관련 변경 |
| ADK Go 릴리스 | https://github.com/google/adk-go/releases | 주 1회 | v0.4.0 이후 패치, 새 에이전트 타입 |
| Live API 변경사항 | https://ai.google.dev/gemini-api/docs/live-session | 주 1회 | 세션 제한 변경, 새 기능 추가, 모델 지원 변경 |
| Rate Limits 변경 | https://ai.google.dev/gemini-api/docs/rate-limits | 주 1회 | 티어별 한도 변경, 새 모델 한도 추가 |
| Gemini 3 Pro Image 업데이트 | https://ai.google.dev/gemini-api/docs/gemini-3 | 주 1회 | 이미지 품질 개선, 새 파라미터, 제한 변경 |
| YouTube Data API 변경 | https://developers.google.com/youtube/v3/revision_history | 주 1회 | API 변경, quota 정책, 새 기능 |
| File API / Video Understanding | https://ai.google.dev/gemini-api/docs/video-understanding | 주 1회 | 토큰 소비 변경, 지원 포맷, 최대 길이 변경 |
| 챌린지 공지사항 | https://devpost.com/software/built-with/google-gemini | 주 1회 | 규칙 변경, 마감일 변경, 추가 요구사항 |

---

*missless.co — 텍스트 박스를 넘어, 그리움의 자리를 채우다.*
