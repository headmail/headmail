// Copyright 2025 JC-Lab
// SPDX-License-Identifier: AGPL-3.0-or-later

package domain

// Template represents a reusable email template.
type Template struct {
	ID       string `json:"id"`        // UUID
	Name     string `json:"name"`      // Template name
	Subject  string `json:"subject"`   // Default subject for the template
	BodyHTML string `json:"body_html"` // HTML content of the template
	BodyText string `json:"body_text"` // Text content of the template
	// BodyMJML holds the MJML source (optional). If provided, MJML will be
	// compiled to HTML for preview/editing/preview.
	BodyMJML  string `json:"body_mjml,omitempty"`
	CreatedAt int64  `json:"created_at"` // Unix timestamp seconds
	UpdatedAt int64  `json:"updated_at"` // Unix timestamp seconds
}
