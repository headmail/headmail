package dto

// CreateDeliveriesRequest defines the request body for creating deliveries for a campaign.
type CreateDeliveriesRequest struct {
	Lists       []string     `json:"lists"`
	Individuals []Individual `json:"individuals"`
	ScheduledAt *int64       `json:"scheduled_at"`
}

// Individual defines an individual recipient for a delivery.
type Individual struct {
	ListID  string                 `json:"listId"`
	Name    string                 `json:"name"`
	Email   string                 `json:"email"`
	Data    map[string]interface{} `json:"data"`
	Headers map[string]string      `json:"headers"`
}
