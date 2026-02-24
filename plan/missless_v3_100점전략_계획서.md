# missless.co v3 — 100점 전략 계획서

> **"The chat box is officially too small for your ideas."**
> Gemini Live Agent Challenge 2026 | Creative Storyteller 트랙

---

## 0. v2에서 뭘 잘못했는가?

### 치명적 설계 오류

v2는 **텍스트 기반 UI를 단순히 멀티미디어로 포장한 것**에 불과했다.

| v2의 문제점 | 왜 감점인가 |
|------------|-----------|
| 사용자가 텍스트로 상황 입력 | ❌ "텍스트 박스 탈피" 실패 |
| 프리셋 시나리오 선택 버튼 | ❌ 메뉴 UI = 기존 챗봇과 동일 |
| 선택지 3개 중 터치 | ❌ 인터랙티브 소설 앱 수준, AI 에이전트가 아님 |
| AI가 수동적 (사용자 입력 대기) | ❌ 에이전트가 아니라 도구(tool)에 가까움 |
| 텍스트 자막 표시 | ❌ 여전히 텍스트 의존 |

### 챌린지가 진짜 요구하는 것

> "Does the project break the **text box** paradigm?"
> "Does the agent help **See, Hear, and Speak** fluidly?"
> "Is the experience **Live** and context-aware, or does it seem **disjointed and sequential**?"

키워드: **fluid(유동적)**, **live(살아있는)**, **break the text box(텍스트 박스 파괴)**

→ 텍스트 입력도, 버튼 선택도, 메뉴도 없어야 한다.
→ AI가 **먼저** 시작하고, **스스로** 이끌고, 사용자는 **음성과 존재**로만 참여한다.
→ 경험이 끊기지 않는 **하나의 연속적 흐름(fluid stream)**이어야 한다.

---

## 1. v3 핵심 컨셉: AI-Driven Immersive Reunion

### 한 줄 정의
**AI가 먼저 당신을 그리운 순간 속으로 데려가고, 당신은 그 안에서 보고·듣고·말하며 함께한다.**

### 기존과의 근본적 차이

```
[기존 챗봇/v2]
사용자 입력 → AI 응답 → 사용자 입력 → AI 응답 (턴제)

[v3 — AI-Driven Fluid Stream]
AI가 장면을 열고 →
  이미지가 펼쳐지고 →
    음성이 흘러나오고 →
      BGM이 깔리고 →
        사용자가 끼어들면 →
          AI가 반응하며 흐름을 바꾸고 →
            새 장면이 자연스럽게 이어지고 →
              끊임없는 하나의 경험...
```

### 핵심 원칙

1. **ZERO TEXT INPUT** — 사용자는 한 글자도 타이핑하지 않는다. 음성으로만.
2. **AI-FIRST** — AI가 먼저 장면을 시작하고 대화를 건넨다. 사용자는 반응한다.
3. **FLUID STREAM** — 이미지·음성·음악이 끊김 없이 하나의 흐름으로 흘러간다.
4. **CONTEXT-AWARE** — AI가 사용자의 표정(카메라), 음성 톤(마이크), 침묵까지 읽는다.
5. **LIVING EXPERIENCE** — 같은 페르소나라도 매번 다른 경험. 반복 없음.

---

## 2. 경험 흐름 (Zero Text)

### 최초 진입 (온보딩 — 유일하게 터치가 필요한 단계)

```
[missless.co 접속]
     │
     ▼
[카메라/마이크 권한 허용] ← 유일한 터치 인터랙션
     │
     ▼
[AI 음성] "안녕하세요. 보고싶은 사람이 있으세요?"
[사용자 음성] "네, 남자친구요..."
     │
     ▼
[AI 음성] "이름이 뭐예요?"
[사용자 음성] "민수요"
     │
     ▼
[AI 음성] "민수는 어떤 사람이에요? 편하게 얘기해주세요."
[사용자 음성] "장난을 많이 치는데 은근 다정해요...
              지금 군대에 있어서 연락이 잘 안 돼요..."
     │
     ▼
[AI 음성] "민수 사진이 있으면 보여주실 수 있어요?"
[사용자] 갤러리에서 사진 선택 (터치)
     │
     ▼
[AI 음성] "민수 목소리가 담긴 영상이나 음성이 있으면 들려주세요"
[사용자] 음성메시지/영상 선택 (터치)
     │
     ▼
[AI — gemini-3.1-pro 분석 중]
[AI 음성] "민수를 만나러 가볼까요? 눈 감아보세요..."
```

