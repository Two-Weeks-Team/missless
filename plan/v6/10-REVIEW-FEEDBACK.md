# 10-REVIEW-FEEDBACK.md — V6 리뷰 피드백 보고서

> **작성일**: 2026-02-24
> **대상**: `09-V6-REVIEW.md` (외부 리뷰)
> **검증 방법**: 각 항목별 공식 문서 직접 조회, Go SDK 소스코드 확인, DevPost 규칙 페이지 조회

---

## 판정 기준

| 표시 | 의미 | 조치 |
|:---:|------|------|
| ✅ 수용 | 리뷰 의견 타당, V6 수정 예정 | `09-V6-ERRATA.md` 기반 수정 진행 |
| 🟡 부분 수용 | 일부 타당하나 리뷰 대안에도 오류 | 정확한 내용으로 별도 수정 |
| ❌ 기각 | 리뷰 의견 부정확, V6 원본 유지 | 출처 근거와 함께 기각 사유 명시 |

---

## 섹션 1: 모델명 및 API 검증

### 1.1 Live API 모델명 — ✅ 수용

> **리뷰 의견**: Developer API `gemini-2.5-flash-native-audio-preview-12-2025`는 프리뷰. 구 프리뷰 `09-2025`는 3/19 폐기.

**검증 결과**: 타당

