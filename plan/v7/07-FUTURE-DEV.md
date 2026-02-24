# 추후개발 항목 — 제출본 완성 후 재검토

> 이 문서에 기록된 항목들은 제출본 완성 후, 제출 직전에 재판단합니다.
> 시간적 여유가 있을 경우에만 구현하며, 핵심 기능에 영향을 주지 않습니다.

---

## 1. 보이스클로닝 (MiniMax / ElevenLabs)

### 현재 상태
- V7에서는 **30개 HD 프리셋 음성 매핑** 방식으로 구현
- VideoAnalyzer가 추출한 음성 메타데이터(성별, 나이대, 톤, 속도 등)를 기반으로 최적 프리셋 선택
- 프리셋 방식은 안정적이나, 실제 인물 목소리와의 유사도에 한계 있음

### 보이스클로닝 옵션

#### Option A: MiniMax Voice Cloning API
- **장점**: 짧은 오디오 샘플(5초~)로 클로닝 가능, 한국어 지원
- **단점**: 외부 API 의존, 실시간 Live API와 통합 복잡
- **통합 방식**:
  1. 온보딩 시 YouTube 영상에서 음성 추출
  2. MiniMax API로 Voice ID 생성
  3. 재회 세션에서 생성된 Voice ID로 TTS 생성
  4. TTS 오디오를 Live API 대신 프론트에 전송
- **문제점**: Live API의 실시간 음성 생성을 대체하면 대화 자연스러움 저하
- **비용**: Pay-as-you-go

#### Option B: ElevenLabs Voice Cloning
- **장점**: 업계 최고 품질, Professional Voice Cloning (11초~)
- **단점**: 비용 높음, 외부 API 의존, 실시간 대화 지원 한계
- **통합 방식**: MiniMax와 유사
- **비용**: $5/월(Starter) ~ $22/월(Creator)

#### Option C: Gemini 자체 보이스 클로닝 (미출시)
- Google이 향후 Gemini Live API에 커스텀 음성 기능을 추가할 가능성
- ADK 로드맵에서 관련 기능 언급 없음 (2026-02 기준)

### 재판단 기준
- [ ] 핵심 기능(온보딩→재회→앨범) E2E 통합 테스트 통과 후
- [ ] 제출 마감 3일 전(D-3) 시점에 여유 시간 확인
- [ ] 30개 프리셋 음성이 데모에서 충분히 자연스러운지 평가
- [ ] 외부 API 추가로 인한 아키텍처 복잡성 vs. 품질 향상 트레이드오프

### 결론
> **V7 기본 전략: 30개 HD 프리셋 음성으로 제출.**
> 프리셋으로도 채점 기준의 "독특한 성격/목소리" 항목을 충분히 충족 가능.
> 보이스클로닝은 제출 후 데모 개선 또는 Grand Prize 타겟 시 재검토.

---

## 2. 다국어 지원 확장

### 현재 상태
- V6는 **한국어(ko-KR)** 기본
- 시스템 프롬프트, UI 텍스트 모두 한국어

### 향후 확장
- 30개 HD 프리셋 음성은 다국어 지원 (영어, 일본어, 중국어 등)
- `SpeechConfig.LanguageCode` 변경으로 언어 전환 가능
- UI 국제화(i18n) 추가 필요

### 재판단 기준
- [ ] 데모 영상이 영어 심사위원 대상이므로, 영어 데모 필요 여부
- [ ] 데모 시나리오에서 영어 대화 포함 여부 결정

---

## 3. 벡터 검색 기반 추억 검색

### 현재 상태
- V6는 Firestore **키워드 매칭** 방식 (전문 검색 미지원)
- 간단하지만 정확도 한계

### 향후 확장
- Vertex AI Vector Search 또는 Firestore Vector Search (GA)
- 임베딩 생성 → 의미론적 검색으로 정확도 향상
- `text-embedding-004` 모델 사용

