# Go 위험요소 분류 + 베스트프랙티스 (전체 태스크)

> missless.co V7 | Go 백엔드 안전성 가이드
> 대상: WebSocket 프록시 + Gemini Live API + 동시성 이미지 생성 + Firestore
> Go 1.22 + google.golang.org/genai v1.47.0

---

## 위험 분류 체계

| 등급 | 설명 | 영향 |
|:---:|------|------|
| 🔴 CRITICAL | 서버 즉시 크래시 또는 데이터 손실 | Panic, Deadlock |
| 🟠 HIGH | 리소스 고갈로 점진적 서비스 장애 | Goroutine Leak, Memory Leak |
| 🟡 MEDIUM | 성능 저하 또는 응답 지연 | Bottleneck, Race Condition |
| 🟢 LOW | 잠재적 문제 (프로덕션에서 발현) | Context 누락, 로깅 부재 |

---

## 1. Panic — 🔴 CRITICAL

### 1.1 위험 설명

Go에서 goroutine 내 미복구 panic은 **전체 프로세스를 종료**시킨다. HTTP 미들웨어의 recovery는 동일 goroutine에서만 동작하며, `go func()` 안에서 발생한 panic은 잡지 못한다.

### 1.2 발생 위치 (태스크별)

| 태스크 | 위치 | 원인 | 위험도 |
|--------|------|------|:---:|
| T02 | `live/proxy.go` — `forwardBrowserToLive()` | Live API 연결 끊김 시 nil conn 접근 | 🔴 |
| T02 | `live/proxy.go` — `forwardLiveToBrowser()` | WebSocket write on closed conn | 🔴 |
| T03 | `live/tools.go` — `handleGenerateScene()` | 이미지 생성 goroutine 내부 panic | 🔴 |
| T05 | `scene/generator.go` — `GeneratePro()` | genai 응답 파싱 중 nil pointer | 🔴 |
| T09 | `onboarding/pipeline.go` — `Run()` | 영상 분석 중 unexpected API 응답 | 🟠 |
| T11 | `session/manager.go` — `TransitionToReunion()` | 세션 교체 중 race + nil | 🔴 |

### 1.3 방어 패턴

#### 패턴 A: `safeGo()` — 모든 goroutine 래퍼

```go
// internal/util/safego.go
package util

import (
    "context"
    "log/slog"
    "runtime/debug"
)

// SafeGo wraps a goroutine with panic recovery.
// 모든 go func() 대신 util.SafeGo(func() { ... }) 사용.
func SafeGo(fn func()) {
    go func() {
        defer func() {
            if r := recover(); r != nil {
                slog.Error("GOROUTINE PANIC",
                    "error", r,
                    "stack", string(debug.Stack()),
                )
                // TODO: metrics counter increment
            }
        }()
        fn()
    }()
}

// SafeGoWithContext는 panic 시 cancel()도 호출한다.
func SafeGoWithContext(cancel context.CancelFunc, fn func()) {
    go func() {
        defer func() {
            if r := recover(); r != nil {
                slog.Error("GOROUTINE PANIC (with cancel)",
                    "error", r,
                    "stack", string(debug.Stack()),
                )
                cancel() // 부모 context에 전파
            }
        }()
        fn()
    }()
}
```

#### 패턴 B: HTTP + WebSocket 미들웨어

```go
// internal/middleware/recovery.go
func Recovery(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        defer func() {
            if err := recover(); err != nil {
                slog.Error("HTTP PANIC",
                    "method", r.Method,
                    "path", r.URL.Path,
                    "error", err,
                    "stack", string(debug.Stack()),
                )
                w.WriteHeader(http.StatusInternalServerError)
                json.NewEncoder(w).Encode(map[string]string{
                    "error": "internal server error",
                })
            }
        }()
        next.ServeHTTP(w, r)
    })
}
```

#### 패턴 C: WebSocket 메시지 루프

```go
// proxy.go — 메시지 수신 루프
func (p *Proxy) forwardBrowserToLive(ctx context.Context) {
    defer func() {
        if r := recover(); r != nil {
            slog.Error("WS_READ PANIC", "error", r)
        }
    }()

    for {
        select {
        case <-ctx.Done():
            return
        default:
        }

        p.browserConn.SetReadDeadline(time.Now().Add(30 * time.Second))
        _, msg, err := p.browserConn.ReadMessage()
        if err != nil {
            return // 정상 종료
        }
        // 개별 메시지 처리도 recover 감싸기
        func() {
            defer func() {
                if r := recover(); r != nil {
                    slog.Error("MSG_PROCESS PANIC", "error", r)
                }
            }()
            p.processMessage(ctx, msg)
        }()
    }
}
```

