package page

import (
	"context"

	"codebasego/internal/common"
)

// Repository defines the data access contract for pages.
type Repository interface {
	FindAll(ctx context.Context, query common.PaginationQuery) ([]Page, int, error)
	FindAllCursor(ctx context.Context, query common.CursorQuery) ([]Page, common.CursorMeta, error)
	FindByID(ctx context.Context, id string) (*Page, error)
	FindBySlug(ctx context.Context, slug string) (*Page, error)
	GetPublishedSlugs(ctx context.Context) ([]string, error)
	Create(ctx context.Context, entity *Page) error
	Update(ctx context.Context, entity *Page) error
	Delete(ctx context.Context, id string) error
}