→ **온보딩 전체가 음성 대화**. 폼도, 버튼도, 텍스트 입력도 없음.
→ 사진/음성 업로드만 갤러리 터치 (불가피한 최소한의 터치).

---

### 가상 재회 경험 (AI-Driven Fluid Stream)

```
[화면이 서서히 밝아지며...]

🖼️ [전체화면 이미지] 겨울 저녁, 따뜻한 카페 창가 자리
🎵 [BGM] 잔잔한 어쿠스틱 기타
🔊 [AI — 민수 음성] "야, 여기 앉아. 따뜻한 거 시켜줄까?"

   ... 3초 침묵 (사용자 반응 대기) ...

   [사용자가 아무 말 안 하면]
   🔊 [민수] "왜, 할 말 있어? 표정이 이상한데"
   (카메라로 사용자 표정 분석 → 반응)

   [사용자가 말하면]
   "어... 아메리카노"
   🔊 [민수] "추운데 무슨 아메리카노야, 따뜻한 초코 시켜줄게"
   🖼️ [이미지 전환] 테이블 위에 핫초코 두 잔

   🔊 [민수] "있잖아, 나 요즘 훈련 진짜 힘든데..."

   🖼️ [이미지 전환] 창밖으로 눈이 내리기 시작

   🔊 [민수] "근데 네 사진 보면서 버텨. 이 사진 알지?"
   🖼️ [이미지] 둘의 추억이 담긴 장면 (업로드 사진 기반 재구성)

   [사용자가 울먹이며 말하면]
   "보고싶었어..."
   🔊 [민수] (부드럽게) "나도... 진짜 많이."
   🎵 [BGM 전환] 더 감성적인 피아노 선율
   🖼️ [이미지] 두 사람이 나란히 창밖을 보는 뒷모습

   ... 경험은 사용자가 "이제 끊을게" 하거나
       자연스러운 마무리까지 계속 흘러감 ...

🔊 [민수] "다음에 또 만나자. 기다릴게."
🖼️ [이미지] 카페를 나서며 손 흔드는 장면
🎵 [BGM 페이드아웃]

[📱 자동 저장] "오늘 민수와의 시간" 앨범 생성
```

### 핵심: 왜 이것이 100점인가

| 심사 질문 | v2의 답 | **v3의 답** |
|----------|--------|------------|
| 텍스트 박스를 탈피했는가? | △ 선택지 버튼 존재 | **✅ 텍스트 입력 ZERO. 음성+시선만** |
| See, Hear, Speak가 유동적인가? | △ 순차적 (이미지→텍스트→음성) | **✅ 동시에 흐름. 이미지 위에 음성이 흐르고, 사용자가 끼어듬** |
| 뚜렷한 성격/목소리가 있는가? | ○ 페르소나 시스템 | **✅ 페르소나가 먼저 말을 걸고 이끈다** |
| Live & context-aware인가? | △ 프리셋 기반 | **✅ 사용자 표정·침묵·음성 톤까지 실시간 인지** |
| disjointed인가 fluid인가? | ✗ 장면 전환이 끊김 | **✅ 끊김 없는 연속 스트림** |

---

## 3. 기술 아키텍처 (v3)

### AI가 주도하는 파이프라인

