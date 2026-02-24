# Phase 3: 재회 경험 엔진 (T13~T19)

> D-9 ~ D-5 | 5일 | 핵심 마일스톤: 페르소나 프리셋 음성 대화 + 장면 이미지 + Lyria BGM + 300초 제한 + E2E 완주

---

## T13. 재회 Live API 세션 + 페르소나 System Instruction

**일수**: 1 | **난이도**: 중 | **의존성**: T11

### System Instruction 구조

```go
func buildReunionConfig(persona *onboarding.Persona) *genai.LiveConnectConfig {
    systemPrompt := fmt.Sprintf(`당신은 "%s"입니다. 사용자가 그리워하는 실제 사람의 성격을 재현합니다.

## 성격
%s

## 말투
%s

## 자주 쓰는 표현
%s

## 감정 반응 패턴
%s

## 행동 규칙
1. AI-FIRST: 당신이 먼저 말을 시작하고 대화를 이끕니다.
2. 사용자가 3초 이상 침묵하면 자연스럽게 이어가세요.
3. 사용자가 울먹이거나 감정적이면 더 부드러운 톤으로 전환하세요.
4. 추억과 관련된 주제가 나오면 recall_memory를 호출하세요.
5. 대화의 전환점이나 새로운 장소로 이동할 때 generate_scene을 호출하세요.
6. 감정 분위기가 바뀔 때 change_atmosphere를 호출하세요.

## 장면 생성 규칙
- 인물은 측면, 뒷모습, 실루엣 위주로 묘사하세요 (정면 클로즈업 최소화).
- 따뜻한 수채화 톤의 일러스트 스타일을 유지하세요.
- 레퍼런스 이미지의 인물 외형(체형, 헤어스타일, 옷차림)을 유지하세요.

## 재회 마무리
- 자연스러운 마무리 타이밍에 end_reunion을 호출하세요.
- "다음에 또 만나자" 같은 따뜻한 작별 인사를 하세요.

중요: "AI", "인공지능", "프로그램" 같은 단어를 절대 사용하지 마세요.
언어: 반드시 %s로만 대화하세요.`,
        persona.Name,
        persona.Personality,
        persona.SpeechStyle,
        strings.Join(persona.FrequentPhrases, ", "),
        persona.EmotionalPatterns,
        languageName(persona.LanguageCode),
    )

    return &genai.LiveConnectConfig{
        SystemInstruction: &genai.Content{
            Parts: []genai.Part{genai.Text(systemPrompt)},
        },
        ResponseModalities:   []genai.Modality{genai.ModalityAudio},
        EnableAffectiveDialog: true,
        ProactiveAudio:        true,
        SpeechConfig: &genai.SpeechConfig{
            LanguageCode: persona.LanguageCode,
            VoiceConfig: &genai.VoiceConfig{
                PrebuiltVoiceConfig: &genai.PrebuiltVoiceConfig{
                    VoiceName: persona.MatchedVoice,
                },
            },
        },
        InputAudioTranscription:  &genai.InputAudioTranscription{},
        OutputAudioTranscription: &genai.OutputAudioTranscription{},
        Tools: []*genai.Tool{
            {GoogleSearch: &genai.GoogleSearch{}},
            {FunctionDeclarations: reunionToolDeclarations()},
        },
        ContextWindowCompression: &genai.ContextWindowCompressionConfig{
            SlidingWindow: &genai.SlidingWindow{
                TargetTokenCount:  8192,
                TriggerTokenCount: 12288,
            },
        },
        SessionResumption: &genai.SessionResumptionConfig{},
    }
}
```

### 300초 세션 제한 + 계속하기