| 확인 사항 | 출처 | 인용 |
|-----------|------|------|
| Developer API 현재 모델명 | [ai.google.dev/gemini-api/docs/live](https://ai.google.dev/gemini-api/docs/live) | 코드 예시에 `gemini-2.5-flash-native-audio-preview-12-2025` 사용 중 |
| Vertex AI GA 모델명 | [docs.cloud.google.com/.../2-5-flash-live-api](https://docs.cloud.google.com/vertex-ai/generative-ai/docs/models/gemini/2-5-flash-live-api) | `gemini-live-2.5-flash-native-audio` — GA, 출시 2025-12-12 |
| 구 프리뷰 폐기 | [ai.google.dev/gemini-api/docs/deprecations](https://ai.google.dev/gemini-api/docs/deprecations) | `gemini-live-2.5-flash-preview` shutdown → `preview-12-2025`로 교체됨 |

**조치**: PRE-12에서 Developer API GA 모델명 존재 여부 확인. V6 코드는 현재 유효하므로 즉시 수정 불필요.

---

### 1.2 Flash Image 모델명 — ✅ 수용

> **리뷰 의견**: `gemini-2.5-flash-preview-image-generation` 프리뷰 종료 → stable 전환

**검증 결과**: 정확. 수정 필수.

| 확인 사항 | 출처 | 인용 |
|-----------|------|------|
| 현재 stable 모델명 | [ai.google.dev/gemini-api/docs/models](https://ai.google.dev/gemini-api/docs/models) | 이미지 생성 모델 = `gemini-2.5-flash-image` (stable) |
| 프리뷰 모델 폐기일 | [ai.google.dev/gemini-api/docs/deprecations](https://ai.google.dev/gemini-api/docs/deprecations) | `gemini-2.5-flash-image-preview` 폐기일 **2026-01-15** (이미 지남) |
| stable 모델 폐기일 | 동일 출처 | `gemini-2.5-flash-image` 폐기일 **2026-10-02** |

**조치**: V6 전체 7곳 수정 → `gemini-2.5-flash-image` (상세: `09-V6-ERRATA.md` C1)

---

### 1.3 Gemini 2.5 Pro 모델명 — ✅ 수용

> **리뷰 의견**: GA가 2025-06 출시. GA 모델명 따로 있을 수 있음.

**검증 결과**: 정확. GA 모델명 확인됨.

| 확인 사항 | 출처 | 인용 |
|-----------|------|------|
| GA 출시 확인 | [cloud.google.com/blog/.../gemini-2-5-flash-lite-flash-pro-ga-vertex-ai](https://cloud.google.com/blog/products/ai-machine-learning/gemini-2-5-flash-lite-flash-pro-ga-vertex-ai) | 2025-06 Gemini 2.5 Pro GA 출시 |
| GA 모델명 | [ai.google.dev/gemini-api/docs/models](https://ai.google.dev/gemini-api/docs/models) | **`gemini-2.5-pro`** (suffix 없음) |
| 폐기일 | [ai.google.dev/gemini-api/docs/deprecations](https://ai.google.dev/gemini-api/docs/deprecations) | 2026-06-17 (해커톤 3/16 이후 → 안전) |

**조치**: V6 전체 11곳 수정 → `gemini-2.5-pro` (상세: `09-V6-ERRATA.md` I2)

---

### 1.4 Lyria API 호출 방식 — ✅ 수용

> **리뷰 의견**: "중대 오류" — GenerateContent가 아닌 WebSocket 방식. Go SDK 미지원 가능.

**검증 결과**: 100% 정확. V6의 T15 코드 전면 재설계 필요.

| 확인 사항 | 출처 | 인용 |
|-----------|------|------|
| Lyria는 WebSocket 전용 | [ai.google.dev/gemini-api/docs/music-generation](https://ai.google.dev/gemini-api/docs/music-generation) | `client.aio.live.music.connect(model='models/lyria-realtime-exp')` — Python/JS 예시만 |
| Go SDK에 Lyria 없음 | [pkg.go.dev/google.golang.org/genai](https://pkg.go.dev/google.golang.org/genai) | "Lyria/music 관련 API 없음" — 패키지 전체 검색 결과 |
| WebSocket 기반 확인 | [dev.to/googleai/lyria-realtime](https://dev.to/googleai/lyria-realtime-the-developers-guide-to-infinite-music-streaming-4m1h) | "persistent, bidirectional, low-latency streaming connection using WebSocket" |

**조치**: T15의 `GenerateContent` 코드 삭제 → 프리셋 BGM 전략으로 전환 (상세: `09-V6-ERRATA.md` C2)

---

## 섹션 2: Go GenAI SDK 인터페이스

### 2.1 Live API 지원 — 🟡 부분 수용

> **리뷰 의견**: V6의 `SendRealtimeInput` 시그니처가 잘못됨. 실제는 `session.Send(&genai.LiveClientMessage{RealtimeInput: &genai.LiveClientRealtimeInput{MediaChunks: []*genai.Blob{...}}})`

**검증 결과**: V6 코드 수정은 필요하나, **리뷰가 제시한 대안 코드도 부정확**.

| 확인 사항 | 출처 | 인용 |
|-----------|------|------|
| 실제 메서드 이름 | [pkg.go.dev/google.golang.org/genai](https://pkg.go.dev/google.golang.org/genai) | `SendRealtimeInput(input LiveRealtimeInput) error` — 헬퍼 메서드 존재 |
| 실제 사용 패턴 | [github.com/googleapis/go-genai/blob/main/live_test.go](https://github.com/googleapis/go-genai/blob/main/live_test.go) | `session.SendRealtimeInput(genai.LiveRealtimeInput{Audio: &genai.Blob{Data: ..., MIMEType: "audio/pcm"}})` |
| Receive 시그니처 | 동일 출처 | `session.Receive()` — **ctx 파라미터 없음** |
| SendToolResponse | 동일 출처 | `session.SendToolResponse(genai.LiveToolResponseInput{...})` |

**리뷰 대안 코드의 오류**:
- 리뷰: `session.Send(&genai.LiveClientMessage{RealtimeInput: &genai.LiveClientRealtimeInput{MediaChunks: ...}})`
- 실제: `session.SendRealtimeInput(genai.LiveRealtimeInput{Audio: &genai.Blob{...}})` — 저수준 `Send` 대신 **전용 헬퍼 메서드** 사용
- `LiveClientRealtimeInput`이 아닌 `LiveRealtimeInput` 타입
- `MediaChunks`가 아닌 `Audio` / `Video` / `Text` 필드

**조치**: V6 코드를 리뷰 제안이 아닌 **실제 SDK 시그니처**로 수정 (상세: `09-V6-ERRATA.md` I3)

---

### 2.2 ADK Go — ✅ 수용

> **리뷰 의견**: 코드에서 ADK의 Agent/Tool 클래스를 실제로 사용하는 부분이 없음

**검증 결과**: 타당한 지적. ADK 의존성은 있으나 실제 활용 코드 부재.

**조치**: Phase 2에서 ADK Agent 클래스 활용 코드 추가 검토. 제출 시 "ADK 사용" 어필을 위해 최소 1개 Agent 클래스 실사용 필요.

---

## 섹션 3: YouTube URL 분석

### 공개상태 제약 — ✅ 수용

> **리뷰 의견**: 공개(public) 영상**만** 분석 가능. 일부공개(unlisted) 미지원.

**검증 결과**: 정확.

| 확인 사항 | 출처 | 인용 (원문) |
|-----------|------|------|
| 공식 문서 원문 | [ai.google.dev/gemini-api/docs/video-understanding](https://ai.google.dev/gemini-api/docs/video-understanding) | **"You can only upload public videos (not private or unlisted videos)."** |
| 커뮤니티 요청 | [discuss.ai.google.dev/t/...unlisted.../105083](https://discuss.ai.google.dev/t/request-allow-gemini-api-to-analyze-unlisted-youtube-videos/105083) | unlisted 지원 요청 중이나 2026-02 현재 미반영 |

**조치**: V6 전체 8곳 수정 (상세: `09-V6-ERRATA.md` C3)

---

## 섹션 4: Live API 고급 기능

### AffectiveDialog, ProactiveAudio — ✅ 리뷰와 일치

> **리뷰 의견**: 확인됨 ✅

**추가 검증**: 정확.

| 기능 | 출처 | 인용 |
|------|------|------|
| `enable_affective_dialog` | [ai.google.dev/gemini-api/docs/live-guide](https://ai.google.dev/gemini-api/docs/live-guide) | "lets Gemini adapt its response style to the input expression and tone" |
| `proactivity.proactive_audio` | 동일 출처 | "Gemini can proactively decide not to respond if the content is not relevant" |

---

### ContextWindowCompression — ❌ 기각 (리뷰 오류)

> **리뷰 의견**: "공식 문서에서 명시적 언급 없음" → ⚠️ 미확인

**검증 결과**: **Go SDK에 구현되어 있으며 테스트 코드도 존재**. 리뷰의 판단 부정확.

| 확인 사항 | 출처 | 인용 |
|-----------|------|------|
| Go SDK 테스트 코드 | [github.com/googleapis/go-genai/blob/main/live_test.go](https://github.com/googleapis/go-genai/blob/main/live_test.go) | `ContextWindowCompressionConfig{TriggerTokens: Ptr[int64](1000), SlidingWindow: &SlidingWindow{TargetTokens: Ptr[int64](500)}}` |
| 세션 관리 문서 | [ai.google.dev/gemini-api/docs/live-session](https://ai.google.dev/gemini-api/docs/live-session) | sliding window로 가장 오래된 턴 자동 프루닝 |
| 웹 검색 결과 | [medium.com/.../session-management-with-googles-multimodal-live-api](https://medium.com/google-cloud/session-management-with-googles-multimodal-live-api-f70a162a4374) | ContextWindowCompression 설정 가이드 |

**결론**: V6의 SlidingWindow 기반 Context 관리 설계는 **정확하고 유효**. 수정 불필요.

**리뷰 권고("수동 요약 + Client Content 재주입") 불필요** — SDK 레벨에서 자동 처리됨.

---

### SessionResumption — ❌ 기각 (리뷰 오류)

> **리뷰 의견**: "공식 문서에서 명시적 언급 없음" → ⚠️ 미확인

**검증 결과**: **Go SDK에 구현되어 있으며 테스트 코드도 존재**.

| 확인 사항 | 출처 | 인용 |
|-----------|------|------|
| Go SDK 테스트 코드 | [github.com/googleapis/go-genai/blob/main/live_test.go](https://github.com/googleapis/go-genai/blob/main/live_test.go) | `SessionResumptionConfig{Handle: "test_handle", Transparent: true}` |
| 세션 관리 문서 | [ai.google.dev/gemini-api/docs/live-session](https://ai.google.dev/gemini-api/docs/live-session) | SessionResumptionUpdate 메시지로 session_id + resumption token 수신 → 재연결 시 사용 |
| Transparent 제한 | 동일 테스트 파일 주석 | "transparent parameter is not supported in Gemini API" (Developer API 한정) |

**결론**: V6의 SessionResumption 설계 **유효**. `Handle` 기반 재연결은 동작함.

⚠️ **단, Developer API에서 `Transparent` 모드 미지원** — 이 점만 V6에 주석 추가 권장.

---

### InputAudioTranscription / OutputAudioTranscription — 리뷰 "확인 필요" → 확인됨

| 기능 | 출처 | 인용 |
|------|------|------|
| `output_audio_transcription` | [ai.google.dev/gemini-api/docs/live-guide](https://ai.google.dev/gemini-api/docs/live-guide) | "enables transcription of the model's audio output" |
| `input_audio_transcription` | 동일 출처 | "enables transcription of the model's audio input" |

**결론**: 두 기능 모두 존재 확인. V6 설계 유효.

---

## 섹션 5: DevPost 공식 규칙

### (A) Grand Prize 금액 — ✅ 수용

> **리뷰 의견**: $50,000 → $25,000

**검증 결과**: 정확.

| 확인 사항 | 출처 | 인용 |
|-----------|------|------|
| Grand Prize 금액 | [geminiliveagentchallenge.devpost.com/rules](https://geminiliveagentchallenge.devpost.com/rules) | **$25,000 USD** + $3,000 크레딧 + 티켓 2장($2,299 each) + 여행비 $3,000×2 |

**조치**: `08-TRACK-ASSESSMENT.md` L221 수정 (상세: `09-V6-ERRATA.md` I1)

---

### (B) Subcategory 상 명칭 — ❌ 기각 (리뷰 오류)

> **리뷰 의견**: "3개 subcategory winner 존재 ($5,000 each), **구체적 명칭 미공개**"

**검증 결과**: **명칭이 DevPost에 명확히 공개되어 있음**. 리뷰의 판단 부정확.

| 확인 사항 | 출처 | 인용 |
|-----------|------|------|
| Subcategory 명칭 | [geminiliveagentchallenge.devpost.com/rules](https://geminiliveagentchallenge.devpost.com/rules) | 1. **Best Multimodal Integration & User Experience** ($5,000) |
| | 동일 출처 | 2. **Best Technical Execution & Agent Architecture** ($5,000) |
| | 동일 출처 | 3. **Best Innovation & Thought Leadership** ($5,000) |

**V6 문서의 명칭 대조**:

| V6 (08-TRACK-ASSESSMENT) | DevPost 공식 | 일치 |
|--------------------------|-------------|:---:|
| Best Multimodal Integration & UX | Best Multimodal Integration & **User Experience** | ✅ (약칭) |
| Best Technical Execution & Architecture | Best Technical Execution & **Agent** Architecture | ✅ (약칭) |
| Best Innovation & Thought Leadership | Best Innovation & Thought Leadership | ✅ 완전일치 |

**결론**: V6 문서의 Subcategory 명칭은 **공식 명칭과 일치**. 리뷰의 "미공개" 주장은 오류.

---

## 섹션 6: 리스크 분류 재평가

### 리뷰의 Critical 분류 재평가

| 리뷰 # | 리뷰 판정 | 검증 결과 | 최종 조치 |
|---------|:---:|:---:|---------|
| C1 | 🔴 Lyria API 전면 오류 | ✅ 정확 | **수용** — 프리셋 BGM 전환 |
| C2 | 🔴 Flash Image 모델명 폐기 | ✅ 정확 | **수용** — 7곳 수정 |
| C3 | 🔴 YouTube 일부공개 미지원 | ✅ 정확 | **수용** — 8곳 수정 |
| C4 | 🔴 Grand Prize 금액 오기 | ✅ 정확 | **수용** — 다만 🟡로 하향 (전략 영향 없음) |

### 리뷰의 High 분류 재평가

| 리뷰 # | 리뷰 판정 | 검증 결과 | 최종 조치 |
|---------|:---:|:---:|---------|
| H1 | 🟡 Developer API 모델명 유효성 | ⚠️ PRE-12 확인 | **수용** — 확인 계속 |
| H2 | 🟡 2.5 Pro GA 모델명 | ✅ `gemini-2.5-pro` 확인 | **수용** — 11곳 수정 |
| H3 | 🟡 Go SDK 시그니처 | 🟡 리뷰 대안도 부정확 | **부분 수용** — 정확한 시그니처로 수정 |
| H4 | 🟡 ContextWindowCompression | ❌ **Go SDK에 존재** | **기각** — V6 유효 |
| H5 | 🟡 SessionResumption | ❌ **Go SDK에 존재** | **기각** — V6 유효 (Transparent만 주의) |
| H6 | 🟡 Lyria Go SDK 지원 | ❌ 미지원 확정 | **수용** — 프리셋 BGM |
| H7 | 🟡 Flash Image stable 모델명 | ✅ `gemini-2.5-flash-image` | **수용** — C1과 동일 |

### 리뷰의 Medium 분류 재평가

| 리뷰 # | 리뷰 판정 | 검증 결과 | 최종 조치 |
|---------|:---:|:---:|---------|
| M1 | 🟡 ADK 실제 활용 | 🟡 타당 | **수용** — ADK 활용 코드 추가 검토 |
| M2 | 🟡 interleaved output 서술 | ✅ 이미 V6 대응 완료 | **수용 불필요** — 이전 세션에서 수정됨 |
| M3 | 🟡 데모 언어 | 🟡 타당 | **수용** — 전략 확정 필요 |
| M4 | 🟡 채점 5.82 비현실적 | 🟡 의견 참고 | **참고** — 보수적 해석 권장하지만 근거는 있음 |
| M5 | 🟡 텍스트 입력 접근성 | 🟢 좋은 제안 | **참고** — 우선순위 낮음 |

---

## 종합 통계

| 판정 | 건수 | 비율 |
|:---:|:---:|:---:|
| ✅ 수용 (수정 진행) | **14건** | 64% |
| 🟡 부분 수용 (리뷰 대안 수정 후 수용) | **2건** | 9% |
| ❌ 기각 (V6 원본 유지) | **3건** | 14% |
| 참고 (전략적 판단 필요) | **3건** | 14% |

**리뷰 품질 평가**: 전반적으로 높은 품질의 기술 리뷰. 특히 Lyria API, Flash Image 모델명, YouTube 공개상태 지적은 **프로젝트 실패를 방지할 수 있는 Critical 발견**. 다만 ContextWindowCompression/SessionResumption 미확인 판정과 Subcategory 명칭 미공개 주장은 **추가 검증 부족**에서 기인한 오류.

---

*수정 작업은 `09-V6-ERRATA.md`의 순서대로 진행합니다.*