### 1.4 적용 규칙

- ✅ **모든** `go func()` → `util.SafeGo()` 교체
- ✅ HTTP 핸들러: `middleware.Recovery()` 적용
- ✅ WebSocket 루프: 외부 `defer recover()` + 메시지별 내부 `recover()`
- ✅ nil 체크: Live API 응답, genai 결과 접근 전 반드시 nil 확인
- ❌ **절대 금지**: `panic()` 직접 사용 (초기화 단계 `config` 검증 외)

---

## 2. Race Condition — 🟡 MEDIUM~🔴 CRITICAL

### 2.1 발생 위치 (태스크별)

| 태스크 | 위치 | 공유 자원 | 동기화 방식 |
|--------|------|-----------|:-----------:|
| T02 | `proxy.go` — `liveSession` | Live API 세션 포인터 | `sync.Mutex` |
| T03 | `tools.go` — `ToolHandler` 등록 | tool map | `sync.RWMutex` |
| T05 | `generator.go` — `CharacterAnchor` | 레퍼런스 이미지 | `sync.RWMutex` |
| T11 | `manager.go` — `State` | 세션 상태 머신 | `sync.Mutex` |
| T14 | `anchor.go` — `lastScene` + `refCrops` | 이미지 체인 | `sync.RWMutex` |
| T16 | `store.go` — Firestore 읽기/쓰기 | 메모리 캐시 | `sync.RWMutex` |
| T18 | `album.go` — `scenes` slice | 앨범 장면 목록 | `sync.Mutex` |

### 2.2 동기화 도구 선택 기준

```
읽기 비율 > 70% → sync.RWMutex
읽기/쓰기 비슷 → sync.Mutex
단일 카운터 → atomic.Int64
producer/consumer → buffered channel
```

### 2.3 핵심 패턴

#### 패턴 A: SessionManager 상태 머신 (Mutex — 쓰기 빈번)

```go
type Manager struct {
    mu    sync.Mutex
    state State
    // ...
}

// 상태 전환은 반드시 Lock 안에서
func (m *Manager) TransitionToReunion(ctx context.Context, persona *Persona) error {
    m.mu.Lock()
    defer m.mu.Unlock()

    if m.state != StateAnalyzing {
        return fmt.Errorf("invalid state transition: %s → Reunion", m.state)
    }

    // 복합 연산을 원자적으로 실행
    m.state = StateTransitioning
    if err := m.swapSession(ctx, persona); err != nil {
        m.state = StateAnalyzing // 롤백
        return err
    }
    m.state = StateReunion
    return nil
}
```

#### 패턴 B: CharacterAnchor (RWMutex — 읽기 빈번)

```go
type CharacterAnchor struct {
    mu        sync.RWMutex
    refCrops  []genai.Part  // 초기 크롭 (읽기 주)
    lastScene string        // 마지막 장면 (쓰기 주)
}

// 읽기: 여러 goroutine 동시 가능
func (a *CharacterAnchor) GetRefParts() []genai.Part {
    a.mu.RLock()
    defer a.mu.RUnlock()
    parts := make([]genai.Part, len(a.refCrops))
    copy(parts, a.refCrops)
    if a.lastScene != "" {
        parts = append(parts, genai.Blob{
            MIMEType: "image/png",
            Data:     decodeBase64(a.lastScene),
        })
    }
    return parts
}

// 쓰기: 배타적 접근
func (a *CharacterAnchor) UpdateLastScene(base64Img string) {
    a.mu.Lock()
    defer a.mu.Unlock()
    a.lastScene = base64Img
}
```

#### 패턴 C: Atomic — 이미지 생성 카운터 (hot path)

```go
type ImageStats struct {
    inFlight  atomic.Int64
    totalGen  atomic.Int64
    totalFail atomic.Int64
}

func (s *ImageStats) StartJob()  { s.inFlight.Add(1) }
func (s *ImageStats) EndJob()    { s.inFlight.Add(-1); s.totalGen.Add(1) }
func (s *ImageStats) FailJob()   { s.inFlight.Add(-1); s.totalFail.Add(1) }
func (s *ImageStats) Pending() int64 { return s.inFlight.Load() }
```

### 2.4 적용 규칙

- ✅ 공유 자원 접근 시 반드시 동기화 도구 사용
- ✅ `go test -race ./...` — **모든 PR에서 실행**
- ✅ Lock 범위 최소화: I/O (API 호출, DB 접근) 전에 Unlock
- ❌ Lock 안에서 채널 send 금지 (데드락 위험)
- ❌ RWMutex 안에서 쓰기 호출 금지 (재진입 데드락)

---

## 3. Goroutine Leak — 🟠 HIGH