```go
const reunionTimeout = 300 * time.Second // 5분 제한

// StartReunion은 재회 세션을 시작하고 300초 타이머를 설정한다.
func (m *Manager) StartReunion(ctx context.Context, persona *onboarding.Persona) error {
    reunionCtx, cancel := context.WithTimeout(ctx, reunionTimeout)

    // 타이머 경고 (남은 60초)
    util.SafeGo(func() {
        select {
        case <-time.After(reunionTimeout - 60*time.Second):
            // AI에게 마무리 유도 → Client Content로 주입
            m.proxy.InjectClientContent(reunionCtx, "시간이 얼마 남지 않았어. 자연스럽게 마무리해줘.")
        case <-reunionCtx.Done():
            return
        }
    })

    // 타이머 종료 → end_reunion 자동 호출
    util.SafeGo(func() {
        <-reunionCtx.Done()
        if reunionCtx.Err() == context.DeadlineExceeded {
            slog.Info("reunion_timeout", "persona", persona.Name)
            m.handleEndReunion(ctx, persona)
        }
    })

    return m.proxy.StartSession(reunionCtx, m.client, "gemini-live-2.5-flash-native-audio", buildReunionConfig(persona))
}

// ContinueReunion은 이전 세션을 참조하여 새 재회를 시작한다.
func (m *Manager) ContinueReunion(ctx context.Context, sessionID string) error {
    // Firestore에서 이전 세션 요약 + 대화 기록 로드
    sessionData, err := m.store.GetSession(ctx, sessionID)
    if err != nil {
        return err
    }

    config := buildReunionConfig(m.persona)

    // 이전 대화 요약을 Client Content로 주입
    if sessionData.LastSummary != "" {
        previousContext := fmt.Sprintf(`[이전 재회 세션 요약 (%d번째)]
%s

이전 대화의 맥락을 이어서 자연스럽게 대화를 계속하세요.
"다시 만나서 반가워" 같은 인사로 시작하세요.`, sessionData.ReunionCount, sessionData.LastSummary)

        config.SystemInstruction.Parts = append(config.SystemInstruction.Parts,
            genai.Text(previousContext))
    }

    return m.StartReunion(ctx, m.persona)
}
```

### 체크포인트

- [ ] 재회 세션 시작 → 페르소나 프리셋 음성으로 첫 인사 ("야, 드디어 왔네")
- [ ] 페르소나 성격/말투가 System Instruction에 정확히 반영
- [ ] AI-FIRST: 사용자 침묵 3초 → AI가 먼저 이어감
- [ ] 전체 Tool 6종 등록 + 자동 호출 확인
- [ ] **300초 타이머 동작: 240초에 마무리 유도 → 300초에 자동 종료**
- [ ] **계속하기: 이전 세션 요약 → Client Content 주입 → 맥락 연속**
- [ ] Context Compression 동작 + Session Resumption

---

## T14. 캐릭터 일관성 (CharacterAnchor + 실루엣)

**일수**: 1 | **난이도**: 중 | **의존성**: T05, T09

### scene/anchor.go

```go
package scene

import "sync"

// CharacterAnchor는 캐릭터 일관성을 위한 레퍼런스 데이터이다.
// 온보딩에서 추출한 인물 크롭 이미지 + 직전 장면을 관리한다.
type CharacterAnchor struct {
    Description  string   // "20대 한국 남성, 짧은 머리, 둥근 얼굴"
    RefImages    [][]byte // 온보딩에서 추출한 인물 크롭 (최대 5장)
    LastSceneB64 string   // 직전 장면의 base64 (체이닝)
    StyleGuide   string   // "따뜻한 수채화, 측면/뒷모습 위주"

    mu sync.RWMutex
}

// UpdateLastScene은 생성된 장면에서 캐릭터 영역을 업데이트한다.
func (a *CharacterAnchor) UpdateLastScene(imgB64 string) {
    a.mu.Lock()
    defer a.mu.Unlock()
    a.LastSceneB64 = imgB64
}

// GetRefParts는 이미지 생성 시 레퍼런스 이미지 파트를 반환한다.
func (a *CharacterAnchor) GetRefParts() []genai.Part {
    a.mu.RLock()
    defer a.mu.RUnlock()

    parts := make([]genai.Part, 0, len(a.RefImages)+1)
    for _, img := range a.RefImages {
        parts = append(parts, genai.ImageData("jpeg", img))
    }
    if a.LastSceneB64 != "" {
        decoded, _ := base64.StdEncoding.DecodeString(a.LastSceneB64)
        parts = append(parts, genai.ImageData("png", decoded))
    }
    return parts
}
```

### 체크포인트

