package domain

type DeliveryType string

const (
	DeliveryTypeCampaign    DeliveryType = "campaign"
	DeliveryTypeTransaction DeliveryType = "transaction"
)

type DeliveryStatus string

const (
	DeliveryStatusScheduled DeliveryStatus = "scheduled"
	DeliveryStatusSending   DeliveryStatus = "sending"
	DeliveryStatusSent      DeliveryStatus = "sent"
	DeliveryStatusDelivered DeliveryStatus = "delivered"
	DeliveryStatusFailed    DeliveryStatus = "failed"
	DeliveryStatusBounced   DeliveryStatus = "bounced"
)

// Delivery represents a single email delivery.
type Delivery struct {
	ID         string                 `json:"id"`                    // UUID
	CampaignID *string                `json:"campaign_id,omitempty"` // Campaign ID (nullable for transactional)
	Type       DeliveryType           `json:"type"`                  // campaign, transactional
	Status     DeliveryStatus         `json:"status"`                // scheduled, sending, sent, delivered, failed, bounced
	Name       string                 `json:"name"`                  // Recipient's name
	Email      string                 `json:"email"`                 // Recipient's email
	Subject    string                 `json:"subject"`               // Actual sent subject
	BodyHTML   string                 `json:"body_html"`             // HTML body
	BodyText   string                 `json:"body_text"`             // Text body
	MessageID  *string                `json:"message_id,omitempty"`  // SMTP Message ID
	Data       map[string]interface{} `json:"data"`                  // Individual data for templates
	Headers    map[string]string      `json:"headers"`               // Individual headers
	Tags       []string               `json:"tags"`                  // Tags for categorization

	// Timestamps
	CreatedAt       int64   `json:"created_at"`             // Creation time
	ScheduledAt     *int64  `json:"scheduled_at,omitempty"` // Scheduled time
	Attempts        int     `json:"attempts"`
	SendScheduledAt *int64  `json:"send_scheduled_at,omitempty"`
	SentAt          *int64  `json:"sent_at,omitempty"`        // Time of sending
	OpenedAt        *int64  `json:"opened_at,omitempty"`      // First open time
	FailedAt        *int64  `json:"failed_at,omitempty"`      // Time of failure
	FailureReason   *string `json:"failure_reason,omitempty"` // Reason for failure

	// Statistics
	OpenCount   int `json:"open_count"`   // Number of opens
	ClickCount  int `json:"click_count"`  // Number of clicks
	BounceCount int `json:"bounce_count"` // Number of bounces
}
