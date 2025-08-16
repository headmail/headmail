// Copyright 2025 JC-Lab
// SPDX-License-Identifier: AGPL-3.0-or-later

package dto

// CampaignStatsResponse is the typed response for campaign stats.
type CampaignStatsResponse struct {
	Labels []int64       `json:"labels"`
	Series []StatsSeries `json:"series"`
}

// StatsSeries represents per-campaign series of opens/clicks aligned to Labels.
type StatsSeries struct {
	CampaignID string  `json:"campaign_id"`
	Opens      []int64 `json:"opens"`
	Clicks     []int64 `json:"clicks"`
}