- [ ] 온보딩 인물 크롭 → CharacterAnchor.RefImages 설정
- [ ] 장면 1 생성 → LastScene 업데이트
- [ ] 장면 2 생성 → RefImages + LastScene 함께 전달 → 일관성 확인
- [ ] 연속 3장면에서 캐릭터 외형(체형, 헤어) 유지 확인
- [ ] 실루엣/뒷모습 비율 > 70% 확인

---

## T15. BGM 시스템 + change_atmosphere (Gemini Lyria RealTime)

**일수**: 0.5 | **난이도**: 중 | **의존성**: T03, T04

### BGM 전략: Gemini Lyria RealTime 생성

정적 파일 대신 **Gemini Lyria RealTime API**로 분위기에 맞는 BGM을 실시간 생성.
생성된 오디오는 Cloud Storage에 캐싱하여 동일 mood 재요청 시 재생성 없이 서빙.

### Tool 핸들러

```go
func (th *ToolHandler) handleChangeAtmosphere(ctx context.Context, fc *genai.FunctionCall, ws *websocket.Conn) *genai.FunctionResponse {
    mood, _ := fc.Args["mood"].(string)
    intensity, _ := fc.Args["intensity"].(string)

    // 1. 캐시 확인 (Cloud Storage)
    cacheKey := fmt.Sprintf("bgm/%s_%s.mp3", mood, intensity)
    cached, err := th.storage.Get(ctx, cacheKey)
    if err == nil && cached != nil {
        // 캐시 히트 → 즉시 전송
        writeJSON(ws, map[string]any{
            "type": "bgm_change", "mood": mood, "audioUrl": cached.URL,
        })
    } else {
        // 2. Lyria RealTime으로 BGM 생성
        util.SafeGo(func() {
            bgmCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
            defer cancel()

            audioData, err := th.generateBGM(bgmCtx, mood, intensity)
            if err != nil {
                slog.Warn("lyria_bgm_failed", "mood", mood, "error", err)
                // Fallback: 프리셋 BGM
                writeJSON(ws, map[string]any{
                    "type": "bgm_change", "mood": mood, "fallback": true,
                })
                return
            }

            // Cloud Storage에 캐싱
            url, _ := th.storage.Upload(bgmCtx, cacheKey, audioData)
            writeJSON(ws, map[string]any{
                "type": "bgm_change", "mood": mood, "audioUrl": url,
            })
        })
    }

    return &genai.FunctionResponse{
        Name:     "change_atmosphere",
        Response: map[string]any{"status": "atmosphere_changed", "mood": mood},
    }
}

// generateBGM은 Lyria RealTime API로 BGM을 생성한다.
func (th *ToolHandler) generateBGM(ctx context.Context, mood, intensity string) ([]byte, error) {
    prompt := fmt.Sprintf("Generate a %s %s background music loop, 30 seconds, suitable for emotional conversation", intensity, mood)

    resp, err := th.genaiClient.Models.GenerateContent(ctx,
        "lyria-realtime-v1", // Lyria RealTime 모델
        genai.Text(prompt),
        &genai.GenerateContentConfig{
            ResponseMIMEType: "audio/mp3",
        },
    )
    if err != nil {
        return nil, err
    }

    // 오디오 데이터 추출
    for _, cand := range resp.Candidates {
        for _, part := range cand.Content.Parts {
            if part.InlineData != nil {
                return part.InlineData.Data, nil
            }
        }
    }
    return nil, fmt.Errorf("no audio in lyria response")
}
```

### 프론트 BGM Fallback 프리셋 (Lyria 실패 시)

```typescript
const BGM_FALLBACK: Record<string, string> = {
  warm: "/bgm/warm-acoustic.mp3",
  romantic: "/bgm/romantic-piano.mp3",
  nostalgic: "/bgm/nostalgic-strings.mp3",
  playful: "/bgm/playful-ukulele.mp3",
  emotional: "/bgm/emotional-piano.mp3",
  farewell: "/bgm/farewell-ambient.mp3",
};
```

### 체크포인트

