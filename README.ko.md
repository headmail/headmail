# Headmail

English | [한국어](README.ko.md)

Headmail은 가벼운 Vue.js 관리자 프런트엔드를 갖춘 오픈소스 이메일 캠페인 관리 백엔드입니다. 캠페인 생성, 구독자/리스트 관리, 템플릿화 및 개인화된 전송, 이벤트 추적(오픈/클릭) 기능을 제공합니다. 백엔드는 Go로 작성되었으며 프런트엔드는 Vue 3 + Vite로 구성되어 있습니다.

이름 "Headmail"은 "head"와 "mailer"의 합성어입니다. 여기서 "head"는 전체 사용자용 UI를 포함한 완전한 시스템이 아닌 API 중심의 '머리'를 의미합니다. 즉, Headmail은 API 우선의 "머리만 있는" 메일러로서 캠페인 생성 및 전송을 위한 API 백엔드를 제공하지만 내장된 계정/사용자 관리를 포함하지 않습니다. 인증, 멀티테넌시 또는 계정 관련 기능이 필요하면 외부 시스템에서 처리하도록 설계되어 있습니다. 또한 시스템은 확장성을 염두에 두고 설계되었으며, 저장소 백엔드, 메일러/리시버 드라이버, 템플릿 시스템 등이 잘 정의된 인터페이스를 통해 플러그인 방식으로 교체·확장 가능합니다.

## 하이라이트

- API, 서비스 로직, 저장소의 명확한 분리 — 비즈니스 로직을 변경하지 않고도 다른 데이터베이스 백엔드를 구현할 수 있습니다.
- 메일 전송/수신 추상화 — SMTP나 써드파티 API 등 다양한 전송 구현을 쉽게 플러그인할 수 있습니다.
- 데이터 병합을 포함한 템플릿 렌더링과 간단한 i18n 헬퍼를 통한 지역화 메시지 지원.
- 점진적 확장과 테스트에 적합한 단순하고 모듈화된 코드 레이아웃.

## 빠른 개요

핵심 구성 요소:
- `cmd/server` — 서버 진입점
- `pkg/api` — HTTP 핸들러 (관리자 및 공용/추적 엔드포인트)
- `pkg/service` — 비즈니스 로직 (캠페인, 리스트, 전송, 추적)
- `pkg/repository` — 리포지토리 인터페이스 및 트랜잭션 유틸리티
- `internal/db/sqlite` — SQLite 리포지토리 구현
- `pkg/template` — 템플릿 렌더링 서비스 (Go 템플릿 + sprig 사용)
- `pkg/mailer`, `pkg/receiver` — 메일 전송 및 수신 추상화
- `frontend` — Vue 3 관리자 UI (Vite)

## 개념 및 데이터 모델 (간단)

- Campaign(캠페인): 제목, HTML/텍스트 본문, 선택적 템플릿 참조, 데이터 맵 및 헤더를 가진 템플릿화된 메시지.
- Template(템플릿): MJML 형식 템플릿 및 GrapesJS 에디터 사용.
- List(리스트): 구독자 그룹.
- Subscriber(구독자): 이메일 + 메타데이터, 여러 리스트에 속할 수 있음.
- Delivery(전송): 캠페인과 수신자 데이터를 기반으로 생성된 예약 전송 인스턴스.
- DeliveryEvent(전송 이벤트): 오픈/클릭 등 애널리틱스를 위한 이벤트 기록.

## 요구사항

- 백엔드:
  - Go 1.24+
- 프런트엔드:
  - Node.js (권장 LTS) 및 pnpm

## 로컬 개발

백엔드
1. Go가 설치되어 있고 `GOPATH`/모듈이 구성되어 있는지 확인합니다.
2. 의존성 다운로드:
   - go mod tidy
3. 서버 실행(개발):
   - go run ./cmd/server
   - 또는 빌드: go build -o headmail ./cmd/server && ./headmail

설정은 `configs/`에서 로드됩니다(예: [configs/config.example.yaml](configs/config.example.yaml)).

프런트엔드
1. cd frontend
2. 패키지 설치:
   - pnpm install
3. 개발 서버 시작:
   - pnpm dev

프런트엔드는 백엔드 API와 통신합니다(`frontend/src/api` 및 `frontend/vite.config.ts`의 프록시/베이스 URL 설정을 확인하세요).

## 테스트

- Go 테스트 실행:
  - go test ./...
- 참고: 테스트는 외부 Go 모듈(예: `github.com/stretchr/testify`)을 사용합니다; 없으면 `go test`가 자동으로 가져옵니다.

## 저장소 / 메일 드라이버 확장

- 저장소(Storage): `pkg/repository/interfaces.go`의 `repository.DB` 및 관련 인터페이스를 구현하면 다른 백엔드(Postgres, MySQL 등)를 추가할 수 있습니다. 서비스 레이어는 인터페이스에만 의존합니다.
- 메일러/리시버(Mailer/Receiver): `pkg/mailer` 및 `pkg/receiver`의 인터페이스를 구현하여 대체 전송 메커니즘이나 서드파티 통합을 추가할 수 있습니다.

## 운영 고려사항

- 템플릿 안전성: 템플릿은 서버 사이드에서 실행됩니다 — 템플릿에 전달되는 데이터를 검증하고 정리하여 비밀 정보가 노출되지 않도록 하십시오.
- 대용량 리스트: 일부 리포지토리/서비스 함수는 구독자를 메모리에 로드하거나 채널을 통해 스트리밍합니다. 매우 큰 리스트를 다룰 경우 페이지네이션/스트리밍 및 배치 전송 생성 구현을 고려하십시오.
- 레이트 리밋/백오프: 고볼륨 전송과 재시도를 위해 큐 또는 워커 풀을 통합하는 것을 권장합니다.

## 프로젝트 구조

- `cmd/server/main.go` — 서버 시작 및 의존 관계 연결
- `pkg/server` — 서버 런타임 및 워커
- `pkg/service/*` — 비즈니스 로직 구현
- `pkg/repository/interfaces.go` — 리포지토리 인터페이스 및 필터
- `internal/db/sqlite` — 참조용 SQLite 리포지토리 구현
- `pkg/template/template.go` — 템플릿 렌더링 서비스 및 테스트

## 기여

- 코드 포맷팅: gofmt/goimports 사용.
- 새로운 로직에는 단위 테스트 포함.
- 리포지토리에 이슈 또는 PR 오픈.

## 라이선스

AGPL-3.0-or-later (자세한 내용은 [LICENSE](LICENSE) 참조)
