// Copyright 2025 JC-Lab
// SPDX-License-Identifier: AGPL-3.0-or-later

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
	GetList(ctx context.Context, id string) (*domain.List, error)
	UpdateList(ctx context.Context, list *domain.List) error
	DeleteList(ctx context.Context, id string) error
	ListLists(ctx context.Context, filter repository.ListFilter, pagination repository.Pagination) ([]*domain.List, int, error)

	GetSubscriberCount(ctx context.Context, listID string) (int, error)
	AddSubscribers(ctx context.Context, subscribers []*domain.Subscriber) error
	GetSubscriber(ctx context.Context, id string) (*domain.Subscriber, error)
	UpdateSubscriber(ctx context.Context, subscriber *domain.Subscriber) error
	DeleteSubscriber(ctx context.Context, id string) error
	ListSubscribers(ctx context.Context, filter repository.SubscriberFilter, pagination repository.Pagination) ([]*domain.Subscriber, int, error)

	// Patch subscribers in a list: add/remove sets of subscriber IDs
	PatchSubscribersInList(ctx context.Context, listID string, add []string, remove []string) error

	// Replace subscribers in a list atomically
	ReplaceSubscribersInList(ctx context.Context, listID string, subscriberIDs []string) error
}

// ListService provides business logic for list management.
type ListService struct {
	db             repository.DB
	listRepo       repository.ListRepository
	subscriberRepo repository.SubscriberRepository
}

// NewListService creates a new ListService.
func NewListService(db repository.DB) *ListService {
	return &ListService{
		db:             db,
		listRepo:       db.ListRepository(),
		subscriberRepo: db.SubscriberRepository(),
	}
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

	return s.listRepo.Create(ctx, list)
}

// GetList retrieves a list by its ID.
func (s *ListService) GetList(ctx context.Context, id string) (*domain.List, error) {
	return s.listRepo.GetByID(ctx, id)
}

// UpdateList updates an existing list.
func (s *ListService) UpdateList(ctx context.Context, list *domain.List) error {
	list.UpdatedAt = time.Now().Unix()
	return s.listRepo.Update(ctx, list)
}

// DeleteList deletes a list by its ID.
func (s *ListService) DeleteList(ctx context.Context, id string) error {
	return s.listRepo.Delete(ctx, id)
}

// ListLists lists all lists.
func (s *ListService) ListLists(ctx context.Context, filter repository.ListFilter, pagination repository.Pagination) ([]*domain.List, int, error) {
	return s.listRepo.List(ctx, filter, pagination)
}

// GetSubscriberCount returns the number of subscribers in a list.
func (s *ListService) GetSubscriberCount(ctx context.Context, listID string) (int, error) {
	return s.listRepo.GetSubscriberCount(ctx, listID)
}

// AddSubscribers adds subscribers to a list.
// It sets CreatedAt and UpdatedAt timestamps and performs bulk upsert within a transaction.
func (s *ListService) AddSubscribers(ctx context.Context, subscribers []*domain.Subscriber) error {
	if len(subscribers) == 0 {
		// nothing to do
		return nil
	}

	now := time.Now().Unix()
	for _, sub := range subscribers {
		// ensure timestamps are set for new/updated subscribers
		sub.CreatedAt = now
		sub.UpdatedAt = now
	}

	return repository.Transactional0(s.db, ctx, func(txCtx context.Context) error {
		return s.subscriberRepo.BulkUpsert(txCtx, subscribers)
	})
}

// GetSubscriber retrieves a subscriber by its ID.
func (s *ListService) GetSubscriber(ctx context.Context, id string) (*domain.Subscriber, error) {
	return s.subscriberRepo.GetByID(ctx, id)
}

// UpdateSubscriber updates an existing subscriber.
func (s *ListService) UpdateSubscriber(ctx context.Context, subscriber *domain.Subscriber) error {
	subscriber.UpdatedAt = time.Now().Unix()
	return s.subscriberRepo.Update(ctx, subscriber)
}

// DeleteSubscriber deletes a subscriber by its ID.
func (s *ListService) DeleteSubscriber(ctx context.Context, id string) error {
	return s.subscriberRepo.Delete(ctx, id)
}

// ListSubscribers lists all subscribers for a list.
func (s *ListService) ListSubscribers(ctx context.Context, filter repository.SubscriberFilter, pagination repository.Pagination) ([]*domain.Subscriber, int, error) {
	return s.subscriberRepo.List(ctx, filter, pagination)
}

// PatchSubscribersInList adds/removes subscribers from a list within a transaction.
func (s *ListService) PatchSubscribersInList(ctx context.Context, listID string, add []string, remove []string) error {
	return repository.Transactional0(s.db, ctx, func(txCtx context.Context) error {
		// Remove first, then add to avoid unique conflicts and to implement requested semantics.
		if len(remove) > 0 {
			if err := s.listRepo.RemoveSubscribers(txCtx, listID, remove); err != nil {
				return err
			}
		}
		if len(add) > 0 {
			if err := s.listRepo.AddSubscribers(txCtx, listID, add); err != nil {
				return err
			}
		}
		return nil
	})
}

// ReplaceSubscribersInList replaces all subscribers for a list atomically.
func (s *ListService) ReplaceSubscribersInList(ctx context.Context, listID string, subscriberIDs []string) error {
	return repository.Transactional0(s.db, ctx, func(txCtx context.Context) error {
		return s.listRepo.ReplaceSubscribers(txCtx, listID, subscriberIDs)
	})
}
