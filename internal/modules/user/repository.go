package user

import (
	"context"

	"codebasego/internal/common"
)

// Repository defines the data access contract for users.
type Repository interface {
	FindAll(ctx context.Context, query common.PaginationQuery) ([]User, int, error)
	FindAllCursor(ctx context.Context, query common.CursorQuery) ([]User, common.CursorMeta, error)
	FindByID(ctx context.Context, id string) (*User, error)
	FindByEmail(ctx context.Context, email string) (*User, error)
	Create(ctx context.Context, entity *User) error
	Update(ctx context.Context, entity *User) error
	Delete(ctx context.Context, id string) error
}

