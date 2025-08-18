// Copyright 2025 JC-Lab
// SPDX-License-Identifier: AGPL-3.0-or-later

package sqlite

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/headmail/headmail/pkg/queue"
	"github.com/headmail/headmail/pkg/repository"
	"gorm.io/gorm"
)

// queueRepository implements queue.Queue using SQLite (GORM).
type queueRepository struct {
	db *DB
}

func NewQueueRepository(db *DB) queue.Queue {
	return &queueRepository{
		db: db,
	}
}

func entityFromDomain(item *queue.QueueItem) *QueueItem {
	return &QueueItem{
		ID:         item.ID,
		Type:       item.Type,
		Payload:    JSON(item.Payload),
		UniqueKey:  item.UniqueKey,
		Status:     item.Status,
		ReservedBy: item.ReservedBy,
		ReservedAt: item.ReservedAt,
		CreatedAt:  item.CreatedAt,
	}
}

func domainFromEntity(e *QueueItem) *queue.QueueItem {
	return &queue.QueueItem{
		ID:         e.ID,
		Type:       e.Type,
		Payload:    json.RawMessage(e.Payload),
		UniqueKey:  e.UniqueKey,
		Status:     e.Status,
		ReservedBy: e.ReservedBy,
		ReservedAt: e.ReservedAt,
		CreatedAt:  e.CreatedAt,
	}
}

// Enqueue inserts a new item into the queue. If UniqueKey is set and an item
// with the same key already exists, the call is a no-op.
func (r *queueRepository) Enqueue(ctx context.Context, item *queue.QueueItem) error {
	db := extractTx(ctx, r.db.DB)

	// If unique key is provided, check for existence first to avoid unique constraint error.
	if item.UniqueKey != nil {
		var existing QueueItem
		err := db.WithContext(ctx).Where("unique_key = ?", *item.UniqueKey).First(&existing).Error
		if err == nil {
			// already exists -> ignore
			return nil
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
	}

	entity := entityFromDomain(item)
	if entity.Status == "" {
		entity.Status = queue.StatusPending
	}
	if entity.CreatedAt == 0 {
		entity.CreatedAt = time.Now().Unix()
	}
	return db.WithContext(ctx).Create(entity).Error
}

// Claim atomically reserves up to `limit` pending items that are ready and returns them.
func (r *queueRepository) Claim(ctx context.Context, workerID string, limit int) ([]*queue.QueueItem, error) {
	return repository.Transactional1[[]*queue.QueueItem](r.db, ctx, func(txCtx context.Context) ([]*queue.QueueItem, error) {
		tx := extractTx(ctx, r.db.DB)

		now := time.Now().Unix()
		var ids []string
		// select candidate ids
		if err := tx.WithContext(ctx).
			Model(&QueueItem{}).
			Where("status = ?", queue.StatusPending).
			Order("created_at ASC").
			Limit(limit).
			Pluck("id", &ids).Error; err != nil {
			return nil, err
		}

		if len(ids) == 0 {
			return []*queue.QueueItem{}, nil
		}

		// attempt to update selected rows to reserved atomically
		if err := tx.WithContext(ctx).
			Model(&QueueItem{}).
			Where("id IN ?", ids).
			Where("status = ?", queue.StatusPending).
			Updates(map[string]interface{}{
				"status":      queue.StatusReserved,
				"reserved_by": workerID,
				"reserved_at": now,
			}).Error; err != nil {
			return nil, err
		}

		// fetch the reserved items to return
		var entities []QueueItem
		if err := tx.WithContext(ctx).Where("id IN ?", ids).Find(&entities).Error; err != nil {
			return nil, err
		}

		var items []*queue.QueueItem
		for _, e := range entities {
			items = append(items, domainFromEntity(&e))
		}
		return items, nil
	})
}

// Ack deletes the queue item (marks done).
func (r *queueRepository) Ack(ctx context.Context, id string) error {
	tx := extractTx(ctx, r.db.DB)
	return tx.WithContext(ctx).
		Model(&QueueItem{}).
		Where("id = ?", id).
		Update("status", queue.StatusDone).
		Error
}

// Fail increments attempts and either reschedules or marks failed.
func (r *queueRepository) Fail(ctx context.Context, id string, reason string) error {
	tx := extractTx(ctx, r.db.DB)
	return tx.WithContext(ctx).
		Model(&QueueItem{}).
		Where("id = ?", id).
		Update("status", queue.StatusFailed).
		Error
}
