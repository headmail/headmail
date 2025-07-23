package sqlite

import "github.com/headmail/headmail/pkg/domain"

// List is the GORM model for a mailing list.
type List struct {
	ID          string `gorm:"column:id;primaryKey"`
	Name        string `gorm:"column:name"`
	Description string `gorm:"column:description"`
	Tags        JSON   `gorm:"column:tags;type:json"`
	CreatedAt   int64  `gorm:"column:created_at"`
	UpdatedAt   int64  `gorm:"column:updated_at"`
	DeletedAt   *int64 `gorm:"column:deleted_at;index"`
}

// Subscriber is the GORM model for a subscriber.
type Subscriber struct {
	ID        string                  `gorm:"column:id;primaryKey"`
	Email     string                  `gorm:"column:email;uniqueIndex"`
	Name      string                  `gorm:"column:name"`
	Status    domain.SubscriberStatus `gorm:"column:status"`
	CreatedAt int64                   `gorm:"column:created_at"`
	UpdatedAt int64                   `gorm:"column:updated_at"`
	Lists     []SubscriberList        `gorm:"foreignKey:SubscriberID"`
}

// SubscriberList is the GORM model for the join table between subscribers and lists.
type SubscriberList struct {
	SubscriberID   string                      `gorm:"column:subscriber_id;primaryKey"`
	ListID         string                      `gorm:"column:list_id;primaryKey"`
	Status         domain.SubscriberListStatus `gorm:"column:status"`
	SubscribedAt   *int64                      `gorm:"column:subscribed_at"`
	UnsubscribedAt *int64                      `gorm:"column:unsubscribed_at"`
	CreatedAt      int64                       `gorm:"column:created_at"`
	UpdatedAt      int64                       `gorm:"column:updated_at"`
}

// Campaign is the GORM model for a campaign.
type Campaign struct {
	ID             string                `gorm:"column:id;primaryKey"`
	Name           string                `gorm:"column:name"`
	TemplateID     *string               `gorm:"column:template_id"`
	Status         domain.CampaignStatus `gorm:"column:status"`
	FromName       string                `gorm:"column:from_name"`
	FromEmail      string                `gorm:"column:from_email"`
	Subject        string                `gorm:"column:subject"`
	TemplateHTML   string                `gorm:"column:template_html"`
	TemplateText   string                `gorm:"column:template_text"`
	Data           JSON                  `gorm:"column:data;type:json"`
	Tags           JSON                  `gorm:"column:tags;type:json"`
	Headers        JSON                  `gorm:"column:headers;type:json"`
	UTMParams      JSON                  `gorm:"column:utm_params;type:json"`
	ScheduledAt    *int64                `gorm:"column:scheduled_at"`
	SentAt         *int64                `gorm:"column:sent_at"`
	CreatedAt      int64                 `gorm:"column:created_at"`
	UpdatedAt      int64                 `gorm:"column:updated_at"`
	DeletedAt      *int64                `gorm:"column:deleted_at;index"`
	RecipientCount int                   `gorm:"column:recipient_count"`
	DeliveredCount int                   `gorm:"column:delivered_count"`
	FailedCount    int                   `gorm:"column:failed_count"`
	OpenCount      int                   `gorm:"column:open_count"`
	ClickCount     int                   `gorm:"column:click_count"`
	BounceCount    int                   `gorm:"column:bounce_count"`
}

// Delivery is the GORM model for a delivery.
type Delivery struct {
	ID            string                `gorm:"column:id;primaryKey"`
	CampaignID    *string               `gorm:"column:campaign_id"`
	Type          domain.DeliveryType   `gorm:"column:type"`
	Status        domain.DeliveryStatus `gorm:"column:status"`
	Name          string                `gorm:"column:name"`
	Email         string                `gorm:"column:email"`
	Subject       string                `gorm:"column:subject"`
	MessageID     *string               `gorm:"column:message_id"`
	Data          JSON                  `gorm:"column:data;type:json"`
	Headers       JSON                  `gorm:"column:headers;type:json"`
	Tags          JSON                  `gorm:"column:tags;type:json"`
	CreatedAt     int64                 `gorm:"column:created_at"`
	ScheduledAt   *int64                `gorm:"column:scheduled_at"`
	SentAt        *int64                `gorm:"column:sent_at"`
	OpenedAt      *int64                `gorm:"column:opened_at"`
	FailedAt      *int64                `gorm:"column:failed_at"`
	FailureReason *string               `gorm:"column:failure_reason"`
	OpenCount     int                   `gorm:"column:open_count"`
	ClickCount    int                   `gorm:"column:click_count"`
	BounceCount   int                   `gorm:"column:bounce_count"`
}

// DeliveryEvent is the GORM model for a delivery event.
type DeliveryEvent struct {
	ID         string           `gorm:"column:id;primaryKey"`
	DeliveryID string           `gorm:"column:delivery_id"`
	EventType  domain.EventType `gorm:"column:event_type"`
	EventData  JSON             `gorm:"column:event_data;type:json"`
	UserAgent  *string          `gorm:"column:user_agent"`
	IPAddress  *string          `gorm:"column:ip_address"`
	URL        *string          `gorm:"column:url"`
	CreatedAt  int64            `gorm:"column:created_at"`
}

// Template is the GORM model for a template.
type Template struct {
	ID        string `gorm:"column:id;primaryKey"`
	CreatedAt int64  `gorm:"column:created_at"`
	UpdatedAt int64  `gorm:"column:updated_at"`
	Name      string `gorm:"column:name"`
	BodyHTML  string `gorm:"column:body_html"`
	BodyText  string `gorm:"column:body_text"`
}
