package sqlite

import (
	"context"
	"strings"

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
	lists := make([]SubscriberList, len(d.Lists))
	for i, l := range d.Lists {
		lists[i] = SubscriberList{
			ListID:         l.ListID,
			Status:         l.Status,
			SubscribedAt:   l.SubscribedAt,
			UnsubscribedAt: l.UnsubscribedAt,
			CreatedAt:      l.CreatedAt,
			UpdatedAt:      l.UpdatedAt,
		}
	}

	return &Subscriber{
		ID:        d.ID,
		Email:     d.Email,
		Name:      d.Name,
		Status:    d.Status,
		CreatedAt: d.CreatedAt,
		UpdatedAt: d.UpdatedAt,
		Lists:     lists,
	}
}

func entityToSubscriberDomain(e *Subscriber) *domain.Subscriber {
	lists := make([]domain.SubscriberList, len(e.Lists))
	for i, l := range e.Lists {
		lists[i] = domain.SubscriberList{
			ListID:         l.ListID,
			Status:         l.Status,
			SubscribedAt:   l.SubscribedAt,
			UnsubscribedAt: l.UnsubscribedAt,
			CreatedAt:      l.CreatedAt,
			UpdatedAt:      l.UpdatedAt,
		}
	}

	return &domain.Subscriber{
		ID:        e.ID,
		Email:     e.Email,
		Name:      e.Name,
		Status:    e.Status,
		CreatedAt: e.CreatedAt,
		UpdatedAt: e.UpdatedAt,
		Lists:     lists,
	}
}

func (r *subscriberRepository) Create(ctx context.Context, subscriber *domain.Subscriber) error {
	entity := domainToSubscriberEntity(subscriber)
	db := extractTx(ctx, r.db)

	// Create the subscriber first, without the lists
	subscriberToCreate := &Subscriber{
		ID:        entity.ID,
		Email:     entity.Email,
		Name:      entity.Name,
		Status:    entity.Status,
		CreatedAt: entity.CreatedAt,
		UpdatedAt: entity.UpdatedAt,
	}
	err := db.WithContext(ctx).Create(subscriberToCreate).Error
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return &repository.ErrUniqueConstraintFailed{Cause: err}
		}
		return err
	}

	// Then, create the associations
	for _, list := range entity.Lists {
		list.SubscriberID = entity.ID
		if err := db.WithContext(ctx).Create(&list).Error; err != nil {
			return err
		}
	}

	return nil
}

func (r *subscriberRepository) GetByID(ctx context.Context, id string) (*domain.Subscriber, error) {
	var entity Subscriber
	db := extractTx(ctx, r.db)
	if err := db.WithContext(ctx).Preload("Lists").First(&entity, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, &repository.ErrNotFound{Entity: "Subscriber", ID: id}
		}
		return nil, err
	}
	return entityToSubscriberDomain(&entity), nil
}

func (r *subscriberRepository) GetByEmail(ctx context.Context, email string) (*domain.Subscriber, error) {
	var entity Subscriber
	db := extractTx(ctx, r.db)
	if err := db.WithContext(ctx).Preload("Lists").First(&entity, "email = ?", email).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, &repository.ErrNotFound{Entity: "Subscriber", ID: email}
		}
		return nil, err
	}
	return entityToSubscriberDomain(&entity), nil
}

func (r *subscriberRepository) Update(ctx context.Context, subscriber *domain.Subscriber) error {
	entity := domainToSubscriberEntity(subscriber)
	db := extractTx(ctx, r.db)

	// Update subscriber basic info
	if err := db.WithContext(ctx).Model(&Subscriber{}).Where("id = ?", entity.ID).Updates(entity).Error; err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return &repository.ErrUniqueConstraintFailed{Cause: err}
		}
		return err
	}

	// Update lists associations
	// First, remove existing associations
	if err := db.WithContext(ctx).Where("subscriber_id = ?", entity.ID).Delete(&SubscriberList{}).Error; err != nil {
		return err
	}

	// Then, add the new associations
	for _, list := range entity.Lists {
		list.SubscriberID = entity.ID
		if err := db.WithContext(ctx).Create(&list).Error; err != nil {
			return err
		}
	}

	return nil
}

func (r *subscriberRepository) Delete(ctx context.Context, id string) error {
	db := extractTx(ctx, r.db)
	if err := db.WithContext(ctx).Delete(&SubscriberList{}, "subscriber_id = ?", id).Error; err != nil {
		return err
	}
	if err := db.WithContext(ctx).Delete(&Subscriber{}, "id = ?", id).Error; err != nil {
		return err
	}
	return nil
}

func (r *subscriberRepository) List(ctx context.Context, filter repository.SubscriberFilter, pagination repository.Pagination) ([]*domain.Subscriber, int, error) {
	var entities []Subscriber
	var total int64

	db := extractTx(ctx, r.db)
	query := db.WithContext(ctx).Model(&Subscriber{}).Preload("Lists")

	if filter.ListID != "" {
		query = query.Joins("JOIN subscriber_lists on subscribers.id = subscriber_lists.subscriber_id").
			Where("subscriber_lists.list_id = ?", filter.ListID)
		if filter.ListStatus != "" {
			query = query.Where("subscriber_lists.status = ?", filter.ListStatus)
		}
	}
	if filter.Status != "" {
		query = query.Where("subscribers.status = ?", filter.Status)
	}
	if filter.Search != "" {
		query = query.Where("subscribers.email LIKE ? OR subscribers.name LIKE ?", "%"+filter.Search+"%", "%"+filter.Search+"%")
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if pagination.Limit > 0 {
		offset := (pagination.Page - 1) * pagination.Limit
		if err := query.Offset(offset).Limit(pagination.Limit).Find(&entities).Error; err != nil {
			return nil, 0, err
		}
	} else {
		if err := query.Find(&entities).Error; err != nil {
			return nil, 0, err
		}
	}

	var subscribers []*domain.Subscriber
	for _, entity := range entities {
		subscribers = append(subscribers, entityToSubscriberDomain(&entity))
	}

	return subscribers, int(total), nil
}

func (r *subscriberRepository) BulkUpsert(ctx context.Context, subscribers []*domain.Subscriber) error {
	db := extractTx(ctx, r.db)
	return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, s := range subscribers {
			// Upsert subscriber
			out := &Subscriber{}
			result := tx.Where(&Subscriber{
				Email: s.Email,
			}).Assign(&Subscriber{
				Name:      s.Name,
				UpdatedAt: s.UpdatedAt,
			}).Attrs(&Subscriber{
				ID:        s.ID,
				Status:    s.Status,
				CreatedAt: s.CreatedAt,
			}).FirstOrCreate(out)
			if result.Error != nil {
				return result.Error
			}
			s.ID = out.ID

			// Upsert subscriber lists
			for _, l := range s.Lists {
				listEntity := SubscriberList{
					SubscriberID: s.ID,
					ListID:       l.ListID,
					Status:       l.Status,
					CreatedAt:    l.CreatedAt,
					UpdatedAt:    l.UpdatedAt,
				}
				if err := tx.Clauses(clause.OnConflict{
					Columns:   []clause.Column{{Name: "subscriber_id"}, {Name: "list_id"}},
					DoUpdates: clause.AssignmentColumns([]string{"updated_at"}),
				}).Create(&listEntity).Error; err != nil {
					return err
				}
			}
		}
		return nil
	})
}
