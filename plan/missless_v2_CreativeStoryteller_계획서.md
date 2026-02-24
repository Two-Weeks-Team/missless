# missless.co v2 — Creative Storyteller 트랙 프로젝트 계획서

> **"그리운 사람과, 다시 만나는 순간."**
> Gemini Live Agent Challenge 2026 | **Creative Storyteller 트랙**

---

## 1. 서비스 개요

### 한 줄 요약
그리운 사람과의 **가상 재회 경험**을 AI가 이미지·음성·텍스트·음악을 엮어 인터랙티브 스토리로 만들어주는 서비스.

### Live Agent 트랙과의 차이

| | Live Agent (v1) | **Creative Storyteller (v2)** |
|---|---|---|
| 핵심 경험 | 실시간 음성 통화 | **가상 재회 장면 생성 + 인터랙티브 스토리** |
| 출력 형태 | 음성 대화 | 이미지 + 음성 + 텍스트 + BGM 통합 |
| 기술 핵심 | Bidi-streaming | **Interleaved multimodal output** |
| 감정 밀도 | 대화 중심 | **시각+청각+스토리 복합 감성** |
| 경쟁 강도 | 높음 (예상) | **낮음~중간** |

### 왜 Creative Storyteller인가?
- 대부분의 출전작이 Live Agent(음성 비서/챗봇)에 몰릴 것으로 예상
- Creative Storyteller는 "뭘 만들어야 하지?" 진입장벽이 높아 경쟁 낮음
- missless의 감성 컨셉은 **멀티미디어 스토리텔링**과 완벽히 맞음
- Gemini의 Interleaved Output을 가장 임팩트있게 보여줄 수 있는 유즈케이스

---

## 2. 핵심 시나리오

### 사용자 여정: "군대 간 남친과 크리스마스"

```
[1] 사용자: "민수와 크리스마스 이브에 같이 트리 장식하고 싶어"

[2] AI가 생성하는 가상 재회 경험:

    🖼️ [이미지] 따뜻한 거실, 반짝이는 크리스마스 트리

    🔊 [음성] (민수 말투로)
        "야, 이 별 장식 어디에 달까? 맨 꼭대기?"

    📝 [텍스트] 민수가 장난스럽게 웃으며 별 장식을 들어 보인다.

    🎵 [BGM] 잔잔한 캐럴 배경음악

    💬 [인터랙션]
        → "꼭대기에 달아줘"
        → "아니, 이쪽에 달자"
        → "그냥 네가 알아서 해"

[3] 사용자가 선택하면 → 스토리가 이어짐

    🖼️ [이미지] 민수가 의자 위에 올라서 별을 다는 장면

    🔊 [음성] "잠깐만, 흔들려... 잡아줘!"

    📝 [텍스트] 민수가 균형을 잃을 뻔하자 당신이 웃으며 잡아준다.

    🖼️ [이미지] 완성된 트리 앞에서 둘이 나란히 앉은 장면

    🔊 [음성] "예쁘다... 내년엔 진짜 같이 하자. 보고싶어."

[4] 경험 종료 후:
    📱 스토리 이미지 + 음성 메시지를 앨범으로 저장
    📤 카카오톡/인스타로 공유 가능
```

### 다양한 재회 시나리오

| 상황 | 그리운 사람 | 장면 |
|------|-----------|------|
| 🎄 크리스마스 이브 | 군대 간 남친 | 트리 장식, 선물 교환 |
| 🌸 벚꽃 피는 날 | 유학 간 친구 | 한강 산책, 치맥 |
| 🥮 추석 명절 | 돌아가신 할머니 | 송편 만들기, 차례상 |
| 🎂 생일 파티 | 바쁜 엄마 | 깜짝 파티, 케이크 |
| 🏖️ 여름 바다 | 장거리 연인 | 해변 산책, 불꽃놀이 |
| 🎓 졸업식 | 먼저 간 친구 | 학사모, 기념사진 |

---

## 3. 핵심 기능 (MVP)

### F1. 페르소나 생성
- 사진 업로드 (3~10장) → 외형 학습 → 장면 이미지에 반영
- 음성 샘플 (10~30초) → 가장 유사한 HD 음성 매핑
- 카카오톡 대화 업로드 (선택) → 말투·성격·기억 학습
- 기본 정보 (이름, 관계, 성격 키워드, 특징적 습관)

