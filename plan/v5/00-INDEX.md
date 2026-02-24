# missless.co V5 — 구현 계획서 (Phase 분할)

> Gemini Live Agent Challenge 2026 | Creative Storyteller 트랙
> 작성일: 2026-02-24 | 마감일: 2026-03-16 (D-20)
> GitHub: https://github.com/Two-Weeks-Team/missless (public)
> V4 → V5 핵심 변경: Ephemeral Token 폐기 → Go 백엔드 프록시 아키텍처

---

## V4 → V5 핵심 아키텍처 변경

| 항목 | V4 | V5 | 이유 |
|------|----|----|------|
| Live API 연결 | 브라우저 직접 (Ephemeral Token) | **Go 백엔드 프록시** | Tool 실행 보안 + 아키텍처 단순화 |
| Tool 실행 | 브라우저가 수신 → 서버에 전달 | **Go 내부에서 직접 실행** | 왕복 레이턴시 제거 + API 키 보안 |
| 이미지 생성 | flash-image 단독 | **2단계 Progressive (flash→pro)** | Creative Storyteller 품질 확보 |
| 세션 전환 | 미정 | **Go SessionManager (2세션)** | 음성 변경 필수 → 세션 분리 |
| 프론트 역할 | Live API 직접 연결 + Tool 중계 | **순수 렌더러 (오디오/이미지 수신만)** | 복잡성 제거 |
| 캐릭터 일관성 | 미정 | **레퍼런스 체이닝 + 실루엣 전략** | pro-image 일관성 한계 대응 |
| 음성 매칭 | 미정 | **30개 HD 프리셋 음성 매핑** | VideoAnalyzer 분석 → VoiceMatcher 매핑 |
| 프론트 서빙 | 미정 | **Next.js static export + Go FileServer** | 동일 오리진, CORS 불필요 |
| 세션 스토어 | 미정 | **Firestore 기반** | OAuth 토큰, 페르소나, 대화 이력 영속화 |
| 재회 시간 | 미정 | **300초(5분) 제한 + 계속하기** | 이전 세션 요약 주입으로 연속성 유지 |
| BGM | 미정 | **Gemini Lyria RealTime 생성** | AI 생성 BGM (프리셋 파일 불필요) |

---

## 전체 시스템 아키텍처 (V5)

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
│  │    - 시스템 음성 (missless 호스트)                     │  │
│  │    - 사용자 안내 + YouTube 영상 선택                   │  │
│  │                                                       │  │
│  │  Phase 2: 재회 Live API 세션                           │  │
│  │    - 페르소나 HD 음성                                  │  │
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
│  │  change_atmosphere → BGM 이벤트 전송                   │ │
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

## 문서 구성

| 파일 | 내용 | 태스크 |
|------|------|--------|
| **`00-INDEX.md`** | 본 문서. 전체 구조, 아키텍처, Phase 개요 | — |
| **`01-PHASE1-INFRA-LIVE.md`** | 인프라 + Live API 기반 + WebSocket 프록시 | T01~T06 |
| **`02-PHASE2-ONBOARDING.md`** | YouTube 분석 + 페르소나 생성 + 온보딩 UX | T07~T12 |
| **`03-PHASE3-REUNION.md`** | 재회 경험 엔진 + 이미지/BGM/앨범 | T13~T19 |
| **`04-PHASE4-DEPLOY-DEMO.md`** | 배포 + 데모 영상 + DevPost 제출 | T20~T24 |
| **`05-GO-SAFETY.md`** | Go 위험요소 분류 (Panic/Race/Bottleneck/Leak) | 전체 |

---

## 태스크 총괄 (24개)

### Phase 1: 인프라 + Live API 기반 (D-20 ~ D-15)

| ID | 태스크 | 난이도 | 일수 | 체크포인트 |
|----|--------|:---:|:---:|------------|
| T01 | GCP 프로젝트 + API 키 + Go scaffolding | 하 | 0.5 | 5개 모델 API 호출 성공 |
| T02 | Go WebSocket 프록시 + Live API 연결 | 상 | 1.5 | 브라우저→Go→Live API 음성 양방향 성공 |
| T03 | Tool 등록 + 서버사이드 실행 기반 | 중 | 1 | Live API Tool Call → Go 핸들러 실행 성공 |
| T04 | Next.js PWA 순수 렌더러 | 중 | 1 | 오디오 재생 + 이미지 표시 + AudioContext |
| T05 | 2단계 이미지 생성 (flash→pro Progressive) | 상 | 1 | preview 1-3초 표시 → final 크로스페이드 |
| T06 | Rate Limit 방어 + 에러 핸들링 기반 | 중 | 0.5 | retryWithBackoff + tool_error 파이프라인 |

### Phase 2: 온보딩 파이프라인 (D-14 ~ D-10)

| ID | 태스크 | 난이도 | 일수 | 체크포인트 |
|----|--------|:---:|:---:|------------|
| T07 | Google OAuth + YouTube 영상 목록 + 공개상태 | 중 | 1 | 로그인→영상 목록+✅🔒 표시 |
| T08 | YouTube URL 직접 분석 + 갤러리 Fallback | 상 | 1.5 | URL→Gemini 분석 성공 + Fallback 동작 |
| T09 | Sequential Agent 온보딩 (VideoAnalyzer→VoiceMatcher) | 중 | 1 | 영상분석→30개 HD 프리셋 음성 매핑 완료 |
| T10 | 온보딩 Live API 세션 (시스템 음성) | 중 | 1 | AI 음성 안내 → YouTube 선택 유도 |
| T11 | SessionManager — 온보딩→재회 전환 | 상 | 1 | 세션 교체 (음성 변경) 1-2초 이내 |
| T12 | 온보딩 UX (진행률 + 하이라이트 + 인물 선택) | 중 | 1 | 실시간 카드 업데이트 + 인물 크롭 선택 |

