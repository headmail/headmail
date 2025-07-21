package service

import (
	"github.com/headmail/headmail/pkg/repository"
)

// DeliveryService provides business logic for delivery management.
type DeliveryService struct {
	repo repository.DeliveryRepository
}

// NewDeliveryService creates a new DeliveryService.
func NewDeliveryService(repo repository.DeliveryRepository) *DeliveryService {
	return &DeliveryService{repo: repo}
}

// TODO: Implement delivery service methods
