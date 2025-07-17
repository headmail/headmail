# Email Campaign API Server

## Overview

이메일 캠페인 전송을 위한 RESTful API 서버입니다. 대량 이메일 발송, 트랜잭셔널 이메일, 구독 관리, 분석 기능을 제공합니다.

계정 관리는 포함되지 않아 headless 로 즉시 사용할 수 있는 특장점이 있습니다.

## Configuration

- **Config Library**: koanf 사용
- **Environment Variables**: `.` 대신 `_` 사용 (예: `SMTP_HOST`, `DB_CONNECTION_STRING`)
- **Config Files**: YAML, JSON, TOML 지원
- **Priority**: 환경변수 > 설정파일 > 기본값

### Configuration Example
```yaml
# config.yaml
server:
  public:
    port: 8080
    url: "https://mailer.example.com"
  admin:
    port: 8081
  
smtp:
  host: "smtp.gmail.com"
  port: 587
  username: "user@example.com"
  password: "mail-password"
  from:
    name: "Example Service"
    email: "noreply@example.com"
  send:
    batch_size: 100 # 한 번에 발송할 이메일 수
    throttle: 50    # 초당 최대 발송 수

database:
  type: "sqlite" # sqlite, mysql, postgresql, mongodb
  url : "file:data.db?cache=shared&mode=rwc"
```

## API Architecture

### Port Configuration
- **PUBLIC API**: 8080 (기본값)
  - 이메일 수신자용 엔드포인트
- **ADMIN API**: 8081 (기본값)
  - 관리자용 엔드포인트

## ADMIN API Reference

### Lists Management

#### GET `/api/lists`
메일링 리스트 목록 조회 (페이지네이션 지원)

**Request:**
```http
GET /api/lists?page=1&limit=20&search=newsletter&tags[]=newsletter&tags[]=weekly
```

**Query Parameters:**
- `page` (integer, optional): 페이지 번호 (기본값: 1)
- `limit` (integer, optional): 페이지당 항목 수 (기본값: 20, 최대: 100)
- `search` (string, optional): 리스트 이름 (포함) 검색
- `tags[]` (string, optional): 리스트 태그 (in-clause) 검색

**Response:**
```json
{
  "data": [
    {
      "id": "6a88c571-f674-40a2-bdd6-de43a082d489",
      "name": "Weekly Newsletter",
      "description": "주간 뉴스레터 구독자",
      "subscriber_count": 1523,
      "created_at": 1623456789,
      "updated_at": 1623456789,
      "tags": ["newsletter", "weekly"]
    }
  ],
  "pagination": {
    "page": 1,
    "total": 87,
    "limit": 20
  }
}
```

#### POST `/api/lists`
새 메일링 리스트 생성

**Request:**
```http
POST /api/lists
Content-Type: application/json

{
  "name": "Product Updates",
  "description": "제품 업데이트 알림 구독자",
  "tags": ["product", "updates"]
}
```

**Response:**
```json
{
  "id": "6a88c571-f674-40a2-bdd6-de43a082d489",
  "name": "Product Updates",
  "description": "제품 업데이트 알림 구독자",
  "subscriber_count": 0,
  "created_at": 1623456789,
  "updated_at": 1623456789,
  "tags": ["product", "updates"]
}
```

#### GET `/api/lists/{list_id}`
특정 리스트 상세 정보

**Response:**
```json
{
  "id": "6a88c571-f674-40a2-bdd6-de43a082d489",
  "name": "Weekly Newsletter",
  "description": "주간 뉴스레터 구독자",
  "subscriber_count": 1523,
  "created_at": 1623456789,
  "updated_at": 1623456789,
  "tags": ["newsletter", "weekly"]
}
```

#### PUT `/api/lists/{list_id}`
리스트 정보 수정

**Request:**
```json
{
  "name": "Updated List Name",
  "description": "Updated description",
  "tags": ["new-tag"]
}
```

