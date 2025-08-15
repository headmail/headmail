package dto

// CreateTemplateRequest defines the request body for creating a new template.
type CreateTemplateRequest struct {
	Name     string `json:"name"`
	BodyHTML string `json:"body_html"`
	BodyText string `json:"body_text"`
	// BodyMJML allows clients to submit MJML source which can be compiled to HTML later.
	BodyMJML string `json:"body_mjml,omitempty"`
}

// UpdateTemplateRequest defines the request body for updating an existing template.
type UpdateTemplateRequest struct {
	Name     string `json:"name"`
	BodyHTML string `json:"body_html"`
	BodyText string `json:"body_text"`
	// BodyMJML allows clients to submit MJML source which can be compiled to HTML later.
	BodyMJML string `json:"body_mjml,omitempty"`
}
