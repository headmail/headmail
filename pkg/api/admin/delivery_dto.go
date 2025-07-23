package admin

import "github.com/headmail/headmail/pkg/domain"

// CreateTransactionalDeliveryRequest is the request for creating a transactional delivery.
type CreateTransactionalDeliveryRequest struct {
	Name    string                 `json:"name"`
	Email   string                 `json:"email"`
	Subject string                 `json:"subject"`
	Data    map[string]interface{} `json:"data"`
	Headers map[string]string      `json:"headers"`
	Tags    []string               `json:"tags"`
}

// Individual defines an individual recipient for a delivery.
type Individual struct {
	ListID  string                 `json:"listId"`
	Name    string                 `json:"name"`
	Email   string                 `json:"email"`
	Data    map[string]interface{} `json:"data"`
	Headers map[string]string      `json:"headers"`
}

// CreateDeliveriesRequest defines the request body for creating deliveries for a campaign.
type CreateDeliveriesRequest struct {
	Lists       []string     `json:"lists"`
	Individuals []Individual `json:"individuals"`
	ScheduledAt *int64       `json:"scheduled_at"`
}

type CreateDeliveriesResponse struct {
	Status            domain.CampaignStatus `json:"status"`
	ScheduledAt       *int64                `json:"scheduled_at"`
	DeliveriesCreated int                   `json:"deliveries_created"`
}