**Response:**
```json
{
  "id": "6a88c571-f674-40a2-bdd6-de43a082d489",
  "name": "Updated List Name",
  "description": "Updated description",
  "subscriber_count": 1523,
  "created_at": 1623456789,
  "updated_at": 1623456789,
  "tags": ["new-tag"]
}
```

#### DELETE `/api/lists/{list_id}`
리스트 삭제 (소프트 삭제)

**Response:**
```json
{
  "deleted": true,
  "message": "List deleted successfully"
}
```

### List Subscribers

#### GET `/api/lists/{list_id}/subscribers`
리스트의 구독자 목록

**Query Parameters:**
- `page` (integer, optional)
- `limit` (integer, optional)
- `status` (string, optional): "active", "unsubscribed"
- `search` (string, optional): 이메일 또는 이름 검색

**Response:**
```json
{
  "data": [
    {
      "id": "e066e6c8-678f-4ea1-8af8-b347453a3eb8",
      "email": "user@example.com",
      "name": "홍길동",
      "status": "active",
      "subscribed_at": 1623456789,
      "unsubscribed_at": null,
      "created_at": 1623456789,
      "updated_at": 1623456789
    }
  ],
  "pagination": {
    "page": 1,
    "total": 1523,
    "limit": 20
  }
}
```

#### POST `/api/lists/{list_id}/subscribers`
구독자 추가 (단일 또는 대량)

**Request:**
```json
{
  "subscribers": [
    {
      "name": "신규 사용자",
      "email": "newuser@example.com",
      "status": "active"
    }
  ],
  "append": true // true-추가, false-전체 교체 (단 email 이 같을 경우 subscriber_id 는 동일하게 유지)
}
```

**Response:**
```json
{
  "created": 1,
  "updated": 0
}
```

### Campaigns Management

#### GET `/api/campaigns`
캠페인 목록 조회 (페이지네이션)

**Query Parameters:**
- `page` (integer, optional)
- `limit` (integer, optional)
- `status[]` (string, optional): "draft", "scheduled", "sending", "sent", "paused", "cancelled" (in-clause 검색)
- `search` (string, optional): 캠페인 이름 검색
- `tags[]` (string, optional): 리스트 태그 (in-clause) 검색


**Response:**
```json
{
  "data": [
    {
      "id": "4554a1e9-4313-4186-99d1-89836b5958d9",
      "name": "여름 세일 알림",
      "status": "sent",
      "created_at": 1623456789,
      "scheduled_at": 1623460000,
      "sent_at": 1623460000,
      "recipient_count": 1523,
      "delivered_count": 1487,
      "open_count": 234, // 한 번이라도 열어본 수신자 수
      "click_count": 45, // 한 번이라도 클릭한 수신자 수
      "bounce_count": 0, // 반송 수
      "tags": ["summer-sale", "promotion"]
    }
  ],
  "pagination": {
    "page": 1,
    "total": 45,
    "limit": 20
  }
}
```

#### POST `/api/campaigns`
새 캠페인 생성

**Request:**

