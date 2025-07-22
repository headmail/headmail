package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/headmail/headmail/pkg/domain"
	"github.com/headmail/headmail/pkg/repository"
)

// CampaignServiceProvider defines the interface for a campaign service.
type CampaignServiceProvider interface {
	CreateCampaign(ctx context.Context, campaign *domain.Campaign) error
	GetCampaign(ctx context.Context, id string) (*domain.Campaign, error)
	UpdateCampaign(ctx context.Context, campaign *domain.Campaign) error
	DeleteCampaign(ctx context.Context, id string) error
	ListCampaigns(ctx context.Context, filter repository.CampaignFilter, pagination repository.Pagination) ([]*domain.Campaign, int, error)
	UpdateCampaignStatus(ctx context.Context, id string, status domain.CampaignStatus) error
}

// CampaignService provides business logic for campaign management.
type CampaignService struct {
	repo repository.CampaignRepository
}

// NewCampaignService creates a new CampaignService.
func NewCampaignService(repo repository.CampaignRepository) *CampaignService {
	return &CampaignService{repo: repo}
}

// CreateCampaign creates a new campaign.
func (s *CampaignService) CreateCampaign(ctx context.Context, campaign *domain.Campaign) error {
	// Generate a new UUID for the campaign
	campaign.ID = uuid.New().String()
	return s.repo.Create(ctx, campaign)
}

// GetCampaign retrieves a campaign by its ID.
func (s *CampaignService) GetCampaign(ctx context.Context, id string) (*domain.Campaign, error) {
	return s.repo.GetByID(ctx, id)
}

// UpdateCampaign updates an existing campaign.
func (s *CampaignService) UpdateCampaign(ctx context.Context, campaign *domain.Campaign) error {
	return s.repo.Update(ctx, campaign)
}

// DeleteCampaign deletes a campaign by its ID.
func (s *CampaignService) DeleteCampaign(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

// ListCampaigns lists all campaigns.
func (s *CampaignService) ListCampaigns(ctx context.Context, filter repository.CampaignFilter, pagination repository.Pagination) ([]*domain.Campaign, int, error) {
	return s.repo.List(ctx, filter, pagination)
}

// UpdateCampaignStatus updates the status of a campaign.
func (s *CampaignService) UpdateCampaignStatus(ctx context.Context, id string, status domain.CampaignStatus) error {
	return s.repo.UpdateStatus(ctx, id, status)
}
