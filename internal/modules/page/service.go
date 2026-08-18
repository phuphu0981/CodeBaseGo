package page

import (
	"context"
	"strings"

	"codebasego/internal/common"
)

// Page status and template constants.
const (
	DefaultSlug     = "home"
	StatusPublished = "published"
	StatusDraft     = "draft"
	DefaultTemplate = "default"
)

// Service contains page business logic.
type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) List(ctx context.Context, query common.PaginationQuery) ([]Page, int, error) {
	return s.repo.FindAll(ctx, query)
}

func (s *Service) ListCursor(ctx context.Context, query common.CursorQuery) ([]Page, common.CursorMeta, error) {
	return s.repo.FindAllCursor(ctx, query)
}

func (s *Service) GetByID(ctx context.Context, id string) (*Page, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *Service) GetBySlug(ctx context.Context, slug string) (*Page, error) {
	cleanSlug := NormalizeSlug(slug)
	return s.repo.FindBySlug(ctx, cleanSlug)
}

func (s *Service) GetPublishedSlugs(ctx context.Context) ([]string, error) {
	return s.repo.GetPublishedSlugs(ctx)
}

func (s *Service) Create(ctx context.Context, req *CreatePageRequest) (*Page, error) {
	if req == nil {
		return nil, common.ErrBadRequest
	}
	if err := req.Validate(); err != nil {
		return nil, err
	}

	status := req.Status
	if status == "" {
		status = StatusPublished
	}
	template := req.Template
	if template == "" {
		template = DefaultTemplate
	}

	entity := &Page{
		Slug:     NormalizeSlug(req.Slug),
		Title:    req.Title,
		Status:   status,
		Template: template,
		Content:  req.Content,
	}

	if err := s.repo.Create(ctx, entity); err != nil {
		return nil, err
	}
	return entity, nil
}

func (s *Service) Update(ctx context.Context, id string, req *UpdatePageRequest) (*Page, error) {
	if req == nil {
		return nil, common.ErrBadRequest
	}
	if err := req.Validate(); err != nil {
		return nil, err
	}

	entity, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.Slug != nil {
		entity.Slug = NormalizeSlug(*req.Slug)
	}
	if req.Title != nil {
		entity.Title = *req.Title
	}
	if req.Status != nil {
		entity.Status = *req.Status
	}
	if req.Template != nil {
		entity.Template = *req.Template
	}
	if req.Content != nil {
		entity.Content = *req.Content
	}

	if err := s.repo.Update(ctx, entity); err != nil {
		return nil, err
	}
	return entity, nil
}

func (s *Service) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

// NormalizeSlug standardizes the page slug.
func NormalizeSlug(slug string) string {
	slug = strings.TrimSpace(slug)
	slug = strings.Trim(slug, "/")
	if slug == "" {
		return DefaultSlug
	}
	return strings.ToLower(slug)
}