```http
POST /api/campaigns
Content-Type: application/json

{
  "name": "여름 세일 알림",
  "from_name": "ABC Store",
  "from_email": "noreply@abcstore.com",
  "subject": "{{ i18n . \"ko\" \"summer_sale.subject\" }}",
  "template_html": "<html><body><h1>{{ i18n . \"ko\" \"greeting\" }} {{ .name }}!</h1><p>{{ i18n . \"ko\" \"summer_sale.body\" }}</p></body></html>",
  "template_text": "{{ i18n \"ko\" .locale \"greeting\" }} {{ .name }}! {{ i18n . \"ko\" \"summer_sale.body\" }}",
  "data": {
    "i18n": {
      "ko": {
        "greeting": "안녕하세요",
        "summer_sale": {
          "subject": "🌞 여름 세일 최대 50% 할인!",
          "preview": "7월 한정 특가 할인을 놓치지 마세요",
          "body": "여름 맞이 특별 할인 이벤트가 진행 중입니다. 모든 상품 최대 50% 할인!"
        }
      },
      "en": {
        "greeting": "Hello",
        "summer_sale": {
          "subject": "🌞 Summer Sale Up to 50% OFF!",
          "preview": "Don't miss July special discounts",
          "body": "Summer special discount event is ongoing. Up to 50% off on all items!"
        }
      }
    }
  },
  "tags": ["summer-sale", "promotion"],
  "headers": {
    "X-Campaign-ID": "{{ .campaignId }}",
    "List-Unsubscribe": "<{{ .unsubscribeUrl }}>"
  },
  "scheduled_at": null,
  "utm_params": {
    "source": "newsletter",
    "medium": "email",
    "campaign": "summer-sale-2024"
  } // 이메일 내 링크 클릭 시 추가할 파라미터
}
```

**Response:**
```json
{
  "id": "4554a1e9-4313-4186-99d1-89836b5958d9",
  "name": "여름 세일 알림",
  "status": "draft",
  "created_at": 1623456789,
  "updated_at": 1623456789,
  "from_name": "ABC Store",
  "from_email": "noreply@abcstore.com",
  "subject": "{{ i18n . \"ko\" \"summer_sale.subject\" }}",
  "template_html": "<html>...</html>",
  "template_text": "...",
  "data": {...},
  "tags": ["summer-sale", "promotion"],
  "headers": {...},
  "utm_params": {...},
  "scheduled_at": null,
  "sent_at": null,
  "recipient_count": 0,
  "delivered_count": 0,
  "failed_count": 0,
  "open_count": 0,
  "click_count": 0,
  "bounce_count": 0
}
```

#### GET `/api/campaigns/{campaign_id}`
특정 캠페인 상세 정보

**Response:**
```json
{
  "id": "4554a1e9-4313-4186-99d1-89836b5958d9",
  "name": "여름 세일 알림",
  "status": "draft",
  "created_at": 1623456789,
  "updated_at": 1623456789,
  
  "from_name": "ABC Store",
  "from_email": "noreply@abcstore.com",
  "subject": "{{ i18n . \"ko\" \"summer_sale.subject\" }}",
  "template_html": "<html>...</html>",
  "template_text": "...",
  "data": {...},
  "tags": ["summer-sale", "promotion"],
  "headers": {...},
  "utm_params": {...},
  "scheduled_at": null,
  "sent_at": null,
  "recipient_count": 0,
  "delivered_count": 0,
  "failed_count": 0,
  "open_count": 0,
  "click_count": 0,
  "bounce_count": 0
}
```

#### PUT `/api/campaigns/{campaign_id}`
캠페인 수정

**Request:**
```json
{
  "name": "수정된 캠페인 이름",
  "subject": "새로운 제목",
  "scheduled_at": 1623460000
}
```

**Response:**
```json
{
  "id": "4554a1e9-4313-4186-99d1-89836b5958d9",
  "name": "수정된 캠페인 이름",
  "status": "draft",
  "created_at": 1623456789,
  "updated_at": 1623456789,
  "from_name": "ABC Store",
  "from_email": "noreply@abcstore.com",
  "subject": "새로운 제목",
  "template_html": "<html>...</html>",
  "template_text": "...",
  "data": {...},
  "tags": ["summer-sale", "promotion"],
  "headers": {...},
  "utm_params": {...},
  "scheduled_at": 1623460000,
  "sent_at": null,
  "recipient_count": 1523,
  "delivered_count": 0,
  "failed_count": 0,
  "open_count": 0,
  "click_count": 0,
  "bounce_count": 0
}
```

#### DELETE `/api/campaigns/{campaign_id}`
캠페인 삭제 (draft 상태만 가능)

**Response:**
```json
{
  "deleted": true,
  "message": "Campaign deleted successfully"
}
```