### F2. 가상 재회 경험 생성 (Creative Storyteller 핵심)
- 상황 설정 (프리셋 또는 직접 입력)
- **Gemini Interleaved Output으로 장면 생성:**
  - 🖼️ 장면 이미지 (Gemini 3 Pro Image)
  - 📝 상황 텍스트 + 대화 (Gemini 3 Flash)
  - 🔊 페르소나 음성 (Gemini Live API TTS)
  - 🎵 분위기 BGM (프리셋 또는 AI 생성)
- 사용자 선택지 제공 → 스토리 분기 → 인터랙티브 경험
- 장면 간 캐릭터/배경 일관성 유지

### F3. 실시간 대화 레이어 (하이브리드)
- 장면 중간에 "직접 말하기" 모드
- Live API로 페르소나와 실시간 음성 대화
- "장면 생성 + 실시간 대화"가 자연스럽게 전환
- → 심사에서 Live Agent 요소도 어필 가능

### F4. 앨범 & 공유
- 생성된 장면 이미지 + 음성을 앨범으로 저장
- "오늘의 재회" 카드 이미지 자동 생성
- 카카오톡/인스타그램 공유 최적화

---

## 4. 기술 아키텍처

### 전체 구조

```
┌────────────────────────────────────────────────────────┐
│           사용자 (모바일 브라우저 / PWA)                    │
│          Next.js 14+ (모바일 퍼스트 웹앱)                  │
│      카메라 + 마이크 + WebSocket + Service Worker         │
└─────────────────────┬──────────────────────────────────┘
                      │ WebSocket + REST API
                      ▼
┌────────────────────────────────────────────────────────┐
│               Go Backend (Cloud Run)                     │
│                                                          │
│  ┌──────────────────────────────────────────────────┐   │
│  │        오케스트레이터 (멀티모델 라우팅)               │   │
│  │                                                    │   │
│  │  ┌────────────┐ ┌────────────┐ ┌──────────────┐  │   │
│  │  │ Persona    │ │ Story      │ │ Realtime     │  │   │
│  │  │ Engine     │ │ Engine     │ │ Voice Engine │  │   │
│  │  │            │ │            │ │              │  │   │
│  │  │ gemini-3.1 │ │ gemini-3   │ │ gemini-live  │  │   │
│  │  │ -pro       │ │ -pro-image │ │ -2.5-flash   │  │   │
│  │  │            │ │ +          │ │ -native-audio│  │   │
│  │  │ 페르소나    │ │ gemini-3   │ │              │  │   │
│  │  │ 분석/생성   │ │ -flash     │ │ 실시간 음성   │  │   │
│  │  └────────────┘ └────────────┘ └──────────────┘  │   │
│  └──────────────────────────────────────────────────┘   │
│                                                          │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐  │
│  │ Session Mgr  │  │ Persona DB   │  │ Media Cache  │  │
│  └──────────────┘  └──────────────┘  └──────────────┘  │
└───────────┬──────────────┬──────────────┬────────────────┘
            │              │              │
            ▼              ▼              ▼
   ┌──────────────┐ ┌──────────────┐ ┌──────────────┐
   │ Cloud        │ │ Firestore    │ │ Cloud        │
   │ Storage      │ │              │ │ Memorystore  │
   │ (미디어)      │ │ (페르소나/    │ │ (세션캐시)    │
   │              │ │  스토리이력)  │ │              │
   └──────────────┘ └──────────────┘ └──────────────┘
```

### 멀티모델 파이프라인 (핵심)