### 3.1 발생 위치 (태스크별)

| 태스크 | 위치 | 리크 원인 |
|--------|------|-----------|
| T02 | `proxy.go` — read/write 루프 | ctx.Done() 미확인 |
| T03 | `tools.go` — 이미지 생성 goroutine | timeout 초과 후 방치 |
| T05 | `generator.go` — flash/pro 병렬 | cancel 안 된 goroutine |
| T10 | 온보딩 Live API 세션 | 세션 종료 시 cleanup 미수행 |
| T11 | `manager.go` — SwapSession | 이전 세션 goroutine 정리 |
| T19 | E2E 통합 | 비정상 종료 시 goroutine 잔류 |

### 3.2 방어 패턴

#### 패턴 A: 결과 채널은 항상 버퍼링

```go
// ❌ 위험: unbuffered 채널 → context 취소 시 goroutine 영원히 블록
resultCh := make(chan string)

// ✅ 안전: 버퍼 1 → goroutine이 결과를 넣고 종료 가능
resultCh := make(chan string, 1)

util.SafeGo(func() {
    result, err := generateImage(ctx, prompt)
    if err != nil {
        return
    }
    select {
    case resultCh <- result:
    case <-ctx.Done():
        // context 취소 시 결과 버리고 종료
    }
})
```

#### 패턴 B: 2단계 이미지 생성 (T05) — 완전한 cleanup

```go
func (h *ToolHandler) handleGenerateScene(ctx context.Context, args map[string]any) {
    // 부모 context에서 파생 — 세션 종료 시 자동 취소
    flashCtx, flashCancel := context.WithTimeout(ctx, 5*time.Second)
    defer flashCancel()

    proCtx, proCancel := context.WithTimeout(ctx, 20*time.Second)
    defer proCancel()

    flashDone := make(chan struct{}, 1)
    proDone := make(chan struct{}, 1)

    // Stage 1: Flash 프리뷰 (gemini-2.5-flash-image)
    util.SafeGo(func() {
        defer func() { flashDone <- struct{}{} }()
        img, err := h.gen.GenerateFlash(flashCtx, prompt, anchor.GetRefParts())
        if err != nil {
            slog.Warn("flash failed", "error", err)
            return
        }
        h.sendEvent(ctx, "scene_preview", img)
    })

    // Stage 2: Pro 고품질 (imagen-4.0-generate-001, Imagen 4)
    util.SafeGo(func() {
        defer func() { proDone <- struct{}{} }()
        // flash 완료 대기 (선택)
        select {
        case <-flashDone:
        case <-proCtx.Done():
            return
        }
        img, err := h.gen.GeneratePro(proCtx, prompt, anchor.GetRefParts())
        if err != nil {
            slog.Warn("pro failed", "error", err)
            return
        }
        h.sendEvent(ctx, "scene_final", img)
        anchor.UpdateLastScene(img)
    })

    // 부모가 취소되면 flashCancel + proCancel 모두 자동 호출 (defer)
}
```

#### 패턴 C: WaitGroup으로 goroutine 추적

```go
type Proxy struct {
    wg sync.WaitGroup
    // ...
}

func (p *Proxy) Start(ctx context.Context) {
    p.wg.Add(2)

    util.SafeGo(func() {
        defer p.wg.Done()
        p.forwardBrowserToLive(ctx)
    })

    util.SafeGo(func() {
        defer p.wg.Done()
        p.forwardLiveToBrowser(ctx)
    })
}

func (p *Proxy) Close() {
    // context cancel → 루프 종료 → wg.Done()
    done := make(chan struct{})
    go func() {
        p.wg.Wait()
        close(done)
    }()

    select {
    case <-done:
        slog.Info("proxy goroutines exited cleanly")
    case <-time.After(5 * time.Second):
        slog.Warn("proxy goroutine shutdown timeout")
    }
}
```

### 3.3 적용 규칙

- ✅ 결과 채널: 항상 `make(chan T, 1)` (최소 버퍼 1)
- ✅ 모든 goroutine에 `context.Context` 전달
- ✅ 루프 안에서 `select { case <-ctx.Done(): return }` 필수
- ✅ `sync.WaitGroup`으로 goroutine 수명 추적
- ✅ 테스트에서 `go.uber.org/goleak` 사용
- ❌ `time.Sleep()` 대신 `time.After()` + `select` 사용

---

## 4. Memory Leak — 🟠 HIGH

### 4.1 발생 위치 (태스크별)

