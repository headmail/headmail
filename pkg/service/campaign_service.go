package service

import (
	"github.com/headmail/headmail/pkg/repository"
)

// CampaignService provides business logic for campaign management.
type CampaignService struct {
	repo repository.CampaignRepository
}

// NewCampaignService creates a new CampaignService.
func NewCampaignService(repo repository.CampaignRepository) *CampaignService {
	return &CampaignService{repo: repo}
}

// TODO: Implement campaign service methods
