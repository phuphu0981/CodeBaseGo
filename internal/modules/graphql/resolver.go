package graphql

import (
	"context"

	"codebasego/internal/common"
	"codebasego/internal/modules/auth"
	"codebasego/internal/modules/user"
)

type UserService interface {
	List(ctx context.Context, query common.PaginationQuery) ([]user.User, int, error)
	ListCursor(ctx context.Context, query common.CursorQuery) ([]user.User, common.CursorMeta, error)
	GetByID(ctx context.Context, id string) (*user.User, error)
	Update(ctx context.Context, id string, req *user.UpdateRequest) (*user.User, error)
	Delete(ctx context.Context, id string) error
}

type AuthService interface {
	Register(ctx context.Context, email, password, name string) (*user.User, error)
	Login(ctx context.Context, email, password string) (*user.User, error)
	GenerateTokenPair(ctx context.Context, userID, email string, role ...string) (*auth.TokenResponse, error)
	RefreshToken(ctx context.Context, refreshTokenStr string) (*auth.TokenResponse, error)
	Logout(ctx context.Context, refreshTokenStr string) error
	LogoutAll(ctx context.Context, userID string) error
	ValidateToken(tokenStr string) (*auth.Claims, error)
}

// Resolver acts as dependency injection container for GraphQL resolvers.
type Resolver struct {
	userService UserService
	authService AuthService
}

func NewResolver(userService UserService, authService AuthService) *Resolver {
	return &Resolver{
		userService: userService,
		authService: authService,
	}
}

// Mutation returns MutationResolver implementation.
func (r *Resolver) Mutation() MutationResolver { return &mutationResolver{r} }

// Query returns QueryResolver implementation.
func (r *Resolver) Query() QueryResolver { return &queryResolver{r} }

type (
	mutationResolver struct{ *Resolver }
	queryResolver    struct{ *Resolver }
)
