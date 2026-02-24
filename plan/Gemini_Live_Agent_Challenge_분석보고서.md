# Gemini Live Agent Challenge 2026 - 종합 분석 보고서

> 작성일: 2026-02-23 | 마감일까지 **21일** 남음

---

## 1. 챌린지 개요

| 항목 | 내용 |
|------|------|
| **챌린지명** | Gemini Live Agent Challenge |
| **주최** | Google (Google AI Developers) via DevPost |
| **형태** | 온라인 해커톤, 공개 참가 |
| **총 상금** | **$80,000** (약 1억 1천만원) |
| **마감일** | **2026년 3월 16일 오후 5:00 PM PDT** (한국시간 3/17 오전 9시) |
| **수상 발표** | 2026년 4월 22일 ~ 4월 24일 (Google Cloud Next 2026 기간) |
| **참가자 수** | 526+ (2월 기준) |
| **해시태그** | #GeminiLiveAgentChallenge |
| **공식 페이지** | https://geminiliveagentchallenge.devpost.com/ |

### 핵심 미션
> "텍스트 박스를 넘어서라." — 단순 텍스트 입출력을 넘어, **보고(See), 듣고(Hear), 말하고(Speak), 만드는(Create)** 차세대 멀티모달 AI 에이전트를 구축하라.

---

## 2. 상금 구조 (총 $80,000)

### Grand Prize — $25,000
- Google Cloud 크레딧 $3,000
- Google 팀과 가상 커피 미팅
- 소셜 미디어 프로모션
- Google Cloud Next 2026 티켓 2장 (각 $2,299 가치, 라스베이거스)
- 여행 경비 2인분 (최대 $3,000/인)
- **Google Cloud Next 2026에서 데모 시연 기회**

### Best Live Agent — $10,000
- Google Cloud 크레딧 $1,000
- Google 팀과 가상 커피 미팅
- 소셜 미디어 프로모션
- Google Cloud Next 2026 티켓 2장

### Best Creative Storyteller — $10,000
- Google Cloud 크레딧 $1,000
- Google 팀과 가상 커피 미팅
- 소셜 미디어 프로모션
- Google Cloud Next 2026 티켓 2장

### Best UI Navigator — $10,000
- Google Cloud 크레딧 $1,000
- Google 팀과 가상 커피 미팅
- 소셜 미디어 프로모션
- Google Cloud Next 2026 티켓 2장

> **참고**: 나머지 $25,000은 추가 카테고리 또는 러너업 상금으로 추정됨.

---

## 3. 프로젝트 카테고리 (3개 트랙)

### 트랙 1: Live Agent (실시간 상호작용)
**핵심**: Multimodal Live API를 활용한 실시간 음성/비전 에이전트

사용자가 자연스럽게 대화하고, 중간에 끼어들 수 있으며, 실시간으로 반응하는 에이전트를 구축합니다.

**예시 프로젝트:**
- 실시간 번역기
- 숙제를 "보는" 비전 기반 개인교사
- 고객 지원 음성 에이전트
- 실시간 의료 상담 보조

### 트랙 2: Creative Storyteller (크리에이티브 스토리텔러)
**핵심**: Gemini의 네이티브 인터리브드(interleaved) 출력을 활용한 크리에이티브 에이전트

텍스트, 이미지, 오디오, 비디오를 하나의 유기적 흐름으로 엮어내는 "크리에이티브 디렉터"처럼 사고하고 창작하는 에이전트입니다.

**예시 프로젝트:**
- 인터랙티브 스토리북
- 마케팅 에셋 생성기
- 교육 콘텐츠 설명 에이전트
- 소셜 콘텐츠 크리에이터

### 트랙 3: UI Navigator (UI 내비게이터)
**핵심**: Gemini AI 멀티모달로 스크린샷/화면 녹화를 해석하여 사용자 대신 화면을 조작하는 에이전트

브라우저나 디바이스 화면을 관찰하고, 시각적 요소를 해석하며, API나 DOM 접근 없이 사용자 의도에 따라 액션을 수행합니다.

**예시 프로젝트:**
- 범용 웹 내비게이터
- 크로스 앱 워크플로우 자동화기
- 비주얼 QA 테스팅 에이전트

---

## 4. 필수 기술 요구사항

### 반드시 사용해야 하는 기술 스택

| 구분 | 요구사항 |
|------|----------|
| **AI 모델** | Gemini 모델 (필수) |
| **SDK** | Google GenAI SDK 또는 ADK (Agent Development Kit) |
| **클라우드** | 최소 1개 이상의 Google Cloud 서비스 |
| **배포** | Google Cloud에서 백엔드 호스팅 (배포 증거 필요) |

