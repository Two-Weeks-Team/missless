# 사전준비 — 개발 시작 전 필수 완료 (PRE-01~PRE-12)

> ⚠️ 이 문서의 모든 항목을 완료해야 Phase 1 개발을 시작할 수 있습니다.
> 예상 소요시간: 약 3~4시간 (순차 진행 시)
> 담당: 프로젝트 리드 (수동 작업)

---

## PRE-01. GCP 프로젝트 생성 + Billing 연결

**소요시간**: 15분

### 작업 내용

1. [Google Cloud Console](https://console.cloud.google.com/) 접속
2. 새 프로젝트 생성
   - 프로젝트 이름: `missless-2026`
   - 프로젝트 ID: `missless-2026` (또는 자동 생성)
   - 조직: 없음 (개인 계정)
3. Billing 계정 연결
   - 새 계정 생성 시 $300 무료 크레딧 자동 부여 (90일)
   - 기존 계정이면 크레딧 잔여 확인
4. 프로젝트 ID 기록

### 확인 항목

```
GCP_PROJECT_ID=missless-2026   # ← 실제 프로젝트 ID 기록
```

### 체크포인트

- [ ] GCP 프로젝트 생성 완료
- [ ] Billing 연결 확인
- [ ] $300 크레딧 또는 잔여 크레딧 확인
- [ ] 프로젝트 ID 메모

---

## PRE-02. Gemini API 활성화 + API 키 발급

**소요시간**: 30분

### 작업 내용

#### A. Vertex AI API 활성화 (권장 — Cloud Run 배포 시)

1. GCP Console → APIs & Services → Library
2. 검색: "Vertex AI API" → **Enable**
3. 검색: "Generative Language API" → **Enable** (Gemini API)

#### B. API 키 발급 (Developer API 방식 — 개발 단계)

1. [Google AI Studio](https://aistudio.google.com/) 접속
2. Get API key → Create API key in new project (또는 기존 프로젝트 선택)
3. API 키 복사 → 안전하게 보관

#### C. Tier 1 승격 (중요!)

> Gemini API 무료 티어는 RPM 한도가 매우 낮음 (15 RPM).
> Tier 1은 150-300 RPM으로 실제 사용에 적합.

1. GCP Console → APIs & Services → Quotas
2. Gemini API 관련 할당량 확인
3. **Pay-as-you-go Billing이 연결되면 자동으로 Tier 1** (대부분의 경우)
4. 할당량 증가가 필요하면: Quotas → Edit Quotas → 요청

#### D. 사용 모델 확인

아래 5개 모델이 모두 호출 가능한지 검증:

| 모델 | ID (Developer API) | ID (Vertex AI) | 용도 |
|------|-------------------|----------------|------|
| Live API | `gemini-2.5-flash-native-audio-preview-12-2025` | `gemini-live-2.5-flash-native-audio` | 음성 대화 |
| Pro (분석) | `gemini-2.5-pro-preview-03-25` | 동일 | 비디오/페르소나 분석 |
| Flash Image | `gemini-2.5-flash-preview-image-generation` | 동일 | 빠른 이미지 생성 |
| Imagen 3 | `imagen-3.0-generate-002` | 동일 | 고품질 이미지 |
| Lyria | `models/lyria-realtime-exp` | 동일 | BGM 생성 (실험적) |

### 확인 항목

```
GEMINI_API_KEY=AIza...           # Developer API 키
# 또는 Vertex AI 방식:
GOOGLE_APPLICATION_CREDENTIALS=./service-account.json
```

### 체크포인트

- [ ] Vertex AI API 활성화 확인
- [ ] Generative Language API 활성화 확인
- [ ] API 키 발급 완료
- [ ] Tier 1 할당량 확인 (RPM 150+)
- [ ] 5개 모델 API 호출 성공 (아래 PRE-12에서 테스트)

---

## PRE-03. YouTube Data API v3 활성화

**소요시간**: 10분

### 작업 내용

1. GCP Console → APIs & Services → Library
2. 검색: "YouTube Data API v3" → **Enable**
3. API 키는 PRE-02에서 발급한 것 공유 가능 (또는 별도 발급)

### 할당량 확인

- 기본 할당량: 10,000 유닛/일 (충분)
- channels.list: 1 유닛/호출
- playlistItems.list: 1 유닛/호출
- videos.list: 1 유닛/호출

### 체크포인트

- [ ] YouTube Data API v3 활성화 확인
- [ ] 할당량 10,000 유닛/일 확인

---

## PRE-04. OAuth 2.0 클라이언트 ID 생성

**소요시간**: 20분

### 작업 내용

1. GCP Console → APIs & Services → Credentials
2. **+ CREATE CREDENTIALS** → OAuth client ID
3. Application type: **Web application**
4. Name: `missless-web`
5. Authorized JavaScript origins:
   ```
   http://localhost:8080
   https://missless.co
   ```
6. Authorized redirect URIs:
   ```
   http://localhost:8080/auth/callback
   https://missless.co/auth/callback
   ```
7. **CREATE** → Client ID + Client Secret 복사

### OAuth 동의 화면 설정

1. APIs & Services → OAuth consent screen
2. User Type: **External**
3. App name: `missless`
4. User support email: 본인 이메일
5. Scopes 추가:
   - `https://www.googleapis.com/auth/youtube.readonly` (YouTube 영상 목록)
6. Test users: 본인 이메일 추가 (앱이 "Testing" 상태일 때 필요)

> ⚠️ **주의**: 앱이 "Testing" 상태면 등록된 테스트 사용자만 로그인 가능.
> 데모 시연 시에는 데모 계정을 테스트 사용자로 미리 등록하거나,
> 심사 시 "Published" 상태로 전환 필요 (Google 심사 최대 6주 소요 — 해커톤에는 불필요할 수 있음).

### 확인 항목

```
YOUTUBE_CLIENT_ID=xxxx.apps.googleusercontent.com
YOUTUBE_CLIENT_SECRET=GOCSPX-xxxx
OAUTH_REDIRECT_URL=http://localhost:8080/auth/callback
```

### 체크포인트

- [ ] OAuth 클라이언트 ID 생성 완료
- [ ] Client ID + Secret 확보
- [ ] Redirect URI 등록 (localhost + production)
- [ ] OAuth 동의 화면 설정 완료
- [ ] youtube.readonly 스코프 추가
- [ ] 테스트 사용자 등록

---

## PRE-05. Cloud Firestore 데이터베이스 생성

**소요시간**: 10분

### 작업 내용

1. GCP Console → Firestore
2. **CREATE DATABASE**
3. 모드: **Native mode** (실시간 쿼리 지원)
4. 위치: `asia-northeast3` (서울) — Cloud Run과 동일 리전
5. 보안 규칙: 기본값 유지 (서버 사이드 접근만 하므로 Admin SDK 사용)

### 컬렉션 구조 (사전 설계)

```
sessions/
  {sessionId}/
    id: string
    oauthToken: { accessToken, refreshToken, expiry }
    persona: { name, personality, speechStyle, matchedVoice, ... }
    state: "onboarding" | "analyzing" | "transitioning" | "reunion" | "album" | "ended"
    transcripts: [ { role, text, timestamp } ]
    reunionCount: number
    lastSummary: string
    createdAt: timestamp
    updatedAt: timestamp

personas/
  {personaId}/
    memories/
      {memoryId}/
        topic: string
        description: string
        timestamp: string
        source: "video_analysis" | "user_input"

albums/
  {albumId}/
    personaId: string
    summary: string
    duration: string
    imageUrls: [string]
    createdAt: timestamp
```

### 확인 항목

```
FIRESTORE_DB=(default)     # 기본 데이터베이스 사용
GCP_PROJECT_ID=missless-2026  # Firestore는 프로젝트 ID로 접근
```

### 체크포인트

- [ ] Firestore Native 모드 데이터베이스 생성 완료
- [ ] 리전: asia-northeast3 (서울) 확인
- [ ] Firestore 콘솔에서 데이터 추가/조회 테스트

---

## PRE-06. Cloud Storage 버킷 생성

**소요시간**: 10분

### 작업 내용

1. GCP Console → Cloud Storage → Buckets
2. **CREATE BUCKET**
3. 이름: `missless-assets` (또는 `missless-2026-assets`)
4. 리전: `asia-northeast3` (서울)
5. Storage class: **Standard**
6. Access control: **Uniform** (IAM 기반)
7. Public access: **일부 공개** (앨범 이미지 공유용)

### 버킷 내 디렉터리 구조

```
missless-assets/
  ├── uploads/          # 갤러리 Fallback 업로드 (임시)
  ├── scenes/           # 장면 이미지 (앨범용)
  ├── bgm/              # Lyria BGM 캐시
  └── albums/           # 앨범 공유 이미지
```

### 공개 접근 설정 (앨범 공유용)

```bash
# albums/ 디렉터리만 공개 접근 허용
gsutil iam ch allUsers:objectViewer gs://missless-assets/albums/
```

### 확인 항목

```
STORAGE_BUCKET=missless-assets
```

### 체크포인트

- [ ] Cloud Storage 버킷 생성 완료
- [ ] 리전: asia-northeast3 확인
- [ ] 파일 업로드/다운로드 테스트
- [ ] albums/ 공개 접근 설정 (배포 후)

---

## PRE-07. 도메인 missless.co DNS 설정

**소요시간**: 30분

### 작업 내용

> 도메인이 이미 구매되어 있다고 가정. 미구매 시 도메인 등록 필요.

1. 도메인 레지스트라 DNS 설정 페이지 접속
2. Cloud Run 커스텀 도메인 매핑 준비:
   ```bash
   gcloud run domain-mappings create \
     --service missless \
     --domain missless.co \
     --region asia-northeast3
   ```
3. 위 명령이 반환하는 CNAME/A 레코드를 DNS에 등록:
   - Type: `CNAME`
   - Name: `@` 또는 비워둠
   - Value: `ghs.googlehosted.com.` (Cloud Run에서 제공하는 값)
4. DNS 전파 대기 (최대 48시간, 보통 1-2시간)

> ⚠️ **배포 전(T20)에 DNS만 미리 설정하면 됩니다.**
> Cloud Run 서비스가 없으면 domain-mapping은 실패하므로,
> 이 단계에서는 **도메인 소유 확인 + DNS 레코드 준비**만 합니다.

### 확인 항목

```
DOMAIN=missless.co
```

### 체크포인트

- [ ] 도메인 missless.co 소유 확인
- [ ] DNS 설정 방법 확인 (레지스트라 접속 가능)
- [ ] Cloud Run 매핑을 위한 CNAME 준비 (T20에서 실행)

---

## PRE-08. GitHub 레포지토리 생성

**소요시간**: 5분

### 작업 내용

1. GitHub에서 새 레포지토리 생성
   - Organization: `Two-Weeks-Team`
   - Name: `missless`
   - Visibility: **Public** (필수 — DevPost 제출 시)
   - Initialize: README 포함
2. 로컬에 클론

> ⚠️ **프로젝트는 2026-02-16 이후 생성되어야 함** (챌린지 규칙).
> Git 히스토리의 첫 커밋이 2/16 이후인지 확인.

### 체크포인트

- [ ] GitHub 레포 생성 (public)
- [ ] 로컬 클론 완료
- [ ] 첫 커밋 날짜가 2026-02-16 이후 확인

---

## PRE-09. 로컬 개발환경 설치

**소요시간**: 20분

### 필수 도구

| 도구 | 최소 버전 | 설치 확인 |
|------|----------|-----------|
| Go | 1.22+ | `go version` |
| Node.js | 20+ | `node --version` |
| npm | 10+ | `npm --version` |
| gcloud CLI | 최신 | `gcloud --version` |
| git | 최신 | `git --version` |

### 설치 명령 (macOS)

```bash
# Go
brew install go

# Node.js
brew install node@20

# gcloud CLI
brew install --cask google-cloud-sdk

# gcloud 로그인 + 프로젝트 설정
gcloud auth login
gcloud config set project missless-2026
gcloud auth application-default login
```

### 체크포인트

- [ ] Go 1.22+ 설치 확인
- [ ] Node.js 20+ 설치 확인
- [ ] gcloud CLI 설치 + 로그인 완료
- [ ] `gcloud config set project` 완료

---

## PRE-10. .env.example 작성 + 환경변수 설정

**소요시간**: 15분

### .env.example (레포에 커밋)

```env
# ===== GCP =====
GCP_PROJECT_ID=your-project-id
GOOGLE_APPLICATION_CREDENTIALS=./service-account.json

# ===== Gemini API =====
GEMINI_API_KEY=your-api-key       # Developer API 방식
# 또는 Vertex AI 방식은 서비스 계정으로 인증

# ===== OAuth 2.0 =====
YOUTUBE_CLIENT_ID=your-client-id.apps.googleusercontent.com
YOUTUBE_CLIENT_SECRET=GOCSPX-your-secret
OAUTH_REDIRECT_URL=http://localhost:8080/auth/callback

# ===== Cloud Storage =====
STORAGE_BUCKET=your-bucket-name

# ===== Firestore =====
FIRESTORE_DB=(default)

# ===== Server =====
PORT=8080
DOMAIN=localhost

# ===== Session =====
SESSION_SECRET=your-random-secret-32chars-minimum
```

### .env (로컬 전용, .gitignore에 추가)

```env
GCP_PROJECT_ID=missless-2026
GEMINI_API_KEY=AIza...실제키
YOUTUBE_CLIENT_ID=실제ID.apps.googleusercontent.com
YOUTUBE_CLIENT_SECRET=GOCSPX-실제시크릿
OAUTH_REDIRECT_URL=http://localhost:8080/auth/callback
STORAGE_BUCKET=missless-assets
FIRESTORE_DB=(default)
PORT=8080
DOMAIN=localhost
SESSION_SECRET=랜덤32자이상문자열
```

### .gitignore 필수 항목

```
.env
service-account.json
*.key
```

### 체크포인트

- [ ] .env.example 작성 완료
- [ ] .env 로컬 설정 완료 (모든 값 입력)
- [ ] .gitignore에 민감 파일 추가
- [ ] `SESSION_SECRET` 32자 이상 랜덤 문자열 생성

---

## PRE-11. 서비스 계정 키 생성 (Vertex AI용)

**소요시간**: 15분

### 작업 내용

> Vertex AI 방식으로 Gemini API를 호출하려면 서비스 계정이 필요합니다.
> Developer API(API 키) 방식을 사용해도 되지만, Cloud Run 배포 시에는 서비스 계정이 기본입니다.

1. GCP Console → IAM & Admin → Service Accounts
2. **+ CREATE SERVICE ACCOUNT**
3. Name: `missless-backend`
4. Role 추가:
   - `Vertex AI User` (Gemini API 호출)
   - `Cloud Firestore User` (Firestore 읽기/쓰기)
   - `Storage Object Admin` (Cloud Storage 읽기/쓰기/삭제)
5. Keys → ADD KEY → Create new key → JSON → 다운로드

> ⚠️ 이 JSON 키 파일은 절대 Git에 커밋하면 안 됩니다.

### Cloud Run에서는?

Cloud Run은 기본 서비스 계정을 사용합니다:
- `missless-backend@missless-2026.iam.gserviceaccount.com`
- 환경변수 대신 Cloud Run 서비스에 이 계정을 연결하면 됨
- 로컬 개발 시에만 JSON 키 파일 사용

### 확인 항목

```
GOOGLE_APPLICATION_CREDENTIALS=./service-account.json   # 로컬 전용
# Cloud Run에서는 불필요 (자동 인증)
```

### 체크포인트

- [ ] 서비스 계정 생성 완료
- [ ] Vertex AI User 역할 부여
- [ ] Firestore User 역할 부여
- [ ] Storage Object Admin 역할 부여
- [ ] JSON 키 파일 다운로드 (로컬 보관)
- [ ] .gitignore에 `service-account.json` 추가 확인

---

## PRE-12. 모든 API 호출 검증 테스트

**소요시간**: 30분

### 검증 스크립트 (Go — 프로젝트 초기화 후 실행)

> 이 검증은 T01(Go scaffolding) 완료 직후에 실행합니다.
> PRE 단계에서는 **AI Studio 웹 UI** 또는 **curl**로 간단히 확인합니다.

#### A. AI Studio에서 모델 확인

1. [AI Studio](https://aistudio.google.com/) 접속
2. 새 프롬프트 → 모델 선택에서 아래 모델들이 표시되는지 확인:
   - `Gemini 2.5 Pro` (비디오 분석용)
   - `Gemini 2.5 Flash` (이미지 생성용)
3. Live API 탭 → `Native Audio` 모델 확인
   - 30개 음성 프리셋 목록 확인 가능

#### B. curl로 API 키 검증

```bash
# Gemini API 기본 테스트
curl -s "https://generativelanguage.googleapis.com/v1beta/models?key=$GEMINI_API_KEY" \
  | jq '.models[] | .name' | head -20

# 특정 모델 테스트 (텍스트 생성)
curl -s -X POST \
  "https://generativelanguage.googleapis.com/v1beta/models/gemini-2.5-flash:generateContent?key=$GEMINI_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"contents":[{"parts":[{"text":"Hello, world!"}]}]}' \
  | jq '.candidates[0].content.parts[0].text'
```

#### C. YouTube Data API 검증

```bash
# YouTube 채널 정보 테스트 (API 키 방식)
curl -s "https://www.googleapis.com/youtube/v3/channels?part=snippet&mine=true" \
  -H "Authorization: Bearer $OAUTH_ACCESS_TOKEN" \
  | jq '.items[0].snippet.title'
```

#### D. Firestore 검증

```bash
# gcloud CLI로 Firestore 연결 테스트
gcloud firestore databases describe --project=missless-2026
```

#### E. Cloud Storage 검증

```bash
# 버킷 접근 테스트
gsutil ls gs://missless-assets/
```

### 체크포인트 (최종)

- [ ] Gemini API 키로 모델 목록 조회 성공
- [ ] Gemini 2.5 Flash 텍스트 생성 성공
- [ ] Gemini 2.5 Pro 텍스트 생성 성공
- [ ] AI Studio에서 Live API 모델 확인
- [ ] AI Studio에서 30개 음성 프리셋 확인
- [ ] YouTube Data API 활성화 상태 확인
- [ ] Firestore 연결 확인
- [ ] Cloud Storage 버킷 접근 확인
- [ ] 서비스 계정 인증 동작 확인

---

## 사전준비 완료 상태 요약

모든 PRE 항목 완료 후 아래 표를 채워 확인합니다:

| 항목 | 값 | 상태 |
|------|:---|:---:|
| GCP Project ID | `missless-2026` | ⬜ |
| Gemini API Key | `AIza...` | ⬜ |
| Tier 1 RPM | `150+` | ⬜ |
| OAuth Client ID | `xxxx.apps.googleusercontent.com` | ⬜ |
| OAuth Client Secret | `GOCSPX-xxxx` | ⬜ |
| Firestore DB | `(default)` @ asia-northeast3 | ⬜ |
| Storage Bucket | `missless-assets` | ⬜ |
| Domain | `missless.co` (DNS 준비) | ⬜ |
| GitHub Repo | `Two-Weeks-Team/missless` (public) | ⬜ |
| Go | 1.22+ | ⬜ |
| Node.js | 20+ | ⬜ |
| gcloud CLI | 로그인 + 프로젝트 설정 | ⬜ |
| Service Account | JSON 키 파일 | ⬜ |
| .env | 모든 값 설정 | ⬜ |
| API 검증 | 5개 모델 + YouTube + Firestore + Storage | ⬜ |
| GDG 멤버십 | 가입 + 공개 프로필 URL | ⬜ |
| GCP 크레딧 | 신청 완료 (마감 3/13 12PM PT) | ⬜ |

> **✅ 모든 항목이 체크되면 Phase 1 (T01) 개발을 시작합니다.**

---

## 보너스 사전준비 (즉시 완료 가능)

### GDG 멤버십 가입 (+0.2점)

**소요시간**: 5분 | **난이도**: 없음

1. https://developers.google.com/community/gdg 접속
2. Google 계정으로 가입
3. 공개 프로필 URL 확보 → 제출 시 기입
4. 모든 팀원이 각각 가입 권장

### GCP 크레딧 신청

**마감**: 2026-03-13 12:00 PM PT (KST 3/14 05:00)

1. DevPost 리소스 페이지에서 신청 폼 확인
2. GCP 프로젝트 ID 기입
3. 마감 전 반드시 완료 (늦으면 크레딧 없음)