### 재판단 기준
- [ ] 키워드 매칭으로 데모에서 충분한 품질인지 확인
- [ ] 추억 검색 빈도가 낮으면 우선순위 하락

---

## 4. 멀티 인물 지원

### 현재 상태
- V6는 **1명의 인물**만 분석/재회 지원
- 영상에서 1명 선택 → 1개 페르소나 생성

### 향후 확장
- 복수 인물 동시 분석 → 여러 페르소나 생성
- "가족 모임" 시나리오: 여러 페르소나가 번갈아 대화
- SessionManager에 복수 페르소나 관리 기능 추가

---

## 5. PWA 오프라인 + 앨범 로컬 저장

### 현재 상태
- V6는 완전 온라인 의존
- 앨범은 Cloud Storage URL 기반

### 향후 확장
- Service Worker로 앨범 이미지 캐싱
- IndexedDB로 대화 기록 로컬 저장
- 오프라인에서 앨범 열람 가능

---

## 6. 감정 분석 대시보드

### 현재 상태
- Affective Dialog는 Live API가 내부적으로 처리
- 감정 데이터를 별도로 저장/시각화하지 않음

### 향후 확장
- 재회 중 감정 변화 타임라인
- 앨범에 "감정 곡선" 포함
- 상담/치유 목적의 감정 리포트

---

## 7. 보너스 점수 관련 콘텐츠 (Optional Developer Contributions, 최대 +1.0점)

### 블로그 포스트 (+0.6점) — dev.to 4편 시리즈 전략

- **전략**: 단일 포스트 대신 **Phase별 4편 시리즈**로 작성 → 더 풍부한 콘텐츠 + 심사위원 어필
- **플랫폼**: **dev.to** (공개 플랫폼, 개발자 커뮤니티 최적)
- **언어**: 영문
- **필수 조건** (모든 포스트에 포함):
  1. **반드시 포함 문구**: "This article was created for purposes of entering this hackathon"
  2. 해시태그: `#GeminiLiveAgentChallenge` `#GeminiAPI` `#GoogleCloud`
  3. 공개적으로 접근 가능해야 함
- **추가 콘텐츠**: 소셜 미디어(Twitter/X, LinkedIn)에 각 포스트 공유 + `#GeminiLiveAgentChallenge`
- **복수 제출 가능**: DevPost에 모든 시리즈 URL 제출

#### 시리즈 전체 구성

| # | 제목 | 작성 시점 | 내용 | 대응 Phase |
|---|------|-----------|------|:---:|
| 1 | **Building a Real-Time Audio Proxy with Go + Gemini Live API** | 3/2 (P1 완성 후) | WebSocket 프록시, Live API 연결, Progressive Rendering | P1 |
| 2 | **YouTube Video Analysis → AI Persona with Sequential Agents** | 3/8 (P2 완성 후) | YouTube URL 분석, Sequential Agent, 30 HD Voice 매핑 | P2 |
| 3 | **Creating Immersive Reunion Experiences with Interleaved Output** | 3/12 (P3 완성 후) | 재회 엔진, Affective Dialog, BGM, 앨범, interleaved output | P3 |
| 4 | **missless.co — Complete Architecture & Lessons Learned** | 3/14-15 (제출 D-2) | 전체 아키텍처, 5모델 통합, 도전과 교훈, 결과 (메인 포스트) | 종합 |

#### Post #1: Real-Time Audio Proxy (3/2 작성)

```markdown
# Building a Real-Time Audio Proxy with Go + Gemini Live API

> This article was created for purposes of entering the
> Gemini Live Agent Challenge hackathon. #GeminiLiveAgentChallenge

## Why Go for a WebSocket Proxy?
[Go 선택 이유 — 동시성, 성능, Google Cloud 네이티브]

## Connecting to Gemini Live API
[Live API 세션 관리, 30 HD 프리셋 음성]

## Progressive Image Rendering
[Flash preview (1-3s) → Pro final (8-12s) crossfade 전략]

## Rate Limiting & Error Handling
[retryWithBackoff, semaphore, 3중 방어 체계]

## What's Next
[시리즈 #2 예고 — YouTube 분석과 Agent Pipeline]
```