| 태스크 | 위치 | 원인 |
|--------|------|------|
| T02 | WebSocket 메시지 버퍼 | 매 메시지마다 새 []byte 할당 |
| T05 | 이미지 base64 | pro-image 응답 ~2MB/장 누적 |
| T08 | YouTube 분석 결과 | 대용량 분석 텍스트 캐싱 |
| T14 | CharacterAnchor refCrops | 이미지 크롭 누적 |
| T16 | Firestore 메모리 캐시 | 무한 캐시 증가 |
| T18 | 앨범 장면 저장 | 장면 이미지 메모리 보유 |

### 4.2 방어 패턴

#### 패턴 A: sync.Pool — WebSocket 메시지 버퍼

```go
var bufPool = sync.Pool{
    New: func() any {
        buf := make([]byte, 0, 4096)
        return &buf
    },
}

func (p *Proxy) readMessage() ([]byte, error) {
    bufPtr := bufPool.Get().(*[]byte)
    buf := (*bufPtr)[:0]
    defer func() {
        // 64KB 초과 버퍼는 풀에 반환하지 않음 (거대 버퍼 누적 방지)
        if cap(buf) <= 64*1024 {
            *bufPtr = buf
            bufPool.Put(bufPtr)
        }
    }()

    _, msg, err := p.browserConn.ReadMessage()
    if err != nil {
        return nil, err
    }
    buf = append(buf, msg...)
    return buf, nil
}
```

#### 패턴 B: LRU 캐시 — Firestore 메모리 (크기 제한)

```go
type MemoryCache struct {
    mu      sync.RWMutex
    cache   map[string]*CacheEntry
    maxSize int
    order   []string // LRU 순서
}

func (mc *MemoryCache) Set(key string, value any) {
    mc.mu.Lock()
    defer mc.mu.Unlock()

    mc.cache[key] = &CacheEntry{Value: value, AccessedAt: time.Now()}
    mc.order = append(mc.order, key)

    // 최대 크기 초과 시 가장 오래된 항목 제거
    for len(mc.cache) > mc.maxSize {
        oldest := mc.order[0]
        mc.order = mc.order[1:]
        delete(mc.cache, oldest)
    }
}
```

#### 패턴 C: 이미지 참조 제한 (CharacterAnchor)

```go
const maxRefCrops = 5

func (a *CharacterAnchor) AddCrop(crop genai.Part) {
    a.mu.Lock()
    defer a.mu.Unlock()

    a.refCrops = append(a.refCrops, crop)
    // 최대 5개 유지 — FIFO
    if len(a.refCrops) > maxRefCrops {
        a.refCrops = a.refCrops[len(a.refCrops)-maxRefCrops:]
    }
}
```

#### 패턴 D: 앨범 장면은 Cloud Storage에 오프로드

```go
func (ag *AlbumGenerator) RecordScene(ctx context.Context, scene SceneData) error {
    ag.mu.Lock()
    defer ag.mu.Unlock()

    // 이미지를 Cloud Storage에 저장하고 URL만 보유
    url, err := ag.storage.Upload(ctx, scene.ImageBase64)
    if err != nil {
        return err
    }

    ag.scenes = append(ag.scenes, AlbumScene{
        ImageURL:    url,        // URL만 메모리에
        Caption:     scene.Caption,
        Timestamp:   time.Now(),
    })
    // scene.ImageBase64는 GC 대상이 됨

    return nil
}
```

### 4.3 적용 규칙

- ✅ hot path 버퍼: `sync.Pool` 사용, 64KB 초과 버퍼 풀 반환 안 함
- ✅ 캐시: 크기 상한 설정 (LRU 또는 TTL)
- ✅ 이미지: 메모리 보유 최소화 → Cloud Storage 오프로드
- ✅ 슬라이스: `append` 누적 시 크기 제한 (FIFO 드랍)
- ❌ `map`을 무제한 캐시로 사용 금지

---

## 5. Deadlock — 🔴 CRITICAL

### 5.1 발생 위치 (태스크별)

| 태스크 | 위치 | 시나리오 |
|--------|------|----------|
| T02 | proxy.go | browserMu Lock → liveSession Lock (순서 불일치) |
| T03 | tools.go | ToolHandler Lock → Generator Lock |
| T05 | generator.go | Anchor RLock → Generator 내부 Lock |
| T11 | manager.go | Manager Lock → Proxy Lock (세션 교체) |
| T18 | album.go | Album Lock 안에서 Storage 호출 (I/O 블로킹) |

### 5.2 방어 패턴

#### 패턴 A: Lock 순서 규칙 (필수)

