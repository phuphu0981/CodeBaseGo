package setting

import (
	"context"

	"codebasego/internal/common"
)

// Repository defines the data access contract for system configurations.
type Repository interface {
	FindAll(ctx context.Context, query common.PaginationQuery, scope, pathPrefix string) ([]CoreConfig, int, error)
	FindAllCursor(ctx context.Context, query common.CursorQuery, scope, pathPrefix string) ([]CoreConfig, common.CursorMeta, error)
	FindByID(ctx context.Context, id string) (*CoreConfig, error)
	FindByPath(ctx context.Context, scope, scopeID, path string) (*CoreConfig, error)
	FindByPrefix(ctx context.Context, scope, scopeID, prefix string) ([]CoreConfig, error)
	Save(ctx context.Context, entity *CoreConfig) error
	Delete(ctx context.Context, id string) error
	DeleteByPath(ctx context.Context, scope, scopeID, path string) error
}