### Campaign Delivery

#### POST `/api/campaigns/{campaign_id}/deliveries`
캠페인 발송 시작

**Request:**
```http
POST /api/campaigns/{campaign_id}/deliveries
Content-Type: application/json

{
  "lists": ["6a88c571-f674-40a2-bdd6-de43a082d489", "6270d21f-805a-43d3-9bd1-544ca306d604"],
  "individuals": [
    {
      "listId": "d6172ca8-5e32-40e1-84e7-32af869c9f1c", // 해당 list 에 subscriber upsert
      "name": "VIP 고객",
      "email": "vip@example.com",
      "data": {
        "locale": "ko",
        "vip_level": "gold"
      },
      "headers": {
        "X-VIP-Level": "gold"
      }
    }
  ],
  "scheduled_at": 1623460000 // null이면 즉시 발송
}
```

**Response:**
```json
{
  "status": "scheduled",
  "scheduled_at": 1623460000,
  "deliveries_created": 1523
}
```

#### GET `/api/campaigns/{campaign_id}/deliveries`
캠페인의 발송 이력

**Response:**
```json
{
  "data": [
    {
      "id": "d8db7b87-8675-48b9-b88f-e07097be0ecf",
      "campaign_id": "4554a1e9-4313-4186-99d1-89836b5958d9",
      "name": "VIP 고객",
      "email": "vip@example.com",
      "subject": "🌞 여름 세일 최대 50% 할인!",
      "status": "delivered",
      "created_at": 1623456789,
      "scheduled_at": 1623460000,
      "sent_at": 1623460000,
      "opened_at": 1623468000,
      "open_count": 1,
      "click_count": 0,
      "bounce_count": 0
    }
  ],
  "pagination": {
    "page": 1,
    "total": 87,
    "limit": 20
  }
}
```

#### GET `/api/campaigns/{campaign_id}/deliveries/{delivery_id}`
특정 발송 상세 정보

**Response:**
```json
{
  "id": "d8db7b87-8675-48b9-b88f-e07097be0ecf",
  "campaign_id": "4554a1e9-4313-4186-99d1-89836b5958d9",
  "type": "campaign",
  "status": "delivered",
  "name": "VIP 고객",
  "email": "vip@example.com",
  "subject": "🌞 여름 세일 최대 50% 할인!",
  "data": {
    "locale": "ko",
    "vip_level": "gold"
  },
  "headers": {
    "X-VIP-Level": "gold"
  },
  "tags": ["summer-sale", "promotion"],
  "created_at": 1623456789,
  "scheduled_at": 1623460000,
  "sent_at": 1623460000,
  "opened_at": 1623468000,
  "failed_at": null,
  "failure_reason": null,
  "open_count": 1,
  "click_count": 0,
  "bounce_count": 0
}
```

### Campaign Status Management

#### PUT `/api/campaigns/{campaign_id}/status`
캠페인 상태 변경

**Request:**
```json
{
  "status": "paused",
  "reason": "일시 중지 - 재검토 필요"
}
```

**Status Values:**
- `draft`: 초안
- `scheduled`: 예약됨
- `sending`: 발송 중
- `sent`: 발송 완료
- `paused`: 일시 중지
- `cancelled`: 취소됨


### Transactional Email

#### POST `/api/tx`
트랜잭셔널 이메일 즉시 발송