```
사용자: "할머니와 추석에 송편 만들기"
                │
                ▼
┌─────────────────────────────────────────────────┐
│  Step 1: 시나리오 설계 (gemini-3-flash-preview)    │
│  → 장면 순서, 대화 흐름, 분기점 설계                 │
│  → thinking_level: medium (속도+품질 균형)          │
│  출력: JSON 시나리오 구조                           │
└─────────────────┬───────────────────────────────┘
                  ▼
┌─────────────────────────────────────────────────┐
│  Step 2: 장면 생성 (gemini-3-pro-image-preview)    │
│  → System Instruction: 페르소나 외형 + 상황 설정     │
│  → ResponseModalities: ["TEXT", "IMAGE"]           │
│  → 텍스트 + 이미지 인터리브드 출력                    │
│  → 캐릭터 일관성 유지 (최대 5인, 4K 해상도)          │
│  출력: 장면 이미지 + 상황 텍스트                     │
└─────────────────┬───────────────────────────────┘
                  ▼
┌─────────────────────────────────────────────────┐
│  Step 3: 음성 생성 (gemini-live-2.5-flash)         │
│  → 페르소나에 가장 유사한 HD 음성 선택 (30종)        │
│  → 대사를 해당 음성으로 TTS 생성                     │
│  → 실시간 대화 모드 전환 가능                        │
│  출력: 음성 오디오 스트림                            │
└─────────────────┬───────────────────────────────┘
                  ▼
┌─────────────────────────────────────────────────┐
│  Step 4: 통합 & 인터랙션 (Go Orchestrator)          │
│  → 이미지 + 텍스트 + 음성 + BGM 동기화              │
│  → 사용자 선택지 → 다음 장면 트리거                   │
│  → 감정 분석 → 분위기 동적 조절                      │
│  출력: 통합된 인터랙티브 경험                        │
└─────────────────────────────────────────────────┘
```

---

## 5. Gemini 모델 사용 맵

### 모델별 역할

| 모델 | 용도 | 왜 이 모델인가 | 비용 |
|------|------|--------------|------|
| **gemini-3.1-pro-preview** | 페르소나 분석 | 사진/대화에서 성격 추출에 최강 추론 필요. 65K 토큰 출력 → 상세 프로파일 | $2/$12 per 1M tokens |
| **gemini-3-pro-image-preview** | 장면 이미지 + 텍스트 생성 | **유일한 Pro급 interleaved output 모델**. 4K 이미지, 캐릭터 일관성, 텍스트 렌더링 최강 | 이미지 생성 과금 |
| **gemini-3-flash-preview** | 시나리오 로직, 감정 분석, 분기 판단 | Pro급 추론 + 3배 속도 + 80% 저비용. thinking_level 조절로 상황별 최적화 | $0.50/$3 per 1M tokens |
| **gemini-live-2.5-flash-native-audio** | 실시간 음성 대화, TTS | **유일한 Live API 네이티브 오디오 모델**. 30개 HD 음성, 24개 언어 | Live API 과금 |
| **gemini-2.5-flash-image** | 보조 이미지 (빠른 생성) | 고속 이미지 생성. 실시간 중 추가 장면 필요 시 | 저비용 |

### thinking_level 활용 전략 (gemini-3-flash)

| 작업 | thinking_level | 이유 |
|------|:---:|------|
| 시나리오 초기 설계 | high | 복잡한 스토리 구조 필요 |
| 사용자 선택지 분기 판단 | medium | 속도 + 품질 균형 |
| 감정 분석 (실시간) | low | 빠른 반응 필요 |
| 단순 텍스트 포맷팅 | minimal | 최저 지연 |

### ⚠️ 주의사항

| 항목 | 내용 |
|------|------|
| **Gemini 2.0 Flash 퇴역** | 2026-03-31 → **사용하지 않음** ✅ |
| **Live API 구 모델 퇴역** | gemini-live-2.5-flash-preview-09-2025 → 2026-03-19 퇴역. `gemini-live-2.5-flash-native-audio` 사용 ✅ |
| **Thought Signatures** | 멀티턴 이미지 생성 시 필수. 빠뜨리면 HTTP 400 에러 |
| **imageConfig 이슈** | gemini-3-pro-image에서 imageSize/aspectRatio가 무시되는 버그 보고 있음. 테스트 필수 |

---

## 6. 기술 스택 상세

| 레이어 | 기술 | 역할 |
|--------|------|------|
| **Frontend** | Next.js 14+ (App Router, PWA) | 모바일 퍼스트 웹앱 |
| **Backend** | Go + google.golang.org/genai SDK | 멀티모델 오케스트레이션 |
| **AI - 추론** | Gemini 3.1 Pro, 3 Flash | 페르소나 분석, 시나리오 로직 |
| **AI - 이미지** | Gemini 3 Pro Image (Nano Banana Pro) | 장면 이미지 생성 (interleaved) |
| **AI - 음성** | Gemini Live 2.5 Flash Native Audio | 실시간 음성 대화, TTS |
| **DB** | Cloud Firestore | 페르소나, 스토리 이력 |
| **파일 저장** | Cloud Storage | 사진, 음성, 생성 이미지 |
| **캐시** | Cloud Memorystore (Redis) | 세션, 생성 이미지 캐시 |
| **배포** | Cloud Run | 오토스케일링, 서버리스 |
| **CI/CD** | Cloud Build | 자동 빌드/배포 |