```
=== missless.co Lock Ordering ===
Lock 순서는 항상 아래 순서를 따른다:

Level 1: Manager.mu        (세션 전체 상태)
Level 2: Proxy.mu          (Live API 세션 포인터)
Level 3: ToolHandler.mu    (Tool 맵)
Level 4: CharacterAnchor.mu (이미지 체인)
Level 5: AlbumGenerator.mu (장면 목록)
Level 6: MemoryStore.mu    (Firestore 캐시)

규칙: 상위 Level Lock을 잡은 상태에서 하위 Lock 획득 가능.
      하위 Lock을 잡은 상태에서 상위 Lock 획득 금지.
      동일 Level Lock 2개 동시 획득 금지.
```

#### 패턴 B: Lock 안에서 I/O 금지

```go
// ❌ 위험: Lock 안에서 외부 호출
func (ag *AlbumGenerator) badRecordScene(ctx context.Context, scene SceneData) {
    ag.mu.Lock()
    defer ag.mu.Unlock()
    url, _ := ag.storage.Upload(ctx, scene.ImageBase64) // ⚠️ I/O under lock
    ag.scenes = append(ag.scenes, AlbumScene{ImageURL: url})
}

// ✅ 안전: I/O를 Lock 밖에서 수행
func (ag *AlbumGenerator) goodRecordScene(ctx context.Context, scene SceneData) error {
    // Step 1: I/O 먼저 (Lock 없이)
    url, err := ag.storage.Upload(ctx, scene.ImageBase64)
    if err != nil {
        return err
    }

    // Step 2: 결과만 Lock 안에서 저장
    ag.mu.Lock()
    ag.scenes = append(ag.scenes, AlbumScene{ImageURL: url})
    ag.mu.Unlock()
    return nil
}
```

#### 패턴 C: 채널 데드락 방지 — select + timeout

```go
// ❌ 위험: unbuffered 채널에 send, receiver 없으면 영원히 블록
ch := make(chan int)
ch <- 1 // DEADLOCK

// ✅ 안전: select + context
func safeSend(ctx context.Context, ch chan<- int, val int) error {
    select {
    case ch <- val:
        return nil
    case <-ctx.Done():
        return ctx.Err()
    }
}
```

### 5.3 적용 규칙

- ✅ Lock 순서 문서화 (위 테이블) — 코드 주석에도 기재
- ✅ Lock 안에서 I/O, 채널 send, 외부 함수 호출 금지
- ✅ 채널 연산은 항상 `select` + `ctx.Done()` 가드
- ✅ 개발 단계에서 `-race` 플래그로 테스트
- ❌ 2개 이상 Lock 동시 획득 시 순서 미준수 금지

---

## 6. Bottleneck — 🟡 MEDIUM

### 6.1 발생 위치 (태스크별)

| 태스크 | 위치 | 병목 원인 |
|--------|------|-----------|
| T02 | Live API WebSocket | 단일 연결 대역폭 한계 |
| T05 | pro-image 생성 | 8~12초/장, 직렬 처리 시 대기 |
| T06 | Rate Limit | Gemini API RPM 제한 (Tier 1: 150 RPM) |
| T08 | YouTube URL 분석 | 영상당 10~30초, 직렬 시 3개 = 90초 |
| T16 | Firestore 읽기 | 개별 문서 get 반복 |
| T20 | Cloud Run cold start | min-instances=0일 때 첫 요청 10초+ |

### 6.2 방어 패턴

#### 패턴 A: HTTP 클라이언트 재사용 (전역 싱글턴)

```go
// internal/config/http.go
var sharedHTTPClient = &http.Client{
    Timeout: 30 * time.Second,
    Transport: &http.Transport{
        MaxIdleConns:        100,
        MaxIdleConnsPerHost: 10,
        MaxConnsPerHost:     20,
        IdleConnTimeout:     90 * time.Second,
    },
}

// 절대로 매 요청마다 http.Client{} 생성하지 않는다.
```

#### 패턴 B: 이미지 생성 워커 풀 (동시 실행 제한)

```go
type ImageGenPool struct {
    sem     chan struct{}   // semaphore
    stats   ImageStats
}

func NewImageGenPool(maxConcurrent int) *ImageGenPool {
    return &ImageGenPool{
        sem: make(chan struct{}, maxConcurrent), // 최대 3 동시 실행
    }
}

func (p *ImageGenPool) Generate(ctx context.Context, req ImageRequest) (string, error) {
    // semaphore 획득 (backpressure)
    select {
    case p.sem <- struct{}{}:
        defer func() { <-p.sem }()
    case <-ctx.Done():
        return "", fmt.Errorf("image queue full: %w", ctx.Err())
    }

    p.stats.StartJob()
    defer p.stats.EndJob()

    return callGeminiImageAPI(ctx, req)
}
```

#### 패턴 C: YouTube 분석 병렬화 + Rate Limit 준수

