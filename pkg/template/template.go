// Copyright 2025 JC-Lab
// SPDX-License-Identifier: AGPL-3.0-or-later

package template

import (
	"bytes"
	"text/template"

	"github.com/Masterminds/sprig/v3"
)

// Service provides template rendering capabilities.
type Service struct{}

// NewService creates a new Service.
func NewService() *Service {
	return &Service{}
}

// Render renders a template string with the given data.
func (s *Service) Render(templateStr string, data map[string]interface{}) (string, error) {
	// Add i18n function to the template context
	funcMap := sprig.TxtFuncMap()
	funcMap["i18n"] = func(data map[string]interface{}, messageID string) (string, error) {
		locale, ok := data["locale"].(string)
		if !ok {
			locale = "en"
		}
		// Extract i18n data from the main data map
		i18nData, ok := data["i18n"].(map[string]interface{})
		if !ok {
			return "", nil // Or return một error
		}

		// Extract locale-specific messages
		localeMessages, ok := i18nData[locale].(map[string]interface{})
		if !ok {
			return "", nil // Or return một error
		}

		// Simple key lookup for now. A more robust solution would handle nested keys.
		message, ok := localeMessages[messageID].(string)
		if !ok {
			// Fallback to default locale if specified
			// This part is simplified. A real implementation would need more logic.
			return messageID, nil
		}
		return message, nil
	}

	tmpl, err := template.New("email").Funcs(funcMap).Parse(templateStr)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}