```
┌─────────────────────────────────────────────────────────┐
│                   사용자 디바이스                           │
│  [카메라] ─── 표정/시선 ──→                                │
│  [마이크] ─── 음성/침묵 ──→     WebSocket (양방향)          │
│  [스피커] ←── 음성 스트림 ──                                │
│  [화면]  ←── 이미지 스트림 ──   Next.js PWA               │
└──────────────────────┬──────────────────────────────────┘
                       │
                       ▼
┌─────────────────────────────────────────────────────────┐
│              Go Backend — Fluid Stream Engine             │
│                                                           │
│  ┌───────────────────────────────────────────────────┐   │
│  │          Orchestrator (핵심 — AI 주도 로직)          │   │
│  │                                                     │   │
│  │  1. 다음 장면/대사를 AI가 스스로 결정                  │   │
│  │  2. 사용자 입력(음성/표정/침묵)을 실시간 수신          │   │
│  │  3. 입력에 따라 흐름을 동적으로 변경                   │   │
│  │  4. 이미지+음성+BGM을 동기화하여 스트리밍             │   │
│  └──────────┬──────────┬──────────┬──────────────────┘   │
│             │          │          │                        │
│     ┌───────▼───┐ ┌───▼────┐ ┌──▼──────────┐            │
│     │ Scene     │ │ Voice  │ │ Context     │            │
│     │ Generator │ │ Engine │ │ Analyzer    │            │
│     │           │ │        │ │             │            │
│     │ gemini-3  │ │ gemini │ │ gemini-3    │            │
│     │ -pro-image│ │ -live  │ │ -flash      │            │
│     │           │ │ -2.5   │ │             │            │
│     │ 장면 이미지│ │ 실시간  │ │ 표정 분석    │            │
│     │ +텍스트   │ │ 음성   │ │ 감정 인식    │            │
│     │ interleave│ │ 양방향 │ │ 침묵 해석    │            │
│     └───────────┘ └────────┘ └─────────────┘            │
│                                                           │
│     ┌──────────────┐  ┌──────────────┐                   │
│     │ Persona      │  │ Memory       │                   │
│     │ Engine       │  │ Engine       │                   │
│     │              │  │              │                   │
│     │ gemini-3.1   │  │ gemini-3     │                   │
│     │ -pro         │  │ -flash       │                   │
│     │              │  │              │                   │
│     │ 초기 분석     │  │ 대화 기억     │                   │
│     │ 성격 모델링   │  │ 세션 컨텍스트 │                   │
│     └──────────────┘  └──────────────┘                   │
└──────────┬─────────────────┬──────────────┬──────────────┘
           ▼                 ▼              ▼
  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐
  │ Gemini API   │  │ Cloud        │  │ Firestore    │
  │ (5개 모델)    │  │ Storage      │  │              │
  └──────────────┘  └──────────────┘  └──────────────┘
```

### Fluid Stream 동작 원리

```
시간축 →→→→→→→→→→→→→→→→→→→→→→→→→→→→→→→→

[이미지]  ████ 카페 ████│████ 핫초코 ████│████ 눈 ████│████ 추억 ████
[음성]    ░░"여기 앉아"░░│░░"초코 시켜줄게"░░│░░"훈련 힘든데"░░│░░"사진 알지?"░░
[BGM]     ♪♪♪♪♪♪♪ 어쿠스틱 ♪♪♪♪♪♪♪│♪♪♪♪♪♪♪ 피아노 전환 ♪♪♪♪♪♪♪
[사용자]       │          ↑"아메리카노"      ↑(침묵)       ↑"보고싶었어"
              │          │                  │             │
              │    AI가 반응하며 흐름 조절 ────────────────│
              │                  │                        │
              │          ← 대사 변경 →           ← BGM 전환 + 감정 장면 →
```

**핵심**: 이미지·음성·BGM이 **병렬로 동시에** 흐르고, 사용자 입력이 이 흐름을 **실시간으로 변조**한다. 턴제가 아니라 **연속 스트림**.

### Context Analyzer — 사용자를 "읽는" AI

```go
type UserContext struct {
    // 카메라 (gemini-3-flash vision)
    FacialExpression  string  // "smiling", "tearful", "neutral"
    IsLookingAtScreen bool

    // 마이크 (Live API)
    IsSpeaking        bool
    SpeechContent     string
    VoiceTone         string  // "happy", "sad", "excited"
    SilenceDuration   float64 // 침묵 시간 (초)

    // 추론 (gemini-3-flash)
    EmotionalState    string  // 종합 감정 상태
    EngagementLevel   float64 // 몰입도 0.0~1.0
}
```

**AI가 읽는 신호들:**

