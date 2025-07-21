package domain

// Subscriber represents a subscriber to a mailing list.
type Subscriber struct {
	ID             string `json:"id"`                        // UUID
	Email          string `json:"email"`                     // Unique email address
	Name           string `json:"name"`                      // Name of the subscriber
	Status         string `json:"status"`                    // active, unsubscribed
	SubscribedAt   int64  `json:"subscribed_at"`             // Unix timestamp of subscription
	UnsubscribedAt *int64 `json:"unsubscribed_at,omitempty"` // Unix timestamp of unsubscription
	CreatedAt      int64  `json:"created_at"`                // Unix timestamp in seconds
	UpdatedAt      int64  `json:"updated_at"`                // Unix timestamp in seconds
}