```go
func (p *Pipeline) analyzeVideos(ctx context.Context, videos []VideoInput) ([]*VideoAnalysis, error) {
    results := make([]*VideoAnalysis, len(videos))
    var wg sync.WaitGroup
    errCh := make(chan error, len(videos))

    // 최대 2개 동시 분석 (RPM 고려)
    sem := make(chan struct{}, 2)

    for i, v := range videos {
        wg.Add(1)
        i, v := i, v
        util.SafeGo(func() {
            defer wg.Done()

            select {
            case sem <- struct{}{}:
                defer func() { <-sem }()
            case <-ctx.Done():
                errCh <- ctx.Err()
                return
            }

            // gemini-2.5-pro 로 YouTube URL 직접 분석
            analysis, err := p.analyzer.AnalyzeYouTubeURL(ctx, v.YouTubeURL)
            if err != nil {
                slog.Warn("video analysis failed", "url", v.YouTubeURL, "error", err)
                return // 개별 실패 허용
            }
            results[i] = analysis
        })
    }

    wg.Wait()
    close(errCh)

    // 실패하지 않은 결과만 수집
    var valid []*VideoAnalysis
    for _, r := range results {
        if r != nil {
            valid = append(valid, r)
        }
    }
    return valid, nil
}
```

#### 패턴 D: Firestore 배치 읽기

```go
func (s *Store) SearchBatch(ctx context.Context, keywords []string) ([]Memory, error) {
    // 개별 Get 대신 Where + 배치
    query := s.client.Collection("memories").
        Where("keywords", "array-contains-any", keywords).
        Limit(10).
        OrderBy("relevance", firestore.Desc)

    docs, err := query.Documents(ctx).GetAll()
    if err != nil {
        return nil, err
    }

    memories := make([]Memory, 0, len(docs))
    for _, doc := range docs {
        var m Memory
        if err := doc.DataTo(&m); err != nil {
            continue
        }
        memories = append(memories, m)
    }
    return memories, nil
}
```

### 6.3 적용 규칙

- ✅ `http.Client` 전역 싱글턴 (절대 per-request 생성 안 함)
- ✅ 이미지 생성: semaphore로 동시 실행 제한 (최대 3)
- ✅ API 호출: `retry.WithBackoff` + 지수 백오프 + jitter
- ✅ Firestore: 배치 읽기/쓰기 활용
- ✅ Cloud Run: `min-instances=1` (cold start 방지)
- ❌ 무제한 goroutine 생성 금지

---

## 7. Context 전파 — 🟢 LOW~🟡 MEDIUM

### 7.1 규칙

```
모든 함수의 첫 번째 파라미터: ctx context.Context

요청 생명주기:
  HTTP Request → r.Context()
    → SessionManager.HandleSession(ctx, ...)
      → Proxy.Start(ctx)
        → forwardBrowserToLive(ctx)
        → forwardLiveToBrowser(ctx)
        → ToolHandler.Handle(ctx, ...)
          → Generator.GenerateFlash(ctx, ...)
          → Generator.GeneratePro(ctx, ...)
          → Store.Search(ctx, ...)

context.Background() 사용 허용 위치:
  1. main() 함수
  2. Graceful Shutdown 시 새 timeout context
  3. 그 외 절대 금지
```

### 7.2 Timeout 체인 설계

```
WebSocket 세션: 부모 context (무제한, cancel으로 종료)
  ├─ Tool 실행: 60초 timeout
  │   ├─ Flash 이미지: 5초 timeout
  │   ├─ Pro 이미지: 20초 timeout
  │   └─ Memory 검색: 5초 timeout
  ├─ YouTube 분석: 30초/영상 timeout
  └─ 세션 전환: 10초 timeout
```

```go
// ❌ 위험: context 체인 끊김
func badHandler(ctx context.Context) {
    // context.Background()는 부모 취소를 무시함
    newCtx := context.Background() // WRONG
    callAPI(newCtx)
}

// ✅ 올바름: 부모 context에서 파생
func goodHandler(ctx context.Context) {
    childCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
    defer cancel() // 반드시 defer
    callAPI(childCtx)
}
```

### 7.3 적용 규칙

- ✅ `context.WithTimeout` 직후 `defer cancel()` 필수
- ✅ `context.Background()` 사용 시 코드 주석으로 이유 명시
- ✅ HTTP/WebSocket 핸들러: `r.Context()` 전파
- ❌ 자식 timeout이 부모보다 길면 안 됨

---

## 8. Graceful Shutdown — 🟡 MEDIUM

### 8.1 종료 순서