### Phase 3: 재회 경험 엔진 (D-9 ~ D-5)

| ID | 태스크 | 난이도 | 일수 | 체크포인트 |
|----|--------|:---:|:---:|------------|
| T13 | 재회 Live API 세션 + 페르소나 System Instruction | 중 | 1 | 페르소나 성격으로 대화 시작 |
| T14 | 캐릭터 일관성 (CharacterAnchor + 실루엣) | 중 | 1 | 연속 3장면 동일 캐릭터 유지 |
| T15 | BGM 시스템 + change_atmosphere | 하 | 0.5 | 분위기 전환 시 BGM 자동 변경 |
| T16 | recall_memory + Firestore 추억 검색 | 중 | 0.5 | 관련 추억 검색 → 대화 반영 |
| T17 | Affective Dialog + Proactive Audio 최적화 | 중 | 0.5 | 감정 인지 대화 + 선택적 응답 |
| T18 | 앨범 생성 + 공유 카드 | 중 | 1 | 장면 저장 → 앨범 페이지 → OG 카드 |
| T19 | E2E 통합 테스트 | 상 | 1 | 온보딩→재회→앨범 전체 완주 |

### Phase 4: 배포 + 데모 + 제출 (D-4 ~ D-0)

| ID | 태스크 | 난이도 | 일수 | 체크포인트 |
|----|--------|:---:|:---:|------------|
| T20 | Cloud Run 배포 + 도메인 + SSL | 중 | 0.5 | missless.co 라이브 접속 |
| T21 | 데모 시나리오 준비 (자체 영상 + 캐싱) | 중 | 1 | YouTube 영상 + Firestore 캐시 |
| T22 | 데모 영상 촬영 (3:50) | 상 | 1.5 | 실제 감정 반응 포함 완성본 |
| T23 | README + 아키텍처 다이어그램 | 하 | 0.5 | GitHub 공개 + spin-up 가이드 |
| T24 | DevPost 제출 | 하 | 0.5 | ✅ 제출 완료 |

---

## 스프린트 일정

| 일자 | 요일 | Phase | 태스크 | 마일스톤 |
|------|:---:|:---:|--------|----------|
| 2/25 | 화 | P1 | T01 | GCP + Go 초기화 |
| 2/26 | 수 | P1 | T02 | WebSocket 프록시 PoC |
| 2/27 | 목 | P1 | T02+T03 | Live API 양방향 + Tool 실행 |
| 2/28 | 금 | P1 | T04+T06 | PWA + 에러 핸들링 |
| 3/1 | 토 | P1 | T05 | 2단계 이미지 Progressive |
| 3/2 | 일 | P1 | T05 | **P1 완성 — Fluid Stream 동작** |
| 3/3 | 월 | P2 | T07 | OAuth + YouTube 그리드 |
| 3/4 | 화 | P2 | T08 | YouTube URL 직접 분석 PoC |
| 3/5 | 수 | P2 | T09 | Sequential Agent (VideoAnalyzer→VoiceMatcher) |
| 3/6 | 목 | P2 | T10+T11 | 온보딩 세션 + SessionManager |
| 3/7 | 금 | P2 | T12 | 온보딩 UX 완성 |
| 3/8 | 토 | P2 | (버퍼) | **P2 완성 — 온보딩→재회 전환** |
| 3/9 | 일 | P3 | T13+T14 | 재회 세션 + 캐릭터 일관성 |
| 3/10 | 월 | P3 | T15+T16 | BGM + 추억 검색 |
| 3/11 | 화 | P3 | T17+T18 | Affective Dialog + 앨범 |
| 3/12 | 수 | P3 | T19 | **P3 완성 — E2E 통합 성공** |
| 3/13 | 목 | P4 | T20+T21 | Cloud Run 배포 + 데모 준비 |
| 3/14 | 금 | P4 | T22 | 데모 영상 촬영 |
| 3/15 | 토 | P4 | T22+T23 | 영상 편집 + README |
| 3/16 | 일 | P4 | T24 | **✅ 최종 제출 (KST 3/17 09:00)** |

---

## 채점 기준 대응 매트릭스

| 채점 항목 (비중) | V5 대응 | Phase |
|-----------------|---------|:---:|
| 텍스트 박스 탈피 (40%) | 전체 경험 텍스트 입력 ZERO. 온보딩도 음성. | P2 |
| See/Hear/Speak 유동성 (40%) | 이미지+음성+BGM Progressive Rendering | P1,P3 |
| 독특한 성격/목소리 (40%) | YouTube 분석 → 30개 HD 프리셋 음성 매핑 (VideoAnalyzer→VoiceMatcher) | P2 |
| Live & context-aware (40%) | Affective Dialog + Proactive Audio + Vision | P3 |
| GenAI SDK/ADK 활용 (30%) | genai v1.29+ / ADK v0.4.0 / 5모델 + URL분석 | P1,P2 |
| Google Cloud 배포 (30%) | Cloud Run + Firestore + Storage + Terraform | P4 |
| 올 Google 생태계 (30%) | OAuth + YouTube API + Gemini + Cloud Run | P2 |
| 할루시네이션 방지 (30%) | YouTube URL 분석 기반 그라운딩 + recall_memory | P2,P3 |
| 문제-솔루션 명확 (30%) | 15초 감성 인트로 + 즉시 이해 | P4 |
| 실제 동작 (30%) | 자체 영상 + 실제 감정 반응 시연 | P4 |