### Go 코드 구조 (google.golang.org/genai SDK 직접 사용)

```go
// 멀티모델 클라이언트 초기화
func NewMisslessEngine(ctx context.Context) (*Engine, error) {
    client, err := genai.NewClient(ctx, &genai.ClientConfig{
        APIKey:  os.Getenv("GEMINI_API_KEY"),
        Backend: genai.BackendGeminiAPI,
    })

    return &Engine{
        // 각 엔진이 다른 모델 사용
        personaEngine: NewPersonaEngine(client, "gemini-3.1-pro-preview"),
        storyEngine:   NewStoryEngine(client, "gemini-3-pro-image-preview"),
        logicEngine:   NewLogicEngine(client, "gemini-3-flash-preview"),
        voiceEngine:   NewVoiceEngine(client, "gemini-live-2.5-flash-native-audio"),
    }, nil
}

// 장면 생성 — interleaved output
func (e *StoryEngine) GenerateScene(ctx context.Context, req SceneRequest) (*Scene, error) {
    resp, err := e.client.Models.GenerateContent(ctx,
        "gemini-3-pro-image-preview",
        genai.Text(req.Prompt),
        &genai.GenerateContentConfig{
            ResponseModalities: []string{
                string(genai.ModalityText),
                string(genai.ModalityImage),
            },
            SystemInstruction: &genai.Content{
                Parts: []genai.Part{genai.Text(req.PersonaSystemPrompt)},
            },
        },
    )

    scene := &Scene{}
    for _, part := range resp.Candidates[0].Content.Parts {
        switch {
        case part.Text != "":
            scene.AddText(part.Text)
        case part.InlineData != nil:
            scene.AddImage(part.InlineData.Data, part.InlineData.MIMEType)
        }
    }
    return scene, nil
}
```

### Go 패키지 구조

```
missless/
├── cmd/
│   └── server/
│       └── main.go                 // 엔트리포인트
├── internal/
│   ├── engine/
│   │   ├── orchestrator.go         // 멀티모델 오케스트레이터
│   │   ├── persona.go              // gemini-3.1-pro (페르소나 분석)
│   │   ├── story.go                // gemini-3-pro-image (장면 생성)
│   │   ├── logic.go                // gemini-3-flash (시나리오 로직)
│   │   └── voice.go                // gemini-live-2.5-flash (음성)
│   ├── handler/
│   │   ├── websocket.go            // WebSocket (실시간 스트리밍)
│   │   ├── api.go                  // REST API
│   │   └── media.go                // 미디어 업로드/다운로드
│   ├── model/
│   │   ├── persona.go              // 페르소나 데이터 모델
│   │   ├── scenario.go             // 시나리오 데이터 모델
│   │   └── scene.go                // 장면 데이터 모델
│   ├── storage/
│   │   ├── firestore.go            // Firestore 클라이언트
│   │   ├── gcs.go                  // Cloud Storage 클라이언트
│   │   └── cache.go                // Redis 캐시
│   └── prompt/
│       ├── persona_system.go       // 페르소나 System Instruction
│       ├── scene_prompt.go         // 장면 생성 프롬프트 빌더
│       └── scenario_template.go    // 시나리오 템플릿
├── web/                             // Next.js 프론트엔드
│   ├── app/
│   │   ├── page.tsx                // 랜딩
│   │   ├── create/page.tsx         // 페르소나 생성
│   │   ├── reunion/[id]/page.tsx   // 가상 재회 메인 화면
│   │   ├── scenarios/page.tsx      // 시나리오 선택
│   │   └── album/page.tsx          // 앨범/공유
│   ├── components/
│   │   ├── ReunionPlayer.tsx       // 재회 경험 플레이어 (핵심 UI)
│   │   ├── SceneRenderer.tsx       // 장면 렌더러 (이미지+텍스트+음성)
│   │   ├── ChoiceSelector.tsx      // 인터랙티브 선택지
│   │   ├── PersonaCreator.tsx      // 페르소나 생성 위저드
│   │   ├── VoiceChat.tsx           // 실시간 음성 대화 모드
│   │   └── AudioVisualizer.tsx     // 음성 시각화
│   ├── lib/
│   │   ├── websocket.ts            // WebSocket 클라이언트
│   │   └── audio.ts                // 오디오 처리 (PCM, AudioWorklet)
│   ├── public/
│   │   └── manifest.json           // PWA 매니페스트
│   └── next.config.js              // PWA 설정 포함
├── Dockerfile
├── go.mod
├── go.sum
└── cloudbuild.yaml
```

