// Copyright 2025 JC-Lab
// SPDX-License-Identifier: AGPL-3.0-or-later

package service

import (
	"context"
	"log"
	"time"

	"github.com/headmail/headmail/pkg/domain"
	"github.com/headmail/headmail/pkg/repository"
)

// TrackingServiceProvider defines the interface for a tracking service.
type TrackingServiceProvider interface {
	LogOpenEvent(ctx context.Context, deliveryID string, ua *string, ip *string) error
	LogClickEvent(ctx context.Context, deliveryID string, ua *string, ip *string, url string) error
}

// TrackingService provides business logic for tracking management.
type TrackingService struct {
	deliveryRepo repository.DeliveryRepository
	eventRepo    repository.EventRepository
	campaignRepo repository.CampaignRepository
}

// NewTrackingService creates a new TrackingService.
func NewTrackingService(db repository.DB) *TrackingService {
	return &TrackingService{
		deliveryRepo: db.DeliveryRepository(),
		eventRepo:    db.EventRepository(),
		campaignRepo: db.CampaignRepository(),
	}
}

func (s *TrackingService) LogOpenEvent(ctx context.Context, deliveryID string, ua *string, ip *string) error {
	now := time.Now()

	ev := &domain.DeliveryEvent{
		DeliveryID: deliveryID,
		EventType:  domain.EventTypeOpened,
		EventData:  map[string]interface{}{},
		UserAgent:  ua,
		IPAddress:  ip,
		CreatedAt:  now.Unix(),
	}

	// Atomically increment open count (repository will set OpenedAt on first open)
	isFirst, err := s.deliveryRepo.IncrementCount(ctx, deliveryID, domain.EventTypeOpened)
	if err != nil {
		log.Printf("tracking: failed to increment open count for %s: %v", deliveryID, err)
	}

	// If this is the first open for the delivery, increment campaign-level open counter.
	if isFirst {
		if d, err := s.deliveryRepo.GetByID(ctx, deliveryID); err == nil {
			if d.CampaignID != nil && *d.CampaignID != "" {
				if err := s.campaignRepo.IncrementStats(ctx, *d.CampaignID, 0, 0, 0, 1, 0, 0); err != nil {
					log.Printf("tracking: failed to increment campaign open count for campaign %s: %v", *d.CampaignID, err)
				}
			}
		} else {
			// non-fatal
			log.Printf("tracking: failed to load delivery %s for campaign increment: %v", deliveryID, err)
		}
	}

	// store event synchronously
	return s.eventRepo.Create(ctx, ev)
}

func (s *TrackingService) LogClickEvent(ctx context.Context, deliveryID string, ua *string, ip *string, url string) error {
	now := time.Now()

	ev := &domain.DeliveryEvent{
		DeliveryID: deliveryID,
		EventType:  domain.EventTypeClicked,
		EventData:  map[string]interface{}{"url": url},
		UserAgent:  ua,
		IPAddress:  ip,
		URL:        &url,
		CreatedAt:  now.Unix(),
	}

	// Atomically increment click count and detect if this is the first click for the delivery
	isFirst, err := s.deliveryRepo.IncrementCount(ctx, deliveryID, domain.EventTypeClicked)
	if err != nil {
		log.Printf("tracking: failed to increment click count for %s: %v", deliveryID, err)
	}

	// If this is the first click for the delivery, increment campaign-level click counter.
	if isFirst {
		if d, err := s.deliveryRepo.GetByID(ctx, deliveryID); err == nil {
			if d.CampaignID != nil && *d.CampaignID != "" {
				if err := s.campaignRepo.IncrementStats(ctx, *d.CampaignID, 0, 0, 0, 0, 1, 0); err != nil {
					log.Printf("tracking: failed to increment campaign click count for campaign %s: %v", *d.CampaignID, err)
				}
			}
		} else {
			// non-fatal
			log.Printf("tracking: failed to load delivery %s for campaign increment: %v", deliveryID, err)
		}
	}

	// record click synchronously
	return s.eventRepo.Create(ctx, ev)
}
