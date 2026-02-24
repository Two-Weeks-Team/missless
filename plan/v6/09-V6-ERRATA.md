# 09-V6-ERRATA.md — V6 계획서 보완 사항 (검증 기반)

> **작성일**: 2026-02-24
> **검증 방법**: 공식 문서 직접 탐색 + Go SDK 소스코드 확인 + DevPost 규칙 페이지 조회
> **용도**: V6 문서 수정 시 이 문서의 각 항목을 순서대로 적용

---

## 🔴 Critical — 즉시 수정 (3건)

---

### C1. Flash Image 모델명 변경

**문제**: V6 전체에서 사용 중인 `gemini-2.5-flash-preview-image-generation`은 이미 **폐기된 프리뷰 모델명**

**출처**:
- [Gemini Models 공식 페이지](https://ai.google.dev/gemini-api/docs/models): 현재 이미지 생성 모델은 `gemini-2.5-flash-image` (stable)
- [Gemini Deprecations](https://ai.google.dev/gemini-api/docs/deprecations): `gemini-2.5-flash-image-preview` 폐기일 **2026-01-15** (이미 지남), stable `gemini-2.5-flash-image` 폐기일 2026-10-02

**변경**: `gemini-2.5-flash-preview-image-generation` → **`gemini-2.5-flash-image`**

**수정 위치 (7곳)**:

| # | 파일 | 라인 | 현재 내용 | 수정 내용 |
|---|------|:---:|----------|----------|
| 1 | `00-INDEX.md` | L16 | `gemini-2.5-flash-preview-image-generation` (GA) | `gemini-2.5-flash-image` (stable) |
| 2 | `00-INDEX.md` | L90 | `gemini-2.5-flash-preview-image-generation` | `gemini-2.5-flash-image` |
| 3 | `01-PREREQUISITES.md` | L76 | `gemini-2.5-flash-preview-image-generation` | `gemini-2.5-flash-image` |
| 4 | `02-PHASE1-INFRA-LIVE.md` | L667 | `// V6 모델: gemini-2.5-flash-preview-image-generation` | `// V6 모델: gemini-2.5-flash-image` |
| 5 | `02-PHASE1-INFRA-LIVE.md` | L672 | `"gemini-2.5-flash-preview-image-generation",` | `"gemini-2.5-flash-image",` |
| 6 | `05-PHASE4-DEPLOY-DEMO.md` | L321 | `gemini-2.5-flash-preview-image-generation,` | `gemini-2.5-flash-image,` |
| 7 | `06-GO-SAFETY.md` | L318 | `// Stage 1: Flash 프리뷰 (gemini-2.5-flash-preview-image-generation)` | `// Stage 1: Flash (gemini-2.5-flash-image)` |

---

### C2. Lyria API 호출 방식 전면 재설계

**문제**: V6의 T15 BGM 코드가 `GenerateContent` API로 Lyria를 호출하지만, Lyria RealTime은 **WebSocket 스트리밍 전용 API**이며 **Go SDK에 Lyria API가 존재하지 않음**

**출처**:
- [Lyria Music Generation 공식 문서](https://ai.google.dev/gemini-api/docs/music-generation): `client.aio.live.music.connect()` — WebSocket 기반 전용 프로토콜
- [Go GenAI SDK (pkg.go.dev)](https://pkg.go.dev/google.golang.org/genai): Lyria/music 관련 API **없음** — Python/JS SDK만 지원
- [Lyria RealTime Dev.to 가이드](https://dev.to/googleai/lyria-realtime-the-developers-guide-to-infinite-music-streaming-4m1h): WebSocket 기반 `WeightedPrompt` + `MusicGenerationConfig` 전용 메시지 사용

**실제 Lyria API 패턴** (Python — Go는 미지원):
```python
async with client.aio.live.music.connect(model='models/lyria-realtime-exp') as session:
    await session.set_weighted_prompts(prompts=[...])
    await session.play()
    # 2초 단위 48kHz 스테레오 오디오 청크 수신
```

**수정 위치 (1개 파일, 다수 라인)**:

| # | 파일 | 라인 | 현재 내용 | 수정 내용 |
|---|------|:---:|----------|----------|
| 1 | `04-PHASE3-REUNION.md` | L218~L295 | Lyria RealTime `GenerateContent` 호출 코드 전체 | **프리셋 BGM 전략**으로 전면 교체 |
| 2 | `04-PHASE3-REUNION.md` | L268~L292 | `generateBGM` 함수: `th.genaiClient.Models.GenerateContent(ctx, "models/lyria-realtime-exp", ...)` | 프리셋 BGM 로딩 + Cloud Storage URL 매핑 함수로 교체 |
| 3 | `04-PHASE3-REUNION.md` | L311 | 체크포인트: "Lyria RealTime BGM 생성 → 브라우저 재생" | "프리셋 BGM 선택 → 브라우저 재생 (Lyria는 Go SDK 미지원)" |
| 4 | `04-PHASE3-REUNION.md` | L313 | 체크포인트: "Lyria 실패 시 → 프리셋 Fallback 동작" | "프리셋 BGM 기본 동작 확인" |

**권장 전략**: 리뷰의 대안 #2 채택 — **프리셋 BGM만 사용**
- V6에 이미 프리셋 Fallback이 설계되어 있음 (L296~L309)
- 프리셋을 기본으로 승격, Lyria는 `07-FUTURE-DEV.md`의 추후개발로 이동
- 데모 영상에서는 프리셋 BGM으로 충분한 품질 확보 가능

---

### C3. YouTube URL 분석 — "공개 영상만" 가능 (일부공개 미지원)

**문제**: V6 전체에서 "공개/일부공개 영상" 분석 가능으로 기재되어 있으나, 실제로는 **공개(public) 영상만** 분석 가능

**출처**:
- [Gemini Video Understanding 공식 문서](https://ai.google.dev/gemini-api/docs/video-understanding): **"You can only upload public videos (not private or unlisted videos)."** (원문 그대로)
- [Google AI Developers Forum](https://discuss.ai.google.dev/t/request-allow-gemini-api-to-analyze-unlisted-youtube-videos/105083): 커뮤니티에서 unlisted 지원 요청 중이나 2026-02 현재 **미지원**

**수정 위치 (8곳)**:

| # | 파일 | 라인 | 현재 내용 | 수정 내용 |
|---|------|:---:|----------|----------|
| 1 | `00-INDEX.md` | L380 | `공개/일부공개 영상만` | `공개(public) 영상만 (일부공개 불가)` |
| 2 | `00-INDEX.md` | L389 | `→ 일부공개 변경 안내 또는 갤러리 업로드` | `→ 공개 변경 안내 또는 갤러리 업로드` |
| 3 | `03-PHASE2-ONBOARDING.md` | L101 | `공개/일부공개/비공개 상태를 분류` | `공개/비공개 상태를 분류 (일부공개는 분석 불가)` |
| 4 | `03-PHASE2-ONBOARDING.md` | L292 | `일부공개(Unlisted) 영상 → 분석 성공` | **삭제** (이 체크포인트 제거) |
| 5 | `03-PHASE2-ONBOARDING.md` | L407 | `공개/일부공개: URL 직접` | `공개만: URL 직접 (일부공개 불가)` |
| 6 | `03-PHASE2-ONBOARDING.md` | L825 | `일부공개 변경 / 갤러리 업로드` | `공개 변경 / 갤러리 업로드` |
| 7 | `05-PHASE4-DEPLOY-DEMO.md` | L78 | `YouTube에 일부공개(Unlisted)로 업로드` | `YouTube에 **공개(Public)**로 업로드` |
| 8 | `05-PHASE4-DEPLOY-DEMO.md` | L159 | `비공개/일부공개 시 심사 불가` | 이미 정확 (유지) |

**추가 영향**: 온보딩 UX에서 사용자에게 "영상을 공개로 변경해주세요" 안내 메시지 필요. 일부공개도 불가하다는 점을 명확히.

---

## 🟡 Important — 수정 권장 (4건)

---

### I1. Grand Prize 금액 수정

**문제**: `08-TRACK-ASSESSMENT.md`에 Grand Prize **$50,000**으로 기재, 실제는 **$25,000**

**출처**:
- [DevPost 공식 규칙](https://geminiliveagentchallenge.devpost.com/rules): Grand Prize — **$25,000 USD** + $3,000 Google Cloud Credits + 컨퍼런스 티켓 2장 ($2,299 each) + 여행비 $3,000×2 + Google 팀 미팅

**수정 위치**:

| # | 파일 | 라인 | 현재 내용 | 수정 내용 |
|---|------|:---:|----------|----------|
| 1 | `08-TRACK-ASSESSMENT.md` | L221 | `**Grand Prize**: $50,000` | `**Grand Prize**: $25,000 + 크레딧/티켓/여행비 (총 $38,598 상당)` |

---

### I2. Gemini 2.5 Pro 모델명 GA 전환

**문제**: V6에서 `gemini-2.5-pro-preview-03-25` (프리뷰) 사용 중이나, **GA 버전이 이미 출시**됨

**출처**:
- [Google Cloud Blog: Gemini 2.5 GA](https://cloud.google.com/blog/products/ai-machine-learning/gemini-2-5-flash-lite-flash-pro-ga-vertex-ai): 2025-06 GA 출시
- [Gemini Models 페이지](https://ai.google.dev/gemini-api/docs/models): `gemini-2.5-pro` (stable, suffix 없음)
- [Gemini Deprecations](https://ai.google.dev/gemini-api/docs/deprecations): `gemini-2.5-pro` 폐기일 2026-06-17 (해커톤 3/16 이후이므로 안전)

**변경**: `gemini-2.5-pro-preview-03-25` → **`gemini-2.5-pro`**

**수정 위치 (10곳)**:

| # | 파일 | 라인 | 현재 내용 |
|---|------|:---:|----------|
| 1 | `00-INDEX.md` | L18 | `gemini-2.5-pro-preview-03-25` |
| 2 | `00-INDEX.md` | L92 | `gemini-2.5-pro-preview-03-25` |
| 3 | `00-INDEX.md` | L93 | `gemini-2.5-pro-preview-03-25` |
| 4 | `01-PREREQUISITES.md` | L75 | `gemini-2.5-pro-preview-03-25` |
| 5 | `02-PHASE1-INFRA-LIVE.md` | (검색) | 존재 시 수정 |
| 6 | `03-PHASE2-ONBOARDING.md` | L206 | `// V6 모델: gemini-2.5-pro-preview-03-25` |
| 7 | `03-PHASE2-ONBOARDING.md` | L215 | `"gemini-2.5-pro-preview-03-25",` |
| 8 | `03-PHASE2-ONBOARDING.md` | L252 | `"gemini-2.5-pro-preview-03-25",` |
| 9 | `03-PHASE2-ONBOARDING.md` | L477+L482 | `gemini-2.5-pro-preview-03-25` |
| 10 | `05-PHASE4-DEPLOY-DEMO.md` | L322 | `gemini-2.5-pro-preview-03-25` |
| 11 | `06-GO-SAFETY.md` | L697 | `gemini-2.5-pro-preview-03-25` |

모든 위치에서 → **`gemini-2.5-pro`** 로 일괄 변경

---

### I3. Go SDK Session 메서드 시그니처 수정

**문제**: V6의 Live API 코드가 실제 Go SDK 시그니처와 불일치

**출처**:
- [Go GenAI SDK pkg.go.dev](https://pkg.go.dev/google.golang.org/genai): Session 타입 — `SendRealtimeInput(LiveRealtimeInput)`, `Receive()`, `SendToolResponse(LiveToolResponseInput)`
- [go-genai/live_test.go](https://github.com/googleapis/go-genai/blob/main/live_test.go): 실제 사용 패턴 확인
- [go-genai/live.go](https://github.com/googleapis/go-genai/blob/main/live.go): Session 메서드 정의 확인

**변경 내용**:

| V6 코드 (현재) | 실제 SDK 시그니처 | 차이점 |
|----------------|-----------------|--------|
| `session.SendRealtimeInput(ctx, &genai.RealtimeInput{Audio: &genai.Blob{...}})` | `session.SendRealtimeInput(genai.LiveRealtimeInput{Audio: &genai.Blob{...}})` | ctx 없음, 타입명 `LiveRealtimeInput` |
| `session.Receive(ctx)` | `session.Receive()` | ctx 파라미터 없음 |
| `session.SendToolResponse(ctx, response)` | `session.SendToolResponse(genai.LiveToolResponseInput{...})` | ctx 없음, 전용 Input 타입 |

**수정 위치**:

| # | 파일 | 라인 | 현재 코드 | 수정 코드 |
|---|------|:---:|----------|----------|
| 1 | `02-PHASE1-INFRA-LIVE.md` | L243 | `session.SendRealtimeInput(ctx, &genai.RealtimeInput{` | `session.SendRealtimeInput(genai.LiveRealtimeInput{` |
| 2 | `02-PHASE1-INFRA-LIVE.md` | L276 | `msg, err := session.Receive(ctx)` | `msg, err := session.Receive()` |
| 3 | `02-PHASE1-INFRA-LIVE.md` | L324 | `session.SendToolResponse(ctx, response)` | `session.SendToolResponse(genai.LiveToolResponseInput{FunctionResponses: response})` |

---

### I4. Category Winner 상금 및 Honorable Mentions 추가

**문제**: V6에 **Category Winner($10,000)** 와 **Honorable Mentions(5×$2,000)** 정보가 누락됨

**출처**:
- [DevPost 공식 규칙](https://geminiliveagentchallenge.devpost.com/rules):
  - **Category Winner** (트랙별 1팀, 3개 트랙): $10,000 + $1,000 크레딧 + 티켓 2장 + Google 팀 미팅
  - **Honorable Mentions** (5팀): $2,000 + $500 크레딧

**추가 위치**:

| # | 파일 | 위치 | 추가 내용 |
|---|------|------|----------|
| 1 | `08-TRACK-ASSESSMENT.md` | 섹션 6 Grand Prize 분석 이후 | Category Winner $10,000 상금 정보 + missless 수상 시 최소 수상 경로 |
| 2 | `00-INDEX.md` | 채점 기준 매트릭스 근처 | 전체 상금 구조 표 추가 |

**추가할 상금 구조 표**:

```markdown
| 상 | 상금 | 수 | missless 타겟 |
|---|:---:|:---:|:---:|
| Grand Prize | $25,000 + 부가혜택 | 1팀 | 🎯 최종 목표 |
| Category Winner (Creative Storyteller) | $10,000 + $1,000 크레딧 | 3팀 (트랙별 1) | 🎯 1차 목표 |
| Subcategory Winner | $5,000 + $500 크레딧 | 3팀 | 🎯 동시 타겟 |
| Honorable Mention | $2,000 + $500 크레딧 | 5팀 | 안전망 |
```

---

## 🟢 참고 사항 — 리뷰 오류 정정 (3건)

> 아래는 리뷰가 잘못 지적한 항목으로, **V6 문서가 올바르며 수정 불필요**

---

### R1. ContextWindowCompression — V6 정확, 리뷰 오류

**리뷰 주장**: "공식 문서에서 명시적 언급 없음" → ⚠️ 미확인

**실제 확인**:
- [Go GenAI SDK live_test.go](https://github.com/googleapis/go-genai/blob/main/live_test.go): `ContextWindowCompressionConfig{TriggerTokens: ..., SlidingWindow: &SlidingWindow{TargetTokens: ...}}` 테스트 코드 존재
- [Gemini Session Management 문서](https://ai.google.dev/gemini-api/docs/live-session): 서버사이드 sliding window로 가장 오래된 턴 자동 프루닝

**결론**: V6의 SlidingWindow 설계는 **정확하고 유효**. 수정 불필요.

---

### R2. SessionResumption — V6 정확, 리뷰 오류

**리뷰 주장**: "공식 문서에서 명시적 언급 없음" → ⚠️ 미확인

**실제 확인**:
- [Go GenAI SDK live_test.go](https://github.com/googleapis/go-genai/blob/main/live_test.go): `SessionResumptionConfig{Handle: "test_handle", Transparent: true}` 테스트 코드 존재
- 단, 테스트 주석: "transparent parameter is not supported in Gemini API" → Developer API에서 Transparent 모드 미지원, Vertex AI에서는 지원 가능

**결론**: V6의 SessionResumption 설계는 **유효**. `Handle` 기반 재연결은 동작함. `Transparent` 옵션만 Developer API에서 미지원이므로 주의.

---

### R3. Subcategory 상 명칭 — V6 정확, 리뷰 오류

**리뷰 주장**: "구체적 명칭 미공개" → 리스크 낮음

**실제 확인**:
- [DevPost 공식 규칙](https://geminiliveagentchallenge.devpost.com/rules): 3개 Subcategory 명칭 **명확히 공개됨**
  1. **Best Multimodal Integration & User Experience** ($5,000)
  2. **Best Technical Execution & Agent Architecture** ($5,000)
  3. **Best Innovation & Thought Leadership** ($5,000)

**결론**: V6 문서(08-TRACK-ASSESSMENT)의 Subcategory 명칭이 **공식 명칭과 일치**. 수정 불필요.

---

## 적용 순서 권장

```
1단계 (즉시): C1 → C2 → C3  (빌드 브레이킹 이슈)
2단계 (PRE 단계): I1 → I2 → I3 → I4  (정확성 개선)
3단계 (확인): R1~R3 참고하여 불필요한 수정 방지
```

### 적용 후 검증

- [ ] `grep -r "gemini-2.5-flash-preview-image-generation" plan/v6/` → 0건
- [ ] `grep -r "일부공개" plan/v6/` → 0건 (데모 영상 업로드 안내 제외)
- [ ] `grep -r "gemini-2.5-pro-preview-03-25" plan/v6/` → 0건
- [ ] `grep -r "50,000" plan/v6/` → 0건
- [ ] `grep -r "GenerateContent.*lyria" plan/v6/` → 0건

---

*본 문서는 2026-02-24 기준 공식 문서 직접 확인을 통해 작성되었습니다. 각 항목의 출처 URL을 통해 재검증 가능합니다.*
