package service

import (
	"context"
	"github.com/headmail/headmail/pkg/api/admin/dto"
	"time"

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
	CreateDeliveries(ctx context.Context, campaignID string, req *dto.CreateDeliveriesRequest) (int, error)
}

// CampaignService provides business logic for campaign management.
type CampaignService struct {
	repo           repository.CampaignRepository
	listRepo       repository.ListRepository
	subscriberRepo repository.SubscriberRepository
	deliveryRepo   repository.DeliveryRepository
}

// NewCampaignService creates a new CampaignService.
func NewCampaignService(
	repo repository.CampaignRepository,
	listRepo repository.ListRepository,
	subscriberRepo repository.SubscriberRepository,
	deliveryRepo repository.DeliveryRepository,
) *CampaignService {
	return &CampaignService{
		repo:           repo,
		listRepo:       listRepo,
		subscriberRepo: subscriberRepo,
		deliveryRepo:   deliveryRepo,
	}
}

// CreateCampaign creates a new campaign.
func (s *CampaignService) CreateCampaign(ctx context.Context, campaign *domain.Campaign) error {
	// Generate a new UUID for the campaign
	campaign.ID = uuid.New().String()
	campaign.CreatedAt = time.Now().Unix()
	campaign.UpdatedAt = campaign.CreatedAt
	return s.repo.Create(ctx, campaign)
}

// GetCampaign retrieves a campaign by its ID.
func (s *CampaignService) GetCampaign(ctx context.Context, id string) (*domain.Campaign, error) {
	return s.repo.GetByID(ctx, id)
}

// UpdateCampaign updates an existing campaign.
func (s *CampaignService) UpdateCampaign(ctx context.Context, campaign *domain.Campaign) error {
	campaign.UpdatedAt = campaign.CreatedAt
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

// CreateDeliveries creates deliveries for a campaign.
func (s *CampaignService) CreateDeliveries(ctx context.Context, campaignID string, req *dto.CreateDeliveriesRequest) (int, error) {
	// 1. Get Campaign
	campaign, err := s.repo.GetByID(ctx, campaignID)
	if err != nil {
		return 0, err
	}

	// 2. Prepare deliveries
	deliveries := make([]*domain.Delivery, 0)
	processedEmails := make(map[string]bool)

	// TODO: Transaction

	// 3. Handle individuals
	if len(req.Individuals) > 0 {
		subscribersToUpsert := make([]*domain.Subscriber, len(req.Individuals))
		for i, individual := range req.Individuals {
			subscribersToUpsert[i] = &domain.Subscriber{
				Email:  individual.Email,
				Name:   individual.Name,
				Status: "active",
			}
			// TODO: Consider how to handle ListID for individual subscribers.
			// For now, we are not associating them with a list directly in this step.
		}
		if err := s.subscriberRepo.BulkUpsert(ctx, subscribersToUpsert); err != nil {
			return 0, err
		}

		for _, individual := range req.Individuals {
			if processedEmails[individual.Email] {
				continue
			}
			delivery := s.createDeliveryFromCampaign(campaign, individual.Name, individual.Email, req.ScheduledAt)
			delivery.Data = individual.Data
			delivery.Headers = individual.Headers
			deliveries = append(deliveries, delivery)
			processedEmails[individual.Email] = true
		}
	}

	// 4. Handle lists
	for _, listID := range req.Lists {
		// TODO: stream
		// We need to fetch all subscribers for the list.
		// Assuming a large list, this should be paginated, but for now, we'll fetch all.
		subscribers, _, err := s.subscriberRepo.List(ctx, repository.SubscriberFilter{
			ListID:     listID,
			Status:     domain.SubscriberStatusEnabled,
			ListStatus: domain.SubscriberListStatusConfirmed,
		}, repository.Pagination{Limit: -1})
		if err != nil {
			return 0, err
		}
		for _, subscriber := range subscribers {
			if processedEmails[subscriber.Email] {
				continue
			}
			delivery := s.createDeliveryFromCampaign(campaign, subscriber.Name, subscriber.Email, req.ScheduledAt)
			deliveries = append(deliveries, delivery)
			processedEmails[subscriber.Email] = true
		}
	}

	// 5. Create deliveries in DB
	// TODO: This should be a transactional operation.
	for _, delivery := range deliveries {
		if err := s.deliveryRepo.Create(ctx, delivery); err != nil {
			// TODO: Handle partial creation
			return 0, err
		}
	}

	return len(deliveries), nil
}

func (s *CampaignService) createDeliveryFromCampaign(campaign *domain.Campaign, name, email string, scheduledAt *int64) *domain.Delivery {
	deliveryID := uuid.New().String()
	campaignID := campaign.ID
	return &domain.Delivery{
		ID:          deliveryID,
		CampaignID:  &campaignID,
		Type:        "campaign",
		Status:      "scheduled",
		Name:        name,
		Email:       email,
		Subject:     campaign.Subject, // This will be templated later
		Data:        campaign.Data,
		Headers:     campaign.Headers,
		Tags:        campaign.Tags,
		ScheduledAt: scheduledAt,
	}
}
