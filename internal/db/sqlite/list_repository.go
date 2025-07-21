package sqlite

import (
	"context"
	"encoding/json"

	"github.com/headmail/headmail/pkg/domain"
	"github.com/headmail/headmail/pkg/repository"
	"gorm.io/gorm"
)

// listRepository implements the repository.ListRepository interface.
type listRepository struct {
	db *gorm.DB
}

// NewListRepository creates a new list repository.
func NewListRepository(db *gorm.DB) repository.ListRepository {
	return &listRepository{db: db}
}

// domainToListEntity converts a domain List to a GORM List entity.
func domainToListEntity(d *domain.List) (*List, error) {
	tagsJSON, err := json.Marshal(d.Tags)
	if err != nil {
		return nil, err
	}
	return &List{
		ID:              d.ID,
		Name:            d.Name,
		Description:     d.Description,
		Tags:            tagsJSON,
		SubscriberCount: d.SubscriberCount,
		CreatedAt:       d.CreatedAt,
		UpdatedAt:       d.UpdatedAt,
		DeletedAt:       d.DeletedAt,
	}, nil
}

// entityToListDomain converts a GORM List entity to a domain List.
func entityToListDomain(e *List) (*domain.List, error) {
	d := &domain.List{
		ID:              e.ID,
		Name:            e.Name,
		Description:     e.Description,
		SubscriberCount: e.SubscriberCount,
		CreatedAt:       e.CreatedAt,
		UpdatedAt:       e.UpdatedAt,
		DeletedAt:       e.DeletedAt,
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
	db := extractTx(ctx, r.db)
	return db.WithContext(ctx).Create(entity).Error
}

func (r *listRepository) GetByID(ctx context.Context, id string) (*domain.List, error) {
	var entity List
	db := extractTx(ctx, r.db)
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
	db := extractTx(ctx, r.db)
	return db.WithContext(ctx).Save(entity).Error
}

func (r *listRepository) Delete(ctx context.Context, id string) error {
	db := extractTx(ctx, r.db)
	return db.WithContext(ctx).Delete(&List{}, "id = ?", id).Error
}

func (r *listRepository) List(ctx context.Context, filter repository.ListFilter, pagination repository.Pagination) ([]*domain.List, int, error) {
	var entities []List
	var total int64

	db := extractTx(ctx, r.db)
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
	panic("implement me")
}

func (r *listRepository) GetSubscribers(ctx context.Context) (chan *domain.Subscriber, error) {
	panic("implement me")
}
