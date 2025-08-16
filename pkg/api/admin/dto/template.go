// Copyright 2025 JC-Lab
// SPDX-License-Identifier: AGPL-3.0-or-later

package dto

// CreateTemplateRequest defines the request body for creating a new template.
type CreateTemplateRequest struct {
	Name     string `json:"name"`
	Subject  string `json:"subject"`
	BodyHTML string `json:"body_html"`
	BodyText string `json:"body_text"`
	// BodyMJML allows clients to submit MJML source which can be compiled to HTML later.
	BodyMJML string `json:"body_mjml,omitempty"`
}

type UpdateTemplateRequest = CreateTemplateRequest