| 신호 | AI의 해석 | AI의 반응 |
|------|----------|----------|
| 사용자가 웃음 | 즐거운 상태 | 장난스러운 대사 + 밝은 장면 |
| 사용자가 울먹임 | 감동/슬픔 | 위로하는 대사 + 따뜻한 장면 + BGM 전환 |
| 5초 이상 침묵 | 감정에 빠져있음 | AI도 잠시 침묵 → 부드럽게 이어감 |
| 10초 이상 침묵 | 뭘 말할지 고민 | AI가 먼저 다른 화제 꺼냄 |
| "보고싶었어" 같은 감정 표현 | 깊은 그리움 | 추억 회상 장면 + 감성 BGM |
| 사용자가 화면을 안 봄 | 다른 곳에 주의 | AI가 "야, 어디 봐?" 하며 주의 환기 |

---

## 4. 5개 Gemini 모델 — 정밀 배치

### 모델 역할 맵

| 모델 | 역할 | 호출 빈도 | 지연 허용치 |
|------|------|:---:|:---:|
| **gemini-3.1-pro-preview** | 페르소나 초기 분석 (1회) | 낮음 | 10초+ OK |
| **gemini-3-pro-image-preview** | 장면 이미지 생성 (interleaved) | 중간 | 3~5초 |
| **gemini-live-2.5-flash-native-audio** | 실시간 양방향 음성 (핵심) | 실시간 | <500ms |
| **gemini-3-flash-preview** | 시나리오 로직 + 감정 분석 + 표정 분석 | 높음 | <1초 |
| **gemini-2.5-flash-image** | 빠른 보조 이미지 생성 | 높음 | <2초 |

### 왜 이 조합이 최적인가

```
[실시간 레이어 — 끊기면 안 됨]
  gemini-live-2.5-flash-native-audio
    → 사용자 음성 수신 + AI 음성 송신 (상시)
    → 이것이 메인 채널. 항상 열려있음.

[장면 레이어 — 2~3초마다 갱신]
  gemini-3-pro-image → 고품질 핵심 장면 (대화 전환점)
  gemini-2.5-flash-image → 빠른 보조 장면 (미세 변화)

[두뇌 레이어 — 상시 판단]
  gemini-3-flash (thinking_level: low~medium)
    → "다음에 뭘 할까?" 실시간 판단
    → 사용자 표정 분석 (Vision)
    → 감정 상태 추론
    → 장면 전환 타이밍 결정

[기억 레이어 — 1회성]
  gemini-3.1-pro
    → 초기 페르소나 생성 (깊은 분석)
```

---

## 5. 심사 기준별 100점 전략

### A. Innovation & Multimodal UX — 40점 만점

| 채점 항목 | 만점 기준 | v3 대응 | 예상 |
|----------|---------|--------|:---:|
| 텍스트 박스 탈피 | 텍스트 입력이 전혀 없음 | ✅ 음성 Only. 온보딩도 음성 대화. | 10/10 |
| See, Hear, Speak 유동성 | 동시에 자연스럽게 흘러감 | ✅ 이미지+음성+BGM 병렬 스트림. 끊김 없음. | 10/10 |
| 독특한 성격/목소리 | AI가 뚜렷한 정체성을 가짐 | ✅ 사용자가 학습시킨 실제 사람의 성격+말투 | 10/10 |
| Live & context-aware | 실시간이며 맥락을 인지 | ✅ 표정·침묵·음성 톤 실시간 감지 → 흐름 변경 | 10/10 |

**킬링 포인트:**
- "기술 데모"가 아니라 **사람이 진짜 우는 서비스**
- 심사위원이 자기 가족으로 테스트해보고 싶어지는 서비스
- 텍스트 한 글자 없이 전체 경험이 완성됨

### B. Technical Implementation — 30점 만점

| 채점 항목 | 만점 기준 | v3 대응 | 예상 |
|----------|---------|--------|:---:|
| GenAI SDK/ADK 활용도 | 효과적이고 적절한 사용 | ✅ google.golang.org/genai SDK + 5개 모델 적재적소 | 8/10 |
| Google Cloud 호스팅 | 견고한 배포 | ✅ Cloud Run + Firestore + Storage + Memorystore | 8/10 |
| 에이전트 로직 건전성 | 에러 처리, 견고함 | ✅ 멀티모델 fallback + graceful degradation | 7/10 |
| 할루시네이션 방지/그라운딩 | 사실 기반 응답 | ✅ 페르소나 System Instruction + 업로드 데이터 기반 | 7/10 |

