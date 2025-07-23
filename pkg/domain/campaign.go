package domain

// CampaignStatus represents the status of a campaign.
type CampaignStatus string

const (
	CampaignStatusDraft     CampaignStatus = "draft"
	CampaignStatusScheduled CampaignStatus = "scheduled"
	CampaignStatusSending   CampaignStatus = "sending"
	CampaignStatusSent      CampaignStatus = "sent"
	CampaignStatusPaused    CampaignStatus = "paused"
	CampaignStatusCancelled CampaignStatus = "cancelled"
)

// Campaign represents an email campaign.
type Campaign struct {
	ID           string                 `json:"id"`                     // UUID
	Name         string                 `json:"name"`                   // Name of the campaign
	Status       CampaignStatus         `json:"status"`                 // Status of the campaign
	FromName     string                 `json:"from_name"`              // Sender's name
	FromEmail    string                 `json:"from_email"`             // Sender's email
	Subject      string                 `json:"subject"`                // Subject template
	TemplateID   *string                `json:"template_id,omitempty"`  // Optional template ID
	TemplateHTML string                 `json:"template_html"`          // HTML template
	TemplateText string                 `json:"template_text"`          // Plain text template
	Data         map[string]interface{} `json:"data"`                   // JSON data for templates
	Tags         []string               `json:"tags"`                   // Tags for categorization
	Headers      map[string]string      `json:"headers"`                // Additional email headers
	UTMParams    map[string]string      `json:"utm_params"`             // UTM parameters for link tracking
	ScheduledAt  *int64                 `json:"scheduled_at,omitempty"` // Scheduled time for sending
	SentAt       *int64                 `json:"sent_at,omitempty"`      // Time when sending was completed
	CreatedAt    int64                  `json:"created_at"`             // Unix timestamp
	UpdatedAt    int64                  `json:"updated_at"`             // Unix timestamp
	DeletedAt    *int64                 `json:"deleted_at,omitempty"`   // For soft deletes

	// Calculated statistics
	RecipientCount int `json:"recipient_count"`
	DeliveredCount int `json:"delivered_count"`
	FailedCount    int `json:"failed_count"`
	OpenCount      int `json:"open_count"`
	ClickCount     int `json:"click_count"`
	BounceCount    int `json:"bounce_count"`
}
