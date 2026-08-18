package seo

import (
	"context"

	"codebasego/internal/common"
)

// Repository defines the data access contract for SEO records.
type Repository interface {
	FindAll(ctx context.Context, query common.PaginationQuery) ([]SEO, int, error)
	FindAllCursor(ctx context.Context, query common.CursorQuery) ([]SEO, common.CursorMeta, error)
	FindByID(ctx context.Context, id string) (*SEO, error)
	FindBySlug(ctx context.Context, slug string) (*SEO, error)
	Create(ctx context.Context, entity *SEO) error
	Update(ctx context.Context, entity *SEO) error
	Delete(ctx context.Context, id string) error
}