```
SIGINT/SIGTERM 수신
    │
    ▼
Phase 1: HTTP 서버 Shutdown (20초)
    │   → 새 요청 거부, 기존 요청 완료 대기
    ▼
Phase 2: WebSocket 세션 종료 (30초)
    │   → Close frame 전송, goroutine 대기
    ▼
Phase 3: 이미지 생성 완료 대기 (20초)
    │   → 진행 중 작업 완료, 새 작업 거부
    ▼
Phase 4: Firestore 플러시 (10초)
    │   → 배치 쓰기 커밋
    ▼
Phase 5: 리소스 정리
    → genai.Client.Close()
    → Firestore Client.Close()
    → 로그 플러시
```

### 8.2 구현

```go
func main() {
    // signal.NotifyContext — Go 1.16+
    ctx, stop := signal.NotifyContext(context.Background(),
        syscall.SIGINT, syscall.SIGTERM)
    defer stop()

    server := NewServer(cfg)

    // 서버 시작
    go func() {
        if err := server.httpServer.ListenAndServe(); err != http.ErrServerClosed {
            slog.Error("HTTP server error", "error", err)
        }
    }()

    slog.Info("server started", "port", cfg.Port)

    // 종료 신호 대기
    <-ctx.Done()
    slog.Info("shutdown signal received")

    // Graceful shutdown (별도 context — 부모는 이미 취소됨)
    shutdownCtx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
    defer cancel()

    // Phase 1: HTTP
    if err := server.httpServer.Shutdown(shutdownCtx); err != nil {
        slog.Error("HTTP shutdown error", "error", err)
    }

    // Phase 2: WebSocket 세션
    server.sessionManager.Shutdown(shutdownCtx)

    // Phase 3: 이미지 생성
    server.imagePool.Shutdown(shutdownCtx)

    // Phase 4: Firestore
    server.memoryStore.Flush(shutdownCtx)

    // Phase 5: 클라이언트 정리
    server.genaiClient.Close()
    server.firestoreClient.Close()

    slog.Info("graceful shutdown complete")
}
```

### 8.3 Cloud Run 특이사항

```go
// Cloud Run은 SIGTERM 후 10초 내 종료해야 함
// --timeout 300 설정 시에도 graceful 기간은 10초
// 따라서 Phase 1~5 합계 < 10초로 제한

const cloudRunGracePeriod = 8 * time.Second // 2초 여유

func (s *Server) shutdownForCloudRun(ctx context.Context) {
    ctx, cancel := context.WithTimeout(ctx, cloudRunGracePeriod)
    defer cancel()

    // 병렬 shutdown (직렬로는 시간 부족)
    var wg sync.WaitGroup
    wg.Add(3)

    util.SafeGo(func() {
        defer wg.Done()
        s.httpServer.Shutdown(ctx)
    })
    util.SafeGo(func() {
        defer wg.Done()
        s.sessionManager.Shutdown(ctx)
    })
    util.SafeGo(func() {
        defer wg.Done()
        s.memoryStore.Flush(ctx)
    })

    wg.Wait()
    s.genaiClient.Close()
    s.firestoreClient.Close()
}
```

### 8.4 적용 규칙

- ✅ `signal.NotifyContext` 사용 (Go 1.16+, Go 1.22에서 안정)
- ✅ Shutdown 전용 context: `context.Background()` + timeout
- ✅ Cloud Run: 병렬 shutdown, 8초 이내
- ✅ WebSocket close frame 전송 후 연결 종료
- ❌ `os.Exit()` 직접 호출 금지

---

## 태스크별 위험요소 매트릭스

| 태스크 | Panic | Race | GoroutineLeak | MemoryLeak | Deadlock | Bottleneck |
|:------:|:-----:|:----:|:-------------:|:----------:|:--------:|:----------:|
| T01 | - | - | - | - | - | - |
| T02 | 🔴 | 🔴 | 🟠 | 🟡 | 🟡 | 🟡 |
| T03 | 🔴 | 🟡 | 🟠 | - | 🟡 | - |
| T04 | - | - | - | - | - | - |
| T05 | 🔴 | 🟡 | 🟠 | 🟠 | - | 🟡 |
| T06 | 🟡 | - | - | - | - | 🟠 |
| T07 | - | - | - | - | - | - |
| T08 | 🟡 | - | 🟡 | 🟡 | - | 🟠 |
| T09 | 🟠 | - | 🟡 | - | - | 🟡 |
| T10 | 🟡 | - | 🟡 | - | - | - |
| T11 | 🔴 | 🔴 | 🟠 | - | 🟠 | - |
| T12 | - | - | - | - | - | - |
| T13 | 🟡 | - | - | - | - | - |
| T14 | - | 🟡 | - | 🟡 | - | - |
| T15 | - | - | - | - | - | - |
| T16 | - | 🟡 | - | 🟡 | - | 🟡 |
| T17 | - | - | - | - | - | - |
| T18 | - | 🟡 | - | 🟠 | 🟡 | - |
| T19 | 🟡 | 🟡 | 🟡 | 🟡 | 🟡 | 🟡 |
| T20 | - | - | - | - | - | 🟡 |
| T21~T24 | - | - | - | - | - | - |