**Request:**
```http
POST /api/tx
Content-Type: application/json

{
  "name": "OTP 발송",
  
  "from_name": "ABC Store",
  "from_email": "noreply@abcstore.com",
  "to": {
    "name": "홍길동",
    "email": "user@example.com"
  },
  "subject": "{{ i18n . .locale \"otp.subject\" }}",
  "template": {
    "html": "<html><body><h1>{{ i18n . .locale \"otp.title\" }}</h1><p>{{ i18n . .locale \"otp.message\" }}: <strong>{{ .otp }}</strong></p></body></html>",
    "text": "{{ i18n . .locale \"otp.title\" }}\n{{ i18n . .locale \"otp.message\" }}: {{ .otp }}"
  },
  "data": {
    "otp": "123456",
    "locale": "ko",
    "i18n": {
      "ko": {
        "otp": {
          "subject": "인증번호가 도착했습니다",
          "title": "인증번호 안내",
          "message": "인증번호는 다음과 같습니다"
        }
      },
      "en": {
        "otp": {
          "subject": "Your verification code",
          "title": "Verification Code",
          "message": "Your verification code is"
        }
      }
    }
  },
  "tags": ["transactional", "otp"],
  "headers": {
    "X-Transaction-ID": "tx-12345"
  }
}
```

**Response:**
```json
{
  "delivery_id": "tx_7f8a9b2c-4d5e-6f7a-8b9c-0d1e2f3a4b5c",
  "status": "sent",
  "sent_at": 1623456789,
  "recipient": {
    "email": "user@example.com",
    "name": "홍길동"
  }
}
```

#### GET `/api/tx/{delivery_id}`
트랜잭셔널 이메일 발송 상태 확인

**Response:**
```json
{
  "id": "tx_7f8a9b2c-4d5e-6f7a-8b9c-0d1e2f3a4b5c",
  "type": "transactional",
  "status": "delivered",
  "name": "홍길동",
  "email": "user@example.com",
  "subject": "인증번호가 도착했습니다",
  "data": {
    "otp": "123456",
    "locale": "ko",
    "i18n": {
      "ko": {
        "otp": {
          "subject": "인증번호가 도착했습니다",
          "title": "인증번호 안내",
          "message": "인증번호는 다음과 같습니다"
        }
      },
      "en": {
        "otp": {
          "subject": "Your verification code",
          "title": "Verification Code",
          "message": "Your verification code is"
        }
      }
    }
  },
  "headers": {
    "X-Transaction-ID": "tx-12345"
  },
  "tags": ["transactional", "otp"],
  "created_at": 1623456789,
  "sent_at": 1623456789,
  "opened_at": 1623456795,
  "open_count": 1,
  "click_count": 0,
  "bounce_count": 0
}
```

# PUBLIC API

## Delivery API

### `/d/{delivery_id}/logo.png`

- 읽음 확인을 위한 더미 이미지

Param
- `preview=y` : 통계 업데이트 안함

### `/d/{delivery_id}/unsubscribe`

- HTML 형식의 unsubscribe 화면.
- 진짜 할건지 확인 후 `/d/{delivery_id}/unsubscribe?confirm=y` 으로 이동하여 발송한 list 에서 unsubscribed 상태로 바꿈

Param
- `preview=y` : 통계 업데이트 안함

### `/d/{delivery_id}/link?to={url}`

- 클릭 확인 및 utm 추가
- 해당 to url 에 utm_params 추가해서 리디렉션

Param
- `preview=y` : 통계 업데이트 안함

# Template

## Functions

### `i18n <context> <default locale> <key>`

`<context>` 는 항상 `.` 이어야 합니다.

data 에는 기본적으로 다음의 key 들이 필요합니다.

- `i18n` : Map<Locale, Map> 형식의 i18n messages 입니다.
- `locale` : 해당 사용자의 locale 입니다.

### 기본 data

- 기본 campaign 의 data 을 상속받으며, 사용자/list 별로 data 을 override 할 수 있음.
- 기본적으로 `.deliveryId` 으로 메일 식별 ID, `.name` 으로 이름, `.mail` 으로 메일 주소 접근 가능

# DB

- 추상화 되어 DB에 종속적이지 않아 다양한 DB (MongoDB / SQLite / MySQL 등) 지원이 가능해야 합니다.
- 따라서 DB 접근은 DTO 와 인터페이스를 활용하세요.

## Abstract Entities

