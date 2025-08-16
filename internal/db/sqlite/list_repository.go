// Copyright 2025 JC-Lab
// SPDX-License-Identifier: AGPL-3.0-or-later

package sqlite

import (
	"context"
	"encoding/json"
	"time"

	"github.com/headmail/headmail/pkg/domain"
	"github.com/headmail/headmail/pkg/repository"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// listRepository implements the repository.ListRepository interface.
type listRepository struct {
	db *DB
}

// NewListRepository creates a new list repository.
func NewListRepository(db *DB) repository.ListRepository {
	return &listRepository{db: db}
}

// domainToListEntity converts a domain List to a GORM List entity.
func domainToListEntity(d *domain.List) (*List, error) {
	tagsJSON, err := json.Marshal(d.Tags)
	if err != nil {
		return nil, err
	}
	return &List{
		ID:          d.ID,
		Name:        d.Name,
		Description: d.Description,
		Tags:        tagsJSON,
		CreatedAt:   d.CreatedAt,
		UpdatedAt:   d.UpdatedAt,
		DeletedAt:   d.DeletedAt,
	}, nil
}

// entityToListDomain converts a GORM List entity to a domain List.
func entityToListDomain(e *List) (*domain.List, error) {
	d := &domain.List{
		ID:          e.ID,
		Name:        e.Name,
		Description: e.Description,
		CreatedAt:   e.CreatedAt,
		UpdatedAt:   e.UpdatedAt,
		DeletedAt:   e.DeletedAt,
	}
	if err := json.Unmarshal(e.Tags, &d.Tags); err != nil {
		return nil, err
	}
	return d, nil
}

// Create saves a new list to the database.
func (r *listRepository) Create(ctx context.Context, list *domain.List) error {
	entity, err := domainToListEntity(list)
	if err != nil {
		return err
	}
	db := extractTx(ctx, r.db.DB)
	return db.WithContext(ctx).Create(entity).Error
}

func (r *listRepository) GetByID(ctx context.Context, id string) (*domain.List, error) {
	var entity List
	db := extractTx(ctx, r.db.DB)
	if err := db.WithContext(ctx).First(&entity, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return entityToListDomain(&entity)
}

func (r *listRepository) Update(ctx context.Context, list *domain.List) error {
	entity, err := domainToListEntity(list)
	if err != nil {
		return err
	}
	db := extractTx(ctx, r.db.DB)
	return db.WithContext(ctx).Save(entity).Error
}

func (r *listRepository) Delete(ctx context.Context, id string) error {
	db := extractTx(ctx, r.db.DB)
	return db.WithContext(ctx).Delete(&List{}, "id = ?", id).Error
}

func (r *listRepository) List(ctx context.Context, filter repository.ListFilter, pagination repository.Pagination) ([]*domain.List, int, error) {
	var entities []List
	var total int64

	db := extractTx(ctx, r.db.DB)
	query := db.WithContext(ctx).Model(&List{})

	if filter.Search != "" {
		query = query.Where("name LIKE ?", "%"+filter.Search+"%")
	}

	if len(filter.Tags) > 0 {
		for _, tag := range filter.Tags {
			query = query.Where("EXISTS (SELECT 1 FROM json_each(tags) WHERE value = ?)", tag)
		}
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (pagination.Page - 1) * pagination.Limit
	if err := query.Offset(offset).Limit(pagination.Limit).Find(&entities).Error; err != nil {
		return nil, 0, err
	}

	var lists []*domain.List
	for _, entity := range entities {
		list, err := entityToListDomain(&entity)
		if err != nil {
			return nil, 0, err
		}
		lists = append(lists, list)
	}

	return lists, int(total), nil
}

func (r *listRepository) GetSubscriberCount(ctx context.Context, listID string) (int, error) {
	var count int64
	db := extractTx(ctx, r.db.DB)
	if err := db.WithContext(ctx).Model(&SubscriberList{}).Where("list_id = ?", listID).Count(&count).Error; err != nil {
		return 0, err
	}
	return int(count), nil
}

func (r *listRepository) GetSubscribers(ctx context.Context) (chan *domain.Subscriber, error) {
	// This implementation is simplified and may not be suitable for large datasets.
	// A more robust implementation would use streaming or pagination.
	subscribersChan := make(chan *domain.Subscriber)

	go func() {
		defer close(subscribersChan)
		var entities []Subscriber
		db := extractTx(ctx, r.db.DB)
		if err := db.WithContext(ctx).Preload("Lists").Find(&entities).Error; err != nil {
			// In a real app, you'd handle this error more gracefully
			return
		}

		for _, entity := range entities {
			subscribersChan <- entityToSubscriberDomain(&entity)
		}
	}()

	return subscribersChan, nil
}

// AddSubscribers adds existing subscribers to the specified list.
// It will ignore duplicates.
func (r *listRepository) AddSubscribers(ctx context.Context, listID string, subscriberIDs []string) error {
	if len(subscriberIDs) == 0 {
		return nil
	}
	now := time.Now().Unix()
	db := extractTx(ctx, r.db.DB)
	return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, sid := range subscriberIDs {
			sl := &SubscriberList{
				SubscriberID: sid,
				ListID:       listID,
				Status:       domain.SubscriberListStatusConfirmed,
				CreatedAt:    now,
				UpdatedAt:    now,
			}
			if err := tx.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "subscriber_id"}, {Name: "list_id"}},
				DoNothing: true,
			}).Create(sl).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// RemoveSubscribers removes subscribers from the specified list.
func (r *listRepository) RemoveSubscribers(ctx context.Context, listID string, subscriberIDs []string) error {
	if len(subscriberIDs) == 0 {
		return nil
	}
	db := extractTx(ctx, r.db.DB)
	return db.WithContext(ctx).Where("list_id = ? AND subscriber_id IN ?", listID, subscriberIDs).Delete(&SubscriberList{}).Error
}

// ReplaceSubscribers replaces the subscribers of the given list with the provided list (atomic).
func (r *listRepository) ReplaceSubscribers(ctx context.Context, listID string, subscriberIDs []string) error {
	db := extractTx(ctx, r.db.DB)
	now := time.Now().Unix()
	return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Delete all current associations for the list
		if err := tx.Where("list_id = ?", listID).Delete(&SubscriberList{}).Error; err != nil {
			return err
		}
		// Insert new associations
		for _, sid := range subscriberIDs {
			sl := &SubscriberList{
				SubscriberID: sid,
				ListID:       listID,
				Status:       domain.SubscriberListStatusConfirmed,
				CreatedAt:    now,
				UpdatedAt:    now,
			}
			if err := tx.Create(sl).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