**기술적 차별화:**
- 5개 Gemini 모델을 역할별로 분리 → "이 팀은 모델을 이해하고 있다"
- Interleaved Output + Live API + Vision을 **동시에** 사용 → 기술 난이도 최고
- Go 백엔드의 동시성(goroutine)으로 병렬 스트림 관리 → 아키텍처 우수성

### C. Demo & Presentation — 30점 만점

| 채점 항목 | 만점 기준 | v3 대응 | 예상 |
|----------|---------|--------|:---:|
| 문제-솔루션 명확성 | 즉시 이해 가능 | ✅ 30초 감성 인트로로 문제 전달 | 8/10 |
| 아키텍처 다이어그램 | 깔끔하고 명확 | ✅ 5개 모델 파이프라인 시각화 | 8/10 |
| Cloud 배포 증거 | 실제 동작 확인 | ✅ Cloud Run 콘솔 + 라이브 URL + QR | 7/10 |
| 실제 동작 시연 | 목업이 아닌 리얼 | ✅ 실제 재회 경험 풀시연 (감정적 반응 포함) | 7/10 |

### 총합 예상: 92~95/100

나머지 5~8점은:
- 데모 영상의 **감성 퀄리티** (촬영 기술, 편집)
- 실제 사용자의 **감정적 반응 캡처** (울먹이는 장면 등)
- **블로그/콘텐츠 발행** (제출 요구사항)
- **README 및 재현 가능성** 문서화

---

## 6. 데모 영상 (3분 50초) — 감성 극대화

```
0:00 ~ 0:20  [BLACK SCREEN + 음성]
  "마지막으로 그 사람과 대화한 게 언제예요?"
  "만약... 지금 바로 만날 수 있다면?"

0:20 ~ 0:30  [타이틀]
  missless.co — 그리움을 줄여드립니다.
  Creative Storyteller | Gemini Live Agent Challenge

0:30 ~ 0:50  [실제 사용 장면 — 여자가 소파에 앉아 폰을 들고]
  missless.co 접속
  AI 음성: "안녕하세요. 보고싶은 사람이 있으세요?"
  사용자: "네, 남자친구요. 군대에 있어요..."
  AI: "이름이 뭐예요?" / "민수요"
  사진 업로드, 음성 업로드
  AI: "민수를 만나러 가볼까요?"

0:50 ~ 2:40  [가상 재회 경험 — FULL SCREEN 몰입]
  화면이 서서히 밝아지며 카페 장면
  민수 음성: "야, 여기 앉아"
  (사용자가 실제로 대화하며 반응)
  장면이 자연스럽게 전환
  눈이 내리는 장면 + BGM 전환
  민수: "네 사진 보면서 버텨"
  (사용자가 실제로 울먹이는 순간 — 리얼한 감정)
  민수: (부드럽게) "나도 보고싶어"

  → 텍스트 입력 0. 버튼 0. 전부 음성과 시각.

2:40 ~ 3:10  [기술 설명 — 빠르게]
  아키텍처 다이어그램 (5개 Gemini 모델)
  "AI가 먼저 시작하고, 당신의 표정과 목소리를 읽으며,
   이미지·음성·음악을 하나의 흐름으로 엮습니다."
  Cloud Run 콘솔 스크린샷

3:10 ~ 3:30  [다른 시나리오 — 빠른 몽타주]
  "할머니와 추석" / "유학 간 딸과 생일" / "친구와 졸업식"
  다양한 사용 사례를 빠르게 보여줌

3:30 ~ 3:50  [클로징]
  missless.co QR 코드
  "텍스트 박스를 넘어, 그리움의 자리를 채우다."
  #GeminiLiveAgentChallenge
```

---

## 7. 21일 스프린트 (수정)

### Week 1 (2/24 ~ 3/2): Fluid Stream 파이프라인