#### Post #2: YouTube → AI Persona (3/8 작성)

```markdown
# YouTube Video Analysis → AI Persona with Sequential Agents

> This article was created for purposes of entering the
> Gemini Live Agent Challenge hackathon. #GeminiLiveAgentChallenge

## The Challenge: Extracting Personality from Video
[YouTube URL 직접 분석, 비디오 다운로드 없이 분석하는 전략]

## Building a Sequential Agent Pipeline
[VideoAnalyzer → VoiceMatcher, ADK v0.5.0]

## Matching 30 HD Preset Voices
[메타데이터 기반 음성 매핑 알고리즘]

## Session Management: Onboarding → Reunion
[SessionManager, 2초 이내 세션 교체]

## What's Next
[시리즈 #3 예고 — 재회 경험과 Interleaved Output]
```

#### Post #3: Immersive Reunion (3/12 작성)

```markdown
# Creating Immersive Reunion Experiences with Interleaved Output

> This article was created for purposes of entering the
> Gemini Live Agent Challenge hackathon. #GeminiLiveAgentChallenge

## The Reunion Engine
[페르소나 System Instruction, Affective Dialog, recall_memory]

## Interleaved Output: Voice + Image + BGM Simultaneously
[Live API 음성 + Tool Call 이미지/BGM → interleaved stream]
[**핵심**: Gemini's interleaved output capabilities + Tool orchestration]

## Lyria RealTime BGM
[감정 적응형 배경음악, change_atmosphere Tool]

## Album Generation: Preserving Digital Memories
[장면 → 앨범 → OG 공유 카드]

## What's Next
[시리즈 #4 예고 — 전체 아키텍처와 교훈]
```

#### Post #4: Complete Architecture (3/14-15 작성, 메인 포스트)

```markdown
# missless.co — Complete Architecture & Lessons Learned

> This article was created for purposes of entering the
> Gemini Live Agent Challenge hackathon. #GeminiLiveAgentChallenge

## The Problem: "When was the last time you talked to them?"
[그리움 문제 정의, 보편적 감정]

## Architecture Overview
[Go 프록시 → 5 Gemini Models → Cloud Services 다이어그램]

## 5 Gemini Models Working Together
[각 모델 역할과 통합 전략]

## Google Cloud Integration
[Cloud Run + Firestore + Storage + OAuth + YouTube API]

## Challenges & Key Learnings
[개발 과정의 도전과 해결, Go 동시성 교훈]

## Results & Demo
[데모 링크, 사용자 반응, 감성적 가치]

## Series Index
- [Part 1: Real-Time Audio Proxy](link)
- [Part 2: YouTube → AI Persona](link)
- [Part 3: Interleaved Reunion](link)
- **Part 4: Complete Architecture** (this post)
```

#### dev.to 작성 체크리스트 (Phase별)

**Post #1 (3/2, P1 완성 후)**:
- [ ] WebSocket 프록시 코드 스니펫 포함
- [ ] Progressive Rendering 시퀀스 다이어그램
- [ ] "created for purposes of entering this hackathon" 문구
- [ ] `#GeminiLiveAgentChallenge` `#GeminiAPI` `#Go` 태그
- [ ] Twitter/LinkedIn 공유

**Post #2 (3/8, P2 완성 후)**:
- [ ] Sequential Agent 파이프라인 다이어그램
- [ ] 30개 HD 프리셋 음성 매핑 로직 설명
- [ ] "created for purposes of entering this hackathon" 문구
- [ ] `#GeminiLiveAgentChallenge` `#GeminiAPI` `#ADK` 태그
- [ ] Twitter/LinkedIn 공유 + Post #1 링크

