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
)

type deliveryRepository struct {
	db *DB
}

func NewDeliveryRepository(db *DB) repository.DeliveryRepository {
	return &deliveryRepository{db: db}
}

func domainToDeliveryEntity(d *domain.Delivery) (*Delivery, error) {
	dataJSON, err := json.Marshal(d.Data)
	if err != nil {
		return nil, err
	}
	headersJSON, err := json.Marshal(d.Headers)
	if err != nil {
		return nil, err
	}
	tagsJSON, err := json.Marshal(d.Tags)
	if err != nil {
		return nil, err
	}

	return &Delivery{
		ID:            d.ID,
		CampaignID:    d.CampaignID,
		Type:          d.Type,
		Status:        d.Status,
		Name:          d.Name,
		Email:         d.Email,
		Subject:       d.Subject,
		BodyHTML:      d.BodyHTML,
		BodyText:      d.BodyText,
		MessageID:     d.MessageID,
		Data:          dataJSON,
		Headers:       headersJSON,
		Tags:          tagsJSON,
		CreatedAt:     d.CreatedAt,
		ScheduledAt:   d.ScheduledAt,
		Attempts:      d.Attempts,
		SentAt:        d.SentAt,
		OpenedAt:      d.OpenedAt,
		FailedAt:      d.FailedAt,
		FailureReason: d.FailureReason,
		OpenCount:     d.OpenCount,
		ClickCount:    d.ClickCount,
		BounceCount:   d.BounceCount,
	}, nil
}

func entityToDeliveryDomain(e *Delivery) (*domain.Delivery, error) {
	var data map[string]interface{}
	if err := json.Unmarshal(e.Data, &data); err != nil {
		return nil, err
	}
	var headers map[string]string
	if err := json.Unmarshal(e.Headers, &headers); err != nil {
		return nil, err
	}
	var tags []string
	if e.Tags != nil {
		if err := json.Unmarshal(e.Tags, &tags); err != nil {
			return nil, err
		}
	}

	return &domain.Delivery{
		ID:            e.ID,
		CampaignID:    e.CampaignID,
		Type:          e.Type,
		Status:        e.Status,
		Name:          e.Name,
		Email:         e.Email,
		Subject:       e.Subject,
		BodyHTML:      e.BodyHTML,
		BodyText:      e.BodyText,
		MessageID:     e.MessageID,
		Data:          data,
		Headers:       headers,
		Tags:          tags,
		CreatedAt:     e.CreatedAt,
		ScheduledAt:   e.ScheduledAt,
		Attempts:      e.Attempts,
		SentAt:        e.SentAt,
		OpenedAt:      e.OpenedAt,
		FailedAt:      e.FailedAt,
		FailureReason: e.FailureReason,
		OpenCount:     e.OpenCount,
		ClickCount:    e.ClickCount,
		BounceCount:   e.BounceCount,
	}, nil
}

func (r *deliveryRepository) Create(ctx context.Context, delivery *domain.Delivery) error {
	entity, err := domainToDeliveryEntity(delivery)
	if err != nil {
		return err
	}
	db := extractTx(ctx, r.db.DB)
	return db.WithContext(ctx).Create(entity).Error
}