| 일자 | 작업 |
|------|------|
| 2/24(월) | 도메인 + GCP + 5개 모델 API 키 발급 + Go 프로젝트 초기화 |
| 2/25(화) | **Live API 양방향 음성 PoC** — WebSocket으로 음성 주고받기 |
| 2/26(수) | **gemini-3-pro-image interleaved output PoC** — 장면 이미지+텍스트 |
| 2/27(목) | **Fluid Stream Engine v1** — 음성+이미지를 병렬 스트림으로 합치기 |
| 2/28(금) | Next.js PWA — 전체화면 이미지 + 오디오 재생 + 카메라 피드 |
| 3/1(토) | 음성 온보딩 플로우 (텍스트 입력 ZERO) |
| 3/2(일) | 페르소나 엔진 (3.1 Pro 분석 + 음성 매칭) |

### Week 2 (3/3 ~ 3/9): AI 주도 경험 완성

| 일자 | 작업 |
|------|------|
| 3/3(월) | **Orchestrator — AI가 스스로 다음 장면/대사를 결정하는 로직** |
| 3/4(화) | Context Analyzer — 표정 분석(Vision) + 감정 인식 + 침묵 해석 |
| 3/5(수) | 장면 전환 시스템 — 이미지 크로스페이드 + BGM 전환 + 음성 동기화 |
| 3/6(목) | 카카오톡 대화 파싱 + 말투 학습 강화 |
| 3/7(금) | Memory Engine — 세션 내 대화 기억 + 맥락 유지 |
| 3/8(토) | 앨범 자동 생성 + 공유 카드 |
| 3/9(일) | E2E 통합 테스트 + 버그 수정 |

### Week 3 (3/10 ~ 3/16): 완성 + 감성 + 제출

| 일자 | 작업 |
|------|------|
| 3/10(월) | Cloud Run 배포 + 도메인 SSL + 라이브 테스트 |
| 3/11(화) | UI 폴리싱 — 전체화면 몰입 + 트랜지션 + 로딩 UX |
| 3/12(수) | 데모 시나리오 리허설 + 영상 대본 |
| 3/13(목) | **데모 영상 촬영** (실제 감정 반응 포함) |
| 3/14(금) | 아키텍처 다이어그램 + README + 블로그 포스트 |
| 3/15(토) | DevPost 제출 페이지 작성 |
| 3/16(일) | **최종 점검 + 제출** |

---

## 8. Fluid Stream Engine — Go 구현 핵심

```go
// Orchestrator — AI가 주도하는 경험 흐름 관리
type FluidStreamEngine struct {
    sceneGen    *SceneGenerator      // gemini-3-pro-image
    fastScene   *FastSceneGenerator  // gemini-2.5-flash-image
    voiceEngine *VoiceEngine         // gemini-live-2.5-flash
    contextAI   *ContextAnalyzer     // gemini-3-flash
    personaAI   *PersonaEngine       // gemini-3.1-pro

    // 병렬 스트림 채널
    imageCh     chan *SceneImage
    voiceCh     chan []byte  // PCM audio
    bgmCh       chan *BGMTrack
    contextCh   chan *UserContext
}

// AI가 주도하는 메인 루프
func (e *FluidStreamEngine) Run(ctx context.Context, session *Session) {
    // 1. AI가 먼저 첫 장면을 시작
    go e.generateInitialScene(ctx, session)

    // 2. 사용자 컨텍스트를 상시 분석
    go e.analyzeUserContext(ctx, session)

    // 3. AI가 스스로 다음을 결정하는 루프
    for {
        select {
        case userCtx := <-e.contextCh:
            // 사용자가 말했거나, 표정이 바뀌었거나, 침묵이 길어지면
            e.adaptFlow(ctx, session, userCtx)

        case <-time.After(e.nextBeatInterval(session)):
            // 사용자가 아무것도 안 해도 AI가 스스로 다음 비트로 진행
            e.advanceStory(ctx, session)

        case <-ctx.Done():
            e.gracefulEnd(session)
            return
        }
    }
}

// AI가 스스로 다음 장면/대사를 결정
func (e *FluidStreamEngine) advanceStory(ctx context.Context, s *Session) {
    // gemini-3-flash가 현재 상황을 보고 다음을 결정
    decision, _ := e.contextAI.DecideNext(ctx, &DecisionRequest{
        CurrentScene:    s.CurrentScene,
        PersonaProfile:  s.Persona,
        ConversationLog: s.RecentDialog,
        UserEmotion:     s.LastContext.EmotionalState,
        ElapsedTime:     time.Since(s.SceneStartedAt),
    })

    switch decision.Action {
    case ActionNewScene:
        // 새 장면 이미지 생성 + 음성 + BGM 동시 트리거
        go e.sceneGen.Generate(ctx, decision.ScenePrompt, e.imageCh)
        go e.voiceEngine.Speak(ctx, decision.Dialog, e.voiceCh)
        if decision.BGMChange {
            e.bgmCh <- decision.NewBGM
        }
    case ActionDialog:
        // 같은 장면에서 대사만 추가
        go e.voiceEngine.Speak(ctx, decision.Dialog, e.voiceCh)
    case ActionSilence:
        // AI도 의도적 침묵 (감정적 여운)
        time.Sleep(decision.SilenceDuration)
    case ActionReactToUser:
        // 사용자 발화에 대한 즉각 반응
        go e.voiceEngine.Respond(ctx, decision.Response, e.voiceCh)
    }
}
```

