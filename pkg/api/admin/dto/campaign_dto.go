// Copyright 2025 JC-Lab
// SPDX-License-Identifier: AGPL-3.0-or-later

package dto

import "github.com/headmail/headmail/pkg/domain"

// CreateCampaignRequest is the request for creating a campaign.
type CreateCampaignRequest struct {
	Name         string                 `json:"name"`
	Status       domain.CampaignStatus  `json:"status"`
	FromName     string                 `json:"from_name"`
	FromEmail    string                 `json:"from_email"`
	Subject      string                 `json:"subject"`
	TemplateID   *string                `json:"template_id,omitempty"`
	TemplateHTML string                 `json:"template_html"`
	TemplateText string                 `json:"template_text"`
	Data         map[string]interface{} `json:"data"`
	Tags         []string               `json:"tags"`
	Headers      map[string]string      `json:"headers"`
	UTMParams    map[string]string      `json:"utm_params"`
	ScheduledAt  *int64                 `json:"scheduled_at,omitempty"`
}

// UpdateCampaignRequest is the request for updating a campaign.
type UpdateCampaignRequest = CreateCampaignRequest
