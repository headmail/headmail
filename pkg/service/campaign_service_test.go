// Copyright 2025 JC-Lab
// SPDX-License-Identifier: AGPL-3.0-or-later

package service

import (
	"context"
	"testing"

	"github.com/headmail/headmail/pkg/domain"
	"github.com/headmail/headmail/pkg/template"
	"github.com/stretchr/testify/assert"
)

func TestCreateDeliveryFromCampaign_RendersTemplates_UsingTestify(t *testing.T) {
	svc := &CampaignService{
		templateService: template.NewService(),
	}

	campaign := &domain.Campaign{
		ID:           "camp-1",
		Subject:      "Hello {{ .name }}",
		TemplateHTML: "<p>Company: {{ .company }} - Hi {{ .name }}</p>",
		TemplateText: "Company: {{ .company }} - Hi {{ .name }}",
		Data: map[string]interface{}{
			"company": "Acme",
		},
		Headers: map[string]string{
			"X-Base": "base",
		},
		Tags: []string{"tag1"},
	}

	individualData := map[string]interface{}{
		"locale": "en",
	}
	individualHeaders := map[string]string{
		"X-User": "user-1",
	}

	delivery, err := svc.createDeliveryFromCampaign(context.Background(), campaign, "Bob", "bob@example.com", nil, individualData, individualHeaders)
	assert.NoError(t, err)

	assert.Equal(t, "bob@example.com", delivery.Email)
	assert.Equal(t, "Bob", delivery.Name)
	assert.Equal(t, "Hello Bob", delivery.Subject)

	assert.Contains(t, delivery.BodyHTML, "Hi Bob")
	assert.Contains(t, delivery.BodyHTML, "Company: Acme")
	assert.Contains(t, delivery.BodyText, "Hi Bob")
	assert.Contains(t, delivery.BodyText, "Company: Acme")

	// Headers should contain both campaign header and individual header
	assert.Equal(t, "base", delivery.Headers["X-Base"])
	assert.Equal(t, "user-1", delivery.Headers["X-User"])

	// Data should include deliveryId, name and email
	_, ok := delivery.Data["deliveryId"]
	assert.True(t, ok, "delivery.Data missing deliveryId")

	assert.Equal(t, "Bob", delivery.Data["name"])
	assert.Equal(t, "bob@example.com", delivery.Data["email"])
}
