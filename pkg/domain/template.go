package domain

// Template represents a reusable email template.
type Template struct {
	ID        string `json:"id"`         // UUID
	Name      string `json:"name"`       // Template name
	BodyHTML  string `json:"body_html"`  // HTML content of the template
	BodyText  string `json:"body_text"`  // Text content of the template
	CreatedAt int64  `json:"created_at"` // Unix timestamp seconds
	UpdatedAt int64  `json:"updated_at"` // Unix timestamp seconds
}
