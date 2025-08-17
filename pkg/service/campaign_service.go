// Copyright 2025 JC-Lab
// SPDX-License-Identifier: AGPL-3.0-or-later

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
	// CreateCampaign creates a campaign. If campaign.ID is empty a new ID will be generated.
	// If campaign.ID is provided and upsert is true, an existing campaign with the same ID
	// will be updated; if upsert is false and the ID already exists, an error will be returned.
	CreateCampaign(ctx context.Context, campaign *domain.Campaign, upsert bool) error
	GetCampaign(ctx context.Context, id string) (*domain.Campaign, error)
	UpdateCampaign(ctx context.Context, campaign *domain.Campaign) error
	DeleteCampaign(ctx context.Context, id string) error
	ListCampaigns(ctx context.Context, filter repository.CampaignFilter, pagination repository.Pagination) ([]*domain.Campaign, int, error)
	UpdateCampaignStatus(ctx context.Context, id string, status domain.CampaignStatus) error
	// ReleaseDueDeliveries sets SendScheduledAt for deliveries whose campaign scheduled time has arrived.
	ReleaseDueDeliveries(ctx context.Context, now int64) (int, error)
	CreateDeliveries(ctx context.Context, campaignID string, req *dto.CreateDeliveriesRequest) (int, error)

	// GetCampaignStats returns time-bucketed opens and clicks for given campaign IDs.
	// granularity: "hour" or "day"
	GetCampaignStats(ctx context.Context, campaignIDs []string, from time.Time, to time.Time, granularity string) (*dto.CampaignStatsResponse, error)
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

// CreateCampaign creates a new campaign or upserts when requested.
func (s *CampaignService) CreateCampaign(ctx context.Context, campaign *domain.Campaign, upsert bool) error {
	// If no ID provided, generate one and create.
	if campaign.ID == "" {
		campaign.ID = uuid.NewString()
		now := time.Now().Unix()
		campaign.CreatedAt = now
		campaign.UpdatedAt = now
		return s.repo.Create(ctx, campaign)
	}

	// ID provided: check existence
	existing, err := s.repo.GetByID(ctx, campaign.ID)
	if err != nil {
		// If not found, create new
		if _, ok := err.(*repository.ErrNotFound); ok {
			now := time.Now().Unix()
			campaign.CreatedAt = now
			campaign.UpdatedAt = now
			return s.repo.Create(ctx, campaign)
		}
		// other error
		return err
	}

	// existing found
	if !upsert {
		// return unique constraint error to indicate conflict
		return &repository.ErrUniqueConstraintFailed{Cause: nil}
	}

	// upsert: preserve CreatedAt, set UpdatedAt and update
	campaign.CreatedAt = existing.CreatedAt
	campaign.UpdatedAt = time.Now().Unix()
	return s.repo.Update(ctx, campaign)
}

// GetCampaign retrieves a campaign by its ID.
func (s *CampaignService) GetCampaign(ctx context.Context, id string) (*domain.Campaign, error) {
	return s.repo.GetByID(ctx, id)
}

// UpdateCampaign updates an existing campaign.
func (s *CampaignService) UpdateCampaign(ctx context.Context, campaign *domain.Campaign) error {
	// When updating, set UpdatedAt to current time
	campaign.UpdatedAt = time.Now().Unix()
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

	// Use TemplateHTML/Text from campaign if present; otherwise, if TemplateID provided fetch missing parts from template.
	if (campaign.TemplateHTML == "" || campaign.TemplateText == "") && campaign.TemplateID != nil && *campaign.TemplateID != "" {
		tmpl, err := s.templateRepo.GetByID(ctx, *campaign.TemplateID)
		if err != nil {
			return 0, err
		}
		if campaign.TemplateHTML == "" {
			campaign.TemplateHTML = tmpl.BodyHTML
		}
		if campaign.TemplateText == "" {
			campaign.TemplateText = tmpl.BodyText
		}
		if campaign.Subject == "" {
			campaign.Subject = tmpl.Subject
		}
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

				delivery, err := s.createDeliveryFromCampaign(txCtx, campaign, individual.Name, individual.Email, individual.Data, individual.Headers)
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
				delivery, err := s.createDeliveryFromCampaign(txCtx, campaign, subscriber.Name, subscriber.Email, nil, nil)
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

		// Increment campaign recipient count atomically by number of unique deliveries created.
		// deliveries are deduplicated by email earlier in this function, so len(deliveries)
		// represents unique recipients for this call.
		if len(deliveries) > 0 {
			if err := s.repo.IncrementStats(txCtx, campaign.ID, len(deliveries), 0, 0, 0, 0, 0); err != nil {
				return 0, err
			}
		}

		return len(deliveries), nil
	})
}

