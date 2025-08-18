// Copyright 2025 JC-Lab
// SPDX-License-Identifier: AGPL-3.0-or-later

package dto

import "github.com/headmail/headmail/pkg/domain"

// CreateTransactionalDeliveryRequest is the request for creating a transactional delivery.
type CreateTransactionalDeliveryRequest struct {
	Name         string                 `json:"name"`
	Email        string                 `json:"email"`
	FromName     *string                `json:"from_name"`
	FromEmail    *string                `json:"from_email"`
	Subject      *string                `json:"subject"`
	TemplateID   *string                `json:"template_id,omitempty"`
	TemplateMJML string                 `json:"template_mjml"`
	Data         map[string]interface{} `json:"data"`
	Tags         []string               `json:"tags"`
	Headers      map[string]string      `json:"headers"`
}

// Individual defines an individual recipient for a delivery.
type Individual struct {
	Name    string                 `json:"name"`
	Email   string                 `json:"email"`
	Data    map[string]interface{} `json:"data"`
	Headers map[string]string      `json:"headers"`
}

// CreateDeliveriesRequest defines the request body for creating deliveries for a campaign.
type CreateDeliveriesRequest struct {
	Lists       []string     `json:"lists"`
	Individuals []Individual `json:"individuals"`
}

type CreateDeliveriesResponse struct {
	Status            domain.CampaignStatus `json:"status"`
	DeliveriesCreated int                   `json:"deliveries_created"`
}