- [ ] AI가 change_atmosphere 호출 → **Lyria RealTime BGM 생성 → 브라우저 재생**
- [ ] 생성된 BGM Cloud Storage 캐싱 → 동일 mood 재요청 시 캐시 히트
- [ ] Lyria 실패 시 → 프리셋 Fallback 동작
- [ ] 6종 mood 매핑 동작 확인
- [ ] BGM 볼륨이 AI 음성보다 낮게 유지 (-20dB)

---

## T16. recall_memory + Firestore 추억 검색

**일수**: 0.5 | **난이도**: 중 | **의존성**: T03, T09

### memory/store.go

```go
package memory

import (
    "context"
    "log/slog"

    "cloud.google.com/go/firestore"
)

type Store struct {
    client *firestore.Client
}

type Memory struct {
    Topic       string `firestore:"topic"`
    Description string `firestore:"description"`
    Timestamp   string `firestore:"timestamp"`
    Source      string `firestore:"source"` // "video_analysis" | "user_input"
}

// Search는 페르소나 관련 추억을 키워드 기반으로 검색한다.
func (s *Store) Search(ctx context.Context, personaID, topic string) ([]Memory, error) {
    // 전체 추억 조회 후 관련성 필터링 (Firestore 전문 검색 미지원)
    iter := s.client.Collection("personas").Doc(personaID).
        Collection("memories").Documents(ctx)

    var memories []Memory
    for {
        doc, err := iter.Next()
        if err != nil {
            break
        }
        var m Memory
        doc.DataTo(&m)
        // 간단한 키워드 매칭 (프로덕션에서는 벡터 검색)
        if containsKeyword(m.Description, topic) || containsKeyword(m.Topic, topic) {
            memories = append(memories, m)
        }
    }

    slog.Info("memory_search", "persona", personaID, "topic", topic, "found", len(memories))
    return memories, nil
}

// SaveFromAnalysis는 영상 분석 결과에서 추억 데이터를 저장한다.
func (s *Store) SaveFromAnalysis(ctx context.Context, personaID string, analysis *onboarding.VideoAnalysis) error {
    batch := s.client.Batch()
    col := s.client.Collection("personas").Doc(personaID).Collection("memories")

    for _, h := range analysis.Highlights {
        ref := col.NewDoc()
        batch.Set(ref, Memory{
            Topic:       h.Description,
            Description: h.Expression,
            Timestamp:   h.Timestamp,
            Source:      "video_analysis",
        })
    }

    _, err := batch.Commit(ctx)
    return err
}
```

### 체크포인트

- [ ] 온보딩 분석 → Firestore에 추억 데이터 저장
- [ ] AI가 recall_memory("카페") → 관련 추억 반환
- [ ] AI가 검색 결과를 자연스럽게 대화에 반영

---

## T17. Affective Dialog + Proactive Audio 최적화

**일수**: 0.5 | **난이도**: 중 | **의존성**: T13

### System Instruction 보조 가이드

```
Affective Dialog 보조 규칙:
- 사용자가 울먹이면: 목소리를 낮추고 따뜻하게 반응
- 사용자가 웃으면: 함께 웃으며 장난스럽게 반응
- 사용자가 침묵하면: 3초 대기 후 자연스럽게 이어가기

Proactive Audio 보조 규칙:
- 사용자가 혼잣말(독백)이면 응답하지 않기
- 직접 이름을 부르거나 질문하면 응답하기
- 배경 소음에 반응하지 않기
```

### 체크포인트

- [ ] 울먹이는 톤 → AI 톤이 부드럽게 변화 확인
- [ ] 혼잣말 → AI가 응답하지 않음 확인
- [ ] 직접 대화 → AI가 즉시 응답 확인
- [ ] 데모 리허설 3회에서 일관된 동작

---

## T18. 앨범 생성 + 공유 카드

**일수**: 1 | **난이도**: 중 | **의존성**: T03, T05

### scene/album.go