### List
```go
type List struct {
    ID              string    `json:"id"`                           // UUID
    Name            string    `json:"name"`                       // 리스트 이름
    Description     string    `json:"description"`         // 리스트 설명
    Tags            []string  `json:"tags"`                       // 태그 배열
    SubscriberCount int       `json:"subscriber_count"` // 구독자 수 (계산 필드)
    CreatedAt       int64     `json:"created_at"`           // Unix timestamp seconds
    UpdatedAt       int64     `json:"updated_at"`           // Unix timestamp seconds
    DeletedAt       *int64    `json:"deleted_at,omitempty"` // 소프트 삭제용
}
```

### Subscriber
```go
type Subscriber struct {
    ID             string  `json:"id"`                               // UUID
    Email          string  `json:"email"`                         // 이메일 (unique)
    Name           string  `json:"name"`                           // 이름
    Status         string  `json:"status"`                       // active, unsubscribed
    SubscribedAt   int64   `json:"subscribed_at"`         // 구독 시간
    UnsubscribedAt *int64  `json:"unsubscribed_at,omitempty"` // 구독 취소 시간 Unix timestamp seconds
    CreatedAt      int64   `json:"created_at"`               // Unix timestamp seconds
    UpdatedAt      int64   `json:"updated_at"`               // Unix timestamp seconds
}
```

### Campaign
```go
type CampaignStatus string

const (
  CampaignStatusDraft     CampaignStatus = "draft"
  CampaignStatusScheduled CampaignStatus = "scheduled"
  CampaignStatusSending   CampaignStatus = "sending"
  CampaignStatusSent      CampaignStatus = "sent"
  CampaignStatusPaused    CampaignStatus = "paused"
  CampaignStatusCancelled CampaignStatus = "cancelled"
)

type Campaign struct {
    ID             string            `json:"id"`                       // UUID
    Name           string            `json:"name"`                   // 캠페인 이름
    Status         CampaignStatus    `json:"status"`
    FromName       string            `json:"from_name"`         // 발신자 이름
    FromEmail      string            `json:"from_email"`       // 발신자 이메일
    Subject        string            `json:"subject"`             // 제목 템플릿
    TemplateHTML   string            `json:"template_html"` // HTML 템플릿
    TemplateText   string            `json:"template_text"` // TEXT 템플릿
    Data           map[string]interface{} `json:"data"`          // JSON 데이터
    Tags           []string          `json:"tags"`                   // 태그 배열
    Headers        map[string]string `json:"headers"`             // 추가 헤더
    UTMParams      map[string]string `json:"utm_params"`       // UTM 파라미터
    ScheduledAt    *int64            `json:"scheduled_at,omitempty"` // 예약 시간
    SentAt         *int64            `json:"sent_at,omitempty"`   // 발송 시간
    CreatedAt      int64             `json:"created_at"`       // Unix timestamp
    UpdatedAt      int64             `json:"updated_at"`       // Unix timestamp
    DeletedAt      *int64            `json:"deleted_at,omitempty"` // 소프트 삭제용
    
    // 통계 필드 (계산 필드)
    RecipientCount int `json:"recipient_count"`
    DeliveredCount int `json:"delivered_count"`
    FailedCount    int `json:"failed_count"`
    OpenCount      int `json:"open_count"`
    ClickCount     int `json:"click_count"`
    BounceCount    int `json:"bounce_count"`
}
```

