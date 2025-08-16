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
}

// NewTrackingService creates a new TrackingService.
func NewTrackingService(db repository.DB) *TrackingService {
	return &TrackingService{
		deliveryRepo: db.DeliveryRepository(),
		eventRepo:    db.EventRepository(),
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
	if err := s.deliveryRepo.IncrementCount(ctx, deliveryID, domain.EventTypeOpened); err != nil {
		log.Printf("tracking: failed to increment open count for %s: %v", deliveryID, err)
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

	// Atomically increment click count
	if err := s.deliveryRepo.IncrementCount(ctx, deliveryID, domain.EventTypeClicked); err != nil {
		log.Printf("tracking: failed to increment click count for %s: %v", deliveryID, err)
	}

	// record click synchronously
	return s.eventRepo.Create(ctx, ev)
}
