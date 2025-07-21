package domain

// EventType represents the type of a delivery event.
type EventType string

const (
	EventTypeSent         EventType = "sent"
	EventTypeDelivered    EventType = "delivered"
	EventTypeOpened       EventType = "opened"
	EventTypeClicked      EventType = "clicked"
	EventTypeBounced      EventType = "bounced"
	EventTypeComplained   EventType = "complained"
	EventTypeUnsubscribed EventType = "unsubscribed"
)

// DeliveryEvent represents an event related to a delivery.
type DeliveryEvent struct {
	ID         string                 `json:"id"`                   // UUID
	DeliveryID string                 `json:"delivery_id"`          // Delivery ID
	EventType  EventType              `json:"event_type"`           // Type of the event
	EventData  map[string]interface{} `json:"event_data"`           // Event-related data
	UserAgent  *string                `json:"user_agent,omitempty"` // User agent for open/click events
	IPAddress  *string                `json:"ip_address,omitempty"` // IP address
	URL        *string                `json:"url,omitempty"`        // Clicked URL for click events
	CreatedAt  int64                  `json:"created_at"`           // Unix timestamp in seconds
}
