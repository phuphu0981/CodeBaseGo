package graphql

import (
	"context"
	"testing"
	"time"

	"codebasego/internal/common"
	"codebasego/internal/modules/auth"
	"codebasego/internal/modules/user"
)

type mockGQLUserService struct{}

func (m *mockGQLUserService) List(ctx context.Context, query common.PaginationQuery) ([]user.User, int, error) {
	users := []user.User{
		{ID: "user-1", Email: "u1@example.com", Name: "User 1", CreatedAt: time.Now(), UpdatedAt: time.Now()},
	}
	return users, 1, nil
}

func (m *mockGQLUserService) ListCursor(ctx context.Context, query common.CursorQuery) ([]user.User, common.CursorMeta, error) {
	users := []user.User{
		{ID: "user-1", Email: "u1@example.com", Name: "User 1", CreatedAt: time.Now(), UpdatedAt: time.Now()},
	}
	return users, common.CursorMeta{HasMore: false, Limit: query.Limit}, nil
}

func (m *mockGQLUserService) GetByID(ctx context.Context, id string) (*user.User, error) {
	if id == "not-found" {
		return nil, common.ErrNotFound
	}
	return &user.User{ID: id, Email: "u1@example.com", Name: "User 1", CreatedAt: time.Now(), UpdatedAt: time.Now()}, nil
}

func (m *mockGQLUserService) Update(ctx context.Context, id string, req *user.UpdateRequest) (*user.User, error) {
	return &user.User{ID: id, Email: "updated@example.com", Name: "Updated", CreatedAt: time.Now(), UpdatedAt: time.Now()}, nil
}

func (m *mockGQLUserService) Delete(ctx context.Context, id string) error {
	return nil
}

type mockGQLAuthService struct{}

func (m *mockGQLAuthService) Register(ctx context.Context, email, password, name string) (*user.User, error) {
	return &user.User{ID: "new-user-id", Email: email, Name: name, CreatedAt: time.Now(), UpdatedAt: time.Now()}, nil
}

func (m *mockGQLAuthService) Login(ctx context.Context, email, password string) (*user.User, error) {
	if password == "wrong" {
		return nil, common.ErrUnauthorized
	}
	return &user.User{ID: "user-1", Email: email, Name: "User 1", CreatedAt: time.Now(), UpdatedAt: time.Now()}, nil
}

func (m *mockGQLAuthService) GenerateTokenPair(ctx context.Context, userID, email string, role ...string) (*auth.TokenResponse, error) {
	return &auth.TokenResponse{
		AccessToken:  "access-token-123",
		RefreshToken: "refresh-token-123",
		ExpiresIn:    900,
		TokenType:    "Bearer",
	}, nil
}

func (m *mockGQLAuthService) RefreshToken(ctx context.Context, refreshTokenStr string) (*auth.TokenResponse, error) {
	return &auth.TokenResponse{
		AccessToken:  "new-access-token",
		RefreshToken: "new-refresh-token",
		ExpiresIn:    900,
		TokenType:    "Bearer",
	}, nil
}

func (m *mockGQLAuthService) Logout(ctx context.Context, refreshTokenStr string) error {
	return nil
}

func (m *mockGQLAuthService) LogoutAll(ctx context.Context, userID string) error {
	return nil
}

func (m *mockGQLAuthService) ValidateToken(tokenStr string) (*auth.Claims, error) {
	if tokenStr == "invalid" {
		return nil, common.ErrUnauthorized
	}
	return &auth.Claims{UserID: "user-1", Email: "u1@example.com"}, nil
}

func TestGraphQLResolver(t *testing.T) {
	res := NewResolver(&mockGQLUserService{}, &mockGQLAuthService{})
	q := res.Query()
	m := res.Mutation()

	t.Run("Query Me - Unauthenticated Error", func(t *testing.T) {
		ctx := context.Background()
		_, err := q.Me(ctx)
		if err != common.ErrUnauthorized {
			t.Fatalf("expected ErrUnauthorized, got %v", err)
		}
	})

	t.Run("Query Me - Success when authenticated", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), userIDCtxKey, "user-1")
		u, err := q.Me(ctx)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if u.ID != "user-1" {
			t.Fatalf("expected user-1, got %s", u.ID)
		}
	})

	t.Run("Query Users - Success", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), userIDCtxKey, "user-1")
		page := 1
		limit := 10
		users, err := q.Users(ctx, &page, &limit)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(users) != 1 {
			t.Fatalf("expected 1 user, got %d", len(users))
		}
	})

	t.Run("Mutation Register - Success", func(t *testing.T) {
		ctx := context.Background()
		pair, err := m.Register(ctx, RegisterInput{
			Email:    "new@example.com",
			Password: "password123",
			Name:     "New User",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if pair.AccessToken == "" {
			t.Fatal("expected access token")
		}
	})

	t.Run("Mutation UpdateUser - Forbidden for different user ID", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), userIDCtxKey, "user-1")
		newName := "New Name"
		_, err := m.UpdateUser(ctx, "user-2", UpdateUserInput{Name: &newName})
		if err != common.ErrForbidden {
			t.Fatalf("expected ErrForbidden, got %v", err)
		}
	})

	t.Run("Mutation UpdateUser - Success for same user ID", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), userIDCtxKey, "user-1")
		newName := "New Name"
		u, err := m.UpdateUser(ctx, "user-1", UpdateUserInput{Name: &newName})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if u.Name != "Updated" {
			t.Fatalf("expected Updated, got %s", u.Name)
		}
	})
}