**Post #3 (3/12, P3 완성 후)**:
- [ ] interleaved output 아키텍처 다이어그램 (**핵심 포스트**)
- [ ] Affective Dialog + BGM 시퀀스 예시
- [ ] "created for purposes of entering this hackathon" 문구
- [ ] `#GeminiLiveAgentChallenge` `#GeminiAPI` `#CreativeStoryteller` 태그
- [ ] Twitter/LinkedIn 공유 + 시리즈 링크

**Post #4 (3/14-15, D-2 메인)**:
- [ ] 전체 아키텍처 다이어그램 (5모델 + Cloud)
- [ ] 시리즈 인덱스 링크 포함
- [ ] "created for purposes of entering this hackathon" 문구
- [ ] `#GeminiLiveAgentChallenge` `#GeminiAPI` `#GoogleCloud` 태그
- [ ] Twitter/LinkedIn 공유 + 시리즈 전체 링크
- [ ] **DevPost 제출 시 4개 URL 모두 등록**

### 자동 배포 스크립트 (+0.2점)

- **요구사항**: Infrastructure-as-code 또는 배포 자동화 스크립트가 **공개 repo에 포함**
- **구현 계획**:
  1. `cloudbuild.yaml` — Cloud Build 파이프라인 정의
  2. `terraform/main.tf` — Cloud Run + Firestore + Storage 인프라 정의
  3. `Makefile` 또는 `deploy.sh` — 원클릭 배포 스크립트
- **현재 상태**: `gcloud run deploy` 수동 명령 → Phase 4에서 IaC로 전환

#### cloudbuild.yaml 예시

```yaml
steps:
  - name: 'gcr.io/cloud-builders/docker'
    args: ['build', '-t', 'gcr.io/$PROJECT_ID/missless', '.']
  - name: 'gcr.io/cloud-builders/docker'
    args: ['push', 'gcr.io/$PROJECT_ID/missless']
  - name: 'gcr.io/google.com/cloudsdktool/cloud-sdk'
    args:
      - 'run'
      - 'deploy'
      - 'missless'
      - '--image=gcr.io/$PROJECT_ID/missless'
      - '--region=asia-northeast3'
      - '--allow-unauthenticated'
      - '--min-instances=1'
      - '--memory=512Mi'
```

### GDG 멤버십 (+0.2점)

- Google Developer Groups 가입 후 **공개** 프로필 링크 제출
- 가입: https://developers.google.com/community/gdg
- 간단한 작업이므로 **즉시 완료 가능** (PRE 단계에서 완료 권장)
- 모든 팀원이 가입하면 더 좋음 (프로필 링크 제출)

---

## 우선순위 요약

### 보너스 점수 (제출 전 반드시 완료)

| 순위 | 항목 | 점수 | 난이도 | 시기 | 비고 |
|:---:|------|:---:|:---:|------|------|
| 1 | **GDG 멤버십** | +0.2 | 없음 | **즉시 (PRE 단계)** | 5분이면 완료, 안 할 이유 없음 |
| 2 | **자동 배포 (IaC)** | +0.2 | 하 | Phase 4 (T20) | cloudbuild.yaml + Terraform |
| 3 | **dev.to 시리즈 (4편)** | +0.6 | 중 | 각 Phase 완성 후 | dev.to 시리즈 전략, 영문 필수 |

### 기능 확장 (제출 후 재검토)

| 순위 | 항목 | 영향도 | 난이도 | 시기 |
|:---:|------|:---:|:---:|------|
| 4 | 영어 데모 시나리오 | 높음 | 중 | Phase 4 |
| 5 | 보이스클로닝 | 중간 | 상 | 제출 후 재검토 |
| 6 | 벡터 검색 | 낮음 | 중 | 제출 후 |
| 7 | 다국어 지원 | 낮음 | 중 | 제출 후 |
| 8 | 멀티 인물 | 낮음 | 상 | 제출 후 |
