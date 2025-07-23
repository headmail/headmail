package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/headmail/headmail/pkg/domain"
	"github.com/headmail/headmail/pkg/repository"
)

// DeliveryServiceProvider defines the interface for a delivery service.
type DeliveryServiceProvider interface {
	CreateDelivery(ctx context.Context, delivery *domain.Delivery) error
	GetDelivery(ctx context.Context, id string) (*domain.Delivery, error)
	UpdateDelivery(ctx context.Context, delivery *domain.Delivery) error
	ListDeliveries(ctx context.Context, filter repository.DeliveryFilter, pagination repository.Pagination) ([]*domain.Delivery, int, error)
	GetDeliveriesByCampaign(ctx context.Context, campaignID string, pagination repository.Pagination) ([]*domain.Delivery, int, error)
	UpdateDeliveryStatus(ctx context.Context, id string, status string) error
}

// DeliveryService provides business logic for delivery management.
type DeliveryService struct {
	repo repository.DeliveryRepository
}

// NewDeliveryService creates a new DeliveryService.
func NewDeliveryService(db repository.DB) *DeliveryService {
	return &DeliveryService{
		repo: db.DeliveryRepository(),
	}
}

// CreateDelivery creates a new delivery.
func (s *DeliveryService) CreateDelivery(ctx context.Context, delivery *domain.Delivery) error {
	delivery.ID = uuid.New().String()
	return s.repo.Create(ctx, delivery)
}

// GetDelivery retrieves a delivery by its ID.
func (s *DeliveryService) GetDelivery(ctx context.Context, id string) (*domain.Delivery, error) {
	return s.repo.GetByID(ctx, id)
}

// UpdateDelivery updates an existing delivery.
func (s *DeliveryService) UpdateDelivery(ctx context.Context, delivery *domain.Delivery) error {
	return s.repo.Update(ctx, delivery)
}

// ListDeliveries lists all deliveries.
func (s *DeliveryService) ListDeliveries(ctx context.Context, filter repository.DeliveryFilter, pagination repository.Pagination) ([]*domain.Delivery, int, error) {
	return s.repo.List(ctx, filter, pagination)
}

// GetDeliveriesByCampaign retrieves all deliveries for a campaign.
func (s *DeliveryService) GetDeliveriesByCampaign(ctx context.Context, campaignID string, pagination repository.Pagination) ([]*domain.Delivery, int, error) {
	return s.repo.GetByCampaignID(ctx, campaignID, pagination)
}

// UpdateDeliveryStatus updates the status of a delivery.
func (s *DeliveryService) UpdateDeliveryStatus(ctx context.Context, id string, status string) error {
	return s.repo.UpdateStatus(ctx, id, status)
}