---

## 7. 사용자 플로우

### Flow 1: 페르소나 생성

```
[missless.co 접속 — 모바일 브라우저]
     │
     ▼
[1] "누구를 만나고 싶으세요?"
    → 이름 입력, 관계 선택
    → 성격 키워드 (다정한, 장난스러운, 차분한...)
     │
     ▼
[2] 학습 자료 업로드 (선택적, 많을수록 정확)
    📷 사진 3~10장
    🎤 음성 메모 10~30초
    💬 카카오톡 대화 내보내기
     │
     ▼
[3] AI 분석 (gemini-3.1-pro)
    "민수님의 페르소나를 만들고 있어요..."
    → 외형 특징, 말투 패턴, 성격 프로파일 추출
     │
     ▼
[4] 페르소나 프리뷰
    🖼️ AI가 생성한 민수의 이미지
    🔊 "이 목소리가 비슷한가요?" (30종 중 매칭)
    → 조정 가능
     │
     ▼
[✅ 페르소나 완성!]
```

### Flow 2: 가상 재회 경험

```
[페르소나 선택]
  "민수와 다시 만나볼까요?"
     │
     ▼
[상황 선택]
  🎄 크리스마스 이브   🌸 벚꽃 산책
  🥮 추석 명절         🎂 생일 파티
  🏖️ 여름 바다         ✏️ 직접 입력
     │
     ▼
[📱 재회 경험 시작]
  ┌──────────────────────────────────┐
  │                                    │
  │  🖼️ [장면 이미지 — 전체 화면]       │
  │                                    │
  │  ┌──────────────────────────┐     │
  │  │ "야, 이 별 장식 어디에     │     │
  │  │  달까? 맨 꼭대기?"        │     │
  │  └──────────────────────────┘     │
  │         🔊 (음성 자동 재생)         │
  │                                    │
  │  ┌────────┐ ┌────────┐ ┌──────┐  │
  │  │꼭대기에 │ │이쪽에   │ │직접  │  │
  │  │달아줘   │ │달자     │ │말하기│  │
  │  └────────┘ └────────┘ └──────┘  │
  │                                    │
  │  [다음 장면] [앨범 저장] [공유]     │
  └──────────────────────────────────┘
     │
     │  "직접 말하기" 선택 시
     ▼
  ┌──────────────────────────────────┐
  │  🎤 실시간 음성 대화 모드          │
  │                                    │
  │  🖼️ [장면 이미지 유지]             │
  │                                    │
  │  🔊 자유롭게 대화                   │
  │     "민수야, 진짜 보고싶었어"       │
  │     → AI가 민수 말투로 실시간 응답   │
  │                                    │
  │  [장면으로 돌아가기]               │
  └──────────────────────────────────┘
     │
     ▼
[재회 종료]
  📱 "오늘 민수와의 재회" 앨범 자동 생성
  📤 카카오톡/인스타 공유 카드
  💾 갤러리에 저장
```

---

## 8. 심사 기준 대응 전략

### Innovation & Multimodal UX — 40% (최중점)

| 심사 포인트 | 대응 | 점수 전략 |
|------------|------|----------|
| "텍스트 박스 탈피" | 텍스트 입력 자체가 거의 없음. 터치 + 음성으로 조작 | ★★★★★ |
| 보고, 듣고, 말하는가 | 이미지(보고) + 음성(듣고) + 실시간대화(말하고) | ★★★★★ |
| 독특한 성격/목소리 | 페르소나 시스템 — 실제 사람의 말투/성격 반영 | ★★★★★ |
| 실시간 & 컨텍스트 인지 | 장면 맥락 유지 + "직접 말하기"에서 실시간 대화 | ★★★★ |
| **차별화 킬링 포인트** | "기술 데모"가 아니라 **사람이 울 수 있는 서비스** | 🎯 |

### Technical Implementation — 30%

