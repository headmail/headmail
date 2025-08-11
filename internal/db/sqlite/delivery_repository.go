package sqlite

import (
	"context"
	"encoding/json"

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
		ID:              d.ID,
		CampaignID:      d.CampaignID,
		Type:            d.Type,
		Status:          d.Status,
		Name:            d.Name,
		Email:           d.Email,
		Subject:         d.Subject,
		MessageID:       d.MessageID,
		Data:            dataJSON,
		Headers:         headersJSON,
		Tags:            tagsJSON,
		CreatedAt:       d.CreatedAt,
		ScheduledAt:     d.ScheduledAt,
		Attempts:        d.Attempts,
		SendScheduledAt: d.SendScheduledAt,
		SentAt:          d.SentAt,
		OpenedAt:        d.OpenedAt,
		FailedAt:        d.FailedAt,
		FailureReason:   d.FailureReason,
		OpenCount:       d.OpenCount,
		ClickCount:      d.ClickCount,
		BounceCount:     d.BounceCount,
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
		ID:              e.ID,
		CampaignID:      e.CampaignID,
		Type:            e.Type,
		Status:          e.Status,
		Name:            e.Name,
		Email:           e.Email,
		Subject:         e.Subject,
		MessageID:       e.MessageID,
		Data:            data,
		Headers:         headers,
		Tags:            tags,
		CreatedAt:       e.CreatedAt,
		ScheduledAt:     e.ScheduledAt,
		Attempts:        e.Attempts,
		SendScheduledAt: e.SendScheduledAt,
		SentAt:          e.SentAt,
		OpenedAt:        e.OpenedAt,
		FailedAt:        e.FailedAt,
		FailureReason:   e.FailureReason,
		OpenCount:       e.OpenCount,
		ClickCount:      e.ClickCount,
		BounceCount:     e.BounceCount,
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
	if err := query.Offset(offset).Limit(pagination.Limit).Find(&entities).Error; err != nil {
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
		Where("send_scheduled_at IS NOT NULL AND send_scheduled_at <= ?", ts).
		Where("status = ?", domain.DeliveryStatusScheduled).
		Order("send_scheduled_at ASC, created_at ASC").
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

func (r *deliveryRepository) UpdateStatus(ctx context.Context, id string, status string) error {
	db := extractTx(ctx, r.db.DB)
	return db.WithContext(ctx).Model(&Delivery{}).Where("id = ?", id).Update("status", status).Error
}
