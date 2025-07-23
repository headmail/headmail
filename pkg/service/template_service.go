package service

import (
	"context"

	"github.com/headmail/headmail/pkg/domain"
	"github.com/headmail/headmail/pkg/repository"
)

// TemplateServiceProvider defines the interface for a template service.
type TemplateServiceProvider interface {
	CreateTemplate(ctx context.Context, template *domain.Template) error
	GetTemplate(ctx context.Context, id string) (*domain.Template, error)
	UpdateTemplate(ctx context.Context, template *domain.Template) error
	DeleteTemplate(ctx context.Context, id string) error
	ListTemplates(ctx context.Context, pagination repository.Pagination) ([]*domain.Template, int, error)
}

// TemplateService provides business logic for template management.
type TemplateService struct {
	repo repository.TemplateRepository
}

// NewTemplateService creates a new TemplateService.
func NewTemplateService(db repository.DB) *TemplateService {
	return &TemplateService{
		repo: db.TemplateRepository(),
	}
}

// CreateTemplate creates a new template.
func (s *TemplateService) CreateTemplate(ctx context.Context, template *domain.Template) error {
	return s.repo.Create(ctx, template)
}

// GetTemplate retrieves a template by its ID.
func (s *TemplateService) GetTemplate(ctx context.Context, id string) (*domain.Template, error) {
	return s.repo.GetByID(ctx, id)
}

// UpdateTemplate updates an existing template.
func (s *TemplateService) UpdateTemplate(ctx context.Context, template *domain.Template) error {
	return s.repo.Update(ctx, template)
}

// DeleteTemplate deletes a template by its ID.
func (s *TemplateService) DeleteTemplate(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

// ListTemplates lists all templates.
func (s *TemplateService) ListTemplates(ctx context.Context, pagination repository.Pagination) ([]*domain.Template, int, error) {
	return s.repo.List(ctx, pagination)
}