| 심사 포인트 | 대응 |
|------------|------|
| GenAI SDK 사용 | google.golang.org/genai SDK (Go 네이티브) |
| 다양한 Gemini 모델 활용 | **5개 모델** 적재적소 사용 (3.1 Pro, 3 Flash, 3 Pro Image, Live 2.5, 2.5 Flash Image) |
| Google Cloud 호스팅 | Cloud Run + Firestore + Cloud Storage + Memorystore |
| 할루시네이션 방지 | 페르소나 System Instruction + 대화 기반 그라운딩 |
| 에이전트 아키텍처 | 멀티모델 오케스트레이션 (모델별 전문화) |

### Demo & Presentation — 30%

**데모 영상 구성 (3분 50초):**

```
0:00 ~ 0:30  [감성 인트로]
  화면: 폰을 들고 있는 손. 연락처에 "민수 ❤️"
  내레이션: "군대 간 남자친구, 유학 간 딸, 바쁜 엄마..."
  "보고 싶은데 지금 만날 수 없다면?"
  "missless가 그 순간을 만들어드립니다."

0:30 ~ 1:00  [서비스 소개 + 아키텍처]
  missless.co 소개
  기술 아키텍처 다이어그램 (5개 Gemini 모델 조합)
  "Creative Storyteller: 텍스트·이미지·음성·음악을
   하나의 인터랙티브 스토리로"

1:00 ~ 1:30  [데모 1: 페르소나 생성]
  스마트폰 화면으로 시연
  사진 업로드 → 음성 샘플 → 카톡 대화
  AI가 페르소나 생성하는 과정

1:30 ~ 3:00  [데모 2: 가상 재회 — "민수와 크리스마스 이브"]
  🖼️ AI가 생성한 크리스마스 거실 장면
  🔊 민수 말투로 "야, 이 별 장식 어디에 달까?"
  💬 사용자가 선택지 터치 → 스토리 진행
  🎤 "직접 말하기" → 실시간 음성 대화 전환
  🖼️ 트리 완성 장면 → "내년엔 진짜 같이 하자"
  (감성적인 BGM과 함께)

3:00 ~ 3:30  [데모 3: 앨범 & 공유]
  자동 생성된 "오늘의 재회" 앨범
  카카오톡 공유 카드

3:30 ~ 3:50  [클로징]
  Cloud Run 콘솔 스크린샷 (배포 증거)
  "missless.co — 그리움을 줄여드립니다"
  QR 코드 (심사위원이 바로 접속)
```

---

## 9. 21일 스프린트 계획

### Week 1 (2/24 ~ 3/2): 기반 + 핵심 파이프라인

| 일자 | 작업 | 산출물 |
|------|------|--------|
| 2/24(월) | 도메인 등록 + GCP 프로젝트 + API 키 발급 | 인프라 셋업 |
| 2/25(화) | Go 프로젝트 초기화 + genai SDK 셋업 | 5개 모델 클라이언트 초기화 |
| 2/26(수) | **gemini-3-pro-image PoC** — interleaved output 테스트 | 텍스트+이미지 생성 확인 |
| 2/27(목) | Next.js PWA 초기화 + 모바일 UI 기본 | ReunionPlayer 컴포넌트 |
| 2/28(금) | 페르소나 생성 플로우 (업로드 + 3.1 Pro 분석) | PersonaCreator |
| 3/1(토) | Firestore + Cloud Storage 연동 | 데이터 영속화 |
| 3/2(일) | **장면 생성 파이프라인 v1** (시나리오→이미지→텍스트) | 기본 재회 경험 |

### Week 2 (3/3 ~ 3/9): 핵심 기능 완성

| 일자 | 작업 | 산출물 |
|------|------|--------|
| 3/3(월) | 인터랙티브 선택지 + 스토리 분기 구현 | ChoiceSelector |
| 3/4(화) | Live API 음성 통합 — TTS + "직접 말하기" | VoiceChat |
| 3/5(수) | 음성-장면 동기화 + BGM 레이어 | 통합 오디오 시스템 |
| 3/6(목) | 카카오톡 대화 파싱 + 말투 학습 강화 | 페르소나 정확도 향상 |
| 3/7(금) | 시나리오 프리셋 6종 구현 | 시나리오 템플릿 |
| 3/8(토) | 앨범 저장 + 공유 카드 생성 | 앨범/공유 기능 |
| 3/9(일) | 통합 테스트 + 버그 수정 | E2E 플로우 |

