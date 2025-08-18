// Copyright 2025 JC-Lab
// SPDX-License-Identifier: AGPL-3.0-or-later

package dto

// CreateTemplateRequest defines the request body for creating a new template.
type CreateTemplateRequest struct {
	Name     string `json:"name"`
	Subject  string `json:"subject"`
	BodyMJML string `json:"body_mjml"`
}

type UpdateTemplateRequest = CreateTemplateRequest