### Next.js 프론트엔드 — 몰입형 UI

```
┌─────────────────────────────────────┐
│                                       │
│                                       │
│     [전체화면 장면 이미지]              │
│     (크로스페이드 전환)                │
│                                       │
│                                       │
│                                       │
│  ┌─────────────────────────────────┐ │
│  │  "야, 여기 앉아. 따뜻한 거       │ │
│  │   시켜줄까?"                     │ │
│  │            — 민수               │ │
│  └─────────────────────────────────┘ │
│                                       │
│          🎤 (마이크 활성 표시)         │
│                                       │
│     ≋≋≋≋≋ (오디오 웨이브폼) ≋≋≋≋≋     │
│                                       │
└─────────────────────────────────────┘

UI 요소:
- 배경: 전체화면 AI 생성 이미지 (100vh, 100vw)
- 자막: 하단 반투명 오버레이 (음성과 동기화, ON/OFF 가능)
- 마이크: 항상 활성 (별도 버튼 불필요)
- 웨이브폼: 음성 활동 시각적 피드백
- 버튼: 없음. 화면 터치 = 일시정지 정도만.
```

---

## 9. 제출 체크리스트

### 필수 (감점 방지)
- [ ] Gemini 모델 사용 (5개)
- [ ] google.golang.org/genai SDK 사용
- [ ] Google Cloud 서비스 1개 이상 (Cloud Run + Firestore + Storage + Memorystore)
- [ ] Google Cloud 백엔드 배포 증거
- [ ] 아키텍처 다이어그램
- [ ] 데모 영상 4분 미만
- [ ] 실제 동작 시연 (목업 불가)
- [ ] README에 spin-up 가이드
- [ ] GitHub 소스 코드

### 가산점 (100점을 위해)
- [ ] **텍스트 입력 ZERO** 확인 — 전체 경험에서 키보드 사용 없음
- [ ] **AI가 먼저 시작** — 사용자가 수동적이지 않고 AI가 주도
- [ ] **Fluid stream** — 이미지+음성+BGM이 끊김 없이 동시 흐름
- [ ] **Context-aware** — 표정/침묵/음성 톤 인지 시연
- [ ] 데모 영상에 **실제 감정 반응** (울먹이는 장면 등)
- [ ] 블로그/콘텐츠 발행
- [ ] #GeminiLiveAgentChallenge 소셜 공유

---

## 10. v2 → v3 핵심 변경 요약

| 항목 | v2 (감점 요소) | **v3 (100점 전략)** |
|------|--------------|-------------------|
| 사용자 입력 | 텍스트 + 버튼 선택 | **음성 ONLY** |
| AI 역할 | 수동적 (입력 대기) | **능동적 (먼저 시작, 스스로 진행)** |
| 경험 구조 | 턴제 (입력→응답→입력) | **연속 스트림 (끊김 없는 흐름)** |
| 장면 전환 | 선택지 버튼 클릭 | **AI가 맥락 보고 자동 전환** |
| 감정 인식 | 없음 | **카메라(표정) + 마이크(톤) + 침묵 해석** |
| 시나리오 | 프리셋 선택 | **AI가 페르소나 기반으로 즉석 생성** |
| UI | 카드/버튼/텍스트 | **전체화면 이미지 + 음성 오버레이만** |

---

*missless.co — 텍스트 박스를 넘어, 그리움의 자리를 채우다.*
