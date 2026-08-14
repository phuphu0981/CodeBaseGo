package auth

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"time"
)

// RefreshTokenRepository defines the data access contract for refresh tokens.
type RefreshTokenRepository interface {
	Create(ctx context.Context, userID string, token string, expiresAt time.Time) (*RefreshToken, error)
	FindByToken(ctx context.Context, token string) (*RefreshToken, error)
	Revoke(ctx context.Context, token string) error
	RevokeAllByUserID(ctx context.Context, userID string) error
	RotateToken(ctx context.Context, oldToken string, userID string, newToken string, expiresAt time.Time) (*RefreshToken, error)
	DeleteExpired(ctx context.Context) (int64, error)
}

func hashToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}