### UI Navigator 트랙 추가 요구사항
- Gemini AI 멀티모달을 사용하여 스크린샷/화면 녹화 해석
- Google Cloud에서 호스팅

---

## 5. 제출 요구사항

### 필수 제출물
1. **아키텍처 다이어그램** — Gemini가 백엔드, 데이터베이스, 프론트엔드와 어떻게 연결되는지 명확하게 표시
2. **데모 영상** (4분 미만) — 다음을 포함:
   - 멀티모달/에이전틱 기능의 **실제 동작** (목업 불가)
   - 프로젝트 피치
3. **Google Cloud 배포 증거** — 백엔드가 Google Cloud에서 실행 중임을 증명
4. **소스 코드** — DevPost 제출 시 포함

---

## 6. 심사 기준

### Innovation & Multimodal User Experience — 40%
- "텍스트 박스" 패러다임을 탈피했는가?
- 에이전트가 유동적으로 "보고, 듣고, 말하는가"?
- 독특한 성격/목소리를 가지고 있는가?
- "실시간(Live)"이고 컨텍스트를 인지하는가?

### Technical Implementation & Agent Architecture — 30%
- Google GenAI SDK 또는 ADK를 효과적으로 사용했는가?
- 백엔드가 Google Cloud에서 견고하게 호스팅되는가?
- 에이전트가 할루시네이션을 방지하는가?
- 그라운딩(Grounding) 증거가 있는가?

### Demo & Presentation — 30%
- 영상이 문제와 솔루션을 명확히 정의하는가?
- 아키텍처 다이어그램이 명확한가?
- 클라우드 배포의 시각적 증거가 있는가?

---

## 7. 핵심 기술 스택 상세

### 7.1 Gemini Multimodal Live API

실시간, 저지연 양방향 음성/비디오 상호작용을 가능하게 하는 API입니다.

**주요 특징:**
- Voice Activity Detection (VAD) — 사용자 발화 감지 및 중단 처리
- Tool Use & Function Calling — 실시간 도구 호출
- Session Management — 장시간 대화 관리
- Ephemeral Tokens — 보안 클라이언트 인증
- 동시 세션: 50~1,000개 (티어에 따라)

**오디오 스펙:**
- 입력: 16-bit PCM, 모노, 16kHz (50~100ms 청크 = 1,600~3,200 바이트)
- 출력: 16-bit PCM, 모노, 24kHz

**공식 문서:** https://ai.google.dev/gemini-api/docs/live

### 7.2 Agent Development Kit (ADK)

Google이 Google Cloud NEXT 2025에서 발표한 오픈소스 에이전트 개발 프레임워크입니다.

**핵심 컴포넌트:**
- **LiveRequestQueue** — 수신 메시지 버퍼링 및 시퀀싱
- **Runner** — 세션 라이프사이클 및 대화 상태 관리
- **LLM Flow** — 복잡한 프로토콜 변환 처리

**주요 기능:**
- Bidi-streaming (양방향 스트리밍) — 실시간 음성/비디오 통신
- 자동 도구 실행 (병렬)
- WebSocket 타임아웃 시 투명한 재연결
- 데이터베이스 세션 영속화
- 멀티 에이전트 워크플로우

**ADK 공식 문서:** https://google.github.io/adk-docs/

### 7.3 Google GenAI SDK
Gemini 모델과 상호작용하기 위한 공식 SDK. Python, JavaScript/TypeScript 등 지원.

---

## 8. 학습 리소스 및 참고 자료

### 공식 가이드 & 코드랩

| 리소스 | 설명 | 링크 |
|--------|------|------|
| ADK Streaming Quickstart | 음성/비디오 스트리밍 빠른 시작 | https://google.github.io/adk-docs/get-started/streaming/quickstart-streaming/ |
| ADK Bidi-Streaming 개발 가이드 (5부작) | 아키텍처부터 프로덕션까지 | https://google.github.io/adk-docs/streaming/ |
| Way Back Home 코드랩 | React 프론트엔드 + Python ADK 백엔드 풀스택 | https://codelabs.developers.google.com/way-back-home-level-3 |
| Personal Expense Assistant | Gemini 2.5 + Firestore + Cloud Run | https://codelabs.developers.google.com/personal-expense-assistant-multimodal-adk |
| 실시간 음성 에이전트 | Gemini + ADK + Google Maps MCP | https://cloud.google.com/blog/products/ai-machine-learning/build-a-real-time-voice-agent-with-gemini-adk |

### ADK Bidi-Streaming 개발 가이드 구성