### Delivery
```go
type Delivery struct {
    ID            string            `json:"id"`                     // UUID
    CampaignID    *string           `json:"campaign_id,omitempty"` // Campaign ID (nullable for transactional)
    Type          string            `json:"type"`                 // campaign, transactional
    Status        string            `json:"status"`             // scheduled, sending, sent, delivered, failed, bounced
    Name          string            `json:"name"`                 // 수신자 이름
    Email         string            `json:"email"`               // 수신자 이메일
    Subject       string            `json:"subject"`           // 실제 발송된 제목
    MessageID     *string           `json:"message_id,omitempty"` // SMTP Message ID
    Data          map[string]interface{} `json:"data"`        // 개별 데이터
    Headers       map[string]string `json:"headers"`           // 개별 헤더
    Tags          []string          `json:"tags"`                 // 태그
    
    // 시간 필드
    CreatedAt     int64   `json:"created_at"`             // 생성 시간
    ScheduledAt   *int64  `json:"scheduled_at,omitempty"` // 예약 시간
    SentAt        *int64  `json:"sent_at,omitempty"`         // 발송 시간
    OpenedAt      *int64  `json:"opened_at,omitempty"`     // 첫 읽음 시간
    FailedAt      *int64  `json:"failed_at,omitempty"`     // 실패 시간
    FailureReason *string `json:"failure_reason,omitempty"` // 실패 사유
    
    // 통계 필드
    OpenCount     int `json:"open_count"`                 // 읽음 횟수
    ClickCount    int `json:"click_count"`               // 클릭 횟수
    BounceCount   int `json:"bounce_count"`             // 반송 횟수
}
```

### DeliveryEvent
```go
type EventType string

const (
    EventTypeSent        EventType = "sent"
    EventTypeDelivered   EventType = "delivered"
    EventTypeOpened      EventType = "opened"
    EventTypeClicked     EventType = "clicked"
    EventTypeBounced     EventType = "bounced"
    EventTypeComplained  EventType = "complained"
    EventTypeUnsubscribed EventType = "unsubscribed"
)

type DeliveryEvent struct {
    ID         string            `json:"id"`                 // UUID
    DeliveryID string            `json:"delivery_id"`        // Delivery ID
    EventType  EventType         `json:"event_type"`         // 이벤트 타입
    EventData  map[string]interface{} `json:"event_data"`    // 이벤트 관련 데이터
    UserAgent  *string           `json:"user_agent,omitempty"` // User Agent (opened, clicked용)
    IPAddress  *string           `json:"ip_address,omitempty"` // IP 주소
    URL        *string           `json:"url,omitempty"`      // 클릭된 URL (clicked용)
    CreatedAt  int64             `json:"created_at"`         // Unix timestamp seconds
}
```

## Database Interfaces

### Repository Interfaces
```go
type Transactionable interface {
    Begin(ctx context.Context) (context.Context, error)
    Commit(ctx context.Context) error
    Rollback(ctx context.Context) error
}

// ListRepository 리스트 저장소 인터페이스
type ListRepository interface {
    Create(ctx context.Context, list *List) error
    GetByID(ctx context.Context, id string) (*List, error)
    Update(ctx context.Context, list *List) error
    Delete(ctx context.Context, id string) error
    List(ctx context.Context, filter ListFilter, pagination Pagination) ([]*List, int, error)
    GetSubscriberCount(ctx context.Context, listID string) (int, error)
    GetSubscribers(ctx context.Context) (chan *Subscriber, error)
}

// SubscriberRepository 구독자 저장소 인터페이스
type SubscriberRepository interface {
    Create(ctx context.Context, subscriber *Subscriber) error
    GetByID(ctx context.Context, id string) (*Subscriber, error)
    GetByEmail(ctx context.Context, email string) (*Subscriber, error)
    Update(ctx context.Context, subscriber *Subscriber) error
    Delete(ctx context.Context, id string) error
    List(ctx context.Context, filter SubscriberFilter, pagination Pagination) ([]*Subscriber, int, error)
    BulkUpsert(ctx context.Context, subscribers []*Subscriber) error
}

// CampaignRepository 캠페인 저장소 인터페이스
type CampaignRepository interface {
    Create(ctx context.Context, campaign *Campaign) error
    GetByID(ctx context.Context, id string) (*Campaign, error)
    Update(ctx context.Context, campaign *Campaign) error
    Delete(ctx context.Context, id string) error
    List(ctx context.Context, filter CampaignFilter, pagination Pagination) ([]*Campaign, int, error)
    UpdateStatus(ctx context.Context, id string, status CampaignStatus) error
}

// DeliveryRepository 발송 저장소 인터페이스
type DeliveryRepository interface {
    Create(ctx context.Context, delivery *Delivery) error
    GetByID(ctx context.Context, id string) (*Delivery, error)
    Update(ctx context.Context, delivery *Delivery) error
    List(ctx context.Context, filter DeliveryFilter, pagination Pagination) ([]*Delivery, int, error)
    GetByCampaignID(ctx context.Context, campaignID string, pagination Pagination) ([]*Delivery, int, error)
    UpdateStatus(ctx context.Context, id string, status string) error
}
```

