package domain

// List represents a mailing list.
type List struct {
	ID              string   `json:"id"`                   // UUID
	Name            string   `json:"name"`                 // Name of the list
	Description     string   `json:"description"`          // Description of the list
	Tags            []string `json:"tags"`                 // Tags for categorization
	SubscriberCount int      `json:"subscriber_count"`     // Calculated field for subscriber count
	CreatedAt       int64    `json:"created_at"`           // Unix timestamp in seconds
	UpdatedAt       int64    `json:"updated_at"`           // Unix timestamp in seconds
	DeletedAt       *int64   `json:"deleted_at,omitempty"` // For soft deletes
}