func (s *CampaignService) createDeliveryFromCampaign(ctx context.Context, campaign *domain.Campaign, name, email string, individualData map[string]interface{}, individualHeaders map[string]string) (*domain.Delivery, error) {
	var err error

	delivery := &domain.Delivery{
		ID:    uuid.New().String(),
		Type:  domain.DeliveryTypeCampaign,
		Name:  name,
		Email: email,
		Tags:  campaign.Tags,
	}
	delivery.CampaignID = &campaign.ID

	// Determine scheduledAt based on campaign status:
	// - sending: enqueue immediately => scheduledAt = nil
	// - scheduled: use campaign.ScheduledAt
	// - other (draft/paused/...): create delivery but do not enqueue; to prevent immediate enqueue
	//   we set scheduledAt to a far future timestamp when campaign.ScheduledAt is nil.
	if campaign.Status == domain.CampaignStatusSending || campaign.Status == domain.CampaignStatusSent {
		// immediate enqueue
		delivery.Status = domain.DeliveryStatusScheduled
		delivery.ScheduledAt = nil
	} else if campaign.Status == domain.CampaignStatusScheduled {
		// inherit campaign schedule (may be in future)
		delivery.Status = domain.DeliveryStatusScheduled
		delivery.ScheduledAt = campaign.ScheduledAt
	} else {
		delivery.Status = domain.DeliveryStatusIdle
		delivery.ScheduledAt = nil
	}

	// Prepare data for templating
	delivery.Data = make(map[string]interface{})
	for k, v := range campaign.Data {
		delivery.Data[k] = v
	}
	if individualData != nil {
		for k, v := range individualData {
			delivery.Data[k] = v
		}
	}
	delivery.Data["deliveryId"] = delivery.ID
	delivery.Data["name"] = name
	delivery.Data["email"] = email

	// Render subject
	delivery.Subject, err = s.templateService.Render(campaign.Subject, delivery.Data)
	if err != nil {
		return nil, err
	}

	// Render HTML content
	delivery.BodyHTML, err = s.templateService.Render(campaign.TemplateHTML, delivery.Data)
	if err != nil {
		return nil, err
	}

	// Render text content
	delivery.BodyText, err = s.templateService.Render(campaign.TemplateText, delivery.Data)
	if err != nil {
		return nil, err
	}

	// Prepare headers
	delivery.Headers = make(map[string]string)
	for k, v := range campaign.Headers {
		delivery.Headers[k] = v
	}
	if individualHeaders != nil {
		for k, v := range individualHeaders {
			delivery.Headers[k] = v
		}
	}

	return delivery, nil
}

// ReleaseDueDeliveries finds campaigns whose ScheduledAt <= now and, for each campaign,
// sets deliveries' send time in a single repository update (atomic per campaign).
// Returns total number of deliveries updated across all campaigns.
func (s *CampaignService) ReleaseDueDeliveries(ctx context.Context, now int64) (int, error) {
	campaigns, err := s.repo.ListScheduledBefore(ctx, now)
	if err != nil {
		return 0, err
	}

	total := 0
	for _, c := range campaigns {
		// run per-campaign update inside a transaction to ensure atomicity
		if err := repository.Transactional0(s.db, ctx, func(txCtx context.Context) error {
			if err := s.repo.UpdateStatus(txCtx, c.ID, domain.CampaignStatusSending); err != nil {
				return err
			}

			updated, err := s.db.DeliveryRepository().UpdateSendScheduledByCampaign(txCtx, c.ID, now)
			if err != nil {
				return err
			}
			total += updated
			return nil
		}); err != nil {
			return total, err
		}
	}

	return total, nil
}

// GetCampaignStats aggregates open and click counts per time bucket for the given campaigns.
// Returns a map with keys:
// - "labels": []int64 (bucket start unix timestamps)
// - "series": []{ "campaign_id": string, "opens": []int64, "clicks": []int64 }
func (s *CampaignService) GetCampaignStats(ctx context.Context, campaignIDs []string, from time.Time, to time.Time, granularity string) (*dto.CampaignStatsResponse, error) {
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
	// sort labels ascending
	for i := 0; i < len(labels); i++ {
		for j := i + 1; j < len(labels); j++ {
			if labels[i] > labels[j] {
				labels[i], labels[j] = labels[j], labels[i]
			}
		}
	}

	series := make([]dto.StatsSeries, 0, len(campaignIDs))
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
		series = append(series, dto.StatsSeries{
			CampaignID: cid,
			Opens:      openSeries,
			Clicks:     clickSeries,
		})
	}

	return &dto.CampaignStatsResponse{
		Labels: labels,
		Series: series,
	}, nil
}