### Week 3 (3/10 ~ 3/16): 완성 & 제출

| 일자 | 작업 | 산출물 |
|------|------|--------|
| 3/10(월) | Cloud Run 배포 + 도메인 연결 + SSL | 라이브 서비스 |
| 3/11(화) | UI 폴리싱 + 모바일 최적화 + PWA 설정 | 완성 프론트엔드 |
| 3/12(수) | 데모 시나리오 스크립트 작성 + 리허설 | 영상 대본 |
| 3/13(목) | **데모 영상 촬영** (스마트폰 화면 녹화) | 3분 50초 영상 |
| 3/14(금) | 아키텍처 다이어그램 정리 + README | 제출 자료 |
| 3/15(토) | DevPost 제출 페이지 작성 | 제출 초안 |
| 3/16(일) | **최종 점검 + 제출** (마감 PDT 5PM = KST 3/17 9AM) | ✅ 제출 |

---

## 10. Google Cloud 서비스 사용 계획

| 서비스 | 용도 | 챌린지 체크 |
|--------|------|:-----------:|
| **Cloud Run** | Go 백엔드 + Next.js 호스팅 | ✅ 필수 |
| **Cloud Firestore** | 페르소나/스토리 이력 | ✅ 추가 GCP 서비스 |
| **Cloud Storage** | 미디어 파일 (사진, 음성, 생성 이미지) | ✅ 추가 GCP 서비스 |
| **Cloud Memorystore** | 세션 캐시, 이미지 캐시 | ✅ 추가 GCP 서비스 |
| **Cloud Build** | CI/CD | ✅ 추가 GCP 서비스 |
| **Gemini API** | 5개 모델 (3.1 Pro, 3 Flash, 3 Pro Image, Live 2.5, 2.5 Flash Image) | ✅ 필수 |

> GCP 신규 가입 시 $300 무료 크레딧. 챌린지 기간 내 충분.

---

## 11. 리스크 & 대응

| 리스크 | 확률 | 대응 |
|--------|:---:|------|
| 장면 간 캐릭터 불일관 | 중 | System Instruction에 외형 상세 기술 + 이전 장면 참조 |
| imageConfig 버그 (크기/비율 무시) | 중 | 프롬프트에 크기 지정 + 후처리 리사이즈 |
| Live API 지연 | 중 | "직접 말하기"는 선택 기능으로, 핵심은 장면 생성에 집중 |
| Thought Signatures 누락 | 고 | 멀티턴 시 반드시 이전 응답의 thought_signature 포함 |
| 음성 클로닝 한계 | 중 | 30개 HD 음성 중 최적 매칭 + 속도/톤 조절 |
| 21일 일정 촉박 | 고 | MVP 핵심(장면 생성 + 인터랙션)에 집중, 앨범/공유는 후순위 |

---

## 12. 윤리 가이드라인

1. **동의 기반**: 페르소나 대상자 본인 동의 또는 가족 동의 (돌아가신 분)
2. **투명성**: "이것은 AI가 만든 가상 경험입니다" 항상 표시
3. **건강한 사용**: 과도한 의존 방지 — "실제 소통도 중요해요" 알림
4. **데이터 보호**: 업로드 미디어 암호화 저장, 삭제 요청 즉시 처리
5. **SynthID**: Gemini가 생성한 모든 이미지에 자동 워터마크 포함

---

## 13. 제출 체크리스트

- [ ] 도메인 등록 (missless.co)
- [ ] GCP 프로젝트 생성 + API 키 발급
- [ ] Go 백엔드 + genai SDK 셋업
- [ ] 5개 Gemini 모델 연동 확인
- [ ] 페르소나 생성 기능
- [ ] **장면 생성 (interleaved output) — 핵심**
- [ ] 인터랙티브 선택지 + 스토리 분기
- [ ] Live API 음성 통합
- [ ] Next.js PWA 모바일 퍼스트 UI
- [ ] Cloud Run 배포 + missless.co 연결
- [ ] 데모 영상 제작 (4분 미만, 스마트폰 화면)
- [ ] 아키텍처 다이어그램
- [ ] Cloud 배포 증거 스크린샷
- [ ] 소스 코드 GitHub 업로드
- [ ] DevPost 제출 페이지 작성
- [ ] **제출** (3/16 PDT 5PM 전)

---

*missless.co — 그리운 사람과, 다시 만나는 순간.*
