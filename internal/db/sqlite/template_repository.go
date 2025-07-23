package sqlite

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/headmail/headmail/pkg/domain"
	"github.com/headmail/headmail/pkg/repository"
)

type templateRepository struct {
	db *gorm.DB
}

func NewTemplateRepository(db *gorm.DB) repository.TemplateRepository {
	return &templateRepository{db: db}
}

func (r *templateRepository) Create(ctx context.Context, template *domain.Template) error {
	template.ID = uuid.New().String()
	template.CreatedAt = time.Now().Unix()
	template.UpdatedAt = template.CreatedAt

	gormTemplate := toTemplateGorm(template)
	return r.db.WithContext(ctx).Create(gormTemplate).Error
}

func (r *templateRepository) GetByID(ctx context.Context, id string) (*domain.Template, error) {
	var gormTemplate Template
	if err := r.db.WithContext(ctx).First(&gormTemplate, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return toTemplateDomain(&gormTemplate), nil
}

func (r *templateRepository) Update(ctx context.Context, template *domain.Template) error {
	template.UpdatedAt = time.Now().Unix()
	gormTemplate := toTemplateGorm(template)
	return r.db.WithContext(ctx).Save(gormTemplate).Error
}

func (r *templateRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&Template{}, "id = ?", id).Error
}

func (r *templateRepository) List(ctx context.Context, pagination repository.Pagination) ([]*domain.Template, int, error) {
	var gormTemplates []*Template
	var total int64

	query := r.db.WithContext(ctx).Model(&Template{})

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if pagination.Limit > 0 {
		query = query.Offset((pagination.Page - 1) * pagination.Limit).Limit(pagination.Limit)
	}

	if err := query.Find(&gormTemplates).Error; err != nil {
		return nil, 0, err
	}

	domainTemplates := make([]*domain.Template, len(gormTemplates))
	for i, t := range gormTemplates {
		domainTemplates[i] = toTemplateDomain(t)
	}

	return domainTemplates, int(total), nil
}

func toTemplateGorm(d *domain.Template) *Template {
	return &Template{
		ID:        d.ID,
		Name:      d.Name,
		BodyHTML:  d.BodyHTML,
		BodyText:  d.BodyText,
		CreatedAt: d.CreatedAt,
		UpdatedAt: d.UpdatedAt,
	}
}

func toTemplateDomain(g *Template) *domain.Template {
	return &domain.Template{
		ID:        g.ID,
		Name:      g.Name,
		BodyHTML:  g.BodyHTML,
		BodyText:  g.BodyText,
		CreatedAt: g.CreatedAt,
		UpdatedAt: g.UpdatedAt,
	}
}
