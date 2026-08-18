package auth

import (
	"context"
	"testing"
	"time"

	"codebasego/internal/common"
	"codebasego/internal/modules/user"
	"codebasego/internal/platform/config"
)

type mockUserService struct {
	users map[string]*user.User
}

func newMockUserService() *mockUserService {
	return &mockUserService{users: make(map[string]*user.User)}
}

func (m *mockUserService) Create(ctx context.Context, req *user.CreateRequest, hashedPassword string) (*user.User, error) {
	for _, u := range m.users {
		if u.Email == req.Email {
			return nil, common.ErrConflict
		}
	}
	u := &user.User{
		ID:        "user-1",
		Email:     req.Email,
		Password:  hashedPassword,
		Name:      req.Name,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	m.users[u.ID] = u
	m.users[u.Email] = u
	return u, nil
}

func (m *mockUserService) GetByID(ctx context.Context, id string) (*user.User, error) {
	if u, ok := m.users[id]; ok {
		return u, nil
	}
	return nil, common.ErrNotFound
}

func (m *mockUserService) GetByEmail(ctx context.Context, email string) (*user.User, error) {
	if u, ok := m.users[email]; ok {
		return u, nil
	}
	return nil, common.ErrNotFound
}

type mockRefreshTokenRepo struct {
	tokens map[string]*RefreshToken
}

func newMockRefreshTokenRepo() *mockRefreshTokenRepo {
	return &mockRefreshTokenRepo{tokens: make(map[string]*RefreshToken)}
}

func (m *mockRefreshTokenRepo) Create(ctx context.Context, userID string, token string, expiresAt time.Time) (*RefreshToken, error) {
	entity := &RefreshToken{
		ID:        "ref-1",
		UserID:    userID,
		Token:     hashToken(token),
		ExpiresAt: expiresAt,
		Revoked:   false,
		CreatedAt: time.Now(),
	}
	m.tokens[token] = entity
	return entity, nil
}

func (m *mockRefreshTokenRepo) FindByToken(ctx context.Context, token string) (*RefreshToken, error) {
	if t, ok := m.tokens[token]; ok {
		return t, nil
	}
	return nil, common.ErrNotFound
}

func (m *mockRefreshTokenRepo) Revoke(ctx context.Context, token string) error {
	if t, ok := m.tokens[token]; ok {
		if t.Revoked {
			return common.ErrUnauthorized
		}
		t.Revoked = true
		return nil
	}
	return common.ErrUnauthorized
}

func (m *mockRefreshTokenRepo) RevokeAllByUserID(ctx context.Context, userID string) error {
	for _, t := range m.tokens {
		if t.UserID == userID {
			t.Revoked = true
		}
	}
	return nil
}

func (m *mockRefreshTokenRepo) RotateToken(ctx context.Context, oldToken string, userID string, newToken string, expiresAt time.Time) (*RefreshToken, error) {
	if err := m.Revoke(ctx, oldToken); err != nil {
		return nil, err
	}
	return m.Create(ctx, userID, newToken, expiresAt)
}

func (m *mockRefreshTokenRepo) DeleteExpired(ctx context.Context) (int64, error) {
	var count int64
	now := time.Now()
	for k, t := range m.tokens {
		if t.ExpiresAt.Before(now) {
			delete(m.tokens, k)
			count++
		}
	}
	return count, nil
}

func TestAuthService(t *testing.T) {
	cfg := &config.Config{
		JWT: config.JWTConfig{
			Secret:             "test-secret-key-12345",
			AccessExpireMinute: 15,
			RefreshExpireDay:   7,
		},
	}
	userSvc := newMockUserService()
	refreshRepo := newMockRefreshTokenRepo()
	svc := NewService(cfg, userSvc, refreshRepo, nil)
	ctx := context.Background()

	t.Run("Register and Login flow", func(t *testing.T) {
		// Register
		u, err := svc.Register(ctx, "test@example.com", "password123", "Test User")
		if err != nil {
			t.Fatalf("unexpected error during register: %v", err)
		}
		if u.Email != "test@example.com" {
			t.Fatalf("expected email test@example.com, got %s", u.Email)
		}

		// Register duplicate email fails
		_, err = svc.Register(ctx, "test@example.com", "password123", "Test User")
		if err == nil {
			t.Fatal("expected conflict error for duplicate email")
		}

		// Login success
		loggedInUser, err := svc.Login(ctx, "test@example.com", "password123")
		if err != nil {
			t.Fatalf("unexpected error during login: %v", err)
		}
		if loggedInUser.ID != u.ID {
			t.Fatalf("expected user id %s, got %s", u.ID, loggedInUser.ID)
		}

		// Login with wrong password fails
		_, err = svc.Login(ctx, "test@example.com", "wrongpass")
		if err == nil {
			t.Fatal("expected unauthorized error for wrong password")
		}

		// Login with non-existent email fails safely
		_, err = svc.Login(ctx, "nonexistent@example.com", "password123")
		if err == nil {
			t.Fatal("expected unauthorized error for non-existent email")
		}
	})

	t.Run("Generate and Validate Token with Role", func(t *testing.T) {
		pair, err := svc.GenerateTokenPair(ctx, "user-1", "test@example.com", "admin")
		if err != nil {
			t.Fatalf("unexpected error generating token pair: %v", err)
		}

		claims, err := svc.ValidateToken(pair.AccessToken)
		if err != nil {
			t.Fatalf("unexpected error validating token: %v", err)
		}
		if claims.UserID != "user-1" || claims.Email != "test@example.com" || claims.Role != "admin" {
			t.Fatalf("claims mismatch: %+v", claims)
		}
	})

	t.Run("RefreshToken Rotation and Reuse Detection", func(t *testing.T) {
		pair, err := svc.GenerateTokenPair(ctx, "user-1", "test@example.com")
		if err != nil {
			t.Fatalf("failed to generate token pair: %v", err)
		}

		// Refresh token successfully
		newPair, err := svc.RefreshToken(ctx, pair.RefreshToken)
		if err != nil {
			t.Fatalf("failed to refresh token: %v", err)
		}
		if newPair.AccessToken == "" || newPair.RefreshToken == "" {
			t.Fatal("expected new token pair")
		}

		// Attempt to reuse old refresh token triggers Reuse Detection (revokes all user tokens)
		_, err = svc.RefreshToken(ctx, pair.RefreshToken)
		if err == nil {
			t.Fatal("expected error when reusing revoked refresh token")
		}

		// Because reuse detection revoked all tokens for user-1, newPair.RefreshToken should also be revoked
		_, err = svc.RefreshToken(ctx, newPair.RefreshToken)
		if err == nil {
			t.Fatal("expected newPair to be revoked due to reuse detection")
		}

		// Generate fresh pair and test Logout
		pair2, err := svc.GenerateTokenPair(ctx, "user-1", "test@example.com")
		if err != nil {
			t.Fatalf("failed to generate pair2: %v", err)
		}
		if err := svc.Logout(ctx, pair2.RefreshToken); err != nil {
			t.Fatalf("logout failed: %v", err)
		}
	})

	t.Run("LogoutAll and DeleteExpiredTokens", func(t *testing.T) {
		pair1, err := svc.GenerateTokenPair(ctx, "user-2", "user2@example.com")
		if err != nil {
			t.Fatalf("failed to generate token pair 1: %v", err)
		}
		pair2, err := svc.GenerateTokenPair(ctx, "user-2", "user2@example.com")
		if err != nil {
			t.Fatalf("failed to generate token pair 2: %v", err)
		}

		// Logout all for user-2
		if err := svc.LogoutAll(ctx, "user-2"); err != nil {
			t.Fatalf("LogoutAll failed: %v", err)
		}

		// Tokens should now fail to refresh
		if _, err := svc.RefreshToken(ctx, pair1.RefreshToken); err == nil {
			t.Fatal("expected pair1 to be revoked after LogoutAll")
		}
		if _, err := svc.RefreshToken(ctx, pair2.RefreshToken); err == nil {
			t.Fatal("expected pair2 to be revoked after LogoutAll")
		}

		// Test DeleteExpiredTokens
		refreshRepo.tokens["expired-token"] = &RefreshToken{
			ID:        "exp-1",
			UserID:    "user-3",
			Token:     hashToken("expired-token"),
			ExpiresAt: time.Now().Add(-1 * time.Hour),
		}
		deletedCount, err := svc.DeleteExpiredTokens(ctx)
		if err != nil {
			t.Fatalf("DeleteExpiredTokens failed: %v", err)
		}
		if deletedCount != 1 {
			t.Fatalf("expected 1 expired token deleted, got %d", deletedCount)
		}
	})
}
