// Copyright 2025 JC-Lab
// SPDX-License-Identifier: AGPL-3.0-or-later

package domain

// Template represents a reusable email template.
type Template struct {
	ID        string `json:"id"`         // UUID
	Name      string `json:"name"`       // Template name
	CreatedAt int64  `json:"created_at"` // Unix timestamp seconds
	UpdatedAt int64  `json:"updated_at"` // Unix timestamp seconds
	Subject   string `json:"subject"`    // Default subject for the template
	BodyMJML  string `json:"body_mjml,omitempty"`
}
