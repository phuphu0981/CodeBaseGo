package seo

import (
	"context"
	"strings"

	"codebasego/internal/common"
)

// SEO default values.
const (
	DefaultSlug        = "home"
	DefaultOGType      = "website"
	DefaultTwitterCard = "summary_large_image"
	DefaultRobots      = "index, follow"
)

// Service contains SEO business logic.
type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) List(ctx context.Context, query common.PaginationQuery) ([]SEO, int, error) {
	return s.repo.FindAll(ctx, query)
}

func (s *Service) ListCursor(ctx context.Context, query common.CursorQuery) ([]SEO, common.CursorMeta, error) {
	return s.repo.FindAllCursor(ctx, query)
}

func (s *Service) GetByID(ctx context.Context, id string) (*SEO, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *Service) GetBySlug(ctx context.Context, slug string) (*SEO, error) {
	cleanSlug := NormalizeSlug(slug)
	return s.repo.FindBySlug(ctx, cleanSlug)
}

func (s *Service) Create(ctx context.Context, req *CreateSEORequest) (*SEO, error) {
	if req == nil {
		return nil, common.ErrBadRequest
	}
	if err := req.Validate(); err != nil {
		return nil, err
	}

	ogType := req.OGType
	if ogType == "" {
		ogType = DefaultOGType
	}
	twitterCard := req.TwitterCard
	if twitterCard == "" {
		twitterCard = DefaultTwitterCard
	}
	robots := req.Robots
	if robots == "" {
		robots = DefaultRobots
	}

	entity := &SEO{
		Slug:           NormalizeSlug(req.Slug),
		Title:          req.Title,
		Description:    req.Description,
		Keywords:       req.Keywords,
		CanonicalURL:   req.CanonicalURL,
		OGTitle:        req.OGTitle,
		OGDescription:  req.OGDescription,
		OGImage:        req.OGImage,
		OGType:         ogType,
		TwitterCard:    twitterCard,
		Robots:         robots,
		StructuredData: req.StructuredData,
	}

	if err := s.repo.Create(ctx, entity); err != nil {
		return nil, err
	}
	return entity, nil
}

func (s *Service) Update(ctx context.Context, id string, req *UpdateSEORequest) (*SEO, error) {
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
	if req.Description != nil {
		entity.Description = *req.Description
	}
	if req.Keywords != nil {
		entity.Keywords = *req.Keywords
	}
	if req.CanonicalURL != nil {
		entity.CanonicalURL = *req.CanonicalURL
	}
	if req.OGTitle != nil {
		entity.OGTitle = *req.OGTitle
	}
	if req.OGDescription != nil {
		entity.OGDescription = *req.OGDescription
	}
	if req.OGImage != nil {
		entity.OGImage = *req.OGImage
	}
	if req.OGType != nil {
		entity.OGType = *req.OGType
	}
	if req.TwitterCard != nil {
		entity.TwitterCard = *req.TwitterCard
	}
	if req.Robots != nil {
		entity.Robots = *req.Robots
	}
	if req.StructuredData != nil {
		entity.StructuredData = *req.StructuredData
	}

	if err := s.repo.Update(ctx, entity); err != nil {
		return nil, err
	}
	return entity, nil
}

func (s *Service) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

// NormalizeSlug cleans and standardizes a URL slug path.
func NormalizeSlug(slug string) string {
	slug = strings.TrimSpace(slug)
	slug = strings.Trim(slug, "/")
	if slug == "" {
		return DefaultSlug
	}
	return strings.ToLower(slug)
}