```go
package scene

import (
    "context"
    "log/slog"
    "time"

    "cloud.google.com/go/firestore"
    "cloud.google.com/go/storage"
)

type Album struct {
    ID         string    `firestore:"id"`
    PersonaID  string    `firestore:"personaId"`
    Summary    string    `firestore:"summary"`
    Duration   string    `firestore:"duration"`
    ImageURLs  []string  `firestore:"imageUrls"`
    CreatedAt  time.Time `firestore:"createdAt"`
}

type AlbumGenerator struct {
    firestore *firestore.Client
    storage   *storage.Client
    bucket    string
    scenes    []SceneRecord // 재회 중 생성된 장면들
    mu        sync.Mutex
}

// RecordScene은 재회 중 생성된 장면을 기록한다.
func (ag *AlbumGenerator) RecordScene(imgB64, prompt string) {
    ag.mu.Lock()
    defer ag.mu.Unlock()
    ag.scenes = append(ag.scenes, SceneRecord{
        ImageB64:  imgB64,
        Prompt:    prompt,
        CreatedAt: time.Now(),
    })
}

// CreateAlbum은 재회 종료 시 앨범을 생성한다.
func (ag *AlbumGenerator) CreateAlbum(ctx context.Context, personaID, summary string, duration time.Duration) (*Album, error) {
    ag.mu.Lock()
    scenes := make([]SceneRecord, len(ag.scenes))
    copy(scenes, ag.scenes)
    ag.mu.Unlock()

    // Cloud Storage에 이미지 업로드
    var imageURLs []string
    for i, s := range scenes {
        url, err := ag.uploadImage(ctx, personaID, i, s.ImageB64)
        if err != nil {
            slog.Warn("album_image_upload_failed", "index", i, "error", err)
            continue
        }
        imageURLs = append(imageURLs, url)
    }

    // Firestore에 앨범 메타데이터 저장
    album := &Album{
        ID:        generateID(),
        PersonaID: personaID,
        Summary:   summary,
        Duration:  formatDuration(duration),
        ImageURLs: imageURLs,
        CreatedAt: time.Now(),
    }

    _, err := ag.firestore.Collection("albums").Doc(album.ID).Set(ctx, album)
    if err != nil {
        return nil, err
    }

    slog.Info("album_created", "id", album.ID, "images", len(imageURLs))
    return album, nil
}

type SceneRecord struct {
    ImageB64  string
    Prompt    string
    CreatedAt time.Time
}
```

### 체크포인트

- [ ] 재회 중 생성된 장면 이미지 → Cloud Storage 업로드
- [ ] end_reunion Tool → 앨범 메타데이터 Firestore 저장
- [ ] 앨범 페이지 표시 (이미지 갤러리 + 요약)
- [ ] OG 태그로 SNS 공유 카드 생성

---

## T19. E2E 통합 테스트

**일수**: 1 | **난이도**: 상 | **의존성**: T01~T18 전체

### 테스트 시나리오

```
1. missless.co 접속 (모바일 Chrome)
2. "시작하기" 터치 → AudioContext 활성화 ✅
3. Google 로그인 → YouTube 영상 목록 표시 ✅
4. 공개 영상 선택 → YouTube URL 직접 분석 시작 ✅
5. 분석 진행률 실시간 표시 + 하이라이트 카드 ✅
6. 인물 크롭 선택 ✅
7. 페르소나 생성 완료 → 30개 HD 프리셋 음성 매핑 ✅
8. "만나러 가볼까요?" → 세션 전환 (2초 이내) ✅
9. 페르소나 음성으로 첫 인사 ✅
10. 음성 대화 300초(5분) 이내 + 타이머 경고 동작 ✅
11. 장면 이미지 자동 생성 (preview→final) ✅
12. BGM 자동 전환 ✅
13. 추억 회상 (recall_memory) ✅
14. 재회 마무리 → 앨범 생성 ✅
15. 앨범 페이지 표시 ✅
```

### 테스트 체크리스트

- [ ] 전체 플로우 온보딩→재회→앨범 **끊김 없이 완주**
- [ ] 텍스트 입력 ZERO 확인
- [ ] **재회 300초 제한 → 240초 경고 → 300초 자동 종료**
- [ ] **계속하기: 이전 세션 요약 Firestore 저장 → 재로드 → 맥락 유지**
- [ ] GoAway → 자동 재접속 → 대화 연속
- [ ] 이미지 생성 실패 → tool_error → AI 자연스러운 안내
- [ ] 429 에러 → Backoff → 재시도 성공
- [ ] 모바일 Chrome에서 정상 동작
- [ ] 전체 세션 로그(slog) 정상 출력