func (r *deliveryRepository) BulkCreate(ctx context.Context, deliveries []*domain.Delivery) error {
	var entities []*Delivery
	for _, delivery := range deliveries {
		entity, err := domainToDeliveryEntity(delivery)
		if err != nil {
			return err
		}
		entities = append(entities, entity)
	}
	db := extractTx(ctx, r.db.DB)
	return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, entity := range entities {
			if err := tx.Create(entity).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *deliveryRepository) GetByID(ctx context.Context, id string) (*domain.Delivery, error) {
	var entity Delivery
	db := extractTx(ctx, r.db.DB)
	if err := db.WithContext(ctx).First(&entity, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return entityToDeliveryDomain(&entity)
}

func (r *deliveryRepository) Update(ctx context.Context, delivery *domain.Delivery) error {
	entity, err := domainToDeliveryEntity(delivery)
	if err != nil {
		return err
	}
	db := extractTx(ctx, r.db.DB)
	return db.WithContext(ctx).Save(entity).Error
}

func (r *deliveryRepository) List(ctx context.Context, filter repository.DeliveryFilter, pagination repository.Pagination) ([]*domain.Delivery, int, error) {
	var entities []Delivery
	var total int64

	db := extractTx(ctx, r.db.DB)
	query := db.WithContext(ctx).Model(&Delivery{})

	if filter.CampaignID != "" {
		query = query.Where("campaign_id = ?", filter.CampaignID)
	}
	if filter.Type != "" {
		query = query.Where("type = ?", filter.Type)
	}
	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}
	if filter.Email != "" {
		query = query.Where("email = ?", filter.Email)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (pagination.Page - 1) * pagination.Limit
	if err := query.Order("created_at DESC").Offset(offset).Limit(pagination.Limit).Find(&entities).Error; err != nil {
		return nil, 0, err
	}

	var deliveries []*domain.Delivery
	for _, entity := range entities {
		delivery, err := entityToDeliveryDomain(&entity)
		if err != nil {
			return nil, 0, err
		}
		deliveries = append(deliveries, delivery)
	}

	return deliveries, int(total), nil
}

func (r *deliveryRepository) GetByCampaignID(ctx context.Context, campaignID string, pagination repository.Pagination) ([]*domain.Delivery, int, error) {
	return r.List(ctx, repository.DeliveryFilter{CampaignID: campaignID}, pagination)
}

// ListScheduledBefore returns deliveries whose scheduled_at is non-null and <= ts.
// It limits results to `limit` items (if limit <= 0 a default is used).
func (r *deliveryRepository) ListScheduledBefore(ctx context.Context, ts int64, limit int) ([]*domain.Delivery, error) {
	if limit <= 0 {
		limit = 1000
	}

	var entities []Delivery
	db := extractTx(ctx, r.db.DB)
	query := db.WithContext(ctx).Model(&Delivery{}).
		Where("scheduled_at IS NOT NULL AND scheduled_at <= ? AND status = ?", ts, domain.DeliveryStatusScheduled).
		Order("scheduled_at ASC, id ASC").
		Limit(limit)

	if err := query.Find(&entities).Error; err != nil {
		return nil, err
	}

	var deliveries []*domain.Delivery
	for _, e := range entities {
		d, err := entityToDeliveryDomain(&e)
		if err != nil {
			return nil, err
		}
		deliveries = append(deliveries, d)
	}

	return deliveries, nil
}

func (r *deliveryRepository) IncrementCount(ctx context.Context, id string, eventType domain.EventType) (bool, error) {
	db := extractTx(ctx, r.db.DB)

	// Load current delivery counters and opened_at to determine if this is the first event of the type.
	var entity Delivery
	if err := db.WithContext(ctx).First(&entity, "id = ?", id).Error; err != nil {
		return false, err
	}

	switch eventType {
	case domain.EventTypeOpened:
		// Always increment delivery open_count, but detect if this is the first open for campaign-level counting.
		isFirst := entity.OpenCount == 0
		now := time.Now().Unix()
		// If opened_at is null, set it to now; always increment open_count.
		if err := db.WithContext(ctx).Model(&Delivery{}).Where("id = ?", id).Updates(map[string]interface{}{
			"open_count": gorm.Expr("open_count + 1"),
			"opened_at":  gorm.Expr("COALESCE(opened_at, ?)", now),
		}).Error; err != nil {
			return false, err
		}
		return isFirst, nil
	case domain.EventTypeClicked:
		isFirst := entity.ClickCount == 0
		if err := db.WithContext(ctx).Model(&Delivery{}).Where("id = ?", id).Update("click_count", gorm.Expr("click_count + 1")).Error; err != nil {
			return false, err
		}
		return isFirst, nil
	case domain.EventTypeBounced:
		isFirst := entity.BounceCount == 0
		if err := db.WithContext(ctx).Model(&Delivery{}).Where("id = ?", id).Update("bounce_count", gorm.Expr("bounce_count + 1")).Error; err != nil {
			return false, err
		}
		return isFirst, nil
	default:
		return false, nil
	}
}

func (r *deliveryRepository) UpdateStatus(ctx context.Context, id string, status domain.DeliveryStatus) error {
	db := extractTx(ctx, r.db.DB)
	return db.WithContext(ctx).Model(&Delivery{}).Where("id = ?", id).Update("status", status).Error
}

func (r *deliveryRepository) UpdateSendScheduledByCampaign(ctx context.Context, campaignID string, ts int64) (int, error) {
	db := extractTx(ctx, r.db.DB)
	res := db.WithContext(ctx).Model(&Delivery{}).
		Where("campaign_id = ? AND status = ?", campaignID, domain.DeliveryStatusIdle).
		Update("scheduled_at", ts).
		Update("status", domain.DeliveryStatusScheduled)
	if res.Error != nil {
		return 0, res.Error
	}
	return int(res.RowsAffected), nil
}
