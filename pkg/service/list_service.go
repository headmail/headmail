package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/headmail/headmail/pkg/domain"
	"github.com/headmail/headmail/pkg/repository"
)

// ListServiceProvider defines the interface for the list service.
type ListServiceProvider interface {
	CreateList(ctx context.Context, list *domain.List) error
}

// ListService provides business logic for list management.
type ListService struct {
	repo repository.ListRepository
}

// NewListService creates a new ListService.
func NewListService(repo repository.ListRepository) *ListService {
	return &ListService{repo: repo}
}

// CreateList creates a new mailing list.
// It assigns a new UUID if the ID is not provided.
func (s *ListService) CreateList(ctx context.Context, list *domain.List) error {
	if list.ID == "" {
		list.ID = uuid.NewString()
	}
	now := time.Now().Unix()
	list.CreatedAt = now
	list.UpdatedAt = now

	return s.repo.Create(ctx, list)
}

// TODO: Implement other list service methods (Get, Update, Delete, List)