### Filter Types
```go
type ListFilter struct {
    Search string   `json:"search,omitempty"`
    Tags   []string `json:"tags,omitempty"`
}

type SubscriberFilter struct {
    ListID string   `json:"list_id,omitempty"`
    Status string   `json:"status,omitempty"`
    Search string   `json:"search,omitempty"`
}

type CampaignFilter struct {
    Status []CampaignStatus `json:"status,omitempty"`
    Search string           `json:"search,omitempty"`
    Tags   []string         `json:"tags,omitempty"`
}

type DeliveryFilter struct {
    CampaignID string `json:"campaign_id,omitempty"`
    Type       string `json:"type,omitempty"`
    Status     string `json:"status,omitempty"`
    Email      string `json:"email,omitempty"`
}

type Pagination struct {
    Page  int `json:"page"`
    Limit int `json:"limit"`
}
```

# Dev Guide

- 주석은 영어만 작성하세요.
- 코드 내의 주석은 꼭 필요한 경우에만 작성하고, 함수 자체에 주석을 주로 사용하세요.
- 코드를 작성할 때 unit test 을 고려하여 mock 이 쉽도록 고려하세요.
- `backlog.md` 을 이용하여 프로젝트를 관리합니다. `backlog.md` 사용법을 모른다면 context7 의 get-library-docs 을 이용해서 `/mrlesk/backlog.md` 의 task cli 사용법을 찾아보세요.

# Project Structure

```
/
├── cmd/
│   └── server/
│       └── main.go             # 애플리케이션 시작점
├── configs/
│   └── config.example.yaml     # 설정 파일 예시
├── pkg/
│   ├── server/                 # 서버 구현 (Admin/Public API 서버 실행)
│   ├── api/                    # API 계층 (HTTP 핸들러, 라우팅, 미들웨어)
│   │   ├── admin/              # Admin API 관련 핸들러
│   │   ├── public/             # Public API 관련 핸들러
│   │   └── dto/                # 데이터 전송 객체 (Request/Response)
│   ├── config/                 # 설정 로딩 및 관리
│   │   └── config.go
│   ├── domain/                 # 핵심 도메인 모델 및 비즈니스 로직
│   │   ├── campaign.go
│   │   ├── delivery.go
│   │   ├── list.go
│   │   └── subscriber.go
│   ├── repository/             # 데이터베이스 접근 계층
│   │   └── interfaces.go       # 리포지토리 인터페이스 정의
│   ├── service/                # 비즈니스 로직 서비스
│   │   ├── campaign_service.go
│   │   ├── list_service.go
│   │   ├── tx_service.go
│   │   └── delivery_service.go
│   ├── mailer/                 # 이메일 발송 관련 로직
│   │   └── smtp.go
│   └── template/               # 이메일 템플릿 처리
│       └── template.go
├── internal/
│   └── db/                     # DB 별 실제 Entity/Repository 구현
│       └── sqlite/             # sqlite 구현
│       └── mongodb/            # mongodb 구현
├── go.mod
└── go.sum
```

# Roadmap

- Double Opt-in
- Common Template (double opt-in 에도 사용)
