package service

import (
	"context"
	"time"

	"github.com/headmail/headmail/pkg/template"

	"github.com/google/uuid"
	"github.com/headmail/headmail/pkg/api/admin/dto"
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

	// GetCampaignStats returns time-bucketed opens and clicks for given campaign IDs.
	// granularity: "hour" or "day"
	GetCampaignStats(ctx context.Context, campaignIDs []string, from time.Time, to time.Time, granularity string) (map[string]interface{}, error)
}

// CampaignService provides business logic for campaign management.
type CampaignService struct {
	db              repository.DB
	repo            repository.CampaignRepository
	listRepo        repository.ListRepository
	subscriberRepo  repository.SubscriberRepository
	templateRepo    repository.TemplateRepository
	deliveryService DeliveryServiceProvider
	templateService *template.Service
}

// NewCampaignService creates a new CampaignService.
func NewCampaignService(
	db repository.DB,
	deliveryService DeliveryServiceProvider,
	templateService *template.Service,
) *CampaignService {
	return &CampaignService{
		db:              db,
		repo:            db.CampaignRepository(),
		listRepo:        db.ListRepository(),
		subscriberRepo:  db.SubscriberRepository(),
		deliveryService: deliveryService,
		templateRepo:    db.TemplateRepository(),
		templateService: templateService,
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

	// If TemplateID is provided, fetch the template and override HTML/Text
	if campaign.TemplateID != nil && *campaign.TemplateID != "" {
		tmpl, err := s.templateRepo.GetByID(ctx, *campaign.TemplateID)
		if err != nil {
			return 0, err
		}
		campaign.TemplateHTML = tmpl.BodyHTML
		campaign.TemplateText = tmpl.BodyText
	}

	// 2. Prepare deliveries
	deliveries := make([]*domain.Delivery, 0)
	processedEmails := make(map[string]bool)

	return repository.Transactional1(s.db, ctx, func(txCtx context.Context) (int, error) {
		// 3. Handle individuals
		if len(req.Individuals) > 0 {
			subscribersToUpsert := make([]*domain.Subscriber, len(req.Individuals))
			for i, individual := range req.Individuals {
				subscribersToUpsert[i] = &domain.Subscriber{
					Email:  individual.Email,
					Name:   individual.Name,
					Status: domain.SubscriberStatusEnabled,
				}
			}
			if err := s.subscriberRepo.BulkUpsert(txCtx, subscribersToUpsert); err != nil {
				return 0, err
			}

			for _, individual := range req.Individuals {
				if processedEmails[individual.Email] {
					continue
				}

				delivery, err := s.createDeliveryFromCampaign(txCtx, campaign, individual.Name, individual.Email, req.ScheduledAt, individual.Data, individual.Headers)
				if err != nil {
					return 0, err // Or handle error more gracefully
				}
				deliveries = append(deliveries, delivery)
				processedEmails[individual.Email] = true
			}
		}

		// 4. Handle lists
		for _, listID := range req.Lists {
			// TODO: stream
			// We need to fetch all subscribers for the list.
			// Assuming a large list, this should be paginated, but for now, we'll fetch all.
			subscribers, err := s.subscriberRepo.ListStream(txCtx, repository.SubscriberFilter{
				ListID:     listID,
				Status:     domain.SubscriberStatusEnabled,
				ListStatus: domain.SubscriberListStatusConfirmed,
			})
			if err != nil {
				return 0, err
			}
			for subscriber := range subscribers {
				if processedEmails[subscriber.Email] {
					continue
				}
				delivery, err := s.createDeliveryFromCampaign(txCtx, campaign, subscriber.Name, subscriber.Email, req.ScheduledAt, nil, nil)
				if err != nil {
					return 0, err // Or handle error more gracefully
				}
				deliveries = append(deliveries, delivery)
				processedEmails[subscriber.Email] = true
			}
		}

		for _, delivery := range deliveries {
			if err := s.deliveryService.CreateDelivery(txCtx, delivery); err != nil {
				return 0, err
			}
		}

		return len(deliveries), nil
	})
}

func (s *CampaignService) createDeliveryFromCampaign(ctx context.Context, campaign *domain.Campaign, name, email string, scheduledAt *int64, individualData map[string]interface{}, individualHeaders map[string]string) (*domain.Delivery, error) {
	deliveryID := uuid.New().String()
	campaignID := campaign.ID

	// Prepare data for templating
	templateData := make(map[string]interface{})
	for k, v := range campaign.Data {
		templateData[k] = v
	}
	if individualData != nil {
		for k, v := range individualData {
			templateData[k] = v
		}
	}
	templateData["deliveryId"] = deliveryID
	templateData["name"] = name
	templateData["email"] = email

	// Render subject
	renderedSubject, err := s.templateService.Render(campaign.Subject, templateData)
	if err != nil {
		return nil, err
	}

	// Render HTML content
	renderedHTML, err := s.templateService.Render(campaign.TemplateHTML, templateData)
	if err != nil {
		return nil, err
	}

	// Render text content
	renderedText, err := s.templateService.Render(campaign.TemplateText, templateData)
	if err != nil {
		return nil, err
	}

	// Prepare headers
	finalHeaders := make(map[string]string)
	for k, v := range campaign.Headers {
		finalHeaders[k] = v
	}
	if individualHeaders != nil {
		for k, v := range individualHeaders {
			finalHeaders[k] = v
		}
	}

	delivery := &domain.Delivery{
		ID:              deliveryID,
		CampaignID:      &campaignID,
		Type:            domain.DeliveryTypeCampaign,
		Status:          domain.DeliveryStatusScheduled,
		Name:            name,
		Email:           email,
		Subject:         renderedSubject,
		BodyHTML:        renderedHTML,
		BodyText:        renderedText,
		Data:            templateData,
		Headers:         finalHeaders,
		Tags:            campaign.Tags,
		ScheduledAt:     scheduledAt,
		SendScheduledAt: scheduledAt,
	}
	return delivery, nil
}

// GetCampaignStats aggregates open and click counts per time bucket for the given campaigns.
// Returns a map with keys:
// - "labels": []int64 (bucket start unix timestamps)
// - "series": []{ "campaign_id": string, "opens": []int64, "clicks": []int64 }
func (s *CampaignService) GetCampaignStats(ctx context.Context, campaignIDs []string, from time.Time, to time.Time, granularity string) (map[string]interface{}, error) {
	fromTs := from.Unix()
	toTs := to.Unix()

	opens, err := s.db.EventRepository().CountByCampaignAndRangeByType(ctx, campaignIDs, string(domain.EventTypeOpened), fromTs, toTs, granularity)
	if err != nil {
		return nil, err
	}
	clicks, err := s.db.EventRepository().CountByCampaignAndRangeByType(ctx, campaignIDs, string(domain.EventTypeClicked), fromTs, toTs, granularity)
	if err != nil {
		return nil, err
	}

	// collect all bucket timestamps
	bucketSet := map[int64]struct{}{}
	for _, m := range []map[string]map[int64]int64{opens, clicks} {
		for _, cmap := range m {
			for b := range cmap {
				bucketSet[b] = struct{}{}
			}
		}
	}
	// produce sorted labels
	labels := make([]int64, 0, len(bucketSet))
	for b := range bucketSet {
		labels = append(labels, b)
	}
	// sort
	for i := 0; i < len(labels); i++ {
		for j := i + 1; j < len(labels); j++ {
			if labels[i] > labels[j] {
				labels[i], labels[j] = labels[j], labels[i]
			}
		}
	}

	series := make([]map[string]interface{}, 0, len(campaignIDs))
	for _, cid := range campaignIDs {
		openSeries := make([]int64, len(labels))
		clickSeries := make([]int64, len(labels))
		if m, ok := opens[cid]; ok {
			for i, b := range labels {
				if v, ok2 := m[b]; ok2 {
					openSeries[i] = v
				} else {
					openSeries[i] = 0
				}
			}
		}
		if m, ok := clicks[cid]; ok {
			for i, b := range labels {
				if v, ok2 := m[b]; ok2 {
					clickSeries[i] = v
				} else {
					clickSeries[i] = 0
				}
			}
		}
		series = append(series, map[string]interface{}{
			"campaign_id": cid,
			"opens":       openSeries,
			"clicks":      clickSeries,
		})
	}

	return map[string]interface{}{
		"labels": labels,
		"series": series,
	}, nil
}
