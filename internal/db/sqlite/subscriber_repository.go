package sqlite

import (
	"context"

	"github.com/headmail/headmail/pkg/domain"
	"github.com/headmail/headmail/pkg/repository"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type subscriberRepository struct {
	db *gorm.DB
}

func NewSubscriberRepository(db *gorm.DB) repository.SubscriberRepository {
	return &subscriberRepository{db: db}
}

func domainToSubscriberEntity(d *domain.Subscriber) *Subscriber {
	return &Subscriber{
		ID:             d.ID,
		Email:          d.Email,
		Name:           d.Name,
		Status:         d.Status,
		SubscribedAt:   d.SubscribedAt,
		UnsubscribedAt: d.UnsubscribedAt,
		CreatedAt:      d.CreatedAt,
		UpdatedAt:      d.UpdatedAt,
	}
}

func entityToSubscriberDomain(e *Subscriber) *domain.Subscriber {
	return &domain.Subscriber{
		ID:             e.ID,
		Email:          e.Email,
		Name:           e.Name,
		Status:         e.Status,
		SubscribedAt:   e.SubscribedAt,
		UnsubscribedAt: e.UnsubscribedAt,
		CreatedAt:      e.CreatedAt,
		UpdatedAt:      e.UpdatedAt,
	}
}

func (r *subscriberRepository) Create(ctx context.Context, subscriber *domain.Subscriber) error {
	entity := domainToSubscriberEntity(subscriber)
	db := extractTx(ctx, r.db)
	return db.WithContext(ctx).Create(entity).Error
}

func (r *subscriberRepository) GetByID(ctx context.Context, id string) (*domain.Subscriber, error) {
	var entity Subscriber
	db := extractTx(ctx, r.db)
	if err := db.WithContext(ctx).First(&entity, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return entityToSubscriberDomain(&entity), nil
}

func (r *subscriberRepository) GetByEmail(ctx context.Context, email string) (*domain.Subscriber, error) {
	var entity Subscriber
	db := extractTx(ctx, r.db)
	if err := db.WithContext(ctx).First(&entity, "email = ?", email).Error; err != nil {
		return nil, err
	}
	return entityToSubscriberDomain(&entity), nil
}

func (r *subscriberRepository) Update(ctx context.Context, subscriber *domain.Subscriber) error {
	entity := domainToSubscriberEntity(subscriber)
	db := extractTx(ctx, r.db)
	return db.WithContext(ctx).Save(entity).Error
}

func (r *subscriberRepository) Delete(ctx context.Context, id string) error {
	db := extractTx(ctx, r.db)
	return db.WithContext(ctx).Delete(&Subscriber{}, "id = ?", id).Error
}

func (r *subscriberRepository) List(ctx context.Context, filter repository.SubscriberFilter, pagination repository.Pagination) ([]*domain.Subscriber, int, error) {
	var entities []Subscriber
	var total int64

	db := extractTx(ctx, r.db)
	query := db.WithContext(ctx).Model(&Subscriber{})

	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}
	if filter.Search != "" {
		query = query.Where("email LIKE ? OR name LIKE ?", "%"+filter.Search+"%", "%"+filter.Search+"%")
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (pagination.Page - 1) * pagination.Limit
	if err := query.Offset(offset).Limit(pagination.Limit).Find(&entities).Error; err != nil {
		return nil, 0, err
	}

	var subscribers []*domain.Subscriber
	for _, entity := range entities {
		subscribers = append(subscribers, entityToSubscriberDomain(&entity))
	}

	return subscribers, int(total), nil
}

func (r *subscriberRepository) BulkUpsert(ctx context.Context, subscribers []*domain.Subscriber) error {
	// GORM's upsert requires specifying conflict columns.
	// This is a basic implementation. A more robust one might handle conflicts differently.
	entities := make([]*Subscriber, len(subscribers))
	for i, s := range subscribers {
		entities[i] = domainToSubscriberEntity(s)
	}

	db := extractTx(ctx, r.db)
	return db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "email"}},
		DoUpdates: clause.AssignmentColumns([]string{"name", "status", "updated_at"}),
	}).Create(&entities).Error
}
