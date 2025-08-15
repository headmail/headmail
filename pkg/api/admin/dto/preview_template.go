// Package dto contains request/response DTOs for admin APIs.
package dto

// PreviewTemplateRequest is the payload for server-side template preview.
// It provides template content and sample subscriber fields used during rendering.
type PreviewTemplateRequest struct {
	TemplateHTML string `json:"templateHtml,omitempty"`
	TemplateText string `json:"templateText,omitempty"`
	Subject      string `json:"subject,omitempty"`
	// Sample subscriber fields used during rendering
	Name  string `json:"name"`
	Email string `json:"email"`
}

// PreviewTemplateResponse contains rendered template outputs returned by the preview endpoint.
type PreviewTemplateResponse struct {
	HTML    string `json:"html"`
	Text    string `json:"text,omitempty"`
	Subject string `json:"subject,omitempty"`
}