1. **Part 1**: 스트리밍 기초 — Live API 기술, ADK 아키텍처, FastAPI 예제
2. **Part 2**: LiveRequestQueue로 메시지 전송 — 텍스트/오디오/비디오, 동시성 패턴
3. **Part 3**: `run_live()` 이벤트 처리 — 자동 도구 실행, 멀티 에이전트 워크플로우
4. **Part 4**: RunConfig — 응답 모달리티, 세션 관리, 컨텍스트 윈도우 압축
5. **Part 5**: 오디오/이미지/비디오 사용법 — VAD, 프로액티브 대화

### 레퍼런스 구현
- **bidi-demo**: FastAPI 기반 프로덕션 레디 참조 구현 (`adk-samples/python/agents/bidi-demo`)
  - WebSocket 통신, 멀티모달 요청, 자동 트랜스크립션, Google Search 연동

### GitHub
- https://github.com/google-gemini — 공식 리포지토리

---

## 9. 관련 해커톤 및 우수 사례 분석

### Gemini 3 Hackathon (2025.12 ~ 2026.02)
- **주최**: Google DeepMind
- **상금**: $100,000
- **심사 기준**: Gemini 통합 40%, 혁신 30%, 실제 영향 20%
- **카테고리**: SOTA Reasoning, Agentic Coding, GenMedia

**싱가포르 해커톤 프로젝트 아이디어:**
- The Legal Eagle — 법정 비디오 영상 및 증거 파일 분석
- The Diagnostic Detective — 환자 이력 및 의료 이미지 분석
- Legacy Lifter — COBOL/Python 2.7을 Go/Rust로 리팩토링
- Dynamic Storyteller — Veo 생성 비디오 기반 인터랙티브 교육

### Gemini API Developer Competition (2024)
**수상작 예시:**
- **Omni** — OS에 통합된 AI 앱
- **AI Shift** — 자동 근무 스케줄링
- **Menu Buddy** — 언어장애인 주문 보조

### Google Cloud AI Hackathon (2025.12)
**수상작:**
- **SurgAgent** — 복강경 비디오에서 수술 도구 추적 AI

### GKE Hackathon
**수상작:**
- **Amie Wei의 Cart-to-Kitchen AI** — 장바구니에서 주방까지 AI 어시스턴트 (대상)
- **V-Commerce Studio** — AI 개인화 채팅, 가상 피팅, 광고 생성

---

## 10. 전략적 권장사항

### 차별화 전략

**1. 트랙 선택 고려사항:**

| 트랙 | 경쟁 강도(예상) | 기술 난이도 | 차별화 가능성 |
|------|----------------|------------|--------------|
| Live Agent | 높음 | 중~고 | 도메인 특화로 차별화 |
| Creative Storyteller | 중간 | 중 | 창의적 UX로 차별화 |
| UI Navigator | 낮음~중간 | 고 | 기술 구현 자체로 차별화 |

**2. 심사 기준 기반 우선순위:**
- **혁신 & 멀티모달 UX (40%)** 가 가장 높은 비중 → "텍스트 박스 탈피"를 극대화
- **기술 구현 (30%)** → ADK 활용, 할루시네이션 방지, 그라운딩 필수
- **데모 & 프레젠테이션 (30%)** → 4분 미만 영상 퀄리티가 매우 중요

**3. 필수 체크리스트:**
- [ ] Gemini 모델 사용
- [ ] Google GenAI SDK 또는 ADK 사용
- [ ] 최소 1개 Google Cloud 서비스 사용
- [ ] Google Cloud 백엔드 배포 및 증거
- [ ] 아키텍처 다이어그램 작성
- [ ] 4분 미만 데모 영상 제작
- [ ] 실제 동작하는 멀티모달/에이전틱 기능 시연

**4. 기술 스택 추천 (Golang + Next.js 기반):**
- 백엔드: Go + Google Cloud Run / GKE
- 프론트엔드: Next.js + WebSocket (ADK bidi-streaming 연동)
- AI: Gemini Live API via Google GenAI SDK
- 인프라: Google Cloud (Cloud Run, Firestore, Cloud Storage 등)
- ADK Python 에이전트를 별도 서비스로 운영하고, Go 백엔드에서 gRPC/REST로 통신

---

## 11. 타임라인

| 날짜 | 이벤트 |
|------|--------|
| 2026-02-23 | 현재 (D-21) |
| 2026-03-16 17:00 PDT | **제출 마감** |
| 2026-04-22 ~ 04-24 | 수상자 발표 (Google Cloud Next 2026) |

---

*이 보고서는 웹 검색을 통해 수집된 정보를 기반으로 작성되었습니다. 공식 규칙 및 최신 정보는 반드시 DevPost 공식 페이지에서 확인하시기 바랍니다.*