**범례**: 🔴 CRITICAL / 🟠 HIGH / 🟡 MEDIUM / - 해당없음

---

## 핵심 위험 태스크 TOP 5

| 순위 | 태스크 | 위험 요약 | 방어 핵심 |
|:---:|--------|-----------|-----------|
| 1 | **T02** (WebSocket 프록시) | Panic + Race + Leak 복합 | `safeGo`, `sync.Mutex`, WaitGroup |
| 2 | **T11** (세션 전환) | 상태 머신 Race + Deadlock | Lock 순서 L1→L2, I/O 밖에서 |
| 3 | **T05** (2단계 이미지) | goroutine 2개 병렬 + 메모리 | context.WithTimeout, 결과 채널 버퍼 |
| 4 | **T03** (Tool 실행) | Panic + Deadlock | recover, Lock 순서 L3→L4 |
| 5 | **T06** (Rate Limit) | API 병목 | 지수 백오프 + jitter + semaphore |

---

## 개발 도구 + 테스트 체크리스트

### 필수 도구

```bash
# Race detector (모든 테스트에 적용)
go test -race ./...

# Goroutine leak 감지
go get go.uber.org/goleak
# 테스트 파일에: defer goleak.VerifyNone(t)

# 정적 분석
go install honnef.co/go/tools/cmd/staticcheck@latest
staticcheck ./...

# 메모리 프로파일링
go tool pprof http://localhost:8080/debug/pprof/heap
go tool pprof http://localhost:8080/debug/pprof/goroutine

# Deadlock 감지 (개발 단계)
go get github.com/sasha-s/go-deadlock
# sync.Mutex → deadlock.Mutex로 교체 (개발 시에만)
```

### CI 파이프라인 체크리스트

- [ ] `go vet ./...` 통과
- [ ] `staticcheck ./...` 통과
- [ ] `go test -race -count=1 ./...` 통과
- [ ] `goleak` 테스트 통과
- [ ] `go build -o /dev/null ./...` 경고 없음

### 프로덕션 모니터링

```go
// internal/handler/debug.go
import _ "net/http/pprof"

func RegisterDebugHandlers(mux *http.ServeMux) {
    // /debug/pprof/* 자동 등록
    // Cloud Run에서는 인증 미들웨어로 보호
    mux.HandleFunc("/debug/goroutines", func(w http.ResponseWriter, r *http.Request) {
        fmt.Fprintf(w, "goroutines: %d\n", runtime.NumGoroutine())
    })
    mux.HandleFunc("/debug/memstats", func(w http.ResponseWriter, r *http.Request) {
        var m runtime.MemStats
        runtime.ReadMemStats(&m)
        json.NewEncoder(w).Encode(map[string]any{
            "alloc_mb":       m.Alloc / 1024 / 1024,
            "sys_mb":         m.Sys / 1024 / 1024,
            "num_gc":         m.NumGC,
            "goroutines":     runtime.NumGoroutine(),
        })
    })
}
```

---

## 요약: 방어 패턴 Quick Reference

| 패턴 | 적용 위치 | 코드 |
|------|-----------|------|
| `util.SafeGo()` | 모든 goroutine | `util.SafeGo(func() { ... })` |
| `sync.Mutex` | 상태 머신, 세션 포인터 | `mu.Lock(); defer mu.Unlock()` |
| `sync.RWMutex` | 읽기 주, 이미지 체인 | `mu.RLock(); defer mu.RUnlock()` |
| `atomic.Int64` | 카운터 (hot path) | `counter.Add(1)` |
| `context.WithTimeout` | 모든 외부 호출 | `ctx, cancel := ...; defer cancel()` |
| `make(chan T, 1)` | goroutine 결과 반환 | 버퍼 1 최소 |
| `select + ctx.Done()` | 모든 채널 연산 | `select { case <-ctx.Done(): return }` |
| `sync.WaitGroup` | goroutine 수명 추적 | `wg.Add(1); defer wg.Done()` |
| `sync.Pool` | 메시지 버퍼 | 64KB 초과 풀 반환 안 함 |
| Lock 순서 L1→L6 | 다중 Lock 획득 | Manager→Proxy→Tool→Anchor→Album→Store |
| `retry.WithBackoff` | API 호출 | 지수 백오프 + jitter |
| semaphore | 동시 실행 제한 | `make(chan struct{}, max)` |
