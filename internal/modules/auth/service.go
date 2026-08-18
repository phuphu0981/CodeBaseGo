package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"codebasego/internal/common"
	"codebasego/internal/modules/user"
	"codebasego/internal/platform/config"
)

// Precomputed bcrypt dummy hash to prevent email enumeration timing attacks
const dummyBcryptHash = "$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy"

// Claims represents the JWT access token payload.
type Claims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

type UserService interface {
	Create(ctx context.Context, req *user.CreateRequest, hashedPassword string) (*user.User, error)
	GetByID(ctx context.Context, id string) (*user.User, error)
	GetByEmail(ctx context.Context, email string) (*user.User, error)
}

// Service handles authentication and token management logic.
type Service struct {
	cfg         *config.Config
	userService UserService
	refreshRepo RefreshTokenRepository
	eventBus    common.EventBus
}

func NewService(cfg *config.Config, userService UserService, refreshRepo RefreshTokenRepository, eventBus common.EventBus) *Service {
	return &Service{
		cfg:         cfg,
		userService: userService,
		refreshRepo: refreshRepo,
		eventBus:    eventBus,
	}
}

func (s *Service) Register(ctx context.Context, email, password, name string) (*user.User, error) {
	req := &user.CreateRequest{
		Email:    email,
		Password: password,
		Name:     name,
	}
	if err := req.Validate(); err != nil {
		return nil, err
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	entity, err := s.userService.Create(ctx, req, string(hashed))
	if err != nil {
		return nil, err
	}

	if s.eventBus != nil {
		s.eventBus.Publish(ctx, common.Event{
			Name: user.EventUserRegistered,
			Payload: user.UserRegisteredPayload{
				UserID:    entity.ID,
				Email:     entity.Email,
				Name:      entity.Name,
				CreatedAt: entity.CreatedAt,
			},
		})
	}

	return entity, nil
}

func (s *Service) Login(ctx context.Context, email, password string) (*user.User, error) {
	entity, err := s.userService.GetByEmail(ctx, email)
	if err != nil {
		// Prevent email enumeration via timing attack by performing a dummy bcrypt check
		_ = bcrypt.CompareHashAndPassword([]byte(dummyBcryptHash), []byte(password))
		return nil, common.ErrUnauthorized
	}

	if err := bcrypt.CompareHashAndPassword([]byte(entity.Password), []byte(password)); err != nil {
		return nil, common.ErrUnauthorized
	}

	return entity, nil
}

// GenerateTokenPair creates both an Access Token and a Refresh Token (persisted to DB).
func (s *Service) GenerateTokenPair(ctx context.Context, userID, email string, role ...string) (*TokenResponse, error) {
	userRole := "user"
	if len(role) > 0 && role[0] != "" {
		userRole = role[0]
	}

	// 1. Generate Access Token (JWT)
	accessExpireDuration := time.Duration(s.cfg.JWT.AccessExpireMinute) * time.Minute
	accessClaims := Claims{
		UserID: userID,
		Email:  email,
		Role:   userRole,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(accessExpireDuration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessTokenStr, err := token.SignedString([]byte(s.cfg.JWT.Secret))
	if err != nil {
		return nil, fmt.Errorf("failed to sign access token: %w", err)
	}

	// 2. Generate Refresh Token (opaque random string)
	refreshTokenStr := uuid.New().String() + "-" + uuid.New().String()
	refreshExpireDuration := time.Duration(s.cfg.JWT.RefreshExpireDay) * 24 * time.Hour
	refreshExpiresAt := time.Now().Add(refreshExpireDuration)

	// Save Refresh Token in Database
	if _, err := s.refreshRepo.Create(ctx, userID, refreshTokenStr, refreshExpiresAt); err != nil {
		return nil, fmt.Errorf("failed to store refresh token: %w", err)
	}

	return &TokenResponse{
		AccessToken:  accessTokenStr,
		RefreshToken: refreshTokenStr,
		ExpiresIn:    int(accessExpireDuration.Seconds()),
		TokenType:    "Bearer",
	}, nil
}

// RefreshToken validates an existing refresh token, revokes it (token rotation), and returns a new token pair atomically.
func (s *Service) RefreshToken(ctx context.Context, refreshTokenStr string) (*TokenResponse, error) {
	storedToken, err := s.refreshRepo.FindByToken(ctx, refreshTokenStr)
	if err != nil {
		return nil, common.ErrUnauthorized
	}

	if storedToken.Revoked {
		// Reuse Detection: If a revoked refresh token is presented after grace period (15s), revoke ALL tokens
		// If presented within 15s grace period (e.g. concurrent requests/retries), deny request without revoking all tokens.
		if time.Since(storedToken.UpdatedAt) >= 15*time.Second {
			_ = s.refreshRepo.RevokeAllByUserID(ctx, storedToken.UserID)
		}
		return nil, common.ErrUnauthorized
	}

	if time.Now().After(storedToken.ExpiresAt) {
		return nil, common.ErrUnauthorized
	}

	// Find associated user
	u, err := s.userService.GetByID(ctx, storedToken.UserID)
	if err != nil {
		return nil, common.ErrUnauthorized
	}

	userRole := u.Role
	if userRole == "" {
		userRole = "user"
	}

	// 1. Generate Access Token (JWT)
	accessExpireDuration := time.Duration(s.cfg.JWT.AccessExpireMinute) * time.Minute
	accessClaims := Claims{
		UserID: u.ID,
		Email:  u.Email,
		Role:   userRole,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(accessExpireDuration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessTokenStr, err := token.SignedString([]byte(s.cfg.JWT.Secret))
	if err != nil {
		return nil, fmt.Errorf("failed to sign access token: %w", err)
	}

	// 2. Generate new Refresh Token and rotate atomically in DB transaction
	newRefreshTokenStr := uuid.New().String() + "-" + uuid.New().String()
	refreshExpireDuration := time.Duration(s.cfg.JWT.RefreshExpireDay) * 24 * time.Hour
	refreshExpiresAt := time.Now().Add(refreshExpireDuration)

	if _, err := s.refreshRepo.RotateToken(ctx, refreshTokenStr, u.ID, newRefreshTokenStr, refreshExpiresAt); err != nil {
		return nil, common.ErrUnauthorized
	}

	return &TokenResponse{
		AccessToken:  accessTokenStr,
		RefreshToken: newRefreshTokenStr,
		ExpiresIn:    int(accessExpireDuration.Seconds()),
		TokenType:    "Bearer",
	}, nil
}

// Logout revokes the given refresh token.
func (s *Service) Logout(ctx context.Context, refreshTokenStr string) error {
	return s.refreshRepo.Revoke(ctx, refreshTokenStr)
}

// LogoutAll revokes all refresh tokens for a user.
func (s *Service) LogoutAll(ctx context.Context, userID string) error {
	return s.refreshRepo.RevokeAllByUserID(ctx, userID)
}

// DeleteExpiredTokens deletes all expired refresh tokens from database.
func (s *Service) DeleteExpiredTokens(ctx context.Context) (int64, error) {
	return s.refreshRepo.DeleteExpired(ctx)
}

func (s *Service) ValidateToken(tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.cfg.JWT.Secret), nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, common.ErrUnauthorized
	}

	return claims, nil
}
